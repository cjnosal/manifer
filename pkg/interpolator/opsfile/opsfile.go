package opsfile

import (
	"fmt"

	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/interpolator"
	"github.com/cjnosal/manifer/pkg/yaml"
	boshopts "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/cppforlife/go-patch/patch"
	"github.com/jessevdk/go-flags"
)

func NewOpsFileInterpolator(y yaml.YamlAccess) interpolator.Interpolator {
	i := &ofInt{
		Yaml: y,
	}
	return &interpolatorWrapper{
		interpolator: i,
	}
}

type interpolatorWrapper struct {
	interpolator opFileInterpolator
}

type ofInt struct {
	Yaml yaml.YamlAccess
}

type opFileInterpolator interface {
	interpolate(template *file.TaggedBytes, snippet *file.TaggedBytes, args []string) ([]byte, error)
}

type opFlags struct {
	// flag string copied from bosh cli ops_flag.go
	Oppaths []string `long:"ops-file" short:"o" value-name:"PATH" description:"Load manifest operations from a YAML file"`
}

func (i *interpolatorWrapper) ParseSnippetFlags(args []string) ([]string, error) {
	opFlags := opFlags{}
	if len(args) > 0 {
		_, err := flags.NewParser(&opFlags, flags.IgnoreUnknown).ParseArgs(args)
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to parse opsfile args", err)
		}
	}
	return opFlags.Oppaths, nil
}

func (i *interpolatorWrapper) Interpolate(template *file.TaggedBytes, snippet *file.TaggedBytes, snippetArgs []string, templateArgs []string) ([]byte, error) {

	var intSnippet *file.TaggedBytes
	if snippet != nil {
		snippetBytes, err := i.interpolator.interpolate(snippet, nil, append(snippetArgs, templateArgs...))
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to interpolate snippet", err)
		}
		intSnippet = &file.TaggedBytes{
			Bytes: snippetBytes,
			Tag:   snippet.Tag,
		}
	}

	templateBytes, err := i.interpolator.interpolate(template, intSnippet, templateArgs)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to interpolate template", err)
	}

	return templateBytes, nil
}

func (i *ofInt) interpolate(templateBytes *file.TaggedBytes, snippetBytes *file.TaggedBytes, args []string) ([]byte, error) {
	template := boshtpl.NewTemplate(templateBytes.Bytes)

	boshOpts := boshopts.InterpolateOpts{}

	var vars boshtpl.Variables = boshtpl.StaticVariables{}
	if len(args) > 0 {
		_, err := flags.NewParser(&boshOpts, flags.None).ParseArgs(append(args, templateBytes.Tag)) // manifest path is a required flag in InterpolateOpts
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to parse args", err)
		}
		vars = boshOpts.VarFlags.AsVariables()
	}

	opDefs := []patch.OpDefinition{}
	ops := patch.Ops{}
	if snippetBytes != nil {
		err := i.Yaml.Unmarshal(snippetBytes.Bytes, &opDefs)
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to parse ops file %s", err, snippetBytes.Tag)
		}

		ops, err = patch.NewOpsFromDefinitions(opDefs)
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to create ops from definitions in %s", err, snippetBytes.Tag)
		}
	}
	if len(ops) == 0 {
		// add nil op so we can still interpolate variables
		ops = append(ops, nil)
	}

	var outBytes []byte
	var err error
	for i, op := range ops {
		outBytes, err = template.Evaluate(vars, op, boshtpl.EvaluateOpts{})
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to evaluate template %s with op %d from %s", err, templateBytes.Tag, i, snippetBytes.Tag)
		}
		template = boshtpl.NewTemplate(outBytes)
	}

	return outBytes, nil
}

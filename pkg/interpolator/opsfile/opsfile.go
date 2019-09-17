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

func NewOpsFileInterpolator(f file.FileAccess, y yaml.YamlAccess) interpolator.Interpolator {
	i := &ofInt{
		File: f,
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
	File file.FileAccess
	Yaml yaml.YamlAccess
}

type opFileInterpolator interface {
	interpolate(inPath string, outPath string, snippetPath string, originalSnippetPath string, args []string, includeOps bool) error
}

func (i *interpolatorWrapper) Interpolate(inPath string, outPath string, snippetPath string, snippetArgs []string, scenarioArgs []string) error {

	intSnippetPath := ""
	if snippetPath != "" {
		intSnippetPath = "/tmp/int_snippet.yml"
		err := i.interpolator.interpolate(snippetPath, intSnippetPath, "", "", append(snippetArgs, scenarioArgs...), false)
		if err != nil {
			return fmt.Errorf("%w\n  while trying to interpolate snippet", err)
		}
	}

	err := i.interpolator.interpolate(inPath, outPath, intSnippetPath, snippetPath, scenarioArgs, true)
	if err != nil {
		return fmt.Errorf("%w\n  while trying to interpolate template", err)
	}

	return nil
}

func (i *ofInt) interpolate(inPath string, outPath string, snippetPath string, originalSnippetPath string, args []string, includeOps bool) error {
	templateBytes, err := i.File.Read(inPath)
	if err != nil {
		return fmt.Errorf("%w\n  while trying to load %s", err, inPath)
	}
	template := boshtpl.NewTemplate(templateBytes)

	boshOpts := boshopts.InterpolateOpts{}

	var vars boshtpl.Variables = boshtpl.StaticVariables{}
	if len(args) > 0 {
		_, err = flags.NewParser(&boshOpts, flags.None).ParseArgs(append(args, inPath)) // manifest path is a required flag in InterpolateOpts
		if err != nil {
			return fmt.Errorf("%w\n  while trying to parse args", err)
		}
		vars = boshOpts.VarFlags.AsVariables()
	}

	opDefs := []patch.OpDefinition{}
	ops := patch.Ops{}
	if snippetPath != "" {
		bytes, err := i.File.Read(snippetPath)
		if err != nil {
			return fmt.Errorf("%w\n  while trying to load ops file %s", err, originalSnippetPath)
		}
		err = i.Yaml.Unmarshal(bytes, &opDefs)
		if err != nil {
			return fmt.Errorf("%w\n  while trying to parse ops file %s", err, originalSnippetPath)
		}

		ops, err = patch.NewOpsFromDefinitions(opDefs)
		if err != nil {
			return fmt.Errorf("%w\n  while trying to create ops from definitions in %s", err, originalSnippetPath)
		}
	}
	if len(ops) == 0 {
		// add nil op so we can still interpolate variables
		ops = append(ops, nil)
	}

	if includeOps {
		passthroughOps := boshOpts.OpsFlags.AsOp()
		ops = append(ops, passthroughOps)
	}

	var outBytes []byte
	for i, op := range ops {
		outBytes, err = template.Evaluate(vars, op, boshtpl.EvaluateOpts{})
		if err != nil {
			return fmt.Errorf("%w\n  while trying to evaluate template %s with op %d from %s", err, inPath, i, originalSnippetPath)
		}
		template = boshtpl.NewTemplate(outBytes)
	}

	err = i.File.Write(outPath, outBytes, 0644)
	if err != nil {
		return fmt.Errorf("%w\n  while trying to write interpolated file %s", err, outPath)
	}

	return nil
}

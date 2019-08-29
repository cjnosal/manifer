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
			return fmt.Errorf("Unable to interpolate snippet: %v", err.Error())
		}
	}

	err := i.interpolator.interpolate(inPath, outPath, intSnippetPath, snippetPath, scenarioArgs, true)
	if err != nil {
		return fmt.Errorf("Unable to interpolate template: %v", err.Error())
	}

	return nil
}

func (i *ofInt) interpolate(inPath string, outPath string, snippetPath string, originalSnippetPath string, args []string, includeOps bool) error {
	templateBytes, err := i.File.Read(inPath)
	if err != nil {
		return fmt.Errorf("Unable to load %s: %s", inPath, err.Error())
	}
	template := boshtpl.NewTemplate(templateBytes)

	boshOpts := boshopts.InterpolateOpts{}

	var vars boshtpl.Variables = boshtpl.StaticVariables{}
	if len(args) > 0 {
		_, err = flags.ParseArgs(&boshOpts, append(args, inPath)) // manifest path is a required flag in InterpolateOpts
		if err != nil {
			return fmt.Errorf("Unable to parse args: %s", err.Error())
		}
		vars = boshOpts.VarFlags.AsVariables()
	}

	opDefs := []patch.OpDefinition{}
	ops := patch.Ops{}
	if snippetPath != "" {

		err = i.Yaml.Load(snippetPath, &opDefs)
		if err != nil {
			return fmt.Errorf("Unable to load ops file %s: %s", originalSnippetPath, err.Error())
		}

		ops, err = patch.NewOpsFromDefinitions(opDefs)
		if err != nil {
			return fmt.Errorf("Unable to create ops from definitions in %s: %s", originalSnippetPath, err.Error())
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
			return fmt.Errorf("Unable to evaluate template %s with op %d from %s: %s", inPath, i, originalSnippetPath, err.Error())
		}
		template = boshtpl.NewTemplate(outBytes)
	}

	err = i.File.Write(outPath, outBytes, 0644)
	if err != nil {
		return fmt.Errorf("Unable to write interpolated file %s: %s", outPath, err.Error())
	}

	return nil
}

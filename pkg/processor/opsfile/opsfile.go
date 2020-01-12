package opsfile

import (
	"fmt"

	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/processor"
	"github.com/cjnosal/manifer/pkg/yaml"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/cppforlife/go-patch/patch"
	"github.com/jessevdk/go-flags"
)

func NewOpsFileProcessor(y yaml.YamlAccess, f file.FileAccess) processor.Processor {
	return &opFileProcessor{
		yaml: y,
		file: f,
	}
}

type opFileProcessor struct {
	yaml yaml.YamlAccess
	file file.FileAccess
}

type opFlags struct {
	// flag string copied from bosh cli ops_flag.go
	Oppaths []string `long:"ops-file" short:"o" value-name:"PATH" description:"Load manifest operations from a YAML file"`

	// from bosh cli opts.go
	Path string `long:"path" value-name:"OP-PATH" description:"Extract value out of template (e.g.: /private_key)"`
}

func (i *opFileProcessor) ValidateSnippet(path string) (bool, error) {
	content, err := i.file.Read(path)
	if err != nil {
		return false, fmt.Errorf("%w\n  while validating opsfile %s", err, path)
	}
	opDefs := []patch.OpDefinition{}
	err = i.yaml.Unmarshal(content, &opDefs)
	return err == nil, nil
}

func (i *opFileProcessor) ParsePassthroughFlags(args []string) (*library.ScenarioNode, []string, error) {
	var node *library.ScenarioNode
	opFlags := opFlags{}
	remainder, err := flags.NewParser(&opFlags, flags.IgnoreUnknown).ParseArgs(args)
	if err != nil {
		return nil, nil, fmt.Errorf("%w\n  while trying to parse opsfile args", err)
	}
	if len(opFlags.Oppaths) > 0 || opFlags.Path != "" {
		snippets := []library.Snippet{}
		for _, o := range opFlags.Oppaths {
			snippets = append(snippets, library.Snippet{
				Path: o,
				Processor: library.Processor{
					Type:    library.OpsFile,
					Options: map[string]interface{}{},
				},
			})
		}
		if opFlags.Path != "" {
			snippets = append(snippets, library.Snippet{
				Processor: library.Processor{
					Type: library.OpsFile,
					Options: map[string]interface{}{
						"path": opFlags.Path,
					},
				},
			})
		}
		node = &library.ScenarioNode{
			Name:        "passthrough opsfile",
			Description: "args passed after --",
			LibraryPath: "<cli>",
			Snippets:    snippets,
		}
	}
	return node, remainder, nil
}

func (i *opFileProcessor) ProcessTemplate(templateBytes *file.TaggedBytes, snippetBytes *file.TaggedBytes, options map[string]interface{}) ([]byte, error) {

	ops := patch.Ops{}
	if snippetBytes != nil {
		opDefs := []patch.OpDefinition{}
		err := i.yaml.Unmarshal(snippetBytes.Bytes, &opDefs)
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to parse ops file %s", err, snippetBytes.Tag)
		}

		ops, err = patch.NewOpsFromDefinitions(opDefs)
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to create ops from definitions in %s", err, snippetBytes.Tag)
		}
	}

	var findPath string
	if options != nil {
		pathInterface := options["path"]
		if pathInterface != nil {
			findPath = pathInterface.(string)
		}
	}

	if len(ops) == 0 && findPath == "" {
		return templateBytes.Bytes, nil
	}

	template := boshtpl.NewTemplate(templateBytes.Bytes)
	var outBytes []byte
	var err error
	for i, op := range ops {
		outBytes, err = template.Evaluate(boshtpl.StaticVariables{}, op, boshtpl.EvaluateOpts{})
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to evaluate template %s with op %d from %s", err, templateBytes.Tag, i, snippetBytes.Tag)
		}
		template = boshtpl.NewTemplate(outBytes)
	}
	if findPath != "" {
		pointer, err := patch.NewPointerFromString(findPath)
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to parse path %s in template %s", err, findPath, templateBytes.Tag)
		}
		evaluateOpts := boshtpl.EvaluateOpts{
			PostVarSubstitutionOp: patch.FindOp{Path: pointer},
			UnescapedMultiline:    true,
		}
		outBytes, err = template.Evaluate(boshtpl.StaticVariables{}, nil, evaluateOpts)
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to find path %s in template %s", err, findPath, templateBytes.Tag)
		}
		template = boshtpl.NewTemplate(outBytes)
	}

	return outBytes, nil
}

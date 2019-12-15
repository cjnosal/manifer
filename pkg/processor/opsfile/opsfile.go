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
	g := &opFileGenerator{
		yaml: y,
	}
	return &opFileProcessor{
		yaml:      y,
		file:      f,
		generator: g,
	}
}

type opFileProcessor struct {
	yaml      yaml.YamlAccess
	file      file.FileAccess
	generator *opFileGenerator
}

type opFlags struct {
	// flag string copied from bosh cli ops_flag.go
	Oppaths []string `long:"ops-file" short:"o" value-name:"PATH" description:"Load manifest operations from a YAML file"`
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

func (i *opFileProcessor) ParsePassthroughFlags(args []string) (*library.ScenarioNode, error) {
	var node *library.ScenarioNode
	if len(args) > 0 {
		opFlags := opFlags{}
		remainder, err := flags.NewParser(&opFlags, flags.IgnoreUnknown).ParseArgs(args)
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to parse opsfile args", err)
		}
		snippets := []library.Snippet{}
		for _, o := range opFlags.Oppaths {
			snippets = append(snippets, library.Snippet{
				Path: o,
				Args: []string{},
			})
		}
		node = &library.ScenarioNode{
			Name:        "passthrough",
			Description: "args passed after --",
			LibraryPath: "<cli>",
			Type:        string(library.OpsFile),
			GlobalArgs:  remainder,
			RefArgs:     []string{},
			Snippets:    snippets,
		}
	}
	return node, nil
}

func (i *opFileProcessor) ProcessTemplate(templateBytes *file.TaggedBytes, snippetBytes *file.TaggedBytes) ([]byte, error) {
	opDefs := []patch.OpDefinition{}
	ops := patch.Ops{}
	err := i.yaml.Unmarshal(snippetBytes.Bytes, &opDefs)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to parse ops file %s", err, snippetBytes.Tag)
	}

	ops, err = patch.NewOpsFromDefinitions(opDefs)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to create ops from definitions in %s", err, snippetBytes.Tag)
	}

	if len(ops) == 0 {
		return templateBytes.Bytes, nil
	}

	template := boshtpl.NewTemplate(templateBytes.Bytes)
	var outBytes []byte
	for i, op := range ops {
		outBytes, err = template.Evaluate(nil, op, boshtpl.EvaluateOpts{})
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to evaluate template %s with op %d from %s", err, templateBytes.Tag, i, snippetBytes.Tag)
		}
		template = boshtpl.NewTemplate(outBytes)
	}

	return outBytes, nil
}

func (i *opFileProcessor) GenerateSnippets(schema *yaml.SchemaNode) ([]*file.TaggedBytes, error) {
	return i.generator.generateSnippets(schema)
}

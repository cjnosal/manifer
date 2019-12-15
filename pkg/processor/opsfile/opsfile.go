package opsfile

import (
	"fmt"

	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/processor"
	"github.com/cjnosal/manifer/pkg/yaml"
	boshopts "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/cppforlife/go-patch/patch"
	"github.com/jessevdk/go-flags"
)

func NewOpsFileProcessor(y yaml.YamlAccess, f file.FileAccess) processor.Processor {
	i := &ofInt{
		Yaml: y,
		File: f,
	}
	g := &opFileGenerator{
		yaml: y,
	}
	return &processorWrapper{
		processor: i,
		file:      f,
		generator: g,
	}
}

type processorWrapper struct {
	processor opFileProcessor
	file      file.FileAccess
	generator *opFileGenerator
}

type ofInt struct {
	Yaml yaml.YamlAccess
	File file.FileAccess
}

type opFileProcessor interface {
	process(template *file.TaggedBytes, snippet *file.TaggedBytes, args []string) ([]byte, error)
}

type opFlags struct {
	// flag string copied from bosh cli ops_flag.go
	Oppaths []string `long:"ops-file" short:"o" value-name:"PATH" description:"Load manifest operations from a YAML file"`
}

func (i *processorWrapper) ValidateSnippet(path string) (bool, error) {
	content, err := i.file.Read(path)
	if err != nil {
		return false, fmt.Errorf("%w\n  while validating opsfile %s", err, path)
	}
	opDefs := []patch.OpDefinition{}
	err = i.processor.(*ofInt).Yaml.Unmarshal(content, &opDefs)
	return err == nil, nil
}

func (i *processorWrapper) ParsePassthroughFlags(args []string) (*library.ScenarioNode, error) {
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

func (i *processorWrapper) ProcessTemplate(template *file.TaggedBytes, snippet *file.TaggedBytes, snippetArgs []string, templateArgs []string) ([]byte, error) {

	var intSnippet *file.TaggedBytes
	if snippet != nil {
		snippetBytes, err := i.processor.process(snippet, nil, append(snippetArgs, templateArgs...))
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to process snippet", err)
		}
		intSnippet = &file.TaggedBytes{
			Bytes: snippetBytes,
			Tag:   snippet.Tag,
		}
	}

	templateBytes, err := i.processor.process(template, intSnippet, templateArgs)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to process template", err)
	}

	return templateBytes, nil
}

func (i *processorWrapper) GenerateSnippets(schema *yaml.SchemaNode) ([]*file.TaggedBytes, error) {
	return i.generator.generateSnippets(schema)
}

func (i *ofInt) process(templateBytes *file.TaggedBytes, snippetBytes *file.TaggedBytes, args []string) ([]byte, error) {
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
		// add nil op so we can still process variables
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

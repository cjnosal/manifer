package factory

import (
	"fmt"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/processor"
	"github.com/cjnosal/manifer/pkg/processor/opsfile"
	"github.com/cjnosal/manifer/pkg/yaml"
)

type ProcessorFactory interface {
	Create(t library.Type) (processor.Processor, error)
	CreateGenerator(t library.Type) (processor.SnippetGenerator, error)
}

type processorFactory struct {
	yaml yaml.YamlAccess
	file file.FileAccess
}

func NewProcessorFactory(yaml yaml.YamlAccess, file file.FileAccess) ProcessorFactory {
	return &processorFactory{
		yaml: yaml,
		file: file,
	}
}

func (i *processorFactory) Create(t library.Type) (processor.Processor, error) {
	if t == library.OpsFile {
		return opsfile.NewOpsFileProcessor(i.yaml, i.file), nil
	}
	return nil, fmt.Errorf("Unknown library type %v", t)
}

func (i *processorFactory) CreateGenerator(t library.Type) (processor.SnippetGenerator, error) {
	if t == library.OpsFile {
		return processor.NewSnippetGenerator(i.yaml, opsfile.NewPathBuilder()), nil
	}
	return nil, fmt.Errorf("Unknown library type %v", t)
}
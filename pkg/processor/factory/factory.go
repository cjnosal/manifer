package factory

import (
	"fmt"
	"github.com/cjnosal/manifer/v2/pkg/file"
	"github.com/cjnosal/manifer/v2/pkg/library"
	"github.com/cjnosal/manifer/v2/pkg/processor"
	"github.com/cjnosal/manifer/v2/pkg/processor/opsfile"
	"github.com/cjnosal/manifer/v2/pkg/processor/yq"
	"github.com/cjnosal/manifer/v2/pkg/yaml"
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
	} else if t == library.Yq {
		return yq.NewYqProcessor(i.yaml, i.file), nil
	}
	return nil, fmt.Errorf("Unknown library type %v", t)
}

func (i *processorFactory) CreateGenerator(t library.Type) (processor.SnippetGenerator, error) {
	if t == library.OpsFile {
		return processor.NewSnippetGenerator(i.yaml, opsfile.NewPathBuilder()), nil
	} else if t == library.Yq {
		return processor.NewSnippetGenerator(i.yaml, yq.NewPathBuilder()), nil
	}
	return nil, fmt.Errorf("Unknown library type %v", t)
}

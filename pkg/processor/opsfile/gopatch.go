package opsfile

import (
	"fmt"
	"github.com/cjnosal/manifer/pkg/processor"
	"github.com/cppforlife/go-patch/patch"
)

func NewPathBuilder() processor.PathBuilder {
	return &gopatchPathBuilder{}
}

type gopatchPathBuilder struct{}

func (pb *gopatchPathBuilder) Root() string {
	return "/"
}

func (pb *gopatchPathBuilder) Append() string {
	return "/-"
}

func (pb *gopatchPathBuilder) Index(index string) string {
	return fmt.Sprintf("/%s", index)
}

func (pb *gopatchPathBuilder) Delimiter() string {
	return "/"
}

func (pb *gopatchPathBuilder) Safe() string {
	return "?"
}

func (pb *gopatchPathBuilder) Marshal(path string, value interface{}) interface{} {
	return []patch.OpDefinition{
		{
			Path:  &path,
			Type:  "replace",
			Value: &value,
		},
	}
}

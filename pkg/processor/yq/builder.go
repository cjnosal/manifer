package yq

import (
	"fmt"
	"github.com/cjnosal/manifer/v2/pkg/processor"
)

func NewPathBuilder() processor.PathBuilder {
	return &yqPathBuilder{}
}

type yqPathBuilder struct{}

func (pb *yqPathBuilder) Root() string {
	return ""
}

func (pb *yqPathBuilder) Append() string {
	return "[+]"
}

func (pb *yqPathBuilder) Index(index string) string {
	return fmt.Sprintf("[%s]", index)
}

func (pb *yqPathBuilder) Delimiter() string {
	return "."
}

func (pb *yqPathBuilder) Safe() string {
	return ""
}

func (pb *yqPathBuilder) Marshal(path string, value interface{}) interface{} {
	return map[interface{}]interface{}{
		path: value,
	}
}

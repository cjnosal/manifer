package processor

import (
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/yaml"
)

type Processor interface {
	ValidateSnippet(path string) (bool, error)
	ParsePassthroughFlags(templateArgs []string) (*library.ScenarioNode, error)
	ProcessTemplate(template *file.TaggedBytes, snippet *file.TaggedBytes) ([]byte, error)
	GenerateSnippets(schema *yaml.SchemaNode) ([]*file.TaggedBytes, error)
}

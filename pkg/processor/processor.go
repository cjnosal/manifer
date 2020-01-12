package processor

import (
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/yaml"
)

type Processor interface {
	ValidateSnippet(path string) (bool, error)
	ParsePassthroughFlags(args []string) (*library.ScenarioNode, []string, error)
	ProcessTemplate(template *file.TaggedBytes, snippet *file.TaggedBytes, options map[string]interface{}) ([]byte, error)
}

type SnippetGenerator interface {
	GenerateSnippets(schema *yaml.SchemaNode) ([]*file.TaggedBytes, error)
}

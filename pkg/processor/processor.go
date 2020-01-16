package processor

import (
	"github.com/cjnosal/manifer/v2/pkg/file"
	"github.com/cjnosal/manifer/v2/pkg/library"
	"github.com/cjnosal/manifer/v2/pkg/yaml"
)

type Processor interface {
	ValidateSnippet(path string) (SnippetHint, error)
	ParsePassthroughFlags(args []string) (*library.ScenarioNode, []string, error)
	ProcessTemplate(template *file.TaggedBytes, snippet *file.TaggedBytes, options map[string]interface{}) ([]byte, error)
}

type SnippetGenerator interface {
	GenerateSnippets(schema *yaml.SchemaNode) ([]*file.TaggedBytes, error)
}

type SnippetHint struct {
	Valid   bool
	Element string
	Action  string
}

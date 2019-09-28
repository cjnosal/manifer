package interpolator

import (
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/library"
)

type Interpolator interface {
	ParsePassthroughFlags(templateArgs []string) (*library.ScenarioNode, error)
	Interpolate(template *file.TaggedBytes, snippet *file.TaggedBytes, snippetArgs []string, templateArgs []string) ([]byte, error)
}

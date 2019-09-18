package interpolator

import (
	"github.com/cjnosal/manifer/pkg/file"
)

type Interpolator interface {
	Interpolate(template *file.TaggedBytes, snippet *file.TaggedBytes, snippetArgs []string, templateArgs []string) ([]byte, error)
}

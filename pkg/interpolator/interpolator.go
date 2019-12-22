package interpolator

import (
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/library"
)

type Interpolator interface {
	Interpolate(templateBytes *file.TaggedBytes, args []string) ([]byte, error)
	ParsePassthroughVars(args []string) (*library.ScenarioNode, error)
}

package interpolator

import (
	"github.com/cjnosal/manifer/pkg/file"
)

type Interpolator interface {
	Interpolate(templateBytes *file.TaggedBytes, args []string) ([]byte, error)
}

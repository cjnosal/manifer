package plan

import (
	"fmt"
	"github.com/cjnosal/manifer/pkg/diff"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/interpolator"
	"io"
)

type Executor interface {
	Execute(showPlan bool, showDiff bool, template *file.TaggedBytes, snippet *file.TaggedBytes, snippetArgs []string, templateArgs []string) ([]byte, error)
}

type InterpolationExecutor struct {
	Interpolator interpolator.Interpolator
	Diff         diff.Diff
	Output       io.Writer
	File         file.FileAccess
}

func (i *InterpolationExecutor) Execute(showPlan bool, showDiff bool, template *file.TaggedBytes, snippet *file.TaggedBytes, snippetArgs []string, templateArgs []string) ([]byte, error) {
	var snippetPath string
	if snippet != nil {
		snippetPath = snippet.Tag
	}
	if showPlan {
		var relpath string
		var err error
		if snippet != nil {
			relpath, err = i.File.ResolveRelativeFromWD(snippetPath)
			if err != nil {
				return nil, fmt.Errorf("%w\n  while resolving relative snippet path %s", err, snippetPath)
			}
		}
		out := fmt.Sprintf("\nSnippet %s; Arg %v; Global %v\n", relpath, snippetArgs, templateArgs)
		i.Output.Write([]byte(out))
	}
	bytes, err := i.Interpolator.Interpolate(template, snippet, snippetArgs, templateArgs)
	if err != nil {
		return nil, err
	}
	if showDiff {
		i.Output.Write([]byte("Diff:\n"))
		diff := i.Diff.StringDiff(string(template.Bytes), string(bytes))
		i.Output.Write([]byte(diff))
	}
	return bytes, nil
}

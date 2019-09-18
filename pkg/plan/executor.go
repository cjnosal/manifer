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
}

func (i *InterpolationExecutor) Execute(showPlan bool, showDiff bool, template *file.TaggedBytes, snippet *file.TaggedBytes, snippetArgs []string, templateArgs []string) ([]byte, error) {
	var snippetPath string
	if snippet != nil {
		snippetPath = snippet.Tag
	}
	if showPlan {
		out := fmt.Sprintf("\nSnippet %s; Arg %v; Global %v\n", snippetPath, snippetArgs, templateArgs)
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

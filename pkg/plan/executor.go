package plan

import (
	"fmt"
	"github.com/cjnosal/manifer/pkg/diff"
	"github.com/cjnosal/manifer/pkg/interpolator"
	"io"
)

type Executor interface {
	Execute(showPlan bool, showDiff bool, inPath string, outPath string, snippetPath string, snippetArgs []string, scenarioArgs []string) error
}

type InterpolationExecutor struct {
	Interpolator interpolator.Interpolator
	Diff         diff.Diff
	Output       io.Writer
}

func (i *InterpolationExecutor) Execute(showPlan bool, showDiff bool, inPath string, outPath string, snippetPath string, snippetArgs []string, scenarioArgs []string) error {
	if showPlan {
		out := fmt.Sprintf("\nSnippet %s; Arg %v; Global %v\n", snippetPath, snippetArgs, scenarioArgs)
		i.Output.Write([]byte(out))
	}
	err := i.Interpolator.Interpolate(inPath, outPath, snippetPath, snippetArgs, scenarioArgs)
	if err != nil {
		return err
	}
	if showDiff {
		i.Output.Write([]byte("Diff:\n"))
		diff, err := i.Diff.FindDiff(inPath, outPath)
		if err != nil {
			return err
		}
		i.Output.Write([]byte(diff))
	}
	return nil
}

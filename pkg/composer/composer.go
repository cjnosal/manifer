package composer

import (
	"fmt"
	"github.com/cjnosal/manifer/v2/pkg/file"
	"github.com/cjnosal/manifer/v2/pkg/library"
	"github.com/cjnosal/manifer/v2/pkg/plan"
)

type Composer interface {
	Compose(
		template *file.TaggedBytes,
		libraryPaths []string,
		scenarioNames []string,
		passthrough []string,
		showPlan bool,
		showDiff bool) ([]byte, error)
}

type ComposerImpl struct {
	Executor plan.Executor
	Resolver ScenarioResolver
	File     file.FileAccess
}

func (c *ComposerImpl) Compose(
	template *file.TaggedBytes,
	libraryPaths []string,
	scenarioNames []string,
	passthrough []string,
	showPlan bool,
	showDiff bool) ([]byte, error) {

	plan, err := c.Resolver.Resolve(libraryPaths, scenarioNames, passthrough)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to resolve scenarios", err)
	}

	in := template
	var out []byte

	if len(plan.Steps) > 0 || !plan.Global.IsZero() {

		for _, step := range plan.Steps {
			var taggedSnippet *file.TaggedBytes
			if step.Snippet != "" {
				taggedSnippet, err = c.File.ReadAndTag(step.Snippet)
				if err != nil {
					return nil, fmt.Errorf("%w\n  while trying to load snippet %s", err, step.Snippet)
				}
			}
			out, err = c.Executor.Execute(showPlan, showDiff, in, taggedSnippet, &step.Processor, step.FlattenParams(), plan.Global)
			if err != nil {
				return nil, fmt.Errorf("%w\n  while trying to apply snippet %s", err, step.Snippet)
			}

			in = &file.TaggedBytes{Tag: in.Tag, Bytes: out}
		}

		if !plan.Global.IsZero() {
			out, err = c.Executor.Execute(showPlan, showDiff, in, nil, nil, library.InterpolatorParams{}, plan.Global)
			if err != nil {
				return nil, fmt.Errorf("%w\n  while trying to apply globals %+v", err, plan.Global)
			}
		}
	} else {
		out = in.Bytes
	}

	return out, nil
}

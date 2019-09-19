package composer

import (
	"fmt"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/plan"
)

type Composer interface {
	Compose(
		templatePath string,
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
	templatePath string,
	libraryPaths []string,
	scenarioNames []string,
	passthrough []string,
	showPlan bool,
	showDiff bool) ([]byte, error) {

	plan, err := c.Resolver.Resolve(libraryPaths, scenarioNames, passthrough)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to resolve scenarios", err)
	}

	in, err := c.File.ReadAndTag(templatePath)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to load template %s", err, templatePath)
	}
	var out []byte

	if len(plan.Snippets) > 0 || len(plan.GlobalArgs) > 0 {

		for _, snippet := range plan.Snippets {
			taggedSnippet, err := c.File.ReadAndTag(snippet.Path)
			if err != nil {
				return nil, fmt.Errorf("%w\n  while trying to load snippet %s", err, snippet.Path)
			}
			out, err = c.Executor.Execute(showPlan, showDiff, in, taggedSnippet, snippet.Args, plan.GlobalArgs)
			if err != nil {
				return nil, fmt.Errorf("%w\n  while trying to apply snippet %s", err, snippet.Path)
			}

			in = &file.TaggedBytes{Tag: in.Tag, Bytes: out}
		}

		if len(plan.GlobalArgs) > 0 {
			out, err = c.Executor.Execute(showPlan, showDiff, in, nil, nil, plan.GlobalArgs)
			if err != nil {
				return nil, fmt.Errorf("%w\n  while trying to apply passthrough args %v", err, plan.GlobalArgs)
			}
		}
	} else {
		out = in.Bytes
	}

	return out, nil
}

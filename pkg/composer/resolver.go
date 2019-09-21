package composer

import (
	"fmt"

	"github.com/cjnosal/manifer/pkg/interpolator"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/scenario"
)

type ScenarioResolver interface {
	Resolve(libPaths []string, scenarioNames []string, passthrough []string) (*scenario.Plan, error)
}

type Resolver struct {
	Loader          library.LibraryLoader
	Selector        scenario.ScenarioSelector
	SnippetResolver interpolator.Interpolator
}

func (r *Resolver) Resolve(libPaths []string, scenarioNames []string, passthrough []string) (*scenario.Plan, error) {
	libraries, err := r.Loader.Load(libPaths)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to load libraries", err)
	}

	plan, err := r.Selector.SelectScenarios(scenarioNames, libraries)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to select scenarios", err)
	}

	snippetPaths, err := r.SnippetResolver.ParseSnippetFlags(passthrough)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to resolve extra snippets", err)
	}
	for _, path := range snippetPaths {
		snippet := library.Snippet{Path: path}
		plan.Snippets = append(plan.Snippets, snippet)
	}

	plan.GlobalArgs = append(plan.GlobalArgs, passthrough...)

	return plan, nil
}

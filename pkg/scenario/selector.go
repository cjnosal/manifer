package scenario

import (
	"fmt"

	"github.com/cjnosal/manifer/pkg/library"
)

type Plan struct {
	GlobalArgs []string
	Snippets   []library.Snippet
}

type ScenarioSelector interface {
	SelectScenarios(scenarioNames []string, loaded *library.LoadedLibrary) (*Plan, error)
}

type Selector struct{}

func (selector *Selector) SelectScenarios(scenarioNames []string, loaded *library.LoadedLibrary) (*Plan, error) {
	plan := &Plan{
		GlobalArgs: []string{},
		Snippets:   []library.Snippet{},
	}

	for _, name := range scenarioNames {
		err := selector.updatePlan(name, loaded, plan, []string{}, nil)
		if err != nil {
			return nil, err
		}
	}

	return plan, nil
}

func (selector *Selector) updatePlan(scenarioName string, loaded *library.LoadedLibrary, plan *Plan, parentArgs []string, parentLib *library.Library) error {

	var rootScenario *library.Scenario
	var lib *library.Library
	if parentLib != nil {
		rootScenario, lib = loaded.GetScenarioFromLib(parentLib, scenarioName)
	} else {
		rootScenario, lib = loaded.GetScenario(scenarioName)
	}
	if rootScenario == nil {
		return fmt.Errorf("Unable to resolve scenario %s", scenarioName)
	}

	for _, dep := range rootScenario.Scenarios {
		args := dep.Args
		args = append(args, rootScenario.Args...)
		args = append(args, parentArgs...)
		err := selector.updatePlan(dep.Name, loaded, plan, args, lib)
		if err != nil {
			return fmt.Errorf("%w\n  while resolving scenario %s", err, scenarioName)
		}
	}

	plan.GlobalArgs = append(plan.GlobalArgs, rootScenario.GlobalArgs...)
	for _, snippet := range rootScenario.Snippets {
		args := snippet.Args
		args = append(args, rootScenario.Args...)
		args = append(args, parentArgs...)
		planSnippet := library.Snippet{
			Path: snippet.Path,
			Args: args,
		}
		plan.Snippets = append(plan.Snippets, planSnippet)
	}

	return nil
}

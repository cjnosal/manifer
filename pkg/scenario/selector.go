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
	SelectScenarios(scenarioNames []string, libraries []library.LoadedLibrary) (*Plan, error)
}

type Selector struct {
	Lookup library.LibraryLookup
}

func (selector *Selector) SelectScenarios(scenarioNames []string, libraries []library.LoadedLibrary) (*Plan, error) {
	plan := &Plan{
		GlobalArgs: []string{},
		Snippets:   []library.Snippet{},
	}

	for _, name := range scenarioNames {
		err := selector.updatePlan(name, libraries, plan, []string{})
		if err != nil {
			return nil, err
		}
	}

	return plan, nil
}

func (selector *Selector) updatePlan(scenarioName string, libraries []library.LoadedLibrary, plan *Plan, parentArgs []string) error {

	lib, err := selector.Lookup.GetContainingLibrary(scenarioName, libraries)
	if err != nil {
		return err
	}

	path := library.SplitName(scenarioName)
	name := path[len(path)-1]

	rootScenario := selector.getScenario(name, lib)

	for _, dep := range rootScenario.Scenarios {
		args := dep.Args
		args = append(args, rootScenario.Args...)
		args = append(args, parentArgs...)
		err := selector.updatePlan(dep.Name, []library.LoadedLibrary{*lib}, plan, args)
		if err != nil {
			return err
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

func (selector *Selector) getScenario(name string, lib *library.LoadedLibrary) *library.Scenario {
	for _, s := range lib.Library.Scenarios {
		if s.Name == name {
			return &s
		}
	}

	panic(fmt.Sprintf("Unable to find scenario %s in %s", name, lib.Path)) // library.Lookup returned wrong library
}

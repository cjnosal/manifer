package composer

import (
	"fmt"

	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/scenario"
)

type ScenarioResolver interface {
	Resolve(libPaths []string, scenarioNames []string, passthrough []string) (*scenario.Plan, error)
}

type Resolver struct {
	Loader   library.LibraryLoader
	Selector scenario.ScenarioSelector
}

func (r *Resolver) Resolve(libPaths []string, scenarioNames []string, passthrough []string) (*scenario.Plan, error) {
	libraries, err := r.Loader.Load(libPaths)
	if err != nil {
		return nil, fmt.Errorf("Unable to load libraries: %v", err)
	}

	plan, err := r.Selector.SelectScenarios(scenarioNames, libraries)
	if err != nil {
		return nil, fmt.Errorf("Unable to select scenarios: %v", err)
	}

	plan.GlobalArgs = append(plan.GlobalArgs, passthrough...)

	return plan, nil
}

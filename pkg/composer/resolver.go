package composer

import (
	"fmt"

	"github.com/cjnosal/manifer/pkg/interpolator"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/plan"
	"github.com/cjnosal/manifer/pkg/processor/factory"
)

type ScenarioResolver interface {
	Resolve(libPaths []string, scenarioNames []string, passthrough []string) (*plan.Plan, error)
}

type Resolver struct {
	Loader           library.LibraryLoader
	ProcessorFactory factory.ProcessorFactory
	Interpolator     interpolator.Interpolator
}

func (r *Resolver) Resolve(libPaths []string, scenarioNames []string, passthrough []string) (*plan.Plan, error) {
	libraries, err := r.Loader.Load(libPaths)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to load libraries", err)
	}

	nodes := library.ScenarioNodes{}
	for _, scenarioName := range scenarioNames {
		node, err := libraries.GetScenarioTree(scenarioName)
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to get scenario tree for %s", err, scenarioName)
		}
		nodes = append(nodes, node)
	}

	for _, t := range library.Types {
		processor, err := r.ProcessorFactory.Create(t)
		if err != nil {
			return nil, err
		}
		passthroughNode, remainder, err := processor.ParsePassthroughFlags(passthrough)
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to parse %s passthrough args", err, t)
		}
		passthrough = remainder
		if passthroughNode != nil {
			nodes = append(nodes, passthroughNode)
		}
	}
	passthroughVars, remainder, err := r.Interpolator.ParsePassthroughVars(passthrough)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to parse passthrough vars", err)
	}
	if passthroughVars != nil {
		nodes = append(nodes, passthroughVars)
	}
	if len(remainder) > 0 {
		return nil, fmt.Errorf("Invalid passthrough arguments %v", remainder)
	}
	executionPlan := &plan.Plan{
		Global: library.InterpolatorParams{
			Vars:      map[string]interface{}{},
			VarFiles:  map[string]string{},
			VarsFiles: []string{},
			VarsEnv:   []string{},
			VarsStore: "",
			RawArgs:   []string{},
		},
		Steps: []*plan.Step{},
	}
	for _, node := range nodes {
		executionPlan = plan.Append(executionPlan, plan.FromScenarioTree(node))
	}

	return executionPlan, nil
}

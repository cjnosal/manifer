package composer

import (
	"fmt"

	"github.com/cjnosal/manifer/pkg/interpolator"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/plan"
)

type ScenarioResolver interface {
	Resolve(libPaths []string, scenarioNames []string, passthrough []string) (*plan.Plan, error)
}

type Resolver struct {
	Loader          library.LibraryLoader
	SnippetResolver interpolator.Interpolator
}

func (r *Resolver) Resolve(libPaths []string, scenarioNames []string, passthrough []string) (*plan.Plan, error) {
	libraries, err := r.Loader.Load(libPaths)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to load libraries", err)
	}

	nodes := []*library.ScenarioNode{}
	for _, scenarioName := range scenarioNames {
		node, err := libraries.GetScenarioTree(scenarioName)
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to get scenario tree for %s", err, scenarioName)
		}
		nodes = append(nodes, node)
	}
	passthroughNode, err := r.SnippetResolver.ParsePassthroughFlags(passthrough)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to parse passthrough args", err)
	}
	if passthroughNode != nil {
		nodes = append(nodes, passthroughNode)
	}
	executionPlan := &plan.Plan{
		Global: plan.ArgSet{
			Tag:  "global",
			Args: []string{},
		},
		Steps: []*plan.Step{},
	}
	for _, node := range nodes {
		executionPlan = plan.Append(executionPlan, plan.FromScenarioTree(node))
	}

	return executionPlan, nil
}

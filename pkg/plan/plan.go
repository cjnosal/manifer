package plan

import (
	"github.com/cjnosal/manifer/pkg/library"
)

type Plan struct {
	Global ArgSet
	Steps  []*Step
}

func Append(a *Plan, b *Plan) *Plan {
	return &Plan{
		Global: ArgSet{
			Tag:  "global",
			Args: append(a.Global.Args, b.Global.Args...),
		},
		Steps: append(a.Steps, b.Steps...),
	}
}

type Step struct {
	Snippet string
	Args    []ArgSet
}

type ArgSet struct {
	Tag  string
	Args []string
}

func FromScenarioTree(node *library.ScenarioNode) *Plan {
	plan := &Plan{
		Global: ArgSet{
			Tag:  "global",
			Args: []string{},
		},
		Steps: []*Step{},
	}
	fromNode(node, plan, []ArgSet{})
	return plan
}

func fromNode(node *library.ScenarioNode, plan *Plan, inherited []ArgSet) {
	scenarioArgs := ArgSet{
		Tag:  node.Name,
		Args: append(node.Args, node.RefArgs...),
	}
	newArgSet := append([]ArgSet{scenarioArgs}, inherited...)
	for _, dep := range node.Dependencies {
		fromNode(dep, plan, newArgSet)
	}
	plan.Global.Args = append(plan.Global.Args, node.GlobalArgs...)
	for _, snippet := range node.Snippets {
		snippetArgs := ArgSet{
			Tag:  "snippet",
			Args: snippet.Args,
		}
		plan.Steps = append(plan.Steps, &Step{
			Snippet: snippet.Path,
			Args:    append([]ArgSet{snippetArgs}, newArgSet...),
		})
	}

}

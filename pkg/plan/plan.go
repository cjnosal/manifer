package plan

import (
	"github.com/cjnosal/manifer/pkg/library"
)

type Plan struct {
	Global library.InterpolatorParams
	Steps  []*Step
}

func Append(a *Plan, b *Plan) *Plan {
	return &Plan{
		Global: a.Global.Merge(b.Global),
		Steps:  append(a.Steps, b.Steps...),
	}
}

type Step struct {
	Snippet   string
	Params    []TaggedParams
	Processor library.Processor
}

func (s *Step) FlattenParams() library.InterpolatorParams {
	args := library.InterpolatorParams{
		Vars:      map[string]interface{}{},
		VarFiles:  map[string]string{},
		VarsFiles: []string{},
		VarsEnv:   []string{},
		VarsStore: "",
		RawArgs:   []string{},
	}
	for _, set := range s.Params {
		args = args.Merge(set.Params)
	}
	return args
}

type TaggedParams struct {
	Tag    string
	Params library.InterpolatorParams
}

func FromScenarioTree(node *library.ScenarioNode) *Plan {
	plan := &Plan{
		Global: library.InterpolatorParams{
			Vars:      map[string]interface{}{},
			VarFiles:  map[string]string{},
			VarsFiles: []string{},
			VarsEnv:   []string{},
			VarsStore: "",
			RawArgs:   []string{},
		},
		Steps: []*Step{},
	}
	fromNode(node, plan, []TaggedParams{})
	return plan
}

func fromNode(node *library.ScenarioNode, plan *Plan, inherited []TaggedParams) {
	scenarioParams := TaggedParams{
		Tag:    node.Name,
		Params: node.Interpolator.Merge(node.RefInterpolator),
	}
	newTaggedParams := append([]TaggedParams{scenarioParams}, inherited...)
	for _, dep := range node.Dependencies {
		fromNode(dep, plan, newTaggedParams)
	}
	plan.Global = plan.Global.Merge(node.GlobalInterpolator)
	for _, snippet := range node.Snippets {
		snippetParams := TaggedParams{
			Tag:    "snippet",
			Params: snippet.Interpolator,
		}
		plan.Steps = append(plan.Steps, &Step{
			Snippet:   snippet.Path,
			Params:    append([]TaggedParams{snippetParams}, newTaggedParams...),
			Processor: snippet.Processor,
		})
	}

}

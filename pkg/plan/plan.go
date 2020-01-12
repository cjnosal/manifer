package plan

import (
	"github.com/cjnosal/manifer/v2/pkg/library"
)

type Plan struct {
	Global library.InterpolatorParams `yaml:"global,omitempty"`
	Steps  []*Step                    `yaml:"steps,omitempty"`
}

func Append(a *Plan, b *Plan) *Plan {
	return &Plan{
		Global: a.Global.Merge(b.Global),
		Steps:  append(a.Steps, b.Steps...),
	}
}

type Step struct {
	Snippet   string            `yaml:"snippet,omitempty"`
	Params    []TaggedParams    `yaml:"params,omitempty"`
	Processor library.Processor `yaml:"processor,omitempty"`
}

func (s *Step) FlattenParams() library.InterpolatorParams {
	intParams := library.InterpolatorParams{
		Vars:      map[string]interface{}{},
		VarFiles:  map[string]string{},
		VarsFiles: []string{},
		VarsEnv:   []string{},
		VarsStore: "",
		RawArgs:   []string{},
	}
	for _, tp := range s.Params {
		intParams = intParams.Merge(tp.Interpolator)
	}
	return intParams
}

type TaggedParams struct {
	Tag          string                     `yaml:"tag,omitempty"`
	Interpolator library.InterpolatorParams `yaml:"interpolator,omitempty"`
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
		Tag:          node.Name,
		Interpolator: node.Interpolator.Merge(node.RefInterpolator),
	}
	newTaggedParams := append([]TaggedParams{scenarioParams}, inherited...)
	for _, dep := range node.Dependencies {
		fromNode(dep, plan, newTaggedParams)
	}
	plan.Global = plan.Global.Merge(node.GlobalInterpolator)
	for _, snippet := range node.Snippets {
		snippetParams := TaggedParams{
			Tag:          "snippet",
			Interpolator: snippet.Interpolator,
		}
		plan.Steps = append(plan.Steps, &Step{
			Snippet:   snippet.Path,
			Params:    append([]TaggedParams{snippetParams}, newTaggedParams...),
			Processor: snippet.Processor,
		})
	}

}

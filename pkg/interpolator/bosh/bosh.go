package bosh

import (
	"fmt"

	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/interpolator"
	"github.com/cjnosal/manifer/pkg/library"
	boshopts "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/jessevdk/go-flags"
)

func NewBoshInterpolator() interpolator.Interpolator {
	return &boshInterpolator{}
}

type boshInterpolator struct{}

func (i *boshInterpolator) Interpolate(templateBytes *file.TaggedBytes, args []string) ([]byte, error) {
	if len(args) == 0 {
		return templateBytes.Bytes, nil
	}

	template := boshtpl.NewTemplate(templateBytes.Bytes)

	boshOpts := boshopts.InterpolateOpts{}

	var vars boshtpl.Variables = boshtpl.StaticVariables{}
	_, err := flags.NewParser(&boshOpts, flags.None).ParseArgs(append(args, templateBytes.Tag)) // manifest path is a required flag in InterpolateOpts
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to parse args", err)
	}
	vars = boshOpts.VarFlags.AsVariables()

	var outBytes []byte
	outBytes, err = template.Evaluate(vars, nil, boshtpl.EvaluateOpts{})
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to evaluate template %s", err, templateBytes.Tag)
	}

	return outBytes, nil
}

func (i *boshInterpolator) ParsePassthroughVars(args []string) (*library.ScenarioNode, error) {
	var node *library.ScenarioNode
	if len(args) > 0 {
		varFlags := boshopts.VarFlags{}
		remainder, err := flags.NewParser(&varFlags, flags.IgnoreUnknown).ParseArgs(args)
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to parse args", err)
		}
		varsArgs := remove(args, remainder)
		if len(varsArgs) > 0 {
			node = &library.ScenarioNode{
				Name:        "passthrough variables",
				Description: "vars passed after --",
				LibraryPath: "<cli>",
				Type:        "",
				GlobalArgs:  varsArgs,
				RefArgs:     []string{},
				Snippets:    []library.Snippet{},
			}
		}
	}
	return node, nil
}

func remove(source []string, discard []string) []string {
	result := []string{}
	for _, s := range source {
		if !contains(discard, s) {
			result = append(result, s)
		}
	}
	return result
}

func contains(collection []string, value string) bool {
	for _, c := range collection {
		if c == value {
			return true
		}
	}
	return false
}

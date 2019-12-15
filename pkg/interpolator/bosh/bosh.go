package bosh

import (
	"fmt"

	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/interpolator"
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

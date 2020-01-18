package bosh

import (
	"fmt"

	"github.com/cjnosal/manifer/v2/pkg/file"
	"github.com/cjnosal/manifer/v2/pkg/interpolator"
	"github.com/cjnosal/manifer/v2/pkg/library"
	boshopts "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/jessevdk/go-flags"
)

func NewBoshInterpolator() interpolator.Interpolator {
	return &boshInterpolator{}
}

type boshInterpolator struct{}

func (i *boshInterpolator) Interpolate(templateBytes *file.TaggedBytes, params library.InterpolatorParams) ([]byte, error) {
	if params.IsZero() {
		return templateBytes.Bytes, nil
	}

	libKVs := []boshtpl.VarKV{}
	for k, v := range params.Vars {
		libKVs = append(libKVs, boshtpl.VarKV{
			Name:  k,
			Value: v,
		})
	}
	libVarFiles := []boshtpl.VarFileArg{}
	for v, p := range params.VarFiles {
		f := &boshtpl.VarFileArg{}
		err := f.UnmarshalFlag(fmt.Sprintf("%s=%s", v, p))
		if err != nil {
			return nil, fmt.Errorf("%w\n  unmarshaling var file %s", err, p)
		}
		libVarFiles = append(libVarFiles, *f)
	}
	libVarsFiles := []boshtpl.VarsFileArg{}
	for _, p := range params.VarsFiles {
		f := &boshtpl.VarsFileArg{}
		err := f.UnmarshalFlag(p)
		if err != nil {
			return nil, fmt.Errorf("%w\n  unmarshaling vars file %s", err, p)
		}
		libVarsFiles = append(libVarsFiles, *f)
	}
	libVarsEnv := []boshtpl.VarsEnvArg{}
	for _, p := range params.VarsEnv {
		e := &boshtpl.VarsEnvArg{}
		err := e.UnmarshalFlag(p)
		if err != nil {
			return nil, fmt.Errorf("%w\n  unmarshaling vars env %s", err, p)
		}
		libVarsEnv = append(libVarsEnv, *e)
	}
	libStore := &boshopts.VarsFSStore{}
	if params.VarsStore != "" {
		err := libStore.UnmarshalFlag(params.VarsStore)
		if err != nil {
			return nil, fmt.Errorf("%w\n  unmarshaling vars store %s", err, params.VarsStore)
		}
	}

	libVarFlags := boshopts.VarFlags{
		VarKVs:      libKVs,
		VarFiles:    libVarFiles,
		VarsFiles:   libVarsFiles,
		VarsEnvs:    libVarsEnv,
		VarsFSStore: *libStore,
	}
	passthroughVarFlags := boshopts.VarFlags{}
	_, err := flags.NewParser(&passthroughVarFlags, flags.None).ParseArgs(params.RawArgs)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to parse vars", err)
	}

	libVarFlags.VarKVs = append(libVarFlags.VarKVs, passthroughVarFlags.VarKVs...)
	libVarFlags.VarFiles = append(libVarFlags.VarFiles, passthroughVarFlags.VarFiles...)
	libVarFlags.VarsFiles = append(libVarFlags.VarsFiles, passthroughVarFlags.VarsFiles...)
	libVarFlags.VarsEnvs = append(libVarFlags.VarsEnvs, passthroughVarFlags.VarsEnvs...)
	if passthroughVarFlags.VarsFSStore.IsSet() {
		libVarFlags.VarsFSStore = passthroughVarFlags.VarsFSStore
	}

	boshVars := libVarFlags.AsVariables()

	template := boshtpl.NewTemplate(templateBytes.Bytes)

	outBytes, err := template.Evaluate(boshVars, nil, boshtpl.EvaluateOpts{})
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to evaluate template %s", err, templateBytes.Tag)
	}

	return outBytes, nil
}

func (i *boshInterpolator) ParsePassthroughVars(args []string) (*library.ScenarioNode, []string, error) {
	varFlags := boshopts.VarFlags{}
	remainder, err := flags.NewParser(&varFlags, flags.IgnoreUnknown).ParseArgs(args)
	if err != nil {
		return nil, nil, fmt.Errorf("%w\n  while trying to parse vars", err)
	}
	varsArgs := remove(args, remainder)
	params := library.InterpolatorParams{
		RawArgs: varsArgs,
	}

	var node *library.ScenarioNode
	if !params.IsZero() {
		node = &library.ScenarioNode{
			Name:               "passthrough variables",
			Description:        "vars passed after --",
			LibraryPath:        "<cli>",
			GlobalInterpolator: params,
		}
	}
	return node, remainder, nil
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

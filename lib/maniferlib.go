package lib

import (
	"github.com/sergi/go-diff/diffmatchpatch"
	"io"

	"github.com/cjnosal/manifer/pkg/composer"
	"github.com/cjnosal/manifer/pkg/diff"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/interpolator/opsfile"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/plan"
	"github.com/cjnosal/manifer/pkg/scenario"
	"github.com/cjnosal/manifer/pkg/yaml"
)

// logger used for Composer's showDiff/showPlan
func NewManifer(logger io.Writer) Manifer {
	return &libImpl{
		composer: newComposer(logger),
		lister:   newLister(),
	}
}

type Manifer interface {
	Compose(
		templatePath string,
		libraryPaths []string,
		scenarioNames []string,
		passthrough []string,
		showPlan bool,
		showDiff bool) ([]byte, error)

	ListScenarios(libraryPaths []string, all bool) ([]scenario.ScenarioEntry, error)
}

type libImpl struct {
	composer composer.Composer
	lister   scenario.ScenarioLister
}

func (l *libImpl) Compose(
	templatePath string,
	libraryPaths []string,
	scenarioNames []string,
	passthrough []string,
	showPlan bool,
	showDiff bool) ([]byte, error) {
	return l.composer.Compose(templatePath, libraryPaths, scenarioNames, passthrough, showPlan, showDiff)
}

func (l *libImpl) ListScenarios(libraryPaths []string, all bool) ([]scenario.ScenarioEntry, error) {
	return l.lister.ListScenarios(libraryPaths, all)
}

func newLister() scenario.ScenarioLister {
	file := &file.FileIO{}
	yaml := &yaml.Yaml{}

	loader := &library.Loader{
		File: file,
		Yaml: yaml,
	}
	return &scenario.Lister{
		Loader: loader,
	}
}

func newComposer(logger io.Writer) composer.Composer {
	file := &file.FileIO{}
	yaml := &yaml.Yaml{}
	patch := diffmatchpatch.New()
	diff := &diff.FileDiff{
		File:  file,
		Patch: patch,
	}
	lookup := &library.Lookup{}
	selector := &scenario.Selector{
		Lookup: lookup,
	}
	loader := &library.Loader{
		File: file,
		Yaml: yaml,
	}
	resolver := &composer.Resolver{
		Loader:   loader,
		Selector: selector,
	}
	opsFileInterpolator := opsfile.NewOpsFileInterpolator(yaml)
	opsFileExecutor := &plan.InterpolationExecutor{
		Interpolator: opsFileInterpolator,
		Diff:         diff,
		Output:       logger,
	}
	return &composer.ComposerImpl{
		Resolver: resolver,
		File:     file,
		Executor: opsFileExecutor,
	}
}

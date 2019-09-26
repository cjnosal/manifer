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
		loader:   newLoader(),
		file:     &file.FileIO{},
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

	GetScenarioTree(libraryPaths []string, name string) (*library.ScenarioNode, error)
}

type libImpl struct {
	composer composer.Composer
	lister   scenario.ScenarioLister
	loader   *library.Loader
	file     *file.FileIO
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

func (l *libImpl) GetScenarioTree(libraryPaths []string, name string) (*library.ScenarioNode, error) {
	loaded, err := l.loader.Load(libraryPaths)
	if err != nil {
		return nil, err
	}
	node, err := loaded.GetScenarioTree(name)
	if err != nil {
		return nil, err
	}
	err = l.makePathsRelative(node)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (l *libImpl) makePathsRelative(node *library.ScenarioNode) error {
	for i, snippet := range node.Snippets {
		rel, err := l.file.ResolveRelativeFromWD(snippet.Path)
		if err != nil {
			return err
		}
		node.Snippets[i].Path = rel
	}
	for _, dep := range node.Dependencies {
		err := l.makePathsRelative(dep)
		if err != nil {
			return err
		}
	}
	rel, err := l.file.ResolveRelativeFromWD(node.LibraryPath)
	if err != nil {
		return err
	}
	node.LibraryPath = rel
	return nil
}

func newLoader() *library.Loader {
	file := &file.FileIO{}
	yaml := &yaml.Yaml{}

	loader := &library.Loader{
		File: file,
		Yaml: yaml,
	}
	return loader
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
	selector := &scenario.Selector{}
	loader := &library.Loader{
		File: file,
		Yaml: yaml,
	}
	opsFileInterpolator := opsfile.NewOpsFileInterpolator(yaml)
	resolver := &composer.Resolver{
		Loader:          loader,
		Selector:        selector,
		SnippetResolver: opsFileInterpolator,
	}
	opsFileExecutor := &plan.InterpolationExecutor{
		Interpolator: opsFileInterpolator,
		Diff:         diff,
		Output:       logger,
		File:         file,
	}
	return &composer.ComposerImpl{
		Resolver: resolver,
		File:     file,
		Executor: opsFileExecutor,
	}
}

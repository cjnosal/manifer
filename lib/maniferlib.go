package lib

import (
	"fmt"
	y "gopkg.in/yaml.v3"
	"io"
	"path/filepath"

	"github.com/cjnosal/manifer/pkg/composer"
	"github.com/cjnosal/manifer/pkg/diff"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/importer"
	"github.com/cjnosal/manifer/pkg/interpolator"
	"github.com/cjnosal/manifer/pkg/interpolator/bosh"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/plan"
	"github.com/cjnosal/manifer/pkg/processor"
	"github.com/cjnosal/manifer/pkg/processor/opsfile"
	"github.com/cjnosal/manifer/pkg/scenario"
	"github.com/cjnosal/manifer/pkg/yaml"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// logger used for Composer's showDiff/showPlan
func NewManifer(logger io.Writer) Manifer {
	fileIO := &file.FileIO{}
	yaml := &yaml.Yaml{}
	opsFileProcessor := opsfile.NewOpsFileProcessor(yaml, fileIO)
	return &libImpl{
		composer:     newComposer(logger),
		lister:       newLister(),
		loader:       newLoader(),
		file:         fileIO,
		yaml:         yaml,
		opProc:       opsFileProcessor,
		importer:     importer.NewImporter(fileIO, opsFileProcessor),
		interpolator: bosh.NewBoshInterpolator(),
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

	ComposeFromBytes(
		template *file.TaggedBytes,
		libraryPaths []string,
		scenarioNames []string,
		passthrough []string,
		showPlan bool,
		showDiff bool) ([]byte, error)

	ListScenarios(libraryPaths []string, all bool) ([]scenario.ScenarioEntry, error)

	GetScenarioTree(libraryPaths []string, name string) (*library.ScenarioNode, error)

	GetScenarioNode(passthroughArgs []string) (*library.ScenarioNode, error)

	GetVarScenarioNode(passthroughArgs []string) (*library.ScenarioNode, error)

	Generate(libType library.Type, templatePath string, libPath string, snippetDir string) (*library.Library, error)

	Import(libType library.Type, path string, recursive bool, outPath string) (*library.Library, error)

	AddScenario(libraryPath string, name string, description string, scenarioDeps []string, passthrough []string) (*library.Library, error)
}

type libImpl struct {
	composer     composer.Composer
	lister       scenario.ScenarioLister
	loader       *library.Loader
	file         *file.FileIO
	yaml         yaml.YamlAccess
	opProc       processor.Processor
	importer     importer.Importer
	interpolator interpolator.Interpolator
}

func (l *libImpl) Compose(
	templatePath string,
	libraryPaths []string,
	scenarioNames []string,
	passthrough []string,
	showPlan bool,
	showDiff bool) ([]byte, error) {

	in, err := l.file.ReadAndTag(templatePath)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to load template %s", err, templatePath)
	}
	return l.ComposeFromBytes(in, libraryPaths, scenarioNames, passthrough, showPlan, showDiff)
}

func (l *libImpl) ComposeFromBytes(
	template *file.TaggedBytes,
	libraryPaths []string,
	scenarioNames []string,
	passthrough []string,
	showPlan bool,
	showDiff bool) ([]byte, error) {
	return l.composer.Compose(template, libraryPaths, scenarioNames, passthrough, showPlan, showDiff)
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

func (l *libImpl) GetScenarioNode(passthroughArgs []string) (*library.ScenarioNode, error) {
	return l.opProc.ParsePassthroughFlags(passthroughArgs)
}

func (l *libImpl) GetVarScenarioNode(passthroughArgs []string) (*library.ScenarioNode, error) {
	return l.interpolator.ParsePassthroughVars(passthroughArgs)
}

func (l *libImpl) Generate(libType library.Type, templatePath string, libPath string, snippetDir string) (*library.Library, error) {
	// parse template
	templateBytes, err := l.file.Read(templatePath)
	if err != nil {
		return nil, err
	}
	node := &y.Node{}
	err = l.yaml.Unmarshal(templateBytes, node)
	if err != nil {
		return nil, err
	}
	schemaBuilder := &yaml.SchemaBuilder{}

	// generate snippets from schema
	err = l.yaml.Walk(node, schemaBuilder.OnVisit)
	if err != nil {
		return nil, err
	}
	ops, err := l.opProc.GenerateSnippets(schemaBuilder.Root)
	if err != nil {
		return nil, err
	}

	// write snippets
	for _, op := range ops {
		snippetPath := filepath.Join(snippetDir, op.Tag)
		dir := filepath.Dir(snippetPath)
		err = l.file.MkDir(dir)
		if err != nil {
			return nil, err
		}
		err = l.file.Write(snippetPath, op.Bytes, 0644)
		if err != nil {
			return nil, err
		}
	}

	return l.Import(libType, snippetDir, true, libPath)
}

func (l *libImpl) Import(libType library.Type, path string, recursive bool, outPath string) (*library.Library, error) {
	return l.importer.Import(libType, path, recursive, outPath)
}

func (l *libImpl) AddScenario(libraryPath string, name string, description string, scenarioDeps []string, passthrough []string) (*library.Library, error) {
	loaded, err := l.loader.Load([]string{libraryPath})
	if err != nil {
		return nil, err
	}
	lib := loaded.TopLibraries[0]

	refs := []library.ScenarioRef{}
	for _, dep := range scenarioDeps {
		_, err = loaded.GetScenarioTree(dep)
		if err != nil {
			return nil, err
		}
		refs = append(refs, library.ScenarioRef{
			Name: dep,
			Args: []string{},
		})
	}

	node, err := l.GetScenarioNode(passthrough)
	if err != nil {
		return nil, err
	}
	varnode, err := l.GetVarScenarioNode(passthrough)
	if err != nil {
		return nil, err
	}
	args := []string{}
	snippets := []library.Snippet{}
	if node != nil {
		snippets = node.Snippets
	}
	if varnode != nil {
		args = varnode.GlobalArgs // passthrough node treats all variables as global but library scenarios need appropriate scope
	}

	scenario := library.Scenario{
		Name:        name,
		Description: description,
		GlobalArgs:  []string{},
		Args:        args,
		Snippets:    snippets,
		Scenarios:   refs,
	}

	lib.Scenarios = append(lib.Scenarios, scenario)

	err = l.makeLibraryPathsRelative(libraryPath, lib)
	if err != nil {
		return nil, err
	}

	return lib, nil
}

func (l *libImpl) makeLibraryPathsRelative(libpath string, lib *library.Library) error {
	for i, libRef := range lib.Libraries {
		rel, err := l.file.ResolveRelativeFrom(libRef.Path, libpath)
		if err != nil {
			return err
		}
		lib.Libraries[i].Path = rel
	}

	for _, scenario := range lib.Scenarios {
		for i, snippet := range scenario.Snippets {
			rel, err := l.file.ResolveRelativeFrom(snippet.Path, libpath)
			if err != nil {
				return err
			}
			scenario.Snippets[i].Path = rel
		}
	}
	return nil
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
	loader := &library.Loader{
		File: file,
		Yaml: yaml,
	}
	opsFileProcessor := opsfile.NewOpsFileProcessor(yaml, file)
	boshInterpolator := bosh.NewBoshInterpolator()
	resolver := &composer.Resolver{
		Loader:          loader,
		SnippetResolver: opsFileProcessor,
		Interpolator:    boshInterpolator,
	}
	opsFileExecutor := &plan.InterpolationExecutor{
		Processor:    opsFileProcessor,
		Interpolator: boshInterpolator,
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

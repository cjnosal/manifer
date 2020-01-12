package library

import (
	"fmt"
	"github.com/cjnosal/manifer/v2/pkg/file"
	"github.com/cjnosal/manifer/v2/pkg/yaml"
	"strings"
)

type LibraryLoader interface {
	Load(paths []string) (*LoadedLibrary, error)
}

type ScenarioNodes []*ScenarioNode

type Loader struct {
	Yaml yaml.YamlAccess
	File file.FileAccess
}

type LoadedLibrary struct {
	TopLibraries []*Library
	Libraries    map[string]*Library
}

type ScenarioNode struct {
	Name               string
	Description        string             `yaml:"description,omitempty"`
	LibraryPath        string             `yaml:"library_path,omitempty"`
	GlobalInterpolator InterpolatorParams `yaml:"global_interpolator,omitempty"`
	RefInterpolator    InterpolatorParams `yaml:"ref_interpolator,omitempty"`
	Interpolator       InterpolatorParams `yaml:"interpolator,omitempty"`
	Snippets           []Snippet          `yaml:"snippets,omitempty"`
	Dependencies       ScenarioNodes      `yaml:"dependencies,omitempty"`
}

func (s *ScenarioNode) GetProcessorTypes() []Type {
	types := []Type{}
	for _, snippet := range s.Snippets {
		types = insert(types, snippet.Processor.Type)
	}
	types = merge(types, s.Dependencies.GetProcessorTypes())
	return types
}

func (n ScenarioNodes) GetProcessorTypes() []Type {
	types := []Type{}
	for _, node := range n {
		types = merge(types, node.GetProcessorTypes())
	}
	return types
}

func (l *LoadedLibrary) GetScenarioTree(name string) (*ScenarioNode, error) {
	return l.getScenarioNode(name, InterpolatorParams{}, nil)
}

func (l *LoadedLibrary) getScenarioNode(name string, refInterpolator InterpolatorParams, parentLib *Library) (*ScenarioNode, error) {
	var scenario *Scenario
	var lib *Library
	if parentLib != nil {
		scenario, lib = l.GetScenarioFromLib(parentLib, name)
	} else {
		scenario, lib = l.GetScenario(name)
	}
	if scenario == nil {
		return nil, fmt.Errorf("Unable to find scenario %s", name)
	}
	deps := []*ScenarioNode{}
	for _, ref := range scenario.Scenarios {
		node, err := l.getScenarioNode(ref.Name, ref.Interpolator, lib)
		if err != nil {
			return nil, fmt.Errorf("%w\n  while finding scenario %s", err, name)
		}
		deps = append(deps, node)
	}
	for i, snippet := range scenario.Snippets {
		if snippet.Processor.Type == "" {
			scenario.Snippets[i].Processor.Type = lib.Type
		}
	}
	scenarioNode := &ScenarioNode{
		Name:               scenario.Name,
		Description:        scenario.Description,
		LibraryPath:        l.GetPath(lib),
		GlobalInterpolator: scenario.GlobalInterpolator,
		RefInterpolator:    refInterpolator,
		Interpolator:       scenario.Interpolator,
		Snippets:           scenario.Snippets,
		Dependencies:       deps,
	}
	return scenarioNode, nil
}

func (l *LoadedLibrary) GetScenario(name string) (*Scenario, *Library) {
	for _, lib := range l.TopLibraries {
		scenario, foundIn := l.GetScenarioFromLib(lib, name)
		if scenario != nil {
			return scenario, foundIn
		}
	}
	return nil, nil
}

func (l *LoadedLibrary) GetScenarioFromLib(lib *Library, name string) (*Scenario, *Library) {
	scenarioPath := SplitName(name)
	if len(scenarioPath) == 1 { // current library
		for _, s := range lib.Scenarios {
			if s.Name == scenarioPath[0] {
				return &s, lib
			}
		}
		return nil, nil
	} else {
		alib := l.GetAliasedLibrary(lib, scenarioPath[0])
		if alib == nil {
			return nil, nil
		}
		return l.GetScenarioFromLib(alib, strings.Join(scenarioPath[1:], "."))
	}
}

func (l *LoadedLibrary) GetAliasedLibrary(lib *Library, alias string) *Library {
	for _, ref := range lib.Libraries {
		if ref.Alias == alias {
			aliasedLib := l.Libraries[ref.Path]
			return aliasedLib
		}
	}
	return nil
}

func (l *LoadedLibrary) GetPath(lib *Library) string {
	for path, loaded := range l.Libraries {
		if lib == loaded {
			return path
		}
	}
	return ""
}

func (l *Loader) Load(paths []string) (*LoadedLibrary, error) {
	loaded := &LoadedLibrary{
		TopLibraries: []*Library{},
		Libraries:    map[string]*Library{},
	}
	wd, err := l.File.GetWorkingDirectory()
	if err != nil {
		return nil, fmt.Errorf("%w\n  while finding working directory", err)
	}
	for _, p := range paths {
		absPath, err := l.File.ResolveRelativeTo(p, wd)
		if err != nil {
			return nil, fmt.Errorf("%w\n  while resolving library path %s from %s", err, p, wd)
		}
		err = l.loadLib(absPath, loaded, true)
		if err != nil {
			return nil, fmt.Errorf("%w\n  while loading library from path %s", err, p)
		}
	}

	return loaded, nil
}

func (l *Loader) loadLib(path string, loaded *LoadedLibrary, top bool) error {
	bytes, err := l.File.Read(path)
	if err != nil {
		return fmt.Errorf("%w\n  while reading library at %s", err, path)
	}
	lib := &Library{}
	err = l.Yaml.Unmarshal(bytes, lib)
	if err != nil {
		return fmt.Errorf("%w\n  while parsing library at %s", err, path)
	}

	for i, scenario := range lib.Scenarios {
		for j, snippet := range scenario.Snippets {
			if snippet.Path != "" {
				absSnippetPath, err := l.File.ResolveRelativeTo(snippet.Path, path)
				if err != nil {
					return fmt.Errorf("%w\n  while resolving snippet path %s from %s", err, snippet.Path, path)
				}
				lib.Scenarios[i].Snippets[j].Path = absSnippetPath
			}
		}
	}

	for i, libref := range lib.Libraries {
		absLibPath, err := l.File.ResolveRelativeTo(libref.Path, path)
		if err != nil {
			return fmt.Errorf("%w\n  while resolving library path %s from %s", err, libref.Path, path)
		}
		lib.Libraries[i].Path = absLibPath
		err = l.loadLib(absLibPath, loaded, false)
		if err != nil {
			return err
		}
	}

	if top {
		loaded.TopLibraries = append(loaded.TopLibraries, lib)
	}
	loaded.Libraries[path] = lib

	return nil
}

func SplitName(scenarioName string) []string {
	return strings.Split(scenarioName, ".")
}

func contains(collection []Type, value Type) bool {
	for _, c := range collection {
		if c == value {
			return true
		}
	}
	return false
}

func merge(a []Type, b []Type) []Type {
	for _, t := range b {
		a = insert(a, t)
	}
	return a
}

func insert(a []Type, b Type) []Type {
	if !contains(a, b) {
		a = append(a, b)
	}
	return a
}

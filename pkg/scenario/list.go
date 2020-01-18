package scenario

import (
	"fmt"
	"github.com/cjnosal/manifer/v2/pkg/library"
)

type ScenarioLister interface {
	ListScenarios(libraries []string, all bool) ([]ScenarioEntry, error)
}

type Lister struct {
	Loader library.LibraryLoader
}

type ScenarioEntry struct {
	Name        string `yaml:"name,omitempty"`
	Description string `yaml:"description,omitempty"`
}

func (l *Lister) ListScenarios(libraryPaths []string, all bool) ([]ScenarioEntry, error) {
	entries := []ScenarioEntry{}
	loadedLibrary, err := l.Loader.Load(libraryPaths)
	if err != nil {
		return nil, fmt.Errorf("%s\n  loading libraries", err)
	}

	for _, lib := range loadedLibrary.TopLibraries {
		l.printLib("", lib, &entries, loadedLibrary, all)
	}

	return entries, nil
}

func (l *Lister) printLib(prefix string, lib *library.Library, entries *[]ScenarioEntry, loadedLibrary *library.LoadedLibrary, all bool) {
	for _, s := range lib.Scenarios {
		entry := ScenarioEntry{
			Name:        prefix + s.Name,
			Description: s.Description,
		}
		*entries = append(*entries, entry)
	}

	if all {
		for _, ref := range lib.Libraries {
			prefix := prefix + ref.Alias + "."
			l.printLib(prefix, loadedLibrary.GetAliasedLibrary(lib, ref.Alias), entries, loadedLibrary, all)
		}
	}
}

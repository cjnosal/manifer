package scenario

import (
	"github.com/cjnosal/manifer/pkg/library"
)

type ScenarioLister interface {
	ListScenarios(libraries []string, all bool) ([]ScenarioEntry, error)
}

type Lister struct {
	Loader library.LibraryLoader
}

type ScenarioEntry struct {
	Name        string
	Description string
}

func (l *Lister) ListScenarios(libraryPaths []string, all bool) ([]ScenarioEntry, error) {
	entries := []ScenarioEntry{}
	libs, err := l.Loader.Load(libraryPaths)
	if err != nil {
		return nil, err
	}

	for _, lib := range libs {
		l.printLib("", &lib, &entries, all)
	}

	return entries, nil
}

func (l *Lister) printLib(prefix string, lib *library.LoadedLibrary, entries *[]ScenarioEntry, all bool) {
	for _, s := range lib.Library.Scenarios {
		entry := ScenarioEntry{
			Name:        prefix + s.Name,
			Description: s.Description,
		}
		*entries = append(*entries, entry)
	}

	if all {
		for ref, lib := range lib.References {
			refPath := prefix + ref + "."
			l.printLib(refPath, lib, entries, all)
		}
	}
}

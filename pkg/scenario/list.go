package scenario

import (
	"github.com/cjnosal/manifer/pkg/library"
	"strings"
)

type ScenarioLister interface {
	ListScenarios(libraries []string) ([]byte, error)
}

type Lister struct {
	Loader library.LibraryLoader
}

func (l *Lister) ListScenarios(libraryPaths []string) ([]byte, error) {
	builder := strings.Builder{}
	libs, err := l.Loader.Load(libraryPaths)
	if err != nil {
		return nil, err
	}

	for _, lib := range libs {
		l.printLib("", &lib, &builder)
	}

	return []byte(builder.String()), nil
}

func (l *Lister) printLib(prefix string, lib *library.LoadedLibrary, builder *strings.Builder) {
	for _, s := range lib.Library.Scenarios {
		builder.WriteString(prefix + s.Name)
		builder.WriteString("\n\t")

		var description string
		if s.Description != "" {
			description = s.Description
		} else {
			description = "no description"
		}
		builder.WriteString(description)

		builder.WriteString("\n")
		builder.WriteString("\n")
	}

	for ref, lib := range lib.References {
		refPath := prefix + ref + "."
		l.printLib(refPath, lib, builder)
	}
}

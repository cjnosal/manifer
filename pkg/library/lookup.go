package library

import (
	"fmt"
	"strings"
)

type LibraryLookup interface {
	GetContainingLibrary(scenarioName string, libraries []LoadedLibrary) (*LoadedLibrary, error)
}

type Lookup struct{}

func (lookup *Lookup) GetContainingLibrary(scenarioName string, libraries []LoadedLibrary) (*LoadedLibrary, error) {
	scenarioPath := SplitName(scenarioName)
	for _, l := range libraries {
		lib := lookup.getLibrary(scenarioPath, &l)
		if lib != nil {
			return lib, nil
		}
	}
	return nil, fmt.Errorf("Unable to find scenario '%s'", scenarioName)
}

func (lookup *Lookup) getLibrary(scenarioPath []string, rootLibrary *LoadedLibrary) *LoadedLibrary {
	if len(scenarioPath) == 1 {
		// no library reference - check rootLibrary
		if lookup.containsScenario(scenarioPath[0], rootLibrary) {
			return rootLibrary
		} else {
			return nil
		}
	} else {
		// follow library reference
		ref := scenarioPath[0]
		sublib := rootLibrary.References[ref]
		if sublib == nil {
			return nil
		}
		return lookup.getLibrary(scenarioPath[1:], sublib)
	}
}

func SplitName(scenarioName string) []string {
	return strings.Split(scenarioName, ".")
}

func (lookup *Lookup) containsScenario(name string, lib *LoadedLibrary) bool {
	for _, s := range lib.Library.Scenarios {
		if s.Name == name {
			return true
		}
	}

	return false
}

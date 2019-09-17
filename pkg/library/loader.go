package library

import (
	"fmt"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/yaml"
)

type LibraryLoader interface {
	Load(paths []string) ([]LoadedLibrary, error)
}

type Loader struct {
	Yaml yaml.YamlAccess
	File file.FileAccess
}

type LoadedLibrary struct {
	Path       string
	Library    *Library
	References map[string]*LoadedLibrary
}

func (l *Loader) Load(paths []string) ([]LoadedLibrary, error) {
	loaded := []LoadedLibrary{}
	for _, p := range paths {
		loadedLib, err := l.loadLib(p)
		if err != nil {
			return nil, err
		}
		loaded = append(loaded, *loadedLib)
	}
	return loaded, nil
}

func (l *Loader) loadLib(path string) (*LoadedLibrary, error) {
	bytes, err := l.File.Read(path)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to read library at %s", err, path)
	}
	lib := &Library{}
	err = l.Yaml.Unmarshal(bytes, lib)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to parse library at %s", err, path)
	}

	for i, scenario := range lib.Scenarios {
		for j, snippet := range scenario.Snippets {
			lib.Scenarios[i].Snippets[j].Path = l.File.ResolveRelativeTo(snippet.Path, path)
		}
	}

	loadedLib := LoadedLibrary{
		Path:       path,
		Library:    lib,
		References: map[string]*LoadedLibrary{},
	}

	for _, libref := range lib.Libraries {
		resolvedPath := l.File.ResolveRelativeTo(libref.Path, path)
		sublib, err := l.loadLib(resolvedPath)
		if err != nil {
			return nil, err
		}
		loadedLib.References[libref.Alias] = sublib
	}
	return &loadedLib, nil
}

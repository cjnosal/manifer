package importer

import (
	"fmt"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/interpolator"
	"github.com/cjnosal/manifer/pkg/library"
	"os"
	"path/filepath"
	"strings"
)

type Importer interface {
	Import(libType library.Type, path string, recursive bool, outPath string) (*library.Library, error)
}

type libraryImporter struct {
	fileIO    file.FileAccess
	validator interpolator.Interpolator
}

func NewImporter(fileIO file.FileAccess, validator interpolator.Interpolator) Importer {
	return &libraryImporter{
		fileIO:    fileIO,
		validator: validator,
	}
}

func (l *libraryImporter) Import(libType library.Type, path string, recursive bool, outPath string) (*library.Library, error) {
	lib := &library.Library{
		Type: libType,
	}

	isDir, err := l.fileIO.IsDir(path)
	if err != nil {
		return nil, fmt.Errorf("%w\n  checking import path %s", err, path)
	}
	if isDir {
		scenarios, err := l.importDir(path, recursive, outPath)
		if err != nil {
			return nil, fmt.Errorf("%w\n  importing directory %s", err, path)
		}
		lib.Scenarios = scenarios
	} else {
		scenario, err := l.importFile(path, filepath.Dir(outPath))
		if err != nil {
			return nil, fmt.Errorf("%w\n  importing file %s", err, path)
		}
		lib.Scenarios = []library.Scenario{}
		if scenario != nil {
			lib.Scenarios = append(lib.Scenarios, *scenario)
		}
	}
	return lib, nil
}

func (l *libraryImporter) importDir(dirPath string, recursive bool, outPath string) ([]library.Scenario, error) {
	scenarios := []library.Scenario{}

	err := l.fileIO.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("%w\n  walking to %s", err, path)
		}
		if info.IsDir() {
			if recursive {
				return nil
			} else {
				return filepath.SkipDir
			}
		}

		scenario, err := l.importFile(path, filepath.Dir(outPath))
		if err != nil {
			return fmt.Errorf("%w\n  importing file %s", err, path)
		}
		if scenario != nil {
			scenarios = append(scenarios, *scenario)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w\n  walking directory %s", err, dirPath)
	}
	return scenarios, nil
}

func (l *libraryImporter) importFile(path string, outPath string) (*library.Scenario, error) {
	valid, err := l.validator.ValidateSnippet(path)
	if err != nil {
		return nil, fmt.Errorf("%w\n  validating file %s", err, path)
	}
	if !valid {
		return nil, nil
	}

	base := filepath.Base(path)
	ext := filepath.Ext(path)
	name := base
	if ext != "" {
		name = base[0:strings.LastIndex(base, ext)]
	}
	relPath, err := l.fileIO.ResolveRelativeFrom(path, outPath)
	if err != nil {
		return nil, fmt.Errorf("%w\n  resolving relative path from %s", err, outPath)
	}
	return &library.Scenario{
		Name:        name,
		Description: fmt.Sprintf("imported from %s", base),
		Snippets: []library.Snippet{
			library.Snippet{
				Path: relPath,
			},
		},
	}, nil
}

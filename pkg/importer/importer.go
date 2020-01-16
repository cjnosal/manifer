package importer

import (
	"fmt"
	"github.com/cjnosal/manifer/v2/pkg/file"
	"github.com/cjnosal/manifer/v2/pkg/library"
	"github.com/cjnosal/manifer/v2/pkg/processor"
	"github.com/cjnosal/manifer/v2/pkg/processor/factory"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Importer interface {
	Import(libType library.Type, path string, recursive bool, outPath string) (*library.Library, error)
}

type libraryImporter struct {
	fileIO    file.FileAccess
	validator factory.ProcessorFactory
}

func NewImporter(fileIO file.FileAccess, procFact factory.ProcessorFactory) Importer {
	return &libraryImporter{
		fileIO:    fileIO,
		validator: procFact,
	}
}

type importedSnippet struct {
	names       []string
	description string
	path        string
}

func (l *libraryImporter) Import(libType library.Type, path string, recursive bool, outPath string) (*library.Library, error) {
	imports := []importedSnippet{}
	validator, err := l.validator.Create(libType)
	if err != nil {
		return nil, err
	}
	isDir, err := l.fileIO.IsDir(path)
	if err != nil {
		return nil, fmt.Errorf("%w\n  checking import path %s", err, path)
	}
	if isDir {
		imps, err := l.importDir(validator, libType, path, recursive, outPath)
		if err != nil {
			return nil, fmt.Errorf("%w\n  importing directory %s", err, path)
		}
		imports = imps
	} else {
		imp, err := l.importFile(validator, libType, path, filepath.Dir(outPath))
		if err != nil {
			return nil, fmt.Errorf("%w\n  importing file %s", err, path)
		}
		if imp != nil {
			imports = append(imports, *imp)
		}
	}

	lib := &library.Library{
		Scenarios: []library.Scenario{},
	}

	candidates := map[string][]importedSnippet{}
	// initialize with first name
	for _, imp := range imports {
		name := imp.names[0]
		if candidates[name] == nil {
			candidates[name] = []importedSnippet{}
		}
		candidates[name] = append(candidates[name], imp)
	}

	// resolve name conflicts
	for name, conflicts := range candidates {
		if len(conflicts) == 1 {
			continue
		}
		resolutions := l.resolve(name, conflicts, 1)
		delete(candidates, name)
		for k, v := range resolutions {
			candidates[k] = v
		}
	}

	// sort names alphabetically (map iterator not deterministic)
	keys := []string{}
	for k := range candidates {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		imp := candidates[name][0]
		scenario := library.Scenario{
			Name:        name,
			Description: imp.description,
			Snippets: []library.Snippet{
				library.Snippet{
					Path: imp.path,
					Processor: library.Processor{
						Type: libType,
					},
				},
			},
		}
		lib.Scenarios = append(lib.Scenarios, scenario)
	}

	return lib, nil
}

func (l *libraryImporter) resolve(name string, conflicts []importedSnippet, nameIndex int) map[string][]importedSnippet {
	if len(conflicts) == 1 {
		return map[string][]importedSnippet{
			name: conflicts,
		}
	}
	candidates := map[string][]importedSnippet{}
	// for each conflict iterate on name if possible, insert into candidates, recurse
	for _, conflict := range conflicts {
		n := conflict.names[len(conflict.names)-1]
		if nameIndex < len(conflict.names) {
			n = conflict.names[nameIndex]
		}
		if candidates[n] == nil {
			candidates[n] = []importedSnippet{}
		}
		candidates[n] = append(candidates[n], conflict)
	}
	for n, c := range candidates {
		resolutions := l.resolve(n, c, nameIndex+1)
		delete(candidates, n)
		for k, v := range resolutions {
			candidates[k] = v
		}
	}
	return candidates
}

func (l *libraryImporter) importDir(validator processor.Processor, libType library.Type, dirPath string, recursive bool, outPath string) ([]importedSnippet, error) {
	imports := []importedSnippet{}

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

		imp, err := l.importFile(validator, libType, path, filepath.Dir(outPath))
		if err != nil {
			return fmt.Errorf("%w\n  importing file %s", err, path)
		}
		if imp != nil {
			imports = append(imports, *imp)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w\n  walking directory %s", err, dirPath)
	}
	return imports, nil
}

func (l *libraryImporter) importFile(validator processor.Processor, libType library.Type, path string, outPath string) (*importedSnippet, error) {
	hint, err := validator.ValidateSnippet(path)
	if err != nil {
		return nil, fmt.Errorf("%w\n  validating file %s", err, path)
	}
	if !hint.Valid {
		return nil, nil
	}

	relPath, err := l.fileIO.ResolveRelativeFrom(path, outPath)
	if err != nil {
		return nil, fmt.Errorf("%w\n  resolving relative path from %s", err, outPath)
	}
	return &importedSnippet{
		names:       l.namesFromPath(relPath),
		path:        relPath,
		description: fmt.Sprintf("%s %s (imported from %s)", hint.Action, hint.Element, relPath),
	}, nil
}

func (l *libraryImporter) namesFromPath(path string) []string {
	cleanPath := filepath.Clean(path)

	base := filepath.Base(cleanPath)
	ext := filepath.Ext(cleanPath)
	name := base
	if ext != "" {
		name = base[0:strings.LastIndex(base, ext)]
	}

	names := []string{name}

	dir := filepath.Dir(cleanPath)
	for {
		if dir == "." {
			break
		}
		dirName := filepath.Base(dir)
		name = fmt.Sprintf("%s_%s", dirName, name)
		names = append(names, name)
		dir = filepath.Dir(dir)
	}

	return names
}

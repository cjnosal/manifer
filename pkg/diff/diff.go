package diff

import (
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type Diff interface {
	FindDiff(path1 string, path2 string) (string, error)
}

type diffMatchPatch interface {
	DiffMain(text1 string, text2 string, checkLines bool) []diffmatchpatch.Diff
	DiffPrettyText([]diffmatchpatch.Diff) string
}

type FileDiff struct {
	File  file.FileAccess
	Patch diffMatchPatch
}

func (f *FileDiff) FindDiff(path1 string, path2 string) (string, error) {
	str1, err := f.File.Read(path1)
	if err != nil {
		return "", err
	}
	str2, err := f.File.Read(path2)
	if err != nil {
		return "", err
	}

	diffs := f.Patch.DiffMain(string(str1), string(str2), true)

	return f.Patch.DiffPrettyText(diffs), nil
}

package diff

import (
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type Diff interface {
	FindDiff(path1 string, path2 string) (string, error)
	StringDiff(str1 string, str2 string) string
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
	b1, err := f.File.Read(path1)
	if err != nil {
		return "", err
	}
	b2, err := f.File.Read(path2)
	if err != nil {
		return "", err
	}

	diff := f.StringDiff(string(b1), string(b2))

	return diff, nil
}

func (f *FileDiff) StringDiff(str1 string, str2 string) string {
	if str1 == str2 {
		return ""
	}
	diffs := f.Patch.DiffMain(str1, str2, true)
	return f.Patch.DiffPrettyText(diffs)
}

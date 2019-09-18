package file

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type FileAccess interface {
	Read(path string) ([]byte, error)
	ReadAndTag(path string) (*TaggedBytes, error)
	Write(path string, content []byte, perms os.FileMode) error
	TempDir(dir string, prefix string) (string, error)
	RemoveAll(dir string) error
	ResolveRelativeTo(targetFile string, sourceFile string) string
}

type FileIO struct{}

type TaggedBytes struct {
	Bytes []byte
	Tag   string
}

func (f *FileIO) ReadAndTag(path string) (*TaggedBytes, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &TaggedBytes{
		Bytes: bytes,
		Tag:   path,
	}, nil
}

func (f *FileIO) Read(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (f *FileIO) Write(path string, content []byte, perms os.FileMode) error {
	return ioutil.WriteFile(path, content, perms)
}

func (f *FileIO) TempDir(dir string, prefix string) (string, error) {
	return ioutil.TempDir(dir, prefix)
}

func (f *FileIO) RemoveAll(dir string) error {
	return os.RemoveAll(dir)
}

func (f *FileIO) ResolveRelativeTo(targetFile string, sourceFile string) string {
	if filepath.IsAbs(targetFile) {
		return targetFile
	} else {
		return filepath.Join(filepath.Dir(sourceFile), filepath.Clean(targetFile))
	}
}

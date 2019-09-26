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
	ResolveRelativeTo(targetFile string, sourceFile string) (string, error)
	ResolveRelativeFrom(targetFile string, sourceFile string) (string, error)
	ResolveRelativeFromWD(targetFile string) (string, error)
	GetWorkingDirectory() (string, error)
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

func (f *FileIO) ResolveRelativeTo(targetFile string, sourceFile string) (string, error) {
	if filepath.IsAbs(targetFile) {
		return targetFile, nil
	} else {
		dir := sourceFile
		dirInfo, err := os.Stat(dir)
		if err != nil {
			return "", err
		}
		if !dirInfo.IsDir() {
			dir = filepath.Dir(sourceFile)
		}
		return filepath.Clean(filepath.Join(dir, targetFile)), nil
	}
}

func (f *FileIO) ResolveRelativeFrom(targetFile string, sourceFile string) (string, error) {
	dir := sourceFile
	dirInfo, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !dirInfo.IsDir() {
		dir = filepath.Dir(sourceFile)
	}
	return filepath.Rel(dir, targetFile)
}

func (f *FileIO) ResolveRelativeFromWD(targetFile string) (string, error) {
	dir, err := f.GetWorkingDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Rel(dir, targetFile)
}

func (f *FileIO) GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

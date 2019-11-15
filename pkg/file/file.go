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
	IsDir(path string) (bool, error)
	Walk(path string, callback func(path string, info os.FileInfo, err error) error) error
	MkDir(path string) error
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
		isDir, err := f.IsDir(dir)
		if err != nil {
			return "", err
		}
		if !isDir {
			dir = filepath.Dir(sourceFile)
		}
		return filepath.Clean(filepath.Join(dir, targetFile)), nil
	}
}

func (f *FileIO) ResolveRelativeFrom(targetFile string, sourceFile string) (string, error) {
	dir, err := filepath.Abs(sourceFile)
	if err != nil {
		return "", err
	}
	isDir, err := f.IsDir(dir)
	if err != nil {
		return "", err
	}
	if !isDir {
		dir, err = filepath.Abs(filepath.Dir(sourceFile))
		if err != nil {
			return "", err
		}
	}
	target, err := filepath.Abs(targetFile)
	if err != nil {
		return "", err
	}
	return filepath.Rel(dir, target)
}

func (f *FileIO) ResolveRelativeFromWD(targetFile string) (string, error) {
	if !filepath.IsAbs(targetFile) {
		return targetFile, nil
	}
	dir, err := f.GetWorkingDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Rel(dir, targetFile)
}

func (f *FileIO) GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

func (f *FileIO) IsDir(path string) (bool, error) {
	pathInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return pathInfo.IsDir(), nil
}

func (f *FileIO) Walk(path string, callback func(path string, info os.FileInfo, err error) error) error {
	return filepath.Walk(path, callback)
}

func (f *FileIO) MkDir(path string) error {
	return os.MkdirAll(path, 0755)
}

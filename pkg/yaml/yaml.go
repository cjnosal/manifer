package yaml

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"

	"github.com/cjnosal/manifer/pkg/file"
)

type YamlAccess interface {
	Load(path string, i interface{}) error
	Write(w io.Writer, i interface{}) error
}

func (l *Yaml) Load(path string, i interface{}) (err error) {
	defer func() {
		// yaml.UnmarshalStrict may panic instead of returning error
		r := recover()
		if r != nil {
			switch t := r.(type) {
			case string:
				err = errors.New(t)
			case error:
				err = t
			}
			err = fmt.Errorf("%w\n  while unmarshalling yaml from %s", err, path)
		}
	}()

	bytes, err := l.File.Read(path)
	if err != nil {
		return fmt.Errorf("%w\n  while loading yaml from %s", err, path)
	}

	err = yaml.UnmarshalStrict(bytes, i)
	if err != nil {
		return fmt.Errorf("%w\n  while unmarshalling yaml from %s", err, path)
	}

	return nil
}

func (l *Yaml) Write(w io.Writer, i interface{}) (err error) {
	defer func() {
		// yaml.Marshall may panic instead of returning error
		r := recover()
		if r != nil {
			switch t := r.(type) {
			case string:
				err = errors.New(t)
			case error:
				err = t
			}
			err = fmt.Errorf("%w\n  while marshalling yaml: %v", err, i)
		}
	}()

	bytes, err := yaml.Marshal(i)
	if err != nil {
		return fmt.Errorf("%w\n  while marshalling yaml: %v", err, i)
	}

	_, err = w.Write(bytes)
	if err != nil {
		return fmt.Errorf("%w\n  while writing yaml", err)
	}

	return nil
}

type Yaml struct {
	File file.FileAccess
}

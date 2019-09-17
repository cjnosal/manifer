package yaml

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
)

type YamlAccess interface {
	Unmarshal(bytes []byte, i interface{}) error
	Marshal(i interface{}) ([]byte, error)
}

func (l *Yaml) Unmarshal(bytes []byte, i interface{}) (err error) {
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
			err = fmt.Errorf("%w\n  while unmarshalling yaml", err)
		}
	}()

	err = yaml.Unmarshal(bytes, i)
	if err != nil {
		return fmt.Errorf("%w\n  while unmarshalling yaml", err)
	}

	return nil
}

func (l *Yaml) Marshal(i interface{}) (b []byte, err error) {
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
		return nil, fmt.Errorf("%w\n  while marshalling yaml: %v", err, i)
	}

	return bytes, nil
}

type Yaml struct {
}

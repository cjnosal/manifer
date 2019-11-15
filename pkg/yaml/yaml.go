package yaml

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
)

type YamlAccess interface {
	Unmarshal(bytes []byte, i interface{}) error
	Marshal(i interface{}) ([]byte, error)
	Walk(n *yaml.Node, visitor NodeVisitor) error
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

type NodeVisitor func(node *yaml.Node, ancestors []Ancestor) error

func (l *Yaml) Walk(n *yaml.Node, visitor NodeVisitor) error {
	return l.visit(n, []Ancestor{}, visitor)
}

func (l *Yaml) visit(n *yaml.Node, ancestors []Ancestor, visitor NodeVisitor) error {
	var err error

	err = visitor(n, ancestors)
	if err != nil {
		return err
	}
	if n.Kind == yaml.AliasNode || n.Kind == yaml.ScalarNode {
		return nil
	}
	for i, c := range n.Content {
		if n.Kind == yaml.SequenceNode {
			err = l.visit(c, prepend(Ancestor{Node: n, Token: fmt.Sprintf("%d", i)}, ancestors), visitor)
		} else if n.Kind == yaml.MappingNode {
			if i%2 == 0 {
				// visit map values with the map itself as ancestor providing key as token
				err = l.visit(n.Content[i+1], prepend(Ancestor{Node: n, Token: c.Value}, ancestors), visitor)
			}
		} else if n.Kind == yaml.DocumentNode {
			err = l.visit(c, prepend(Ancestor{Node: n}, ancestors), visitor)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func prepend(n Ancestor, nodes []Ancestor) []Ancestor {
	return append([]Ancestor{n}, nodes...)
}

type Yaml struct {
}

type Ancestor struct {
	Node  *yaml.Node
	Token string
}

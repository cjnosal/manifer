package yaml

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"strconv"
)

type SchemaNode struct {
	Kind     yaml.Kind
	Contents map[string]*SchemaNode
}

func (n *SchemaNode) Insert(tokens []string, kind yaml.Kind) error {
	key := tokens[0]
	if len(tokens) == 1 {
		if n.Contents[key] == nil {
			n.Contents[key] = &SchemaNode{
				Kind:     kind,
				Contents: map[string]*SchemaNode{},
			}
		} else {
			if kind != n.Contents[key].Kind {
				return fmt.Errorf("Different types (%d vs %d)", kind, n.Contents[key].Kind)
			}
		}
	} else {
		child := n.Contents[key]
		if child == nil {
			return fmt.Errorf("Missing node %s", key)
		}
		err := child.Insert(tokens[1:], kind)
		if err != nil {
			return fmt.Errorf("%w\n  inserting %v into schema at %v", err, kind, tokens)
		}
	}
	return nil
}

func tokens(ancestors []Ancestor) []string {
	tokens := []string{}
	for i := len(ancestors) - 2; i >= 0; i-- { // no token for root mapping/sequence
		_, err := strconv.Atoi(ancestors[i].Token)
		if err == nil {
			tokens = append(tokens, "!!element")
		} else {
			tokens = append(tokens, ancestors[i].Token)
		}
	}
	return tokens
}

type SchemaBuilder struct {
	Root *SchemaNode
}

func (s *SchemaBuilder) OnVisit(node *yaml.Node, ancestors []Ancestor) error {
	if node.Kind == yaml.DocumentNode {
		return nil
	} else if s.Root == nil {
		s.Root = &SchemaNode{
			Kind:     node.Kind,
			Contents: map[string]*SchemaNode{},
		}
	} else {
		t := tokens(ancestors)
		err := s.Root.Insert(t, node.Kind)
		if err != nil {
			return fmt.Errorf("%w\n  inserting %v", err, t)
		}
	}
	return nil
}

package yaml

import (
	"github.com/google/go-cmp/cmp"
	y "gopkg.in/yaml.v3"
	"testing"
)

func TestWalkSchemaBuilder(t *testing.T) {

	t.Run("map with scalar sequence", func(t *testing.T) {
		input := `---
foo:
- bizz
- bazz`
		node := &y.Node{}
		yaml := &Yaml{}
		err := yaml.Unmarshal([]byte(input), node)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		schemaBuilder := &SchemaBuilder{}

		err = yaml.Walk(node, schemaBuilder.OnVisit)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expected := &SchemaNode{
			Kind: y.MappingNode,
			Contents: map[string]*SchemaNode{
				"foo": &SchemaNode{
					Kind: y.SequenceNode,
					Contents: map[string]*SchemaNode{
						"!!element": &SchemaNode{
							Kind:     y.ScalarNode,
							Contents: map[string]*SchemaNode{},
						},
					},
				},
			},
		}

		if !cmp.Equal(*expected, *schemaBuilder.Root) {
			t.Errorf("Expected:\n'%v'\nActual:\n'%v'\nDiff:\n%s\n", *expected, *schemaBuilder.Root, cmp.Diff(*expected, *schemaBuilder.Root))
		}
	})

	t.Run("sequence with scalar mappings", func(t *testing.T) {
		input := `---
- foo: 1
- bar: 2`
		node := &y.Node{}
		yaml := &Yaml{}
		err := yaml.Unmarshal([]byte(input), node)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		schemaBuilder := &SchemaBuilder{}

		err = yaml.Walk(node, schemaBuilder.OnVisit)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expected := &SchemaNode{
			Kind: y.SequenceNode,
			Contents: map[string]*SchemaNode{
				"!!element": &SchemaNode{
					Kind: y.MappingNode,
					Contents: map[string]*SchemaNode{
						"foo": &SchemaNode{
							Kind:     y.ScalarNode,
							Contents: map[string]*SchemaNode{},
						},
						"bar": &SchemaNode{
							Kind:     y.ScalarNode,
							Contents: map[string]*SchemaNode{},
						},
					},
				},
			},
		}

		if !cmp.Equal(*expected, *schemaBuilder.Root) {
			t.Errorf("Expected:\n'%v'\nActual:\n'%v'\nDiff:\n%s\n", *expected, *schemaBuilder.Root, cmp.Diff(*expected, *schemaBuilder.Root))
		}
	})
}

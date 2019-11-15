package yaml

import (
	"errors"
	"github.com/cjnosal/manifer/test"
	"github.com/google/go-cmp/cmp"
	y "gopkg.in/yaml.v3"
	"testing"
)

func TestSchemaBuilder(t *testing.T) {

	t.Run("map with scalar sequence", func(t *testing.T) {
		/**
		---
		foo:
		- bizz
		- bazz
		*/
		subject := &SchemaBuilder{}

		doc := &y.Node{Kind: y.DocumentNode}
		err := subject.OnVisit(doc, []Ancestor{})
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		docMap := &y.Node{Kind: y.MappingNode}
		err = subject.OnVisit(docMap, []Ancestor{{Node: doc}})
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		bar := &y.Node{Kind: y.SequenceNode}
		err = subject.OnVisit(bar, []Ancestor{{Node: docMap, Token: "foo"}, {Node: doc}})
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		bizz := &y.Node{Kind: y.ScalarNode, Value: "bizz"}
		err = subject.OnVisit(bizz, []Ancestor{{Node: bar, Token: "0"}, {Node: docMap, Token: "foo"}, {Node: doc}})
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		bazz := &y.Node{Kind: y.ScalarNode, Value: "bazz"}
		err = subject.OnVisit(bazz, []Ancestor{{Node: bar, Token: "1"}, {Node: docMap, Token: "foo"}, {Node: doc}})
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

		if !cmp.Equal(*expected, *subject.Root) {
			t.Errorf("Expected:\n'%v'\nActual:\n'%v'\nDiff:\n%s\n", *expected, *subject.Root, cmp.Diff(*expected, *subject.Root))
		}
	})

	t.Run("sequence with scalar mappings", func(t *testing.T) {
		/**
		---
		- foo: 1
		- bar: 2
		*/
		subject := &SchemaBuilder{}

		doc := &y.Node{Kind: y.DocumentNode}
		err := subject.OnVisit(doc, []Ancestor{})
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		docSequence := &y.Node{Kind: y.SequenceNode}
		err = subject.OnVisit(docSequence, []Ancestor{{Node: doc}})
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		map0 := &y.Node{Kind: y.MappingNode}
		err = subject.OnVisit(map0, []Ancestor{{Node: docSequence, Token: "0"}, {Node: doc}})
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		fooValue := &y.Node{Kind: y.ScalarNode}
		err = subject.OnVisit(fooValue, []Ancestor{{Node: map0, Token: "foo"}, {Node: docSequence, Token: "0"}, {Node: doc}})
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		map1 := &y.Node{Kind: y.MappingNode}
		err = subject.OnVisit(map1, []Ancestor{{Node: docSequence, Token: "1"}, {Node: doc}})
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		barValue := &y.Node{Kind: y.ScalarNode}
		err = subject.OnVisit(barValue, []Ancestor{{Node: map1, Token: "bar"}, {Node: docSequence, Token: "1"}, {Node: doc}})
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

		if !cmp.Equal(*expected, *subject.Root) {
			t.Errorf("Expected:\n'%v'\nActual:\n'%v'\nDiff:\n%s\n", *expected, *subject.Root, cmp.Diff(*expected, *subject.Root))
		}
	})

	t.Run("Type mismatch", func(t *testing.T) {
		/**
		---
		- foo: []
		- foo: {}
		*/
		subject := &SchemaBuilder{}

		doc := &y.Node{Kind: y.DocumentNode}
		subject.OnVisit(doc, []Ancestor{})

		docSequence := &y.Node{Kind: y.SequenceNode}
		subject.OnVisit(docSequence, []Ancestor{{Node: doc}})

		map0 := &y.Node{Kind: y.MappingNode}
		subject.OnVisit(map0, []Ancestor{{Node: docSequence, Token: "0"}, {Node: doc}})

		fooValue := &y.Node{Kind: y.SequenceNode}
		subject.OnVisit(fooValue, []Ancestor{{Node: map0, Token: "foo"}, {Node: docSequence, Token: "0"}, {Node: doc}})

		map1 := &y.Node{Kind: y.MappingNode}
		subject.OnVisit(map1, []Ancestor{{Node: docSequence, Token: "1"}, {Node: doc}})

		barValue := &y.Node{Kind: y.MappingNode}
		err := subject.OnVisit(barValue, []Ancestor{{Node: map1, Token: "foo"}, {Node: docSequence, Token: "1"}, {Node: doc}})

		expectedError := errors.New("Different types (4 vs 2) inserting [!!element foo]")
		if !cmp.Equal(&expectedError, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedError, err)
		}
	})

	t.Run("Child visited before parent", func(t *testing.T) {
		/**
		---
		- foo: []
		*/
		subject := &SchemaBuilder{}

		doc := &y.Node{Kind: y.DocumentNode}
		subject.OnVisit(doc, []Ancestor{})

		docSequence := &y.Node{Kind: y.SequenceNode}
		subject.OnVisit(docSequence, []Ancestor{{Node: doc}})

		map0 := &y.Node{Kind: y.MappingNode}
		fooValue := &y.Node{Kind: y.SequenceNode}
		err := subject.OnVisit(fooValue, []Ancestor{{Node: map0, Token: "foo"}, {Node: docSequence, Token: "0"}, {Node: doc}})

		expectedError := errors.New("Missing node !!element inserting [!!element foo]")
		if !cmp.Equal(&expectedError, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedError, err)
		}
	})
}

package yaml

import (
	"errors"
	"fmt"
	"github.com/cjnosal/manifer/v2/test"
	"github.com/google/go-cmp/cmp"
	y "gopkg.in/yaml.v3"
	"strings"
	"testing"
)

var (
	subject *Yaml
)

func setup(t *testing.T) {
	subject = &Yaml{}
}

func TestUnmarshal(t *testing.T) {

	t.Run("Invalid Yaml", func(t *testing.T) {
		setup(t)

		data := map[string]string{}
		err := subject.Unmarshal([]byte(":::not yaml"), &data)

		if err == nil {
			t.Error("Unmarshal should return error if yaml is not valid")
		}

		expected := "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `:::not ...` into map[string]string\n  while unmarshalling yaml"
		actual := err.Error()
		if actual != expected {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expected, actual)
		}
	})

	t.Run("Valid Yaml", func(t *testing.T) {
		setup(t)

		actual := map[string]string{}
		err := subject.Unmarshal([]byte("key: value\nname: test\n"), &actual)

		if err != nil {
			t.Error("Unmarshal should not return error if yaml is valid")
		}

		expected := map[string]string{
			"key":  "value",
			"name": "test",
		}

		if !cmp.Equal(actual, expected) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, actual)
		}
	})

}

func TestMarshal(t *testing.T) {

	t.Run("Marshal Error", func(t *testing.T) {
		setup(t)

		_, err := subject.Marshal(&Unmarshallable{A: "foo", B: "bar"})

		if err == nil {
			t.Error("Marshal should return error if marshal failed")
		}

		expected := "duplicated key 'a' in struct yaml.Unmarshallable\n  while marshalling yaml: &{foo bar}"
		actual := err.Error()
		if actual != expected {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expected, actual)
		}
	})

	t.Run("Valid Yaml", func(t *testing.T) {
		setup(t)

		data := map[string]string{
			"key":  "value",
			"name": "test",
		}
		bytes, err := subject.Marshal(&data)

		if err != nil {
			t.Error("Marshal should not return error if marshal succeeded")
		}

		actual := string(bytes)
		expected := "key: value\nname: test\n"
		if actual != expected {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expected, actual)
		}
	})
}

func TestWalk(t *testing.T) {
	t.Run("Map document", func(t *testing.T) {
		setup(t)

		input := `key: value
name: test`

		node := &y.Node{}
		err := subject.Unmarshal([]byte(input), node)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		builder := strings.Builder{}
		callback := func(node *y.Node, ancestors []Ancestor) error {
			builder.WriteString(fmt.Sprintf("(%d %s)", node.Kind, node.Value))
			builder.WriteString("[")
			for _, a := range ancestors {
				builder.WriteString(fmt.Sprintf("(%d %s)", a.Node.Kind, a.Token))
			}
			builder.WriteString("]")
			return nil
		}

		expected := `(1 )[](4 )[(1 )](8 value)[(4 key)(1 )](8 test)[(4 name)(1 )]`
		err = subject.Walk(node, callback)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		actual := builder.String()
		if !cmp.Equal(actual, expected) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, actual)
		}

	})

	t.Run("Sequence document", func(t *testing.T) {
		setup(t)

		input := `- key
- name`

		node := &y.Node{}
		err := subject.Unmarshal([]byte(input), node)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		builder := strings.Builder{}
		callback := func(node *y.Node, ancestors []Ancestor) error {
			builder.WriteString(fmt.Sprintf("(%d %s)", node.Kind, node.Value))
			builder.WriteString("[")
			for _, a := range ancestors {
				builder.WriteString(fmt.Sprintf("(%d %s)", a.Node.Kind, a.Token))
			}
			builder.WriteString("]")
			return nil
		}

		expected := `(1 )[](2 )[(1 )](8 key)[(2 0)(1 )](8 name)[(2 1)(1 )]`
		err = subject.Walk(node, callback)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		actual := builder.String()
		if !cmp.Equal(actual, expected) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, actual)
		}

	})

	t.Run("Nested document", func(t *testing.T) {
		setup(t)

		input := `---
- key:
    value: foo
- name:
    test:
    - bar
`

		node := &y.Node{}
		err := subject.Unmarshal([]byte(input), node)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		builder := strings.Builder{}
		callback := func(node *y.Node, ancestors []Ancestor) error {
			builder.WriteString(fmt.Sprintf("(%d %s)", node.Kind, node.Value))
			builder.WriteString("[")
			for _, a := range ancestors {
				builder.WriteString(fmt.Sprintf("(%d %s)", a.Node.Kind, a.Token))
			}
			builder.WriteString("]")
			return nil
		}

		expected := `(1 )[](2 )[(1 )](4 )[(2 0)(1 )](4 )[(4 key)(2 0)(1 )](8 foo)[(4 value)(4 key)(2 0)(1 )](4 )[(2 1)(1 )](4 )[(4 name)(2 1)(1 )](2 )[(4 test)(4 name)(2 1)(1 )](8 bar)[(2 0)(4 test)(4 name)(2 1)(1 )]`
		err = subject.Walk(node, callback)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		actual := builder.String()
		if !cmp.Equal(actual, expected) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, actual)
		}

	})

	t.Run("callback error", func(t *testing.T) {
		setup(t)

		input := `---
- key: value`

		node := &y.Node{}
		err := subject.Unmarshal([]byte(input), node)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expectedError := errors.New("oops")
		callback := func(node *y.Node, ancestors []Ancestor) error {
			return expectedError
		}

		err = subject.Walk(node, callback)
		if !cmp.Equal(&expectedError, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedError, err)
		}

	})
}

type Unmarshallable struct {
	A string `yaml:"a"`
	B string `yaml:"a"`
}

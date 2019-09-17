package yaml

import (
	"reflect"
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

		if !reflect.DeepEqual(actual, expected) {
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

type Unmarshallable struct {
	A string `yaml:"a"`
	B string `yaml:"a"`
}

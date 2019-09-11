package yaml

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/test"
)

var (
	ctrl     *gomock.Controller
	mockFile *file.MockFileAccess
	subject  *Yaml
)

func setup(t *testing.T) {
	ctrl = gomock.NewController(t)
	mockFile = file.NewMockFileAccess(ctrl)
	subject = &Yaml{
		File: mockFile,
	}
}

func TestLoad(t *testing.T) {
	t.Run("Invalid Path", func(t *testing.T) {
		setup(t)
		defer ctrl.Finish()

		mockFile.EXPECT().Read("doesnotexist").Times(1).Return(nil, errors.New("test"))

		data := map[string]string{}
		err := subject.Load("doesnotexist", &data)

		if err == nil {
			t.Error("Load should return error if unable to read file")
		}

		expected := "test\n  while loading yaml from doesnotexist"
		actual := err.Error()
		if actual != expected {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expected, actual)
		}
	})

	t.Run("Invalid Yaml", func(t *testing.T) {
		setup(t)
		defer ctrl.Finish()

		mockFile.EXPECT().Read("malformed").Times(1).Return([]byte(":::not yaml"), nil)

		data := map[string]string{}
		err := subject.Load("malformed", &data)

		if err == nil {
			t.Error("Load should return error if yaml is not valid")
		}

		expected := "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `:::not ...` into map[string]string\n  while unmarshalling yaml from malformed"
		actual := err.Error()
		if actual != expected {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expected, actual)
		}
	})

	t.Run("Valid Yaml", func(t *testing.T) {
		setup(t)
		defer ctrl.Finish()

		mockFile.EXPECT().Read("valid").Times(1).Return([]byte("key: value\nname: test\n"), nil)

		actual := map[string]string{}
		err := subject.Load("valid", &actual)

		if err != nil {
			t.Error("Load should not return error if yaml is valid")
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

func TestWrite(t *testing.T) {
	t.Run("Writer Error", func(t *testing.T) {
		setup(t)
		defer ctrl.Finish()

		writer := test.BrokenWriter{}
		data := map[string]string{}
		err := subject.Write(&writer, &data)

		if err == nil {
			t.Error("Write should return error if writer failed")
		}

		expected := "broken writer failed\n  while writing yaml"
		actual := err.Error()
		if actual != expected {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expected, actual)
		}
	})

	t.Run("Marshaller Error", func(t *testing.T) {
		setup(t)
		defer ctrl.Finish()

		writer := test.StringWriter{}
		err := subject.Write(&writer, &Unmarshallable{A: "foo", B: "bar"})

		if err == nil {
			t.Error("Write should return error if writer failed")
		}

		expected := "Duplicated key 'a' in struct yaml.Unmarshallable\n  while marshalling yaml: &{foo bar}"
		actual := err.Error()
		if actual != expected {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expected, actual)
		}
	})

	t.Run("Valid Yaml", func(t *testing.T) {
		setup(t)
		defer ctrl.Finish()

		writer := test.StringWriter{}
		data := map[string]string{
			"key":  "value",
			"name": "test",
		}
		err := subject.Write(&writer, &data)

		if err != nil {
			t.Error("Write should not return error if writer succeeded")
		}

		actual := writer.String()
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

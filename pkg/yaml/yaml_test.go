package yaml

import (
	"errors"
	"io/ioutil"
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

		expected := "Error loading yaml from doesnotexist: test"
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

		expected := "Error unmarshalling yaml from malformed: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `:::not ...` into map[string]string"
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

		expected := "Error writing yaml: broken writer failed"
		actual := err.Error()
		if actual != expected {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expected, actual)
		}
	})

	t.Run("Marshaller Error", func(t *testing.T) {
		setup(t)
		defer ctrl.Finish()

		writer := test.StringWriter{}
		err := subject.Write(&writer, &Unmarshallable{})

		if err == nil {
			t.Error("Write should return error if writer failed")
		}

		expected := "Duplicated key 'a' in struct yaml.Unmarshallable"
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

		bytes, err := ioutil.ReadFile("../../test/data/valid.yml")
		if err != nil {
			t.Error("Unable to load test data")
		}

		actual := writer.String()
		expected := string(bytes)
		if actual != expected {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expected, actual)
		}
	})
}

type Unmarshallable struct {
	A string `yaml:"a"`
	B string `yaml:"a"`
}

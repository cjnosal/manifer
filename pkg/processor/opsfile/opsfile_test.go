package opsfile

import (
	"errors"
	"testing"

	"github.com/cppforlife/go-patch/patch"
	"github.com/golang/mock/gomock"

	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/yaml"
	"github.com/cjnosal/manifer/test"
	"github.com/google/go-cmp/cmp"
)

func litpnt(i interface{}) *interface{} {
	return &i
}

func newOpDefinition(t string, p string, i interface{}) patch.OpDefinition {
	return patch.OpDefinition{
		Type:  t,
		Path:  &p,
		Value: &i,
	}
}

func TestParsePassthroughFlags(t *testing.T) {

	t.Run("op files", func(t *testing.T) {
		subject := opFileProcessor{}
		flags := []string{"-ofoo", "-o", "bar", "--ops-file=bizz"}
		node, err := subject.ParsePassthroughFlags(flags)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expectedNode := &library.ScenarioNode{
			GlobalArgs:  []string{},
			RefArgs:     []string{},
			Name:        "passthrough",
			Description: "args passed after --",
			LibraryPath: "<cli>",
			Type:        string(library.OpsFile),
			Snippets: []library.Snippet{
				{
					Path: "foo",
					Args: []string{},
				},
				{
					Path: "bar",
					Args: []string{},
				},
				{
					Path: "bizz",
					Args: []string{},
				},
			},
		}
		if !cmp.Equal(*expectedNode, *node) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", *expectedNode, *node)
		}
	})

	t.Run("set other flags as globals", func(t *testing.T) {
		subject := opFileProcessor{}
		flags := []string{"-ofoo", "-vbar"}
		node, err := subject.ParsePassthroughFlags(flags)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expectedNode := &library.ScenarioNode{
			GlobalArgs:  []string{"-vbar"},
			RefArgs:     []string{},
			Name:        "passthrough",
			Description: "args passed after --",
			LibraryPath: "<cli>",
			Type:        string(library.OpsFile),
			Snippets: []library.Snippet{
				{
					Path: "foo",
					Args: []string{},
				},
			},
		}
		if !cmp.Equal(*expectedNode, *node) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", *expectedNode, *node)
		}
	})

	t.Run("parse error", func(t *testing.T) {
		subject := opFileProcessor{}
		flags := []string{"-o"}
		_, err := subject.ParsePassthroughFlags(flags)

		expectedError := "expected argument for flag `-o, --ops-file'\n  while trying to parse opsfile args"
		if err == nil || err.Error() != expectedError {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expectedError, err)
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("valid ops file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockYaml := yaml.NewMockYamlAccess(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)

		subject := NewOpsFileProcessor(mockYaml, mockFile)

		mockFile.EXPECT().Read("/foo").Times(1).Return([]byte{1}, nil)
		mockYaml.EXPECT().Unmarshal([]byte{1}, &[]patch.OpDefinition{}).Times(1).Return(nil)

		valid, err := subject.ValidateSnippet("/foo")

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		if !valid {
			t.Errorf("Expected ValidateSnippet to return true")
		}
	})

	t.Run("invalid ops file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockYaml := yaml.NewMockYamlAccess(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)

		subject := NewOpsFileProcessor(mockYaml, mockFile)

		mockFile.EXPECT().Read("/foo").Times(1).Return([]byte{1}, nil)
		mockYaml.EXPECT().Unmarshal([]byte{1}, &[]patch.OpDefinition{}).Times(1).Return(errors.New("oops"))

		valid, err := subject.ValidateSnippet("/foo")

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		if valid {
			t.Errorf("Expected ValidateSnippet to return false")
		}
	})

	t.Run("file error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockYaml := yaml.NewMockYamlAccess(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)

		subject := NewOpsFileProcessor(mockYaml, mockFile)

		mockFile.EXPECT().Read("/foo").Times(1).Return(nil, errors.New("oops"))

		valid, err := subject.ValidateSnippet("/foo")

		expectedError := errors.New("oops\n  while validating opsfile /foo")
		if !cmp.Equal(&expectedError, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected error:\n'''%s'''\nActual:\n'''%s'''\n", expectedError, err)
		}

		if valid {
			t.Errorf("Expected ValidateSnippet to return false")
		}
	})
}

func TestProcessTemplate(t *testing.T) {

	validTemplate := "foo: bar\n\n"
	invalidTemplate := ":::not yaml"
	cases := []struct {
		name    string
		in      *file.TaggedBytes
		snippet *file.TaggedBytes

		opDefinitions []patch.OpDefinition

		parseSnippetError error

		expectedError error
		expectedOut   []byte
	}{
		{
			name:          "empty op",
			in:            &file.TaggedBytes{Tag: "../../../test/data/template.yml", Bytes: []byte(validTemplate)},
			snippet:       &file.TaggedBytes{Tag: "opsfile.yml", Bytes: []byte(validTemplate)},
			opDefinitions: []patch.OpDefinition{},
			expectedOut:   []byte("foo: bar\n\n"),
		},
		{
			name:    "single op",
			in:      &file.TaggedBytes{Tag: "../../../test/data/template.yml", Bytes: []byte(validTemplate)},
			snippet: &file.TaggedBytes{Tag: "opsfile.yml", Bytes: []byte(validTemplate)},
			opDefinitions: []patch.OpDefinition{
				newOpDefinition("replace", "/bizz?", "bazz"),
			},
			expectedOut: []byte("bizz: bazz\nfoo: bar\n"),
		},
		{
			name:    "multiple ops in file",
			in:      &file.TaggedBytes{Tag: "../../../test/data/template.yml", Bytes: []byte(validTemplate)},
			snippet: &file.TaggedBytes{Tag: "opsfile.yml", Bytes: []byte(validTemplate)},
			opDefinitions: []patch.OpDefinition{
				newOpDefinition("replace", "/bizz?", "bazz"),
				newOpDefinition("replace", "/bazz?", "buzz"),
			},
			expectedOut: []byte("bazz: buzz\nbizz: bazz\nfoo: bar\n"),
		},
		{
			name:              "parse snippet error",
			in:                &file.TaggedBytes{Tag: "../../../test/data/template.yml", Bytes: []byte(validTemplate)},
			parseSnippetError: errors.New("test"),
			snippet:           &file.TaggedBytes{Tag: "/originalsnippet", Bytes: []byte(invalidTemplate)},
			expectedError:     errors.New("test\n  while trying to parse ops file /originalsnippet"),
		},
		{
			name:    "invalid snippet error",
			in:      &file.TaggedBytes{Tag: "../../../test/data/template.yml", Bytes: []byte(validTemplate)},
			snippet: &file.TaggedBytes{Tag: "opsfile.yml", Bytes: []byte("foo: bar")},
			opDefinitions: []patch.OpDefinition{
				newOpDefinition("", "", ""),
			},
			expectedError: errors.New(`Unknown operation [0] with type '' within
{
  "Path": "",
  "Value": "<redacted>"
}
  while trying to create ops from definitions in opsfile.yml`),
		},
		{
			name:    "template evalution error",
			in:      &file.TaggedBytes{Tag: "invalid.yml", Bytes: []byte(invalidTemplate)},
			snippet: &file.TaggedBytes{Tag: "opsfile.yml", Bytes: []byte("foo: bar")},
			opDefinitions: []patch.OpDefinition{
				newOpDefinition("replace", "/bizz?", "bazz"),
			},
			expectedError: errors.New("Expected to find a map at path '/bizz?' but found 'string'\n  while trying to evaluate template invalid.yml with op 0 from opsfile.yml"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockYaml := yaml.NewMockYamlAccess(ctrl)
			mockFile := file.NewMockFileAccess(ctrl)

			subject := NewOpsFileProcessor(mockYaml, mockFile)

			mockYaml.EXPECT().Unmarshal(c.snippet.Bytes, &[]patch.OpDefinition{}).Times(1).Return(c.parseSnippetError).Do(func(bytes []byte, o *[]patch.OpDefinition) {
				*o = c.opDefinitions
			})

			templateBytes, err := subject.ProcessTemplate(c.in, c.snippet)

			if !cmp.Equal(&c.expectedError, &err, cmp.Comparer(test.EqualMessage)) {
				t.Errorf("Expected error:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedError, err)
			}

			if err == nil && !cmp.Equal(templateBytes, c.expectedOut) {
				t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedOut, templateBytes)
			}
		})
	}
}

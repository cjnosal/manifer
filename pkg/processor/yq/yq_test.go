package yq

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	y2 "github.com/mikefarah/yaml/v2"

	"github.com/cjnosal/manifer/v2/pkg/file"
	"github.com/cjnosal/manifer/v2/pkg/library"
	"github.com/cjnosal/manifer/v2/pkg/processor"
	"github.com/cjnosal/manifer/v2/test"
	"github.com/google/go-cmp/cmp"
)

func litpnt(i interface{}) *interface{} {
	return &i
}

func newMapItem(key string, value interface{}) y2.MapItem {
	return y2.MapItem{
		Key:   key,
		Value: value,
	}
}

func TestParsePassthroughFlags(t *testing.T) {

	t.Run("yq", func(t *testing.T) {
		subject := yqProcessor{}
		flags := []string{"-sfoo", "-s", "bar", "--script=bizz"}
		node, remainder, err := subject.ParsePassthroughFlags(flags)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expectedNode := &library.ScenarioNode{
			Name:        "passthrough yq",
			Description: "args passed after --",
			LibraryPath: "<cli>",
			Snippets: []library.Snippet{
				{
					Path: "foo",
					Processor: library.Processor{
						Type: library.Yq,
						Options: map[string]interface{}{
							"command": "write",
						},
					},
				},
				{
					Path: "bar",
					Processor: library.Processor{
						Type: library.Yq,
						Options: map[string]interface{}{
							"command": "write",
						},
					},
				},
				{
					Path: "bizz",
					Processor: library.Processor{
						Type: library.Yq,
						Options: map[string]interface{}{
							"command": "write",
						},
					},
				},
			},
		}
		if !cmp.Equal(*expectedNode, *node) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", *expectedNode, *node)
		}

		expectedRemainder := []string{}
		if err == nil && !cmp.Equal(remainder, expectedRemainder) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedRemainder, remainder)
		}
	})

	t.Run("ignore other flags", func(t *testing.T) {
		subject := yqProcessor{}
		flags := []string{"-sfoo", "-vbar"}
		node, remainder, err := subject.ParsePassthroughFlags(flags)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expectedNode := &library.ScenarioNode{
			Name:        "passthrough yq",
			Description: "args passed after --",
			LibraryPath: "<cli>",
			Snippets: []library.Snippet{
				{
					Path: "foo",
					Processor: library.Processor{
						Type: library.Yq,
						Options: map[string]interface{}{
							"command": "write",
						},
					},
				},
			},
		}
		if !cmp.Equal(*expectedNode, *node) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", *expectedNode, *node)
		}

		expectedRemainder := []string{"-vbar"}
		if err == nil && !cmp.Equal(remainder, expectedRemainder) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedRemainder, remainder)
		}
	})

	t.Run("no snippets or path", func(t *testing.T) {
		subject := yqProcessor{}
		flags := []string{"-vbar"}
		node, remainder, err := subject.ParsePassthroughFlags(flags)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		var expectedNode *library.ScenarioNode
		if !cmp.Equal(expectedNode, node) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedNode, node)
		}

		expectedRemainder := []string{"-vbar"}
		if err == nil && !cmp.Equal(remainder, expectedRemainder) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedRemainder, remainder)
		}
	})

	t.Run("parse error", func(t *testing.T) {
		subject := yqProcessor{}
		flags := []string{"-s"}
		_, _, err := subject.ParsePassthroughFlags(flags)

		expectedError := "expected argument for flag `-s, --script'\n  while trying to parse yq args"
		if err == nil || err.Error() != expectedError {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expectedError, err)
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("valid yq script", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFile := file.NewMockFileAccess(ctrl)

		subject := NewYqProcessor(mockFile)

		commands := y2.MapSlice{
			y2.MapItem{
				Key:   "bar",
				Value: "asdf",
			},
		}
		bytes, _ := y2.Marshal(commands)

		mockFile.EXPECT().Read("/foo").Times(1).Return(bytes, nil)

		hint, err := subject.ValidateSnippet("/foo")

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expectedHint := processor.SnippetHint{
			Valid:   true,
			Element: "bar",
			Action:  "write",
		}

		if !cmp.Equal(expectedHint, hint) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedHint, hint)
		}
	})

	t.Run("invalid yq script", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFile := file.NewMockFileAccess(ctrl)

		subject := NewYqProcessor(mockFile)

		mockFile.EXPECT().Read("/foo").Times(1).Return([]byte{1}, nil)

		hint, err := subject.ValidateSnippet("/foo")

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		if hint.Valid {
			t.Errorf("Expected ValidateSnippet to return false")
		}
	})

	t.Run("file error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFile := file.NewMockFileAccess(ctrl)

		subject := NewYqProcessor(mockFile)

		mockFile.EXPECT().Read("/foo").Times(1).Return(nil, errors.New("oops"))

		hint, err := subject.ValidateSnippet("/foo")

		expectedError := errors.New("oops\n  while validating yq script /foo")
		if !cmp.Equal(&expectedError, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected error:\n'''%s'''\nActual:\n'''%s'''\n", expectedError, err)
		}

		if hint.Valid {
			t.Errorf("Expected ValidateSnippet to return false")
		}
	})
}

func TestProcessTemplate(t *testing.T) {

	validTemplate := "foo: bar\n"
	invalidTemplate := ":::not yaml"
	cases := []struct {
		name    string
		in      *file.TaggedBytes
		snippet *file.TaggedBytes

		processorOptions map[string]interface{}

		expectedError error
		expectedOut   []byte
	}{
		{
			name:        "empty script",
			in:          &file.TaggedBytes{Tag: "../../../test/data/v2/template.yml", Bytes: []byte(validTemplate)},
			snippet:     &file.TaggedBytes{Tag: "yq.yml", Bytes: []byte("")},
			expectedOut: []byte("foo: bar\n"),
		},
		{
			name:        "single default command",
			in:          &file.TaggedBytes{Tag: "../../../test/data/v2/template.yml", Bytes: []byte(validTemplate)},
			snippet:     &file.TaggedBytes{Tag: "yq.yml", Bytes: []byte("bizz: bazz")},
			expectedOut: []byte("foo: bar\nbizz: bazz\n"),
		},
		{
			name:    "single write command",
			in:      &file.TaggedBytes{Tag: "../../../test/data/v2/template.yml", Bytes: []byte(validTemplate)},
			snippet: &file.TaggedBytes{Tag: "yq.yml", Bytes: []byte("bizz: bazz")},
			processorOptions: map[string]interface{}{
				"command": "write",
			},
			expectedOut: []byte("foo: bar\nbizz: bazz\n"),
		},
		{
			name: "single read command",
			in:   &file.TaggedBytes{Tag: "../../../test/data/v2/template.yml", Bytes: []byte(validTemplate)},
			processorOptions: map[string]interface{}{
				"command": "read",
				"path":    "foo",
			},
			expectedOut: []byte("bar\n"),
		},
		{
			name: "single delete command",
			in:   &file.TaggedBytes{Tag: "../../../test/data/v2/template.yml", Bytes: []byte(validTemplate)},
			processorOptions: map[string]interface{}{
				"command": "delete",
				"path":    "foo",
			},
			expectedOut: []byte("{}\n"),
		},
		{
			name:    "single merge command",
			in:      &file.TaggedBytes{Tag: "../../../test/data/v2/template.yml", Bytes: []byte(validTemplate)},
			snippet: &file.TaggedBytes{Tag: "yq.yml", Bytes: []byte("bizz: bazz\nfoo: new")},
			processorOptions: map[string]interface{}{
				"command": "merge",
			},
			expectedOut: []byte("bizz: bazz\nfoo: bar\n"),
		},
		{
			name: "single prefix command",
			in:   &file.TaggedBytes{Tag: "../../../test/data/v2/template.yml", Bytes: []byte(validTemplate)},
			processorOptions: map[string]interface{}{
				"command": "prefix",
				"prefix":  "nest",
			},
			expectedOut: []byte("nest:\n  foo: bar\n"),
		},
		{
			name:        "multiple commands in file",
			in:          &file.TaggedBytes{Tag: "../../../test/data/v2/template.yml", Bytes: []byte(validTemplate)},
			snippet:     &file.TaggedBytes{Tag: "yq.yml", Bytes: []byte("bizz: bazz\nbazz[+]: buzz")},
			expectedOut: []byte("foo: bar\nbizz: bazz\nbazz:\n- buzz\n"),
		},
		{
			name:          "parse snippet error",
			in:            &file.TaggedBytes{Tag: "../../../test/data/v2/template.yml", Bytes: []byte(validTemplate)},
			snippet:       &file.TaggedBytes{Tag: "/originalsnippet", Bytes: []byte(invalidTemplate)},
			expectedError: errors.New("yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `:::not ...` into yaml.MapSlice\n  unmarshaling snippet"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFile := file.NewMockFileAccess(ctrl)

			subject := NewYqProcessor(mockFile)

			templateBytes, err := subject.ProcessTemplate(c.in, c.snippet, c.processorOptions)

			if !cmp.Equal(&c.expectedError, &err, cmp.Comparer(test.EqualMessage)) {
				t.Errorf("Expected error:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedError, err)
			}

			if err == nil && !cmp.Equal(templateBytes, c.expectedOut) {
				t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedOut, templateBytes)
			}
		})
	}
}

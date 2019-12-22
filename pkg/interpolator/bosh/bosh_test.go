package bosh

import (
	"errors"
	"testing"

	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/test"
	"github.com/google/go-cmp/cmp"
)

func TestInterpolate(t *testing.T) {

	t.Run("no args", func(t *testing.T) {
		template := &file.TaggedBytes{Tag: "../../../test/data/template.yml", Bytes: []byte("foo: bar\n\n")}

		bytes, err := NewBoshInterpolator().Interpolate(template, []string{})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expectedBytes := []byte("foo: bar\n\n")
		if err == nil && !cmp.Equal(bytes, expectedBytes) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expectedBytes, bytes)
		}
	})

	t.Run("valid args", func(t *testing.T) {
		template := &file.TaggedBytes{Tag: "../../../test/data/template_with_var.yml", Bytes: []byte("foo: ((bar))\n")}

		bytes, err := NewBoshInterpolator().Interpolate(template, []string{"-vbar=bizz"})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expectedBytes := []byte("foo: bizz\n")
		if err == nil && !cmp.Equal(bytes, expectedBytes) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expectedBytes, bytes)
		}
	})

	t.Run("invalid arg", func(t *testing.T) {
		template := &file.TaggedBytes{Tag: "../../../test/data/template.yml", Bytes: []byte("foo: bar\n\n")}

		_, err := NewBoshInterpolator().Interpolate(template, []string{"--invalid"})

		expectedError := errors.New("unknown flag `invalid'\n  while trying to parse args")
		if !cmp.Equal(&expectedError, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected error:\n'''%s'''\nActual:\n'''%s'''\n", expectedError, err)
		}
	})
}

func TestParsePassthroughVars(t *testing.T) {

	t.Run("no args", func(t *testing.T) {
		node, err := NewBoshInterpolator().ParsePassthroughVars([]string{})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		var expectedNode *library.ScenarioNode
		if err == nil && !cmp.Equal(node, expectedNode) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedNode, node)
		}
	})

	t.Run("valid args", func(t *testing.T) {
		node, err := NewBoshInterpolator().ParsePassthroughVars([]string{"-vbar=bizz"})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expectedNode := &library.ScenarioNode{
			Name:        "passthrough variables",
			Description: "vars passed after --",
			LibraryPath: "<cli>",
			Type:        "",
			GlobalArgs:  []string{"-vbar=bizz"},
			RefArgs:     []string{},
			Snippets:    []library.Snippet{},
		}
		if err == nil && !cmp.Equal(node, expectedNode) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedNode, node)
		}
	})

	t.Run("ignore invalid arg", func(t *testing.T) {
		node, err := NewBoshInterpolator().ParsePassthroughVars([]string{"--invalid"})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		var expectedNode *library.ScenarioNode
		if err == nil && !cmp.Equal(node, expectedNode) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedNode, node)
		}
	})
}

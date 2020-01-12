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
		template := &file.TaggedBytes{Tag: "../../../test/data/v2/template.yml", Bytes: []byte("foo: bar\n\n")}

		bytes, err := NewBoshInterpolator().Interpolate(template, library.InterpolatorParams{})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expectedBytes := []byte("foo: bar\n\n")
		if err == nil && !cmp.Equal(bytes, expectedBytes) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expectedBytes, bytes)
		}
	})

	t.Run("valid vars", func(t *testing.T) {
		template := &file.TaggedBytes{Tag: "../../../test/data/v2/template_with_var.yml", Bytes: []byte("foo: ((bar))\n")}

		bytes, err := NewBoshInterpolator().Interpolate(template, library.InterpolatorParams{Vars: map[string]interface{}{"bar": "bizz"}})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expectedBytes := []byte("foo: bizz\n")
		if err == nil && !cmp.Equal(bytes, expectedBytes) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expectedBytes, bytes)
		}
	})

	t.Run("valid args", func(t *testing.T) {
		template := &file.TaggedBytes{Tag: "../../../test/data/v2/template_with_var.yml", Bytes: []byte("foo: ((bar))\n")}

		bytes, err := NewBoshInterpolator().Interpolate(template, library.InterpolatorParams{RawArgs: []string{"-vbar=bizz"}})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expectedBytes := []byte("foo: bizz\n")
		if err == nil && !cmp.Equal(bytes, expectedBytes) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expectedBytes, bytes)
		}
	})

	t.Run("non-kv args do not override vars", func(t *testing.T) {
		template := &file.TaggedBytes{Tag: "../../../test/data/v2/template_with_var.yml", Bytes: []byte("foo: ((bar))\n")}

		bytes, err := NewBoshInterpolator().Interpolate(template, library.InterpolatorParams{Vars: map[string]interface{}{"bar": "bizz"}, RawArgs: []string{"-l=../../../test/data/v2/vars.yml"}})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expectedBytes := []byte("foo: bizz\n")
		if err == nil && !cmp.Equal(bytes, expectedBytes) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expectedBytes, bytes)
		}
	})

	t.Run("invalid arg", func(t *testing.T) {
		template := &file.TaggedBytes{Tag: "../../../test/data/v2/template_with_var.yml", Bytes: []byte("foo: ((bar))\n")}

		_, err := NewBoshInterpolator().Interpolate(template, library.InterpolatorParams{RawArgs: []string{"--invalid"}})

		expectedError := errors.New("unknown flag `invalid'\n  while trying to parse vars")
		if !cmp.Equal(&err, &expectedError, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("'%s'", cmp.Diff(err.Error(), expectedError.Error()))
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expectedError, err)
		}
	})
}

func TestParsePassthroughVars(t *testing.T) {

	t.Run("no args", func(t *testing.T) {
		node, remainder, err := NewBoshInterpolator().ParsePassthroughVars([]string{})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		var expectedNode *library.ScenarioNode
		if err == nil && !cmp.Equal(node, expectedNode) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedNode, node)
		}

		expectedRemainder := []string{}
		if err == nil && !cmp.Equal(remainder, expectedRemainder) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedRemainder, remainder)
		}
	})

	t.Run("valid args", func(t *testing.T) {
		node, remainder, err := NewBoshInterpolator().ParsePassthroughVars([]string{"-vbar=bizz"})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expectedNode := &library.ScenarioNode{
			Name:        "passthrough variables",
			Description: "vars passed after --",
			LibraryPath: "<cli>",
			GlobalInterpolator: library.InterpolatorParams{
				RawArgs: []string{"-vbar=bizz"},
			},
		}
		if err == nil && !cmp.Equal(node, expectedNode) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedNode, node)
		}

		expectedRemainder := []string{}
		if err == nil && !cmp.Equal(remainder, expectedRemainder) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedRemainder, remainder)
		}
	})

	t.Run("ignore invalid arg", func(t *testing.T) {
		node, remainder, err := NewBoshInterpolator().ParsePassthroughVars([]string{"--invalid"})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		var expectedNode *library.ScenarioNode
		if err == nil && !cmp.Equal(node, expectedNode) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedNode, node)
		}

		expectedRemainder := []string{"--invalid"}
		if err == nil && !cmp.Equal(remainder, expectedRemainder) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedRemainder, remainder)
		}
	})
}

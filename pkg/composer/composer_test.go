package composer

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/plan"
	"github.com/cjnosal/manifer/test"
)

type interpolation struct {
	showPlan     bool
	showDiff     bool
	in           *file.TaggedBytes
	out          []byte
	snippet      *file.TaggedBytes
	snippetArgs  []string
	templateArgs []string
	err          error
}

func TestCompose(t *testing.T) {

	planWithGlobals := &plan.Plan{
		Global: plan.ArgSet{
			Tag:  "global",
			Args: []string{"global arg", "cli arg"},
		},
		Steps: []*plan.Step{
			{
				Snippet: "/snippet",
				Args: []plan.ArgSet{
					{
						Tag:  "snippet",
						Args: []string{"snippet args"},
					},
				},
			},
		},
	}

	planWithoutGlobals := &plan.Plan{
		Global: plan.ArgSet{
			Tag:  "global",
			Args: []string{},
		},
		Steps: []*plan.Step{
			{
				Snippet: "/snippet",
				Args: []plan.ArgSet{
					{
						Tag:  "snippet",
						Args: []string{"snippet args"},
					},
				},
			},
		},
	}

	t.Run("no-op template", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockExecutor := plan.NewMockExecutor(ctrl)
		mockResolver := NewMockScenarioResolver(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		subject := ComposerImpl{
			Resolver: mockResolver,
			File:     mockFile,
			Executor: mockExecutor,
		}

		expectedOut := []byte("base")
		template := "/tmp/base.yml"
		taggedTemplate := &file.TaggedBytes{Tag: template, Bytes: expectedOut}

		mockResolver.EXPECT().Resolve(nil, nil, nil).Times(1).Return(&plan.Plan{}, nil)
		mockFile.EXPECT().ReadAndTag(template).Times(1).Return(taggedTemplate, nil)

		out, err := subject.Compose(template, nil, nil, nil, false, false)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		} else {
			if !cmp.Equal(expectedOut, out) {
				t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
					expectedOut, out, cmp.Diff(expectedOut, out))
			}
		}
	})

	t.Run("template and one scenario", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockExecutor := plan.NewMockExecutor(ctrl)
		mockResolver := NewMockScenarioResolver(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		subject := ComposerImpl{
			Resolver: mockResolver,
			File:     mockFile,
			Executor: mockExecutor,
		}
		libraries := []string{
			"/tmp/library/lib.yml",
		}
		scenarioNames := []string{
			"a scenario",
		}
		expectedOut := []byte("base")
		template := "/tmp/base.yml"
		taggedTemplate := &file.TaggedBytes{Tag: template, Bytes: []byte("in")}
		taggedSnippet := &file.TaggedBytes{Tag: planWithoutGlobals.Steps[0].Snippet, Bytes: []byte("op")}

		mockResolver.EXPECT().Resolve(libraries, scenarioNames, nil).Times(1).Return(planWithoutGlobals, nil)
		mockFile.EXPECT().ReadAndTag(template).Times(1).Return(taggedTemplate, nil)
		mockFile.EXPECT().ReadAndTag(taggedSnippet.Tag).Times(1).Return(taggedSnippet, nil)
		mockExecutor.EXPECT().Execute(false, false, taggedTemplate, taggedSnippet, planWithoutGlobals.Steps[0].FlattenArgs(), []string{}).Times(1).Return(expectedOut, nil)
		out, err := subject.Compose(template, libraries, scenarioNames, nil, false, false)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		} else {
			if !cmp.Equal(expectedOut, out) {
				t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
					expectedOut, out, cmp.Diff(expectedOut, out))
			}
		}
	})

	t.Run("post snippet args", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockExecutor := plan.NewMockExecutor(ctrl)
		mockResolver := NewMockScenarioResolver(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		subject := ComposerImpl{
			Resolver: mockResolver,
			File:     mockFile,
			Executor: mockExecutor,
		}
		libraries := []string{
			"/tmp/library/lib.yml",
		}
		scenarioNames := []string{
			"a scenario",
		}
		passthrough :=
			[]string{
				"cli arg",
			}
		expectedOut := []byte("base")
		template := "/tmp/base.yml"
		taggedTemplate := &file.TaggedBytes{Tag: template, Bytes: []byte("in")}
		taggedSnippet := &file.TaggedBytes{Tag: planWithoutGlobals.Steps[0].Snippet, Bytes: []byte("op")}

		mockResolver.EXPECT().Resolve(libraries, scenarioNames, passthrough).Times(1).Return(planWithGlobals, nil)
		mockFile.EXPECT().ReadAndTag(template).Times(1).Return(taggedTemplate, nil)
		mockFile.EXPECT().ReadAndTag(taggedSnippet.Tag).Times(1).Return(taggedSnippet, nil)
		mockExecutor.EXPECT().Execute(false, false, taggedTemplate, taggedSnippet, planWithGlobals.Steps[0].FlattenArgs(), planWithGlobals.Global.Args).Times(1).Return([]byte("transient"), nil)
		mockExecutor.EXPECT().Execute(false, false, &file.TaggedBytes{Tag: template, Bytes: []byte("transient")}, nil, nil, planWithGlobals.Global.Args).Times(1).Return(expectedOut, nil)
		out, err := subject.Compose(template, libraries, scenarioNames, passthrough, false, false)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		} else {
			if !cmp.Equal(expectedOut, out) {
				t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
					expectedOut, out, cmp.Diff(expectedOut, out))
			}
		}
	})

	t.Run("scenario resolution error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockExecutor := plan.NewMockExecutor(ctrl)
		mockResolver := NewMockScenarioResolver(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		subject := ComposerImpl{
			Resolver: mockResolver,
			File:     mockFile,
			Executor: mockExecutor,
		}
		libraries := []string{
			"/tmp/library/lib.yml",
		}
		scenarioNames := []string{
			"a scenario",
		}
		passthrough :=
			[]string{
				"cli arg",
			}
		template := "/tmp/base.yml"
		resolverError := errors.New("test")
		expectedError := errors.New("test\n  while trying to resolve scenarios")

		mockResolver.EXPECT().Resolve(libraries, scenarioNames, passthrough).Times(1).Return(nil, resolverError)
		_, err := subject.Compose(template, libraries, scenarioNames, passthrough, false, false)

		if !cmp.Equal(&err, &expectedError, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("'%s'", cmp.Diff(err.Error(), expectedError.Error()))
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expectedError, err)
		}
	})

	t.Run("read template error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockExecutor := plan.NewMockExecutor(ctrl)
		mockResolver := NewMockScenarioResolver(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		subject := ComposerImpl{
			Resolver: mockResolver,
			File:     mockFile,
			Executor: mockExecutor,
		}
		libraries := []string{
			"/tmp/library/lib.yml",
		}
		scenarioNames := []string{
			"a scenario",
		}
		passthrough :=
			[]string{
				"cli arg",
			}
		template := "/tmp/base.yml"
		loadError := errors.New("test")
		expectedError := errors.New("test\n  while trying to load template /tmp/base.yml")

		mockResolver.EXPECT().Resolve(libraries, scenarioNames, passthrough).Times(1).Return(planWithGlobals, nil)
		mockFile.EXPECT().ReadAndTag(template).Times(1).Return(nil, loadError)
		_, err := subject.Compose(template, libraries, scenarioNames, passthrough, false, false)

		if err == nil || err.Error() != expectedError.Error() {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expectedError, err)
		}
	})

	t.Run("load snippet error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockExecutor := plan.NewMockExecutor(ctrl)
		mockResolver := NewMockScenarioResolver(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		subject := ComposerImpl{
			Resolver: mockResolver,
			File:     mockFile,
			Executor: mockExecutor,
		}
		libraries := []string{
			"/tmp/library/lib.yml",
		}
		scenarioNames := []string{
			"a scenario",
		}
		passthrough :=
			[]string{
				"cli arg",
			}
		template := "/tmp/base.yml"
		taggedTemplate := &file.TaggedBytes{Tag: template, Bytes: []byte("in")}
		snippetError := errors.New("test")
		expectedError := errors.New("test\n  while trying to load snippet /snippet")

		mockResolver.EXPECT().Resolve(libraries, scenarioNames, passthrough).Times(1).Return(planWithGlobals, nil)
		mockFile.EXPECT().ReadAndTag(template).Times(1).Return(taggedTemplate, nil)
		mockFile.EXPECT().ReadAndTag("/snippet").Times(1).Return(nil, snippetError)
		_, err := subject.Compose(template, libraries, scenarioNames, passthrough, false, false)

		if err == nil || err.Error() != expectedError.Error() {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expectedError, err)
		}
	})

	t.Run("interpolate snippet error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockExecutor := plan.NewMockExecutor(ctrl)
		mockResolver := NewMockScenarioResolver(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		subject := ComposerImpl{
			Resolver: mockResolver,
			File:     mockFile,
			Executor: mockExecutor,
		}
		libraries := []string{
			"/tmp/library/lib.yml",
		}
		scenarioNames := []string{
			"a scenario",
		}
		passthrough :=
			[]string{
				"cli arg",
			}
		template := "/tmp/base.yml"
		taggedTemplate := &file.TaggedBytes{Tag: template, Bytes: []byte("in")}
		taggedSnippet := &file.TaggedBytes{Tag: planWithoutGlobals.Steps[0].Snippet, Bytes: []byte("op")}
		snippetError := errors.New("test")
		expectedError := errors.New("test\n  while trying to apply snippet /snippet")

		mockResolver.EXPECT().Resolve(libraries, scenarioNames, passthrough).Times(1).Return(planWithGlobals, nil)
		mockFile.EXPECT().ReadAndTag(template).Times(1).Return(taggedTemplate, nil)
		mockFile.EXPECT().ReadAndTag(taggedSnippet.Tag).Times(1).Return(taggedSnippet, nil)
		mockExecutor.EXPECT().Execute(false, false, taggedTemplate, taggedSnippet, planWithGlobals.Steps[0].FlattenArgs(), planWithGlobals.Global.Args).Times(1).Return(nil, snippetError)

		_, err := subject.Compose(template, libraries, scenarioNames, passthrough, false, false)

		if err == nil || err.Error() != expectedError.Error() {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expectedError, err)
		}
	})

	t.Run("interpolate template error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockExecutor := plan.NewMockExecutor(ctrl)
		mockResolver := NewMockScenarioResolver(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		subject := ComposerImpl{
			Resolver: mockResolver,
			File:     mockFile,
			Executor: mockExecutor,
		}
		libraries := []string{
			"/tmp/library/lib.yml",
		}
		scenarioNames := []string{
			"a scenario",
		}
		passthrough :=
			[]string{
				"cli arg",
			}
		template := "/tmp/base.yml"
		taggedTemplate := &file.TaggedBytes{Tag: template, Bytes: []byte("in")}
		taggedSnippet := &file.TaggedBytes{Tag: planWithoutGlobals.Steps[0].Snippet, Bytes: []byte("op")}
		intError := errors.New("test")
		expectedError := errors.New("test\n  while trying to apply passthrough args [global arg cli arg]")

		mockResolver.EXPECT().Resolve(libraries, scenarioNames, passthrough).Times(1).Return(planWithGlobals, nil)
		mockFile.EXPECT().ReadAndTag(template).Times(1).Return(taggedTemplate, nil)
		mockFile.EXPECT().ReadAndTag(taggedSnippet.Tag).Times(1).Return(taggedSnippet, nil)
		mockExecutor.EXPECT().Execute(false, false, taggedTemplate, taggedSnippet, planWithGlobals.Steps[0].FlattenArgs(), planWithGlobals.Global.Args).Times(1).Return([]byte("transient"), nil)
		mockExecutor.EXPECT().Execute(false, false, &file.TaggedBytes{Tag: template, Bytes: []byte("transient")}, nil, nil, planWithGlobals.Global.Args).Times(1).Return(nil, intError)

		_, err := subject.Compose(template, libraries, scenarioNames, passthrough, false, false)

		if err == nil || err.Error() != expectedError.Error() {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expectedError, err)
		}
	})

}

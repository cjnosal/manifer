package composer

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/plan"
	"github.com/cjnosal/manifer/pkg/scenario"
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

	planWithGlobals := &scenario.Plan{
		GlobalArgs: []string{"global arg", "cli arg"},
		Snippets: []library.Snippet{
			{
				Path: "/snippet",
				Args: []string{
					"snippet",
					"args",
				},
			},
		},
	}

	planWithoutGlobals := &scenario.Plan{
		GlobalArgs: []string{},
		Snippets: []library.Snippet{
			{
				Path: "/snippet",
				Args: []string{
					"snippet",
					"args",
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
		}

		expectedOut := []byte("base")
		template := "/tmp/base.yml"
		taggedTemplate := &file.TaggedBytes{Tag: template, Bytes: expectedOut}

		mockResolver.EXPECT().Resolve(nil, nil, nil).Times(1).Return(&scenario.Plan{}, nil)
		mockFile.EXPECT().ReadAndTag(template).Times(1).Return(taggedTemplate, nil)

		out, err := subject.Compose(mockExecutor, template, nil, nil, nil, false, false)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		} else {
			if !reflect.DeepEqual(expectedOut, out) {
				t.Errorf("Expected output:\n'''%s'''\nActual:\n'''%s'''\n", expectedOut, out)
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
		taggedSnippet := &file.TaggedBytes{Tag: planWithoutGlobals.Snippets[0].Path, Bytes: []byte("op")}

		mockResolver.EXPECT().Resolve(libraries, scenarioNames, nil).Times(1).Return(planWithoutGlobals, nil)
		mockFile.EXPECT().ReadAndTag(template).Times(1).Return(taggedTemplate, nil)
		mockFile.EXPECT().ReadAndTag(taggedSnippet.Tag).Times(1).Return(taggedSnippet, nil)
		mockExecutor.EXPECT().Execute(false, false, taggedTemplate, taggedSnippet, planWithoutGlobals.Snippets[0].Args, []string{}).Times(1).Return(expectedOut, nil)
		out, err := subject.Compose(mockExecutor, template, libraries, scenarioNames, nil, false, false)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		} else {
			if !reflect.DeepEqual(expectedOut, out) {
				t.Errorf("Expected output:\n'''%s'''\nActual:\n'''%s'''\n", expectedOut, out)
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
		taggedSnippet := &file.TaggedBytes{Tag: planWithoutGlobals.Snippets[0].Path, Bytes: []byte("op")}

		mockResolver.EXPECT().Resolve(libraries, scenarioNames, passthrough).Times(1).Return(planWithGlobals, nil)
		mockFile.EXPECT().ReadAndTag(template).Times(1).Return(taggedTemplate, nil)
		mockFile.EXPECT().ReadAndTag(taggedSnippet.Tag).Times(1).Return(taggedSnippet, nil)
		mockExecutor.EXPECT().Execute(false, false, taggedTemplate, taggedSnippet, planWithGlobals.Snippets[0].Args, planWithGlobals.GlobalArgs).Times(1).Return([]byte("transient"), nil)
		mockExecutor.EXPECT().Execute(false, false, &file.TaggedBytes{Tag: template, Bytes: []byte("transient")}, nil, nil, planWithGlobals.GlobalArgs).Times(1).Return(expectedOut, nil)
		out, err := subject.Compose(mockExecutor, template, libraries, scenarioNames, passthrough, false, false)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		} else {
			if !reflect.DeepEqual(expectedOut, out) {
				t.Errorf("Expected output:\n'''%s'''\nActual:\n'''%s'''\n", expectedOut, out)
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
		_, err := subject.Compose(mockExecutor, template, libraries, scenarioNames, passthrough, false, false)

		if err == nil || err.Error() != expectedError.Error() {
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
		_, err := subject.Compose(mockExecutor, template, libraries, scenarioNames, passthrough, false, false)

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
		_, err := subject.Compose(mockExecutor, template, libraries, scenarioNames, passthrough, false, false)

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
		taggedSnippet := &file.TaggedBytes{Tag: planWithoutGlobals.Snippets[0].Path, Bytes: []byte("op")}
		snippetError := errors.New("test")
		expectedError := errors.New("test\n  while trying to apply snippet /snippet")

		mockResolver.EXPECT().Resolve(libraries, scenarioNames, passthrough).Times(1).Return(planWithGlobals, nil)
		mockFile.EXPECT().ReadAndTag(template).Times(1).Return(taggedTemplate, nil)
		mockFile.EXPECT().ReadAndTag(taggedSnippet.Tag).Times(1).Return(taggedSnippet, nil)
		mockExecutor.EXPECT().Execute(false, false, taggedTemplate, taggedSnippet, planWithGlobals.Snippets[0].Args, planWithGlobals.GlobalArgs).Times(1).Return(nil, snippetError)

		_, err := subject.Compose(mockExecutor, template, libraries, scenarioNames, passthrough, false, false)

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
		taggedSnippet := &file.TaggedBytes{Tag: planWithoutGlobals.Snippets[0].Path, Bytes: []byte("op")}
		intError := errors.New("test")
		expectedError := errors.New("test\n  while trying to apply passthrough args [global arg cli arg]")

		mockResolver.EXPECT().Resolve(libraries, scenarioNames, passthrough).Times(1).Return(planWithGlobals, nil)
		mockFile.EXPECT().ReadAndTag(template).Times(1).Return(taggedTemplate, nil)
		mockFile.EXPECT().ReadAndTag(taggedSnippet.Tag).Times(1).Return(taggedSnippet, nil)
		mockExecutor.EXPECT().Execute(false, false, taggedTemplate, taggedSnippet, planWithGlobals.Snippets[0].Args, planWithGlobals.GlobalArgs).Times(1).Return([]byte("transient"), nil)
		mockExecutor.EXPECT().Execute(false, false, &file.TaggedBytes{Tag: template, Bytes: []byte("transient")}, nil, nil, planWithGlobals.GlobalArgs).Times(1).Return(nil, intError)

		_, err := subject.Compose(mockExecutor, template, libraries, scenarioNames, passthrough, false, false)

		if err == nil || err.Error() != expectedError.Error() {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", expectedError, err)
		}
	})

}

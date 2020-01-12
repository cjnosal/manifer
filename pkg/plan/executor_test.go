package plan

import (
	"errors"
	"testing"

	"github.com/cjnosal/manifer/v2/pkg/diff"
	"github.com/cjnosal/manifer/v2/pkg/file"
	"github.com/cjnosal/manifer/v2/pkg/interpolator"
	"github.com/cjnosal/manifer/v2/pkg/library"
	"github.com/cjnosal/manifer/v2/pkg/processor"
	"github.com/cjnosal/manifer/v2/pkg/processor/factory"
	"github.com/cjnosal/manifer/v2/pkg/yaml"
	"github.com/cjnosal/manifer/v2/test"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
)

func TestExecute(t *testing.T) {

	t.Run("Show plan", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDiff := diff.NewMockDiff(ctrl)
		mockProcessor := processor.NewMockProcessor(ctrl)
		mockInterpolator := interpolator.NewMockInterpolator(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockYaml := yaml.NewMockYamlAccess(ctrl)
		writer := &test.StringWriter{}
		defer ctrl.Finish()

		subject := &InterpolationExecutor{
			Diff:             mockDiff,
			ProcessorFactory: mockProcessorFactory,
			Interpolator:     mockInterpolator,
			Output:           writer,
			File:             mockFile,
			Yaml:             mockYaml,
		}

		in := &file.TaggedBytes{Tag: "in", Bytes: []byte("foo: bar")}
		processedIn := &file.TaggedBytes{Tag: "in", Bytes: []byte("bytes")}
		snippet := &file.TaggedBytes{Tag: "snippet", Bytes: []byte("bizz: bazz")}
		intSnippet := &file.TaggedBytes{Tag: "snippet", Bytes: []byte("intSnippetBytes")}
		globals := library.InterpolatorParams{
			Vars:    map[string]interface{}{"global": "gargs"},
			RawArgs: []string{"-vfoo=bar"},
		}
		snippetVars := library.InterpolatorParams{
			Vars: map[string]interface{}{"snippet": "sargs"},
		}
		snippetProcessor := &library.Processor{
			Type:    library.OpsFile,
			Options: map[string]interface{}{},
		}
		executorStep := ExecutorStep{
			Processor:    snippetProcessor,
			Interpolator: snippetVars.Merge(globals),
			Snippet:      "../snippet",
		}

		mockInterpolator.EXPECT().Interpolate(snippet, library.InterpolatorParams{Vars: map[string]interface{}{"snippet": "sargs", "global": "gargs"}, RawArgs: []string{"-vfoo=bar"}}).Times(1).Return([]byte("intSnippetBytes"), nil)
		mockInterpolator.EXPECT().Interpolate(processedIn, library.InterpolatorParams{Vars: map[string]interface{}{"global": "gargs"}, RawArgs: []string{"-vfoo=bar"}}).Times(1).Return([]byte("intTemplateBytes"), nil)
		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockProcessor.EXPECT().ProcessTemplate(in, intSnippet, snippetProcessor.Options).Times(1).Return([]byte("bytes"), nil)
		mockFile.EXPECT().ResolveRelativeFromWD("snippet").Times(1).Return("../snippet", nil)
		mockYaml.EXPECT().Marshal(executorStep).Times(1).Return([]byte("yamlstep"), nil)
		bytes, err := subject.Execute(true, false, in, snippet, snippetProcessor, snippetVars, globals)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		} else if !cmp.Equal(bytes, []byte("intTemplateBytes")) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", "bytes", string(bytes))
		}

		expectedStep := "\nyamlstep"
		if writer.String() != expectedStep {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedStep, writer.String())
		}

	})

	t.Run("Show diff", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDiff := diff.NewMockDiff(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		mockProcessor := processor.NewMockProcessor(ctrl)
		mockInterpolator := interpolator.NewMockInterpolator(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockYaml := yaml.NewMockYamlAccess(ctrl)
		writer := &test.StringWriter{}
		defer ctrl.Finish()

		subject := &InterpolationExecutor{
			Diff:             mockDiff,
			ProcessorFactory: mockProcessorFactory,
			Interpolator:     mockInterpolator,
			Output:           writer,
			File:             mockFile,
			Yaml:             mockYaml,
		}

		expectedDiff := "\nDiff:\ndiff"
		in := &file.TaggedBytes{Tag: "in", Bytes: []byte("foo: bar")}
		processedIn := &file.TaggedBytes{Tag: "in", Bytes: []byte("bytes")}
		snippet := &file.TaggedBytes{Tag: "snippet", Bytes: []byte("bizz: bazz")}
		intSnippet := &file.TaggedBytes{Tag: "snippet", Bytes: []byte("intSnippetBytes")}
		globals := library.InterpolatorParams{
			Vars:    map[string]interface{}{"global": "gargs"},
			RawArgs: []string{"-vfoo=bar"},
		}
		snippetVars := library.InterpolatorParams{
			Vars: map[string]interface{}{"snippet": "sargs"},
		}
		snippetProcessor := &library.Processor{
			Type:    library.OpsFile,
			Options: map[string]interface{}{},
		}

		mockInterpolator.EXPECT().Interpolate(snippet, library.InterpolatorParams{Vars: map[string]interface{}{"snippet": "sargs", "global": "gargs"}, RawArgs: []string{"-vfoo=bar"}}).Times(1).Return([]byte("intSnippetBytes"), nil)
		mockInterpolator.EXPECT().Interpolate(processedIn, library.InterpolatorParams{Vars: map[string]interface{}{"global": "gargs"}, RawArgs: []string{"-vfoo=bar"}}).Times(1).Return([]byte("intTemplateBytes"), nil)
		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockProcessor.EXPECT().ProcessTemplate(in, intSnippet, snippetProcessor.Options).Times(1).Return([]byte("bytes"), nil)
		mockDiff.EXPECT().StringDiff("foo: bar", "intTemplateBytes").Times(1).Return("diff")

		bytes, err := subject.Execute(false, true, in, snippet, snippetProcessor, snippetVars, globals)

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		} else if !cmp.Equal(bytes, []byte("intTemplateBytes")) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", "bytes", string(bytes))
		}

		if writer.String() != expectedDiff {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedDiff, writer.String())
		}

	})

	t.Run("Interpolate snippet error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDiff := diff.NewMockDiff(ctrl)
		mockInterpolator := interpolator.NewMockInterpolator(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockYaml := yaml.NewMockYamlAccess(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		writer := &test.StringWriter{}
		defer ctrl.Finish()

		subject := &InterpolationExecutor{
			Diff:             mockDiff,
			ProcessorFactory: mockProcessorFactory,
			Interpolator:     mockInterpolator,
			Output:           writer,
			File:             mockFile,
			Yaml:             mockYaml,
		}

		expectedError := errors.New("test\n  while trying to interpolate snippet")
		in := &file.TaggedBytes{Tag: "in", Bytes: []byte("foo: bar")}
		snippet := &file.TaggedBytes{Tag: "snippet", Bytes: []byte("bizz: bazz")}
		globals := library.InterpolatorParams{
			Vars:    map[string]interface{}{"global": "gargs"},
			RawArgs: []string{"-vfoo=bar"},
		}
		snippetVars := library.InterpolatorParams{
			Vars: map[string]interface{}{"snippet": "sargs"},
		}
		snippetProcessor := &library.Processor{
			Type:    library.OpsFile,
			Options: map[string]interface{}{},
		}

		mockInterpolator.EXPECT().Interpolate(snippet, library.InterpolatorParams{Vars: map[string]interface{}{"snippet": "sargs", "global": "gargs"}, RawArgs: []string{"-vfoo=bar"}}).Times(1).Return(nil, errors.New("test"))

		_, err := subject.Execute(false, false, in, snippet, snippetProcessor, snippetVars, globals)

		if !cmp.Equal(&expectedError, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedError, err)
		}

	})

	t.Run("Process error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDiff := diff.NewMockDiff(ctrl)
		mockProcessor := processor.NewMockProcessor(ctrl)
		mockInterpolator := interpolator.NewMockInterpolator(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockYaml := yaml.NewMockYamlAccess(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		writer := &test.StringWriter{}
		defer ctrl.Finish()

		subject := &InterpolationExecutor{
			Diff:             mockDiff,
			ProcessorFactory: mockProcessorFactory,
			Interpolator:     mockInterpolator,
			Output:           writer,
			File:             mockFile,
			Yaml:             mockYaml,
		}

		expectedError := errors.New("test\n  while trying to process template")
		in := &file.TaggedBytes{Tag: "in", Bytes: []byte("foo: bar")}
		snippet := &file.TaggedBytes{Tag: "snippet", Bytes: []byte("bizz: bazz")}
		intSnippet := &file.TaggedBytes{Tag: "snippet", Bytes: []byte("intSnippetBytes")}
		globals := library.InterpolatorParams{
			Vars:    map[string]interface{}{"global": "gargs"},
			RawArgs: []string{"-vfoo=bar"},
		}
		snippetVars := library.InterpolatorParams{
			Vars: map[string]interface{}{"snippet": "sargs"},
		}
		snippetProcessor := &library.Processor{
			Type:    library.OpsFile,
			Options: map[string]interface{}{},
		}

		mockInterpolator.EXPECT().Interpolate(snippet, library.InterpolatorParams{Vars: map[string]interface{}{"snippet": "sargs", "global": "gargs"}, RawArgs: []string{"-vfoo=bar"}}).Times(1).Return([]byte("intSnippetBytes"), nil)
		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockProcessor.EXPECT().ProcessTemplate(in, intSnippet, snippetProcessor.Options).Times(1).Return(nil, errors.New("test"))

		_, err := subject.Execute(false, false, in, snippet, snippetProcessor, snippetVars, globals)

		if !cmp.Equal(&expectedError, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedError, err)
		}

	})

	t.Run("Interpolate template error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDiff := diff.NewMockDiff(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		mockProcessor := processor.NewMockProcessor(ctrl)
		mockInterpolator := interpolator.NewMockInterpolator(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockYaml := yaml.NewMockYamlAccess(ctrl)
		writer := &test.StringWriter{}
		defer ctrl.Finish()

		subject := &InterpolationExecutor{
			Diff:             mockDiff,
			ProcessorFactory: mockProcessorFactory,
			Interpolator:     mockInterpolator,
			Output:           writer,
			File:             mockFile,
			Yaml:             mockYaml,
		}

		in := &file.TaggedBytes{Tag: "in", Bytes: []byte("foo: bar")}
		processedIn := &file.TaggedBytes{Tag: "in", Bytes: []byte("bytes")}
		snippet := &file.TaggedBytes{Tag: "snippet", Bytes: []byte("bizz: bazz")}
		intSnippet := &file.TaggedBytes{Tag: "snippet", Bytes: []byte("intSnippetBytes")}
		globals := library.InterpolatorParams{
			Vars:    map[string]interface{}{"global": "gargs"},
			RawArgs: []string{"-vfoo=bar"},
		}
		snippetVars := library.InterpolatorParams{
			Vars: map[string]interface{}{"snippet": "sargs"},
		}
		snippetProcessor := &library.Processor{
			Type:    library.OpsFile,
			Options: map[string]interface{}{},
		}

		mockInterpolator.EXPECT().Interpolate(snippet, library.InterpolatorParams{Vars: map[string]interface{}{"snippet": "sargs", "global": "gargs"}, RawArgs: []string{"-vfoo=bar"}}).Times(1).Return([]byte("intSnippetBytes"), nil)
		mockInterpolator.EXPECT().Interpolate(processedIn, library.InterpolatorParams{Vars: map[string]interface{}{"global": "gargs"}, RawArgs: []string{"-vfoo=bar"}}).Times(1).Return(nil, errors.New("test"))

		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockProcessor.EXPECT().ProcessTemplate(in, intSnippet, snippetProcessor.Options).Times(1).Return([]byte("bytes"), nil)

		expectedError := errors.New("test\n  while trying to interpolate template")
		_, err := subject.Execute(false, false, in, snippet, snippetProcessor, snippetVars, globals)

		if !cmp.Equal(&expectedError, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedError, err)
		}
	})
}

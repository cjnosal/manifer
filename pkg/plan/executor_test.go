package plan

import (
	"errors"
	"testing"

	"github.com/cjnosal/manifer/pkg/diff"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/processor"
	"github.com/cjnosal/manifer/test"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
)

func TestStringDiff(t *testing.T) {

	t.Run("Show plan", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDiff := diff.NewMockDiff(ctrl)
		mockProcessor := processor.NewMockProcessor(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		writer := &test.StringWriter{}
		defer ctrl.Finish()

		subject := &InterpolationExecutor{
			Diff:      mockDiff,
			Processor: mockProcessor,
			Output:    writer,
			File:      mockFile,
		}

		in := &file.TaggedBytes{Tag: "in", Bytes: []byte("foo: bar")}
		snippet := &file.TaggedBytes{Tag: "snippet", Bytes: []byte("bizz: bazz")}
		mockProcessor.EXPECT().ProcessTemplate(in, snippet, []string{"snippet args"}, []string{"global args"}).Times(1).Return([]byte("bytes"), nil)
		mockFile.EXPECT().ResolveRelativeFromWD("snippet").Times(1).Return("../snippet", nil)
		bytes, err := subject.Execute(true, false, in, snippet, []string{"snippet args"}, []string{"global args"})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		} else if !cmp.Equal(bytes, []byte("bytes")) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", "bytes", string(bytes))
		}

		expectedStep := "\nSnippet ../snippet; Arg [snippet args]; Global [global args]\n"
		if writer.String() != expectedStep {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedStep, writer.String())
		}

	})

	t.Run("Interpolation error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDiff := diff.NewMockDiff(ctrl)
		mockProcessor := processor.NewMockProcessor(ctrl)
		writer := &test.StringWriter{}
		defer ctrl.Finish()

		subject := &InterpolationExecutor{
			Diff:      mockDiff,
			Processor: mockProcessor,
			Output:    writer,
		}

		expectedError := errors.New("test")
		in := &file.TaggedBytes{Tag: "in", Bytes: []byte("foo: bar")}
		snippet := &file.TaggedBytes{Tag: "snippet", Bytes: []byte("bizz: bazz")}
		mockProcessor.EXPECT().ProcessTemplate(in, snippet, []string{"snippet args"}, []string{"global args"}).Times(1).Return(nil, expectedError)

		_, err := subject.Execute(false, false, in, snippet, []string{"snippet args"}, []string{"global args"})

		if !cmp.Equal(&expectedError, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedError, err)
		}

	})

	t.Run("Show diff", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDiff := diff.NewMockDiff(ctrl)
		mockProcessor := processor.NewMockProcessor(ctrl)
		writer := &test.StringWriter{}
		defer ctrl.Finish()

		subject := &InterpolationExecutor{
			Diff:      mockDiff,
			Processor: mockProcessor,
			Output:    writer,
		}

		expectedDiff := "Diff:\ndiff"
		in := &file.TaggedBytes{Tag: "in", Bytes: []byte("foo: bar")}
		snippet := &file.TaggedBytes{Tag: "snippet", Bytes: []byte("bizz: bazz")}
		mockProcessor.EXPECT().ProcessTemplate(in, snippet, []string{"snippet args"}, []string{"global args"}).Times(1).Return([]byte("bytes"), nil)
		mockDiff.EXPECT().StringDiff("foo: bar", "bytes").Times(1).Return("diff")

		bytes, err := subject.Execute(false, true, in, snippet, []string{"snippet args"}, []string{"global args"})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		} else if !cmp.Equal(bytes, []byte("bytes")) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", "bytes", string(bytes))
		}

		if writer.String() != expectedDiff {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedDiff, writer.String())
		}

	})

}

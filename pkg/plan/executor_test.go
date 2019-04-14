package plan

import (
	"errors"
	"reflect"
	"testing"

	"github.com/cjnosal/manifer/pkg/diff"
	"github.com/cjnosal/manifer/pkg/interpolator"
	"github.com/cjnosal/manifer/test"
	"github.com/golang/mock/gomock"
)

func TestFindDiff(t *testing.T) {

	t.Run("Show plan", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDiff := diff.NewMockDiff(ctrl)
		mockInterpolator := interpolator.NewMockInterpolator(ctrl)
		writer := &test.StringWriter{}
		defer ctrl.Finish()

		subject := &InterpolationExecutor{
			Diff:         mockDiff,
			Interpolator: mockInterpolator,
			Output:       writer,
		}

		mockInterpolator.EXPECT().Interpolate("in", "out", "snippet", []string{"snippet args"}, []string{"global args"}).Times(1).Return(nil)

		err := subject.Execute(true, false, "in", "out", "snippet", []string{"snippet args"}, []string{"global args"})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		expectedStep := "\nSnippet snippet; Arg [snippet args]; Global [global args]\n"
		if writer.String() != expectedStep {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedStep, writer.String())
		}

	})

	t.Run("Interpolation error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDiff := diff.NewMockDiff(ctrl)
		mockInterpolator := interpolator.NewMockInterpolator(ctrl)
		writer := &test.StringWriter{}
		defer ctrl.Finish()

		subject := &InterpolationExecutor{
			Diff:         mockDiff,
			Interpolator: mockInterpolator,
			Output:       writer,
		}

		expectedError := errors.New("test")
		mockInterpolator.EXPECT().Interpolate("in", "out", "snippet", []string{"snippet args"}, []string{"global args"}).Times(1).Return(expectedError)

		err := subject.Execute(false, false, "in", "out", "snippet", []string{"snippet args"}, []string{"global args"})

		if !reflect.DeepEqual(expectedError, err) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedError, err)
		}

	})

	t.Run("Show diff", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDiff := diff.NewMockDiff(ctrl)
		mockInterpolator := interpolator.NewMockInterpolator(ctrl)
		writer := &test.StringWriter{}
		defer ctrl.Finish()

		subject := &InterpolationExecutor{
			Diff:         mockDiff,
			Interpolator: mockInterpolator,
			Output:       writer,
		}

		expectedDiff := "Diff:\ndiff"
		mockInterpolator.EXPECT().Interpolate("in", "out", "snippet", []string{"snippet args"}, []string{"global args"}).Times(1).Return(nil)
		mockDiff.EXPECT().FindDiff("in", "out").Times(1).Return("diff", nil)

		err := subject.Execute(false, true, "in", "out", "snippet", []string{"snippet args"}, []string{"global args"})

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		if writer.String() != expectedDiff {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedDiff, writer.String())
		}

	})

	t.Run("Diff error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockDiff := diff.NewMockDiff(ctrl)
		mockInterpolator := interpolator.NewMockInterpolator(ctrl)
		writer := &test.StringWriter{}
		defer ctrl.Finish()

		subject := &InterpolationExecutor{
			Diff:         mockDiff,
			Interpolator: mockInterpolator,
			Output:       writer,
		}

		expectedError := errors.New("test")
		mockInterpolator.EXPECT().Interpolate("in", "out", "snippet", []string{"snippet args"}, []string{"global args"}).Times(1).Return(nil)
		mockDiff.EXPECT().FindDiff("in", "out").Times(1).Return("", expectedError)

		err := subject.Execute(false, true, "in", "out", "snippet", []string{"snippet args"}, []string{"global args"})

		if !reflect.DeepEqual(expectedError, err) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedError, err)
		}

	})

}

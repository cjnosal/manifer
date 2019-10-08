package diff

import (
	"errors"
	"testing"

	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/test"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestFindDiff(t *testing.T) {

	t.Run("Strings", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockFile := file.NewMockFileAccess(ctrl)
		mockPatch := NewMockdiffMatchPatch(ctrl)
		defer ctrl.Finish()

		subject := &FileDiff{
			File:  mockFile,
			Patch: mockPatch,
		}

		diff := diffmatchpatch.Diff{

			Type: diffmatchpatch.DiffInsert,
			Text: "diff",
		}
		expectedDiff := "pretty diff"

		mockPatch.EXPECT().DiffMain("content1", "content2", true).Times(1).Return([]diffmatchpatch.Diff{
			diff,
		})
		mockPatch.EXPECT().DiffPrettyText([]diffmatchpatch.Diff{
			diff,
		}).Times(1).Return(expectedDiff)

		prettyDiff := subject.StringDiff("content1", "content2")

		if expectedDiff != prettyDiff {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedDiff, prettyDiff)
		}
	})

	t.Run("Identical Strings", func(t *testing.T) {
		subject := &FileDiff{}

		diff := subject.StringDiff("content1", "content1")
		expectedDiff := ""
		if expectedDiff != diff {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedDiff, diff)
		}
	})

	t.Run("Readable Files", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockFile := file.NewMockFileAccess(ctrl)
		mockPatch := NewMockdiffMatchPatch(ctrl)
		defer ctrl.Finish()

		subject := &FileDiff{
			File:  mockFile,
			Patch: mockPatch,
		}

		diff := diffmatchpatch.Diff{

			Type: diffmatchpatch.DiffInsert,
			Text: "diff",
		}
		expectedDiff := "pretty diff"

		mockFile.EXPECT().Read("file1").Times(1).Return([]byte("content1"), nil)
		mockFile.EXPECT().Read("file2").Times(1).Return([]byte("content2"), nil)
		mockPatch.EXPECT().DiffMain("content1", "content2", true).Times(1).Return([]diffmatchpatch.Diff{
			diff,
		})
		mockPatch.EXPECT().DiffPrettyText([]diffmatchpatch.Diff{
			diff,
		}).Times(1).Return(expectedDiff)

		prettyDiff, err := subject.FindDiff("file1", "file2")

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		if expectedDiff != prettyDiff {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedDiff, prettyDiff)
		}

	})

	t.Run("First file error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockFile := file.NewMockFileAccess(ctrl)
		mockPatch := NewMockdiffMatchPatch(ctrl)
		defer ctrl.Finish()

		subject := &FileDiff{
			File:  mockFile,
			Patch: mockPatch,
		}
		expectedError := errors.New("test")

		mockFile.EXPECT().Read("file1").Times(1).Return(nil, expectedError)

		_, err := subject.FindDiff("file1", "file2")

		if !cmp.Equal(&expectedError, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedError, err)
		}

	})

	t.Run("Second file error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockFile := file.NewMockFileAccess(ctrl)
		mockPatch := NewMockdiffMatchPatch(ctrl)
		defer ctrl.Finish()

		subject := &FileDiff{
			File:  mockFile,
			Patch: mockPatch,
		}

		expectedError := errors.New("test")
		mockFile.EXPECT().Read("file1").Times(1).Return([]byte("content1"), nil)
		mockFile.EXPECT().Read("file2").Times(1).Return(nil, expectedError)

		_, err := subject.FindDiff("file1", "file2")

		if !cmp.Equal(&expectedError, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedError, err)
		}

	})

}

package composer

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/cjnosal/manifer/pkg/interpolator"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/scenario"
)

func TestResolve(t *testing.T) {

	cases := []struct {
		name              string
		libraryPaths      []string
		scenarioNames     []string
		passthrough       []string
		yamlError         error
		expectedLibraries []library.LoadedLibrary
		resolveError      error
		expectedSnippets  []string
		parseError        error
		expectedPlan      *scenario.Plan
		expectedError     error
	}{
		{
			name: "generate plan",
			libraryPaths: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			passthrough: []string{
				"extra",
			},
			expectedSnippets: []string{},
			expectedLibraries: []library.LoadedLibrary{
				{
					Path: "./lib/lib.yml",
				},
			},
			expectedPlan: &scenario.Plan{
				GlobalArgs: []string{
					"extra",
				},
				Snippets: []library.Snippet{
					{
						Path: "./lib/snippet.yml",
					},
				},
			},
		},
		{
			name: "append passthrough snippets",
			libraryPaths: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			passthrough: []string{
				"extra",
			},
			expectedSnippets: []string{"foo"},
			expectedLibraries: []library.LoadedLibrary{
				{
					Path: "./lib/lib.yml",
				},
			},
			expectedPlan: &scenario.Plan{
				GlobalArgs: []string{
					"extra",
				},
				Snippets: []library.Snippet{
					{
						Path: "./lib/snippet.yml",
					},
					{
						Path: "foo",
					},
				},
			},
		},
		{
			name: "passthrough snippets error",
			libraryPaths: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			passthrough: []string{
				"extra",
			},
			parseError: errors.New("test"),
			expectedLibraries: []library.LoadedLibrary{
				{
					Path: "./lib/lib.yml",
				},
			},
			expectedError: errors.New("test\n  while trying to resolve extra snippets"),
		},
		{
			name: "yaml error",
			libraryPaths: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			yamlError:     errors.New("test"),
			expectedError: errors.New("test\n  while trying to load libraries"),
		},
		{
			name: "resolve error",
			libraryPaths: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			expectedLibraries: []library.LoadedLibrary{
				{
					Path: "./lib/lib.yml",
				},
			},
			resolveError:  errors.New("test"),
			expectedError: errors.New("test\n  while trying to select scenarios"),
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLoader := library.NewMockLibraryLoader(ctrl)
			mockSelector := scenario.NewMockScenarioSelector(ctrl)
			mockInterpolator := interpolator.NewMockInterpolator(ctrl)

			mockLoader.EXPECT().Load(c.libraryPaths).Times(1).Return(c.expectedLibraries, c.yamlError)
			if c.yamlError == nil {
				mockSelector.EXPECT().SelectScenarios(c.scenarioNames, c.expectedLibraries).Times(1).Return(c.expectedPlan, c.resolveError)
				if c.resolveError == nil {
					mockInterpolator.EXPECT().ParseSnippetFlags(c.passthrough).Times(1).Return(c.expectedSnippets, c.parseError)
				}
			}

			subject := Resolver{
				Loader:          mockLoader,
				Selector:        mockSelector,
				SnippetResolver: mockInterpolator,
			}

			plan, err := subject.Resolve(c.libraryPaths, c.scenarioNames, c.passthrough)

			if !(c.expectedError == nil && err == nil) && !(c.expectedError != nil && err != nil && c.expectedError.Error() == err.Error()) {
				t.Errorf("Expected error:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedError, err)
			}

			if err == nil {
				if !reflect.DeepEqual(c.expectedPlan, plan) {
					t.Errorf("Expected plan:\n'''%v'''\nActual:\n'''%v'''\n", c.expectedPlan, plan)
				}
			}
		})
	}
}

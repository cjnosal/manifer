package composer

import (
	"errors"
	"testing"

	"github.com/cjnosal/manifer/test"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/cjnosal/manifer/pkg/interpolator"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/plan"
)

func TestResolve(t *testing.T) {

	cases := []struct {
		name              string
		libraryPaths      []string
		scenarioNames     []string
		passthrough       []string
		yamlError         error
		expectedLibraries *library.LoadedLibrary
		expectedNode      *library.ScenarioNode
		parseError        error
		expectedPlan      *plan.Plan
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
			passthrough:  []string{},
			expectedNode: nil,
			expectedLibraries: &library.LoadedLibrary{
				TopLibraries: []*library.Library{
					{
						Scenarios: []library.Scenario{
							{
								Name: "a scenario",
								Args: []string{},
								Snippets: []library.Snippet{
									{
										Path: "/foo.yml",
										Args: []string{"arg"},
									},
								},
							},
						},
					},
				},
			},
			expectedPlan: &plan.Plan{
				Global: plan.ArgSet{
					Tag:  "global",
					Args: []string{},
				},
				Steps: []*plan.Step{
					{
						Snippet: "/foo.yml",
						Args: []plan.ArgSet{
							{
								Tag:  "snippet",
								Args: []string{"arg"},
							},
							{
								Tag:  "a scenario",
								Args: []string{},
							},
						},
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
			expectedNode: &library.ScenarioNode{
				GlobalArgs: []string{"extra"},
				Snippets: []library.Snippet{
					{
						Path: "/bar.yml",
						Args: []string{"arg2"},
					},
				},
				Name:        "passthrough",
				Description: "args passed after --",
				LibraryPath: "<cli>",
				Type:        string(library.OpsFile),
				Args:        []string{},
				RefArgs:     []string{},
			},
			expectedLibraries: &library.LoadedLibrary{
				TopLibraries: []*library.Library{
					{
						Scenarios: []library.Scenario{
							{
								Name: "a scenario",
								Args: []string{},
								Snippets: []library.Snippet{
									{
										Path: "/foo.yml",
										Args: []string{"arg"},
									},
								},
							},
						},
					},
				},
			},
			expectedPlan: &plan.Plan{
				Global: plan.ArgSet{
					Tag:  "global",
					Args: []string{"extra"},
				},
				Steps: []*plan.Step{
					{
						Snippet: "/foo.yml",
						Args: []plan.ArgSet{
							{
								Tag:  "snippet",
								Args: []string{"arg"},
							},
							{
								Tag:  "a scenario",
								Args: []string{},
							},
						},
					},
					{
						Snippet: "/bar.yml",
						Args: []plan.ArgSet{
							{
								Tag:  "snippet",
								Args: []string{"arg2"},
							},
							{
								Tag:  "passthrough",
								Args: []string{},
							},
						},
					},
				},
			},
		},
		{
			name: "passthrough snippets error",
			libraryPaths: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{},
			passthrough: []string{
				"extra",
			},
			parseError:        errors.New("test"),
			expectedLibraries: &library.LoadedLibrary{},
			expectedError:     errors.New("test\n  while trying to parse passthrough args"),
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
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLoader := library.NewMockLibraryLoader(ctrl)
			mockInterpolator := interpolator.NewMockInterpolator(ctrl)

			mockLoader.EXPECT().Load(c.libraryPaths).Times(1).Return(c.expectedLibraries, c.yamlError)
			if c.yamlError == nil {
				mockInterpolator.EXPECT().ParsePassthroughFlags(c.passthrough).Times(1).Return(c.expectedNode, c.parseError)
			}

			subject := Resolver{
				Loader:          mockLoader,
				SnippetResolver: mockInterpolator,
			}

			plan, err := subject.Resolve(c.libraryPaths, c.scenarioNames, c.passthrough)

			if !cmp.Equal(&c.expectedError, &err, cmp.Comparer(test.EqualMessage)) {
				t.Errorf("Expected error:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedError, err)
			}

			if err == nil {
				if !cmp.Equal(c.expectedPlan, plan) {
					t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\nDiff:\n'''%s'''\n",
						c.expectedPlan, plan, cmp.Diff(c.expectedPlan, plan))
				}
			}
		})
	}
}

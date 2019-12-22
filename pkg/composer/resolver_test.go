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
	"github.com/cjnosal/manifer/pkg/processor"
)

func TestResolve(t *testing.T) {

	cases := []struct {
		name                    string
		libraryPaths            []string
		scenarioNames           []string
		passthrough             []string
		yamlError               error
		expectedLibraries       *library.LoadedLibrary
		expectedPassthroughNode *library.ScenarioNode
		expectedVarNode         *library.ScenarioNode
		parseError              error
		expectedPlan            *plan.Plan
		expectedError           error
		parseVarError           error
	}{
		{
			name: "generate plan",
			libraryPaths: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			passthrough:             []string{},
			expectedPassthroughNode: nil,
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
				"-oextra=o",
			},
			expectedPassthroughNode: &library.ScenarioNode{
				GlobalArgs: []string{},
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
				"-oextra=o",
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
		{
			name: "append passthrough variables",
			libraryPaths: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			passthrough: []string{
				"-vextra=e",
			},
			expectedVarNode: &library.ScenarioNode{
				GlobalArgs:  []string{"extra"},
				Snippets:    []library.Snippet{},
				Name:        "passthrough variables",
				Description: "vars passed after --",
				LibraryPath: "<cli>",
				Type:        "",
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
				},
			},
		},
		{
			name: "passthrough variables error",
			libraryPaths: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{},
			passthrough: []string{
				"-vextra=e",
			},
			parseVarError:     errors.New("test"),
			expectedLibraries: &library.LoadedLibrary{},
			expectedError:     errors.New("test\n  while trying to parse passthrough vars"),
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLoader := library.NewMockLibraryLoader(ctrl)
			mockProcessor := processor.NewMockProcessor(ctrl)
			mockInterpolator := interpolator.NewMockInterpolator(ctrl)

			mockLoader.EXPECT().Load(c.libraryPaths).Times(1).Return(c.expectedLibraries, c.yamlError)
			if c.yamlError == nil {
				mockProcessor.EXPECT().ParsePassthroughFlags(c.passthrough).Times(1).Return(c.expectedPassthroughNode, c.parseError)
			}
			if c.yamlError == nil && c.parseError == nil {
				mockInterpolator.EXPECT().ParsePassthroughVars(c.passthrough).Times(1).Return(c.expectedVarNode, c.parseVarError)
			}

			subject := Resolver{
				Loader:          mockLoader,
				SnippetResolver: mockProcessor,
				Interpolator:    mockInterpolator,
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

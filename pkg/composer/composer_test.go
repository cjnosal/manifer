package composer

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/interpolator"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/scenario"
)

type interpolation struct {
	in           string
	out          string
	snippet      string
	snippetArgs  []string
	templateArgs []string
	err          error
}

func TestCompose(t *testing.T) {

	planWithGlobals := &scenario.Plan{
		GlobalArgs: []string{"global arg"},
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

	cases := []struct {
		name           string
		template       string
		libraries      []string
		scenarioNames  []string
		plan           *scenario.Plan
		planError      error
		interpolations []interpolation
		passthrough    []string
		tmpError       error
		readError      error
		outputPath     string
		expectedOut    []byte
		expectedError  error
	}{
		{
			name:     "template and one scenario",
			template: "/tmp/base.yml",
			libraries: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			plan:        planWithoutGlobals,
			passthrough: []string{},
			interpolations: []interpolation{
				{
					in:      "/tmp/base.yml",
					out:     "/tmp/composed_0.yml",
					snippet: "/snippet",
					snippetArgs: []string{
						"snippet",
						"args",
					},
					templateArgs: []string{},
				},
			},
			outputPath:  "/tmp/composed_0.yml",
			expectedOut: []byte("composed"),
		},
		{
			name:     "post snippet args",
			template: "/tmp/base.yml",
			libraries: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			passthrough: []string{
				"cli arg",
			},
			plan: planWithGlobals,
			interpolations: []interpolation{
				{
					in:      "/tmp/base.yml",
					out:     "/tmp/composed_0.yml",
					snippet: "/snippet",
					snippetArgs: []string{
						"snippet",
						"args",
						"global arg",
						"cli arg",
					},
					templateArgs: []string{
						"global arg",
						"cli arg",
					},
				},
				{
					in:      "/tmp/composed_0.yml",
					out:     "/tmp/composed_final.yml",
					snippet: "",
					templateArgs: []string{
						"global arg",
						"cli arg",
					},
				},
			},
			outputPath:  "/tmp/composed_final.yml",
			expectedOut: []byte("composed"),
		},
		{
			name:     "scenario resolution error",
			template: "/tmp/base.yml",
			libraries: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			planError:     errors.New("test"),
			expectedError: errors.New("Unable to resolve scenarios: test"),
		},
		{
			name:     "tempdir error",
			template: "/tmp/base.yml",
			libraries: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			plan:          planWithoutGlobals,
			tmpError:      errors.New("test"),
			expectedError: errors.New("Unable to create temporary directory: test"),
		},
		{
			name:     "interpolation error",
			template: "/tmp/base.yml",
			libraries: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			plan:        planWithoutGlobals,
			passthrough: []string{},
			interpolations: []interpolation{
				{
					in:      "/tmp/base.yml",
					out:     "/tmp/composed_0.yml",
					snippet: "/snippet",
					snippetArgs: []string{
						"snippet",
						"args",
					},
					err:          errors.New("test"),
					templateArgs: []string{},
				},
			},
			expectedError: errors.New("Unable to apply snippet /snippet: test"),
		},
		{
			name:     "read error",
			template: "/tmp/base.yml",
			libraries: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			plan:        planWithoutGlobals,
			passthrough: []string{},
			interpolations: []interpolation{
				{
					in:      "/tmp/base.yml",
					out:     "/tmp/composed_0.yml",
					snippet: "/snippet",
					snippetArgs: []string{
						"snippet",
						"args",
					},
					templateArgs: []string{},
				},
			},
			outputPath:    "/tmp/composed_0.yml",
			readError:     errors.New("test"),
			expectedError: errors.New("Unable to read composed output: test"),
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockInterpolator := interpolator.NewMockInterpolator(ctrl)
			mockResolver := NewMockScenarioResolver(ctrl)
			mockFile := file.NewMockFileAccess(ctrl)
			subject := ComposerImpl{
				Resolver: mockResolver,
				File:     mockFile,
			}

			shouldLoad := true
			for _, i := range c.interpolations {
				if i.err != nil {
					shouldLoad = false
				}
				mockInterpolator.EXPECT().Interpolate(i.in, i.out, i.snippet, i.snippetArgs, i.templateArgs).Times(1).Return(i.err)
			}

			mockResolver.EXPECT().Resolve(c.libraries, c.scenarioNames).Times(1).Return(c.plan, c.planError)

			if c.planError == nil {
				mockFile.EXPECT().TempDir("", "manifer").Times(1).Return("/tmp", c.tmpError)
				if c.tmpError == nil {
					mockFile.EXPECT().RemoveAll("/tmp")
					if shouldLoad {
						mockFile.EXPECT().Read(c.outputPath).Times(1).Return(c.expectedOut, c.readError)
					}
				}
			}
			out, err := subject.Compose(mockInterpolator, c.template, c.libraries, c.scenarioNames, c.passthrough)

			if !reflect.DeepEqual(c.expectedError, err) {
				t.Errorf("Expected error:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedError, err)
			}

			if err == nil {
				if !reflect.DeepEqual(c.expectedOut, out) {
					t.Errorf("Expected output:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedOut, out)
				}
			}
		})
	}
}

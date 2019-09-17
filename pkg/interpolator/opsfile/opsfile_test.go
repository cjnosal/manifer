package opsfile

import (
	"errors"
	"os"
	"testing"

	"github.com/cppforlife/go-patch/patch"
	"github.com/golang/mock/gomock"

	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/yaml"
)

func litpnt(i interface{}) *interface{} {
	return &i
}

func newOpDefinition(t string, p string, i interface{}) patch.OpDefinition {
	return patch.OpDefinition{
		Type:  t,
		Path:  &p,
		Value: &i,
	}
}

func TestWrapper(t *testing.T) {
	cases := []struct {
		name             string
		in               string
		out              string
		snippet          string
		snippetArgs      []string
		scenarioArgs     []string
		intSnippetError  error
		intTemplateError error
		expectedError    error
	}{
		{
			name:    "with snippet",
			in:      "/template.yml",
			out:     "/out.yml",
			snippet: "/snippet.yml",
			snippetArgs: []string{
				"arg",
			},
			scenarioArgs: []string{
				"another",
			},
		},
		{
			name: "no snippet",
			in:   "/template.yml",
			out:  "/out.yml",
			scenarioArgs: []string{
				"another",
			},
		},
		{
			name:    "snippet error",
			in:      "/template.yml",
			out:     "/out.yml",
			snippet: "/snippet.yml",
			snippetArgs: []string{
				"arg",
			},
			intSnippetError: errors.New("test"),
			expectedError:   errors.New("test\n  while trying to interpolate snippet"),
		},
		{
			name:    "template error",
			in:      "/template.yml",
			out:     "/out.yml",
			snippet: "/snippet.yml",
			snippetArgs: []string{
				"arg",
			},
			scenarioArgs: []string{
				"another",
			},
			intTemplateError: errors.New("test"),
			expectedError:    errors.New("test\n  while trying to interpolate template"),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockInt := NewMockopFileInterpolator(ctrl)
			subject := interpolatorWrapper{
				interpolator: mockInt,
			}

			expectedSnippet := ""
			if c.snippet != "" {
				expectedSnippet = "/tmp/int_snippet.yml"
				mockInt.EXPECT().interpolate(c.snippet, expectedSnippet, "", "", append(c.snippetArgs, c.scenarioArgs...), false).Times(1).Return(c.intSnippetError)
			}
			if c.intSnippetError == nil {
				mockInt.EXPECT().interpolate(c.in, c.out, expectedSnippet, c.snippet, c.scenarioArgs, true).Times(1).Return(c.intTemplateError)
			}

			err := subject.Interpolate(c.in, c.out, c.snippet, c.snippetArgs, c.scenarioArgs)
			if !(c.expectedError == nil && err == nil) && !(c.expectedError != nil && err != nil && c.expectedError.Error() == err.Error()) {
				t.Errorf("Expected error:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedError, err)
			}
		})
	}
}

func TestInterpolate(t *testing.T) {

	validTemplate := "foo: bar\n\n"
	invalidTemplate := ":::not yaml"
	cases := []struct {
		name            string
		in              string
		inString        string
		out             string
		snippet         string
		originalSnippet string
		args            []string
		includeOps      bool

		opDefinitions []patch.OpDefinition

		readTemplateError  error
		parseArgsError     error
		readSnippetError   error
		parseSnippetError  error
		writeTemplateError error

		expectedError error
		expectedOut   string
	}{
		{
			name:       "vars only",
			in:         "../../../test/data/template_with_var.yml",
			inString:   validTemplate,
			out:        "/tmp/manifer_out.yml",
			includeOps: true,
			args: []string{
				"-v",
				"bar=bar",
			},
			expectedOut: "foo: bar\n",
		},
		{
			name:            "single op",
			in:              "../../../test/data/template.yml",
			inString:        validTemplate,
			out:             "/tmp/manifer_out.yml",
			snippet:         "interpolated_snippet.yml",
			originalSnippet: "opsfile.yml",
			opDefinitions: []patch.OpDefinition{
				newOpDefinition("replace", "/bizz?", "bazz"),
			},
			expectedOut: "bizz: bazz\nfoo: bar\n",
		},
		{
			name:            "multiple ops in file",
			in:              "../../../test/data/template.yml",
			inString:        validTemplate,
			out:             "/tmp/manifer_out.yml",
			snippet:         "interpolated_snippet.yml",
			originalSnippet: "opsfile.yml",
			opDefinitions: []patch.OpDefinition{
				newOpDefinition("replace", "/bizz?", "bazz"),
				newOpDefinition("replace", "/bazz?", "buzz"),
			},
			expectedOut: "bazz: buzz\nbizz: bazz\nfoo: bar\n",
		},
		{
			name:            "ignored passthrough ops",
			in:              "../../../test/data/template_with_var.yml",
			inString:        validTemplate,
			out:             "/tmp/manifer_out.yml",
			snippet:         "interpolated_snippet.yml",
			originalSnippet: "opsfile.yml",
			opDefinitions: []patch.OpDefinition{
				newOpDefinition("replace", "/bizz?", "bazz"),
			},
			includeOps: false,
			args: []string{
				"-v",
				"bar=bar",
				"-o",
				"../../../test/data/opsfile_with_vars.yml",
			},
			expectedOut: "bizz: bazz\nfoo: bar\n",
		},
		{
			name:       "include passthrough ops",
			in:         "../../../test/data/template.yml",
			inString:   validTemplate,
			out:        "/tmp/manifer_out.yml",
			includeOps: true,
			args: []string{
				"-o",
				"../../../test/data/opsfile.yml",
			},
			expectedOut: "bazz: buzz\nbizz: bazz\nfoo: bar\n",
		},
		{
			name:              "read template error",
			in:                "/doesnotexist",
			readTemplateError: errors.New("test"),
			expectedError:     errors.New("test\n  while trying to load /doesnotexist"),
		},
		{
			name: "parse args error",
			in:   "template.yml",
			args: []string{
				"--invalid",
			},
			expectedError: errors.New("unknown flag `invalid'\n  while trying to parse args"),
		},
		{
			name:             "read snippet error",
			in:               "../../../test/data/template.yml",
			readSnippetError: errors.New("test"),
			snippet:          "/missingsnippet",
			originalSnippet:  "/originalsnippet",
			expectedError:    errors.New("test\n  while trying to load ops file /originalsnippet"),
		},
		{
			name:              "parse snippet error",
			in:                "../../../test/data/template.yml",
			parseSnippetError: errors.New("test"),
			snippet:           "/missingsnippet",
			originalSnippet:   "/originalsnippet",
			expectedError:     errors.New("test\n  while trying to parse ops file /originalsnippet"),
		},
		{
			name:            "invalid snippet error",
			in:              "../../../test/data/template.yml",
			snippet:         "interpolated_snippet.yml",
			originalSnippet: "opsfile.yml",
			opDefinitions: []patch.OpDefinition{
				newOpDefinition("", "", ""),
			},
			expectedError: errors.New(`Unknown operation [0] with type '' within
{
  "Path": "",
  "Value": "<redacted>"
}
  while trying to create ops from definitions in opsfile.yml`),
		},
		{
			name:            "../../../test/data/template evalution error",
			in:              "invalid.yml",
			inString:        invalidTemplate,
			snippet:         "interpolated_snippet.yml",
			originalSnippet: "opsfile.yml",
			opDefinitions: []patch.OpDefinition{
				newOpDefinition("replace", "/bizz?", "bazz"),
			},
			expectedError: errors.New("Expected to find a map at path '/bizz?' but found 'string'\n  while trying to evaluate template invalid.yml with op 0 from opsfile.yml"),
		},
		{
			name:               "write error",
			in:                 "../../../test/data/template.yml",
			inString:           validTemplate,
			out:                "/tmp/manifer_out.yml",
			expectedOut:        "bizz: bazz\nfoo: bar\n",
			writeTemplateError: errors.New("test"),
			snippet:            "interpolated_snippet.yml",
			originalSnippet:    "opsfile.yml",
			opDefinitions: []patch.OpDefinition{
				newOpDefinition("replace", "/bizz?", "bazz"),
			},
			expectedError: errors.New("test\n  while trying to write interpolated file /tmp/manifer_out.yml"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockYaml := yaml.NewMockYamlAccess(ctrl)
			mockFile := file.NewMockFileAccess(ctrl)
			subject := ofInt{
				File: mockFile,
				Yaml: mockYaml,
			}

			mockFile.EXPECT().Read(c.in).Times(1).Return([]byte(c.inString), c.readTemplateError)

			if c.snippet != "" {
				mockFile.EXPECT().Read(c.snippet).Times(1).Return([]byte("bytes"), c.readSnippetError)
				if c.readSnippetError == nil {
					mockYaml.EXPECT().Unmarshal([]byte("bytes"), &[]patch.OpDefinition{}).Times(1).Return(c.parseSnippetError).Do(func(bytes []byte, o *[]patch.OpDefinition) {
						*o = c.opDefinitions
					})
				}
			}
			if c.out != "" {
				mockFile.EXPECT().Write(c.out, []byte(c.expectedOut), os.FileMode(0644)).Times(1).Return(c.writeTemplateError)
			}

			err := subject.interpolate(c.in, c.out, c.snippet, c.originalSnippet, c.args, c.includeOps)

			if !(c.expectedError == nil && err == nil) && !(c.expectedError != nil && err != nil && c.expectedError.Error() == err.Error()) {
				t.Errorf("Expected error:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedError, err)
			}
		})
	}
}

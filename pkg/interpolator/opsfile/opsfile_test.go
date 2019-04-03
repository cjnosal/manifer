package opsfile

import (
	"errors"
	"io/ioutil"
	"os"
	"reflect"
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

func loadTestFile(path string) []byte {
	bytes, _ := ioutil.ReadFile(path)
	return bytes
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
			expectedError:   errors.New("Unable to interpolate snippet: test"),
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
			expectedError:    errors.New("Unable to interpolate template: test"),
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
			if !reflect.DeepEqual(c.expectedError, err) {
				t.Errorf("Expected error:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedError, err)
			}
		})
	}
}

func TestInterpolate(t *testing.T) {

	cases := []struct {
		name            string
		in              string
		out             string
		snippet         string
		originalSnippet string
		args            []string
		includeOps      bool

		opDefinitions []patch.OpDefinition

		readTemplateError  error
		parseArgsError     error
		readSnippetError   error
		writeTemplateError error

		expectedError error
		expectedOut   string
	}{
		{
			name:       "vars only",
			in:         "../../../test/data/template_with_var.yml",
			out:        "/tmp/manifer_out.yml",
			includeOps: true,
			args: []string{
				"-v",
				"bar=bar",
			},
			expectedOut: "../../../test/data/template.yml",
		},
		{
			name:            "single op",
			in:              "../../../test/data/template.yml",
			out:             "/tmp/manifer_out.yml",
			snippet:         "../../../test/data/interpolated_snippet.yml",
			originalSnippet: "../../../test/data/opsfile.yml",
			opDefinitions: []patch.OpDefinition{
				newOpDefinition("replace", "/bizz?", "bazz"),
			},
			expectedOut: "../../../test/data/interpolated_bizz.yml",
		},
		{
			name:            "multiple ops in file",
			in:              "../../../test/data/template.yml",
			out:             "/tmp/manifer_out.yml",
			snippet:         "../../../test/data/interpolated_snippet.yml",
			originalSnippet: "../../../test/data/opsfile.yml",
			opDefinitions: []patch.OpDefinition{
				newOpDefinition("replace", "/bizz?", "bazz"),
				newOpDefinition("replace", "/bazz?", "buzz"),
			},
			expectedOut: "../../../test/data/interpolated_bizz_bazz.yml",
		},
		{
			name:            "ignored passthrough ops",
			in:              "../../../test/data/template_with_var.yml",
			out:             "/tmp/manifer_out.yml",
			snippet:         "../../../test/data/interpolated_snippet.yml",
			originalSnippet: "../../../test/data/opsfile.yml",
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
			expectedOut: "../../../test/data/interpolated_bizz.yml",
		},
		{
			name:       "include passthrough ops",
			in:         "../../../test/data/template.yml",
			out:        "/tmp/manifer_out.yml",
			includeOps: true,
			args: []string{
				"-o",
				"../../../test/data/opsfile.yml",
			},
			expectedOut: "../../../test/data/interpolated_bizz_bazz.yml",
		},
		{
			name:              "read template error",
			in:                "/doesnotexist",
			readTemplateError: errors.New("test"),
			expectedError:     errors.New("Unable to load /doesnotexist: test"),
		},
		{
			name: "parse args error",
			in:   "../../../test/data/template.yml",
			args: []string{
				"--invalid",
			},
			expectedError: errors.New("Unable to parse args: unknown flag `invalid'"),
		},
		{
			name:             "read snippet error",
			in:               "../../../test/data/template.yml",
			readSnippetError: errors.New("test"),
			snippet:          "/missingsnippet",
			originalSnippet:  "/originalsnippet",
			expectedError:    errors.New("Unable to load ops file /originalsnippet: test"),
		},
		{
			name:            "parse snippet error",
			in:              "../../../test/data/template.yml",
			snippet:         "../../../test/data/interpolated_snippet.yml",
			originalSnippet: "../../../test/data/opsfile.yml",
			opDefinitions: []patch.OpDefinition{
				newOpDefinition("", "", ""),
			},
			expectedError: errors.New(`Unable to create ops from definitions in ../../../test/data/opsfile.yml: Unknown operation [0] with type '' within
{
  "Path": "",
  "Value": "<redacted>"
}`),
		},
		{
			name:            "template evalution error",
			in:              "../../../test/data/invalid.yml",
			snippet:         "../../../test/data/interpolated_snippet.yml",
			originalSnippet: "../../../test/data/opsfile.yml",
			opDefinitions: []patch.OpDefinition{
				newOpDefinition("replace", "/bizz?", "bazz"),
			},
			expectedError: errors.New("Unable to evaluate template ../../../test/data/invalid.yml with op 0 from ../../../test/data/opsfile.yml: Expected to find a map at path '/bizz?' but found 'string'"),
		},
		{
			name:               "write error",
			in:                 "../../../test/data/template.yml",
			out:                "/tmp/manifer_out.yml",
			expectedOut:        "../../../test/data/interpolated_bizz.yml",
			writeTemplateError: errors.New("test"),
			snippet:            "../../../test/data/interpolated_snippet.yml",
			originalSnippet:    "../../../test/data/opsfile.yml",
			opDefinitions: []patch.OpDefinition{
				newOpDefinition("replace", "/bizz?", "bazz"),
			},
			expectedError: errors.New("Unable to write interpolated file /tmp/manifer_out.yml: test"),
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

			mockFile.EXPECT().Read(c.in).Times(1).Return(loadTestFile(c.in), c.readTemplateError)

			if c.snippet != "" {
				mockYaml.EXPECT().Load(c.snippet, &[]patch.OpDefinition{}).Times(1).Return(c.readSnippetError).Do(func(path string, o *[]patch.OpDefinition) {
					*o = c.opDefinitions
				})
			}
			if c.out != "" {
				mockFile.EXPECT().Write(c.out, loadTestFile(c.expectedOut), os.FileMode(0644)).Times(1).Return(c.writeTemplateError)
			}

			err := subject.interpolate(c.in, c.out, c.snippet, c.originalSnippet, c.args, c.includeOps)

			if !reflect.DeepEqual(c.expectedError, err) {
				t.Errorf("Expected error:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedError, err)
			}
		})
	}
}

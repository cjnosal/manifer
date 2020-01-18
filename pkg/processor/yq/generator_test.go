package yq

import (
	"errors"
	"github.com/cjnosal/manifer/v2/pkg/processor"
	"github.com/cjnosal/manifer/v2/pkg/yaml"
	"github.com/cjnosal/manifer/v2/test"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	y "gopkg.in/yaml.v3"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	t.Run("GenerateSnippets", func(t *testing.T) {
		subject := processor.NewSnippetGenerator(
			&yaml.Yaml{},
			&yqPathBuilder{},
		)

		schema := &yaml.SchemaNode{
			Kind: y.SequenceNode,
			Contents: map[string]*yaml.SchemaNode{
				"!!element": &yaml.SchemaNode{
					Kind: y.MappingNode,
					Contents: map[string]*yaml.SchemaNode{
						"foo": &yaml.SchemaNode{
							Kind:     y.ScalarNode,
							Contents: map[string]*yaml.SchemaNode{},
						},
						"bar": &yaml.SchemaNode{
							Kind:     y.ScalarNode,
							Contents: map[string]*yaml.SchemaNode{},
						},
						"a": &yaml.SchemaNode{
							Kind: y.MappingNode,
							Contents: map[string]*yaml.SchemaNode{
								"b": &yaml.SchemaNode{
									Kind:     y.ScalarNode,
									Contents: map[string]*yaml.SchemaNode{},
								},
								"c": &yaml.SchemaNode{
									Kind: y.SequenceNode,
									Contents: map[string]*yaml.SchemaNode{
										"!!element": &yaml.SchemaNode{
											Kind:     y.ScalarNode,
											Contents: map[string]*yaml.SchemaNode{},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		serializedOps, err := subject.GenerateSnippets(schema)
		if err != nil {
			t.Errorf("Unexpected error %w", err)
		}
		builder := strings.Builder{}

		for _, op := range serializedOps {
			builder.WriteString(op.Tag)
			builder.WriteString("\n")
			builder.WriteString(string(op.Bytes))
			builder.WriteString("\n")
		}
		result := builder.String()

		expected := `./add_root.yml
'[+]':
    bar: ((root_bar))
    foo: ((root_foo))

./set_root.yml
'[((root_index))]':
    bar: ((root_bar))
    foo: ((root_foo))

./root/set_a.yml
'[((root_index))].a':
    b: ((a_b))

./root/a/add_c.yml
'[((root_index))].a.c[+]': ((a_c))

`

		if !cmp.Equal(expected, result) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\nDiff:\n'''%v'''\n",
				expected, result, cmp.Diff(expected, result))
		}
	})

	t.Run("marshaling error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockYaml := yaml.NewMockYamlAccess(ctrl)
		mockYaml.EXPECT().Marshal(gomock.Any()).Times(1).Return(nil, errors.New("oops"))
		subject := processor.NewSnippetGenerator(
			mockYaml,
			&yqPathBuilder{},
		)

		schema := &yaml.SchemaNode{
			Kind: y.MappingNode,
			Contents: map[string]*yaml.SchemaNode{
				"foo": &yaml.SchemaNode{
					Kind:     y.ScalarNode,
					Contents: map[string]*yaml.SchemaNode{},
				},
			},
		}

		_, err := subject.GenerateSnippets(schema)

		expectedError := errors.New("oops\n  marshaling snippet\n  generating snippet for scalar foo at foo")
		if !cmp.Equal(&expectedError, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected error:\n'''%s'''\nActual:\n'''%s'''\n", expectedError, err)
		}
	})
}

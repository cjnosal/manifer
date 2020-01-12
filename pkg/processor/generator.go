package processor

import (
	"fmt"
	"github.com/cjnosal/manifer/v2/pkg/file"
	"github.com/cjnosal/manifer/v2/pkg/yaml"
	y "gopkg.in/yaml.v3"
	"os"
	"strings"
)

func NewSnippetGenerator(y yaml.YamlAccess, builder PathBuilder) SnippetGenerator {
	return &snippetGenerator{
		yaml:        y,
		pathBuilder: builder,
	}
}

type snippetGenerator struct {
	yaml        yaml.YamlAccess
	pathBuilder PathBuilder
}

func (i *snippetGenerator) GenerateSnippets(schema *yaml.SchemaNode) ([]*file.TaggedBytes, error) {
	snippets := []*file.TaggedBytes{}
	if schema.Kind == y.MappingNode {
		for k, v := range schema.Contents {
			if v.Kind == y.ScalarNode {
				path := fmt.Sprintf("%s%s%s", i.pathBuilder.Root(), k, i.pathBuilder.Safe())
				file := fmt.Sprintf(".%sset_%s.yml", string(os.PathSeparator), k)
				snippet, err := i.generate(path, file, "", k, v)
				if err != nil {
					return nil, err
				}
				snippets = append(snippets, snippet)
			} else {
				p := fmt.Sprintf("%s%s", i.pathBuilder.Root(), k)
				d := "."
				o, err := i.visit(p, d, "", k, v)
				if err != nil {
					return nil, err
				}
				snippets = append(snippets, o...)
			}
		}
	} else { // sequence
		element := schema.Contents["!!element"]
		appendpath := i.pathBuilder.Append()
		elementName := "root"
		file := fmt.Sprintf(".%sadd_%s.yml", string(os.PathSeparator), elementName)
		appendsnippet, err := i.generate(appendpath, file, "", elementName, element)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, appendsnippet)

		if element != nil {
			indexvar := fmt.Sprintf("((%s_index))", elementName)
			editpath := i.pathBuilder.Index(indexvar)
			editsnippets, err := i.visit(editpath, ".", elementName, elementName, element)
			if err != nil {
				return nil, err
			}
			snippets = append(snippets, editsnippets...)
		}
	}
	return snippets, nil
}

func (i *snippetGenerator) visit(path string, dir string, parent string, name string, node *yaml.SchemaNode) ([]*file.TaggedBytes, error) {
	snippets := []*file.TaggedBytes{}

	if node.Kind == y.SequenceNode {
		element := node.Contents["!!element"]
		appendpath := fmt.Sprintf("%s%s%s", path, i.pathBuilder.Safe(), i.pathBuilder.Append())
		elementName := naiveSingular(name)
		file := fmt.Sprintf("%s%sadd_%s.yml", dir, string(os.PathSeparator), elementName)
		appendsnippet, err := i.generate(appendpath, file, parent, elementName, element)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, appendsnippet)

		if element != nil {
			indexvar := fmt.Sprintf("((%s_index))", elementName)
			editpath := fmt.Sprintf("%s%s", path, i.pathBuilder.Index(indexvar))
			editsnippets, err := i.visit(editpath, dir, elementName, elementName, element)
			if err != nil {
				return nil, err
			}
			snippets = append(snippets, editsnippets...)
		}
	} else if node.Kind == y.MappingNode {
		optionalPath := path
		if !strings.HasSuffix(path, "))") {
			optionalPath = fmt.Sprintf("%s%s", path, i.pathBuilder.Safe())
		}
		file := fmt.Sprintf("%s%sset_%s.yml", dir, string(os.PathSeparator), name)
		snippet, err := i.generate(optionalPath, file, parent, name, node)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, snippet)

		// recurse sequence/mapping nodes
		for k, v := range node.Contents {
			p := fmt.Sprintf("%s%s%s", path, i.pathBuilder.Delimiter(), k)
			d := fmt.Sprintf("%s%s%s", dir, string(os.PathSeparator), name)
			o, err := i.visit(p, d, name, k, v)
			if err != nil {
				return nil, err
			}
			snippets = append(snippets, o...)
		}
	}

	return snippets, nil
}

func (i *snippetGenerator) generate(path string, filePath string, parent string, name string, node *yaml.SchemaNode) (*file.TaggedBytes, error) {
	var value interface{}

	// generate snippetsfile that inserts all scalars into a mapping
	// snippetsfile to insert key=value for empty mapping node
	if node == nil || node.Kind == y.ScalarNode {
		var variable string
		if parent != "" {
			variable = fmt.Sprintf("((%s_%s))", parent, name)
		} else {
			variable = fmt.Sprintf("((%s))", name)
		}
		value = &variable
	} else {
		values := &map[string]string{}
		for k, v := range node.Contents {
			if v.Kind == y.ScalarNode {
				variable := fmt.Sprintf("((%s_%s))", name, k)
				(*values)[k] = variable
			}
		}
		value = values
	}
	snippet := i.pathBuilder.Marshal(path, value)
	bytes, err := i.yaml.Marshal(snippet)
	if err != nil {
		return nil, fmt.Errorf("%w\n  marshaling %s", err, path)
	}
	return &file.TaggedBytes{Bytes: bytes, Tag: filePath}, nil
}

func naiveSingular(name string) string {
	if strings.HasSuffix(name, "s") {
		return name[:len(name)-1]
	}
	return name
}

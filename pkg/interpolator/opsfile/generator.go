package opsfile

import (
	"fmt"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/yaml"
	"github.com/cppforlife/go-patch/patch"
	y "gopkg.in/yaml.v3"
	"strings"
)

type opFileGenerator struct {
	yaml yaml.YamlAccess
}

func (i *opFileGenerator) generateSnippets(schema *yaml.SchemaNode) ([]*file.TaggedBytes, error) {
	ops := []*file.TaggedBytes{}
	if schema.Kind == y.MappingNode {
		for k, v := range schema.Contents {
			if v.Kind == y.ScalarNode {
				opPath := fmt.Sprintf("/%s?", k)
				file := fmt.Sprintf("./set_%s.yml", k)
				op, err := i.generate(opPath, file, "", k, v)
				if err != nil {
					return nil, err
				}
				ops = append(ops, op)
			} else {
				p := fmt.Sprintf("/%s", k)
				d := "."
				o, err := i.visit(p, d, "", k, v)
				if err != nil {
					return nil, err
				}
				ops = append(ops, o...)
			}
		}
	} else { // sequence
		element := schema.Contents["!!element"]
		appendpath := "/-"
		elementName := "root"
		file := fmt.Sprintf("./add_%s.yml", elementName)
		appendop, err := i.generate(appendpath, file, "", elementName, element)
		if err != nil {
			return nil, err
		}
		ops = append(ops, appendop)

		if element != nil {
			indexvar := fmt.Sprintf("((%s_index))", elementName)
			editpath := fmt.Sprintf("/%s", indexvar)
			editops, err := i.visit(editpath, ".", elementName, elementName, element)
			if err != nil {
				return nil, err
			}
			ops = append(ops, editops...)
		}
	}
	return ops, nil
}

func (i *opFileGenerator) visit(opPath string, dir string, parent string, name string, node *yaml.SchemaNode) ([]*file.TaggedBytes, error) {
	ops := []*file.TaggedBytes{}

	if node.Kind == y.SequenceNode {
		element := node.Contents["!!element"]
		appendpath := fmt.Sprintf("%s?/%s", opPath, "-")
		elementName := naiveSingular(name)
		file := fmt.Sprintf("%s/add_%s.yml", dir, elementName)
		appendop, err := i.generate(appendpath, file, parent, elementName, element)
		if err != nil {
			return nil, err
		}
		ops = append(ops, appendop)

		if element != nil {
			indexvar := fmt.Sprintf("((%s_index))", elementName)
			editpath := fmt.Sprintf("%s/%s", opPath, indexvar)
			editops, err := i.visit(editpath, dir, elementName, elementName, element)
			if err != nil {
				return nil, err
			}
			ops = append(ops, editops...)
		}
	} else if node.Kind == y.MappingNode {
		optionalPath := opPath
		if !strings.HasSuffix(opPath, "))") {
			optionalPath = fmt.Sprintf("%s?", opPath)
		}
		file := fmt.Sprintf("%s/set_%s.yml", dir, name)
		op, err := i.generate(optionalPath, file, parent, name, node)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)

		// recurse sequence/mapping nodes
		for k, v := range node.Contents {
			p := fmt.Sprintf("%s/%s", opPath, k)
			d := fmt.Sprintf("%s/%s", dir, name)
			o, err := i.visit(p, d, name, k, v)
			if err != nil {
				return nil, err
			}
			ops = append(ops, o...)
		}
	}

	return ops, nil
}

func (i *opFileGenerator) generate(opPath string, filePath string, parent string, name string, node *yaml.SchemaNode) (*file.TaggedBytes, error) {
	var value interface{}

	// generate opsfile that inserts all scalars into a mapping
	// opsfile to insert key=value for empty mapping node
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
	opdef := patch.OpDefinition{
		Path:  &opPath,
		Type:  "replace",
		Value: &value,
	}
	bytes, err := i.yaml.Marshal([]patch.OpDefinition{opdef})
	if err != nil {
		return nil, fmt.Errorf("%w\n  marshaling %s", err, opPath)
	}
	return &file.TaggedBytes{Bytes: bytes, Tag: filePath}, nil
}

func naiveSingular(name string) string {
	if strings.HasSuffix(name, "s") {
		return name[:len(name)-1]
	}
	return name
}

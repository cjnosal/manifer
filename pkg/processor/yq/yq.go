package yq

import (
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/processor"
	"github.com/cjnosal/manifer/pkg/yaml"

	"github.com/jessevdk/go-flags"
	y2 "github.com/mikefarah/yaml/v2" // for MapSlice - replaced by Node in yaml v3
	"github.com/mikefarah/yq/v2/pkg/yqlib"
	logging "gopkg.in/op/go-logging.v1"

	"bytes"
	"fmt"
	"reflect"
)

type yqInt struct {
	yaml yaml.YamlAccess
	file file.FileAccess
}

type writeCommand struct {
	path  string
	value interface{}
}

func NewYqProcessor(y yaml.YamlAccess, f file.FileAccess) processor.Processor {
	return &yqInt{
		yaml: y,
		file: f,
	}
}

func (y *yqInt) ValidateSnippet(path string) (bool, error) {
	content, err := y.file.Read(path)
	if err != nil {
		return false, fmt.Errorf("%w\n  while validating yq script %s", err, path)
	}
	var rawCommands y2.MapSlice
	err = y2.Unmarshal(content, &rawCommands)
	if err != nil {
		return false, nil
	}
	if len(rawCommands) == 0 {
		return false, nil
	}
	for _, command := range rawCommands {
		strKey, ok := command.Key.(string)
		if !ok || len(strKey) == 0 {
			return false, nil
		}
	}
	return true, nil
}

type scriptFlags struct {
	// flag string copied from yq
	ScriptPaths []string `long:"script" short:"s" value-name:"PATH" description:"yaml write script for updating yaml"`
}

func (i *yqInt) ParsePassthroughFlags(args []string) (*library.ScenarioNode, []string, error) {
	var node *library.ScenarioNode
	scriptFlags := scriptFlags{}
	remainder, err := flags.NewParser(&scriptFlags, flags.IgnoreUnknown).ParseArgs(args)
	if err != nil {
		return nil, nil, fmt.Errorf("%w\n  while trying to parse yq args", err)
	}
	if len(scriptFlags.ScriptPaths) > 0 {
		snippets := []library.Snippet{}
		for _, o := range scriptFlags.ScriptPaths {
			snippets = append(snippets, library.Snippet{
				Path: o,
				Processor: library.Processor{
					Type: library.Yq,
					Options: map[string]interface{}{
						"command": "write",
					},
				},
			})
		}
		node = &library.ScenarioNode{
			Name:        "passthrough yq",
			Description: "args passed after --",
			LibraryPath: "<cli>",
			Snippets:    snippets,
		}
	}
	return node, remainder, nil
}

func (y *yqInt) ProcessTemplate(templateBytes *file.TaggedBytes, snippetBytes *file.TaggedBytes, options map[string]interface{}) ([]byte, error) {
	yql := yqlib.NewYqLib(logging.MustGetLogger("yq"))
	logging.SetLevel(logging.ERROR, "yq")
	y2.DefaultMapType = reflect.TypeOf(y2.MapSlice{})

	command, commandFound := getOptionString(options, "command")
	if !commandFound {
		command = "write"
	}

	var template interface{}
	err := y2.Unmarshal(templateBytes.Bytes, &template)
	if err != nil {
		return nil, fmt.Errorf("%w\n  unmarshaling template", err)
	}

	if command == "write" {
		writeCommands := []writeCommand{}

		var rawCommands y2.MapSlice
		err := y2.Unmarshal(snippetBytes.Bytes, &rawCommands)
		if err != nil {
			return nil, fmt.Errorf("%w\n  unmarshaling snippet", err)
		}
		for _, c := range rawCommands {
			writeCommands = append(writeCommands, writeCommand{
				path:  c.Key.(string),
				value: c.Value,
			})
		}

		for _, wc := range writeCommands {
			template = yql.WritePath(template, wc.path, wc.value)
		}

	} else if command == "read" {
		path, pathFound := getOptionString(options, "path")
		if !pathFound {
			return nil, fmt.Errorf("read path missing")
		}
		template, err = yql.ReadPath(template, path)
		if err != nil {
			return nil, fmt.Errorf("%w\n  reading path %s", err, path)
		}

	} else if command == "delete" {
		path, pathFound := getOptionString(options, "path")
		if !pathFound {
			return nil, fmt.Errorf("delete path missing")
		}
		template, err = yql.DeletePath(template, path)
		if err != nil {
			return nil, fmt.Errorf("%w\n  deleting path %s", err, path)
		}

	} else if command == "merge" {
		y2.DefaultMapType = reflect.TypeOf(map[interface{}]interface{}{})
		err = y2.Unmarshal(templateBytes.Bytes, &template)
		if err != nil {
			return nil, fmt.Errorf("%w\n  re-unmarshaling template", err)
		}

		overwriteFlag := getOptionBool(options, "overwrite")
		appendFlag := getOptionBool(options, "append")

		// lib assumes top level map?
		templateWrapper := map[interface{}]interface{}{}
		templateWrapper["root"] = template

		var parsedSnippet interface{}
		err := y2.Unmarshal(snippetBytes.Bytes, &parsedSnippet)
		if err != nil {
			return nil, fmt.Errorf("%w\n  unmarshaling snippet", err)
		}

		snippetWrapper := map[interface{}]interface{}{}
		snippetWrapper["root"] = parsedSnippet

		err = yql.Merge(&templateWrapper, snippetWrapper, overwriteFlag, appendFlag)
		if err != nil {
			return nil, fmt.Errorf("%w\n  merging snippet %s", err, snippetBytes.Tag)
		}
		template = templateWrapper["root"]

	} else if command == "prefix" {
		prefix, prefixFound := getOptionString(options, "prefix")
		if !prefixFound {
			return nil, fmt.Errorf("prefix missing")
		}

		template = yql.PrefixPath(template, prefix)

	} else {
		return nil, fmt.Errorf("Unsupported yq command %s", command)
	}

	bytes, err := y2.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("%w\n  marshaling template", err)
	}

	return bytes, nil
}

type byteWriter struct {
	buffer bytes.Buffer
}

func (s *byteWriter) Write(b []byte) (int, error) {
	s.buffer.Write(b)
	return len(b), nil
}

func (s *byteWriter) Bytes() []byte {
	return s.buffer.Bytes()
}

func getOptionString(options map[string]interface{}, opt string) (string, bool) {
	var optString string
	if options != nil {
		optInterface := options[opt]
		if optInterface != nil {
			optString = optInterface.(string)
			return optString, true
		}
	}

	return "", false
}

func getOptionBool(options map[string]interface{}, opt string) bool {
	if options != nil {
		optInterface := options[opt]
		if optInterface != nil {
			return optInterface.(bool)
		}
	}

	return false
}

package plan

import (
	"fmt"
	"github.com/cjnosal/manifer/pkg/diff"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/interpolator"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/processor"
	"io"
)

type Executor interface {
	Execute(showPlan bool, showDiff bool, template *file.TaggedBytes, snippet *file.TaggedBytes, snippetProcessor *library.Processor, snippetVars library.InterpolatorParams, globals library.InterpolatorParams) ([]byte, error)
}

type InterpolationExecutor struct {
	Processor    processor.Processor
	Interpolator interpolator.Interpolator
	Diff         diff.Diff
	Output       io.Writer
	File         file.FileAccess
}

func (i *InterpolationExecutor) Execute(showPlan bool, showDiff bool, template *file.TaggedBytes, snippet *file.TaggedBytes, snippetProcessor *library.Processor, snippetVars library.InterpolatorParams, globals library.InterpolatorParams) ([]byte, error) {
	var snippetPath string
	if snippet != nil {
		snippetPath = snippet.Tag
	}
	if showPlan {
		var relpath string
		var err error
		if snippet != nil {
			relpath, err = i.File.ResolveRelativeFromWD(snippetPath)
			if err != nil {
				return nil, fmt.Errorf("%w\n  while resolving relative snippet path %s", err, snippetPath)
			}
		}
		out := fmt.Sprintf("\nSnippet %s; Params %+v; Processor %+v\n", relpath, snippetVars.Merge(globals), snippetProcessor)
		i.Output.Write([]byte(out))
	}
	bytes, err := i.processSnippet(template, snippet, snippetProcessor, snippetVars, globals)
	if err != nil {
		return nil, err
	}
	if showDiff {
		i.Output.Write([]byte("Diff:\n"))
		diff := i.Diff.StringDiff(string(template.Bytes), string(bytes))
		i.Output.Write([]byte(diff))
	}
	return bytes, nil
}

func (i *InterpolationExecutor) processSnippet(template *file.TaggedBytes, snippet *file.TaggedBytes, snippetProcessor *library.Processor, snippetVars library.InterpolatorParams, globals library.InterpolatorParams) ([]byte, error) {

	var processedTemplate *file.TaggedBytes
	var intSnippet *file.TaggedBytes
	if snippet != nil {
		snippetBytes, err := i.Interpolator.Interpolate(snippet, snippetVars.Merge(globals))
		if err != nil {
			return nil, fmt.Errorf("%w\n  while trying to interpolate snippet", err)
		}
		intSnippet = &file.TaggedBytes{
			Bytes: snippetBytes,
			Tag:   snippet.Tag,
		}
	}

	var processorOptions map[string]interface{}
	if snippetProcessor != nil {
		processorOptions = snippetProcessor.Options
	}

	processedBytes, err := i.Processor.ProcessTemplate(template, intSnippet, processorOptions)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to process template", err)
	}
	processedTemplate = &file.TaggedBytes{
		Bytes: processedBytes,
		Tag:   template.Tag,
	}

	intTemplateBytes, err := i.Interpolator.Interpolate(processedTemplate, globals)
	if err != nil {
		return nil, fmt.Errorf("%w\n  while trying to interpolate template", err)
	}

	return intTemplateBytes, nil
}

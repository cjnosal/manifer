package interpolator

type Interpolator interface {
	Interpolate(inPath string, outPath string, snippetPath string, snippetArgs []string, scenarioArgs []string) error
}

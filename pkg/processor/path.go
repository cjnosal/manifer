package processor

type PathBuilder interface {
	Root() string
	Append() string
	Index(index string) string
	Delimiter() string
	Safe() string
	Marshal(path string, value interface{}) interface{}
}

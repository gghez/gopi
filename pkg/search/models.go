package search

type Searcher[T any] interface {
	Source() string
	Search(query string) ([]T, error)
}

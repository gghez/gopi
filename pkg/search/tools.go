package search

import (
	"github.com/gghez/gopi/internal/dump"
)

func SearchAndDump[T any](searcher Searcher[T], query string, dumpsDirPath string) error {
	if results, err := searcher.Search(query); err != nil {
		return err
	} else if err := dump.DumpYAML(results, dumpsDirPath, searcher.Source()); err != nil {
		return err
	}
	return nil
}

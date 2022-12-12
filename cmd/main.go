package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/gghez/gopi/pkg/search"
	"github.com/gghez/gopi/pkg/sources"
	"github.com/rs/zerolog/log"
)

const dumpsDir = ".local/dumps"

func main() {
	query := flag.String("q", "", "query terms")
	flag.Parse()

	runDumpsDir := fmt.Sprintf("%s/%s", dumpsDir, time.Now().Format(time.RFC3339))

	ukRegistrySearchEngine := sources.NewUKRegistrySearchEngine()
	if err := search.SearchAndDump(ukRegistrySearchEngine, *query, runDumpsDir); err != nil {
		log.Error().Err(err).Msg("failed to search and dump uk registry")
	}

}

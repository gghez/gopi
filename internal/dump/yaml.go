package dump

import (
	"fmt"

	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Dumps results as YAML
func DumpYAML(results interface{}, dumpsDir string, source string) error {
	data, err := yaml.Marshal(results)
	if err != nil {
		log.Fatal().Err(err)
	}

	if err := os.MkdirAll(dumpsDir, 0700); err != nil {
		return err
	}

	filepath := fmt.Sprintf("%s/%s.yml", dumpsDir, source)
	if err := os.WriteFile(filepath, data, 0600); err != nil {
		return err
	}

	log.Info().Str("source", source).Str("path", filepath).Msg("dump successful")

	return nil
}

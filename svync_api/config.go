package svync_api

import (
	"log"
	"os"

	cli "github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

// Read the configuration file, cast it to its struct and validate
func ReadConfig(Cctx *cli.Context) Config {
	logger := log.New(os.Stderr, "", 0)
	configFile, err := os.ReadFile(Cctx.String("config"))
	if err != nil {
		logger.Fatalf("Failed to open the config file: %v", err)
	}

	var config Config

	if err := yaml.Unmarshal(configFile, &config); err != nil {
		logger.Fatalf("Failed to parse the config file: %v", err)
	}

	return config
}

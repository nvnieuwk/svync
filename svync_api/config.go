package svync_api

import (
	"log"
	"os"

	cli "github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

// Read the configuration file, cast it to its struct and validate
func ReadConfig(Cctx *cli.Context) *Config {
	logger := log.New(os.Stderr, "", 0)
	configFile, err := os.ReadFile(Cctx.String("config"))
	if err != nil {
		logger.Fatalf("Failed to open the config file: %v", err)
	}

	var config Config

	if err := yaml.Unmarshal(configFile, &config); err != nil {
		logger.Fatalf("Failed to parse the config file: %v", err)
	}

	config.defineMissing()
	return &config
}

// Define all missing mandatory fields
func (config *Config) defineMissing() {
	// Info fields
	if _, ok := config.Info["SVTYPE"]; !ok {
		config.Info["SVTYPE"] = ConfigInput{
			Value:       "$INFO/SVTYPE",
			Number:      "1",
			Type:        "String",
			Description: "Type of structural variant",
		}
	}
	if _, ok := config.Info["SVLEN"]; !ok {
		config.Info["SVLEN"] = ConfigInput{
			Value:       "$INFO/SVLEN",
			Number:      "1",
			Type:        "String",
			Description: "Type of structural variant",
		}
	}
	if _, ok := config.Info["END"]; !ok {
		config.Info["END"] = ConfigInput{
			Value:       "$INFO/END",
			Number:      "1",
			Type:        "Integer",
			Description: "End position of the variant described in this record",
			Alts: map[string]string{
				"BND": "",
			},
		}
	}
	if _, ok := config.Info["CHR2"]; !ok {
		config.Info["CHR2"] = ConfigInput{
			Value:       "",
			Number:      "1",
			Type:        "Integer",
			Description: "End position of the variant described in this record",
			Alts: map[string]string{
				"TRA": "$INFO/CHR2",
			},
		}
	}
	if _, ok := config.Info["IMPRECISE"]; !ok {
		config.Info["IMPRECISE"] = ConfigInput{
			Value:       "$INFO/IMPRECISE",
			Number:      "0",
			Type:        "Flag",
			Description: "Imprecise structural variation",
		}
	}

	// Format fields
	if _, ok := config.Format["GT"]; !ok {
		config.Format["GT"] = ConfigInput{
			Value:       "$FORMAT/GT",
			Number:      "1",
			Type:        "String",
			Description: "Genotype",
		}
	}
}

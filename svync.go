package main

import (
	"log"
	"os"

	"github.com/nvnieuwk/svync/svync_api"
	cli "github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:            "svync",
		Usage:           "A tool to standardize VCF files from structural variant callers",
		HideHelpCommand: true,
		Version:         "0.1.0dev",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Usage:    "The location to the output VCF file, defaults to stdout",
				Category: "Optional",
			},
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "Configuration file (YAML) to use for the parsing of INFO and FORMAT fields",
				Required: true,
				Category: "Required",
			},
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Usage:    "The input VCF file to standardize",
				Required: true,
				Category: "Required",
			},
		},
		Action: func(Cctx *cli.Context) error {
			config := svync_api.ReadConfig(Cctx) // Outputs Config
			vcf := svync_api.ReadVcf(Cctx)       // Outputs VCF
			vcf.Standardize(config, Cctx)        // Standardize VCF and write to output file
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"log"
	"os"
	"slices"
	"strings"

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
			&cli.BoolFlag{
				Name:     "nodate",
				Aliases:  []string{"nd"},
				Usage:    "Don't add the current date to the output VCF header",
				Category: "Optional",
			},
			&cli.StringFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Usage:    "The location to the output VCF file, defaults to stdout",
				Category: "Optional",
			},
			&cli.StringFlag{
				Name:     "notation",
				Aliases:  []string{"n"},
				Usage:    "The notation to use for the output VCF file. Must be one of: breakpoint, breakend. By default the notation isn't changed",
				Category: "Optional",
				Action: func(c *cli.Context, input string) error {
					validNotations := []string{"breakpoint", "breakend"}
					if slices.Contains(validNotations, input) {
						return nil
					}
					return cli.Exit("Invalid notation '"+input+"', must be one of: "+strings.Join(validNotations, ", "), 1)
				},
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
			config := svync_api.ReadConfig(Cctx)
			vcf := svync_api.ReadVcf(Cctx)
			vcf.StandardizeAndOutput(config, Cctx) // Standardize VCF and write to output file
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.New(os.Stderr, "", 0).Fatal(err)
	}
}

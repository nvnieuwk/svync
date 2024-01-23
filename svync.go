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
		Version:         "0.1.2",
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
			// TODO re-add this when conversion is implemented
			// &cli.BoolFlag{
			// 	Name:     "to-breakpoint",
			// 	Aliases:  []string{"tb"},
			// 	Usage:    "Convert pairs of breakends to a single breakpoint variant. WARNING: this will cause some loss of data.",
			// 	Category: "Optional",
			// },
			&cli.BoolFlag{
				Name:     "mute-warnings",
				Aliases:  []string{"mw"},
				Usage:    "Mute all warnings.",
				Category: "Optional",
			},
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "Configuration file (YAML) used for standardizing the VCF",
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
			svync_api.Execute(Cctx, config)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.New(os.Stderr, "", 0).Fatal(err)
	}
}

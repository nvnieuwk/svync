package svync_api

import (
	"fmt"
	"log"
	"os"

	cli "github.com/urfave/cli/v2"
)

// Standardize the VCF file and write it to the output file
func (vcf *VCF) StandardizeAndOutput(config *Config, Cctx *cli.Context) {
	logger := log.New(os.Stderr, "", 0)
	stdout := true
	var file *os.File
	var err error
	if Cctx.String("output") != "" {
		stdout = false
		file, err = os.Create(Cctx.String("output"))
		if err != nil {
			logger.Fatalf("Failed to create the output file: %v", err)
		}
		defer file.Close()
	}

	// Write the info fields of the config
	// for _, infoField := range config.Info {
	// }

	for _, variant := range vcf.Variants {
		// Standardize the variant
		line := variant.standardizeToString(config, Cctx)
		writeLine(line, file, stdout)
	}
}

// Standardize the variant and return it as a string
func (variant *Variant) standardizeToString(config *Config, Cctx *cli.Context) string {
	// standardizedVariant := newVariant()
	return "hello"

}

// Initialize a new Variant
func newVariant() *Variant {
	return &Variant{
		Info:   map[string][]string{},
		Format: map[string]VariantFormat{},
	}
}

// Write a line to the output file or stdout
func writeLine(line string, file *os.File, stdout bool) {
	if stdout {
		fmt.Println(line)
	} else {
		file.WriteString(line + "\n")
	}
}

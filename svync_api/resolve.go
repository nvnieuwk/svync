package svync_api

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	cli "github.com/urfave/cli/v2"
)

// Resolve a value
func ResolveValue(input string, variant *Variant, format *VariantFormat, Cctx *cli.Context, config *Config) string {
	logger := log.New(os.Stderr, "", 0)

	// Replace all the FORMAT fields
	formatRegex := regexp.MustCompile(`\$FORMAT/[\w\d]+(/\d+)?`)
	allFormats := formatRegex.FindAllString(input, -1)
	if len(allFormats) > 0 && format == nil {
		logger.Fatalf("Cannot use a FORMAT field in a non-FORMAT context, please check your config file")
	}
	for _, rawField := range allFormats {
		fieldSlice := strings.Split(rawField, "/")

		field := fieldSlice[1]
		formatValue, ok := format.Content[field]

		if !ok {
			// Check if the field is a default value
			defaults := config.Format[field].Defaults
			if defaultValue, ok := defaults[rawField]; ok {
				formatValue = []string{defaultValue}
			} else if !Cctx.Bool("mute-warnings") {
				logger.Printf("The field %s is not present in the FORMAT fields of the variant with ID %s, excluding it from this variant. Supply a default to mute this warning", field, variant.Id)
			}
		} else if len(fieldSlice) > 2 {
			index, err := strconv.ParseInt(fieldSlice[2], 0, 64)
			if err != nil {
				logger.Fatal(err)
			}
			formatValue = []string{formatValue[index]}
		}
		input = strings.ReplaceAll(input, rawField, strings.Join(formatValue, ","))
	}

	// Replace all the INFO fields
	infoRegex := regexp.MustCompile(`\$INFO/[\w\d]+(/\d+)?`)
	allInfos := infoRegex.FindAllString(input, -1)
	for _, rawField := range allInfos {
		fieldSlice := strings.Split(rawField, "/")

		field := fieldSlice[1]
		info, ok := variant.Info[field]

		if !ok {
			// Check if the field is a default value
			defaults := config.Info[field].Defaults
			infoType := variant.Header.Info[field].Type
			if defaultValue, ok := defaults[rawField]; ok {
				info = []string{defaultValue}
			} else if infoType != "Flag" && !Cctx.Bool("mute-warnings") {
				logger.Printf("The field %s is not present in the FORMAT fields of the variant with ID %s, excluding it from this variant. Supply a default to mute this warning", field, variant.Id)
			}
		} else if len(fieldSlice) > 2 {
			index, err := strconv.ParseInt(fieldSlice[2], 0, 64)
			if err != nil {
				logger.Fatal(err)
			}
			info = []string{info[index]}
		}
		input = strings.ReplaceAll(input, rawField, strings.Join(info, ","))
	}

	// Replace POS fields
	input = strings.ReplaceAll(input, "$POS", fmt.Sprint(variant.Pos))

	// Replace CHROM fields
	input = strings.ReplaceAll(input, "$CHROM", variant.Chromosome)

	// Replace REF fields
	input = strings.ReplaceAll(input, "$REF", variant.Ref)

	// Replace ALT fields
	input = strings.ReplaceAll(input, "$ALT", variant.Alt)

	// Replace QUAL fields
	input = strings.ReplaceAll(input, "$QUAL", variant.Qual)

	// Replace FILTER fields
	input = strings.ReplaceAll(input, "$FILTER", variant.Filter)

	functionToken := "~"
	if !strings.Contains(input, functionToken) {
		return input
	}
	return resolveFunction(input, functionToken)
}

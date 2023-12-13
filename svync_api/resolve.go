package svync_api

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func ResolveValue(input string, variant *Variant, format *VariantFormat) string {
	logger := log.New(os.Stderr, "", 0)

	// Replace all the FORMAT fields
	formatRegex := regexp.MustCompile(`\$FORMAT/[\w\d]+`)
	allFormats := formatRegex.FindAllString(input, -1)
	if len(allFormats) > 0 && format == nil {
		logger.Fatalf("Cannot use a FORMAT field in a non-FORMAT context, please check your config file")
	}
	for _, stringToReplace := range allFormats {
		field := strings.Replace(stringToReplace, "$FORMAT/", "", 1)
		formatValue, ok := format.Content[field]
		// TODO implement some alternative way to handle missing fields
		if !ok {
			logger.Fatalf("The field %s is not present in the FORMAT fields of the variant with ID %s", field, variant.Id)
		}
		input = strings.ReplaceAll(input, stringToReplace, strings.Join(formatValue, ","))
	}

	// Replace all the INFO fields
	infoRegex := regexp.MustCompile(`\$INFO/[\w\d]+`)
	allInfos := infoRegex.FindAllString(input, -1)
	for _, stringToReplace := range allInfos {
		field := strings.Replace(stringToReplace, "$INFO/", "", 1)
		info, ok := variant.Info[field]
		// TODO implement some alternative way to handle missing fields
		if !ok {
			infoType := variant.Header.Info[field].Type
			if infoType != "Flag" {
				logger.Fatalf("The field %s is not present in the INFO fields of the variant with ID %s", field, variant.Id)
			}
		}
		input = strings.ReplaceAll(input, stringToReplace, strings.Join(info, ","))
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

	return input
}

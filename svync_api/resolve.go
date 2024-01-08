package svync_api

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Resolve a value
func ResolveValue(input string, variant *Variant, format *VariantFormat) string {
	logger := log.New(os.Stderr, "", 0)

	// Replace all the FORMAT fields
	formatRegex := regexp.MustCompile(`\$FORMAT/[\w\d]+(/\d+)?`)
	allFormats := formatRegex.FindAllString(input, -1)
	if len(allFormats) > 0 && format == nil {
		logger.Fatalf("Cannot use a FORMAT field in a non-FORMAT context, please check your config file")
	}
	for _, stringToReplace := range allFormats {
		fieldSlice := strings.Split(stringToReplace, "/")

		field := fieldSlice[1]
		formatValue, ok := format.Content[field]

		// TODO implement some alternative way to handle missing fields
		if !ok {
			logger.Printf("The field %s is not present in the FORMAT fields of the variant with ID %s, excluding it from this variant", field, variant.Id)
		} else if len(fieldSlice) > 2 {
			index, err := strconv.ParseInt(fieldSlice[2], 0, 64)
			if err != nil {
				logger.Fatal(err)
			}
			formatValue = []string{formatValue[index]}
		}
		input = strings.ReplaceAll(input, stringToReplace, strings.Join(formatValue, ","))
	}

	// Replace all the INFO fields
	infoRegex := regexp.MustCompile(`\$INFO/[\w\d]+(/\d+)?`)
	allInfos := infoRegex.FindAllString(input, -1)
	for _, stringToReplace := range allInfos {
		fieldSlice := strings.Split(stringToReplace, "/")

		field := fieldSlice[1]
		info, ok := variant.Info[field]

		// TODO implement some alternative way to handle missing fields
		if !ok {
			infoType := variant.Header.Info[field].Type
			if infoType != "Flag" {
				logger.Printf("The field %s is not present in the INFO fields of the variant with ID %s, excluding it from this variant", field, variant.Id)
			}
		} else if len(fieldSlice) > 2 {
			index, err := strconv.ParseInt(fieldSlice[2], 0, 64)
			if err != nil {
				logger.Fatal(err)
			}
			info = []string{info[index]}
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

	functionToken := "~"
	if !strings.Contains(input, functionToken) {
		return input
	}
	return resolveFunction(input, functionToken)
}

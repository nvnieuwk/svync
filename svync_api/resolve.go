package svync_api

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func ResolveInfo(input string, variant *Variant) string {
	logger := log.New(os.Stderr, "", 0)

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

	logger.Println(input)
	return input
}

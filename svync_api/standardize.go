package svync_api

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	cli "github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

	// VCF version
	writeLine("##fileformat=VCFv4.2", file, stdout)

	// Date of file creation
	if !Cctx.Bool("nodate") {
		cT := time.Now()
		dateLine := fmt.Sprintf("##fileDate=%d%02d%02d", cT.Year(), cT.Month(), cT.Day())
		writeLine(dateLine, file, stdout)
	}

	descriptionRegex := regexp.MustCompile(`["']?([^"']*)["']?`)

	// ALT header lines
	for _, alt := range vcf.Header.Alt {
		description := descriptionRegex.FindStringSubmatch(alt.Description)[1]
		altLine := fmt.Sprintf("##ALT=<ID=%s,Description=\"%s\">", alt.Id, description)
		writeLine(altLine, file, stdout)
	}

	// FILTER header lines
	for _, filter := range vcf.Header.Filter {
		description := descriptionRegex.FindStringSubmatch(filter.Description)[1]
		filterLine := fmt.Sprintf("##FILTER=<ID=%s,Description=\"%s\">", filter.Id, description)
		writeLine(filterLine, file, stdout)
	}

	// Write the info fields of the config
	for name, info := range config.Info {
		description := descriptionRegex.FindStringSubmatch(info.Description)[1]
		infoType := cases.Title(language.English, cases.Compact).String(strings.ToLower(info.Type))
		infoLine := fmt.Sprintf("##INFO=<ID=%s,Number=%s,Type=%s,Description=\"%s\">", name, info.Number, infoType, description)
		writeLine(infoLine, file, stdout)
	}

	// Write the format fields of the config
	for name, format := range config.Format {
		description := descriptionRegex.FindStringSubmatch(format.Description)[1]
		formatType := cases.Title(language.English, cases.Compact).String(strings.ToLower(format.Type))
		formatLine := fmt.Sprintf("##FORMAT=<ID=%s,Number=%s,Type=%s,Description=\"%s\">", name, format.Number, formatType, description)
		writeLine(formatLine, file, stdout)
	}

	// Write the contig fields
	for _, contig := range vcf.Header.Contig {
		contigLine := fmt.Sprintf("##contig=<ID=%s,length=%d>", contig.Id, contig.Length)
		writeLine(contigLine, file, stdout)
	}

	// Write the column headers
	columnHeaders := []string{"#CHROM", "POS", "ID", "REF", "ALT", "QUAL", "FILTER", "INFO", "FORMAT"}
	columnHeaders = append(columnHeaders, vcf.Header.Samples...)
	writeLine(strings.Join(columnHeaders, "\t"), file, stdout)

	// Write the variants
	variantCount := 0
	for _, variant := range vcf.Variants {
		// Standardize the variant
		line := variant.standardizeToString(config, Cctx, variantCount)
		variantCount++
		writeLine(line, file, stdout)
	}
}

// Standardize the variant and return it as a string
func (variant *Variant) standardizeToString(config *Config, Cctx *cli.Context, count int) string {
	// logger := log.New(os.Stderr, "", 0)
	standardizedVariant := newVariant()
	standardizedVariant.Chromosome = variant.Chromosome
	standardizedVariant.Pos = variant.Pos
	standardizedVariant.Ref = variant.Ref
	standardizedVariant.Alt = variant.Alt
	standardizedVariant.Qual = variant.Qual
	standardizedVariant.Filter = variant.Filter
	standardizedVariant.Header = variant.Header
	standardizedVariant.Id = fmt.Sprintf("%s_%v", ResolveValue(config.Id, variant, nil), count)

	// Add info fields
	for name, info := range config.Info {
		infoArray := []string{}
		for _, split := range strings.Split(info.Value, ",") {
			infoArray = append(infoArray, ResolveValue(split, variant, nil))
		}
		standardizedVariant.Info[name] = infoArray
	}

	// Add format fields
	for sample, format := range variant.Format {
		newFormat := newVariantFormat()
		newFormat.Sample = sample

		for name, formatConfig := range config.Format {
			formatArray := []string{}
			for _, split := range strings.Split(formatConfig.Value, ",") {
				formatArray = append(formatArray, ResolveValue(split, variant, &format))
			}
			newFormat.Content[name] = formatArray
		}
		standardizedVariant.Format[sample] = *newFormat
	}
	return standardizedVariant.String(config)
}

// Initialize a new Variant
func newVariant() *Variant {
	return &Variant{
		Info:   map[string][]string{},
		Format: map[string]VariantFormat{},
	}
}

// Initialize a new VariantFormat
func newVariantFormat() *VariantFormat {
	return &VariantFormat{
		Sample:  "",
		Content: map[string][]string{},
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

// Convert a variant to a string
func (v *Variant) String(config *Config) string {
	// logger := log.New(os.Stderr, "", 0)

	// Make sure the order of the info fields is respected
	infoSlice := []string{}
	infoLength := len(v.Info)

	infoKeys := make([]string, infoLength)
	l := 0
	for k := range v.Info {
		infoKeys[l] = k
		l++
	}
	sort.Strings(infoKeys)

	for _, key := range infoKeys {
		if config.Info[key].Type == "Flag" {
			infoSlice = append(infoSlice, key)
			continue
		}
		infoSlice = append(infoSlice, fmt.Sprintf("%s=%s", key, strings.Join(v.Info[key], ",")))
	}

	// Make sure the order of the format fields is respected
	samples := v.Header.Samples
	sort.Strings(samples)

	formatKeys := []string{}
	for k := range v.Format[samples[0]].Content {
		formatKeys = append(formatKeys, k)
	}
	sort.Strings(formatKeys)

	formatString := ""
	formatString += strings.Join(formatKeys, ":")

	for _, sample := range samples {
		sampleArray := []string{}
		for _, key := range formatKeys {
			sampleArray = append(sampleArray, strings.Join(v.Format[sample].Content[key], ","))
		}
		formatString += fmt.Sprintf("\t%s", strings.Join(sampleArray, ":"))
	}

	return fmt.Sprintf(
		"%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v",
		v.Chromosome,
		v.Pos,
		v.Id,
		v.Ref,
		v.Alt,
		v.Qual,
		v.Filter,
		strings.Join(infoSlice, ";"),
		formatString,
	)
}

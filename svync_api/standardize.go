package svync_api

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	cli "github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func writeHeader(config *Config, Cctx *cli.Context, header *Header, file *os.File, stdout bool) {
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
	for _, alt := range header.Alt {
		altId := alt.Id
		if newAlt, ok := config.Alt[altId]; ok {
			altId = newAlt
		}
		description := descriptionRegex.FindStringSubmatch(alt.Description)[1]
		altLine := fmt.Sprintf("##ALT=<ID=%s,Description=\"%s\">", altId, description)
		writeLine(altLine, file, stdout)
	}

	// FILTER header lines
	for _, filter := range header.Filter {
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
	for _, contig := range header.Contig {
		contigLine := fmt.Sprintf("##contig=<ID=%s,length=%d>", contig.Id, contig.Length)
		writeLine(contigLine, file, stdout)
	}

	// Write the column headers
	columnHeaders := []string{"#CHROM", "POS", "ID", "REF", "ALT", "QUAL", "FILTER", "INFO", "FORMAT"}
	columnHeaders = append(columnHeaders, header.Samples...)
	writeLine(strings.Join(columnHeaders, "\t"), file, stdout)
}

// Standardize the VCF file and write it to the output file
func standardizeAndOutput(config *Config, Cctx *cli.Context, variant *Variant, file *os.File, stdout bool, variantCount int) {

	// Standardize the variant
	line := variant.standardize(config, Cctx, variantCount).String(config)
	writeLine(line, file, stdout)

}

// Standardize the variant using the config
func (variant *Variant) standardize(config *Config, Cctx *cli.Context, count int) *Variant {
	standardizedVariant := newVariant()
	standardizedVariant.Chromosome = variant.Chromosome
	standardizedVariant.Pos = variant.Pos
	standardizedVariant.Ref = variant.Ref
	standardizedVariant.Alt = variant.Alt
	standardizedVariant.Qual = variant.Qual
	standardizedVariant.Filter = variant.Filter
	standardizedVariant.Header = variant.Header

	sVType := variant.Info["SVTYPE"][0]

	if config.Alt != nil {
		if alt, ok := config.Alt[sVType]; ok {
			sVType = alt
			standardizedVariant.Alt = "<" + alt + ">"
			variant.Info["SVTYPE"] = []string{alt}
		}
	}

	standardizedVariant.Id = fmt.Sprintf("%s_%v", ResolveValue(config.Id, variant, nil), count)

	// Add info fields
	for name, infoConfig := range config.Info {
		value := infoConfig.Value
		if val, ok := config.Info[name].Alts[sVType]; ok {
			value = val
		}
		// Don't add INFO fields with empty values
		if value == "" {
			continue
		}
		standardizedVariant.Info[name] = []string{ResolveValue(value, variant, nil)}
	}

	// Add format fields
	for sample, format := range variant.Format {
		newFormat := newVariantFormat()
		newFormat.Sample = sample

		for name, formatConfig := range config.Format {
			value := formatConfig.Value
			if val, ok := formatConfig.Alts[sVType]; ok {
				value = val
			}
			newFormat.Content[name] = []string{ResolveValue(value, variant, &format)}
		}
		standardizedVariant.Format[sample] = *newFormat
	}
	return standardizedVariant
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
		value := v.Info[key]
		if value[0] == "" && len(value) == 1 {
			continue
		}
		infoSlice = append(infoSlice, fmt.Sprintf("%s=%s", key, strings.Join(value, ",")))
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

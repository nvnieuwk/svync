package svync_api

import (
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
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
		if variant.Parsed {
			continue
		}
		line := ""
		if Cctx.String("notation") == "breakpoint" {
			if variant.Info["SVTYPE"][0] == "BND" {
				line = variant.toBreakPoint(vcf).standardize(config, Cctx, variantCount).String(config)
			} else {
				line = variant.standardize(config, Cctx, variantCount).String(config)
			}
		} else if Cctx.String("notation") == "breakend" {
			variant1, variant2 := variant.standardize(config, Cctx, variantCount).toBreakEnd()
			line = variant1.String(config) + "\n" + variant2.String(config)
		} else {
			line = variant.standardize(config, Cctx, variantCount).String(config)
		}
		variantCount++
		writeLine(line, file, stdout)
	}
}

func (variant *Variant) toBreakPoint(vcf *VCF) *Variant {
	logger := log.New(os.Stderr, "", 0)
	mateIds := variant.Info["MATEID"]

	// Don't convert if less or more than 1 mate are found
	if len(mateIds) == 0 || len(mateIds) > 1 {
		return variant
	}

	mateId := mateIds[0]
	mateVariant, ok := vcf.Variants[mateId]
	if ok {
		mateVariant.Parsed = true
		vcf.Variants[mateId] = mateVariant
	}

	alt := variant.Alt
	altRegex := regexp.MustCompile(`(\[|\])(?P<chr>[^:]*):(?P<pos>[0-9]*)`)
	altGroups := altRegex.FindStringSubmatch(alt)

	chr := variant.Chromosome
	pos := variant.Pos
	chr2 := altGroups[2]
	pos2, err := strconv.ParseInt(altGroups[3], 0, 64)
	if err != nil {
		logger.Fatalf("Couldn't convert string to integer: %v", err)
	}
	bracket := altGroups[1]
	strand1 := ""
	strand2 := ""

	if strings.HasSuffix(alt, "[") || strings.HasSuffix(alt, "]") {
		strand1 = "-"
	} else {
		strand1 = "+"
	}

	if bracket == string('[') {
		strand2 = "-"
	} else {
		strand2 = "+"
	}

	filter := "."
	if variant.Filter == mateVariant.Filter {
		filter = variant.Filter
	}

	varQual, err := strconv.ParseFloat(variant.Qual, 64)
	mateQual, err := strconv.ParseFloat(mateVariant.Qual, 64)
	qual := "."
	if err == nil {
		qual = fmt.Sprintf("%f", (varQual+mateQual)/2)
	}

	breakpointVariant := &Variant{
		Chromosome: chr,
		Pos:        pos,
		Id:         variant.Id,
		Ref:        variant.Ref,
		Filter:     filter,
		Qual:       qual,
		Header:     variant.Header,
		Info:       variant.Info,
		Format:     variant.Format,
	}

	breakpointVariant.Info["END"] = []string{fmt.Sprint(pos2)}
	breakpointVariant.Info["CHR2"] = []string{chr2}

	// Define all types and determine their svlen
	svtype := ""
	svlen := strconv.FormatFloat(math.Abs(float64(pos2-pos)), 'g', -1, 64)
	if chr != chr2 {
		svtype = "TRA"
		svlen = "0"
	} else if strand1 == strand2 {
		svtype = "INV"
	} else if float64(getInsLen(alt, strand1, bracket)) > math.Abs(float64(pos2-pos))*0.5 {
		svtype = "INS"
		svlen = fmt.Sprint(getInsLen(alt, strand1, bracket))
	} else if pos < pos2 && strand1 == "-" && strand2 == "+" {
		svtype = "DUP"
	} else if pos > pos2 && strand1 == "+" && strand2 == "-" {
		svtype = "DUP"
	} else {
		svtype = "DEL"
		svlen = strconv.FormatFloat(-math.Abs(float64(pos2-pos)), 'g', -1, 64)
	}

	if svtype != "" {
		breakpointVariant.Alt = svtype
		breakpointVariant.Info["SVTYPE"] = []string{svtype}
		breakpointVariant.Info["SVLEN"] = []string{svlen}
		return breakpointVariant
	}

	return variant
}

// Determining the ALT fields is impossible for breakpoint -> breakend conversion
func (variant *Variant) toBreakEnd() (*Variant, *Variant) {
	logger := log.New(os.Stderr, "", 0)

	id1 := variant.Id + "_01"
	id2 := variant.Id + "_02"

	end, err := strconv.ParseInt(variant.Info["END"][0], 0, 64)
	if err != nil {
		logger.Fatal(err)
	}

	chr := variant.Chromosome
	chr2 := ""
	chrom2, ok := variant.Info["CHR2"]
	if !ok {
		chr2 = variant.Chromosome
	} else {
		chr2 = chrom2[0]
	}

	info1 := map[string][]string{}
	info2 := map[string][]string{}
	for name, info := range variant.Info {
		if slices.Contains([]string{"CHR2", "END", "SVLEN"}, name) {
			continue
		}
		info1[name] = info
		info2[name] = info
	}

	info1["SVTYPE"] = []string{"BND"}
	info2["SVTYPE"] = []string{"BND"}
	info1["MATEID"] = []string{id2}
	info2["MATEID"] = []string{id1}

	alt1 := "."
	alt2 := "."

	breakend1 := &Variant{
		Chromosome: chr,
		Pos:        variant.Pos,
		Id:         id1,
		Alt:        alt1,
		Ref:        variant.Ref,
		Qual:       variant.Qual,
		Filter:     variant.Filter,
		Header:     variant.Header,
		Info:       info1,
		Format:     variant.Format,
	}

	breakend2 := &Variant{
		Chromosome: chr2,
		Pos:        end,
		Id:         id2,
		Alt:        alt2,
		Ref:        "N", // TODO find a good way to determine the breakend2 REF
		Qual:       variant.Qual,
		Filter:     variant.Filter,
		Header:     variant.Header,
		Info:       info2,
		Format:     variant.Format,
	}

	return breakend1, breakend2
}

func getInsLen(alt string, strand string, bracket string) int {
	if strand == "-" {
		return len(alt[strings.LastIndex(alt, bracket):])
	}
	return len(alt[:strings.LastIndex(alt, bracket)])
}

// Standardize the variant and return it as a string
func (variant *Variant) standardize(config *Config, Cctx *cli.Context, count int) *Variant {
	// logger := log.New(os.Stderr, "", 0)
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

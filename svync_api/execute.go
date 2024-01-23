package svync_api

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/biogo/hts/bgzf"
	cli "github.com/urfave/cli/v2"
)

// Read the VCF file and return it as a VCF struct
func Execute(Cctx *cli.Context, config *Config) {
	logger := log.New(os.Stderr, "", 0)

	file := Cctx.String("input")
	inputVcf, err := os.Open(file)
	defer inputVcf.Close()
	if err != nil {
		logger.Fatal(err)
	}
	header := newHeader()
	breakEndVariants := &map[string]Variant{}
	headerIsMade := false
	variantCount := 0

	stdout := true
	var outputFile *os.File
	if Cctx.String("output") != "" {
		stdout = false
		outputFile, err = os.Create(Cctx.String("output"))
		if err != nil {
			logger.Fatalf("Failed to create the output file: %v", err)
		}
		defer outputFile.Close()
	}

	if strings.HasSuffix(file, ".gz") {
		bgReader, err := bgzf.NewReader(inputVcf, 1)
		if err != nil {
			logger.Fatal(err)
		}
		defer bgReader.Close()

		for {
			b, _, err := readBgzipLine(bgReader)
			if err != nil {
				if err == io.EOF {
					break
				}
				logger.Fatal(string(b[:]))
			}

			parseLine(
				string(bytes.TrimSpace(b[:])),
				header,
				breakEndVariants,
				config,
				Cctx,
				&headerIsMade,
				outputFile,
				stdout,
				&variantCount,
			)
		}
	} else {
		scanner := bufio.NewScanner(inputVcf)
		const maxCapacity = 8 * 1000000 // 8 MB
		scanner.Buffer(make([]byte, maxCapacity), maxCapacity)
		for scanner.Scan() {
			parseLine(
				scanner.Text(),
				header,
				breakEndVariants,
				config,
				Cctx,
				&headerIsMade,
				outputFile,
				stdout,
				&variantCount,
			)
		}

		if err := scanner.Err(); err != nil {
			logger.Fatal(err)
		}
	}

	if !headerIsMade {
		writeHeader(config, Cctx, header, outputFile, stdout)
		headerIsMade = true
	}

}

// readBgzipLine reads a line from a bgzip file
func readBgzipLine(r *bgzf.Reader) ([]byte, bgzf.Chunk, error) {
	tx := r.Begin()
	var (
		data []byte
		b    byte
		err  error
	)
	for {
		b, err = r.ReadByte()
		if err != nil {
			break
		}
		data = append(data, b)
		if b == '\n' {
			break
		}
	}
	chunk := tx.End()
	return data, chunk, err
}

// Parse the line and add it to the VCF struct
func parseLine(
	line string,
	header *Header,
	breakEndVariants *map[string]Variant,
	config *Config,
	Cctx *cli.Context,
	headerIsMade *bool,
	outputFile *os.File,
	stdout bool,
	variantCount *int,
) {
	if strings.HasPrefix(line, "#") {
		header.parse(line)
	} else {
		if !*headerIsMade {
			writeHeader(config, Cctx, header, outputFile, stdout)
			*headerIsMade = true
		}
		// id := strings.Split(line, "\t")[2]
		variant := createVariant(line, header, Cctx)

		// TODO continue work on this later
		// Convert breakends to breakpoints if the --to-breakpoint flag is set
		// if Cctx.Bool("to-breakpoint") && variant.Info["SVTYPE"][0] == "BND" && len(variant.Info["MATEID"]) == 1 {
		// 	mateid := variant.Info["MATEID"][0]
		// 	if mate, ok := (*breakEndVariants)[mateid]; ok {
		// 		variant = toBreakPoint(variant, &mate)
		// 		delete(*breakEndVariants, mateid)
		// 	} else {
		// 		(*breakEndVariants)[id] = *variant
		// 		return
		// 	}
		// }
		*variantCount++
		standardizeAndOutput(config, Cctx, variant, outputFile, stdout, *variantCount)

		// Standardize and output the variant
	}
}

// Parse the line and add it to the Variant struct
func createVariant(line string, header *Header, Cctx *cli.Context) *Variant {
	logger := log.New(os.Stderr, "", 0)

	variant := new(Variant)
	variant.Header = header

	data := strings.Split(line, "\t")
	variant.Chromosome = data[0]

	var err error
	variant.Pos, err = strconv.ParseInt(data[1], 0, 64)
	if err != nil {
		logger.Fatal(err)
	}
	variant.Id = data[2]
	variant.Ref = data[3]
	variant.Alt = data[4]
	variant.Qual = data[5]
	variant.Filter = data[6]

	variant.Info = map[string][]string{}
	info := strings.Split(data[7], ";")
	for _, i := range info {
		split := strings.Split(i, "=")
		field := split[0]
		value := ""
		if len(split) > 1 {
			value = split[1]
		}
		variant.Info[field] = parseInfoFormat(field, value, variant.Header.Info, Cctx)
	}

	variant.Format = map[string]VariantFormat{}
	formatHeaders := strings.Split(data[8], ":")
	formatValues := data[9:]
	for index, value := range formatValues {
		sample := variant.Header.Samples[index]
		variant.Format[sample] = VariantFormat{
			Sample:  sample,
			Content: map[string][]string{},
		}
		for idx, val := range strings.Split(value, ":") {
			header := formatHeaders[idx]
			variant.Format[sample].Content[header] = parseInfoFormat(header, val, variant.Header.Format, Cctx)
		}
	}

	return variant

}

// Parse the value of the INFO or FORMAT field and return it as a slice of strings
func parseInfoFormat(header string, value string, infoFormatLines map[string]HeaderLineIdNumberTypeDescription, Cctx *cli.Context) []string {
	logger := log.New(os.Stderr, "", 0)
	headerLine := infoFormatLines[header]
	if headerLine == (HeaderLineIdNumberTypeDescription{}) {
		if !Cctx.Bool("mute-warnings") {
			logger.Printf("Field %s not found in header, defaulting to Type 'String' and Number '1'", header)
		}
		headerLine = HeaderLineIdNumberTypeDescription{
			Id:          header,
			Number:      "1",
			Type:        "String",
			Description: "",
		}
	}

	if headerLine.Type == "Flag" {
		return []string{}
	}

	infoNumber, err := strconv.ParseInt(headerLine.Number, 0, 64)
	if err != nil {
		infoNumber = -1
	}
	return strings.SplitN(value, ",", int(infoNumber))
}

// Parse the header line and add it to the Header struct
func (header *Header) parse(line string) {
	if strings.HasPrefix(line, "#CHROM") {
		header.Samples = strings.Split(line, "\t")[9:]
		return
	}

	r := regexp.MustCompile(`^##(?P<headerType>[^=]*)=<(?P<content>.*)>$`)
	matches := r.FindStringSubmatch(line)

	if len(matches) == 0 {
		if header.Other == nil {
			header.Other = []string{}
		}
		header.Other = append(header.Other, line)
		return
	}

	headerType := matches[1]
	content := matches[2]
	contentMap := convertLineToMap(content)

	switch headerType {
	case "INFO":
		header.Info[contentMap["id"]] = HeaderLineIdNumberTypeDescription{
			Id:          contentMap["id"],
			Number:      contentMap["number"],
			Type:        contentMap["type"],
			Description: contentMap["description"],
		}
	case "FORMAT":
		header.Format[contentMap["id"]] = HeaderLineIdNumberTypeDescription{
			Id:          contentMap["id"],
			Number:      contentMap["number"],
			Type:        contentMap["type"],
			Description: contentMap["description"],
		}
	case "ALT":
		header.Alt[contentMap["id"]] = HeaderLineIdDescription{
			Id:          contentMap["id"],
			Description: contentMap["description"],
		}
	case "FILTER":
		header.Filter[contentMap["id"]] = HeaderLineIdDescription{
			Id:          contentMap["id"],
			Description: contentMap["description"],
		}
	case "contig":
		length, err := strconv.ParseInt(contentMap["length"], 0, 64)
		if err != nil {
			log.Fatalf("Could not convert contig length to an integer: %v", err)
		}
		header.Contig = append(header.Contig, HeaderLineIdLength{
			Id:     contentMap["id"],
			Length: length,
		})
	}

}

// convertLineToMap converts the header line contents to a map suitable to transform to a struct
func convertLineToMap(line string) map[string]string {
	data := map[string]string{}
	word := ""
	key := ""
	quote := ""
	for _, letter := range strings.Split(line, "") {
		if letter == "=" {
			key = strings.ToLower(word)
			word = ""
			continue
		} else if letter == "," && quote == "" {
			data[key] = word
			key = ""
			word = ""
			continue
		}

		word += letter

		if letter == quote {
			quote = ""
		} else if letter == "\"" || letter == "'" {
			quote = letter
		}
	}
	data[key] = word

	return data
}

// Create a new header struct
func newHeader() *Header {
	return &Header{
		Info:    map[string]HeaderLineIdNumberTypeDescription{},
		Format:  map[string]HeaderLineIdNumberTypeDescription{},
		Alt:     map[string]HeaderLineIdDescription{},
		Filter:  map[string]HeaderLineIdDescription{},
		Contig:  []HeaderLineIdLength{},
		Other:   []string{},
		Samples: []string{},
	}
}

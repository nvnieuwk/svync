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
func ReadVcf(Cctx *cli.Context) *VCF {
	logger := log.New(os.Stderr, "", 0)

	file := Cctx.String("input")
	openFile, err := os.Open(file)
	defer openFile.Close()
	if err != nil {
		logger.Fatal(err)
	}

	vcf := newVCF()
	if strings.HasSuffix(file, ".gz") {
		vcf.readBgzip(openFile)
	} else {
		vcf.readPlain(openFile)
	}

	return vcf
}

// Initialize a new VCF
func newVCF() *VCF {
	return &VCF{
		Header: Header{
			Info:   map[string]HeaderLineIdNumberTypeDescription{},
			Format: map[string]HeaderLineIdNumberTypeDescription{},
			Alt:    map[string]HeaderLineIdDescription{},
			Filter: map[string]HeaderLineIdDescription{},
			Contig: []HeaderLineIdLength{},
		},
		Variants: map[string]Variant{},
	}
}

// Read the VCF file in bgzip format and convert it to a VCF struct
func (vcf *VCF) readBgzip(input *os.File) {
	logger := log.New(os.Stderr, "", 0)

	bgReader, err := bgzf.NewReader(input, 1)
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

		vcf.parse(string(bytes.TrimSpace(b[:])))
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

// Read the VCF file in plain text format and convert it to a VCF struct
func (vcf *VCF) readPlain(input *os.File) {
	logger := log.New(os.Stderr, "", 0)

	scanner := bufio.NewScanner(input)
	const maxCapacity = 8 * 1000000 // 8 MB
	scanner.Buffer(make([]byte, maxCapacity), maxCapacity)
	for scanner.Scan() {
		vcf.parse(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		logger.Fatal(err)
	}

}

// Parse the line and add it to the VCF struct
func (vcf *VCF) parse(line string) {
	if strings.HasPrefix(line, "#") {
		vcf.Header.parse(line)
	} else {
		id := strings.Split(line, "\t")[2]
		variant := &Variant{}
		variant.Header = &vcf.Header
		variant.parse(line)
		vcf.Variants[id] = *variant
		// logger.Println(vcf.Variants[id])
	}
}

// Parse the line and add it to the Variant struct
func (variant *Variant) parse(line string) {
	logger := log.New(os.Stderr, "", 0)

	err := error(nil)
	data := strings.Split(line, "\t")
	variant.Chromosome = data[0]
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
		variant.Info[field] = parseInfoFormat(field, value, variant.Header.Info)
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
			variant.Format[sample].Content[header] = parseInfoFormat(header, val, variant.Header.Format)
		}
	}

}

// Parse the value of the INFO or FORMAT field and return it as a slice of strings
func parseInfoFormat(header string, value string, infoFormatLines map[string]HeaderLineIdNumberTypeDescription) []string {
	logger := log.New(os.Stderr, "", 0)
	headerLine := infoFormatLines[header]
	if headerLine == (HeaderLineIdNumberTypeDescription{}) {
		logger.Printf("Field %s not found in header, defaulting to Type 'String' and Number '1'", header)
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

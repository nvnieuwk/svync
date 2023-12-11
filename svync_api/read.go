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

func Read(Cctx *cli.Context) *VCF {
	logger := log.New(os.Stderr, "", 0)

	file := Cctx.String("input")
	openFile, err := os.Open(file)
	defer openFile.Close()
	if err != nil {
		logger.Fatal(err)
	}

	vcf := NewVCF()
	if strings.HasSuffix(file, ".gz") {
		vcf.readBgzip(openFile)
	} else {
		vcf.readPlain(openFile)
	}

	return vcf
}

func NewVCF() *VCF {
	return &VCF{
		Header: Header{
			Info:   map[string]HeaderLineIdNumberTypeDescription{},
			Format: map[string]HeaderLineIdNumberTypeDescription{},
			Alt:    map[string]HeaderLineIdDescription{},
			Filter: map[string]HeaderLineIdDescription{},
			Contig: map[string]HeaderLineIdLength{},
		},
		Variants: map[string]Variant{},
	}
}

func (vcf *VCF) readBgzip(input *os.File) {
	logger := log.New(os.Stderr, "", 0)

	bgReader, err := bgzf.NewReader(input, 1)
	if err != nil {
		logger.Fatal(err)
	}
	defer bgReader.Close()

	for {
		b, _, err := readLine(bgReader)
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Fatal(string(b[:]))
		}

		vcf.parse(string(bytes.TrimSpace(b[:])))
	}

}

func readLine(r *bgzf.Reader) ([]byte, bgzf.Chunk, error) {
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

func (vcf *VCF) parse(line string) {
	// logger := log.New(os.Stderr, "", 0)

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
		headerLine := variant.Header.Info[field]
		if headerLine == (HeaderLineIdNumberTypeDescription{}) {
			logger.Printf("Field INFO/%s not found in header, defaulting to Type 'String' and Number '1'", field)
			headerLine = HeaderLineIdNumberTypeDescription{
				Id:          field,
				Number:      "1",
				Type:        "String",
				Description: "",
			}
		}

		if headerLine.Type == "Flag" {
			variant.Info[field] = []string{}
			continue
		}

		logger.Print(headerLine.Number)
		infoNumber, err := strconv.ParseInt(headerLine.Number, 0, 64)
		if err != nil {
			infoNumber = -1
		}
		variant.Info[field] = strings.SplitN(split[1], ",", int(infoNumber))
		logger.Print(variant.Info[field])
	}

	variant.Format = map[string]VariantFormat{}

}

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
		header.Contig[contentMap["id"]] = HeaderLineIdLength{
			Id:     contentMap["id"],
			Length: 0,
		}
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

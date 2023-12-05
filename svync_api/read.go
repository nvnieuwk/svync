package svync_api

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/biogo/hts/bgzf"
	cli "github.com/urfave/cli/v2"
)

func Read(Cctx *cli.Context) {
	logger := log.New(os.Stderr, "", 0)

	file := Cctx.String("input")
	openFile, err := os.Open(file)
	defer openFile.Close()
	if err != nil {
		logger.Fatal(err)
	}

	vcf := VCF{}
	if strings.HasSuffix(file, ".gz") {
		vcf = readBgzip(openFile)
	} else {
		vcf = readPlain(openFile)
	}
	logger.Println(vcf)
}

func readBgzip(input *os.File) VCF {
	logger := log.New(os.Stderr, "", 0)

	bgReader, err := bgzf.NewReader(input, 1)
	if err != nil {
		logger.Fatal(err)
	}
	defer bgReader.Close()

	vcf := VCF{}
	for {
		b, _, err := readLine(bgReader)
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Fatal(string(b[:]))
		}

		vcf.Parse(string(bytes.TrimSpace(b[:])))
	}

	return vcf

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

func readPlain(input *os.File) VCF {
	logger := log.New(os.Stderr, "", 0)

	vcf := VCF{}
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		vcf.Parse(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		logger.Fatal(err)
	}

	return vcf
}

func (vcf *VCF) Parse(line string) {
	// logger := log.New(os.Stderr, "", 0)

	if strings.HasPrefix(line, "#") {
		vcf.Header.Parse(line)
	}
}

func (header *Header) Parse(line string) {
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
		if header.Info == nil {
			header.Info = map[string]HeaderLineIdNumberTypeDescription{}
		}
		header.Info[contentMap["id"]] = HeaderLineIdNumberTypeDescription{
			Id:          contentMap["id"],
			Number:      contentMap["number"],
			Type:        contentMap["type"],
			Description: contentMap["description"],
		}
	case "FORMAT":
		if header.Format == nil {
			header.Format = map[string]HeaderLineIdNumberTypeDescription{}
		}
		header.Format[contentMap["id"]] = HeaderLineIdNumberTypeDescription{
			Id:          contentMap["id"],
			Number:      contentMap["number"],
			Type:        contentMap["type"],
			Description: contentMap["description"],
		}
	case "ALT":
		if header.Alt == nil {
			header.Alt = map[string]HeaderLineIdDescription{}
		}
		header.Alt[contentMap["id"]] = HeaderLineIdDescription{
			Id:          contentMap["id"],
			Description: contentMap["description"],
		}
	case "FILTER":
		if header.Filter == nil {
			header.Filter = map[string]HeaderLineIdDescription{}
		}
		header.Filter[contentMap["id"]] = HeaderLineIdDescription{
			Id:          contentMap["id"],
			Description: contentMap["description"],
		}
	case "contig":
		if header.Contig == nil {
			header.Contig = map[string]HeaderLineIdLength{}
		}
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

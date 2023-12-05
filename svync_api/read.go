package svync_api

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
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
	lines := []string{}
	if strings.HasSuffix(file, ".gz") {
		lines = readBgzip(openFile)
	} else {
		lines = readPlain(openFile)
	}
	logger.Println(strings.Join(lines, "\n"))
}

func readBgzip(input *os.File) []string {
	logger := log.New(os.Stderr, "", 0)

	bgReader, err := bgzf.NewReader(input, 1)
	if err != nil {
		logger.Fatal(err)
	}
	defer bgReader.Close()

	var data []string

	var n int
	for {
		n++
		b, _, err := readLine(bgReader)
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Fatal(string(b[:]))
		}

		data = append(data, string(bytes.TrimSpace(b[:])))
	}

	return data

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

func readPlain(input *os.File) []string {
	logger := log.New(os.Stderr, "", 0)

	var data []string
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		data = append(data, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		logger.Fatal(err)
	}

	return data
}

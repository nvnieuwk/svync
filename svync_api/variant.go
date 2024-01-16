package svync_api

import (
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Convert a breakend variant pair to one breakpoint
func toBreakPoint(mate1 *Variant, mate2 *Variant) *Variant {
	logger := log.New(os.Stderr, "", 0)

	alt := mate1.Alt
	altRegex := regexp.MustCompile(`(\[|\])(?P<chr>[^:]*):(?P<pos>[0-9]*)`)
	altGroups := altRegex.FindStringSubmatch(alt)

	chr := mate1.Chromosome
	pos := mate1.Pos
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
	if mate1.Filter == mate2.Filter {
		filter = mate1.Filter
	}

	varQual, err := strconv.ParseFloat(mate1.Qual, 64)
	mateQual, err := strconv.ParseFloat(mate2.Qual, 64)
	qual := "."
	if err == nil {
		qual = fmt.Sprintf("%f", (varQual+mateQual)/2)
	}

	breakpointVariant := &Variant{
		Chromosome: chr,
		Pos:        pos,
		Id:         mate1.Id,
		Ref:        mate1.Ref,
		Filter:     filter,
		Qual:       qual,
		Header:     mate1.Header,
		Info:       mate1.Info,
		Format:     mate1.Format,
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

	breakpointVariant.Alt = svtype
	breakpointVariant.Info["SVTYPE"] = []string{svtype}
	breakpointVariant.Info["SVLEN"] = []string{svlen}
	return breakpointVariant
}

// Get the length of an insertion
func getInsLen(alt string, strand string, bracket string) int {
	if strand == "-" {
		return len(alt[strings.LastIndex(alt, bracket):])
	}
	return len(alt[:strings.LastIndex(alt, bracket)])
}

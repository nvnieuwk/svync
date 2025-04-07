package svync_api

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Convert a breakend variant pair to one breakpoint
func toBreakPoint(mate1 *Variant, mate2 *Variant) *Variant {

	// Swap mates if the first one comes after the second one
	if mate1.Chromosome == mate2.Chromosome && mate1.Pos > mate2.Pos {
		mate1, mate2 = mate2, mate1
	}

	alt := mate1.Alt
	chr := mate1.Chromosome
	pos := mate1.Pos
	chr2 := mate2.Chromosome
	pos2 := mate2.Pos

	altRegex := regexp.MustCompile(`(\[|\])(?P<chr>[^:]*):(?P<pos>[0-9]*)`)
	altGroups := altRegex.FindStringSubmatch(alt)
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

	varQual, err1 := strconv.ParseFloat(mate1.Qual, 64)
	mateQual, err2 := strconv.ParseFloat(mate2.Qual, 64)
	qual := "."
	if err1 == nil && err2 == nil {
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

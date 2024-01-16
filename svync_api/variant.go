package svync_api

import (
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"slices"
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

// Convert a breakpoint to variant to its breakend pairs
//
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

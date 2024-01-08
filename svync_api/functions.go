package svync_api

import (
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func resolveFunction(input string, token string) string {
	logger := log.New(os.Stderr, "", 0)

	result := ""
	prefixRegex := regexp.MustCompile(`^([^~]*)~`)
	prefixResults := prefixRegex.FindStringSubmatch(input)
	if len(prefixResults) > 0 {
		result += prefixResults[1]
	}
	functionRegex := regexp.MustCompile(`~(\w+):([^;]*)`)
	functionResults := functionRegex.FindStringSubmatch(input)
	if len(functionResults) == 0 {
		logger.Fatalf("No function found in '%s'", input)
	}
	function := functionResults[1]
	value := strings.Split(functionResults[2], ",")
	for v := range value {
		if strings.Contains(value[v], token) {
			value[v] = resolveFunction(strings.Join(value[v:], ","), token)
		}
	}

	switch function {
	case "sub":
		result += sub(value)
	case "sum":
		result += sum(value)
	default:
		logger.Fatalf("The function '%s' is not supported", value[1:])
	}
	return result
}

func sub(input []string) string {
	result := stringToFloat(input[0])
	for i := 1; i < len(input); i++ {
		result -= stringToFloat(input[i])
	}
	return floatToString(result)
}

func sum(input []string) string {
	result := stringToFloat(input[0])
	for i := 1; i < len(input); i++ {
		result += stringToFloat(input[i])
	}
	return floatToString(result)
}

func stringToFloat(input string) float64 {
	result, err := strconv.ParseFloat(input, 64)
	if err != nil {
		log.Fatalf("Cannot convert '%s' to float64", input)
	}
	return result
}

func floatToString(input float64) string {
	return strconv.FormatFloat(input, 'f', -1, 64)
}

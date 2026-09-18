package cli

import (
	"regexp"
	"strings"
)

const defaultStartStatus = "In Progress"

var shortFlagGroupPattern = regexp.MustCompile(`^-[A-Za-z]+$`)

func NormalizeOptionalValue(args []string) []string {
	if len(args) == 0 || args[0] != "create" {
		return args
	}

	normalized := make([]string, 0, len(args))

	for index := 0; index < len(args); index++ {
		token := args[index]

		if token == "--" {
			return append(normalized, args[index:]...)
		}

		prefix, isStart := splitStartToken(token)
		if !isStart {
			normalized = append(normalized, token)
			continue
		}

		if prefix != "" {
			normalized = append(normalized, prefix)
		}

		value := defaultStartStatus
		if index+1 < len(args) && !strings.HasPrefix(args[index+1], "-") {
			value = args[index+1]
			index++
		}

		normalized = append(normalized, "--start="+value)
	}

	return normalized
}

func splitStartToken(token string) (string, bool) {
	if token == "--start" || token == "-s" {
		return "", true
	}

	if !shortFlagGroupPattern.MatchString(token) || !strings.HasSuffix(token, "s") {
		return "", false
	}

	return "-" + token[1:len(token)-1], true
}

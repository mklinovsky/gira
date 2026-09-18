package cli

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const jsSpace = `\t\n\v\f\r\p{Zs}\x{2028}\x{2029}\x{feff}`

var (
	branchStripPattern = regexp.MustCompile(`[^a-zA-Z0-9` + jsSpace + `{1}_-]+`)
	branchDashPattern  = regexp.MustCompile(`[` + jsSpace + `_]+`)
	jiraKeyPattern     = regexp.MustCompile(`^([A-Z]+-[0-9]+)`)
	jiraSummaryPattern = regexp.MustCompile(`^([A-Z]+-[0-9]+)-(.+)`)
	titleSplitPattern  = regexp.MustCompile(`[/_-]+`)
)

var strippableBranchPrefixes = map[string]bool{
	"feat":     true,
	"feature":  true,
	"fix":      true,
	"bugfix":   true,
	"chore":    true,
	"refactor": true,
	"docs":     true,
	"test":     true,
	"ci":       true,
	"build":    true,
	"perf":     true,
	"style":    true,
	"hotfix":   true,
	"release":  true,
	"wip":      true,
}

func createBranchName(key, summary string) string {
	slug := branchStripPattern.ReplaceAllString(summary, "")
	slug = strings.TrimFunc(slug, isJSSpace)
	slug = branchDashPattern.ReplaceAllString(slug, "-")

	return key + "-" + strings.ToLower(slug)
}

func jiraKeyFromBranchName(branchName string) string {
	match := jiraKeyPattern.FindStringSubmatch(branchName)
	if match == nil {
		return ""
	}

	return match[1]
}

func jiraSummaryFromBranchName(branchName string) string {
	match := jiraSummaryPattern.FindStringSubmatch(branchName)
	if match == nil {
		return ""
	}

	return upperFirst(strings.ReplaceAll(match[2], "-", " "))
}

func createTitleFromBranchName(branchName string) (string, error) {
	segments := nonEmpty(strings.Split(strings.TrimFunc(branchName, isJSSpace), "/"))

	var titlePrefix string
	for len(segments) > 1 {
		first := strings.ToLower(segments[0])
		if !strippableBranchPrefixes[first] {
			break
		}
		if titlePrefix == "" {
			titlePrefix = first
		}
		segments = segments[1:]
	}

	title := strings.Join(nonEmpty(titleSplitPattern.Split(strings.Join(segments, "/"), -1)), " ")
	if title == "" {
		return "", errors.New("No title found.")
	}

	title = upperFirst(title)
	if titlePrefix != "" {
		return titlePrefix + ": " + title, nil
	}

	return title, nil
}

func isJSSpace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', 0x2028, 0x2029, 0xFEFF:
		return true
	}

	return unicode.Is(unicode.Zs, r)
}

func upperFirst(s string) string {
	first, size := utf8.DecodeRuneInString(s)
	if size == 0 {
		return s
	}

	return string(unicode.ToUpper(first)) + s[size:]
}

func nonEmpty(values []string) []string {
	out := values[:0:0]
	for _, value := range values {
		if value != "" {
			out = append(out, value)
		}
	}

	return out
}

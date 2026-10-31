package domain

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// ToolContentResult describes semantic diagnostics independently of execution status.
type ToolContentResult struct {
	HasContentError bool
	Severity        FailureSeverity
	Category        string
	ErrorSnippet    *string
}

var contentPatterns = []struct {
	category string
	pattern  *regexp.Regexp
}{
	{"traceback", regexp.MustCompile(`(?m)^\s*Traceback \(most recent call last\):`)},
	{"panic", regexp.MustCompile(`(?m)^\s*(?:panic:|fatal error:)\s+\S`)},
	{"exception", regexp.MustCompile(`(?m)^\s*(?:UnhandledPromiseRejection(?:Warning)?|Uncaught (?:\w*Error|Exception)|Unhandled exception|(?:TypeError|ReferenceError|AssertionError):)`)},
	{"compiler", regexp.MustCompile(`(?im)^\s*(?:(?:[^\s]+:\d+(?::\d+)?:?|[^\r\n]+?\(\d+,\d+\):)\s*)?(?:error(?:\s+TS\d+)?\s*:|syntax error\b|cannot find package\b|build failed\b|compilation failed\b)`)},
	{"test", regexp.MustCompile(`(?m)^\s*(?:--- FAIL:|FAIL\s+\S|AssertionError\b|FAILURES\s*$)`)},
	{"stack_trace", regexp.MustCompile(`(?m)^\s+at\s+\S.*\([^\n]+:\d+:\d+\)\s*$`)},
}

// ClassifyToolContent recognizes diagnostic output without altering execution reliability.
func ClassifyToolContent(toolName, output string) ToolContentResult {
	name := strings.ToLower(strings.TrimSpace(toolName))
	for _, inspection := range []string{"read", "read_file", "readfile", "grep", "glob", "search", "web", "webfetch", "websearch"} {
		if name == inspection || strings.HasSuffix(name, "."+inspection) {
			return ToolContentResult{Severity: SeverityNone}
		}
	}
	for _, rule := range contentPatterns {
		if loc := rule.pattern.FindStringIndex(output); loc != nil {
			end := loc[0] + 2048
			if end > len(output) {
				end = len(output)
			}
			for end > loc[0] && end < len(output) && !utf8.RuneStart(output[end]) {
				end--
			}
			snippet := strings.TrimSpace(output[loc[0]:end])
			return ToolContentResult{HasContentError: true, Severity: SeverityMedium, Category: rule.category, ErrorSnippet: &snippet}
		}
	}
	return ToolContentResult{Severity: SeverityNone}
}

// ClassifyCompletedToolContent excludes failed and unfinished invocations from content totals.
func ClassifyCompletedToolContent(toolName, status, output string, executionError bool) ToolContentResult {
	if executionError || (status != "completed" && status != "success") {
		return ToolContentResult{Severity: SeverityNone}
	}
	return ClassifyToolContent(toolName, output)
}

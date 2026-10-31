package domain

import "testing"

func TestClassifyToolContentSeparatesSuccessfulDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		name, tool, output, category string
		want                         bool
	}{
		{"compiler", "bash", "main.go:3:2: syntax error: unexpected name", "compiler", true},
		{"typescript", "exec_command", "src/app.ts(3,2): error TS2345: bad argument", "compiler", true},
		{"traceback", "bash", "Traceback (most recent call last):\n  File \"x.py\", line 2\nValueError: invalid", "traceback", true},
		{"panic", "bash", "panic: invalid pointer", "panic", true},
		{"rejection", "bash", "UnhandledPromiseRejectionWarning: failed", "exception", true},
		{"assertion", "bash", "--- FAIL: TestBuild (0.00s)", "test", true},
		{"stack", "bash", "    at run (/tmp/app.js:12:3)", "stack_trace", true},
		{"ordinary prose", "bash", "This checks syntax error handling", "", false},
		{"inspection", "read", "panic: source example", "", false},
		{"search", "grep", "error: documentation example", "", false},
		{"clean", "bash", "all tests passed", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ClassifyToolContent(tc.tool, tc.output)
			if got.HasContentError != tc.want || got.Category != tc.category {
				t.Fatalf("classification=%+v", got)
			}
			execution := ClassifyToolExecution(tc.tool, "completed", tc.output, 0)
			if execution.IsError {
				t.Fatalf("content changed execution: %+v", execution)
			}
			if tc.want && (got.ErrorSnippet == nil || got.Severity == SeverityNone) {
				t.Fatalf("missing diagnostic: %+v", got)
			}
			if failed := ClassifyCompletedToolContent(tc.tool, "error", tc.output, true); failed.HasContentError {
				t.Fatal("failed call counted as successful content error")
			}
		})
	}
}

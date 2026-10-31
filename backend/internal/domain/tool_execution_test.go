package domain

import "testing"

func TestClassifyToolExecution(t *testing.T) {
	cases := []struct {
		name, tool, status, output string
		streak                     int
		wantError                  bool
		wantSeverity               FailureSeverity
	}{
		{"ordinary completion", "bash", "completed", "syntax error in a file", 0, false, SeverityNone},
		{"explicit failure", "read", "error", "", 0, true, SeverityMedium},
		{"nonzero process exit", "bash", "completed", "exit status 2", 0, true, SeverityMedium},
		{"grep no match exit", "grep", "completed", "exit status 1", 0, false, SeverityNone},
		{"glob no matches", "glob", "completed", "0 matches", 0, false, SeverityNone},
		{"timeout", "bash", "completed", "command timed out", 0, true, SeverityMedium},
		{"process crash", "bash", "completed", "segmentation fault", 0, true, SeverityHigh},
		{"escalation", "edit", "failed", "denied", 2, true, SeverityHigh},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ClassifyToolExecution(tc.tool, tc.status, tc.output, tc.streak)
			if got.IsError != tc.wantError || got.Severity != tc.wantSeverity {
				t.Fatalf("got error=%v severity=%s, want error=%v severity=%s", got.IsError, got.Severity, tc.wantError, tc.wantSeverity)
			}
		})
	}
}

func TestClassifyToolExecutionCapsErrorMessage(t *testing.T) {
	result := ClassifyToolExecution("bash", "error", string(make([]byte, 4096)), 0)
	if !result.IsError || result.ErrorMessage == nil || len(*result.ErrorMessage) != 2048 {
		t.Fatalf("expected an error message capped at 2048 characters, got %#v", result.ErrorMessage)
	}
}

func TestClassifyToolContent(t *testing.T) {
	for _, tc := range []struct {
		name, tool, output string
		want               bool
	}{
		{"compiler", "bash", "main.go:12:4: syntax error: unexpected token", true},
		{"typescript", "exec_command", "error TS2322: incompatible types", true},
		{"traceback", "bash", "Traceback (most recent call last):\n  File test.py", true},
		{"panic", "bash", "panic: invariant violated", true},
		{"rejection", "bash", "UnhandledPromiseRejectionWarning: missing await", true},
		{"test assertion", "bash", "--- FAIL: TestInvariant (0.00s)", true},
		{"inspection", "read", "panic: invariant violated", false},
		{"search", "grep", "error: copied diagnostic", false},
		{"ordinary prose", "bash", "No syntax errors were found", false},
		{"clean", "bash", "all checks passed", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ClassifyToolContent(tc.tool, tc.output)
			if got.HasContentError != tc.want {
				t.Fatalf("got %+v, want content error %v", got, tc.want)
			}
			if tc.want {
				if got.ErrorSnippet == nil || got.Category == "" || got.Severity == SeverityNone {
					t.Fatalf("missing diagnostic evidence: %+v", got)
				}
				if execution := ClassifyToolExecution(tc.tool, "completed", tc.output, 0); execution.IsError {
					t.Fatalf("content altered execution: %+v", execution)
				}
			}
			if got := ClassifyCompletedToolContent(tc.tool, "error", tc.output, true); got.HasContentError {
				t.Fatal("failed execution counted as content warning")
			}
		})
	}
}

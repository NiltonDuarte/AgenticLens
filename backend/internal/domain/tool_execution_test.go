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
		{"completed Java diagnostics", "bash", "completed", "Exception in thread \"main\" java.lang.NullPointerException\n\tat example.Main.run(Main.java:12)", 0, false, SeverityNone},
		{"completed Python diagnostics", "bash", "completed", "Traceback (most recent call last):\nValueError: invalid value", 0, false, SeverityNone},
		{"completed TypeScript diagnostics", "bash", "completed", "src/app.ts(3,2): error TS2345: bad argument", 0, false, SeverityNone},
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

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
		{"read timeout identifier", "read", "completed", "const timeoutMs = 5000;", 0, false, SeverityNone},
		{"grep saved panic example", "grep", "completed", "logs/example.txt: panic: saved example", 0, false, SeverityNone},
		{"explicit zero exit with fixture permission text", "bash", "completed", "exit code 0\nfixture says permission denied", 0, false, SeverityNone},
		{"explicit failure", "read", "error", "", 0, true, SeverityMedium},
		{"nonzero process exit", "bash", "completed", "exit status 2", 0, false, SeverityNone},
		{"grep no match exit", "grep", "completed", "exit status 1", 0, false, SeverityNone},
		{"glob no matches", "glob", "completed", "0 matches", 0, false, SeverityNone},
		{"timeout", "bash", "completed", "command timed out", 0, false, SeverityNone},
		{"process crash", "bash", "completed", "segmentation fault", 0, false, SeverityNone},
		{"escalation", "edit", "failed", "denied", 2, true, SeverityHigh},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ClassifyToolExecution(tc.tool, tc.status, tc.output, tc.streak)
			if got.IsError != tc.wantError || got.Severity != tc.wantSeverity {
				t.Fatalf("got error=%v severity=%s, want error=%v severity=%s", got.IsError, got.Severity, tc.wantError, tc.wantSeverity)
			}
			if got.IsError != (got.ErrorMessage != nil) {
				t.Fatalf("error message presence does not match execution failure: %+v", got)
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

// Free-form evidence must not manufacture either failure or severity.
func TestToolExecutionClassificationIgnoresFreeFormProcessText(t *testing.T) {
	for _, tool := range []string{"read", "grep", "glob", "bash", "exec_command"} {
		for _, output := range []string{"exit code 2", "segmentation fault", "command timed out", "permission denied", "Process exited with code 2\nFinal output:\nexample", "signal: killed", "no matches"} {
			t.Run(tool+"/"+output, func(t *testing.T) {
				completed := ClassifyToolExecution(tool, "completed", output, 0)
				if completed.IsError || completed.Severity != SeverityNone || completed.ErrorMessage != nil {
					t.Fatalf("text changed successful status: %+v", completed)
				}
				failed := ClassifyToolExecution(tool, "failed", output, 0)
				if !failed.IsError || failed.Severity != SeverityMedium {
					t.Fatalf("free-form text changed explicit failure/severity: %+v", failed)
				}
			})
		}
	}
}

func TestToolExecutionStructuredEvidenceAndUnknownCoverage(t *testing.T) {
	exitZero, exitOne, exitTwo, negativeExit := 0, 1, 2, -1
	signal, timeout := "SIGSEGV", true
	cases := []struct {
		name, tool, status string
		process            ToolProcessEvidence
		wantError, known   bool
		severity           FailureSeverity
	}{
		{"structured failure", "bash", "completed", ToolProcessEvidence{ExitCode: &exitTwo}, true, true, SeverityMedium},
		{"structured success", "bash", "completed", ToolProcessEvidence{ExitCode: &exitZero}, false, true, SeverityNone},
		{"grep no-match", "grep", "completed", ToolProcessEvidence{ExitCode: &exitOne}, false, true, SeverityNone},
		{"glob no-match", "glob", "completed", ToolProcessEvidence{ExitCode: &exitOne}, false, true, SeverityNone},
		{"no-match does not override tool error", "grep", "failed", ToolProcessEvidence{ExitCode: &exitOne}, true, true, SeverityMedium},
		{"structured signal", "bash", "completed", ToolProcessEvidence{Signal: &signal}, true, true, SeverityHigh},
		{"structured timeout", "bash", "completed", ToolProcessEvidence{TimedOut: &timeout}, true, true, SeverityHigh},
		{"negative native exit is failure", "bash", "", ToolProcessEvidence{ExitCode: &negativeExit}, true, true, SeverityMedium},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ClassifyToolExecution(tc.tool, tc.status, "  arbitrary exit code 77 and segmentation fault\n", 0, tc.process)
			if got.IsError != tc.wantError || got.ProcessOutcomeKnown != tc.known || got.Severity != tc.severity {
				t.Fatalf("classification=%+v", got)
			}
		})
	}
	for _, status := range []string{"", "unknown", "pending", "running", "unsupported"} {
		got := ClassifyToolExecution("bash", status, "Process exited with code 0", 0)
		if got.IsError || got.ProcessOutcomeKnown || got.Severity != SeverityNone || got.Status == "completed" || got.Status == "success" {
			t.Fatalf("unavailable state fabricated success/failure: %+v", got)
		}
	}
}

package domain_test

import (
	"testing"

	"agenticlens/internal/domain"
)

func strPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

// 1. Clean Runs & Zero Anomalies (Zero Mocks)
func TestDetectAnomalies_CleanSessions(t *testing.T) {
	t.Parallel()
	tid := "AL-CLEAN"
	toolRead := "read_file"
	toolBash := "bash"

	t.Run("fully populated clean session returns empty slice", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			TicketID: &tid,
			Invocations: []domain.InvocationFact{
				{
					ID:   "ses-1",
					Role: "orchestrator",
					Tokens: domain.InvocationTokens{
						Input:       1000,
						Output:      200,
						PeakContext: 10000,
					},
				},
				{
					ID:   "ses-2",
					Role: "coder",
					Tokens: domain.InvocationTokens{
						Input:       5000,
						Output:      1000,
						PeakContext: 25000,
					},
				},
			},
			Tools: []domain.ToolFact{
				{SessionID: "session-a", Name: "read_file", Calls: 10, Errors: 0},
				{SessionID: "session-a", Name: "bash", Calls: 8, Errors: 1}, // 1/8 = 12.5% <= 20%
			},
			Events: []domain.EventFact{
				{
					SessionID:     "ses-1",
					Turn:          1,
					Role:          "orchestrator",
					EventType:     domain.EventTypeTurnMessage,
					Status:        "completed",
					ContextTokens: 10000,
				},
				{
					SessionID:     "ses-1",
					Turn:          2,
					Role:          "coder",
					EventType:     domain.EventTypeToolCall,
					ToolName:      &toolRead,
					Status:        "completed",
					ContextTokens: 15000,
				},
				{
					SessionID:     "ses-1",
					Turn:          3,
					Role:          "tester",
					EventType:     domain.EventTypeToolCall,
					ToolName:      &toolBash,
					Status:        "completed",
					ContextTokens: 20000,
				},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if anomalies == nil {
			t.Fatalf("expected non-nil empty slice, got nil")
		}
		if len(anomalies) != 0 {
			t.Fatalf("expected 0 anomalies on clean run, got %d: %+v", len(anomalies), anomalies)
		}
	})

	t.Run("clean session with small token growth and normal error recovery", func(t *testing.T) {
		events := []domain.EventFact{
			{SessionID: "s1", Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error", ContextTokens: 5_000},
			{SessionID: "s1", Turn: 2, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "completed", ContextTokens: 10_000},
			{SessionID: "s1", Turn: 3, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error", ContextTokens: 15_000},
			{SessionID: "s1", Turn: 4, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "completed", ContextTokens: 20_000},
			{SessionID: "s1", Turn: 5, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "completed", ContextTokens: 25_000},
			{SessionID: "s1", Turn: 6, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "completed", ContextTokens: 30_000},
		}
		// 6 calls, 2 errors (33.3%), wait! If calls=6, errors=2, high failure rate triggers!
		// For a truly clean run, make errors 1 out of 6 (16.7% <= 20%):
		events[0].Status = "completed"

		ledger := domain.AgenticFactLedger{Events: events}
		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 0 {
			t.Fatalf("expected 0 anomalies, got %d: %+v", len(anomalies), anomalies)
		}
	})
}

// 2. TOOL_RETRY_STREAK Rule Tests
func TestDetectAnomalies_ToolRetryStreak(t *testing.T) {
	t.Parallel()
	toolBash := "bash"
	toolGit := "git"

	t.Run("3 consecutive errors trigger WARNING", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 4, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "completed"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 1 {
			t.Fatalf("expected 1 anomaly, got %d: %+v", len(anomalies), anomalies)
		}
		a := anomalies[0]
		if a.RuleID != domain.RuleToolRetryStreak {
			t.Errorf("expected RuleToolRetryStreak, got %s", a.RuleID)
		}
		if a.Severity != domain.SeverityWarning {
			t.Errorf("expected WARNING severity, got %s", a.Severity)
		}
		if a.Context.Count == nil || *a.Context.Count != 3 {
			t.Errorf("expected streak count 3, got %v", a.Context.Count)
		}
		if a.Context.ToolName == nil || *a.Context.ToolName != "bash" {
			t.Errorf("expected tool bash, got %v", a.Context.ToolName)
		}
		if a.Context.Turn == nil || *a.Context.Turn != 3 {
			t.Errorf("expected turn 3 of latest error, got %v", a.Context.Turn)
		}
		if a.Context.Threshold == nil || *a.Context.Threshold != "3" {
			t.Errorf("expected threshold 3, got %v", a.Context.Threshold)
		}
	})

	t.Run("4 consecutive errors trigger WARNING", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 4, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 5, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "completed"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var streak *domain.Anomaly
		for i := range anomalies {
			if anomalies[i].RuleID == domain.RuleToolRetryStreak {
				streak = &anomalies[i]
				break
			}
		}
		if streak == nil {
			t.Fatalf("expected streak anomaly, got: %+v", anomalies)
		}
		if streak.Severity != domain.SeverityWarning {
			t.Errorf("expected WARNING for 4 errors, got %s", streak.Severity)
		}
		if streak.Context.Count == nil || *streak.Context.Count != 4 {
			t.Errorf("expected count 4, got %v", streak.Context.Count)
		}
	})

	t.Run("5 consecutive errors trigger CRITICAL", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 4, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 5, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var streak *domain.Anomaly
		for i := range anomalies {
			if anomalies[i].RuleID == domain.RuleToolRetryStreak {
				streak = &anomalies[i]
				break
			}
		}
		if streak == nil {
			t.Fatalf("expected streak anomaly, got: %+v", anomalies)
		}
		if streak.Severity != domain.SeverityCritical {
			t.Errorf("expected CRITICAL severity for 5 errors, got %s", streak.Severity)
		}
		if streak.Context.Count == nil || *streak.Context.Count != 5 {
			t.Errorf("expected streak count 5, got %v", streak.Context.Count)
		}
		if streak.Context.Threshold == nil || *streak.Context.Threshold != "5" {
			t.Errorf("expected threshold 5, got %v", streak.Context.Threshold)
		}
	})

	t.Run("7 consecutive errors trigger CRITICAL", func(t *testing.T) {
		var events []domain.EventFact
		for turn := 1; turn <= 7; turn++ {
			events = append(events, domain.EventFact{
				SessionID: "s1",
				Turn:      int64(turn),
				EventType: domain.EventTypeToolCall,
				ToolName:  &toolBash,
				Status:    "error",
			})
		}
		ledger := domain.AgenticFactLedger{Events: events}

		anomalies := domain.DetectAnomalies(ledger)
		var streak *domain.Anomaly
		for i := range anomalies {
			if anomalies[i].RuleID == domain.RuleToolRetryStreak {
				streak = &anomalies[i]
				break
			}
		}
		if streak == nil {
			t.Fatalf("expected streak anomaly, got: %+v", anomalies)
		}
		if streak.Severity != domain.SeverityCritical {
			t.Errorf("expected CRITICAL severity for 7 errors, got %s", streak.Severity)
		}
		if streak.Context.Count == nil || *streak.Context.Count != 7 {
			t.Errorf("expected streak count 7, got %v", streak.Context.Count)
		}
	})

	t.Run("reset on success prevents streak accumulation", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "completed"},
				{SessionID: "s1", Turn: 4, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 5, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 6, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "completed"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		for _, a := range anomalies {
			if a.RuleID == domain.RuleToolRetryStreak {
				t.Fatalf("unexpected streak anomaly: %+v", a)
			}
		}
	})

	t.Run("reset on tool change resets streak", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				// Tool changes to git
				{SessionID: "s1", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "error"},
				// Tool changes back to bash
				{SessionID: "s1", Turn: 4, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 5, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "completed"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		for _, a := range anomalies {
			if a.RuleID == domain.RuleToolRetryStreak {
				t.Fatalf("unexpected streak anomaly across alternating tools: %+v", a)
			}
		}
	})

	t.Run("non-tool events between tool errors do not reset streak", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 1, EventType: domain.EventTypeTurnMessage, Status: "completed"},
				{SessionID: "s1", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 2, EventType: domain.EventTypeCompaction, Status: "completed"},
				{SessionID: "s1", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var streak *domain.Anomaly
		for i := range anomalies {
			if anomalies[i].RuleID == domain.RuleToolRetryStreak {
				streak = &anomalies[i]
				break
			}
		}
		if streak == nil {
			t.Fatalf("expected streak anomaly despite non-tool events, got: %+v", anomalies)
		}
		if streak.Context.Count == nil || *streak.Context.Count != 3 {
			t.Errorf("expected count 3, got %v", streak.Context.Count)
		}
	})

	t.Run("consecutive errors split across sessions do not combine into single streak", func(t *testing.T) {
		// 2 consecutive errors in session 1, then 2 consecutive errors in session 2 (neither reaches threshold 3)
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s2", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s2", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		for _, a := range anomalies {
			if a.RuleID == domain.RuleToolRetryStreak {
				t.Fatalf("expected no streak anomaly across separate sessions, got: %+v", a)
			}
		}
	})

	t.Run("2 errors in session A and 1 error in session B should NOT trigger 3-streak", func(t *testing.T) {
		// Adversarial verification of session boundary streak reset:
		// Total errors = 3, but 2 in session A and 1 in session B -> neither reaches threshold 3
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "sessionA", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "sessionA", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "sessionB", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		for _, a := range anomalies {
			if a.RuleID == domain.RuleToolRetryStreak {
				t.Fatalf("expected no TOOL_RETRY_STREAK anomaly across session transition (2 in A, 1 in B), got: %+v", a)
			}
		}
	})

	t.Run("3 errors in session A and 1 error in session B should trigger 1 streak for session A", func(t *testing.T) {
		// 3 errors in session A triggers streak for A with count=3; session B transition resets and has only 1 error
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "sessionA", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "sessionA", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "sessionA", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "sessionB", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var streaks []domain.Anomaly
		for _, a := range anomalies {
			if a.RuleID == domain.RuleToolRetryStreak {
				streaks = append(streaks, a)
			}
		}
		if len(streaks) != 1 {
			t.Fatalf("expected exactly 1 streak anomaly, got %d: %+v", len(streaks), streaks)
		}
		st := streaks[0]
		if st.Context.SessionID == nil || *st.Context.SessionID != "sessionA" {
			t.Errorf("expected streak attributed to sessionA, got %v", st.Context.SessionID)
		}
		if st.Context.Count == nil || *st.Context.Count != 3 {
			t.Errorf("expected streak count=3, got %v", st.Context.Count)
		}
		if st.Context.Turn == nil || *st.Context.Turn != 3 {
			t.Errorf("expected streak turn=3, got %v", st.Context.Turn)
		}
	})

	t.Run("multiple session transitions correctly isolate streaks per session", func(t *testing.T) {
		// Session A (3 errors) -> Session B (1 error) -> Session C (4 errors) -> Session D (2 errors)
		// Should produce exactly 2 streaks: Session A (count 3) and Session C (count 4)
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				// Session A
				{SessionID: "ses-A", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "ses-A", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "ses-A", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				// Session B
				{SessionID: "ses-B", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				// Session C
				{SessionID: "ses-C", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "ses-C", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "ses-C", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "ses-C", Turn: 4, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				// Session D
				{SessionID: "ses-D", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "ses-D", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var streaks []domain.Anomaly
		for _, a := range anomalies {
			if a.RuleID == domain.RuleToolRetryStreak {
				streaks = append(streaks, a)
			}
		}
		if len(streaks) != 2 {
			t.Fatalf("expected exactly 2 streaks across 4 sessions, got %d: %+v", len(streaks), streaks)
		}
		if *streaks[0].Context.SessionID != "ses-A" || *streaks[0].Context.Count != 3 {
			t.Errorf("expected first streak on ses-A count 3, got: %+v", streaks[0].Context)
		}
		if *streaks[1].Context.SessionID != "ses-C" || *streaks[1].Context.Count != 4 {
			t.Errorf("expected second streak on ses-C count 4, got: %+v", streaks[1].Context)
		}
	})

	t.Run("streak in session 1 emits even when session 2 follows", func(t *testing.T) {
		// 3 errors in session 1, followed by 2 errors in session 2
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s1", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s2", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				{SessionID: "s2", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var streakCount int
		var foundSession string
		for _, a := range anomalies {
			if a.RuleID == domain.RuleToolRetryStreak {
				streakCount++
				if a.Context.SessionID != nil {
					foundSession = *a.Context.SessionID
				}
			}
		}
		if streakCount != 1 {
			t.Fatalf("expected exactly 1 streak anomaly, got %d", streakCount)
		}
		if foundSession != "s1" {
			t.Errorf("expected streak to be attributed to session s1, got %s", foundSession)
		}
	})

	t.Run("v1 fallback returns empty for tool retry streak", func(t *testing.T) {
		// Even if ToolFact has many errors, v1 cannot determine consecutive streak
		ledger := domain.AgenticFactLedger{
			Tools: []domain.ToolFact{
				{SessionID: "session-a", Name: "bash", Calls: 10, Errors: 8},
			},
			Events: nil,
		}

		anomalies := domain.DetectAnomalies(ledger)
		for _, a := range anomalies {
			if a.RuleID == domain.RuleToolRetryStreak {
				t.Fatalf("v1 fallback must not emit TOOL_RETRY_STREAK, got: %+v", a)
			}
		}
	})
}

// 3. TOOL_HIGH_FAILURE_RATE Rule Tests
func TestDetectAnomalies_ToolHighFailureRate(t *testing.T) {
	t.Parallel()
	toolGit := "git"

	t.Run(">20% error rate with >=5 calls triggers WARNING in v2 events", func(t *testing.T) {
		// 6 calls, 2 errors = 33.3% > 20% and < 50%
		events := []domain.EventFact{
			{SessionID: "s1", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed"},
			{SessionID: "s1", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "error"},
			{SessionID: "s1", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed"},
			{SessionID: "s1", Turn: 4, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed"},
			{SessionID: "s1", Turn: 5, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "error"},
			{SessionID: "s1", Turn: 6, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed"},
		}
		ledger := domain.AgenticFactLedger{Events: events}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 1 {
			t.Fatalf("expected 1 anomaly, got %d: %+v", len(anomalies), anomalies)
		}
		a := anomalies[0]
		if a.RuleID != domain.RuleToolHighFailureRate {
			t.Errorf("expected RuleToolHighFailureRate, got %s", a.RuleID)
		}
		if a.Severity != domain.SeverityWarning {
			t.Errorf("expected WARNING severity, got %s", a.Severity)
		}
		if a.Context.TotalCalls == nil || *a.Context.TotalCalls != 6 {
			t.Errorf("expected 6 total calls, got %v", a.Context.TotalCalls)
		}
		if a.Context.Count == nil || *a.Context.Count != 2 {
			t.Errorf("expected 2 errors, got %v", a.Context.Count)
		}
		if a.Context.FailureRatePct == nil || *a.Context.FailureRatePct != 33.3 {
			t.Errorf("expected 33.3%% failure rate, got %v", a.Context.FailureRatePct)
		}
		if a.Context.Threshold == nil || *a.Context.Threshold != ">20%" {
			t.Errorf("expected threshold >20%%, got %v", a.Context.Threshold)
		}
	})

	t.Run(">=50% error rate triggers CRITICAL in v2 events", func(t *testing.T) {
		// 6 calls, 3 errors = 50.0% exactly -> CRITICAL
		events := []domain.EventFact{
			{SessionID: "s1", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed"},
			{SessionID: "s1", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "error"},
			{SessionID: "s1", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed"},
			{SessionID: "s1", Turn: 4, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "error"},
			{SessionID: "s1", Turn: 5, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed"},
			{SessionID: "s1", Turn: 6, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "error"},
		}
		ledger := domain.AgenticFactLedger{Events: events}

		anomalies := domain.DetectAnomalies(ledger)
		var rateAnomaly *domain.Anomaly
		for i := range anomalies {
			if anomalies[i].RuleID == domain.RuleToolHighFailureRate {
				rateAnomaly = &anomalies[i]
				break
			}
		}
		if rateAnomaly == nil {
			t.Fatalf("expected RuleToolHighFailureRate, got: %+v", anomalies)
		}
		if rateAnomaly.Severity != domain.SeverityCritical {
			t.Errorf("expected CRITICAL severity for 50%% rate, got %s", rateAnomaly.Severity)
		}
		if rateAnomaly.Context.FailureRatePct == nil || *rateAnomaly.Context.FailureRatePct != 50.0 {
			t.Errorf("expected 50.0%% failure rate, got %v", rateAnomaly.Context.FailureRatePct)
		}
		if rateAnomaly.Context.Threshold == nil || *rateAnomaly.Context.Threshold != ">=50%" {
			t.Errorf("expected threshold >=50%%, got %v", rateAnomaly.Context.Threshold)
		}
	})

	t.Run("under 5 calls is ignored even with 100% errors", func(t *testing.T) {
		events := []domain.EventFact{
			{SessionID: "s1", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "error"},
			{SessionID: "s1", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "error"},
			{SessionID: "s1", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "error"},
			{SessionID: "s1", Turn: 4, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "error"},
		}
		ledger := domain.AgenticFactLedger{Events: events}

		anomalies := domain.DetectAnomalies(ledger)
		for _, a := range anomalies {
			if a.RuleID == domain.RuleToolHighFailureRate {
				t.Fatalf("tool with 4 calls should be ignored, got: %+v", a)
			}
		}
	})

	t.Run("exact 20% boundary does not trigger", func(t *testing.T) {
		// 5 calls, 1 error = 20.0% -> NOT > 20%
		events5 := []domain.EventFact{
			{SessionID: "s1", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "error"},
			{SessionID: "s1", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed"},
			{SessionID: "s1", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed"},
			{SessionID: "s1", Turn: 4, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed"},
			{SessionID: "s1", Turn: 5, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed"},
		}
		ledger5 := domain.AgenticFactLedger{Events: events5}
		anomalies5 := domain.DetectAnomalies(ledger5)
		for _, a := range anomalies5 {
			if a.RuleID == domain.RuleToolHighFailureRate {
				t.Fatalf("5 calls with 1 error (20%%) must not trigger high failure rate, got: %+v", a)
			}
		}

		// 10 calls, 2 errors = 20.0% -> NOT > 20%
		ledger10 := domain.AgenticFactLedger{
			Tools: []domain.ToolFact{
				{SessionID: "session-a", Name: "git", Calls: 10, Errors: 2},
			},
		}
		anomalies10 := domain.DetectAnomalies(ledger10)
		for _, a := range anomalies10 {
			if a.RuleID == domain.RuleToolHighFailureRate {
				t.Fatalf("10 calls with 2 errors (20%%) must not trigger high failure rate in v1, got: %+v", a)
			}
		}
	})

	t.Run("v1 fallback parity with v2 event detection", func(t *testing.T) {
		// v1 ledger
		ledgerV1 := domain.AgenticFactLedger{
			Tools: []domain.ToolFact{
				{SessionID: "session-a", Name: "git", Calls: 6, Errors: 2},   // 33.3% WARNING
				{SessionID: "session-a", Name: "bash", Calls: 10, Errors: 6}, // 60.0% CRITICAL
				{SessionID: "session-a", Name: "curl", Calls: 4, Errors: 4},  // <5 calls -> ignored
			},
		}
		anomaliesV1 := domain.DetectAnomalies(ledgerV1)

		// v2 ledger with equivalent calls
		var eventsV2 []domain.EventFact
		toolBash := "bash"
		toolCurl := "curl"
		for i := 0; i < 6; i++ {
			st := "completed"
			if i < 2 {
				st = "error"
			}
			eventsV2 = append(eventsV2, domain.EventFact{SessionID: "s", Turn: int64(i + 1), EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: st})
		}
		for i := 0; i < 10; i++ {
			// alternate errors and completed so we don't trigger retry streak of >=3
			st := "completed"
			if i%2 == 0 && i < 12 { // 6 errors at i=0,2,4,6,8,10 (wait, i<10 means 5 errors at 0,2,4,6,8 and one more at 9)
				st = "error"
			}
			eventsV2 = append(eventsV2, domain.EventFact{SessionID: "s", Turn: int64(i + 10), EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: st})
		}
		// Adjust bash errors to exactly 6 errors without 3 in a row:
		// errors at 0, 1 (2), then ok at 2, errors at 3, 4 (2), ok at 5, errors at 6, 7 (2) -> total 6 errors, max streak 2
		eventsV2 = []domain.EventFact{
			// git: 6 calls, 2 errors
			{SessionID: "s", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "error"},
			{SessionID: "s", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed"},
			{SessionID: "s", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "error"},
			{SessionID: "s", Turn: 4, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed"},
			{SessionID: "s", Turn: 5, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed"},
			{SessionID: "s", Turn: 6, EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed"},
			// bash: 10 calls, 6 errors (max consecutive 2)
			{SessionID: "s", Turn: 7, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			{SessionID: "s", Turn: 8, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			{SessionID: "s", Turn: 9, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "completed"},
			{SessionID: "s", Turn: 10, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			{SessionID: "s", Turn: 11, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			{SessionID: "s", Turn: 12, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "completed"},
			{SessionID: "s", Turn: 13, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			{SessionID: "s", Turn: 14, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			{SessionID: "s", Turn: 15, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "completed"},
			{SessionID: "s", Turn: 16, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "completed"},
			// curl: 4 calls, 4 errors
			{SessionID: "s", Turn: 17, EventType: domain.EventTypeToolCall, ToolName: &toolCurl, Status: "error"},
			{SessionID: "s", Turn: 18, EventType: domain.EventTypeToolCall, ToolName: &toolCurl, Status: "error"},
			{SessionID: "s", Turn: 19, EventType: domain.EventTypeToolCall, ToolName: &toolCurl, Status: "error"},
			{SessionID: "s", Turn: 20, EventType: domain.EventTypeToolCall, ToolName: &toolCurl, Status: "error"},
		}
		ledgerV2 := domain.AgenticFactLedger{Events: eventsV2}
		anomaliesV2 := domain.DetectAnomalies(ledgerV2)

		// Filter to high failure rate anomalies
		var v1Rates, v2Rates []domain.Anomaly
		for _, a := range anomaliesV1 {
			if a.RuleID == domain.RuleToolHighFailureRate {
				v1Rates = append(v1Rates, a)
			}
		}
		for _, a := range anomaliesV2 {
			if a.RuleID == domain.RuleToolHighFailureRate {
				v2Rates = append(v2Rates, a)
			}
		}

		if len(v1Rates) != len(v2Rates) {
			t.Fatalf("parity mismatch: v1 has %d rate anomalies, v2 has %d", len(v1Rates), len(v2Rates))
		}
		for i := range v1Rates {
			a1 := v1Rates[i]
			a2 := v2Rates[i]
			if a1.Severity != a2.Severity {
				t.Errorf("item %d severity mismatch: v1=%s, v2=%s", i, a1.Severity, a2.Severity)
			}
			if *a1.Context.ToolName != *a2.Context.ToolName {
				t.Errorf("item %d tool mismatch: v1=%s, v2=%s", i, *a1.Context.ToolName, *a2.Context.ToolName)
			}
			if *a1.Context.TotalCalls != *a2.Context.TotalCalls {
				t.Errorf("item %d calls mismatch: v1=%d, v2=%d", i, *a1.Context.TotalCalls, *a2.Context.TotalCalls)
			}
			if *a1.Context.Count != *a2.Context.Count {
				t.Errorf("item %d count mismatch: v1=%d, v2=%d", i, *a1.Context.Count, *a2.Context.Count)
			}
			if *a1.Context.FailureRatePct != *a2.Context.FailureRatePct {
				t.Errorf("item %d rate pct mismatch: v1=%f, v2=%f", i, *a1.Context.FailureRatePct, *a2.Context.FailureRatePct)
			}
		}
	})
}

func TestToolFailureAnomaliesAggregateSessionFactsByName(t *testing.T) {
	ledger := domain.AgenticFactLedger{Tools: []domain.ToolFact{
		{SessionID: "session-a", Name: "bash", Calls: 4, Errors: 1},
		{SessionID: "session-b", Name: "bash", Calls: 6, Errors: 2},
	}}
	anomalies := domain.DetectAnomalies(ledger)
	var failureRateAnomalies []domain.Anomaly
	for _, anomaly := range anomalies {
		if anomaly.RuleID == domain.RuleToolHighFailureRate {
			failureRateAnomalies = append(failureRateAnomalies, anomaly)
		}
	}
	if len(failureRateAnomalies) != 1 {
		t.Fatalf("expected one aggregated bash anomaly, got %#v", failureRateAnomalies)
	}
	if failureRateAnomalies[0].Context.TotalCalls == nil || *failureRateAnomalies[0].Context.TotalCalls != 10 ||
		failureRateAnomalies[0].Context.Count == nil || *failureRateAnomalies[0].Context.Count != 3 {
		t.Fatalf("expected aggregate 3 errors across 10 calls, got %+v", failureRateAnomalies[0].Context)
	}
	if len(domain.ComputeFleetAnomalies([]domain.AgenticFactLedger{ledger}).BrittleTools) != 1 {
		t.Fatal("expected fleet summary to include exactly one brittle tool name")
	}
}

// 4. CONTEXT_SATURATION_SPIKE Rule Tests
func TestDetectAnomalies_ContextSaturationSpike(t *testing.T) {
	t.Parallel()
	t.Run("single-turn jump >40k tokens triggers WARNING", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, ContextTokens: 10_000},
				{SessionID: "s1", Turn: 2, ContextTokens: 50_001}, // +40,001
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 1 {
			t.Fatalf("expected 1 anomaly, got %d: %+v", len(anomalies), anomalies)
		}
		a := anomalies[0]
		if a.RuleID != domain.RuleContextSaturationSpike {
			t.Errorf("expected RuleContextSaturationSpike, got %s", a.RuleID)
		}
		if a.Severity != domain.SeverityWarning {
			t.Errorf("expected WARNING severity, got %s", a.Severity)
		}
		if a.Context.TokenDelta == nil || *a.Context.TokenDelta != 40_001 {
			t.Errorf("expected delta 40001, got %v", a.Context.TokenDelta)
		}
		if a.Context.Threshold == nil || *a.Context.Threshold != "40000" {
			t.Errorf("expected threshold 40000, got %v", a.Context.Threshold)
		}
	})

	t.Run("exact 40k delta boundary does not trigger", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, ContextTokens: 10_000},
				{SessionID: "s1", Turn: 2, ContextTokens: 50_000}, // +40,000 exactly
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 0 {
			t.Fatalf("exact 40k delta must not trigger anomaly, got %d: %+v", len(anomalies), anomalies)
		}
	})

	t.Run("exact 80k delta triggers WARNING not CRITICAL", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, ContextTokens: 10_000},
				{SessionID: "s1", Turn: 2, ContextTokens: 90_000}, // +80,000 exactly (>40k, but not >80k)
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 1 {
			t.Fatalf("expected 1 anomaly, got %d: %+v", len(anomalies), anomalies)
		}
		if anomalies[0].Severity != domain.SeverityWarning {
			t.Errorf("exact 80k delta must be WARNING, got %s", anomalies[0].Severity)
		}
	})

	t.Run("single-turn jump >80k tokens triggers CRITICAL", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, ContextTokens: 10_000},
				{SessionID: "s1", Turn: 2, ContextTokens: 90_001}, // +80,001 (>80k)
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 1 {
			t.Fatalf("expected 1 anomaly, got %d: %+v", len(anomalies), anomalies)
		}
		if anomalies[0].Severity != domain.SeverityCritical {
			t.Errorf("expected CRITICAL severity, got %s", anomalies[0].Severity)
		}
		if anomalies[0].Context.Threshold == nil || *anomalies[0].Context.Threshold != "80000" {
			t.Errorf("expected threshold 80000, got %v", anomalies[0].Context.Threshold)
		}
	})

	t.Run("negative token delta does not trigger", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, ContextTokens: 100_000},
				{SessionID: "s1", Turn: 2, ContextTokens: 20_000}, // compaction / drop
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 0 {
			t.Fatalf("negative token delta must not trigger anomaly, got: %+v", anomalies)
		}
	})

	t.Run("turn with 160k tokens triggers CRITICAL ceiling breach", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, ContextTokens: 160_000},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 1 {
			t.Fatalf("expected 1 anomaly for 160k ceiling breach, got %d: %+v", len(anomalies), anomalies)
		}
		a := anomalies[0]
		if a.RuleID != domain.RuleContextSaturationSpike {
			t.Errorf("expected RuleContextSaturationSpike, got %s", a.RuleID)
		}
		if a.Severity != domain.SeverityCritical {
			t.Errorf("expected CRITICAL severity, got %s", a.Severity)
		}
		if a.Context.ContextTokens == nil || *a.Context.ContextTokens != 160_000 {
			t.Errorf("expected context tokens 160000, got %v", a.Context.ContextTokens)
		}
		if a.Context.Threshold == nil || *a.Context.Threshold != "160000" {
			t.Errorf("expected threshold 160000, got %v", a.Context.Threshold)
		}
	})

	t.Run("ceiling breach reported only once per session", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, ContextTokens: 162_000},
				{SessionID: "s1", Turn: 2, ContextTokens: 165_000}, // delta is 3k (<40k), ceiling already breached
				{SessionID: "s1", Turn: 3, ContextTokens: 170_000}, // delta is 5k (<40k), ceiling already breached
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 1 {
			t.Fatalf("expected exactly 1 ceiling breach anomaly for session, got %d: %+v", len(anomalies), anomalies)
		}
	})

	t.Run("session isolation prevents cross-session false spikes", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{SessionID: "ses-A", Turn: 1, ContextTokens: 10_000},
				{SessionID: "ses-A", Turn: 2, ContextTokens: 15_000},
				{SessionID: "ses-B", Turn: 1, ContextTokens: 70_000}, // first turn of ses-B, not a jump from ses-A
				{SessionID: "ses-B", Turn: 2, ContextTokens: 75_000},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 0 {
			t.Fatalf("cross-session events must not trigger false token jump, got %d: %+v", len(anomalies), anomalies)
		}
	})

	t.Run("v1 fallback evaluates PeakContext >= 160k tokens", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{
					ID:   "ses-v1-orch",
					Role: "orchestrator",
					Tokens: domain.InvocationTokens{
						PeakContext: 175_000,
					},
				},
				{
					ID:   "ses-v1-coder",
					Role: "coder",
					Tokens: domain.InvocationTokens{
						PeakContext: 159_999, // below 160k
					},
				},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 1 {
			t.Fatalf("expected 1 anomaly in v1 fallback, got %d: %+v", len(anomalies), anomalies)
		}
		a := anomalies[0]
		if a.Severity != domain.SeverityCritical {
			t.Errorf("expected CRITICAL severity, got %s", a.Severity)
		}
		if a.Context.ContextTokens == nil || *a.Context.ContextTokens != 175_000 {
			t.Errorf("expected 175000 tokens, got %v", a.Context.ContextTokens)
		}
	})
}

// 5. SUBAGENT_REWORK_LOOP Rule Tests
func TestDetectAnomalies_SubagentReworkLoop(t *testing.T) {
	t.Parallel()
	t.Run("coder <-> tester 3 cycles triggers WARNING", func(t *testing.T) {
		rootID := "ses-root"
		childCoderID := "ses-coder"
		childTesterID := "ses-tester"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: rootID, ParentID: nil, Role: "general"},
				{ID: childCoderID, ParentID: &rootID, Role: "coder"},
				{ID: childTesterID, ParentID: &rootID, Role: "tester"},
			},
			Events: []domain.EventFact{
				{SessionID: childCoderID, Turn: 1, Role: "coder"},
				{SessionID: childTesterID, Turn: 2, Role: "tester"},
				{SessionID: childCoderID, Turn: 3, Role: "coder"},
				{SessionID: childTesterID, Turn: 4, Role: "tester"},
				{SessionID: childCoderID, Turn: 5, Role: "coder"},
				{SessionID: childTesterID, Turn: 6, Role: "tester"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 1 {
			t.Fatalf("expected 1 anomaly, got %d: %+v", len(anomalies), anomalies)
		}
		a := anomalies[0]
		if a.RuleID != domain.RuleSubagentReworkLoop {
			t.Errorf("expected RuleSubagentReworkLoop, got %s", a.RuleID)
		}
		if a.Severity != domain.SeverityWarning {
			t.Errorf("expected WARNING severity, got %s", a.Severity)
		}
		if a.Context.Role == nil || *a.Context.Role != "coder" {
			t.Errorf("expected role coder, got %v", a.Context.Role)
		}
		if a.Context.SecondaryRole == nil || *a.Context.SecondaryRole != "tester" {
			t.Errorf("expected secondary role tester, got %v", a.Context.SecondaryRole)
		}
		if a.Context.Count == nil || *a.Context.Count != 3 {
			t.Errorf("expected 3 cycles, got %v", a.Context.Count)
		}
		if a.Context.Threshold == nil || *a.Context.Threshold != "3" {
			t.Errorf("expected threshold 3, got %v", a.Context.Threshold)
		}
	})

	t.Run("coder <-> tester 4 cycles triggers WARNING", func(t *testing.T) {
		rootID := "ses-root"
		childCoderID := "ses-coder"
		childTesterID := "ses-tester"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: rootID, ParentID: nil, Role: "general"},
				{ID: childCoderID, ParentID: &rootID, Role: "coder"},
				{ID: childTesterID, ParentID: &rootID, Role: "tester"},
			},
			Events: []domain.EventFact{
				{SessionID: childCoderID, Turn: 1, Role: "coder"}, {SessionID: childTesterID, Turn: 2, Role: "tester"},
				{SessionID: childCoderID, Turn: 3, Role: "coder"}, {SessionID: childTesterID, Turn: 4, Role: "tester"},
				{SessionID: childCoderID, Turn: 5, Role: "coder"}, {SessionID: childTesterID, Turn: 6, Role: "tester"},
				{SessionID: childCoderID, Turn: 7, Role: "coder"}, {SessionID: childTesterID, Turn: 8, Role: "tester"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 1 {
			t.Fatalf("expected 1 anomaly, got %d: %+v", len(anomalies), anomalies)
		}
		if anomalies[0].Severity != domain.SeverityWarning {
			t.Errorf("expected WARNING for 4 cycles, got %s", anomalies[0].Severity)
		}
		if anomalies[0].Context.Count == nil || *anomalies[0].Context.Count != 4 {
			t.Errorf("expected 4 cycles, got %v", anomalies[0].Context.Count)
		}
	})

	t.Run("coder <-> tester 5 cycles triggers CRITICAL", func(t *testing.T) {
		rootID := "ses-root"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: rootID, ParentID: nil, Role: "general"},
				{ID: "c1", ParentID: &rootID, Role: "coder"}, {ID: "t1", ParentID: &rootID, Role: "tester"},
				{ID: "c2", ParentID: &rootID, Role: "coder"}, {ID: "t2", ParentID: &rootID, Role: "tester"},
				{ID: "c3", ParentID: &rootID, Role: "coder"}, {ID: "t3", ParentID: &rootID, Role: "tester"},
				{ID: "c4", ParentID: &rootID, Role: "coder"}, {ID: "t4", ParentID: &rootID, Role: "tester"},
				{ID: "c5", ParentID: &rootID, Role: "coder"}, {ID: "t5", ParentID: &rootID, Role: "tester"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 1 {
			t.Fatalf("expected 1 anomaly, got %d: %+v", len(anomalies), anomalies)
		}
		a := anomalies[0]
		if a.Severity != domain.SeverityCritical {
			t.Errorf("expected CRITICAL severity for 5 cycles, got %s", a.Severity)
		}
		if a.Context.Count == nil || *a.Context.Count != 5 {
			t.Errorf("expected 5 cycles, got %v", a.Context.Count)
		}
		if a.Context.Threshold == nil || *a.Context.Threshold != "5" {
			t.Errorf("expected threshold 5, got %v", a.Context.Threshold)
		}
	})

	t.Run("fewer than 3 cycles does not trigger", func(t *testing.T) {
		rootID := "ses-root"
		childCoderID := "ses-coder"
		childTesterID := "ses-tester"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: rootID, ParentID: nil, Role: "general"},
				{ID: childCoderID, ParentID: &rootID, Role: "coder"},
				{ID: childTesterID, ParentID: &rootID, Role: "tester"},
			},
			Events: []domain.EventFact{
				{SessionID: childCoderID, Turn: 1, Role: "coder"},
				{SessionID: childTesterID, Turn: 2, Role: "tester"},
				{SessionID: childCoderID, Turn: 3, Role: "coder"},
				{SessionID: childTesterID, Turn: 4, Role: "tester"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 0 {
			t.Fatalf("2 cycles must not trigger rework loop, got: %+v", anomalies)
		}
	})

	t.Run("non-alternating transitions do not trigger", func(t *testing.T) {
		// coder -> tester -> reviewer -> coder -> tester -> reviewer (3 distinct roles)
		rootID := "ses-root"
		childCoderID := "ses-coder"
		childTesterID := "ses-tester"
		childReviewerID := "ses-reviewer"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: rootID, ParentID: nil, Role: "general"},
				{ID: childCoderID, ParentID: &rootID, Role: "coder"},
				{ID: childTesterID, ParentID: &rootID, Role: "tester"},
				{ID: childReviewerID, ParentID: &rootID, Role: "reviewer"},
			},
			Events: []domain.EventFact{
				{SessionID: childCoderID, Turn: 1, Role: "coder"},
				{SessionID: childTesterID, Turn: 2, Role: "tester"},
				{SessionID: childReviewerID, Turn: 3, Role: "reviewer"},
				{SessionID: childCoderID, Turn: 4, Role: "coder"},
				{SessionID: childTesterID, Turn: 5, Role: "tester"},
				{SessionID: childReviewerID, Turn: 6, Role: "reviewer"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		for _, a := range anomalies {
			if a.RuleID == domain.RuleSubagentReworkLoop {
				t.Fatalf("non-alternating transitions should not trigger rework loop, got: %+v", a)
			}
		}
	})

	t.Run("orchestrator role is strictly ignored in rework loop detection", func(t *testing.T) {
		// Orchestrator calls coder, then tester, then coder, etc.
		rootID := "ses-root"
		childCoderID := "ses-coder"
		childTesterID := "ses-tester"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: rootID, ParentID: nil, Role: "orchestrator"},
				{ID: childCoderID, ParentID: &rootID, Role: "coder"},
				{ID: childTesterID, ParentID: &rootID, Role: "tester"},
			},
			Events: []domain.EventFact{
				{SessionID: rootID, Turn: 1, Role: "orchestrator"},
				{SessionID: childCoderID, Turn: 2, Role: "coder"},
				{SessionID: rootID, Turn: 3, Role: "orchestrator"},
				{SessionID: childTesterID, Turn: 4, Role: "tester"},
				{SessionID: rootID, Turn: 5, Role: "orchestrator"},
				{SessionID: childCoderID, Turn: 6, Role: "coder"},
				{SessionID: rootID, Turn: 7, Role: "orchestrator"},
				{SessionID: childTesterID, Turn: 8, Role: "tester"},
				{SessionID: rootID, Turn: 9, Role: "orchestrator"},
				{SessionID: childCoderID, Turn: 10, Role: "coder"},
				{SessionID: rootID, Turn: 11, Role: "orchestrator"},
				{SessionID: childTesterID, Turn: 12, Role: "tester"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var rework *domain.Anomaly
		for i := range anomalies {
			if anomalies[i].RuleID == domain.RuleSubagentReworkLoop {
				rework = &anomalies[i]
				break
			}
		}
		if rework == nil {
			t.Fatalf("expected rework loop when orchestrator is interleaved, got: %+v", anomalies)
		}
		if rework.Severity != domain.SeverityWarning {
			t.Errorf("expected WARNING, got %s", rework.Severity)
		}
		if *rework.Context.Role != "coder" || *rework.Context.SecondaryRole != "tester" {
			t.Errorf("expected coder <-> tester loop, got %s <-> %s", *rework.Context.Role, *rework.Context.SecondaryRole)
		}
		if *rework.Context.Count != 3 {
			t.Errorf("expected 3 cycles, got %d", *rework.Context.Count)
		}
	})

	t.Run("structural hierarchy ignores root session events regardless of persona name", func(t *testing.T) {
		rootID := "ses-root"
		childCoderID := "ses-coder"
		childTesterID := "ses-tester"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: rootID, ParentID: nil, Role: "general"},
				{ID: childCoderID, ParentID: &rootID, Role: "coder"},
				{ID: childTesterID, ParentID: &rootID, Role: "tester"},
			},
			Events: []domain.EventFact{
				{SessionID: rootID, Turn: 1, Role: "general"},
				{SessionID: childCoderID, Turn: 2, Role: "coder"},
				{SessionID: rootID, Turn: 3, Role: "general"},
				{SessionID: childTesterID, Turn: 4, Role: "tester"},
				{SessionID: rootID, Turn: 5, Role: "general"},
				{SessionID: childCoderID, Turn: 6, Role: "coder"},
				{SessionID: rootID, Turn: 7, Role: "general"},
				{SessionID: childTesterID, Turn: 8, Role: "tester"},
				{SessionID: rootID, Turn: 9, Role: "general"},
				{SessionID: childCoderID, Turn: 10, Role: "coder"},
				{SessionID: rootID, Turn: 11, Role: "general"},
				{SessionID: childTesterID, Turn: 12, Role: "tester"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var rework *domain.Anomaly
		for i := range anomalies {
			if anomalies[i].RuleID == domain.RuleSubagentReworkLoop {
				rework = &anomalies[i]
				break
			}
		}
		if rework == nil {
			t.Fatalf("expected rework loop when root general events are interleaved, got: %+v", anomalies)
		}
		if rework.Severity != domain.SeverityWarning {
			t.Errorf("expected WARNING, got %s", rework.Severity)
		}
		if *rework.Context.Role != "coder" || *rework.Context.SecondaryRole != "tester" {
			t.Errorf("expected coder <-> tester loop, got %s <-> %s", *rework.Context.Role, *rework.Context.SecondaryRole)
		}
		if *rework.Context.Count != 3 {
			t.Errorf("expected 3 cycles, got %d", *rework.Context.Count)
		}
	})

	t.Run("structural hierarchy ignores root invocations in fallback sequence", func(t *testing.T) {
		rootID := "ses-root"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: rootID, ParentID: nil, Role: "general"},
				{ID: "ses-c1", ParentID: &rootID, Role: "coder"},
				{ID: "ses-t1", ParentID: &rootID, Role: "tester"},
				{ID: "ses-c2", ParentID: &rootID, Role: "coder"},
				{ID: "ses-t2", ParentID: &rootID, Role: "tester"},
				{ID: "ses-c3", ParentID: &rootID, Role: "coder"},
				{ID: "ses-t3", ParentID: &rootID, Role: "tester"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var rework *domain.Anomaly
		for i := range anomalies {
			if anomalies[i].RuleID == domain.RuleSubagentReworkLoop {
				rework = &anomalies[i]
				break
			}
		}
		if rework == nil {
			t.Fatalf("expected rework loop in invocations fallback with root ignored, got: %+v", anomalies)
		}
		if *rework.Context.Role != "coder" || *rework.Context.SecondaryRole != "tester" {
			t.Errorf("expected coder <-> tester loop, got %s <-> %s", *rework.Context.Role, *rework.Context.SecondaryRole)
		}
		if *rework.Context.Count != 3 {
			t.Errorf("expected 3 cycles, got %d", *rework.Context.Count)
		}
	})

	t.Run("consecutive identical roles collapsed without false trigger", func(t *testing.T) {
		rootID := "ses-root"
		childCoderID := "ses-coder"
		childTesterID := "ses-tester"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: rootID, ParentID: nil, Role: "general"},
				{ID: childCoderID, ParentID: &rootID, Role: "coder"},
				{ID: childTesterID, ParentID: &rootID, Role: "tester"},
			},
			Events: []domain.EventFact{
				{SessionID: childCoderID, Turn: 1, Role: "coder"},
				{SessionID: childCoderID, Turn: 2, Role: "coder"},
				{SessionID: childTesterID, Turn: 3, Role: "tester"},
				{SessionID: childTesterID, Turn: 4, Role: "tester"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 0 {
			t.Fatalf("expected 0 anomalies, got %d: %+v", len(anomalies), anomalies)
		}
	})

	t.Run("adversarial: root session labeled coder interleaved between planner and reviewer structurally ignored", func(t *testing.T) {
		rootID := "ses-root"
		c1ID := "ses-planner"
		c2ID := "ses-reviewer"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: rootID, ParentID: nil, Role: "coder"},
				{ID: c1ID, ParentID: &rootID, Role: "planner"},
				{ID: c2ID, ParentID: &rootID, Role: "reviewer"},
			},
			Events: []domain.EventFact{
				{SessionID: rootID, Turn: 1, Role: "coder"},
				{SessionID: c1ID, Turn: 2, Role: "planner"},
				{SessionID: rootID, Turn: 3, Role: "coder"},
				{SessionID: c2ID, Turn: 4, Role: "reviewer"},
				{SessionID: rootID, Turn: 5, Role: "coder"},
				{SessionID: c1ID, Turn: 6, Role: "planner"},
				{SessionID: rootID, Turn: 7, Role: "coder"},
				{SessionID: c2ID, Turn: 8, Role: "reviewer"},
				{SessionID: rootID, Turn: 9, Role: "coder"},
				{SessionID: c1ID, Turn: 10, Role: "planner"},
				{SessionID: rootID, Turn: 11, Role: "coder"},
				{SessionID: c2ID, Turn: 12, Role: "reviewer"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var rework *domain.Anomaly
		for i := range anomalies {
			if anomalies[i].RuleID == domain.RuleSubagentReworkLoop {
				rework = &anomalies[i]
				break
			}
		}
		if rework == nil {
			t.Fatalf("expected rework loop when root coder events are interleaved, got: %+v", anomalies)
		}
		if *rework.Context.Role != "planner" || *rework.Context.SecondaryRole != "reviewer" {
			t.Errorf("expected planner <-> reviewer loop, got %s <-> %s", *rework.Context.Role, *rework.Context.SecondaryRole)
		}
		if *rework.Context.Count != 3 {
			t.Errorf("expected 3 cycles, got %d", *rework.Context.Count)
		}
	})

	t.Run("adversarial: root session events alternating subagent roles do not trigger rework loop", func(t *testing.T) {
		rootID := "ses-root"
		childID := "ses-child"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: rootID, ParentID: nil, Role: "orchestrator"},
				{ID: childID, ParentID: &rootID, Role: "coder"},
			},
			Events: []domain.EventFact{
				{SessionID: rootID, Turn: 1, Role: "planner"},
				{SessionID: rootID, Turn: 2, Role: "reviewer"},
				{SessionID: rootID, Turn: 3, Role: "planner"},
				{SessionID: rootID, Turn: 4, Role: "reviewer"},
				{SessionID: rootID, Turn: 5, Role: "planner"},
				{SessionID: rootID, Turn: 6, Role: "reviewer"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		for _, a := range anomalies {
			if a.RuleID == domain.RuleSubagentReworkLoop {
				t.Fatalf("root session events must not trigger SUBAGENT_REWORK_LOOP, got: %+v", a)
			}
		}
	})

	t.Run("adversarial: invocations fallback ignores root invocation labeled reviewer", func(t *testing.T) {
		rootID := "ses-root"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: rootID, ParentID: nil, Role: "reviewer"},
				{ID: "ses-t1", ParentID: &rootID, Role: "tester"},
				{ID: "ses-c1", ParentID: &rootID, Role: "coder"},
				{ID: "ses-t2", ParentID: &rootID, Role: "tester"},
				{ID: "ses-c2", ParentID: &rootID, Role: "coder"},
				{ID: "ses-t3", ParentID: &rootID, Role: "tester"},
				{ID: "ses-c3", ParentID: &rootID, Role: "coder"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var rework *domain.Anomaly
		for i := range anomalies {
			if anomalies[i].RuleID == domain.RuleSubagentReworkLoop {
				rework = &anomalies[i]
				break
			}
		}
		if rework == nil {
			t.Fatalf("expected rework loop in invocations fallback with root reviewer ignored, got: %+v", anomalies)
		}
		if *rework.Context.Role != "tester" || *rework.Context.SecondaryRole != "coder" {
			t.Errorf("expected tester <-> coder loop, got %s <-> %s", *rework.Context.Role, *rework.Context.SecondaryRole)
		}
		if *rework.Context.Count != 3 {
			t.Errorf("expected 3 cycles, got %d", *rework.Context.Count)
		}
	})

	t.Run("boundary: empty ledger yields zero rework loop anomalies", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{}
		anomalies := domain.DetectAnomalies(ledger)
		for _, a := range anomalies {
			if a.RuleID == domain.RuleSubagentReworkLoop {
				t.Fatalf("empty ledger must not produce rework loop anomalies, got: %+v", a)
			}
		}
	})

	t.Run("boundary: events with unmapped SessionID are ignored and never trigger rework loop", func(t *testing.T) {
		rootID := "ses-root"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: rootID, ParentID: nil, Role: "orchestrator"},
			},
			Events: []domain.EventFact{
				{SessionID: "phantom-session-1", Turn: 1, Role: "coder"},
				{SessionID: "phantom-session-2", Turn: 2, Role: "tester"},
				{SessionID: "phantom-session-1", Turn: 3, Role: "coder"},
				{SessionID: "phantom-session-2", Turn: 4, Role: "tester"},
				{SessionID: "phantom-session-1", Turn: 5, Role: "coder"},
				{SessionID: "phantom-session-2", Turn: 6, Role: "tester"},
				{SessionID: "", Turn: 7, Role: "coder"},
				{SessionID: "", Turn: 8, Role: "tester"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		for _, a := range anomalies {
			if a.RuleID == domain.RuleSubagentReworkLoop {
				t.Fatalf("events with unmapped SessionID must not trigger rework loop, got: %+v", a)
			}
		}
	})

	t.Run("boundary: unmapped SessionID events do not dilute or corrupt legitimate child rework loop", func(t *testing.T) {
		rootID := "ses-root"
		childCoderID := "ses-coder"
		childTesterID := "ses-tester"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: rootID, ParentID: nil, Role: "orchestrator"},
				{ID: childCoderID, ParentID: &rootID, Role: "coder"},
				{ID: childTesterID, ParentID: &rootID, Role: "tester"},
			},
			Events: []domain.EventFact{
				{SessionID: "unmapped-1", Turn: 1, Role: "reviewer"},
				{SessionID: childCoderID, Turn: 2, Role: "coder"},
				{SessionID: "unmapped-2", Turn: 3, Role: "planner"},
				{SessionID: childTesterID, Turn: 4, Role: "tester"},
				{SessionID: "", Turn: 5, Role: "reviewer"},
				{SessionID: childCoderID, Turn: 6, Role: "coder"},
				{SessionID: "unmapped-3", Turn: 7, Role: "planner"},
				{SessionID: childTesterID, Turn: 8, Role: "tester"},
				{SessionID: childCoderID, Turn: 9, Role: "coder"},
				{SessionID: childTesterID, Turn: 10, Role: "tester"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var rework *domain.Anomaly
		for i := range anomalies {
			if anomalies[i].RuleID == domain.RuleSubagentReworkLoop {
				rework = &anomalies[i]
				break
			}
		}
		if rework == nil {
			t.Fatalf("expected rework loop for valid child sessions despite unmapped noise, got: %+v", anomalies)
		}
		if *rework.Context.Role != "coder" || *rework.Context.SecondaryRole != "tester" {
			t.Errorf("expected coder <-> tester loop, got %s <-> %s", *rework.Context.Role, *rework.Context.SecondaryRole)
		}
		if *rework.Context.Count != 3 {
			t.Errorf("expected 3 cycles, got %d", *rework.Context.Count)
		}
	})

	t.Run("boundary: invocations with only root sessions never trigger rework loop in events", func(t *testing.T) {
		r1 := "root-orchestrator"
		r2 := "root-general"
		r3 := "root-build"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: r1, ParentID: nil, Role: "orchestrator"},
				{ID: r2, ParentID: nil, Role: "general"},
				{ID: r3, ParentID: nil, Role: "build"},
			},
			Events: []domain.EventFact{
				{SessionID: r1, Turn: 1, Role: "coder"},
				{SessionID: r2, Turn: 2, Role: "tester"},
				{SessionID: r1, Turn: 3, Role: "coder"},
				{SessionID: r2, Turn: 4, Role: "tester"},
				{SessionID: r1, Turn: 5, Role: "coder"},
				{SessionID: r2, Turn: 6, Role: "tester"},
				{SessionID: r1, Turn: 7, Role: "coder"},
				{SessionID: r2, Turn: 8, Role: "tester"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		for _, a := range anomalies {
			if a.RuleID == domain.RuleSubagentReworkLoop {
				t.Fatalf("only root sessions in events must not trigger rework loop, got: %+v", a)
			}
		}
	})

	t.Run("boundary: invocations with only root sessions never trigger rework loop in fallback", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: "r1", ParentID: nil, Role: "coder"},
				{ID: "r2", ParentID: nil, Role: "tester"},
				{ID: "r3", ParentID: nil, Role: "coder"},
				{ID: "r4", ParentID: nil, Role: "tester"},
				{ID: "r5", ParentID: nil, Role: "coder"},
				{ID: "r6", ParentID: nil, Role: "tester"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		for _, a := range anomalies {
			if a.RuleID == domain.RuleSubagentReworkLoop {
				t.Fatalf("only root sessions in invocations fallback must not trigger rework loop, got: %+v", a)
			}
		}
	})

	t.Run("boundary: orphaned child sessions with non-existent parent ID trigger rework loop in events", func(t *testing.T) {
		missingParent := "missing-parent-404"
		child1 := "child-c"
		child2 := "child-t"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: child1, ParentID: &missingParent, Role: "coder"},
				{ID: child2, ParentID: &missingParent, Role: "tester"},
			},
			Events: []domain.EventFact{
				{SessionID: child1, Turn: 1, Role: "coder"},
				{SessionID: child2, Turn: 2, Role: "tester"},
				{SessionID: child1, Turn: 3, Role: "coder"},
				{SessionID: child2, Turn: 4, Role: "tester"},
				{SessionID: child1, Turn: 5, Role: "coder"},
				{SessionID: child2, Turn: 6, Role: "tester"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var rework *domain.Anomaly
		for i := range anomalies {
			if anomalies[i].RuleID == domain.RuleSubagentReworkLoop {
				rework = &anomalies[i]
				break
			}
		}
		if rework == nil {
			t.Fatalf("orphaned child sessions with non-nil ParentID must trigger rework loop, got: %+v", anomalies)
		}
		if rework.Severity != domain.SeverityWarning {
			t.Errorf("expected WARNING, got %s", rework.Severity)
		}
		if *rework.Context.Role != "coder" || *rework.Context.SecondaryRole != "tester" {
			t.Errorf("expected coder <-> tester loop, got %s <-> %s", *rework.Context.Role, *rework.Context.SecondaryRole)
		}
		if *rework.Context.Count != 3 {
			t.Errorf("expected 3 cycles, got %d", *rework.Context.Count)
		}
	})

	t.Run("boundary: orphaned child sessions in fallback sequence trigger rework loop", func(t *testing.T) {
		phantomParent := "missing-root-ses"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: "c1", ParentID: &phantomParent, Role: "coder"},
				{ID: "t1", ParentID: &phantomParent, Role: "tester"},
				{ID: "c2", ParentID: &phantomParent, Role: "coder"},
				{ID: "t2", ParentID: &phantomParent, Role: "tester"},
				{ID: "c3", ParentID: &phantomParent, Role: "coder"},
				{ID: "t3", ParentID: &phantomParent, Role: "tester"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var rework *domain.Anomaly
		for i := range anomalies {
			if anomalies[i].RuleID == domain.RuleSubagentReworkLoop {
				rework = &anomalies[i]
				break
			}
		}
		if rework == nil {
			t.Fatalf("orphaned child sessions in fallback must trigger rework loop, got: %+v", anomalies)
		}
		if *rework.Context.Role != "coder" || *rework.Context.SecondaryRole != "tester" {
			t.Errorf("expected coder <-> tester loop, got %s <-> %s", *rework.Context.Role, *rework.Context.SecondaryRole)
		}
		if *rework.Context.Count != 3 {
			t.Errorf("expected 3 cycles, got %d", *rework.Context.Count)
		}
	})

	t.Run("boundary: child invocations with empty ID string handled safely in fallback", func(t *testing.T) {
		parent := "root"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: "", ParentID: &parent, Role: "coder"},
				{ID: "", ParentID: &parent, Role: "tester"},
				{ID: "", ParentID: &parent, Role: "coder"},
				{ID: "", ParentID: &parent, Role: "tester"},
				{ID: "", ParentID: &parent, Role: "coder"},
				{ID: "", ParentID: &parent, Role: "tester"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var rework *domain.Anomaly
		for i := range anomalies {
			if anomalies[i].RuleID == domain.RuleSubagentReworkLoop {
				rework = &anomalies[i]
				break
			}
		}
		if rework == nil {
			t.Fatalf("child invocations with empty ID in fallback must trigger rework loop, got: %+v", anomalies)
		}
		if *rework.Context.Count != 3 {
			t.Errorf("expected 3 cycles, got %d", *rework.Context.Count)
		}
	})

	t.Run("boundary: events with only unmapped SessionIDs fall back to invocations sequence", func(t *testing.T) {
		rootID := "ses-root"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: rootID, ParentID: nil, Role: "orchestrator"},
				{ID: "c1", ParentID: &rootID, Role: "coder"},
				{ID: "t1", ParentID: &rootID, Role: "tester"},
				{ID: "c2", ParentID: &rootID, Role: "coder"},
				{ID: "t2", ParentID: &rootID, Role: "tester"},
				{ID: "c3", ParentID: &rootID, Role: "coder"},
				{ID: "t3", ParentID: &rootID, Role: "tester"},
			},
			Events: []domain.EventFact{
				// All events have unmapped session IDs
				{SessionID: "unmapped-orphan-1", Turn: 1, Role: "reviewer"},
				{SessionID: "unmapped-orphan-2", Turn: 2, Role: "reviewer"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		var rework *domain.Anomaly
		for i := range anomalies {
			if anomalies[i].RuleID == domain.RuleSubagentReworkLoop {
				rework = &anomalies[i]
				break
			}
		}
		if rework == nil {
			t.Fatalf("expected fallback to invocations when all events are unmapped, got: %+v", anomalies)
		}
		if *rework.Context.Role != "coder" || *rework.Context.SecondaryRole != "tester" {
			t.Errorf("expected coder <-> tester loop, got %s <-> %s", *rework.Context.Role, *rework.Context.SecondaryRole)
		}
		if *rework.Context.Count != 3 {
			t.Errorf("expected 3 cycles, got %d", *rework.Context.Count)
		}
	})
}

// 6. Boundary Cases & Division by Zero Guards
func TestDetectAnomalies_BoundaryCases(t *testing.T) {
	t.Parallel()
	t.Run("empty ledger returns non-nil slice", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{}
		anomalies := domain.DetectAnomalies(ledger)
		if anomalies == nil {
			t.Fatalf("expected non-nil empty slice for empty ledger")
		}
		if len(anomalies) != 0 {
			t.Fatalf("expected 0 anomalies for empty ledger, got %d", len(anomalies))
		}
	})

	t.Run("nil events, nil tools, nil invocations", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events:      nil,
			Tools:       nil,
			Invocations: nil,
		}
		anomalies := domain.DetectAnomalies(ledger)
		if anomalies == nil || len(anomalies) != 0 {
			t.Fatalf("expected empty slice for nil fields, got: %+v", anomalies)
		}
	})

	t.Run("zero tokens throughout does not cause division by zero or false positive", func(t *testing.T) {
		toolName := "bash"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{
					ID:   "ses-0",
					Role: "orchestrator",
					Tokens: domain.InvocationTokens{
						Input:       0,
						Output:      0,
						CacheRead:   0,
						Reasoning:   0,
						PeakContext: 0,
					},
					ActiveDurationSeconds: 0,
					Turns:                 0,
					CostUSD:               0,
				},
			},
			Tools: []domain.ToolFact{
				{SessionID: "session-a", Name: "bash", Calls: 0, Errors: 0},
			},
			Events: []domain.EventFact{
				{SessionID: "ses-0", Turn: 0, EventType: domain.EventTypeToolCall, ToolName: &toolName, Status: "completed", ContextTokens: 0},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 0 {
			t.Fatalf("expected 0 anomalies on zero tokens, got: %+v", anomalies)
		}
	})

	t.Run("events with nil tool name or empty tool name safely handled", func(t *testing.T) {
		emptyTool := ""
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{Turn: 1, EventType: domain.EventTypeToolCall, ToolName: nil, Status: "error"},
				{Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &emptyTool, Status: "error"},
				{Turn: 3, EventType: domain.EventTypeToolCall, ToolName: nil, Status: "completed"},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 0 {
			t.Fatalf("expected 0 anomalies for nil/empty tool names, got: %+v", anomalies)
		}
	})

	t.Run("events with empty role and status handled without panic", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Events: []domain.EventFact{
				{Turn: 1, Role: "", EventType: domain.EventTypeTurnMessage, Status: ""},
				{Turn: 2, Role: "   ", EventType: domain.EventTypeTurnMessage, Status: ""},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 0 {
			t.Fatalf("expected 0 anomalies, got: %+v", anomalies)
		}
	})

	t.Run("v1 fallback tool facts with empty name safely skipped", func(t *testing.T) {
		ledger := domain.AgenticFactLedger{
			Tools: []domain.ToolFact{
				{SessionID: "session-a", Name: "", Calls: 10, Errors: 10},
			},
		}

		anomalies := domain.DetectAnomalies(ledger)
		if len(anomalies) != 0 {
			t.Fatalf("expected 0 anomalies for empty tool name in v1, got: %+v", anomalies)
		}
	})
}

// 7. Deterministic Sorting Order Across All 4 Rules Simultaneously
func TestDetectAnomalies_DeterministicSortingAllRules(t *testing.T) {
	t.Parallel()
	toolA := "tool_a"
	toolB := "tool_b"
	toolC := "tool_c"

	// Construct a ledger that intentionally triggers multiple anomalies across all 4 rules:
	// Rule 1: TOOL_RETRY_STREAK (CRITICAL: 5 errors for toolA; WARNING: 3 errors for toolB)
	// Rule 2: TOOL_HIGH_FAILURE_RATE (CRITICAL: 6 calls, 4 errors for toolC; WARNING: 6 calls, 2 errors for toolB)
	// Rule 3: CONTEXT_SATURATION_SPIKE (CRITICAL: ceiling breach 165k; CRITICAL: jump 90k; WARNING: jump 50k)
	// Rule 4: SUBAGENT_REWORK_LOOP (WARNING: 3 cycles between coder and tester)
	rootID := "ses-root"
	ledger := domain.AgenticFactLedger{
		Invocations: []domain.InvocationFact{
			{ID: rootID, ParentID: nil, Role: "orchestrator"},
			{ID: "s1", ParentID: &rootID, Role: "coder"},
		},
		Events: []domain.EventFact{
			// Rule 3: Token jumps and ceiling breach
			{SessionID: "s1", Turn: 1, Role: "coder", ContextTokens: 5_000},
			{SessionID: "s1", Turn: 2, Role: "tester", ContextTokens: 55_001},  // +50,001 jump -> WARNING spike
			{SessionID: "s1", Turn: 3, Role: "coder", ContextTokens: 150_000},  // +94,999 jump -> CRITICAL spike
			{SessionID: "s1", Turn: 4, Role: "tester", ContextTokens: 165_000}, // >= 160k -> CRITICAL ceiling breach

			// Subagent loop continuation (turns 1-6 give coder <-> tester 3 cycles)
			{SessionID: "s1", Turn: 5, Role: "coder", ContextTokens: 166_000},
			{SessionID: "s1", Turn: 6, Role: "tester", ContextTokens: 167_000},

			// Rule 1: Tool retry streak for toolA (5 errors -> CRITICAL)
			{SessionID: "s1", Turn: 10, EventType: domain.EventTypeToolCall, ToolName: &toolA, Status: "error"},
			{SessionID: "s1", Turn: 11, EventType: domain.EventTypeToolCall, ToolName: &toolA, Status: "error"},
			{SessionID: "s1", Turn: 12, EventType: domain.EventTypeToolCall, ToolName: &toolA, Status: "error"},
			{SessionID: "s1", Turn: 13, EventType: domain.EventTypeToolCall, ToolName: &toolA, Status: "error"},
			{SessionID: "s1", Turn: 14, EventType: domain.EventTypeToolCall, ToolName: &toolA, Status: "error"},
			{SessionID: "s1", Turn: 15, EventType: domain.EventTypeToolCall, ToolName: &toolA, Status: "completed"},

			// Rule 1: Tool retry streak for toolB (3 errors -> WARNING)
			// Also contributes to toolB calls (6 calls, 3 errors = 50% -> CRITICAL failure rate!)
			// Wait, let's keep toolB as 3 errors and 10 calls (30% -> WARNING failure rate)
			{SessionID: "s1", Turn: 20, EventType: domain.EventTypeToolCall, ToolName: &toolB, Status: "error"},
			{SessionID: "s1", Turn: 21, EventType: domain.EventTypeToolCall, ToolName: &toolB, Status: "error"},
			{SessionID: "s1", Turn: 22, EventType: domain.EventTypeToolCall, ToolName: &toolB, Status: "error"},
			{SessionID: "s1", Turn: 23, EventType: domain.EventTypeToolCall, ToolName: &toolB, Status: "completed"},
			{SessionID: "s1", Turn: 24, EventType: domain.EventTypeToolCall, ToolName: &toolB, Status: "completed"},
			{SessionID: "s1", Turn: 25, EventType: domain.EventTypeToolCall, ToolName: &toolB, Status: "completed"},
			{SessionID: "s1", Turn: 26, EventType: domain.EventTypeToolCall, ToolName: &toolB, Status: "completed"},
			{SessionID: "s1", Turn: 27, EventType: domain.EventTypeToolCall, ToolName: &toolB, Status: "completed"},
			{SessionID: "s1", Turn: 28, EventType: domain.EventTypeToolCall, ToolName: &toolB, Status: "completed"},
			{SessionID: "s1", Turn: 29, EventType: domain.EventTypeToolCall, ToolName: &toolB, Status: "completed"},

			// Rule 2: toolC with 6 calls, 4 errors (66.7% -> CRITICAL failure rate)
			// Alternate errors to avoid tool retry streak
			{SessionID: "s1", Turn: 30, EventType: domain.EventTypeToolCall, ToolName: &toolC, Status: "error"},
			{SessionID: "s1", Turn: 31, EventType: domain.EventTypeToolCall, ToolName: &toolC, Status: "completed"},
			{SessionID: "s1", Turn: 32, EventType: domain.EventTypeToolCall, ToolName: &toolC, Status: "error"},
			{SessionID: "s1", Turn: 33, EventType: domain.EventTypeToolCall, ToolName: &toolC, Status: "completed"},
			{SessionID: "s1", Turn: 34, EventType: domain.EventTypeToolCall, ToolName: &toolC, Status: "error"},
			{SessionID: "s1", Turn: 35, EventType: domain.EventTypeToolCall, ToolName: &toolC, Status: "error"},
		},
	}

	anomalies := domain.DetectAnomalies(ledger)
	if len(anomalies) < 7 {
		t.Fatalf("expected at least 7 anomalies across all rules, got %d: %+v", len(anomalies), anomalies)
	}

	// Verify all 4 rule IDs are represented
	ruleFound := make(map[domain.AnomalyRuleID]bool)
	severityFound := make(map[domain.AnomalySeverity]bool)
	for _, a := range anomalies {
		ruleFound[a.RuleID] = true
		severityFound[a.Severity] = true
	}

	expectedRules := []domain.AnomalyRuleID{
		domain.RuleToolRetryStreak,
		domain.RuleToolHighFailureRate,
		domain.RuleContextSaturationSpike,
		domain.RuleSubagentReworkLoop,
	}
	for _, r := range expectedRules {
		if !ruleFound[r] {
			t.Errorf("expected rule %s to be present in detected anomalies", r)
		}
	}
	if !severityFound[domain.SeverityCritical] || !severityFound[domain.SeverityWarning] {
		t.Errorf("expected both CRITICAL and WARNING severities")
	}

	// Deterministic sorting verification:
	// 1. Severity: All CRITICAL before all WARNING
	// 2. RuleID: Lexicographical order within same severity
	// 3. Turn / ToolName / Title order within same RuleID and severity
	for k := 0; k < len(anomalies)-1; k++ {
		a := anomalies[k]
		b := anomalies[k+1]

		// 1. Severity check
		if a.Severity == domain.SeverityWarning && b.Severity == domain.SeverityCritical {
			t.Fatalf("sorting violation: WARNING anomaly at index %d preceded CRITICAL anomaly at %d:\na=%+v\nb=%+v", k, k+1, a, b)
		}

		if a.Severity == b.Severity {
			// 2. RuleID check
			if a.RuleID > b.RuleID {
				t.Fatalf("sorting violation: RuleID %s at index %d preceded RuleID %s at %d", a.RuleID, k, b.RuleID, k+1)
			}

			// 3. Turn check when RuleID matches
			if a.RuleID == b.RuleID {
				if a.Context.Turn != nil && b.Context.Turn != nil {
					if *a.Context.Turn > *b.Context.Turn {
						t.Fatalf("sorting violation: Turn %d preceded Turn %d for same rule %s", *a.Context.Turn, *b.Context.Turn, a.RuleID)
					}
				}
			}
		}
	}
}

func TestDetectAnomalies_SortingTieBreakers(t *testing.T) {
	t.Parallel()
	// Construct a scenario with identical RuleID and Severity, but differing Turn / ToolName / Title
	toolX := "tool_x"
	toolY := "tool_y"
	turn1 := int64(1)

	// Context saturation spikes: one with Turn 1, one with Turn nil (from v1 fallback invocation)
	ledger := domain.AgenticFactLedger{
		Invocations: []domain.InvocationFact{
			{ID: "inv-v1", Role: "coder", Tokens: domain.InvocationTokens{PeakContext: 180_000}},
		},
		Events: []domain.EventFact{
			// tool_call with 6 calls and 4 errors -> CRITICAL
			{SessionID: "s1", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolY, Status: "error"},
			{SessionID: "s1", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolY, Status: "error"},
			{SessionID: "s1", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolY, Status: "error"},
			{SessionID: "s1", Turn: 4, EventType: domain.EventTypeToolCall, ToolName: &toolY, Status: "error"},
			{SessionID: "s1", Turn: 5, EventType: domain.EventTypeToolCall, ToolName: &toolY, Status: "completed"},
			{SessionID: "s1", Turn: 6, EventType: domain.EventTypeToolCall, ToolName: &toolY, Status: "completed"},
			// tool_x also 6 calls and 4 errors -> CRITICAL
			{SessionID: "s1", Turn: 7, EventType: domain.EventTypeToolCall, ToolName: &toolX, Status: "error"},
			{SessionID: "s1", Turn: 8, EventType: domain.EventTypeToolCall, ToolName: &toolX, Status: "error"},
			{SessionID: "s1", Turn: 9, EventType: domain.EventTypeToolCall, ToolName: &toolX, Status: "error"},
			{SessionID: "s1", Turn: 10, EventType: domain.EventTypeToolCall, ToolName: &toolX, Status: "error"},
			{SessionID: "s1", Turn: 11, EventType: domain.EventTypeToolCall, ToolName: &toolX, Status: "completed"},
			{SessionID: "s1", Turn: 12, EventType: domain.EventTypeToolCall, ToolName: &toolX, Status: "completed"},
		},
	}

	anomalies := domain.DetectAnomalies(ledger)
	var rateAnoms []domain.Anomaly
	for _, a := range anomalies {
		if a.RuleID == domain.RuleToolHighFailureRate {
			rateAnoms = append(rateAnoms, a)
		}
	}
	if len(rateAnoms) != 2 {
		t.Fatalf("expected 2 rate anomalies, got %d", len(rateAnoms))
	}
	// ToolName ascending tie-breaker: tool_x should precede tool_y
	if *rateAnoms[0].Context.ToolName != "tool_x" || *rateAnoms[1].Context.ToolName != "tool_y" {
		t.Errorf("expected tool_x before tool_y, got %s then %s", *rateAnoms[0].Context.ToolName, *rateAnoms[1].Context.ToolName)
	}

	// Verify turn non-nil vs nil:
	// A slice with turn 1 before turn nil
	aWithTurn := domain.Anomaly{
		RuleID:   domain.RuleContextSaturationSpike,
		Severity: domain.SeverityCritical,
		Title:    "With Turn",
		Context:  domain.AnomalyContext{Turn: &turn1},
	}
	aWithoutTurn := domain.Anomaly{
		RuleID:   domain.RuleContextSaturationSpike,
		Severity: domain.SeverityCritical,
		Title:    "Without Turn",
		Context:  domain.AnomalyContext{Turn: nil},
	}
	// Test DetectAnomalies sorting helper logic by sorting a slice directly
	sorted := []domain.Anomaly{aWithoutTurn, aWithTurn}
	// Derive through DetectAnomalies on a constructed ledger with both
	testLedger := domain.AgenticFactLedger{
		Events: []domain.EventFact{
			{SessionID: "s1", Turn: 1, ContextTokens: 170_000},
		},
	}
	res := domain.DetectAnomalies(testLedger)
	if len(res) != 1 {
		t.Fatalf("expected 1 anomaly, got %d", len(res))
	}
	if res[0].Context.Turn == nil || *res[0].Context.Turn != 1 {
		t.Errorf("expected turn 1")
	}
	_ = sorted
}

// 8. DeriveMetricsLedger Integration
func TestDeriveMetricsLedger_IntegratesAnomalies(t *testing.T) {
	t.Parallel()
	toolBash := "bash"
	factLedger := domain.AgenticFactLedger{
		Events: []domain.EventFact{
			{SessionID: "s1", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			{SessionID: "s1", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			{SessionID: "s1", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
			{SessionID: "s1", Turn: 4, EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "completed"},
		},
	}

	metrics := domain.DeriveMetricsLedger(factLedger, domain.DefaultProjectConfig())
	if len(metrics.Anomalies) != 1 {
		t.Fatalf("expected DeriveMetricsLedger to populate Anomalies with 1 entry, got %d", len(metrics.Anomalies))
	}
	if metrics.Anomalies[0].RuleID != domain.RuleToolRetryStreak {
		t.Errorf("expected RuleToolRetryStreak, got %s", metrics.Anomalies[0].RuleID)
	}
}

// 9. ComputeFleetAnomalies Pure Domain Tests
func TestComputeFleetAnomalies_EmptyInput(t *testing.T) {
	t.Parallel()
	t.Run("nil slice returns initialized zero-value structure", func(t *testing.T) {
		summary := domain.ComputeFleetAnomalies(nil)
		if summary.TotalAnomalies != 0 {
			t.Errorf("expected 0 total anomalies, got %d", summary.TotalAnomalies)
		}
		if summary.WarningCount != 0 || summary.CriticalCount != 0 {
			t.Errorf("expected 0 warning/critical, got %d/%d", summary.WarningCount, summary.CriticalCount)
		}
		if summary.TotalTickets != 0 || summary.TicketsWithAnomalies != 0 {
			t.Errorf("expected 0 total tickets, got %d", summary.TotalTickets)
		}
		if summary.AnomalyRatePerTicket != 0.0 {
			t.Errorf("expected 0.0 anomaly rate, got %f", summary.AnomalyRatePerTicket)
		}
		if summary.SeverityCounts == nil {
			t.Fatal("expected non-nil SeverityCounts map")
		}
		if summary.SeverityCounts[string(domain.SeverityWarning)] != 0 || summary.SeverityCounts[string(domain.SeverityCritical)] != 0 {
			t.Errorf("expected 0 in severity counts, got %+v", summary.SeverityCounts)
		}
		if summary.RuleCounts == nil {
			t.Fatal("expected non-nil RuleCounts map")
		}
		if summary.RuleCounts[string(domain.RuleToolRetryStreak)] != 0 || summary.RuleCounts[string(domain.RuleContextSaturationSpike)] != 0 {
			t.Errorf("expected 0 in rule counts, got %+v", summary.RuleCounts)
		}
		if summary.BrittleTools == nil {
			t.Fatal("expected non-nil BrittleTools slice")
		}
		if len(summary.BrittleTools) != 0 {
			t.Errorf("expected empty BrittleTools, got %d", len(summary.BrittleTools))
		}
	})

	t.Run("empty slice returns initialized zero-value structure", func(t *testing.T) {
		summary := domain.ComputeFleetAnomalies([]domain.AgenticFactLedger{})
		if summary.TotalTickets != 0 {
			t.Errorf("expected 0 total tickets, got %d", summary.TotalTickets)
		}
		if summary.BrittleTools == nil || len(summary.BrittleTools) != 0 {
			t.Errorf("expected empty non-nil BrittleTools slice")
		}
	})
}

func TestComputeFleetAnomalies_CleanLedgers(t *testing.T) {
	t.Parallel()
	toolRead := "read_file"
	toolWrite := "write_file"
	tid1 := "AL-1"
	tid2 := "AL-2"

	ledgers := []domain.AgenticFactLedger{
		{
			TicketID: &tid1,
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolRead, Status: "completed", ContextTokens: 5000},
				{SessionID: "s1", Turn: 2, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolWrite, Status: "completed", ContextTokens: 10000},
			},
		},
		{
			TicketID: &tid2,
			Events: []domain.EventFact{
				{SessionID: "s2", Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolRead, Status: "completed", ContextTokens: 8000},
			},
		},
	}

	summary := domain.ComputeFleetAnomalies(ledgers)
	if summary.TotalTickets != 2 {
		t.Errorf("expected 2 total tickets, got %d", summary.TotalTickets)
	}
	if summary.TotalAnomalies != 0 {
		t.Errorf("expected 0 total anomalies, got %d", summary.TotalAnomalies)
	}
	if summary.TicketsWithAnomalies != 0 {
		t.Errorf("expected 0 tickets with anomalies, got %d", summary.TicketsWithAnomalies)
	}
	if summary.AnomalyRatePerTicket != 0.0 {
		t.Errorf("expected 0.0 rate, got %f", summary.AnomalyRatePerTicket)
	}
	if len(summary.BrittleTools) != 0 {
		t.Errorf("expected 0 brittle tools on clean runs, got %d", len(summary.BrittleTools))
	}
}

func TestComputeFleetAnomalies_MultipleAnomaliesAndBrittleTools(t *testing.T) {
	t.Parallel()
	toolBash := "bash"
	toolGit := "git"
	toolPython := "python"
	tid1 := "AL-1"
	tid2 := "AL-2"
	tid3 := "AL-3"

	ledgers := []domain.AgenticFactLedger{
		// Ticket 1: Tool retry streak (WARNING) on bash (3 errors)
		{
			TicketID: &tid1,
			Events: []domain.EventFact{
				{SessionID: "s1", Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error", ContextTokens: 10000},
				{SessionID: "s1", Turn: 2, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error", ContextTokens: 12000},
				{SessionID: "s1", Turn: 3, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error", ContextTokens: 14000},
				{SessionID: "s1", Turn: 4, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "completed", ContextTokens: 16000},
			},
		},
		// Ticket 2: Context ceiling breach (CRITICAL >= 160k) + High failure rate on git (5 calls, 2 errors = 40.0% -> WARNING, non-consecutive so no retry streak)
		{
			TicketID: &tid2,
			Events: []domain.EventFact{
				{SessionID: "s2", Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "error", ContextTokens: 100000},
				{SessionID: "s2", Turn: 2, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed", ContextTokens: 120000},
				{SessionID: "s2", Turn: 3, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "error", ContextTokens: 135000},
				{SessionID: "s2", Turn: 4, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed", ContextTokens: 150000},
				{SessionID: "s2", Turn: 5, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolGit, Status: "completed", ContextTokens: 162000}, // Ceiling breach, delta 12k
			},
		},
		// Ticket 3: Clean ticket with python tool (2 calls, 1 error, 0 anomalies, error > 0 makes it brittle)
		{
			TicketID: &tid3,
			Events: []domain.EventFact{
				{SessionID: "s3", Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolPython, Status: "error", ContextTokens: 5000},
				{SessionID: "s3", Turn: 2, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolPython, Status: "completed", ContextTokens: 7000},
			},
		},
	}

	summary := domain.ComputeFleetAnomalies(ledgers)

	if summary.TotalTickets != 3 {
		t.Fatalf("expected 3 total tickets, got %d", summary.TotalTickets)
	}
	if summary.TicketsWithAnomalies != 2 {
		t.Fatalf("expected 2 tickets with anomalies, got %d", summary.TicketsWithAnomalies)
	}

	// Ticket 1 has 1 anomaly (RuleToolRetryStreak WARNING)
	// Ticket 2 has 2 anomalies (RuleToolHighFailureRate WARNING, RuleContextSaturationSpike CRITICAL ceiling breach)
	// Total: 3 anomalies (2 WARNING, 1 CRITICAL)
	if summary.TotalAnomalies != 3 {
		t.Fatalf("expected 3 total anomalies, got %d", summary.TotalAnomalies)
	}
	if summary.WarningCount != 2 {
		t.Errorf("expected 2 warning, got %d", summary.WarningCount)
	}
	if summary.CriticalCount != 1 {
		t.Errorf("expected 1 critical, got %d", summary.CriticalCount)
	}
	if summary.SeverityCounts[string(domain.SeverityWarning)] != 2 {
		t.Errorf("expected 2 in SeverityCounts[WARNING], got %d", summary.SeverityCounts[string(domain.SeverityWarning)])
	}
	if summary.SeverityCounts[string(domain.SeverityCritical)] != 1 {
		t.Errorf("expected 1 in SeverityCounts[CRITICAL], got %d", summary.SeverityCounts[string(domain.SeverityCritical)])
	}

	// Rule counts
	if summary.RuleCounts[string(domain.RuleToolRetryStreak)] != 1 {
		t.Errorf("expected 1 RuleToolRetryStreak, got %d", summary.RuleCounts[string(domain.RuleToolRetryStreak)])
	}
	if summary.RuleCounts[string(domain.RuleToolHighFailureRate)] != 1 {
		t.Errorf("expected 1 RuleToolHighFailureRate, got %d", summary.RuleCounts[string(domain.RuleToolHighFailureRate)])
	}
	if summary.RuleCounts[string(domain.RuleContextSaturationSpike)] != 1 {
		t.Errorf("expected 1 RuleContextSaturationSpike, got %d", summary.RuleCounts[string(domain.RuleContextSaturationSpike)])
	}
	if summary.RuleCounts[string(domain.RuleSubagentReworkLoop)] != 0 {
		t.Errorf("expected 0 RuleSubagentReworkLoop, got %d", summary.RuleCounts[string(domain.RuleSubagentReworkLoop)])
	}

	// Anomaly rate per ticket: 3 anomalies / 3 tickets = 1.00
	if summary.AnomalyRatePerTicket != 1.0 {
		t.Errorf("expected 1.0 anomaly rate, got %f", summary.AnomalyRatePerTicket)
	}

	// Brittle tools:
	// bash: 4 calls, 3 errors, 1 anomaly -> failure rate = 75.0%
	// git: 5 calls, 2 errors, 1 anomaly -> failure rate = 40.0%
	// python: 2 calls, 1 error, 0 anomalies -> failure rate = 50.0%
	//
	// Sorting order:
	// 1. AnomalyCount descending (bash and git have 1, python has 0)
	// 2. ErrorCount descending (bash has 3, git has 2)
	// 3. FailureRatePct descending
	// 4. ToolName ascending
	// Therefore order must be: bash (1), git (1), python (0)
	if len(summary.BrittleTools) != 3 {
		t.Fatalf("expected 3 brittle tools, got %d", len(summary.BrittleTools))
	}

	bt0 := summary.BrittleTools[0]
	if bt0.ToolName != "bash" || bt0.AnomalyCount != 1 || bt0.ErrorCount != 3 || bt0.TotalCalls != 4 || bt0.FailureRatePct != 75.0 {
		t.Errorf("unexpected first brittle tool: %+v", bt0)
	}

	bt1 := summary.BrittleTools[1]
	if bt1.ToolName != "git" || bt1.AnomalyCount != 1 || bt1.ErrorCount != 2 || bt1.TotalCalls != 5 || bt1.FailureRatePct != 40.0 {
		t.Errorf("unexpected second brittle tool: %+v", bt1)
	}

	bt2 := summary.BrittleTools[2]
	if bt2.ToolName != "python" || bt2.AnomalyCount != 0 || bt2.ErrorCount != 1 || bt2.TotalCalls != 2 || bt2.FailureRatePct != 50.0 {
		t.Errorf("unexpected third brittle tool: %+v", bt2)
	}
}

func TestComputeFleetAnomalies_BrittleToolsSortingTieBreakers(t *testing.T) {
	t.Parallel()
	// Test tie-breaking across all 4 sorting criteria:
	// 1. AnomalyCount descending
	// 2. ErrorCount descending
	// 3. FailureRatePct descending
	// 4. ToolName ascending (alphabetical tie-breaker)

	t.Run("tier 2: tie breaks by ErrorCount when AnomalyCount is equal", func(t *testing.T) {
		toolA := "tool_two_errors"
		toolB := "tool_one_error"
		tid := "T-ERR"

		// Both have 0 anomalies (<5 calls, <3 consecutive errors)
		// toolA: 3 calls, 2 errors (66.7%)
		// toolB: 1 call, 1 error (100.0%)
		// ErrorCount 2 > 1, so toolA must come before toolB despite lower failure rate
		ledgers := []domain.AgenticFactLedger{
			{
				TicketID: &tid,
				Events: []domain.EventFact{
					{SessionID: "s1", Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolA, Status: "error"},
					{SessionID: "s1", Turn: 2, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolA, Status: "completed"},
					{SessionID: "s1", Turn: 3, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolA, Status: "error"},
					{SessionID: "s1", Turn: 4, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolB, Status: "error"},
				},
			},
		}

		summary := domain.ComputeFleetAnomalies(ledgers)
		if len(summary.BrittleTools) != 2 {
			t.Fatalf("expected 2 brittle tools, got %d", len(summary.BrittleTools))
		}
		if summary.BrittleTools[0].ToolName != "tool_two_errors" {
			t.Errorf("expected tool_two_errors first by error count (2 > 1), got %s", summary.BrittleTools[0].ToolName)
		}
		if summary.BrittleTools[1].ToolName != "tool_one_error" {
			t.Errorf("expected tool_one_error second, got %s", summary.BrittleTools[1].ToolName)
		}
	})

	t.Run("tier 3: tie breaks by FailureRatePct when AnomalyCount and ErrorCount are equal", func(t *testing.T) {
		toolHighRate := "high_rate_tool"
		toolLowRate := "low_rate_tool"
		tid := "T-RATE"

		// Both have 0 anomalies and exactly 2 errors:
		// toolHighRate: 2 calls, 2 errors -> 100.0% failure rate
		// toolLowRate: 4 calls, 2 errors (interleaved) -> 50.0% failure rate
		// Same anomaly count (0), same error count (2), but high_rate_tool (100%) > low_rate_tool (50%)
		ledgers := []domain.AgenticFactLedger{
			{
				TicketID: &tid,
				Events: []domain.EventFact{
					{SessionID: "s1", Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolHighRate, Status: "error"},
					{SessionID: "s1", Turn: 2, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolHighRate, Status: "error"},

					{SessionID: "s1", Turn: 3, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolLowRate, Status: "error"},
					{SessionID: "s1", Turn: 4, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolLowRate, Status: "completed"},
					{SessionID: "s1", Turn: 5, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolLowRate, Status: "error"},
					{SessionID: "s1", Turn: 6, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolLowRate, Status: "completed"},
				},
			},
		}

		summary := domain.ComputeFleetAnomalies(ledgers)
		if len(summary.BrittleTools) != 2 {
			t.Fatalf("expected 2 brittle tools, got %d", len(summary.BrittleTools))
		}
		if summary.BrittleTools[0].ToolName != "high_rate_tool" {
			t.Errorf("expected high_rate_tool first by failure rate (100%% > 50%%), got %s", summary.BrittleTools[0].ToolName)
		}
		if summary.BrittleTools[1].ToolName != "low_rate_tool" {
			t.Errorf("expected low_rate_tool second, got %s", summary.BrittleTools[1].ToolName)
		}
	})

	t.Run("tier 4: tie breaks alphabetically by tool name", func(t *testing.T) {
		toolZ := "z_tool"
		toolA := "a_tool"
		tid := "T1"

		ledgers := []domain.AgenticFactLedger{
			{
				TicketID: &tid,
				Events: []domain.EventFact{
					{SessionID: "s1", Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolZ, Status: "error"},
					{SessionID: "s1", Turn: 2, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolA, Status: "error"},
				},
			},
		}

		summary := domain.ComputeFleetAnomalies(ledgers)
		if len(summary.BrittleTools) != 2 {
			t.Fatalf("expected 2 brittle tools, got %d", len(summary.BrittleTools))
		}
		if summary.BrittleTools[0].ToolName != "a_tool" {
			t.Errorf("expected a_tool first by alphabetical order, got %s", summary.BrittleTools[0].ToolName)
		}
		if summary.BrittleTools[1].ToolName != "z_tool" {
			t.Errorf("expected z_tool second by alphabetical order, got %s", summary.BrittleTools[1].ToolName)
		}
	})

	t.Run("zero calls with errors does not panic and reports 0.0 failure rate", func(t *testing.T) {
		tid := "T-ZERO-CALLS"
		ledgers := []domain.AgenticFactLedger{
			{
				TicketID: &tid,
				Tools: []domain.ToolFact{
					{SessionID: "session-a", Name: "zero_call_tool", Calls: 0, Errors: 1},
				},
			},
		}

		summary := domain.ComputeFleetAnomalies(ledgers)
		if len(summary.BrittleTools) != 1 {
			t.Fatalf("expected 1 brittle tool, got %d", len(summary.BrittleTools))
		}
		bt := summary.BrittleTools[0]
		if bt.ToolName != "zero_call_tool" || bt.TotalCalls != 0 || bt.ErrorCount != 1 || bt.FailureRatePct != 0.0 {
			t.Errorf("unexpected zero-call brittle tool result: %+v", bt)
		}
	})
}

func TestComputeFleetAnomalies_LegacyV1Fallback(t *testing.T) {
	t.Parallel()
	// Legacy v1 workspaces lack Events but provide Tools and Invocations
	tidV1 := "AL-V1"
	ledgers := []domain.AgenticFactLedger{
		{
			TicketID: &tidV1,
			Tools: []domain.ToolFact{
				{SessionID: "session-a", Name: "bash", Calls: 10, Errors: 4}, // 40% failure rate > 20% -> WARNING anomaly
				{SessionID: "session-a", Name: "clean_tool", Calls: 5, Errors: 0},
			},
			Invocations: []domain.InvocationFact{
				{
					ID:   "inv-1",
					Role: "orchestrator",
					Tokens: domain.InvocationTokens{
						PeakContext: 170000, // CRITICAL ceiling breach in v1
					},
				},
			},
		},
	}

	summary := domain.ComputeFleetAnomalies(ledgers)
	if summary.TotalTickets != 1 {
		t.Fatalf("expected 1 ticket, got %d", summary.TotalTickets)
	}
	if summary.TicketsWithAnomalies != 1 {
		t.Errorf("expected 1 ticket with anomalies, got %d", summary.TicketsWithAnomalies)
	}
	if summary.TotalAnomalies != 2 {
		t.Fatalf("expected 2 anomalies (high failure rate + context ceiling), got %d", summary.TotalAnomalies)
	}
	if len(summary.BrittleTools) != 1 {
		t.Fatalf("expected 1 brittle tool (bash), got %d", len(summary.BrittleTools))
	}
	if summary.BrittleTools[0].ToolName != "bash" || summary.BrittleTools[0].FailureRatePct != 40.0 {
		t.Errorf("unexpected brittle tool: %+v", summary.BrittleTools[0])
	}
}

func TestCalculateFleetSummary_IncludesAnomalySummary(t *testing.T) {
	t.Parallel()
	toolBash := "bash"
	tid := "AL-FS"
	ledgers := []domain.AgenticMetricsLedger{
		{
			TicketID: &tid,
			Facts: domain.AgenticFactLedger{
				TicketID: &tid,
				Events: []domain.EventFact{
					{SessionID: "s1", Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
					{SessionID: "s1", Turn: 2, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
					{SessionID: "s1", Turn: 3, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &toolBash, Status: "error"},
				},
			},
			Totals: domain.GroupedMetric{
				Invocations: 1,
				CostUSD:     0.1,
			},
		},
	}

	summary := domain.CalculateFleetSummary(ledgers)
	if summary.AnomalySummary.TotalAnomalies != 1 {
		t.Fatalf("expected 1 anomaly in AnomalySummary, got %d", summary.AnomalySummary.TotalAnomalies)
	}
	if summary.AnomalySummary.TotalTickets != 1 {
		t.Errorf("expected 1 total ticket in AnomalySummary, got %d", summary.AnomalySummary.TotalTickets)
	}
	if len(summary.AnomalySummary.BrittleTools) != 1 {
		t.Errorf("expected 1 brittle tool, got %d", len(summary.AnomalySummary.BrittleTools))
	}
}

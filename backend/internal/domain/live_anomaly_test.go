package domain_test

import (
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

func TestConsecutiveToolErrorsRule(t *testing.T) {
	t.Run("zero tool errors produces no alert", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s1",
			Messages: []domain.LiveMessageFact{
				{
					ID:        "m1",
					Role:      "assistant",
					Timestamp: "2026-10-19T10:00:00Z",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: false},
					},
				},
			},
		}

		alerts := domain.EvaluateConsecutiveToolErrors(trace)
		if len(alerts) != 0 {
			t.Fatalf("expected 0 alerts, got %d", len(alerts))
		}
	})

	t.Run("1 and 2 consecutive tool errors produce no alert", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s1",
			Messages: []domain.LiveMessageFact{
				{
					ID:        "m1",
					Role:      "assistant",
					Timestamp: "2026-10-19T10:00:00Z",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("bash"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateConsecutiveToolErrors(trace)
		if len(alerts) != 0 {
			t.Fatalf("expected 0 alerts, got %d", len(alerts))
		}
	})

	t.Run("3 consecutive tool errors trigger HIGH alert", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s1",
			Messages: []domain.LiveMessageFact{
				{
					ID:        "m1",
					Role:      "assistant",
					Timestamp: "2026-10-19T10:00:00Z",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p3", Type: "tool", ToolName: strPtr("bash"), IsError: true, ErrorMessage: strPtr("command failed")},
					},
				},
			},
		}

		alerts := domain.EvaluateConsecutiveToolErrors(trace)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert, got %d", len(alerts))
		}

		alert := alerts[0]
		if alert.RuleName != domain.RuleConsecutiveToolErrors {
			t.Errorf("expected rule %s, got %s", domain.RuleConsecutiveToolErrors, alert.RuleName)
		}
		if alert.Severity != domain.LiveSeverityHigh {
			t.Errorf("expected severity HIGH, got %s", alert.Severity)
		}
		if alert.SessionID != "s1" {
			t.Errorf("expected session_id s1, got %s", alert.SessionID)
		}
		if !strings.Contains(alert.Message, "3 consecutive failures") {
			t.Errorf("expected message to mention 3 consecutive failures, got %s", alert.Message)
		}
		if !strings.Contains(alert.TriggerEvent, "p3") {
			t.Errorf("expected trigger event to contain part ID p3, got %s", alert.TriggerEvent)
		}
	})

	t.Run("4 consecutive tool errors trigger HIGH alert", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s1",
			Messages: []domain.LiveMessageFact{
				{
					ID:        "m1",
					Role:      "assistant",
					Timestamp: "2026-10-19T10:00:00Z",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("read"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("read"), IsError: true},
						{ID: "p3", Type: "tool", ToolName: strPtr("read"), IsError: true},
						{ID: "p4", Type: "tool", ToolName: strPtr("read"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateConsecutiveToolErrors(trace)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert, got %d", len(alerts))
		}
		if alerts[0].Severity != domain.LiveSeverityHigh {
			t.Errorf("expected severity HIGH for 4 errors, got %s", alerts[0].Severity)
		}
	})

	t.Run("5 consecutive tool errors trigger CRITICAL alert", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s1",
			Messages: []domain.LiveMessageFact{
				{
					ID:        "m1",
					Role:      "assistant",
					Timestamp: "2026-10-19T10:00:00Z",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("read"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("read"), IsError: true},
						{ID: "p3", Type: "tool", ToolName: strPtr("read"), IsError: true},
						{ID: "p4", Type: "tool", ToolName: strPtr("read"), IsError: true},
						{ID: "p5", Type: "tool", ToolName: strPtr("read"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateConsecutiveToolErrors(trace)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert, got %d", len(alerts))
		}
		if alerts[0].Severity != domain.LiveSeverityCritical {
			t.Errorf("expected severity CRITICAL for 5 errors, got %s", alerts[0].Severity)
		}
	})

	t.Run("streak resets on success: 2 errors, 1 success, 2 errors produces zero alerts", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s1",
			Messages: []domain.LiveMessageFact{
				{
					ID:        "m1",
					Role:      "assistant",
					Timestamp: "2026-10-19T10:00:00Z",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p3", Type: "tool", ToolName: strPtr("bash"), IsError: false},
						{ID: "p4", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p5", Type: "tool", ToolName: strPtr("bash"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateConsecutiveToolErrors(trace)
		if len(alerts) != 0 {
			t.Fatalf("expected 0 alerts after streak reset, got %d", len(alerts))
		}
	})

	t.Run("errors across multiple different tools accumulate in streak", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s1",
			Messages: []domain.LiveMessageFact{
				{
					ID:        "m1",
					Role:      "assistant",
					Timestamp: "2026-10-19T10:00:00Z",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("read"), IsError: true},
						{ID: "p3", Type: "tool", ToolName: strPtr("edit"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateConsecutiveToolErrors(trace)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert across different tools, got %d", len(alerts))
		}
		if alerts[0].Severity != domain.LiveSeverityHigh {
			t.Errorf("expected HIGH severity, got %s", alerts[0].Severity)
		}
		if !strings.Contains(alerts[0].Message, "edit") {
			t.Errorf("expected last failing tool 'edit' in message, got %s", alerts[0].Message)
		}
	})

	t.Run("nil trace or empty messages safely returns empty slice", func(t *testing.T) {
		if alerts := domain.EvaluateConsecutiveToolErrors(nil); len(alerts) != 0 {
			t.Errorf("expected 0 alerts for nil trace, got %d", len(alerts))
		}
		if alerts := domain.EvaluateConsecutiveToolErrors(&domain.LiveTraceView{}); len(alerts) != 0 {
			t.Errorf("expected 0 alerts for empty trace, got %d", len(alerts))
		}
	})
}

func TestContextCeilingApproachRule(t *testing.T) {
	testCases := []struct {
		name             string
		tokens           int64
		ceiling          int64
		expectedAlerts   int
		expectedSeverity string
	}{
		{
			name:           "0 tokens produces zero alerts and no division by zero",
			tokens:         0,
			ceiling:        200_000,
			expectedAlerts: 0,
		},
		{
			name:           "100k tokens (50% of ceiling) produces zero alerts",
			tokens:         100_000,
			ceiling:        200_000,
			expectedAlerts: 0,
		},
		{
			name:             "160k tokens (80% of ceiling) triggers WARNING alert",
			tokens:           160_000,
			ceiling:          200_000,
			expectedAlerts:   1,
			expectedSeverity: domain.LiveSeverityWarning,
		},
		{
			name:             "199k tokens (99.5% of ceiling) triggers WARNING alert",
			tokens:           199_000,
			ceiling:          200_000,
			expectedAlerts:   1,
			expectedSeverity: domain.LiveSeverityWarning,
		},
		{
			name:             "200k tokens (100% of ceiling) triggers CRITICAL alert",
			tokens:           200_000,
			ceiling:          200_000,
			expectedAlerts:   1,
			expectedSeverity: domain.LiveSeverityCritical,
		},
		{
			name:             "250k tokens (125% of ceiling) triggers CRITICAL alert",
			tokens:           250_000,
			ceiling:          200_000,
			expectedAlerts:   1,
			expectedSeverity: domain.LiveSeverityCritical,
		},
		{
			name:             "custom ceiling 100k: 80k triggers WARNING",
			tokens:           80_000,
			ceiling:          100_000,
			expectedAlerts:   1,
			expectedSeverity: domain.LiveSeverityWarning,
		},
		{
			name:             "custom ceiling 100k: 100k triggers CRITICAL",
			tokens:           100_000,
			ceiling:          100_000,
			expectedAlerts:   1,
			expectedSeverity: domain.LiveSeverityCritical,
		},
		{
			name:             "fallback ceiling when zero or negative passed",
			tokens:           160_000,
			ceiling:          0,
			expectedAlerts:   1,
			expectedSeverity: domain.LiveSeverityWarning,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trace := &domain.LiveTraceView{
				SessionID: "s_ctx",
				Messages: []domain.LiveMessageFact{
					{
						ID:            "m_ctx_1",
						Role:          "assistant",
						Timestamp:     "2026-10-19T10:00:00Z",
						ContextTokens: tc.tokens,
					},
				},
			}

			alerts := domain.EvaluateContextCeilingApproach(trace, tc.ceiling)
			if len(alerts) != tc.expectedAlerts {
				t.Fatalf("expected %d alerts, got %d", tc.expectedAlerts, len(alerts))
			}

			if tc.expectedAlerts > 0 {
				alert := alerts[0]
				if alert.RuleName != domain.RuleContextCeilingApproach {
					t.Errorf("expected rule %s, got %s", domain.RuleContextCeilingApproach, alert.RuleName)
				}
				if alert.Severity != tc.expectedSeverity {
					t.Errorf("expected severity %s, got %s", tc.expectedSeverity, alert.Severity)
				}
				if alert.SessionID != "s_ctx" {
					t.Errorf("expected session_id s_ctx, got %s", alert.SessionID)
				}
				if !strings.Contains(alert.Message, "Context window utilization reached") {
					t.Errorf("expected utilization message, got %s", alert.Message)
				}
			}
		})
	}
}

func TestTokenVelocityRunawayRule(t *testing.T) {
	t.Run("normal turns with 2000 tokens produce zero alerts", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_vel",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", TotalTokens: 2000, ContextTokens: 2000},
				{ID: "m2", Role: "assistant", TotalTokens: 2000, ContextTokens: 4000},
			},
		}

		alerts := domain.EvaluateTokenVelocityRunaway(trace, 0, 0)
		if len(alerts) != 0 {
			t.Fatalf("expected 0 alerts, got %d", len(alerts))
		}
	})

	t.Run("sudden surge turn with 35,000 tokens triggers WARNING alert", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_vel",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", TotalTokens: 2000, ContextTokens: 2000},
				{ID: "m2", Role: "assistant", TotalTokens: 35000, ContextTokens: 10000},
			},
		}

		alerts := domain.EvaluateTokenVelocityRunaway(trace, 0, 0)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert, got %d", len(alerts))
		}
		if alerts[0].Severity != domain.LiveSeverityWarning {
			t.Errorf("expected WARNING severity, got %s", alerts[0].Severity)
		}
		if !strings.Contains(alerts[0].Message, "Abnormally high token velocity") {
			t.Errorf("expected high token velocity message, got %s", alerts[0].Message)
		}
		if !strings.Contains(alerts[0].TriggerEvent, "turn 2") {
			t.Errorf("expected trigger event to reference turn 2, got %s", alerts[0].TriggerEvent)
		}
	})

	t.Run("extreme surge turn with 65,000 tokens triggers CRITICAL alert", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_vel",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", TotalTokens: 65000, ContextTokens: 65000},
			},
		}

		alerts := domain.EvaluateTokenVelocityRunaway(trace, 0, 0)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert, got %d", len(alerts))
		}
		if alerts[0].Severity != domain.LiveSeverityCritical {
			t.Errorf("expected CRITICAL severity, got %s", alerts[0].Severity)
		}
	})

	t.Run("turn with 0 tokens produces zero alerts and no panic", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_vel",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", TotalTokens: 0, ContextTokens: 0},
			},
		}

		alerts := domain.EvaluateTokenVelocityRunaway(trace, 0, 0)
		if len(alerts) != 0 {
			t.Fatalf("expected 0 alerts, got %d", len(alerts))
		}
	})

	t.Run("context surge delta greater than 40k triggers WARNING", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_vel",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", TotalTokens: 5000, ContextTokens: 10000},
				{ID: "m2", Role: "assistant", TotalTokens: 5000, ContextTokens: 55000},
			},
		}

		alerts := domain.EvaluateTokenVelocityRunaway(trace, 0, 0)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert for context delta, got %d", len(alerts))
		}
		if alerts[0].Severity != domain.LiveSeverityWarning {
			t.Errorf("expected WARNING, got %s", alerts[0].Severity)
		}
	})

	t.Run("context surge delta greater than 80k triggers CRITICAL", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_vel",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", TotalTokens: 5000, ContextTokens: 10000},
				{ID: "m2", Role: "assistant", TotalTokens: 5000, ContextTokens: 95000},
			},
		}

		alerts := domain.EvaluateTokenVelocityRunaway(trace, 0, 0)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert for critical delta, got %d", len(alerts))
		}
		if alerts[0].Severity != domain.LiveSeverityCritical {
			t.Errorf("expected CRITICAL, got %s", alerts[0].Severity)
		}
	})

	t.Run("negative context delta from compaction is safely ignored", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_vel",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", TotalTokens: 5000, ContextTokens: 100000},
				{ID: "m2", Role: "assistant", TotalTokens: 2000, ContextTokens: 20000},
			},
		}

		alerts := domain.EvaluateTokenVelocityRunaway(trace, 0, 0)
		if len(alerts) != 0 {
			t.Fatalf("expected 0 alerts for compaction reduction, got %d", len(alerts))
		}
	})
}

func TestToolThrashingRule(t *testing.T) {
	t.Run("same tool called 3 times successfully produces zero alerts", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_thrash",
			Messages: []domain.LiveMessageFact{
				{
					ID: "m1",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: false},
						{ID: "p2", Type: "tool", ToolName: strPtr("bash"), IsError: false},
						{ID: "p3", Type: "tool", ToolName: strPtr("bash"), IsError: false},
					},
				},
			},
		}

		alerts := domain.EvaluateToolThrashing(trace)
		if len(alerts) != 0 {
			t.Fatalf("expected 0 alerts, got %d", len(alerts))
		}
	})

	t.Run("different tools failing once each produces zero thrashing alerts", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_thrash",
			Messages: []domain.LiveMessageFact{
				{
					ID: "m1",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("read"), IsError: true},
						{ID: "p3", Type: "tool", ToolName: strPtr("edit"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateToolThrashing(trace)
		if len(alerts) != 0 {
			t.Fatalf("expected 0 alerts for different tools, got %d", len(alerts))
		}
	})

	t.Run("same tool failing 3 times triggers WARNING thrashing alert", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_thrash",
			Messages: []domain.LiveMessageFact{
				{
					ID: "m1",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p3", Type: "tool", ToolName: strPtr("bash"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateToolThrashing(trace)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 thrashing alert, got %d", len(alerts))
		}
		if alerts[0].RuleName != domain.RuleToolThrashing {
			t.Errorf("expected rule %s, got %s", domain.RuleToolThrashing, alerts[0].RuleName)
		}
		if alerts[0].Severity != domain.LiveSeverityWarning {
			t.Errorf("expected WARNING severity, got %s", alerts[0].Severity)
		}
		if !strings.Contains(alerts[0].Message, "Tool thrashing detected on tool bash with 3 failing invocations") {
			t.Errorf("expected message to mention thrashing, got %s", alerts[0].Message)
		}
		if !strings.Contains(alerts[0].TriggerEvent, "p1, p2, p3") {
			t.Errorf("expected trigger event to list part IDs, got %s", alerts[0].TriggerEvent)
		}
	})

	t.Run("same tool failing 5 times triggers CRITICAL thrashing alert", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_thrash",
			Messages: []domain.LiveMessageFact{
				{
					ID: "m1",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("git"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("git"), IsError: true},
						{ID: "p3", Type: "tool", ToolName: strPtr("git"), IsError: true},
						{ID: "p4", Type: "tool", ToolName: strPtr("git"), IsError: true},
						{ID: "p5", Type: "tool", ToolName: strPtr("git"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateToolThrashing(trace)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert, got %d", len(alerts))
		}
		if alerts[0].Severity != domain.LiveSeverityCritical {
			t.Errorf("expected CRITICAL severity, got %s", alerts[0].Severity)
		}
	})

	t.Run("tool fails 2 times, succeeds once, fails once produces no thrashing alert", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_thrash",
			Messages: []domain.LiveMessageFact{
				{
					ID: "m1",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p3", Type: "tool", ToolName: strPtr("bash"), IsError: false},
						{ID: "p4", Type: "tool", ToolName: strPtr("bash"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateToolThrashing(trace)
		if len(alerts) != 0 {
			t.Fatalf("expected 0 alerts when intermediate success occurred, got %d", len(alerts))
		}
	})

	t.Run("tool fails 3 times within last 5 calls without progress triggers alert", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_thrash",
			Messages: []domain.LiveMessageFact{
				{
					ID: "m1",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("read"), IsError: true},
						{ID: "p3", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p4", Type: "tool", ToolName: strPtr("edit"), IsError: true},
						{ID: "p5", Type: "tool", ToolName: strPtr("bash"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateToolThrashing(trace)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert for window thrashing, got %d", len(alerts))
		}
		if alerts[0].Severity != domain.LiveSeverityWarning {
			t.Errorf("expected WARNING severity, got %s", alerts[0].Severity)
		}
	})
}

func TestCompositeEvaluateLiveAnomalies(t *testing.T) {
	t.Run("nil tree and nil trace returns empty slice without panicking", func(t *testing.T) {
		alerts := domain.EvaluateLiveAnomalies(nil, nil)
		if alerts == nil {
			t.Fatal("expected non-nil empty slice")
		}
		if len(alerts) != 0 {
			t.Fatalf("expected 0 alerts, got %d", len(alerts))
		}
	})

	t.Run("populated tree and nil trace handles cleanly", func(t *testing.T) {
		tree := &domain.LiveSessionTree{
			ActiveAnomalies: []domain.LiveAnomalyAlert{
				{SessionID: "s1", RuleName: "TestRule", Severity: domain.LiveSeverityWarning},
			},
		}
		alerts := domain.EvaluateLiveAnomalies(tree, nil)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert, got %d", len(alerts))
		}
	})

	t.Run("nil tree and populated trace executes cleanly", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s1",
			Messages: []domain.LiveMessageFact{
				{
					ID:            "m1",
					Role:          "assistant",
					ContextTokens: 250_000,
				},
			},
		}
		alerts := domain.EvaluateLiveAnomalies(nil, trace)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert, got %d", len(alerts))
		}
		if alerts[0].RuleName != domain.RuleContextCeilingApproach {
			t.Errorf("expected ContextCeilingApproach rule, got %s", alerts[0].RuleName)
		}
	})

	t.Run("multiple rules triggered simultaneously are deterministically sorted", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_multi",
			Messages: []domain.LiveMessageFact{
				{
					ID:            "m1",
					Role:          "assistant",
					Timestamp:     "2026-10-19T10:01:00Z",
					ContextTokens: 165_000, // triggers ContextCeilingApproach: WARNING
					TotalTokens:   70_000,  // triggers TokenVelocityRunaway: CRITICAL
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p3", Type: "tool", ToolName: strPtr("bash"), IsError: true}, // triggers ConsecutiveToolErrors (HIGH) and ToolThrashing (WARNING)
					},
				},
			},
		}

		alerts := domain.EvaluateLiveAnomalies(nil, trace)
		if len(alerts) < 3 {
			t.Fatalf("expected at least 3 alerts, got %d", len(alerts))
		}

		// First alert must be CRITICAL
		if alerts[0].Severity != domain.LiveSeverityCritical {
			t.Errorf("expected first alert to be CRITICAL, got %s", alerts[0].Severity)
		}

		// Subsequent alert must be HIGH
		if alerts[1].Severity != domain.LiveSeverityHigh {
			t.Errorf("expected second alert to be HIGH, got %s", alerts[1].Severity)
		}

		// Remaining alerts must be WARNING
		for i := 2; i < len(alerts); i++ {
			if alerts[i].Severity != domain.LiveSeverityWarning {
				t.Errorf("expected alert %d to be WARNING, got %s", i, alerts[i].Severity)
			}
		}

		// Verify deterministic secondary sort by RuleName ascending among WARNINGs
		for i := 2; i < len(alerts)-1; i++ {
			if alerts[i].Severity == alerts[i+1].Severity {
				if alerts[i].RuleName > alerts[i+1].RuleName {
					t.Errorf("expected RuleName ascending: %s > %s", alerts[i].RuleName, alerts[i+1].RuleName)
				}
			}
		}
	})
}

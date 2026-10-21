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
		if alert.Count == nil || *alert.Count != 3 {
			t.Errorf("expected alert Count to be 3, got %v", alert.Count)
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
		if alerts[0].Count == nil || *alerts[0].Count != 4 {
			t.Errorf("expected alert Count to be 4, got %v", alerts[0].Count)
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
		if alerts[0].Count == nil || *alerts[0].Count != 5 {
			t.Errorf("expected alert Count to be 5, got %v", alerts[0].Count)
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
		if alerts[0].Count == nil || *alerts[0].Count != 3 {
			t.Errorf("expected alert Count to be 3 across different tools, got %v", alerts[0].Count)
		}
	})

	t.Run("consecutive tool errors span multiple messages and record total count", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s-multi-msg",
			Messages: []domain.LiveMessageFact{
				{
					ID: "m1",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("bash"), IsError: true},
					},
				},
				{
					ID: "m2",
					Parts: []domain.LivePartFact{
						{ID: "p3", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p4", Type: "tool", ToolName: strPtr("bash"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateConsecutiveToolErrors(trace)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert spanning 2 messages, got %d", len(alerts))
		}
		if alerts[0].Count == nil || *alerts[0].Count != 4 {
			t.Errorf("expected alert Count to be 4 across messages, got %v", alerts[0].Count)
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
		if alerts[0].Count == nil || *alerts[0].Count != 3 {
			t.Errorf("expected alert Count to be 3, got %v", alerts[0].Count)
		}
	})

	t.Run("tool fails 4 times without progress triggers alert with count 4", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_thrash_4",
			Messages: []domain.LiveMessageFact{
				{
					ID: "m1",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p3", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p4", Type: "tool", ToolName: strPtr("bash"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateToolThrashing(trace)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert for 4 thrashing failures, got %d", len(alerts))
		}
		if alerts[0].Count == nil || *alerts[0].Count != 4 {
			t.Errorf("expected alert Count to be 4, got %v", alerts[0].Count)
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

func TestConsecutiveToolErrorsRule_Adversarial(t *testing.T) {
	t.Run("non-tool parts interleaved between tool errors do not break error streak", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_interleaved",
			Messages: []domain.LiveMessageFact{
				{
					ID:        "m1",
					Role:      "assistant",
					Timestamp: "2026-10-19T10:00:00Z",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p2", Type: "text", OutputSnippet: "thinking about next step"},
						{ID: "p3", Type: "reasoning", OutputSnippet: "let me analyze the error"},
						{ID: "p4", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p5", Type: "subtask", InputSnippet: "subtask info"},
						{ID: "p6", Type: "tool", ToolName: strPtr("bash"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateConsecutiveToolErrors(trace)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert despite non-tool parts in streak, got %d", len(alerts))
		}
		if alerts[0].Severity != domain.LiveSeverityHigh {
			t.Errorf("expected HIGH severity, got %s", alerts[0].Severity)
		}
	})

	t.Run("nil or whitespace tool name defaults safely to unknown", func(t *testing.T) {
		emptyTool := "   "
		trace := &domain.LiveTraceView{
			SessionID: "s_nil_tool",
			Messages: []domain.LiveMessageFact{
				{
					ID:        "m1",
					Role:      "assistant",
					Timestamp: "2026-10-19T10:00:00Z",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: nil, IsError: true},
						{ID: "p2", Type: "tool", ToolName: &emptyTool, IsError: true},
						{ID: "p3", Type: "tool", ToolName: nil, IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateConsecutiveToolErrors(trace)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert, got %d", len(alerts))
		}
		if !strings.Contains(alerts[0].Message, "last failing tool: unknown") {
			t.Errorf("expected message to mention unknown tool, got %s", alerts[0].Message)
		}
	})

	t.Run("long error message is truncated to 256 characters in trigger event", func(t *testing.T) {
		hugeError := strings.Repeat("F", 500)
		trace := &domain.LiveTraceView{
			SessionID: "s_long_err",
			Messages: []domain.LiveMessageFact{
				{
					ID:        "m1",
					Role:      "assistant",
					Timestamp: "2026-10-19T10:00:00Z",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p3", Type: "tool", ToolName: strPtr("bash"), IsError: true, ErrorMessage: &hugeError},
					},
				},
			},
		}

		alerts := domain.EvaluateConsecutiveToolErrors(trace)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert, got %d", len(alerts))
		}
		// trigger_event format is "p3: <snippet up to 256 chars>"
		expectedPrefix := "p3: "
		if !strings.HasPrefix(alerts[0].TriggerEvent, expectedPrefix) {
			t.Errorf("expected prefix %q, got %s", expectedPrefix, alerts[0].TriggerEvent)
		}
		snippet := strings.TrimPrefix(alerts[0].TriggerEvent, expectedPrefix)
		if len(snippet) > 256 {
			t.Errorf("expected snippet length <= 256, got %d", len(snippet))
		}
	})

	t.Run("consecutive errors across multiple messages", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_multi_msg",
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
				{
					ID:        "m2",
					Role:      "assistant",
					Timestamp: "2026-10-19T10:01:00Z",
					Parts: []domain.LivePartFact{
						{ID: "p3", Type: "tool", ToolName: strPtr("bash"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateConsecutiveToolErrors(trace)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert across messages, got %d", len(alerts))
		}
		if alerts[0].Timestamp != "2026-10-19T10:01:00Z" {
			t.Errorf("expected timestamp of latest failing message, got %s", alerts[0].Timestamp)
		}
	})

	t.Run("post recovery single failure after streak reset does not alert", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_post_rec",
			Messages: []domain.LiveMessageFact{
				{
					ID: "m1",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p3", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p4", Type: "tool", ToolName: strPtr("bash"), IsError: false}, // recovery
						{ID: "p5", Type: "tool", ToolName: strPtr("bash"), IsError: true},  // streak = 1
					},
				},
			},
		}

		alerts := domain.EvaluateConsecutiveToolErrors(trace)
		if len(alerts) != 0 {
			t.Fatalf("expected 0 alerts after streak was reset by recovery, got %d", len(alerts))
		}
	})
}

func TestContextCeilingApproachRule_Adversarial(t *testing.T) {
	t.Run("exact boundary: 159999 tokens on 200k ceiling produces 0 alerts", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_boundary",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", ContextTokens: 159_999},
			},
		}
		alerts := domain.EvaluateContextCeilingApproach(trace, 200_000)
		if len(alerts) != 0 {
			t.Fatalf("expected 0 alerts for 159999 tokens, got %d", len(alerts))
		}
	})

	t.Run("exact boundary: 160000 tokens on 200k ceiling produces WARNING", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_boundary",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", ContextTokens: 160_000},
			},
		}
		alerts := domain.EvaluateContextCeilingApproach(trace, 200_000)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert for 160000 tokens, got %d", len(alerts))
		}
		if alerts[0].Severity != domain.LiveSeverityWarning {
			t.Errorf("expected WARNING, got %s", alerts[0].Severity)
		}
	})

	t.Run("exact boundary: 199999 tokens on 200k ceiling produces WARNING", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_boundary",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", ContextTokens: 199_999},
			},
		}
		alerts := domain.EvaluateContextCeilingApproach(trace, 200_000)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert, got %d", len(alerts))
		}
		if alerts[0].Severity != domain.LiveSeverityWarning {
			t.Errorf("expected WARNING, got %s", alerts[0].Severity)
		}
	})

	t.Run("exact boundary: 200000 tokens on 200k ceiling produces CRITICAL", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_boundary",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", ContextTokens: 200_000},
			},
		}
		alerts := domain.EvaluateContextCeilingApproach(trace, 200_000)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert, got %d", len(alerts))
		}
		if alerts[0].Severity != domain.LiveSeverityCritical {
			t.Errorf("expected CRITICAL, got %s", alerts[0].Severity)
		}
	})

	t.Run("negative ceiling tokens safely falls back to default 200k ceiling", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_neg_ceil",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", ContextTokens: 160_000},
			},
		}
		alerts := domain.EvaluateContextCeilingApproach(trace, -50_000)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert with fallback ceiling, got %d", len(alerts))
		}
		if alerts[0].Severity != domain.LiveSeverityWarning {
			t.Errorf("expected WARNING with fallback ceiling, got %s", alerts[0].Severity)
		}
	})

	t.Run("first breach message is selected for timestamp and triggerEvent", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_first_breach",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Timestamp: "2026-10-19T10:00:00Z", ContextTokens: 100_000},
				{ID: "m2", Timestamp: "2026-10-19T10:05:00Z", ContextTokens: 170_000}, // First breach of 160k
				{ID: "m3", Timestamp: "2026-10-19T10:10:00Z", ContextTokens: 190_000}, // Peak
			},
		}
		alerts := domain.EvaluateContextCeilingApproach(trace, 200_000)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert, got %d", len(alerts))
		}
		if alerts[0].TriggerEvent != "m2" {
			t.Errorf("expected trigger event m2 (first breach), got %s", alerts[0].TriggerEvent)
		}
		if alerts[0].Timestamp != "2026-10-19T10:05:00Z" {
			t.Errorf("expected timestamp of first breach m2, got %s", alerts[0].Timestamp)
		}
	})
}

func TestTokenVelocityRunawayRule_Adversarial(t *testing.T) {
	t.Run("exact boundary for total tokens: 30000 produces 0 alerts, 30001 produces WARNING", func(t *testing.T) {
		trace30k := &domain.LiveTraceView{
			SessionID: "s_vel_b",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", TotalTokens: 30_000},
			},
		}
		alerts30k := domain.EvaluateTokenVelocityRunaway(trace30k, 0, 0)
		if len(alerts30k) != 0 {
			t.Fatalf("expected 0 alerts for exactly 30000 tokens, got %d", len(alerts30k))
		}

		trace30k1 := &domain.LiveTraceView{
			SessionID: "s_vel_b",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", TotalTokens: 30_001},
			},
		}
		alerts30k1 := domain.EvaluateTokenVelocityRunaway(trace30k1, 0, 0)
		if len(alerts30k1) != 1 {
			t.Fatalf("expected 1 alert for 30001 tokens, got %d", len(alerts30k1))
		}
		if alerts30k1[0].Severity != domain.LiveSeverityWarning {
			t.Errorf("expected WARNING, got %s", alerts30k1[0].Severity)
		}
	})

	t.Run("exact boundary for total tokens: 60000 produces WARNING, 60001 produces CRITICAL", func(t *testing.T) {
		trace60k := &domain.LiveTraceView{
			SessionID: "s_vel_b",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", TotalTokens: 60_000},
			},
		}
		alerts60k := domain.EvaluateTokenVelocityRunaway(trace60k, 0, 0)
		if len(alerts60k) != 1 || alerts60k[0].Severity != domain.LiveSeverityWarning {
			t.Fatalf("expected WARNING for exactly 60000 tokens, got %+v", alerts60k)
		}

		trace60k1 := &domain.LiveTraceView{
			SessionID: "s_vel_b",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", TotalTokens: 60_001},
			},
		}
		alerts60k1 := domain.EvaluateTokenVelocityRunaway(trace60k1, 0, 0)
		if len(alerts60k1) != 1 || alerts60k1[0].Severity != domain.LiveSeverityCritical {
			t.Fatalf("expected CRITICAL for 60001 tokens, got %+v", alerts60k1)
		}
	})

	t.Run("exact boundary for context delta: 40000 produces 0 alerts, 40001 produces WARNING", func(t *testing.T) {
		trace40k := &domain.LiveTraceView{
			SessionID: "s_vel_d",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", ContextTokens: 10_000, TotalTokens: 5_000},
				{ID: "m2", Role: "assistant", ContextTokens: 50_000, TotalTokens: 5_000}, // delta 40k
			},
		}
		alerts40k := domain.EvaluateTokenVelocityRunaway(trace40k, 0, 0)
		if len(alerts40k) != 0 {
			t.Fatalf("expected 0 alerts for exactly 40k delta, got %d", len(alerts40k))
		}

		trace40k1 := &domain.LiveTraceView{
			SessionID: "s_vel_d",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", ContextTokens: 10_000, TotalTokens: 5_000},
				{ID: "m2", Role: "assistant", ContextTokens: 50_001, TotalTokens: 5_000}, // delta 40001
			},
		}
		alerts40k1 := domain.EvaluateTokenVelocityRunaway(trace40k1, 0, 0)
		if len(alerts40k1) != 1 {
			t.Fatalf("expected 1 alert for 40001 delta, got %d", len(alerts40k1))
		}
		if alerts40k1[0].Severity != domain.LiveSeverityWarning {
			t.Errorf("expected WARNING, got %s", alerts40k1[0].Severity)
		}
	})

	t.Run("exact boundary for context delta: 80000 produces WARNING, 80001 produces CRITICAL", func(t *testing.T) {
		trace80k := &domain.LiveTraceView{
			SessionID: "s_vel_d",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", ContextTokens: 10_000, TotalTokens: 5_000},
				{ID: "m2", Role: "assistant", ContextTokens: 90_000, TotalTokens: 5_000}, // delta 80k
			},
		}
		alerts80k := domain.EvaluateTokenVelocityRunaway(trace80k, 0, 0)
		if len(alerts80k) != 1 || alerts80k[0].Severity != domain.LiveSeverityWarning {
			t.Fatalf("expected WARNING for 80k delta, got %+v", alerts80k)
		}

		trace80k1 := &domain.LiveTraceView{
			SessionID: "s_vel_d",
			Messages: []domain.LiveMessageFact{
				{ID: "m1", Role: "assistant", ContextTokens: 10_000, TotalTokens: 5_000},
				{ID: "m2", Role: "assistant", ContextTokens: 90_001, TotalTokens: 5_000}, // delta 80001
			},
		}
		alerts80k1 := domain.EvaluateTokenVelocityRunaway(trace80k1, 0, 0)
		if len(alerts80k1) != 1 || alerts80k1[0].Severity != domain.LiveSeverityCritical {
			t.Fatalf("expected CRITICAL for 80001 delta, got %+v", alerts80k1)
		}
	})

	t.Run("user and system messages are ignored for velocity checks", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_roles",
			Messages: []domain.LiveMessageFact{
				{ID: "m_user", Role: "user", TotalTokens: 100_000, ContextTokens: 100_000},
				{ID: "m_sys", Role: "system", TotalTokens: 100_000, ContextTokens: 100_000},
			},
		}
		alerts := domain.EvaluateTokenVelocityRunaway(trace, 0, 0)
		if len(alerts) != 0 {
			t.Fatalf("expected 0 alerts for non-assistant roles, got %d", len(alerts))
		}
	})
}

func TestToolThrashingRule_Adversarial(t *testing.T) {
	t.Run("two distinct tools thrashing simultaneously are sorted alphabetically", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_dual_thrash",
			Messages: []domain.LiveMessageFact{
				{
					ID: "m1",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("read"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p3", Type: "tool", ToolName: strPtr("read"), IsError: true},
						{ID: "p4", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p5", Type: "tool", ToolName: strPtr("read"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateToolThrashing(trace)
		// "read" has 3 failures in 5 calls (WARNING)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert for read thrashing, got %d", len(alerts))
		}
		if alerts[0].RuleName != domain.RuleToolThrashing {
			t.Errorf("expected ToolThrashing, got %s", alerts[0].RuleName)
		}
		if !strings.Contains(alerts[0].Message, "tool read") {
			t.Errorf("expected message to mention tool read, got %s", alerts[0].Message)
		}
	})

	t.Run("nil tool name in thrashing trace defaults to unknown", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_nil_thrash",
			Messages: []domain.LiveMessageFact{
				{
					ID: "m1",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: nil, IsError: true},
						{ID: "p2", Type: "tool", ToolName: nil, IsError: true},
						{ID: "p3", Type: "tool", ToolName: nil, IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateToolThrashing(trace)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 alert for unknown tool thrashing, got %d", len(alerts))
		}
		if !strings.Contains(alerts[0].Message, "unknown") {
			t.Errorf("expected message to identify unknown tool, got %s", alerts[0].Message)
		}
	})

	t.Run("fewer than 3 tool calls total produces 0 alerts", func(t *testing.T) {
		trace := &domain.LiveTraceView{
			SessionID: "s_few_calls",
			Messages: []domain.LiveMessageFact{
				{
					ID: "m1",
					Parts: []domain.LivePartFact{
						{ID: "p1", Type: "tool", ToolName: strPtr("bash"), IsError: true},
						{ID: "p2", Type: "tool", ToolName: strPtr("bash"), IsError: true},
					},
				},
			},
		}

		alerts := domain.EvaluateToolThrashing(trace)
		if len(alerts) != 0 {
			t.Fatalf("expected 0 alerts for 2 calls, got %d", len(alerts))
		}
	})
}

func TestSortLiveAnomalies_MultiKeyOrdering(t *testing.T) {
	alerts := []domain.LiveAnomalyAlert{
		{
			SessionID:    "ses_b",
			RuleName:     domain.RuleTokenVelocityRunaway,
			Severity:     domain.LiveSeverityWarning,
			Timestamp:    "2026-10-19T10:00:00Z",
			TriggerEvent: "turn 1",
		},
		{
			SessionID:    "ses_a",
			RuleName:     domain.RuleConsecutiveToolErrors,
			Severity:     domain.LiveSeverityCritical,
			Timestamp:    "2026-10-19T10:01:00Z",
			TriggerEvent: "part 1",
		},
		{
			SessionID:    "ses_c",
			RuleName:     domain.RuleContextCeilingApproach,
			Severity:     domain.LiveSeverityHigh,
			Timestamp:    "2026-10-19T10:02:00Z",
			TriggerEvent: "turn 2",
		},
		{
			SessionID:    "ses_d",
			RuleName:     domain.RuleConsecutiveToolErrors,
			Severity:     domain.LiveSeverityWarning,
			Timestamp:    "2026-10-19T10:00:00Z",
			TriggerEvent: "part 2",
		},
		{
			SessionID:    "ses_a",
			RuleName:     domain.RuleConsecutiveToolErrors,
			Severity:     domain.LiveSeverityWarning,
			Timestamp:    "2026-10-19T10:05:00Z", // Newer timestamp
			TriggerEvent: "part 3",
		},
	}

	domain.SortLiveAnomalies(alerts)

	// 1. First alert must be CRITICAL
	if alerts[0].Severity != domain.LiveSeverityCritical {
		t.Fatalf("expected alerts[0] to be CRITICAL, got %s", alerts[0].Severity)
	}

	// 2. Second alert must be HIGH
	if alerts[1].Severity != domain.LiveSeverityHigh {
		t.Fatalf("expected alerts[1] to be HIGH, got %s", alerts[1].Severity)
	}

	// 3. Third alert must be WARNING with RuleConsecutiveToolErrors and newer timestamp (10:05:00Z)
	if alerts[2].Severity != domain.LiveSeverityWarning || alerts[2].Timestamp != "2026-10-19T10:05:00Z" {
		t.Errorf("expected alerts[2] to be WARNING with newer timestamp, got %+v", alerts[2])
	}

	// 4. Fourth alert must be WARNING with RuleConsecutiveToolErrors (10:00:00Z)
	if alerts[3].RuleName != domain.RuleConsecutiveToolErrors {
		t.Errorf("expected alerts[3] to be ConsecutiveToolErrors, got %s", alerts[3].RuleName)
	}

	// 5. Fifth alert must be TokenVelocityRunaway (RuleName ascending after ConsecutiveToolErrors)
	if alerts[4].RuleName != domain.RuleTokenVelocityRunaway {
		t.Errorf("expected alerts[4] to be TokenVelocityRunaway, got %s", alerts[4].RuleName)
	}
}

func TestCompositeEvaluateLiveAnomalies_DeepHierarchyAndDeduplication(t *testing.T) {
	t.Run("nested sub-sessions anomalies are collected recursively", func(t *testing.T) {
		grandchild := &domain.LiveSessionNode{
			ID: "ses_gc",
			Anomalies: []domain.LiveAnomalyAlert{
				{SessionID: "ses_gc", RuleName: "GCRule", Severity: domain.LiveSeverityWarning, Timestamp: "2026-10-19T10:00:00Z"},
			},
		}
		child := &domain.LiveSessionNode{
			ID: "ses_child",
			Anomalies: []domain.LiveAnomalyAlert{
				{SessionID: "ses_child", RuleName: "ChildRule", Severity: domain.LiveSeverityHigh, Timestamp: "2026-10-19T10:01:00Z"},
			},
			SubSessions: []*domain.LiveSessionNode{grandchild, nil},
		}
		root := &domain.LiveSessionNode{
			ID: "ses_root",
			Anomalies: []domain.LiveAnomalyAlert{
				{SessionID: "ses_root", RuleName: "RootRule", Severity: domain.LiveSeverityCritical, Timestamp: "2026-10-19T10:02:00Z"},
			},
			SubSessions: []*domain.LiveSessionNode{child},
		}

		tree := &domain.LiveSessionTree{
			RootSessions: []*domain.LiveSessionNode{root, nil},
		}

		alerts := domain.EvaluateLiveAnomalies(tree, nil)
		if len(alerts) != 3 {
			t.Fatalf("expected 3 alerts collected recursively, got %d", len(alerts))
		}
		if alerts[0].Severity != domain.LiveSeverityCritical {
			t.Errorf("expected CRITICAL first, got %s", alerts[0].Severity)
		}
		if alerts[1].Severity != domain.LiveSeverityHigh {
			t.Errorf("expected HIGH second, got %s", alerts[1].Severity)
		}
		if alerts[2].Severity != domain.LiveSeverityWarning {
			t.Errorf("expected WARNING third, got %s", alerts[2].Severity)
		}
	})

	t.Run("exact duplicate alerts are deduplicated", func(t *testing.T) {
		dupAlert := domain.LiveAnomalyAlert{
			SessionID:    "s_dup",
			RuleName:     domain.RuleContextCeilingApproach,
			Severity:     domain.LiveSeverityCritical,
			Timestamp:    "2026-10-19T10:00:00Z",
			TriggerEvent: "m1",
		}

		tree := &domain.LiveSessionTree{
			ActiveAnomalies: []domain.LiveAnomalyAlert{dupAlert, dupAlert},
			RootSessions: []*domain.LiveSessionNode{
				{
					ID:        "s_dup",
					Anomalies: []domain.LiveAnomalyAlert{dupAlert},
				},
			},
		}

		alerts := domain.EvaluateLiveAnomalies(tree, nil)
		if len(alerts) != 1 {
			t.Fatalf("expected 1 deduplicated alert, got %d", len(alerts))
		}
	})

	t.Run("LiveAnomalyRuleName typed constants and struct field verification", func(t *testing.T) {
		canonicalRules := []domain.LiveAnomalyRuleName{
			domain.RuleConsecutiveToolErrors,
			domain.RuleContextCeilingApproach,
			domain.RuleTokenVelocityRunaway,
			domain.RuleToolThrashing,
		}

		expectedValues := []string{
			"ConsecutiveToolErrors",
			"ContextCeilingApproach",
			"TokenVelocityRunaway",
			"ToolThrashing",
		}

		for i, rule := range canonicalRules {
			if string(rule) != expectedValues[i] {
				t.Errorf("expected rule value %q, got %q", expectedValues[i], string(rule))
			}
		}

		alert := domain.LiveAnomalyAlert{
			SessionID:    "s1",
			RuleName:     domain.RuleConsecutiveToolErrors,
			Severity:     domain.LiveSeverityHigh,
			Message:      "test",
			TriggerEvent: "p1",
		}

		var typedRule domain.LiveAnomalyRuleName = alert.RuleName
		if typedRule != domain.RuleConsecutiveToolErrors {
			t.Errorf("expected alert.RuleName to have type LiveAnomalyRuleName with value %s", domain.RuleConsecutiveToolErrors)
		}
	})
}

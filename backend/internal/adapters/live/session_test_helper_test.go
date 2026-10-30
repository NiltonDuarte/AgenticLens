package live_test

import (
	"agenticlens/internal/domain"
	"fmt"
)

func sessionFacts(invocations []domain.InvocationFact, tools []domain.ToolFact, events []domain.EventFact) []domain.SessionFact {
	for i := range invocations {
		if invocations[i].Model == "" {
			invocations[i].Model = "fixture-model"
		}
		if invocations[i].StartedAt == nil {
			value := "fixture-started"
			invocations[i].StartedAt = &value
		}
		if invocations[i].CompletedAt == nil {
			value := "fixture-completed"
			invocations[i].CompletedAt = &value
		}
	}
	for i := range tools {
		if tools[i].SessionID == "" {
			tools[i].SessionID = "fixture-session"
		}
	}
	for i := range events {
		if events[i].SessionID == "" {
			events[i].SessionID = "fixture-session"
		}
		if events[i].Timestamp == "" {
			events[i].Timestamp = fmt.Sprintf("fixture-%06d", i)
		}
	}
	sessions, err := domain.SessionsFromFacts(fixtureInvocationSessions(invocations), tools, events)
	if err != nil {
		panic(err)
	}
	for i := range sessions {
		if sessions[i].User == "" {
			sessions[i].User = "fixture-user"
		}
		if sessions[i].Anomalies == nil {
			sessions[i].Anomalies = []domain.Anomaly{}
		}
	}
	return sessions
}

func sessionFactsWithSource(invocations []domain.InvocationFact, tools []domain.ToolFact, events []domain.EventFact, source domain.SourceHarness) []domain.SessionFact {
	for i := range invocations {
		if invocations[i].Model == "" {
			invocations[i].Model = "fixture-model"
		}
		if invocations[i].StartedAt == nil {
			value := "fixture-started"
			invocations[i].StartedAt = &value
		}
		if invocations[i].CompletedAt == nil {
			value := "fixture-completed"
			invocations[i].CompletedAt = &value
		}
	}
	for i := range tools {
		if tools[i].SessionID == "" {
			tools[i].SessionID = "fixture-session"
		}
	}
	for i := range events {
		if events[i].SessionID == "" {
			events[i].SessionID = "fixture-session"
		}
		if events[i].Timestamp == "" {
			events[i].Timestamp = fmt.Sprintf("fixture-%06d", i)
		}
	}
	sessions, err := domain.SessionsFromFactsWithSource(source, fixtureInvocationSessions(invocations), tools, events)
	if err != nil {
		panic(err)
	}
	for i := range sessions {
		if sessions[i].User == "" {
			sessions[i].User = "fixture-user"
		}
		if sessions[i].Anomalies == nil {
			sessions[i].Anomalies = []domain.Anomaly{}
		}
	}
	return sessions
}

func setEventFacts(ledger *domain.AgenticFactLedger, events []domain.EventFact) {
	for i := range events {
		if events[i].SessionID == "" {
			events[i].SessionID = "fixture-session"
		}
		if events[i].Timestamp == "" {
			events[i].Timestamp = fmt.Sprintf("fixture-%06d", i)
		}
	}
	sessions, err := domain.SessionsFromFacts(ledger.Sessions, ledger.ToolFacts(), events)
	if err != nil {
		panic(err)
	}
	for i := range sessions {
		if sessions[i].Anomalies == nil {
			sessions[i].Anomalies = []domain.Anomaly{}
		}
	}
	ledger.Sessions = sessions
}

func fixtureInvocationSessions(invocations []domain.InvocationFact) []domain.SessionFact {
	if len(invocations) == 0 {
		return nil
	}
	return []domain.SessionFact{{ID: "fixture-session", Source: domain.SourceOpenCode, User: "fixture-user", Invocations: invocations}}
}

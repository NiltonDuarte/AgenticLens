package backfill

import (
	"agenticlens/internal/domain"
	"fmt"
)

func sessionFacts(invocations []domain.InvocationFact, tools []domain.ToolFact, events []domain.EventFact) []domain.SessionFact {
	for i := range invocations {
		if invocations[i].ID == "" {
			invocations[i].ID = fmt.Sprintf("fixture-session-%d", i)
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
	sessions, err := domain.SessionsFromFacts(invocations, tools, events)
	if err != nil {
		panic(err)
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
	sessions, err := domain.SessionsFromFacts(ledger.InvocationFacts(), ledger.ToolFacts(), events)
	if err != nil {
		panic(err)
	}
	ledger.Sessions = sessions
}

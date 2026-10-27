package domain_test

import (
	"testing"

	"agenticlens/internal/domain"
)

func TestSessionsFromFactsAggregatesDuplicateToolNamesWithinSession(t *testing.T) {
	sessions, err := domain.SessionsFromFacts(nil, []domain.ToolFact{
		{SessionID: "session-a", Name: "bash", Calls: 3, Errors: 1},
		{SessionID: "session-a", Name: "bash", Calls: 2, Errors: 1},
		{SessionID: "session-b", Name: "bash", Calls: 4, Errors: 0},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 || len(sessions[0].Tools) != 1 || sessions[0].Tools[0].Calls != 5 || sessions[0].Tools[0].Errors != 2 {
		t.Fatalf("same-session rows must aggregate once by tool name without crossing sessions: %+v", sessions)
	}
	if sessions[1].Tools[0].Calls != 4 || sessions[1].Tools[0].Errors != 0 {
		t.Fatalf("other session totals changed: %+v", sessions[1])
	}
}

package adapters_test

import (
	"agenticlens/internal/domain"
	"fmt"
)

func sessionFacts(invocations []domain.InvocationFact, tools []domain.ToolFact, events []domain.EventFact) []domain.SessionFact {
	for i := range invocations {
		if invocations[i].ID == "" {
			invocations[i].ID = fmt.Sprintf("fixture-session-%d", i)
		}
		if invocations[i].Role == "" {
			invocations[i].Role = "orchestrator"
		}
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
	sessions, err := domain.SessionsFromFacts(invocations, tools, events)
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

type testInvFact struct {
	ID                    string
	ParentID              *string
	Role                  domain.AgentRole
	Model                 string
	User                  string
	StartedAt             *string
	CompletedAt           *string
	ActiveDurationSeconds int64
	Turns                 int64
	Tokens                domain.InvocationTokens
	CostUSD               float64
	Compactions           int64
	UserIterations        int64
}

func sessionFactsWithInvs(invs []testInvFact, tools []domain.ToolFact, events []domain.EventFact) []domain.SessionFact {
	var sessions []domain.SessionFact
	for i, inv := range invs {
		sessID := inv.ID
		if sessID == "" {
			sessID = fmt.Sprintf("fixture-session-%d", i)
		}
		invID := fmt.Sprintf("%s-inv-1", sessID)
		sessUser := inv.User
		if sessUser == "" {
			sessUser = "fixture-user"
		}
		model := inv.Model
		if model == "" {
			model = "fixture-model"
		}
		startedAt := inv.StartedAt
		if startedAt == nil {
			v := "fixture-started"
			startedAt = &v
		}
		completedAt := inv.CompletedAt
		if completedAt == nil {
			v := "fixture-completed"
			completedAt = &v
		}
		sess := domain.SessionFact{
			ID:       sessID,
			ParentID: inv.ParentID,
			Source:   domain.SourceOpenCode,
			User:     sessUser,
			Invocations: []domain.InvocationFact{
				{
					ID:                    invID,
					Role:                  inv.Role,
					Model:                 model,
					StartedAt:             startedAt,
					CompletedAt:           completedAt,
					ActiveDurationSeconds: inv.ActiveDurationSeconds,
					Turns:                 inv.Turns,
					Tokens:                inv.Tokens,
					CostUSD:               inv.CostUSD,
					Compactions:           inv.Compactions,
					UserIterations:        inv.UserIterations,
				},
			},
			Tools:     []domain.SessionToolFact{},
			Events:    []domain.SessionEventFact{},
			Anomalies: []domain.Anomaly{},
		}
		sessions = append(sessions, sess)
	}
	for _, tool := range tools {
		for i := range sessions {
			if sessions[i].ID == tool.SessionID {
				sessions[i].Tools = append(sessions[i].Tools, domain.SessionToolFact{
					Name:   tool.Name,
					Calls:  tool.Calls,
					Errors: tool.Errors,
				})
			}
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
	sessions, err := domain.SessionsFromFacts(ledger.InvocationFacts(), ledger.ToolFacts(), events)
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

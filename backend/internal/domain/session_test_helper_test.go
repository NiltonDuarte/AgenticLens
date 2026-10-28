package domain_test

import (
	"agenticlens/internal/domain"
	"fmt"
)

func sessionFacts(invocations []domain.InvocationFact, tools []domain.ToolFact, events []domain.EventFact) []domain.SessionFact {
	for i := range invocations {
		if invocations[i].ID == "" {
			invocations[i].ID = fmt.Sprintf("fixture-session-%d", i)
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

type testSessionOpt func(*domain.SessionFact, *domain.InvocationFact)

func withTokens(tokens domain.InvocationTokens) testSessionOpt {
	return func(s *domain.SessionFact, inv *domain.InvocationFact) {
		inv.Tokens = tokens
	}
}

func withCost(cost float64) testSessionOpt {
	return func(s *domain.SessionFact, inv *domain.InvocationFact) {
		inv.CostUSD = cost
	}
}

func withTurns(turns int64) testSessionOpt {
	return func(s *domain.SessionFact, inv *domain.InvocationFact) {
		inv.Turns = turns
	}
}

func withModel(model string) testSessionOpt {
	return func(s *domain.SessionFact, inv *domain.InvocationFact) {
		inv.Model = model
	}
}

func withActiveDuration(secs int64) testSessionOpt {
	return func(s *domain.SessionFact, inv *domain.InvocationFact) {
		inv.ActiveDurationSeconds = secs
	}
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
	return sessionFactsFromSessions(sessions, tools, events)
}

func testSession(id string, parentID *string, role domain.AgentRole, opts ...testSessionOpt) domain.SessionFact {
	sess := domain.SessionFact{
		ID:       id,
		ParentID: parentID,
		Source:   domain.SourceOpenCode,
		User:     "fixture-user",
		Invocations: []domain.InvocationFact{
			{
				ID:    id + "-inv-1",
				Role:  role,
				Model: "fixture-model",
			},
		},
		Tools:     []domain.SessionToolFact{},
		Events:    []domain.SessionEventFact{},
		Anomalies: []domain.Anomaly{},
	}
	for _, opt := range opts {
		opt(&sess, &sess.Invocations[0])
	}
	return sess
}

func sessionFactsFromSessions(sessions []domain.SessionFact, tools []domain.ToolFact, events []domain.EventFact) []domain.SessionFact {
	sessMap := make(map[string]*domain.SessionFact, len(sessions))
	for i := range sessions {
		if sessions[i].User == "" {
			sessions[i].User = "fixture-user"
		}
		if sessions[i].Anomalies == nil {
			sessions[i].Anomalies = []domain.Anomaly{}
		}
		if sessions[i].Tools == nil {
			sessions[i].Tools = []domain.SessionToolFact{}
		}
		if sessions[i].Events == nil {
			sessions[i].Events = []domain.SessionEventFact{}
		}
		sessMap[sessions[i].ID] = &sessions[i]
	}
	for _, tool := range tools {
		if sess, ok := sessMap[tool.SessionID]; ok {
			sess.Tools = append(sess.Tools, domain.SessionToolFact{
				Name:   tool.Name,
				Calls:  tool.Calls,
				Errors: tool.Errors,
			})
		}
	}
	for _, ev := range events {
		if sess, ok := sessMap[ev.SessionID]; ok {
			sess.Events = append(sess.Events, domain.SessionEventFact{
				EventID:       ev.EventID,
				Timestamp:     ev.Timestamp,
				MessageID:     ev.MessageID,
				Turn:          ev.Turn,
				Role:          ev.Role,
				EventType:     ev.EventType,
				ToolName:      ev.ToolName,
				Status:        ev.Status,
				ErrorMessage:  ev.ErrorMessage,
				DurationMs:    ev.DurationMs,
				ContextTokens: ev.ContextTokens,
			})
		}
	}
	return sessions
}

func setEventFacts(ledger *domain.AgenticFactLedger, events []domain.EventFact) {
	if events == nil {
		for i := range ledger.Sessions {
			ledger.Sessions[i].Events = nil
		}
		return
	}
	sessMap := make(map[string]*domain.SessionFact, len(ledger.Sessions))
	for i := range ledger.Sessions {
		ledger.Sessions[i].Events = []domain.SessionEventFact{}
		sessMap[ledger.Sessions[i].ID] = &ledger.Sessions[i]
	}
	for i := range events {
		ev := events[i]
		if ev.SessionID == "" {
			ev.SessionID = "fixture-session"
		}
		if ev.Timestamp == "" {
			ev.Timestamp = fmt.Sprintf("fixture-%06d", i)
		}
		if sess, ok := sessMap[ev.SessionID]; ok {
			sess.Events = append(sess.Events, domain.SessionEventFact{
				EventID:       ev.EventID,
				Timestamp:     ev.Timestamp,
				MessageID:     ev.MessageID,
				Turn:          ev.Turn,
				Role:          ev.Role,
				EventType:     ev.EventType,
				ToolName:      ev.ToolName,
				Status:        ev.Status,
				ErrorMessage:  ev.ErrorMessage,
				DurationMs:    ev.DurationMs,
				ContextTokens: ev.ContextTokens,
			})
		}
	}
}

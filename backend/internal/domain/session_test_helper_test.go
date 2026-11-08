package domain_test

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
	return fixtureSourceSummaries(sessions, events)
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
	SessionID             string
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
		sessID := inv.SessionID
		if sessID == "" {
			sessID = fmt.Sprintf("fixture-session-%d", i)
		}
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
				Role:  role,
				Model: "fixture-model",
			},
		},
		Tools:     []domain.SessionToolFact{},
		Anomalies: []domain.Anomaly{},
	}
	for _, opt := range opts {
		opt(&sess, &sess.Invocations[0])
	}
	return sess
}

func sessionFactsFromSessions(sessions []domain.SessionFact, tools []domain.ToolFact, events []domain.EventFact) []domain.SessionFact {
	for i := range sessions {
		if sessions[i].User == "" {
			sessions[i].User = "fixture-user"
		}
		if sessions[i].Anomalies == nil {
			sessions[i].Anomalies = []domain.Anomaly{}
		}
		for _, tool := range tools {
			if tool.SessionID == sessions[i].ID {
				sessions[i].Tools = append(sessions[i].Tools, domain.SessionToolFact{Name: tool.Name, Calls: tool.Calls, Errors: tool.Errors})
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
	sessions, err := domain.SessionsFromFacts(ledger.Sessions, ledger.ToolFacts(), events)
	if err != nil {
		panic(err)
	}
	ledger.Sessions = sessions
}

func fixtureInvocationSessions(invocations []domain.InvocationFact) []domain.SessionFact {
	if len(invocations) == 0 {
		return nil
	}
	return []domain.SessionFact{{ID: "fixture-session", Source: domain.SourceOpenCode, User: "fixture-user", Invocations: invocations}}
}

func fixtureSourceSummaries(sessions []domain.SessionFact, events []domain.EventFact) []domain.SessionFact {
	ids := []string{}
	for _, session := range sessions {
		ids = append(ids, session.ID)
	}
	ledger, err := domain.PersistSourceAnomalySummaries(domain.AgenticFactLedger{Sessions: sessions}, domain.SourceEvents{Events: events, CompleteSessionIDs: ids}, fixtureAnomalyPolicy())
	if err != nil {
		panic(err)
	}
	return ledger.Sessions
}

// Fixture policy is explicit test input, never an application fallback.
func fixtureAnomalyPolicy() domain.AnomalyThresholdsConfig {
	return domain.AnomalyThresholdsConfig{ErrorRateMinCalls: 5, ErrorRateThreshold: .25, ErrorRateCriticalThreshold: .50}
}
func detectFixtureAnomaliesFromEvents(ledger domain.AgenticFactLedger, events []domain.EventFact, policy ...domain.AnomalyThresholdsConfig) []domain.Anomaly {
	if len(policy) == 0 {
		policy = []domain.AnomalyThresholdsConfig{fixtureAnomalyPolicy()}
	}
	return domain.DetectAnomaliesFromEvents(ledger, events, policy...)
}
func computeFixtureFleetAnomalies(ledgers []domain.AgenticFactLedger, policy ...domain.AnomalyThresholdsConfig) domain.FleetAnomalySummary {
	if len(policy) == 0 {
		policy = []domain.AnomalyThresholdsConfig{fixtureAnomalyPolicy()}
	}
	return domain.ComputeFleetAnomalies(ledgers, policy...)
}

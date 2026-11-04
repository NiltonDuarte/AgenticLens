package migration_test

import (
	"agenticlens/internal/domain"
	"agenticlens/internal/domain/migration"
	"encoding/json"
	"fmt"
)

func sessionFacts(invocations []domain.InvocationFact, tools []domain.ToolFact, events []domain.EventFact) []migration.SessionInput {
	for i := range invocations {
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
	sessions, err := migrationSessionFacts(fixtureInvocationSessions(invocations), tools, events)
	if err != nil {
		panic(err)
	}
	for i := range sessions {
		if sessions[i].Anomalies == nil {
			sessions[i].Anomalies = []domain.Anomaly{}
		}
	}
	return sessions
}

func setEventFacts(ledger *migration.LedgerInput, events []domain.EventFact) {
	for i := range events {
		if events[i].SessionID == "" {
			events[i].SessionID = "fixture-session"
		}
		if events[i].Timestamp == "" {
			events[i].Timestamp = fmt.Sprintf("fixture-%06d", i)
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
	sessions, err := migrationSessionFacts(ledger.Sessions, ledger.ToolFacts(), events)
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

func setInvocationFacts(ledger *migration.LedgerInput, invocations []domain.InvocationFact) {
	for i := range invocations {
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
	sessions, err := migrationSessionFacts(fixtureInvocationSessions(invocations), ledger.ToolFacts(), ledger.EventFacts())
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

func setToolFacts(ledger *migration.LedgerInput, tools []domain.ToolFact) {
	for i := range tools {
		if tools[i].SessionID == "" {
			tools[i].SessionID = "fixture-session"
		}
	}
	sessions, err := migrationSessionFacts(ledger.Sessions, tools, ledger.EventFacts())
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

func fixtureInvocationSessions(invocations []domain.InvocationFact) []migration.SessionInput {
	if len(invocations) == 0 {
		return nil
	}
	return []migration.SessionInput{{ID: "fixture-session", Source: domain.SourceOpenCode, User: "fixture-user", Invocations: invocations}}
}

func migrationSessionFacts(owned []migration.SessionInput, tools []domain.ToolFact, events []domain.EventFact) ([]migration.SessionInput, error) {
	data, err := json.Marshal(owned)
	if err != nil {
		return nil, err
	}
	var canonical []domain.SessionFact
	if err = json.Unmarshal(data, &canonical); err != nil {
		return nil, err
	}
	grouped, err := domain.SessionsFromFacts(canonical, tools, events)
	if err != nil {
		return nil, err
	}
	result := []migration.SessionInput{}
	for _, session := range grouped {
		var out migration.SessionInput
		for _, previous := range owned {
			if previous.ID == session.ID {
				out = previous
			}
		}
		out.ID = session.ID
		out.ParentID = session.ParentID
		out.Source = session.Source
		out.User = session.User
		out.Invocations = session.Invocations
		out.Tools = session.Tools
		out.Events = nil
		for _, event := range events {
			if event.SessionID == session.ID {
				data, _ := json.Marshal(event)
				var input migration.EventInput
				if err := json.Unmarshal(data, &input); err != nil {
					return nil, err
				}
				out.Events = append(out.Events, input)
			}
		}
		result = append(result, out)
	}
	return result, nil
}

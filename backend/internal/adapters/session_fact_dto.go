package adapters

import (
	"agenticlens/internal/domain"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type SessionToolFactDTO struct {
	Name   string `json:"name"`
	Calls  int64  `json:"calls"`
	Errors int64  `json:"errors"`
}

// UnmarshalJSON accepts the retired content count only on reads. It is discarded
// so older supported ledgers remain readable without producing the removed metric.
func (dto *SessionToolFactDTO) UnmarshalJSON(data []byte) error {
	type toolAlias SessionToolFactDTO
	var decoded struct {
		toolAlias
		ContentErrors json.RawMessage `json:"content_errors"`
	}
	decoder := newStrictFactDecoder(data)
	if err := decoder.Decode(&decoded); err != nil {
		return fmt.Errorf("decoding session tool fact: %w", err)
	}
	if err := requireFactJSONEOF(decoder, "session tool fact"); err != nil {
		return err
	}
	*dto = SessionToolFactDTO(decoded.toolAlias)
	return nil
}

type SessionFactDTO struct {
	AnomalyAvailability domain.AnomalyAvailability `json:"anomaly_availability,omitempty"`
	Skills              []SkillFactDTO             `json:"skills,omitempty"`
	Delegations         []DelegationFactDTO        `json:"delegations,omitempty"`
	ID                  string                     `json:"id"`
	ParentID            *string                    `json:"parent_id,omitempty"`
	Source              string                     `json:"source,omitempty"`
	User                string                     `json:"user,omitempty"`
	Invocations         []InvocationFactDTO        `json:"invocations"`
	Tools               []SessionToolFactDTO       `json:"tools"`
	Anomalies           []AnomalyDTO               `json:"anomalies"`
}

// ErrToolSessionIDRequired identifies a fact ledger that has tool aggregates
// without the session identity required by schema v4.
var ErrToolSessionIDRequired = errors.New("tool fact requires session_id")

func (session SessionFactDTO) ToDomain() domain.SessionFact {
	sessionSource := domain.SourceHarness(strings.TrimSpace(session.Source))
	user := strings.TrimSpace(session.User)
	parentID := session.ParentID
	result := domain.SessionFact{
		AnomalyAvailability: session.AnomalyAvailability,
		Skills:              skillFactDTOsToDomain(session.Skills), Delegations: delegationFactDTOsToDomain(session.Delegations),
		ID:       session.ID,
		ParentID: parentID,
		Source:   sessionSource,
		User:     user,
	}
	if session.Anomalies != nil {
		result.Anomalies = []domain.Anomaly{}
	}
	for _, inv := range session.Invocations {
		result.Invocations = append(result.Invocations, inv.ToDomain())
	}
	for _, tool := range session.Tools {
		name := tool.Name
		if name == "" {
			name = "unknown"
		}
		result.Tools = append(result.Tools, domain.SessionToolFact{Name: name, Calls: tool.Calls, Errors: tool.Errors})
	}
	for _, anomaly := range session.Anomalies {
		detected := anomaly.ToDomain()
		detected.Context.SessionID = nil
		result.Anomalies = append(result.Anomalies, detected)
	}
	return result
}

func SessionFactDTOFromDomain(session domain.SessionFact) SessionFactDTO {
	sessionSource := strings.TrimSpace(string(session.Source))
	result := SessionFactDTO{
		AnomalyAvailability: session.AnomalyAvailability,
		Skills:              skillFactDTOsFromDomain(session.Skills), Delegations: delegationFactDTOsFromDomain(session.Delegations),
		ID:          session.ID,
		ParentID:    session.ParentID,
		Source:      sessionSource,
		User:        session.User,
		Invocations: []InvocationFactDTO{},
		Tools:       []SessionToolFactDTO{},
		Anomalies:   []AnomalyDTO{},
	}
	for _, inv := range session.Invocations {
		result.Invocations = append(result.Invocations, InvocationFactDTOFromDomain(inv))
	}
	for _, tool := range session.Tools {
		result.Tools = append(result.Tools, SessionToolFactDTO{Name: tool.Name, Calls: tool.Calls, Errors: tool.Errors})
	}
	if session.AnomalyAvailability == domain.AnomalyUnavailable {
		result.Anomalies = nil
	}
	for _, anomaly := range session.Anomalies {
		anomaly.Context.SessionID = nil
		result.Anomalies = append(result.Anomalies, AnomalyDTOFromDomain(anomaly))
	}
	return result
}

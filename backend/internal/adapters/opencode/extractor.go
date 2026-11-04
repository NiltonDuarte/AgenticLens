package opencode

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"agenticlens/internal/domain"
)

func sourceEventID(kind, sourceID string) string {
	if strings.TrimSpace(sourceID) == "" {
		return ""
	}
	return "opencode:" + kind + ":" + sourceID
}

type sessionModelAccumulator struct {
	input         int64
	output        int64
	cache         int64
	reasoning     int64
	peak          int64
	cost          float64
	costPresent   int64
	costMissing   int64
	activeMS      int64
	provisionalMS int64
	provisional   bool
	durationLimit string
	firstTS       *int64
	lastTS        *int64
	asstMsgs      int64
}

type toolAccumulator struct {
	calls             int64
	errors            int64
	consecutiveErrors int
}

type toolAccumulatorKey struct {
	sessionID string
	toolName  string
}

func sanitizeSessionID(id string) (string, error) {
	clean := strings.TrimSpace(id)
	if clean == "" {
		return "", errors.New("empty session id")
	}
	for i := 0; i < len(clean); i++ {
		c := clean[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			continue
		}
		return "", fmt.Errorf("invalid character '%c' in session id %q", c, clean)
	}
	return clean, nil
}

func buildQuotedIDList(threadIDs []string) (string, error) {
	if len(threadIDs) == 0 {
		return "", errors.New("empty thread id list")
	}
	var b strings.Builder
	for i, id := range threadIDs {
		clean, err := sanitizeSessionID(id)
		if err != nil {
			return "", err
		}
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteByte('\'')
		b.WriteString(clean)
		b.WriteByte('\'')
	}
	return b.String(), nil
}

func formatEpochMilli(ms int64) string {
	t := time.UnixMilli(ms).UTC()
	return t.Format("2006-01-02T15:04:05.000000+00:00")
}

func normalizeToolInput(raw json.RawMessage) *string {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil
	}
	v := string(raw)
	return &v
}

// ExtractSessions queries and builds normalized domain.SessionFact trees for either supported OpenCode schema family.
func ExtractSessions(db *sql.DB, threadIDs []string, defaultUser string, userBySessionID map[string]string) ([]domain.SessionFact, *string, *string, error) {
	return extractSessions(db, threadIDs, defaultUser, userBySessionID, false)
}

// ExtractSessionsStrict rejects malformed source records instead of silently
// omitting them. Backfill uses this boundary before replacing stored snapshots.
func ExtractSessionsStrict(db *sql.DB, threadIDs []string, defaultUser string, userBySessionID map[string]string) ([]domain.SessionFact, *string, *string, error) {
	return extractSessions(db, threadIDs, defaultUser, userBySessionID, true)
}

func extractSessions(db *sql.DB, threadIDs []string, defaultUser string, userBySessionID map[string]string, strict bool, source ...*domain.SourceEvents) ([]domain.SessionFact, *string, *string, error) {
	return extractSessionsAt(db, threadIDs, defaultUser, userBySessionID, strict, time.Now().UTC(), source...)
}

func extractSessionsAt(db *sql.DB, threadIDs []string, defaultUser string, userBySessionID map[string]string, strict bool, cutoff time.Time, source ...*domain.SourceEvents) ([]domain.SessionFact, *string, *string, error) {
	if db == nil {
		return nil, nil, nil, errors.New("database connection is nil")
	}
	family, err := detectSchemaFamily(db)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(threadIDs) == 0 {
		return []domain.SessionFact{}, nil, nil, nil
	}
	var sessions []domain.SessionFact
	var events []domain.EventFact
	var sink *[]domain.EventFact
	if len(source) > 0 {
		sink = &events
	}
	var started, completed *string
	if family == schemaFamilyV2 {
		sessions, started, completed, err = extractSchemaV2SessionsWithEventsAt(db, threadIDs, defaultUser, userBySessionID, strict, sink, cutoff)
	} else {
		sessions, started, completed, err = extractSchemaV1SessionsWithEventsAt(db, threadIDs, defaultUser, userBySessionID, strict, sink, cutoff)
	}
	if err != nil {
		return nil, nil, nil, err
	}
	if len(source) > 0 {
		fresh := domain.SourceEvents{Events: events}
		for _, owner := range sessions {
			fresh.CompleteSessionIDs = append(fresh.CompleteSessionIDs, owner.ID)
		}
		*source[0] = domain.MergeSourceEvents(*source[0], fresh)
	}
	return sessions, started, completed, nil
}

// ExtractSteeringTurns calculates human steering turns from message records for given root threads.
func ExtractSteeringTurns(db *sql.DB, threadIDs []string) (int64, error) {
	if db == nil {
		return 0, errors.New("database connection is nil")
	}
	family, err := detectSchemaFamily(db)
	if err != nil {
		return 0, err
	}
	if len(threadIDs) == 0 {
		return 0, nil
	}
	if family == schemaFamilyV2 {
		return countSchemaV2SteeringTurns(db, threadIDs)
	}
	return countSchemaV1SteeringTurns(db, threadIDs)
}

func (acc sessionModelAccumulator) costEvidence() domain.CostEvidence {
	amount := math.Round(acc.cost*1e6) / 1e6
	if acc.costPresent == 0 {
		return domain.UnavailableCost()
	}
	if acc.costMissing > 0 {
		return domain.PartialCost(amount)
	}
	return domain.ObservedCost(amount)
}

// ExtractSessionSourceEvents reads one exact owner and returns source-only trace
// rows. Successful empty extraction is distinct from a missing native session.
func ExtractSessionSourceEvents(db *sql.DB, id, user string) ([]domain.EventFact, error) {
	family, err := captureSchemaFamily(context.Background(), db, id)
	if err != nil {
		return nil, err
	}
	events := []domain.EventFact{}
	var sessions []domain.SessionFact
	if family == schemaFamilyV2 {
		sessions, _, _, err = extractSchemaV2SessionsWithEvents(db, []string{id}, user, nil, true, &events, true)
	} else {
		sessions, _, _, err = extractSchemaV1SessionsWithEvents(db, []string{id}, user, nil, true, &events, true)
	}
	if err != nil {
		strictErr := err
		if len(events) == 0 {
			var partialErr error
			if family == schemaFamilyV2 {
				_, _, _, partialErr = extractSchemaV2SessionsWithEvents(db, []string{id}, user, nil, false, &events, true)
			} else {
				_, _, _, partialErr = extractSchemaV1SessionsWithEvents(db, []string{id}, user, nil, false, &events, true)
			}
			if partialErr != nil {
				strictErr = errors.Join(strictErr, partialErr)
			}
		}
		return events, strictErr
	}

	if len(sessions) != 1 || sessions[0].ID != id {
		return nil, fmt.Errorf("native session %q unavailable", id)
	}
	return events, nil
}

// nativeMessageText decodes only the source message's own content envelope.
func nativeMessageText(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &parts) != nil {
		return ""
	}
	result := []string{}
	for _, part := range parts {
		if part.Type == "text" || part.Type == "input_text" || part.Type == "output_text" {
			result = append(result, part.Text)
		}
	}
	return strings.Join(result, "\n")
}

// ExtractSessionsStrictWithSource returns canonical owners and a separate
// request-local event input for final aggregation across retained owners.
func ExtractSessionsStrictWithSource(db *sql.DB, ids []string, user string, users map[string]string, source *domain.SourceEvents) ([]domain.SessionFact, *string, *string, error) {
	return extractSessions(db, ids, user, users, true, source)
}

// ExtractSessionsStrictWithSourceAt uses one explicit duration snapshot cutoff.
func ExtractSessionsStrictWithSourceAt(db *sql.DB, ids []string, user string, users map[string]string, source *domain.SourceEvents, cutoff time.Time) ([]domain.SessionFact, *string, *string, error) {
	return extractSessionsAt(db, ids, user, users, true, cutoff, source)
}

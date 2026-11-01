package opencode

import (
	"bytes"
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
	input       int64
	output      int64
	cache       int64
	reasoning   int64
	peak        int64
	cost        float64
	costPresent int64
	costMissing int64
	activeMS    int64
	firstTS     *int64
	lastTS      *int64
	asstMsgs    int64
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
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) == nil {
		if path, ok := fields["path"]; ok {
			canonical, _ := json.Marshal(struct {
				Path json.RawMessage `json:"path"`
			}{path})
			v := string(canonical)
			return &v
		}
		if path, ok := fields["filePath"]; ok {
			canonical, _ := json.Marshal(struct {
				Path json.RawMessage `json:"path"`
			}{path})
			v := string(canonical)
			return &v
		}
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

func extractSessions(db *sql.DB, threadIDs []string, defaultUser string, userBySessionID map[string]string, strict bool) ([]domain.SessionFact, *string, *string, error) {
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
	if family == schemaFamilyV2 {
		return extractSchemaV2Sessions(db, threadIDs, defaultUser, userBySessionID)
	}
	return extractSchemaV1Sessions(db, threadIDs, defaultUser, userBySessionID, strict)
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

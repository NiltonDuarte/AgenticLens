package opencode

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"agenticlens/internal/domain"
)

type rawTimeInfo struct {
	Created   *float64 `json:"created"`
	Completed *float64 `json:"completed"`
}

type rawCacheInfo struct {
	Read *int64 `json:"read"`
}

type rawTokensInfo struct {
	Input     *int64        `json:"input"`
	Output    *int64        `json:"output"`
	Cache     *rawCacheInfo `json:"cache"`
	Reasoning *int64        `json:"reasoning"`
}

type rawMessageData struct {
	ModelID *string        `json:"modelID"`
	Role    *string        `json:"role"`
	Time    *rawTimeInfo   `json:"time"`
	Tokens  *rawTokensInfo `json:"tokens"`
	Cost    *float64       `json:"cost"`
}

type rawPartState struct {
	Status *string `json:"status"`
	Output *string `json:"output"`
}

type rawPartData struct {
	Type  *string       `json:"type"`
	Tool  *string       `json:"tool"`
	State *rawPartState `json:"state"`
}

type rawModelJSON struct {
	ID *string `json:"id"`
}

type sessionModelAccumulator struct {
	input     int64
	output    int64
	cache     int64
	reasoning int64
	peak      int64
	cost      float64
	activeMS  int64
	firstTS   *int64
	lastTS    *int64
	asstMsgs  int64
}

type toolAccumulator struct {
	calls  int64
	errors int64
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

// ExtractFacts queries invocation and tool facts from OpenCode SQLite database for given root threads.
func ExtractFacts(
	db *sql.DB,
	threadIDs []string,
	defaultUser string,
	userBySessionID map[string]string,
) ([]domain.InvocationFact, []domain.ToolFact, *string, *string, error) {
	if db == nil {
		return nil, nil, nil, nil, errors.New("database connection is nil")
	}
	if len(threadIDs) == 0 {
		return []domain.InvocationFact{}, []domain.ToolFact{}, nil, nil, nil
	}

	quotedList, err := buildQuotedIDList(threadIDs)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("validating thread ids: %w", err)
	}

	// 1. Earliest start time and latest completed time across sessions
	qTimes := fmt.Sprintf(`
		SELECT MIN(time_created), MAX(time_updated)
		FROM session
		WHERE id IN (%s) OR parent_id IN (%s)
	`, quotedList, quotedList)

	var minCreated, maxUpdated sql.NullInt64
	err = db.QueryRow(qTimes).Scan(&minCreated, &maxUpdated)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, nil, nil, fmt.Errorf("querying session time boundaries: %w", err)
	}

	var startedAtISO, completedAtISO *string
	if minCreated.Valid {
		formatted := formatEpochMilli(minCreated.Int64)
		startedAtISO = &formatted
	}
	if maxUpdated.Valid {
		formatted := formatEpochMilli(maxUpdated.Int64)
		completedAtISO = &formatted
	}

	// 2. Compactions per session
	qComp := fmt.Sprintf(`
		SELECT p.session_id, p.data
		FROM part p
		JOIN session s ON p.session_id = s.id
		WHERE (s.id IN (%s) OR s.parent_id IN (%s))
		  AND p.data LIKE '%%"type":"compaction"%%'
	`, quotedList, quotedList)

	compMap := make(map[string]int64)
	compRows, err := db.Query(qComp)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("querying session compactions: %w", err)
	}
	defer compRows.Close()

	for compRows.Next() {
		var sid string
		var rawData sql.NullString
		if err := compRows.Scan(&sid, &rawData); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("scanning compaction row: %w", err)
		}
		if rawData.Valid && strings.TrimSpace(rawData.String) != "" {
			var rawPD rawPartData
			if err := json.Unmarshal([]byte(rawData.String), &rawPD); err == nil {
				if rawPD.Type != nil && *rawPD.Type == "compaction" {
					compMap[sid]++
				}
			}
		}
	}
	if err := compRows.Err(); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("iterating compaction rows: %w", err)
	}

	// 3. Sessions
	qSess := fmt.Sprintf(`
		SELECT s.id, s.parent_id, s.agent, s.model, s.time_created
		FROM session s
		WHERE s.id IN (%s) OR s.parent_id IN (%s)
		ORDER BY s.time_created ASC
	`, quotedList, quotedList)

	sessRows, err := db.Query(qSess)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("querying sessions: %w", err)
	}
	defer sessRows.Close()

	type sessRecord struct {
		id          string
		parentID    *string
		agent       string
		model       string
		timeCreated int64
	}

	var sessions []sessRecord
	for sessRows.Next() {
		var sid string
		var pid, agent, model sql.NullString
		var timeCreated int64
		if err := sessRows.Scan(&sid, &pid, &agent, &model, &timeCreated); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("scanning session row: %w", err)
		}
		var parentPtr *string
		if pid.Valid && strings.TrimSpace(pid.String) != "" {
			val := strings.TrimSpace(pid.String)
			parentPtr = &val
		}
		sessions = append(sessions, sessRecord{
			id:          sid,
			parentID:    parentPtr,
			agent:       agent.String,
			model:       model.String,
			timeCreated: timeCreated,
		})
	}
	if err := sessRows.Err(); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("iterating session rows: %w", err)
	}

	var invocations []domain.InvocationFact

	for _, s := range sessions {
		role := "orchestrator"
		if s.parentID != nil && s.agent != "build" && s.agent != "orchestrator" {
			if strings.TrimSpace(s.agent) != "" {
				role = strings.TrimSpace(s.agent)
			}
		}

		sessionUser := defaultUser
		if u, ok := userBySessionID[s.id]; ok && strings.TrimSpace(u) != "" {
			sessionUser = strings.TrimSpace(u)
		}

		qMsgs := `
			SELECT m.data
			FROM message m
			WHERE m.session_id = ?
			ORDER BY m.time_created ASC
		`
		msgRows, err := db.Query(qMsgs, s.id)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("querying messages for session %s: %w", s.id, err)
		}

		byModel := make(map[string]*sessionModelAccumulator)
		var modelOrder []string
		var totalSessionMsgs int64

		for msgRows.Next() {
			totalSessionMsgs++
			var mData sql.NullString
			if err := msgRows.Scan(&mData); err != nil {
				msgRows.Close()
				return nil, nil, nil, nil, fmt.Errorf("scanning message row: %w", err)
			}
			if !mData.Valid || strings.TrimSpace(mData.String) == "" {
				continue
			}

			var rawMD rawMessageData
			if err := json.Unmarshal([]byte(mData.String), &rawMD); err != nil {
				continue
			}

			if rawMD.Role == nil || *rawMD.Role != "assistant" {
				continue
			}

			mName := "default"
			if rawMD.ModelID != nil && strings.TrimSpace(*rawMD.ModelID) != "" {
				mName = strings.TrimSpace(*rawMD.ModelID)
			}

			acc, exists := byModel[mName]
			if !exists {
				acc = &sessionModelAccumulator{}
				byModel[mName] = acc
				modelOrder = append(modelOrder, mName)
			}
			acc.asstMsgs++

			if rawMD.Time != nil {
				if rawMD.Time.Created != nil {
					c := int64(*rawMD.Time.Created)
					if acc.firstTS == nil || c < *acc.firstTS {
						acc.firstTS = &c
					}
					if rawMD.Time.Completed != nil {
						comp := int64(*rawMD.Time.Completed)
						if acc.lastTS == nil || comp > *acc.lastTS {
							acc.lastTS = &comp
						}
						if comp >= c {
							acc.activeMS += comp - c
						}
					} else if acc.lastTS == nil || c > *acc.lastTS {
						acc.lastTS = &c
					}
				}
			}

			if rawMD.Tokens != nil {
				var inp, out, cache, reas int64
				if rawMD.Tokens.Input != nil {
					inp = *rawMD.Tokens.Input
				}
				if rawMD.Tokens.Output != nil {
					out = *rawMD.Tokens.Output
				}
				if rawMD.Tokens.Cache != nil && rawMD.Tokens.Cache.Read != nil {
					cache = *rawMD.Tokens.Cache.Read
				}
				if rawMD.Tokens.Reasoning != nil {
					reas = *rawMD.Tokens.Reasoning
				}

				acc.input += inp
				acc.output += out
				acc.cache += cache
				acc.reasoning += reas
				if inp+cache > acc.peak {
					acc.peak = inp + cache
				}
			}

			if rawMD.Cost != nil {
				acc.cost += *rawMD.Cost
			}
		}
		msgRows.Close()

		if len(byModel) == 0 {
			fallbackModel := "default"
			if strings.TrimSpace(s.model) != "" {
				var mJSON rawModelJSON
				if err := json.Unmarshal([]byte(s.model), &mJSON); err == nil && mJSON.ID != nil && strings.TrimSpace(*mJSON.ID) != "" {
					fallbackModel = strings.TrimSpace(*mJSON.ID)
				} else {
					fallbackModel = strings.TrimSpace(s.model)
				}
			}
			tCreated := s.timeCreated
			byModel[fallbackModel] = &sessionModelAccumulator{
				firstTS: &tCreated,
				lastTS:  &tCreated,
			}
			modelOrder = append(modelOrder, fallbackModel)
		}

		sessionCompactions := compMap[s.id]
		for idx, mName := range modelOrder {
			st := byModel[mName]
			fTS := s.timeCreated
			if st.firstTS != nil {
				fTS = *st.firstTS
			}
			lTS := s.timeCreated
			if st.lastTS != nil {
				lTS = *st.lastTS
			}

			stISO := formatEpochMilli(fTS)
			endISO := formatEpochMilli(lTS)

			var turns int64
			if len(byModel) == 1 {
				turns = totalSessionMsgs
			} else {
				turns = st.asstMsgs
			}

			var compactions int64
			if idx == 0 {
				compactions = sessionCompactions
			}

			activeSecs := int64(math.Round(float64(st.activeMS) / 1000.0))

			invocations = append(invocations, domain.InvocationFact{
				ID:                    s.id,
				ParentID:              s.parentID,
				Role:                  role,
				Model:                 mName,
				User:                  sessionUser,
				StartedAt:             &stISO,
				CompletedAt:           &endISO,
				ActiveDurationSeconds: activeSecs,
				Turns:                 turns,
				Tokens: domain.InvocationTokens{
					Input:       st.input,
					Output:      st.output,
					CacheRead:   st.cache,
					Reasoning:   st.reasoning,
					PeakContext: st.peak,
				},
				CostUSD:     math.Round(st.cost*1e6) / 1e6,
				Compactions: compactions,
			})
		}
	}

	// 4. Tool calls
	qTools := fmt.Sprintf(`
		SELECT p.data
		FROM part p
		JOIN session s ON p.session_id = s.id
		WHERE (s.id IN (%s) OR s.parent_id IN (%s))
		  AND p.data LIKE '%%"type":"tool"%%'
	`, quotedList, quotedList)

	toolRows, err := db.Query(qTools)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("querying tool calls: %w", err)
	}
	defer toolRows.Close()

	toolMap := make(map[string]*toolAccumulator)

	for toolRows.Next() {
		var pData sql.NullString
		if err := toolRows.Scan(&pData); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("scanning tool row: %w", err)
		}
		if !pData.Valid || strings.TrimSpace(pData.String) == "" {
			continue
		}

		var rawPD rawPartData
		if err := json.Unmarshal([]byte(pData.String), &rawPD); err != nil {
			continue
		}

		toolName := "unknown"
		if rawPD.Tool != nil && strings.TrimSpace(*rawPD.Tool) != "" {
			toolName = strings.TrimSpace(*rawPD.Tool)
		}

		tAcc, exists := toolMap[toolName]
		if !exists {
			tAcc = &toolAccumulator{}
			toolMap[toolName] = tAcc
		}
		tAcc.calls++

		if rawPD.State != nil {
			stat := ""
			out := ""
			if rawPD.State.Status != nil {
				stat = *rawPD.State.Status
			}
			if rawPD.State.Output != nil {
				out = *rawPD.State.Output
			}

			outPrefix := out
			if len(outPrefix) > 50 {
				outPrefix = outPrefix[:50]
			}

			if stat == "error" || strings.Contains(out, "Traceback") || strings.Contains(outPrefix, "Error:") {
				tAcc.errors++
			}
		}
	}
	if err := toolRows.Err(); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("iterating tool rows: %w", err)
	}

	var tools []domain.ToolFact
	for tName, tStats := range toolMap {
		tools = append(tools, domain.ToolFact{
			Name:   tName,
			Calls:  tStats.calls,
			Errors: tStats.errors,
		})
	}

	sort.Slice(tools, func(i, j int) bool {
		if tools[i].Calls == tools[j].Calls {
			return tools[i].Name < tools[j].Name
		}
		return tools[i].Calls > tools[j].Calls
	})

	var maxCompleted *string
	for _, inv := range invocations {
		if inv.CompletedAt != nil && *inv.CompletedAt != "" {
			if maxCompleted == nil || *inv.CompletedAt > *maxCompleted {
				maxCompleted = inv.CompletedAt
			}
		}
	}
	if maxCompleted != nil {
		completedAtISO = maxCompleted
	}

	return invocations, tools, startedAtISO, completedAtISO, nil
}

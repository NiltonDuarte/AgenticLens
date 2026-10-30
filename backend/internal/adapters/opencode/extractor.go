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

func sourceEventID(kind, sourceID string) string {
	if strings.TrimSpace(sourceID) == "" {
		return ""
	}
	return "opencode:" + kind + ":" + sourceID
}

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
	Status *string         `json:"status"`
	Output *string         `json:"output"`
	Input  json.RawMessage `json:"input"`
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

func tableHasColumn(db *sql.DB, tableName, columnName string) bool {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err == nil {
			if strings.EqualFold(name, columnName) {
				return true
			}
		}
	}
	return false
}

// ExtractSessions queries and builds normalized domain.SessionFact trees directly from OpenCode SQLite.
func ExtractSessions(
	db *sql.DB,
	threadIDs []string,
	defaultUser string,
	userBySessionID map[string]string,
) ([]domain.SessionFact, *string, *string, error) {
	if db == nil {
		return nil, nil, nil, errors.New("database connection is nil")
	}
	family, familyErr := detectSchema(db)
	if familyErr != nil {
		return nil, nil, nil, familyErr
	}
	if len(threadIDs) == 0 {
		return []domain.SessionFact{}, nil, nil, nil
	}
	if family == familyV2 {
		return extractV2Sessions(db, threadIDs, defaultUser, userBySessionID)
	}

	quotedList, err := buildQuotedIDList(threadIDs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("validating thread ids: %w", err)
	}

	msgTimeUpdatedCol := "0 AS time_updated"
	if tableHasColumn(db, "message", "time_updated") {
		msgTimeUpdatedCol = "m.time_updated"
	}

	partTimeUpdatedCol := "0 AS time_updated"
	if tableHasColumn(db, "part", "time_updated") {
		partTimeUpdatedCol = "p.time_updated"
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
		return nil, nil, nil, fmt.Errorf("querying session time boundaries: %w", err)
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

	// 2. Query sessions
	qSess := fmt.Sprintf(`
		SELECT s.id, s.parent_id, s.agent, s.model, s.time_created
		FROM session s
		WHERE s.id IN (%s) OR s.parent_id IN (%s)
		ORDER BY s.time_created ASC
	`, quotedList, quotedList)

	sessRows, err := db.Query(qSess)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("querying sessions: %w", err)
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
			return nil, nil, nil, fmt.Errorf("scanning session row: %w", err)
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
		return nil, nil, nil, fmt.Errorf("iterating session rows: %w", err)
	}

	type msgMeta struct {
		turn          int64
		contextTokens int64
		timestamp     string
	}

	msgLookup := make(map[string]msgMeta)
	var events []domain.EventFact
	sessionRoleMap := make(map[string]domain.AgentRole)
	sessionModels := make(map[string]map[string]*sessionModelAccumulator)
	sessionModelOrders := make(map[string][]string)
	sessionTurns := make(map[string]int64)
	sessionUserIterations := make(map[string]int64)

	// 3. Process messages per session
	for _, s := range sessions {
		cleanAgent := strings.TrimSpace(s.agent)
		var role domain.AgentRole
		if cleanAgent != "" {
			role = domain.AgentRole(cleanAgent)
		} else if s.parentID == nil {
			role = domain.RoleRootAgent
		} else {
			role = domain.RoleUnassignedSubAgent
		}
		sessionRoleMap[s.id] = role

		qMsgs := fmt.Sprintf(`
			SELECT m.id, m.time_created, %s, m.data
			FROM message m
			WHERE m.session_id = ?
			ORDER BY m.time_created ASC
		`, msgTimeUpdatedCol)
		msgRows, err := db.Query(qMsgs, s.id)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("querying messages for session %s: %w", s.id, err)
		}

		byModel := make(map[string]*sessionModelAccumulator)
		var modelOrder []string
		var totalSessionMsgs int64
		var sessionAssistantTurn int64
		var sessionUserTurns int64

		for msgRows.Next() {
			totalSessionMsgs++
			var mid string
			var mCreated, mUpdated int64
			var mData sql.NullString
			if err := msgRows.Scan(&mid, &mCreated, &mUpdated, &mData); err != nil {
				msgRows.Close()
				return nil, nil, nil, fmt.Errorf("scanning message row: %w", err)
			}
			if !mData.Valid || strings.TrimSpace(mData.String) == "" {
				continue
			}

			var rawMD rawMessageData
			if err := json.Unmarshal([]byte(mData.String), &rawMD); err != nil {
				continue
			}

			if rawMD.Role != nil && strings.EqualFold(strings.TrimSpace(*rawMD.Role), "user") {
				sessionUserTurns++
				continue
			}

			if rawMD.Role == nil || *rawMD.Role != "assistant" {
				continue
			}

			sessionAssistantTurn++
			turn := sessionAssistantTurn

			var inp, out, cache, reas int64
			if rawMD.Tokens != nil {
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
			}
			ctxTokens := inp + cache

			msgTSMillis := mCreated
			if rawMD.Time != nil && rawMD.Time.Created != nil && int64(*rawMD.Time.Created) > 0 {
				msgTSMillis = int64(*rawMD.Time.Created)
			}
			msgTS := formatEpochMilli(msgTSMillis)

			var msgDurMs *int64
			if rawMD.Time != nil && rawMD.Time.Created != nil && rawMD.Time.Completed != nil {
				dur := int64(*rawMD.Time.Completed) - int64(*rawMD.Time.Created)
				if dur >= 0 {
					msgDurMs = &dur
				}
			} else if mUpdated >= mCreated && mCreated > 0 {
				dur := mUpdated - mCreated
				msgDurMs = &dur
			}

			if mid != "" {
				msgLookup[mid] = msgMeta{
					turn:          turn,
					contextTokens: ctxTokens,
					timestamp:     msgTS,
				}
			}

			var msgIDPtr *string
			if mid != "" {
				mIDVal := mid
				msgIDPtr = &mIDVal
			}

			events = append(events, domain.EventFact{
				EventID:       sourceEventID("message", mid),
				Timestamp:     msgTS,
				SessionID:     s.id,
				MessageID:     msgIDPtr,
				Turn:          turn,
				Role:          role,
				EventType:     domain.EventTypeTurnMessage,
				ToolName:      nil,
				Status:        "completed",
				ErrorMessage:  nil,
				DurationMs:    msgDurMs,
				ContextTokens: ctxTokens,
			})

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

			acc.input += inp
			acc.output += out
			acc.cache += cache
			acc.reasoning += reas
			if inp+cache > acc.peak {
				acc.peak = inp + cache
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

		sessionModels[s.id] = byModel
		sessionModelOrders[s.id] = modelOrder
		sessionTurns[s.id] = totalSessionMsgs
		sessionUserIterations[s.id] = sessionUserTurns
	}

	// 4. Query parts (tool calls and compactions)
	qParts := fmt.Sprintf(`
		SELECT p.id, p.session_id, p.message_id, p.time_created, %s, p.data
		FROM part p
		JOIN session s ON p.session_id = s.id
		WHERE (s.id IN (%s) OR s.parent_id IN (%s))
		  AND (p.data LIKE '%%"type":"tool"%%' OR p.data LIKE '%%"type":"compaction"%%')
		ORDER BY p.time_created ASC
	`, partTimeUpdatedCol, quotedList, quotedList)

	partRows, err := db.Query(qParts)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("querying parts: %w", err)
	}
	defer partRows.Close()

	compMap := make(map[string]int64)
	toolMap := make(map[toolAccumulatorKey]*toolAccumulator)

	for partRows.Next() {
		var pid, sid, mid sql.NullString
		var tCreated, tUpdated sql.NullInt64
		var pData sql.NullString
		if err := partRows.Scan(&pid, &sid, &mid, &tCreated, &tUpdated, &pData); err != nil {
			return nil, nil, nil, fmt.Errorf("scanning part row: %w", err)
		}
		if !pData.Valid || strings.TrimSpace(pData.String) == "" {
			continue
		}

		var rawPD rawPartData
		if err := json.Unmarshal([]byte(pData.String), &rawPD); err != nil {
			continue
		}

		if rawPD.Type == nil {
			continue
		}

		sID := ""
		if sid.Valid {
			sID = sid.String
		}
		mID := ""
		if mid.Valid {
			mID = mid.String
		}
		partID := ""
		if pid.Valid {
			partID = sourceEventID("part", pid.String)
		}

		role := domain.RoleRootAgent
		if r, ok := sessionRoleMap[sID]; ok {
			role = r
		}

		turn := int64(1)
		var ctxTokens int64
		if meta, ok := msgLookup[mID]; ok {
			turn = meta.turn
			ctxTokens = meta.contextTokens
		}

		ts := ""
		if tCreated.Valid && tCreated.Int64 > 0 {
			ts = formatEpochMilli(tCreated.Int64)
		} else if meta, ok := msgLookup[mID]; ok && meta.timestamp != "" {
			ts = meta.timestamp
		} else {
			ts = formatEpochMilli(0)
		}

		var msgIDPtr *string
		if mID != "" {
			mIDVal := mID
			msgIDPtr = &mIDVal
		}

		if *rawPD.Type == "compaction" {
			compMap[sID]++
			events = append(events, domain.EventFact{
				EventID:       partID,
				Timestamp:     ts,
				SessionID:     sID,
				MessageID:     msgIDPtr,
				Turn:          turn,
				Role:          role,
				EventType:     domain.EventTypeCompaction,
				ToolName:      nil,
				Status:        "completed",
				ErrorMessage:  nil,
				DurationMs:    nil,
				ContextTokens: ctxTokens,
			})
			continue
		}

		if *rawPD.Type == "tool" {
			if sID == "" {
				return nil, nil, nil, errors.New("tool part has no owning session ID")
			}
			toolName := "unknown"
			if rawPD.Tool != nil && strings.TrimSpace(*rawPD.Tool) != "" {
				toolName = strings.TrimSpace(*rawPD.Tool)
			}

			key := toolAccumulatorKey{sessionID: sID, toolName: toolName}
			tAcc, exists := toolMap[key]
			if !exists {
				tAcc = &toolAccumulator{}
				toolMap[key] = tAcc
			}
			tAcc.calls++

			stat := ""
			out := ""
			if rawPD.State != nil {
				if rawPD.State.Status != nil {
					stat = *rawPD.State.Status
				}
				if rawPD.State.Output != nil {
					out = *rawPD.State.Output
				}
			}

			outPrefix := out
			if len(outPrefix) > 50 {
				outPrefix = outPrefix[:50]
			}

			trimmedOut := strings.TrimSpace(out)
			cleanStat := strings.ToLower(strings.TrimSpace(stat))
			cleanTool := strings.ToLower(toolName)

			var isError bool
			if (cleanTool == "read" || cleanTool == "grep" || cleanTool == "glob" || cleanTool == "webfetch") && cleanStat == "completed" {
				isError = false
			} else {
				isError = stat == "error" || strings.Contains(out, "Traceback") || strings.HasPrefix(trimmedOut, "Error:") || strings.Contains(outPrefix, "Error:")
			}

			if isError {
				tAcc.errors++
			}

			eventStatus := "completed"
			if isError {
				eventStatus = "error"
			}

			var durMs *int64
			if tCreated.Valid && tUpdated.Valid && tCreated.Int64 > 0 && tUpdated.Int64 >= tCreated.Int64 {
				dur := tUpdated.Int64 - tCreated.Int64
				durMs = &dur
			}
			var toolInput, toolOutput *string
			if rawPD.State != nil {
				toolInput = normalizeToolInput(rawPD.State.Input)
				if rawPD.State.Output != nil {
					outputValue := *rawPD.State.Output
					toolOutput = &outputValue
				}
			}

			tName := toolName
			events = append(events, domain.EventFact{
				EventID:       partID,
				Timestamp:     ts,
				SessionID:     sID,
				MessageID:     msgIDPtr,
				Turn:          turn,
				Role:          role,
				EventType:     domain.EventTypeToolCall,
				ToolName:      &tName,
				Status:        eventStatus,
				ErrorMessage:  nil,
				DurationMs:    durMs,
				ContextTokens: ctxTokens,
				ToolInput:     toolInput,
				ToolOutput:    toolOutput,
			})
			continue
		}
	}
	if err := partRows.Err(); err != nil {
		return nil, nil, nil, fmt.Errorf("iterating part rows: %w", err)
	}

	// 5. Build tools per session
	toolsBySession := make(map[string][]domain.SessionToolFact)
	for key, tStats := range toolMap {
		toolsBySession[key.sessionID] = append(toolsBySession[key.sessionID], domain.SessionToolFact{
			Name:   key.toolName,
			Calls:  tStats.calls,
			Errors: tStats.errors,
		})
	}
	for sid := range toolsBySession {
		sort.Slice(toolsBySession[sid], func(i, j int) bool {
			if toolsBySession[sid][i].Calls == toolsBySession[sid][j].Calls {
				return toolsBySession[sid][i].Name < toolsBySession[sid][j].Name
			}
			return toolsBySession[sid][i].Calls > toolsBySession[sid][j].Calls
		})
	}

	// 6. Build events per session
	eventsBySession := make(map[string][]domain.SessionEventFact)
	for _, ev := range events {
		eventsBySession[ev.SessionID] = append(eventsBySession[ev.SessionID], domain.SessionEventFact{
			EventID:          ev.EventID,
			Timestamp:        ev.Timestamp,
			MessageID:        ev.MessageID,
			Turn:             ev.Turn,
			Role:             ev.Role,
			EventType:        ev.EventType,
			ToolName:         ev.ToolName,
			Status:           ev.Status,
			ErrorMessage:     ev.ErrorMessage,
			DurationMs:       ev.DurationMs,
			ContextTokens:    ev.ContextTokens,
			PromptContent:    ev.PromptContent,
			ReasoningContent: ev.ReasoningContent,
			ToolInput:        ev.ToolInput,
			ToolOutput:       ev.ToolOutput,
		})
	}
	for sid := range eventsBySession {
		sort.SliceStable(eventsBySession[sid], func(i, j int) bool {
			if eventsBySession[sid][i].Timestamp != eventsBySession[sid][j].Timestamp {
				return eventsBySession[sid][i].Timestamp < eventsBySession[sid][j].Timestamp
			}
			return eventsBySession[sid][i].Turn < eventsBySession[sid][j].Turn
		})
	}

	// 7. Build sessions and invocations
	var sessionFacts []domain.SessionFact
	for _, s := range sessions {
		sessionUser := defaultUser
		if u, ok := userBySessionID[s.id]; ok && strings.TrimSpace(u) != "" {
			sessionUser = strings.TrimSpace(u)
		}
		role := sessionRoleMap[s.id]
		byModel := sessionModels[s.id]
		modelOrder := sessionModelOrders[s.id]
		totalSessionMsgs := sessionTurns[s.id]
		sessionCompactions := compMap[s.id]

		var invocations []domain.InvocationFact
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
			var userIterations int64
			if idx == 0 {
				compactions = sessionCompactions
				if s.parentID == nil {
					userIterations = sessionUserIterations[s.id]
				}
			}

			activeSecs := int64(math.Round(float64(st.activeMS) / 1000.0))

			invocations = append(invocations, domain.InvocationFact{
				Role:                  role,
				Model:                 mName,
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
				CostUSD:        math.Round(st.cost*1e6) / 1e6,
				Compactions:    compactions,
				UserIterations: userIterations,
			})
		}

		sTools := toolsBySession[s.id]
		if sTools == nil {
			sTools = []domain.SessionToolFact{}
		}
		sEvents := eventsBySession[s.id]
		if sEvents == nil {
			sEvents = []domain.SessionEventFact{}
		}

		sessionFacts = append(sessionFacts, domain.SessionFact{
			ID:          s.id,
			ParentID:    s.parentID,
			Source:      domain.SourceOpenCode,
			User:        sessionUser,
			Invocations: invocations,
			Tools:       sTools,
			Events:      sEvents,
			Anomalies:   []domain.Anomaly{},
		})
	}

	// 8. Determine overall completed timestamp
	var maxCompleted *string
	for _, sf := range sessionFacts {
		for _, inv := range sf.Invocations {
			if inv.CompletedAt != nil && *inv.CompletedAt != "" {
				if maxCompleted == nil || *inv.CompletedAt > *maxCompleted {
					maxCompleted = inv.CompletedAt
				}
			}
		}
	}
	if maxCompleted != nil {
		completedAtISO = maxCompleted
	}

	return sessionFacts, startedAtISO, completedAtISO, nil
}

// ExtractFacts queries invocation, tool, and chronological event facts from OpenCode SQLite database for given root threads.
func ExtractFacts(
	db *sql.DB,
	threadIDs []string,
	defaultUser string,
	userBySessionID map[string]string,
) ([]domain.InvocationFact, []domain.ToolFact, []domain.EventFact, *string, *string, error) {
	sessions, startedAtISO, completedAtISO, err := ExtractSessions(db, threadIDs, defaultUser, userBySessionID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	ledger := domain.AgenticFactLedger{Sessions: sessions}
	tools := ledger.ToolFacts()
	sort.Slice(tools, func(i, j int) bool {
		if tools[i].SessionID != tools[j].SessionID {
			return tools[i].SessionID < tools[j].SessionID
		}
		if tools[i].Calls == tools[j].Calls {
			return tools[i].Name < tools[j].Name
		}
		return tools[i].Calls > tools[j].Calls
	})
	events := ledger.EventFacts()
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].Timestamp != events[j].Timestamp {
			return events[i].Timestamp < events[j].Timestamp
		}
		if events[i].Turn != events[j].Turn {
			return events[i].Turn < events[j].Turn
		}
		return events[i].SessionID < events[j].SessionID
	})
	return ledger.InvocationFacts(), tools, events, startedAtISO, completedAtISO, nil
}

// ExtractSteeringTurns calculates human steering turns from message records for given root threads.
func ExtractSteeringTurns(db *sql.DB, threadIDs []string) (int64, error) {
	if db == nil {
		return 0, errors.New("database connection is nil")
	}
	family, familyErr := detectSchema(db)
	if familyErr != nil {
		return 0, familyErr
	}
	if len(threadIDs) == 0 {
		return 0, nil
	}
	if family == familyV2 {
		quoted, err := buildQuotedIDList(threadIDs)
		if err != nil {
			return 0, fmt.Errorf("validating thread ids: %w", err)
		}
		var count int64
		err = db.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM session_message WHERE type='user' AND session_id IN (SELECT id FROM session_v2 WHERE id IN (%s) OR parent_id IN (%s))`, quoted, quoted)).Scan(&count)
		if err != nil {
			return 0, fmt.Errorf("counting v2 user messages: %w", err)
		}
		return count, nil
	}

	quotedList, err := buildQuotedIDList(threadIDs)
	if err != nil {
		return 0, fmt.Errorf("validating thread ids: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT m.data
		FROM message m
		WHERE m.session_id IN (%s)
		ORDER BY m.time_created ASC
	`, quotedList)

	rows, err := db.Query(query)
	if err != nil {
		return 0, fmt.Errorf("querying steering turns: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		var mData sql.NullString
		if err := rows.Scan(&mData); err != nil {
			return 0, fmt.Errorf("scanning message data: %w", err)
		}
		if !mData.Valid || strings.TrimSpace(mData.String) == "" {
			continue
		}

		var rawMD rawMessageData
		if err := json.Unmarshal([]byte(mData.String), &rawMD); err != nil {
			continue
		}

		if rawMD.Role != nil && strings.EqualFold(strings.TrimSpace(*rawMD.Role), "user") {
			count++
		}
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterating message rows: %w", err)
	}

	return count, nil
}

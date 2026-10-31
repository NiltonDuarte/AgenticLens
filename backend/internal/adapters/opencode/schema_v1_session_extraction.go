package opencode

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"agenticlens/internal/domain"
)

type schemaV1TimeInfo struct {
	Created   *float64 `json:"created"`
	Completed *float64 `json:"completed"`
}

type schemaV1CacheInfo struct {
	Read *int64 `json:"read"`
}

type schemaV1TokensInfo struct {
	Input     *int64             `json:"input"`
	Output    *int64             `json:"output"`
	Cache     *schemaV1CacheInfo `json:"cache"`
	Reasoning *int64             `json:"reasoning"`
}

type schemaV1ModelJSON struct {
	ID *string `json:"id"`
}

type schemaV1MessageData struct {
	ModelID *string             `json:"modelID"`
	Role    *string             `json:"role"`
	Time    *schemaV1TimeInfo   `json:"time"`
	Tokens  *schemaV1TokensInfo `json:"tokens"`
	Cost    *float64            `json:"cost"`
}

type schemaV1PartState struct {
	Status *string         `json:"status"`
	Output *string         `json:"output"`
	Input  json.RawMessage `json:"input"`
}

type schemaV1PartData struct {
	Type  *string            `json:"type"`
	Tool  *string            `json:"tool"`
	State *schemaV1PartState `json:"state"`
}

func extractSchemaV1Sessions(
	db *sql.DB,
	threadIDs []string,
	defaultUser string,
	userBySessionID map[string]string,
) ([]domain.SessionFact, *string, *string, error) {

	quotedList, err := buildQuotedIDList(threadIDs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("OpenCode schema family v1: validating thread ids: %w", err)
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
		return nil, nil, nil, fmt.Errorf("OpenCode schema family v1: querying session time boundaries: %w", err)
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
		return nil, nil, nil, fmt.Errorf("OpenCode schema family v1: querying sessions: %w", err)
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
			return nil, nil, nil, fmt.Errorf("OpenCode schema family v1: scanning session row: %w", err)
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
		return nil, nil, nil, fmt.Errorf("OpenCode schema family v1: iterating session rows: %w", err)
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

		qMsgs := `
			SELECT m.id, m.time_created, m.time_updated, m.data
			FROM message m
			WHERE m.session_id = ?
			ORDER BY m.time_created ASC
		`
		msgRows, err := db.Query(qMsgs, s.id)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("OpenCode schema family v1: querying messages for session %s: %w", s.id, err)
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
				return nil, nil, nil, fmt.Errorf("OpenCode schema family v1: scanning message row: %w", err)
			}
			if !mData.Valid || strings.TrimSpace(mData.String) == "" {
				continue
			}

			var rawMD schemaV1MessageData
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
				var mJSON schemaV1ModelJSON
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
		SELECT p.id, p.session_id, p.message_id, p.time_created, p.time_updated, p.data
		FROM part p
		JOIN session s ON p.session_id = s.id
		WHERE (s.id IN (%s) OR s.parent_id IN (%s))
		  AND (p.data LIKE '%%"type":"tool"%%' OR p.data LIKE '%%"type":"compaction"%%')
		ORDER BY p.time_created ASC
	`, quotedList, quotedList)

	partRows, err := db.Query(qParts)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("OpenCode schema family v1: querying parts: %w", err)
	}
	defer partRows.Close()

	compMap := make(map[string]int64)
	toolMap := make(map[toolAccumulatorKey]*toolAccumulator)

	for partRows.Next() {
		var pid, sid, mid sql.NullString
		var tCreated, tUpdated sql.NullInt64
		var pData sql.NullString
		if err := partRows.Scan(&pid, &sid, &mid, &tCreated, &tUpdated, &pData); err != nil {
			return nil, nil, nil, fmt.Errorf("OpenCode schema family v1: scanning part row: %w", err)
		}
		if !pData.Valid || strings.TrimSpace(pData.String) == "" {
			continue
		}

		var rawPD schemaV1PartData
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
				return nil, nil, nil, errors.New("OpenCode schema family v1: tool part has no owning session ID")
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

			classification := domain.ClassifyToolExecution(toolName, stat, out, tAcc.consecutiveErrors)
			isError := classification.IsError
			content := domain.ClassifyCompletedToolContent(toolName, stat, out, isError)
			if content.HasContentError {
				tAcc.contentErrors++
			}

			if isError {
				tAcc.errors++
				tAcc.consecutiveErrors++
			} else {
				tAcc.consecutiveErrors = 0
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
				EventID:      partID,
				Timestamp:    ts,
				SessionID:    sID,
				MessageID:    msgIDPtr,
				Turn:         turn,
				Role:         role,
				EventType:    domain.EventTypeToolCall,
				ToolName:     &tName,
				Status:       eventStatus,
				ErrorMessage: classification.ErrorMessage, HasContentError: content.HasContentError, ContentSeverity: string(content.Severity), ContentErrorSnippet: content.ErrorSnippet,
				DurationMs:    durMs,
				ContextTokens: ctxTokens,
				ToolInput:     toolInput,
				ToolOutput:    toolOutput,
			})
			continue
		}
	}
	if err := partRows.Err(); err != nil {
		return nil, nil, nil, fmt.Errorf("OpenCode schema family v1: iterating part rows: %w", err)
	}

	// 5. Build tools per session
	toolsBySession := make(map[string][]domain.SessionToolFact)
	for key, tStats := range toolMap {
		toolsBySession[key.sessionID] = append(toolsBySession[key.sessionID], domain.SessionToolFact{
			Name:   key.toolName,
			Calls:  tStats.calls,
			Errors: tStats.errors, ContentErrors: tStats.contentErrors,
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
			EventID:   ev.EventID,
			Timestamp: ev.Timestamp,
			MessageID: ev.MessageID,
			Turn:      ev.Turn,
			Role:      ev.Role,
			EventType: ev.EventType,
			ToolName:  ev.ToolName,
			Status:    ev.Status, HasContentError: ev.HasContentError, ContentSeverity: ev.ContentSeverity, ContentErrorSnippet: ev.ContentErrorSnippet,
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

// countSchemaV1SteeringTurns calculates human steering turns from message records for given root threads.
func countSchemaV1SteeringTurns(db *sql.DB, threadIDs []string) (int64, error) {

	quotedList, err := buildQuotedIDList(threadIDs)
	if err != nil {
		return 0, fmt.Errorf("OpenCode schema family v1: validating thread ids: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT m.data
		FROM message m
		WHERE m.session_id IN (%s)
		ORDER BY m.time_created ASC
	`, quotedList)

	rows, err := db.Query(query)
	if err != nil {
		return 0, fmt.Errorf("OpenCode schema family v1: querying steering turns: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		var mData sql.NullString
		if err := rows.Scan(&mData); err != nil {
			return 0, fmt.Errorf("OpenCode schema family v1: scanning message data: %w", err)
		}
		if !mData.Valid || strings.TrimSpace(mData.String) == "" {
			continue
		}

		var rawMD schemaV1MessageData
		if err := json.Unmarshal([]byte(mData.String), &rawMD); err != nil {
			continue
		}

		if rawMD.Role != nil && strings.EqualFold(strings.TrimSpace(*rawMD.Role), "user") {
			count++
		}
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("OpenCode schema family v1: iterating message rows: %w", err)
	}

	return count, nil
}

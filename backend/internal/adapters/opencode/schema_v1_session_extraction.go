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
	Content          json.RawMessage     `json:"content"`
	Reasoning        *string             `json:"reasoning"`
	ReasoningContent *string             `json:"reasoning_content"`
	ModelID          *string             `json:"modelID"`
	Role             *string             `json:"role"`
	Agent            string              `json:"agent"`
	Time             *schemaV1TimeInfo   `json:"time"`
	Tokens           *schemaV1TokensInfo `json:"tokens"`
	Cost             *float64            `json:"cost"`
}

type schemaV1PartState struct {
	Status *string         `json:"status"`
	Output *string         `json:"output"`
	Input  json.RawMessage `json:"input"`
}

type schemaV1PartData struct {
	Text  *string            `json:"text"`
	Type  *string            `json:"type"`
	Tool  *string            `json:"tool"`
	State *schemaV1PartState `json:"state"`
}

func extractSchemaV1SessionsWithEvents(db *sql.DB, ids []string, user string, users map[string]string, strict bool, sink *[]domain.EventFact, exact ...bool) ([]domain.SessionFact, *string, *string, error) {
	return extractSchemaV1SessionsWithPolicyAt(db, ids, user, users, strict, sink, nil, time.Now().UTC(), exact...)
}

func extractSchemaV1SessionsWithEventsAt(db *sql.DB, threadIDs []string, defaultUser string, userBySessionID map[string]string, strict bool, eventSink *[]domain.EventFact, cutoff time.Time, exact ...bool) ([]domain.SessionFact, *string, *string, error) {
	return extractSchemaV1SessionsWithPolicyAt(db, threadIDs, defaultUser, userBySessionID, strict, eventSink, nil, cutoff, exact...)
}

func extractSchemaV1SessionsWithPolicy(db *sql.DB, threadIDs []string, defaultUser string, userBySessionID map[string]string, strict bool, eventSink *[]domain.EventFact, policy *domain.GenerationPolicy, exact ...bool) ([]domain.SessionFact, *string, *string, error) {
	return extractSchemaV1SessionsWithPolicyAt(db, threadIDs, defaultUser, userBySessionID, strict, eventSink, policy, time.Now().UTC(), exact...)
}

func extractSchemaV1SessionsWithPolicyAt(db *sql.DB, threadIDs []string, defaultUser string, userBySessionID map[string]string, strict bool, eventSink *[]domain.EventFact, policy *domain.GenerationPolicy, cutoff time.Time, exact ...bool) ([]domain.SessionFact, *string, *string, error) {

	incompleteOwners := make(map[string]bool)
	quotedList, err := buildQuotedIDList(threadIDs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("OpenCode schema family v1: validating thread ids: %w", err)
	}

	parentList := quotedList
	if len(exact) > 0 && exact[0] {
		parentList = "NULL"
	}

	// 1. Earliest start time and latest completed time across sessions
	qTimes := fmt.Sprintf(`
		SELECT MIN(time_created), MAX(time_updated)
		FROM session
		WHERE id IN (%s) OR parent_id IN (%s)
	`, quotedList, parentList)

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
	`, quotedList, parentList)

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
	messageContent := make(map[string]string)
	promptSource := make(map[string]string)
	var events []domain.EventFact
	defer func() {
		if eventSink != nil {
			for i := range events {
				event := &events[i]
				if event.EventType != domain.EventTypeTurnMessage || event.MessageID == nil {
					continue
				}
				mid := *event.MessageID
				source := mid
				if previous, ok := promptSource[mid]; ok {
					source = previous
				}
				if content := messageContent[source]; content != "" {
					value := content
					event.PromptContent = &value
				}
			}
			*eventSink = append(*eventSink, events...)
		}
	}()
	sessionRoleMap := make(map[string]domain.AgentRole)
	sessionModels := make(map[string]map[string]*sessionModelAccumulator)
	sessionModelOrders := make(map[string][]string)
	sessionTurns := make(map[string]int64)
	sessionUserIterations := make(map[string]int64)

	// Native running tools or open streaming-part timing establish an unfinished
	// execution lifecycle. Native activity start (or record creation for older
	// tool records without a native start) bounds that evidence; a generic
	// progress update after cutoff does not erase an earlier open interval.
	// Only a later message at or before cutoff can supersede that evidence.
	runningMessages := map[string]bool{}
	runningRows, runErr := db.Query(fmt.Sprintf(`SELECT DISTINCT p.message_id FROM part p
      JOIN message m ON m.id=p.message_id AND m.session_id=p.session_id
      WHERE (p.session_id IN (%s) OR p.session_id IN (SELECT id FROM session WHERE parent_id IN (%s)))
      AND json_valid(p.data)
      AND ((json_extract(p.data,'$.type')='tool' AND json_extract(p.data,'$.state.status')='running'
            AND coalesce(json_extract(p.data,'$.state.time.start'),p.time_created)>0
            AND coalesce(json_extract(p.data,'$.state.time.start'),p.time_created)<=?)
        OR (json_extract(p.data,'$.type') IN ('text','reasoning') AND json_type(p.data,'$.time.start')='integer' AND json_extract(p.data,'$.time.start')<=? AND json_type(p.data,'$.time.end') IS NULL))
      AND NOT EXISTS (SELECT 1 FROM message newer WHERE newer.session_id=m.session_id AND newer.time_created>m.time_created AND newer.time_created<=?)`, quotedList, parentList), cutoff.UnixMilli(), cutoff.UnixMilli(), cutoff.UnixMilli())
	if runErr != nil {
		return nil, nil, nil, fmt.Errorf("reading v1 activity evidence: %w", runErr)
	}
	for runningRows.Next() {
		var id string
		if err := runningRows.Scan(&id); err != nil {
			runningRows.Close()
			return nil, nil, nil, err
		}
		runningMessages[id] = true
	}
	if err := runningRows.Err(); err != nil {
		runningRows.Close()
		return nil, nil, nil, err
	}
	if err := runningRows.Close(); err != nil {
		return nil, nil, nil, err
	}

	// 3. Process messages per session
	for _, s := range sessions {
		role := domain.AgentRole(strings.TrimSpace(s.agent))
		var persona domain.PersonaEvidence
		eventStart := len(events)

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
		lastUserMessage := ""

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
				if strict {
					return nil, nil, nil, errors.Join(fmt.Errorf("OpenCode schema family v1: empty message %s in session %s", mid, s.id), msgRows.Close())
				}
				incompleteOwners[s.id] = true
				continue
			}

			var rawMD schemaV1MessageData
			if err := json.Unmarshal([]byte(mData.String), &rawMD); err != nil {
				if strict {
					return nil, nil, nil, errors.Join(fmt.Errorf("OpenCode schema family v1: decoding message %s in session %s: %w", mid, s.id, err), msgRows.Close())
				}
				incompleteOwners[s.id] = true
				continue
			}

			if strict && (rawMD.Role == nil || (*rawMD.Role != "user" && *rawMD.Role != "assistant")) {
				return nil, nil, nil, errors.Join(fmt.Errorf("OpenCode schema family v1: unsupported role in message %s in session %s", mid, s.id), msgRows.Close())
			}
			if rawMD.Role != nil && strings.EqualFold(strings.TrimSpace(*rawMD.Role), "user") {
				sessionUserTurns++
				lastUserMessage = mid
				messageContent[mid] = nativeMessageText(rawMD.Content)
				continue
			}

			if rawMD.Role == nil || *rawMD.Role != "assistant" {
				incompleteOwners[s.id] = true
				continue
			}

			messageContent[mid] = nativeMessageText(rawMD.Content)
			if lastUserMessage != "" {
				promptSource[mid] = lastUserMessage
				lastUserMessage = ""
			}
			reasoning := rawMD.ReasoningContent
			if reasoning == nil {
				reasoning = rawMD.Reasoning
			}
			persona = persona.WithCandidate(rawMD.Agent)
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
				EventID:          sourceEventID("message", mid),
				Timestamp:        msgTS,
				SessionID:        s.id,
				MessageID:        msgIDPtr,
				Turn:             turn,
				Role:             role,
				EventType:        domain.EventTypeTurnMessage,
				ToolName:         nil,
				Status:           "completed",
				ErrorMessage:     nil,
				DurationMs:       msgDurMs,
				ContextTokens:    ctxTokens,
				ReasoningContent: reasoning,
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

			if rawMD.Time != nil && rawMD.Time.Created != nil {
				c := int64(*rawMD.Time.Created)
				if acc.firstTS == nil || c < *acc.firstTS {
					acc.firstTS = &c
				}
				var finish *int64
				if rawMD.Time.Completed != nil {
					v := int64(*rawMD.Time.Completed)
					finish = &v
				}
				completed, provisional, pending, durationErr := domain.DurationInterval(c, finish, cutoff.UnixMilli(), runningMessages[mid])
				if durationErr != nil {
					return nil, nil, nil, errors.Join(fmt.Errorf("OpenCode schema family v1 message %s activity: %w", mid, durationErr), msgRows.Close())
				}
				acc.activeMS += completed
				acc.provisionalMS += provisional
				acc.provisional = acc.provisional || pending
				if finish == nil && !pending {
					acc.durationIncomplete = true
				}
				if finish != nil && *finish >= c && *finish <= cutoff.UnixMilli() && (acc.lastTS == nil || *finish > *acc.lastTS) {
					acc.lastTS = finish
				}
			} else {
				acc.durationIncomplete = true
			}

			acc.input += inp
			acc.output += out
			acc.cache += cache
			acc.reasoning += reas
			if inp+cache > acc.peak {
				acc.peak = inp + cache
			}

			if rawMD.Cost != nil {
				if !domain.ValidCostAmount(*rawMD.Cost) {
					return nil, nil, nil, errors.Join(fmt.Errorf("invalid cost in message %s session %s", mid, s.id), msgRows.Close())
				}
				acc.costPresent++
				acc.cost += *rawMD.Cost
			} else {
				acc.costMissing++
			}
		}
		if err := msgRows.Err(); err != nil {
			return nil, nil, nil, errors.Join(fmt.Errorf("OpenCode schema family v1: iterating messages in session %s: %w", s.id, err), msgRows.Close())
		}
		if err := msgRows.Close(); err != nil {
			return nil, nil, nil, fmt.Errorf("OpenCode schema family v1: closing messages in session %s: %w", s.id, err)
		}

		role = persona.Resolve(s.agent)
		sessionRoleMap[s.id] = role
		for i := eventStart; i < len(events); i++ {
			events[i].Role = role
		}

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

	// Both extraction modes inspect every part: a malformed record cannot
	// establish complete optional-category evidence by escaping a tool filter.
	// Strict mode rejects it; permissive mode records incomplete owner coverage.
	// 4. Query parts (tool calls and compactions)
	qParts := fmt.Sprintf(`
		SELECT p.id, p.session_id, p.message_id, p.time_created, p.time_updated, p.data
		FROM part p
		JOIN session s ON p.session_id = s.id
		WHERE (s.id IN (%s) OR s.parent_id IN (%s))
		ORDER BY p.time_created ASC
	`, quotedList, parentList)

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
			if strict {
				return nil, nil, nil, fmt.Errorf("OpenCode schema family v1: empty part %s in session %s", pid.String, sid.String)
			}
			incompleteOwners[sid.String] = true
			continue
		}

		var rawPD schemaV1PartData
		if err := json.Unmarshal([]byte(pData.String), &rawPD); err != nil {
			if strict {
				return nil, nil, nil, fmt.Errorf("OpenCode schema family v1: decoding part %s in session %s: %w", pid.String, sid.String, err)
			}
			incompleteOwners[sid.String] = true
			continue
		}

		if rawPD.Type == nil {
			if strict {
				return nil, nil, nil, fmt.Errorf("OpenCode schema family v1: missing type in part %s in session %s", pid.String, sid.String)
			}
			incompleteOwners[sid.String] = true
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

		role := sessionRoleMap[sID]

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

		if rawPD.Text != nil {
			if *rawPD.Type == "text" {
				if messageContent[mID] != "" {
					messageContent[mID] += "\n"
				}
				messageContent[mID] += *rawPD.Text
			}
			if *rawPD.Type == "reasoning" {
				for i := range events {
					event := &events[i]
					if event.MessageID != nil && *event.MessageID == mID && event.EventType == domain.EventTypeTurnMessage {
						content := *rawPD.Text
						if event.ReasoningContent != nil {
							content = *event.ReasoningContent + "\n" + content
						}
						event.ReasoningContent = &content
					}
				}
			}
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
			if stat == "running" || stat == "pending" {
				eventStatus = stat
				durMs = nil
			}
			var toolInput, toolOutput *string
			if rawPD.State != nil {
				toolInput = normalizeToolInput(rawPD.State.Input)
				if rawPD.State.Output != nil {
					outputValue := *rawPD.State.Output
					toolOutput = &outputValue
				}
			}

			var invokedAt *string
			if tCreated.Valid && tCreated.Int64 > 0 {
				value := formatEpochMilli(tCreated.Int64)
				invokedAt = &value
			}
			tName := toolName
			events = append(events, domain.EventFact{
				InvokedAt:     invokedAt,
				EventID:       partID,
				Timestamp:     ts,
				SessionID:     sID,
				MessageID:     msgIDPtr,
				Turn:          turn,
				Role:          role,
				EventType:     domain.EventTypeToolCall,
				ToolName:      &tName,
				Status:        eventStatus,
				ErrorMessage:  classification.ErrorMessage,
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

			activeSecs, provisionalSeconds := domain.MeasuredDuration(st.activeMS, st.provisionalMS)
			var completedAt *string
			if st.lastTS != nil && !st.provisional && !st.durationIncomplete {
				completedAt = &endISO
			}

			invocations = append(invocations, domain.InvocationFact{
				Role:                  role,
				Model:                 mName,
				StartedAt:             &stISO,
				CompletedAt:           completedAt,
				ActiveDurationSeconds: activeSecs,
				ProvisionalSeconds:    provisionalSeconds,
				Turns:                 turns,
				Tokens: domain.InvocationTokens{
					Input:       st.input,
					Output:      st.output,
					CacheRead:   st.cache,
					Reasoning:   st.reasoning,
					PeakContext: st.peak,
				},
				CostUSD:        math.Round(st.cost*1e6) / 1e6,
				CostEvidence:   st.costEvidence(),
				Compactions:    compactions,
				UserIterations: userIterations,
			})
		}

		sTools := toolsBySession[s.id]
		if sTools == nil {
			sTools = []domain.SessionToolFact{}
		}

		sessionFacts = append(sessionFacts, domain.SessionFact{
			ID:          s.id,
			ParentID:    s.parentID,
			Source:      domain.SourceOpenCode,
			User:        sessionUser,
			Invocations: invocations,
			Tools:       sTools,
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

	if err := proveSourceChildLinks(db, "session", events); err != nil {
		return nil, nil, nil, err
	}
	projected, err := projectSourceFacts(sessionFacts, events, strict, policy, incompleteOwners)
	if err != nil {
		return nil, nil, nil, err
	}
	return projected, startedAtISO, completedAtISO, nil
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

func extractSchemaV1Sessions(db *sql.DB, ids []string, user string, users map[string]string, strict bool, exact ...bool) ([]domain.SessionFact, *string, *string, error) {
	return extractSchemaV1SessionsWithEvents(db, ids, user, users, strict, nil, exact...)
}

package opencode

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"agenticlens/internal/domain"
)

type schemaV2SessionRow struct {
	id               string
	parent           sql.NullString
	agent, model     sql.NullString
	created, updated int64
}
type schemaV2MessageRow struct {
	id, kind, data        string
	seq, created, updated int64
}
type schemaV2MessagePayload struct {
	Time struct {
		Created   *int64 `json:"created"`
		Completed *int64 `json:"completed"`
	} `json:"time"`
	Text  *string `json:"text"`
	Agent string  `json:"agent"`
	Model struct {
		ID       string `json:"id"`
		Provider string `json:"providerID"`
	} `json:"model"`
	Tokens *struct {
		Input     *int64 `json:"input"`
		Output    *int64 `json:"output"`
		Reasoning *int64 `json:"reasoning"`
		Cache     *struct {
			Read *int64 `json:"read"`
		} `json:"cache"`
	} `json:"tokens"`
	Cost    *float64 `json:"cost"`
	Outcome string   `json:"outcome"`
	Content []struct {
		Type, ID, Name string
		Time           struct{ Created, Ran, Completed *int64 } `json:"time"`
		Text           string                                   `json:"text"`
		State          struct {
			Status  string          `json:"status"`
			Input   json.RawMessage `json:"input"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			Error json.RawMessage                          `json:"error"`
			Time  struct{ Created, Ran, Completed *int64 } `json:"time"`
		} `json:"state"`
	} `json:"content"`
}

type schemaV2ModelJSON struct {
	ID *string `json:"id"`
}

func extractSchemaV2SessionsWithEvents(db *sql.DB, roots []string, defaultUser string, userBySession map[string]string, strict bool, eventSink *[]domain.EventFact, exact ...bool) ([]domain.SessionFact, *string, *string, error) {
	return extractSchemaV2SessionsWithPolicyAt(db, roots, defaultUser, userBySession, strict, eventSink, nil, time.Now().UTC(), exact...)
}

func extractSchemaV2SessionsWithEventsAt(db *sql.DB, roots []string, defaultUser string, userBySession map[string]string, strict bool, eventSink *[]domain.EventFact, cutoff time.Time, exact ...bool) ([]domain.SessionFact, *string, *string, error) {
	return extractSchemaV2SessionsWithPolicyAt(db, roots, defaultUser, userBySession, strict, eventSink, nil, cutoff, exact...)
}

func extractSchemaV2SessionsWithPolicy(db *sql.DB, roots []string, defaultUser string, userBySession map[string]string, strict bool, eventSink *[]domain.EventFact, policy *domain.GenerationPolicy, exact ...bool) ([]domain.SessionFact, *string, *string, error) {
	return extractSchemaV2SessionsWithPolicyAt(db, roots, defaultUser, userBySession, strict, eventSink, policy, time.Now().UTC(), exact...)
}

func extractSchemaV2SessionsWithPolicyAt(db *sql.DB, roots []string, defaultUser string, userBySession map[string]string, strict bool, eventSink *[]domain.EventFact, policy *domain.GenerationPolicy, cutoff time.Time, exact ...bool) ([]domain.SessionFact, *string, *string, error) {
	ids, err := buildQuotedIDList(roots)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("validating thread ids: %w", err)
	}
	parentIDs := ids
	if len(exact) > 0 && exact[0] {
		parentIDs = "NULL"
	}
	// v2 explicitly records native idle lifecycle state. Streaming text has no
	// per-item start/end fields, so message creation alone is insufficient.
	activeSessions := map[string]bool{}
	columns, columnErr := tableColumns(db)
	if columnErr != nil {
		return nil, nil, nil, columnErr
	}
	if columns["session_v2"]["time_idle"] {
		activityRows, activityErr := db.Query(fmt.Sprintf(`SELECT id FROM session_v2 WHERE time_idle IS NULL AND (id IN (%s) OR parent_id IN (%s))`, ids, parentIDs))
		if activityErr != nil {
			return nil, nil, nil, activityErr
		}
		for activityRows.Next() {
			var id string
			if err := activityRows.Scan(&id); err != nil {
				activityRows.Close()
				return nil, nil, nil, err
			}
			activeSessions[id] = true
		}
		if err := activityRows.Err(); err != nil {
			activityRows.Close()
			return nil, nil, nil, err
		}
		if err := activityRows.Close(); err != nil {
			return nil, nil, nil, err
		}
	}
	rows, err := db.Query(fmt.Sprintf(`SELECT id,parent_id,agent,model,time_created,time_updated FROM session_v2 WHERE id IN (%s) OR parent_id IN (%s) ORDER BY time_created,id`, ids, parentIDs))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("querying OpenCode schema family v2 sessions: %w", err)
	}
	var sessions []schemaV2SessionRow
	for rows.Next() {
		var s schemaV2SessionRow
		if err := rows.Scan(&s.id, &s.parent, &s.agent, &s.model, &s.created, &s.updated); err != nil {
			rows.Close()
			return nil, nil, nil, fmt.Errorf("scanning OpenCode schema family v2 session: %w", err)
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, nil, nil, fmt.Errorf("iterating OpenCode schema family v2 sessions: %w", err)
	}
	rows.Close()
	var start, end *string
	var minCreated, maxUpdated int64
	for i, s := range sessions {
		if i == 0 || s.created < minCreated {
			minCreated = s.created
		}
		if i == 0 || s.updated > maxUpdated {
			maxUpdated = s.updated
		}
	}
	if len(sessions) > 0 {
		v := formatEpochMilli(minCreated)
		start = &v
		v2 := formatEpochMilli(maxUpdated)
		end = &v2
	}
	var invocationSessions []domain.SessionFact
	var tools []domain.ToolFact
	var events []domain.EventFact
	defer func() {
		if eventSink != nil {
			*eventSink = append(*eventSink, events...)
		}
	}()
	for _, s := range sessions {
		role := domain.AgentRole(strings.TrimSpace(s.agent.String))
		var persona domain.PersonaEvidence
		eventStart := len(events)
		mr, e := db.Query(`SELECT id,type,seq,time_created,time_updated,data FROM session_message WHERE session_id=? ORDER BY seq,time_created,id`, s.id)
		if e != nil {
			return nil, nil, nil, fmt.Errorf("querying OpenCode schema family v2 messages for session %s: %w", s.id, e)
		}
		var msgs []schemaV2MessageRow
		for mr.Next() {
			var m schemaV2MessageRow
			if e := mr.Scan(&m.id, &m.kind, &m.seq, &m.created, &m.updated, &m.data); e != nil {
				mr.Close()
				return nil, nil, nil, fmt.Errorf("scanning OpenCode schema family v2 message in session %s: %w", s.id, e)
			}
			msgs = append(msgs, m)
		}
		if e := mr.Err(); e != nil {
			mr.Close()
			return nil, nil, nil, fmt.Errorf("iterating OpenCode schema family v2 messages in session %s: %w", s.id, e)
		}
		mr.Close()
		models := map[string]*sessionModelAccumulator{}
		order := []string{}
		var userTurns, assistantTurns int64
		lookup := map[string]int64{}
		toolStats := map[string]*toolAccumulator{}
		for mi, m := range msgs {
			var p schemaV2MessagePayload
			if e := json.Unmarshal([]byte(m.data), &p); e != nil {
				return nil, nil, nil, fmt.Errorf("OpenCode schema family v2: decoding %s message %s in session %s: %w", m.kind, m.id, s.id, e)
			}
			ts := m.created
			if p.Time.Created != nil {
				ts = *p.Time.Created
			}
			stamp := formatEpochMilli(ts)
			switch m.kind {
			case "user":
				userTurns++
				var prompt *string
				if p.Text != nil {
					prompt = p.Text
				}
				v := m.id
				events = append(events, domain.EventFact{EventID: sourceEventID("message", m.id), Timestamp: stamp, SessionID: s.id, MessageID: &v, Turn: userTurns, Role: role, EventType: domain.EventTypeTurnMessage, Status: "completed", PromptContent: prompt})
			case "idle":
				continue
			case "assistant":
				persona = persona.WithCandidate(p.Agent)
				if p.Model.ID == "" {
					return nil, nil, nil, fmt.Errorf("OpenCode schema family v2: assistant message %s in session %s is missing model.id", m.id, s.id)
				}
				assistantTurns++
				lookup[m.id] = assistantTurns
				acc, ok := models[p.Model.ID]
				if !ok {
					acc = &sessionModelAccumulator{}
					models[p.Model.ID] = acc
					order = append(order, p.Model.ID)
				}
				acc.asstMsgs++
				if p.Tokens != nil {
					if p.Tokens.Input != nil {
						acc.input += *p.Tokens.Input
					}
					if p.Tokens.Output != nil {
						acc.output += *p.Tokens.Output
					}
					if p.Tokens.Reasoning != nil {
						acc.reasoning += *p.Tokens.Reasoning
					}
					if p.Tokens.Cache != nil && p.Tokens.Cache.Read != nil {
						acc.cache += *p.Tokens.Cache.Read
					}
				}
				ctx := int64(0)
				if p.Tokens != nil {
					ctx = schemaV2Value(p.Tokens.Input)
					if p.Tokens.Cache != nil {
						ctx += schemaV2Value(p.Tokens.Cache.Read)
					}
				}
				if ctx > acc.peak {
					acc.peak = ctx
				}
				if p.Cost != nil {
					if !domain.ValidCostAmount(*p.Cost) {
						return nil, nil, nil, fmt.Errorf("invalid cost in message %s session %s", m.id, s.id)
					}
					acc.costPresent++
					acc.cost += *p.Cost
				} else {
					acc.costMissing++
				}
				if acc.firstTS == nil || ts < *acc.firstTS {
					v := ts
					acc.firstTS = &v
				}
				running := activeSessions[s.id]
				for _, part := range p.Content {
					if part.Type == "tool" && part.State.Status == "running" {
						running = true
					}
				}
				// A later idle/user/assistant record terminates current execution evidence.
				for _, later := range msgs[mi+1:] {
					if later.created <= cutoff.UnixMilli() {
						running = false
						break
					}
				}
				completedMS, provisionalMS, pending, limitation := domain.DurationInterval(ts, p.Time.Completed, cutoff.UnixMilli(), running)
				acc.activeMS += completedMS
				acc.provisionalMS += provisionalMS
				acc.provisional = acc.provisional || pending
				if limitation != "" {
					acc.durationLimit = limitation
				}
				if p.Time.Completed != nil && *p.Time.Completed >= ts && *p.Time.Completed <= cutoff.UnixMilli() && (acc.lastTS == nil || *p.Time.Completed > *acc.lastTS) {
					v := *p.Time.Completed
					acc.lastTS = &v
				}
				var dp *int64
				status := "unsupported"
				if p.Time.Completed != nil && *p.Time.Completed >= ts && *p.Time.Completed <= cutoff.UnixMilli() {
					d := completedMS
					dp = &d
					status = "completed"
				} else if pending {
					status = "running"
				}

				var mid *string
				v := m.id
				mid = &v
				var messageText []string
				for _, part := range p.Content {
					if part.Type == "text" {
						messageText = append(messageText, part.Text)
					}
				}
				var assistantText *string
				if len(messageText) > 0 {
					assistantText = schemaV2StringPointer(strings.Join(messageText, "\n"))
				}
				events = append(events, domain.EventFact{EventID: sourceEventID("message", m.id), Timestamp: stamp, SessionID: s.id, MessageID: mid, Turn: assistantTurns, Role: role, EventType: domain.EventTypeTurnMessage, Status: status, DurationMs: dp, ContextTokens: ctx, PromptContent: assistantText})
				for _, part := range p.Content {
					if part.Type != "text" && part.Type != "tool" {
						return nil, nil, nil, fmt.Errorf("unsupported OpenCode schema family v2: assistant content type %q in message %s session %s", part.Type, m.id, s.id)
					}
					if part.Type != "tool" {
						continue
					}
					name := strings.TrimSpace(part.Name)
					if name == "" {
						return nil, nil, nil, fmt.Errorf("OpenCode schema family v2 tool content in message %s session %s has no name", m.id, s.id)
					}
					stat := strings.ToLower(part.State.Status)
					if stat == "" {
						return nil, nil, nil, fmt.Errorf("OpenCode schema family v2 tool %s in message %s session %s is missing state.status", name, m.id, s.id)
					}
					var failed bool
					switch stat {
					case "completed":
					case "running", "pending":
					case "error", "failed":
						failed = true
					default:
						return nil, nil, nil, fmt.Errorf("OpenCode schema family v2: unsupported tool status %q for tool %s in message %s session %s", stat, name, m.id, s.id)
					}
					ta, ok := toolStats[name]
					if !ok {
						ta = &toolAccumulator{}
						toolStats[name] = ta
					}
					ta.calls++
					if failed {
						ta.errors++
					}
					status := stat
					if failed {
						status = "error"
					}
					var toolName = name
					// Native v2 stores timing on the content item; retain the
					// older supported state.time form for imported snapshots.
					timing := part.Time
					if timing.Created == nil {
						timing = part.State.Time
					}
					var duration *int64
					if timing.Created != nil && timing.Completed != nil && *timing.Completed >= *timing.Created {
						d := *timing.Completed - *timing.Created
						duration = &d
					}
					outputParts := make([]string, 0, len(part.State.Content))
					for _, c := range part.State.Content {
						if c.Type != "text" {
							return nil, nil, nil, fmt.Errorf("OpenCode schema family v2: unsupported tool result type %q in tool %s message %s session %s", c.Type, name, m.id, s.id)
						}
						outputParts = append(outputParts, c.Text)
					}
					output := strings.Join(outputParts, "\n")
					input := normalizeToolInput(part.State.Input)
					var invokedAt *string
					if timing.Created != nil && *timing.Created > 0 {
						value := formatEpochMilli(*timing.Created)
						invokedAt = &value
					}
					events = append(events, domain.EventFact{InvokedAt: invokedAt, EventID: sourceEventID("tool", part.ID), Timestamp: stamp, SessionID: s.id, MessageID: mid, Turn: assistantTurns, Role: role, EventType: domain.EventTypeToolCall, ToolName: &toolName, Status: status, DurationMs: duration, ContextTokens: ctx, ToolInput: input, ToolOutput: schemaV2StringPointer(output)})
				}
			default:
				return nil, nil, nil, fmt.Errorf("OpenCode schema family v2: unsupported message type %q for message %s in session %s", m.kind, m.id, s.id)
			}
		}
		role = persona.Resolve(s.agent.String)
		for i := eventStart; i < len(events); i++ {
			events[i].Role = role
		}
		model := strings.TrimSpace(s.model.String)
		if len(models) == 0 {
			fallback := "default"
			if model != "" {
				var raw schemaV2ModelJSON
				if err := json.Unmarshal([]byte(model), &raw); err == nil && raw.ID != nil && strings.TrimSpace(*raw.ID) != "" {
					fallback = strings.TrimSpace(*raw.ID)
				} else {
					fallback = model
				}
			}
			models[fallback] = &sessionModelAccumulator{firstTS: &s.created, lastTS: &s.created}
			order = append(order, fallback)
		}
		user := defaultUser
		if value, ok := userBySession[s.id]; ok && strings.TrimSpace(value) != "" {
			user = strings.TrimSpace(value)
		}
		var invocations []domain.InvocationFact
		for i, name := range order {
			a := models[name]
			first, last := s.created, s.created
			if a.firstTS != nil {
				first = *a.firstTS
			}
			if a.lastTS != nil {
				last = *a.lastTS
			}
			turns := int64(len(msgs))
			if len(order) > 1 {
				turns = a.asstMsgs
			}
			var iterations int64
			if i == 0 && !s.parent.Valid {
				iterations = userTurns
			}
			activeSeconds, durationEvidence := domain.MeasuredDuration(a.activeMS, a.provisionalMS, a.provisional, cutoff.UTC().Format(time.RFC3339Nano), "opencode:v2-assistant-message", a.durationLimit)
			var completedAt *string
			if a.lastTS != nil && !a.provisional && a.durationLimit == "" {
				completedAt = schemaV2StringPointer(formatEpochMilli(last))
			}
			invocations = append(invocations, domain.InvocationFact{Role: role, Model: name, StartedAt: schemaV2StringPointer(formatEpochMilli(first)), CompletedAt: completedAt, ActiveDurationSeconds: activeSeconds, DurationEvidence: durationEvidence, Turns: turns, Tokens: domain.InvocationTokens{Input: a.input, Output: a.output, CacheRead: a.cache, Reasoning: a.reasoning, PeakContext: a.peak}, CostUSD: math.Round(a.cost*1e6) / 1e6, CostEvidence: a.costEvidence(), UserIterations: iterations})
		}
		var parent *string
		if s.parent.Valid {
			v := s.parent.String
			parent = &v
		}
		invocationSessions = append(invocationSessions, domain.SessionFact{ID: s.id, ParentID: parent, Source: domain.SourceOpenCode, User: user, Invocations: invocations})
		for name, a := range toolStats {
			tools = append(tools, domain.ToolFact{SessionID: s.id, Name: name, Calls: a.calls, Errors: a.errors})
		}
	}
	sort.Slice(tools, func(i, j int) bool {
		if tools[i].SessionID != tools[j].SessionID {
			return tools[i].SessionID < tools[j].SessionID
		}
		return tools[i].Name < tools[j].Name
	})
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Timestamp < events[j].Timestamp
	})
	result, err := domain.SessionsFromFactsWithSource(domain.SourceOpenCode, invocationSessions, tools, events)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("grouping OpenCode schema family v2 facts into sessions: %w", err)
	}
	if err := proveSourceChildLinks(db, "session_v2", events); err != nil {
		return nil, nil, nil, err
	}
	projected, err := projectSourceFacts(result, events, strict, policy)
	if err != nil {
		return nil, nil, nil, err
	}
	return projected, start, end, nil
}

func schemaV2StringPointer(s string) *string { return &s }
func schemaV2Value(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func countSchemaV2SteeringTurns(db *sql.DB, threadIDs []string) (int64, error) {
	quoted, err := buildQuotedIDList(threadIDs)
	if err != nil {
		return 0, fmt.Errorf("validating thread ids: %w", err)
	}
	var count int64
	err = db.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM session_message WHERE type='user' AND session_id IN (SELECT id FROM session_v2 WHERE id IN (%s) OR parent_id IN (%s))`, quoted, quoted)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting OpenCode schema family v2 user messages: %w", err)
	}
	return count, nil
}

func extractSchemaV2Sessions(db *sql.DB, ids []string, user string, users map[string]string, exact ...bool) ([]domain.SessionFact, *string, *string, error) {
	return extractSchemaV2SessionsWithEvents(db, ids, user, users, true, nil, exact...)
}

package opencode

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

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
		Text           string `json:"text"`
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

func extractSchemaV2Sessions(db *sql.DB, roots []string, defaultUser string, userBySession map[string]string) ([]domain.SessionFact, *string, *string, error) {
	ids, err := buildQuotedIDList(roots)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("validating thread ids: %w", err)
	}
	rows, err := db.Query(fmt.Sprintf(`SELECT id,parent_id,agent,model,time_created,time_updated FROM session_v2 WHERE id IN (%s) OR parent_id IN (%s) ORDER BY time_created,id`, ids, ids))
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
	for _, s := range sessions {
		role := domain.AgentRole(strings.TrimSpace(s.agent.String))
		if role == "" {
			if s.parent.Valid {
				role = domain.RoleUnassignedSubAgent
			} else {
				role = domain.RoleRootAgent
			}
		}
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
		for _, m := range msgs {
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
					acc.cost += *p.Cost
				}
				if acc.firstTS == nil || ts < *acc.firstTS {
					v := ts
					acc.firstTS = &v
				}
				completed := m.updated
				if p.Time.Completed != nil {
					completed = *p.Time.Completed
				}
				if completed >= ts {
					acc.activeMS += completed - ts
				}
				if acc.lastTS == nil || completed > *acc.lastTS {
					v := completed
					acc.lastTS = &v
				}
				d := completed - ts
				var dp *int64
				if d >= 0 {
					dp = &d
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
				events = append(events, domain.EventFact{EventID: sourceEventID("message", m.id), Timestamp: stamp, SessionID: s.id, MessageID: mid, Turn: assistantTurns, Role: role, EventType: domain.EventTypeTurnMessage, Status: "completed", DurationMs: dp, ContextTokens: ctx, PromptContent: assistantText})
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
					status := "completed"
					if failed {
						status = "error"
					}
					var toolName = name
					var duration *int64
					if part.State.Time.Created != nil && part.State.Time.Completed != nil && *part.State.Time.Completed >= *part.State.Time.Created {
						d := *part.State.Time.Completed - *part.State.Time.Created
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
					content := domain.ClassifyCompletedToolContent(name, stat, output, failed)
					if content.HasContentError {
						ta.contentErrors++
					}
					input := normalizeToolInput(part.State.Input)
					events = append(events, domain.EventFact{EventID: sourceEventID("tool", part.ID), Timestamp: stamp, SessionID: s.id, MessageID: mid, Turn: assistantTurns, Role: role, EventType: domain.EventTypeToolCall, ToolName: &toolName, Status: status, DurationMs: duration, ContextTokens: ctx, HasContentError: content.HasContentError, ContentSeverity: string(content.Severity), ContentErrorSnippet: content.ErrorSnippet, ToolInput: input, ToolOutput: schemaV2StringPointer(output)})
				}
			default:
				return nil, nil, nil, fmt.Errorf("OpenCode schema family v2: unsupported message type %q for message %s in session %s", m.kind, m.id, s.id)
			}
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
			invocations = append(invocations, domain.InvocationFact{Role: role, Model: name, StartedAt: schemaV2StringPointer(formatEpochMilli(first)), CompletedAt: schemaV2StringPointer(formatEpochMilli(last)), ActiveDurationSeconds: int64(math.Round(float64(a.activeMS) / 1000)), Turns: turns, Tokens: domain.InvocationTokens{Input: a.input, Output: a.output, CacheRead: a.cache, Reasoning: a.reasoning, PeakContext: a.peak}, CostUSD: math.Round(a.cost*1e6) / 1e6, UserIterations: iterations})
		}
		var parent *string
		if s.parent.Valid {
			v := s.parent.String
			parent = &v
		}
		invocationSessions = append(invocationSessions, domain.SessionFact{ID: s.id, ParentID: parent, Source: domain.SourceOpenCode, User: user, Invocations: invocations})
		for name, a := range toolStats {
			tools = append(tools, domain.ToolFact{SessionID: s.id, Name: name, Calls: a.calls, Errors: a.errors, ContentErrors: a.contentErrors})
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
	return result, start, end, nil
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

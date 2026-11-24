package claude

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"agenticlens/internal/domain"
)

// Extraction is the request-local result of one proven tree. Sessions hold
// only owners read completely; Events and CheckpointEvidence are transient.
type Extraction struct {
	Sessions           []domain.SessionFact
	Events             []domain.EventFact
	CheckpointEvidence []domain.CheckpointInvocationEvidence
	CompleteSessionIDs []string
	// RootIncomplete lists reasons that prevent replacing the root snapshot.
	RootIncomplete []string
	// Diagnostics name unresolved, unavailable and incomplete subagents.
	Diagnostics      []string
	StartedAt        string
	WorkingDirectory string
}

type ownerInput struct {
	transcript Transcript
	parentID   string
	role       domain.AgentRole
	root       bool
}

// Facts builds session facts for the root and every proven child that was read
// completely. It is pure: the cutoff and generation policy are explicit inputs.
// Incomplete owners are omitted, never recorded as zero.
func (tree ProvenTree) Facts(cutoff time.Time, policy *domain.GenerationPolicy) (Extraction, error) {
	cutoff = cutoff.UTC()
	result := Extraction{WorkingDirectory: tree.Root.WorkingDirectory, Diagnostics: append([]string(nil), tree.Diagnostics...)}
	owners := []ownerInput{{transcript: tree.Root, root: true}}
	for _, child := range tree.Children {
		owners = append(owners, ownerInput{transcript: child.Transcript, parentID: child.ParentID, role: domain.AgentRole(child.Role)})
	}
	for _, owner := range owners {
		facts := buildOwner(owner, tree.childByToolUse, cutoff)
		if len(facts.incomplete) > 0 {
			if owner.root {
				result.RootIncomplete = facts.incomplete
			} else {
				result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("Claude Code subagent %s incomplete; stored facts preserved: %s", owner.transcript.OwnerID, strings.Join(facts.incomplete, "; ")))
			}
			continue
		}
		sessions, err := domain.SessionsFromFactsWithSource(domain.SourceClaude, []domain.SessionFact{facts.session}, facts.tools, facts.events)
		if err != nil {
			return Extraction{}, fmt.Errorf("grouping Claude Code facts for %s: %w", owner.transcript.OwnerID, err)
		}
		for i := range sessions {
			// Anomaly summaries are calculated by the owning operation after
			// reconciliation with the project thresholds.
			sessions[i].Anomalies = nil
		}
		if policy != nil {
			sessions = domain.ProjectGenerationFacts(domain.AgenticFactLedger{Sessions: sessions}, *policy).Sessions
		}
		result.Sessions = append(result.Sessions, sessions...)
		result.Events = append(result.Events, facts.events...)
		result.CheckpointEvidence = append(result.CheckpointEvidence, facts.evidence...)
		result.CompleteSessionIDs = append(result.CompleteSessionIDs, owner.transcript.OwnerID)
		if facts.startedAt != "" && (result.StartedAt == "" || earlier(facts.startedAt, result.StartedAt)) {
			result.StartedAt = facts.startedAt
		}
	}
	sort.SliceStable(result.Events, func(i, j int) bool { return earlier(result.Events[i].Timestamp, result.Events[j].Timestamp) })
	return result, nil
}

func earlier(left, right string) bool {
	l, lErr := time.Parse(time.RFC3339Nano, left)
	r, rErr := time.Parse(time.RFC3339Nano, right)
	if lErr != nil || rErr != nil {
		return left < right
	}
	return l.Before(r)
}

type ownerFacts struct {
	session    domain.SessionFact
	tools      []domain.ToolFact
	events     []domain.EventFact
	evidence   []domain.CheckpointInvocationEvidence
	incomplete []string
	startedAt  string
}

type stamp struct {
	at  time.Time
	raw string
}

type modelAccumulator struct {
	tokens            domain.InvocationTokens
	turns             map[int]bool
	activeMillis      int64
	provisionalMillis int64
	provisional       bool
	started           stamp
	completed         stamp
	intervals         []domain.CheckpointInterval
	intervalsUnknown  bool
}

func (acc *modelAccumulator) observeStart(at time.Time, raw string) {
	if acc.started.raw == "" || at.Before(acc.started.at) {
		acc.started = stamp{at: at, raw: raw}
	}
}

func (acc *modelAccumulator) observeEnd(at time.Time, raw string) {
	if acc.completed.raw == "" || at.After(acc.completed.at) {
		acc.completed = stamp{at: at, raw: raw}
	}
}

type durationState struct {
	models     map[string]*modelAccumulator
	steering   int64
	current    string
	byResponse map[string]response
}

func (state *durationState) model(name string) *modelAccumulator {
	acc := state.models[name]
	if acc == nil {
		acc = &modelAccumulator{turns: make(map[int]bool)}
		state.models[name] = acc
	}
	return acc
}

type openTurn struct {
	index     int
	start     timelineEntry
	fallback  string
	model     string
	lastOwned time.Time
	lastRaw   string
}

type closedTurn struct {
	acc *modelAccumulator
	end time.Time
	key string
}

func buildOwner(owner ownerInput, childByToolUse map[string]string, cutoff time.Time) ownerFacts {
	t := owner.transcript
	facts := ownerFacts{incomplete: append([]string(nil), t.Incomplete...)}
	for _, use := range t.toolUses {
		if use.name == "Skill" && skillSucceeded(t.results, use.id) && use.skill == "" {
			facts.incomplete = append(facts.incomplete, "successful Skill call "+use.id+" has no skill name")
		}
	}
	if len(facts.incomplete) > 0 {
		return facts
	}
	state := &durationState{models: make(map[string]*modelAccumulator), byResponse: make(map[string]response, len(t.responses))}
	for _, current := range t.responses {
		state.byResponse[current.key] = current
		acc := state.model(current.model)
		acc.tokens.Input += current.input
		acc.tokens.Output += current.output
		acc.tokens.CacheRead += current.cacheRead
		acc.tokens.Reasoning += current.thinking
		acc.tokens.PeakContext = max(acc.tokens.PeakContext, current.context())
		if at, err := time.Parse(time.RFC3339Nano, current.timestamp); err == nil {
			acc.observeStart(at, current.timestamp)
			acc.observeEnd(at, current.timestamp)
		}
	}
	measureTurns(t, state, owner.root, cutoff)
	if len(state.models) == 0 {
		state.model("")
	}
	names := make([]string, 0, len(state.models))
	for name := range state.models {
		names = append(names, name)
	}
	sort.Strings(names)
	ownerID := t.OwnerID
	session := domain.SessionFact{ID: ownerID, Source: domain.SourceClaude, Skills: []domain.SkillFact{}, Delegations: []domain.DelegationFact{}}
	if owner.parentID != "" {
		parent := owner.parentID
		session.ParentID = &parent
	}
	for index, name := range names {
		acc := state.models[name]
		active, provisional := domain.MeasuredDuration(acc.activeMillis, acc.provisionalMillis)
		invocation := domain.InvocationFact{
			Role:                  owner.role,
			Model:                 name,
			ActiveDurationSeconds: active,
			ProvisionalSeconds:    provisional,
			Turns:                 int64(len(acc.turns)),
			Tokens:                acc.tokens,
			CostEvidence:          domain.UnavailableCost(),
		}
		if acc.started.raw != "" {
			started := acc.started.raw
			invocation.StartedAt = &started
		}
		if acc.completed.raw != "" && !acc.provisional {
			completed := acc.completed.raw
			invocation.CompletedAt = &completed
		}
		if index == 0 {
			invocation.Compactions = int64(len(t.compactions))
			if owner.root {
				invocation.UserIterations = state.steering
			}
		}
		session.Invocations = append(session.Invocations, invocation)
		facts.evidence = append(facts.evidence, domain.CheckpointInvocationEvidence{
			Source: domain.SourceClaude, SessionID: ownerID, InvocationRow: index + 1, Model: name,
			Cutoff: cutoff, Complete: !acc.intervalsUnknown, Intervals: acc.intervals,
		})
		if acc.started.raw != "" && (facts.startedAt == "" || earlier(acc.started.raw, facts.startedAt)) {
			facts.startedAt = acc.started.raw
		}
	}
	callerRole := domain.KnownString(string(owner.role))
	toolStats := make(map[string]*domain.ToolFact)
	for _, use := range t.toolUses {
		stats := toolStats[use.name]
		if stats == nil {
			stats = &domain.ToolFact{SessionID: ownerID, Name: use.name}
			toolStats[use.name] = stats
		}
		stats.Calls++
		result, finished := t.results[use.id]
		failed := finished && toolFailed(use, result)
		if failed {
			stats.Errors++
		}
		name := use.name
		event := domain.EventFact{
			EventID: "claude-tool:" + use.id, CallID: use.id, Timestamp: use.timestamp, SessionID: ownerID,
			Turn: use.turn, Role: owner.role, EventType: domain.EventTypeToolCall, ToolName: &name,
			Status: "running", ContextTokens: state.byResponse[use.responseKey].context(),
		}
		if finished {
			event.Status = "completed"
			if failed {
				event.Status = "error"
			}
			if duration, ok := elapsedMillis(use.timestamp, result.timestamp); ok {
				event.DurationMs = &duration
			}
		}
		switch use.name {
		case "Skill":
			if finished && !failed {
				session.Skills = append(session.Skills, domain.SkillFact{Name: use.skill, CallerRole: callerRole, InvokedAt: domain.KnownString(use.timestamp)})
				invoked := use.timestamp
				facts.events = append(facts.events, domain.EventFact{
					EventID: "claude-skill:" + use.id, CallID: use.id, Timestamp: use.timestamp, InvokedAt: &invoked, SessionID: ownerID,
					Turn: use.turn, Role: owner.role, SkillName: use.skill, EventType: domain.EventTypeSkillActivation,
					Status: "completed", ContextTokens: event.ContextTokens,
				})
			}
		case "Agent":
			child := childByToolUse[use.id]
			if child == "" {
				child = result.agentID
			}
			agentType := use.subagentType
			if agentType == "" {
				agentType = result.agentType
			}
			session.Delegations = append(session.Delegations, domain.DelegationFact{AgentType: domain.KnownString(agentType), Description: use.description, InvokedAt: domain.KnownString(use.timestamp), ChildSession: child})
			event.ChildSessionID = child
		}
		facts.events = append(facts.events, event)
	}
	for index, raw := range t.compactions {
		facts.events = append(facts.events, domain.EventFact{
			EventID: fmt.Sprintf("claude-compaction:%s:%d", ownerID, index+1), Timestamp: raw, SessionID: ownerID,
			Role: owner.role, EventType: domain.EventTypeCompaction, Status: "completed",
		})
	}
	names = names[:0]
	for name := range toolStats {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		facts.tools = append(facts.tools, *toolStats[name])
	}
	facts.session = session
	return facts
}

func skillSucceeded(results map[string]toolResult, id string) bool {
	result, finished := results[id]
	return finished && !toolFailed(toolUse{name: "Skill"}, result)
}

// toolFailed treats an error result, or a Skill result reporting success:false,
// as a failed call.
func toolFailed(use toolUse, result toolResult) bool {
	if result.isError {
		return true
	}
	return use.name == "Skill" && result.success != nil && !*result.success
}

func elapsedMillis(startRaw, endRaw string) (int64, bool) {
	start, startErr := time.Parse(time.RFC3339Nano, startRaw)
	end, endErr := time.Parse(time.RFC3339Nano, endRaw)
	if startErr != nil || endErr != nil || end.Before(start) {
		return 0, false
	}
	return end.Sub(start).Milliseconds(), true
}

// measureTurns applies AL-64 duration rules. A turn starts at a qualifying
// prompt and ends at the owner's terminal assistant record or at an interrupt. A turn superseded
// by a new prompt ends at its last owned record, a final open turn adds
// provisional time up to the cutoff, and idle time between turns is excluded.
// The cutoff gates lifecycle records only; usage is not filtered.
func measureTurns(t Transcript, state *durationState, root bool, cutoff time.Time) {
	// A turn without a served response is attributed to the model last served
	// before it, otherwise to the owner's first served model.
	firstModel := ""
	if len(t.responses) > 0 {
		firstModel = t.responses[0].model
	}
	fallback := func() string {
		if state.current != "" {
			return state.current
		}
		return firstModel
	}
	var open *openTurn
	var closed *closedTurn
	turnIndex := 0
	finish := func(end time.Time, endRaw string) *modelAccumulator {
		name := open.model
		if name == "" {
			name = open.fallback
		}
		acc := state.model(name)
		if !end.Before(open.start.at) {
			acc.activeMillis += end.Sub(open.start.at).Milliseconds()
		}
		acc.observeStart(open.start.at, open.start.raw)
		acc.observeEnd(end, endRaw)
		open = nil
		return acc
	}
	for _, entry := range t.timeline {
		if entry.kind == entryPrompt && root && !entry.system {
			state.steering++
		}
		if entry.at.After(cutoff) {
			continue
		}
		if entry.kind == entryResume && open != nil {
			if !entry.at.Before(open.lastOwned) {
				open.lastOwned, open.lastRaw = entry.at, entry.raw
			}
			continue
		}
		switch entry.kind {
		case entryPrompt, entryResume:
			if open != nil {
				finish(open.lastOwned, open.lastRaw)
			}
			closed = nil
			turnIndex++
			open = &openTurn{index: turnIndex, start: entry, fallback: fallback(), lastOwned: entry.at, lastRaw: entry.raw}
		case entryInterrupt:
			// Esc ends the owner's open turn at the interrupt record. It is not a
			// response, so a later record cannot extend that turn.
			if open != nil {
				finish(entry.at, entry.raw)
			}
			closed = nil
		case entryAssistant:
			if entry.model != "" {
				state.current = entry.model
			}
			if open == nil {
				if closed != nil && entry.responseKey != "" && entry.responseKey == closed.key && entry.at.After(closed.end) {
					closed.acc.activeMillis += entry.at.Sub(closed.end).Milliseconds()
					closed.acc.observeEnd(entry.at, entry.raw)
					closed.end = entry.at
				}
				continue
			}
			if entry.responseKey != "" {
				if served, ok := state.byResponse[entry.responseKey]; ok {
					if open.model == "" {
						open.model = served.model
					}
					state.model(served.model).turns[open.index] = true
				}
			}
			open.lastOwned, open.lastRaw = entry.at, entry.raw
			if entry.terminal {
				acc := finish(entry.at, entry.raw)
				closed = &closedTurn{acc: acc, end: entry.at, key: entry.responseKey}
			}
		default:
			if open != nil && !entry.at.Before(open.lastOwned) {
				open.lastOwned, open.lastRaw = entry.at, entry.raw
			}
		}
	}
	if open == nil {
		return
	}
	name := open.model
	if name == "" {
		name = open.fallback
	}
	acc := state.model(name)
	acc.observeStart(open.start.at, open.start.raw)
	acc.provisionalMillis += cutoff.Sub(open.start.at).Milliseconds()
	acc.provisional = true
	interval := domain.CheckpointInterval{Kind: "turn", ID: open.start.nativeID, StartedAt: open.start.at, SourceOpen: open.start.nativeID != ""}
	if interval.ID == "" {
		acc.intervalsUnknown = true
	}
	acc.intervals = append(acc.intervals, interval)
}

package migration

import (
	"fmt"
	"reflect"

	"agenticlens/internal/domain"
)

// InvariantValidator audits domain ledgers before and after a migration transition
// to guarantee that no identity, timestamp, code, invocation, tool, or event data is corrupted or lost.
type InvariantValidator struct{}

// NewInvariantValidator creates an InvariantValidator instance.
func NewInvariantValidator() *InvariantValidator {
	return &InvariantValidator{}
}

// Validate compares pre-migration and post-migration states against strict preservation invariants.
func (v *InvariantValidator) Validate(before domain.AgenticFactLedger, after domain.AgenticFactLedger, expectedTargetVersion int) error {
	// 1. Schema version invariant
	if after.SchemaVersion == nil {
		return fmt.Errorf("%w: schema_version is nil after migration; expected %d",
			ErrInvariantViolation, expectedTargetVersion)
	}
	if *after.SchemaVersion != expectedTargetVersion {
		return fmt.Errorf("%w: schema_version is %d after migration; expected %d",
			ErrInvariantViolation, *after.SchemaVersion, expectedTargetVersion)
	}

	// 2. Ticket identification invariant
	if !ptrEqual(before.TicketID, after.TicketID) {
		return fmt.Errorf("%w: ticket_id modified from %v to %v",
			ErrInvariantViolation, ptrString(before.TicketID), ptrString(after.TicketID))
	}
	if !ptrEqual(before.TicketTitle, after.TicketTitle) {
		return fmt.Errorf("%w: ticket_title modified from %v to %v",
			ErrInvariantViolation, ptrString(before.TicketTitle), ptrString(after.TicketTitle))
	}
	if !ptrEqual(before.TicketType, after.TicketType) {
		return fmt.Errorf("%w: ticket_type modified from %v to %v",
			ErrInvariantViolation, ptrString(before.TicketType), ptrString(after.TicketType))
	}

	// 3. Timestamps invariant
	if !ptrEqual(before.StartedAt, after.StartedAt) {
		return fmt.Errorf("%w: started_at modified from %v to %v",
			ErrInvariantViolation, ptrString(before.StartedAt), ptrString(after.StartedAt))
	}
	if !ptrEqual(before.CompletedAt, after.CompletedAt) {
		return fmt.Errorf("%w: completed_at modified from %v to %v",
			ErrInvariantViolation, ptrString(before.CompletedAt), ptrString(after.CompletedAt))
	}
	if before.UpdatedAt != "" && after.UpdatedAt == "" {
		return fmt.Errorf("%w: updated_at was cleared during migration",
			ErrInvariantViolation)
	}

	// 4. Git code facts invariant
	if before.Code != after.Code {
		return fmt.Errorf("%w: code facts modified from %+v to %+v",
			ErrInvariantViolation, before.Code, after.Code)
	}

	// 5. Session ownership and invocation invariants
	if len(before.Sessions) != len(after.Sessions) {
		return fmt.Errorf("%w: session count changed from %d to %d", ErrInvariantViolation, len(before.Sessions), len(after.Sessions))
	}
	for i, b := range before.Sessions {
		a := after.Sessions[i]
		if b.ID != a.ID || !ptrEqual(b.ParentID, a.ParentID) || b.Source != a.Source || b.User != a.User {
			return fmt.Errorf("%w: session[%d] identity or ownership modified", ErrInvariantViolation, i)
		}
		if len(b.Invocations) != len(a.Invocations) {
			return fmt.Errorf("%w: session[%d] invocation count changed from %d to %d", ErrInvariantViolation, i, len(b.Invocations), len(a.Invocations))
		}
	}
	afterInvocations, beforeInvocations := after.InvocationFacts(), before.InvocationFacts()
	if len(afterInvocations) != len(beforeInvocations) {
		return fmt.Errorf("%w: invocation count changed from %d to %d",
			ErrInvariantViolation, len(beforeInvocations), len(afterInvocations))
	}
	for i := range beforeInvocations {
		b := beforeInvocations[i]
		a := afterInvocations[i]
		if b.Role != a.Role ||
			b.Model != a.Model ||
			!ptrEqual(b.StartedAt, a.StartedAt) ||
			!ptrEqual(b.CompletedAt, a.CompletedAt) ||
			b.ActiveDurationSeconds != a.ActiveDurationSeconds ||
			b.Turns != a.Turns ||
			b.Tokens != a.Tokens ||
			b.CostUSD != a.CostUSD ||
			b.Compactions != a.Compactions ||
			(expectedTargetVersion < 4 && b.UserIterations != a.UserIterations) {
			return fmt.Errorf("%w: invocation[%d] data modified",
				ErrInvariantViolation, i)
		}
	}

	// 6. Tools invariant
	afterTools, beforeTools := after.ToolFacts(), before.ToolFacts()
	if len(afterTools) != len(beforeTools) {
		return fmt.Errorf("%w: tools count changed from %d to %d",
			ErrInvariantViolation, len(beforeTools), len(afterTools))
	}
	for i := range beforeTools {
		b := beforeTools[i]
		a := afterTools[i]
		if b != a {
			return fmt.Errorf("%w: tool[%d] (%s) modified from %+v to %+v",
				ErrInvariantViolation, i, b.Name, b, a)
		}
	}

	// 7. Event facts invariant
	if expectedTargetVersion >= 3 {
		// Version 3+ invariant: event pruning and error message stripping
		afterEvents, beforeEvents := after.EventFacts(), before.EventFacts()
		if len(afterEvents) > len(beforeEvents) {
			return fmt.Errorf("%w: events count increased in v3 migration from %d to %d",
				ErrInvariantViolation, len(beforeEvents), len(afterEvents))
		}
		matched := make(map[int]bool)
		for i, a := range afterEvents {
			if a.ErrorMessage != nil {
				return fmt.Errorf("%w: event[%d] in schema v3 has non-nil error_message: %s",
					ErrInvariantViolation, i, *a.ErrorMessage)
			}
			found := false
			for j, b := range beforeEvents {
				if matched[j] {
					continue
				}
				if b.Timestamp == a.Timestamp &&
					b.SessionID == a.SessionID &&
					b.Turn == a.Turn &&
					b.Role == a.Role &&
					b.EventType == a.EventType &&
					ptrEqual(b.ToolName, a.ToolName) &&
					b.Status == a.Status &&
					ptrEqual(b.DurationMs, a.DurationMs) &&
					b.ContextTokens == a.ContextTokens {
					matched[j] = true
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("%w: pruned event[%d] in schema v3 does not correspond to any pre-migration event",
					ErrInvariantViolation, i)
			}
		}

		totalsBefore := domain.ComputeTotals(beforeInvocations)
		totalsAfter := domain.ComputeTotals(afterInvocations)
		if !reflect.DeepEqual(totalsBefore, totalsAfter) {
			return fmt.Errorf("%w: ComputeTotals mismatch after migration", ErrInvariantViolation)
		}

		effBefore := domain.ComputeWorkflowEfficiency(before)
		effAfter := domain.ComputeWorkflowEfficiency(after)
		if !reflect.DeepEqual(effBefore, effAfter) {
			return fmt.Errorf("%w: ComputeWorkflowEfficiency mismatch after migration", ErrInvariantViolation)
		}
	} else {
		// Schema v1/v2: events count cannot decrease and all prior events must be strictly preserved
		afterEvents, beforeEvents := after.EventFacts(), before.EventFacts()
		if len(afterEvents) < len(beforeEvents) {
			return fmt.Errorf("%w: events count decreased from %d to %d",
				ErrInvariantViolation, len(beforeEvents), len(afterEvents))
		}
		for i := range beforeEvents {
			b := beforeEvents[i]
			a := afterEvents[i]
			if b.Timestamp != a.Timestamp ||
				b.SessionID != a.SessionID ||
				b.Turn != a.Turn ||
				b.Role != a.Role ||
				b.EventType != a.EventType ||
				!ptrEqual(b.ToolName, a.ToolName) ||
				b.Status != a.Status ||
				!ptrEqual(b.ErrorMessage, a.ErrorMessage) ||
				!ptrEqual(b.DurationMs, a.DurationMs) ||
				b.ContextTokens != a.ContextTokens {
				return fmt.Errorf("%w: pre-existing event[%d] modified",
					ErrInvariantViolation, i)
			}
		}
	}

	return nil
}

func ptrEqual[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrString[T comparable](p *T) string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprint(*p)
}

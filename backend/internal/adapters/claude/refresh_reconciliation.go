package claude

import (
	"fmt"

	"agenticlens/internal/domain"
)

// RetainChangedParents keeps the stored snapshot of every recorded Claude Code
// owner whose source parent link differs from its stored one. Such owners are
// removed from the fresh set and named in a diagnostic, so a refresh never
// relinks or replaces them. The function is pure.
func RetainChangedParents(stored, fresh []domain.SessionFact) ([]domain.SessionFact, []string) {
	recorded := make(map[string]domain.SessionFact, len(stored))
	for _, owner := range stored {
		if owner.Source == domain.SourceClaude {
			recorded[owner.ID] = owner
		}
	}
	kept := make([]domain.SessionFact, 0, len(fresh))
	var diagnostics []string
	for _, owner := range fresh {
		old, known := recorded[owner.ID]
		if known && parentChanged(old.ParentID, owner.ParentID) {
			diagnostics = append(diagnostics, fmt.Sprintf("Claude Code owner %s retained: source parent linkage changed from %s to %s", owner.ID, parentLabel(old.ParentID), parentLabel(owner.ParentID)))
			continue
		}
		kept = append(kept, owner)
	}
	return kept, diagnostics
}

func parentChanged(stored, fresh *string) bool {
	if (stored == nil) != (fresh == nil) {
		return true
	}
	return stored != nil && *stored != *fresh
}

func parentLabel(parent *string) string {
	if parent == nil {
		return "no parent"
	}
	return *parent
}

// ReplacementNotices names every recorded Claude Code owner that a complete
// read replaces with fewer invocation rows or a lower token total, so a smaller
// snapshot, such as a restored older transcript, is reported rather than
// silent. Token totals add input, output and cache-read tokens. The function
// is pure.
func ReplacementNotices(stored, fresh []domain.SessionFact) []string {
	recorded := make(map[string]domain.SessionFact, len(stored))
	for _, owner := range stored {
		if owner.Source == domain.SourceClaude {
			recorded[owner.ID] = owner
		}
	}
	var notices []string
	for _, owner := range fresh {
		old, known := recorded[owner.ID]
		if !known {
			continue
		}
		before, after := tokenTotal(old), tokenTotal(owner)
		if len(owner.Invocations) < len(old.Invocations) || after < before {
			notices = append(notices, fmt.Sprintf("Claude Code owner %s replaced by a complete read with smaller totals: invocations %d -> %d, tokens %d -> %d", owner.ID, len(old.Invocations), len(owner.Invocations), before, after))
		}
	}
	return notices
}

func tokenTotal(owner domain.SessionFact) int64 {
	var total int64
	for _, invocation := range owner.Invocations {
		total += invocation.Tokens.Input + invocation.Tokens.Output + invocation.Tokens.CacheRead
	}
	return total
}

// WithoutOwners removes the named owners' transient events and completion
// evidence, so retained owners never receive refreshed summaries.
func WithoutOwners(source domain.SourceEvents, removed map[string]bool) domain.SourceEvents {
	if len(removed) == 0 {
		return source
	}
	result := domain.SourceEvents{}
	for _, event := range source.Events {
		if !removed[event.SessionID] {
			result.Events = append(result.Events, event)
		}
	}
	for _, evidence := range source.CheckpointEvidence {
		if !removed[evidence.SessionID] {
			result.CheckpointEvidence = append(result.CheckpointEvidence, evidence)
		}
	}
	for _, id := range source.CompleteSessionIDs {
		if !removed[id] {
			result.CompleteSessionIDs = append(result.CompleteSessionIDs, id)
		}
	}
	return result
}

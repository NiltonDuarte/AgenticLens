package live

import (
	"sync"

	"agenticlens/internal/domain"
)

// TraceBuffer provides a thread-safe, strictly in-memory transient ring buffer
// retaining recent trace messages for live sessions bounded by capacity.
type TraceBuffer struct {
	mu            sync.RWMutex
	maxPerSession int
	sessions      map[string][]domain.LiveMessageFact
}

// NewTraceBuffer constructs an in-memory trace buffer with max capacity per session.
func NewTraceBuffer(maxPerSession int) *TraceBuffer {
	if maxPerSession <= 0 {
		maxPerSession = 500
	}
	return &TraceBuffer{
		maxPerSession: maxPerSession,
		sessions:      make(map[string][]domain.LiveMessageFact),
	}
}

// SetMessages stores or replaces trace messages for a session, retaining at most maxPerSession messages.
func (b *TraceBuffer) SetMessages(sessionID string, messages []domain.LiveMessageFact) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(messages) > b.maxPerSession {
		messages = messages[len(messages)-b.maxPerSession:]
	}

	copied := make([]domain.LiveMessageFact, len(messages))
	copy(copied, messages)
	b.sessions[sessionID] = copied
}

// AppendMessages appends new trace messages to an existing session trace, enforcing capacity bounds.
func (b *TraceBuffer) AppendMessages(sessionID string, messages ...domain.LiveMessageFact) {
	if len(messages) == 0 {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	existing := b.sessions[sessionID]
	combined := append(existing, messages...)

	if len(combined) > b.maxPerSession {
		combined = combined[len(combined)-b.maxPerSession:]
	}

	b.sessions[sessionID] = combined
}

// GetMessages returns a snapshot copy of buffered trace messages for a given session.
func (b *TraceBuffer) GetMessages(sessionID string) []domain.LiveMessageFact {
	b.mu.RLock()
	defer b.mu.RUnlock()

	msgs, ok := b.sessions[sessionID]
	if !ok || len(msgs) == 0 {
		return []domain.LiveMessageFact{}
	}

	result := make([]domain.LiveMessageFact, len(msgs))
	copy(result, msgs)
	return result
}

// Clear evicts buffered messages for a specific session ID.
func (b *TraceBuffer) Clear(sessionID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.sessions, sessionID)
}

// Len returns the count of buffered messages for a session.
func (b *TraceBuffer) Len(sessionID string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.sessions[sessionID])
}

package codex

import (
	"errors"
	"fmt"
	"io"
)

// ErrDiscoveryBudget means a scan did not establish a complete selected family.
// Callers retain known facts and report incomplete recovery rather than accepting
// a prefix of the scanned store as complete evidence.
var ErrDiscoveryBudget = errors.New("configured Codex source discovery budget reached")

// DiscoveryBudget is operation-local resource accounting. SeenPaths prevents
// repeated inspection of one file from consuming additional session slots; all
// actual decompressed reads still consume the byte allowance. It grants no
// authority to accept, replace or reject persisted ledger facts.
type DiscoveryBudget struct {
	MaxSessions     int
	MaxSourceBytes  int64
	SessionsRead    int
	SourceBytesRead int64
	SeenPaths       map[string]bool
}

func (b *DiscoveryBudget) validate() error {
	if b == nil {
		return nil
	}
	if b.MaxSessions < 1 || b.MaxSourceBytes < 1 || b.SessionsRead < 0 || b.SourceBytesRead < 0 || b.SessionsRead > b.MaxSessions || b.SourceBytesRead > b.MaxSourceBytes {
		return fmt.Errorf("Codex discovery requires positive explicit session/byte limits and valid operation-local usage")
	}
	seen := 0
	for _, counted := range b.SeenPaths {
		if counted {
			seen++
		}
	}
	if seen > b.SessionsRead {
		return fmt.Errorf("Codex discovery seen paths exceed the recorded session allowance usage")
	}
	return nil
}

func (b *DiscoveryBudget) claim(path string) error {
	if b == nil {
		return nil
	}
	if b.SeenPaths[path] {
		return nil
	}
	if b.SessionsRead >= b.MaxSessions {
		return ErrDiscoveryBudget
	}
	if b.SeenPaths == nil {
		b.SeenPaths = make(map[string]bool)
	}
	b.SeenPaths[path] = true
	b.SessionsRead++
	return nil
}

type discoveryBudgetReader struct {
	reader        io.Reader
	budget        *DiscoveryBudget
	remainingSize int64
	eof           bool
}

func (r *discoveryBudgetReader) Read(body []byte) (int, error) {
	if len(body) == 0 {
		return 0, nil
	}
	remaining := r.budget.MaxSourceBytes - r.budget.SourceBytesRead
	if remaining == 0 {
		if r.eof || r.remainingSize == 0 {
			return 0, io.EOF
		}
		return 0, ErrDiscoveryBudget
	}
	if int64(len(body)) > remaining {
		body = body[:remaining]
	}
	n, err := r.reader.Read(body)
	r.budget.SourceBytesRead += int64(n)
	if r.remainingSize >= 0 {
		r.remainingSize -= int64(n)
	}
	r.eof = errors.Is(err, io.EOF)
	return n, err
}

package codex

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/sys/unix"
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
	Context         context.Context
	MaxSessions     int
	MaxSourceBytes  int64
	SessionsRead    int
	SourceBytesRead int64
	SeenPaths       map[string]bool
}

func (b *DiscoveryBudget) checkDeadline() error {
	return b.readContext().Err()
}

func (b *DiscoveryBudget) readContext() context.Context {
	if b != nil && b.Context != nil {
		return b.Context
	}
	return context.Background()
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
	if err := b.checkDeadline(); err != nil {
		return err
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

// A rollout source must be a regular file. Nonblocking open lets the deadline
// remain effective when a source path changes before the descriptor is opened.
func openRolloutSource(path string) (*os.File, os.FileInfo, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, nil, err
	}
	info, err := f.Stat()
	if err == nil && !info.Mode().IsRegular() {
		err = errors.New("Codex source is not a regular file")
	}
	if err != nil {
		return nil, nil, errors.Join(err, f.Close())
	}
	return f, info, nil
}

// A regular-file read is synchronous. Bound each read and check before and
// after it, including reads made by gzip before it returns decompressed bytes.
type rolloutContextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r rolloutContextReader) Read(body []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	const maxReadBytes = 64 * 1024
	n, err := r.reader.Read(body[:min(len(body), maxReadBytes)])
	if cancelled := r.ctx.Err(); cancelled != nil {
		return n, cancelled
	}
	return n, err
}

func openRolloutReader(ctx context.Context, path string) (io.Reader, int64, func() error, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, nil, err
	}
	file, info, err := openRolloutSource(path)
	if err != nil {
		return nil, 0, nil, err
	}
	reader := rolloutContextReader{ctx: ctx, reader: file}
	if !strings.HasSuffix(path, ".gz") {
		return reader, info.Size(), file.Close, nil
	}
	compressed, err := gzip.NewReader(reader)
	if err != nil {
		return nil, 0, nil, errors.Join(fmt.Errorf("opening compressed Codex rollout: %w", err), file.Close())
	}
	return compressed, -1, func() error { return errors.Join(compressed.Close(), file.Close()) }, nil
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
	if err := r.budget.checkDeadline(); err != nil {
		return 0, err
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

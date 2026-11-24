package codex

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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

// headerReadChunk is the read granularity of capture header identification.
// Charging follows the bytes these reads return, so a header costs about its
// own length rather than a full structural record bound.
const headerReadChunk = 4096

// RolloutHeader holds the identity facts of a rollout's first session_meta
// record within the structural header bound. It never retains instructions,
// messages or event payloads.
type RolloutHeader struct {
	CreatedAt         time.Time
	ID                string
	ParentID          string
	WorkingDirectory  string
	First             bool
	ConflictingParent bool
}

// SessionContext applies the exact-owner metadata rules: the first record is a
// session_meta with an identity and working directory, and both ancestry
// spellings agree.
func (header RolloutHeader) SessionContext() (SessionMetadataContext, error) {
	if !header.First || header.ID == "" || header.WorkingDirectory == "" {
		return SessionMetadataContext{}, errors.New("exact native session metadata unavailable")
	}
	if header.ConflictingParent {
		return SessionMetadataContext{}, fmt.Errorf("%w: conflicting native parent metadata", ErrSessionContextMismatch)
	}
	return SessionMetadataContext{ID: header.ID, ParentID: header.ParentID, WorkingDirectory: header.WorkingDirectory}, nil
}

// HeaderRead reports one header lookup. Bytes is what this call read from the
// source; a cached result reads nothing.
type HeaderRead struct {
	Header RolloutHeader
	Bytes  int64
	Cached bool
}

type headerCacheKey struct {
	path    string
	size    int64
	modTime int64
}

type headerCacheEntry struct {
	err    error
	header RolloutHeader
}

// HeaderCache serves each rollout header at most once per reconciliation pass.
// Entries are keyed by physical path, size and modification time, so a file
// that changes is read again. Nothing in it is persisted. A nil cache reads
// every request.
type HeaderCache struct {
	entries   map[headerCacheKey]headerCacheEntry
	oversized map[headerCacheKey]string
}

// NewHeaderCache returns an empty pass-local header cache.
func NewHeaderCache() *HeaderCache {
	return &HeaderCache{entries: make(map[headerCacheKey]headerCacheEntry), oversized: make(map[headerCacheKey]string)}
}

// errHeaderBound reports a first session_meta beyond the structural bound.
var errHeaderBound = errors.New("native session metadata exceeds its structural record limit")

// ReadHeader returns the rollout's header, charging the bytes actually read to
// budget's byte allowance, never more than the 64 KiB structural bound. A read
// starts only when the smaller of the file size and that bound fits in the
// remaining allowance; otherwise it reads nothing and returns
// ErrDiscoveryBudget, which is not cached. It does not claim a session slot.
func (cache *HeaderCache) ReadHeader(ctx context.Context, path string, budget *DiscoveryBudget) (HeaderRead, error) {
	if err := ctx.Err(); err != nil {
		return HeaderRead{}, err
	}
	key, keyErr := headerKey(path)
	if keyErr != nil {
		return HeaderRead{}, keyErr
	}
	if cache != nil {
		if entry, ok := cache.entries[key]; ok {
			return HeaderRead{Header: entry.header, Cached: true}, entry.err
		}
	}
	// Start a read only when its structural bound fits, so a header is never
	// abandoned part way through. The charge is still the bytes actually read.
	if budget != nil && budget.MaxSourceBytes-budget.SourceBytesRead < min(key.size, maxSessionMetadataBytes) {
		return HeaderRead{}, ErrDiscoveryBudget
	}
	header, consumed, err := readRolloutHeader(ctx, path, budget)
	read := HeaderRead{Header: header, Bytes: consumed}
	if cache != nil && !errors.Is(err, ErrDiscoveryBudget) && ctx.Err() == nil {
		// A file that changed during the read is keyed by its new state, so a
		// later request reads it again.
		if after, afterErr := headerKey(path); afterErr == nil && after == key {
			if cache.entries == nil {
				cache.entries = make(map[headerCacheKey]headerCacheEntry)
			}
			cache.entries[key] = headerCacheEntry{header: header, err: err}
		}
	}
	return read, err
}

func headerKey(path string) (headerCacheKey, error) {
	physical, err := physicalRolloutPath(path)
	if err != nil {
		return headerCacheKey{}, err
	}
	info, err := os.Stat(physical)
	if err != nil {
		return headerCacheKey{}, err
	}
	if !info.Mode().IsRegular() {
		return headerCacheKey{}, errors.New("Codex source is not a regular file")
	}
	return headerCacheKey{path: physical, size: info.Size(), modTime: info.ModTime().UnixNano()}, nil
}

// countingReader records consumed bytes when no budget accounts for them.
type countingReader struct {
	reader io.Reader
	bytes  int64
}

func (r *countingReader) Read(body []byte) (int, error) {
	n, err := r.reader.Read(body)
	r.bytes += int64(n)
	return n, err
}

// readRolloutHeader scans records within the structural bound until the first
// session_meta. A first record that is not session_meta leaves First false.
func readRolloutHeader(ctx context.Context, path string, budget *DiscoveryBudget) (header RolloutHeader, consumed int64, err error) {
	reader, size, closeReader, err := openRolloutReader(ctx, path)
	if err != nil {
		return RolloutHeader{}, 0, err
	}
	defer func() { err = errors.Join(err, closeReader()) }()
	counter := &countingReader{reader: reader}
	source := io.Reader(counter)
	if budget != nil {
		source = &discoveryBudgetReader{reader: counter, budget: budget, remainingSize: size}
	}
	lines := bufio.NewReaderSize(io.LimitReader(source, maxSessionMetadataBytes), headerReadChunk)
	for record := 0; ; record++ {
		if err := ctx.Err(); err != nil {
			return RolloutHeader{}, counter.bytes, err
		}
		line, readErr := lines.ReadBytes('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return RolloutHeader{}, counter.bytes, readErr
		}
		if trimmed := bytes.TrimSpace(line); len(trimmed) > 0 {
			var parsed rolloutRecord
			if json.Unmarshal(trimmed, &parsed) == nil && parsed.Type == "session_meta" {
				if metadata, decodeErr := decodeSessionMetadata(parsed.Payload); decodeErr == nil {
					return headerFromMetadata(metadata, record == 0), counter.bytes, nil
				}
			}
		}
		if readErr != nil {
			if counter.bytes >= maxSessionMetadataBytes && (size < 0 || size > maxSessionMetadataBytes) {
				return RolloutHeader{}, counter.bytes, errHeaderBound
			}
			if record == 0 && len(bytes.TrimSpace(line)) == 0 {
				return RolloutHeader{}, counter.bytes, errors.New("session metadata has not flushed")
			}
			return RolloutHeader{}, counter.bytes, errors.New("rollout has no readable session metadata")
		}
	}
}

func headerFromMetadata(metadata sessionMetadata, first bool) RolloutHeader {
	header := RolloutHeader{ID: metadata.ID, ParentID: metadata.ParentID, WorkingDirectory: metadata.CWD, First: first}
	if created, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(metadata.Timestamp)); err == nil {
		header.CreatedAt = created.UTC()
	}
	var nested struct {
		Subagent struct {
			ThreadSpawn struct {
				ParentID string `json:"parent_thread_id"`
			} `json:"thread_spawn"`
		} `json:"subagent"`
	}
	if err := json.Unmarshal(metadata.Source, &nested); err == nil && nested.Subagent.ThreadSpawn.ParentID != "" && metadata.ParentID != "" && nested.Subagent.ThreadSpawn.ParentID != metadata.ParentID {
		header.ConflictingParent = true
	}
	return header
}

// oversizedThreadID identifies a rollout whose first session_meta exceeds the
// structural header bound by scanning its records, as manual discovery does.
// The identity is cached for the pass; only the first scan is charged.
func (cache *HeaderCache) oversizedThreadID(ctx context.Context, path string, budget *DiscoveryBudget) (string, error) {
	key, keyErr := headerKey(path)
	if keyErr != nil {
		return "", keyErr
	}
	if cache != nil {
		if id, ok := cache.oversized[key]; ok {
			return id, nil
		}
	}
	candidate, err := inspectRollout(ctx, path, "", budget)
	if err == nil && cache != nil {
		if after, afterErr := headerKey(path); afterErr == nil && after == key {
			if cache.oversized == nil {
				cache.oversized = make(map[headerCacheKey]string)
			}
			cache.oversized[key] = candidate.threadID
		}
	}
	return candidate.threadID, err
}

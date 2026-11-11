package backfill

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"agenticlens/internal/adapters/codex"
)

var errSourceBudget = errors.New("configured source recovery budget reached; retained owners remain incomplete")

type sourceReadBudget struct {
	sessions  int
	bytes     int64
	codexSeen map[string]bool
}

func (b *sourceReadBudget) beginOwner() error {
	if b == nil || b.sessions < 1 || b.bytes < 1 {
		return errSourceBudget
	}
	b.sessions--
	return nil
}

func (b *sourceReadBudget) parseCodex(path string, cutoff time.Time) (transcript codex.Transcript, err error) {
	if b == nil || b.bytes < 1 {
		return transcript, errSourceBudget
	}
	if !b.codexSeen[path] {
		if err := b.beginOwner(); err != nil {
			return transcript, err
		}
		if b.codexSeen == nil {
			b.codexSeen = make(map[string]bool)
		}
		b.codexSeen[path] = true
	}
	file, err := os.Open(path)
	if err != nil {
		return transcript, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	info, err := file.Stat()
	if err != nil {
		return transcript, err
	}
	if !info.Mode().IsRegular() {
		return transcript, fmt.Errorf("source rollout %q is not a regular file", path)
	}
	compressedSource := strings.HasSuffix(path, ".gz")
	var reader io.Reader = file
	if compressedSource {
		compressed, err := gzip.NewReader(file)
		if err != nil {
			return transcript, err
		}
		defer func() { err = errors.Join(err, compressed.Close()) }()
		reader = compressed
	}
	observed := &sourceEOFReader{reader: reader}
	limit := &io.LimitedReader{R: observed, N: b.bytes}
	remaining := b.bytes
	transcript, err = codex.ParseRolloutAt(limit, cutoff)
	consumed := remaining - limit.N
	b.bytes -= consumed
	if limit.N == 0 && !observed.eof && (compressedSource || info.Size() > remaining) {
		return codex.Transcript{}, errSourceBudget
	}
	return transcript, err
}

func (b *sourceReadBudget) discoveryLimit() int {
	limit := b.sessions
	if limit < int(^uint(0)>>1) {
		limit++
	}
	return limit
}

type sourceEOFReader struct {
	reader io.Reader
	eof    bool
}

func (r *sourceEOFReader) Read(body []byte) (int, error) {
	n, err := r.reader.Read(body)
	r.eof = errors.Is(err, io.EOF)
	return n, err
}

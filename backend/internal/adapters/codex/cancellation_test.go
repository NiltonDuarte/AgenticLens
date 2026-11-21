package codex_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"agenticlens/internal/adapters/codex"
)

type rolloutReadBoundary struct {
	context.Context
	budget *codex.DiscoveryBudget
	cancel context.CancelFunc
}

func (ctx rolloutReadBoundary) Err() error {
	if ctx.budget.SourceBytesRead > 0 {
		ctx.cancel()
	}
	return ctx.Context.Err()
}

const cancellationOwnerRecord = "{\"type\":\"session_meta\",\"payload\":{\"id\":\"owner\"}}\n"

const cancellationCallRecord = "{\"type\":\"response_item\",\"payload\":{\"type\":\"function_call\",\"call_id\":\"call\",\"name\":\"read\",\"arguments\":\"{}\"}}\n"

func TestCodexDiscoveryCancellationAfterNativeRead(t *testing.T) {
	for _, suffix := range []string{".jsonl", ".jsonl.gz"} {
		t.Run(suffix, func(t *testing.T) {
			home := t.TempDir()
			path := filepath.Join(home, "sessions", "owner"+suffix)
			if err := os.Mkdir(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			body := []byte(cancellationOwnerRecord + cancellationCallRecord)
			if suffix == ".jsonl.gz" {
				var compressed bytes.Buffer
				writer := gzip.NewWriter(&compressed)
				_, err := writer.Write(body)
				if err := errors.Join(err, writer.Close()); err != nil {
					t.Fatal(err)
				}
				body = compressed.Bytes()
			}
			if err := os.WriteFile(path, body, 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			budget := &codex.DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: 1 << 20}
			budget.Context = rolloutReadBoundary{ctx, budget, cancel}
			options := codex.DiscoveryOptions{CodexHome: home, SessionID: "owner", Budget: budget}
			paths, err := codex.DiscoverRolloutPaths(options)
			if !errors.Is(err, context.Canceled) || len(paths) != 0 || budget.SourceBytesRead == 0 {
				t.Fatalf("native discovery accepted an interrupted result: paths=%v bytes=%d error=%v", paths, budget.SourceBytesRead, err)
			}
			if seen, err := codex.DiscoverExactRolloutPathContext(ctx, home, "owner"); seen != "" || !errors.Is(err, context.Canceled) {
				t.Fatalf("already cancelled exact selection returned %q, %v", seen, err)
			}
			if transcript, err := codex.ParseRolloutPathContext(ctx, path); transcript.ThreadID != "" || !errors.Is(err, context.Canceled) {
				t.Fatalf("already cancelled native parser returned %q, %v", transcript.ThreadID, err)
			}

			options.Budget = nil
			paths, err = codex.DiscoverRolloutPaths(options)
			if err != nil || !slices.Equal(paths, []string{path}) {
				t.Fatalf("independent discovery failed: paths=%v error=%v", paths, err)
			}
			transcript, err := codex.ParseRolloutPathContext(context.Background(), path)
			if err != nil || transcript.ThreadID != "owner" || len(transcript.Warnings) != 0 || len(transcript.Events) != 1 {
				t.Fatalf("independent parsing failed: owner=%q events=%v warnings=%v error=%v", transcript.ThreadID, transcript.Events, transcript.Warnings, err)
			}
		})
	}
}

type rolloutRecordReader struct {
	records   []string
	readCount int
	cancel    context.CancelFunc
}

func (r *rolloutRecordReader) Read(dest []byte) (int, error) {
	r.readCount++
	if len(r.records) == 0 {
		return 0, io.EOF
	}
	n := copy(dest, r.records[0])
	r.records = r.records[1:]
	if r.readCount == 2 {
		r.cancel()
	}
	return n, nil
}

func TestCodexParserCancellationDuringRead(t *testing.T) {
	records := []string{cancellationOwnerRecord, cancellationCallRecord}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reader := &rolloutRecordReader{records: records, cancel: cancel}
	transcript, err := codex.ParseRolloutContext(ctx, reader)
	if !errors.Is(err, context.Canceled) || reader.readCount != 2 || len(transcript.Events) != 0 {
		t.Fatalf("interrupted record was decoded or empty success returned: reads=%d events=%v error=%v", reader.readCount, transcript.Events, err)
	}
	transcript, err = codex.ParseRolloutContext(context.Background(), strings.NewReader(strings.Join(records, "")))
	if err != nil || transcript.ThreadID != "owner" || len(transcript.Events) != 1 || transcript.Events[0].CallID != "call" {
		t.Fatalf("healthy reader did not parse the retained control record: owner=%q events=%v error=%v", transcript.ThreadID, transcript.Events, err)
	}
}

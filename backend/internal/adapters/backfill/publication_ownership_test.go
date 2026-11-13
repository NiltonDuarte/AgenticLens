package backfill

import (
	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type publicationWaitContext struct {
	context.Context
	waiting chan struct{}
	once    sync.Once
}

func (c *publicationWaitContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.waiting) })
	return c.Context.Done()
}

func TestManualWritersWaitForFormatterPublication(t *testing.T) {
	for _, writer := range []string{"migration", "backfill"} {
		t.Run(writer, func(t *testing.T) {
			root := newBackfillProject(t)
			paths := []string{}
			for _, id := range []string{"A", "B"} {
				path := filepath.Join(root, id, "agentic_metrics.json")
				paths = append(paths, path)
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				body := `{"schema_version":4,"ticket_id":"` + id + `","ticket_title":"fixture","sessions":[],"anomalies":[]}`
				if err := os.WriteFile(path, []byte(body), 0644); err != nil {
					t.Fatal(err)
				}
			}
			owner, err := capture.AcquirePublication(context.Background(), paths[1])
			if err != nil {
				t.Fatal(err)
			}
			held := true
			defer func() {
				if held {
					if err := owner.Close(); err != nil {
						t.Error(err)
					}
				}
			}()
			deadline, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			formatCtx := &publicationWaitContext{Context: deadline, waiting: make(chan struct{})}
			formatDone := make(chan error, 1)
			go func() {
				_, err := adapters.FormatLedgerFiles(formatCtx, adapters.LedgerFormatOptions{Project: root, All: true})
				formatDone <- err
			}()
			select {
			case <-formatCtx.waiting:
			case <-deadline.Done():
				t.Fatal("formatter did not reach B contention after reading A")
			}
			writeCtx := &publicationWaitContext{Context: deadline, waiting: make(chan struct{})}
			writeDone := make(chan error, 1)
			go func() {
				if writer == "migration" {
					res, issues := NewMigrationRunner(nil).migrateWorkspace(writeCtx, paths[0], MigrationOptions{}, domain.CurrentSchemaVersion, domain.ProjectConfig{})
					if len(issues) > 0 {
						writeDone <- errors.New(issues[0])
						return
					}
					if res.Status != "migrated" {
						writeDone <- errors.New("migration did not publish")
						return
					}
					writeDone <- res.Error
				} else {
					res := NewEngine().processSingleWorkspace(writeCtx, paths[0], BackfillOptions{}, nil, "test-user", nil, "2026-01-02T03:04:05Z", domain.ProjectConfig{})
					writeDone <- res.Error
				}
			}()
			select {
			case <-writeCtx.waiting:
			case err := <-writeDone:
				t.Fatalf("manual writer bypassed formatter ownership: %v", err)
			case <-deadline.Done():
				t.Fatal("manual writer never attempted ownership")
			}
			if err := owner.Close(); err != nil {
				t.Fatal(err)
			}
			held = false
			select {
			case err := <-formatDone:
				if err != nil {
					t.Fatal(err)
				}
			case <-deadline.Done():
				t.Fatal("formatter deadlocked")
			}
			select {
			case err := <-writeDone:
				if err != nil {
					t.Fatal(err)
				}
			case <-deadline.Done():
				t.Fatal("manual writer deadlocked")
			}
			data, err := os.ReadFile(paths[0])
			if err != nil {
				t.Fatal(err)
			}
			var facts struct {
				SchemaVersion int    `json:"schema_version"`
				UpdatedAt     string `json:"updated_at"`
			}
			if err := json.Unmarshal(data, &facts); err != nil {
				t.Fatal(err)
			}
			if facts.SchemaVersion != domain.CurrentSchemaVersion {
				t.Fatalf("formatter restored stale schema: %s", data)
			}
			if writer == "backfill" && facts.UpdatedAt != "2026-01-02T03:04:05Z" {
				t.Fatalf("backfill update lost: %s", data)
			}
		})
	}
}

func TestManualWritersPreservePendingPublication(t *testing.T) {
	for _, writer := range []string{"migration", "backfill"} {
		t.Run(writer, func(t *testing.T) {
			path := filepath.Join(newBackfillProject(t), "agentic_metrics.json")
			body := []byte(`{"schema_version":4,"ticket_id":"A","ticket_title":"fixture","sessions":[],"anomalies":[]}`)
			if err := os.WriteFile(path, body, 0644); err != nil {
				t.Fatal(err)
			}
			owner, err := capture.AcquirePublication(context.Background(), path)
			if err != nil {
				t.Fatal(err)
			}
			if err := adapters.BeginPublication(path, []string{path}); err != nil {
				t.Fatal(err)
			}
			if err := owner.Close(); err != nil {
				t.Fatal(err)
			}
			if writer == "migration" {
				res, _ := NewMigrationRunner(nil).migrateWorkspace(context.Background(), path, MigrationOptions{}, domain.CurrentSchemaVersion, domain.ProjectConfig{})
				err = res.Error
			} else {
				res := NewEngine().processSingleWorkspace(context.Background(), path, BackfillOptions{}, nil, "", nil, "", domain.ProjectConfig{})
				err = res.Error
			}
			if err == nil {
				t.Fatal("writer accepted interrupted publication")
			}
			data, err := os.ReadFile(path)
			if err != nil || string(data) != string(body) {
				t.Fatal("writer changed journal-owned baseline")
			}
		})
	}
}

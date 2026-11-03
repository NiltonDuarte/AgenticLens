package capture

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// RequestPublication commits a generation before native worker launch. A worker
// acknowledges only the generation captured before its source/publication pass.
func (s *Store) RequestPublication(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS publication_request (worktree_id INTEGER NOT NULL,singleton INTEGER NOT NULL CHECK(singleton=1), generation INTEGER NOT NULL, acknowledged INTEGER NOT NULL, requested_at INTEGER NOT NULL,PRIMARY KEY(worktree_id,singleton)); INSERT INTO publication_request VALUES(?,1,1,0,?) ON CONFLICT(worktree_id,singleton) DO UPDATE SET generation=generation+1,requested_at=excluded.requested_at`, s.worktreeID, time.Now().UnixMilli())
	return err
}

type PublicationRequest struct {
	Generation   int64
	Acknowledged int64
	RequestedAt  int64
}

func (s *Store) PublicationRequest(ctx context.Context) (PublicationRequest, error) {
	var r PublicationRequest
	var exists int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='publication_request'`).Scan(&exists); err != nil {
		return r, err
	}
	if exists == 0 {
		return r, nil
	}
	err := s.db.QueryRowContext(ctx, `SELECT generation,acknowledged,requested_at FROM publication_request WHERE worktree_id=? AND singleton=1`, s.worktreeID).Scan(&r.Generation, &r.Acknowledged, &r.RequestedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return r, nil
	}
	return r, err
}
func (s *Store) AcknowledgePublication(ctx context.Context, generation int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE publication_request SET acknowledged=? WHERE worktree_id=? AND singleton=1 AND acknowledged<?`, generation, s.worktreeID, generation)
	return err
}

// AcquireWorker never queues a second worker. Kernel ownership releases on crash.
func AcquireWorker(directory string) (*PublicationLock, error) {
	file, err := os.OpenFile(filepath.Join(directory, "worker.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	held, err := tryPublicationLock(file)
	if err != nil {
		return nil, errors.Join(err, file.Close())
	}
	if !held {
		return nil, file.Close()
	}
	return &PublicationLock{file}, nil
}

// LaunchWorker disconnects callback pipes and process group. It never invokes a
// wrapper, model, shell, network client or separately installed service.
func LaunchWorker(binary, project, state string) error {
	cmd := exec.Command(binary, "capture", "worker", "--repo", project, "--state-dir", state)
	cmd.Dir = project
	configureWorkerProcess(cmd)
	logPath := filepath.Join(state, "worker.log")
	flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if info, e := os.Stat(logPath); e == nil && info.Size() > 1024*1024 {
		flags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	}
	log, err := os.OpenFile(logPath, flags, 0600)
	if err != nil {
		return err
	}
	cmd.Stdout = log
	cmd.Stderr = log
	if err = cmd.Start(); err != nil {
		return errors.Join(err, log.Close())
	}
	// Wait reaps children when the host lives longer than a callback.
	go func() {
		if err := cmd.Wait(); err != nil {
			fmt.Fprintln(os.Stderr, "native capture worker:", err)
		}
	}()
	return log.Close()
}

// WorkerProgress is local operational evidence, never source telemetry.
type WorkerProgress struct {
	PID         int    `json:"pid"`
	State       string `json:"state"`
	Heartbeat   int64  `json:"heartbeat"`
	Deadline    int64  `json:"deadline"`
	NextAttempt int64  `json:"next_attempt"`
	Detail      string `json:"detail"`
}

func WriteWorkerProgress(directory string, progress WorkerProgress) error {
	body, err := json.Marshal(progress)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(directory, ".worker-status-")
	if err != nil {
		return err
	}
	name := f.Name()
	defer func() {
		if e := os.Remove(name); e != nil && !errors.Is(e, os.ErrNotExist) {
			fmt.Fprintln(os.Stderr, "worker status cleanup:", e)
		}
	}()
	_, err = f.Write(body)
	if e := errors.Join(err, f.Close()); e != nil {
		return e
	}
	return os.Rename(name, filepath.Join(directory, "worker-status.json"))
}
func WorkerStatus(directory string) (WorkerProgress, error) {
	var p WorkerProgress
	body, err := os.ReadFile(filepath.Join(directory, "worker-status.json"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return p, err
	}
	if len(body) > 0 {
		if err = json.Unmarshal(body, &p); err != nil {
			return p, err
		}
	}
	// Report recorded evidence only. Probing ownership would create/acquire a lock.
	if p.State == "" {
		p.State = "unobserved"
	} else if p.Deadline > 0 && time.Now().UnixMilli() >= p.Deadline {
		p.State = "expired"
	} else if p.State == "running" {
		p.State = "recorded_running"
		p.Detail += " Recorded heartbeat only; live process ownership is unverified by read-only status."
	}

	return p, nil
}

// AutomaticEnabled preserves compatibility with registries created before install
// manifests existed. Once installed, at least one owned integration must be active.
func AutomaticEnabled(directory string) (bool, error) {
	found := false
	for _, source := range []string{"codex", "opencode"} {
		m, err := ReadManifest(directory, source)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return false, err
		}
		found = true
		if m.Active {
			return true, nil
		}
	}
	return !found, nil
}

// PublicationTickets rotates the ticket pass across bounded worker lifetimes.
// Persisting the attempted ticket prevents an expensive early ticket from
// consuming every lifetime before another ticket receives an attempt.
func (s *Store) PublicationTickets(ctx context.Context, tickets []string) ([]string, error) {
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS publication_cursor(worktree_id INTEGER NOT NULL,singleton INTEGER NOT NULL CHECK(singleton=1),ticket TEXT NOT NULL,PRIMARY KEY(worktree_id,singleton))`); err != nil {
		return nil, err
	}
	var cursor string
	err := s.db.QueryRowContext(ctx, `SELECT ticket FROM publication_cursor WHERE worktree_id=? AND singleton=1`, s.worktreeID).Scan(&cursor)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	split := 0
	for split < len(tickets) && tickets[split] <= cursor {
		split++
	}
	return append(append([]string{}, tickets[split:]...), tickets[:split]...), nil
}
func (s *Store) AttemptPublicationTicket(ctx context.Context, ticket string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO publication_cursor VALUES(?,1,?) ON CONFLICT(worktree_id,singleton) DO UPDATE SET ticket=excluded.ticket`, s.worktreeID, ticket)
	return err
}

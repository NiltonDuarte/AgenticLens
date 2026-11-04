package capture

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"time"

	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

const diagnosticSchema = `CREATE TABLE IF NOT EXISTS capture_diagnostics(sequence INTEGER PRIMARY KEY AUTOINCREMENT,worktree_id INTEGER NOT NULL,ticket TEXT NOT NULL,phase TEXT NOT NULL,outcome TEXT NOT NULL,record BLOB NOT NULL);
CREATE TABLE IF NOT EXISTS capture_diagnostic_heads(worktree_id INTEGER NOT NULL,ticket TEXT NOT NULL,phase TEXT NOT NULL,owner TEXT NOT NULL,kind TEXT NOT NULL CHECK(kind IN('latest','committed','failed','accepted','completed','current','generic_failed')),sequence INTEGER NOT NULL,PRIMARY KEY(worktree_id,ticket,phase,owner,kind));
CREATE TABLE IF NOT EXISTS capture_diagnostic_failures(worktree_id INTEGER NOT NULL,sequence INTEGER NOT NULL,PRIMARY KEY(worktree_id,sequence));
CREATE TABLE IF NOT EXISTS capture_diagnostic_index(worktree_id INTEGER NOT NULL,singleton INTEGER NOT NULL CHECK(singleton=1),cursor INTEGER NOT NULL,upper_sequence INTEGER NOT NULL,known_max INTEGER NOT NULL,complete INTEGER NOT NULL,PRIMARY KEY(worktree_id,singleton));
CREATE INDEX IF NOT EXISTS capture_diagnostics_scope_sequence ON capture_diagnostics(worktree_id,sequence);`

// RecordDiagnostic retains current observation and terminal success/failure per
// ticket/phase and exact source owner, plus 32 recent failures. Pending retries
// and unrelated history cannot acknowledge a confirmed terminal failure.

func (s *Store) RecordDiagnostic(ctx context.Context, d domain.CaptureDiagnostic) error {
	if d.ObservedAt == "" {
		d.ObservedAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	body, err := json.Marshal(d)
	if err != nil {
		return err
	}
	tx, err := s.beginDiagnosticWrite(ctx)
	if err != nil {
		return err
	}
	state, err := s.readDiagnosticIndexTx(ctx, tx)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}
	if err = s.detectDiagnosticIndexGap(ctx, tx, &state); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO capture_diagnostics(worktree_id,ticket,phase,outcome,record)VALUES(?,?,?,?,?)`, s.worktreeID, d.TicketID, d.Phase, d.Outcome, body)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}
	d.Sequence, err = result.LastInsertId()
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}
	if err = s.indexDiagnostic(ctx, tx, d); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	state.known = d.Sequence
	state.attemptKnown = d.Sequence
	if err = s.writeDiagnosticIndex(ctx, tx, state); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	if state.complete {
		if err = s.pruneIndexedDiagnostics(ctx, tx); err != nil {
			return errors.Join(err, tx.Rollback())
		}
	}
	return tx.Commit()
}

func (s *Store) Diagnostics(ctx context.Context) ([]domain.CaptureDiagnostic, error) {
	var exists int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='capture_diagnostics'`).Scan(&exists); err != nil {
		return nil, err
	}
	result := []domain.CaptureDiagnostic{}
	if exists == 0 {
		return result, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT sequence,record FROM capture_diagnostics WHERE worktree_id=? ORDER BY sequence`, s.worktreeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var seq int64
		var body []byte
		var d domain.CaptureDiagnostic
		if err = rows.Scan(&seq, &body); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(body, &d); err != nil {
			return nil, err
		}
		d.Sequence = seq
		result = append(result, d)
	}
	return result, rows.Err()
}

// OpenReadOnlyProject never creates an absent registry, initializes schema or
// writes domain facts. Standard SQLite read transactions may maintain ordinary
// WAL/SHM coordination metadata; no AgenticLens lock/status files are created.
func OpenReadOnlyProject(directory, project string) (*Store, error) {
	cfg, err := adaptersconfig.Load(project)
	if err != nil {
		return nil, err
	}
	project, err = ProjectDirectory(project)
	if err != nil {
		return nil, err
	}
	directory, err = namedProjectState(directory, cfg.ProjectName, project)
	if err != nil {
		return nil, err
	}
	return openRegistry(directory, cfg.ProjectName, project, true)
}

func (s *Store) RecordPublication(ctx context.Context, ticket, outcome, destination, summary string, owners []domain.SessionFact, detail string, refreshed ...[]domain.SessionFact) error {
	rows, err := s.List(ctx)
	if err != nil {
		return err
	}
	d := domain.CaptureDiagnostic{TicketID: ticket, Phase: "publication", Outcome: outcome, Destination: destination, Summary: summary, Detail: detail}
	req, err := s.PublicationRequest(ctx)
	if err != nil {
		return err
	}
	d.Generation = req.Generation
	compact, err := domain.PersistAnomalySummaries(domain.AgenticFactLedger{Sessions: owners})
	if err != nil {
		return err
	}
	type ownerKey struct {
		source domain.SourceHarness
		id     string
	}
	candidates := map[ownerKey][]domain.CaptureRegistration{}
	for _, row := range rows {
		if row.TicketID == ticket {
			key := ownerKey{row.Identity.Source, row.Identity.SessionID}
			candidates[key] = append(candidates[key], row)
		}
	}
	refreshIDs := make(map[string]bool)
	for _, snapshot := range refreshed {
		for _, owner := range snapshot {
			refreshIDs[string(owner.Source)+"\x00"+owner.ID] = true
		}
	}
	for _, owner := range compact.Sessions {
		recorded := domain.CapturePublishedOwner{Source: owner.Source, SessionID: owner.ID, Association: "unknown"}
		if len(refreshed) > 0 {
			recorded.RefreshState = "retained"
			if refreshIDs[string(owner.Source)+"\x00"+owner.ID] {
				recorded.RefreshState = "refreshed"
			}
		}
		if checkpoint, known := domain.CaptureOwnerCheckpoint(owner); known {
			recorded.Checkpoint = checkpoint.UTC().Format(time.RFC3339Nano)
		}
		matches := 0
		for _, row := range candidates[ownerKey{owner.Source, owner.ID}] {
			snapshotFacts, err := domain.PersistAnomalySummaries(domain.AgenticFactLedger{Sessions: row.Snapshot})
			if err != nil {
				return err
			}
			for _, snapshot := range snapshotFacts.Sessions {
				equal, err := domain.CaptureOwnerFactsEqual(snapshot, owner)
				if err != nil {
					return err
				}
				if !equal {
					continue
				}
				matches++
				recorded.SourceFingerprint = row.SourceFingerprint
				recorded.Store = row.Identity.Store
				recorded.InstallationID = row.SourceInstallationID
				recorded.Association = "matching_registry_snapshot"
				break
			}
		}
		if matches > 1 {
			recorded.SourceFingerprint = ""
			recorded.Store = ""
			recorded.InstallationID = ""
			recorded.Association = "unknown"
		}
		d.Owners = append(d.Owners, recorded)
	}
	return s.RecordDiagnostic(ctx, d)
}

// InstallationIdentity binds callback evidence to the exact installed manifest.
// It is local operational evidence and never appears in the public summary.
func InstallationIdentity(directory, source string) (string, string, error) {
	body, err := os.ReadFile(manifestPath(directory, source))
	if err != nil {
		return "", "", err
	}
	var identity struct {
		SourceStore string `json:"source_store"`
		Active      bool   `json:"active"`
	}
	if err = json.Unmarshal(body, &identity); err != nil {
		return "", "", err
	}
	if !identity.Active {
		return "", identity.SourceStore, nil
	}
	digest := sha256.Sum256(body)
	return hex.EncodeToString(digest[:]), identity.SourceStore, nil
}

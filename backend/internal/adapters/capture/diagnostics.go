package capture

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"agenticlens/internal/domain"
)

const diagnosticSchema = `CREATE TABLE IF NOT EXISTS capture_diagnostics(sequence INTEGER PRIMARY KEY AUTOINCREMENT,ticket TEXT NOT NULL,phase TEXT NOT NULL,outcome TEXT NOT NULL,record BLOB NOT NULL);
CREATE TABLE IF NOT EXISTS capture_diagnostic_heads(ticket TEXT NOT NULL,phase TEXT NOT NULL,owner TEXT NOT NULL,kind TEXT NOT NULL CHECK(kind IN('latest','committed','failed','accepted','completed','current','generic_failed')),sequence INTEGER NOT NULL,PRIMARY KEY(ticket,phase,owner,kind));
CREATE TABLE IF NOT EXISTS capture_diagnostic_failures(sequence INTEGER PRIMARY KEY);
CREATE TABLE IF NOT EXISTS capture_diagnostic_index(singleton INTEGER PRIMARY KEY CHECK(singleton=1),cursor INTEGER NOT NULL,upper_sequence INTEGER NOT NULL,known_max INTEGER NOT NULL,complete INTEGER NOT NULL);
INSERT OR IGNORE INTO capture_diagnostic_index SELECT 1,0,coalesce(max(sequence),0),coalesce(max(sequence),0),CASE WHEN count(*)=0 THEN 1 ELSE 0 END FROM (SELECT sequence FROM capture_diagnostics ORDER BY sequence DESC LIMIT 1);`

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
	state, err := readDiagnosticIndexTx(ctx, tx)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}
	if err = detectDiagnosticIndexGap(ctx, tx, &state); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO capture_diagnostics(ticket,phase,outcome,record)VALUES(?,?,?,?)`, d.TicketID, d.Phase, d.Outcome, body)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}
	d.Sequence, err = result.LastInsertId()
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}
	if err = indexDiagnostic(ctx, tx, d); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	state.known = d.Sequence
	state.attemptKnown = d.Sequence
	if err = writeDiagnosticIndex(ctx, tx, state); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	if state.complete {
		if err = pruneIndexedDiagnostics(ctx, tx); err != nil {
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
	rows, err := s.db.QueryContext(ctx, `SELECT sequence,record FROM capture_diagnostics ORDER BY sequence`)
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
	p, err := filepath.Abs(filepath.Join(directory, "registry.db"))
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("capture registry is not a regular file")
	}
	project, err = ProjectDirectory(project)
	if err != nil {
		return nil, err
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(p)}
	db, err := sql.Open("sqlite", u.String()+"?mode=ro&_pragma=query_only(1)&_pragma=busy_timeout(250)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	store := &Store{db: db, project: project, directory: directory}
	var version int
	if err = db.QueryRow(`SELECT version FROM capture_meta`).Scan(&version); err != nil {
		return nil, errors.Join(fmt.Errorf("reading capture protocol: %w", err), store.Close())
	}
	if version != ProtocolVersion {
		return nil, errors.Join(fmt.Errorf("unsupported capture registry protocol %d; refresh setup explicitly", version), store.Close())
	}
	var owner string
	if err = db.QueryRow(`SELECT directory FROM project_owner WHERE singleton=1`).Scan(&owner); err != nil {
		return nil, errors.Join(fmt.Errorf("capture project association unavailable: %w", err), store.Close())
	}
	if owner != project {
		return nil, errors.Join(errors.New("capture registry belongs to another project"), store.Close())
	}
	return store, nil
}

func (s *Store) RecordPublication(ctx context.Context, ticket, outcome, destination, summary string, owners []domain.SessionFact, detail string) error {
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
	for _, owner := range compact.Sessions {
		recorded := domain.CapturePublishedOwner{Source: owner.Source, SessionID: owner.ID, Association: "unknown"}
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

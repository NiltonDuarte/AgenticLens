package capture

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const publicationAuxiliarySchema = `CREATE TABLE IF NOT EXISTS capture_auxiliary_publications(worktree_id INTEGER NOT NULL,ticket TEXT NOT NULL,destination TEXT NOT NULL,record BLOB NOT NULL,PRIMARY KEY(worktree_id,ticket,destination));`
const auxiliaryMetadataVersion = 1
const maxAuxiliaryMetadataBytes = 64 << 20
const maxAuxiliaryGenerationBytes = 32 << 20
const auxiliaryRecordQuery = "SELECT CASE WHEN length(record)<=67108864 THEN record ELSE NULL END FROM capture_auxiliary_publications WHERE worktree_id=? AND ticket=? AND destination=?"
const maxAuxiliaryArtifactBytes = 128 << 20

// ErrAuxiliaryArtifactMismatch never grants evidence. A caller may rebind only
// after independently proving facts without any receipt for this artifact.
var ErrAuxiliaryArtifactMismatch = errors.New("ledger does not match a trusted receipt generation; independent evidence or explicit recovery required")

// AuxiliaryGeneration is immutable publication proof, not a source occurrence.
type AuxiliaryGeneration struct {
	ID     string         `json:"id"`
	Digest string         `json:"digest"`
	Facts  AuxiliaryFacts `json:"facts"`
}
type auxiliaryPublication struct {
	Version int                  `json:"version"`
	Active  *AuxiliaryGeneration `json:"active"`
	Pending *AuxiliaryGeneration `json:"pending"`
}

func exactReceiptFields(body []byte, fields ...string) error {
	d := json.NewDecoder(bytes.NewReader(body))
	token, err := d.Token()
	if err != nil {
		return err
	}
	if token != json.Delim('{') {
		return errors.New("receipt metadata must be an object")
	}
	seen := map[string]bool{}
	for d.More() {
		token, err = d.Token()
		if err != nil {
			return err
		}
		key, ok := token.(string)
		if !ok {
			return errors.New("invalid receipt field")
		}
		allowed := false
		for _, f := range fields {
			if key == f {
				allowed = true
			}
		}
		if !allowed || seen[key] {
			return fmt.Errorf("unknown or duplicate receipt field %q", key)
		}
		seen[key] = true
		var raw json.RawMessage
		if err = d.Decode(&raw); err != nil {
			return err
		}
	}
	if _, err = d.Token(); err != nil {
		return err
	}
	if len(seen) != len(fields) {
		return errors.New("incomplete receipt metadata")
	}
	var extra json.RawMessage
	if err = d.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("trailing receipt metadata")
	}
	return nil
}
func (g *AuxiliaryGeneration) UnmarshalJSON(body []byte) error {
	if err := exactReceiptFields(body, "id", "digest", "facts"); err != nil {
		return err
	}
	type wire AuxiliaryGeneration
	var value wire
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	if err := d.Decode(&value); err != nil {
		return err
	}
	var fields struct {
		Facts json.RawMessage `json:"facts"`
	}
	if err := json.Unmarshal(body, &fields); err != nil {
		return err
	}
	if err := exactReceiptFields(fields.Facts, "skills", "task_links"); err != nil {
		return err
	}
	if len(value.ID) != 32 || !hexReceiptToken(value.ID) || len(value.Digest) != 64 || !hexReceiptToken(value.Digest) {
		return errors.New("invalid receipt generation identity or digest")
	}
	*g = AuxiliaryGeneration(value)
	return nil
}
func hexReceiptToken(value string) bool {
	b, err := hex.DecodeString(value)
	return err == nil && len(b) > 0 && value == strings.ToLower(value)
}
func decodeAuxiliaryPublication(body []byte) (auxiliaryPublication, error) {
	var value auxiliaryPublication
	if len(body) > maxAuxiliaryMetadataBytes {
		return value, errors.New("receipt metadata exceeds bound")
	}
	if err := exactReceiptFields(body, "version", "active", "pending"); err != nil {
		return value, err
	}
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	if err := d.Decode(&value); err != nil {
		return value, err
	}
	if value.Version != auxiliaryMetadataVersion {
		return value, fmt.Errorf("unsupported auxiliary receipt metadata version %d", value.Version)
	}
	if value.Active != nil && value.Pending != nil {
		if value.Active.ID == value.Pending.ID {
			return value, errors.New("duplicate receipt generation identity")
		}
		if value.Active.Digest == value.Pending.Digest {
			equal, err := sameAuxiliaryFacts(value.Active.Facts, value.Pending.Facts)
			if err != nil {
				return value, err
			}
			if !equal {
				return value, errors.New("identical artifact digest has conflicting receipt snapshots")
			}
		}
	}
	for _, generation := range []*AuxiliaryGeneration{value.Active, value.Pending} {
		if generation != nil {
			encoded, err := json.Marshal(generation)
			if err != nil {
				return value, err
			}
			if len(encoded) > maxAuxiliaryGenerationBytes {
				return value, errors.New("receipt generation exceeds bound")
			}
		}
	}
	return value, nil
}

func (s *Store) auxiliaryDestination(ctx context.Context, ticket, destination string) (string, error) {
	if s.project == "" || strings.TrimSpace(ticket) == "" || ticket != strings.TrimSpace(ticket) {
		return "", errors.New("receipt requires a verified worktree and exact ticket")
	}
	if err := validateWorktreeOwnership(s.db, s.worktreeID, s.project); err != nil {
		return "", err
	}
	if err := ValidateDestination(s.project, destination); err != nil {
		return "", err
	}
	return ProjectDirectory(destination)
}
func auxiliaryArtifactDigest(destination string) (digest string, present bool, err error) {
	f, err := os.Open(destination)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	info, err := f.Stat()
	if err != nil {
		return "", false, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxAuxiliaryArtifactBytes {
		return "", false, errors.New("receipt ledger is not a bounded regular artifact")
	}
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, maxAuxiliaryArtifactBytes+1))
	if err != nil {
		return "", false, err
	}
	if n > maxAuxiliaryArtifactBytes {
		return "", false, errors.New("receipt ledger exceeds bound")
	}
	return hex.EncodeToString(h.Sum(nil)), true, nil
}
func receiptJournalAbsent(destination string) error {
	if _, err := os.Lstat(destination + ".publication-pending"); err == nil {
		return errors.New("receipt selection requires successful file publication recovery")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
func sameAuxiliaryFacts(a, b AuxiliaryFacts) (bool, error) {
	x, err := json.Marshal(a)
	if err != nil {
		return false, err
	}
	y, err := json.Marshal(b)
	return bytes.Equal(x, y), err
}
func newAuxiliaryGeneration(digest string, facts AuxiliaryFacts) (AuxiliaryGeneration, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return AuxiliaryGeneration{}, err
	}
	return AuxiliaryGeneration{ID: hex.EncodeToString(id[:]), Digest: digest, Facts: facts}, nil
}

// changeAuxiliaryPublication runs under the caller's ledger publication lock and
// an immediate SQLite transaction. Existing generations remain bounded and exact.
func (s *Store) changeAuxiliaryPublication(ctx context.Context, ticket, destination string, change func(auxiliaryPublication, bool) (auxiliaryPublication, error)) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			fmt.Fprintln(os.Stderr, "receipt transaction rollback:", err)
		}
	}()
	var body []byte
	err = tx.QueryRowContext(ctx, auxiliaryRecordQuery, s.worktreeID, ticket, destination).Scan(&body)
	exists := err == nil
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	value := auxiliaryPublication{Version: auxiliaryMetadataVersion}
	if exists {
		value, err = decodeAuxiliaryPublication(body)
		if err != nil {
			return err
		}
	}
	value, err = change(value, exists)
	if err != nil {
		return err
	}
	body, err = json.Marshal(value)
	if err != nil {
		return err
	}
	if _, err = decodeAuxiliaryPublication(body); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO capture_auxiliary_publications(worktree_id,ticket,destination,record)VALUES(?,?,?,?) ON CONFLICT(worktree_id,ticket,destination) DO UPDATE SET record=excluded.record", s.worktreeID, ticket, destination, body); err != nil {
		return err
	}
	return tx.Commit()
}

// SelectPublicationAuxiliary is a write-side recovery operation, called only
// after validated file recovery under the ledger lock. A bound mismatch never
// falls back to the legacy ticket-wide receipt.
func (s *Store) SelectPublicationAuxiliary(ctx context.Context, ticket, destination string) ([]AuxiliaryFacts, bool, error) {
	path, err := s.auxiliaryDestination(ctx, ticket, destination)
	if err != nil {
		return nil, false, err
	}
	if err = receiptJournalAbsent(path); err != nil {
		return nil, false, err
	}
	var exists int
	if err = s.db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE name='capture_auxiliary_publications'").Scan(&exists); err != nil {
		return nil, false, err
	}
	var body []byte
	if exists > 0 {
		err = s.db.QueryRowContext(ctx, auxiliaryRecordQuery, s.worktreeID, ticket, path).Scan(&body)
	} else {
		err = sql.ErrNoRows
	}
	if errors.Is(err, sql.ErrNoRows) {
		legacy, e := s.legacyAuxiliaryEvidence(ctx, ticket)
		return legacy, false, e
	}
	if err != nil {
		return nil, false, err
	}
	digest, present, err := auxiliaryArtifactDigest(path)
	if err != nil {
		return nil, true, err
	}
	var selected []AuxiliaryFacts
	err = s.changeAuxiliaryPublication(ctx, ticket, path, func(state auxiliaryPublication, _ bool) (auxiliaryPublication, error) {
		if !present {
			if state.Active != nil {
				return state, errors.New("accepted receipt ledger is missing; explicit recovery required")
			}
			state.Pending = nil
			return state, nil
		}
		if state.Active != nil && state.Active.Digest == digest {
			selected = []AuxiliaryFacts{state.Active.Facts}
			state.Pending = nil
			return state, nil
		}
		if state.Pending != nil && state.Pending.Digest == digest {
			state.Active = state.Pending
			state.Pending = nil
			selected = []AuxiliaryFacts{state.Active.Facts}
			return state, nil
		}
		return state, ErrAuxiliaryArtifactMismatch
	})
	return selected, true, err
}

// MigratePublicationAuxiliary explicitly establishes a destination association
// now, after the caller independently authorizes the accepted artifact's facts.
func (s *Store) MigratePublicationAuxiliary(ctx context.Context, ticket, destination string, facts AuxiliaryFacts) (AuxiliaryGeneration, error) {
	return s.bindPublicationAuxiliary(ctx, ticket, destination, facts, false)
}

// RebindPublicationAuxiliary is an explicit write-side recovery operation. Its
// caller must independently authorize exact facts from registered or pinned Git
// evidence (never either receipt or the raw ledger) and hold the ledger lock.
// It cannot replace a generation while a file journal or pending proof exists.
func (s *Store) RebindPublicationAuxiliary(ctx context.Context, ticket, destination string, facts AuxiliaryFacts) (AuxiliaryGeneration, error) {
	return s.bindPublicationAuxiliary(ctx, ticket, destination, facts, true)
}

func (s *Store) bindPublicationAuxiliary(ctx context.Context, ticket, destination string, facts AuxiliaryFacts, rebind bool) (AuxiliaryGeneration, error) {
	path, err := s.auxiliaryDestination(ctx, ticket, destination)
	if err != nil {
		return AuxiliaryGeneration{}, err
	}
	if err = receiptJournalAbsent(path); err != nil {
		return AuxiliaryGeneration{}, err
	}
	digest, present, err := auxiliaryArtifactDigest(path)
	if err != nil {
		return AuxiliaryGeneration{}, err
	}
	if !present {
		return AuxiliaryGeneration{}, errors.New("cannot migrate receipt without accepted ledger")
	}
	generation, err := newAuxiliaryGeneration(digest, facts)
	if err != nil {
		return generation, err
	}
	err = s.changeAuxiliaryPublication(ctx, ticket, path, func(state auxiliaryPublication, exists bool) (auxiliaryPublication, error) {
		if exists {
			if state.Active != nil && state.Active.Digest == digest {
				equal, e := sameAuxiliaryFacts(state.Active.Facts, facts)
				if e != nil {
					return state, e
				}
				if equal && state.Pending == nil {
					generation = *state.Active
					return state, nil
				}
				if !equal {
					return state, errors.New("identical artifact digest has different trusted receipt facts")
				}
			}
			if !rebind || state.Pending != nil {
				return state, errors.New("destination already has receipt metadata or unfinished generation; explicit recovery required")
			}
		} else if rebind {
			return state, errors.New("cannot rebind an unassociated receipt destination")
		}
		state.Active = &generation
		return state, nil
	})
	return generation, err
}

// StagePublicationAuxiliary preserves active proof while durably recording the
// exact, already-authorized next snapshot before the existing file commit.
func (s *Store) StagePublicationAuxiliary(ctx context.Context, ticket, destination string, facts AuxiliaryFacts) (AuxiliaryGeneration, error) {
	path, err := s.auxiliaryDestination(ctx, ticket, destination)
	if err != nil {
		return AuxiliaryGeneration{}, err
	}
	digest, present, err := auxiliaryArtifactDigest(path)
	if err != nil {
		return AuxiliaryGeneration{}, err
	}
	if !present {
		return AuxiliaryGeneration{}, errors.New("cannot stage receipt without written ledger")
	}
	generation, err := newAuxiliaryGeneration(digest, facts)
	if err != nil {
		return generation, err
	}
	err = s.changeAuxiliaryPublication(ctx, ticket, path, func(state auxiliaryPublication, _ bool) (auxiliaryPublication, error) {
		if state.Pending != nil && state.Pending.Digest != digest {
			return state, errors.New("unfinished receipt generation requires file recovery")
		}
		for _, prior := range []*AuxiliaryGeneration{state.Active, state.Pending} {
			if prior != nil && prior.Digest == digest {
				equal, e := sameAuxiliaryFacts(prior.Facts, facts)
				if e != nil {
					return state, e
				}
				if !equal {
					return state, errors.New("identical ledger digest has different trusted receipt facts")
				}
				generation = *prior
				return state, nil
			}
		}
		if state.Pending != nil {
			return state, errors.New("unfinished receipt generation requires file recovery")
		}
		state.Pending = &generation
		return state, nil
	})
	return generation, err
}

// FinalizePublicationAuxiliary follows successful file commit; recovery can
// promote the same pending proof if the process stops immediately before this.
func (s *Store) FinalizePublicationAuxiliary(ctx context.Context, ticket, destination, generationID string) error {
	path, err := s.auxiliaryDestination(ctx, ticket, destination)
	if err != nil {
		return err
	}
	if err = receiptJournalAbsent(path); err != nil {
		return err
	}
	digest, present, err := auxiliaryArtifactDigest(path)
	if err != nil {
		return err
	}
	if !present {
		return errors.New("cannot finalize missing receipt artifact")
	}
	return s.changeAuxiliaryPublication(ctx, ticket, path, func(state auxiliaryPublication, exists bool) (auxiliaryPublication, error) {
		if !exists {
			return state, errors.New("receipt generation missing")
		}
		if state.Pending != nil && state.Pending.ID == generationID && state.Pending.Digest == digest {
			state.Active = state.Pending
			state.Pending = nil
			return state, nil
		}
		if state.Active != nil && state.Active.ID == generationID && state.Active.Digest == digest && state.Pending == nil {
			return state, nil
		}
		return state, errors.New("receipt generation does not match committed artifact")
	})
}

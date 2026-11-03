package capture

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	adaptersgit "agenticlens/internal/adapters/git"

	"agenticlens/internal/domain"
)

// RepositoryProof identifies an actual shared Git directory, not copied objects,
// remote URLs or a configured project label. It is private local provenance.
type RepositoryProof struct {
	CommonDirectory   string
	DirectoryIdentity string
}

type HistoryOwner struct {
	Source    domain.SourceHarness
	SessionID string
}

type HistoryEvidence struct {
	Owner      HistoryOwner
	Store      string
	Repository RepositoryProof
}

// RepositoryProof verifies plumbing against the already validated physical
// ownership association. No source store or transcript is opened.
func (s *Store) RepositoryProof(ctx context.Context) (RepositoryProof, error) {
	if err := validateWorktreeOwnership(s.db, s.worktreeID, s.project); err != nil {
		return RepositoryProof{}, err
	}
	commonPath, err := adaptersgit.ResolveHistoryCommonDirectory(ctx, s.project)
	if err != nil {
		return RepositoryProof{}, fmt.Errorf("history repository proof unavailable: %w", err)
	}
	common, err := filepath.EvalSymlinks(commonPath)
	if err != nil {
		return RepositoryProof{}, err
	}
	identity, err := fileIdentity(common)
	if err != nil {
		return RepositoryProof{}, err
	}
	proof := RepositoryProof{common, identity}
	var recorded string
	if err = s.db.QueryRowContext(ctx, "SELECT ownership FROM worktree_ownership WHERE worktree_id=?", s.worktreeID).Scan(&recorded); err != nil {
		return RepositoryProof{}, err
	}
	var owner worktreeOwnership
	if err = json.Unmarshal([]byte(recorded), &owner); err != nil {
		return RepositoryProof{}, err
	}
	if owner.Repository != proof {
		return RepositoryProof{}, errors.New("history repository proof missing or changed; verify live ownership and explicitly recover provenance")
	}
	return proof, nil
}

// AuthorizeCommittedHistory is an exact, bounded read of foreign metadata. The
// caller must separately prove equality to the pinned committed SessionFact.
// It never opens a foreign scope, hydrates a source or changes operational rows.
func (s *Store) AuthorizeCommittedHistory(ctx context.Context, ticket string, candidates []domain.SessionFact, current RepositoryProof) ([]HistoryEvidence, error) {
	if err := s.requireHistoryIndex(ctx); err != nil {
		return nil, err
	}
	actual, err := s.RepositoryProof(ctx)
	if err != nil {
		return nil, err
	}
	if actual != current || current.CommonDirectory == "" || current.DirectoryIdentity == "" {
		return nil, errors.New("history repository proof conflicts with current worktree")
	}
	result := make([]HistoryEvidence, 0, len(candidates))
	for _, fact := range candidates {
		if fact.ID == "" || (fact.Source != domain.SourceCodex && fact.Source != domain.SourceOpenCode) {
			return nil, errors.New("history owner requires exact supported source/session identity")
		}
		var competing int
		if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM (SELECT 1 FROM registrations WHERE session_id=? AND (source<>? OR ticket<>? OR worktree_id=?) LIMIT 1)`, fact.ID, fact.Source, ticket, s.worktreeID).Scan(&competing); err != nil {
			return nil, err
		}
		if competing > 0 {
			return nil, fmt.Errorf("history owner %s has competing source/ticket or current-scope association; explicit recovery required", fact.ID)
		}
		rows, err := s.db.QueryContext(ctx, `SELECT r.worktree_id,r.source,r.store,r.ticket,r.record,coalesce(w.directory,''),coalesce(o.ownership,'') FROM registrations r LEFT JOIN worktrees w ON w.id=r.worktree_id LEFT JOIN worktree_ownership o ON o.worktree_id=r.worktree_id WHERE r.session_id=? AND r.source=? AND r.ticket=? LIMIT 129`, fact.ID, fact.Source, ticket)
		if err != nil {
			return nil, err
		}
		var evidence *HistoryEvidence
		count := 0
		for rows.Next() {
			count++
			var scope int64
			var source, store, boundTicket, directory, ownership string
			var body []byte
			if err = rows.Scan(&scope, &source, &store, &boundTicket, &body, &directory, &ownership); err != nil {
				return nil, errors.Join(err, rows.Close())
			}
			fail := func(message string) ([]HistoryEvidence, error) {
				return nil, errors.Join(fmt.Errorf("history owner %s: %s; explicit provenance recovery required", fact.ID, message), rows.Close())
			}
			if count > 128 || len(body) > 32*1024*1024 || len(ownership) > MaxPayloadBytes {
				return fail("provenance lookup exceeds bound")
			}
			if scope == s.worktreeID {
				return fail("current-scope owner cannot be reclassified as history")
			}
			if source != string(fact.Source) || boundTicket != ticket {
				return fail("competing source or ticket association")
			}
			r, e := decode(body)
			if e != nil {
				return nil, errors.Join(e, rows.Close())
			}
			if r.Identity.Source != fact.Source || r.Identity.SessionID != fact.ID || r.Identity.Store != store || !acceptedHistoryAssociation(r, directory, ticket) {
				return fail("missing verified source/project/ticket association")
			}
			var owner worktreeOwnership
			if e = json.Unmarshal([]byte(ownership), &owner); e != nil {
				return nil, errors.Join(e, rows.Close())
			}
			if owner.DirectoryIdentity == "" || owner.GitDirectory == "" || owner.GitIdentity == "" || owner.Repository != current {
				return fail("missing or competing actual repository proof")
			}
			canonical, e := CanonicalStore(store)
			if e != nil {
				return nil, errors.Join(e, rows.Close())
			}
			if evidence != nil && evidence.Store != canonical {
				return fail("ambiguous source store")
			}
			next := HistoryEvidence{HistoryOwner{fact.Source, fact.ID}, canonical, current}
			evidence = &next
		}
		if err = errors.Join(rows.Err(), rows.Close()); err != nil {
			return nil, err
		}
		if evidence == nil {
			return nil, fmt.Errorf("history owner %s lacks verified provenance; explicit recovery required", fact.ID)
		}
		result = append(result, *evidence)
	}
	return result, nil
}

func verifiedHistoryBinding(binding string) bool {
	return binding == "explicit_agent_response" || binding == "explicit_manual_selection" || binding == "source_parent"
}

func acceptedHistoryAssociation(r domain.CaptureRegistration, directory, ticket string) bool {
	// ProjectDirectory is the accepted association recorded by verified binding
	// and source producers. ProjectProof describes the latest optional metadata
	// observation and can become unverified without erasing that association.
	return directory != "" && r.ProjectDirectory == directory && r.TicketID == ticket &&
		(r.ProjectProof == domain.CaptureProjectVerified || r.ProjectProof == domain.CaptureProjectUnverified) &&
		verifiedHistoryBinding(r.Binding) && (r.Binding != "source_parent" || r.ParentID != "")
}

// RejectForeignSessionCollisions checks identities before fresh selection can
// create a local binding. A schema5 owner has no store discriminator, so an
// existing foreign association cannot silently become a second fresh owner.
func (s *Store) RejectForeignSessionCollisions(ctx context.Context, owners []domain.SessionFact) error {
	if len(owners) == 0 {
		return nil
	}
	if err := s.requireHistoryIndex(ctx); err != nil {
		return err
	}
	for _, owner := range owners {
		var count int
		if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM (SELECT 1 FROM registrations WHERE session_id=? AND worktree_id<>? LIMIT 1)", owner.ID, s.worktreeID).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("fresh/current session %s collides with a foreign source/store association; explicit ownership recovery required", owner.ID)
		}
	}
	return nil
}

func (s *Store) requireHistoryIndex(ctx context.Context) error {
	var exists int
	if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='index' AND name='registrations_history_owner'").Scan(&exists); err != nil {
		return err
	}
	if exists != 1 {
		return errors.New("history provenance index unavailable; open the verified current scope to upgrade metadata before read-only history authorization")
	}
	return nil
}

// AuxiliaryFacts contains owner-derived projections plus unresolved legacy rows.
// Receipts retain exact source/session provenance and occurrence multiplicity;
// accepted publication evidence is never inferred from role/time/child text.
type AuxiliaryFacts struct {
	Skills    []domain.SkillFact    `json:"skills"`
	TaskLinks []domain.TaskLinkFact `json:"task_links"`
}

const auxiliaryReceiptSchema = `CREATE TABLE IF NOT EXISTS capture_auxiliary_receipts(worktree_id INTEGER NOT NULL,ticket TEXT NOT NULL,record BLOB NOT NULL,PRIMARY KEY(worktree_id,ticket));`

func (s *Store) RecordAuxiliaryReceipt(ctx context.Context, ticket string, facts AuxiliaryFacts) error {
	body, err := json.Marshal(facts)
	if err != nil {
		return err
	}
	if len(body) > 32*1024*1024 {
		return errors.New("publication auxiliary receipt exceeds bound")
	}
	if _, err = s.db.ExecContext(ctx, auxiliaryReceiptSchema); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO capture_auxiliary_receipts(worktree_id,ticket,record)VALUES(?,?,?) ON CONFLICT(worktree_id,ticket) DO UPDATE SET record=excluded.record`, s.worktreeID, ticket, body)
	return err
}

func (s *Store) AuxiliaryEvidence(ctx context.Context, ticket string) ([]AuxiliaryFacts, error) {
	registered, err := s.RegisteredAuxiliaryEvidence(ctx, ticket)
	if err != nil {
		return nil, err
	}
	legacy, err := s.legacyAuxiliaryEvidence(ctx, ticket)
	return append(registered, legacy...), err
}

func (s *Store) RegisteredAuxiliaryEvidence(ctx context.Context, ticket string) ([]AuxiliaryFacts, error) {
	rows, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	result := []AuxiliaryFacts{}
	accepted := AuxiliaryFacts{}
	for _, row := range rows {
		if acceptedHistoryAssociation(row, s.project, ticket) {
			ledger := domain.AgenticFactLedger{Sessions: row.Snapshot, Skills: row.Skills, TaskLinks: row.TaskLinks}
			accepted.Skills = append(accepted.Skills, ledger.SkillFacts()...)
			accepted.TaskLinks = append(accepted.TaskLinks, ledger.TaskLinkFacts()...)
		}
	}
	result = append(result, accepted)
	return result, nil
}

func (s *Store) legacyAuxiliaryEvidence(ctx context.Context, ticket string) ([]AuxiliaryFacts, error) {
	var err error
	var exists int
	if err = s.db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE name='capture_auxiliary_receipts'").Scan(&exists); err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, nil
	}
	var body []byte
	err = s.db.QueryRowContext(ctx, "SELECT record FROM capture_auxiliary_receipts WHERE worktree_id=? AND ticket=? AND length(record)<=?", s.worktreeID, ticket, 32*1024*1024).Scan(&body)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var receipt AuxiliaryFacts
	if err = json.Unmarshal(body, &receipt); err != nil {
		return nil, err
	}
	return []AuxiliaryFacts{receipt}, nil
}

package capture

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"agenticlens/internal/domain"
)

type DiagnosticOwner struct {
	TicketID string
	Identity domain.CaptureIdentity
}

type DiagnosticReadOptions struct {
	ExactOwners []DiagnosticOwner
	Tickets     []string
	Owners      []domain.CaptureIdentity
	Current     bool
	Before      int64
	Limit       int
	MaxBytes    int64
}

type DiagnosticPage struct {
	MetadataComplete bool
	HeadLookups      int
	Available        bool
	Records          []domain.CaptureDiagnostic
	More             bool
	NextBefore       int64
	Bytes            int64
}

const diagnosticHeadQuery = `SELECT kind,sequence FROM capture_diagnostic_heads WHERE ticket=? AND phase=? AND owner=? ORDER BY kind LIMIT 7`

// diagnosticMetadataComplete uses only schema metadata and primary-key extrema.
// Legacy/direct writer gaps never certify stale heads as current evidence.
func (s *Store) diagnosticMetadataComplete(ctx context.Context) (bool, bool, error) {
	var tables int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name IN('capture_diagnostic_heads','capture_diagnostic_index')`).Scan(&tables); err != nil {
		return false, false, err
	}
	if tables != 2 {
		return false, false, nil
	}
	var complete int
	var known, actual, attemptKnown int64
	err := s.db.QueryRowContext(ctx, `SELECT complete,known_max,coalesce((SELECT sequence FROM capture_diagnostic_heads WHERE ticket='' AND phase='index' AND owner=? AND kind='latest'),-1),(SELECT coalesce(max(sequence),0) FROM capture_diagnostics) FROM capture_diagnostic_index WHERE singleton=1`, attemptIndexOwner).Scan(&complete, &known, &attemptKnown, &actual)
	if errors.Is(err, sql.ErrNoRows) {
		return true, false, nil
	}
	return true, complete == 1 && known == actual && attemptKnown == actual, err
}

// ReadDiagnostics bounds indexed head lookups before loading any JSON payload.
// Legacy histories remain readable without deriving metadata during inspection.
func (s *Store) ReadDiagnostics(ctx context.Context, options DiagnosticReadOptions) (page DiagnosticPage, returnErr error) {
	page.Records = []domain.CaptureDiagnostic{}
	if options.Limit < 1 || options.MaxBytes < 1 || options.Before < 0 || options.Limit == int(^uint(0)>>1) {
		return page, errors.New("diagnostic inspection requires bounded positive rows/bytes and a nonnegative cursor")
	}
	var exists int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='capture_diagnostics'`).Scan(&exists); err != nil {
		return page, err
	}
	if exists == 0 {
		return page, nil
	}
	page.Available = true
	type metadata struct{ sequence, bytes int64 }
	records := []metadata{}
	var err error
	if options.Current {
		available, complete, e := s.diagnosticMetadataComplete(ctx)
		if e != nil {
			return page, e
		}
		page.MetadataComplete = complete
		page.More = !complete
		if !available {
			return page, nil
		}
		selected := map[int64]bool{}
		seen := map[string]bool{}
		maxLookups := options.Limit
		lookup := func(ticket, phase, owner string) error {
			key := ticket + "\x00" + phase + "\x00" + owner
			if seen[key] {
				return nil
			}
			seen[key] = true
			if page.HeadLookups >= maxLookups {
				page.More = true
				return nil
			}
			page.HeadLookups++
			rows, e := s.db.QueryContext(ctx, diagnosticHeadQuery, ticket, phase, owner)
			if e != nil {
				return e
			}
			for rows.Next() {
				var kind string
				var sequence int64
				if e = rows.Scan(&kind, &sequence); e != nil {
					return errors.Join(e, rows.Close())
				}
				selected[sequence] = true
			}
			return errors.Join(rows.Err(), rows.Close())
		}
		for _, phase := range []string{"publication", "reconciliation", "worker"} {
			if err = lookup("", phase, ""); err != nil {
				return page, err
			}
		}
		for i, ticket := range options.Tickets {
			if i >= maxLookups {
				page.More = true
				break
			}
			for _, phase := range []string{"publication", "reconciliation", "worker"} {
				if err = lookup(ticket, phase, ""); err != nil {
					return page, err
				}
			}
		}
		for i, ticket := range options.Tickets {
			if i >= maxLookups {
				page.More = true
				break
			}
			if err = lookup(ticket, "reconciliation", CollectionAttemptOwner); err != nil {
				return page, err
			}
		}
		for i, owner := range options.ExactOwners {
			if i >= maxLookups {
				page.More = true
				break
			}
			if err = lookup(owner.TicketID, "reconciliation", IdentityKey(owner.Identity)); err != nil {
				return page, err
			}
		}
		// Compatibility for callers which supplied identities and ticket selections
		// separately. The lookup budget also bounds this Cartesian selection.
		for i, owner := range options.Owners {
			if i >= maxLookups {
				page.More = true
				break
			}
			for _, ticket := range options.Tickets {
				if page.HeadLookups >= maxLookups {
					page.More = true
					break
				}
				if err = lookup(ticket, "reconciliation", IdentityKey(owner)); err != nil {
					return page, err
				}
			}
		}
		sequences := make([]int64, 0, len(selected))
		for sequence := range selected {
			sequences = append(sequences, sequence)
		}
		sort.Slice(sequences, func(i, j int) bool { return sequences[i] > sequences[j] })
		if len(sequences) > options.Limit+1 {
			sequences = sequences[:options.Limit+1]
			page.More = true
		}
		for _, sequence := range sequences {
			var size int64
			if err = s.db.QueryRowContext(ctx, `SELECT length(record) FROM capture_diagnostics WHERE sequence=?`, sequence).Scan(&size); errors.Is(err, sql.ErrNoRows) {
				page.More = true
				page.MetadataComplete = false
				continue
			} else if err != nil {
				return page, err
			}
			records = append(records, metadata{sequence, size})
		}
	} else {
		page.NextBefore = options.Before
		rows, e := s.db.QueryContext(ctx, `SELECT sequence,length(record) FROM capture_diagnostics WHERE sequence<? ORDER BY sequence DESC LIMIT ?`, func() int64 {
			if options.Before == 0 {
				return int64(^uint64(0) >> 1)
			}
			return options.Before
		}(), options.Limit+1)
		if e != nil {
			return page, e
		}
		for rows.Next() {
			var item metadata
			if err = rows.Scan(&item.sequence, &item.bytes); err != nil {
				return page, errors.Join(err, rows.Close())
			}
			records = append(records, item)
		}
		if err = errors.Join(rows.Err(), rows.Close()); err != nil {
			return page, err
		}
	}
	if len(records) > options.Limit {
		records = records[:options.Limit]
		page.More = true
	}
	for _, item := range records {
		if item.bytes > options.MaxBytes-page.Bytes {
			page.More = true
			break
		}
		var body []byte
		if err = s.db.QueryRowContext(ctx, `SELECT record FROM capture_diagnostics WHERE sequence=? AND length(record)<=?`, item.sequence, options.MaxBytes-page.Bytes).Scan(&body); errors.Is(err, sql.ErrNoRows) {
			page.More = true
			break
		} else if err != nil {
			return page, err
		}
		var diagnostic domain.CaptureDiagnostic
		if err = json.Unmarshal(body, &diagnostic); err != nil {
			return page, fmt.Errorf("reading retained capture diagnostic: %w", err)
		}
		diagnostic.Sequence = item.sequence
		page.Records = append(page.Records, diagnostic)
		page.Bytes += int64(len(body))
		page.NextBefore = item.sequence
	}
	return page, nil
}

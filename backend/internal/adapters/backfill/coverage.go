package backfill

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

type OwnerRefreshStatus string

const (
	OwnerRefreshed          OwnerRefreshStatus = "refreshed"
	OwnerVerifiedUnchanged  OwnerRefreshStatus = "verified unchanged"
	OwnerVerifiedUncompared OwnerRefreshStatus = "verified read, comparison unavailable"
	OwnerRetained           OwnerRefreshStatus = "retained"
	OwnerDiscovered         OwnerRefreshStatus = "discovered"
	OwnerProposed           OwnerRefreshStatus = "verified change not committed"
	OwnerProposedDiscovered OwnerRefreshStatus = "discovered not committed"
)

type CollectionRefreshStatus string

const (
	CollectionVerified    CollectionRefreshStatus = "verified"
	CollectionEmpty       CollectionRefreshStatus = "verified empty"
	CollectionRetained    CollectionRefreshStatus = "retained"
	CollectionUnavailable CollectionRefreshStatus = "unavailable"
)

type HarnessRefreshStatus string

const (
	HarnessComplete    HarnessRefreshStatus = "complete"
	HarnessIncomplete  HarnessRefreshStatus = "incomplete"
	HarnessNoOwners    HarnessRefreshStatus = "no recorded owners"
	HarnessUnavailable HarnessRefreshStatus = "unavailable"
	HarnessBlocked     HarnessRefreshStatus = "blocked"
)

type OwnerRefreshCoverage struct {
	SessionID string                  `json:"session_id"`
	Status    OwnerRefreshStatus      `json:"status"`
	Reason    string                  `json:"reason,omitempty"`
	Skills    CollectionRefreshStatus `json:"skills"`
	Tasks     CollectionRefreshStatus `json:"tasks"`
}

type RefreshCounts struct {
	Recorded           int `json:"recorded"`
	Refreshed          int `json:"refreshed"`
	VerifiedUnchanged  int `json:"verified_unchanged"`
	VerifiedUncompared int `json:"verified_uncompared,omitempty"`
	Retained           int `json:"retained"`
	Discovered         int `json:"discovered"`
	Proposed           int `json:"proposed"`
	ProposedDiscovered int `json:"proposed_discovered"`
}

type HarnessRefreshTotals struct {
	Harness                 domain.SourceHarness `json:"harness"`
	UnknownOwnerInventories int                  `json:"unknown_owner_inventories,omitempty"`
	RefreshCounts
}

type HarnessRefreshCoverage struct {
	Harness               domain.SourceHarness   `json:"harness"`
	Status                HarnessRefreshStatus   `json:"status"`
	Reason                string                 `json:"reason,omitempty"`
	Owners                []OwnerRefreshCoverage `json:"owners,omitempty"`
	BudgetLimited         bool                   `json:"budget_limited,omitempty"`
	OwnerInventoryUnknown bool                   `json:"owner_inventory_unknown,omitempty"`
	InventoryError        string                 `json:"inventory_error,omitempty"`
	RefreshCounts
}

// Collection presence must be observed before source recovery fills in any
// historical facts that the native reader could not verify.
func collectionCoverage[T domain.SkillFact | domain.DelegationFact](fresh []T, prior []T) CollectionRefreshStatus {
	if fresh == nil {
		if prior != nil {
			return CollectionRetained
		}
		return CollectionUnavailable
	}
	if len(fresh) == 0 {
		return CollectionEmpty
	}
	return CollectionVerified
}

func selectedHarnesses(source string) []domain.SourceHarness {
	switch source {
	case "opencode":
		return []domain.SourceHarness{domain.SourceOpenCode}
	case "codex":
		return []domain.SourceHarness{domain.SourceCodex}
	case "both", "all":
		return []domain.SourceHarness{domain.SourceOpenCode, domain.SourceCodex}
	default:
		return nil
	}
}

func (r *sourceRecovery) accept(fresh, prior domain.SessionFact) {
	r.reads[fresh.ID] = OwnerRefreshCoverage{
		SessionID: fresh.ID,
		Skills:    collectionCoverage(fresh.Skills, prior.Skills),
		Tasks:     collectionCoverage(fresh.Delegations, prior.Delegations),
	}
}

var errOwnerUnavailable = errors.New("exact source session unavailable")

// A nil changed map means the persisted comparison could not run. An empty
// map means the comparison established that the candidate was equivalent.
func recoveredCoverage(harness domain.SourceHarness, stored []domain.SessionFact, recovered sourceRecovery, changed map[string]bool) HarnessRefreshCoverage {
	report := HarnessRefreshCoverage{Harness: harness, Status: HarnessComplete}
	if !recovered.complete {
		report.Status = HarnessIncomplete
	}
	if len(stored) == 0 && recovered.failure == nil {
		report.Status = HarnessNoOwners
	}
	if recovered.failure != nil {
		report.Reason = recovered.failure.Error()
	}
	report.BudgetLimited = isSourceBudgetError(recovered.failure)
	for _, failure := range recovered.failures {
		report.BudgetLimited = report.BudgetLimited || isSourceBudgetError(failure)
	}
	report.Recorded = len(stored)
	recorded := make(map[string]bool, len(stored))
	for _, prior := range stored {
		recorded[prior.ID] = true
		if owner, ok := recovered.reads[prior.ID]; ok {
			switch {
			case changed == nil:
				owner.Status = OwnerVerifiedUncompared
				report.VerifiedUncompared++
			case changed[prior.ID]:
				owner.Status = OwnerRefreshed
				report.Refreshed++
			default:
				owner.Status = OwnerVerifiedUnchanged
				report.VerifiedUnchanged++
			}
			if owner.Skills == CollectionRetained || owner.Tasks == CollectionRetained {
				report.Status = HarnessIncomplete
			}
			report.Owners = append(report.Owners, owner)
			continue
		}
		reason := recovered.failures[prior.ID]
		if reason == nil {
			reason = recovered.failure
		}
		if reason == nil {
			reason = errOwnerUnavailable
		}
		report.Retained++
		report.Owners = append(report.Owners, OwnerRefreshCoverage{
			SessionID: prior.ID, Status: OwnerRetained, Reason: reason.Error(),
			Skills: collectionCoverage(nil, prior.Skills),
			Tasks:  collectionCoverage(nil, prior.Delegations),
		})
	}
	for _, owner := range recovered.reads {
		if recorded[owner.SessionID] {
			continue
		}
		switch {
		case changed == nil:
			owner.Status = OwnerProposedDiscovered
			report.ProposedDiscovered++
		case changed[owner.SessionID]:
			owner.Status = OwnerDiscovered
			report.Discovered++
		default:
			continue
		}
		report.Owners = append(report.Owners, owner)
	}
	slices.SortFunc(report.Owners, func(a, b OwnerRefreshCoverage) int {
		return cmp.Compare(a.SessionID, b.SessionID)
	})
	return report
}

func sourceCoverage(source string, storedOpenCode, storedCodex []domain.SessionFact, openCode, codex sourceRecovery, changed map[string]bool) []HarnessRefreshCoverage {
	var coverage []HarnessRefreshCoverage
	for _, harness := range selectedHarnesses(source) {
		if harness == domain.SourceOpenCode {
			coverage = append(coverage, recoveredCoverage(harness, storedOpenCode, openCode, changed))
		} else {
			coverage = append(coverage, recoveredCoverage(harness, storedCodex, codex, changed))
		}
	}
	return coverage
}

func blockedCoverage(source string, err error) []HarnessRefreshCoverage {
	var coverage []HarnessRefreshCoverage
	for _, harness := range selectedHarnesses(source) {
		coverage = append(coverage, HarnessRefreshCoverage{Harness: harness, Status: HarnessBlocked, Reason: err.Error(), OwnerInventoryUnknown: true})
	}
	return coverage
}

func decodeCoverageSessions(data []byte) ([]adapters.SessionFactDTO, error) {
	var ledger struct {
		Sessions []adapters.SessionFactDTO `json:"sessions"`
	}
	if err := json.Unmarshal(data, &ledger); err != nil {
		return nil, fmt.Errorf("decoding persisted sessions for refresh coverage: %w", err)
	}
	return ledger.Sessions, nil
}

func compareCoverageSessions(migrated []adapters.SessionFactDTO, accepted, candidate []byte, currentSchema bool) (map[string]bool, error) {
	baseline := migrated
	if currentSchema {
		var err error
		baseline, err = decodeCoverageSessions(accepted)
		if err != nil {
			return nil, err
		}
	}
	sessions, err := decodeCoverageSessions(candidate)
	if err != nil {
		return nil, err
	}
	prior := make(map[string]adapters.SessionFactDTO, len(baseline))
	for _, owner := range baseline {
		prior[owner.ID] = owner
	}
	changed := make(map[string]bool, len(sessions))
	for _, owner := range sessions {
		if old, recorded := prior[owner.ID]; !recorded || !reflect.DeepEqual(old, owner) {
			changed[owner.ID] = true
		}
	}
	return changed, nil
}

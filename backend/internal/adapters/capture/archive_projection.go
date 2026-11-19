package capture

import (
	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/domain"
)

// archiveReadGroup keeps the native inputs and prior facts until reconciliation
// takes its final registry view. Nothing here is durable registry evidence.
type archiveReadGroup struct {
	owners      map[string]domain.CaptureRegistration
	transcripts []codex.Transcript
	previous    []domain.SessionFact
	policy      domain.GenerationPolicy
}

func (result *ReconcileResult) recordArchiveReadGroup(registrations []domain.CaptureRegistration, refreshed map[string]bool, transcripts []codex.Transcript, previous []domain.SessionFact, policy domain.GenerationPolicy) {
	owners := make(map[string]domain.CaptureRegistration, len(transcripts))
	for _, row := range registrations {
		if row.Identity.Source == domain.SourceCodex && refreshed[IdentityKey(row.Identity)] {
			owners[row.Identity.SessionID] = row
		}
	}
	result.archiveReadGroups = append(result.archiveReadGroups, archiveReadGroup{
		owners: owners, transcripts: transcripts, previous: previous, policy: policy,
	})
}

// reprojectArchiveFamilies corrects request-local evidence when a recovered
// owner is superseded between annotation commit and the final registry read.
// Regrouping distinguishes an inferred child link from a link independently
// proved by a healthy parent's own rollout. Registry snapshots stay untouched.
func (result *ReconcileResult) reprojectArchiveFamilies(registrations []domain.CaptureRegistration, current map[string]domain.CaptureRegistration, refreshed, stale map[string]bool) error {
	if len(stale) == 0 || len(result.archiveReadGroups) == 0 {
		return nil
	}
	indexes := make(map[string]int, len(registrations))
	for i, row := range registrations {
		indexes[IdentityKey(row.Identity)] = i
	}
	for _, group := range result.archiveReadGroups {
		if !group.hasStaleOwner(stale) {
			continue
		}
		for ownerID, observed := range group.owners {
			if stale[ownerID] {
				continue
			}
			key := IdentityKey(observed.Identity)
			latest, exists := current[key]
			if exists && sameCaptureReadObservation(latest, observed) {
				registrations[indexes[key]] = latest
				continue
			}
			delete(refreshed, key)
			stale[ownerID] = true
			if exists {
				registrations[indexes[key]] = pendingChangedObservation(latest)
			}
		}
		if err := result.regroupArchiveFamily(group, registrations, indexes, stale); err != nil {
			return err
		}
	}
	return nil
}

func (group archiveReadGroup) hasStaleOwner(stale map[string]bool) bool {
	for ownerID := range group.owners {
		if stale[ownerID] {
			return true
		}
	}
	return false
}

func (result *ReconcileResult) regroupArchiveFamily(group archiveReadGroup, registrations []domain.CaptureRegistration, indexes map[string]int, stale map[string]bool) error {
	transcripts := make([]codex.Transcript, 0, len(group.transcripts))
	previous := append([]domain.SessionFact(nil), group.previous...)
	for _, transcript := range group.transcripts {
		id := transcript.ThreadID
		if id == "" {
			id = transcript.SessionID
		}
		if stale[id] {
			if transcript.ParentThread != "" {
				// Even an owner with no prior snapshot makes its observed
				// parent's group incomplete once this read is discarded.
				parent := transcript.ParentThread
				previous = append(previous, domain.SessionFact{ID: id, ParentID: &parent})
			}
			continue
		}
		transcripts = append(transcripts, transcript)
	}
	if len(transcripts) == 0 {
		return nil
	}
	combined, err := codex.CombineTranscriptsWithPolicy(transcripts, group.policy)
	if err != nil {
		return err
	}
	owners, err := combined.SessionFacts(group.policy)
	if err != nil {
		return err
	}
	result.SourceEvents = domain.MergeSourceEvents(result.SourceEvents, domain.SourceEvents{
		Events: combined.Events, CheckpointEvidence: combined.CheckpointEvidence, CompleteSessionIDs: transcriptSessionIDs(transcripts),
	})

	type occurrenceProjection struct {
		row   int
		index int
	}
	positions := make([]occurrenceProjection, 0, len(owners))
	fresh := make([]domain.SessionFact, 0, len(owners))
	for _, owner := range owners {
		observed, ok := group.owners[owner.ID]
		if !ok {
			continue
		}
		row, ok := indexes[IdentityKey(observed.Identity)]
		if !ok {
			continue
		}
		for i, accepted := range registrations[row].Snapshot {
			if accepted.ID == owner.ID {
				accepted.Delegations = owner.Delegations
				fresh = append(fresh, accepted)
				positions = append(positions, occurrenceProjection{row: row, index: i})
			}
		}
	}
	for i, owner := range domain.RetainIncompleteGroupSummaries(previous, fresh) {
		position := positions[i]
		row := &registrations[position.row]
		row.Snapshot = append([]domain.SessionFact(nil), row.Snapshot...)
		row.Snapshot[position.index] = owner
	}
	return nil
}

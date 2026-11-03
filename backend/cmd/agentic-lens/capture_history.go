package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/git"
	"agenticlens/internal/domain"
)

type generationBaseline struct {
	Current            []domain.SessionFact
	History            []domain.SessionFact
	HistoryDTOs        []adapters.SessionFactDTO
	CommittedAuxiliary capture.AuxiliaryFacts
	Evidence           []capture.HistoryEvidence
}

func classifyGenerationBaseline(project, state, destination, ticket string, facts *adapters.ExistingLedgerFacts) (generationBaseline, error) {
	var result generationBaseline
	if facts == nil {
		return result, nil
	}
	if facts.TicketID == nil || *facts.TicketID != ticket {
		return result, errors.New("baseline ticket conflicts with publication")
	}
	s, err := capture.OpenProject(state, project)
	if err != nil {
		return result, err
	}
	defer func() {
		if err := s.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "history registry close:", err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rows, err := s.List(ctx)
	if err != nil {
		return result, err
	}
	local := map[string]bool{}
	for _, row := range rows {
		local[row.Identity.SessionID] = true
	}
	seen := map[string]bool{}
	for _, owner := range facts.Sessions {
		if seen[owner.ID] {
			return result, fmt.Errorf("duplicate baseline session identity %s", owner.ID)
		}
		seen[owner.ID] = true
		if local[owner.ID] {
			result.Current = append(result.Current, owner)
		} else {
			result.History = append(result.History, owner)
		}
	}
	if err = validateGenerationOwners(project, rows, result.Current, ticket); err != nil {
		return result, err
	}
	if err = s.RejectForeignSessionCollisions(ctx, result.Current); err != nil {
		return result, err
	}
	if len(result.History) == 0 && len(facts.Skills) == 0 && len(facts.TaskLinks) == 0 {
		return result, nil
	}
	auxiliary, err := s.AuxiliaryEvidence(ctx, ticket)
	if err != nil {
		return result, err
	}
	// Git is needed only to authorize imported facts, never to hydrate current
	// capture. A missing committed ledger is harmless when scoped evidence covers
	// the entire baseline, including ownerless rows.
	aux := capture.AuxiliaryFacts{Skills: facts.Skills, TaskLinks: facts.TaskLinks}
	auxiliaryErr := validateAuxiliaryFacts(aux, auxiliary)
	needsCommitted := len(result.History) > 0 || auxiliaryErr != nil
	if !needsCommitted {
		return result, nil
	}
	// Every committed fact, including ownerless auxiliary rows, needs actual
	// repository authority. Current scoped receipts above remain Git independent.
	proof, err := s.RepositoryProof(ctx)
	if err != nil {
		return result, err
	}
	committed, err := committedGenerationBaseline(ctx, project, destination)
	if err != nil {
		if len(result.History) == 0 && auxiliaryErr != nil {
			return result, fmt.Errorf("baseline auxiliary provenance requires explicit recovery: %w", errors.Join(auxiliaryErr, fmt.Errorf("committed evidence unavailable: %w", err)))
		}
		return result, fmt.Errorf("committed history unavailable: %w", err)
	}
	if committed.TicketID == nil || *committed.TicketID != ticket {
		return result, errors.New("committed history ticket conflicts with publication")
	}
	committedByID := map[string]domain.SessionFact{}
	for _, owner := range committed.Sessions {
		if _, exists := committedByID[owner.ID]; exists {
			return result, errors.New("committed history contains ambiguous session identities")
		}
		committedByID[owner.ID] = owner
	}
	workingDTOs := make(map[string]adapters.SessionFactDTO, len(facts.PersistedSessions))
	committedDTOs := make(map[string]adapters.SessionFactDTO, len(committed.PersistedSessions))
	for _, owner := range facts.PersistedSessions {
		workingDTOs[owner.ID] = owner
	}
	for _, owner := range committed.PersistedSessions {
		committedDTOs[owner.ID] = owner
	}
	for _, owner := range result.History {
		prior, exists := committedByID[owner.ID]
		if !exists || !reflect.DeepEqual(owner, prior) {
			return result, fmt.Errorf("foreign owner %s differs from the pinned committed fact; explicit recovery required", owner.ID)
		}
	}
	if len(result.History) > 0 {
		result.Evidence, err = s.AuthorizeCommittedHistory(ctx, ticket, result.History, proof)
		if err != nil {
			return result, err
		}
	}
	for _, owner := range result.History {
		workingDTO, workingExists := workingDTOs[owner.ID]
		committedDTO, committedExists := committedDTOs[owner.ID]
		if !workingExists || !committedExists || facts.SchemaVersion == nil || committed.SchemaVersion == nil {
			return result, fmt.Errorf("foreign owner %s persisted evidence differs from pinned committed history", owner.ID)
		}
		// Compare the original semantic facts above, then compare supported
		// publication representations. This accepts our own normalized v5 output
		// against pinned v4 HEAD without accepting a different cost or evidence.
		workingPublication, err := adapters.NormalizeHistorySessionForPublication(workingDTO, *facts.SchemaVersion)
		if err != nil {
			return result, err
		}
		immutable, err := adapters.NormalizeHistorySessionForPublication(committedDTO, *committed.SchemaVersion)
		if err != nil {
			return result, err
		}
		if !reflect.DeepEqual(workingPublication, immutable) {
			return result, fmt.Errorf("foreign owner %s persisted evidence differs from pinned committed history", owner.ID)
		}
		result.HistoryDTOs = append(result.HistoryDTOs, immutable)
	}
	result.CommittedAuxiliary = capture.AuxiliaryFacts{Skills: committed.Skills, TaskLinks: committed.TaskLinks}
	auxiliary = append(auxiliary, result.CommittedAuxiliary)
	if err = validateAuxiliaryFacts(aux, auxiliary); err != nil {
		return result, err
	}
	return result, nil
}

func committedGenerationBaseline(ctx context.Context, project, destination string) (*adapters.ExistingLedgerFacts, error) {
	if err := capture.ValidateDestination(project, destination); err != nil {
		return nil, err
	}
	if info, err := os.Lstat(destination); err == nil {
		if !info.Mode().IsRegular() {
			return nil, errors.New("history ledger destination must be a regular file")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	physical, err := capture.ProjectDirectory(destination)
	if err != nil {
		return nil, err
	}
	relative, err := filepath.Rel(project, physical)
	if err != nil {
		return nil, err
	}
	commit, err := git.ResolveHistoryCommit(ctx, project)
	if err != nil {
		return nil, err
	}
	blob, err := git.ResolveCommittedLedger(ctx, project, commit, filepath.ToSlash(relative))
	if err != nil {
		return nil, err
	}
	return adapters.DecodeExistingFactLedger(blob.Data)
}

func validateAuxiliaryFacts(candidate capture.AuxiliaryFacts, evidence []capture.AuxiliaryFacts) error {
	skills := [][]domain.SkillFact{}
	links := [][]domain.TaskLinkFact{}
	for _, rows := range evidence {
		skills = append(skills, rows.Skills)
		links = append(links, rows.TaskLinks)
	}
	if err := validateAuxiliaryRows(candidate.Skills, skills); err != nil {
		return fmt.Errorf("baseline skills lack committed or current scoped publication evidence: %w", err)
	}
	if err := validateAuxiliaryRows(candidate.TaskLinks, links); err != nil {
		return fmt.Errorf("baseline task links lack committed or current scoped publication evidence: %w", err)
	}
	return nil
}

func validateAuxiliaryRows[T domain.SkillFact | domain.TaskLinkFact](candidate []T, evidence [][]T) error {
	// Each evidence snapshot can repeat the same accepted publication. Use the
	// greatest observed multiplicity per row, rather than counting copies as new
	// permission for extra baseline rows.
	allowed := map[string]int{}
	for _, snapshot := range evidence {
		counts := map[string]int{}
		for _, row := range snapshot {
			body, err := json.Marshal(row)
			if err != nil {
				return err
			}
			counts[string(body)]++
		}
		for key, count := range counts {
			allowed[key] = max(allowed[key], count)
		}
	}
	for _, row := range candidate {
		body, err := json.Marshal(row)
		if err != nil {
			return err
		}
		key := string(body)
		if allowed[key] == 0 {
			return errors.New("unmatched or modified ownerless fact")
		}
		allowed[key]--
	}
	return nil
}

func validateHistorySelection(history, selected []domain.SessionFact) error {
	ids := map[string]bool{}
	for _, owner := range history {
		ids[owner.ID] = true
	}
	for _, owner := range selected {
		if ids[owner.ID] {
			return fmt.Errorf("fresh session %s collides with retained committed history; explicit ownership recovery required", owner.ID)
		}
	}
	return nil
}

func validateFreshGenerationOwners(project, state string, history, selected []domain.SessionFact) error {
	if len(selected) == 0 {
		return nil
	}
	if err := validateHistorySelection(history, selected); err != nil {
		return err
	}
	s, err := capture.OpenReadOnlyProject(state, project)
	if errors.Is(err, os.ErrNotExist) {
		s, err = capture.OpenProject(state, project)
	}
	if err != nil {
		return err
	}
	return errors.Join(s.RejectForeignSessionCollisions(context.Background(), selected), s.Close())
}

func historySummary(baseline generationBaseline) string {
	if len(baseline.History) == 0 {
		return ""
	}
	return fmt.Sprintf("\nRetained committed history: %d session(s), unchanged from the pinned committed ticket snapshot. These sessions were not refreshed and do not establish capture coverage in this worktree.\n", len(baseline.History))
}

func recordGenerationHistory(project, state, ticket string, baseline generationBaseline) error {
	if len(baseline.History) == 0 {
		return nil
	}
	s, err := capture.OpenProject(state, project)
	if err != nil {
		return err
	}
	d := domain.CaptureDiagnostic{TicketID: ticket, Phase: "history", Outcome: "retained", Detail: fmt.Sprintf("Retained %d committed session(s) unchanged after exact fact and same-repository provenance checks. No foreign source hydration, binding, refresh or acknowledgement was performed.", len(baseline.History))}
	return errors.Join(s.RecordDiagnostic(context.Background(), d), s.Close())
}

func recordGenerationAuxiliary(project, state, ticket string, facts capture.AuxiliaryFacts) error {
	if len(facts.Skills) == 0 && len(facts.TaskLinks) == 0 {
		return nil
	}
	s, err := capture.OpenProject(state, project)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return errors.Join(s.RecordAuxiliaryReceipt(ctx, strings.TrimSpace(ticket), facts), s.Close())
}

func ownPublicationAuxiliary(all, committed capture.AuxiliaryFacts) (capture.AuxiliaryFacts, error) {
	skills, err := subtractAuxiliaryRows(all.Skills, committed.Skills)
	if err != nil {
		return capture.AuxiliaryFacts{}, err
	}
	links, err := subtractAuxiliaryRows(all.TaskLinks, committed.TaskLinks)
	return capture.AuxiliaryFacts{Skills: skills, TaskLinks: links}, err
}

func subtractAuxiliaryRows[T domain.SkillFact | domain.TaskLinkFact](all, historical []T) ([]T, error) {
	counts := map[string]int{}
	for _, row := range historical {
		body, err := json.Marshal(row)
		if err != nil {
			return nil, err
		}
		counts[string(body)]++
	}
	result := []T{}
	for _, row := range all {
		body, err := json.Marshal(row)
		if err != nil {
			return nil, err
		}
		if counts[string(body)] > 0 {
			counts[string(body)]--
			continue
		}
		result = append(result, row)
	}
	return result, nil
}

func sessionsWithoutHistory(all, historical []domain.SessionFact) []domain.SessionFact {
	ids := map[string]bool{}
	for _, row := range historical {
		ids[row.ID] = true
	}
	result := []domain.SessionFact{}
	for _, row := range all {
		if !ids[row.ID] {
			result = append(result, row)
		}
	}
	return result
}

package main

import (
	"context"
	"errors"
	"fmt"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

type generationBaseline struct {
	Current     []domain.SessionFact
	History     []domain.SessionFact
	HistoryDTOs []adapters.SessionFactDTO
}

// classifyGenerationBaseline trusts the validated current ledger as portable
// recorded history. Private capture receipts and Git file bytes are not an
// authenticity requirement for preserving or exactly refreshing these facts.
func classifyGenerationBaseline(project, state, destination, ticket string, facts *adapters.ExistingLedgerFacts) (generationBaseline, error) {
	var result generationBaseline
	if facts == nil {
		return result, nil
	}
	if facts.TicketID == nil || *facts.TicketID != ticket {
		return result, errors.New("baseline ticket conflicts with publication")
	}
	if destination != "" {
		cfg, err := adaptersconfig.LoadShared(project)
		if err != nil {
			return result, err
		}
		if err := adapters.ValidateLedgerDestinationScope(cfg, destination); err != nil {
			return result, err
		}
	}
	result.Current = append([]domain.SessionFact(nil), facts.Sessions...)
	result.HistoryDTOs = append([]adapters.SessionFactDTO(nil), facts.PersistedSessions...)
	return result, nil
}

// Only owners not completely refreshed keep their original persisted DTO. This
// preserves sparse baseline evidence without letting it veto source authority.
func retainUnrefreshedGenerationOwners(baseline *generationBaseline, fresh []domain.SessionFact) {
	refreshed := make(map[string]bool)
	for _, owner := range fresh {
		refreshed[owner.ID] = true
	}
	baseline.History = nil
	for _, owner := range baseline.Current {
		if !refreshed[owner.ID] {
			baseline.History = append(baseline.History, owner)
		}
	}
	preserved := []adapters.SessionFactDTO{}
	for _, owner := range baseline.HistoryDTOs {
		if !refreshed[owner.ID] {
			preserved = append(preserved, owner)
		}
	}
	baseline.HistoryDTOs = preserved
}

func historySummary(baseline generationBaseline) string {
	if len(baseline.History) == 0 {
		return ""
	}
	return fmt.Sprintf("\nRetained recorded history: %d session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.\n", len(baseline.History))
}

func recordGenerationHistory(project, state, ticket string, baseline generationBaseline) error {
	if len(baseline.History) == 0 {
		return nil
	}
	s, err := capture.OpenProject(state, project)
	if err != nil {
		return err
	}
	d := domain.CaptureDiagnostic{TicketID: ticket, Phase: "history", Outcome: "retained", Detail: fmt.Sprintf("Retained %d recorded session(s) unchanged outside the complete source refresh set. No foreign source hydration, binding, refresh or acknowledgement was performed.", len(baseline.History))}
	return errors.Join(s.RecordDiagnostic(context.Background(), d), s.Close())
}

// validateRegisteredGenerationBaseline checks positive current-scope bindings
// even when automatic selection finds no fresh owner. Missing registration is
// not evidence against portable history. A target-ticket binding identifies
// the actual selected store; unrelated stores may reuse native identifiers.
func validateRegisteredGenerationBaseline(project, state, ticket string, baseline []domain.SessionFact, configuredStores map[domain.SourceHarness]string) error {
	if len(baseline) == 0 {
		return nil
	}
	registry, err := capture.OpenProject(state, project)
	if err != nil {
		return err
	}
	rows, readErr := registry.List(context.Background())
	if err = errors.Join(readErr, registry.Close()); err != nil {
		return err
	}
	for _, owner := range baseline {
		selectedStores := []string{}
		if store := configuredStores[owner.Source]; store != "" {
			selectedStores = append(selectedStores, store)
		}
		for _, row := range rows {
			if row.Identity.Source == owner.Source && row.Identity.SessionID == owner.ID && row.TicketID == ticket {
				selectedStores = append(selectedStores, row.Identity.Store)
			}
		}
		for _, row := range rows {
			if row.Identity.Source != owner.Source || row.Identity.SessionID != owner.ID || row.TicketID == "" || row.TicketID == ticket {
				continue
			}
			if len(selectedStores) == 0 {
				// With no configured selector, automatic mode uses physical
				// local bindings. Only a positively verified source/project
				// association is a conflict; absent/stale receipts are not.
				if row.ProjectDirectory == project && capture.ValidateOwner(context.Background(), project, row) == nil {
					return fmt.Errorf("recorded source owner %s is positively bound in its local source store to ticket %s; automatic generation cannot relabel it as %s", owner.ID, row.TicketID, ticket)
				}
				continue
			}
			for _, selectedStore := range selectedStores {
				same, compareErr := capture.SameStore(row.Identity.Store, selectedStore)
				if compareErr != nil {
					return compareErr
				}
				if same {
					return fmt.Errorf("recorded source owner %s in selected store is bound to ticket %s, refusing %s", owner.ID, row.TicketID, ticket)
				}
			}
		}
	}
	return nil
}

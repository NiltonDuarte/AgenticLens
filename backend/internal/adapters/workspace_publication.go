package adapters

import (
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/filesystem"
	"agenticlens/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func workspaceArtifactPaths(directory, output, summary string) ([]string, error) {
	paths := []string{filepath.Join(directory, "agentic_metrics.json")}
	for _, path := range []string{output, summary} {
		if path == "" {
			continue
		}
		physical, err := filesystem.PhysicalPath(path)
		if err != nil {
			return nil, err
		}
		if !insideWorkspace(directory, physical) || physical == paths[0] {
			return nil, fmt.Errorf("computed export/summary must be inside the selected workspace and cannot replace its source ledger")
		}
		paths = append(paths, physical)
	}
	return paths, nil
}

// RecoverWorkspaceArtifacts requires explicit location and complete member selection.
// It is used before persisted reads when repeating an interrupted export invocation.
func RecoverWorkspaceArtifacts(project, directory, output, summary string) (err error) {
	_, cfg, err := LoadWorkspaceProject(project)
	if err != nil {
		return err
	}
	directory, err = filesystem.PhysicalPath(directory)
	if err != nil {
		return err
	}
	ledger := filepath.Join(directory, "agentic_metrics.json")
	if err = ValidateLedgerDestinationScope(cfg, ledger); err != nil {
		return err
	}
	paths, err := workspaceArtifactPaths(directory, output, summary)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	owner, err := capture.AcquirePublication(ctx, ledger, project)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, owner.Close()) }()
	_, err = RecoverPublication(ledger, paths)
	return err
}

// WorkspaceArtifactSelection retains physical selection without holding a read
// lock across writer admission. It is request-scoped and must be closed.
type WorkspaceArtifactSelection struct {
	Ledger    domain.AgenticMetricsLedger
	location  WorkspaceLocation
	reference *os.File
}

func (s *WorkspaceArtifactSelection) Close() error {
	if s == nil || s.reference == nil {
		return nil
	}
	err := s.reference.Close()
	s.reference = nil
	return err
}

// LoadWorkspaceArtifactSelection is the write-intent counterpart of the simple
// read-only LoadTicketWorkspace API.
func LoadWorkspaceArtifactSelection(ticket, project, selection string) (result *WorkspaceArtifactSelection, err error) {
	_, cfg, err := LoadWorkspaceProject(project)
	if err != nil {
		return nil, err
	}
	location, reference, err := resolveTicketReadLocation(cfg, ticket, selection)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, reference.Close())
		}
	}()
	ledger, err := readWorkspaceLedger(location, cfg)
	if err != nil {
		return nil, err
	}
	if !sameTicket(*ledger.TicketID, ticket, cfg.Ticket) {
		return nil, &WorkspaceResolutionError{Outcome: "identity_conflict", TicketID: ticket, Candidates: []string{location.Path}}
	}
	return &WorkspaceArtifactSelection{Ledger: ledger, location: location, reference: reference}, nil
}

// PublishWorkspaceArtifacts keeps source ledger, computed export and summary in one
// accepted set. Source facts are never replaced by their computed projection.
func PublishWorkspaceArtifacts(project string, selection *WorkspaceArtifactSelection, output, summary string) (err error) {
	if selection == nil || selection.reference == nil {
		return fmt.Errorf("artifact publication requires an explicitly resolved ledger")
	}
	_, cfg, err := LoadWorkspaceProject(project)
	if err != nil {
		return err
	}
	paths, err := workspaceArtifactPaths(selection.location.Path, output, summary)
	if err != nil {
		return err
	}
	ledger := paths[0]
	if err = ValidateLedgerDestinationScope(cfg, ledger); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	owner, err := capture.AcquirePublication(ctx, ledger, project)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, owner.Close()) }()
	// Compare the retained inode with the owned root before recovery can write.
	selectedInfo, selectedErr := selection.reference.Stat()
	ownedInfo, ownedErr := capture.StatPublicationFile(selection.location.Path)
	if selectedErr != nil || ownedErr != nil || !os.SameFile(selectedInfo, ownedInfo) {
		return errors.Join(fmt.Errorf("workspace changed after selection; resolve and retry"), selectedErr, ownedErr)
	}

	if _, err = RecoverPublication(ledger, paths); err != nil {
		return err
	}
	body, err := readOwnedLedger(ledger)
	if err != nil {
		return err
	}
	if err = ValidateRuntimeFactLedger(body); err != nil {
		return err
	}
	var dto AgenticFactLedgerDTO
	if err = json.Unmarshal(body, &dto); err != nil {
		return err
	}
	if err = dto.Validate(); err != nil {
		return err
	}
	facts, err := domain.NormalizeLedgerForRead(dto.ToDomain())
	if err != nil {
		return err
	}
	if facts.TicketID == nil || !sameTicket(*facts.TicketID, selection.location.TicketID, cfg.Ticket) {
		return fmt.Errorf("selected ledger identity changed before artifact publication")
	}
	current := domain.DeriveMetricsLedger(facts, cfg)
	if err = BeginPublication(ledger, paths); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_, recoveryErr := RecoverPublication(ledger, paths)
			err = errors.Join(err, recoveryErr)
		}
	}()
	if output != "" {
		data, e := json.Marshal(SerializeLedger(current))
		if e != nil {
			return e
		}
		data, e = FormatLedgerJSON(data)
		if e != nil {
			return e
		}
		if e = durableReplace(output, data); e != nil {
			return e
		}
	}
	if summary != "" {
		if err = capture.MkdirPublicationDirectories(filepath.Dir(summary), 0755); err != nil {
			return fmt.Errorf("creating selected summary directory: %w", err)
		}
		table := FormatMarkdownMetricsTable(current, cfg.RoleOrder)
		existing, e := capture.ReadPublicationFile(summary)
		if e != nil && !errors.Is(e, os.ErrNotExist) {
			return e
		}
		updated, e := replaceSummaryMetrics(string(existing), table)
		if e != nil {
			return e
		}
		if e = durableReplace(summary, []byte(updated)); e != nil {
			return e
		}
	}
	return CommitPublication(ledger, paths)
}

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

// PublishWorkspaceArtifacts keeps source ledger, computed export and summary in one
// accepted set. Source facts are never replaced by their computed projection.
func PublishWorkspaceArtifacts(project string, selected domain.AgenticMetricsLedger, output, summary string) (err error) {
	if selected.WorkspacePath == "" || selected.TicketID == nil {
		return fmt.Errorf("artifact publication requires an explicitly resolved ledger")
	}
	_, cfg, err := LoadWorkspaceProject(project)
	if err != nil {
		return err
	}
	paths, err := workspaceArtifactPaths(selected.WorkspacePath, output, summary)
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
	if facts.TicketID == nil || !sameTicket(*facts.TicketID, *selected.TicketID, cfg.Ticket) {
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

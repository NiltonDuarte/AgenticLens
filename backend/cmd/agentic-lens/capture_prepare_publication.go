package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func capturePreparePublication(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("capture prepare-publication", flag.ContinueOnError)
	flags.SetOutput(output)
	repo := flags.String("repo", ".", "Project whose configured workspace roots authorize the destination")
	ticket := flags.String("ticket-id", "", "Exact ticket ID to publish on first native capture")
	path := flags.String("json-output", "", "Canonical agentic_metrics.json in an existing empty ledger directory")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 || strings.TrimSpace(*path) == "" {
		return errors.New("prepare-publication requires --ticket-id and --json-output without positional arguments")
	}
	project, cfg, err := adapters.LoadWorkspaceProject(*repo)
	if err != nil {
		return err
	}
	if matched, ok := domain.NormalizeTicketID(*ticket, cfg.Ticket); !ok || matched != *ticket {
		return errors.New("prepare-publication requires an exact configured ticket ID")
	}
	explicit, err := adaptersconfig.ResolveConfiguredPath(project, *path)
	if err != nil {
		return err
	}
	if filepath.Base(explicit) != "agentic_metrics.json" {
		return errors.New("--json-output must name agentic_metrics.json")
	}
	var protection *capture.ProtectedPublication
	home, err := adaptersconfig.ProtectedCodexProjectHome(project)
	if err != nil {
		return err
	}
	if home != "" {
		admitted, err := capture.AdmitProtectedPublication(project, home)
		if err != nil {
			return err
		}
		protection = &admitted
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := adapters.PrepareWorkspacePublication(ctx, cfg, project, *ticket, explicit, protection); err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "First publication prepared for %s. Launch the configured native session and bind it to the ticket.\n", *ticket)
	return err
}

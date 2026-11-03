package main

import (
	"agenticlens/internal/adapters"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func executeFormat(args []string, out, diagnostics io.Writer) int {
	flags := flag.NewFlagSet("format", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	all := flags.Bool("all", false, "Select all discovered ticket ledgers")
	ticket := flags.String("ticket-id", "", "Select exactly one ticket ledger")
	tickets := flags.String("ticket-ids", "", "Select comma-separated ticket IDs (trimmed and deduplicated)")
	workspaces := flags.String("workspaces", "", "Workspace root (defaults to project configuration)")
	dryRun := flags.Bool("dry-run", false, "Report formatting differences without writing")
	check := flags.Bool("check", false, "Exit 1 for formatting differences; exit 2 for input/I/O failures; never write")
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Usage: agentic-lens format (--all | --ticket-id ID | --ticket-ids ID1,ID2) [flags]\n\nApply the invoked release's ledger layout without sources, migration or fact changes.\nOpt-in only; installs no hook or mandatory CI check and never updates the release pin.")
		flags.PrintDefaults()
	}
	if len(args) == 1 && args[0] == "help" {
		args = []string{"--help"}
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	selected := map[string]bool{}
	flags.Visit(func(f *flag.Flag) { selected[f.Name] = true })
	selectors := 0
	for _, key := range []string{"all", "ticket-id", "ticket-ids"} {
		if selected[key] {
			selectors++
		}
	}
	if selectors != 1 || (selected["all"] && !*all) || (*dryRun && *check) || flags.NArg() != 0 {
		fmt.Fprintln(diagnostics, "format: require exactly one selection mode and mutually exclusive --dry-run/--check, without positional arguments")
		return 2
	}
	ids := []string{}
	if selected["ticket-id"] {
		ids = append(ids, *ticket)
	}
	if selected["ticket-ids"] {
		for _, id := range strings.Split(*tickets, ",") {
			id = strings.TrimSpace(id)
			if id == "" {
				fmt.Fprintln(diagnostics, "format: --ticket-ids contains an empty ticket ID")
				return 2
			}
			ids = append(ids, id)
		}
	}
	mode := adapters.LedgerFormatWrite
	if *dryRun {
		mode = adapters.LedgerFormatDryRun
	}
	if *check {
		mode = adapters.LedgerFormatCheck
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	results, err := adapters.FormatLedgerFiles(ctx, adapters.LedgerFormatOptions{Workspaces: *workspaces, All: *all, TicketIDs: ids, Mode: mode})
	changed, unchanged, failed := 0, 0, 0
	for _, result := range results {
		status := "unchanged"
		switch {
		case result.Error != nil:
			status = "failed"
			failed++
			fmt.Fprintf(diagnostics, "%s: %v\n", result.Path, result.Error)
		case result.Changed:
			changed++
			status = "would change"
			if result.Written {
				status = "changed"
			}
			if *check {
				status = "needs formatting"
			}
		default:
			unchanged++
		}
		fmt.Fprintf(out, "%s: %s\n", result.Path, status)
	}
	fmt.Fprintf(out, "%d ledgers: %d changed, %d unchanged, %d failed\n", len(results), changed, unchanged, failed)
	if err != nil {
		fmt.Fprintln(diagnostics, "format:", err)
		return 2
	}
	if *check && changed > 0 {
		return 1
	}
	return 0
}
func runFormat(args []string) {
	if code := executeFormat(args, os.Stdout, os.Stderr); code != 0 {
		os.Exit(code)
	}
}

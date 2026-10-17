package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

func main() {
	ticketIDFlag := flag.String("ticket-id", "", "Target ticket identifier to filter and report on")
	workspacesFlag := flag.String("workspaces", "", "Path to root workspaces directory containing agentic_metrics.json ledgers")
	appendFlag := flag.String("append", "", "Path to markdown summary file to idempotently update with metrics table")
	jsonOutputFlag := flag.String("json-output", "", "Path to write computed metrics ledger as JSON")
	listFlag := flag.Bool("list", false, "List all discovered ticket workspaces and summary metrics")

	flag.Parse()

	workspacesDir := adapters.ResolveWorkspacesDir(*workspacesFlag)
	report := adapters.ScanWorkspaces(workspacesDir)

	if len(report.Ledgers) == 0 {
		fmt.Fprintf(os.Stderr, "No valid agentic_metrics.json ledgers found in %s\n", workspacesDir)
		os.Exit(1)
	}

	if *listFlag {
		fmt.Printf("Found %d ticket workspaces in %s:\n\n", len(report.Ledgers), workspacesDir)
		header := fmt.Sprintf("%-15s %-12s %-10s %-12s %-10s %s", "Ticket ID", "Duration", "Cost", "Health", "Iterations", "Title")
		fmt.Println(header)
		fmt.Println(strings.Repeat("-", len(header)))
		for _, ledger := range report.Ledgers {
			tid := "UNKNOWN"
			if ledger.TicketID != nil && *ledger.TicketID != "" {
				tid = *ledger.TicketID
			}
			dur := ledger.WorkflowEfficiency.ActiveDurationFormatted
			cost := fmt.Sprintf("$%.2f", ledger.Totals.CostUSD)
			health := ledger.WorkflowEfficiency.ContextWindowHealth.Status
			iters := fmt.Sprintf("%d", ledger.WorkflowEfficiency.SDLCIterations)
			title := "N/A"
			if ledger.TicketTitle != nil && *ledger.TicketTitle != "" {
				title = *ledger.TicketTitle
			}
			fmt.Printf("%-15s %-12s %-10s %-12s %-10s %s\n", tid, dur, cost, health, iters, title)
		}
		return
	}

	var targetLedger *domain.AgenticMetricsLedger
	if *ticketIDFlag != "" {
		for i := range report.Ledgers {
			l := &report.Ledgers[i]
			if l.TicketID != nil && *l.TicketID == *ticketIDFlag {
				targetLedger = l
				break
			}
		}
		if targetLedger == nil {
			fmt.Fprintf(os.Stderr, "Error: Ticket '%s' not found in scanned workspaces (%s)\n", *ticketIDFlag, workspacesDir)
			os.Exit(1)
		}
	} else {
		// Default to most recently updated ledger
		var latest *domain.AgenticMetricsLedger
		for i := range report.Ledgers {
			l := &report.Ledgers[i]
			if latest == nil || l.UpdatedAt > latest.UpdatedAt {
				latest = l
			}
		}
		targetLedger = latest
	}

	tableMD := adapters.FormatMarkdownMetricsTable(*targetLedger)

	if *appendFlag != "" {
		if err := adapters.UpdateSummaryFile(*appendFlag, tableMD); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating summary file %s: %v\n", *appendFlag, err)
			os.Exit(1)
		}
		fmt.Printf("AgenticLens metrics table successfully updated in %s\n", *appendFlag)
	}

	if *jsonOutputFlag != "" {
		parentDir := filepath.Dir(*jsonOutputFlag)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directories for %s: %v\n", *jsonOutputFlag, err)
			os.Exit(1)
		}

		serialized := adapters.SerializeLedger(*targetLedger)
		data, err := json.MarshalIndent(serialized, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}
		data = append(data, '\n')
		if err := os.WriteFile(*jsonOutputFlag, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON to %s: %v\n", *jsonOutputFlag, err)
			os.Exit(1)
		}
		fmt.Printf("AgenticLens computed ledger written to %s\n", *jsonOutputFlag)
	}

	if *appendFlag == "" && *jsonOutputFlag == "" {
		fmt.Println(tableMD)
	}
}

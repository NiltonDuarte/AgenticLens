package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"agenticlens/internal/adapters"
	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/adapters/web"
	"agenticlens/internal/domain"
)

func runServer(args []string) {
	serverFlags := flag.NewFlagSet("server", flag.ExitOnError)
	portFlag := serverFlags.Int("port", 8080, "Port to listen on")
	workspacesFlag := serverFlags.String("workspaces", "", "Path to workspaces directory")

	if err := serverFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing server flags: %v\n", err)
		os.Exit(1)
	}

	workspacesDir := adapters.ResolveWorkspacesDir(*workspacesFlag)

	staticFS, err := web.GetFS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing embedded web filesystem: %v\n", err)
		os.Exit(1)
	}

	srv := adaptershttp.NewServer(workspacesDir, staticFS)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("Starting AgenticLens server on http://localhost:%d (workspaces: %s)\n", *portFlag, workspacesDir)
		if err := srv.ListenAndServe(*portFlag); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
			os.Exit(1)
		}
	}()

	<-stop
	fmt.Println("\nShutting down AgenticLens server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server graceful shutdown failed: %v\n", err)
	}
	fmt.Println("AgenticLens server stopped.")
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "server" {
		runServer(os.Args[2:])
		return
	}

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

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
	"strconv"
	"strings"
	"syscall"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/backfill"
	"agenticlens/internal/adapters/git"
	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/adapters/web"
	"agenticlens/internal/domain"
)

func runGenerate(args []string) {
	genFlags := flag.NewFlagSet("generate", flag.ExitOnError)
	ticketIDFlag := genFlags.String("ticket-id", "", "Target ticket identifier to lock session resolution to")
	nonceFlag := genFlags.String("nonce", "", "Unique random execution nonce to lock session resolution to")
	sessionIDFlag := genFlags.String("session-id", "", "Explicit root session ID")
	ticketTitleFlag := genFlags.String("ticket-title", "", "Human-readable ticket title or descriptive name")
	ticketTypeFlag := genFlags.String("ticket-type", "", "Ticket type classification (e.g., feature, bug, chore, refactor)")
	startedAtFlag := genFlags.String("started-at", "", "Explicit start timestamp ISO format")
	completedAtFlag := genFlags.String("completed-at", "", "Explicit completion timestamp ISO format")
	jsonOutputFlag := genFlags.String("json-output", "", "Path to write/update structured JSON ledger")
	appendFlag := genFlags.String("append", "", "Path to markdown file to append/update metrics table in")
	dbPathFlag := genFlags.String("db-path", "", "Path to OpenCode SQLite database")

	if err := genFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing generate flags: %v\n", err)
		os.Exit(1)
	}

	var cachedFacts *adapters.ExistingLedgerFacts
	if *jsonOutputFlag != "" {
		loaded, err := adapters.LoadExistingFactLedger(*jsonOutputFlag)
		if err == nil && loaded != nil {
			cachedFacts = loaded
		}
	}

	activeTicketID := strings.TrimSpace(*ticketIDFlag)
	if activeTicketID == "" && cachedFacts != nil && cachedFacts.TicketID != nil {
		activeTicketID = strings.TrimSpace(*cachedFacts.TicketID)
	}

	var activeTicketTitle *string
	if strings.TrimSpace(*ticketTitleFlag) != "" {
		clean := strings.TrimSpace(*ticketTitleFlag)
		activeTicketTitle = &clean
	} else if cachedFacts != nil && cachedFacts.TicketTitle != nil {
		activeTicketTitle = cachedFacts.TicketTitle
	} else {
		activeTicketTitle = git.InferTicketTitle(activeTicketID)
	}

	var activeTicketType *string
	if strings.TrimSpace(*ticketTypeFlag) != "" {
		clean := strings.TrimSpace(*ticketTypeFlag)
		activeTicketType = &clean
	} else if cachedFacts != nil && cachedFacts.TicketType != nil {
		activeTicketType = cachedFacts.TicketType
	} else {
		activeTicketType = git.InferTicketType(activeTicketID)
	}

	currentUser := git.GetGitUser()

	var activeStartedAt *string
	if strings.TrimSpace(*startedAtFlag) != "" {
		clean := strings.TrimSpace(*startedAtFlag)
		activeStartedAt = &clean
	} else if cachedFacts != nil && cachedFacts.StartedAt != nil {
		activeStartedAt = cachedFacts.StartedAt
	}

	var activeCompletedAt *string
	if strings.TrimSpace(*completedAtFlag) != "" {
		clean := strings.TrimSpace(*completedAtFlag)
		activeCompletedAt = &clean
	} else if cachedFacts != nil && cachedFacts.CompletedAt != nil {
		activeCompletedAt = cachedFacts.CompletedAt
	}

	var knownThreadIDs []string
	userBySessionID := make(map[string]string)
	var invocations []domain.InvocationFact
	var tools []domain.ToolFact
	var fallbackCode *domain.CodeFacts

	if cachedFacts != nil {
		knownThreadIDs = append(knownThreadIDs, cachedFacts.RootThreadIDs...)
		for _, inv := range cachedFacts.Invocations {
			userBySessionID[inv.ID] = inv.User
		}
		invocations = append(invocations, cachedFacts.Invocations...)
		tools = append(tools, cachedFacts.Tools...)
		fallbackCode = cachedFacts.Code
	}

	dbClient, err := opencode.OpenDB(*dbPathFlag)
	if err == nil {
		defer dbClient.Close()

		rootID, err := opencode.ResolveActiveRootSession(
			dbClient.DB(),
			*nonceFlag,
			activeTicketID,
			*sessionIDFlag,
		)
		if err == nil && rootID != nil && *rootID != "" {
			alreadyKnown := false
			for _, id := range knownThreadIDs {
				if id == *rootID {
					alreadyKnown = true
					break
				}
			}
			if !alreadyKnown {
				knownThreadIDs = append(knownThreadIDs, *rootID)
			}
		}

		if len(knownThreadIDs) > 0 {
			dbInvs, dbTools, dbStarted, dbCompleted, err := opencode.ExtractFacts(
				dbClient.DB(),
				knownThreadIDs,
				currentUser,
				userBySessionID,
			)
			if err == nil && len(dbInvs) > 0 {
				invocations = dbInvs
				tools = dbTools
				if strings.TrimSpace(*startedAtFlag) == "" && dbStarted != nil {
					activeStartedAt = dbStarted
				}
				if strings.TrimSpace(*completedAtFlag) == "" && dbCompleted != nil {
					activeCompletedAt = dbCompleted
				}
			}
		}
	}

	if len(invocations) > 0 && strings.TrimSpace(*startedAtFlag) == "" {
		var minStarted *string
		for _, inv := range invocations {
			if inv.StartedAt != nil && *inv.StartedAt != "" {
				if minStarted == nil || *inv.StartedAt < *minStarted {
					minStarted = inv.StartedAt
				}
			}
		}
		if minStarted != nil {
			if activeStartedAt == nil || *minStarted < *activeStartedAt {
				activeStartedAt = minStarted
			}
		}
	}

	if len(invocations) > 0 && strings.TrimSpace(*completedAtFlag) == "" {
		var maxCompleted *string
		for _, inv := range invocations {
			if inv.CompletedAt != nil && *inv.CompletedAt != "" {
				if maxCompleted == nil || *inv.CompletedAt > *maxCompleted {
					maxCompleted = inv.CompletedAt
				}
			}
		}
		if maxCompleted != nil {
			if activeCompletedAt == nil || *maxCompleted > *activeCompletedAt {
				activeCompletedAt = maxCompleted
			}
		}
	}

	cfg, _ := git.LoadConfig("")
	codeFacts := git.ComputeCodeFacts(activeTicketID, cfg, fallbackCode)

	updatedAtISO := time.Now().UTC().Format("2006-01-02T15:04:05.000000+00:00")

	var ticketIDPtr *string
	if activeTicketID != "" {
		ticketIDPtr = &activeTicketID
	}

	schemaVer := domain.CurrentSchemaVersion
	factLedger := domain.AgenticFactLedger{
		SchemaVersion: &schemaVer,
		TicketID:      ticketIDPtr,
		TicketTitle:   activeTicketTitle,
		TicketType:    activeTicketType,
		StartedAt:     activeStartedAt,
		CompletedAt:   activeCompletedAt,
		UpdatedAt:     updatedAtISO,
		Code:          codeFacts,
		Invocations:   invocations,
		Tools:         tools,
	}

	metricsLedger := domain.DeriveMetricsLedger(factLedger)
	tableMD := adapters.FormatMarkdownMetricsTable(metricsLedger)

	if *jsonOutputFlag != "" {
		parentDir := filepath.Dir(*jsonOutputFlag)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directories for %s: %v\n", *jsonOutputFlag, err)
			os.Exit(1)
		}

		data, err := json.MarshalIndent(factLedger, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}
		data = append(data, '\n')
		if err := os.WriteFile(*jsonOutputFlag, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON to %s: %v\n", *jsonOutputFlag, err)
			os.Exit(1)
		}
		fmt.Printf("Structured metrics ledger written to %s\n", *jsonOutputFlag)
	}

	if *appendFlag != "" {
		if err := adapters.UpdateSummaryFile(*appendFlag, tableMD); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating summary file %s: %v\n", *appendFlag, err)
			os.Exit(1)
		}
		fmt.Printf("Metrics table successfully updated in %s\n", *appendFlag)
	} else if *jsonOutputFlag == "" {
		fmt.Println(tableMD)
	}
}

func runServer(args []string) {
	defaultPort := 8085
	if envPort := strings.TrimSpace(os.Getenv("PORT")); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			defaultPort = p
		} else {
			defaultPort = -1
		}
	}

	serverFlags := flag.NewFlagSet("server", flag.ExitOnError)
	portFlag := serverFlags.Int("port", defaultPort, "Port to listen on")
	workspacesFlag := serverFlags.String("workspaces", "", "Path to workspaces directory")

	if err := serverFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing server flags: %v\n", err)
		os.Exit(1)
	}

	if *portFlag < 1 || *portFlag > 65535 {
		fmt.Fprintf(os.Stderr, "Error: invalid port %d: must be between 1 and 65535\n", *portFlag)
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
			if errors.Is(err, syscall.EADDRINUSE) || strings.Contains(strings.ToLower(err.Error()), "address already in use") {
				fmt.Fprintf(os.Stderr, "Server failed: port %d is already in use. Specify an alternate port via --port or the PORT environment variable.\n", *portFlag)
			} else {
				fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
			}
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

func runBackfill(args []string) {
	backfillFlags := flag.NewFlagSet("backfill", flag.ExitOnError)
	workspacesFlag := backfillFlags.String("workspaces", "", "Path to root workspaces directory containing agentic_metrics.json ledgers")
	ticketIDFlag := backfillFlags.String("ticket-id", "", "Target single ticket identifier to filter backfill operation")
	recomputeGitFlag := backfillFlags.Bool("recompute-git", false, "Recompute Git code facts and SDLC iterations from repository history")
	recomputeDBFlag := backfillFlags.Bool("recompute-db", false, "Recompute OpenCode invocations and tool facts from local SQLite DB")
	dryRunFlag := backfillFlags.Bool("dry-run", false, "Preview planned migrations without writing changes to disk")
	updateSummariesFlag := backfillFlags.Bool("update-summaries", false, "Idempotently update SUMMARY.md markdown metrics tables")
	dbPathFlag := backfillFlags.String("db-path", "", "Explicit path to OpenCode SQLite database file")
	forceFlag := backfillFlags.Bool("force", false, "Force migration and recompute even if ledger is already completed")

	if err := backfillFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing backfill flags: %v\n", err)
		os.Exit(1)
	}

	opts := backfill.BackfillOptions{
		WorkspacesDir:   *workspacesFlag,
		TicketID:        *ticketIDFlag,
		RecomputeGit:    *recomputeGitFlag,
		RecomputeDB:     *recomputeDBFlag,
		DryRun:          *dryRunFlag,
		UpdateSummaries: *updateSummariesFlag,
		DBPath:          *dbPathFlag,
		Force:           *forceFlag,
	}

	engine := backfill.NewEngine()
	report, err := engine.Execute(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Backfill failed: %v\n", err)
		os.Exit(1)
	}

	targetDir := adapters.ResolveWorkspacesDir(opts.WorkspacesDir)
	modeStr := "EXECUTION"
	if opts.DryRun {
		modeStr = "DRY-RUN PREVIEW"
	}

	fmt.Println("================================================================================")
	fmt.Printf("AGENTIC LENS - TELEMETRY BACKFILL & SCHEMA MIGRATION [%s]\n", modeStr)
	fmt.Println("================================================================================")
	fmt.Printf("Target Workspaces Directory: %s\n", targetDir)
	if opts.TicketID != "" {
		fmt.Printf("Filter Ticket ID:            %s\n", opts.TicketID)
	}
	fmt.Printf("Recompute Git Facts:         %t\n", opts.RecomputeGit)
	fmt.Printf("Recompute OpenCode DB:       %t\n", opts.RecomputeDB)
	fmt.Printf("Update Markdown Summaries:   %t\n", opts.UpdateSummaries)
	fmt.Printf("Force Migration:             %t\n", opts.Force)
	fmt.Println("--------------------------------------------------------------------------------")

	if len(report.Results) == 0 {
		fmt.Println("No matching ticket ledgers found to backfill.")
		return
	}

	header := fmt.Sprintf("%-15s %-14s %-12s %-8s %-8s %-8s %s", "Ticket ID", "Prev Status", "New Status", "Git", "DB", "Skipped", "Notes")
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))

	for _, res := range report.Results {
		gitStr := "no"
		if res.GitUpdated {
			gitStr = "yes"
		}
		dbStr := "no"
		if res.DBUpdated {
			dbStr = "updated"
		} else if res.DBFound {
			dbStr = "found"
		} else if opts.RecomputeDB {
			dbStr = "missing"
		}
		skipStr := "no"
		if res.Skipped {
			skipStr = "yes"
		}
		notes := res.Notes
		if len(res.PendingSteps) > 0 {
			notes = fmt.Sprintf("pending: [%s]", strings.Join(res.PendingSteps, ", "))
		}
		if res.Error != nil {
			notes = fmt.Sprintf("error: %v", res.Error)
		}

		fmt.Printf("%-15s %-14s %-12s %-8s %-8s %-8s %s\n",
			res.TicketID,
			res.PreviousStatus,
			res.NewStatus,
			gitStr,
			dbStr,
			skipStr,
			notes,
		)
	}

	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("Summary: Scanned: %d | Updated: %d | Skipped: %d | Completed: %d | Partial: %d\n",
		report.Scanned, report.Updated, report.Skipped, report.Completed, report.Partial)

	if len(report.Warnings) > 0 {
		fmt.Printf("\nWarnings / Non-fatal Errors (%d):\n", len(report.Warnings))
		for _, w := range report.Warnings {
			fmt.Printf("- %s\n", w)
		}
	}
	fmt.Println("================================================================================")
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "server" {
		runServer(os.Args[2:])
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "generate" {
		runGenerate(os.Args[2:])
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "backfill" {
		runBackfill(os.Args[2:])
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

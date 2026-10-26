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
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/backfill"
	"agenticlens/internal/adapters/codex"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/git"
	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/adapters/live"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/adapters/web"
	"agenticlens/internal/domain"
)

func runGenerate(args []string) {
	genFlags := flag.NewFlagSet("generate", flag.ExitOnError)
	if len(args) > 0 && args[0] == "help" {
		genFlags.Usage()
		return
	}

	ticketIDFlag := genFlags.String("ticket-id", "", "Target ticket identifier to lock session resolution to")
	codeRefFlag := genFlags.String("code-ref", "HEAD", "Git branch or commit to measure against the configured base (defaults to HEAD; committed changes only)")
	sourceFlag := genFlags.String("source", "opencode", "Telemetry harness to read: opencode or codex")
	nonceFlag := genFlags.String("nonce", "", "Unique random execution nonce to lock session resolution to")
	sessionIDFlag := genFlags.String("session-id", "", "Explicit root session ID")
	ticketTitleFlag := genFlags.String("ticket-title", "", "Human-readable ticket title or descriptive name")
	ticketTypeFlag := genFlags.String("ticket-type", "", "Ticket type classification (e.g., feat, fix, docs, chore, refactor, test, ci, perf, build)")
	strictTypeFlag := genFlags.Bool("strict-type", false, "Require a valid ticket type, failing if omitted and cannot be inferred")
	startedAtFlag := genFlags.String("started-at", "", "Explicit start timestamp ISO format")
	completedAtFlag := genFlags.String("completed-at", "", "Explicit completion timestamp ISO format")
	jsonOutputFlag := genFlags.String("json-output", "", "Path to write/update structured JSON ledger")
	appendFlag := genFlags.String("append", "", "Path to markdown file to append/update metrics table in")
	dbPathFlag := genFlags.String("db-path", "", "Path to OpenCode SQLite database")
	codexHomeFlag := genFlags.String("codex-home", "", "Path to CODEX_HOME (defaults to $CODEX_HOME or ~/.codex)")
	anomaliesFlag := genFlags.Bool("anomalies", false, "Display anomalies and health flags for target ticket")
	allowEmptyFlag := genFlags.Bool("allow-empty", false, "Allow empty metrics generation even if no root session or invocations are resolved (test escape hatch)")

	if err := genFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing generate flags: %v\n", err)
		os.Exit(1)
	}
	activeSource := strings.ToLower(strings.TrimSpace(*sourceFlag))
	if activeSource != "opencode" && activeSource != "codex" {
		fmt.Fprintf(os.Stderr, "Error: unsupported source %q. Use opencode or codex.\n", *sourceFlag)
		os.Exit(1)
	}

	var cachedFacts *adapters.ExistingLedgerFacts
	if *jsonOutputFlag != "" {
		loaded, err := adapters.LoadExistingFactLedger(*jsonOutputFlag)
		if errors.Is(err, adapters.ErrToolSessionIDRequired) {
			fmt.Fprintf(os.Stderr, "Error loading existing fact ledger: %v\n", err)
			os.Exit(1)
		}
		if loaded != nil {
			cachedFacts = loaded
		}
	}

	projCfg, err := adaptersconfig.Load("")
	if err != nil {
		if errors.Is(err, adaptersconfig.ErrConfigNotFound) {
			fmt.Fprintln(os.Stderr, "Project configuration not found. Please run 'agentic-lens init' to initialize the project.")
		} else {
			var valErr *adaptersconfig.ConfigValidationError
			if errors.As(err, &valErr) {
				for _, f := range valErr.MissingFields {
					if f == "workspaces.src_folders" {
						fmt.Fprintf(os.Stderr, "Error: workspaces.src_folders is empty or unconfigured: configure 'workspaces.src_folders' in .agenticlens-config.yaml\n")
						os.Exit(1)
					}
				}
			}
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		}
		os.Exit(1)
	}

	activeTicketID := strings.TrimSpace(*ticketIDFlag)
	if activeTicketID == "" && cachedFacts != nil && cachedFacts.TicketID != nil {
		activeTicketID = strings.TrimSpace(*cachedFacts.TicketID)
	}
	if activeTicketID == "" {
		fmt.Fprintf(os.Stderr, "Error: missing ticket ID. Specify --ticket-id explicitly.\n")
		os.Exit(1)
	}

	var activeTicketTitle *string
	if strings.TrimSpace(*ticketTitleFlag) != "" {
		clean := strings.TrimSpace(*ticketTitleFlag)
		activeTicketTitle = &clean
	} else if cachedFacts != nil && cachedFacts.TicketTitle != nil {
		clean := strings.TrimSpace(*cachedFacts.TicketTitle)
		if clean != "" {
			activeTicketTitle = &clean
		}
	}
	if activeTicketTitle == nil || strings.TrimSpace(*activeTicketTitle) == "" {
		fmt.Fprintf(os.Stderr, "Error: missing ticket title for ticket '%s'. Specify --ticket-title explicitly.\n", activeTicketID)
		os.Exit(1)
	}

	var activeTicketType *domain.ConventionalTicketType
	if strings.TrimSpace(*ticketTypeFlag) != "" {
		normalized, ok := domain.NormalizeTicketType(*ticketTypeFlag)
		if !ok {
			var names []string
			for _, t := range domain.CanonicalTicketTypes {
				names = append(names, string(t))
			}
			fmt.Fprintf(os.Stderr, "Error: invalid ticket type '%s'. Must be one of: %s\n", *ticketTypeFlag, strings.Join(names, ", "))
			os.Exit(1)
		}
		activeTicketType = &normalized
	} else if cachedFacts != nil && cachedFacts.TicketType != nil {
		if normalized, ok := domain.NormalizeTicketType(string(*cachedFacts.TicketType)); ok {
			activeTicketType = &normalized
		} else {
			activeTicketType = cachedFacts.TicketType
		}
	}

	if activeTicketType == nil && activeTicketID != "" {
		activeTicketType = git.InferTicketType(activeTicketID, projCfg.Ticket.Prefixes...)
	}

	if activeTicketType == nil {
		targetTicket := activeTicketID
		if *strictTypeFlag {
			fmt.Fprintf(os.Stderr, "Error: missing ticket type for ticket '%s'. Specify --ticket-type explicitly.\n", targetTicket)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "WARNING: Missing ticket type for ticket %s. Specify --ticket-type explicitly.\n", targetTicket)
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
	var sessions []domain.SessionFact
	var fallbackCode *domain.CodeFacts

	if cachedFacts != nil {
		knownThreadIDs = append(knownThreadIDs, cachedFacts.RootThreadIDs...)
		sessions = append(sessions, cachedFacts.Sessions...)
		cachedLedger := domain.AgenticFactLedger{Sessions: cachedFacts.Sessions}
		for _, inv := range cachedLedger.InvocationFacts() {
			userBySessionID[inv.ID] = inv.User
		}
		fallbackCode = cachedFacts.Code
		if cachedFacts.Migration != nil && strings.Contains(cachedFacts.Migration.Notes, "Removed ") {
			fmt.Fprintf(os.Stderr, "WARNING: %s\n", cachedFacts.Migration.Notes)
		}
	}

	var localInvocations []domain.InvocationFact
	var localTools []domain.ToolFact
	var localEvents []domain.EventFact
	var hasLocalExtraction bool

	if activeSource == "codex" {
		codexHome := strings.TrimSpace(*codexHomeFlag)
		if codexHome == "" {
			codexHome = strings.TrimSpace(os.Getenv("CODEX_HOME"))
		}
		if codexHome == "" {
			if home, homeErr := os.UserHomeDir(); homeErr == nil {
				codexHome = filepath.Join(home, ".codex")
			}
		}
		rolloutPaths, discoverErr := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{
			CodexHome: codexHome, Nonce: *nonceFlag, SessionID: *sessionIDFlag,
		})
		if discoverErr != nil {
			fmt.Fprintf(os.Stderr, "Error discovering Codex sessions: %v\n", discoverErr)
			os.Exit(1)
		}
		transcripts := make([]codex.Transcript, 0, len(rolloutPaths))
		for _, path := range rolloutPaths {
			transcript, parseErr := codex.ParseRolloutPath(path)
			if parseErr != nil {
				fmt.Fprintf(os.Stderr, "Error reading discovered Codex session: %v\n", parseErr)
				os.Exit(1)
			}
			transcripts = append(transcripts, transcript)
			knownThreadIDs = append(knownThreadIDs, transcript.ThreadID)
		}
		combined, combineErr := codex.CombineTranscripts(transcripts)
		if combineErr != nil {
			fmt.Fprintf(os.Stderr, "Error combining Codex session facts: %v\n", combineErr)
			os.Exit(1)
		}
		for _, warning := range combined.Warnings {
			fmt.Fprintf(os.Stderr, "Warning: Codex rollout: %s\n", warning)
		}
		localInvocations = combined.Invocations
		localTools = combined.Tools
		localEvents = combined.Events
		hasLocalExtraction = true
		if strings.TrimSpace(*startedAtFlag) == "" && combined.StartedAt != "" {
			activeStartedAt = &combined.StartedAt
		}
		if strings.TrimSpace(*completedAtFlag) == "" && combined.CompletedAt != "" {
			activeCompletedAt = &combined.CompletedAt
		}
		fmt.Fprintf(os.Stderr, "Discovered %d Codex thread rollout(s) for session %s.\n", len(rolloutPaths), combined.ThreadID)
	} else {
		dbClient, dbErr := opencode.OpenDB(*dbPathFlag)
		if dbErr == nil {
			defer dbClient.Close()

			rootID, resolveErr := opencode.ResolveActiveRootSession(
				dbClient.DB(),
				*nonceFlag,
				activeTicketID,
				*sessionIDFlag,
			)
			if resolveErr == nil && rootID != nil && *rootID != "" {
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
				dbInvs, dbTools, dbEvents, dbStarted, dbCompleted, extractErr := opencode.ExtractFacts(
					dbClient.DB(), knownThreadIDs, currentUser, userBySessionID,
				)
				if extractErr == nil && len(dbInvs) > 0 {
					localInvocations = dbInvs
					localTools = dbTools
					localEvents = dbEvents
					hasLocalExtraction = true
					if strings.TrimSpace(*startedAtFlag) == "" && dbStarted != nil {
						activeStartedAt = dbStarted
					}
					if strings.TrimSpace(*completedAtFlag) == "" && dbCompleted != nil {
						activeCompletedAt = dbCompleted
					}
				}
			}
		}
	}

	if hasLocalExtraction {
		freshSessions, buildErr := domain.SessionsFromFacts(localInvocations, localTools, localEvents)
		if buildErr != nil {
			fmt.Fprintf(os.Stderr, "Error grouping extracted session facts: %v\n", buildErr)
			os.Exit(1)
		}
		sessions = mergeSessions(sessions, freshSessions)
	}
	invocations := (domain.AgenticFactLedger{Sessions: sessions}).InvocationFacts()

	if len(knownThreadIDs) == 0 && len(invocations) == 0 && !*allowEmptyFlag {
		fmt.Fprintf(os.Stderr, "Error: unable to identify root session for ticket '%s'. Provide --session-id, --nonce, or ensure workspace files exist in OpenCode database. Telemetry generation aborted.\n", activeTicketID)
		os.Exit(1)
	}

	if len(invocations) == 0 {
		fmt.Fprintf(os.Stderr, "WARNING: Zero invocations recorded for ticket %s. This is not expected.\n", activeTicketID)
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

	cfg, err := git.LoadConfig("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	for _, warning := range git.CheckUnconfiguredPatterns(cfg) {
		fmt.Fprintf(os.Stderr, "WARNING: %s\n", warning)
	}
	codeFacts, err := git.ComputeCodeFactsAtRef(strings.TrimSpace(*codeRefFlag), cfg)
	if err != nil {
		if fallbackCode != nil {
			fmt.Fprintf(os.Stderr, "WARNING: Failed to compute git code facts (%v); falling back to cached code facts\n", err)
			codeFacts = *fallbackCode
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	var totalSDLC int64
	for _, inv := range invocations {
		if inv.ParentID == nil {
			totalSDLC += inv.UserIterations
		}
	}

	codeFacts.SDLCIterations = totalSDLC
	if totalSDLC == 0 && cachedFacts != nil && (cachedFacts.SchemaVersion == nil || *cachedFacts.SchemaVersion < 4) {
		fmt.Fprintf(os.Stderr, "WARNING: Existing fact ledger has schema version < 4 and must be migrated via 'agentic-lens migrate' or 'agentic-lens backfill'.\n")
	}

	updatedAtISO := time.Now().UTC().Format("2006-01-02T15:04:05.000000+00:00")

	if activeCompletedAt == nil || strings.TrimSpace(*activeCompletedAt) == "" {
		activeCompletedAt = &updatedAtISO
	} else {
		clean := strings.TrimSpace(*activeCompletedAt)
		activeCompletedAt = &clean
	}

	if activeCompletedAt == nil || strings.TrimSpace(*activeCompletedAt) == "" {
		fmt.Fprintf(os.Stderr, "Error: mandatory completed_at timestamp could not be resolved\n")
		os.Exit(1)
	}

	var ticketIDPtr *string
	if activeTicketID != "" {
		ticketIDPtr = &activeTicketID
	}

	schemaVer := domain.CurrentSchemaVersion
	factLedger := domain.AgenticFactLedger{
		SchemaVersion: &schemaVer,
		Source:        activeSource,
		TicketID:      ticketIDPtr,
		TicketTitle:   activeTicketTitle,
		TicketType:    activeTicketType,
		StartedAt:     activeStartedAt,
		CompletedAt:   activeCompletedAt,
		UpdatedAt:     updatedAtISO,
		Code:          codeFacts,
		Sessions:      sessions,
	}

	metricsLedger := domain.DeriveMetricsLedger(factLedger, projCfg)
	tableMD := adapters.FormatMarkdownMetricsTable(metricsLedger, projCfg.RoleOrder)

	if *jsonOutputFlag != "" {
		persistedFactLedger, pruneErr := domain.PersistAnomalySummaries(factLedger)
		if pruneErr != nil {
			fmt.Fprintf(os.Stderr, "Error preparing session-owned telemetry: %v\n", pruneErr)
			os.Exit(1)
		}
		if err := adapters.SaveFactLedger(*jsonOutputFlag, persistedFactLedger); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON to %s: %v\n", *jsonOutputFlag, err)
			os.Exit(1)
		}
		fmt.Printf("Structured metrics ledger written to %s\n", *jsonOutputFlag)
	}

	if *anomaliesFlag {
		tid := "UNKNOWN"
		if metricsLedger.TicketID != nil && *metricsLedger.TicketID != "" {
			tid = *metricsLedger.TicketID
		}
		printAnomalyDiagnostics(tid, metricsLedger.Anomalies)
	}

	if *appendFlag != "" {
		if err := adapters.UpdateSummaryFile(*appendFlag, tableMD); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating summary file %s: %v\n", *appendFlag, err)
			os.Exit(1)
		}
		fmt.Printf("Metrics table successfully updated in %s\n", *appendFlag)
	} else if *jsonOutputFlag == "" && !*anomaliesFlag {
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
	dbPathFlag := serverFlags.String("db-path", "", "Path to OpenCode SQLite database (overrides config)")
	if len(args) > 0 && args[0] == "help" {
		serverFlags.Usage()
		return
	}

	if err := serverFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing server flags: %v\n", err)
		os.Exit(1)
	}

	if *portFlag < 1 || *portFlag > 65535 {
		fmt.Fprintf(os.Stderr, "Error: invalid port %d: must be between 1 and 65535\n", *portFlag)
		os.Exit(1)
	}

	projCfg, err := adaptersconfig.Load("")
	if err != nil {
		if errors.Is(err, adaptersconfig.ErrConfigNotFound) {
			fmt.Fprintln(os.Stderr, "Project configuration not found. Please run 'agentic-lens init' to initialize the project.")
		} else {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		}
		os.Exit(1)
	}

	workspacesDir, err := adapters.ResolveWorkspacesDir(*workspacesFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: workspaces.dir is empty or unconfigured: configure 'workspaces.dir' in .agenticlens-config.yaml or provide --workspaces flag. Error: workspaces directory must be configured in project configuration (workspaces.dir) or provided via --workspaces flag. Server cannot start without this information.\n")
		os.Exit(1)
	}

	rawFlag := strings.TrimSpace(*workspacesFlag)
	if rawFlag != "" {
		projCfg.Workspaces.Dir = workspacesDir
	} else if !filepath.IsAbs(workspacesDir) {
		cfgPath, _ := adaptersconfig.FindConfigFile("")
		if cfgPath != "" {
			baseDir := filepath.Dir(cfgPath)
			candidate := filepath.Join(baseDir, workspacesDir)
			if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
				workspacesDir = candidate
			}
		}
	}

	fi, err := os.Stat(workspacesDir)
	if err != nil || !fi.IsDir() {
		rawDir := strings.TrimSpace(*workspacesFlag)
		if rawDir == "" {
			rawDir = strings.TrimSpace(projCfg.Workspaces.Dir)
		}
		fmt.Fprintf(os.Stderr, "Error: workspaces directory %q does not exist. Server cannot start.\n", rawDir)
		os.Exit(1)
	}

	staticFS, err := web.GetFS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing embedded web filesystem: %v\n", err)
		os.Exit(1)
	}

	dbPath := strings.TrimSpace(*dbPathFlag)
	if dbPath == "" {
		dbPath = strings.TrimSpace(projCfg.Database.Path)
	}
	if dbPath == "" {
		if defPath, err := opencode.DefaultDBPath(); err == nil {
			dbPath = defPath
		}
	}

	cleanDBPath := dbPath
	if strings.HasPrefix(cleanDBPath, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			cleanDBPath = filepath.Join(home, cleanDBPath[2:])
		}
	}

	liveService := live.NewLiveService(nil, projCfg)
	if cleanDBPath != "" {
		if _, err := os.Stat(cleanDBPath); err == nil {
			busyTimeout := projCfg.Database.BusyTimeoutMs
			if busyTimeout <= 0 {
				busyTimeout = 5000
			}
			hook, err := live.OpenReadOnlyHook(cleanDBPath, busyTimeout)
			if err != nil {
				fmt.Printf("Notice: could not initialize live SQLite hook (%v); live telemetry disabled\n", err)
			} else {
				liveService = live.NewLiveService(hook, projCfg)
				liveService.Start(context.Background())
				fmt.Printf("Live telemetry hook active on %s (polling every %dms)\n", cleanDBPath, projCfg.Live.PollIntervalMs)
			}
		}
	}

	srv := adaptershttp.NewServerWithConfig(workspacesDir, staticFS, projCfg, liveService)

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

	if liveService != nil {
		if err := liveService.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning closing live service: %v\n", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server graceful shutdown failed: %v\n", err)
	}
	fmt.Println("AgenticLens server stopped.")
}

func runBackfill(args []string) {
	backfillFlags := flag.NewFlagSet("backfill", flag.ExitOnError)
	backfillFlags.Usage = func() {
		out := backfillFlags.Output()
		fmt.Fprintf(out, "Usage: agentic-lens backfill [flags]\n\n")
		fmt.Fprintf(out, "Batch telemetry backfill, Git code facts recomputation, and schema upgrades.\n\n")
		fmt.Fprintf(out, "Flags:\n")
		fmt.Fprintf(out, "  --workspaces <path>    Path to root workspaces directory containing agentic_metrics.json ledgers\n")
		fmt.Fprintf(out, "  --ticket-id <id>       Target single ticket identifier to filter backfill operation\n")
		fmt.Fprintf(out, "  --recompute-git        Recompute Git code facts and SDLC iterations from repository history\n")
		fmt.Fprintf(out, "  --code-ref <ref>        Required with --recompute-git; branch or commit to compare with configured base\n")
		fmt.Fprintf(out, "  --recompute-db         Recompute OpenCode invocations and tool facts from local SQLite DB\n")
		fmt.Fprintf(out, "  --migrate-schema       Upgrade telemetry ledger schema versions to latest supported version\n")
		fmt.Fprintf(out, "  --dry-run              Preview planned migrations without writing changes to disk\n")
		fmt.Fprintf(out, "  --update-summaries     Idempotently update SUMMARY.md markdown metrics tables\n")
		fmt.Fprintf(out, "  --db-path <path>       Explicit path to OpenCode SQLite database file\n")
		fmt.Fprintf(out, "  --force                Force migration and recompute even if ledger is already completed\n")
	}
	if len(args) > 0 && (args[0] == "help" || args[0] == "-h" || args[0] == "--help" || args[0] == "-help") {
		backfillFlags.SetOutput(os.Stdout)
		backfillFlags.Usage()
		return
	}
	workspacesFlag := backfillFlags.String("workspaces", "", "Path to root workspaces directory containing agentic_metrics.json ledgers")
	ticketIDFlag := backfillFlags.String("ticket-id", "", "Target single ticket identifier to filter backfill operation")
	recomputeGitFlag := backfillFlags.Bool("recompute-git", false, "Recompute Git code facts and SDLC iterations from repository history")
	codeRefFlag := backfillFlags.String("code-ref", "", "Git branch or commit to compare with the configured base when recomputing code facts")
	recomputeDBFlag := backfillFlags.Bool("recompute-db", false, "Recompute OpenCode invocations and tool facts from local SQLite DB")
	migrateSchemaFlag := backfillFlags.Bool("migrate-schema", false, "Upgrade telemetry ledger schema versions to latest supported version")
	dryRunFlag := backfillFlags.Bool("dry-run", false, "Preview planned migrations without writing changes to disk")
	updateSummariesFlag := backfillFlags.Bool("update-summaries", false, "Idempotently update SUMMARY.md markdown metrics tables")
	dbPathFlag := backfillFlags.String("db-path", "", "Explicit path to OpenCode SQLite database file")
	forceFlag := backfillFlags.Bool("force", false, "Force migration and recompute even if ledger is already completed")

	if err := backfillFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing backfill flags: %v\n", err)
		os.Exit(1)
	}
	if *recomputeGitFlag {
		if strings.TrimSpace(*codeRefFlag) == "" {
			fmt.Fprintln(os.Stderr, "Error: --code-ref is required with --recompute-git")
			os.Exit(1)
		}
		if strings.TrimSpace(*ticketIDFlag) == "" {
			fmt.Fprintln(os.Stderr, "Error: --ticket-id is required when --recompute-git is specified")
			os.Exit(1)
		}
	}

	opts := backfill.BackfillOptions{
		WorkspacesDir:   *workspacesFlag,
		TicketID:        *ticketIDFlag,
		RecomputeGit:    *recomputeGitFlag,
		CodeRef:         strings.TrimSpace(*codeRefFlag),
		RecomputeDB:     *recomputeDBFlag,
		MigrateSchema:   *migrateSchemaFlag,
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

	targetDir, err := adapters.ResolveWorkspacesDir(opts.WorkspacesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving workspaces directory: %v\n", err)
		os.Exit(1)
	}
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
	if opts.RecomputeGit {
		fmt.Printf("Git Code Ref:                %s\n", opts.CodeRef)
	}
	fmt.Printf("Recompute OpenCode DB:       %t\n", opts.RecomputeDB)
	fmt.Printf("Migrate Schema:              %t\n", opts.MigrateSchema)
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

	if report.SchemaMigration != nil {
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Printf("Schema Migration: Migrated: %d | Up to date: %d | Skipped: %d | Errors: %d\n",
			report.SchemaMigration.Migrated, report.SchemaMigration.UpToDate,
			report.SchemaMigration.Skipped, len(report.SchemaMigration.Errors))
	}

	if len(report.Warnings) > 0 {
		fmt.Printf("\nWarnings / Non-fatal Errors (%d):\n", len(report.Warnings))
		for _, w := range report.Warnings {
			fmt.Printf("- %s\n", w)
		}
	}
	fmt.Println("================================================================================")
}

func runMigrate(args []string) {
	migrateFlags := flag.NewFlagSet("migrate", flag.ExitOnError)
	migrateFlags.Usage = func() {
		out := migrateFlags.Output()
		fmt.Fprintf(out, "Usage: agentic-lens migrate [flags]\n\n")
		fmt.Fprintf(out, "Universal stepwise schema migration engine for telemetry ledgers.\n\n")
		fmt.Fprintf(out, "Flags:\n")
		fmt.Fprintf(out, "  --workspaces <path>      Path to root workspaces directory containing agentic_metrics.json ledgers\n")
		fmt.Fprintf(out, "  --ticket-id <id>         Target single ticket identifier to filter migration\n")
		fmt.Fprintf(out, "  --target-version <int>   Target schema version integer (default %d)\n", domain.CurrentSchemaVersion)
		fmt.Fprintf(out, "  --dry-run                Preview planned migrations without writing changes to disk\n")
		fmt.Fprintf(out, "  --backup                 Create .bak backup copy of original file before writing\n")
		fmt.Fprintf(out, "  --update-summaries       Idempotently update SUMMARY.md markdown metrics tables\n")
	}
	if len(args) > 0 && (args[0] == "help" || args[0] == "-h" || args[0] == "--help" || args[0] == "-help") {
		migrateFlags.SetOutput(os.Stdout)
		migrateFlags.Usage()
		return
	}
	workspacesFlag := migrateFlags.String("workspaces", "", "Path to root workspaces directory containing agentic_metrics.json ledgers")
	ticketIDFlag := migrateFlags.String("ticket-id", "", "Target single ticket identifier to filter migration")
	targetVersionFlag := migrateFlags.Int("target-version", domain.CurrentSchemaVersion, "Target schema version integer")
	dryRunFlag := migrateFlags.Bool("dry-run", false, "Preview planned migrations without writing changes to disk")
	backupFlag := migrateFlags.Bool("backup", false, "Create .bak backup copy of original file before writing")
	updateSummariesFlag := migrateFlags.Bool("update-summaries", false, "Idempotently update SUMMARY.md markdown metrics tables")

	if err := migrateFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing migrate flags: %v\n", err)
		os.Exit(1)
	}

	opts := backfill.MigrationOptions{
		WorkspacesDir:   *workspacesFlag,
		TicketID:        *ticketIDFlag,
		TargetVersion:   *targetVersionFlag,
		DryRun:          *dryRunFlag,
		Backup:          *backupFlag,
		UpdateSummaries: *updateSummariesFlag,
	}

	runner := backfill.NewMigrationRunner(nil)
	report, err := runner.Run(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
		os.Exit(1)
	}

	targetDir, err := adapters.ResolveWorkspacesDir(opts.WorkspacesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving workspaces directory: %v\n", err)
		os.Exit(1)
	}
	modeStr := "EXECUTION"
	if opts.DryRun {
		modeStr = "DRY-RUN PREVIEW"
	}

	fmt.Println("================================================================================")
	fmt.Printf("AGENTIC LENS - UNIVERSAL SCHEMA MIGRATION [%s]\n", modeStr)
	fmt.Println("================================================================================")
	fmt.Printf("Target Workspaces Directory: %s\n", targetDir)
	if opts.TicketID != "" {
		fmt.Printf("Filter Ticket ID:            %s\n", opts.TicketID)
	}
	fmt.Printf("Target Schema Version:       v%d\n", opts.TargetVersion)
	fmt.Printf("Create Backups (.bak):       %t\n", opts.Backup)
	fmt.Printf("Update Markdown Summaries:   %t\n", opts.UpdateSummaries)
	fmt.Println("--------------------------------------------------------------------------------")

	if len(report.Results) == 0 {
		fmt.Println("No matching ticket workspaces found to migrate.")
		return
	}

	header := fmt.Sprintf("%-15s %-12s %-12s %-18s %s", "Ticket ID", "From Version", "To Version", "Status", "Notes")
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))

	for _, res := range report.Results {
		fromStr := "unknown"
		if res.SourceVersion == 0 {
			fromStr = "unversioned"
		} else if res.SourceVersion > 0 {
			fromStr = fmt.Sprintf("v%d", res.SourceVersion)
		}

		toStr := fmt.Sprintf("v%d", res.TargetVersion)
		if res.TargetVersion == 0 {
			toStr = "-"
		}

		notes := res.DiffSummary
		if res.Error != nil {
			notes = fmt.Sprintf("error: %v", res.Error)
		}

		fmt.Printf("%-15s %-12s %-12s %-18s %s\n",
			res.TicketID,
			fromStr,
			toStr,
			res.Status,
			notes,
		)
	}

	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("Summary: Scanned: %d | Migrated: %d | Up to date: %d | Skipped: %d | Errors: %d\n",
		report.Scanned, report.Migrated, report.UpToDate, report.Skipped, len(report.Errors))

	if len(report.Errors) > 0 {
		fmt.Printf("\nWarnings / Errors (%d):\n", len(report.Errors))
		for _, w := range report.Errors {
			fmt.Printf("- %s\n", w)
		}
	}
	fmt.Println("================================================================================")
}

func formatAnomalyListSummary(anomalies []domain.Anomaly) string {
	if len(anomalies) == 0 {
		return "0"
	}
	var crit, warn int
	for _, a := range anomalies {
		if a.Severity == domain.SeverityCritical {
			crit++
		} else {
			warn++
		}
	}
	var parts []string
	if crit > 0 {
		parts = append(parts, fmt.Sprintf("%d crit", crit))
	}
	if warn > 0 {
		parts = append(parts, fmt.Sprintf("%d warn", warn))
	}
	return strings.Join(parts, ", ")
}

func printAnomalyDiagnostics(ticketID string, anomalies []domain.Anomaly) {
	if len(anomalies) == 0 {
		fmt.Printf("No anomalies detected for ticket %s.\n", ticketID)
		return
	}
	fmt.Printf("Found %d anomalies for ticket %s:\n\n", len(anomalies), ticketID)
	header := fmt.Sprintf("%-10s %-28s %-40s %s", "Severity", "Rule ID", "Title", "Description")
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))
	for _, a := range anomalies {
		fmt.Printf("%-10s %-28s %-40s %s\n", string(a.Severity), string(a.RuleID), a.Title, a.Description)
	}
}

func printRootUsage() {
	out := flag.CommandLine.Output()
	fmt.Fprintln(out, "AgenticLens Developer Telemetry and Observability CLI")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  agentic-lens <subcommand> [flags]")
	fmt.Fprintln(out, "  agentic-lens [flags]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Subcommands:")
	fmt.Fprintln(out, "  init      Scaffold and generate a new .agenticlens-config.yaml configuration")
	fmt.Fprintln(out, "  server    Start the HTTP dashboard server and API service")
	fmt.Fprintln(out, "  generate  Ingest and compute facts from OpenCode or Codex and Git history")
	fmt.Fprintln(out, "  backfill  Batch telemetry backfill, Git code facts recomputation, and schema upgrades")
	fmt.Fprintln(out, "  migrate   Universal stepwise schema migration engine for telemetry ledgers")
	fmt.Fprintln(out, "  version   Print version and build information")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Root Workspace Inspection Flags:")
	fmt.Fprintln(out, "  --workspaces <path>   Path to root workspaces directory containing agentic_metrics.json ledgers")
	fmt.Fprintln(out, "  --ticket-id <id>      Target ticket identifier to filter and report on")
	fmt.Fprintln(out, "  --list                List all discovered ticket workspaces and summary metrics")
	fmt.Fprintln(out, "  --anomalies           Display anomalies and health flags for target ticket")
	fmt.Fprintln(out, "  --append <path>       Path to markdown summary file to idempotently update with metrics table")
	fmt.Fprintln(out, "  --json-output <path>  Path to write computed metrics ledger as JSON")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Run 'agentic-lens <subcommand> --help' for subcommand-specific flags and options.")
}

func runInit(args []string) {
	initFlags := flag.NewFlagSet("init", flag.ExitOnError)
	projectNameFlag := initFlags.String("project-name", "", "Human-readable name of the project (required)")
	workspacesFlag := initFlags.String("workspaces", "workspaces", "Directory where ticket artifacts and session telemetry ledgers are stored")
	dbPathFlag := initFlags.String("db-path", "~/.local/share/opencode/opencode.db", "Path to local OpenCode SQLite database file")
	srcFoldersFlag := initFlags.String("src-folders", "", "Comma-separated list of monitored source code directories (required, e.g. web,backend)")
	ticketPrefixesFlag := initFlags.String("ticket-prefixes", "", "Comma-separated list of ticket prefix identifiers (required, e.g. AL,GEN,PO)")
	outputFlag := initFlags.String("output", ".agenticlens-config.yaml", "Destination path for generated configuration file")
	forceFlag := initFlags.Bool("force", false, "Overwrite existing configuration file without prompting")
	dryRunFlag := initFlags.Bool("dry-run", false, "Print generated configuration YAML to stdout without writing")

	if len(args) > 0 && (args[0] == "help" || args[0] == "-h" || args[0] == "--help") {
		initFlags.Usage()
		return
	}

	if err := initFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing init flags: %v\n", err)
		os.Exit(1)
	}

	projectName := strings.TrimSpace(*projectNameFlag)
	if projectName == "" {
		fmt.Fprintln(os.Stderr, "Error: missing required flag --project-name")
		os.Exit(1)
	}

	var srcFolders []string
	if strings.TrimSpace(*srcFoldersFlag) != "" {
		parts := strings.Split(*srcFoldersFlag, ",")
		for _, p := range parts {
			clean := strings.TrimSpace(p)
			if clean != "" {
				srcFolders = append(srcFolders, clean)
			}
		}
	}
	if len(srcFolders) == 0 {
		fmt.Fprintln(os.Stderr, "Error: missing required flag --src-folders")
		os.Exit(1)
	}

	ticketPrefixesVal := strings.TrimSpace(*ticketPrefixesFlag)

	var prefixes []domain.TicketPrefix
	if ticketPrefixesVal != "" {
		parts := strings.Split(ticketPrefixesVal, ",")
		for _, p := range parts {
			clean := strings.TrimSpace(p)
			if clean != "" {
				upper := strings.ToUpper(clean)
				prefixes = append(prefixes, domain.TicketPrefix{
					Prefix:      upper,
					Description: fmt.Sprintf("%s project development features and bugfixes", upper),
				})
			}
		}
	}
	if len(prefixes) == 0 {
		fmt.Fprintln(os.Stderr, "Error: missing required flag --ticket-prefixes")
		os.Exit(1)
	}

	yamlData, err := adaptersconfig.GenerateDefaultYAML(projectName, *workspacesFlag, *dbPathFlag, srcFolders, prefixes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating configuration: %v\n", err)
		os.Exit(1)
	}

	if *dryRunFlag {
		fmt.Print(string(yamlData))
		return
	}

	outputPath := *outputFlag
	if strings.TrimSpace(outputPath) == "" {
		outputPath = ".agenticlens-config.yaml"
	}

	if err := adaptersconfig.WriteConfigFile(outputPath, yamlData, *forceFlag); err != nil {
		if errors.Is(err, adaptersconfig.ErrFileExists) {
			fmt.Fprintf(os.Stderr, "Error: configuration file already exists at %s. Use --force to overwrite.\n", outputPath)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error writing configuration file %s: %v\n", outputPath, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully initialized AgenticLens configuration for '%s' at %s\n", projectName, outputPath)
}

func runVersion(args []string) {
	versionFlags := flag.NewFlagSet("version", flag.ExitOnError)
	jsonFlag := versionFlags.Bool("json", false, "Output version information in JSON format")
	shortFlag := versionFlags.Bool("short", false, "Output short version string only")

	if err := versionFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing version flags: %v\n", err)
		os.Exit(1)
	}

	info := domain.GetBuildInfo()

	if *jsonFlag {
		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling version JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
		return
	}

	if *shortFlag {
		fmt.Println(domain.FormatShortVersion(info))
		return
	}

	fmt.Println(domain.FormatBuildInfo(info))
}

func main() {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "help" || arg == "-h" || arg == "--help" || arg == "-help" {
			if len(os.Args) > 2 {
				switch os.Args[2] {
				case "init":
					runInit([]string{"--help"})
					return
				case "server":
					runServer([]string{"--help"})
					return
				case "generate":
					runGenerate([]string{"--help"})
					return
				case "backfill":
					runBackfill([]string{"--help"})
					return
				case "migrate":
					runMigrate([]string{"--help"})
					return
				case "version":
					runVersion([]string{"--help"})
					return
				}
			}
			flag.CommandLine.SetOutput(os.Stdout)
			printRootUsage()
			os.Exit(0)
		}

		switch arg {
		case "init":
			runInit(os.Args[2:])
			return
		case "version":
			runVersion(os.Args[2:])
			return
		case "-v", "--version", "-version":
			runVersion(nil)
			return
		case "server":
			runServer(os.Args[2:])
			return
		case "generate":
			runGenerate(os.Args[2:])
			return
		case "backfill":
			runBackfill(os.Args[2:])
			return
		case "migrate":
			runMigrate(os.Args[2:])
			return
		}
	}

	flag.Usage = printRootUsage

	ticketIDFlag := flag.String("ticket-id", "", "Target ticket identifier to filter and report on")
	workspacesFlag := flag.String("workspaces", "", "Path to root workspaces directory containing agentic_metrics.json ledgers")
	appendFlag := flag.String("append", "", "Path to markdown summary file to idempotently update with metrics table")
	jsonOutputFlag := flag.String("json-output", "", "Path to write computed metrics ledger as JSON")
	listFlag := flag.Bool("list", false, "List all discovered ticket workspaces and summary metrics")
	anomaliesFlag := flag.Bool("anomalies", false, "Display anomalies and health flags for target ticket")
	versionFlag := flag.Bool("v", false, "Print version information and exit")
	versionLongFlag := flag.Bool("version", false, "Print version information and exit")

	flag.Parse()

	if *versionFlag || *versionLongFlag {
		runVersion(nil)
		return
	}

	projCfg, err := adaptersconfig.Load("")
	if err != nil {
		if errors.Is(err, adaptersconfig.ErrConfigNotFound) {
			fmt.Fprintln(os.Stderr, "Project configuration not found. Please run 'agentic-lens init' to initialize the project.")
		} else {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		}
		os.Exit(1)
	}

	workspacesDir, err := adapters.ResolveWorkspacesDir(*workspacesFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: workspaces.dir is empty or unconfigured: configure 'workspaces.dir' in .agenticlens-config.yaml or provide --workspaces flag\n")
		os.Exit(1)
	}

	rawWorkspacesDir := strings.TrimSpace(*workspacesFlag)
	if rawWorkspacesDir == "" {
		rawWorkspacesDir = strings.TrimSpace(projCfg.Workspaces.Dir)
	}

	if !filepath.IsAbs(workspacesDir) {
		cfgPath, _ := adaptersconfig.FindConfigFile("")
		if cfgPath != "" {
			baseDir := filepath.Dir(cfgPath)
			candidate := filepath.Join(baseDir, workspacesDir)
			if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
				workspacesDir = candidate
			}
		}
	}

	fi, err := os.Stat(workspacesDir)
	if err != nil || !fi.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: workspaces directory %q does not exist. Cannot proceed.\n", rawWorkspacesDir)
		os.Exit(1)
	}
	report := adapters.ScanWorkspaces(workspacesDir)

	if len(report.Ledgers) == 0 {
		fmt.Fprintf(os.Stderr, "No valid agentic_metrics.json ledgers found in %s\n", workspacesDir)
		os.Exit(1)
	}

	if *listFlag {
		fmt.Printf("Found %d ticket workspaces in %s:\n\n", len(report.Ledgers), workspacesDir)
		header := fmt.Sprintf("%-15s %-12s %-10s %-12s %-10s %-16s %s", "Ticket ID", "Duration", "Cost", "Health", "Iterations", "Anomalies", "Title")
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
			anomSummary := formatAnomalyListSummary(ledger.Anomalies)
			title := "N/A"
			if ledger.TicketTitle != nil && *ledger.TicketTitle != "" {
				title = *ledger.TicketTitle
			}
			fmt.Printf("%-15s %-12s %-10s %-12s %-10s %-16s %s\n", tid, dur, cost, health, iters, anomSummary, title)
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

	if *anomaliesFlag {
		tid := "UNKNOWN"
		if targetLedger.TicketID != nil && *targetLedger.TicketID != "" {
			tid = *targetLedger.TicketID
		}
		printAnomalyDiagnostics(tid, targetLedger.Anomalies)
		return
	}

	tableMD := adapters.FormatMarkdownMetricsTable(*targetLedger, projCfg.RoleOrder)

	if *appendFlag != "" {
		if err := adapters.UpdateSummaryFile(*appendFlag, tableMD); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating summary file %s: %v\n", *appendFlag, err)
			os.Exit(1)
		}
		fmt.Printf("AgenticLens metrics table successfully updated in %s\n", *appendFlag)
	}

	if *jsonOutputFlag != "" {
		if err := adapters.SaveMetricsLedger(*jsonOutputFlag, *targetLedger); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON to %s: %v\n", *jsonOutputFlag, err)
			os.Exit(1)
		}
		fmt.Printf("AgenticLens computed ledger written to %s\n", *jsonOutputFlag)
	}

	if *appendFlag == "" && *jsonOutputFlag == "" {
		fmt.Println(tableMD)
	}
}

type invocationKey struct {
	id    string
	model string
}

type toolFactKey struct {
	sessionID string
	name      string
}

func mergeInvocations(cached []domain.InvocationFact, local []domain.InvocationFact) []domain.InvocationFact {
	merged := make([]domain.InvocationFact, len(cached))
	copy(merged, cached)

	indexMap := make(map[invocationKey]int)
	for i, inv := range merged {
		indexMap[invocationKey{id: inv.ID, model: inv.Model}] = i
	}

	for _, lInv := range local {
		k := invocationKey{id: lInv.ID, model: lInv.Model}
		if idx, exists := indexMap[k]; exists {
			merged[idx] = lInv
		} else {
			indexMap[k] = len(merged)
			merged = append(merged, lInv)
		}
	}

	return merged
}

func mergeTools(cached []domain.ToolFact, local []domain.ToolFact) []domain.ToolFact {
	merged := make([]domain.ToolFact, len(cached))
	copy(merged, cached)

	indexMap := make(map[toolFactKey]int)
	for i, t := range merged {
		indexMap[toolFactKey{sessionID: t.SessionID, name: t.Name}] = i
	}

	for _, lt := range local {
		key := toolFactKey{sessionID: lt.SessionID, name: lt.Name}
		if idx, exists := indexMap[key]; exists {
			merged[idx] = lt
		} else {
			indexMap[key] = len(merged)
			merged = append(merged, lt)
		}
	}

	return merged
}

type eventKey struct {
	timestamp string
	sessionID string
	turn      int64
	eventType domain.EventType
}

func mergeEvents(cached []domain.EventFact, local []domain.EventFact) []domain.EventFact {
	merged := make([]domain.EventFact, len(cached))
	copy(merged, cached)

	seen := make(map[eventKey]bool)
	for _, e := range merged {
		seen[eventKey{
			timestamp: e.Timestamp,
			sessionID: e.SessionID,
			turn:      e.Turn,
			eventType: e.EventType,
		}] = true
	}

	for _, le := range local {
		k := eventKey{
			timestamp: le.Timestamp,
			sessionID: le.SessionID,
			turn:      le.Turn,
			eventType: le.EventType,
		}
		if !seen[k] {
			seen[k] = true
			merged = append(merged, le)
		}
	}

	return merged
}

func mergeSessions(cached []domain.SessionFact, local []domain.SessionFact) []domain.SessionFact {
	byID := make(map[string]domain.SessionFact, len(cached)+len(local))
	for _, session := range cached {
		byID[session.ID] = session
	}
	for _, session := range local {
		byID[session.ID] = session
	}
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	merged := make([]domain.SessionFact, 0, len(ids))
	for _, id := range ids {
		merged = append(merged, byID[id])
	}
	return merged
}

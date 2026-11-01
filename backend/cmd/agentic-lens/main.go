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
	"agenticlens/internal/adapters/capture"
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
	sourceFlag := genFlags.String("source", "", "Telemetry harness to read: opencode or codex (defaults to local configuration, then opencode)")
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
	captureDirFlag := genFlags.String("capture-dir", capture.DefaultDirectory, "Local lifecycle registry to reconcile before final publication")
	allowEmptyFlag := genFlags.Bool("allow-empty", false, "Allow empty metrics generation even if no root session or invocations are resolved (test escape hatch)")

	if err := genFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing generate flags: %v\n", err)
		os.Exit(1)
	}
	// All writers take ownership before loading the canonical baseline or reading sources.
	// This prevents an earlier source view from publishing after a later generation.
	if *jsonOutputFlag != "" {
		lockCtx, lockCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		publicationLock, lockErr := capture.AcquirePublication(lockCtx, *jsonOutputFlag)
		lockCancel()
		if lockErr != nil {
			fmt.Fprintln(os.Stderr, "Error acquiring ledger publication ownership:", lockErr)
			os.Exit(1)
		}
		defer func() {
			if closeErr := publicationLock.Close(); closeErr != nil {
				fmt.Fprintln(os.Stderr, "Publication unlock:", closeErr)
			}
		}()
	}
	activeSource := strings.ToLower(strings.TrimSpace(*sourceFlag))

	var cachedFacts *adapters.ExistingLedgerFacts
	if *jsonOutputFlag != "" {
		loaded, err := adapters.LoadExistingFactLedger(*jsonOutputFlag)
		if err != nil {
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
					if f == "code_metrics.source_roots" {
						fmt.Fprintf(os.Stderr, "Error: source roots are empty or unconfigured: configure required 'code_metrics.source_roots' in .agenticlens-config.yaml\n")
						os.Exit(1)
					}
				}
			}
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		}
		os.Exit(1)
	}
	activeSource, err = adaptersconfig.ResolveSource(*sourceFlag, projCfg.Telemetry.DefaultSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.\n", err)
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
	var skills []domain.SkillFact
	var taskLinks []domain.TaskLinkFact
	var fallbackCode *domain.CodeFacts

	if cachedFacts != nil {
		knownThreadIDs = append(knownThreadIDs, cachedFacts.RootThreadIDs...)
		sessions = append(sessions, cachedFacts.Sessions...)
		skills = append(skills, cachedFacts.Skills...)
		taskLinks = append(taskLinks, cachedFacts.TaskLinks...)
		for _, s := range cachedFacts.Sessions {
			userBySessionID[s.ID] = s.User
		}
		fallbackCode = cachedFacts.Code
		if cachedFacts.Migration != nil && strings.Contains(cachedFacts.Migration.Notes, "Removed ") {
			fmt.Fprintf(os.Stderr, "WARNING: %s\n", cachedFacts.Migration.Notes)
		}
	}

	baselineSkills := append([]domain.SkillFact{}, skills...)
	baselineTaskLinks := append([]domain.TaskLinkFact{}, taskLinks...)
	var freshSessions []domain.SessionFact
	captureResult, captureErr := reconcileCaptureGeneration(*captureDirFlag, activeTicketID)
	if captureErr != nil {
		fmt.Fprintf(os.Stderr, "Error reconciling capture registry; no coverage report published: %v\n", captureErr)
		os.Exit(1)
	}
	if precedenceErr := filterCapturePrecedence(sessions, captureResult); precedenceErr != nil {
		fmt.Fprintln(os.Stderr, "Capture checkpoint comparison:", precedenceErr)
		os.Exit(1)
	}
	registeredOnly := hasBoundRegistrations(captureResult) && strings.TrimSpace(*nonceFlag) == "" && strings.TrimSpace(*sessionIDFlag) == ""
	if registeredOnly {
		freshSessions = captureResult.RefreshedSessions
	} else if activeSource == "codex" {
		builtinCodexHome := ""
		if home, homeErr := os.UserHomeDir(); homeErr == nil {
			builtinCodexHome = filepath.Join(home, ".codex")
		}
		codexHome := adaptersconfig.ResolveCodexHome(*codexHomeFlag, projCfg.Telemetry.Sources.CodexHome, os.Getenv("CODEX_HOME"), builtinCodexHome)
		rolloutPaths, discoverErr := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{
			CodexHome: codexHome, Nonce: *nonceFlag, SessionID: *sessionIDFlag,
		})
		if discoverErr != nil {
			fmt.Fprintf(os.Stderr, "Error discovering selected Codex source at %s: %v. Check telemetry.sources.codex.home or --codex-home.\n", codexHome, discoverErr)
			os.Exit(1)
		}
		if len(rolloutPaths) == 0 {
			fmt.Fprintf(os.Stderr, "Error: no matching rollouts found in selected Codex source at %s. Check telemetry.sources.codex.home, --codex-home, and the required --nonce or --session-id.\n", codexHome)
			os.Exit(1)
		}
		transcripts := make([]codex.Transcript, 0, len(rolloutPaths))
		for _, path := range rolloutPaths {
			transcript, parseErr := codex.ParseRolloutPath(path)
			if parseErr != nil {
				fmt.Fprintf(os.Stderr, "Error reading selected Codex source rollout %s: %v. Check telemetry.sources.codex.home or --codex-home.\n", path, parseErr)
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
		freshCodex, factErr := combined.SessionFacts()
		if factErr != nil {
			fmt.Fprintf(os.Stderr, "Error grouping Codex session facts: %v\n", factErr)
			os.Exit(1)
		}
		if ownerErr := validateCaptureSelection(*captureDirFlag, activeTicketID, domain.SourceCodex, codexHome, freshCodex); ownerErr != nil {
			fmt.Fprintln(os.Stderr, "Capture ownership conflict:", ownerErr)
			os.Exit(1)
		}
		resolveCodexSessionUsers(freshCodex, userBySessionID, currentUser)
		freshSessions = freshCodex
		if strings.TrimSpace(*startedAtFlag) == "" && combined.StartedAt != "" {
			activeStartedAt = &combined.StartedAt
		}
		if strings.TrimSpace(*completedAtFlag) == "" && combined.CompletedAt != "" {
			activeCompletedAt = &combined.CompletedAt
		}
		fmt.Fprintf(os.Stderr, "Discovered %d Codex thread rollout(s) for session %s.\n", len(rolloutPaths), combined.ThreadID)
	} else {
		dbPath := adaptersconfig.ResolveOpenCodePath(*dbPathFlag, projCfg.Telemetry.Sources.OpenCodeDatabasePath, projCfg.Database.Path)
		dbClient, dbErr := opencode.OpenDB(dbPath)
		if dbErr == nil {
			defer dbClient.Close()

			rootID, resolveErr := opencode.ResolveActiveRootSession(
				dbClient.DB(),
				*nonceFlag,
				activeTicketID,
				*sessionIDFlag,
			)
			if resolveErr != nil && !errors.Is(resolveErr, opencode.ErrSessionNotFound) {
				fmt.Fprintf(os.Stderr, "Error querying selected OpenCode source at %s: %v. Check telemetry.sources.opencode.database_path or --db-path.\n", dbPath, resolveErr)
				os.Exit(1)
			}
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
				extractedSessions, dbStarted, dbCompleted, extractErr := opencode.ExtractSessions(
					dbClient.DB(), knownThreadIDs, currentUser, userBySessionID,
				)
				if extractErr != nil {
					fmt.Fprintf(os.Stderr, "Error reading selected OpenCode source at %s: %v. Check the database and telemetry.sources.opencode.database_path.\n", dbPath, extractErr)
					os.Exit(1)
				}
				if len(extractedSessions) > 0 {
					if ownerErr := validateCaptureSelection(*captureDirFlag, activeTicketID, domain.SourceOpenCode, dbPath, extractedSessions); ownerErr != nil {
						fmt.Fprintln(os.Stderr, "Capture ownership conflict:", ownerErr)
						os.Exit(1)
					}
					freshSessions = extractedSessions
					skills, taskLinks = opencode.ExtractSkillTaskFacts(extractedSessions)
					if strings.TrimSpace(*startedAtFlag) == "" && dbStarted != nil {
						activeStartedAt = dbStarted
					}
					if strings.TrimSpace(*completedAtFlag) == "" && dbCompleted != nil {
						activeCompletedAt = dbCompleted
					}
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error opening selected OpenCode source at %s: %v. Check telemetry.sources.opencode.database_path or --db-path.\n", dbPath, dbErr)
			os.Exit(1)
		}
	}

	if captureResult != nil {
		if err := capture.ValidateSessionMerge(sessions, captureResult.Sessions); err != nil {
			fmt.Fprintln(os.Stderr, "Capture identity collision:", err)
			os.Exit(1)
		}
		// Retained local snapshots are fallback only; they never replace canonical or explicit facts.
		sessions = fillMissingCaptureSessions(sessions, captureResult.Sessions)
		// Explicit extraction runs after capture and has precedence for a duplicate refreshed owner.
		freshSessions = mergeSessions(captureResult.RefreshedSessions, freshSessions)
		for _, session := range captureResult.Sessions {
			knownThreadIDs = append(knownThreadIDs, session.ID)
		}
		var mergeErr error
		skills, mergeErr = mergeCaptureRows(skills, captureResult.RefreshedSkills)
		if mergeErr != nil {
			fmt.Fprintln(os.Stderr, "Capture skill merge:", mergeErr)
			os.Exit(1)
		}
		taskLinks, mergeErr = mergeCaptureRows(taskLinks, captureResult.RefreshedTaskLinks)
		if mergeErr != nil {
			fmt.Fprintln(os.Stderr, "Capture task merge:", mergeErr)
			os.Exit(1)
		}
	}
	if len(freshSessions) > 0 {
		if err := capture.ValidateSessionMerge(sessions, freshSessions); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		var checkpointErr error
		sourceRefreshes := freshSessions
		freshSessions, checkpointErr = rejectOlderExplicitOwners(sessions, freshSessions)
		if checkpointErr != nil {
			fmt.Fprintln(os.Stderr, "Source checkpoint comparison:", checkpointErr)
			os.Exit(1)
		}
		if len(freshSessions) < len(sourceRefreshes) {
			// v5 auxiliary rows have no owner ID; retain baseline on partial unordered
			// refresh rather than publish rows from a rejected source snapshot.
			skills = baselineSkills
			taskLinks = baselineTaskLinks
			acceptedIDs := map[string]bool{}
			for _, owner := range freshSessions {
				acceptedIDs[owner.ID] = true
			}
			if captureResult != nil {
				for i := range captureResult.Registrations {
					if !acceptedIDs[captureResult.Registrations[i].Identity.SessionID] {
						captureResult.Registrations[i].State = domain.CapturePending
						captureResult.Registrations[i].Detail = "Selected source checkpoint could not establish newer canonical facts; accepted baseline preserved"
					}
				}
				captureResult.Limits = append(captureResult.Limits, "Explicit source refresh had unordered/older checkpoints; retained canonical source facts and ownerless auxiliary rows.")
			}
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
	for _, s := range sessions {
		if s.ParentID == nil {
			for _, inv := range s.Invocations {
				totalSDLC += inv.UserIterations
			}
		}
	}

	codeFacts.SDLCIterations = totalSDLC
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
		TicketID:      ticketIDPtr,
		TicketTitle:   activeTicketTitle,
		TicketType:    activeTicketType,
		StartedAt:     activeStartedAt,
		CompletedAt:   activeCompletedAt,
		UpdatedAt:     updatedAtISO,
		Code:          codeFacts,
		Sessions:      sessions,
		Skills:        skills,
		TaskLinks:     taskLinks,
	}

	metricsLedger := domain.DeriveMetricsLedger(factLedger, projCfg)
	tableMD := adapters.FormatMarkdownMetricsTable(metricsLedger, projCfg.RoleOrder)
	if captureResult != nil {
		tableMD += capture.CoverageSummary(*captureResult)
	}

	if *jsonOutputFlag != "" {
		persistedFactLedger, pruneErr := domain.PersistAnomalySummaries(factLedger)
		if pruneErr != nil {
			fmt.Fprintf(os.Stderr, "Error preparing session-owned telemetry: %v\n", pruneErr)
			os.Exit(1)
		}
		if err := saveGenerationLedger(*jsonOutputFlag, persistedFactLedger); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON to %s: %v\n", *jsonOutputFlag, err)
			os.Exit(1)
		}
		if captureResult != nil {
			if err := publishCaptureReport(*jsonOutputFlag, *captureResult); err != nil {
				fmt.Fprintf(os.Stderr, "Ledger written but capture report publication failed; regenerate to recover: %v\n", err)
				os.Exit(1)
			}
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

func resolveCodexSessionUsers(fresh []domain.SessionFact, existingUsers map[string]string, gitUser string) {
	for i := range fresh {
		fresh[i].User = git.ResolveSessionUser(fresh[i].User, existingUsers[fresh[i].ID], gitUser)
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
		dbPath = strings.TrimSpace(projCfg.Telemetry.Sources.OpenCodeDatabasePath)
	}
	if dbPath == "" {
		dbPath = strings.TrimSpace(projCfg.Database.Path)
	}
	if dbPath == "" {
		if defPath, err := opencode.DefaultDBPath(); err == nil {
			dbPath = defPath
		}
	}

	cleanDBPath := adaptersconfig.ExpandHomePath(dbPath)

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
		fmt.Fprintf(out, "  --source <harness>     Required telemetry harness to backfill: opencode, codex, or both\n")
		fmt.Fprintf(out, "  --codex-home <path>    Path to CODEX_HOME (defaults to $CODEX_HOME or ~/.codex)\n")
		fmt.Fprintf(out, "  --workspaces <path>    Path to root workspaces directory containing agentic_metrics.json ledgers\n")
		fmt.Fprintf(out, "  --ticket-id <id>       Target single ticket identifier to filter backfill operation\n")
		fmt.Fprintf(out, "  --recompute-git        Recompute Git code facts and SDLC iterations from repository history\n")
		fmt.Fprintf(out, "  --code-ref <ref>        Required with --recompute-git; branch or commit to compare with configured base\n")
		fmt.Fprintf(out, "  --recompute-db         Refresh complete selected-source snapshots without --force; retain unavailable history\n")
		fmt.Fprintf(out, "  --migrate-schema       Upgrade telemetry ledger schema versions to latest supported version\n")
		fmt.Fprintf(out, "  --dry-run              Preview planned migrations without writing changes to disk\n")
		fmt.Fprintf(out, "  --update-summaries     Idempotently update SUMMARY.md markdown metrics tables\n")
		fmt.Fprintf(out, "  --db-path <path>       Explicit path to OpenCode SQLite database file\n")
		fmt.Fprintf(out, "  --force                Repeat completed-ledger processing; never bypass preservation or enable fallbacks\n")
		fmt.Fprintf(out, "  --allow-fallbacks      Allow heuristic session discovery fallbacks (title match, workspace commit range) when exact session IDs are not found in DB\n")
	}
	if len(args) > 0 && (args[0] == "help" || args[0] == "-h" || args[0] == "--help" || args[0] == "-help") {
		backfillFlags.SetOutput(os.Stdout)
		backfillFlags.Usage()
		return
	}
	workspacesFlag := backfillFlags.String("workspaces", "", "Path to root workspaces directory containing agentic_metrics.json ledgers")
	sourceFlag := backfillFlags.String("source", "", "Telemetry harness to backfill: opencode, codex, or both (required)")
	codexHomeFlag := backfillFlags.String("codex-home", "", "Path to CODEX_HOME (defaults to $CODEX_HOME or ~/.codex)")
	ticketIDFlag := backfillFlags.String("ticket-id", "", "Target single ticket identifier to filter backfill operation")
	recomputeGitFlag := backfillFlags.Bool("recompute-git", false, "Recompute Git code facts and SDLC iterations from repository history")
	codeRefFlag := backfillFlags.String("code-ref", "", "Git branch or commit to compare with the configured base when recomputing code facts")
	recomputeDBFlag := backfillFlags.Bool("recompute-db", false, "Refresh complete selected-source snapshots without --force; retain unavailable history")
	migrateSchemaFlag := backfillFlags.Bool("migrate-schema", false, "Upgrade telemetry ledger schema versions to latest supported version")
	dryRunFlag := backfillFlags.Bool("dry-run", false, "Preview planned migrations without writing changes to disk")
	updateSummariesFlag := backfillFlags.Bool("update-summaries", false, "Idempotently update SUMMARY.md markdown metrics tables")
	dbPathFlag := backfillFlags.String("db-path", "", "Explicit path to OpenCode SQLite database file")
	forceFlag := backfillFlags.Bool("force", false, "Repeat completed-ledger processing; never bypass preservation or enable fallbacks")
	allowFallbacksFlag := backfillFlags.Bool("allow-fallbacks", false, "Allow heuristic session discovery fallbacks (title match, workspace commit range) when exact session IDs are not found in DB")

	if err := backfillFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing backfill flags: %v\n", err)
		os.Exit(1)
	}
	cleanSource := strings.ToLower(strings.TrimSpace(*sourceFlag))
	if cleanSource == "" {
		fmt.Fprintln(os.Stderr, "Error: --source flag is required for backfill. Specify 'opencode', 'codex', or 'both'.")
		os.Exit(1)
	}
	if cleanSource != "opencode" && cleanSource != "codex" && cleanSource != "both" && cleanSource != "all" {
		fmt.Fprintf(os.Stderr, "Error: invalid --source %q. Specify 'opencode', 'codex', or 'both'.\n", *sourceFlag)
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
	resolvedDBPath := strings.TrimSpace(*dbPathFlag)
	resolvedCodexHome := strings.TrimSpace(*codexHomeFlag)
	cfg, loadErr := adaptersconfig.Load("")
	if loadErr != nil && !errors.Is(loadErr, adaptersconfig.ErrConfigNotFound) {
		fmt.Fprintf(os.Stderr, "Error loading project configuration for backfill: %v\n", loadErr)
		os.Exit(1)
	}
	if loadErr == nil {
		if resolvedDBPath == "" {
			resolvedDBPath = strings.TrimSpace(cfg.Telemetry.Sources.OpenCodeDatabasePath)
		}
		if resolvedDBPath == "" {
			resolvedDBPath = strings.TrimSpace(cfg.Database.Path)
		}
		if resolvedCodexHome == "" {
			resolvedCodexHome = strings.TrimSpace(cfg.Telemetry.Sources.CodexHome)
		}
	}
	resolvedDBPath = adaptersconfig.ExpandHomePath(resolvedDBPath)
	resolvedCodexHome = adaptersconfig.ExpandHomePath(resolvedCodexHome)

	opts := backfill.BackfillOptions{
		WorkspacesDir:   *workspacesFlag,
		TicketID:        *ticketIDFlag,
		Source:          cleanSource,
		CodexHome:       resolvedCodexHome,
		RecomputeGit:    *recomputeGitFlag,
		CodeRef:         strings.TrimSpace(*codeRefFlag),
		RecomputeDB:     *recomputeDBFlag,
		MigrateSchema:   *migrateSchemaFlag,
		DryRun:          *dryRunFlag,
		UpdateSummaries: *updateSummariesFlag,
		DBPath:          resolvedDBPath,
		Force:           *forceFlag,
		AllowFallbacks:  *allowFallbacksFlag,
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
	fmt.Printf("Recompute Source Facts:       %t\n", opts.RecomputeDB)
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
			notes = strings.TrimSpace(fmt.Sprintf("%s; pending: [%s]", notes, strings.Join(res.PendingSteps, ", ")))
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
	fmt.Printf("Refresh diagnostics: Incomplete: %d | Retained skill/task snapshots: %d (session trees may be newer)\n", report.RefreshIncomplete, report.CollectionsRetained)

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
	header := fmt.Sprintf("%-10s %-28s %s", "Severity", "Rule ID", "Summary")
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))
	for _, a := range anomalies {
		title, description := domain.AnomalyDisplayText(a)
		fmt.Printf("%-10s %-28s %s — %s\n", string(a.Severity), string(a.RuleID), title, description)
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
	fmt.Fprintln(out, "  capture   Install hooks, bind exact sessions, reconcile local snapshots and diagnose coverage")
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
	sharedDestination, sharedPathErr := filepath.Abs(filepath.Clean(*outputFlag))
	localDestination, localPathErr := filepath.Abs(".agenticlens-config.local.yaml")
	if sharedPathErr == nil && localPathErr == nil && sharedDestination == localDestination {
		fmt.Fprintln(os.Stderr, "Error: --output cannot target .agenticlens-config.local.yaml; shared and local configuration outputs must be distinct.")
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
	localYAML := []byte("# Machine-local telemetry source settings. Do not commit this file.\ntelemetry:\n  default_source: opencode\n  sources:\n    opencode:\n      database_path: ~/.local/share/opencode/opencode.db\n    codex:\n      home: ~/.codex\n")

	if *dryRunFlag {
		fmt.Printf("# .agenticlens-config.yaml\n%s\n# .agenticlens-config.local.yaml\n%s", string(yamlData), string(localYAML))
		return
	}

	outputPath := *outputFlag
	if strings.TrimSpace(outputPath) == "" {
		outputPath = ".agenticlens-config.yaml"
	}

	localPath := ".agenticlens-config.local.yaml"
	for _, path := range []string{outputPath, localPath} {
		if _, err := os.Stat(path); err == nil && !*forceFlag {
			fmt.Fprintf(os.Stderr, "Error: configuration file already exists at %s. Use --force to overwrite.\n", path)
			os.Exit(1)
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "Error checking configuration file %s: %v\n", path, err)
			os.Exit(1)
		}
	}
	if err := writeConfigPair(outputPath, yamlData, localPath, localYAML, *forceFlag); err != nil {
		if errors.Is(err, adaptersconfig.ErrFileExists) {
			fmt.Fprintf(os.Stderr, "Error writing configuration files: %v. Use --force to overwrite.\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error writing configuration files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully initialized AgenticLens configuration for '%s' at %s and %s\n", projectName, outputPath, localPath)
}

func writeConfigPair(firstPath string, firstData []byte, secondPath string, secondData []byte, force bool) (resultErr error) {
	type stagedFile struct {
		target, temp, backup string
		installed            bool
	}
	items := []struct {
		path string
		data []byte
	}{{firstPath, firstData}, {secondPath, secondData}}
	// Validate the complete pair before moving either existing configuration.
	for _, item := range items {
		info, err := os.Stat(item.path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("checking %s: %w", item.path, err)
		}
		if info.IsDir() {
			return fmt.Errorf("target %s is a directory", item.path)
		}
		if !force {
			return &adaptersconfig.FileExistsError{Path: item.path}
		}
	}
	staged := make([]stagedFile, 0, len(items))
	committed := false
	defer func() {
		for i := len(staged) - 1; i >= 0; i-- {
			item := staged[i]
			if !committed {
				if item.installed {
					if err := os.Remove(item.target); err != nil && !errors.Is(err, os.ErrNotExist) {
						resultErr = errors.Join(resultErr, fmt.Errorf("removing replacement %s: %w", item.target, err))
					}
				}
				if item.backup != "" {
					if err := os.Rename(item.backup, item.target); err != nil {
						resultErr = errors.Join(resultErr, fmt.Errorf("restoring %s from %s: %w", item.target, item.backup, err))
					}
				}
			} else if item.backup != "" {
				if err := os.Remove(item.backup); err != nil {
					resultErr = errors.Join(resultErr, fmt.Errorf("removing backup %s: %w", item.backup, err))
				}
			}
			if err := os.Remove(item.temp); err != nil && !errors.Is(err, os.ErrNotExist) {
				resultErr = errors.Join(resultErr, fmt.Errorf("removing staged file %s: %w", item.temp, err))
			}
		}
	}()
	for _, item := range items {
		if err := os.MkdirAll(filepath.Dir(item.path), 0755); err != nil {
			return fmt.Errorf("preparing directory for %s: %w", item.path, err)
		}
		file, err := os.CreateTemp(filepath.Dir(item.path), ".agenticlens-config-*.tmp")
		if err != nil {
			return fmt.Errorf("staging %s: %w", item.path, err)
		}
		staged = append(staged, stagedFile{target: item.path, temp: file.Name()})
		_, writeErr := file.Write(item.data)
		if err := errors.Join(writeErr, file.Close()); err != nil {
			return fmt.Errorf("staging %s: %w", item.path, err)
		}
	}
	for i := range staged {
		item := &staged[i]
		info, err := os.Stat(item.target)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("checking %s: %w", item.target, err)
		}
		if info.IsDir() {
			return fmt.Errorf("target %s is a directory", item.target)
		}
		if !force {
			return &adaptersconfig.FileExistsError{Path: item.target}
		}
		backup, err := os.CreateTemp(filepath.Dir(item.target), ".agenticlens-config-backup-*.tmp")
		if err != nil {
			return fmt.Errorf("preparing backup for %s: %w", item.target, err)
		}
		backupPath := backup.Name()
		if err := errors.Join(backup.Close(), os.Remove(backupPath)); err != nil {
			return fmt.Errorf("preparing backup for %s: %w", item.target, err)
		}
		if err := os.Rename(item.target, backupPath); err != nil {
			return fmt.Errorf("backing up %s: %w", item.target, err)
		}
		item.backup = backupPath
	}
	for i := range staged {
		item := &staged[i]
		if err := os.Rename(item.temp, item.target); err != nil {
			return fmt.Errorf("replacing %s: %w", item.target, err)
		}
		item.installed = true
	}
	committed = true
	return nil
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
				case "capture":
					runCapture([]string{"status", "--help"})
					return
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
		case "capture":
			runCapture(os.Args[2:])
			return
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
			cost := adapters.FormatCostSummary(domain.EffectiveCostSummary(ledger.Totals.CostSummary, ledger.Totals.CostUSD, ledger.Totals.Invocations))
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

func mergeSessions(cached []domain.SessionFact, local []domain.SessionFact) []domain.SessionFact {
	byID := make(map[string]domain.SessionFact, len(cached)+len(local))
	for _, session := range cached {
		byID[session.ID] = session
	}
	// Each extracted session is authoritative, including repeated or empty invocation rows.
	for _, session := range local {
		if existing, ok := byID[session.ID]; ok {
			if session.Source == "" && existing.Source != "" {
				session.Source = existing.Source
			}
		}
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

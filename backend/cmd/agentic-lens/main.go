package main

import (
	"bytes"
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
	"strings"
	"syscall"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/backfill"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/codex"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/filesystem"
	"agenticlens/internal/adapters/git"
	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/adapters/live"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/adapters/web"
	"agenticlens/internal/domain"
)

func runGenerate(args []string) {
	snapshotCutoff := time.Now().UTC()
	genFlags := flag.NewFlagSet("generate", flag.ExitOnError)

	var cliTags tagFlags
	genFlags.Var(&cliTags, "tag", "Ticket tag KEY=VALUE (repeatable; CLI overrides local and shared per key)")
	ticketIDFlag := genFlags.String("ticket-id", "", "Target ticket identifier to lock session resolution to")
	codeRefFlag := genFlags.String("code-ref", "HEAD", "Git branch or commit to measure against the configured base (defaults to HEAD; committed changes only)")
	sourceFlag := genFlags.String("source", "", "Telemetry harness: explicit value or selected local telemetry.default_source")
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
	codexHomeFlag := genFlags.String("codex-home", "", "Explicit Codex source store; otherwise selected local telemetry.sources.codex.home")
	anomaliesFlag := genFlags.Bool("anomalies", false, "Display anomalies and health flags; generation.metrics.anomalies=false reports disabled and never re-enables analysis")
	captureDirFlag := genFlags.String("capture-dir", "", "Local lifecycle registry to reconcile before final publication")
	autoFlag := genFlags.Bool("capture-auto", false, "Native lifecycle publication (registered owners only)")
	maxSessionsFlag := genFlags.Int("max-sessions", 0, "Explicit positive capture session budget; otherwise selected local policy")
	maxBytesFlag := genFlags.Int64("max-source-bytes", 0, "Explicit positive capture source-byte budget; otherwise selected local policy")

	defaultUsage := genFlags.Usage
	genFlags.Usage = func() {
		defaultUsage()
		fmt.Fprintln(genFlags.Output(), "Checkpoint feedback: generation.feedback.provisional_threshold_seconds must be explicit; 0 disables human stderr diagnostics. --session-id selects an extraction root, not a verified caller.")
	}
	if len(args) > 0 && args[0] == "help" {
		genFlags.Usage()
		return
	}
	if err := genFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing generate flags: %v\n", err)
		os.Exit(1)
	}
	project, projectErr := os.Getwd()
	if projectErr != nil {
		fmt.Fprintln(os.Stderr, projectErr)
		os.Exit(1)
	}
	project, projectErr = adaptersconfig.ProjectRoot(project)
	if projectErr != nil {
		fmt.Fprintln(os.Stderr, projectErr)
		os.Exit(1)
	}
	projCfg, err := adaptersconfig.Load(project)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading configuration:", err)
		os.Exit(1)
	}
	projCfg.Tags = domain.MergeTags(projCfg.Tags, map[string]string(cliTags))
	if err := adaptersconfig.ValidateGenerationPolicy(projCfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := adaptersconfig.ValidateContextPolicy(projCfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !*autoFlag {
		if err := adaptersconfig.ValidateFeedbackPolicy(projCfg); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if projCfg.Generation.Metrics.Anomalies {
		if err := adaptersconfig.ValidateAnomalyPolicy(projCfg); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	provided := map[string]bool{}
	genFlags.Visit(func(f *flag.Flag) { provided[f.Name] = true })
	for _, choice := range []struct{ flag, value string }{{"ticket-type", *ticketTypeFlag}, {"ticket-id", *ticketIDFlag}, {"ticket-title", *ticketTitleFlag}, {"nonce", *nonceFlag}, {"session-id", *sessionIDFlag}, {"capture-dir", *captureDirFlag}, {"json-output", *jsonOutputFlag}} {
		if provided[choice.flag] && strings.TrimSpace(choice.value) == "" {
			fmt.Fprintf(os.Stderr, "--%s must be nonblank when supplied\n", choice.flag)
			os.Exit(1)
		}
	}
	if provided["source"] && strings.TrimSpace(*sourceFlag) == "" {
		fmt.Fprintln(os.Stderr, "--source must be nonblank when supplied")
		os.Exit(1)
	}
	if provided["db-path"] && strings.TrimSpace(*dbPathFlag) == "" {
		fmt.Fprintln(os.Stderr, "--db-path must be nonblank when supplied")
		os.Exit(1)
	}
	if provided["codex-home"] && strings.TrimSpace(*codexHomeFlag) == "" {
		fmt.Fprintln(os.Stderr, "--codex-home must be nonblank when supplied")
		os.Exit(1)
	}
	if *jsonOutputFlag != "" {
		expanded, pathErr := adaptersconfig.ExpandHomePath(*jsonOutputFlag)
		if pathErr != nil {
			fmt.Fprintln(os.Stderr, pathErr)
			os.Exit(1)
		}
		physical, pathErr := filesystem.PhysicalPath(expanded)
		if pathErr != nil {
			fmt.Fprintln(os.Stderr, pathErr)
			os.Exit(1)
		}
		*jsonOutputFlag = physical
	}
	selectedSource, err := adaptersconfig.ResolveSource(*sourceFlag, projCfg.Telemetry.DefaultSource)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	effectiveSummary, err := resolveGenerationSummaryPath(project, projCfg, *appendFlag, provided["append"], *jsonOutputFlag, *ticketIDFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	*appendFlag = effectiveSummary
	admissionCtx, admissionCancel := context.WithTimeout(context.Background(), 2*time.Second)
	admission, admissionErr := adaptersconfig.AcquireGenerationAdmission(admissionCtx, project)
	defer admissionCancel()
	if admissionErr != nil {
		if errors.Is(admissionErr, adaptersconfig.ErrConfigNotFound) {
			fmt.Fprintln(os.Stderr, "Project configuration not found. Please run 'agentic-lens init' to initialize the project.")
		} else {
			fmt.Fprintln(os.Stderr, "Capture configuration admission:", admissionErr)
		}
		os.Exit(1)
	}
	defer func() {
		if admission != nil {
			if err := admission.Close(); err != nil {
				fmt.Fprintln(os.Stderr, "Capture configuration unlock:", err)
			}
		}
	}()
	policy, policyErr := admission.ResolveGenerationCapture(*captureDirFlag)
	if *autoFlag {
		policy, policyErr = admission.ResolveCapture(*captureDirFlag)
	}
	if policyErr != nil {
		if errors.Is(policyErr, adaptersconfig.ErrConfigNotFound) {
			fmt.Fprintln(os.Stderr, "Project configuration not found. Please run 'agentic-lens init' to initialize the project.")
		} else {
			fmt.Fprintln(os.Stderr, policyErr)
		}
		os.Exit(1)
	}
	if *autoFlag && !policy.Enabled {
		fmt.Fprintln(os.Stderr, "Automatic publication disabled by selected local policy")
		os.Exit(1)
	}
	if provided["max-sessions"] {
		if *maxSessionsFlag < 1 {
			fmt.Fprintln(os.Stderr, "--max-sessions must be a positive integer")
			os.Exit(1)
		}
		policy.MaxSessions = *maxSessionsFlag
		if policy.Settings != nil {
			policy.Settings["max_sessions"] = true
		}
	}
	if provided["max-source-bytes"] {
		if *maxBytesFlag < 1 {
			fmt.Fprintln(os.Stderr, "--max-source-bytes must be a positive integer")
			os.Exit(1)
		}
		policy.MaxSourceBytes = *maxBytesFlag
		if policy.Settings != nil {
			policy.Settings["max_source_bytes"] = true
		}
	}
	if err := errors.Join(policy.ValidatePass(), policy.ValidateBudgets()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	resolvedState, stateErr := capture.ProjectState(policy.StateDir, project)
	if stateErr != nil {
		fmt.Fprintln(os.Stderr, stateErr)
		os.Exit(1)
	}
	// Only explicit manual recovery may use independent healthy state when
	// configured coordinate mappings are unsupported. Other guards still apply.
	if e := checkCaptureInstallationScope(project, resolvedState, *captureDirFlag != "", *captureDirFlag != "" && !*autoFlag); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	if *autoFlag {
		enabled, e := automaticCaptureEnabled(project, resolvedState)
		if e != nil || !enabled {
			fmt.Fprintln(os.Stderr, "Automatic publication disabled or unavailable:", e)
			os.Exit(1)
		}
	}

	// Materialize the selected scope while configuration admission is held.
	// Subsequent path-only maintenance must observe this database and refuse
	// detaching it, so no global config lease spans source parsing/publication.
	admissionStore, admissionStoreErr := capture.OpenProjectContext(admissionCtx, resolvedState, project)
	if admissionStoreErr != nil {
		fmt.Fprintln(os.Stderr, "Capture state admission:", admissionStoreErr)
		os.Exit(1)
	}
	admissionCloseErr := errors.Join(admissionStore.Close(), admission.Close())
	admission = nil
	admissionCancel()
	if admissionCloseErr != nil {
		fmt.Fprintln(os.Stderr, "Capture state admission:", admissionCloseErr)
		os.Exit(1)
	}
	*captureDirFlag = resolvedState
	destinations := []string{*jsonOutputFlag, *appendFlag}
	for _, dest := range destinations {
		if dest != "" {
			if e := capture.ValidateDestination(project, dest); e != nil {
				fmt.Fprintln(os.Stderr, e)
				os.Exit(1)
			}
		}
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
	if *jsonOutputFlag != "" {
		recovered, e := adapters.RecoverPublicationWithLegacyReport(project, *jsonOutputFlag, authorizedPublicationPaths(*jsonOutputFlag, *appendFlag))
		if e != nil {
			fmt.Fprintln(os.Stderr, "Interrupted publication recovery failed:", e)
			os.Exit(1)
		}
		if recovered {
			fmt.Fprintln(os.Stderr, "Recovered interrupted publication: restored prior accepted artifact set before source reads")
		}
	}
	activeSource := selectedSource

	var cachedFacts *adapters.ExistingLedgerFacts
	if *jsonOutputFlag != "" {
		loaded, err := adapters.LoadOwnedFactLedger(*jsonOutputFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading existing fact ledger: %v\n", err)
			os.Exit(1)
		}
		if loaded != nil {
			cachedFacts = loaded
		}
	}

	activeTicketID := strings.TrimSpace(*ticketIDFlag)
	if activeTicketID == "" && cachedFacts != nil && cachedFacts.TicketID != nil {
		activeTicketID = strings.TrimSpace(*cachedFacts.TicketID)
	}
	if activeTicketID == "" {
		fmt.Fprintf(os.Stderr, "Error: missing ticket ID. Specify --ticket-id explicitly.\n")
		os.Exit(1)
	}

	if cachedFacts != nil && (cachedFacts.TicketID == nil || strings.TrimSpace(*cachedFacts.TicketID) != activeTicketID) {
		fmt.Fprintln(os.Stderr, "Existing ledger ticket conflicts with requested ticket; refusing relabel")
		os.Exit(1)
	}

	if *jsonOutputFlag != "" {
		if e := recordGenerationPublication(project, *captureDirFlag, activeTicketID, "preparing", *jsonOutputFlag, *appendFlag, nil, "Generation began before source extraction. No replacement has been acknowledged; inspect latest reconciliation and retry generation if this attempt remains unfinished."); e != nil {
			fmt.Fprintln(os.Stderr, "Cannot persist generation attempt diagnostic:", e)
			os.Exit(1)
		}
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
	if !*autoFlag && (activeTicketTitle == nil || strings.TrimSpace(*activeTicketTitle) == "") {
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
	} else if cachedFacts != nil && cachedFacts.TicketType != nil && strings.TrimSpace(string(*cachedFacts.TicketType)) != "" {
		activeTicketType = cachedFacts.TicketType
	}

	if activeTicketType == nil && activeTicketID != "" {
		value, found, typeErr := projCfg.Ticket.ResolveTicketTypeMapping(activeTicketID)
		if typeErr != nil {
			fmt.Fprintln(os.Stderr, typeErr)
			os.Exit(1)
		}
		if found {
			activeTicketType = &value
		}
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

	baseline, baselineErr := classifyGenerationBaseline(project, *captureDirFlag, *jsonOutputFlag, activeTicketID, cachedFacts)
	if baselineErr != nil {
		fmt.Fprintln(os.Stderr, "Baseline authorization:", baselineErr)
		os.Exit(1)
	}

	if cachedFacts != nil {
		sessions = append(sessions, baseline.Current...)
		for _, owner := range baseline.Current {
			if owner.ParentID == nil && (owner.Source == domain.SourceOpenCode || owner.Source == "") {
				knownThreadIDs = append(knownThreadIDs, owner.ID)
			}
			userBySessionID[owner.ID] = owner.User
		}
		if cachedFacts.Migration != nil && strings.Contains(cachedFacts.Migration.Notes, "Removed ") {
			fmt.Fprintf(os.Stderr, "WARNING: %s\n", cachedFacts.Migration.Notes)
		}
	}

	var freshSessions []domain.SessionFact
	exactRolloutPaths := make(map[string]string)
	var sourceEvents domain.SourceEvents
	generationPolicy := projCfg.Generation.Metrics
	captureResult, captureErr := reconcileCaptureGenerationAtWithPolicy(project, *captureDirFlag, activeTicketID, snapshotCutoff, policy, generationPolicy)
	if captureErr != nil {
		fmt.Fprintf(os.Stderr, "Error reconciling capture registry; publication not completed: %v\n", captureErr)
		os.Exit(1)
	}
	if precedenceErr := filterCapturePrecedence(sessions, captureResult); precedenceErr != nil {
		fmt.Fprintln(os.Stderr, "Capture checkpoint comparison:", precedenceErr)
		os.Exit(1)
	}
	registeredOnly := (*autoFlag || hasBoundRegistrations(captureResult)) && strings.TrimSpace(*nonceFlag) == "" && strings.TrimSpace(*sessionIDFlag) == ""
	if registeredOnly {
		configuredStores := make(map[domain.SourceHarness]string)
		if *codexHomeFlag != "" || projCfg.Telemetry.Sources.CodexHome != "" {
			value, pathErr := adaptersconfig.ResolveCodexHome(*codexHomeFlag, projCfg.Telemetry.Sources.CodexHome)
			if pathErr != nil {
				fmt.Fprintln(os.Stderr, pathErr)
				os.Exit(1)
			}
			configuredStores[domain.SourceCodex] = value
		}
		if *dbPathFlag != "" || projCfg.Telemetry.Sources.OpenCodeDatabasePath != "" {
			value, pathErr := adaptersconfig.ResolveOpenCodePath(*dbPathFlag, projCfg.Telemetry.Sources.OpenCodeDatabasePath)
			if pathErr != nil {
				fmt.Fprintln(os.Stderr, pathErr)
				os.Exit(1)
			}
			configuredStores[domain.SourceOpenCode] = value
		}
		if associationErr := validateRegisteredGenerationBaseline(project, *captureDirFlag, activeTicketID, baseline.Current, configuredStores); associationErr != nil {
			fmt.Fprintln(os.Stderr, "Baseline ticket association conflict:", associationErr)
			os.Exit(1)
		}
		freshSessions = captureResult.RefreshedSessions
	} else if activeSource == "codex" {
		codexHome, pathErr := adaptersconfig.ResolveCodexHome(*codexHomeFlag, projCfg.Telemetry.Sources.CodexHome)
		if pathErr != nil {
			fmt.Fprintln(os.Stderr, pathErr)
			os.Exit(1)
		}
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
			transcript, parseErr := codex.ParseRolloutPathAt(path, snapshotCutoff)
			if parseErr != nil {
				fmt.Fprintf(os.Stderr, "Error reading selected Codex source rollout %s: %v. Check telemetry.sources.codex.home or --codex-home.\n", path, parseErr)
				os.Exit(1)
			}
			if len(transcript.Warnings) > 0 {
				fmt.Fprintf(os.Stderr, "Error: incomplete Codex owner %s; stored snapshots preserved: %s\n", transcript.ThreadID, strings.Join(transcript.Warnings, "; "))
				os.Exit(1)
			}
			transcripts = append(transcripts, transcript)
			exactRolloutPaths[transcript.ThreadID] = path
			knownThreadIDs = append(knownThreadIDs, transcript.ThreadID)
		}
		combined, combineErr := codex.CombineTranscriptsWithPolicy(transcripts, generationPolicy)
		if combineErr != nil {
			fmt.Fprintf(os.Stderr, "Error combining Codex session facts: %v\n", combineErr)
			os.Exit(1)
		}
		for _, warning := range combined.Warnings {
			fmt.Fprintf(os.Stderr, "Warning: Codex rollout: %s\n", warning)
		}
		freshCodex, factErr := combined.SessionFacts(generationPolicy)
		if factErr != nil {
			fmt.Fprintf(os.Stderr, "Error grouping Codex session facts: %v\n", factErr)
			os.Exit(1)
		}
		if collisionErr := capture.ValidateSessionMerge(sessions, freshCodex); collisionErr != nil {
			fmt.Fprintln(os.Stderr, collisionErr)
			os.Exit(1)
		}
		if ownerErr := validateCaptureSelection(project, *captureDirFlag, activeTicketID, domain.SourceCodex, codexHome, freshCodex, baseline.Current, exactRolloutPaths); ownerErr != nil {
			fmt.Fprintln(os.Stderr, "Capture ownership conflict:", ownerErr)
			os.Exit(1)
		}
		resolveCodexSessionUsers(freshCodex, userBySessionID, currentUser)
		freshSessions = freshCodex
		sourceEvents.Events = combined.Events
		sourceEvents.CheckpointEvidence = combined.CheckpointEvidence
		for _, owner := range freshCodex {
			sourceEvents.CompleteSessionIDs = append(sourceEvents.CompleteSessionIDs, owner.ID)
		}
		if strings.TrimSpace(*startedAtFlag) == "" && combined.StartedAt != "" {
			activeStartedAt = &combined.StartedAt
		}
		if strings.TrimSpace(*completedAtFlag) == "" && combined.CompletedAt != "" {
			activeCompletedAt = &combined.CompletedAt
		}
		fmt.Fprintf(os.Stderr, "Discovered %d Codex thread rollout(s) for session %s.\n", len(rolloutPaths), combined.ThreadID)
	} else {
		dbPath, pathErr := adaptersconfig.ResolveOpenCodePath(*dbPathFlag, projCfg.Telemetry.Sources.OpenCodeDatabasePath)
		if pathErr != nil {
			fmt.Fprintln(os.Stderr, pathErr)
			os.Exit(1)
		}
		if err := adaptersconfig.ValidateDatabasePolicy(projCfg); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		dbClient, dbErr := opencode.OpenDB(dbPath, projCfg.Database.BusyTimeoutMs)
		if dbErr == nil {
			defer dbClient.Close()

			rootID, resolveErr := opencode.ResolveActiveRootSession(
				dbClient.DB(),
				*nonceFlag,
				activeTicketID,
				*sessionIDFlag,
			)
			if resolveErr != nil {
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
				extractedSessions, dbStarted, dbCompleted, extractErr := opencode.ExtractSessionsStrictWithPolicyAt(
					dbClient.DB(), knownThreadIDs, currentUser, userBySessionID, &sourceEvents, generationPolicy, snapshotCutoff,
				)
				if extractErr != nil {
					fmt.Fprintf(os.Stderr, "Error reading selected OpenCode source at %s: %v. Check the database and telemetry.sources.opencode.database_path.\n", dbPath, extractErr)
					os.Exit(1)
				}
				if len(extractedSessions) > 0 {
					if collisionErr := capture.ValidateSessionMerge(sessions, extractedSessions); collisionErr != nil {
						fmt.Fprintln(os.Stderr, collisionErr)
						os.Exit(1)
					}
					if ownerErr := validateCaptureSelection(project, *captureDirFlag, activeTicketID, domain.SourceOpenCode, dbPath, extractedSessions, baseline.Current, nil); ownerErr != nil {
						fmt.Fprintln(os.Stderr, "Capture ownership conflict:", ownerErr)
						os.Exit(1)
					}
					freshSessions = extractedSessions
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
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := capture.ValidateSessionMerge(sessions, captureResult.Sessions); err != nil {
			fmt.Fprintln(os.Stderr, "Capture identity collision:", err)
			os.Exit(1)
		}
		sourceEvents = domain.MergeSourceEvents(captureResult.SourceEvents, sourceEvents)
		// Retained snapshots fill missing owners only; canonical and explicit facts remain authoritative.
		sessions = fillMissingCaptureSessions(sessions, captureResult.Sessions)
		// Explicit extraction runs after capture and has precedence for a duplicate refreshed owner.
		freshSessions = mergeSessions(captureResult.RefreshedSessions, freshSessions)
		for _, session := range captureResult.Sessions {
			knownThreadIDs = append(knownThreadIDs, session.ID)
		}

	}
	if len(freshSessions) > 0 {
		if err := capture.ValidateSessionMerge(sessions, freshSessions); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// Explicit complete local snapshots are authoritative for their owner,
		// including older restored sources and legitimate removal to empty.

		sessions = mergeSessions(sessions, freshSessions)
	}
	// Retained facts join only after owned extraction and checkpoint decisions.
	retainUnrefreshedGenerationOwners(&baseline, freshSessions)
	if err := recordGenerationHistory(project, *captureDirFlag, activeTicketID, baseline); err != nil {
		fmt.Fprintln(os.Stderr, "Cannot record retained history:", err)
		os.Exit(1)
	}
	invocations := (domain.AgenticFactLedger{Sessions: sessions}).InvocationFacts()

	if len(sessions) == 0 {
		fmt.Fprintf(os.Stderr, "Error: unable to identify root session for ticket '%s'. Provide --session-id, --nonce, with an exact existing source session. Telemetry generation aborted.\n", activeTicketID)
		os.Exit(1)
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

	cfg, err := git.LoadConfig(project)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading code-metric policy:", err)
		os.Exit(1)
	}
	codeFacts, err := git.ComputeCodeFactsAtRefInDir(project, strings.TrimSpace(*codeRefFlag), cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Fresh Git calculation failed; existing ledger/summary preserved:", err)
		os.Exit(1)
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

	// Completion is this observability snapshot finalization; source invocation
	// completion timestamps remain untouched.
	activeCompletedAt = &updatedAtISO

	var ticketIDPtr *string
	if activeTicketID != "" {
		ticketIDPtr = &activeTicketID
	}

	schemaVer := domain.CurrentSchemaVersion
	var recordedTags map[string]string
	if cachedFacts != nil {
		recordedTags = cachedFacts.Tags
	}
	factLedger := domain.AgenticFactLedger{
		Tags:          domain.MergeTags(recordedTags, projCfg.Tags),
		SchemaVersion: &schemaVer,
		TicketID:      ticketIDPtr,
		TicketTitle:   activeTicketTitle,
		TicketType:    activeTicketType,
		StartedAt:     activeStartedAt,
		CompletedAt:   activeCompletedAt,
		UpdatedAt:     updatedAtISO,
		Code:          codeFacts,
		Sessions:      sessions,
	}

	if cachedFacts != nil && cachedFacts.Anomalies != nil {
		factLedger.Anomalies = append([]domain.Anomaly{}, cachedFacts.Anomalies...)
	}
	factLedger, summaryErr := domain.PersistSourceAnomalySummariesWithPolicy(factLedger, sourceEvents, generationPolicy, projCfg.AnomalyThresholds)
	if summaryErr != nil {
		fmt.Fprintln(os.Stderr, "Cannot calculate source summaries:", summaryErr)
		os.Exit(1)
	}
	metricsLedger := domain.DeriveMetricsLedger(domain.ProjectGenerationFacts(factLedger, generationPolicy), projCfg)
	tableMD := adapters.FormatMarkdownMetricsTableWithPolicy(metricsLedger, generationPolicy, projCfg.RoleOrder)
	if captureResult != nil {
		tableMD += capture.CoverageSummary(*captureResult)
	}
	tableMD += historySummary(baseline)
	if !generationPolicy.Skills || !generationPolicy.Delegations || !generationPolicy.Anomalies {
		tableMD = strings.ReplaceAll(tableMD, "unchanged because their sources were not refreshed", "outside the source refresh set; optional categories follow the current generate settings")
	}

	if *autoFlag {
		enabled, e := automaticCaptureEnabled(project, *captureDirFlag)
		if e != nil || !enabled {
			fmt.Fprintln(os.Stderr, "Automatic publication disabled or unavailable:", e)
			os.Exit(1)
		}
	}

	publicationPaths := []string{}
	if *jsonOutputFlag != "" {
		publicationPaths = append(publicationPaths, *jsonOutputFlag)
		if *appendFlag != "" {
			publicationPaths = append(publicationPaths, *appendFlag)
		}
		if e := recordGenerationPublication(project, *captureDirFlag, activeTicketID, "staging", *jsonOutputFlag, *appendFlag, factLedger.Sessions, "Artifact replacement started; success is unconfirmed until journal commit and diagnostic acknowledgement. If interrupted, rerun generate to recover the prior accepted set.", freshSessions); e != nil {
			fmt.Fprintln(os.Stderr, "Cannot persist publication diagnostic:", e)
			os.Exit(1)
		}
		if e := adapters.BeginPublication(*jsonOutputFlag, publicationPaths); e != nil {
			fmt.Fprintln(os.Stderr, "Cannot stage recoverable publication:", e)
			os.Exit(1)
		}
	}
	if *jsonOutputFlag != "" {
		if err := saveGenerationLedgerPreservingHistory(*jsonOutputFlag, factLedger, baseline.HistoryDTOs); err != nil {
			fmt.Fprintln(os.Stderr, recoverFailedGenerationPublication(*jsonOutputFlag, publicationPaths, fmt.Errorf("writing JSON to %s: %w", *jsonOutputFlag, err)))
			os.Exit(1)
		}
		fmt.Printf("Structured metrics ledger written to %s\n", *jsonOutputFlag)
	}

	if *anomaliesFlag {
		tid := "UNKNOWN"
		if metricsLedger.TicketID != nil && *metricsLedger.TicketID != "" {
			tid = *metricsLedger.TicketID
		}
		if !generationPolicy.Anomalies {
			fmt.Printf("Anomaly analysis is disabled for %s by generation.metrics.anomalies=false.\n", tid)
		} else {
			printLedgerAnomalyDiagnostics(tid, metricsLedger.Anomalies, metricsLedger.Facts)
		}
	}

	if *appendFlag != "" {
		summaryPath, err := capture.ProjectDirectory(*appendFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, recoverFailedGenerationPublication(*jsonOutputFlag, publicationPaths, fmt.Errorf("resolving summary file %s: %w", *appendFlag, err)))
			os.Exit(1)
		}
		if err := adapters.UpdateSummaryFile(summaryPath, tableMD); err != nil {
			fmt.Fprintln(os.Stderr, recoverFailedGenerationPublication(*jsonOutputFlag, publicationPaths, fmt.Errorf("updating summary file %s: %w", *appendFlag, err)))
			os.Exit(1)
		}
		fmt.Printf("Metrics table successfully updated in %s\n", *appendFlag)
	} else if *jsonOutputFlag == "" && !*anomaliesFlag {
		fmt.Println(tableMD)
	}
	if *jsonOutputFlag != "" {
		if e := adapters.CommitPublication(*jsonOutputFlag, publicationPaths); e != nil {
			fmt.Fprintln(os.Stderr, recoverFailedGenerationPublication(*jsonOutputFlag, publicationPaths, fmt.Errorf("committing artifact publication: %w", e)))
			os.Exit(1)
		}
		if e := recordGenerationPublication(project, *captureDirFlag, activeTicketID, "committed", *jsonOutputFlag, *appendFlag, factLedger.Sessions, "Ledger and requested summary replacement committed.", freshSessions); e != nil {
			fmt.Fprintln(os.Stderr, "Artifacts committed but diagnostic acknowledgement failed; publication verification remains incomplete:", e)
			os.Exit(1)
		}
	}
	if !*autoFlag {
		// Generic generate has no independently verified native caller/interval
		// transport. Source selection and process environment are not proof.
		feedback := domain.SelectCheckpointFeedback(factLedger.Sessions, sourceEvents.CheckpointEvidence, domain.CheckpointCaller{}, snapshotCutoff, projCfg.Generation.Feedback.ProvisionalThresholdSeconds)
		if err := writeCheckpointFeedback(os.Stderr, feedback); err != nil {
			fmt.Fprintln(os.Stderr, "Optional checkpoint feedback unavailable:", err)
		}
	}
}

func resolveCodexSessionUsers(fresh []domain.SessionFact, existingUsers map[string]string, gitUser string) {
	for i := range fresh {
		fresh[i].User = git.ResolveSessionUser(fresh[i].User, existingUsers[fresh[i].ID], gitUser)
	}
}

func runServer(args []string) {

	serverFlags := flag.NewFlagSet("server", flag.ExitOnError)
	portFlag := serverFlags.Int("port", 8085, "Port to listen on")
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

	serverFlags.Visit(func(f *flag.Flag) {
		if (f.Name == "workspaces" && strings.TrimSpace(*workspacesFlag) == "") || (f.Name == "db-path" && strings.TrimSpace(*dbPathFlag) == "") {
			fmt.Fprintf(os.Stderr, "--%s must be nonblank when supplied\n", f.Name)
			os.Exit(1)
		}
	})

	if *portFlag < 1 || *portFlag > 65535 {
		fmt.Fprintf(os.Stderr, "Error: invalid port %d: must be between 1 and 65535\n", *portFlag)
		os.Exit(1)
	}

	workspacesDir, projCfg, err := adapters.ResolveWorkspaceContext(*workspacesFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Server workspace policy:", err)
		os.Exit(1)
	}
	staticFS, err := web.GetFS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing embedded web filesystem: %v\n", err)
		os.Exit(1)
	}

	if optional, localErr := adaptersconfig.Load(filepath.Dir(projCfg.ConfigPath)); localErr != nil {
		fmt.Fprintln(os.Stderr, "Optional hydration unavailable:", localErr)
	} else {
		projCfg = optional
	}
	liveService := live.NewLiveServiceWithCodexHome(nil, projCfg, projCfg.Telemetry.Sources.CodexHome)
	if strings.TrimSpace(*dbPathFlag) != "" || projCfg.Telemetry.Sources.OpenCodeDatabasePath != "" {
		dbPath, pathErr := adaptersconfig.ResolveOpenCodePath(*dbPathFlag, projCfg.Telemetry.Sources.OpenCodeDatabasePath)
		policyErr := errors.Join(adaptersconfig.ValidateDatabasePolicy(projCfg), adaptersconfig.ValidateActivityPolicy(projCfg), adaptersconfig.ValidatePollingPolicy(projCfg), adaptersconfig.ValidateAnomalyPolicy(projCfg))
		if pathErr != nil || policyErr != nil {
			fmt.Fprintln(os.Stderr, "Optional live source unavailable:", errors.Join(pathErr, policyErr))
		} else if hook, hookErr := live.OpenReadOnlyHook(dbPath, projCfg.Database.BusyTimeoutMs); hookErr != nil {
			fmt.Fprintln(os.Stderr, "Optional live source unavailable:", hookErr)
		} else {
			liveService = live.NewLiveServiceWithCodexHome(hook, projCfg, projCfg.Telemetry.Sources.CodexHome)
			if startErr := liveService.Start(context.Background()); startErr != nil {
				fmt.Fprintln(os.Stderr, startErr)
				os.Exit(1)
			}
		}
	}

	srv := adaptershttp.NewServerWithConfig(workspacesDir, staticFS, projCfg, liveService)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	listener, bindErr := srv.Bind(*portFlag)
	if bindErr != nil {
		fmt.Fprintln(os.Stderr, "Server bind failed:", bindErr)
		os.Exit(1)
	}
	fmt.Printf("AgenticLens server listening at http://%s (workspaces: %s)\n", listener.Addr().String(), workspacesDir)
	go func() {
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			if errors.Is(err, syscall.EADDRINUSE) || strings.Contains(strings.ToLower(err.Error()), "address already in use") {
				fmt.Fprintf(os.Stderr, "Server failed: port %d is already in use. Specify an alternate port via --port.\n", *portFlag)
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
		fmt.Fprintf(out, "  --codex-home <path>    Explicit Codex source store; otherwise selected local telemetry.sources.codex.home\n")
		fmt.Fprintf(out, "  --workspaces <path>    Path to root workspaces directory containing agentic_metrics.json ledgers\n")
		fmt.Fprintf(out, "  --ticket-id <id>       Target single ticket identifier to filter backfill operation\n")
		fmt.Fprintf(out, "  --recompute-git        Recompute Git code facts and SDLC iterations from repository history\n")
		fmt.Fprintf(out, "  --code-ref <ref>        Required with --recompute-git; branch or commit to compare with configured base\n")
		fmt.Fprintf(out, "  --recompute-db         Refresh complete selected-source snapshots without --force; retain unavailable history\n")
		fmt.Fprintf(out, "  --migrate-schema       Upgrade telemetry ledger schema versions to latest supported version\n")
		fmt.Fprintf(out, "  --dry-run              Preview planned migrations without writing changes to disk\n")
		fmt.Fprintf(out, "  --update-summaries     Idempotently update the configured summary filename\n")
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
	codexHomeFlag := backfillFlags.String("codex-home", "", "Explicit Codex source store; otherwise selected local telemetry.sources.codex.home")
	ticketIDFlag := backfillFlags.String("ticket-id", "", "Target single ticket identifier to filter backfill operation")
	recomputeGitFlag := backfillFlags.Bool("recompute-git", false, "Recompute Git code facts and SDLC iterations from repository history")
	codeRefFlag := backfillFlags.String("code-ref", "", "Git branch or commit to compare with the configured base when recomputing code facts")
	recomputeDBFlag := backfillFlags.Bool("recompute-db", false, "Refresh complete selected-source snapshots without --force; retain unavailable history")
	migrateSchemaFlag := backfillFlags.Bool("migrate-schema", false, "Upgrade telemetry ledger schema versions to latest supported version")
	dryRunFlag := backfillFlags.Bool("dry-run", false, "Preview planned migrations without writing changes to disk")
	updateSummariesFlag := backfillFlags.Bool("update-summaries", false, "Idempotently update the configured summary filename")
	dbPathFlag := backfillFlags.String("db-path", "", "Explicit path to OpenCode SQLite database file")
	forceFlag := backfillFlags.Bool("force", false, "Repeat completed-ledger processing; never bypass preservation or enable fallbacks")
	allowFallbacksFlag := backfillFlags.Bool("allow-fallbacks", false, "Allow heuristic session discovery fallbacks (title match, workspace commit range) when exact session IDs are not found in DB")

	if err := backfillFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing backfill flags: %v\n", err)
		os.Exit(1)
	}
	backfillFlags.Visit(func(f *flag.Flag) {
		if (f.Name == "workspaces" && strings.TrimSpace(*workspacesFlag) == "") || (f.Name == "db-path" && strings.TrimSpace(*dbPathFlag) == "") || (f.Name == "codex-home" && strings.TrimSpace(*codexHomeFlag) == "") {
			fmt.Fprintf(os.Stderr, "--%s must be nonblank when supplied\n", f.Name)
			os.Exit(1)
		}
	})
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
	var pathErr error
	resolvedDBPath, pathErr = adaptersconfig.ExpandHomePath(resolvedDBPath)
	if pathErr != nil {
		fmt.Fprintln(os.Stderr, pathErr)
		os.Exit(1)
	}
	resolvedCodexHome, pathErr = adaptersconfig.ExpandHomePath(resolvedCodexHome)
	if pathErr != nil {
		fmt.Fprintln(os.Stderr, pathErr)
		os.Exit(1)
	}

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
		fmt.Println("No readable matching ticket ledgers found to backfill.")
		if len(report.Warnings) == 0 {
			return
		}
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
			notes = strings.TrimSpace(fmt.Sprintf("%s; error: %v", notes, res.Error))
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
		fmt.Printf("  Ledger: %s\n", res.LedgerPath)
	}

	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("Summary: Scanned: %d | Updated: %d | Skipped: %d | Completed: %d | Partial: %d\n",
		report.Scanned, report.Updated, report.Skipped, report.Completed, report.Partial)
	fmt.Printf("Refresh diagnostics: Incomplete: %d | Retained skill/task snapshots: %d (session trees may be newer)\n", report.RefreshIncomplete, report.CollectionsRetained)

	if report.SchemaMigration != nil {
		fmt.Println("--------------------------------------------------------------------------------")
		for _, res := range report.SchemaMigration.Results {
			notes := res.DiffSummary
			if res.Error != nil {
				notes = strings.TrimSpace(fmt.Sprintf("%s; error: %v", notes, res.Error))
			}
			fmt.Printf("  Schema ledger: %s; %s; %s\n", res.LedgerPath, res.Status, notes)
		}
		fmt.Printf("Schema Migration: Migrated: %d | Up to date: %d | Skipped: %d | Errors: %d\n",
			report.SchemaMigration.Migrated, report.SchemaMigration.UpToDate,
			report.SchemaMigration.Skipped, len(report.SchemaMigration.Errors))
	}

	if opts.DryRun {
		fmt.Println("Dry run: no ledgers or backups were written.")
	} else {
		fmt.Println("Rollback: backfill creates no backups; restore ledger originals from Git or another backup. Before schema changes, use migrate --backup to retain originals.")
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
		fmt.Fprintf(out, "  --update-summaries       Idempotently update the configured summary filename\n")
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
	updateSummariesFlag := migrateFlags.Bool("update-summaries", false, "Idempotently update the configured summary filename")

	if err := migrateFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing migrate flags: %v\n", err)
		os.Exit(1)
	}

	migrateFlags.Visit(func(f *flag.Flag) {
		if f.Name == "workspaces" && strings.TrimSpace(*workspacesFlag) == "" {
			fmt.Fprintln(os.Stderr, "--workspaces must be nonblank when supplied")
			os.Exit(1)
		}
	})

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
		fmt.Println("No readable matching ticket workspaces found to migrate.")
		if len(report.Errors) == 0 {
			return
		}
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
			notes = strings.TrimSpace(fmt.Sprintf("%s; error: %v", notes, res.Error))
		}

		fmt.Printf("%-15s %-12s %-12s %-18s %s\n",
			res.TicketID,
			fromStr,
			toStr,
			res.Status,
			notes,
		)
		fmt.Printf("  Ledger: %s\n", res.LedgerPath)
		if res.BackupCreated {
			fmt.Printf("  Backup: %s; restore this file over the ledger to revert it.\n", res.BackupPath)
		}
	}

	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("Summary: Scanned: %d | Migrated: %d | Up to date: %d | Skipped: %d | Errors: %d\n",
		report.Scanned, report.Migrated, report.UpToDate, report.Skipped, len(report.Errors))

	if opts.DryRun {
		fmt.Println("Dry run: no ledgers or backups were written. Use --backup when executing to retain originals.")
	} else {
		fmt.Println("Rollback: restore the printed .bak file when available, or recover the original ledger from Git or another backup. Console audit cannot reconstruct discarded payloads.")
		if !opts.Backup {
			fmt.Println("No backups were requested. Use --backup before executing schema changes to retain originals.")
		}
	}

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

func printLedgerAnomalyDiagnostics(ticketID string, anomalies []domain.Anomaly, facts domain.AgenticFactLedger) {
	if !facts.HasExternalParentReference() {
		printAnomalyDiagnostics(ticketID, anomalies, facts.UnrecordedAnomalyScopes())
		return
	}
	fmt.Printf("Parent/group anomaly context outside ticket %s is partial; recorded summaries remain available.\n", ticketID)
	if count := facts.UnrecordedAnomalyScopes(); count > 0 {
		fmt.Printf("Anomaly analysis has not been recorded for %d owner/group scope(s).\n", count)
	}
	if len(anomalies) == 0 {
		fmt.Printf("No complete clean-session conclusion is available for ticket %s.\n", ticketID)
		return
	}
	printAnomalyDiagnostics(ticketID, anomalies)
}

func printAnomalyDiagnostics(ticketID string, anomalies []domain.Anomaly, unavailableSessions ...int) {
	if len(unavailableSessions) > 0 && unavailableSessions[0] > 0 {
		fmt.Printf("Anomaly analysis was not recorded for %d owner/group scope(s) in %s; complete source-backed refresh is required.\n", unavailableSessions[0], ticketID)
		if len(anomalies) == 0 {
			return
		}
	}
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
	fmt.Fprintln(out, "  format    Reformat selected persisted ledgers without changing facts")
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
	if err := initCommand(args); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func initCommand(args []string) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	settings := flags.String("settings", "", "Explicit approved shared YAML candidate; required for shared configuration generation")
	localSettings := flags.String("local-settings", "", "Explicit approved local YAML choices; fills only missing choices")
	name := flags.String("project-name", "", "Explicit project label override for the shared candidate")
	version := flags.String("agenticlens-version", "", "Approved exact release pin; otherwise preserve candidate/existing valid pin")
	workspaces := flags.String("workspaces", "", "Explicit workspace root (required for new shared configuration)")
	output := flags.String("output", ".agenticlens-config.yaml", "Canonical shared YAML output path (.agenticlens-config.yaml or .agenticlens-config.yml)")
	localOnly := flags.Bool("local-only", false, "Fill missing explicitly supplied local choices; preserve valid existing choices")
	state := flags.String("state-dir", "", "Explicit local capture state directory")
	enabled := flags.String("capture-enabled", "", "Explicit true or false; configuration does not install capture integration")
	source := flags.String("source", "", "Explicit local telemetry source: codex or opencode")
	codexHome := flags.String("codex-home", "", "Explicit Codex source store")
	database := flags.String("db-path", "", "Explicit OpenCode source database")
	force := flags.Bool("force", false, "Overwrite the selected shared file with the approved candidate")
	complete := flags.Bool("complete-scaffold", false, "Complete only an unchanged owned bootstrap scaffold; preserve valid local settings")
	updateCapture := flags.Bool("update-capture-config", false, "Apply explicit state path maintenance without reinitializing configuration")
	captureState := flags.String("capture-state-dir", "", "Explicit replacement used with --update-capture-config")
	dry := flags.Bool("dry-run", false, "Show validated output without writing or installing anything")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected init arguments")
	}
	provided := map[string]bool{}
	flags.Visit(func(f *flag.Flag) { provided[f.Name] = true })
	for _, choice := range []struct{ flag, value string }{{"settings", *settings}, {"local-settings", *localSettings}, {"project-name", *name}, {"agenticlens-version", *version}, {"workspaces", *workspaces}, {"state-dir", *state}, {"capture-enabled", *enabled}, {"source", *source}, {"codex-home", *codexHome}, {"db-path", *database}, {"capture-state-dir", *captureState}} {
		if provided[choice.flag] && strings.TrimSpace(choice.value) == "" {
			return fmt.Errorf("--%s must be nonblank when supplied", choice.flag)
		}
	}
	if *localOnly {
		for key := range provided {
			switch key {
			case "local-only", "local-settings", "state-dir", "capture-enabled", "source", "codex-home", "db-path", "output", "dry-run":
			default:
				return fmt.Errorf("--local-only cannot be combined with --%s", key)
			}
		}
	}
	if strings.TrimSpace(*output) == "" {
		return errors.New("--output must be an explicit nonblank canonical YAML path")
	}
	expandedOutput, err := adaptersconfig.ExpandHomePath(*output)
	if err != nil {
		return err
	}
	outputPath, err := filesystem.PhysicalPath(expandedOutput)
	if err != nil {
		return err
	}
	base := filepath.Base(outputPath)
	if base != ".agenticlens-config.yaml" && base != ".agenticlens-config.yml" {
		return errors.New("--output must use .agenticlens-config.yaml or .agenticlens-config.yml")
	}
	project := filepath.Dir(outputPath)
	other := ".agenticlens-config.yml"
	if base == other {
		other = ".agenticlens-config.yaml"
	}
	if _, err := os.Lstat(filepath.Join(project, other)); err == nil {
		return fmt.Errorf("ambiguous shared configuration: %s and %s would coexist", outputPath, filepath.Join(project, other))
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := adaptersconfig.CheckBaseSetup(project); err != nil {
		return err
	}
	if *updateCapture {
		for key := range provided {
			if key != "update-capture-config" && key != "capture-state-dir" && key != "output" && key != "dry-run" {
				return fmt.Errorf("--update-capture-config cannot be combined with --%s", key)
			}
		}
		return runCaptureConfigUpdate(*output, *captureState, *dry)
	}
	if provided["capture-state-dir"] {
		return errors.New("--capture-state-dir requires --update-capture-config")
	}
	// CLI path choices belong to the invocation directory, even when output is elsewhere.
	for _, choice := range []struct {
		flag  string
		value *string
	}{{"workspaces", workspaces}, {"codex-home", codexHome}, {"db-path", database}, {"state-dir", state}} {
		if !provided[choice.flag] {
			continue
		}
		expanded, pathErr := adaptersconfig.ExpandHomePath(*choice.value)
		if pathErr != nil {
			return pathErr
		}
		physical, pathErr := filesystem.PhysicalPath(expanded)
		if pathErr != nil {
			return fmt.Errorf("--%s: %w", choice.flag, pathErr)
		}
		*choice.value = physical
	}
	for _, candidate := range []*string{settings, localSettings} {
		if *candidate != "" {
			expanded, pathErr := adaptersconfig.ExpandHomePath(*candidate)
			if pathErr != nil {
				return pathErr
			}
			*candidate = expanded
		}
	}
	localYAML, err := adaptersconfig.InitialLocalYAMLWithSettings(project, *localSettings, *enabled, *source, *codexHome, *database, *state)
	if err != nil {
		return err
	}
	localPath := filepath.Join(project, ".agenticlens-config.local.yaml")
	before, localErr := os.ReadFile(localPath)
	if localErr != nil && !errors.Is(localErr, os.ErrNotExist) {
		return localErr
	}
	if *localOnly {
		if *dry {
			fmt.Print(string(localYAML))
			return nil
		}
		if localErr == nil {
			if bytes.Equal(before, localYAML) {
				fmt.Println("Existing valid local choices retained.")
				return nil
			}
			return adaptersconfig.WriteCaptureConfigUpdate(localPath, adaptersconfig.CaptureConfigUpdate{Original: before, Data: localYAML, Changed: true})
		}
		return writeConfigPair(localPath, localYAML, "", nil, false)
	}
	if strings.TrimSpace(*settings) == "" {
		return errors.New("--settings is required: review an explicit shared YAML candidate containing prefixes/descriptions, classification lists and runtime policy before init")
	}
	cfg, err := adaptersconfig.ReadConfigCandidate(*settings)
	if err != nil {
		return fmt.Errorf("approved shared settings: %w", err)
	}
	if provided["project-name"] {
		if strings.TrimSpace(*name) == "" {
			return errors.New("--project-name cannot be blank")
		}
		cfg.ProjectName = *name
	}
	if provided["workspaces"] {
		if strings.TrimSpace(*workspaces) == "" {
			return errors.New("--workspaces cannot be blank")
		}
		cfg.Workspaces.Dir = *workspaces
	} else if !*complete {
		return errors.New("new configuration requires an explicit nonblank --workspaces")
	}
	if provided["agenticlens-version"] {
		if err := adaptersconfig.ValidateReleaseVersion(*version); err != nil {
			return err
		}
		cfg.AgenticLens.Version = *version
	}
	if !provided["agenticlens-version"] {
		if existing, err := adaptersconfig.ReadReleasePin(outputPath); err == nil {
			cfg.AgenticLens.Version = existing
		} else if !errors.Is(err, os.ErrNotExist) && !*complete {
			return err
		}
	}
	if err := adaptersconfig.ValidateReleaseVersion(cfg.AgenticLens.Version); err != nil {
		return fmt.Errorf("explicit approved agenticlens.version is required: %w", err)
	}
	if *complete && *force {
		return errors.New("--complete-scaffold cannot be combined with --force")
	}
	var scaffold []byte
	if *complete {
		scaffold, err = os.ReadFile(outputPath)
		if err != nil {
			return err
		}
		if err := adaptersconfig.ValidateBootstrapScaffold(scaffold, cfg.AgenticLens.Version); err != nil {
			return err
		}
	}
	yamlData, err := adaptersconfig.GenerateConfigYAML(cfg)
	if err != nil {
		return err
	}
	if *dry {
		fmt.Printf("# %s\n%s\n# %s\n%s", outputPath, yamlData, localPath, localYAML)
		return nil
	}
	if localErr == nil && !bytes.Equal(before, localYAML) {
		return errors.New("existing local configuration requires explicit missing choices; run init --local-only first; neither output was written")
	}
	if localErr == nil {
		localPath = ""
	}
	if *complete {
		err = writeConfigPair(outputPath, yamlData, localPath, localYAML, false, scaffold)
	} else {
		err = writeConfigPair(outputPath, yamlData, localPath, localYAML, *force)
	}
	if err != nil {
		return err
	}
	fmt.Printf("Initialized approved AgenticLens configuration at %s. Capture enablement, integration installation and manual generation are separate choices.\n", outputPath)
	return nil
}

func writeConfigPair(firstPath string, firstData []byte, secondPath string, secondData []byte, force bool, ownedScaffold ...[]byte) (resultErr error) {
	if len(ownedScaffold) > 1 {
		return errors.New("invalid scaffold replacement authority")
	}
	mayReplace := func(index int, path string) (bool, error) {
		if index != 0 || len(ownedScaffold) == 0 {
			return force, nil
		}
		current, err := os.ReadFile(path)
		if err != nil {
			return false, err
		}
		if !bytes.Equal(current, ownedScaffold[0]) {
			return false, errors.New("owned scaffold changed before replacement; existing configuration preserved")
		}
		return true, nil
	}
	type stagedFile struct {
		target, temp, backup string
		installed            bool
	}
	items := []struct {
		path string
		data []byte
	}{{firstPath, firstData}}
	if secondPath != "" {
		items = append(items, struct {
			path string
			data []byte
		}{secondPath, secondData})
	}
	// Validate the complete pair before moving either existing configuration.
	for index, item := range items {
		info, err := os.Stat(item.path)
		if errors.Is(err, os.ErrNotExist) {
			if index == 0 && len(ownedScaffold) > 0 {
				return errors.New("owned scaffold disappeared before replacement; configuration preserved")
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("checking %s: %w", item.path, err)
		}
		if info.IsDir() {
			return fmt.Errorf("target %s is a directory", item.path)
		}
		allowed, err := mayReplace(index, item.path)
		if err != nil {
			return err
		}
		if !allowed {
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
			if i == 0 && len(ownedScaffold) > 0 {
				return errors.New("owned scaffold disappeared before replacement; configuration preserved")
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("checking %s: %w", item.target, err)
		}
		if info.IsDir() {
			return fmt.Errorf("target %s is a directory", item.target)
		}
		allowed, err := mayReplace(i, item.target)
		if err != nil {
			return err
		}
		if !allowed {
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
	if len(os.Args) == 3 && os.Args[1] == "engine" && os.Args[2] == "protocol" {
		fmt.Println("1")
		return
	}
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
				case "format":
					runFormat([]string{"--help"})
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
		case "manager", "dispatch":
			fmt.Fprintln(os.Stderr, "This executable is the AgenticLens engine; installation and routing belong to the agentic-lens manager.")
			os.Exit(2)
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
		case "format":
			runFormat(os.Args[2:])
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
	anomaliesFlag := flag.Bool("anomalies", false, "Display recorded anomalies and availability without source refresh")
	versionFlag := flag.Bool("v", false, "Print version information and exit")
	versionLongFlag := flag.Bool("version", false, "Print version information and exit")

	flag.Parse()

	if *versionFlag || *versionLongFlag {
		runVersion(nil)
		return
	}

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "workspaces" && strings.TrimSpace(*workspacesFlag) == "" {
			fmt.Fprintln(os.Stderr, "--workspaces must be nonblank when supplied")
			os.Exit(1)
		}
	})
	workspacesDir, projCfg, err := adapters.ResolveWorkspaceContext(*workspacesFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error selecting workspace policy:", err)
		os.Exit(1)
	}
	report := adapters.ScanWorkspaces(workspacesDir)
	if report.FatalError != "" {
		fmt.Fprintf(os.Stderr, "Error scanning workspaces: %s\n", report.FatalError)
		os.Exit(1)
	}
	for _, diagnostic := range report.Diagnostics {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", diagnostic.Message)
	}
	if *ticketIDFlag != "" {
		for _, diagnostic := range report.Diagnostics {
			if diagnostic.TicketID == *ticketIDFlag {
				fmt.Fprintf(os.Stderr, "Error: Ticket '%s' was rejected during workspace scanning: %s\n", *ticketIDFlag, diagnostic.Message)
				os.Exit(1)
			}
		}
	}

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
			dur := ledger.WorkflowEfficiency.ActiveDurationFormatted + " (" + domain.DurationDescription(ledger.WorkflowEfficiency.ProvisionalSeconds) + ")"
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
		printLedgerAnomalyDiagnostics(tid, targetLedger.Anomalies, targetLedger.Facts)
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
	local = domain.RetainIncompleteGroupSummaries(cached, local)
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

func authorizedPublicationPaths(ledger, summary string) []string {
	paths := []string{ledger}
	if summary != "" {
		paths = append(paths, summary)
	}
	return paths
}

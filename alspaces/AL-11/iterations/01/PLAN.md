# Implementation Plan: AL-11 Native Go Telemetry Extractor and Active Session Metric Generation

## 1. Goal and Iteration Scope
The goal of iteration 01 is to deliver the complete native Go telemetry extraction engine and active session generator, fulfilling ticket AL-11. This encompasses creating the pure Go OpenCode SQLite adapter (internal/adapters/opencode/) using modernc.org/sqlite, creating the Git analysis and hybrid classification adapter (internal/adapters/git/), implementing the generate subcommand in cmd/agentic-lens/main.go, ensuring idempotent Markdown summary table updates, and validating behavior against both synthesized SQLite test fixtures and live development sessions.

Upon completion, telemetry extraction and summary generation for AgenticLens can run natively using the compiled Go binary without requiring Python 3 or scripts/extract_agentic_metrics.py.

## 2. Grounded Reference Architecture & Paths
The implementation maps directly to the following repository locations:
- Dependency Manifest: go.mod (adding pure Go SQLite driver modernc.org/sqlite)
- Domain Models & Aggregations:
  - internal/domain/models.go (existing InvocationTokens, InvocationFact, ToolFact, CodeFacts, AgenticFactLedger, AgenticMetricsLedger)
  - internal/domain/aggregations.go (ComputeTotals, ComputeWorkflowEfficiency, DeriveMetricsLedger)
  - internal/domain/constants.go (HealthyCeilingTokens, formatters)
  - internal/domain/enums.go (AgentRole, ContextHealthStatus)
- SQLite OpenCode Adapter:
  - internal/adapters/opencode/client.go (read-only SQLite connection management and schema querying)
  - internal/adapters/opencode/resolver.go (session resolution via nonce, ticket ID, explicit session ID, and fallback)
  - internal/adapters/opencode/extractor.go (extraction of invocations, tool facts, compactions, and timestamps)
  - internal/adapters/opencode/extractor_test.go (table-driven tests using in-memory SQLite database)
- Git & Workspace Adapter:
  - internal/adapters/git/analyzer.go (git process execution, revision range resolution, ticket numstat)
  - internal/adapters/git/classifier.go (file classification, test vs prod breakdown, config-driven overrides)
  - internal/adapters/git/inference.go (ticket title, ticket type, user identification, SDLC iteration counting)
  - internal/adapters/git/analyzer_test.go (table-driven tests for numstat parsing and file classification)
- Repository & Formatting Adapters:
  - internal/adapters/schemas.go (DTO mapping, SerializeLedger, AgenticFactLedgerDTO)
  - internal/adapters/formatters.go (FormatMarkdownMetricsTable, UpdateSummaryFile)
  - internal/adapters/repository.go (ResolveWorkspacesDir, loading existing ledger facts)
- CLI Entrypoint:
  - cmd/agentic-lens/main.go (subcommand generate parsing, orchestration, output dispatch)
  - cmd/agentic-lens/main_test.go (CLI flag parsing tests)
- Root Automation:
  - Makefile (ensuring check, vet, test, and build targets include new adapter packages)

## 3. Phased Implementation Tasks

### Phase 1: Pure Go SQLite Driver Integration and Domain Verification
- Task 1.1: Add modernc.org/sqlite to go.mod using go get modernc.org/sqlite. Verify that compilation succeeds under CGO_ENABLED=0 with go vet ./....
- Task 1.2: Audit internal/domain/models.go and internal/domain/aggregations.go to confirm full representation of extraction facts. Confirm that InvocationFact, ToolFact, CodeFacts, and AgenticFactLedger satisfy all requirements without requiring domain changes.

### Phase 2: OpenCode SQLite Adapter Implementation (internal/adapters/opencode/)
- Task 2.1: Implement internal/adapters/opencode/client.go:
  - Define OpenCodeDB struct encapsulating *sql.DB.
  - Implement OpenDB function accepting database path string (defaulting to ~/.local/share/opencode/opencode.db).
  - Configure connection string with pure Go driver name sqlite and read-only URI parameters (file:...?...&mode=ro&_pragma=query_only(1)).
  - Implement Close method.
- Task 2.2: Implement internal/adapters/opencode/resolver.go:
  - Implement ResolveActiveRootSession function with prioritized resolution strategies:
    - Priority 1: Exact Nonce Matching querying part and session tables for parts containing nonce substring in data where root session parent_id is null.
    - Priority 2: Explicit session ID passed directly as argument.
    - Priority 3: Ticket ID matching querying session table for parent_id is null and title like ticket ID pattern, ordered by time_updated descending.
    - Priority 4: Fallback to most recent message session parent or ID only if both nonce and ticket ID are empty.
- Task 2.3: Implement internal/adapters/opencode/extractor.go:
  - Implement ExtractFacts function accepting database connection, root thread IDs, fallback git user, and cached user mapping.
  - Query earliest time_created and latest time_updated across sessions for started_at and completed_at timestamps.
  - Query compaction count per session from part table where json_extract(data, '$.type') = 'compaction'.
  - Query sessions belonging to threads ordered by time_created ascending. Deduce role (orchestrator for root or build/orchestrator agent, else agent name).
  - Query messages per session ordered by time_created ascending. Parse JSON message data for assistant messages: accumulate input, output, cache read, and reasoning tokens; track peak context; compute active duration from created and completed timestamps; accumulate financial costs.
  - Formulate InvocationFact records per session-model combination, assigning session compactions to the primary invocation.
  - Query tool executions from part records where json_extract(data, '$.type') = 'tool', aggregating calls and failure counts.
  - Return extracted slice of InvocationFact, ToolFact, started_at ISO string, and completed_at ISO string.
- Task 2.4: Implement internal/adapters/opencode/extractor_test.go:
  - Setup in-memory SQLite database creating session, message, and part tables.
  - Insert deterministic fixture sessions, child subagent sessions, assistant messages with token JSON, compactions, and tool executions.
  - Test nonce resolution, ticket title resolution, and message token extraction.
  - Verify exact match of accumulated tokens, costs, peak context, and tool call counts.

### Phase 3: Git and Workspace Analysis Adapter (internal/adapters/git/)
- Task 3.1: Implement internal/adapters/git/classifier.go:
  - Define ClassificationConfig struct with patterns for include, exclude, test_patterns, and prod_patterns.
  - Implement LoadConfig function reading optional code_metrics from .agenticlens/config.json.
  - Implement ClassifyFile function applying glob patterns:
    - Identify ignored files: markdown documentation, JSON schemas/ledgers, lockfiles, workspaces directory contents.
    - Identify test files: Go tests (*_test.go), TypeScript tests (*.test.ts, *.test.tsx, *.spec.ts, *.spec.tsx), Python tests (test_*.py, *_test.py), test directory paths (tests/, test/, __tests__/).
    - Identify production files: .go, .ts, .tsx, .py, .js, .jsx, .html, .css, Makefile (excluding test files).
- Task 3.2: Implement internal/adapters/git/analyzer.go:
  - Implement GetTicketNumstat function executing git commands to resolve revision range:
    - Check active branch name against ticket ID, diffing against merge-base with origin/main or HEAD.
    - Query git log origin/main with --grep for merged ticket commits.
    - Inspect local and remote feature branches matching ticket ID.
    - Query git log --all with --grep for local commits.
  - Parse numstat output into lines added, lines deleted, and relative path.
  - Apply ClassifyFile to compute aggregate lines added, lines deleted, test lines added/deleted, prod lines added/deleted, and file count.
- Task 3.3: Implement internal/adapters/git/inference.go:
  - Implement GetGitUser reading git config user.email with fallback to user.name, defaulting to unknown.
  - Implement InferTicketTitle inspecting workspaces/<ticket_id>/SUMMARY.md first heading, CONTRACT.md first heading, or git log commit subjects.
  - Implement InferTicketType checking branch name prefix (feat/, fix/, chore/, refactor/, docs/) or conventional commit messages.
  - Implement ComputeSDLCIterations inspecting workspaces/<ticket_id>/iterations/ subdirectories.
- Task 3.4: Implement internal/adapters/git/analyzer_test.go:
  - Table-driven tests for numstat text parsing.
  - Unit tests verifying classification of Go, TypeScript, Python, and web asset files into test vs production categories.
  - Unit tests verifying ignore rules for documentation, lockfiles, and workspace files.
  - Unit tests verifying ticket title and type inference.

### Phase 4: CLI Engine and Subcommand Wiring (cmd/agentic-lens/main.go)
- Task 4.1: Extend cmd/agentic-lens/main.go to route subcommands:
  - If first argument is server, invoke runServer.
  - If first argument is generate, invoke runGenerate.
  - Otherwise, preserve default behavior (scan/list/filter of existing ledgers).
- Task 4.2: Implement runGenerate function in cmd/agentic-lens/main.go:
  - Parse flags using dedicated flag.FlagSet: --ticket-id, --nonce, --session-id, --ticket-title, --ticket-type, --started-at, --completed-at, --json-output, --append, and --db-path.
  - If --json-output exists on disk, load existing ledger facts (ticket ID, title, type, started_at, completed_at, root thread IDs, cached invocations, cached tools, cached code facts) via adapters.LoadExistingLedger.
  - Connect to OpenCode SQLite database via opencode.OpenDB. Resolve root session and extract facts. Fall back to cached facts if database is unavailable.
  - Execute Git analysis via git.ComputeCodeFacts, git.InferTicketTitle, git.InferTicketType, and git.ComputeSDLCIterations.
  - Assemble domain.AgenticFactLedger and project full metrics ledger via domain.DeriveMetricsLedger.
  - If --json-output is specified, serialize the atomic fact ledger and write formatted JSON.
  - If --append is specified, format Markdown metrics table via adapters.FormatMarkdownMetricsTable and update summary file idempotently via adapters.UpdateSummaryFile.
  - If neither --append nor --json-output is specified, print Markdown table to stdout.
- Task 4.3: Implement internal/adapters/repository.go helper LoadExistingFactLedger:
  - Safely read and deserialize an existing agentic_metrics.json file into raw fact components, supporting both current schema and legacy thread schemas.

### Phase 5: Verification, Quality Gates & Testing
- Task 5.1: Run go vet ./... across all packages.
- Task 5.2: Run go test -race ./... verifying that all domain tests, opencode tests, git analyzer tests, and repository tests pass cleanly.
- Task 5.3: Verify CGO_ENABLED=0 compilation: build binary using env CGO_ENABLED=0 go build -o bin/agentic-lens ./cmd/agentic-lens.
- Task 5.4: Test live extraction against the active ticket workspace:
  - Run bin/agentic-lens generate --ticket-id AL-11 --json-output workspaces/AL-11/agentic_metrics.json.
  - Verify that workspaces/AL-11/agentic_metrics.json is written with non-zero code facts, valid invocations, and valid tools.
- Task 5.5: Run full verification suite using make check, make test, and make build.

## 4. Key Risks and Defensive Mitigations
1. SQLite CGO Dependency Risk:
   - Risk: Standard Go SQLite drivers require CGO, breaking static compilation, cross-compilation, and single-binary packaging.
   - Mitigation: Mandate modernc.org/sqlite pure Go driver. Enforce CGO_ENABLED=0 in build and CI verification.
2. Database Lock Contention Risk:
   - Risk: Concurrent OpenCode agent processes holding write locks on opencode.db could cause database locked errors in the CLI extractor.
   - Mitigation: Open connections strictly with read-only URI parameters (file:...?...&mode=ro&_pragma=query_only(1)). Use short query timeouts.
3. Git Revision Resolution Drift:
   - Risk: In different branch workflows (unpushed commits, detached HEAD, squashed commits), simple git diff origin/main...HEAD returns empty numstats.
   - Mitigation: Implement multi-tiered resolution: active branch merge-base, git log commit grep, local/remote branch list, and git log --all fallback.
4. Schema Incompatibility Risk:
   - Risk: Extracted JSON output deviating from the schema expected by the web dashboard and repository scanner.
   - Mitigation: Use the existing domain.AgenticFactLedger and domain.AgenticMetricsLedger domain structs directly for serialization. Validate output against existing workspace ledgers.
5. Accidental Emoji Usage Risk:
   - Risk: Introduction of decorative emojis in log messages, CLI output, or markdown generation.
   - Mitigation: Strict zero-emoji review across all code strings, test fixtures, and markdown outputs.

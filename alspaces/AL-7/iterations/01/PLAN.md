# Implementation Plan: AL-7 Domain Anomaly Detection and Rule Engine

## 1. Goal and Iteration Scope

Deliver a pure functional anomaly detection rule engine in the Go domain evaluating deterministic rules against AgenticFactLedger, complete with boundary serialization, markdown summary formatting, CLI diagnostics, repository README documentation covering both existing and new capabilities, and TypeScript domain contract parity. This iteration covers the entire scope of ticket AL-7.

---

## 2. Grounded Reference Architecture and Concrete Paths

The implementation maps to the following existing packages and modules:

### Domain Core Layer (backend/internal/domain/)
- backend/internal/domain/enums.go: Add AnomalySeverity and AnomalyRuleID typed string enums with typed constants.
- backend/internal/domain/anomalies.go: Create pure domain models (Anomaly, AnomalyContext) and pure evaluation engine DetectAnomalies(ledger AgenticFactLedger) []Anomaly.
- backend/internal/domain/models.go: Add Anomalies []Anomaly to AgenticMetricsLedger.
- backend/internal/domain/aggregations.go: Update DeriveMetricsLedger to invoke DetectAnomalies and assign the result to the Anomalies field of AgenticMetricsLedger.

### Boundary Adapters Layer (backend/internal/adapters/)
- backend/internal/adapters/schemas.go: Implement AnomalyDTO and AnomalyContextDTO with bidirectional mapping methods to and from domain types. Update AgenticMetricsLedgerDTO with Anomalies []AnomalyDTO.
- backend/internal/adapters/formatters.go: Update FormatMarkdownMetricsTable to render the "Anomalies & Health Flags" section with severity badges and descriptive context. Ensure UpdateSummaryFile updates this section cleanly.

### CLI Application Layer (backend/cmd/agentic-lens/)
- backend/cmd/agentic-lens/main.go: Add --anomalies flag to root and generate command line flag sets. Add Anomalies column to --list output table. Render anomaly diagnostics when --anomalies is supplied or anomalies are detected.

### TypeScript Presentation Layer (web/src/domain/)
- web/src/domain/types.ts: Add AnomalySeverity, AnomalyRuleID, AnomalyContext, and Anomaly type definitions matching Go models. Add optional readonly anomalies?: readonly Anomaly[] to AgenticMetricsLedger.

### Documentation Layer (README.md)
- README.md: Document the full system architecture, existing capabilities (telemetry extraction from OpenCode SQLite/git, schema v1/v2 with chronological events, backfill engine, single-binary embedded web UI, CLI commands generate/backfill/server/scan/list), and new AL-7 anomaly detection features (pure rule engine, CLI --anomalies flag, anomalies column, markdown summary health section).

### Test Suites
- backend/internal/domain/anomalies_test.go: Comprehensive table-driven unit tests for all four rules, clean sessions, v1 fallback behavior, multi-anomaly sorting, and edge cases. Zero mocks.
- backend/internal/adapters/formatters_test.go: Unit tests asserting correct markdown formatting for clean runs and anomaly runs.
- backend/cmd/agentic-lens/main_test.go: CLI integration tests verifying --anomalies flag, --list column output, and --append markdown updates.

---

## 3. Phased Implementation Tasks

### Phase 1: Domain Core Enums and Models (backend/internal/domain/)
- Task 1.1: Define AnomalySeverity enum type in backend/internal/domain/enums.go:
  - Constant SeverityWarning = "WARNING"
  - Constant SeverityCritical = "CRITICAL"
- Task 1.2: Define AnomalyRuleID enum type in backend/internal/domain/enums.go:
  - Constant RuleToolRetryStreak = "TOOL_RETRY_STREAK"
  - Constant RuleToolHighFailureRate = "TOOL_HIGH_FAILURE_RATE"
  - Constant RuleContextSaturationSpike = "CONTEXT_SATURATION_SPIKE"
  - Constant RuleSubagentReworkLoop = "SUBAGENT_REWORK_LOOP"
- Task 1.3: Create backend/internal/domain/anomalies.go:
  - Define AnomalyContext struct with fields:
    - ToolName of type *string (json tag tool_name,omitempty)
    - SessionID of type *string (json tag session_id,omitempty)
    - Turn of type *int64 (json tag turn,omitempty)
    - Role of type *string (json tag role,omitempty)
    - SecondaryRole of type *string (json tag secondary_role,omitempty)
    - Count of type *int64 (json tag count,omitempty)
    - TotalCalls of type *int64 (json tag total_calls,omitempty)
    - TokenDelta of type *int64 (json tag token_delta,omitempty)
    - ContextTokens of type *int64 (json tag context_tokens,omitempty)
    - Threshold of type *string (json tag threshold,omitempty)
    - FailureRatePct of type *float64 (json tag failure_rate_pct,omitempty)
  - Define Anomaly struct with fields:
    - RuleID of type AnomalyRuleID (json tag rule_id)
    - Severity of type AnomalySeverity (json tag severity)
    - Title of type string (json tag title)
    - Description of type string (json tag description)
    - Context of type AnomalyContext (json tag context)
- Task 1.4: Update AgenticMetricsLedger in backend/internal/domain/models.go:
  - Add field Anomalies of type []Anomaly (json tag anomalies,omitempty).

### Phase 2: Domain Rule Evaluation Engine (backend/internal/domain/anomalies.go)
- Task 2.1: Implement rule evaluation helper detectToolRetryStreaks(events []EventFact) []Anomaly:
  - Traverse chronological tool_call events.
  - Track consecutive error count for the current tool.
  - Reset counter to 0 on tool call completed.
  - When consecutive error count reaches 3: emit WARNING anomaly (or CRITICAL if streak reaches 5 or more).
- Task 2.2: Implement rule evaluation helper detectToolHighFailureRates(ledger AgenticFactLedger) []Anomaly:
  - If ledger.Events is present: compute calls and errors per tool from tool_call events.
  - If ledger.Events is empty: fall back to ledger.Tools []ToolFact.
  - For each tool with calls >= 5 and error rate > 0.20: emit WARNING anomaly (or CRITICAL if error rate >= 0.50).
- Task 2.3: Implement rule evaluation helper detectContextSaturationSpikes(ledger AgenticFactLedger) []Anomaly:
  - If ledger.Events is present: track context tokens per session across successive turns. Flag single-turn deltas > 40k tokens (WARNING) and > 80k tokens (CRITICAL). Flag context token ceiling breaches >= 160k tokens (CRITICAL).
  - If ledger.Events is empty: fall back to checking Invocations for PeakContext >= 160k tokens.
- Task 2.4: Implement rule evaluation helper detectSubagentReworkLoops(ledger AgenticFactLedger) []Anomaly:
  - Extract ordered subagent roles from turn messages or invocations, filtering out orchestrator.
  - Detect alternating ping-pong role pairs with >= 3 consecutive cycles (WARNING) or >= 5 cycles (CRITICAL).
- Task 2.5: Implement pure root function DetectAnomalies(ledger AgenticFactLedger) []Anomaly:
  - Combine results of all four rule helpers.
  - Sort anomalies deterministically: Severity descending (CRITICAL before WARNING), RuleID ascending, Turn/ToolName ascending.
  - Return empty non-nil slice when zero anomalies detected.
- Task 2.6: Update DeriveMetricsLedger in backend/internal/domain/aggregations.go:
  - Invoke DetectAnomalies(factLedger) and assign to Anomalies field on AgenticMetricsLedger.

### Phase 3: Boundary Serialization and DTO Mapping (backend/internal/adapters/)
- Task 3.1: Define AnomalyContextDTO in backend/internal/adapters/schemas.go:
  - Mirror domain.AnomalyContext with JSON tags matching specification.
  - Implement ToDomain() domain.AnomalyContext and standalone AnomalyContextDTOFromDomain(domain.AnomalyContext) AnomalyContextDTO.
- Task 3.2: Define AnomalyDTO in backend/internal/adapters/schemas.go:
  - Mirror domain.Anomaly with RuleID, Severity, Title, Description, and Context.
  - Implement ToDomain() domain.Anomaly and AnomalyDTOFromDomain(domain.Anomaly) AnomalyDTO.
- Task 3.3: Update AgenticMetricsLedgerDTO in backend/internal/adapters/schemas.go:
  - Add field Anomalies of type []AnomalyDTO (json tag anomalies,omitempty).
  - Update ToDomain() and SerializeLedger mapping functions to convert between domain and DTO representations.

### Phase 4: Markdown Summary Formatter and CLI Diagnostics (backend/internal/adapters/, backend/cmd/agentic-lens/)
- Task 4.1: Update FormatMarkdownMetricsTable in backend/internal/adapters/formatters.go:
  - Add "### Anomalies & Health Flags" section directly following "### Workflow Efficiency".
  - If ledger.Anomalies is empty: render "- **Status:** No execution anomalies detected (clean session)."
  - If ledger.Anomalies has entries: render summary line with total count and breakdown by severity.
  - List each anomaly as a bullet: format severity badge, rule identifier in bold, title, description, and key context parameters.
- Task 4.2: Update UpdateSummaryFile in backend/internal/adapters/formatters.go:
  - Ensure section header replacement includes the anomalies section when updating an existing SUMMARY.md file.
- Task 4.3: Update backend/cmd/agentic-lens/main.go:
  - Add anomaliesFlag := flag.Bool("anomalies", false, "Display detected anomaly diagnostic details") to root CLI flags and generate flags.
  - In --list output: add "Anomalies" column to the formatted table (displaying count or severity summary).
  - In runGenerate: evaluate anomalies on the derived metrics ledger. If --anomalies is set, print formatted anomaly diagnostics to stdout.
  - In default single-ticket inspection mode: if --anomalies is passed or anomalies exist, output the anomaly diagnostic table.

### Phase 5: TypeScript Presentation Layer Parity (web/src/domain/)
- Task 5.1: Update web/src/domain/types.ts:
  - Define type AnomalySeverity = 'WARNING' | 'CRITICAL'.
  - Define type AnomalyRuleID = 'TOOL_RETRY_STREAK' | 'TOOL_HIGH_FAILURE_RATE' | 'CONTEXT_SATURATION_SPIKE' | 'SUBAGENT_REWORK_LOOP'.
  - Define interface AnomalyContext with optional readonly fields matching domain models.
  - Define interface Anomaly with readonly rule_id: AnomalyRuleID, severity: AnomalySeverity, title: string, description: string, context: AnomalyContext.
  - Extend AgenticMetricsLedger interface with optional readonly anomalies?: readonly Anomaly[].
- Task 5.2: Verify TypeScript strict compilation across all web modules (npm run typecheck).

### Phase 6: Repository Documentation Update (README.md)
- Task 6.1: Update README.md with comprehensive documentation of existing capabilities:
  - Document OpenCode SQLite session database and git metadata extraction pipeline.
  - Document telemetry schema evolution across v1 aggregate metrics and v2 chronological turn-level event facts.
  - Document automated backfill engine for historical ticket workspace reconciliation.
  - Document single-binary distribution with compiled TypeScript web dashboard embedded via Go embed.FS.
  - Document CLI commands: generate (telemetry ingestion), backfill (historical reconciliation), server (HTTP API and embedded web UI), scan (workspace discovery), and list (tabular metric overview).
- Task 6.2: Update README.md with AL-7 anomaly detection engine documentation:
  - Document pure functional rule engine and its four deterministic rules: TOOL_RETRY_STREAK, TOOL_HIGH_FAILURE_RATE, CONTEXT_SATURATION_SPIKE, and SUBAGENT_REWORK_LOOP.
  - Document anomaly severity classifications (WARNING, CRITICAL) and deterministic ordering semantics.
  - Document CLI diagnostics usage via --anomalies flag and the --list output table anomalies column.
  - Document automated markdown summary integration rendering the Anomalies and Health Flags section in SUMMARY.md.
  - Ensure zero emojis and clear operational instructions across all added documentation sections.

### Phase 7: Automated Verification, Adversarial Testing, and Quality Strategy
- Task 7.1: Unit testing in backend/internal/domain/anomalies_test.go:
  - Clean run test: Ledger with zero errors, small context, normal workflow returns empty slice.
  - Tool retry streak test: 3 consecutive tool errors trigger WARNING; 5 trigger CRITICAL; error followed by success resets streak.
  - Tool high failure rate test: Tool with 6 calls and 2 errors (>20%) triggers WARNING; 6 calls and 4 errors (>50%) triggers CRITICAL; tool with <5 calls ignored even if 100% error.
  - Context saturation spike test: Single-turn jump from 10k to 55k tokens triggers WARNING; jump from 10k to 95k tokens triggers CRITICAL; turn with 165k tokens triggers CRITICAL ceiling breach.
  - Subagent rework loop test: Coder -> tester -> coder -> tester -> coder -> tester sequence triggers WARNING.
  - Legacy v1 fallback test: Ledger with nil Events evaluates tool failure rates and peak context without panic.
  - Empty ledger test: Empty fact ledger returns empty slice without panic.
  - Deterministic sort order test: Multiple mixed anomalies returned in exact severity, rule ID, and turn order.
- Task 7.2: Adapter testing in backend/internal/adapters/formatters_test.go:
  - Test FormatMarkdownMetricsTable with empty anomalies (asserts clean session text).
  - Test FormatMarkdownMetricsTable with populated anomalies (asserts formatted bullets with severity tags).
- Task 7.3: CLI integration testing in backend/cmd/agentic-lens/main_test.go:
  - Test --list output includes Anomalies column.
  - Test --anomalies flag outputs anomaly details.
  - Test --append writes anomalies section into summary file.
- Task 7.4: Full quality gate execution:
  - Execute go vet ./... across the repository.
  - Execute go test -race ./... across all Go packages.
  - Execute npm run typecheck and npm test in web/.
  - Verify zero emojis in code, comments, test logs, output strings, and README.md.
  - Verify zero any or empty interface{} in Go and TypeScript.

---

## 4. Execution Dependencies and Delivery Sequence

1. Task 1.1 to 1.4 (Domain Enums and Models) -> unblocks Domain Rule Evaluators.
2. Task 2.1 to 2.6 (Domain Rule Logic and Unit Tests) -> unblocks Adapters and CLI.
3. Task 3.1 to 3.3 (DTOs and Serialization) -> unblocks Formatter and CLI.
4. Task 4.1 to 4.3 (Markdown Formatter and CLI Flags) -> completes backend slice.
5. Task 5.1 to 5.2 (TypeScript Domain Types) -> establishes frontend contract parity.
6. Task 6.1 to 6.2 (Repository Documentation in README.md) -> synchronizes architectural and operational reference manual.
7. Task 7.1 to 7.4 (Full Test Execution and Verification) -> ensures clean delivery.

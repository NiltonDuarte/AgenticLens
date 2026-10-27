# Ticket AL-29: Conventional Commit Ticket Type Classification in Telemetry Engine

## Summary of Completed Work

### Pure Domain Layer (`backend/internal/domain`)
- **Conventional Commit Types Enum (`enums.go`)**: Defined `ConventionalTicketType` string enum and canonical constants: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`.
- **Canonical Ordering**: Added `CanonicalTicketTypes` slice representing standard conventional commit categories in deterministic order.
- **Pure Normalization & Alias Mapping (`NormalizeTicketType`)**: Trims and lowercases input, mapping common aliases (`feature` -> `feat`, `bug`/`bugfix` -> `fix`, `doc`/`documentation` -> `docs`, `task` -> `chore`, `tests` -> `test`, `performance` -> `perf`) to canonical types.
- **Validation**: Implemented `IsValidTicketType` reporting whether a string maps to a valid conventional commit token.
- **Decoupled Prefix Classification (`config.go`)**:
  - Extended `TicketPrefix` struct to include `TicketType string` (`json:"ticket_type,omitempty"` `yaml:"ticket_type,omitempty"`).
  - Implemented pure, parameter-driven `InferTicketTypeFromPrefix(ticketID string, prefixes []TicketPrefix) (ConventionalTicketType, bool)` with zero hardcoded repository prefixes in domain logic.
  - Added delimiter protection (`-` / `_`) ensuring false prefix substrings (e.g. `POD-1`, `GENTLE-1`) are safely rejected.
  - Added `TicketConfig.InferTicketType(ticketID)` helper method.
  - Populated domain project baseline defaults in `DefaultProjectConfig()` (`GEN` -> `chore`, `PO` -> `docs`).
- **Domain Unit Tests (`enums_test.go`, `config_test.go`)**: Pure table-driven tests verifying alias mapping, prefix isolation, whitespace resilience, delimiter protection, and invalid input rejection with zero mocks and zero dependencies.

### Boundary Adapters Layer (`backend/internal/adapters`)
- **Configuration Loader Extension (`config/loader.go`)**:
  - Extended `rawPrefixItem` for YAML mapping/scalar and JSON object/string unmarshaling to extract `ticket_type`.
  - Propagated `TicketType` across `mergeConfig` into domain `TicketPrefix` instances.
  - Unit tests in `loader_test.go` covering scalar/mapping YAML and JSON configurations.
- **Repository Configuration (`.agenticlens/config.yaml`)**:
  - Configured `ticket_type: chore` for prefix `GEN`.
  - Configured `ticket_type: docs` for prefix `PO`.
- **Deterministic Git Inference (`git/inference.go`)**:
  - Eliminated legacy commit message grepping (`git log --grep`) completely.
  - Implemented branch prefix inspection with `InferTicketTypeFromBranchList`, parsing category prefixes (e.g., `feat/`, `fix/`) and using regex boundary matching to prevent cross-ticket name collisions.
  - Parameterized `InferTicketType(ticketID string, prefixes ...domain.TicketPrefix)`: evaluates branch category first, falling back to caller-supplied configured prefixes only when branch detection does not resolve.
  - Banished silent default to `"feature"`, returning `nil` when no tiers match.
- **Serialization Boundary Testing (`schemas_test.go`)**: Added tests asserting `ticket_type: null` persistence and verifying round-trip JSON serialization across all conventional types and rejecting malformed payloads.
- **Adapter Unit Tests (`inference_test.go`, `analyzer_test.go`)**: Validated branch parsing across nested slashes, remote branches, detached HEAD states, prefix boundaries, and configured prefix backups.

### CLI Application Layer (`backend/cmd/agentic-lens`)
- **Early Configuration Ingestion**: Loaded project configuration via `adaptersconfig.Load("")` early in `runGenerate` and reused downstream.
- **Strict Ticket Type Flag (`--strict-type`)**: Added `--strict-type` flag in `main.go` to enforce explicit classification when types cannot be inferred.
- **Input Normalization & Validation**: Normalized `--ticket-type` input flag; aborts with exit code 1 and lists valid conventional types on unrecognized inputs.
- **Deterministic 4-Tier Fallback Hierarchy**:
  - Tier 1: CLI flag `--ticket-type` or cached ledger facts (highest priority).
  - Tier 2: Git branch prefix.
  - Tier 3: Configured project prefix mapping from `.agenticlens/config.yaml` as backup.
  - Tier 4: Unclassified fallback: Emits `WARNING: Missing ticket type for ticket <TICKET_ID>. Specify --ticket-type explicitly.` and writes `ticket_type: null` in output JSON; under `--strict-type`, emits `Error: missing ticket type for ticket '<TICKET_ID>'. Specify --ticket-type explicitly.` and exits with status 1.
- **CLI Integration Tests (`main_test.go`)**: Comprehensive tests for flag aliases, Tier 1 overriding Tier 2, Tier 2 overriding Tier 3, automatic `PO-*`/`GEN-*` resolution via config, branch inference, `--strict-type` gating, and unclassified warnings.

### Quality & Verification
- **Test Suite**: All backend tests pass with `-race` (`go test -count=1 -race ./...`). Frontend tests pass cleanly (337 vitest tests).
- **Static Invariants**: Zero `any` or empty `interface{}`, zero emojis, and zero skipped tests across the codebase.
- **Ticket Tracking**: Updated `workspaces/plan/TICKETS.md` reflecting active development status.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 10
- **Active Agent Duration:** 3h 51m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 177.1k tokens (89% of 200k healthy ceiling, 3 compactions) | Subagents peak 172.9k (Coder)
- **Role Effort Distribution (Subagents):** Planner 11.5%, Coder 42.1%, Tester 30.4%, Reviewer 15.9%
- **Subagent Turn Velocity & Thrashing:** 50.0 avg turns/run (peak: 91 turns) | Planner 21.3, Coder 84.0, Tester 68.0, Reviewer 41.0
- **Test-to-Production Code Ratio:** 10.3:1 (+2,111 test lines / +205 prod lines)
- **Code Delivered:** +2,318 / -78 lines across 13 code files (excluding docs)
- **Reasoning Effort:** 194,915 reasoning tokens (60.5% of generated output)
- **Tool Health & Self-Correction:** 679 tool calls (15 self-corrections, 97.8% success rate)

### Anomalies & Health Flags
- **Status:** 6 execution anomalies detected (3 CRITICAL, 3 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.0k tokens) — Context tokens reached 160.0k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.2k tokens) — Context tokens reached 160.2k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.5k tokens) — Context tokens reached 160.5k, breaching the 160k critical ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+42.5k tokens) — Context tokens surged by 42.5k to 42.5k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+46.4k tokens) — Context tokens surged by 46.4k to 46.4k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+72.4k tokens) — Context tokens surged by 72.4k to 72.4k tokens.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 2,652,687 | 35,239 | 14,125,355 | 84.2% | 177.1k | ELEVATED | $3.30 |
| **planner** | 3 | 365,901 | 16,981 | 2,008,396 | 84.6% | 79.0k | HEALTHY | $0.65 |
| **coder** | 2 | 1,429,637 | 38,145 | 12,822,367 | 90.0% | 172.9k | ELEVATED | $2.37 |
| **tester** | 2 | 995,931 | 29,449 | 9,039,833 | 90.1% | 168.0k | ELEVATED | $1.71 |
| **reviewer** | 2 | 559,285 | 7,307 | 4,947,419 | 89.8% | 127.4k | ELEVATED | $0.90 |
| **Total** | **10** | **6,003,441** | **127,121** | **42,943,370** | **87.7%** | **177.1k** | **ELEVATED** | **$8.93** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 10 | 6,003,441 | 127,121 | 42,943,370 | 87.7% | $8.93 |
| **Total** | **10** | **6,003,441** | **127,121** | **42,943,370** | **87.7%** | **$8.93** |

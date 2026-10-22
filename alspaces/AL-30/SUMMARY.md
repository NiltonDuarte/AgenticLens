# Ticket AL-30: Fail Telemetry Generation on Unresolved Root Session

## Summary of Completed Work

### Adapters & Shell Layer (`backend/internal/adapters/opencode`)
- **Automatic Workspace Path Fallback (`resolver.go`)**: Enhanced `ResolveActiveRootSession` under Priority 3 (ticket ID matching). When querying `session` by title returns `sql.ErrNoRows`, it automatically falls back to invoking `ResolveSessionByWorkspacePath(db, cleanTicket, 0, 0)`.
- **Error Wrapping**: Wrapped query errors from the workspace path fallback with context `resolving session by workspace path fallback: %w`.
- **Documentation**: Updated function docstrings on `ResolveActiveRootSession` documenting Priority 3 workspace path fallback logic.

### CLI Application Layer (`backend/cmd/agentic-lens`)
- **Strict Root Session Resolution Enforcement (`main.go`)**: Added guard after session extraction in `runGenerate` checking `len(knownThreadIDs) == 0 && len(invocations) == 0 && !*allowEmptyFlag`.
- **Standardized Actionable Error**: Emits error message to `os.Stderr`:
  `Error: unable to identify root session for ticket '<TICKET_ID>'. Provide --session-id, --nonce, or ensure workspace files exist in OpenCode database. Telemetry generation aborted.`
  (where `<TICKET_ID>` falls back to `"UNKNOWN"` when empty).
- **Process Termination**: Aborts execution with exit code 1 (`os.Exit(1)`).
- **Side-Effect Prevention**: Positioned validation gate strictly prior to code metrics computation, ledger derivation, and disk writes (`adapters.SaveFactLedger` and `adapters.UpdateSummaryFile`), ensuring failed invocations never create or mutate output files.
- **Escape Hatch Flag (`--allow-empty`)**: Registered boolean flag `--allow-empty` (strictly defaulting to `false`) to allow isolated synthetic unit tests and dry runs to bypass the live database.

### Quality & Verification
- **Adapter Tests (`extractor_test.go`)**: Added unit tests verifying `ResolveActiveRootSession` resolves sessions via workspace path fallback when titles differ, returns `(nil, nil)` on unresolved tickets, and returns descriptive wrapped errors on database failures.
- **CLI Adversarial Test Suite (`main_test.go`)**: Implemented `TestCLIGenerate_UnresolvedSessionEnforcement` asserting:
  - Exit status 1 and stderr message on unresolved root session.
  - Zero output files (`--json-output`, `--append`) created on failure.
  - Existing summary markdown and JSON ledger files are preserved unmodified on failure.
  - `--allow-empty` allows empty generation to proceed with exit status 0.
  - Valid tickets with identifiable sessions continue to succeed.
  - Missing ticket ID displays `UNKNOWN` in error message.
  - Cached ledger facts with invocations allow generation without DB session.
  - Cached ledger facts with zero invocations fail without DB session.
  - Database errors during fallback fail cleanly with status 1.
- **Synthetic Test Adaptation**: Updated existing tests (`TestCLIAnomaliesCommand`, `TestCLIGenerate_AdversarialEdgeCases`) with `--allow-empty`.
- **Static Invariants**: Zero `any` or empty `interface{}`, zero emojis, and zero skipped tests across the codebase.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 1
- **Active Agent Duration:** 57m 40s (excludes user wait time)
- **Context Window Health:** ELEVATED — Orchestrator peak 108.2k tokens (54% of 200k healthy ceiling, 0 compactions) | Subagents peak 112.1k (Coder)
- **Role Effort Distribution (Subagents):** Planner 19.8%, Coder 30.8%, Tester 27.4%, Reviewer 22.0%
- **Subagent Turn Velocity & Thrashing:** 56.0 avg turns/run (peak: 66 turns) | Planner 42.0, Coder 66.0, Tester 60.0, Reviewer 56.0
- **Test-to-Production Code Ratio:** 3.4:1 (+1,375 test lines / +402 prod lines)
- **Code Delivered:** +1,778 / -149 lines across 24 code files (excluding docs)
- **Reasoning Effort:** 94,802 reasoning tokens (65.4% of generated output)
- **Tool Health & Self-Correction:** 314 tool calls (9 self-corrections, 97.1% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 536,604 | 15,665 | 6,165,226 | 92.0% | 108.2k | ELEVATED | $0.97 |
| **planner** | 1 | 248,884 | 10,367 | 1,833,917 | 88.1% | 89.6k | HEALTHY | $0.45 |
| **coder** | 1 | 401,695 | 10,921 | 3,932,097 | 90.7% | 112.1k | ELEVATED | $0.70 |
| **tester** | 1 | 384,773 | 8,065 | 3,036,282 | 88.8% | 89.3k | HEALTHY | $0.63 |
| **reviewer** | 1 | 288,570 | 5,062 | 2,556,962 | 89.9% | 84.8k | HEALTHY | $0.50 |
| **Total** | **5** | **1,860,526** | **50,080** | **17,524,484** | **90.4%** | **112.1k** | **ELEVATED** | **$3.25** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 5 | 1,860,526 | 50,080 | 17,524,484 | 90.4% | $3.25 |
| **Total** | **5** | **1,860,526** | **50,080** | **17,524,484** | **90.4%** | **$3.25** |

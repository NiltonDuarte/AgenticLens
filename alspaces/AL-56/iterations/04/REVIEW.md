# AL-56 Iteration 04 Review: Session-Scoped Tool Facts and Exact Iterations

## Scope

Reviewed iterations 01–03 changes and the iteration 04 implementation against `CONTRACT.md`, the approved plan, and `TECHNICAL_GUIDELINES.md`. The user explicitly approved the schema-v4 contract tightening: every persisted tool fact requires a nonempty session ID, and legacy unscoped tool totals are rejected rather than assigned fabricated provenance.

## Findings Resolved

### Exact SDLC iteration totals

`runGenerate` now assigns `code.sdlc_iterations` from the exact sum of `user_iterations` on root invocations, including zero. DTO conversion and legacy workflow-efficiency loading preserve zero, and OpenCode backfill records zero steering turns instead of retaining a stale positive count. Regression tests cover positive cached counts replaced by a fresh zero and zero-preserving conversions.

### Session-scoped tool aggregation

OpenCode and Codex extraction now emit per-session tool facts. Generation replaces refreshed facts by `(session_id, tool_name)` and retains other sessions, so distinct-session totals add while repeat refreshes do not duplicate counts. Domain metrics and anomaly aggregation sum session rows by tool name; the dashboard counts unique tool names and keeps the existing per-tool presentation.

### Required v4 session provenance

The v4 DTO boundary rejects missing, null, empty, and whitespace-only `session_id` values. Validation is applied when scanning workspaces, loading a single ticket for the HTTP API, loading an existing generation ledger, saving facts, backfilling, and running schema migrations. Migration validation runs before the already-up-to-date shortcut; invalid source ledgers are rejected without mutation and receive source-log re-extraction guidance. Valid and no-tool migrations remain supported.

The recovery instructions now explain that users must preserve the old ledger, regenerate root sessions into a fresh output path, reuse that path to merge session facts, verify the resulting session IDs/tool rows, and replace the old ledger only after verification. The documented commands include the required ticket title.

## Verification

- The latest full `make` passed after the `LoadTicketWorkspace` validation change: Go vet, all backend race tests including adapters/repository, 719 frontend tests, typecheck/lint, frontend build, and multi-platform builds. `git diff --check` passed after the documentation recovery update.
- A post-change Codex CLI smoke run against local rollout data discovered 11 rollouts and emitted schema v4 with 13 invocations, 26 tool facts with nonempty session IDs, and 70 root user iterations matching `code.sdlc_iterations`. The exported events contained no prompt, reasoning, tool-input/output, or error-message payload fields.
- Repeated snapshots of the selected local Codex thread changed while checking it, so they are not stable-rollout idempotency evidence. The fixed-fixture tests verify repeat-generation idempotency for invocation, iteration, and tool facts.
- Dependency installation reported six npm audit advisories (three moderate, one high, two critical); the checks/build still passed.
- Checked the changed source and tests for AL-56-specific coupling; ticket identity remains fixture/input data and is not used by production logic.

## Verdict

`APPROVED`

The two prior P1 findings are resolved, strict session provenance is enforced across ledger boundaries, and the final single-ticket loader path is covered. The active Codex extraction works with the updated adapter. The running local rollout limited stable-input live idempotency evidence, which is covered by deterministic fixture tests.

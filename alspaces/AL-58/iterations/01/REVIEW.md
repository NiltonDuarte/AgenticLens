# AL-58 Iteration 01 Review

## Verdict: APPROVED

Round 02 re-audited the current implementation against the frozen contract, plan, retained v1.19.5/v2.2.0 runtime evidence, and round 01 findings. No unresolved implementation findings remain. The reported tester run independently confirmed the corrections and passed the focused adapter/CLI tests and shell syntax/cleanup checks.

## Round 01 findings and resolution

### [Major] v2 text content was silently discarded — Resolved

- User and assistant text now maps to `EventFact.PromptContent` in `backend/internal/adapters/opencode/v2.go`.
- `TestV2RuntimeEvidenceExtractsNormalizedFacts` asserts the retained user and assistant text exactly.

### [Major] nonterminal tool states were emitted as completed successes — Resolved

- v2 tool handling accepts `completed` as success and `error`/`failed` as errors. Other statuses return contextual errors instead of producing successful facts.
- `TestV2ToolPendingRunningAndUnknownStatusesFailClosed` covers pending, running, and unknown status values.

### [Medium] equal-time event sorting could reorder source chronology — Resolved

- Final event sorting is stable and compares timestamps only, preserving source append order for equal timestamps. Message reads use `seq`, then timestamp and ID as deterministic ordering keys; tool events are appended after their containing assistant message.
- `TestV2TiedTimestampsPreserveSourceAndSameMessageOrder` asserts user, assistant, and tool event order for tied timestamps.

### [Medium] cleanup could leak temporary state when the provider PID was unset — Resolved

- `provider_pid` is initialized to an empty value before the EXIT trap is installed. Cleanup conditionally stops it and removes the temporary directory.
- The round 02 tester reports cleanup checks with `OPENCODE_V1_BIN=/bin/false` passed, covering failure before provider startup and the cleanup path.

## Contract and architecture audit

- Schema-family inference validates required table and column sets and rejects unknown, partial, mixed, or malformed layouts with diagnostics. `ExtractFacts` and `ExtractSteeringTurns` validate schema before returning for an empty selection; `TestEmptyThreadSelectionStillValidatesSchema` covers this regression.
- v1 and v2 extraction remain in the adapter layer and normalize to existing domain fact types. The retained v2 fixture test asserts two assistant responses, a completed read and result, model, 46 input tokens, 22 output tokens, zero cost, user text, and assistant text. Tool path normalization covers v1 `filePath` and v2 `path`.
- `OpenDB` retains its existing default path and opens SQLite with `mode=ro`; `TestOpenDBRemainsReadOnly` verifies a write is rejected. No version override flag or domain-layer changes were found in the diff.
- Runtime workflows use independent temporary HOME, XDG data/config/state roots, project roots, and absolute database paths. Cleanup is installed before setup can fail. Docs state the evidence window and unverified limitations.
- Reviewed changes contain no Go `any`/`interface{}` or emoji additions. No out-of-scope frontend, telemetry contract, or database mutation artifacts were found. The existing plan/ticket edits and other working-tree changes were left untouched by this review.

## Verification evidence

- Reviewed the full current diff, `CONTRACT.md`, iteration `PLAN.md`, this review's round 01 findings, runtime evidence and fixtures, `schema.go`, `v2.go`, `v2_test.go`, the v1 extractor/resolver changes, `client.go`, compatibility workflow, Makefile, and documentation.
- Round 02 tester reports passing `cd backend && go test ./internal/adapters/opencode ./cmd/agentic-lens`, `bash -n scripts/opencode-compat.sh`, and cleanup checks using `OPENCODE_V1_BIN=/bin/false`. The coder's prior report states broader checks also passed. These test commands were not rerun during this read-only review.
- `git diff --check` passed during this review. ShellCheck is unavailable, as reported by the tester.

## Remaining findings

None.

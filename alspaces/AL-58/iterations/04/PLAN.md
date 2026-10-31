# AL-58 Iteration 04 Delta Plan: Resolve Unknown Session Users from Trusted Identity Sources

## Goal and boundary

Fix Codex telemetry generation and backfill when a previously migrated session contains the placeholder user `unknown`. Treat a trimmed, case-insensitive `unknown` value the same as an absent identity. Prefer a known identity already attached to the same session, then use the existing Git `user.name` fallback when it is known, and retain `unknown` only when neither source supplies a trustworthy identity. Preserve the existing session and telemetry schemas, migration history, source database read-only behavior, and identity behavior for sessions that already have a known user.

This iteration covers identity resolution in generation, Codex backfill, regression coverage, and a source-backed repair of the affected AL-58 ledger. It does not change OpenCode extraction, infer identity from message content, alter other providers' identity rules, or rewrite the v3-to-v4 migration placeholder.

## Grounding

- `backend/internal/domain/migration/step_v3_v4.go` assigns `unknown` when a migrated session has no user. This is an intentional compatibility placeholder and remains unchanged.
- In `backend/cmd/agentic-lens/main.go`, the generation flow builds `userBySessionID` from existing session facts and resolves newly extracted Codex sessions near the `freshCodex` handling block. The session-specific map currently accepts `unknown`; the fallback checks only for a non-empty mapped value, so the placeholder prevents the available `currentUser` from being used. `currentUser` comes from `backend/internal/adapters/git/inference.go` through `git.GetGitUser()`.
- In `backend/internal/adapters/backfill/engine.go`, Codex processing constructs `existingUsers` from prior session facts and reconciles freshly extracted sessions. The map currently accepts any non-empty value, including `unknown`; when the fresh session also has no known identity, this can copy the placeholder instead of the available `gitUser`.
- `backend/internal/adapters/git/inference.go` returns `unknown` when Git has no configured `user.name`, so the fallback itself must be checked using the same known-identity rule.
- The telemetry `SessionFact.User` field is already present in the normalized domain model. This fix should remain in the existing CLI/backfill imperative-shell resolution paths and should not add database or filesystem work to the domain layer.

## Planned changes

### Shared identity validity rule

1. Define or reuse a small helper at the appropriate imperative boundary that classifies an identity as known only when trimming whitespace leaves a non-empty value other than `unknown`, compared without case sensitivity.
2. Apply that rule consistently to freshly extracted values, cached session values, and the Git fallback. Preserve the existing output spelling of known names apart from trimming whitespace where the current resolution path already trims values.
3. Do not interpret arbitrary strings such as `Unknown User` as the sentinel; the reserved placeholder is the exact word `unknown`, ignoring case and surrounding whitespace.

### Generation identity resolution

1. In `backend/cmd/agentic-lens/main.go`, avoid inserting unknown cached users into `userBySessionID`, or ensure they are ignored at lookup. A known cached identity for the same session remains the first choice.
2. For a fresh Codex session with a blank or unknown identity, use a known session-specific identity if present; otherwise use `currentUser` only when it passes the known-identity rule.
3. If neither source is known, retain the normalized placeholder `unknown`. Never replace a known identity supplied by the extractor with a different identity from the cache or Git configuration.

### Codex backfill identity resolution

1. In `backend/internal/adapters/backfill/engine.go`, include only known prior session identities in `existingUsers`.
2. For a refreshed Codex session with no known extracted identity, preserve a known identity for that same session when available; otherwise use the known `gitUser` already obtained by the backfill flow.
3. Retain `unknown` when neither the refreshed fact, prior session fact, nor Git configuration has a known identity. Do not let a placeholder overwrite any known session-specific user.
4. Keep non-Codex provider handling and backfill selection behavior unchanged.

### Regression coverage

1. Add generation coverage proving an `unknown` cached session user falls through to a known Git identity.
2. Assert a known same-session cached identity takes precedence over the Git identity, and a known freshly extracted identity is preserved.
3. Add backfill coverage proving a prior `unknown` user does not block a known Git fallback for a refreshed Codex session.
4. Assert backfill preserves a known prior session-specific user ahead of the Git fallback and retains `unknown` when all identity sources are blank or unknown.
5. Cover trimming and case-insensitive sentinel recognition, including that a distinct name containing the word “unknown” is not treated as the sentinel.
6. Keep tests deterministic and local. Do not require a live OpenCode installation or mutate an OpenCode database.

### Repair of the existing AL-58 ledger

1. After the implementation is in place, use the supported AgenticLens Codex backfill/generation workflow against the original Codex source evidence for the affected session and regenerate the AL-58 telemetry ledger through the CLI. Confirm the exact supported command and overwrite/append semantics from the current CLI help and implementation before running it.
2. Ensure the repair reads the Codex source database read-only and updates only the AgenticLens-owned AL-58 `agentic_metrics.json` and its documented summary artifacts. Do not hand-edit JSON or alter the OpenCode database.
3. Confirm the regenerated record for session `01a184e0-3822-7631-a3f3-cb33bb3e947f` carries the known Git identity when available. If no trusted identity source is present in the reproducible run, keep `unknown` and record that the available evidence cannot recover the identity.
4. Preserve unrelated ledger facts and follow the workspace's final-step telemetry-ingestion rule for any new iteration telemetry; do not run telemetry before all implementation, verification, and artifact updates are complete.

## Verification

- Run focused generation and backfill tests that exercise the identity precedence and sentinel cases.
- Run `make test` for the repository suites.
- Run the supported Codex backfill/regeneration flow on the isolated AL-58 evidence and inspect the target session record plus neighboring records for unintended changes.
- Confirm the OpenCode source DB remains unmodified and that no other source-provider behavior changed.
- Run `git diff --check`.

## Acceptance criteria

1. Blank, whitespace-only, and case variants of the exact `unknown` sentinel are treated as missing identities in both Codex generation and backfill.
2. Identity precedence is: known freshly extracted session identity; otherwise known identity already associated with that same session; otherwise known Git identity; otherwise `unknown`. Existing known extracted identities are never overwritten.
3. Regression tests cover cached `unknown` falling through to Git, session-specific known identity preference, known fresh identity preservation, and final `unknown` fallback when no trusted identity exists.
4. The migration's `unknown` placeholder, normalized telemetry schema, OpenCode extraction, and read-only source database boundary remain unchanged.
5. The AL-58 ledger is repaired only by supported source-backed AgenticLens processing, with the affected session verified and no manual JSON edits.
6. Focused tests, `make test`, source/ledger checks, and `git diff --check` pass.

## Handoff

This is a focused identity-resolution correction for AL-58. Implementation starts after the user approves this iteration plan.

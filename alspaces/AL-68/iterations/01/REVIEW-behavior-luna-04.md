# AL-68 Behavior and Integration Re-review — behavior-luna-04

## Verdict

**Scoped code verdict: APPROVED. Overall acceptance: INCOMPLETE.** No blocking behavior defect was substantiated in the named CAP-01/03/04/11/12 correction interactions on this candidate. Required real Codex trust, callback, agent-binding and source comparison evidence is still unavailable, so this review does not establish Codex runtime support or complete acceptance.

## Revision, scope, and identity

- Candidate: `CANDIDATE-04.json`; requested fingerprint `c34cc949278608f4c777f7428a620e2ae29542e7fe33875d5b13ae9f0d17e6ec`; baseline `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Independently recomputed the candidate's sorted path/NUL/content/NUL digest from its listed repository paths; it exactly matches the requested fingerprint.
- Requested reviewer profile: `codex / gpt-6-luna / medium`. This is the requested setting; I do not claim independent source-record confirmation of model, effort, usage, or cost.
- Read the behavior persona, shared review procedure, approved contract/plan and the specified correction, finding, implementation, test, smoke, setup and verification evidence. Did not read Sol-04 closing conclusions. Product and test files remained read-only; only this report was written. No child agents, full-suite run, commit, stage, telemetry, publication or user-controlled trust action.
- Review boundary: recheck named counterexamples and their affected ownership/recovery interactions; not a whole-repository audit.

## Corrected IDs and evidence

| ID | Disposition | Behavioral basis |
| --- | --- | --- |
| CAP-01 | Fixed | Registered refreshes are accepted against the canonical owner using source-derived invocation completion checkpoints. The explicit-selection path applies the same rule. Differing facts require a strictly later checkpoint; equal or unknown checkpoints preserve the canonical owner and expose unresolved coverage. Identical facts and genuinely later facts are accepted. The implementation does not infer precedence from totals, ledger time, filesystem time, or callback arrival. This covers the previously reported successful reread of old source bytes after a newer explicit publication. |
| CAP-03 | Fixed for known actual-store conflicts | Explicit Codex/OpenCode selections are checked against exact source/session ownership before merge/publication. `SameStore` resolves existing symlinks and also compares filesystem identity. Registry changes inspect legacy alias rows inside the write transaction and reject duplicate actual-store identities instead of selecting one arbitrarily. The regression covers both stored-alias directions and permits the same ID in a genuinely distinct store. |
| CAP-04 | Fixed for source-proven recovery | `SubagentStart` stores only opaque agent ID, callback parent and source-store evidence. Recovery discovers an actual source session and requires its authoritative parent to equal the registered parent before inheriting the ticket. The interface ID is never treated as a source ID. The start-without-stop regression checks this path and preserves unresolved evidence. |
| CAP-11 | Fixed | Ticket reconciliation refreshes and publishes only owners bound to that ticket. Unassigned registrations remain visible locally, and do not consume another ticket's extraction budget or enter its report. Source-proven child discovery is the bounded route from a bound parent into that ticket's scope. |
| CAP-12 | Fixed for both reported starvation schedules | Durable work-class phase rotation gives refresh, queued-stop recovery, start metadata discovery and OpenCode child discovery each an eventual first chance with the full configured session/byte allowance. Persistent failed or oversized work therefore cannot monopolize every pass. Per-class cursors preserve continuation, and unresolved entries remain pending. |

### Hook-to-agent identity and ticket binding

The OpenCode binding flow follows the actual plugin interface recorded in the evidence: `experimental.chat.system.transform` supplies per-session agent context; `agenticlens_bind` takes the runtime `context.sessionID` and exact request token. The real OpenCode 1.19.5 Linux smoke recorded independent roots, a native child, completed binding exchanges and source-verified parent/usage facts. It is valid evidence for that tested interface and platform only.

Codex source handling keeps callback `session_id` as the parent on `SubagentStart`; the callback's opaque `agent_id` does not become the child identity. The documented `additionalContext` request is limited to the exact unassigned session on `SessionStart`/`UserPromptSubmit`. Source-level checks and fixtures support the no-guessing behavior, but there was no real trusted Codex callback/agent exchange or source comparison. No trust setting was changed or bypassed.

## Verification

I ran the targeted affected race tests from `backend/` (all passed):

- `go test -race ./internal/domain -run '^TestCaptureSnapshotPrecedenceRejectsTruncationAndOlderSource$' -count=1`
- `go test -race ./internal/adapters/capture -run '^Test(QueuedCodexChildRecoveryContinuesPastMalformedAndMismatchedMetadata|ReconcileSchedulerEventuallyRefreshesOwnerWhenFailedMetadataConsumesByteAllowance|ReconcileSchedulerEventuallyDiscoversStartOnlyChildWithPersistentFailedStop|ReconcileRecoversSourceProvenChildAfterStartWithoutStop)$' -count=1`
- `go test -race ./cmd/agentic-lens -run '^TestCaptureGeneration(ResolvesStoreAliasesWithoutConflatingDistinctStores|RejectsExplicitSelectionOfAnotherTicketOwner|RetainedRegistryCannotOverrideNewerCanonicalOrExplicitSnapshot)$' -count=1`

The targeted tests returned `ok` for domain, capture and CLI. These exercise the equal/older snapshot guard, actual-store aliases and distinct-store control, both CAP-12 schedules, start-only source-proven recovery, and conflicting explicit selection. I reused TEST_REPORT.md's coordinated six-package race results, all 14 distribution tests and parent-reported integrated `make` rather than duplicating broader suites. The evidence records later parent integration/make; that verification is outside this reviewer run.

## Remaining limit

The concrete remaining gap is user-controlled real Codex trust/callback/agent-binding/native-child source evidence. The OpenCode 1.19.5 Linux interaction does not close it. Other platform and harness-version claims remain limited to the explicit evidence in SMOKE.md. With that acceptance evidence unavailable, the scoped code corrections pass, while the overall result remains **INCOMPLETE**.

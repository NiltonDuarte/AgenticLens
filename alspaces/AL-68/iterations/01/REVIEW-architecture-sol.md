# AL-68 architecture review

Outcome: CHANGES_REQUESTED. Independent first pass, run `architecture-sol-01`.

Requested profile: `codex/gpt-6.1-sol/high`. Actual runtime model, effort, usage and cost are not exposed to this reviewer and are not inferred from the requested profile or persona.

Baseline HEAD: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`. Reviewed working-tree candidate: `c7dade3313a1d65a4269d595c782750655be8a044c020dc1f3ac502464cf325a`. Recomputed the CANDIDATE.json path/NUL/content/NUL hash and obtained an exact match both before verification and after writing this report. Product code and tests remained read-only; this report is the only repository output owned by this reviewer. No sibling review was read.

## Findings

### A1 — P1: Retained capture snapshots override successful explicit source extraction

Location: `backend/cmd/agentic-lens/main.go:312`, using `mergeSessions` at line 1464; retained snapshots originate in `backend/internal/adapters/capture/reconcile.go:135` and are returned at line 255 even when their refresh failed.

Trigger: A registered owner has an accepted snapshot, its registered transcript path becomes unavailable, and `generate --session-id` or `--nonce` successfully discovers a newer authoritative transcript for the same owner. Codex discovery explicitly supports `sessions` and `archived_sessions`. The same ordering problem applies when capture's byte budget rejects a large owner while explicit extraction succeeds.

Consequence: `mergeSessions(freshSessions, captureResult.Sessions)` gives the registry's retained snapshot unconditional precedence over the newly extracted source facts. The command exits successfully and persists stale usage. Keeping old facts on source failure is required, but those facts must not replace a successfully refreshed owner from the other supported extraction path. Snapshot authority is currently decided independently in registry reconciliation and command composition.

Executed evidence: built the frozen source with `go build -o /tmp/al68-architecture-sol-review ./cmd/agentic-lens` from backend. A disposable native CLI fixture at `/tmp/al68-arch-explicit-jxnvloh3` registered and explicitly bound `root`, then reconciled an input-token snapshot of 10. It moved the source from `source/sessions/root.jsonl` to `source/archived_sessions/root.jsonl` and updated its authoritative usage to 20. `generate --source codex --session-id root --codex-home <fixture source> --capture-dir <fixture state> --ticket-id TEST-100 --ticket-title Probe --ticket-type feat --json-output <fixture ledger>` exited 0. The published canonical v5 session contained input 10; the sidecar reported the old registered path's owner as failed. Direct discovery had succeeded, so this is observable stale publication, not merely unavailable-source preservation. Request tokens and raw source contents are not included here.

Required correction: establish a single precedence rule for duplicate owners across capture and explicit extraction. Retained failed/pending snapshots may supply otherwise unavailable owners; they must not overwrite a successful fresher extraction. Add a self-contained regression that reaches both paths for the same owner and checks the published ledger, including the unavailable registered-path case.

### A2 — P2: Failed source parses bypass the reconciliation pass byte budget

Location: `backend/internal/adapters/capture/reconcile.go:133` through line 144; Codex extraction returns zero revision after fully reading a warning-bearing source at lines 286–292.

Trigger: Two or more bound registered sources each fit the remaining byte budget individually but contain malformed source records. `extractExact` reads a source before returning its parse failure. The failure branch returns without charging any bytes; only successful reads reach `byteBudget -= revision`.

Consequence: The documented per-pass source byte limit does not bound actual source reads on failure, precisely when retries are likely. Each owner can consume the full configured allowance again. A bounded session count does not satisfy the separate byte budget in CONTRACT.md. The same accounting structure discards OpenCode's measured source size when extraction returns an error.

Executed evidence: disposable native CLI fixture `/tmp/al68-arch-budget-3j7unihg` registered and bound `owner-a` and `owner-b`. Each valid-identity Codex source was 411 bytes and ended with a malformed record. One `capture reconcile --ticket-id TEST-100 --max-source-bytes 411 --max-sessions 2` parsed both files, returned both owners as failed, and emitted no byte-budget limit. Reaching both warning-bearing parse failures requires reading 822 source bytes under a 411-byte allowance. A correctly charged first read would leave the second owner pending without parsing it.

Required correction: make consumed/reserved source bytes independent of snapshot revision and extraction success. Charge failed reads against the same pass budget and leave unprocessed owners pending with continuation. Preserve the existing snapshot-preservation behavior. Extend the existing successful-read budget regression to multiple failed reads.

## Assessed architecture hypotheses

- FCIS: the new domain module uses value models, deterministic binding and timestamp comparison; no I/O, clock access or outward adapter dependency was found. Clocks, SQLite, filesystem, process execution and template embedding remain in adapters/commands. No layering blocker found in the inspected paths.
- Source identity and ownership: hooks do not infer tickets from titles, paths, roles or times. OpenCode binding gets its session ID from runtime tool context; Codex parent callback IDs are not treated as child identities. Exact OpenCode extraction reuses existing schema-family parsers and refuses IDs owned by multiple complete families. The general title/workspace fallback is not used by registry extraction. Proven-parent propagation is source/store scoped. No additional ownership finding established in this focus.
- Canonical preservation: canonical schema version remains v5; capture installation/correlation metadata lives in the local registry and separately versioned report. Existing source parsers and AL-62 skill/task extraction remain reused. Report generation copies registrations before redacting request/store fields. A digest ties the sidecar to serialized ledger bytes; a sidecar write error produces a nonzero publication result. A1 exposes a distinct precedence defect despite those protections.
- Embedded harness adapters: the OpenCode plugin is embedded in the Go binary and uses native argument-vector execution rather than a network/updating wrapper. Lifecycle delivery and agent feedback use separate channels. SMOKE.md contains a real OpenCode 1.19.5 independent-root/native-child binding exchange compared against source records. This supports the reported Linux CLI path only. Codex callback/agent interaction remains unverified pending human trust review; no trust was changed or bypassed here.
- Installation ownership: additive Codex groups and OpenCode content hashes protect unrelated configuration and edited owned files; handler protocol is checked before installation writes. Setup/downloaders delegate lifecycle installation to the native handler; wrappers' capture branch bypasses updates/downloads. No further blocker established by the inspected successful/conflict paths. The tester explicitly leaves fault injection after the first installer write unexecuted, so rollback inspection is not claimed as a runtime fault-path proof.
- Handoff contention remains a follow-up risk rather than an additional substantiated finding in this bounded pass: `Store.Change` acquires `BEGIN IMMEDIATE` before invoking the reconciliation closure, and that closure performs full source extraction while holding the registry writer. Callback writes use the same registry with a 1000 ms busy timeout. Long source extraction can therefore contend with handoff; callback loss under an actual prolonged extraction was not measured in this review. Existing concurrent-writer tests do not establish that latency boundary.

## Verification and limits

Read the contract, approved iteration plan and user no-guessing/actual-interaction clarification; assigned persona/shared procedure; technical/regression guidance; production diff and nonignored untracked capture files; relevant unchanged Codex discovery/parser, OpenCode client/family dispatch/extractors and canonical command merge. Read TEST_REPORT.md, SMOKE.md and SETUP_IMPLEMENTATION.md as evidence rather than conclusions.

Executed the native build and two isolated CLI probes above. Reused the tester's affected-package race checks and the parent's full `make` run recorded in VERIFICATION.md on the same frozen candidate: Go race/vet, frontend typecheck/lint/build, 781 frontend tests, distribution/release tests and native builds passed. Those green checks do not detect A1 or A2, so no full suite was repeated.

Required real Codex trust/callback/binding/native-child evidence is still pending. Other platforms, OpenCode v2 native execution, resumed-session interaction and real interrupted-hook recovery remain outside the real smoke evidence. These are explicit acceptance gaps, not inferred support. The parent owns consolidation, fixes, revised-candidate verification and delivery; this scoped review does not approve merging or claim whole-repository correctness.

# AL-68 data integrity review: data-sol-01

Verdict: **CHANGES_REQUESTED**. Three reproduced defects violate persisted-fact preservation, publication ordering and exclusive ticket ownership. This verdict covers the assigned data/failure scope only.

## Revision, requirements and identity

- Phase: independent discovery, first pass. Baseline HEAD: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Reviewed candidate: `c7dade3313a1d65a4269d595c782750655be8a044c020dc1f3ac502464cf325a`. Independently recomputed the CANDIDATE.json path/NUL/content/NUL hash and obtained that exact fingerprint. The candidate includes untracked capture product/test files; this was not a tracked-diff-only review.
- Requirements: CONTRACT.md, iteration 01 PLAN.md including the explicit user instruction that hooks interact with the agent and never guess, and the AL-68 ticket specification. Read the assigned data persona, shared review context, canonical review skill and applicable technical/regression rules. No sibling review/persona conclusions were loaded.
- Requested execution: `codex/gpt-6.1-sol/high`, run `data-sol-01`. Observed environment: Codex agent tool interface and shared Linux checkout. Actual runtime model and reasoning-effort metadata were not exposed by the available runtime evidence; requested settings are not independently observed settings. No session ID, usage or cost is invented.
- Product and tests remained read-only. The only repository write owned by this reviewer is this report. Probes used synthetic source fixtures and a disposable native binary outside the checkout.

## Findings

### D1 — P1: a failed registered refresh overwrites newer persisted facts with an older local snapshot

Locations: `backend/internal/adapters/capture/reconcile.go:247` through the unconditional snapshot collection at line 255; `backend/cmd/agentic-lens/main.go:312` and `:333`.

Trigger: the local registry has an accepted cumulative snapshot older than the canonical ledger, and the next registered refresh cannot read its source. The canonical ledger can legitimately advance independently through the retained explicit generation path, a different configured registry, or existing backfill.

Reproduction against the candidate native CLI:

1. Deliver SessionStart for exact synthetic Codex session `root`, bind its challenge to `TEST-100`, and generate through registry A. Source-derived input usage is 10.
2. Grow the same valid source to input usage 30. Generate the same canonical ledger with `--capture-dir <absent-registry> --session-id root --codex-home <source-store>`. The resulting persisted input usage is 30; registry A remains at 10.
3. Remove the source file. Generate the same ledger through registry A without an explicit selector.
4. The command exits zero, reports the registration as `failed`, and replaces the persisted input usage 30 with 10.

Observed output: `initial_registered_generation 10`, `newer_explicit_generation 30`, `registered_generation_missing_source 10`, `source_state failed`.

The registry correctly retains its own cached facts on extraction failure, but ReconcileResult exposes those cached rows as authoritative fresh sessions. Generation then overwrites both explicit/cached rows by ID without source-checkpoint provenance or refresh disposition. A failed, pending or skipped refresh must preserve a newer canonical owner. Keep fallback facts distinct from successful authoritative replacements and establish precedence when both registry and ledger contain that owner. Existing tests compare a registry-derived ledger with the same registry snapshot; they do not exercise an independently newer persisted ledger.

### D2 — P1: an older generation can publish after and overwrite a newer accepted snapshot

Locations: `backend/cmd/agentic-lens/main.go:196`, `:453` and `:458`; `backend/cmd/agentic-lens/capture.go:235` and `:241`. The per-owner write transaction in `backend/internal/adapters/capture/store.go:100` ends before final publication.

Trigger: two final-generation processes for the same ticket/output overlap while the source grows. Reconciliation serializes individual registry changes, but there is no publication lock or checkpoint comparison covering ledger load, reconciliation, ledger save and report publication.

Deterministic isolated schedule:

1. Process A reconciles source input usage 10, then pauses at the subsequent `git rev-parse --show-toplevel` code-facts boundary. A disposable PATH shim forwards every Git call to `/usr/bin/git` and pauses only that invocation; no product mutation or source-parser replacement is involved.
2. Grow the valid source to input usage 50. Process B reconciles and publishes the same ticket/output successfully; the canonical ledger now contains 50.
3. Release A. It exits zero and publishes its earlier result, replacing the canonical ledger with usage 10. The durable registry still holds the accepted 50 snapshot.

Observed output: `concurrent_newer_generation 50`, `older_generation_exit 0`, `concurrent_final_ledger 10`.

The report digest does not prevent this: the older process can publish an internally matching but regressed ledger/report pair. Protect final publication with recoverable ownership and snapshot precedence, including reload/merge of the persisted baseline within that protection. Merely serializing each SQLite owner update is insufficient. The existing concurrent-writer/crash tests cover registration transactions, not this publication schedule.

### D3 — P1: an explicit selector bypasses an existing conflicting ticket binding

Locations: `backend/internal/adapters/capture/reconcile.go:110` and `:241` filter other-ticket registrations; `backend/cmd/agentic-lens/main.go:201` falls through to explicit extraction without checking the selected owners' registered bindings.

Trigger: an exact source session is already bound to ticket A, then generation for ticket B explicitly selects that session with `--session-id` or the equivalent nonce path. This is a known ownership conflict, rather than an unregistered legacy session.

Reproduction: register and explicitly bind `root` to `TEST-100`, then run:

```text
agentic-lens generate --source codex --capture-dir <same-registry> \
  --ticket-id TEST-200 --ticket-title 'Other ticket' --ticket-type feat \
  --json-output <other-ledger> --session-id root --codex-home <exact-store>
```

The command exits zero. The new canonical ledger has `ticket_id: TEST-200` and session `root`, while capture status still says `root` is bound to `TEST-100`. The generated TEST-200 capture report contains zero registrations, so the known conflict is omitted entirely. Observed output: `published_ticket TEST-200 published_sessions ['root'] report_registration_count 0`, `registry_ticket TEST-100`.

Retaining explicit nonce/session selection must not bypass a conflicting exact binding. Validate selected root/child ownership against all relevant registrations before publishing; fail visibly when the selected owner belongs to a different ticket. Existing Bind tests reject this conflict only through the bind command.

## Hypotheses assessed and evidence reused

- Exact source selection: traced Codex hook normalization, deferred child source verification, exact transcript ID comparison and native-parent extraction. Traced OpenCode capture-only schema selection, exact session membership and the single-owner v1/v2 extraction options. The new registered path does not use the existing title/workspace fallback. The explicit publication path still has D3.
- Source/store isolation: registration keys include harness, store and exact ID. Parent inheritance includes the same harness/store scope. Simultaneously registered indistinguishable canonical IDs are rejected. Schema v5 remains the canonical format; no migration to v6 was introduced. Reviewed the existing ledger load/save, session merge, v5 cost migration and source extraction dependencies where relevant.
- Snapshot preservation: accepted registry revisions reject shrinking/older source checkpoints; warning-bearing Codex parses and OpenCode extraction errors retain the local snapshot. Missing and ambiguous owners are covered by the tester's real SQLite fixtures. These protections hold locally and fail at the final merge/publication boundaries in D1/D2.
- Retry/idempotence and budgets: inspected cumulative replacement, persisted continuation rotation, shared byte/session limits and pending states. Reused the unchanged/resumed/truncated and successive bounded-pass test evidence. No additional whole-suite execution was warranted.
- Concurrency/crash recovery: inspected SQLite BEGIN IMMEDIATE read/modify/write and rollback paths. Reused TEST_REPORT.md evidence for independent connections, aborted transactions and subprocess termination/reopen. This supports local registration durability; it does not cover D2.
- Unknown facts: inspected source-derived unavailable-cost handling, v5 preservation and report/metric separation. TEST_REPORT.md establishes that a missing Codex cost remains unavailable, and SMOKE.md explicitly limits unavailable effort/runtime claims. No source usage/model/cost was inferred from the requested review profile.
- Final privacy and pair consistency: inspected normalized storage, redacted report copies, digest generation and report temp-file sync/rename. Existing fixture evidence validates a sequential matching digest and avoids publishing store paths/challenges. Those assertions do not establish publication ordering or prevent D1/D3.

## Verification and limits

- Built the candidate directly with `go build -o /tmp/al68-data-review-cli ./cmd/agentic-lens` from backend; exit zero. Executed the isolated Python/CLI probes above; all reproduced. Reproduction driver: `/tmp/al68-data-probe.py`; first fixture directory: `/tmp/al68-data-probe-roac3yhm`. These contain synthetic records only and remain outside Git.
- Reused raw TEST_REPORT.md, SMOKE.md and SETUP_IMPLEMENTATION.md evidence. Parent VERIFICATION.md records full `make` exit zero on the same unchanged candidate fingerprint; no duplicate broad suite was run here. Passing suites do not refute the schedules above.
- Real OpenCode evidence documents two independent roots, actual binding-tool calls and a source-proven native child with unchanged repeated snapshots. It establishes that specific version/interface path, rather than all source families/platforms.
- Real Codex trust, callbacks and agent binding remain pending. No trust bypass, user request or fabricated completion was attempted. Real resumed-harness interaction and additional platform/schema-family compatibility remain unproven as recorded in SMOKE.md.
- The three reproduced product defects independently block approval. The known real-Codex evidence gap additionally prevents a complete supported-harness claim. The parent owns corrections, consolidation, revised-candidate verification, telemetry and publication.

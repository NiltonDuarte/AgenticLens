# AL-84 iteration 01 behavior review — CHANGES_REQUESTED

Fresh independent diff review of `163181428ee9a5a599a2da1e91d3af5e618a9682` → `2b4118736b5bde24d17cc382da7445c8046ce884` for PR144. Role: reviewer-behavior. Runtime identifies gpt-6-astra, high effort, matching the request. No separate run identifier was supplied. Requirements, raw source/diff and hypotheses preceded reading TEST_RESULTS.md; no sibling or prior review reports were read. Shared tasks.md modifications were preserved. Product and test sources were not edited.

## Finding AB-01: pending homes disappear from uniqueness and absence decisions

**Introduced; severity 1 (P1); unresolved.** Locations: `backend/internal/adapters/workspace_locations.go:413–425,450–452,559–565`. A complete directory enumeration is incorrectly treated as complete identity knowledge even when a discovered current-format ledger cannot be read because its publication is pending.

Reachable trigger: configure a depth-1 root containing two valid current-format ledgers A and B with the same recorded ticket. Initially implicit resolution correctly reports ambiguity. Add a current-format pending indicator to B (for example, an interrupted copied workspace without its original private state). `readWorkspaceLedgerSnapshot` rejects B, but the resolver only appends a diagnostic and continues; `incomplete` remains false. It then resolves A successfully. This reaches implicit CLI detail/generate, HTTP detail/events and automatic capture publication (`capture_worker.go:225`). The latter can refresh a copy the user never explicitly selected. If B is the sole existing ledger, `ValidateInitialLedgerDestination` likewise ignores its read error and admits first generation at a different empty home, including a recreated former home after a move.

Expected: partial fleet listing may retain A with diagnostics; an unresolved pending home must not establish unique identity or absence. Require explicit existing-location selection or recovery before these implicit decisions. Preserve the distinction between a definitively invalid ledger and an operationally unavailable current ledger, so invalid unrelated siblings do not unnecessarily block valid reads.

Concrete proof: disposable Go overlay `/tmp/al84-astra-behavior-34mo_h84/overlay.json`, source `probe_test.go`, uses isolated temporary schema-6 fixtures and marker `{"version":2,"transaction":"0123456789abcdef0123456789abcdef"}`. `TestAstraPendingIdentity` first verifies the ordinary duplicate control rejects; after adding the marker, resolution returns A with **nil error**, plus B's incomplete-publication diagnostic. Removing A's ledger then makes first-destination validation at C return **nil**. Both expected rejection assertions fail, exit 1.

Refuting guards checked: CheckPublication correctly rejects B; explicit B selection correctly fails; directory-budget incompleteness correctly blocks implicit resolution. None changes `incomplete` for operational read errors. Generation's later selected-ledger identity, scope and ownership checks validate A or C, not unresolved B; they do not repair the incorrect global selection. Maintenance's identity-only resolver differs and does not exhibit this exact pending-marker elision.

## Other hypotheses and verification

- Folder-derived identity, root-order duplicate precedence, incomplete-budget selection, stale explicit selection, and private-state-dependent portable reads: inspected resolver/configuration/repository/HTTP paths; existing focused tests pass. Physical directory retention and final identity checks guard replacement reads.
- Dashboard location loss: traced list/comparison/scatter selection through App and deep dive to detail/events query parameters; location keys and stale-request guards preserve explicit selection. Live navigation supplies identity and reaches the same ambiguity/recovery selector rather than choosing a copy itself.
- Maintenance caller mismatch: inspected format, backfill and nested migration selection; identity-only admission and owned final validation preserve historical migration access and explicit selection.
- Pages caller mismatch: inspected builder, Pages config, workflow and regression cases. Configured `roadmap_path` supports the trusted no-flag caller; explicit flag overrides it; containment remains enforced. Reused reported final Pages/hosted checks; did not rerun Ruby builds.

Executed from backend: `go test ./internal/adapters ./internal/adapters/http ./internal/adapters/backfill ./internal/adapters/config -run 'Test(Workspace|LedgerLocations|Discovery|.*Selection)' -count=1` — all four packages passed. The failing overlay command was `go test -overlay /tmp/al84-astra-behavior-34mo_h84/overlay.json ./internal/adapters -run '^TestAstraPendingIdentity$' -count=1 -v`.

Executed from web: `npx vitest run --config vitest.integration.config.ts src/views/TicketDeepDive.test.tsx src/api/client.test.ts` — 301 tests passed. Reused reported make at 273d90aa and final Pages/hosted evidence at 2b41187; these do not refute AB-01. No full make rerun, native macOS execution, new source capture, or end-to-end publication probe was performed. The counterexample directly exercises the production admission functions; publication consequences are source-traced. Parent owns fixes, consolidation, telemetry and publication.

PR https://github.com/NiltonDuarte/AgenticLens/pull/144, head 2b4118736b5bde24d17cc382da7445c8046ce884, checks previously passed; scoped review CHANGES_REQUESTED.

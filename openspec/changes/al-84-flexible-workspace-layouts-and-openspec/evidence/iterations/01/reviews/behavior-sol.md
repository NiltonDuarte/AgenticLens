# AL-84 behavior discovery review

Outcome: **CHANGES_REQUESTED**. Run `al84_review_behavior_sol`, iteration 01 discovery; requested profile `gpt-6.1-sol` / high. Runtime model/effort metadata is unavailable to this reviewer. Scope: CLI/API/UI/capture/maintenance agreement on ledger-backed homes, duplicates and movement; no sibling reports were read.

Reviewed baseline/HEAD `163181428ee9a5a599a2da1e91d3af5e618a9682` plus frozen tracked/untracked candidate. Independently checked all 669 manifest entries: zero drift; supplied manifest SHA-256 `13a044baddc800f298f54fba03ee41d4c09092d41c89228afde1a79520a821a6`. Reviewed the selected design/specs/tasks, technical/regression rules and TEST_RESULTS. Product and tests remained read-only.

## Finding B-SOL-1: generation can successfully publish outside the canonical ledger filename

Severity **6/10**, required-contract defect. Location: `backend/cmd/agentic-lens/main.go:157–173`, `backend/internal/adapters/workspace_locations.go:376–406`; consequential reader at `workspace_locations.go:116`.

An explicit `generate --json-output <configured-home>/alternate.json` bypasses ticket-location resolution. Existing outputs receive only directory-scope validation; initial-destination validation likewise never requires the canonical ledger filename. Generation then loads, refreshes and publishes that chosen file as its raw fact ledger, including the configured summary. Every discovery consumer instead looks exclusively for `agentic_metrics.json`. Consequently a successful raw-fact publication can remain absent from listing, detail, events, maintenance and automatic publication. This conflicts with the approved shared-reader/writer location contract and the promise that successful generation makes its directory discoverable. The root inspection command's computed JSON export is a separate, intentionally supported artifact and is not this finding.

Executed isolated native counterexample in `/tmp/al84-behavior-w_t295lr`: temporary Git project on `main`, explicit `records` root at depth 0/budget 300, private state confined to the fixture, and supported two-harness history copied to `records/alternate.json`. Used the existing exact Codex source session and the receipt binary, whose independently checked SHA-256 is `f8b01613b6e3739b9bb55453850e86631d837af01f409e2814f4d8cf2e5240ac`:

```text
generate --source codex --session-id 01a1d6b4-27d6-7112-9e68-48d8c31ed9da --codex-home <native-recovery>/codex-home --ticket-id AL-902 --ticket-type chore --json-output <fixture>/records/alternate.json
```

Generation exited 0, reporting successful ledger and `delivery.md` publication. Immediately running the same binary with `--list` exited 1: `No valid agentic_metrics.json ledgers found`. The canonical file remained absent. Full generation output is in the fixture's `generate.log`. This is an executed refresh counterexample; the equivalent first-publication gap follows from the inspected initial-destination guard, not a separately executed first-publication claim. Require generation destinations to agree with the canonical ledger home/file, preserving the separate computed-export interface; add a regression for this accepted-but-undiscoverable output.

## Other hypotheses assessed

- Planning-only folders and physical aliases: scanner requires a regular canonical ledger; readers validate supported facts, union aliases and retain per-location diagnostics. No planning placeholder or naming-derived identity path was found.
- Duplicate/stale selection: resolver collects all matching ledgers and rejects ambiguity; explicit stale selection cannot fall back to another duplicate. HTTP detail/events forward the same selection; UI keys and navigation retain directory coordinates. Ticket-only live navigation reaches the recovery selector rather than choosing a duplicate.
- Moves/private state: reviewed held-directory publication admission, scope revalidation, relative member I/O and pending checks. Reused the matching native receipt's archive/external arbitrary-folder refresh and source-free/state-free read evidence. CLI capture resolves existing ledgers before spawning generation; maintenance uses shared discovery and recorded identity, with separate explicit schema handling. No old-parent creation path was found in these traced consumers.

Limits: no broad suites duplicated; parent owns integrated `make`. Installed automatic hooks and native macOS execution were not exercised; those paths received source review only. This scoped review does not approve repository-wide correctness or publication.

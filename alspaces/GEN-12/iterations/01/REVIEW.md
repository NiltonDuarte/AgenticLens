# GEN-12 session audit and instruction review

## Scope and method

Snapshot: 2026-10-31. Parsed all 88 existing `alspaces/*/agentic_metrics.json` ledgers (627 session entries). Excluded testdata fixtures. Twenty tickets contain 114 Codex entries; all 114 session IDs resolve to local rollouts. Indexed 289 AgenticLens rollouts selected by session metadata cwd or ledger ID, including 175 without a matching checked-in Codex ledger entry. These may be child sessions, ongoing work, or unrecorded work; absence is not proof of data loss. The audit's own ongoing session is included in the inventory but excluded from findings.

`../../inventory.json` records every task's source coverage, Codex tool/error/compaction counts, rollout identifiers, file names, and SHA-256 hashes. Local source: `~/.codex/sessions/`, recursively. All selected JSONL files were parsed; user steering and selected execution commands were extracted. Qualitative reading concentrated on corrections, telemetry, build/test work, and handoff. This is not a claim that every transcript line received manual review. Raw transcripts remain local.

Deduplicate identical user messages before interpreting frequency: child rollouts can replay parent history. Reported ledger counts are diagnostic signals, not validated productivity measures. Tool errors can include deliberate negative tests; compactions can reflect scope and duration. Zero cost is not proof of free execution. OpenCode-only work has ledger coverage here, not transcript review. No causal ranking of agents or models is justified by these data.

## Findings and corresponding changes

### 1. Keep configuration and fixtures independent of project history

AL-56 root `01a17362-baa4-72b3-934d-3cb040c899ea`, JSONL lines 4585 and 5041: user explicitly objects that changing a ticket ID required source/test refactoring and that test/prod paths were hardcoded. These are concrete corrections, not inferred from error counts.

Change: AGENTS workspace isolation and grounding rules; TECHNICAL_GUIDELINES fixture boundaries. Ticket relocation must not require runtime or test changes; use isolated configuration and fixtures.

### 2. Validate telemetry provenance and preserve source coverage

Same root, lines 2590, 2927, 4688, 6012, and 9101: user questions whether samples are real and from the current session, requires committing metrics, flags silent unknown/missing records, and asks why OpenCode entries disappeared. FIX-8 root `01a18171-3980-7a53-87b6-f8d1c9156789`, line 1159, asks why the root session is missing. GEN-9 root `01a18887-68ef-7523-92e4-fe41fdde6380`, lines 138 and 242, clarifies local-source backfill and retaining partial migration when source data is unavailable.

Change: explicit harness selection, source/root/child validation, preservation of other harness sessions, committed ledger/summary, and explicit incomplete states. Do not fabricate missing evidence. Existing final-step ingestion mandate retained.

### 3. Avoid ceremony for clarification; preserve accepted scope

GEN-10 root `01a188b5-1a07-7163-af04-069719c119fc`, line 499, asks to merge plans/reviews and delete dangling folders. GEN-9 root, lines 190 and 223, asks to merge plans and delete empty folders. AL-56 root, lines 1238 and 1693, corrects accidental ticket changes and excessive research planning.

Change: distinguish acceptance-changing iterations from status/clarification/rework; retain approval across compaction; clarify that documentation maintenance does not require application-development delegation. The application development confirmation gate and role boundaries remain intact. AL-49 root `01a18f23-474f-7f50-83b0-a64af4954d37`, lines 92 and 111, shows explicit pre-approval followed by another approval question. Clarify that existing explicit approval satisfies the gate within its stated scope.

### 4. Keep optional enrichment optional and ship the verified assets

FIX-12 root `01a18759-f52b-7e43-a160-2ee1d4cea623`, line 19, specifies that hydration is a bonus and stored waterfall data must remain usable. Local rollout `01a18793-3350-7671-9ff2-117eda38b8dd`, line 19, records tester discovery that embedded assets were stale after build artifacts had been restored.

Change: test baseline data with hydration unavailable; preserve canonical embedded assets after frontend changes. Verified current `web/package.json` build copies `web/dist` to `backend/internal/adapters/web/dist`. Fix Go command cwd to `backend/` in technical guidance. GEN-11 root `01a18f52-031d-75e1-ac6c-85dc85d233fa`, line 17, explicitly requests ignoring generated assets; FIX-13 root `01a18f71-3099-7f92-86cc-fe6936352000`, line 17, reports the missing embed input during CI vet. Current checkout ignores generated assets, so guidance respects current tracking policy rather than requiring generated files to be committed. The first full `make` attempt additionally reproduced a clean-checkout failure: Go vet runs before frontend build and cannot resolve `all:dist`. Documented dependency installation and frontend build before the first full gate.

### 5. Keep delegation and delivery ownership explicit

FIX-12 root, lines 948 and 960, contains tester-task handoff messages about PR delivery and continuing layout work; later lines 1209 and 1238 identify that the PR merged before follow-up work finished. This shows confusing handoff evidence, not proof of unauthorized actions by every child.

Change: explicit write ownership and parent integration/PR responsibility; check current PR state and compare undelivered commits before follow-up delivery. Stage owned paths rather than indiscriminately staging other work.

### 6. Name version-specific responsibilities precisely

AL-58 rollout `01a18f13-5529-71c3-be59-36e39f8727f0`, lines 126, 134, and 168: user rejects both vague `v2.go` and unversioned `session_message.go`, requesting version and responsibility separation.

Change: schema family plus responsibility in filenames and tests; shared comments reflect shared behavior.

### 7. Preserve verification evidence and bound repeated work

AL-56 ledger reports 2,953 Codex tool calls and 21 compactions. This warrants targeted checkpoints, not a claim that those calls were wasted. FIX-12 rollout `01a18793-3350-7671-9ff2-117eda38b8dd`, line 19, carries prior focused and full-suite results into a narrowly scoped asset fix. AL-58 rollout `01a18ede-d177-7580-8a36-7c53161d241a`, line 19, records a third review round for comment corrections.

Change: preserve command/result/baseline evidence, rerun affected checks after changes, avoid duplicate whole-suite execution without a concrete concern, retain the full integration gate. Use bounded searches and checkpoint unresolved hypotheses. No arbitrary performance thresholds or model mandates added.

## Diff review

Documentation-only scope. Architecture, zero-any, zero-emoji, PO separation, and application-development gate retained. Historical artifacts and application files untouched. Instructions describe observed build paths and CLI options. No new test suite is appropriate for prose edits; repository delivery checks are recorded in SUMMARY.md. This is an orchestrator self-review, not an independent reviewer verdict.

# GEN-16 iteration 02 architecture review

- **Scope:** architecture discovery for persona resolution and presentation; requested profile `gpt-6-luna/medium` (`architecture_luna_6_medium`). The request and harness acknowledged these settings; source-observed runtime model metadata was unavailable, so it is not independently asserted.
- **Candidate:** baseline `99fbcd7be0d0ed05639e4a0cb3be8744a3d5371e`; product/test fingerprint `785b903e14bb4af8923e054f2d8d5fb69f44210150b172d431d75aa5416d6f9a` as supplied. Inspected production diff and untracked `backend/internal/domain/persona.go`, plus unchanged aggregation, event reconciliation, API-facing and live-source dependencies.
- **Outcome: APPROVED (scoped architecture focus).** No blocking architecture finding in inspected paths. This is not whole-review or merge approval.

## Findings

None.

## Architecture assessment and hypotheses

- **FCIS placement:** `domain.PersonaEvidence` is a pure value accumulator/resolver: trims explicit candidates, records disagreement, prioritizes explicit session metadata, and returns unknown when evidence is absent or conflicting. SQLite/JSON reads remain in adapters; role mapping and `formatRoleLabel`/CLI labels remain at presentation boundaries. This follows the configured domain and adapter responsibilities without leaking source dictionaries or I/O into domain.
- **Topology:** root/child calculations continue to use `SessionFact.ParentID`; live tree nodes retain parent links. No new persisted root marker was introduced. Unknown raw roles remain blank through domain grouping, while frontend and markdown renderers supply `Unknown` at display time. Legacy role constants remain available for historical values.
- **Redundant state / live rescans hypothesis — refuted:** live assistant evidence is cached in `tracePersonas` with the trace snapshot, under the same mutex, source revision invalidation, and eviction paths. `readSessionTrace` already reads the session messages; evidence is derived from that same slice. Unchanged polls return the cached trace and do not rescan history. Hydration groups the already-fetched `rawMsgs` once and resolves per verified session; it does not issue another history query per role. The extra map is ephemeral adapter cache state, not persisted topology.
- **Unnecessary complexity hypothesis — not substantiated:** the evidence helper captures a real contract (explicit persona precedence, agreement across assistant messages, sticky conflict) used across OpenCode extraction and live projection. Consolidating that rule into a small pure domain value avoids divergent adapter conditionals. The live projection consults the cached evidence after trace loading; I found no lock inversion or stale-cache path in revision invalidation/eviction.
- **Baseline preservation:** `ReconcileEvents` retains recorded event roles while enriching only content fields; when local hydration cannot produce events, the response falls back to committed ledger facts. Hydration's blank role projection therefore does not rewrite matched legacy event personas. Source refresh extraction can emit new blank roles under the approved contract.

## Verification and limits

Reused iteration 02 test evidence: all 12 Go race packages, 35 frontend files / 783 tests, typecheck, lint, and diff check reported passing. No suites were rerun. Source inspection covered Codex/OpenCode extraction, domain role grouping, live cache and hydration, event reconciliation, markdown and frontend role presentation. No isolated probe was warranted by a concrete unresolved architecture concern.

Not verified here: integrated `make`, frontend embedded asset refresh/standalone binary behavior, live browser layout, or actual source-session identity recovery. Those are outside this architecture-only conclusion and remain parent delivery/review gates.

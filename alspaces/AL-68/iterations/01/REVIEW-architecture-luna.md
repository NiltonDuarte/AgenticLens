# AL-68 architecture review — architecture-luna-01

## Scope and identity

- Requested model/effort: `gpt-6-luna / medium`.
- Runtime-reported model/effort: `gpt-6-luna / medium`; independent runtime attestation is unavailable.
- Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Candidate fingerprint: `c7dade3313a1d65a4269d595c782750655be8a044c020dc1f3ac502464cf325a`, recomputed from the 370 listed sorted paths using path/NUL/content/NUL.
- Inspected `git diff HEAD`, nonignored untracked product/tests, ticket spec, CONTRACT, PLAN, setup implementation, test report, smoke evidence, verification, and relevant unchanged Codex/OpenCode extraction and CLI paths. Did not inspect sibling focus reports, `FINDINGS.json`, or consolidated `REVIEW.md`.
- Working tree contains shared ticket/product changes; this review wrote only this report.

## Assessment: INCOMPLETE

No blocking architecture defect was established in the inspected candidate. The requested hook-to-agent interaction is demonstrated for OpenCode 1.19.5 in the raw smoke evidence, including an explicit binding tool call and a source-verified native child. Codex 0.166.0 has only configuration/capability evidence: trust review is pending, and no callback or active-agent binding round trip was run. This is an acceptance evidence gap for the Codex path, not evidence that its implementation is defective. Do not bypass trust; keep Codex runtime compatibility unverified/degraded until that smoke is completed.

## Architecture findings

None substantiated in this focus.

## Consequential hypotheses investigated

- **FCIS / dependency direction:** `backend/internal/domain/capture.go` contains value validation, binding conflict rules, and timestamp precedence; it imports no adapters or I/O. Store, filesystem, SQLite, clocks, harness parsing, and source reads are under `backend/internal/adapters/capture` and `backend/internal/adapters/opencode`; CLI wiring and publication are under `backend/cmd/agentic-lens`. No domain boundary breach found.
- **Source reuse and preservation:** Codex reconciliation reuses `codex.ParseRollout` and `Transcript.SessionFacts`; OpenCode capture dispatches by complete schema family and uses exact session membership. Existing general extraction keeps its current behavior; exact capture selection is isolated in `CaptureSessionSnapshot` and uses the established v1/v2 extractors. Raw fixtures and smoke evidence specifically exercise coexisting families and ambiguous/missing exact owners. This is a reasonable reuse boundary; no title/workspace inference enters registered capture.
- **Bounded work / state ownership:** Hook input is capped at 64 KiB; capture CLI enforces a timeout; source extraction is deferred to reconciliation. Reconciliation has session/byte budgets, continuation state, and source snapshot acceptance before replacing stored snapshots. Local registration and source-derived snapshot data are separated from canonical ledger facts; publication emits a digest-bound sidecar. Static trace plus reported fixture evidence supports these boundaries.
- **Embedded compatibility / setup:** OpenCode plugin glue is embedded as a Go asset and smoke typechecked against the published 1.19.5 plugin API. The tested interaction uses runtime-provided `sessionID` and its supported system-transform/tool channels. Codex handlers use explicit source/store paths and do not modify trust. Live Codex callback semantics remain unverified as described above.
- **Maintainability:** The v1/v2 extractor’s new variadic `exact ...bool` parameter is less explicit than a named exact-session entry point, but only the capture adapter supplies it and no behavior defect or material tracing burden was established. Not raised as a blocker or requested cleanup.

## Verification and limits

- Recomputed candidate fingerprint: exact match.
- `git diff --check HEAD`: passed.
- Reused iteration evidence: full `make` passed on the frozen candidate; raw verification records Go race suite, frontend checks/build, release and distribution checks, and Linux/Darwin cross-builds. I did not rerun the suite because no concrete architecture hypothesis required it.
- Raw OpenCode smoke: real 1.19.5 agent binding interaction and native child/source identity verified. Raw Codex smoke: trust remains human-controlled and pending; no runtime interaction evidence.
- This is an architecture/maintainability assessment of the listed candidate, not whole-repository approval. Codex callback delivery, Codex agent response, other OpenCode schema family runtime execution, and non-Linux runtime behavior remain unverified.

# docs(audit): verify cross-harness source capabilities (AL-63)

## Outcome

Codex historical hydration returned 232 events for all five recorded AL-55 owners. The v1 and v2 OpenCode records returned 27 and four events respectively from single-family copies. Current historical reads reject an intact mixed-family store even when the owner ID is known. Live observation works for v1, fails for v2 and has no Codex reader. [Full matrix, versions and source IDs](CAPABILITIES.md)

## Verification and review

The November 17 review used application revision `387624b3`, which integrates main `0e9c8b36`. Exact-owner backfill completed twice for every source on disposable ledgers. Every second run preserved the entire session array. The OpenCode checks use real rows in family-isolated copies and constructed selector ledgers. The Codex check uses a copied real ticket and real rollout files. The original ticket is byte-identical to its initial revision. AL-55 and AL-62 match main, including the upstream tag addition.

Focused checks used `go test -count=1` from `backend/`. The CLI, backfill, OpenCode and live packages passed the selected cases for child read errors, manual generation, duplicate location selection, insufficient discovery coverage, child proof, structured execution and mixed-source hydration. The three additional backfill cases for skill/task snapshots, v2 grandchildren and growing duration passed. The November 17 service probes and the CLI command are recorded in the matrix.

Eight independent reviews evaluated the original audit revision. Their findings, identities and initial incomplete outcomes remain attributed. The audit owner rechecked the affected paths and discovered the mixed-store historical qualification during this refresh. [Review record](evidence/iterations/02/REVIEW.md) · [Findings and dispositions](evidence/iterations/02/REVIEW_SUMMARY.json)

Full `make` exited 0 on the integrated revision, including backend race and integration tests, 231 frontend unit tests, 804 frontend integration tests and four native build targets. Strict OpenSpec validation passed, as did all three repository workflow tests. The report and retained evidence links resolve locally.

## Telemetry

The change carries the existing AL-63 ledger. The integrated engine publishes the refresh because the shared installed pin predates the OpenSpec workspace layout. This run uses a temporary worktree-local copy of the selected machine policy with a private publication state directory. Capture stays disabled. The November 8 native graph comprises the root and four direct reviewers. Service-level delegates remain unattributed in that graph.

## Limits

OpenCode v1 metadata-backed child links still disappear. Codex failed/error function outputs still leave tool error counts at zero. OpenCode v2 lacks a native positive skill/delegation sample in the selected records. Malformed live-part diagnostics have only source-inspection evidence. A newly active live session, browser navigation and heuristic backfill were outside this replay. Four service-delegated reviewer sessions lack proven native linkage.

## Pull request

This audit updates the eight-capability matrix for OpenCode v1, OpenCode v2 and Codex and transfers AL-63 execution to OpenSpec. A replay against integrated main corrected historical OpenCode coverage. Both single-family stores can hydrate their recorded owners, while an intact mixed store returns unavailable with a diagnostic. Codex hydrated all five AL-55 owners and exposed partial coverage when four rollouts were missing.

Exact-owner backfill passed twice per source on disposable copies. The report retains the broken v1 child-link and Codex error-count cases, the v2 live limit and the bounds of the native evidence. Original roadmap acceptance is preserved.

Full `make` passed (231 frontend unit and 804 integration tests), alongside strict OpenSpec validation and three workflow tests.

[Capability matrix](https://github.com/NiltonDuarte/AgenticLens/blob/AL-63/verify-source-capabilities-across-harnesses/openspec/changes/al-63-verify-source-capabilities-across-harnesses/CAPABILITIES.md) · [AL-63 pull request](https://github.com/NiltonDuarte/AgenticLens/pull/133)

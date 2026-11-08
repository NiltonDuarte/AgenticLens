# AL-63: Verify source capabilities across harnesses

## Goal

As an operator, I want a versioned capability matrix with evidence and limits so that I can choose sources and diagnose gaps accurately.

## Context

The user reports Codex lacks trace hydration and requests a capability comparison. Identical capabilities across harnesses are not assumed. Incorporate the merged Codex-fact fix and granular ownership delivery instead of rerunning unchanged evidence. The standalone active OpenCode walkthrough has been removed from this plan; this capability audit retains its original matrix scope and reports unexecuted live checks explicitly.

## Acceptance criteria

* Provide a three-source matrix covering ingestion, historical trace hydration, live observability, skill facts, delegation facts, child-session navigation, execution-error reporting, and source-backed recovery.
* Every cell distinguishes supported and verified, supported but failing, unsupported, unavailable source evidence, or not yet executed. Cite declared support separately from runtime evidence, including build, source version/schema, command, IDs and result.
* Check persisted-data behavior with optional hydration available and unavailable; supported stored facts remain visible when enrichment fails.
* For both OpenCode schema families, check missing and malformed source records, unavailable children, retention of previous trees and skill/task snapshots, and explicit incomplete/mixed-age diagnostics. Do not treat unavailable lookup as successful empty data.
* Check exact stored-session identity, parent/source preservation, other-harness facts, growing-source refresh, unchanged repeat runs, and explicit opt-in for heuristic recovery using disposable evidence copies.
* Use AL-62 diagnosis/implemented-fix verification; record residual gaps without asserting a successful backfill or preservation check based on the original notes. Findings propose follow-up tickets rather than silently expanding this audit.

## Implementation notes

Discovery only: no application, schema, runtime or new test changes. Use existing evidence and disposable probes; label unavailable and unexecuted cases. Preserve the original AL-55 ledger. [Prior reconciliation](../../PO-19/SPEC.md) records shipped foundations and historical limits.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [relevant regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. Historical findings require revalidation; this rewrite is not runtime verification or implementation approval.

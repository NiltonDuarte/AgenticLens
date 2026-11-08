# AL-63: Verify source capabilities across harnesses

## Goal

Operators can choose a source and diagnose missing telemetry using a versioned capability matrix that separates declared support from observed behavior and evidence gaps.

## Context

The original report that Codex lacks trace hydration needs revalidation against the shipped AL-62 fact and ownership fixes. The three comparison columns are OpenCode v1, OpenCode v2 and Codex; they need not offer identical capabilities. Historical trace hydration and live observation are separate capabilities. Prior fixture coverage does not establish native runtime compatibility.

## Acceptance criteria

- Deliver a matrix covering ingestion, historical trace hydration, live observability, skill facts, delegation facts, child-session navigation, execution-error reporting and source-backed recovery for all three source families.
- Every cell separates declared support from verification and identifies supported-and-verified, supported-but-failing, unsupported, source-evidence-unavailable or not-executed behavior. Qualify by operation/version where results differ. Link evidence with AgenticLens build/revision, harness version and schema family (or explicitly unknown), command or API operation, relevant session/parent IDs, expected and observed result, and limitations. Distinguish native observations, synthetic fixtures and reused historical evidence; record why reused evidence still applies.
- Compare persisted-data reads with optional hydration available, partial and unavailable. Supported stored facts remain visible; distinguish complete-empty traces from missing or failed lookup, and do not synthesize traces from lean facts or persist enrichment during reads.
- For both OpenCode families, check missing/malformed records, unavailable children, retention of previous trees and skill/task snapshots, and explicit incomplete/mixed-age diagnostics. Unavailable lookup must never count as successful empty data.
- On disposable copies, check exact stored-session identity, parent/source preservation, unavailable owners from the same harness, other-harness facts, growing-source refresh and unchanged-repeat idempotence. Include complete smaller/empty snapshots versus partial/failed reads, local-source authority and whole-ledger aggregates under the portable-ledger contract. Heuristic recovery requires explicit opt-in; preserve the original AL-55 ledger.
- Reuse AL-62 diagnosis and implemented-fix verification with stated revision limits. Report residual failures and unexecuted live checks honestly; original notes alone prove neither successful backfill nor preservation. Conclude with evidence gaps and bounded follow-up recommendations, referencing existing owners before proposing new tickets. Supply findings usable by the backfill-defaults investigation without changing defaults.

## Implementation notes

Discovery only: no application, schema, runtime or new test changes. No reinstatement of the removed standalone active OpenCode walkthrough. [AL-62 evidence](../../AL-62/SUMMARY.md) and [roadmap reconciliation](../../PO-19/SPEC.md) provide historical foundations. [FIX-18](FIX-18%20report%20unsupported%20live%20telemetry%20accurately.md) owns live-support corrections; [REV-7](REV-7%20review%20source%20read%20guarantees.md) owns read-contract investigation. Neither is a prerequisite to recording current limits.

Preserve the [portable-ledger contract](../../../TECHNICAL_GUIDELINES.md#portable-ledger-contract) and select relevant [regression scenarios](../../../docs/REGRESSION_PREVENTION.md). This specification is planning, not runtime verification or implementation approval.

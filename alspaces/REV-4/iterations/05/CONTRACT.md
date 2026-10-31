# REV-4 iteration 05 contract delta

Baseline `0c4d1fea90b8ec259e052c484361299435a7b84c`, initial working tree clean. Existing PR #72 is open. Workspace root remains `alspaces`; the existing REV-4 identifier, branch and prior contracts are preserved.

The user asks whether all tests scoring at most four can confidently be deleted or merged into existing tests. The session already authorizes evidence-backed reduction; this request approves auditing and applying the bounded, assertion-preserving deletions/merges in CANDIDATES.md. It does not authorize blanket erasure of a score band. No renewed question is needed within these boundaries.

Audit outcome: 47 declarations examined, 42 donor declarations selected for deletion or consolidation, five retained independent guards. The exact CANDIDATES.md ledger is authoritative. Keep their existing scores; do not inflate scores to manufacture a clear threshold. CSS source assertions must survive unchanged in the existing responsive-wrapper test, not be discarded as if jsdom proved layout. Positive narrowing assignments and the separate negative compiler directives must survive under tsc.

No production, runtime, schema, CI, Makefile, dependency or fixture-data changes. Add no top-level test declarations or new helper suites. Preserve meaningful unique assertions, with only duplicate fixture literals and explicit alias/language overconstraints dropped as documented. Merge only into coherent existing recipients sharing component, fixture or contract; no large unrelated mega-test just to meet a numeric target. If a proposed merge is awkward or cannot preserve the oracle, stop that candidate and report it; retaining additional justified cases is safer than weakening assertions.

Expected declaration change if all selected donors are removed: five Go and 37 TypeScript, from 499 Go plus 698 TypeScript (1,197) to 494 Go plus 661 TypeScript (1,155). Frontend expanded cases should fall from 749 to 712 because selected TS donors are plain declarations. Count actual AST/runtime results independently; never alter assertions to fit arithmetic.

Two coders have disjoint test-only ownership listed in PLAN.md. The tester owns TEST_RESULTS.md, reviewer owns REVIEW.md, and parent owns integration, other ticket documentation, telemetry, commits and delivery. All share the checkout and preserve one another's changes. No subagent publication or nested delegation.

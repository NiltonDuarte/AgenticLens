# GEN-16 iteration 02 independent test effectiveness review

## Assessment

- **Requested reviewer profile:** `gpt-6-luna`, medium reasoning effort. Runtime profile metadata was not available independently; this records the requested profile, not source-observed identity.
- **Scope:** test effectiveness for source persona fallback, unknown-role propagation, topology, hydration, and role filtering. Product and test files were read-only.
- **Revision:** base `99fbcd7be0d0ed05639e4a0cb3be8744a3d5371e`; candidate fingerprint `785b903e14bb4af8923e054f2d8d5fb69f44210150b172d431d75aa5416d6f9a`, independently recomputed from the specified backend/web diff and untracked-file byte stream.
- **Requirements:** `alspaces/GEN-17/CONTRACT.md` and controlling `alspaces/GEN-17/iterations/02/PLAN.md`.
- **Inspected:** changed Go extraction/live/domain tests and implementations, `backend/internal/domain/persona.go`, relevant unchanged `BuildSessionTree`, changed frontend filtering/display tests and components, and the supplied `TEST_REPORT.md`. No sibling reports, consolidated review, or findings files were read.
- **Confidence:** high for the covered extraction, aggregation, hydration and UI test boundaries; moderate for complete topology regression coverage.

## Findings

### TEST-01 — Partial-tree promotion does not have a preserving-parent regression oracle

- **Classification:** required-contract test gap for this iteration; no production defect demonstrated.
- **Location:** `web/src/components/LiveSessionTree.test.tsx` hierarchy fixture; unchanged `backend/internal/domain/live.go:BuildSessionTree`.
- **Trigger:** a child session has a nonblank `parent_id`, but its parent is missing from the supplied partial node set.
- **Consequence:** the tree builder promotes the child to the displayed root list. The contract says promotion does not establish that the source session is a root; retaining the original `parent_id` is the evidence that preserves that distinction. Existing changed tree tests cover null-parent roots and children whose parent is present, but not this partial-tree case. A future regression that clears the child’s parent link during promotion could pass the current hierarchy tests.
- **Proof:** source inspection shows `BuildSessionTree` currently appends a node with a missing parent to `roots` while retaining the shallow-copied `ParentID`. The updated live tree test covers nested parent links only. This is an oracle gap, not evidence that the candidate implementation clears the link.
- **Requested test evidence:** an isolated tree fixture with a child whose parent is absent; assert it appears at top level while its `parent_id` remains the original nonblank source ID. The assertion should exercise the production tree builder or live endpoint boundary.

## Hypotheses investigated

- **Codex root source fallback / explicit precedence / unknown preservation:** covered through `ParseRollout` using source metadata fixtures; assertions reach extracted invocation and event roles, tokens, and parent provenance.
- **OpenCode v1/v2 unambiguous assistant fallback and conflicts:** covered through isolated SQLite databases and `ExtractSessions`; fixtures cover agreement, conflict, explicit session metadata precedence, ignored user transport metadata, root/child unknown roles, event parity, and retained ancestry.
- **Live source fallback and bounded repeat polling:** live SQLite-backed tests exercise assistant persona projection into the tree and cache invalidation after an in-place message update. The combined trace/persona cache plus pointer reuse assertion protects against rebuilding unchanged history; no isolated query-count assertion was found, but code inspection confirms persona evidence is computed during source snapshot reads and reused with the trace revision.
- **Unknown totals, root/child separation, and role metrics:** domain tests assert blank and whitespace roles stay grouped without losing invocation/token/cost facts, and exercise unknown root and child sessions with structural workflow metrics. Formatter tests assert visible Unknown rows and distributions.
- **Hydration and persisted-data fallback:** adapter tests exercise explicit source precedence, assistant conflict, historical RootAgent/UnassignedSubAgent preservation when source is unavailable, event count and role projection, and non-mutation of stored facts. Frontend tests cover ledger rendering while optional hydration rejects.
- **Role filtering:** a rendered `TraceWaterfall` test selects Unknown, retains blank and whitespace rows, excludes a named role, then recovers rows through No filter.

## Verification

Reused `TEST_REPORT.md` for the reported full race run across 12 Go packages, frontend run across 35 files / 783 tests, typecheck, lint, and final focused TraceWaterfall rerun. No broad suites were duplicated. The report identifies that the full frontend suite preceded a final fixture-only adjustment and that the focused test plus typecheck/lint covered the final state. I ran no additional tests.

## Outcome

**CHANGES_REQUESTED**, scoped to test effectiveness: add the preserving-parent oracle in TEST-01 because this is an explicit topology guarantee and the current tests do not exercise the missing-parent promotion boundary. No production correctness defect was demonstrated. The remaining requested fallback, unknown-role, hydration, and filtering boundaries have meaningful production-boundary assertions. Parent owns any test correction and subsequent review/delivery decision.

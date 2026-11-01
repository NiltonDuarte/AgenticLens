# REV-5 independent integration testing

Scope: main 47af923 integrated onto premerge aa86f4b within iteration 02. Read the contract/integration plan, cost semantics, backend/domain implementation maps and current frontend code/tests. Reviewed both explicit conflict decisions and their effects on incoming cost/schema behavior. Tester changed only this report; no new tests were necessary because the relevant regressions already have coherent active-path cases. No source, staging, commit or publication by this child.

## Test-value and boundary audit

- The split persisted-fact DTOs retain implicit JSON hooks, presence tracking and v4/v5 cost validation. TestCostEvidenceBoundaryCompatibility still exercises missing/null/numeric-zero/positive historical costs, partial evidence, v5 missing/contradictory members, task-link metadata and malformed input. The removed computed-ledger reverse-conversion conflict cases described an output-only API that main removed; actual fact ingestion assertions remain. No test-only reverse codec was recreated.
- The migration conflict keeps TestCostEvidenceMigrationPreservesObservedHistoryAndMissingPresence intact, including earlier-version cost presence, repeated owner rows, task exclusion, v5 contradictions and v4-to-v5 recovery semantics. Replacing SummarizeCost with ComputeTotals(...).CostSummary retains the independent expected result while exercising the live aggregate. The serialized-ledger fixture helper performs only encoding, not migration logic.
- The removed CalculateRoleEffortRollup assertions now target active role/fleet APIs. Inspection confirms the intentional scalar distinction is recorded: fleet sums per-ledger rounded amounts, while role aggregation computes its scalar from underlying facts; precise summaries remain separately asserted. This does not claim preservation of the dead wrapper's scalar behavior.
- Incoming extraction cases use real SQLite fixtures through ExtractSessions and production fact projections rather than restoring ExtractFacts. Domain/backend race suites recorded by coders cover these paths and the retained configuration rollback/current cache tests; reused those results without another full suite.
- Frontend client validation reads each response once and composes concrete shape checks with cost checks at explicit schema locations. Current client cases cover complete observed zero, partial/unavailable/empty projections, contradictory summaries, malformed non-cost nested fields, v5 required invocation/task evidence, legacy absent/null/zero facts, opaque model/role map keys, opaque event metadata and thread model validation. Missing legacy invocation cost remains absent/unavailable, not coerced to observed zero.
- Existing diagnostic cases still preserve valid data and valid sibling warnings when header JSON or entries are malformed. Existing omitted artifact-description compatibility cases remain. Removal of obsolete fetcher tests does not erase public server-route coverage or active fetchWorkspaces projection/error checks.

## Actual-wire evidence and limits

Frontend owner reported a freshly built merged Go CLI smoke using an isolated temporary adopting repository: v4 observed-zero, v4 missing-cost and v5 partial-evidence fixtures; fleet, ledger-list and each workspace-detail endpoint accepted by the merged TS decoders. Totals were respectively numeric zero/complete, null/unavailable and null/partial. No live ticket workspace or developer database was used, and the server/temporary project were removed. This is reused coder evidence, not an independent second server execution by tester. INTEGRATION_FRONTEND.md owns the detailed smoke record.

No concrete functional integration regression found during this independent inspection. A trailing blank line in costValidation.ts was reported to its production owner and corrected; tester repeated backend/web git diff --check successfully. Parent/reviewer must still assess the full incoming-code diff and execute the integrated make gate after resolution bookkeeping; focused success alone is not a claim of full integration approval, exhaustive numerical proof, browser geometry or provider source compatibility.

## Executed verification

- From backend: go test -race ./internal/adapters -run TestCostEvidenceBoundaryCompatibility -count=1 passed (exit 0).
- From backend: go test -race ./internal/domain/migration -run TestCostEvidenceMigrationPreservesObservedHistoryAndMissingPresence -count=1 passed (exit 0).
- From web: npm test -- --run src/api/client.test.ts src/components/WorkspaceLoadWarning.test.tsx passed: 113 cases across two files, no skips.
- Earlier coder domain/adapters race, frontend typecheck/lint/build and affected view/domain checks are recorded in their integration reports. No unexecuted check is represented as passed here.

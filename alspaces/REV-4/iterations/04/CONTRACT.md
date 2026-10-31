# REV-4 iteration 04 contract delta

Baseline: `e4fb51060afc40726153954872f400e6b36bdca3`, initially clean. Existing branch and open PR #72 remain the delivery target. Configuration resolves workspaces to `alspaces`; use the explicit existing REV-4 ticket. The root contract and prior iteration contracts remain historical and unchanged.

The user explicitly requests deleting release_verification_test.go, discarding TestDateFiltersAndAggregations, and merging TestOrderRoles_And_OrderSubagentRoles_AdversarialEdgeCases into the enum suite. This authorizes this concrete scope and satisfies the existing approval gate under AGENTS.md; do not ask again.

Only three code-tree paths may change: backend/cmd/agentic-lens/release_verification_test.go, backend/internal/domain/aggregations_test.go, and backend/internal/domain/enums_test.go. Delete the release file and its four remaining tests/helper, remove the redundant date preset smoke, and eliminate the redundant role-ordering wrapper after consolidating its meaningful assertions into the existing enum table. No production, CI workflow, Makefile, frontend, dependency, fixture-data, or runtime changes are included.

No new top-level tests or additional table rows are needed. Preserve all meaningful ordering contracts and existing tests outside the requested targets. Expected top-level result: 505 Go minus six = 499 Go; TypeScript stays at 698 declarations and 749 runtime cases, making 1,197 total declarations. The user accepts removing the four release-file guards; do not claim their deletion improves CI validation. Actual release/publish coverage remains limited to existing CI/build behavior.

Coder owns only the three test paths. Tester owns iterations/04/TEST_RESULTS.md and adds no tests. Reviewer owns iterations/04/REVIEW.md and otherwise audits read-only. Parent owns documentation, telemetry, commits, integration and PR updates. Everyone shares the checkout and must preserve others' changes.

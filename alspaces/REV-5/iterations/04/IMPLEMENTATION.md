# Iteration 04 implementation

Implemented ADP-01 at the existing recovery boundary: any source transcript warning rejects that owner-tree refresh before combination/replacement. The existing retained-tree path exposes diagnostics and incomplete status; healthy independent trees still refresh. Both malformed complete records and a partial final write defer recovery until a clean retry. The tolerant parser is unchanged except for its misleading final-line comment.

Extended the existing global recovery test with successful parsing that contains warnings. Added a real-parser-to-persisted-ledger table covering malformed middle/final records, an incomplete unterminated tail and invalid token usage. It checks atomic parent/child retention, repeated invocation rows and counts, user/source identity, other-harness facts, independent healthy refresh, successful repaired retry and unchanged repeat without double-counting. A valid final record without newline is accepted. Fixtures are temporary and independent of ticket artifacts.

Implemented D3 in Release Please: forced rebuild bypasses release creation; explicit tag selection precedes tooling/dependencies/checks/tests; exact refs/tags identity is resolved to a commit and checked out detached. Empty/invalid/missing/non-tag references fail. Metadata rechecks HEAD and tag identity and has no invented version fallback. Forced rebuild additionally requires an existing GitHub release, and cannot move latest or sweep release candidates. Input uses environment variables rather than executable expression interpolation.

Added scripts/tests/test_release_revision.py, a standard-library test harness that executes the actual named workflow shell blocks in an isolated two-commit/tag Git repository. It covers forced older annotated tag precedence, normal output, metadata identity, wrong-HEAD rejection and empty/missing/non-tag/invalid input. Structural assertions guard execution order and forced side effects. It is integrated as make test-release-workflow, a prerequisite of make check. No production wrapper exists solely for these tests. The harness intentionally fails if the named literal run-block layout changes.

Validation on this implementation diff:
- go test -race ./internal/adapters/backfill ./internal/adapters/codex: passed.
- make test-release-workflow: four tests passed.
- git diff --check: passed.

Initial test development exposed an unsupported synthetic record type; replaced it with a supported turn_context record so malformed-record cases reach warning-bearing successful parses. No remote workflow/release/upload/tag mutation was performed. GitHub release-existence lookup and hosted action execution remain structural evidence only. Parent owns independent testing/review, integrated make, telemetry and delivery. Other findings remain open; no listener/authentication change.

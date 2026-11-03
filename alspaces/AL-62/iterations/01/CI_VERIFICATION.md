# AL-62 independent Backend CI verification

Status: focused independent verification PASS on the frozen CI correction; full integration gates remain pending.

Tester: inherited gpt-6.1-sol/high; serving identity not independently exposed. Baseline: 22f571ae649d5dc53071fc6410f187816905aa9e. Owns this report only; product/tests read-only.

## Inspected evidence and hypotheses

- Durable Backend CI job log at /tmp/al62-ci-backend-original.log matches all four user-reported cases, including all five supported v4 cost/evidence variants. Before-correction targeted log /tmp/al62-ci-before.log reproduces them. These differ from earlier native capability/worker timeout failures.
- Root auxiliary assertions/mutations were stale after canonical session ownership. Real OpenCode skill/task source records require populated nested facts, exact projected source identifiers/name/outcome/child checks and no duplicate canonical root writes. Empty root slices do not establish missing extraction.
- Schema5 output literals were stale after accepted schema6. Supported input remains version4 and retains every absent/null/zero/nonzero evidence variant, cost and historical DTO equality. Current output schema is the accepted contract, not a semantic change.
- Independent Codex roots source records usage only. Empty auxiliary facts are expected; explicit no-fabrication guards complement identity/usage/idempotence/unavailable-source assertions.
- Corrected nested tamper mutations expose a product defect: legacy-root-only current-owner fast path and receipt recording/evidence omit canonical facts. The original intermediate coder log was later overwritten by its successful correction run; the preserved verifier overlay proof at /tmp/al62-ci-product-before.log shows a modified nested skill accepted with source unavailable. Exact source-backed authorization must hold; correction must project owner provenance into receipts/evidence without relaxing multiplicity or committed foreign history.

## Independent assertion inspection

Manual test checks canonical nonempty owner, exact source event/message/outcome and explicit child reference; source-unavailable and Git-unavailable positives; name/task-description/multiplicity tamper rejection; and byte preservation on rejection. Tester requested source-ID/outcome tamper cases in that coherent table. Coder reports event/message/call/child IDs and both multiplicities added, outcome next. Redundant nested owner fields are normalized from containment; meaningful owner separation requires two actual registered sessions, not mutating an overwritten redundant projection. No product/test edits made by tester.

## Planned frozen checks

Run four affected cases with race detection, then focused related capture/history authorization tests. No full make, unrelated native worker/capability tests, predecessor merge or publication. Exact tested diff fingerprint, command/results and limitations follow after final freeze.

## Frozen verification results

Baseline commit: `22f571ae649d5dc53071fc6410f187816905aa9e`. Frozen backend patch SHA-256: `5e8f85373d93a0e3f70faa76b193579ee9511cdf6145a884dd3971f61c0f9d90` from `git diff -- backend`; the same hash was verified after execution. Product/test edits remained frozen throughout. This patch hash covers the tracked changed backend source/tests, including the new adapter test added to an already tracked file; report/plan edits are outside that hash.

Executed from backend/:

- `go test -race ./cmd/agentic-lens -run '^TestCapture(ManualAuxiliaryPublicationSurvivesUnavailableSource|CommittedHistory.*|CommittedV4HistoryNormalizesLikeCurrentUnavailableSource|MergedCommittedLedgerRetainsBothWorktrees|CLIGenerateIncludesBoundIndependentSessionsAndSummaryDiagnostics|GenerationRetainedRegistryCannotOverrideNewerCanonicalOrExplicitSnapshot|GenerationPreservesDifferingSnapshotsWithoutProvenSourceOrder)$' -count=1 -v` — PASS, package result 18.810s. Raw log `/tmp/al62-ci-verifier-cli-race.log`. All four reported cases ran and passed, including all five v4 cost/evidence variants. Related committed foreign history, unauthorized edits, native parent, stored evidence, raw-body rejection, repeated merged histories, newer canonical/explicit snapshots and equal/unknown checkpoint controls also ran and passed. No skipped test was reported.
- `go test -race ./internal/adapters/capture -run 'TestAuxiliaryEvidence' -count=1 -v` — PASS, package result 1.167s. Raw log `/tmp/al62-ci-verifier-evidence-race.log`. The actual SQLite-backed accepted-association test protects two distinct owners with shared call IDs/equal values, repeated accepted occurrences, unresolved legacy rows without invented ownership, explicit unavailable child references and exclusion of another ticket.
- `git diff --check` — PASS.

Assertion inspection confirms source-unavailable positive publication retains nonempty canonical skills/tasks and exact source identity. Canonical name/outcome/event/message/call/child mutations and skill/task count inflation must fail, with ledger bytes preserved after failure. Native parent remains explicit source ancestry; descriptive task children never become native ancestry. V4 fixture input remains version4; current output expectation follows the accepted schema6 contract while original cost/evidence/DTO equality and pinned-commit authority remain intact. The usage-only independent-root fixture explicitly refuses invented auxiliary activity. No accepted semantic requirement was relaxed.

## Additional hypothesis disposition

Parent raised whether aggregation of overlapping registered snapshots could multiply authorization evidence. Independent source trace found current Codex extraction parses one exact rollout, OpenCode capture extraction selects one exact session ID, and Reconcile rejects one session owner present across distinct registration identities before publication. No normal-source reachable overlap-inflation case was substantiated. Splitting legacy ownerless evidence indiscriminately could alter legitimate known multiplicity across independent registrations; no speculative semantic change was made. This is a bounded source-trace conclusion, not proof about arbitrarily corrupted trusted registry databases.

## Limits and remaining gates

The `-race` flag instruments the Go test process and the adapter test's production library calls. Existing cmd TestMain builds its CLI subprocess using plain `go build`; this run does not claim CLI subprocess race instrumentation. Actual CLI filesystem/SQLite/Git publication and rejection paths were exercised. The TestMain build succeeding is not evidence that the previously skipped CI binary/version job stage passed.

No full make, unrelated native-version/capability/worker timeout cases, frontend suite/build, predecessor integration, final standalone verification, remote CI rerun or publication performed by this tester. User's hold and manager-owned merge authority remain unchanged. Raw source logs are outside Git. No product/test edits, commits, subagents or external messages were made by tester. Parent retains all four focus/profile reviews and consolidated delivery gate.

## Preserved failing-before product proof

At the parent's request, the tester reproduced the actual authorization failure independently of stale fixture assertions. A disposable Go overlay replaces only these three PRODUCT files with their contents from baseline commit `22f571ae649d5dc53071fc6410f187816905aa9e`, keeping all corrected tests and other source files current:

- backend/cmd/agentic-lens/capture_history.go
- backend/cmd/agentic-lens/main.go
- backend/internal/adapters/capture/history.go

Overlay manifest: `/tmp/al62-ci-product-before-overlay-0rsbuk7c/overlay.json`. Baseline files were read using git show into the disposable directory; shared source was untouched.

Command from backend/: `GOFLAGS=-overlay=/tmp/al62-ci-product-before-overlay-0rsbuk7c/overlay.json go test ./cmd/agentic-lens -run '^TestCaptureManualAuxiliaryPublicationSurvivesUnavailableSource$' -count=1 -v`. Existing TestMain does not override cmd.Env, so its child `go build` inherits GOFLAGS and builds the old product paths with current corrected tests.

Result: expected FAIL, exit1, package1.315s. Preserved raw output `/tmp/al62-ci-product-before.log` reaches `capture_history_test.go:322`: `unproven auxiliary skill name accepted`, followed by successful structured ledger publication. The canonical nonempty source assertions and unavailable-source preservation setup pass first; the failure is the real nested-tamper acceptance oracle, not a stale root/schema assertion or fixture setup failure. Current corrected code already passed this same test in `/tmp/al62-ci-verifier-cli-race.log`; no redundant current suite rerun was needed. This is verification of the coder-supplied regression, not a new independent discovery.

Backend patch hash after this proof remains `5e8f85373d93a0e3f70faa76b193579ee9511cdf6145a884dd3971f61c0f9d90`; no product/test edits, commits or environment persistence were performed. GOFLAGS was scoped to this one command.

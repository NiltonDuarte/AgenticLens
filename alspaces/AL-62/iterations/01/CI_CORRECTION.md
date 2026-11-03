# AL-62 draft Backend CI correction

Baseline: 22f571ae649d5dc53071fc6410f187816905aa9e, initially clean. Draft PR103 remains open; predecessor integration and expensive full make stay on hold. This is a correction within iteration01, not a changed contract.

## Reproduction and diagnosis

The exact four user-reported tests failed locally before changes, including all five v4 cost/evidence variants. Raw output: /tmp/al62-ci-before.log. These are separate from manager-reported native-version/capability/worker-timeout failures.

Three failures were stale output-schema oracles: the accepted contract selects current schema6, while tests still expected5. The v4 input remains explicitly version4 and its absent/null/zero/nonzero/null-evidence cost mapping, exact non-cost history DTO comparison, pinned Git authority, changed-cost rejection and native parent assertions are preserved. The independent-root fixture records usage only, so its empty auxiliary collections are accurate; it now explicitly rejects fabricated auxiliary facts while preserving root identities, totals, repeat/missing-source behavior and summary diagnostics.

The manual positive test read legacy root collections although new source-owned facts are canonical under manual-root. Corrected assertions require exactly one canonical skill and task, no duplicate root writes, exact source/session/event/message identities, recorded successful skill outcome, descriptive child reference and unchanged facts after source/Git become unavailable.

Correcting that stale oracle exposed a real authorization regression: changing a canonical skill name was accepted with its source unavailable. Original receipt rejection assertions now mutate canonical nested rows, rather than obsolete root rows. Before the product fix, the actual CLI succeeded and rewrote the tampered ledger. Intermediate reproduction: /tmp/al62-ci-after.log was subsequently replaced by the final passing affected-case output; the original observed failure was capture_history_test.go:308, "unproven auxiliary skill name accepted".

## Product correction and preserved guarantees

- capture_history.go previously skipped auxiliary validation when legacy root collections and imported history were empty. Both the candidate and committed evidence now use owner-derived SkillFacts/TaskLinkFacts projections, covering canonical nested facts plus unresolved legacy rows.
- main.go records that same full projection in the existing scoped publication receipt. No second canonical fact store is introduced.
- capture/history.go projects accepted registration snapshots with their owning session and source, while retaining legacy registration slices. Existing receipt representation, size bounds, scoped association and exact row/max-multiplicity authorization remain intact.
- No metric authorization, source checkpoint, parent ownership, raw historical DTO normalization, publication lock or external repair behavior changed.
- Older root-only receipts remain compatible with unresolved root facts. Pre-fix canonical manual facts with unavailable sources and no accepted owner-bearing receipt/snapshot evidence cannot be retroactively trusted: generation rejects unknown evidence and requires explicit recovery. It does not silently authorize or overwrite those stored facts.

The same manual table still rejects altered names/descriptions and skill multiplicity, and now also rejects task multiplicity, source event/message/invocation IDs, skill outcome, child identity and owner-source changes. Rejection must leave the submitted baseline bytes unchanged. A SQLite adapter test covers two accepted owners sharing equal call IDs and equal-looking repeated skills, retains their separate session/source provenance, excludes another ticket and preserves unresolved legacy history.

## Executed verification

- Before: go test ./cmd/agentic-lens -run '^(TestCaptureManualAuxiliaryPublicationSurvivesUnavailableSource|TestCaptureCommittedHistoryPreservesNativeParent|TestCaptureCommittedV4HistoryNormalizesLikeCurrentUnavailableSource|TestCaptureCLIGenerateIncludesBoundIndependentSessionsAndSummaryDiagnostics)$' -count=1: FAIL all four reported cases, including five v4 subcases (5.127s).
- After product and fixture correction, same four-test command: PASS (6.761s), /tmp/al62-ci-after.log.
- Broader affected CLI: go test ./cmd/agentic-lens -run '^TestCapture(ManualAuxiliary|Committed|Merged|Generation|AuxiliaryOnly|CLIGenerate)' -count=1: PASS (26.377s), /tmp/al62-ci-capture-history.log. Includes final skill-outcome tamper guard.
- Affected adapter checks: go test ./internal/adapters/capture -run '^Test(Reconcile|.*History|.*Publication|.*Auxiliary)' -count=1: PASS (9.886s), /tmp/al62-ci-adapter-capture.log. This ran before the final new adapter test was added; that test passed separately below.
- New snapshot evidence regression: go test ./internal/adapters/capture -run '^TestAuxiliaryEvidencePreservesSnapshotOwnersAndRepeatedOccurrences$' -count=1: PASS (0.077s), /tmp/al62-ci-auxiliary-evidence.log.
- gofmt on changed Go paths and git diff --check: PASS.

Source fixtures remain self-contained; no live ticket workspace is a fixture. No full make, predecessor integration, publication, commit, push, telemetry finalization or standalone binary/version gate was performed by this implementation subagent. Parent owns independent verification/review, publication and delivery. Product and test candidate frozen after the checks above.

## Refuted overlap hypothesis

Parent asked whether root and child registration snapshots could overlap and inflate receipt permission. My initial confirmation incorrectly assumed general discovery/tree extraction also served capture. Independent tester tracing corrected this: capture.extractExact uses one ParseRollout owner (not CombineTranscripts), OpenCode uses exact-ID CaptureSessionSnapshot, and reconciliation rejects duplicate owners across registrations before publication. No ordinary supported overlap trigger was established. The hypothesis is refuted; original legacy registration aggregation and independent ownerless occurrence multiplicity remain unchanged. No defensive split or schema/behavior change was made.

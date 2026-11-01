# REV-5 independent test assessment

Baseline: e8d9003ea6388d1a750273478ec572863e965a72. Assessed the uncommitted iteration-02 diff after backend/domain/frontend handoff, including CONTRACT.md, PLAN.md, the original audit and all three implementation disposition reports. Tester owns only the test additions described here and this artifact. Production correction was implemented by the frontend coder. Parent owns the final integrated gate.

## Findings and corrections

Found a concrete compatibility regression in the new detail decoder: backend TicketArtifactItem.Description uses json omitempty, while validation required a description string. A valid artifact with omitted description therefore rejected the entire successful workspace response. Extended the existing optional-field client test to retain its omitted-items case and add a described artifact with the description omitted. Before correction, the client run failed exactly this new case (84 other cases passed). Frontend coder corrected decodeDetail to normalize only absent description to the frontend's empty string representation. The test now passes; a separate case rejects null description so normalization does not conceal malformed values. This was a review correction within approved F7, not a new scope request.

Checked adjacent Go omitempty fields against guards: session source/user/parent/events, task child/duration/cost, skill errors, migration notes, anomaly context/evidence and artifact content fields. Source/user omission is accepted; config handler explicitly emits empty role/prefix arrays, so those arrays need no missing-field fallback. Required null pointer metadata remains accepted. No further required-vs-omitted mismatch identified in that inspection. Actual Go HTTP smoke evidence from the coder remains valid except that its described artifact fixture did not exercise this omitted-description branch; the new test closes that gap.

Added TestWriteConfigPairCleansStagedFilesAfterTargetBecomesDirectory to config_internal_test.go. The first staged path is nested beneath the second target; staging creates that directory after initial prevalidation, so later target validation fails after staging has actually begun. The test asserts no partial first configuration and no leaked staged/backup files. It uses a real temporary filesystem without permission-dependent skips, sleeps or global injection hooks. Existing absent/existing shared-config CLI cases still assert exact original bytes, path preservation and error context; collision/force tests cover typed errors and replacement of both targets.

## Acceptance and test-reduction assessment

- F1: Original failed-init condition is covered with both absent and existing first config; exact preservation and staging cleanup exercised. A failure after backup movement or an OS restoration failure is not deterministically injected; implementation reports those errors, but this assessment does not claim every filesystem failure path was executed.
- F2: Removed buffer tests had only an unused buffer as their oracle. Active source trace, no-write, concurrency, revision-cache invalidation and bounded on-demand cache tests remain. Removing a buffer-size assertion from the service test does not remove its actual trace/no-disk assertions.
- F3: Obsolete progression client's mocked ledger response was not the server contract. Active ledger fetch/projection failure and query cases remain, as do server progression and frontend bucket/view tests. No route is removed.
- F4: Old flat merge tests were unreachable algorithm contracts; live mergeSessions ownership, multiplicity, snapshot replacement, idempotence and unavailable-owner tests remain. No old merge implementation was copied into test code.
- F5: Reviewed representative domain retargeting and extraction/migration diffs against the disposition maps. Explicit-ceiling calls preserve numerical assertions; completion filtering preserves no-fallback/boundary fixtures while discarding obsolete preset dispatch; registry tests use PlanChain and migration tests serialize fixtures into actual Engine.Migrate. SQLite extraction tests call ExtractSessions then production fact projections, retaining real schema fixtures and error handling. Codex compact persistence tests call SessionFacts/PersistAnomalySummaries rather than retaining an unused facade. Live hydration now checks message/event identity and session/turn provenance in addition to existing event content. These are named behavior recipients, not test-count reduction claims.
- F6: Dormant formatter/DTO/validator/API helper tests are removed only with recorded obsolete-contract rationale. Implicit JSON/YAML methods remain, with real decoder/loader tests. No new checks of dead code were added. No test source reference to the renamed GEN-14/REV-5 workspace or live alspaces history found in the targeted search.
- F7: Existing malformed-body table covers all active client fetch families and nested fields; event cases preserve legacy arrays and reject malformed envelopes. Optional malformed diagnostic headers preserve data and show response warnings; component tests cover warning presentation/recovery. Added the omitted-description and malformed-description guards described above. No browser-geometry or real-provider hydration claim.

The three implementation maps supply detailed per-symbol recipients and obsolete contracts. Inspection did not establish formal mutation equivalence or a complete binding-aware dead-code proof; counts and passing tests were not used as deletion quotas.

## Executed verification

- Initial npx vitest run src/api/client.test.ts: one expected regression failure, 84 passes, proving the omitted-description rejection before production correction.
- After correction: npm run typecheck passed.
- npx vitest run src/api/client.test.ts src/components/WorkspaceLoadWarning.test.tsx: 90 cases passed across two files, no skips (87 client, 3 warning).
- go test -race ./cmd/agentic-lens -run 'TestWriteConfigPair|TestInitCommand': passed, including new staging cleanup and existing CLI regression cases.
- git diff --check passed.

Reused coder evidence for unchanged domain/adapter race suites and the pre-correction full frontend 743-case run. The focused reruns cover tester changes and the single production decoder correction; root must run the integrated make gate and canonical frontend build after all changes. No performance profiling was requested or performed.

Status: verification corrections resolved; ready for independent reviewer and parent integration gate.

# AL-70 A12 behavior review: APPROVED

Run `a12_behavior_sol`, independent discovery, reviewer-behavior. Requested/dispatch-acknowledged identity: `gpt-6.1-sol`, high; source-session model metadata, usage and cost were not independently observed. Scope: A12 producer-to-consumer behavior against baseline `54d58be40c74864d645c5168ba23820a21389c01`, frozen candidate `974d1449fc440115b563067461d43b9b6381baf60e178ff45e6854f884b10e6a`. All 29 manifest files matched their SHA-256 hashes at entry and exit. Product/tests remained read-only; no sibling conclusions were consulted.

No blocking findings or candidate findings. The scoped verdict is APPROVED.

Investigated counterexamples and refutation:

- Missing metadata might remove zero-token invocation rows or reset counters. OpenCode v1/v2 still create their existing accumulator and project it through `SessionFact`; only its absent name changes. Real SQLite cases preserve rows, user turns, tokens, cost and duration. The v2 assistant `model.id` guard remains unchanged and rejects missing IDs.
- Empty Codex keys might be mistaken for absent turns, or native `default`/`unknown` turns might move when later context arrives. `assignTurnModel` now requires a present empty-valued entry; `turnNumber` uses map presence. Targeted rollouts cover known-name non-reassignment, late attribution retaining earlier usage/duration, and references to an earlier unknown-model turn after a second turn starts.
- API/UI/Markdown might collapse identities or expose blank labels. Unchanged invocation/model DTOs retain empty strings; frontend validation accepts strings. Domain aggregation groups by raw model, including empty. Ticket/thread/live/distribution labels use Unknown only at display; keys, color inputs, costs, counts and sorting retain original identities. Optional debug-prompt models still omit undefined values. Metadata JSON interpretation remains unchanged.
- Historical `default`/`unknown` might be relabeled or rewritten. Repeated production workspace-handler reads preserve both recorded groups, counters, amounts and exact ledger bytes without local sources. No schema/migration path changes occur.

Executed diagnostic probes, all PASS:

```text
cd backend
go test -count=1 ./internal/adapters/codex ./internal/adapters/opencode ./internal/adapters/live ./internal/adapters -run '^(TestParseRolloutMissingAndObservedModelIdentity|TestParseRolloutObservedNamesAreNotProvisional|TestParseRolloutLateModelPreservesEarlierUsageAndDuration|TestParseRolloutUnknownModelTurnLookupUsesPresence|TestOpenCodeMetadataOnlyModelIdentity|TestSchemaV1AssistantMissingModelAndObservedNames|TestSchemaV1MissingModelDoesNotCollideWithObservedDefault|TestSchemaV2AssistantStillRequiresNativeModelID|TestLiveModelResolutionPreservesUnknownAndObservedNames|TestRecordedModelNamesAreNotRelabeledOnRead|TestModelSummaryLabelsUnknownWithoutChangingGroups)$'
go test -count=1 ./internal/adapters/live -run '^TestRecordedModelNamesAndCountersSurviveRepeatedReads$'
```

The first command included a nonexistent historical-test selector; the second executes its actual name and passes. Reused tester evidence `/tmp/al70-a12-tester-evidence.json` and raw logs: affected backend/domain/HTTP checks, 112 frontend units and 149 consumer tests pass. Reused owner typecheck/lint/build logs confirm current embedded assets, also covered by the manifest. These are bounded diagnostic checks, not the public integrated gate.

Limits: no independent browser or standalone-binary launch, upstream integration, canonical `make`, telemetry or publication. Parent owns those checks and consolidation. Unchanged AL-70 configuration/capture/installation behavior is carried forward only as an unchanged dependency, not independently recertified here.

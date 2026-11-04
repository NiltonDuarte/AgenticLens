# AL-67 iteration 01 plan

## State and scope

Planning baseline: a84adbe5be46c2ef52ede0987a417562715d38e1. Initial working tree was clean. Working branch: docs/AL-67-telemetry-repository-review. Contract: ../../CONTRACT.md. Approval: explicitly granted by the user on 2026-11-04 for this documentation-only audit. Scope: architecture assessment and recommendations only. No coder, application implementation or new tests are planned. This is a broad audit despite its documentation-only delivery, so all four core review focuses remain applicable.

## Grounded investigation map

- Configuration: backend/internal/adapters/config/loader.go, Load and local overlay handling; config/source.go, ResolveSource, ResolveOpenCodePath, ResolveCodexHome and ProjectSourcePath. Compare each command's actual usage; helper comments do not establish consumer consistency.
- CLI composition: backend/cmd/agentic-lens/main.go, runGenerate, runBackfill and runServer. Follow capture selection/collection where these paths reach it. Trace native results into domain persistence/aggregation and rendered outputs.
- Ledger access: backend/internal/adapters/repository.go, ResolveWorkspacesDir and ScanWorkspaces. Distinguish its filesystem ledger responsibility from a native telemetry-source port; inventory other read/write entry points reached by consumers.
- OpenCode: adapters/opencode/client.go, OpenDB and OpenCodeDB; schema_family_detection.go, detectSchemaFamily; resolver.go, ResolveActiveRootSession; extractor.go, ExtractSessions, ExtractSessionSourceEvents and ExtractSessionsStrictWithSource; schema_v1 and schema_v2 session resolution/extraction files. Inspect exact schema dispatch and unknown/partial/mixed layout handling rather than relying on naming.
- Codex: adapters/codex/discovery.go, DiscoverRolloutPaths and DiscoveryOptions; rollout.go, ParseRollout, SessionFacts and CombineTranscripts; source_occurrences.go and occurrences.go as needed for normalized ownership and completeness. Follow children and exact selection independently of OpenCode dispatch.
- Backfill: adapters/backfill/engine.go, NewEngine and Execute; recovery.go and options.go. Trace per-owner source partitioning, recovery snapshots, unavailable owners and publication semantics.
- Live/hydration: adapters/live/service.go, NewLiveService, NewLiveServiceWithCodexHome, Start, Close, Subscribe, GetSessionTree, GetSessionTrace and HydrateWorkspaceEvents; live/sqlite.go, OpenReadOnlyHook, SourceRevision, FetchSessionTrace and ResolveChildSessionIDs. Inspect the actual live capability per family, resource ownership, cancellation, cache invalidation and subscription work bounds.
- Consumer contracts: adapters/http/live_handlers.go, LiveServiceInterface; handlers.go, WorkspaceEventHydrator and WorkspaceEventsHandler; server.go, NewServerWithConfig. Follow API DTOs into web/src/api/client.ts, liveApi.ts, types.ts and validation.ts, plus views/LiveObservability.tsx and hooks/useLiveTrace.ts where needed. No UI redesign is proposed.

## Execution sequence after explicit approval

1. Freeze the reviewed application revision and requirements. Parent records the review roster and execution settings in private working records, preserving unavailable cells as incomplete. Read relevant adjacent-ticket summaries as historical leads only; keep AL-55 intact. Capture current built CLI command help before planning execution commands against it.
2. Dispatch independent discovery runs using the roster below, within capacity waves. Every run receives the same revision, contract, plan, source pointers, canonical review skill and its own persona/shared rules. First passes must not read sibling conclusions. Reviewers own only their unique reports; application and tests remain read-only. Parent coordinates shared test execution to avoid duplicate whole suites.
3. Parent verifies and deduplicates candidate findings against reachable callers and counterevidence. Record all reporters, numeric severity, confirmed/refuted/unresolved status, pre-existing/proposal classification and verification evidence. Existing architectural debt is documented with a proposed remediation owner/slice, not silently fixed. Material unknowns remain explicit.
4. Parent writes ARCHITECTURE_ASSESSMENT.md with current and proposed dependency diagrams, operation/source matrix and concrete v1/v2/Codex request walkthroughs. Include separate generate, backfill, server/hydration and live selection tables; distinguish source harness selection from schema dispatch. Explain legitimate dispatch and capability differences as well as demonstrated gaps. For each recommendation give affected consumers, preserved behavior, required validation and bounded order.
5. Validate the assessment against evidence, ensure claims and diagrams match current wiring, and reconcile focused review coverage. A reviewer may report no findings; that is evidence only for its completed scope. Unresolved material evidence or a missing model/focus cell leaves review incomplete. This docs ticket's approval means the assessment is accurate, not that pre-existing defects have been remediated. Confirmed pre-existing debt alone does not block documentation delivery when accurately assessed with ownership and limits; inaccurate claims, unresolved material evidence and incomplete required review coverage do block it.
6. Parent performs delivery: update summary/PR draft, commit owned documentation, integrate origin/main, run required make gate and assess evidence affected by integration. Retain compact REVIEW_SUMMARY.json with every candidate/reporting tuple and dispositions, plus revision, verification actors, approval and limits in SUMMARY.md. Remove disposable review records only after approved consolidation; ARCHITECTURE_ASSESSMENT.md remains durable. Complete ticket lifecycle only when deliverables satisfy acceptance, generate/check final telemetry immediately before final delivery commit, then push and publish/link the documentation PR. Do not merge.

## Independent review roster

Every run inspects relevant unchanged dependencies, records actual acknowledged/observed model settings and owns only the named report. Parent owns consolidated REVIEW.md and all integration, telemetry, commits and publication. Report paths below are relative to this iteration; working reports may instead live in private temporary storage per the model attribution protocol, with actual paths recorded by the parent.

| Focus owner | Model / effort | Risk question and primary scope | Unique proposed report |
| --- | --- | --- | --- |
| reviewer-behavior | gpt-6.1-sol / high | Does each entry point select the intended harness/schema and deliver explicit supported outcomes through CLI/API consumers? | reports/behavior-sol.md |
| reviewer-behavior | gpt-6-luna / medium | Same independent behavior investigation and request walkthroughs. | reports/behavior-luna.md |
| reviewer-data | gpt-6.1-sol / high | Do normalized results preserve owner identity, multiplicity, tree evidence, partial/empty semantics and persisted baseline under refresh/hydration failure? | reports/data-sol.md |
| reviewer-data | gpt-6-luna / medium | Same independent data/failure investigation. | reports/data-luna.md |
| reviewer-architecture | gpt-6.1-sol / high | Where are real ports, concrete construction, schema dispatch and lifecycle responsibilities; what minimal target composition preserves capability differences? | reports/architecture-sol.md |
| reviewer-architecture | gpt-6-luna / medium | Same independent architecture investigation and target-design assessment. | reports/architecture-luna.md |
| reviewer-tests | gpt-6.1-sol / high | What existing tests actually prove selection, schema rejection, mixed ownership and failure visibility; where does evidence stop? | reports/tests-sol.md |
| reviewer-tests | gpt-6-luna / medium | Same independent test-effectiveness investigation. | reports/tests-luna.md |

Coordinator reserves one slot; run up to three cells at a time and queue remaining cells in waves against the same frozen revision. Use host-required delegation for explicit model/effort settings. Check capabilities before dispatch; unavailable configured settings are an explicit incomplete cell, never silently substituted. No security/performance specialist is added absent a demonstrated boundary needing it; operational work bounds remain in the architecture focus.

## Verification selection and evidence limits

No tests run during planning and none are added by this ticket. After approval, inspect existing coverage before choosing focused runs, including:

- config/source_test.go, TestResolveTelemetrySources, and local-overlay tests in the config package: incorrect precedence or hidden source fallback.
- opencode/schema_family_detection_test.go, TestSchemaFamilyRejectsPartialAndMixedLayouts and TestSchemaFamilyRejectsEveryRequiredTableColumnOmission; schema_v1/v2 resolution/extraction tests: wrong schema acceptance, root selection and ordering/ownership differences.
- codex/discovery_test.go and rollout_test.go: exact session/nonce selection, child evidence, parse completeness and repeated occurrences. CLI main_codex_test.go includes TestGenerateSelectedSourceDoesNotFallback and identity/snapshot preservation tests.
- backfill/engine_test.go: SourceValidation, SourceOpenCodeSkipsCodexSessions, SourceCodexSkipsOpenCodeSessions, MultiMachineLifecycle, discovery/parse failure preservation and complete-parent/missing-child preservation. Check test bodies and assertions before claiming coverage.
- live/lean_hydration_test.go, TestHydrationMixedSourcesRetainsUsableOwnersAndNativePayloads; live/sqlite_test.go includes explicit Codex home, missing root/child, SourceLessLedgerStaysUnavailable, HydrationAvailabilityPreservesLeanFactsAndLedgerBytes, child lookup failure and snapshot invalidation cases. Inspect live schema-family support independently; these names do not establish v2 parity.
- http/server_test.go, TestLeanFactsSurviveSourceTraceAvailabilityWithoutGETWrites and end-to-end hydration tests; relevant live handler tests and web API/live tests: persisted visibility and consumer representation under unavailable enrichment.

Prepare embedded assets with canonical make web-assets after ensuring dependencies exist before direct Go commands. Select bounded go test -race package/test filters for the above hypotheses and record exact commands, tested revision, results and skips. Reuse successful unchanged checks. The required integrated delivery gate remains make. If prerequisites fail, inspect the cause and report the stage rather than treating conditional skips as passing. Do not modify tracked files produced by dependency installation/build unless separately authorized; preserve and report unrelated changes.

Real-source runs are optional evidence only when available and safely read-only. Do not generate new live sessions or treat fixtures as proof of installed OpenCode/Codex compatibility. Record unsupported runtime/source/version claims, missing sources and unexecuted checks. No application path may depend on live AL-67 artifacts. Final telemetry must identify this Codex root and expected children, preserve prior harness sessions where present, and record source limitations rather than manufacture a ledger.

## Approval request

Approve the documentation-only assessment and independent review plan above. Approval authorizes read-only investigation, existing tests, assessment artifacts and documentation PR delivery. It does not authorize application refactoring, schema/runtime changes or new tests; those require separately approved development scope.

## Execution evidence

All four focuses ran on both configured profiles at the frozen application revision. Initial Luna data pass was INCOMPLETE; a new same-profile completion run closed its named source/HTTP/UI gaps. Nine runs total; all eight cells now have completed evidence. Coordinator verified consequential claims and preserved duplicate-file/child-diagnostic dissent in the durable assessment. Focused race checks and full Codex package passed; full integrated make and publication remain pending. No code or tests were changed.

# REV-4: Repository-wide test necessity review

Iteration 01/02 removal history. For the latest 0-10 ratings of the remaining suite, see [SCORE_REVIEW.md](SCORE_REVIEW.md) and [TEST_SCORES.md](TEST_SCORES.md).

## Outcome

Iteration 02 applies the user-approved removal of **50 test declarations**: 44 self-confirming or disconnected tests, one redundant repetition test, and five duplicated component cases. Each entry below preserves the original reason and retained protection. Conditional consolidation and rework candidates remain unchanged. Actual deletion and reduced-suite verification are recorded in the iteration 02 completion section.

The corrected baseline contains **1,253 test declarations**: 511 Go test functions and 742 TypeScript test declarations, across 92 tracked test/support files (56 Go, 36 TypeScript). Parameterized declarations expand into more runtime cases. Helpers and TestMain are not tests. The companion [TEST_INVENTORY.md](TEST_INVENTORY.md) records every declaration, source location, disposition, and observed assertion/contract. All 23 tests in `web/src/domain/live.test.ts` and all seven in `web/src/domain/migration.test.ts` belong to the approved deletion scope; actual live parsing and backend migration coverage remain.

Baseline: `3481919832272d3a545f8d7a1c6c98a1f678fc58`, clean checkout on `REV-4/reduce-low-value-tests`.

## Decision criteria and scope

A useful test must fail for a plausible regression in application behavior or an explicit wire/compiler contract. Small tests, zero-input cases, real I/O at adapter boundaries, and tests of the same feature at different layers are not automatically redundant. Go JSON round trips can protect real field tags and custom codecs; TypeScript `JSON.parse(...) as SomeType` does not validate an application type at runtime.

The audit inventories all declarations and screens assertion/call/table-case evidence across the suite, with detailed body inspection for the proposed removals and selected overlap clusters. KEEP in the inventory means retain conservatively because no deletion case was established; it is **not** a proof that every case is minimal. Iteration 01 performed no mutation campaign or reduced-suite run; iteration 02 reduced-suite results are recorded separately. Table rows and dynamically expanded cases are represented by their owning declaration, not individually certified for deletion. Passing the baseline establishes execution health, not test usefulness. Nothing in this audit establishes the authors' intent or whether an agent wrote a test merely to pass.

Iteration 01 delivered review recommendations. The user subsequently selected "Review and delete approved candidates", explicitly authorizing R01-R50. Iteration 02 uses `agentic-lens-dev` with planner, coder, tester, and reviewer roles under that existing approval; [the delta contract](iterations/02/CONTRACT.md) freezes this scope. No production changes or conditional consolidations are included.

## Removal ledger

R01-R50 are the approved whole-declaration deletion ledger for iteration 02. Locations refer to the original baseline, including files deleted completely; use the baseline commit to inspect historical source. Shared helpers and fixtures used by retained tests remain. All 50 declarations are removed. Iteration 02 completion evidence below records verification.

### R01: `TestReleaseStable_SemVerRegexValidation`

- Location: `backend/cmd/agentic-lens/release_verification_test.go:15`.
- Status: Removed in iteration 02.
- Reason: Compiles a regex copied into this test; never executes the workflow validator. Editing the real release validation can leave all cases green.
- Retained protection / limitation: Retain workflow integrity and Makefile missing-input checks as limited wiring checks, plus actual release CI. Those are not full end-to-end release validation. Deleting this case loses no executed production release behavior; release success-path coverage remains a known gap.
### R02: `TestReleaseCandidate_VersionFormatRegex`

- Location: `backend/cmd/agentic-lens/release_verification_test.go:70`.
- Status: Removed in iteration 02.
- Reason: Checks versions against a test-local regex. It never invokes actual candidate version generation.
- Retained protection / limitation: Retain workflow integrity and Makefile missing-input checks as limited wiring checks, plus actual release CI. Those are not full end-to-end release validation. Deleting this case loses no executed production release behavior; release success-path coverage remains a known gap.
### R03: `TestReleaseCandidate_BaseTagDerivation`

- Location: `backend/cmd/agentic-lens/release_verification_test.go:114`.
- Status: Removed in iteration 02.
- Reason: Reimplements shell sed/prefix logic in Go. Tests the copied implementation rather than the workflow.
- Retained protection / limitation: Retain workflow integrity and Makefile missing-input checks as limited wiring checks, plus actual release CI. Those are not full end-to-end release validation. Deleting this case loses no executed production release behavior; release success-path coverage remains a known gap.
### R04: `TestChecksumVerification_DetectsCorruptedArchives`

- Location: `backend/cmd/agentic-lens/release_verification_test.go:299`.
- Status: Removed in iteration 02.
- Reason: Writes a synthetic file and SHA256SUMS, then invokes sha256sum directly. It never calls make release-verify or the packaging pipeline; it tests the external utility.
- Retained protection / limitation: Retain workflow integrity and Makefile missing-input checks as limited wiring checks, plus actual release CI. Those are not full end-to-end release validation. Deleting this case loses no executed production release behavior; release success-path coverage remains a known gap.
### R05: `TestWorkspaceArtifactConfig`

- Location: `backend/internal/domain/config_test.go:244`.
- Status: Removed in iteration 02.
- Reason: Assigns fields on two plain structs and reads them back. Go struct assignment is the only behavior tested; no defaults, parsing, or artifact loading executes.
- Retained protection / limitation: Keep config writer TestGenerateConfigYAML_WorkspaceArtifacts and repository TestReadTicketArtifacts.
### R06: `TestProjectConfig_HelperFixture`

- Location: `backend/internal/domain/config_test.go:264`.
- Status: Removed in iteration 02.
- Reason: Checks values constructed by testProjectConfig, a test-only helper. Duplicates fixture maintenance and fails when the helper changes even if application behavior is correct. Retain the helper for actual normalization tests.
- Retained protection / limitation: Keep TestDefaultProjectConfig for production defaults, config loader tests for parsing, and existing consumers of testProjectConfig.
### R07: `dispatches onSelectTab when tab is clicked`

- Location: `web/src/components/Components.test.tsx:368`.
- Status: Removed in iteration 02.
- Reason: Repeats the overview navigation callback exercised across tabs in the dedicated Navbar suite.
- Retained protection / limitation: Navbar.test.tsx: invokes onSelectTab when user clicks navigation tabs.
### R08: `renders empty placeholder when data array is empty`

- Location: `web/src/components/Components.test.tsx:379`.
- Status: Removed in iteration 02.
- Reason: Duplicates the same TrendChart empty data/custom emptyMessage contract with only different labels.
- Retained protection / limitation: TrendChart.test.tsx: renders empty message when data is empty.
### R09: `renders bar series rect elements`

- Location: `web/src/components/Components.test.tsx:414`.
- Status: Removed in iteration 02.
- Reason: Only asserts that some rect exists, which could even be a hover target. The dedicated positive chart test verifies exactly two data bars and a line.
- Retained protection / limitation: TrendChart.test.tsx: renders standard positive-only chart with grid lines, bars, and line series. Keep mixed-sign geometry and hover tests.
### R10: `renders markdown artifact content and toggles between contract and summary`

- Location: `web/src/components/Components.test.tsx:967`.
- Status: Removed in iteration 02.
- Reason: Repeats rich text display and CONTRACT/SUMMARY switching in the dedicated artifact suite.
- Retained protection / limitation: ArtifactViewer.test.tsx: renders rich document headings, paragraphs, blockquotes, and lists by default; switches between CONTRACT.md and SUMMARY.md documents.
### R11: `renders empty notice when content is null`

- Location: `web/src/components/Components.test.tsx:986`.
- Status: Removed in iteration 02.
- Reason: Repeats null-artifact fallback and switching to empty SUMMARY covered by the dedicated boundary tests.
- Retained protection / limitation: ArtifactViewer.test.tsx: defensively handles null, undefined, empty, or whitespace contents; displays empty note in both rich and raw view modes when tab has no content.
### R12: `correctly discriminates event_type enums`

- Location: `web/src/domain/events.test.ts:96`.
- Status: Removed in iteration 02.
- Reason: Maps event_type from a locally constructed array and compares with its literal sequence. No production discrimination executes.
- Retained protection / limitation: Keep the same file: projects every nested field with its owner and stable chronological ordering without mutation; it invokes production sessionEvents. Keep reconcileEvents and source extraction/sorting tests.
### R13: `preserves tool call error excerpt and execution duration`

- Location: `web/src/domain/events.test.ts:100`.
- Status: Removed in iteration 02.
- Reason: Reads mockEvents[1] and checks fields assigned directly in that fixture; cannot detect payload loss in production.
- Retained protection / limitation: Keep the same file: projects every nested field with its owner and stable chronological ordering without mutation; it invokes production sessionEvents. Keep reconcileEvents and source extraction/sorting tests.
### R14: `preserves compaction events with turn and context volume`

- Location: `web/src/domain/events.test.ts:109`.
- Status: Removed in iteration 02.
- Reason: Reads mockEvents[2] and checks its own fields; no extraction or projection executes.
- Retained protection / limitation: Keep the same file: projects every nested field with its owner and stable chronological ordering without mutation; it invokes production sessionEvents. Keep reconcileEvents and source extraction/sorting tests.
### R15: `verifies chronological ascending ordering and turn tie-breaking logic`

- Location: `web/src/domain/events.test.ts:118`.
- Status: Removed in iteration 02.
- Reason: Implements a sort comparator inside the test, then tests that comparator. Production sorting can break while this remains green.
- Retained protection / limitation: Keep the same file: projects every nested field with its owner and stable chronological ordering without mutation; it invokes production sessionEvents. Keep reconcileEvents and source extraction/sorting tests.
### R16: `accepts arbitrary role strings under AgentRole`

- Location: `web/src/domain/formatters.test.ts:1011`.
- Status: Removed in iteration 02.
- Reason: Checks locally assigned role strings and array length; no role mapping/formatting code executes.
- Retained protection / limitation: Keep canonicalTicketType formatter tests, theme status mapping, role-aware metrics/UI tests, and npm run typecheck. Removing these loses a few positive compile examples, not a runtime oracle. Preserve the separate negative @ts-expect-error anomaly contract.
### R17: `accepts all 4 valid severity values`

- Location: `web/src/domain/formatters.test.ts:1026`.
- Status: Removed in iteration 02.
- Reason: Iterates a local array and verifies membership in the same literal list; cannot reject widening of the production union.
- Retained protection / limitation: Keep canonicalTicketType formatter tests, theme status mapping, role-aware metrics/UI tests, and npm run typecheck. Removing these loses a few positive compile examples, not a runtime oracle. Preserve the separate negative @ts-expect-error anomaly contract.
### R18: `accepts all 4 health status values`

- Location: `web/src/domain/formatters.test.ts:1035`.
- Status: Removed in iteration 02.
- Reason: Asserts the length of a four-element literal; no health classification executes.
- Retained protection / limitation: Keep canonicalTicketType formatter tests, theme status mapping, role-aware metrics/UI tests, and npm run typecheck. Removing these loses a few positive compile examples, not a runtime oracle. Preserve the separate negative @ts-expect-error anomaly contract.
### R19: `supports canonical conventional commit types and custom ticket types`

- Location: `web/src/domain/formatters.test.ts:1046`.
- Status: Removed in iteration 02.
- Reason: Asserts a nonempty literal array and two local string assignments. It does not test canonicalTicketType or reject invalid types.
- Retained protection / limitation: Keep canonicalTicketType formatter tests, theme status mapping, role-aware metrics/UI tests, and npm run typecheck. Removing these loses a few positive compile examples, not a runtime oracle. Preserve the separate negative @ts-expect-error anomaly contract.
### R20: `deserializes alert across all severity tiers and canonical rule names`

- Location: `web/src/domain/live.test.ts:16`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(alert.session_id).toBe(`ses_alert_${i}`);). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R21: `deserializes alert with canonical live anomaly rule`

- Location: `web/src/domain/live.test.ts:47`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(alert.rule_name).toBe('ConsecutiveToolErrors');). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R22: `handles hostile trigger_event strings including path traversal and SQL injection patterns`

- Location: `web/src/domain/live.test.ts:62`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(alert.trigger_event).toBe(hostileTrigger);). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R23: `deserializes complete tool execution part with input and output snippets`

- Location: `web/src/domain/live.test.ts:89`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(part.id).toBe('part_tool_01');). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R24: `deserializes skill and task delegation attributes while preserving legacy omission`

- Location: `web/src/domain/live.test.ts:114`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(legacy.skill_name).toBeUndefined();). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R25: `deserializes failing tool part with error message and high severity`

- Location: `web/src/domain/live.test.ts:129`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(part.is_error).toBe(true);). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R26: `handles optional fields omitted or null gracefully`

- Location: `web/src/domain/live.test.ts:149`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(part.id).toBe('part_reasoning_01');). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R27: `deserializes failing tool part where error_message is null`

- Location: `web/src/domain/live.test.ts:167`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(part.id).toBe('part_fail_no_err_msg');). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R28: `deserializes varied non-tool part types such as text, step_finish, and compaction`

- Location: `web/src/domain/live.test.ts:187`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(part.type).toBe(partType);). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R29: `deserializes message fact with full token breakdown and reasoning content`

- Location: `web/src/domain/live.test.ts:207`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(msg.id).toBe('msg_01');). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R30: `deserializes message fact when reasoning_content is omitted or null`

- Location: `web/src/domain/live.test.ts:249`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(msgWithout.reasoning_content).toBeUndefined();). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R31: `deserializes message fact with zero tokens across all token fields without failure`

- Location: `web/src/domain/live.test.ts:287`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(msg.context_tokens).toBe(0);). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R32: `deserializes message fact with saturated context window exceeding healthy ceiling`

- Location: `web/src/domain/live.test.ts:311`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(msg.context_tokens).toBe(215000);). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R33: `deserializes message fact containing multiple mixed tool and non-tool parts`

- Location: `web/src/domain/live.test.ts:333`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(msg.parts).toHaveLength(3);). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R34: `deserializes root session node with null or omitted parent_id`

- Location: `web/src/domain/live.test.ts:386`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(nodeOmitted.id).toBe('ses_root_10');). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R35: `deserializes child sub-session node with populated parent_id and attached anomalies`

- Location: `web/src/domain/live.test.ts:430`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(childNode.id).toBe('ses_child_20');). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R36: `deserializes deeply nested 3-tier session hierarchy (root -> child -> grandchild)`

- Location: `web/src/domain/live.test.ts:463`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(root.sub_sessions).toHaveLength(1);). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R37: `deserializes nested multi-level tree hierarchy and aggregated active anomalies`

- Location: `web/src/domain/live.test.ts:532`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(tree.total_active_sessions).toBe(3);). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R38: `deserializes empty session tree safely`

- Location: `web/src/domain/live.test.ts:615`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(tree.root_sessions).toHaveLength(0);). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R39: `deserializes session tree with multiple concurrent root sessions and distinct ticket formats`

- Location: `web/src/domain/live.test.ts:629`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(tree.root_sessions).toHaveLength(2);). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R40: `deserializes trace view with message history, tool statistics, and anomalies`

- Location: `web/src/domain/live.test.ts:677`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(trace.session_id).toBe('ses_trace_01');). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R41: `deserializes empty trace view safely`

- Location: `web/src/domain/live.test.ts:734`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(trace.session_id).toBe('ses_empty_trace');). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R42: `deserializes trace with 100% tool failure rate and multiple recent errors`

- Location: `web/src/domain/live.test.ts:753`.
- Status: Removed in iteration 02.
- Reason: This case constructs its own JSON or object and reads the assigned properties back (expect(trace.total_tool_calls).toBe(5);). It imports domain types only; no production parser, classifier, renderer, or security boundary executes. The stated scenario is therefore not exercised.
- Retained protection / limitation: Keep web/src/api/liveApi.test.ts production parseLiveSessionTree/parseLiveTraceView and fetch/stream tests; LiveSessionTree, LiveTraceInspector, and FailedToolAlertCard component tests. These remain behavioral coverage, not a claim of identical fixture coverage.
### R43: `deserializes raw historical v1 ledger without schema_version or events`

- Location: `web/src/domain/migration.test.ts:5`.
- Status: Removed in iteration 02.
- Reason: Only JSON.parse(JSON.stringify(test-local historical data)) runs. It cannot detect an application rejecting or misreading historical ledgers.
- Retained protection / limitation: Keep backend/internal/domain/migration/engine_test.go actual migration/invariant tests, backend/internal/domain/migration_test.go migration status tests, and production-consuming progression/events/API tests. There is no runtime coverage of migration to lose here; compile-only fixture examples disappear.
### R44: `accurately parses version 2 ledger with migration metadata and empty events array`

- Location: `web/src/domain/migration.test.ts:69`.
- Status: Removed in iteration 02.
- Reason: Only a test-local v2 object is JSON round-tripped. No production migration or decoder is invoked.
- Retained protection / limitation: Keep backend/internal/domain/migration/engine_test.go actual migration/invariant tests, backend/internal/domain/migration_test.go migration status tests, and production-consuming progression/events/API tests. There is no runtime coverage of migration to lose here; compile-only fixture examples disappear.
### R45: `accurately handles partial migration status representation`

- Location: `web/src/domain/migration.test.ts:110`.
- Status: Removed in iteration 02.
- Reason: Assigns status=partial and the two step arrays, then asserts those same literals. No migration state transition executes.
- Retained protection / limitation: Keep backend/internal/domain/migration/engine_test.go actual migration/invariant tests, backend/internal/domain/migration_test.go migration status tests, and production-consuming progression/events/API tests. There is no runtime coverage of migration to lose here; compile-only fixture examples disappear.
### R46: `accurately parses version 4 ledger with atomic user_iterations on invocations and migration metadata`

- Location: `web/src/domain/migration.test.ts:126`.
- Status: Removed in iteration 02.
- Reason: JSON round-trips a v4 fixture and uses test-local flatMap. Neither invocation attribution nor the migration engine runs.
- Retained protection / limitation: Keep backend/internal/domain/migration/engine_test.go actual migration/invariant tests, backend/internal/domain/migration_test.go migration status tests, and production-consuming progression/events/API tests. There is no runtime coverage of migration to lose here; compile-only fixture examples disappear.
### R47: `handles legacy ledgers where user_iterations is undefined without runtime error`

- Location: `web/src/domain/migration.test.ts:206`.
- Status: Removed in iteration 02.
- Reason: Checks a missing property and applies ?? 0 inside the test itself. A regression in an application fallback would leave this green.
- Retained protection / limitation: Keep backend/internal/domain/migration/engine_test.go actual migration/invariant tests, backend/internal/domain/migration_test.go migration status tests, and production-consuming progression/events/API tests. There is no runtime coverage of migration to lose here; compile-only fixture examples disappear.
### R48: `handles empty code facts defensively`

- Location: `web/src/domain/migration.test.ts:238`.
- Status: Removed in iteration 02.
- Reason: Constructs a zero-valued CodeFacts literal and asserts two assigned zeros. No defensive calculation executes.
- Retained protection / limitation: Keep backend/internal/domain/migration/engine_test.go actual migration/invariant tests, backend/internal/domain/migration_test.go migration status tests, and production-consuming progression/events/API tests. There is no runtime coverage of migration to lose here; compile-only fixture examples disappear.
### R49: `preserves populated chronological events during schema upgrades`

- Location: `web/src/domain/migration.test.ts:252`.
- Status: Removed in iteration 02.
- Reason: Builds an already ordered event array and reads it back. No schema upgrade or chronological sorting occurs.
- Retained protection / limitation: Keep backend/internal/domain/migration/engine_test.go actual migration/invariant tests, backend/internal/domain/migration_test.go migration status tests, and production-consuming progression/events/API tests. There is no runtime coverage of migration to lose here; compile-only fixture examples disappear.
### R50: `is strictly deterministic across 50 repetitions`

- Location: `web/src/utils/theme.test.ts:101`.
- Status: Removed in iteration 02.
- Reason: Repeats the same stateless role hash input 50 times. Earlier custom-role and complex-tag tests already compare repeated calls and validate color shape/palette membership.
- Retained protection / limitation: Keep returns deterministic hex colors for custom roles and handles special characters and complex model tags deterministically in this file.
## Conditional consolidation and tests to preserve

The following are not included in the 50 removals. Counting merged test wrappers as deleted coverage would exaggerate the reduction.

| Declaration | Disposition | Reason and condition | Retained owner |
|---|---|---|---|
| `backend/internal/domain/tool_execution_test.go` / `TestClassifyToolExecution` | CONSOLIDATE | Substantial overlap with external-package TestClassifyToolExecution in live_test.go. Preserve the explicit segmentation-fault case before removing the smaller table; the other suite uses panic text instead. | backend/internal/domain/live_test.go TestClassifyToolExecution |
| `backend/internal/domain/aggregations_test.go` / `TestOrderRoles` | CONSOLIDATE | Role ordering is also tested in enums_test.go. Transfer deduplication and custom-priority cases before deleting this entry. | backend/internal/domain/enums_test.go ordering/comprehensive tests |
| `backend/internal/domain/aggregations_test.go` / `TestOrderSubagentRoles` | CONSOLIDATE | Duplicates the role ordering suite; preserve both custom and default ordering for mixed-case arbitrary roles before deletion. | backend/internal/domain/enums_test.go TestOrderSubagentRoles_OrderingAndPreservation and comprehensive tests |
| `web/src/domain/anomalies.test.ts` / `narrows candidate to ToolFailureAnomalyRule via type guard across all valid members` | CONSOLIDATE | Runtime valid/non-tool membership is already tested directly. Keep compile-time narrowing assertions in a file checked by tsc before removing the runtime wrapper. | Same file membership tests plus a tsc-checked narrowing contract |
| `web/src/domain/anomalies.test.ts` / `strictly enforces LiveAnomalyAlert.rule_name type safety` | KEEP TYPE CONTRACT | Runtime literal assertions are weak, but @ts-expect-error proves invalid rule assignments are rejected. Deleting the whole test would remove meaningful negative compiler checks. | npm run typecheck includes src test files |
| `web/src/views/TicketDeepDive.test.tsx` / `ensures .table-card and .thread-accordion-item in web/src/index.css have overflow: visible to prevent tooltip clipping` | REWORK | Regex checks source spelling, not rendered tooltip clipping. Do not label this safe deletion: it is currently the explicit CSS regression guard. Replace with visual/geometry evidence before removing. | Existing DOM tooltip tests cannot verify CSS clipping in jsdom |
| `web/src/views/TicketDeepDive.test.tsx` / `ensures .table-responsive, .data-table, and .thread-models-table rules in web/src/index.css satisfy responsive constraints` | REWORK | Locks exact CSS values and text without testing layout. Keep responsive intent; replace with a browser layout check before deletion. | Existing wrapper tests cover DOM structure only |

Other overlap clusters worth reducing only after assertion transfer:

- `Components.test.tsx` and `Navbar.test.tsx` both test workspace pills/loading. The broad suite additionally asserts the absence of old Live Server/Mock Mode labels. Transfer those assertions before deleting its two Navbar render tests; the dedicated suite is not currently an identical superset.
- `Components.test.tsx` TrendChart line-render smoke includes a reference-line label, so retain that case until the dedicated chart suite owns the reference-line assertion. The bar-only smoke can be removed now because the dedicated suite checks actual data bars.
- `MetricProgression.test.tsx` repeats granularity toggles, zero-ledger states, and empty-filter recovery. Consolidate equivalent setups but retain distinct initial-empty, filtered-empty, recovery, and API-refetch transitions.
- `TraceWaterfall.test.tsx` repeats single-select and multiselect permutations. Keep one realistic conjunction flow plus independent include/exclude, selection, keyboard, and baseline denominator contracts. Removing scenarios merely because their titles sound similar would be unsafe without mapping the assertion differences.
- `config/loader_test.go` repeats discovery, precedence, partial defaults, and malformed inputs. Consolidate within one filesystem matrix, preserving root versus legacy configuration locations, explicit path priority, and wrapped error identity. Do not remove the real filesystem boundary.
- `api/client.test.ts` repeats network/HTTP/JSON failure scaffolding per endpoint. A shared parameterized contract can reduce maintenance lines, but callers still require route-specific URL/query and response/error tests. It would not justify claiming all endpoint error coverage is unnecessary.

Keep the high-value contracts:

- Session ownership, duplicate invocation multiplicity, explicit source selection, child inheritance exclusion, and refresh replacing snapshots: incorrect behavior changes attributed usage and persisted facts.
- Actual migration transformations, invariant validation, partial/unavailable source state, atomic persistence, dry-run byte preservation, and repeated-run idempotence: these protect historical data.
- Missing optional hydration with committed facts still visible, stale async responses ignored, and ambiguity/provenance preserved: unit projection tests cannot replace view/HTTP integration wiring.
- Tool classification negatives, exact threshold boundaries, streak resets, and independent session/parent groups: these prevent false diagnostics. Keep crash variants when merging overlapping tables.
- Read-only database access, concurrent subscribers, malicious session/path inputs, Markdown escaping, clipboard rejection, and SSE reconnect/unsubscribe: these cross real trust/resource boundaries.
- Real UI selection/callback/filter behavior and keyboard accessibility. A click test does not establish Space/Enter behavior or event-propagation isolation.
- Compiled frontend embedding and standalone serving. A development component test cannot establish a usable single-binary distribution.

## Performance diagnostics

Baseline timing is local and affected by concurrent profiling and the initial build; no before/after speedup is claimed. Removing fixture-only tests mostly reduces noise and maintenance, not the dominant execution time. A slow test may still be one of the most valuable tests.

- Uncached Go profile: `cd backend && go test -count=1 -json ./...`; 511 top-level test completions, 2291 test/subtest completion events, no failures or skips; observed JSON event span 14.17s (excludes any compilation before first event).
- Required `make` frontend run: 36 files, 793 cases passed; Vitest reported 27.15s. Separate JSON profile: 793 passed, 0 failed; event span approximately 35.76s under concurrent load.
- Performance target: reduce maintenance and redundant assertions first; set a runtime budget only after a controlled post-removal profile. No numeric speed target is justified by this run.

| Go duration | Test | Assessment |
|---|---|---|
| 2.32s | `agenticlens/cmd/agentic-lens/TestCLIGenerate_ThreeDeviceHandoffAccumulation` | Keep: integration/data preservation; profile fixture cost separately. |
| 1.60s | `agenticlens/cmd/agentic-lens/TestCLIGenerate_SessionRefreshMatchesFreshExtraction` | Keep: integration/data preservation; profile fixture cost separately. |
| 1.32s | `agenticlens/cmd/agentic-lens/TestCLIGenerate_InvocationsUpdateInPlaceOnSameDevice` | Keep: integration/data preservation; profile fixture cost separately. |
| 0.91s | `agenticlens/internal/adapters/live/TestHydrateWorkspaceEventsMixedSourcesPreserveSQLiteDetailsOverCommittedEcho` | Keep: integration/data preservation; profile fixture cost separately. |
| 0.72s | `agenticlens/cmd/agentic-lens/TestCLIGenerate_StrictZeroFallback` | Keep: integration/data preservation; profile fixture cost separately. |
| 0.68s | `agenticlens/cmd/agentic-lens/TestCLIGenerate_MultiDeviceInvocationsMergeAndSDLC` | Keep: integration/data preservation; profile fixture cost separately. |
| 0.64s | `agenticlens/cmd/agentic-lens/TestCLIGenerateCommand` | Keep: integration/data preservation; profile fixture cost separately. |
| 0.59s | `agenticlens/cmd/agentic-lens/TestCLIGenerate_SparseAnomalyEvidenceAndZeroErrorMessage` | Keep: integration/data preservation; profile fixture cost separately. |

Slow frontend examples from the separate JSON profile (individual durations under parallel load):

| Duration | Case | Assessment |
|---|---|---|
| 2.34s | `TicketDeepDive.test.tsx`: FIX-12 committed trace and enrichment lifecycle keeps baseline facts after rejected enrichment | Preserve behavior; optimize fixture/mount size only after measurement. |
| 2.11s | `TicketComparisons.test.tsx`: TicketComparisons View Component filters comparison rows by ticket type and workspace prefix dropdowns | Preserve behavior; optimize fixture/mount size only after measurement. |
| 1.77s | `TicketDeepDive.test.tsx`: FIX-12 committed trace and enrichment lifecycle never creates a summary list for an event-bearing trace with zero filter matches (hydrated=true) | Preserve behavior; optimize fixture/mount size only after measurement. |
| 1.61s | `TicketDeepDive.test.tsx`: FIX-12 committed trace and enrichment lifecycle keeps baseline facts after empty offline enrichment | Preserve behavior; optimize fixture/mount size only after measurement. |
| 1.59s | `ArtifactViewer.test.tsx`: ArtifactViewer Component renders large multiline documents and tables with high node counts and preserves copy fidelity | Preserve behavior; optimize fixture/mount size only after measurement. |

No sleep/skip removal is proposed without a demonstrated equivalent contract. Real SQLite, Git, filesystem, and HTTP fixtures belong at their adapter boundaries; moving them to in-memory mocks merely to improve timings could conceal real regressions. Domain timings rounded to 0.00s in Go JSON do not prove a microsecond SLA. This review does not claim a comprehensive architecture audit outside test necessity.

## Verification and implementation acceptance

- Baseline `make`: install, canonical frontend build, Go vet, frontend typecheck/lint, Go race suite, frontend tests, and binary builds. Final integration result is recorded in SUMMARY.md.
- Separate Go and frontend profiles produced the timings above. Raw execution output stays outside Git; these figures describe the original suite.
- Inventory validation: 511 Go functions and 742 TypeScript declarations. Iteration 02 restored the multiline Go TestFilterLedgersByCompletionDate declaration omitted by the initial regex (the original runtime profile already reported 511). The TypeScript inventory used the installed compiler AST, including multiline it.each. No benchmark or fuzz declarations were found. Reconcile static declarations separately from table expansion; do not add them as if both were executed-test counts.
- Confirm the five component duplicates retain every relevant assertion. Re-check unused imports after a later deletion. Keep shared test helpers until reference search proves them unused.
- Iteration 02 acceptance: affected suites and required `make` must pass after all 50 removals, retaining compiler negative tests and the critical contracts listed above. Record actual removed cases and lines.
- Regression detection loss is assessed statically. No coverage-equivalence or mutation-score claim is made. The original timing profile describes the pre-removal suite; post-removal verification is recorded separately.

## Original implementation recommendations (iteration 01)

1. Remove the 44 disconnected/literal-only cases, preserving actual parser/codec and negative compiler tests.
2. Remove the one repeated hash case and five component duplicates; retain reference-line coverage and transfer special Navbar assertions before considering further deletion.
3. Consolidate the specifically identified overlapping tables with all unique cases retained.
4. Re-profile under consistent conditions and inspect new candidates by the plausible production fault they would catch. Stop deleting when the remaining test protects an independent contract.

## Iteration 02 executed removals

The user approved the removal ledger after reviewing iteration 01. Exactly R01-R50 were removed, and no conditional candidates were applied. The eight changed test paths contain 1 line added and 1,611 lines deleted, for 1,610 net lines removed. The added line is an import-list cleanup. Production code, manifests, fixture data, and runtime scripts are unchanged.

Two files were deleted in full: `web/src/domain/live.test.ts` (23 local JSON/type-cast cases) and `web/src/domain/migration.test.ts` (seven fixture-only cases). The six other paths lose only the named tests and now-unreferenced setup, imports, and empty wrappers.

The corrected baseline is 1,253 declarations (511 Go and 742 TypeScript); the reduced source set is 1,203 declarations (505 Go and 698 TypeScript). The 50 REMOVED rows in TEST_INVENTORY.md keep their original source references and reasons. `TestFilterLedgersByCompletionDate`, missed by the original multiline-insensitive inventory, remains in the suite.

Affected Go packages passed the race detector, and frontend typecheck/lint passed. Independent exact-set and survivor-body verification, full frontend results, and the final integration gate are recorded in [iteration 02 test results](iterations/02/TEST_RESULTS.md), [iteration 02 review](iterations/02/REVIEW.md), and [SUMMARY.md](SUMMARY.md). These checks establish that the remaining suite passes and that removals match the approved list; they do not establish mutation-equivalent coverage or a controlled performance improvement.

Final integration gate: full `make` passed at `2a78bad`, including Go race checks, all 749 frontend cases, and four binary builds. The Linux standalone binary contains the exact current embedded HTML/JS/CSS bytes. The reviewed implementation fingerprint remained unchanged after integration.

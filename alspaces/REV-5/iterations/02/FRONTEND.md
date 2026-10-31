# REV-5 frontend implementation

Baseline: e8d9003ea6388d1a750273478ec572863e965a72. Owned web/** and this report; preserved concurrent backend/root changes. All frontend F3/F5/F6 candidates are removed; F7 active client.ts boundaries now validate unknown payloads before returning domain values. No server routes, liveApi SSE/polling behavior, or schema versions changed.

## Candidate and test dispositions

- fetchProgressionLedgers: removed. The deleted client tests supplied ledgers for /api/progression, which actually returns time buckets. That mocked contract was invalid. MetricProgression still fetches workspace ledgers and uses the active aggregateMetricsOverTime calculation; progression.test.ts and MetricProgression.test.ts remain.
- fetchComparisonRows: removed; its route-specific client tests guard an unused client path. Active fetchWorkspaces tests retain query encoding, sorting/projection, network/HTTP failure behavior, and views retain filtering/navigation tests. The public comparison route is untouched.
- fetchWorkspaceAnomalies: removed; its redundant fetching/defaulting wrapper tests are obsolete. fetchWorkspaceById still carries validated anomalies, with existing HTTP 404/network/500 tests and new malformed-anomaly rejection. TicketDeepDive and anomaly presentation tests remain.
- fetchProgressionMetrics: removed with its unused request-construction tests. The public progression route is untouched; active frontend progression uses ledger data and its existing time-bucket tests. No replacement copy of this client exists in tests.
- computeSkillAdoption: removed. Production skill adoption comes from the server. Backend TestComputeSkillAdoption preserves repeated calls, unique-ticket counts, explicit success/failure, ticketless ledgers, input immutability and fleet projection; TestSkillAdoptionOrderingAndEmptyProjections preserves rank ties and empty projections. The two frontend analytics tests were rewritten to directly retain the meaningful live helper oracles: skillNames deduplication/Unicode ordering/input immutability/missing facts, and skillSuccessRate invocation-weighted rates/empty denominators. No test-only aggregation implementation was introduced.
- formatIsoDateOnly: removed with its date-only rendering test because no production UI uses that output contract. The adjacent formatIsoDateTimeShort assertions remain unchanged.
- filterLedgersByCompletionDate: removed with obsolete standalone filter calls. The shared mixed test retains timezone/leap-year/completion-sort assertions, dropping only inverted-range calls into this dead function. Active backend TestFilterLedgersByCompletionDate_Ranges and completion-date tests retain completion ownership and inclusive selection; frontend progression.test.ts retains start-only/end-only/inverted/nonmatching ranges through aggregateMetricsOverTime. Existing getEffectiveCompletionDateMs and compareTicketCompletionDate guards remain.
- extractDistinctPrefixes: removed with its unused ledger-to-prefix projection test. extractTicketPrefix tests retain normalization and missing-prefix parsing; combineDistinctPrefixes retains deduplication, case/whitespace handling and sorted discovery/config merging.
- filterComparisonRows: removed with tests of the unused row filtering path. Existing filterLedgersByTicketTypes and filterLedgersByPrefixes tests retain empty/all/alias/case/null-ID and null-type cases; Adversarial Multi-Dimensional Filtering Edge Cases retain whitespace, empty input, disjoint selection and alias symmetry. TicketComparisons.test.ts retains composed filters, ticket-type/prefix selection and empty-result recovery through the actual view path.
- InfoIcon: removed unused alias. InfoTooltip and its component interaction/accessibility coverage remain.

The original syntax-aware function inventory was re-run after edits: 90 TS/TSX files parsed, no zero-production-reference named function candidates reported. This remains a conservative name-based screen, not a full binding/reachability proof; variable exports and implicit component use require separate interpretation.

## HTTP boundary behavior

validation.ts uses concrete domain-shape guards and small shared scalar/collection checks. Active fleet, ledger-list, comparison-projection input, workspace detail, version, configuration and event fetches reject malformed successful responses. Validation reaches nested session invocation tokens, workflow statistics, threads/model rows, artifacts, anomalies, skills/task links and migration metadata. Numeric data must be finite; supported enum discriminants and collection element shapes are checked. Unknown additive fields are preserved. No type assertion is used as a substitute for payload checking.

Known Go nil collection positions are normalized to empty collections before validation, including session invocation/tool/anomaly collections, optional fact collections, ledger projections, migration step arrays and metric maps. Required absent fields still fail. Legacy bare event arrays and optional event payload/ID fields remain supported. Unknown event envelopes now reject instead of becoming empty success.

Malformed optional diagnostic headers preserve valid workspace data. Each valid diagnostic element survives; malformed JSON/container/elements add an explicit response-level warning. WorkspaceDiagnostic distinguishes local response warnings from wire ticket diagnostics with a union, so no ticket ID is invented. WorkspaceLoadWarning renders/deduplicates that state and clears it on recovery while retaining the existing skipped-ticket wording for ordinary diagnostics.

## Verification

- npm run typecheck: passed.
- npm run lint: passed.
- Focused client/formatter/analytics/warning run: 160 cases passed across four files.
- Full npm test: 743 cases passed across 35 files; no skips.
- npm run build: passed. web/dist and backend/internal/adapters/web/dist matched with diff -qr; generated files remain ignored.
- git diff --check -- web: passed.
- Actual Go HTTP smoke: started the baseline built CLI with a temporary copy of testdata/workspaces and project configuration, and an unavailable temporary SQLite path. Captured /api/version, /api/config, /api/fleet, /api/workspaces, /api/workspaces/AL-1 and events responses; stopped the process and removed temporary project afterward. New guards/decoders accepted actual version/config/fleet, five ledgers including completed migration metadata, and workspace detail. Events response was the committed empty baseline (zero events), so this smoke does not claim source-hydration or nonempty event proof; existing event client tests cover nonempty hydrated/legacy payloads. No live alspaces history or developer source database was used as a test fixture.

The full suite validates view error recovery and existing interactions; browser geometry and actual provider source hydration were not exercised. Root owns final integrated make and independent tester/reviewer gates.

## Tester correction

The tester identified TicketArtifactItem.description as a Go omitempty string despite its required frontend string type. decodeDetail now normalizes only absent descriptions to the Go zero value (empty string) before validation; null and non-string descriptions still fail. Tester owns the regression cases. Adjacent artifact content remains required nullable, while contract_content/summary_content/items retain existing optional handling. Rechecked other omitempty fields in domain/models.go: parent/source/user, event IDs and payloads, skill error, task child/duration/cost, migration notes/schema/version, and optional fact collections are already optional in the relevant guards. No changes to tests by this coder during the correction.

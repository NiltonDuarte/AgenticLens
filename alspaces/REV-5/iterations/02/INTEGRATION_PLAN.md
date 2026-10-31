# REV-5 upstream integration and incoming-code review

## Authorization and baseline

User steering: "merge main into branch and review the new code". This is routine integration within iteration 02, not a new ticket or a new remediation scope. Premerge branch head is aa86f4b71abe6cbb91a43e855b91bbc11fb31515; incoming origin/main is 47af923 (AL-57, unavailable telemetry costs and schema v5). The explicit merge request authorizes necessary conflict resolution and integration corrections. Preserve the earlier fixes and verification for unchanged code; verify affected boundaries again. No additional approval is needed.

The earlier contract's prohibition on introducing a new schema version applies to REV-5 inventing a version, not to rejecting main's shipped v5 support. Integration must preserve main's documented schema-v5 contract in docs/COST_SEMANTICS.md. Unknown costs remain unknown; historical observed numeric zero remains observed zero. No new cost inference, rate-card estimation, endpoint, schema version, or compatibility policy is authorized.

## Ownership

- Backend coder: backend/cmd/** and backend/internal/adapters/**, including affected tests; own INTEGRATION_BACKEND.md only among ticket artifacts.
- Domain coder: backend/internal/domain/**, including affected tests; own INTEGRATION_DOMAIN.md.
- Frontend coder: web/**, including affected tests and canonical generated asset build; own INTEGRATION_FRONTEND.md.
- Parent: staging/resolution bookkeeping, merge commit, shared documentation and ticket artifacts, full integration gate, telemetry and PR delivery. Incoming AL-57 artifacts are part of the merge, not a new ticket assignment; preserve them without edits.
- Independent tester follows coder handoff, owning only test corrections and INTEGRATION_TESTING.md. Reviewer owns only INTEGRATION_REVIEW.md and reviews incoming code plus resolution, not just conflict markers.

All agents share the checkout, must preserve the existing staged incoming changes and each other's edits, and must not stage, commit, publish or launch children. Root marks resolved paths after their owners finish. Coordinate API changes directly across owners.

## Required resolution

### Backend

Resolve extractor.go imports in favor of the actual remaining extraction implementation: keep cost parsing/math only where needed; do not resurrect ExtractFacts or its sorting solely to satisfy incoming tests. Preserve source-backed partial/observed/unavailable cost evidence in schema-v1/v2 extraction and task links. Retarget upstream test assertions to ExtractSessions and existing session projections as REV-5 already did.

Honor main's schemas.go split into fact_ledger_dto.go, invocation_fact_dto.go, session_fact_dto.go, task_link_fact_dto.go, fact cost/decode helpers, metric/export DTOs and anomaly_dto.go. Delete the obsolete monolith only after reconciling REV-5 removals with the replacement files. Avoid duplicate declarations and do not silently reintroduce dormant validators, converters or test-only DTO surfaces that REV-5 removed. Retain implicit JSON/YAML codecs and required reverse conversions for persisted fact ingestion. Computed exports and HTTP DTOs are distinct output contracts; do not turn them into speculative second ingestion APIs.

Resolve schemas_test.go by retaining upstream cost-evidence, schema-v4/v5 compatibility and malformed-cost rejection oracles, alongside REV-5 removal rationale and surviving codec/ingestion assertions. Review automatic merges in main.go, rollout tests, HTTP tests and formatters: ensure references to removed APIs are replaced with active equivalents without weakening assertions. Preserve configuration rollback and live revision cache fixes unchanged.

### Domain

The aggregations.go conflict is an incoming modification to deleted CalculateRoleEffortRollup. Recheck actual references; retain its deletion if it still has no runtime caller, and keep incoming cost-summary accumulation in active fleet/role/model/comparison/time aggregation paths. Distinguish precise known subtotal from legacy rounded scalar compatibility; do not recover missing cost by treating numeric internal defaults as observed.

Resolve migration/engine_test.go using the active Engine.Migrate serialized boundary rather than restoring removed MigrateLedger. Preserve new v4-to-v5 upgrade, null/zero and evidence validations, owner identity and no source-recovery requirement solely for missing cost. Review auto-merged engine.go, cost_decode.go, step_v4_v5.go, validator.go, models.go and aggregation tests for old CurrentSchemaVersion assumptions and lost assertions. Existing v1-v3 recovery requirements and v4 read compatibility must remain.

### Frontend

client.ts conflicts combine REV-5's full shape decoders with main's cost-specific response readers. Integrate both validation responsibilities while reading a Response body only once. General validation must accept nullable cost fields and maps at the explicit domain/API locations now defined by main, and must retain all non-cost validation and Go null/omitted collection compatibility. Cost evidence/summary validators must reject invalid states, mismatched totals, negative/nonfinite amounts and invalid null/zero combinations. Dynamic role/model map keys are opaque identifiers, never a signal to recursively reinterpret arbitrary objects.

Keep removed fetchComparisonRows, fetchProgressionLedgers, fetchWorkspaceAnomalies and fetchProgressionMetrics absent unless a real incoming application caller is found. In particular do not restore the wrong progression-ledger contract. Transfer new cost assertions from obsolete fetcher tests to actual active fetch boundaries or existing server/domain coverage. Remove any cost response-wrapper helpers made unused by this integration, while keeping shared cost validation used by production. Preserve malformed diagnostic warnings with valid sibling data and omitted artifact descriptions.

Review automatically merged formatters, progression/scatter calculations, comparison averages, cards/charts/views and types, not only client conflicts. Preserve main's complete-only Average Spend denominator (including observed zero), partial/unavailable rows in unrelated metrics, explicit coverage, unavailable distributions and cumulative unknown costs. Ensure no numeric-only REV-5 guard rejects valid new API output. Extend active tests using actual upstream wire shapes rather than adapting production to obsolete mocks.

## Verification and review

Each coder records conflict decisions, automatically merged risks inspected, meaningful test assertion recipients and focused results. Backend/domain focused race suites must protect source extraction, persisted v4/v5 compatibility, migration, real HTTP nullable outputs, aggregate coverage and initialization rollback. Frontend typecheck/lint and affected client/view/domain tests must cover valid complete/partial/unavailable/empty costs together with malformed non-cost responses, diagnostic recovery and existing optional artifact fields.

Use a temporary adopting workspace with representative v4/v5 fixtures to check actual Go-produced JSON against merged frontend decoders; include observed zero, unknown, partial and ordinary non-cost metrics. Do not use live alspaces or developer source stores as fixtures. Canonical frontend build must refresh embedded assets; parent validates source/embed equality and standalone served assets.

After tester and reviewer gates, parent completes the merge commit with all resolved incoming paths, runs required make, and records actual passes/failures/skips. Review the incoming delta from the previous merge base through 47af923, plus the final integration diff against both parents. Re-run declaration/reference screening in changed modules for code-smell and test-only production regressions; implicit codecs are not dead code. Any unrelated upstream defect discovered is reported with severity and evidence rather than silently broadening this merge into a new feature. Final telemetry preserves original root, earlier children and newly created integration children; update the existing open PR #92 after checking its state.

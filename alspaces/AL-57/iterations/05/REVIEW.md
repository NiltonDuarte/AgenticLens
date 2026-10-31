# AL-57 iteration 05: Independent structural and assertion-transfer audit

Status: APPROVED. No open findings. Full integration and delivery verification remain parent-owned.

## Authority and reviewed snapshot

Baseline: 412bc08e3199af04c03009ed60ca1bd3b592dd85. Read the approved iteration 05 PLAN.md, frozen CONTRACT.md, AGGREGATION.md, FRONTEND.md, BACKEND.md and independent TESTING.md. The user explicitly approved the full six-part concrete plan. Cost semantics, iteration 04's filtered comparison benchmark, v4 compatibility, strict v5 facts, source ownership and wire field order remain unchanged.

Reviewed production and test bodies against that baseline, including new untracked files and baseline bodies of deleted assertions. No production/test edits by this reviewer. Final frozen backend/web source SHA256: f49d9f8a69b019dbeba07fa2e5891c1da6509a40683197bc6bfffed4c02998ad across 230 existing files. Independently computed sorted unique git ls-files -co --exclude-standard -- backend web paths, hashing each existing path, NUL, bytes, NUL; ignored dependencies/assets and deleted paths are excluded. Parent independently matches this fingerprint.

The tester's a14d1909f8693e0521b58a72b8d26ae39e600959a6488439960af929f57f2aca is an earlier audited snapshot, not an identical final-source claim. The subsequent bounded reviewer correction deleted two unreachable anomaly evidence-reference helpers; backend coder ran focused adapter/HTTP/backfill tests afterward. The parent full make gate will verify the final combined fingerprint.

## Six concerns: actual dispositions

| Concern | Disposition and concrete evidence |
| --- | --- |
| 1. Recursive frontend financial inference | FIXED. client.ts chooses five endpoint-specific readers. costValidation.ts follows only declared fleet maps, metrics, workflow, thread/model, fact-session-invocation/task, comparison and time-bucket locations. Dynamic role/model names are opaque keys with own-entry pairing checks, not schema selectors. V5 requires both own cost members for invocation and task rows; v4/no-version numeric compatibility remains. Free-form source events/artifacts are not walked. |
| 2. DTO responsibility | FIXED. Shared Grouped/Role/Model/Workflow output projections live in metric_projection_dto.go; compact computed export lives in metrics_export_dto.go. HTTP-only workspace/fact/session/thread, fleet, comparison and progression DTOs and converters are actually defined in their http package files. ResponsePayload and handlers use local HTTP types. No root-adapter forwarding facade or adapter-to-HTTP production dependency recreates the old catchall. |
| 3. Fact codec fragmentation | FIXED. Invocation tokens/fact definitions, codec and conversions colocate in invocation_fact_dto.go; task counterparts in task_link_fact_dto.go. Session/tool/event definitions and live session conversion are in session_fact_dto.go; root/code/migration DTOs and live root validation/conversion are in fact_ledger_dto.go. fact_cost.go contains only shared known-cost primitives. cost_fact_dto.go and schemas.go are removed. |
| 4. Redundant local cost work | FIXED. accumulateInvocations merges cost evidence during the token loop and projects its scalar once; ComputeTotals uses its row count directly. Thread model projection reuses that accumulator. Bucket subtotal cache and independent cumulative subtotal variable are removed. Invocation converters and HTTP fleet/cumulative converters compute their local effective evidence/summaries once. Rounding, unrounded cost ordering and non-cost projections remain. |
| 5. Unused computed reverse/dead boundary code | FIXED. Computed metrics ToDomain, its exclusive reverse subtree, scalar reverse helpers and SerializedLedgerDTO alias are removed. Dead FleetAnomaly/BrittleTool and root Tool/Event DTOs, unreachable raw root collections and validateAnomalyEvidence are removed. Meaningful assertions now reach actual emitted JSON or live fact reads; no test-only reverse graph replaces them. Live fact/anomaly conversion remains. The reviewer residue finding and correction are recorded below. |
| 6. Repeated decode mechanics | FIXED. fact_decode_helpers.go owns only configured strict Decoder construction, contextual EOF validation and the fixed named cost-member probe. Concrete invocation/task aliases remain in their owning codecs. Named v4/v5 wrappers share one canonical root body while supplying explicit optional/required evidence policy. No generic target decoder, reflection, byte-kind dispatch, arbitrary recursion or object-kind guessing appears. |

### Retained with justification and deferred scope

- RETAINED WITH JUSTIFICATION: persisted compact facts, richer admitted HTTP facts and compact computed export are distinct privacy/public-shape responsibilities. Shared metric projections are reused; merging root/session envelopes would lose admitted events or add HTTP-only facts/threads to export.
- RETAINED WITH JUSTIFICATION: AgenticFactLedgerDTO.ToDomain, SessionFactDTO conversion and live AnomalyDTO/AnomalyContextDTO/ticket-type helper chains have repository/backfill production callers. Their bidirectional fact behavior must survive removal of the unused computed reverse graph. Domain migration still supports its legacy ToolFact/EventFact inputs.
- RETAINED WITH JUSTIFICATION: CalculateFleetSummary and CalculateRoleEffortRollup still sum already-rounded per-ledger scalar projections alongside raw summaries. Two actual source ledgers with 0.00004 each produce rounded scalar ledger/fleet/role/model sums 0, but source-precision summary 0.00008. Projecting a merged raw subtotal back into those legacy public scalars would change them to 0.0001.
- RETAINED WITH JUSTIFICATION: EffectiveCost/EffectiveCostSummary remain independent pure in-memory compatibility policies at construction, rollup/sort and output boundaries. Their presence does not substitute for wire presence decoding; numeric zero and empty zero-count projections retain different meanings.
- RETAINED WITH JUSTIFICATION: explicit version and record policies, plus the separately tolerant historic migration decoder, remain. Sharing mechanics must not impose runtime strictness on tolerated historical metadata or import adapters into domain migration.
- DEFERRED: wholesale scalar-plus-summary representation replacement and rounding redesign are outside the accepted bounded cleanup. There is no delivery-blocking deferred defect or unexplained partial fix.

## Reviewer finding and correction

FIXED: the first split moved evidenceRefsToDomain and evidenceRefsFromDomain into anomaly_dto.go even though a backend-wide symbol search found only their definitions. Backend coder deleted these two unexported unreachable helpers. Rechecked there are no remaining declarations/callers. AnomalyEvidenceRefDTO and live anomaly/context/fact conversions remain intact. No new test is needed for unreachable helper deletion; focused adapter/HTTP/backfill tests passed after the correction. No open finding remains.

## Contract and architecture checks

Compared the JSON tag/order sequences of all 25 retained DTO structures with their baseline definitions; they match. Reviewed forward assignments, nil/empty collection handling and actual owners. HTTP lists/details retain facts, threads and admitted events/anomalies; compact persistence/export retain their established privacy and omission behavior. Field moves do not add a facade, import cycle or a new nullable/financial authority.

Frontend endpoint readers validate existing financial fields rather than every JSON object. Own-key checks prevent inherited or cost-named dynamic map keys from becoming schema policy. Explicit v5 invocation and task validation distinguishes omission, null and contradictory evidence. Older numeric response compatibility does not silently convert missing cost to zero or globally require new aggregate metadata from all older numeric callers. Comparison card/average behavior and view populations are unchanged.

The domain aggregation diff remains pure and deterministic. Thread ceiling/peak/cache calculations use the same accumulator and configured ceiling; row multiplicity and complete-first source-precision ordering remain. Hourly bucket filtering, chronological merges, empty code-only buckets and cumulative unknown behavior retain their previous contracts. No schema/source/migration production changes are hidden in the aggregation cleanup.

Strict decoding is consolidated at known input locations, preserving unknown-field and trailing-value rejection. Presence is checked before nullable numeric projection; required-v5 cost/evidence policy stays explicit. The pure migration layer is unchanged in this iteration, retains its prior tolerant historical-field behavior and does not gain an adapter dependency. No new any/interface{}, generic JSON engine or production source-shape guessing was introduced.

## Removed assertion audit: baseline bodies and actual recipients

Inspected the baseline bodies and current meaningful assertions, not only declaration names, counts or green suites.

| Baseline removed/changed assertion | Actual recipient or specific retirement rationale |
| --- | --- |
| Computed anomaly reverse roundtrip | TestAgenticMetricsLedgerDTO_AnomaliesSerialization now inspects the actual emitted/parsed rule, severity, tool, turn and count. Omitted presentation strings and empty-anomaly omitempty guards remain. |
| Unsupported legacy computed ingestion and malformed computed-input tests | Retired with the unused output reverse graph; an output DTO unmarshal was not a supported ingestion contract. Malformed anomaly collection rejection reaches the real fact decoder in TestSchemaMigrationDTOBoundaries. Live fact/API negative owners remain. |
| Computed custom roles/models reconstructed into domain | The same dynamic serialization owner directly asserts emitted parsed counts, custom names and order. No replacement reverse helper is created. |
| Computed reverse ticket-type assertion | TestTicketTypeBoundarySchemaAndSerialization marshals actual computed JSON and checks ticket_type. All live fact bidirectional/null/omitted assertions remain. |
| Dead BrittleToolSummaryDTO field mapping/standard-library roundtrip | Actual TestResponseHelpersAndSerialization/WriteFleetResponse checks two emitted tools field by field: tool name, anomaly count, calls, errors and failure rate. Domain arithmetic guards remain separate. |
| Dead FleetAnomalySummaryDTO mapping/empty shape | Actual WriteFleetResponse populated case checks total/warning/critical counts, full severity/rule maps, total/affected tickets, anomaly rate and nested tool fields. Its empty companion preserves zero scalars, explicit zero map entries and non-nil empty tool array. Existing domain and end-to-end /api/fleet owners remain. |
| Dead root EventFactDTO custom role | TestCanonicalV4NestedEventsReadCompatibilityAndMetadata reads security-auditor through the real session event decoder and retains the complete owner/metadata oracle. The dead root DTO's empty-role/status defaults are intentionally retired, not promised as a live runtime behavior. |
| Root Tool/Event DTO/unused anomaly helper removal | Existing root events/tools/invocations rejection loop and active anomaly validation negatives remain. Renamed TestSessionEventPersistenceAndRepositoryHandling keeps its persistence/read and unscoped-tool rejection body. Historic domain migration facts are untouched. |
| SerializedLedgerDTO alias | Three declarations use canonical AgenticMetricsLedgerDTO; their actual CLI/persistence/output assertions remain. |
| Local accumulation and thread reuse | Existing TestComputeTotals evidence table, tiny-cost model/thread case and configurable-ceiling owner now also assert scalar/count/token/cache/peak/health outputs independently. No original meaningful assertion was removed. |
| Rounded fleet/role compatibility | TestCalculateFleetSummary derives two real source ledgers before fleet/role rollup and asserts scalar 0 versus summary 0.00008, nonempty role metadata and legacy-zero versus empty fallback. |
| Bucket cache removal | Existing hourly owner supplies unordered complete/partial tiny costs, unknown, zero and code-only bucket; asserts chronological precision, counts, empty evidence and persistent cumulative incompleteness. |

Inspected actual client cases for opaque cost-like role/model names, deliberately invalid matching summary values, v5 required invocation/task members, supported legacy presence, unrelated source/artifact content, all ledger-array entry points, nested thread/model cost, comparison summary counts and independent cumulative buckets. These reach client functions, not only direct guard calls. No live workspace fixture, copied validator oracle or test quota was used.

## Verification reuse and final limits

Reused unchanged-source coder and independent tester evidence, recorded separately in AGGREGATION.md, BACKEND.md, FRONTEND.md and TESTING.md:

- Domain: focused aggregate owners, go test -race ./internal/domain/... and go vet ./internal/domain/... passed after final domain additions.
- Backend: adapter/HTTP/backfill/CLI suites under race and full Go vet passed after the DTO moves/assertion transfers; focused adapter/HTTP/backfill tests passed after deleting the two unreachable helpers.
- Frontend: all 100 API client cases, typecheck, lint and canonical build/embed-tree comparison passed with no skips.
- Reviewer: git diff --check passed; exact symbol/caller and retained-DTO tag/order checks were inspection aids, not source-text unit tests.

No suite was duplicated merely to occupy the reviewer role. No browser geometry or measured performance improvement is claimed. The parent must run the final current-main integrated make gate, standalone/embedded-asset verification and final real-source telemetry, then commit and update the existing PR. Historical full gates and the tester's earlier fingerprint remain distinct from this final source snapshot.

Final verdict: APPROVED.

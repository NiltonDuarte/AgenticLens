# AL-77 iteration 02: Require recorded ticket IDs and remove path fallback

Status: Authorized by explicit user steering; implementation may proceed within this delta without asking the same approval again.

## Steering authority and superseded decision

The user explicitly instructed: "Do not fill missing id from directory name. Ticket id is a required field on the data". This governs the correction and authorizes required-ticket validation and removal of fallback. It supersedes this iteration's earlier, still-proposed recorded_ticket_id metadata solution and the prior assumption that otherwise supported current ledgers with missing ticket IDs should remain accepted. No recorded_ticket_id field, provenance metadata, new endpoint or response shape is needed.

The source issue was independently discovered by the behavior reviewer using gpt-6-luna/medium on candidate AL77-candidate-1-4a5906d5ba2f237a: directory fallback can turn missing recorded identity into apparent cross-ticket membership. Parent source tracing confirmed the path. Preserve the finding and its discovery credit while reviews remain open. This is the same correction iteration; the clarification does not create a third iteration or rewrite the frozen [CONTRACT.md](../../CONTRACT.md). The explicit user requirement takes precedence where the baseline or earlier plan assumed no rejection for missing ticket identity.

## Grounding and required data contract

- backend/internal/adapters/fact_ledger_dto.go AgenticFactLedgerDTO.Validate is the canonical typed current-schema validator. It currently requires session identity and collections but does not require TicketID. ValidatePersistedEvidence delegates to it, and ValidateRuntimeFactLedger finishes by decoding the DTO and calling Validate. One canonical nonblank ticket-ID check therefore covers typed validation, persisted output validation and ordinary runtime ingestion without a second competing raw-JSON check.
- backend/internal/adapters/repository.go ScanWorkspaces calls ValidateRuntimeFactLedger and DTO validation before filling omitted/empty factLedger.TicketID from the directory. LoadTicketWorkspace validates the same data before filling it from the requested workspace ID. Remove both fallback blocks. Directory names can continue identifying a rejected workspace in diagnostics; they must not become ledger facts.
- ScanWorkspaces already skips a rejected individual ledger and reports WorkspaceDiagnostic while preserving valid siblings. WorkspacesHandler exposes those diagnostics, and the new warning marks inspected scope partial. LoadTicketWorkspace already returns a descriptive load error to the detail handler. Reuse these failure paths rather than changing response envelopes or inventing successful empty data.
- backend/internal/adapters/repository.go SaveFactLedger already rejects nil/blank TicketID before writing. Apply the same required-data rule at the canonical read validator. Domain value structs and unvalidated transformation fixtures need not become stricter constructors.
- docs/OWNED_OCCURRENCES.md is the current schema-6 ledger/runtime/workspace API contract. Add its concise root-identity requirement there; TECHNICAL_GUIDELINES.md already requires validation of required fields and does not need a competing rule copy.

A current persisted ledger must contain a string ticket_id with at least one non-whitespace character. Absent, null, empty and whitespace-only values are invalid. Wrong JSON types remain invalid through existing typed decoding. Do not trim, replace, guess or write identity during reads. This delta does not introduce a new rule comparing a valid recorded ID to a directory name, a new normalization rule for surrounding whitespace, or any machine/source-authenticity prerequisite.

Invalid records are excluded with explicit diagnostics; valid persisted records remain usable without their original sources. Valid sibling metrics and original bytes remain unchanged. The frontend continues matching exact session IDs across distinct available ticket_id values; it must not add new identity metadata or inference.

## Implementation ownership and minimal paths

1. Coder owns backend/internal/adapters/fact_ledger_dto.go for one canonical required nonblank TicketID check, backend/internal/adapters/repository.go for deletion of the two fallback blocks, and docs/OWNED_OCCURRENCES.md for the concise required-root-field rule. Preserve current schema version, HTTP DTO shape, domain models, extraction, persistence layout and aggregate policy. The user instruction authorizes these focused backend changes.
2. Tester owns focused updates in backend/internal/adapters/schemas_test.go, repository_test.go and http/server_test.go. Positive current-schema fixtures reaching validation need explicit synthetic ticket IDs. Replace the existing parent-directory-inference expectation with rejection/diagnostic evidence. Update additional self-contained positive fixtures in backend tests only where the stronger canonical validation exposes their omitted ticket field; preserve each test's original oracle and do not weaken unrelated error assertions. Historical legacy migration input and pure ToDomain/derivation fixtures do not need arbitrary changes unless they cross current-schema validation.
3. Existing web/src/domain/sessionOverlaps.test.ts and web/src/views/FleetOverview.test.tsx may gain focused assertions for mixed valid/rejected project scope if missing. Preserve production frontend behavior: returned diagnostics already make the loaded scope partial. Do not implement the superseded metadata proposal or bulk-edit unrelated web fixtures.
4. Parent owns contract/plan bookkeeping, review records, integration, telemetry, commits and PR delivery. Agents share the checkout, preserve others' changes and do not start unrelated work.

## Verification by reachable regression

| Regression | Focused evidence owner |
| --- | --- |
| Missing data ID is accepted through typed or ordinary runtime validation | Extend coherent schema validation cases for absent, null, empty and whitespace-only ticket_id; assert required-field error from DTO Validate and ValidateRuntimeFactLedger, with a valid explicit-ID control |
| Two unidentified ledger files sharing a session become a cross-ticket warning | Repository/HTTP fixtures place those files in different directories; neither becomes an accepted member, both get workspace diagnostics, repeated GET leaves files unchanged |
| An unidentified ledger and a valid identified ledger sharing a session create a false conflict | Existing HTTP listing fixture preserves the valid sibling and its totals, reports the invalid sibling, and produces no invented membership; actual consumer shows partial scope without conflict for the returned valid membership |
| Direct ticket load silently borrows the requested path ID | Extend existing LoadTicketWorkspace tests for missing/null/blank identity and assert a load error, no returned ledger and byte preservation |
| Invalid sibling aborts all reads or silently appears as a complete empty dataset | Existing workspace diagnostics tests verify valid siblings remain visible and inspection scope is partial; all-invalid input still exposes rejection diagnostics |
| Existing valid persisted records need private source state after validation tightens | Reuse/extend the existing source-unavailable repeated-GET fixture with explicit recorded IDs; shared sessions and mixed source labels remain returned, totals and bytes remain unchanged |
| Positive tests fail before exercising their intended schema/session error | Add explicit ticket IDs to current-schema validated fixture setup; retain existing failure substring and behavior assertions so new validation does not mask their original oracle |

No live ticket workspace or developer source database becomes a fixture. No automatic repair, migration or deletion is part of this correction. Required-field validation is authorized directly by the user; it is not a new authenticity condition inferred from tests or review opinion.

## Evidence reuse and review continuity

Candidate 1 passed focused frontend 23/23, full frontend 833/833, final strengthened domain 2/2, focused HTTP workspace race tests, frontend typecheck/lint/build and embedded-tree comparison. Those results remain evidence for unchanged UI/request behavior; they do not establish the new validation boundary. Rerun affected schema/repository/HTTP suites, then the required integrated make gate after upstream integration. Regenerate and compare frontend/embedded assets through the canonical build and verify the built standalone binary serves the current assets. The prior browser/assistive-technology limitation remains explicit: user-event/jsdom did not reproduce native summary keyboard activation.

Retain the first-candidate review findings and all discovering persona/model/effort tuples. Freeze the corrected revision before review, map changed validation and loaders to behavior, data/failure, architecture and test-effectiveness reviewers under both configured models, gpt-6.1-sol/high and gpt-6-luna/medium. The frontend async/accessibility specialist remains selected; unchanged UI evidence can carry forward only with an explicit dependency rationale. Missing discovery cells or affected revision gaps remain incomplete. Parent verifies findings, records dispositions and issues consolidated approval only after required integrated checks. Final lifecycle moves, telemetry and PR publication remain deferred to delivery; this authorization does not permit merging.

## Corrected-candidate review ownership

All runs inspect the same frozen corrected candidate and the complete AL-77 diff. First-pass reports remain independent. The three earlier completed runs are re-reviewed for the correction; missing persona/model cells receive discovery runs. Parent owns consolidated verdict and records. Each child owns only its named private report under the active private review directory, with three children at a time.

| Persona / responsibility | Risk question | Sol high report | Luna medium report |
| --- | --- | --- | --- |
| reviewer-behavior | Can invalid IDs create membership, defeat diagnostics, or regress filters/navigation/reload? | c2-reviewer-behavior-sol.md | c2-reviewer-behavior-luna.md |
| reviewer-data | Are missing IDs rejected without source dependence, writes or loss of valid siblings/totals? | c2-reviewer-data-sol.md | c2-reviewer-data-luna.md |
| reviewer-architecture | Is validation canonical and fallback fully removed without unnecessary metadata or changed aggregation policy? | c2-reviewer-architecture-sol.md | c2-reviewer-architecture-luna.md |
| reviewer-tests | Do missing-ID fixtures reach real scan/load/HTTP paths and preserve unrelated error oracles? | c2-reviewer-tests-sol.md | c2-reviewer-tests-luna.md |
| generic reviewer, frontend async/accessibility specialty | Can delayed reloads restore conflicts, or can keyboard users lose disclosure/navigation? Follow invalid-ledger diagnostics to actual UI. | c2-reviewer-frontend-sol.md | c2-reviewer-frontend-luna.md |

Models remain exact gpt-6.1-sol/high and gpt-6-luna/medium. No new policy or omitted coverage is implied by these private report names.

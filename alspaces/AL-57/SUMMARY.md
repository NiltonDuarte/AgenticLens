# AL-57: Evidence-aware telemetry cost

## Status and authorization

Implementation authorized. User selected AL-57, reviewed the proposed scope, and answered the approval question: "remove legacy_unverified type. treat legacy costs as observed. keep it simple". The first approved contract includes this amendment; iteration 02 records the acceptance delta. Valid legacy numeric values, including zero and historical Codex zero, are observed. No historical special state, count or display is implemented. AL-55 is delivered and AL2 development remains queued.

## Baseline and ownership

Initial Git status was clean at 45cb679f21475cd14b7dd4cbc25996a39d734403. The existing worktree was retained and the ticket branch is fix/AL-57-cost-semantics. Workspace root is alspaces, read from .agenticlens-config.yaml. Backend, frontend, relevant compatibility documentation, AL-57 workspace documents and its index/spec lifecycle paths are task-owned. Parent owns integration and delivery; backend and frontend coders have disjoint writable application paths.

## Work

- Read the attached PO-17 thread through its final page and the delivered local recommendations.
- Assigned the planner sole ownership of CONTRACT.md and iterations/01/PLAN.md, with zero application, schema, or test writes.
- Created the proposed CONTRACT.md and iterations/01/PLAN.md. Parent reviewed them for consistency with the planning-only authorization and PO-17 hold.
- Proposal includes partial coalesced invocation evidence, known subtotal and coverage, legacy numeric amounts treated as observed, v5 persistence/API changes, and explicit read-only canonical-v4 compatibility.
- Recorded explicit amended acceptance in CONTRACT.md and iterations/02/PLAN.md; dispatched separate backend and frontend coders with disjoint ownership.
- The original iteration 01 proposal is historical, superseded where it conflicts with the accepted legacy numeric policy.
- Initial planning whitespace/scope checks passed. Implementation is now in progress; no delivery commit or PR is claimed yet.

## Implementation

- Domain: three cost evidence states and composable invocation coverage, nullable complete-cost projections, precise complete-first cost ordering, unavailable cost-derived ratios and cumulative totals.
- Adapters: presence-aware OpenCode/Codex extraction, partial coalesced source amounts, strict v5 persistence/API validation, read-only v4 compatibility, explicit v4-v5 migration, legacy v1-v3 presence preservation, and exact-owner refresh safeguards. Valid historical numeric amounts remain observed.
- Presentation: CLI/Markdown and dashboard show known subtotals with coverage, preserve true zero, sort incomplete values last, and leave chart gaps for unavailable costs. Partial task references show known subtotals without invocation coverage and remain excluded from invocation totals.
- Documentation: cost/API compatibility guide linked from README and setup guidance; original unapproved history policy explicitly superseded.

## Verification and limits

Coder Go package tests/vet and frontend typecheck/lint/build passed. Independent tester's make test passed 12 Go packages under race and 754 frontend tests; subsequent focused domain/HTTP race tests and 64 task-panel tests passed after review corrections. See iterations/02/TESTING.md for tested source fingerprints and commands; the final integrated make gate is pending.

Review corrections: fleet null-map values require matching coverage; cost_usd HTTP query selects the actual comparator; micro-cost sorting uses complete precision with stable identity ties; partial task amounts retain their subtotal and stay excluded from invocation totals; built-in migration registration errors are handled explicitly.

Current source fingerprint (217 backend/web paths; sorted path + NUL + bytes + NUL, ignored generated files excluded): 82d6d7261258a2b33dd8350b64a5215185c2e24929a5f5301e0f25d9f6ac5c81. No browser geometry claim. Existing frontend React act warnings are recorded by tester. No live historical ledgers were bulk migrated. Old binaries reject v5; numeric-only API clients must update.

Telemetry extraction is intentionally deferred to final delivery; no ledger has been fabricated.

## Telemetry source evidence (not an extracted ledger)

The actual Codex root is 01a192c4-c9a6-7022-97da-19f433d6c5c3 in the normal ~/.codex/sessions store. session_meta source is vscode. Child session_meta records explicitly link parent_thread_id to that root:

- Planner: 01a192c6-4542-7501-99f4-9e010bd1977f.
- Backend coder: 01a192d0-c616-7413-b55b-cfc48d4f0a22.
- Frontend coder: 01a192d0-f684-7b60-b166-7429e80de682.

Tester/reviewer source IDs and final ledger verification will be recorded at delivery. Raw rollout files remain outside Git.

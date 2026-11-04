# AL-62: Granular skill and delegation facts

## Initial state and request

- Baseline: 19d70574d0e22b00f185c32d048baaf62792743e.
- Existing ticket branch: AL-62/fix-codex-skill-delegation-facts.
- Initial Git status: clean. No unrelated changes observed.
- Source: user request and AL-62 specification. Deliver granular source ownership and an evidence-backed Codex missing-fact fix together.
- Planner owns CONTRACT.md and iterations/01/PLAN.md. Parent owns integration, ticket artifacts, approval recording and delivery. Application changes remain unstarted pending approval of the concrete plan.

## Verification and limits

Planning complete; [contract](CONTRACT.md) and [iteration plan](iterations/01/PLAN.md) await explicit user approval. git diff --check passed. No implementation tests or completed review claimed. Bare agentic-lens is unavailable on PATH; use the repository build for command grounding and final extraction. Telemetry is deferred to the final delivery step.

## Planning findings

Planner (gpt-6.1-sol, high; inherited runtime configuration) reports that schema 5 already includes AL-57 changes, and proposes sequential schema 6 migration. Codex Transcript/ParseRollout currently produces neither SkillFact nor TaskLinkFact, readResponseItem discards function-call arguments, and CLI/capture auxiliary extraction is OpenCode-only. These are source inspection findings, not yet implemented or regression-tested.

Privacy-safe source inspection found the AL-55 root 01a19267-636b-7d61-85c6-db8d3d3eba77 and its four children locally available, recorded with Codex 0.166.0. Four explicit spawn_agent calls are present in the root; child rollouts replay parent history and must not own those inherited calls. Skill catalogue/path mentions do not establish activation; AL-55 skill activation remains unproven. Raw source logs stay outside Git.

Local command grounding: bare agentic-lens --help failed because the executable is absent from PATH. go run ./cmd/agentic-lens --help from backend failed because backend/internal/adapters/web/dist is absent. Implementation must prepare canonical embedded assets before invoking the local CLI.

## Approval and execution

User explicitly approved CONTRACT.md and iterations/01/PLAN.md on 2026-11-04: "Approve this scope and start implementation." Approval includes schema v6, session-owned facts/native provenance, supported Codex fixes, migration/partial refresh, CLI/API/dashboard and regression verification. Explicit skill requests retain unknown outcomes unless success is recorded. No new acceptance boundaries introduced.

Backend coder owns backend/internal/domain, backend/internal/adapters and backend/cmd/agentic-lens plus directly relevant backend schema documentation. Frontend ownership starts after the backend contract is settled. Tester receives test ownership after coders pause. Parent retains integration, ticket artifacts, telemetry and delivery.

## Environment preparation

Parent completed npm ci and make web-assets successfully, producing the canonical ignored embedded assets. No dependency manifest or lock changes. origin/main was still baseline 19d70574d0e22b00f185c32d048baaf62792743e when checked after approval; schema 5 remains available for sequential v6 extension.

Frontend coder exclusively owns web/src and iterations/01/FRONTEND.md; backend coder retains backend paths and iterations/01/BACKEND.md. Stable shared shape: session-owned collections; native provenance fields; authoritative explicit outcome; unresolved_facts for ownerless historical rows; ticket-wide API arrays derived from owned facts.

Review capability discovery confirmed Codex provider offers both configured profiles (gpt-6.1-sol/high and gpt-6-luna/medium), with reasoningEffort as the supported dispatch setting. Review is not started and no approval is claimed.

## Iteration 02: drop compatibility and ownerless migration history

Source: explicit user steering on 2026-11-04: "Delete ownerless records on migration. Drop backward compatibility." This directly authorizes the changed acceptance boundaries and supersedes iteration01 unresolved-history preservation and compatible runtime/API root projections. Implementation paused for the planner delta.

Current schema6 runtime, persisted DTO/API and frontend will use only owned session collections; no root skills/task_links compatibility copies or unresolved_facts. Older decoding remains solely for explicit migration, which drops ownerless auxiliary facts and reports discarded counts. Reliably proven owned facts may relocate. Current reads of old schemas require explicit migration. Preserve unavailable sessions' owned facts, complete-owner replacement, mixed-harness/remote history, legitimate multiplicity, provenance and unknown outcome guarantees. Tests and review now target this accepted delta; prior incompatible assumptions do not establish a regression.

Real source evidence added by backend coder: AL-62 Codex0.166.0 root 01a1a3dc-3627-79c0-8b42-8c9cea2041f3 contains direct selection (response_item line17, msg_01a1a3dc-3cc3-71f2-95e8-43f79fb7905f) and structured attachment (line19, msg_01a1a3dc-3cd7-7953-8be3-a7c4578f04f1), one activation-request envelope with unknown outcome. AL-55 spawn outputs contain task_name namespace only; an exact explicit parent_thread_id/agent_path join may resolve UUID, preserving raw child_reference if unavailable. Raw source logs and live ledgers unchanged.

## Iteration 02 implementation evidence (pre-review)

Backend production go build passed. Current persistence/runtime/API use v6 owned facts; explicit migration alone decodes old versions and reports dropped ownerless counts. Source producers retain native occurrence provenance, explicit outcomes and parent-directed child links. Session agent_path permits exact persisted metadata resolution; relationship diagnostics preserve raw source values and disable contradictory navigation. Optional source_ref is a diagnostic locator, not a native identity.

Frontend coder completed strict6 owned-only consumers. All810 frontend tests, typecheck, lint and canonical build passed on the initial iteration02 frontend diff. After the bounded child_link_diagnostic correction, all279 affected API/panel/trace tests, typecheck/lint/build passed again; unchanged scope retains earlier evidence. Embedded tree equals web/dist. Standalone binary and integrated make remain pending.

Tester (gpt-6.1-sol/high, inherited configuration) discovered a real CC-02 publication failure: valid committed other-source owned facts were rejected solely because local private registration/provenance was unavailable. Backend coder (gpt-6.1-sol/high, inherited configuration) fixed the committed-history authorization path while retaining actual repository and competing-association checks. Command-level growth/empty regression uses explicit exact session extraction; separate capture tests retain incomplete-checkpoint safeguards. Final publication regression result still pending. No independent review approval yet.

Test ownership was divided explicitly to avoid concurrent writes: frontend coder owns all web/src; backend coder owns adapter root tests except schemas_test.go/repository_test.go, HTTP tests and backfill/migration_runner_test.go; tester retains remaining backend tests. Existing obsolete root receipt and implicit legacy runtime assertions are superseded by user-approved owned-only behavior; retained assertions move to current owned snapshot/migration/security recipients documented in TESTING.md.

## Frozen review candidate

Initial iteration02 review candidate: baseline19d70574d0e22b00f185c32d048baaf62792743e with frozen439-file source/documentation/fixture manifest fingerprint03ea89d54fdb06f626e8b2ded7f9e76aefe727d70352b41f84fd2c4865d84a7b. Product/test edits paused. Behavior, data/failure and architecture first passes launched independently as gpt-6.1-sol/high, with test-effectiveness and paired gpt-6-luna/medium cells queued for capacity-limited waves. Detailed review evidence lives in private temporary storage until consolidation; no final approval claimed.

## Initial review outcome and corrections

All eight independent discovery cells completed on fingerprint03ea89d54fdb06f626e8b2ded7f9e76aefe727d70352b41f84fd2c4865d84a7b: behavior/data/architecture/test-effectiveness each on gpt-6.1-sol/high and gpt-6-luna/medium. Explicit dispatch settings accepted; source execution metadata not independently inspected. Seven requested changes, test-effectiveness gpt-6-luna/medium reported no blocking findings. Parent consolidated twelve contract defects; no publication approval.

Correction ownership: backend coder all backend production/docs; tester all backend tests/fixtures; frontend coder all web source/tests. Frozen source lifted only after every discovery pass completed. Shared attribution retained: distinct skill requests coalescing discovered by data, behavior and test-effectiveness reviewers (all gpt-6.1-sol/high); derived child UUID persistence discovered by behavior and architecture (gpt-6.1-sol/high); private history gates discovered by data and architecture (gpt-6.1-sol/high). Other concrete reporter/finding tuples and dispositions remain in private active-review records pending final compact retention.

Confirmed changes cover malformed native skill decoding/completeness; receipt/pinned-file gates; per-owner backfill despite missing children; skill envelope multiplicity; complete equal/older capture authority; derived child identity storage; child namespace contradictions; omitted OpenCode capture producer; native activation timeline producer shape; unknown opaque outcomes (data gpt-6-luna/medium); opaque agent identity vs child session (behavior gpt-6-luna/medium); and required current task outcome (architecture gpt-6-luna/medium). All are corrections to existing acceptance, not a new scope iteration.

## Corrected candidate and re-review

Corrected442-file fingerprint904aded9b1c80a32f3638b5310260a10b63d161d13a23bd1cffc68660780177f frozen for revision-aware paired re-review. Initial twelve review findings addressed. Subsequent source-integration checks also corrected native load_skill/call_id badge joins (orchestrator gpt-6.1-sol/high), sparse-event anomaly completeness (backend coder gpt-6.1-sol/high), full-ledger retained-owner anomaly aggregation and positive scoped cross-ticket association with empty refresh (tester gpt-6.1-sol/high). No unresolved test failure after affected reruns.

Backend final broad non-race run: all internal packages passed; CLI72.883s had one actual positive-association guard failure, corrected and all affected guard/portability/store/project cases passed3.297s. Frontend full815 passed before the final native-call/partial-label change, then all310 affected tests, typecheck/lint/build and embedded-tree equivalence passed. Required integrated make/race/fullCLI/standalone checks remain pending. Re-review and final approval are not yet claimed.

## Integration-ready review result

All required correction focus/profile cells APPROVED: behavior, data/failure, architecture and test-effectiveness each on gpt-6.1-sol/high and gpt-6-luna/medium. Final source fingerprint7bc5d7ef512cdc2a031f087c9e6ab1826a2ae6e7c4f25dc1eaccf1a3b7104168 (442 files). Residual DATA-01 outer-type case was CHANGES_REQUESTED at the prior checkpoint, then fixed and APPROVED by data gpt-6.1-sol/high and gpt-6-luna/medium; actual raw CLI whole-byte preservation and complete-empty controls passed. Sol behavior/architecture approvals carry from candidate2 because candidate3 changed only Codex malformed outer/required-field error warnings and tests; successful source, schema, ownership and consumer paths are unchanged. Other final cells reviewed candidate3 directly.

No unresolved finding or material review gap remains. Parent final consolidated approval is pending the required integrated make/race/full-suite and standalone asset gates. Implementation commit precedes upstream integration; telemetry remains deferred.

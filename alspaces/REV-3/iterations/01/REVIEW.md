# REV-3 Review: Lessons from REV and FIX tickets

## Scope and evidence

Baseline: 3481919. Reviewed the summaries and PR descriptions for every existing REV/FIX workspace (17 tickets: REV-1, REV-2, FIX-1 through FIX-15), the REV-1 report, iteration review findings, and relevant contracts. Compared conclusions with current AGENTS.md, CONTRIBUTING.md, TECHNICAL_GUIDELINES.md, Makefile, web/package.json, configuration, and the current REV-2 source anchors. No raw transcripts were used to infer session behavior. Historical pass claims below are ticket evidence, not verification rerun by REV-3.

Documentation-only audit: the architecture review skill was scoped to historical ticket/process evidence. Fresh application architecture scanning and test performance profiling were not part of this request. No subagents were launched under the repository's documentation maintenance exception.

## Ticket-by-ticket inventory

| Ticket and evidence | Documented failure or corrective work | Prevention lesson |
|---|---|---|
| [REV-1](../../../REV-1/REVIEW_REPORT.md), [final review](../../../REV-1/iterations/02/REVIEW.md) | Completion-date/type fallbacks invented meaning; latest-session and path guessing could select unrelated data. Final review records strict workspace resolution. | Specify authority and required/unknown behavior per field. Historical audit proposals can be superseded by later accepted iterations. |
| [REV-2](../../../REV-2/iterations/01/REVIEW.md) | Full trace reconstruction per poll, discarded child lookup errors, weak/unpopulated fields, and missing embed prerequisite. | Define runtime work bounds and partial-data semantics; use source-backed fields and clean-asset verification. |
| [FIX-1](../../../FIX-1/SUMMARY.md) | Missing ID/title could be persisted; missing type/empty invocations needed diagnostics; a root-config test was removed. | Validate metadata at persistence and CLI boundaries; use isolated config fixtures. |
| [FIX-2](../../../FIX-2/SUMMARY.md) | Synthetic production API fallback hid failures; workspace defaults guessed configuration. | Keep fixtures in tests and use the same production path for custom projects. |
| [FIX-3](../../../FIX-3/SUMMARY.md), [classification review](../../../FIX-3/iterations/01/REVIEW.md) | Inspection output caused false failures; streaks and anomaly selection needed session scoping; live/offline logic and types drifted. Later iterations expanded prompts and typing. | Share semantic predicates and enums, use actual owner/rule counts, and test source/status provenance. Not every later iteration represents a defect. |
| [FIX-4](../../../FIX-4/SUMMARY.md) | Workspace rename broke tests coupled to live history; configuration loading and reported active paths diverged. | Hermetic fixtures and custom-root propagation checks must survive ticket relocation. |
| [FIX-5](../../../FIX-5/SUMMARY.md) | Shell-expanded nonce was absent from recorded tool arguments. | Supply a literal nonce because telemetry observes the pre-execution input. |
| [FIX-6](../../../FIX-6/SUMMARY.md) | Release Please needed binary cross-compilation and asset attachment. | Trace the actual release producer through packaging, upload, checksums, and execution. |
| [FIX-7](../../../FIX-7/SUMMARY.md) | Downloader complexity and hardcoded initial ticket assumptions required correction; invalid IDs must remain errors. | Verify consumer URLs and custom catalogs; preserve explicit not-found behavior. |
| [FIX-8](../../../FIX-8/SUMMARY.md), [tree review](../../../FIX-8/iterations/01/REVIEW.md) | Codex refresh lost tree members; an empty OpenCode ledger later required regeneration with Codex; artifact paths required reconciliation. | Refresh complete trees transactionally; explicitly select the harness and configured workspace root. Review notes unforced combine/grouping failure branches. |
| [FIX-9](../../../FIX-9/SUMMARY.md) | Codex needed discoverable wrappers with stable canonical role names and shared persona guidance. | Verify harness discovery and inherited permissions rather than copying unsupported metadata. Historical verification did not run automated suites. |
| [FIX-10](../../../FIX-10/SUMMARY.md), [round-2 review](../../../FIX-10/iterations/04/REVIEW.md) | Multi-root truncation and implicit fallback corrupted backfill; growing snapshots double-counted; synthetic identity removal affected unrelated IDs; tests initially missed intended paths. | Preserve owners and multiplicity, compare growth/repeat refresh to fresh extraction, and prove fixtures reach the asserted branch. Limit structural edits to the owning model. |
| [FIX-11](../../../FIX-11/SUMMARY.md), [canonical-root review](../../../FIX-11/iterations/05/REVIEW.md) | Hardcoded ceilings and legacy roots duplicated configuration authority. | Trace non-default values through every consumer and writer; remove retired aliases consistently from docs/templates. |
| [FIX-12](../../../FIX-12/CONTRACT.md), [summary](../../../FIX-12/SUMMARY.md) | Read compaction and hydration dependency hid accepted ledger events; empty traces could imply legacy schema. | Preserve baseline facts on read and optional enrichment failure; test pending/partial responses and obsolete selection races. |
| [FIX-13](../../../FIX-13/SUMMARY.md) | Backend CI lacked generated assets before Go vet/tests. | Check package-load prerequisites, not only the final binary build. |
| [FIX-14](../../../FIX-14/SUMMARY.md) | Frontend copy assumed the embed destination already existed. | Verify absent destination plus repeated build, not just a warm developer checkout. |
| [FIX-15](../../../FIX-15/iterations/02/REVIEW.md) | Local backend checks needed the same prerequisite as CI; four workflows required inspection. | Share the Make dependency graph between CI and local verification. |

## Highest-value documentation improvements

1. **Make fact fidelity a technical invariant.** FCIS and strict typing alone permit incorrect source selection or data replacement. Added authority, session/tree preservation, cumulative-snapshot idempotency, and read-versus-write requirements to TECHNICAL_GUIDELINES.md.
2. **Select adversarial cases while planning.** Added docs/REGRESSION_PREVENTION.md with explicit cases for telemetry growth/replay, missing source members, other harnesses, hydration failures, non-default configuration, and build cold starts. Existing general instructions covered these unevenly.
3. **Use current executable entry points.** CONTRIBUTING.md lacked FIX-15's backend-check/web-assets targets; technical guidance still led with manual build preparation. Both now point to the shared prerequisite graph and explain direct Go command setup.
4. **Keep operational instructions short and discoverable.** AGENTS.md links the checklist and asks maintainers to reconcile active references instead of copying full historical reports. The checklist maps document ownership and distinguishes workspace artifact defaults from strict runtime configuration.
5. **Separate review approval from proof.** Reviews must record tested diff, actual command stages, environment limits, and open findings. Historical success, empty telemetry, and conditional stages cannot stand in for current evidence.

## Remaining limitations and follow-up work

REV-3 changes documentation only. At the audited baseline, service.go still calls broadcastTraceUpdates from polling and suppresses childErr; models.go still declares SkillFact.CallerRole as string and TaskLinkFact.CostUSD without a populated producer. These REV-2 recommendations remain open. The build prerequisite finding has a recorded FIX-15 correction. No runtime remediation is claimed here.

The skill-local technical guide is a separate copy, and active setup/persona references can drift. This change names the root technical guide as canonical and establishes a reconciliation rule; it does not rewrite every skill or create synchronization tooling. Automated documentation checks and stronger regression enforcement are possible follow-ups, requiring their own approved implementation scope.

Historical ticket links and facts remain unchanged. No matching REV-3 plan specification/index entry exists, so none is created or moved. Documentation cannot guarantee prevention: boundary tests and runtime contracts must enforce these rules in subsequent development.

## REV-3 verification

Pending final documentation checks and integration gate; results will be recorded before delivery.

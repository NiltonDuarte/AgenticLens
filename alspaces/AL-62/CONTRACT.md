# AL-62: Source-owned skill and delegation facts

## Current acceptance: iteration 02 supersedes legacy compatibility

User steering on 2026-11-04: "Delete ownerless records on migration. Drop backward compatibility" in direct response to the proposal to retain ownerless history explicitly unresolved. This is explicit approval of the changed boundary. [Iteration 02 PLAN.md](iterations/02/PLAN.md) governs implementation and verification; no additional approval is required for this requested change.

Current v6 persistence, runtime reads, API and frontend expose skill/task facts only inside their owning sessions. Remove root skills/task_links compatibility projections and unresolved_facts storage, API fields, selectors and UI treatment. Older-schema decoding remains solely an explicit migration input boundary; ordinary old-schema reads fail with a migration-required diagnostic. Migration discards ownerless skill/task rows and reports counts by collection; reliably owned facts may relocate without invented ownership. Transient old fields may exist only as nonserialized migration input for the existing older migration chain. They are not runtime aggregation inputs or canonical stores.

The previous preservation/expand-contract decisions below remain historical approval context and are superseded wherever they require ownerless retention, runtime old-schema compatibility, transitional root projections or unresolved-history totals. This user-authorized deletion applies to ownerless legacy rows only. Preserve unavailable owners' already canonical facts, complete-owner refresh/CC-02 protections, source-grounded Codex fixes, native occurrence identity, legitimate multiplicity, unknown outcomes, parent-directed child links and all unrelated cost semantics. Explicit migration does not authorize dropping a known owner's unavailable history or silently rewriting files on GET.

## Proposed contract and approval boundary

Status: approved by the user on 2026-11-04: "Approve this scope and start implementation." Baseline: 19d70574d0e22b00f185c32d048baaf62792743e. Existing branch: AL-62/fix-codex-skill-delegation-facts. Initial Git status was clean. Workspace root is alspaces from .agenticlens-config.yaml.

The user selected two inseparable outcomes: move skills and task_links from ticket-wide canonical storage to their narrowest proven owner, and diagnose and fix missing Codex skill/delegation facts. The linked AL-62 specification, including REV-5 CC-02, supplies acceptance boundaries. Investigation or schema relocation alone does not complete this ticket. AL-63 is downstream; it is not an implementation dependency.

## Grounded diagnosis and limits

backend/internal/domain/models.go defines ticket-wide SkillFact and TaskLinkFact collections without owner identifiers. InvocationFact is a role/model metrics row without an occurrence identity; it must not be mistaken for a raw tool invocation.

backend/internal/adapters/opencode/skills_tasks.go, ExtractSkillTaskFacts, visits session events but discards their session/event/message identities when returning ticket-wide rows. It recognizes OpenCode skill/task calls. It currently requires a locally present direct child to retain the task output's child-session reference, losing an explicit reference when the child is unavailable.

backend/internal/adapters/codex/rollout.go has no skill/delegation output collections. readResponseItem reads call identity/name/output but does not decode arguments. The Codex generation branch in backend/cmd/agentic-lens/main.go does not derive skill/task facts; OpenCode alone calls ExtractSkillTaskFacts. Capture refresh in backend/internal/adapters/capture/reconcile.go likewise derives these collections only for OpenCode. Backfill refreshes Codex sessions without an equivalent Codex auxiliary-fact producer. These are confirmed missing producer/wiring paths. The implementation must reproduce representative supported records and inspect persistence/read behavior separately.

Privacy-safe source inspection found the AL-55 root 01a19267-636b-7d61-85c6-db8d3d3eba77 and four direct children in the local Codex rollout store, all recording cli_version 0.166.0. Root response_item/function_call records contain four explicit spawn_agent calls at JSONL lines 27, 302, 451 and 616, with agent_type planner, coder, tester and reviewer. Their call IDs are respectively call_20e6a816d17f4aeab01f137a43a936e7, call_26ae903280ec481888b576342f285385, call_d97ff4c20c014963baed601d837c43d2 and call_5dda838dc993444380b221d4e6705a84. This proves recorded parent delegation calls independently of child metadata. Result-to-child matching remains to be verified; no match is assumed from order or proximity.

The direct children are 01a19267-a66a-71b1-a7e8-cf880007ce6b, 01a19286-06d4-7881-acd0-580111b793e6, 01a19288-d362-7201-a246-3b42d554ef27 and 01a1928d-556d-7232-9753-245d9283cea7. Their rollouts contain replayed parent calls, so new extraction must preserve ParseRollout's inherited-history exclusion rather than count replay as child activity.

AL-55 messages contain skill catalogue/path references, but the bounded inspection found no structured <skill> attachment. A path mention or a listed skill does not establish activation or success. Exact Codex skill evidence formats, including explicit user invocation, structured attachment and successful skill-document loading where recorded, require source-grounded classification during implementation. Unsupported formats and unavailable outcomes must remain explicit. No raw transcript is copied to Git; source logs and the live AL-55 ledger remain unchanged. Synthetic tests describe parser behavior and do not replace this real-source evidence.

## Canonical ownership and identity

Use session-owned skills and task_links as canonical collections, following SessionFact's existing tools/events ownership. Each occurrence retains available source event, message, tool-call or explicit invocation identity, original turn/message positions, timestamp and recorded values. A source call identity is finer occurrence provenance inside the session; it is not an invented InvocationFact metrics-row ID. If a source proves an unambiguous invocation owner, preserve that explicit identity without coalescing distinct occurrences. Do not attach facts to a metrics row using model, role, timestamp or local index.

| Source family | Proven owner | Finer provenance and constraints |
| --- | --- | --- |
| OpenCode v1 | Session owning the message/part | Retain actual message ID and tool-part/event ID from schema_v1_session_extraction.go. Preserve repeated parts and raw status/error; no owner inferred from local indexes. |
| OpenCode v2 | Session owning the extracted message/content occurrence | Retain actual message/tool/content identifiers when supplied by schema_v2_session_extraction.go. A generated ordering key may retain ordering, but is not described as a native invocation identity. |
| Codex | Rollout session after inherited-history exclusion | Retain native call_id, recorded message/item ID and turn identity when present. Pair outputs by exact call_id within that session. Activation evidence without a call has message/session provenance and unknown execution outcome unless the source establishes it. |
| Historical ownerless facts | Explicit unresolved legacy compatibility collection | Preserve original rows, ordering, multiplicity and values; do not invent an owner. Migration may relocate only when reliable source identity proves a unique owner. |

A delegation belongs to its parent call/session and may reference an unavailable child. Preserve an explicit output child identity without requiring that child to be locally present. Contradictory relationship evidence requires a diagnostic. Child sessions alone never synthesize delegation calls. Follow-up or messaging calls do not become spawn calls merely because they mention a child.

An occurrence's identity is source, owning native session and native occurrence identity when available. Preserve legitimate equal-looking occurrences. Content equality is not a deduplication key across owners or distinct calls. Without a native occurrence ID, preserve source sequence/multiplicity without inventing provenance.

## Evidence and outcomes

Separate recorded activation/attempt from known successful execution. Preserve OpenCode success/error values. Codex requests, skill catalogues, copied prompts, path mentions and ordinary reads cannot automatically become successful skill executions. Recognized explicit activation or loading records may create facts with an evidence kind and unknown outcome when success is unrecorded. Exclude unknown outcomes from a measured success denominator; show unavailable success rate when no known outcomes exist. Keep no recorded facts, unsupported recording and unavailable/partial source evidence distinct in projections. Retain available original values and errors; do not fabricate cost, duration, completion or child identity.

The selected Codex skill producer recognizes a structured user skill attachment with explicit name/path and a direct user skill invocation token naming a selected skill as activation-request evidence, owned by the native user message. It also recognizes an explicit skill tool/load occurrence when that native source format records one. These are distinct evidence kinds: a request/attachment does not prove tool execution, completion or success. Parse only the activation envelope/direct request rather than scanning quoted specifications, catalogues or inherited messages for arbitrary dollar-prefixed words. When the source supplies both attachment and invocation token for one request, retain one activation occurrence with its evidence references, not two counted requests. A plain SKILL.md command read is insufficient by itself; the implementation must not invent a generic shell-content skill execution detector. Validate these selected formats against available real records, report unavailable variants as unverified, and retain deterministic self-contained regression coverage for recognized formats.

## Schema, compatibility and portability

CurrentSchemaVersion is 5; AL-57's implemented v4-to-v5 cost-evidence migration is already present. Add v5-to-v6 through backend/internal/domain/migration's sequential registry; preserve AL-57 cost semantics and all older transitions. Recheck upstream schema ownership before implementation/integration; a new version collision requires reconciliation before publishing, not overwriting a transition.

Expand reads first to accept legacy omitted/empty/populated root collections, v6 owned collections and deliberate transitional projections. New canonical writes target v6 and owned collections only. Store unassignable historical rows in an explicit unresolved legacy representation with ownership diagnostics. Known legacy totals remain available without claiming session attribution. Unknown-outcome representation and evidence kinds are part of v6; legacy bool success remains a recorded historical value.

Pure flattening/projection helpers supply ticket/fleet and compatible API views from canonical owners plus unresolved history exactly once. Root API compatibility collections are derived views, never a second persisted store. Transitional copies must be identified by explicit provenance/representation policy; never suppress equal-looking legitimate occurrences. Ambiguous overlap cannot silently be called a new known occurrence. Document the deterministic precedence/diagnostic rule and test it before removing root writes.

Read/GET performs no persistent migration or hydration write. Explicit migration retains source-unavailable history and resumable diagnostics. Failed transformations do not truncate files. Preserve original ledgers through the existing migration publication/recovery mechanism. Old binaries need not read v6; rollback requires the preserved pre-migration copy or a version-aware exporter, not silent lossy downgrade. Publish adapter/consumer changes as one integrated slice.

## Refresh, generation, capture and backfill

Apply TECHNICAL_GUIDELINES.md's portable ledger contract. Successful complete local extraction replaces only its exact session's snapshot, including legitimate reduction or deletion to empty. Preserve all owners outside that complete set: missing root/children, cross-device sessions, other sources, incomplete/malformed reads and unresolved legacy rows. Partial success is not authoritative coverage of unavailable owners. Use completeness at the owner boundary rather than collection-wide clearing or cumulative union.

Repeated unchanged refresh is stable; growing snapshots replace their own prior occurrences without accumulation. Preserve previous trees and mixed-age/incomplete diagnostics. Optional hydration adds evidence without hiding stored facts when unavailable, pending or failed. Exact identity is the default; heuristic recovery remains explicit opt-in. Local receipts/caches are not authenticity prerequisites for valid committed history. Capture registration/history and publication ownership validation must carry the owned representation and preserve legacy compatibility without weakening unrelated ticket ownership checks.

## Consumers and implementation boundaries

Pure Go domain owns occurrence/provenance values, flattening, migration transformations and aggregations. Adapters own native parsing, validation, I/O, refresh completeness and persistence. CLI generation/formatting, API DTOs, repository projections, capture history, live ingestion and backfill must use owned facts. TypeScript mirrors the contract with strict types and shared selectors; Skills & Subagents, ticket/session views, fleet analytics, comparisons, timeline badges and child navigation use the same projections. Show owning session and available finer source identity; session filters exclude unresolved ownership while ticket totals may include explicitly unresolved recorded history.

Maintain Go/TypeScript FCIS boundaries, zero any/interface{}, no domain I/O or mocks, zero emojis and single-binary embedded assets. Tests use self-contained fixtures, never live ticket workspaces. No unrelated analytics redesign, AL-63 implementation, or bulk modification of historical source ledgers is authorized.

## Acceptance and delivery

The iteration plan maps regressions to existing coverage owners. Required proof follows populated source records through extraction, generate, save/reload, CLI/API and UI/fleet/filter views; validates ownership, replay exclusion, multiplicity, ambiguous migration, child unavailable navigation, hydration failure and partial generate/backfill preservation. Include complete authoritative deletion and unchanged repeat stability. Record real source versions/references and evidence limits separately from synthetic results.

Complete the integrated make gate after upstream integration, canonical frontend build and embedded standalone verification. Independent behavior, data/failure, architecture and test-effectiveness review covers the frozen revision under the configured profile matrix. Parent owns consolidated approval, telemetry and PR publication. Build a local CLI after approval because agentic-lens is not installed and direct Go execution currently lacks embedded dist assets. Telemetry is the final pre-commit step, explicitly selects Codex, checks root/children/source preservation, and never manufactures a ledger when extraction is incomplete.

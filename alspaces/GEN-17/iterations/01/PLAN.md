# GEN-16 iteration 01: Preserve source persona and derive root status

Status: planning complete; awaiting explicit approval of CONTRACT.md before coder dispatch.

## Grounded findings

- backend/internal/domain/models.go, SessionFact.ParentID, already persists hierarchy. backend/internal/domain/live.go, BuildSessionTree, also promotes missing-parent and self-parent nodes for display, so top-level placement alone cannot prove source-root identity.
- backend/internal/adapters/opencode/schema_v1_session_extraction.go, extractSchemaV1Sessions, already preserves nonempty session.agent but assigns RoleRootAgent when a root's agent is absent. Its orphan-part fallback also starts with RoleRootAgent. schemaV1MessageData currently models transport role, not assistant agent.
- backend/internal/adapters/opencode/schema_v2_session_extraction.go, extractSchemaV2Sessions, has the same missing-session-agent fallback. schemaV2MessagePayload.Agent exists but is unused for role resolution.
- backend/internal/adapters/codex/rollout.go, ParseRollout and readSessionMetadata, initialize RootAgent, honor explicit agent_role, and permit agent_path fallback only when a parent exists.
- backend/internal/adapters/live/service.go, GetSessionTree and HydrateWorkspaceEvents, use configured role mappings and child-specific fallbacks. Hydration currently prefers stored invocation roles and can retain legacy RootAgent even when new source evidence is available. Preserve baseline behavior when that evidence is unavailable.
- backend/internal/domain/config.go, RoleMappings.ResolveRole, returns UnassignedSubAgent for blanks; changing that global mapping contract is unnecessary. Resolve unknown roots before applying presentation mappings.
- backend/internal/adapters/invocation_fact_dto.go accepts raw strings. Existing schemas_test.go tests explicitly preserve empty roles. web/src/api/validation.ts checks role as a string without a nonempty constraint. backend/internal/domain/aggregations.go, ComputeRoleMetricsWithCeiling, groups empty roles without dropping their totals; workflow root metrics already use ParentID. web/src/domain/formatters.ts, formatSubagentRole, presents blanks as Unknown, while LiveSessionTree.tsx renders node.role directly. Verify all affected direct renderers and CLI summary paths.

## Implementation sequence after approval

1. Coder owns the affected production Go adapter files above, any necessary pure domain helper, and minimal frontend role presentation changes. Audit current command help and exact CLI/refresh ownership before changing behavior. Parent reports that bin/agentic-lens is absent locally; build the canonical binary before final telemetry. Implement source-persona resolution with explicit metadata precedence, bounded unambiguous message fallback, root-safe Codex agent_path extraction, and no new RootAgent assignments. Retain legacy constant/read compatibility if still needed.
2. Apply the same decision to invocations, events, live nodes and hydration without persisting mapped display roles. Preserve legacy facts during unavailable enrichment. Do not expand this work into per-turn persona histories or refresh redesign. Presentation shows Unknown for blank roles and may add Root beside the persona using source parent linkage, never tree depth alone.
3. Tester owns only necessary changes to existing relevant test files. Inspect coverage before adding cases. Extend coherent tables and fixtures to prove the missing contract behavior; preserve existing factual and boundary assertions.
4. Freeze the reviewed diff/revision and obtain independent focused reviews described below. Address findings through coder/tester, reverify affected checks, and integrate origin/main before required make gate. Parent owns integration, review consolidation, workspace records, telemetry, commits and PR publication.

## Targeted verification

- Extend TestSchemaV1ExtractSessions_RawPersonaPreservation in schema_v1_session_extraction_test.go for named/unnamed roots and child controls, unique assistant-agent fallback, conflicting personas, and matching event role. Existing RootAgent expectations must change only where they describe new extraction, not historical compatibility.
- Extend schema_v2_session_extraction_test.go runtime-evidence fixture to omit session.agent while retaining assistant Agent. Verify extracted invocation/event role and unchanged tokens/tools; include absent/conflicting assistant candidates if not covered coherently elsewhere.
- Extend TestCodexExtractionPreservesRawPersonaAndExplicitRole in codex/rollout_test.go for root agent_path, explicit precedence, and unknown root, preserving existing child evidence. Update main_codex_test.go root expectations through the real generate path.
- Extend live/sqlite_test.go, TestHydrateWorkspaceEvents_FallbackRoleBehaviorForNamelessSessions and TestHydrateWorkspaceEvents_SessionRoleMapUsesSessionID, to verify source-backed root roles, unknown roots, legacy stored RootAgent with and without available enrichment, and unchanged child fallbacks. Reuse existing persisted-data/hydration-unavailable tests rather than adding a parallel suite.
- Verify existing domain aggregation and session-tree tests with a root persona equal to a child persona: topology, root peaks and child effort must remain correct, and unknown-role rows must retain totals. Extend an existing aggregation test only for missing assertions.
- Extend LiveSessionTree.test.tsx and relevant role formatter tests to exercise Unknown and the structural Root label if added, plus a missing-parent child promoted to top level. Keep interaction/collapse behavior covered by existing tests.
- Run affected Go and frontend checks, then make after integration. If frontend changes, run canonical npm build, inspect backend/internal/adapters/web/dist and verify the standalone binary serves rebuilt assets. Record tested diff/commit, actual commands, results and environmental limits. Synthetic source fixtures prove deterministic behavior, not recovery of this real session's persona.

Applicable regression checklist: source-to-persistence fidelity; explicit unavailable values; hierarchy independent of persona; successful refresh ownership and other-source preservation; baseline reads with optional enrichment unavailable; frontend asset freshness. Historical records remain untouched.

## Review assignments and evidence

Before dispatch, parent loads .agents/skills/agentic-lens-review/references/parallel-review.md and model-policy.yaml and expands every selected focus across all configured model/effort profiles. Current defaults are gpt-6.1-sol/high and gpt-6-luna/medium. If unavailable, surface the evidence gap; do not silently substitute or approve an incomplete review.

- Behavior reviewer owns working reports iterations/01/REVIEW_BEHAVIOR_<profile>.md: role precedence, extraction/live parity, user-visible labels and root/child behavior.
- Data and failure reviewer owns REVIEW_DATA_<profile>.md: unavailable/conflicting metadata, legacy compatibility, hydration failures, refresh ownership and provenance.
- Architecture reviewer owns REVIEW_ARCHITECTURE_<profile>.md: FCIS placement, reuse, no redundant topology fields, raw versus mapped role boundaries and bounded live work.
- Test effectiveness reviewer owns REVIEW_TESTS_<profile>.md: production-path assertions, existing coverage reuse, unknown-role retention and regression detection.

Use waves within capacity, one writer per report, common frozen requirements/revision and fresh contexts. Parent consolidates independently verified findings and final verdict; no publication with material unresolved gaps. After approval, retain concise finding/actor/model/fix/verification records in SUMMARY.md and remove completed detailed review evidence per protocol. Telemetry is delayed to final delivery and must use the actual Codex source and conventional type chore.

## Approval boundary

Approve or revise: preserve actual role; remove synthetic RootAgent generation; use parent_id for Root display without a new persisted field; represent unavailable root persona as blank/Unknown; allow only unambiguous assistant-agent fallback; retain legacy ledgers without automatic rewriting. No application changes or implementation tests have run at this planning checkpoint.

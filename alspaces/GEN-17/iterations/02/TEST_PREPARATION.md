# GEN-16 test preparation

Read-only preparation against baseline 99fbcd7 and approved iteration 02. No application/test edits or suites run in this preparation pass. Existing peers were inspected before selecting additions; this is a targeted coverage assessment, not a full audit.

## Existing recipients and missing assertions

- OpenCode v1: extend TestSchemaV1ExtractSessions_RawPersonaPreservation. It already exercises named roots/children and blank roots/children with invocation/event assertions, but currently requires synthetic labels. Preserve its token/tool controls. Add source assistant-agent agreement, explicit session override, conflicting assistant candidates, and user-agent exclusion through ExtractSessions with its real SQLite fixture.
- OpenCode v2: TestSchemaV2SessionRuntimeEvidenceExtractsNormalizedFacts already has assistant agent in its real-format payload and verifies tools/tokens/text. Extend with session.agent absent and matching invocation/event persona; use isolated subcases for conflict and missing child persona. Preserve runtime-evidence controls rather than replacing the fixture with a role-only one.
- Codex: extend TestCodexExtractionPreservesRawPersonaAndExplicitRole. Existing coverage proves child agent_path and explicit agent_role precedence plus private path omission. Add root agent_path, root/child absent and whitespace metadata, explicit root override, and resulting event/invocation parity. Existing child rollout tests still expect UnassignedSubAgent and need intentional expectation updates.
- Live: TestHydrateWorkspaceEvents_FallbackRoleBehaviorForNamelessSessions already owns NULL/whitespace root and child fixtures with actual tool events. Update those expectations to blank and assert no event disappears. Extend coherent hydration/source cases for legacy RootAgent and UnassignedSubAgent surviving unavailable optional enrichment; successful source-backed resolution must not leak configured display mappings into raw facts.
- Domain: config_test.go has multiple blank ResolveRole cases; update their contract without deleting nonblank mappings/identity cases. Existing aggregations_test.go unnamed-child cases at the ses-child-blank and combined-unnamed-children setups assert effort, velocity and context peak; change their expected grouping to blank and retain numerical assertions. A root and child sharing blank persona must still count child-only metrics using ancestry.
- UI: LiveSessionTree.test.tsx already owns nested roots/children; extend with blank root and child display and an orphan whose parent_id is retained. TraceWaterfall.test.tsx owns role-filter composition; add selection of Unknown and verify blank-role events survive that selection while named events filter out, then recover All. TicketDeepDive.test.tsx owns rendered role summaries; assert blank role rows and by-role turns show Unknown rather than an empty label. Existing formatSubagentRole has no dedicated imported coverage in formatters.test.ts; favor UI production-path assertions unless a separate formatting boundary remains uncovered.
- CLI: existing summary/render test tables should verify Unknown appears for blank role without dropping its numerical totals. Preserve stored legacy label tests rather than globally replacing every constant in fixtures.

## Concrete implementation risk found during preparation

TraceWaterfall.tsx currently excludes blank roles from availableRoles and rejects falsy ev.role under an active role filter. A formatter-only change would make Unknown visible in other UI but impossible to select here. Test the actual dropdown and filtered event result, not a duplicated filtering helper.

## Verification sequencing

After coder handoff, inspect its diff and extend only uncovered assertions above. Run focused Go adapter/domain/CLI tests and focused frontend suites, then the parent runs the integrated make gate. Frontend changes require rebuilt embed assets and standalone serving evidence. Synthetic fixtures demonstrate deterministic behavior; they do not prove recovery of this development session's source persona. No live ticket artifacts are used as fixtures.

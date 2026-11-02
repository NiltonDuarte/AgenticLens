# GEN-17: Preserve root session personas

Status: proposed baseline; implementation requires explicit user approval.

## Request and allocation

The user requested the next generic ticket, preservation of root sessions' actual roles instead of assigning RootAgent, and a recommendation about moving or removing that label. The parent verified ticket workspaces, planning index and branch references: GEN-14 is reserved remotely and GEN-15 exists, so GEN-16 is next. Configured workspace root: alspaces. Branch: chore/GEN-17-preserve-root-session-role. Baseline: 99fbcd7be0d0ed05639e4a0cb3be8744a3d5371e. Initial Git status: clean.

## Proposed decision

Remove RootAgent as a newly synthesized role. Preserve the source persona in role. Use existing parent_id to identify structural roots; do not add is_root, root_agent, or another persisted topology field. A displayed Root label may accompany the persona without replacing it. A node promoted to the top level because its parent is absent from a partial tree is not thereby proven to be a source root.

For a root whose persona is unavailable, propose the existing empty-string representation for unknown role, rendered as Unknown at presentation boundaries. Do not guess orchestrator, infer a persona from a skill invocation or user prompt, or reuse the child-specific UnassignedSubAgent label for roots. The existing child fallback remains unchanged. Empty roles already survive InvocationFactDTO conversion, frontend string validation, and ComputeRoleMetricsWithCeiling grouping; formatSubagentRole already renders them as Unknown. Direct renderers and live fallback logic still require correction and verification.

## Source resolution and fidelity

Explicit nonblank session persona metadata wins. Preserve its spelling after existing whitespace trimming; configured role mappings remain presentation policy and must not rewrite raw persisted facts. Codex agent_role remains authoritative; a source agent_path basename must also be usable for roots, not only children.

When OpenCode session.agent is absent, allow fallback to an explicit assistant-message agent only when all nonblank assistant persona candidates agree. User/assistant transport role is not a persona. Conflicting candidates remain unknown; this ticket does not invent a current persona or introduce per-turn role-history segmentation. Metadata absent everywhere remains unknown. Apply consistent resolution to invocation and event facts and to live reads. Keep source lookup work bounded and avoid repeated full-history scans on every poll.

## Compatibility and acceptance

1. Named roots retain their real persona in newly generated facts, API reads, role grouping and rendered summaries; root classification continues to depend on parent linkage.
2. Codex root agent_path and unambiguous OpenCode assistant agent can recover source-backed personas when the primary metadata is absent.
3. Truly unnamed roots produce no synthetic RootAgent role and render a useful Unknown label. Children keep their existing fallback. Unknown roots remain included in totals.
4. Existing ledgers containing RootAgent remain readable and unchanged when source data is unavailable. An explicit successful refresh may replace a source-owned snapshot with recovered facts through existing refresh semantics. No automatic bulk rewrite, fabricated historical persona recovery, schema-version bump or new persisted field.
5. Optional hydration failure preserves persisted facts. Source sessions, IDs, parent links, tokens, costs, invocation multiplicity and other-harness facts remain intact.

## Scope and invariants

Adapter scope: Codex rollout parsing; OpenCode schema-family v1/v2 extraction; live session and hydration role projection; CLI/summary rendering only where required for unknown-role presentation. Domain scope: any minimal pure shared role-resolution or presentation helper and aggregation guards required by the contract. Frontend scope: role display and optional Root badge derived from parent_id. Existing wire fields and configured mapping semantics remain compatible.

Follow TECHNICAL_GUIDELINES.md: pure immutable domain functions, I/O only at adapters, Go and strict TypeScript with zero any, zero emojis, and embedded standalone frontend delivery. Tests use isolated source fixtures, never this workspace. Broader role-history redesign, new topology statuses, global mapping redesign, and historical ledger repair are outside this proposal.

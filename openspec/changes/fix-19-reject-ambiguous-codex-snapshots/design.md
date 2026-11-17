# FIX-19: Technical design

## Scope and approval

Implement the existing [roadmap acceptance](../../../alspaces/plan/to-do/FIX-19%20reject%20ambiguous%20codex%20session%20snapshots.md), selected in the [specification](specs/codex-owner-snapshots/spec.md). The user explicitly instructed: “Implement FIX-19 using the existing roadmap ticket” and stated that this authorizes its existing acceptance criteria, with approval recorded here rather than requested again. This approval covers iteration 01; material acceptance changes require renewed approval.

## Decisions and preserved guarantees

Reject multiple readable native snapshots for one Codex owner before cumulative combination or complete-owner certification, including identical copies, divergent copies and warning-bearing copies. Identity is native session identity; filename, directory, modification time and content equality establish no authority. Discovery must detect ambiguity before child ancestry selection can silently resolve conflicting parent relationships. Keep exact-discovery rejection. Combination also needs a guard against duplicate native owners presented directly or inside already combined inputs; preserve supported nested combination with disjoint owners.

Generation fails before publication. Backfill retains accepted facts for owners outside a successful complete refresh and exposes ambiguity diagnostics. Hydration remains read-only, reports incomplete coverage for ambiguous owners, and continues to expose unaffected owners. Use exact owner selection where a family-level failure would otherwise hide an unambiguous owner. No private receipt or source corroboration is required to retain supported Git-shared history. Repeated invocations, skills and delegations within one transcript retain multiplicity.

## Ownership and iteration 01

Planner owns initial execution artifacts, ticket backlink and In Progress index update. Coder owns scoped changes under backend/internal/adapters/codex, backend/internal/adapters/live, backend/internal/adapters/backfill and backend/cmd/agentic-lens as needed. Tester owns adversarial regression additions in those test packages; coder and tester coordinate writes sequentially. Parent owns integration, final evidence, telemetry, commits and PR publication. All preserve others’ work.

Select extraction/refresh, optional hydration and portable-ledger regression boundaries from docs/REGRESSION_PREVENTION.md. Verification must cover active/archive JSONL/gzip, duplicate root/child and conflicting relationships, warning-bearing copies, genuine repeated occurrences, unaffected same-harness and other-harness history, failed publication preservation and unchanged repeats through actual consumers. Four independent review focuses (behavior, data, architecture, tests), each Sol 6.1/high and Luna 6/medium, investigate a frozen candidate in capacity-limited waves. Parent records actual model availability and evidence; missing coverage stays incomplete.

## Current-main revalidation

Baseline and fetched origin/main are b6ab0471224c8bcea7bc00adce867d3ba0de4a37 (AL-63, after AL-84). Source inspection confirms the counterexample remains reachable: DiscoverRolloutPaths appends every selected candidate without rejecting repeated threadID; CombineTranscriptsWithPolicy appends session facts, invocations and tools and sums ToolCalls per transcript; CLI generation and backfill readCodexOwners consume this family. HydrateWorkspaceEvents discovers the family then breaks on the first parsed matching owner, certifying it complete when warning-free. DiscoverExactRolloutPathWithBudget already rejects a second matching owner. Executable before/after proof remains a required tester step rather than a claimed runtime result.

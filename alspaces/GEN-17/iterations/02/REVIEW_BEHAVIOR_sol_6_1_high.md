# GEN-16 iteration 02 behavior review

Verdict: CHANGES_REQUESTED. One substantiated source-persona integration defect remains in the scoped investigation.

Run: behavior_sol_6_1_high; phase: independent discovery; focus: behavior and integration. Requested and dispatch-configured identity: codex / gpt-6.1-sol / high. Source-observed model, effort, usage and cost were not available to this reviewer and are not asserted.

Baseline: 99fbcd7be0d0ed05639e4a0cb3be8744a3d5371e. Reviewed source fingerprint: 785b903e14bb4af8923e054f2d8d5fb69f44210150b172d431d75aa5416d6f9a, independently recomputed with the REVIEW_RUNS.json method: SHA-256 over git diff HEAD --binary -- backend web, followed by sorted untracked backend/web path bytes, NUL and file bytes. Requirements: CONTRACT.md as superseded by iterations/02/PLAN.md for unknown children. Sibling personas and reports were not read.

## BEH-SOL-01: Source persona recovery is discarded for matched persisted events

Priority: P2. Required-contract defect. Locations: backend/internal/adapters/live/service.go:963-968, withLocalEvents at service.go:842, and unchanged backend/internal/domain/events.go:118-131.

Trigger: a workspace ledger already contains an event that matches an available source event by source event ID or unique occurrence, while the source session now supplies a persona different from the stored event role. This includes an old RootAgent event or an unknown blank event whose persona can now be recovered.

Trace: HydrateWorkspaceEvents resolves explicit session.agent ahead of stored roles and assistant evidence, writes the result to sessionRoleMap, and builds the local event with that resolved Role. Its withLocalEvents callback then calls ReconcileEvents. The matching branch invokes enrichEvent, which retains recorded.Role and copies only payload, references and duration from the local event. The recovered role is therefore discarded for existing events. Newly appended events use the recovered role, so the same session can expose inconsistent personas depending on whether its events were previously recorded.

Consequence: successful API hydration can continue displaying Unknown or RootAgent in TraceWaterfall and its role filter despite authoritative source metadata being available. A stale named persona likewise wins over explicit source metadata. This contradicts consistent source-backed role resolution in live reads. No persisted rewrite is needed to correct the read projection; unavailable source data must continue preserving baseline facts.

Evidence: a disposable Go overlay added TestGEN16ReviewMatchedSourceRole to package live_test without writing product or test files in the checkout. The probe constructs an isolated SQLite database using createTestDBFile, inserts session.agent="Build" and assistant agent="coder", and hydrates a ledger containing the matching event ID opencode:message:m_review. The stored role is varied across blank, RootAgent and stale-persona. Each result has IsHydrated=true, exactly one event and the source prompt payload, proving successful source lookup and event reconciliation. All three retain the stored role instead of Build:

```text
hydrated=true rows=1 role="" source_content="source payload"
hydrated=true rows=1 role="RootAgent" source_content="source payload"
hydrated=true rows=1 role="stale-persona" source_content="source payload"
```

Executed from backend:

```text
go test -overlay /tmp/gen16-behavior-r_89l0uf/overlay.json ./internal/adapters/live -run '^TestGEN16ReviewMatchedSourceRole$' -v
```

Exit status: 1, with all three expected-source-role assertions failing. Overlay fixture: /tmp/gen16-behavior-r_89l0uf/probe_test.go. These are synthetic deterministic fixtures, not observations about real session personas.

Classification: enrichEvent's baseline-role retention predates this diff. The new service role-resolution path exposes an unaddressed integration dependency and does not satisfy the newly accepted source-wins behavior for matched events. This is not an introduced persistence corruption or a request for bulk historical rewriting.

Required correction: ensure authoritative, successfully resolved source persona evidence reaches the final hydrated event projection for matched events, including a resolved unknown when assistant evidence conflicts. Preserve unmatched baseline occurrences, input immutability and unavailable-source fallback. Add coverage with actual baseline events; the extended TestHydrateWorkspaceEvents_SessionRoleMapUsesSessionID currently supplies invocations but no baseline events, so its rows are appended and bypass this failure.

## Other consequential hypotheses assessed

- New synthetic placeholders: production searches found neither placeholder synthesized by Codex, OpenCode v1/v2, blank RoleMappings resolution or workflow aggregation. Remaining constants support old stored labels and hydration compatibility decisions.
- Source fidelity: extraction trims explicit session personas, keeps raw spelling, considers assistant-only agent candidates, preserves conflicts, and updates invocation/event roles together. Codex root agent_path now reaches the basename fallback; explicit nonblank agent_role wins.
- Unknown totals and topology: RoleMappings preserves blank input; role ordering includes its blank bucket; aggregation groups trimmed blank roles without dropping invocations. Root/child cost, turns and peak separation still uses ParentID. BuildSessionTree retains parent links when promoting missing-parent children for display.
- Persistence and optional hydration: DTOs preserve blank strings and old labels. DeriveMetricsLedger applies configured mappings to copies and returns original Facts. Unavailable source fallback returns committed events; no write path was introduced by hydration. Existing missing-source and legacy-label tests cover that branch.
- Presentation and filters: formatRoleLabel presents blank/whitespace as Unknown, while dropdown values retain the raw blank key; TraceWaterfall compares trimmed event roles without rejecting blanks. MultiSelectDropdown supports the blank string option. Role, effort, peak, thread and markdown renderers retain the unknown grouping. The integration finding above concerns the values supplied to this presentation, not the blank label itself.

## Verification and limits

Reused TEST_REPORT.md evidence for this fingerprint: backend go test -race ./... passed all 12 packages; frontend npm test passed 35 files / 783 tests; final selected Unknown-filter regression passed after the test-only context_tokens correction; typecheck and lint passed on the final test state. The reviewer inspected relevant test setup and assertions rather than treating suite status as a correctness conclusion. Independently executed the isolated failing hydration probe, source fingerprint recomputation and git diff --check (passed).

Scope included Codex rollout extraction, both OpenCode schema-family extraction paths, live session persona cache/projection and workspace hydration; domain mapping/role ordering/workflow and role/thread aggregation; DTO/repository reads and CLI markdown formatting; frontend Unknown displays and trace filters, with relevant unchanged reconciliation, topology and dropdown dependencies. Product and shared test files remained read-only; only this assigned report was written.

No full-suite rerun, actual user source-session recovery, browser geometry, rebuilt embedded frontend or standalone serving was performed by this reviewer. Parent owns the integrated make/build/assets gate, fixes, consolidation and publication. This verdict addresses the assigned behavior scope only.

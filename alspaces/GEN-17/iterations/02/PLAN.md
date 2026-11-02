# GEN-16 iteration 02: Unknown personas for every session

Status: explicitly preapproved; proceed to implementation under this delta and the preserved baseline contract.

## Steering and supersession

Source: the user said "apply this recommendation" after the parent recommended Unknown for missing roles in both root and child sessions, with ancestry represented separately. That instruction approves the baseline proposal and expands its missing-persona rule to children. No further scope approval is needed for this implementation.

Changed acceptance: newly extracted facts and live projections must synthesize neither RootAgent nor UnassignedSubAgent when persona evidence is absent. Preserve an empty raw role and render Unknown for both roots and children. RoleMappings.ResolveRole must also preserve the unknown result for blank input instead of manufacturing a child role. Aggregation and downstream fallbacks must not reintroduce either synthetic label for these new unknown facts.

Superseded decision: CONTRACT.md's "existing child fallback remains unchanged" and iteration 01's child-fallback preservation are historical proposed boundaries, superseded only by this explicit user instruction. The frozen files remain intact. This delta governs that conflict; it does not authorize unrelated mapping or history redesign.

## Preserved guarantees

- Explicit nonblank source persona wins and remains a raw fact; existing configured mappings for nonblank names retain their presentation semantics.
- Codex source-backed agent_path and unambiguous OpenCode assistant-agent fallback remain approved. Conflicting or unavailable candidates remain unknown. Transport user/assistant roles and skill names are not persona evidence.
- Existing parent_id remains the topology authority; no new persisted root field. A missing-parent child promoted for display remains a child according to its retained source parent link.
- Existing ledgers containing either legacy label remain readable without automatic rewriting. Unavailable optional hydration preserves accepted persisted facts. Successful source refresh follows existing ownership rules, preserving other sessions, harnesses and invocation multiplicity.
- No schema bump, bulk historical rewrite, per-turn persona history, new topology statuses or fabricated recovery. FCIS, strict typing, embedded frontend requirements and all remaining baseline boundaries still apply.

## Implementation and ownership

Coder owns affected production paths: backend/internal/adapters/codex/rollout.go; OpenCode schema_v1_session_extraction.go and schema_v2_session_extraction.go; live/service.go and any grounded live role-projection dependencies; backend/internal/domain/config.go and aggregations.go where blank fallback currently creates UnassignedSubAgent; affected CLI/summary and frontend role display paths. Follow iteration 01's source-grounded resolution and bounded live lookup requirements. Inspect every active use of both constants to distinguish legacy compatibility from new synthesis before changing it. Keep constants only where compatibility or explicit source values require them.

Tester owns necessary existing test-file updates. Parent retains integration, review consolidation, telemetry, staging, commits and PR publication. Everyone shares the checkout and must preserve other owners' changes.

## Verification impact

Extend iteration 01 coverage to missing and whitespace-only child personas in Codex, both OpenCode schema families, live session trees and hydration. Test named children and roots as controls; explicit source precedence and conflicting assistant metadata must remain correct. Existing config_test.go blank-input cases for RoleMappings.ResolveRole must now expect an empty role, while nonblank configured mappings remain covered unchanged.

Update existing domain aggregation cases to ensure unnamed child invocations remain included in child effort, costs, peaks and role totals without a synthesized UnassignedSubAgent grouping. Root versus child calculations must still use parent linkage, including when both personas are unknown or equal. Retain compatibility cases for stored RootAgent and UnassignedSubAgent with source enrichment unavailable.

Presentation tests must verify useful Unknown output for both root and child blanks through relevant UI and CLI rendering; role filters must still expose and retain unknown rows. Extend coherent tests before adding declarations and use isolated fixtures. Reuse valid unchanged checks, run affected checks after changes, then the integrated make gate and canonical embedded-asset verification if the frontend changes.

## Review and delivery

Use iteration 01's four independent review responsibilities and configured model profiles, with working reports owned under iterations/02 instead of iterations/01. Parent loads the current parallel-review protocol and model policy before dispatch, freezes common requirements/revision, records actual model/effort evidence and consolidates approval. Include blank child-role propagation and legacy preservation in behavior, data/failure and test reviews. Follow the baseline final telemetry and delivery sequence; do not generate telemetry during planning or implementation.

# GEN-17: Preserve source session personas

## Delivered behavior

- Domain: retain the unknown role bucket in ordering, role/thread totals, effort and peaks; structural root/child calculations continue to use parent_id. Shared pure persona evidence preserves explicit metadata and rejects conflicting assistant candidates.
- Adapters: Codex supports root agent_path fallback; OpenCode v1/v2 recover only unambiguous assistant personas. New facts synthesize neither RootAgent nor UnassignedSubAgent. Successful source hydration projects recovered roles onto matched events through an explicit source-only reconciler. Failed metadata reads preserve committed OpenCode facts and independent Codex enrichment. Live trace and persona evidence share one immutable cached snapshot.
- Presentation: missing roles display Unknown in dashboard, trace filters, metrics and CLI summaries. Existing known ledger personas remain when optional live evidence adds no persona; explicit source evidence overrides them. Historical labels remain readable. No new persisted root field, schema bump or bulk ledger rewrite.

## Scope, ownership and ticket correction

User approval: "apply this recommendation" approved Unknown for missing roles in roots and children, expanding the proposed baseline's child fallback. [Iteration 02](iterations/02/PLAN.md) controls that delta; CONTRACT.md retains the proposed baseline. Planner authored the plans, coder owned production implementation, tester owned regression changes, independent reviewers owned scoped investigation, and the parent owned integration and delivery.

Initial Git status was clean at 99fbcd7be0d0ed05639e4a0cb3be8744a3d5371e on generated branch GEN-13/root-session-role. Initial local workspace/index/reference checks allocated GEN-16. User later explicitly corrected this: "GEN-16 was taken. change the workspace to GEN-17 and merge main into the branch". The same work moved to alspaces/GEN-17 and chore/GEN-17-preserve-root-session-role; no application/test changes were required for relocation. Historical finding IDs retain GEN16 as provenance. Checkpoint b8c087a committed owned work; merge 124419d integrated origin/main 8f0fd4f cleanly. Incoming independent GEN-16 artifacts were preserved. No matching GEN-17 planning specification or index entry required lifecycle updates.

## Consolidated review: APPROVED

Reviewed final content: integrated 124419da3aadd170b9fe6e15e0d7bdcc1ebef200 plus backend/web diff fingerprint 3b083d30617cefbe537a8838c4fc441881988aff7f2588fcbaa5d60916260f60. The fingerprint is SHA-256 of git diff HEAD --binary restricted to backend/web followed by sorted untracked source path, NUL and bytes. Application/test content was frozen through the final reviews. Integration changed guidance only. Sol behavior/architecture correction approvals carry forward across the final failclosed metadata branch; final data/test investigations and targeted controls cover that branch.

[Compact findings and every reporting persona/model/effort tuple](iterations/02/REVIEW_SUMMARY.json) permanently retain all six candidates, including shared discovery, refutation and accepted debt. Model/effort attribution below is dispatch-acknowledged configuration, not independently source-verified identity.

| Finding | Discovering responsibility and model/effort | Disposition and verification |
| --- | --- | --- |
| GEN16-01 | Behavior, data/failure and test-effectiveness reviewers, each gpt-6.1-sol/high | Matched events discarded recovered persona. Coder gpt-6.1-sol/high added explicit source reconciliation; tester gpt-6.1-sol/high added stable-ID, unique legacy and appended controls. Independent pre-fix probes failed; parent and reviewers verified corrected behavior. |
| GEN16-02 | Architecture reviewer, gpt-6.1-sol/high | Split cache reads could lose a stable persona. Coder gpt-6.1-sol/high paired trace/evidence snapshots; tester gpt-6.1-sol/high added deterministic real-SQLite invalidation guard. Parent bounded concurrent race probe and focused reviewers verified correction. |
| GEN16-03 | Data/failure reviewer, gpt-6-luna/medium | Confirmed pre-existing global invalidation cost. Parent verified identical baseline behavior; accepted debt owned by live-service maintenance for future per-session invalidation work. No new history query introduced. |
| GEN16-04 | Test-effectiveness reviewer, gpt-6-luna/medium | Test gap only. Tester gpt-6.1-sol/high extended the existing orphan builder test to assert original parent_id and input ancestry. Focused reviewers verified the oracle; no production defect claimed. |
| GEN17-05 | Data/failure re-reviewer, gpt-6.1-sol/high | Failed metadata lookup erased an event-only recorded role. Coder gpt-6.1-sol/high now fails closed for OpenCode; tester gpt-6.1-sol/high added partial-failure and mixed-source controls. Pre-fix overlay failed three cases; corrected races and final data/test reviews passed. This distinct defect was discovered during re-review. |
| GEN17-06 | Data/failure re-reviewer, gpt-6-luna/medium | Refuted interpretation: absent optional persona evidence does not prove a known recorded role stale. Existing persisted-planner control and baseline-preservation contract support retention; final test review agreed. No product change. |

Initial eight discovery runs completed: all four Sol scopes requested changes; Luna behavior/architecture approved, data raised existing performance debt, tests identified the ancestry assertion gap. Re-review verified corrections under both profiles. Final behavior, architecture and test scopes approve under both profiles; final Sol data approves. Luna data's completed dissent is retained and independently refuted as above. Parent consolidated APPROVED only after integrated make and standalone checks passed; this does not authorize merging the PR. Verification actors are separate from discovering reporters; no supplied-finding verification is counted as independent discovery.

## Verification

- Integrated make passed on the final reviewed content: all 12 Go packages under the race detector; frontend 35 files / 783 tests; typecheck and ESLint; Python release and installation checks; canonical frontend build and Linux/macOS amd64/arm64 builds. Log: /tmp/gen17-make.log (local evidence, not a committed transcript).
- Parent verified frontend and Go embed output trees byte-for-byte, then ran the Linux/amd64 standalone binary with Node absent from PATH in a disposable workspace. Current index, JavaScript and CSS served byte-identically; server terminated. Generated assets remain ignored per policy.
- Targeted regression and before/after evidence: [TEST_REPORT.md](iterations/02/TEST_REPORT.md). Final metadata-failure and mixed-source controls preserve ledger bytes, accepted facts and other-source multiplicity. Legacy labels remain readable with enrichment unavailable. git diff --check passed.
- The original 2000-poll race/count3 stress attempt timed out after 40 seconds and is not a pass. A bounded 20-poll race/count1 copy passed in 1.889 seconds; the durable snapshot test uses deterministic interleaving. Browser geometry and actual external-session persona recovery were not exercised.

## Remaining limits and telemetry

Database-wide invalidation can reread histories of unchanged active sessions after unrelated writes; idle polls reuse snapshots. This is existing accepted performance debt, not changed-session-only polling. Pricing, cost, source identity and session completeness must come from final extraction, not reviewer assignments. Final telemetry is generated after all review/integration work; extraction facts and limits are recorded below when available.

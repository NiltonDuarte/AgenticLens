# AL-67 iteration 02: follow-up ticket creation

## Steering delta and approval

Source: user follow-up after PR112 publication: "Create tickets to fix the issues found in the assessment"; multiple FIX bug tickets and REV investigations are explicitly authorized, related issues should stay together, explanations should be concise and examples are welcome.

Changed acceptance: create actionable downstream specifications and registry entries, link every assessment finding to an owner, and enrich existing overlapping tickets instead of creating duplicates. This supersedes the initial assessment's statement that no follow-up tickets had yet been created. The user request pre-approves this documentation scope; no further confirmation is needed.

Preserved guarantees: no application/test/schema/runtime changes or downstream implementation; no switch from active AL-67 into a new PO/FIX/REV working session. Use product-owner ticket format as guidance under AL-67, with no premature developer workspaces/contracts. FIX and REV prefixes are explicitly user-authorized; existing numbering and active roadmap are checked before allocation. Historical findings, severities, discovery credit, verification and source limitations remain intact.

Baseline: 74f6d985fdf9851b241e11d9b0eb01d13c0f3d26 on docs/AL-67-telemetry-repository-review, initially clean. PR112 is open. origin/main remains a84adbe5be46c2ef52ede0987a417562715d38e1. Existing reserved IDs through FIX-17 and REV-6 were checked across workspace paths, specifications and registry. No external issue service is available; repository plan specifications and TICKETS.md are the ticket system used here.

## Planned result

1. Create FIX-18 for explicit live schema/capability reporting (F03), FIX-19 for duplicate Codex owner snapshots (F04), FIX-20 for native-read cancellation (F06), and REV-7 for shared read contracts/snapshot/ordering and architecture boundaries (F07/F08/F09/F12).
2. Keep configured source selection/path resolution and stronger no-fallback proof together in AL-70 (F01/F02/F11). Keep failed child-lookup visibility in existing AL-69 (F05), and required-column test-oracle repair in existing AL-74 (F10). Add examples and current evidence. Correct AL-74's stale test-only accessor assumption: production hydration now calls that accessor.
3. Give each ticket a brief problem, illustrative example, goal, scope, measurable acceptance and explicit exclusions. New fixes/investigation are Ready with no invented blocking dependencies. Review-only work does not authorize a refactor.
4. Add a concise ownership map, update existing compact dispositions without changing findings/reporters, and update summary/PR to describe the final documentation delivery.
5. Validate numbering, all twelve finding owners, titles/statuses/index paths, links and unchanged application scope. Commit only task-owned docs. Reuse valid integrated application checks from iteration 01 because no application/build inputs changed; no duplicate full suite or new tests are needed for ticket prose. Integrate current main; record any new impact rather than claiming old results cover changed code.
6. Refresh source-grounded AL-67 telemetry at the final step, preserving existing native owners and reporting the known T3 linkage/cost limitations. Commit/push and link the open PR; do not implement or merge child work.

## Review proportionality

This iteration authors task specifications and tracking metadata; it does not repeat or expand the architecture audit. Coordinator validates requirements, evidence mapping and scope directly. No new delegated reviewers are required for this narrow documentation follow-up; the eight audit focus/model cells from iteration 01 remain valid for unchanged application evidence and are not relabeled as reviews of these new specifications.

## Verification result

Coordinator validation passed: four unique new Ready specifications and matching index entries, no premature child workspaces, every F01-F12 mapped exactly once, all 28 existing discovery tuples and severities preserved, local links resolve, PR draft below 3000 characters, whitespace clean and changed paths limited to owned planning/artifact documentation. No application code, tests, runtime or configuration changed. No new suite execution is claimed; prior full make and standalone-asset evidence remains valid for identical application/build inputs.

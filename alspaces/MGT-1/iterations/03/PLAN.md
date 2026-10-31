# MGT-1 Iteration 03: supervise ticket threads through merge

Source: user on 2026-10-31 explicitly requested control of each open PR/ticket thread until merge, beginning with a reusable skill. Release PRs are strictly human controlled and must never be merged by agents.

Baseline: 40ee4b6; earlier audit and reclassification accepted. Acceptance delta: create agentic-lens-management, map all current PRs to owning tasks, send user-proxy instructions to existing independent threads, coordinate review fixes/upstream integration, and merge ready non-release PRs serially. Supersedes the original review-only/no-message/no-merge boundaries, not the original findings or ownership rules. No manager application implementation.

Preserve: independent ticket roots/telemetry, owners' approved product scope, source evidence, adjacent features, no release merge or auto-merge, no main checkout. Current candidate sequence: #76, #79, #75, #77, #78 after repairs; #81 documentation audit must also be assessed; management #80 integrates the final skill/control record. Re-evaluate against current heads.

Verification: skill frontmatter validation and scenario review, canonical guidance alignment, owner return evidence and exact-head CI/review/merge gates. Full make required after integration; application changes stay in their owning development workflow. Record PR/thread inventory and tool results in CONTROL.md and iteration REVIEW.md. Hold unavailable owners explicitly and continue unblocked tickets.

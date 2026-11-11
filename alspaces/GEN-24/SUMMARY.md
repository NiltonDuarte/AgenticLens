# GEN-24: Telemetry-led presentation portal

[Pull request #138](https://github.com/NiltonDuarte/AgenticLens/pull/138) · [Live preview](https://niltonduarte.github.io/AgenticLens/preview/138/)

The presentation portal now uses the platform’s navy surfaces and blue/cyan accents. Skills, agents and roles, models, and context-window use lead the page. Project evolution and ticket comparisons explain how recorded code changes, cost, tokens, cache use, duration and developer steering help teams assess changes to guidance and practices. Anomalies remain a supporting capability. Git-versioned ledgers provide the shared history for teammates and engineering leads. The user approved this scope with “$agentic-lens-dev approved” after the [origin-story correction](iterations/01/PLAN.md#scope-clarification). Application and dashboard behavior are unchanged.

## Preview delivery

Same-repository PRs targeting main publish beneath `/AgenticLens/preview/<PR-number>/`. Production builds from main; isolated read-only PR builds feed a serialized static publisher. A durable state branch preserves other open previews. Pending transaction markers prevent a later publication from silently losing state after an interrupted deployment. Recovery and the 100-pending-run queue bound are documented in the [Pages guide](../../.github/pages/README.md).

The initial bootstrap used the explicitly approved, reviewed feature branch and a dedicated `github-pages-preview` environment restricted to that exact branch. Production’s `github-pages` policy remains main-only. Automatic lifecycle handling starts once this workflow reaches main; fork previews are excluded.

Hosted run [43023112171](https://github.com/NiltonDuarte/AgenticLens/actions/runs/43023112171) passed all four jobs and published reviewed implementation `80e83c662274e381be96d038f912b9393bc3fc26`. Public HTTP checks verified seven local destinations and their fragments, the telemetry hero and navy stylesheet. Production homepage, presentation CSS, roadmap index and roadmap CSS matched the pre-deployment SHA256 hashes. Durable state records the expected main/preview revisions and has no pending marker. A subsequent publication is being verified before final handoff.

## Verification and review

Baseline `167c204fc958da34eb82d97129dedd85ac9b8c58` had a clean checkout. Reviewed candidate `80e83c662274e381be96d038f912b9393bc3fc26` integrated origin/main without changes. Both initial and post-integration `make` passed; all six hosted CI checks passed on that revision. Later ticket-artifact-only changes reuse this evidence because product, tests and workflow are unchanged.

Native Jekyll passed 8 runs / 453 assertions; Python passed 9 tests, including actual local Git pending-state handling and retarget rejection/removal. The actual candidate build shell passed on pre-change source lacking new composer files (8 runs / 335 assertions and a prefixed build). Ruby 3.3 ran in Docker. Playwright at 1440, 768 and 375 pixels loaded every image without horizontal overflow or page errors, with first-Tab skip focus; desktop/mobile screenshots were inspected. [Test evidence](iterations/01/TEST_REPORT.md) preserves boundaries. Latest actionlint lacks the documented queue key; the rest validated locally and successful hosted execution established exact workflow acceptance.

Four review focuses ran on both configured models; source metadata confirmed execution profiles. C1 publication findings were corrected and every affected focus approved C2. `gpt-6.1-sol`/high behavior, data and architecture reviewers found queue loss; its behavior reviewer found older-head incompatibility. Data reviewers on `gpt-6-luna`/medium and Sol found retarget inconsistency, with initial impact dissent. A Luna bootstrap-trust claim was refuted against the explicit reviewed-branch boundary. Test reviewers verified distinct assertions with disposable mutations. The parent consolidated approval after integrated checks and hosted evidence; [retained findings](iterations/01/REVIEW_SUMMARY.json) preserve all discovery attribution and dissent. No unresolved product findings remain; telemetry extraction is deferred to final delivery.

# MGT-2 management contract

User authorization, 2026-11-03: supervise ongoing AgenticLens threads, guide owners, choose merge order, report progress, and continue until eligible changes are merged. Routine owner instructions and gated non-release merges are authorized. Product scope changes still require a user decision.

In scope: current independent owners of AL-60, AL-61, AL-62, AL-76, and reporting the AL-68 release-marker work. Rediscover late-arriving PRs and verify ownership before including them. No new ticket threads requested.

Iteration 02 scope approval, 2026-11-03: the user answered “Authorize AL-60 shared capture repair (Recommended)” to the explicit proposal to expand AL-60 into a diagnosed shared capture repair before remaining merges. AL-60 owns the minimal capture repair and related deterministic regression coverage; no duplicate capture repairs by other owners. This supersedes the previous read-only repair hold. All review, integration, preservation and release exclusions remain in force.

Release PRs, including the explicitly release-only AL-68 marker and Release Please #55, are human_only. This manager's own PR requires external review and merge. No self-merge, auto-merge, protection bypass or application implementation by this manager.

Owners retain implementation, tests, review and telemetry. Manager owns only alspaces/MGT-2 coordination artifacts, instructions, integration sequencing and eligible merge actions. Require fresh exact-head review/check/base/dependency evidence before each merge; verify actual merge SHA afterward.

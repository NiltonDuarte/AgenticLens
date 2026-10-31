# chore(telemetry): revert automatic content-error classification (AL-43)

Reverts the automatic content-error classification introduced in PR #78 at the user's request. Stack-specific output heuristics do not provide a consistent platform metric across adopting projects.

Removes the classifier, content-warning metrics and API fields, extraction wiring, and dashboard indicators. Retains current-main execution classification exactly, per the user’s approved correction: "retain main’s execution classifier while removing content classification, and add these regression cases." This supersedes iteration 02’s mechanical pre-PR #78 restoration. Regression coverage protects successful reads containing timeoutMs, grep output quoting panic, and explicitly zero-exit output quoting permission denied, including live/offline counters and failure streaks. Previously written ledgers remain readable through decode-only support for retired fields. Historical AL-43 records remain available.

Contract: [CONTRACT.md](CONTRACT.md)
Implementation and verification: [SUMMARY.md](SUMMARY.md)

# chore(telemetry): revert automatic content-error classification (AL-43)

Reverts the automatic content-error classification introduced in PR #78 at the user's request. Stack-specific output heuristics do not provide a consistent platform metric across adopting projects.

Removes the classifier, content-warning metrics and API fields, extraction wiring, and dashboard indicators. Restores execution classification to its pre-PR #78 behavior. Previously written ledgers remain readable through decode-only support for retired fields. Historical AL-43 records remain available.

Contract: [CONTRACT.md](CONTRACT.md)
Implementation and verification: [SUMMARY.md](SUMMARY.md)

# fix(telemetry): preserve source-owned skill and delegation facts (AL-62)

Codex recorded skill loading and delegation evidence was discarded before reaching skill analytics. Preserve source occurrence identities under their owning sessions and carry those facts through generation, persistence, refresh and presentation.

Migrate legacy history without guessing ownership, retain unavailable owners during partial refresh, and distinguish recorded loading from unknown execution outcomes.

Contract: [CONTRACT.md](CONTRACT.md). Implementation and verification: [SUMMARY.md](SUMMARY.md).

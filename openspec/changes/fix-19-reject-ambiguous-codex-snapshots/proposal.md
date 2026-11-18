# FIX-19: Reject ambiguous Codex session snapshots

Codex active and archived rollout copies can describe the same native owner. Prevent those cumulative copies from being combined or chosen arbitrarily while preserving accepted portable history and real repeated occurrences.

The selected scope is the existing [FIX-19 roadmap ticket](../../../alspaces/plan/done/FIX-19%20reject%20ambiguous%20codex%20session%20snapshots.md). Historical source findings are linked there; current source revalidation is recorded in [design](design.md). Generation, refresh/backfill and historical hydration are the affected consumers. No authority precedence, row deduplication, private receipt requirement or source-store cleanup is introduced.

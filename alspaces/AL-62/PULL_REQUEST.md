# fix(telemetry): preserve granular skill and delegation facts (AL-62)

Codex extraction omitted skill and delegation facts, and ticket-level collections lost their source ownership. Schema v6 stores occurrences within their owning session, retaining native call/message provenance and explicit unknown outcomes. Codex and OpenCode producers, refresh/backfill, CLI/API and dashboard consume the owned representation.

Current runtime requires v6. Explicit migration drops ownerless legacy skill/task records and reports counts; root compatibility arrays and unresolved-history storage are removed by explicit user request. Partial refresh preserves unavailable owners' already-owned facts. Child references remain explicit when the child is unavailable.

Contract and accepted steering: [CONTRACT.md](alspaces/AL-62/CONTRACT.md). Delivery details and verification: [SUMMARY.md](alspaces/AL-62/SUMMARY.md).

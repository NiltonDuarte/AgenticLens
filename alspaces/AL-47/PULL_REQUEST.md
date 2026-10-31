# feat(live): refresh active session traces automatically (AL-47)

Selected active sessions could show stale traces when SSE updates were missed. Keep the inspector current with a coalesced two-second REST recovery loop alongside the existing SSE stream, and stop automatic refresh when the session is no longer active or streaming is paused.

Protect selected traces from obsolete REST and manual-refresh responses, retain valid data on refresh errors, and preserve expanded reasoning and reading position when messages grow. Reuse unchanged backend snapshots through read-only SQLite change detection, avoiding duplicate full-trace reconstruction on idle polls without changing the API or ledger schema.

Contract: [CONTRACT.md](CONTRACT.md). Implementation and verification: [SUMMARY.md](SUMMARY.md).

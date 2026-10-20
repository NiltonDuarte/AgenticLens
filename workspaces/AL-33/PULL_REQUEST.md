# feat(live,web): dynamic local database event hydration and dual-plane waterfall (AL-33)

## Summary of Changes

Delivers an end-to-end vertical slice for dynamic local database event hydration and dual-plane waterfall visualization across the Go backend and TypeScript frontend for ticket AL-33:

1. **Domain Core**:
   - Added `WorkspaceEventsResponse` struct (`ticket_id`, `is_hydrated`, `events`, `data_source`) and extended `EventFact` with optional in-memory diagnostic pointers (`PromptContent`, `ReasoningContent`, `ToolInput`, `ToolOutput`).
   - Defined canonical data source constants `DataSourceLocalSQLite` and `DataSourceCommittedLedger`.
   - Updated `PruneEventsForPersistence` to sanitize all diagnostic payload pointers to `nil` before disk serialization, guaranteeing zero sensitive context leakage into committed Git ledgers.

2. **Boundary Adapters & HTTP Service**:
   - Implemented non-blocking SQLite trace querying in `adapters/live/sqlite.go` using read-uncommitted and query-only pragmas (`PRAGMA query_only = ON;`, `PRAGMA busy_timeout = 5000;`, `PRAGMA read_uncommitted = ON;`).
   - Implemented `HydrateWorkspaceEvents` on `LiveService` to discover root and child sub-sessions across invocation trees and reconstruct rich chronological `EventFact` instances in memory with zero disk writes to `workspaces/`.
   - Updated `WorkspaceEventsHandler` to resolve dynamic hydration when local SQLite is available, degrading gracefully to committed ledger events when offline, and enforcing strict path traversal validation.

3. **Presentation & Web Dashboard**:
   - Defined `WorkspaceEventsResponse` and `DataSource` TypeScript types, extending `EventFact` with optional diagnostic fields.
   - Updated `api/client.ts` to parse dual-plane responses with backward-compatible normalization.
   - Enhanced `TraceWaterfall.tsx` to detect `is_hydrated`: rendering an informational offline banner ("Committed Git Ledger — Local Database Offline") and compact error timeline when unhydrated, and revealing interactive expandable payload drawers for user prompts, model reasoning, and tool execution parameters/results when hydrated.
   - Threaded hydration state through `TicketDeepDive.tsx` and updated `AnomalyDrawer.tsx`.

## Architectural Invariants
- FCIS compliance: pure domain core with zero I/O imports; all SQLite and HTTP logic confined to imperative shell adapters.
- Zero disk persistence: transient hydrated data remains 100% in-memory and is never persisted to workspace ledgers.
- Type safety: zero `any` or `interface{}` across Go and TypeScript.
- Design system: zero emojis across code, tests, and UI strings.

## Documentation References
- Contract: [CONTRACT.md](workspaces/AL-33/CONTRACT.md)
- Work Summary: [SUMMARY.md](workspaces/AL-33/SUMMARY.md)

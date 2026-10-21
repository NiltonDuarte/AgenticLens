# Ticket Summary: REV-1

## Scope & Objective
Comprehensive architecture, code, and design audit of all fallback chains across AgenticLens (Go backend, CLI, domain, adapters, and TypeScript frontend).

## Status
Completed (Audit & Decision Invariants Locked; Zero Code Patches Applied)

## Deliverables Generated
- `workspaces/REV-1/REVIEW_REPORT.md`: Comprehensive audit cataloging all 24 fallback cascades across the system, grading their resilience, detailing 5 critical blind spots, and establishing concrete architectural directives based on user decisions.
- `workspaces/REV-1/SUMMARY.md`: Summary log of the audit findings, architectural decisions, and downstream ticket definitions.

## Key Architectural Decisions & Directives Locked

1. **Date Resolution Precedence Mandate**:
   - `completed_at` must always be available and non-zero when telemetry is generated.
   - Telemetry generator (`agentic-lens generate`) must validate that `completed_at` is populated.
   - **Zero Fallback**: Domain aggregations and frontend charts must never fall back to `updated_at` or `started_at` for ticket completion. Fallback cascades in `GetLedgerCompletionDate`, `GetTicketEffectiveDate`, `formatters.ts`, and `progression.ts` are slated for complete removal.

2. **Session Resolver Tier 4 Deletion**:
   - Drop Tier 4 (`SELECT session_id FROM messages ORDER BY created_at DESC LIMIT 1`) entirely from `backend/internal/adapters/opencode/resolver.go`.
   - Fail fast immediately with `ErrSessionNotFound` when explicit session identifiers (nonce, session ID, ticket title/directory basename) are absent.
   - Eliminates race conditions and cross-ticket telemetry contamination in concurrent multi-agent runs.

3. **Frontend Ticket Type Masking**:
   - Remove `ledger.ticket_type ?? 'feature'` in `TicketDeepDive.tsx:221`.
   - Explicitly display `'unclassified'` with dedicated neutral badge styling when `ticket_type` is nil or missing.

4. **Workspaces Directory Resolution Order**:
   - Prioritize `.agenticlens-config.yaml` (`cfg.Workspaces.Dir`) over host filesystem container checks (`/app/workspaces`).

5. **Default Build Version Synchronization**:
   - Align frontend fallback version (`Navbar.tsx`) with backend domain constant `v0.0.0-dev`.

## Downstream Remediation Tickets (Prepared for Execution)
- `REV-FIX-1`: Enforce Mandatory `completed_at` in Telemetry & Dismantle Date Fallback Chains.
- `REV-FIX-2`: Drop Session Resolver Tier 4 & Fail Fast.
- `REV-FIX-3`: Eliminate UI Ticket Type Fallback to `'feature'`.
- `REV-FIX-4`: Fix Workspaces Container Directory Sniffing Precedence.
- `REV-FIX-5`: Harmonize Default Version Strings across Frontend and Backend.

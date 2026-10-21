# refactor(core): eliminate brittle fallback cascades and enforce deterministic contracts (REV-1)

## Summary of Changes

This pull request resolves the critical fallback cascades identified in `workspaces/REV-1/REVIEW_REPORT.md` across the Go backend and TypeScript web dashboard, replacing ad-hoc fallback chains with deterministic invariants, fail-fast error boundaries, and strict type safety:

1. **Mandatory Telemetry Completion Timestamp**:
   - `backend/cmd/agentic-lens/main.go`: `generateCmd` now enforces a valid, non-empty ISO-8601 UTC `completed_at` timestamp on every generated fact ledger.

2. **Dismantling Completion Date Fallback Cascades**:
   - `backend/internal/domain/aggregations.go`: `GetLedgerCompletionDate` and `GetTicketEffectiveDate` strictly evaluate `CompletedAt`. All fallbacks to `updated_at` or `started_at` are completely removed. Tickets lacking `completed_at` are treated strictly as uncompleted, preventing temporal distortion in time-window aggregations.
   - `web/src/domain/formatters.ts`: `getEffectiveCompletionDateMs` evaluates only `completed_at` and returns `null` if missing or unparseable.
   - `web/src/domain/progression.ts`: Progression trends allocate delivered code and invocation metrics strictly by `completed_at`.

3. **Session Resolver Tier 4 Deletion & Fail-Fast Sentinel**:
   - `backend/internal/adapters/opencode/resolver.go`: Deleted Priority Tier 4 database fallback (`SELECT ... ORDER BY time_created DESC LIMIT 1`). When Tiers 1-3 fail, `ResolveActiveRootSession` immediately returns `ErrSessionNotFound`, eliminating foreign session cross-contamination in concurrent multi-agent executions.

4. **Workspaces Directory Precedence**:
   - `backend/internal/adapters/repository.go`: `ResolveWorkspacesDir` prioritizes explicit CLI overrides and configured project directory (`cfg.Workspaces.Dir`) before checking `/app/workspaces`, preventing container paths from hijacking host project configurations.

5. **Elimination of Silent UI Ticket Classification**:
   - `web/src/views/TicketDeepDive.tsx`: Replaced `{ledger.ticket_type ?? 'feature'}` with `{ledger.ticket_type ?? 'unclassified'}`.
   - `web/src/index.css`: Added dedicated `.ticket-type-pill.unclassified` warning badge styling.

6. **Default Build Version Harmonization**:
   - Synchronized default version fallback to `v0.0.0-dev` across `domain.DefaultVersion`, `Makefile`, and `Navbar.tsx`.

---

## Artifact Pointers
- **Contract & Architecture Boundaries**: [CONTRACT.md](workspaces/REV-1/CONTRACT.md)
- **Comprehensive Audit**: [REVIEW_REPORT.md](workspaces/REV-1/REVIEW_REPORT.md)
- **Detailed Execution Summary**: [SUMMARY.md](workspaces/REV-1/SUMMARY.md)
- **Iteration 01 Plan**: [PLAN.md](workspaces/REV-1/iterations/01/PLAN.md)
- **Iteration 01 Review**: [REVIEW.md](workspaces/REV-1/iterations/01/REVIEW.md)

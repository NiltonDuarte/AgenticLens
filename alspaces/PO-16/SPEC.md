# PO-16: Reconcile ticket index and to-do planning records

## Problem

Merged AL-49 and AL-50 still appeared as Ready, and outstanding planning records disagreed on dependencies and current schema assumptions. Maintainers need an accurate backlog to select work without repeating delivered features.

## Scope and acceptance

- Remove confirmed completed entries while keeping their existing done specifications.
- Keep every remaining to-do specification indexed exactly once with valid paths, consistent statuses, and outstanding blockers.
- Move dependency-free AL-57 to Ready; synchronize AL-52/53/54 and AL-55 with completed AL-51.
- Distinguish partial AL-47 behavior and superseded AL-55 assumptions from completion. Preserve product goals and the shipped v4 runtime boundary.
- Record the audit, review evidence, PR, and current session telemetry under PO-16. Restore AL-49's original artifacts and ledger unchanged.

Documentation-only maintenance: no Go, TypeScript, tests, schemas, runtime scripts, or configuration changes. No new implementation tickets are created. Existing unfinished tickets remain subject to their development workflow; this audit does not approve a replacement architecture.

## Delivery

PR #74 contains the planning-record reconciliation. Per-ticket decisions and validation are in `iterations/01/REVIEW.md`. This task is complete planning maintenance and therefore does not add a Ready entry or a new development specification to `plan/`.

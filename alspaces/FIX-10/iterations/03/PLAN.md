# Iteration 03 Plan: Remove Schema Fallbacks and Normalize All Existing alspaces Ledgers

## Overview
Remove backward-compatibility fallbacks in Go and TypeScript, and normalize all historical ledgers in `alspaces/` to canonical schema v4 without invocation-level `user` or `parent_id`:
1. **Remove Go Hoisting & Invocations Boundary Fallbacks (`backend/internal/adapters/schemas.go`)**:
   - Remove `ParentID` and `User` fields from `InvocationFactDTO`.
   - Remove the hoisting loop in `AgenticFactLedgerDTO.ToDomain()` that lifted `User` and `ParentID` from child invocations.
   - Enforce in `ValidateRuntimeFactLedger` and `AgenticFactLedgerDTO.Validate()` that `invocations` must not contain legacy `parent_id` or `user` fields, ensuring pure canonical v4 ingestion.
2. **Remove TypeScript Boundary Fallbacks (`web/src/domain/types.ts`)**:
   - Remove optional `parent_id?: string | null` and `user?: string` from `InvocationFact` interface in TypeScript.
   - Update any frontend formatters/tests referencing `inv.user` or `inv.parent_id`.
3. **Normalize All Existing Ledgers in `alspaces/`**:
   - Upgrade and format all 71 `agentic_metrics.json` ledgers under `alspaces/`:
     - Hoist `user` to `session.user`.
     - Remove `parent_id` and `user` keys from all `invocations` arrays.
     - Normalize invocation IDs to `<session_id>-inv-<index>`.
   - Verify every ledger passes `ValidateRuntimeFactLedger` without diagnostics or skips.
4. **Test Suite & Verification**:
   - Update any backend tests in `adapters` and `domain` that pass legacy invocation `ParentID` or `User` through DTOs.
   - Run `make` (Go vet, race tests, TypeScript typecheck, ESLint, Vitest, multi-platform build).
   - Verify all workspaces in `alspaces/` scan cleanly via `agentic-lens --workspaces alspaces --list`.

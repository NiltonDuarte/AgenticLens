# Iteration 03 Review: Remove Schema Fallbacks and Normalize All Existing alspaces Ledgers

## Verdict: STATUS: APPROVED

Iteration 03 successfully eliminated legacy schema fallbacks for invocation-level `parent_id` and `user` across Go adapters and TypeScript domain types, normalized all 71 production workspace ledgers under `alspaces/` and testdata fixtures to canonical schema v4, updated all test fixtures, and verified zero diagnostics during complete workspace scanning and test execution.

---

## Axis 1: Spec & Contract Alignment

### 1.1 Pure Canonical Schema v4 Boundary Ingestion
- **Go Boundary DTOs (`backend/internal/adapters/schemas.go`)**:
  - Removed `ParentID` and `User` fields from `InvocationFactDTO`.
  - Removed the fallback hoisting loop in `AgenticFactLedgerDTO.ToDomain()` that previously lifted `ParentID` and `User` from child invocations.
  - Updated `ValidateRuntimeFactLedger` and `AgenticFactLedgerDTO.Validate()` to enforce that `session.user` is non-empty and stripped of leading/trailing whitespace, and that invocations contain only execution-level fields.
  - Because `json.NewDecoder` is configured with `DisallowUnknownFields()`, any invocation payload containing `parent_id` or `user` is strictly rejected at the boundary.
- **TypeScript Boundary Types (`web/src/domain/types.ts`)**:
  - Removed optional `parent_id?: string | null` and `user?: string` fields from the `InvocationFact` interface.
  - `SessionFact` maintains `user: string`, `parent_id?: string | null`, and `source?: string`.

### 1.2 Ledger Normalization Across All Workspaces
- Normalized all 71 workspace ledgers in `alspaces/*/agentic_metrics.json`:
  - Hoisted missing or invocation-level user to `session.user` (defaulting historical unannotated commits to repository author "Nilton Duarte").
  - Stripped `parent_id` and `user` keys from all `invocations` arrays.
  - Normalized invocation IDs to canonical `<session_id>-inv-<index>` format.
  - Verified every workspace ledger decodes cleanly through `ValidateRuntimeFactLedger`.
- Normalized all mock and parity ledgers under `testdata/workspaces/` and `testdata/parity_workspaces/`.

### 1.3 Test Fixture Updates
- Updated backend tests in `backend/internal/adapters/schemas_test.go`, `backend/cmd/agentic-lens/main_test.go`, `backend/internal/adapters/repository_test.go`, `backend/internal/adapters/backfill/engine_test.go`, and `backend/internal/adapters/http/server_test.go`.
- Updated frontend test fixtures in `web/src/test/fixtures/mockData.ts` and unit test assertions in `web/src/domain/progression.test.ts`.

---

## Axis 2: Engineering & Quality Standards

### 2.1 Functional Core, Imperative Shell (FCIS)
- Pure domain logic in `internal/domain` remains completely free of I/O, network, or clock operations.
- Boundary adapters in `internal/adapters` strictly validate and translate external DTOs into domain models.

### 2.2 Typing and Strictness
- Zero `any` or `interface{}` across Go backend and TypeScript frontend.
- Strict typecheck (`tsc --noEmit`) passes with zero errors.
- ESLint passes with zero warnings (`--max-warnings 0`).

### 2.3 Verification Results
- `go test -race ./...`: All backend packages pass with 100% success.
- `vitest run`: 32 test files passed, 734 tests passed.
- `agentic-lens --workspaces alspaces --list`: Successfully scanned 71 ticket workspaces with zero warnings or diagnostic errors.
- Multi-platform binaries compiled cleanly for `linux/amd64`, `linux/arm64`, `darwin/amd64`, and `darwin/arm64`.

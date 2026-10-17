# AgenticLens Technical Guidelines

Mandatory technical guidelines for developing AgenticLens using **Go** (backend, domain engine, CLI, and HTTP services) and **TypeScript** (frontend dashboard and UI presentation). All code must strictly adhere to these rules.

---

## 1. Architectural Boundaries and Layer Separation

Dependencies flow strictly inward across layers:

### Backend (Go)
1. **CLI / Service Layer (`cmd/agentic-lens/`)**:
   - Application entrypoint, CLI commands (`scan`, `metrics`, `summary-append`, `server`), HTTP/gRPC server wiring.
2. **Adapter Layer (`internal/adapters/`)**:
   - External boundary serialization (JSON encoding/decoding), filesystem repository scanner (`workspaces/` traversal), HTTP handlers, API clients.
3. **Domain Layer / Functional Core (`internal/domain/`)**:
   - Pure domain models, value objects, aggregation algorithms, health score derivations, constants, and typed errors.
   - **Zero External Dependencies**: Must NOT import outer layers (`internal/adapters`, `cmd/`) or network/disk libraries (`net/http`, `os`, `database/sql`).

### Frontend (TypeScript)
1. **Views & Components (`web/src/components/`, `web/src/views/`)**:
   - Interactive analytics dashboard, KPI widgets, comparative charts, filtering bars.
2. **API & Boundary Layer (`web/src/api/`, `web/src/adapters/`)**:
   - Network clients, JSON boundary validation (via Zod or strict type guards).
3. **Domain Types & Pure Formatters (`web/src/domain/`, `web/src/types/`)**:
   - Strictly typed data contracts mirroring the Go domain model, pure mathematical or date formatting utilities.

---

## 2. Functional Core, Imperative Shell (FCIS)

### Go Functional Core
- All mathematical rollups, context window health evaluations (HEALTHY, ELEVATED, BLOATED, CRITICAL), token formatting, turn velocity, and financial cost derivations must be pure functions in `internal/domain`.
- Pure functions receive immutable structs/slices and return computed models or explicit errors.
- No disk I/O, no network calls, no clocks (`time.Now()`), and no random number generators in the domain core. Inject timestamps or random seeds as function arguments.
- Defensively guard against division by zero (e.g. empty workspaces, zero turns, zero duration, zero tokens).

### Imperative Shell (Go & TypeScript)
- The shell handles filesystem traversal, directory scanning, JSON deserialization, HTTP routing, and UI rendering.
- Errors from the shell must be caught, wrapped with descriptive context (`fmt.Errorf("...: %w", err)`), and handled gracefully.

---

## 3. Go Design & Idioms

### Strong Typing & Value Semantics
- Target Go 1.23+.
- Use value semantics and immutable struct patterns for domain data entities.
- Zero `any` or `interface{}` across domain and adapters. Use concrete structs, generics, or type-safe sum types / interfaces.
- Define interfaces where they are consumed, keeping interfaces small (1-2 methods).
- Use typed constants / string enums (`type TimeWindow string`) rather than untyped raw string primitives.

### Error Handling & Validation
- Never discard errors (`_ = ...` is forbidden for I/O, parsing, or domain execution).
- Wrap errors with domain-relevant context: `fmt.Errorf("parsing fact ledger for ticket %s: %w", ticketID, err)`.
- Differentiate between non-fatal scan warnings (e.g. corrupted single JSON file) and fatal execution errors.

---

## 4. TypeScript Design & Idioms

### Strict Typing & No Any
- Target TypeScript 5.x+ in strict mode (`"strict": true`, `"noImplicitAny": true`, `"exactOptionalPropertyTypes": true`).
- The `any` type is strictly forbidden across all TypeScript code. Use `unknown` with runtime type narrowing or Zod schemas at boundaries.
- Discriminated unions for multi-state UI representation (Loading, Success, Empty, Error).

### UI Presentation Rules
- Pure presentation components separated from network fetching.
- Declarative state management and reactive chart updates.
- Zero emojis permitted in UI titles, tab labels, metric cards, tables, charts, tooltips, or console logs.
- Defensive presentation: handle missing or empty datasets gracefully with empty-state messages without uncaught runtime errors.

---

## 5. Quality Checks & Verification

Before submitting any iteration or PR, all quality gates must pass:

### Go Checks
```bash
# Verify compilation and format
go vet ./...
# Run test suite with race detector
go test -race ./...
# Build CLI binary
go build -o bin/agentic-lens ./cmd/agentic-lens
```

### TypeScript Checks
```bash
# Typecheck
npm run typecheck  # or npx tsc --noEmit
# Linting
npm run lint
# Unit tests
npm test
```

### Invariant Checks
- Verify zero `any` in Go and TypeScript.
- Verify zero emojis in code, UI, docstrings, and log messages.

---

## Summary Checklist

| Topic | Decision | Rule |
|---|---|---|
| **Backend Architecture** | Inward Dependencies | `internal/domain` has zero I/O imports. `internal/adapters` handles JSON and disk. `cmd/` handles CLI/server. |
| **Separation** | FCIS | Pure deterministic logic in domain; filesystem I/O and HTTP in shell. |
| **Go Standards** | Go 1.23+, Value Semantics | No `any`/`interface{}`, wrap errors with `%w`, zero unhandled errors. |
| **Frontend Architecture** | TypeScript Strict | `"strict": true`, zero `any`, Zod/type guards at boundaries. |
| **UI & Styling** | Modern Web / Zero Emojis | Pure components, graceful empty states, zero emojis. |
| **Testing** | Deterministic Pure Core | Table-driven tests in Go with zero mocks for domain. Vitest/Jest for TypeScript. |

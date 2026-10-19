# feat(cli): conventional commit ticket type classification in telemetry engine (AL-29)

## Overview
Refactors ticket type resolution across AgenticLens into a deterministic 4-tier hierarchy (CLI flag -> branch prefix -> configured project prefix -> unclassified fallback), completely banishing brittle git log commit grepping and eliminating silent fallbacks to `feature`.

All hardcoded repository prefixes (`PO`, `GEN`) have been purged from domain engine code. Prefix-to-ticket-type mappings are purely project-configurable via `.agenticlens/config.yaml` (`ticket.prefixes[].ticket_type`) and serve strictly as a tertiary backup when caller flags and git branches are absent. Unmatched tickets emit an explicit warning to stderr and persist `ticket_type: null` rather than fabricating inaccurate data. An optional `--strict-type` flag forces CLI failure with exit code 1 when a ticket type cannot be resolved. In addition, canonical conventional commit types (`feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`) and verbose aliases (`feature`, `bug`, `documentation`, `task`, `performance`, `tests`) are strictly validated and normalized at the domain boundary.

- Contract: [workspaces/AL-29/CONTRACT.md](workspaces/AL-29/CONTRACT.md)
- Telemetry: [workspaces/AL-29/SUMMARY.md](workspaces/AL-29/SUMMARY.md)

## Summary of Changes

### Pure Functional Core Layer (`backend/internal/domain/`)
- Added `ConventionalTicketType` string enum and canonical tokens (`feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`).
- Implemented `NormalizeTicketType` mapping verbose aliases and validating canonical tokens.
- Implemented `IsValidTicketType` predicate.
- Extended `TicketPrefix` struct with `TicketType string` (`json:"ticket_type,omitempty"` `yaml:"ticket_type,omitempty"`).
- Implemented purely data-driven `InferTicketTypeFromPrefix(ticketID string, prefixes []TicketPrefix)` with delimiter protection (`-` / `_`) and zero hardcoded prefix literals in engine code.
- Added helper method `(t TicketConfig) InferTicketType(ticketID string)`.

### Boundary Adapters Layer (`backend/internal/adapters/`)
- In `config/loader.go`, extended `rawPrefixItem` to decode `ticket_type` from YAML/JSON configurations.
- In `.agenticlens/config.yaml`, configured `ticket_type: chore` for prefix `GEN` and `ticket_type: docs` for prefix `PO`.
- In `git/inference.go`, completely removed `git log --all --grep` commit search heuristic and parameterized `InferTicketType(ticketID string, prefixes ...domain.TicketPrefix)` to check git branch category first, delegating to configured prefixes only when branch detection does not resolve.
- In `schemas_test.go`, verified schema serialization stability of `ticket_type: null`.

### CLI Application Layer (`backend/cmd/agentic-lens/`)
- Registered `--strict-type` flag in `runGenerate`.
- Ingested project configuration early and enforced deterministic 4-tier resolution:
  1. CLI flag `--ticket-type` / cached ledger facts (highest priority).
  2. Git branch prefix (`feat/*`, `fix/*`).
  3. Configured project prefix backup (`.agenticlens/config.yaml`).
  4. Fallback: stderr warning and `ticket_type: null` persistence (or stderr error and exit 1 under `--strict-type`).
- Added input validation for `--ticket-type` terminating with exit code 1 on unrecognized tokens.

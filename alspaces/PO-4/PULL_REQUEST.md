# PR: PO-4: CI/CD Pipelines, Multi-Version Schema Migration, Review Skill, Live Observability, YAML Config Bridge, and Rich Markdown Artifacts

## Summary
Defines the product specification, architectural contracts, and 10 vertical tracer-bullet Jira tickets across CI/CD quality gates, automated release pipelines with semantic versioning, universal schema migration for `agentic_metrics.json`, codebase and skill review via the provided `agentic-lens-review` skill, YAML configuration migration with an agent-backend bridge schema, non-locking live observability, and rich Markdown rendering for workspace artifacts.

## Initiatives Covered

1. **Continuous Integration & Semantic Versioning (AL-13)**:
   - Compile-time `-ldflags` injection for `Version`, `GitCommit`, and `BuildDate`.
   - CLI `agentic-lens version` and `/api/version` HTTP endpoint.
   - GitHub Actions CI workflow executing `make check`, `make test` (`-race`), and `make build` on PRs and `main`.

2. **Automated Release Workflows (AL-14)**:
   - Release Candidate workflow on pushes to `main` generating pre-releases with single-binary artifacts.
   - Stable Release workflow on `v*.*.*` tags with compiled binaries and a floating `latest` tag/pointer.

3. **Universal Multi-Version Schema Migration Engine (AL-15)**:
   - Pure domain stepwise migration pipeline (`v1 -> v2 -> ... -> vLatest`).
   - `agentic-lens migrate` subcommand for non-destructive, idempotent ledger upgrades.
   - Fixture test suites validating forward migration from all historical versions.

4. **Codebase & Skill Review, Refactoring with Provided Review Skill (AL-16)**:
   - Uses the externally provided `agentic-lens-review` skill to audit and refactor the Go backend and TypeScript frontend for FCIS adherence, typing rules (zero `any`), and zero emojis.

5. **YAML Configuration Migration & Agent-Backend Bridge (AL-21)**:
   - Migrates `.agenticlens/config.json` to human-readable `.agenticlens/config.yaml` with JSON fallback.
   - Defines concrete bridge schemas: case-insensitive ticket prefix matching, database pragmas, live thresholds, and agent-to-role mappings.

6. **Live Observability Platform (AL-17, AL-18, AL-19, AL-20)**:
   - **AL-17**: Non-locking SQLite database hook with read-only pragmas, multi-root session tree discovery, config-driven case-insensitive ticket guessing, and SSE stream.
   - **AL-18**: In-memory transient trace inspector for messages, reasoning thoughts, and prioritized failed tool calls.
   - **AL-19**: Pure domain real-time anomaly detection engine (tool error streaks, context saturation, token runaway).
   - **AL-20**: Reactive TypeScript Live Observability view with multi-root session tree, under-the-hood inspector, and failed tool call badges.

7. **Rich Formatted Markdown for Workspace Artifacts (AL-22)**:
   - Upgrades `ArtifactViewer.tsx` to render `CONTRACT.md` and `SUMMARY.md` as rich formatted Markdown with tables, task checkboxes, code syntax highlighting, and source/rich view toggles.

## Ticket Breakdown
- Ready: `AL-13`, `AL-15`, `AL-16`, `AL-21`, `AL-22`
- Pending: `AL-14` (blocked by AL-13), `AL-17` (blocked by AL-21), `AL-18` (blocked by AL-17), `AL-19` (blocked by AL-18), `AL-20` (blocked by AL-18, AL-19)

## Pointers
- Specification: `workspaces/PO-4/SPEC.md`
- Shared Roadmap: `workspaces/plan/TICKETS.md`
- Session Summary: `workspaces/PO-4/SUMMARY.md`

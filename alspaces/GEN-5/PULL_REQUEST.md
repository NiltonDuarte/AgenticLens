# GEN-5: Restructure README as Platform Showcase and Split Internals into ARCHITECTURE and CONTRIBUTING

## Summary
Restructures `README.md` to serve as an appealing platform advertisement and showcase for AgenticLens, highlighting its value in the autonomous agentic software development lifecycle (SDLC). Moves internal architecture, execution anomaly detection, backfill engine, and schema migration documentation to `ARCHITECTURE.md`. Moves developer installation, local development mode, Makefile target references, quality standards, and release packaging documentation to `CONTRIBUTING.md`.

## Changes
- **`README.md`**:
  - Re-anchored the document around the agentic SDLC value proposition, solving unmonitored agentic failure modes (runaway token spend, context saturation decay, tool retry death spirals, subagent ping-pong thrashing).
  - Highlighted core capabilities: real-time execution observability, deterministic anomaly detection, multi-agent SDLC analytics, automated pull request telemetry ledgers, universal schema migrations, and single binary distribution.
  - Added multi-agent SDLC workflow diagram showing the telemetry bus spanning Inception, Plan, Code, Test, and Review stages.
  - Streamlined Quick Start with one-liner binary downloads and standard make commands.
  - Added clean CLI overview and deep-dive documentation directory navigation.
- **`ARCHITECTURE.md`**:
  - Extracted Functional Core, Imperative Shell (FCIS) architecture overview and subsystem boundaries.
  - Detailed domain layer pure functions, aggregations, context health models, cost calculations, and migration framework.
  - Documented imperative adapter layers (SQLite extractor, Git analyzer, workspace traversal, HTTP server).
  - Documented execution anomaly detection rule engine (`TOOL_RETRY_STREAK`, `TOOL_HIGH_FAILURE_RATE`, `CONTEXT_SATURATION_SPIKE`, `SUBAGENT_REWORK_LOOP`) and CLI diagnostics.
  - Documented telemetry backfill engine and universal schema migration engine flags, invariants, and CLI usage.
- **`CONTRIBUTING.md`**:
  - Detailed prerequisites and developer environment setup.
  - Provided step-by-step instructions for installation (`make install`) and local development mode with hot reload.
  - Added comprehensive Makefile targets reference table.
  - Documented engineering standards and invariants (FCIS, zero `any`, data race freedom, zero emojis).
  - Documented release tiers, supported platform matrix, and local release validation commands.
- **Verification**:
  - Validated repository hygiene with `make check`.
  - Executed Go test suite with data race detection (`make test-backend`).
  - Confirmed strictly zero emojis across all markdown artifacts.

# AL-58 Iteration 01 Consolidated Implementation Plan

## Goal and frozen boundary

Implement complete OpenCode v1/v2 telemetry extraction, deterministic fixture coverage, isolated compatibility workflows, and documentation in accordance with [`../../CONTRACT.md`](../../CONTRACT.md). Evidence targets are supplied OpenCode v1.19.5 and npm `@opencode/cli` v2.2.0, resolved/pinned as registry `latest` on 2026-10-29. The user-wide v2 runtime is installed as `opencode2` without overwriting v1. Storage detection identifies a schema family, never an exact binary release.

The runtime evidence under `../../evidence/` contains populated isolated runs made with a deterministic local OpenAI-compatible provider. v1 uses `session`/`message`/`part`, with tool arguments named `filePath`; v2 uses `session_v2`/`session_message`, with tool arguments named `path`. Both record a successful read tool result, two assistant responses, equivalent totals of 46 input and 22 output tokens, and zero cost. Consult `RUNTIME_EVIDENCE.md` and both JSON fixtures for exact observed fields and provenance. The v2 migration transform is [the pinned v2.2.0 upstream migration source](https://github.com/anomalyco/opencode/blob/v2.2.0/packages/core/src/database/v1-migration.bun.ts).

This plan consolidates prior discovery and implementation notes. Earlier v1.19.3/v2.0.32 empty-database observations are historical evidence only and must not be relabeled or used as the target contract. The former branch (b) schema-only contingency and its implementation-readiness conclusion are superseded by populated runtime evidence. Full v2 extraction and the existing acceptance criteria remain required.

## Implementation sequence

### 1. Read-only storage-family detection and CLI wiring

1. In `backend/internal/adapters/opencode/`, add a focused schema inspection/classification boundary that reads SQLite metadata read-only. Define the required table sets from the pinned evidence and extraction/resolution queries: v1 requires `session`, `message`, and `part`; v2 requires `session_v2` and `session_message`. Validate required columns actually used by each query, not table names alone.
2. Automatically select exactly one complete family and pass the same selection to session resolution and extraction. Do not add a normal OpenCode version override flag. Preserve `--db-path`, its current default behavior, and existing session/nonce/ticket/title selection semantics.
3. Return typed, contextual errors for missing, partial, mixed/conflicting, malformed, or unknown layouts. Include observed/missing tables and columns when available. Never silently fall back, report empty success for unsupported schemas, guess from installed binaries, or migrate/repair source state.
4. Keep `client.go` opening the source read-only. Confirm all inspection and extraction queries remain read-only. Do not infer a separate v2 default path absent runtime evidence.

### 2. Full extraction into shared facts

1. Preserve the existing v1 path and normalized output. Add a v2-specific adapter reader for `session_v2` and ordered `session_message` records using observed relationship/order fields.
2. Decode evidence-backed v2 user, assistant, and idle message variants, typed text/tool content, timing, model, token/cost, finish state, and session relationships when represented. Convert these to existing `domain.InvocationFact`, `domain.ToolFact`, and `domain.EventFact` contracts.
3. Map v1 `filePath` and v2 `path` source arguments to the existing normalized tool representation. Preserve source IDs and deterministic event order; allow source payload and ID differences across versions.
4. Handle absent optional values deliberately. Malformed required payloads, unknown required variants, or missing required fields must produce record/session-context errors rather than silent drops. Do not claim behavior for unobserved historical variants.
5. Keep SQL rows, decoding structs, schema branches, and I/O out of `backend/internal/domain`; do not add domain version branches or broaden telemetry contracts without a demonstrated need.

### 3. Fixtures and deterministic adapter/CLI tests

1. Build deterministic SQLite fixtures from the retained runtime evidence (or an equivalent harness that exercises the real query/decoder path). Preserve the evidence files unchanged; do not fabricate additional record shapes or edit fixtures to fit implementation assumptions.
2. Cover v1 session/message/part and v2 session_v2/session_message extraction, root/session resolution, chronological events, two assistant responses, successful read invocation/result, model identity, observed timestamps, token/cost facts, and normalized domain outputs.
3. Assert semantic parity for shared facts: 46 input tokens, 22 output tokens, zero cost, and one completed read call/result, while respecting different source IDs and encodings.
4. Add adapter and CLI cases for unique-family auto-selection; missing and partial table/column sets; mixed/conflicting families; malformed JSON/variants; missing required fields; unknown schema; empty sessions; SQL/read failures; stable ordering; and read-only behavior. Assert failures are actionable and never become silent fallback/empty success.
5. Retain current v1 regression cases and prove fresh v1-only and v2-only databases work without relying on the other family's tables. No OpenCode binary is required for normal unit tests.

### 4. Isolated runtime workflows

1. Add independently callable v1-only, v2-only, and combined compatibility entry points in the root `Makefile` and/or focused scripts. The combined workflow runs both isolated variants plus the deterministic suite; each individual target can run independently.
2. Runtime setup must use pinned or explicitly configured binaries and independent temporary `HOME`, XDG data/config/state roots, project roots, and absolute per-version `OPENCODE_DB`. Use the deterministic local provider with no credentials. Clean temporary state on success and failure. Never reuse or inspect a developer's ordinary OpenCode DB.
3. Runtime tests may be opt-in or provisioned; document prerequisites. Ordinary fixture/unit tests remain deterministic and offline without installed binaries.

### 5. Documentation

1. Update `ARCHITECTURE.md` and relevant `docs/SETUP_MANUAL.md` descriptions that currently couple extraction to v1 tables. Explain supported evidence versions, family inference, required schema validation, automatic selection and rejection behavior, read-only access, `--db-path`, and unchanged default path behavior.
2. Document v1-only/v2-only/combined workflow commands, isolated state roots, deterministic fixture versus runtime integration evidence, and evidence provenance.
3. State that v2.2.0 was registry `latest` on 2026-10-29 and pinned for this evidence; do not imply that latest is timeless or that schema detection proves an exact binary version.
4. Explicitly disclose that nonzero native pricing, imported history, partial migration, duplicate handling, and releases outside v1.19.5/v2.2.0 were not verified. These limits do not downgrade full extraction of the pinned populated evidence.

## Expected change areas

- `backend/internal/adapters/opencode/`: schema inventory/family selection, read-only query routing, v1/v2 extraction and session resolution, focused fixture helpers and adapter tests.
- `backend/cmd/agentic-lens/`: automatic adapter selection and actionable CLI errors; preserve `--db-path` and existing options. Update CLI tests.
- `backend/internal/domain/` and configuration: only narrow typed value additions proven necessary; no I/O, SQLite, runtime payload, or version-specific aggregation.
- Root `Makefile` and focused runtime scripts: distinct v1, v2, combined entry points with enforced isolation.
- `ARCHITECTURE.md` and `docs/SETUP_MANUAL.md`: supported evidence window, behavior, workflow, and limitations.
- Retained evidence references: `../../evidence/RUNTIME_EVIDENCE.md`, `v1-1.19.5.json`, `v2-2.2.0.json`, provider configs, and `mock_openai_server.py`.

## Verification plan (for a separately authorized implementation)

- Run focused adapter/CLI tests and the normal deterministic test suite; verify pinned evidence fixtures map to equivalent normalized outputs.
- Run the isolated v1-only and v2-only workflows independently, then the combined workflow. Confirm each uses its isolated roots and absolute DB path, and verify source databases remain unchanged/read-only.
- Run repository checks/build (`make check`, `make test`, `make build`, or project-equivalent aggregate target) after implementation. Ordinary tests must not require OpenCode installations.
- Review errors for every unsupported/partial/mixed/malformed case and ensure no fallback or empty-success path.
- Review the diff against `../../CONTRACT.md`, FCIS and zero-`any` constraints, no emojis, path behavior, and evidence limitations. No test suite is requested or run as part of this planning consolidation.

## Handoff

User confirmation is still required before implementation.

# AL-58 Frozen Technical Contract

## Objective and supported evidence window

Add concurrent OpenCode v1/v2 telemetry extraction while preserving the existing normalized AgenticLens fact contracts and v1 behavior. The supplied v1 target is OpenCode **1.19.5**. The v2 target is npm `@opencode/cli` **2.2.0**, resolved as the registry `latest` release on **2026-10-29** and pinned for this work. The persistent v2 executable is installed as `opencode2`; the supplied v1 executable remains intact and was not overwritten. These are evidence targets, not a promise that every release has identical storage.

## Evidence and limitations

Runtime evidence is retained in `evidence/RUNTIME_EVIDENCE.md`, `evidence/v1-1.19.5.json`, `evidence/v2-2.2.0.json`, the provider configurations, and the deterministic local provider server. Both isolated runtimes used a local OpenAI-compatible provider and produced populated records, two assistant responses, a successful `read` tool invocation/result, 46 aggregate input tokens, 22 output tokens, and zero cost. v1 persisted `session`/`message`/`part`; its successful tool input uses `filePath`. v2 persisted `session_v2`/`session_message`; its successful tool input uses `path`. The v2 migration transform is documented in the pinned upstream source: [OpenCode v2.2.0 v1-migration.bun.ts](https://github.com/anomalyco/opencode/blob/v2.2.0/packages/core/src/database/v1-migration.bun.ts).

Evidence does not establish nonzero native-provider pricing, imported-history extraction after migration, partial migration states, duplicate handling, or behavior outside the pinned releases. Runtime-created records establish the observed payload shapes; they do not establish every optional or historical record variant. Schema-family detection identifies a storage family, not the exact binary release that created a database. Keep these limitations visible in implementation documentation and tests without reducing the full extraction acceptance below.

## Frozen architectural boundaries

1. Keep SQLite inspection, storage-family inference, SQL, path resolution, payload decoding, and compatibility diagnostics in the imperative shell, principally `backend/internal/adapters/opencode` and CLI wiring.
2. Normalize both storage families into the existing domain fact contracts. Domain calculations remain pure and version-agnostic. Add only narrow typed provenance values if a concrete traceability requirement needs them; never expose SQL rows or runtime-specific payloads to domain aggregation.
3. Open source databases read-only. Never create, migrate, repair, rename, or write OpenCode-owned tables.
4. Preserve `--db-path` as the explicit database-path override and preserve existing default path behavior. OpenCode provides no normal version override flag for this selection. Do not add one. Do not assume v2 requires a different default directory.
5. Infer and automatically select exactly one supported storage family based on its complete required table and column set. Reject missing, partial, mixed, conflicting, malformed, and unknown schemas with actionable diagnostics. Never silently fall back, return empty success for an unsupported family, choose from installed binary/version guesses, or claim an exact release from schema metadata.
6. Keep deterministic unit/boundary tests independent of installed OpenCode binaries. Any runtime workflow must isolate each version's `HOME`, XDG data/config/state roots, project, and absolute `OPENCODE_DB`; never point it at a developer's normal database.
7. Maintain Go 1.23+ FCIS boundaries, strict TypeScript, zero `any`/`interface{}`, and zero emojis. Avoid frontend/domain changes unless implementation demonstrates a specific need.

## Acceptance boundaries

- Existing v1 extraction keeps equivalent normalized output for the v1.19.5 evidence records.
- v2 extraction fully consumes the evidence-backed v2.2.0 session/message representation, including session resolution, chronology, user/assistant and idle variants, text, successful tool invocation/result, timing, model, token and cost fields where represented, and session relationships supported by evidence.
- v1 `filePath` and v2 `path` tool inputs map correctly into the normalized representation without conflating their source payloads.
- Exactly one complete supported family is selected. Required tables and every column required by actual resolution/extraction queries are validated. Missing/partial/mixed/conflicting/malformed/unknown layouts fail with contextual diagnostics identifying observed and missing structures; no silent fallback or migration occurs.
- Source DB access remains read-only; `--db-path` remains supported; no normal OpenCode version override flag is introduced.
- Runtime-produced fixtures assert semantic parity for shared facts: two assistant responses, one successful read call/result, 46 aggregate input tokens, 22 output tokens, and zero persisted cost. Source IDs and payload encodings may differ.
- Provide isolated v1-only, v2-only, and combined compatibility workflows. Ordinary deterministic unit tests need no installed OpenCode runtime.
- Documentation identifies the evidence versions, schema-family inference behavior, path behavior, isolated commands, fixture/runtime distinction, and known evidence limitations. Do not claim compatibility beyond the pinned evidence window without additional runtime evidence.
- No OpenCode DB migration/repair, broad telemetry-contract rewrite, acceptance downgrade, or branch (b) schema-only alternative.

## Out of scope

- Converting, repairing, backfilling, or writing to OpenCode databases.
- Claiming imported-history parity, complete migration, duplicate policy, nonzero native pricing, or support for unverified releases.
- Managing end-user OpenCode installation/configuration or rewriting user config.
- Unrelated AgenticLens domain, frontend, or telemetry-ledger redesign.

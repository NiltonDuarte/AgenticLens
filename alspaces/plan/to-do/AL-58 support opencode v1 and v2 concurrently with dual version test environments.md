h3. Title: AL-58: Support OpenCode v1 and v2 Concurrently with Dual-Version Test Environments

*TARGET*: AgenticLens (OpenCode | domain | adapters | CLI/config | tests | documentation)

*CONTEXT*

OpenCode v2 is released, while AgenticLens currently supports the v1 runtime/data environment. Users need both versions supported during migration. Developers also need both versions runnable in the current environment so they can exercise compatibility and regression tests without state from one version contaminating the other. Exact v2 schema and storage differences must be verified from the released runtime as part of implementation discovery.

*GOAL*

Add concurrent OpenCode v1 and v2 compatibility while preserving v1 behavior, and provide an isolated, repeatable local workflow for running each version's compatibility tests independently or together.

*VERTICAL SLICE SCOPE*

* OpenCode boundary adapters: identify source/runtime version and handle version-specific configuration, storage, and schema differences; normalize both versions into existing domain telemetry contracts.
* Domain: keep calculations version-agnostic and pure; introduce only value-level distinctions needed to represent normalized facts or source version.
* CLI/configuration: allow explicit selection where discovery is ambiguous and surface actionable errors for unsupported or malformed source data.
* Test workflow: provide isolated v1 and v2 environments or fixtures, dedicated test entry points, and a combined compatibility run. Ensure each version has separate configuration and state. Unit tests must not require installed runtime binaries.
* Presentation/diagnostics: expose source version where it aids traceability and explain supported versions and test setup in repository documentation.
* Tests: preserve v1 regression coverage; add representative v2 compatibility and cross-version parity coverage. Exercise missing runtime, ambiguous selection, malformed input, and state isolation.

*ACCEPTANCE CRITERIA*

# Existing v1 fixtures and representative v1 data continue to ingest with equivalent normalized telemetry.
# Representative v2 data ingests into the same domain contracts, with schema differences handled explicitly and verified against released v2.
# Users can independently select v1 or v2 sources; unsupported, malformed, or ambiguous inputs produce actionable diagnostics rather than silent fallback.
# A documented repeatable workflow runs v1 and v2 compatibility suites independently and together in the current development environment.
# Each version uses isolated runtime configuration/database state; running one suite does not overwrite or contaminate the other.
# Unit tests remain deterministic and binary-independent; integration tests exercise both versions using isolated runtimes or versioned fixtures.
# Domain logic remains pure and version-specific I/O remains in adapters; Go and TypeScript stay strictly typed with zero `interface{}` / `any`.
# No emojis are added to code, comments, documentation, commands, or output.

*DEFINITION OF DONE*

Implementation and tests are reviewed; v1 regression and v2 compatibility suites pass; both per-version and combined test commands are documented and verified; configuration/state isolation and migration limitations are documented.

*Blocked by*: None

*Blocks*: None

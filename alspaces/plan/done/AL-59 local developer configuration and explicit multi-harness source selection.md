h3. Title: AL-59: Local Developer Configuration and Explicit Multi-Harness Source Selection

*TARGET*: AgenticLens (config | adapters | cli | documentation | tests)

*CONTEXT*

The shared `.agenticlens-config.yaml` currently contains an OpenCode-specific SQLite database path alongside project-wide settings. This path is machine-specific, and Codex users have no equivalent configuration entry. Developers with both Codex and OpenCode installed need predictable source selection without editing shared configuration or accidentally reading a different database when the selected source is unavailable.

*GOAL*

Separate local machine-specific telemetry source settings from shared project configuration. Support independently configured Codex and OpenCode database paths, a local default source, and explicit command-level selection with documented precedence and safe failure behavior.

*VERTICAL SLICE SCOPE*

* Configuration: Add an optional ignored local developer config and tracked sanitized example. Define per-harness source paths and default source. Preserve current shared config compatibility and specify precedence as CLI override, local value, shared value, then built-in default where applicable.
* Adapters: Load and validate configured source locations. Keep harness-specific database access at the adapter boundary. If a selected source is missing or invalid, return a clear error without silently selecting another configured source.
* CLI: Allow telemetry commands to accept an explicit source selector and use the configured local default when omitted. Support a checkout configured for either or both harnesses.
* Documentation: Explain local setup, source selection, precedence, ignored local state, and missing-source behavior. Do not put credentials in config.
* Tests: Cover precedence, Codex-only/OpenCode-only/both-source configurations, invalid or absent selected databases, and compatibility with existing OpenCode configuration.

*ACCEPTANCE CRITERIA*

* Codex and OpenCode database paths can be configured independently in a local-only file.
* A configured default source is used when no CLI source is supplied; an explicit CLI source takes precedence.
* Explicitly selecting an unavailable source produces an actionable error and does not fall back silently.
* Existing shared OpenCode configuration remains usable during migration.
* The real local config is ignored by version control and a sanitized example is tracked.
* Tests verify source configuration, precedence, failure paths, and backward compatibility.
* Filesystem and database I/O remain in adapters; domain calculations remain pure.
* Go and TypeScript remain strictly typed with no `any` or `interface{}`; no emojis are added.

*DEFINITION OF DONE*

* Implementation and tests are delivered with setup documentation.
* Quality gates pass, including the repository's standard `make` checks.
* Ticket artifacts and telemetry are recorded according to repository workflow.

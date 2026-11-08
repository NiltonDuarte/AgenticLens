# GEN-22: Configurable user attribution

Status: Proposed; awaiting explicit scope approval before implementation.

## Request and interpretation requiring approval

The user requested: “Add user tag, both shared config and from cli”. This proposal interprets a user tag as the human attribution label already represented by SessionFact.User and displayed by the dashboard as User. It is an explicit attribution override, not an additional arbitrary metadata field. Confirm this interpretation before implementation; an additive tag would require a different persisted-data contract and a scope revision.

## Proposed behavior

- Support optional top-level user_tag in the shared .agenticlens-config.yaml and optional --user-tag on generate.
- Resolve the effective attribution override from a nonblank CLI value, then a nonblank shared configuration value. Trim surrounding whitespace; blank values behave as absent. Preserve internal spaces and case.
- Apply an effective override to each successfully refreshed session in this generation, including refreshed children and either supported harness. Persist it in the existing session user field. It takes precedence over extracted, cached and Git-derived attribution for those refreshed owners.
- When neither input supplies a tag, retain the existing source-specific user-resolution behavior unchanged, including unknown handling.
- Preserve every session outside the successful refresh set with its recorded attribution. Never relabel unavailable historical owners, other developers' unavailable owners or unrelated harness history merely because a config value changed.
- Automatic capture publication through generate receives the shared value too. A manual CLI override is invocation-scoped and does not modify shared configuration or capture registrations.
- Reads and optional hydration continue exposing stored attribution without requiring the originating source. Existing API session/thread user projections and the existing dashboard label display the result without a new schema or UI control.
- Backfill, migration, live-source attribution, local config overrides and new dashboard tag filtering are outside this initial scope. Existing recorded attribution must continue surviving backfill; this work does not authorize historical relabeling.

## Compatibility and architecture

The field is optional. Existing configurations and canonical schema 6 ledgers remain supported without migration or schema-version changes. Generated configuration preserves an explicitly configured value and documents the optional setting without embedding a developer-specific default. The writer must correctly quote YAML-sensitive labels.

Domain configuration remains typed and pure. CLI/configuration/persistence adapters own external inputs and I/O. Preserve the portable ledger contract in TECHNICAL_GUIDELINES.md: only successfully complete refreshed owner snapshots replace stored owners; all other owners survive. Preserve IDs, source labels, hierarchy, invocation multiplicity, derived totals and failure reporting. Use no any types and no emojis.

## Acceptance

Demonstrate shared-only attribution, CLI precedence and no-setting compatibility through actual generation paths for Codex and OpenCode. Include the registered capture publication path, refreshed child attribution, preserved unavailable owners and unrelated harness owners, save/reload visibility through existing API projections and idempotent repeat generation. Verify config load/write round trip with a non-default YAML-sensitive value and optional omission. Use isolated fixtures, never live ticket artifacts or private session stores as tests.

Implementation, tests and independent review follow approval. The parent owns integration, telemetry, commits and PR publication.

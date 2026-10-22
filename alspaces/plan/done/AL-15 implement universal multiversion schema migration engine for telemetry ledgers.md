h3. Title: AL-15: Implement Universal Multi-Version Schema Migration Engine for Telemetry Ledgers

*TARGET*: AgenticLens (domain | adapters | cli | tests)

*CONTEXT*

As AgenticLens evolves, the schema of `agentic_metrics.json` changes across iterations (e.g., from unversioned v1 without events, to v2 with chronological events and migration metadata, to future extensions). Users have existing workspaces with ledgers from older versions that become incompatible or incomplete. The platform requires a universal, stepwise schema migration engine that can upgrade `agentic_metrics.json` from any previous historical schema version up to the latest active version deterministically and idempotently.

*GOAL*

Deliver a pure domain stepwise migration pipeline and an adapter backfill command in Go capable of detecting any historical schema version in `agentic_metrics.json`, executing sequential version migrations (`v1 -> v2 -> ... -> vLatest`), validating schema invariants at each step, and atomically persisting upgraded files with dry-run support.

*VERTICAL SLICE SCOPE*

* Domain Core:
  - Create `internal/domain/migration/` defining the `MigrationStep` interface and migration registry.
  - Implement sequential migration chain: `Migrate(input []byte, targetVersion int) (AgenticFactLedger, error)`.
  - Step implementations:
    - Step `v1_to_v2`: Upgrades unversioned or v1 ledgers by adding missing `schema_version: 2`, initial `migration` metadata, and initializing empty `events` list.
    - Future step hooks (`v2_to_v3`, etc.) adhering to the uniform step interface.
  - Invariant validator: ensures non-destructive field transformation, preserving all custom and git metrics across migrations.
* Boundary Adapter & CLI:
  - Extend `internal/adapters/backfill/engine.go` with migration runner discovering all workspace ledgers.
  - Expose `agentic-lens migrate` subcommand (or `agentic-lens backfill --migrate-schema`):
    - `--workspaces <dir>`: Target workspaces directory.
    - `--target-version <int>`: Target schema version (defaults to `CurrentSchemaVersion`).
    - `--dry-run`: Previews schema transformations and field additions without disk mutation.
    - `--backup`: Creates `.bak` copies before overwriting ledgers.
* Quality:
  - Table-driven unit tests in Go using fixtures for unversioned v1 ledgers, partial v2 ledgers, and malformed ledgers.
  - Invariant test asserting that re-running migration on an already up-to-date ledger produces zero mutations (idempotence).
  - Race condition tests with `go test -race ./...`.

*ACCEPTANCE CRITERIA*

* Running `agentic-lens migrate --workspaces workspaces/` upgrades all historical `agentic_metrics.json` files to `CurrentSchemaVersion`.
* Legacy unversioned files and v1 files migrate cleanly to v2 without data loss or field corruption.
* Already upgraded ledgers remain unchanged when migration is executed repeatedly.
* Dry-run mode (`--dry-run`) reports migration diffs and version upgrades without modifying files on disk.
* Corrupted JSON files trigger descriptive error logs while allowing other workspaces to continue migrating.
* All Go tests and static checks pass cleanly with zero `any`.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`go vet ./...`, `go test -race ./...`).
* Verified end-to-end migration across test fixture workspaces.
* Telemetry tested with `agentic-lens` CLI.

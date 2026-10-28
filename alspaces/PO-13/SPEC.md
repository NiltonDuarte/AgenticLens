# Feature Specification: Concurrent OpenCode v1 and v2 Support

## Problem Statement

OpenCode has released v2 while AgenticLens currently targets the v1 runtime and data environment. Users need to migrate support without losing v1 compatibility, and developers need both versions available in the current environment to run comparable test suites during the transition. Without explicit version isolation, configuration and local runtime state can be mixed, making regressions and compatibility gaps difficult to identify.

## User Stories

- As an AgenticLens user, I want the platform to ingest and present telemetry from OpenCode v1 and v2, so that I can continue analyzing existing workspaces while adopting v2.
- As a developer, I want to run v1 and v2 OpenCode environments and their tests side by side, so that I can verify compatibility and migration behavior in one development environment.
- As a maintainer, I want the selected OpenCode version and any unsupported version to be explicit, so that failures are diagnosable and data is not silently misinterpreted.

## Functional Scope

### Runtime and data compatibility

- Discover and identify OpenCode v1 and v2 runtime/data sources without confusing their version-specific schemas or locations.
- Preserve v1 ingestion and behavior while adding v2 ingestion, including version-specific schema interpretation and normalization into the existing AgenticLens domain contracts.
- Support explicit version selection where automatic detection is ambiguous, and report unsupported or malformed runtime data with actionable diagnostics.
- Keep version-specific filesystem/database access in adapters; domain calculations remain pure and consume normalized values.

### Local development and test environments

- Provide a documented, reproducible way to provision or select both OpenCode versions in the current development environment.
- Allow v1 and v2 integration tests to run independently and in a combined compatibility workflow without overwriting each other's configuration, databases, or fixtures.
- Make the version under test visible in commands and test output, and fail clearly when a requested runtime is unavailable.
- Keep unit tests deterministic and independent of installed OpenCode binaries; use real versioned fixtures or isolated integration environments for compatibility tests.

### Presentation and operations

- Expose the detected/source OpenCode version in relevant diagnostic or trace metadata when useful for understanding ingestion behavior.
- Document supported versions, setup, test commands, and any known compatibility limits.

## Acceptance Criteria

1. Existing v1 fixtures and representative v1 runtime data continue to ingest with equivalent normalized telemetry.
2. Representative v2 data is ingested into the same domain-level telemetry contracts, with version-specific differences handled explicitly.
3. A user can configure or select v1 and v2 sources independently, and ambiguous or unsupported versions produce actionable errors rather than silent fallback.
4. The repository provides a repeatable workflow to run v1 and v2 compatibility tests independently and together in the current environment.
5. Running either test workflow does not mutate or reuse the other version's database/configuration state unintentionally.
6. Compatibility tests cover both versions and a regression suite demonstrates v1 behavior is preserved.
7. Version-specific I/O stays in adapters; domain logic has no filesystem, process, network, or clock dependencies.
8. Go and TypeScript remain strictly typed, with no `interface{}` or `any`; no emojis are introduced.

## Out of Scope

- Removing v1 support or forcing an immediate migration of existing user data.
- Rewriting OpenCode itself or managing production OpenCode upgrades.
- Changing telemetry meaning unrelated to compatibility between the two versions.

## Risks and Open Decisions for Development

- Confirm exact v2 database/configuration locations and schema changes against the installed/released v2 runtime before freezing the technical contract.
- Prefer separate isolated runtime state for each version; the implementation may use local installations, containers, or versioned fixtures according to repository and environment constraints, but must offer one documented workflow that works in the current environment.

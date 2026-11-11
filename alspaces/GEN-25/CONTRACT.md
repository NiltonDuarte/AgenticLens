# GEN-25: Support macOS and Linux only

AgenticLens will support Linux and macOS on amd64 and arm64. Windows support remains in the build, native manager, capture integrations and setup tooling even though its CI gate was removed. This ticket removes those active implementation and distribution paths so the advertised platform set matches the maintained product.

## Observable acceptance

- Default contributor builds and release packaging produce manager and engine binaries for exactly linux/amd64, linux/arm64, darwin/amd64 and darwin/arm64. Release archives and checksum generation work with tar.gz output only. Explicit unsupported platform requests fail clearly before downloads or installation changes.
- The manager, native validation and setup downloader no longer accept Windows artifacts or platform selections. Remove the PowerShell bootstrap, Windows executable naming, PE admission and Windows-specific implementation files, process handling, shell quoting and hook/plugin transport branches. Build constraints must express the maintained platform boundary without adding another operating system.
- Linux and macOS retain exact project pin selection, separate manager/engine identity, checksum/protocol validation, atomic installation, ownership preservation, cancellation, bounded handoff and capture publication. Supported hook commands preserve literal argument values and project context.
- Active user/contributor documentation, setup/operator skills and generated installation guidance consistently describe macOS/Linux support. Remove obsolete Windows execution tests and Windows-only assertions; preserve distinct supported-platform and data-portability assertions in mixed tests.

## Approved iteration 02 boundary

The user's subsequent explicit instruction establishes that Windows has never been installed and no Windows compatibility is needed. Remove `commandWindows` inspection and historical multi-platform hook recognition/fixtures/tests. Only the current supported hook template is owned automatically; unknown or edited hooks remain protected. Remove obsolete Windows-only test environment scaffolding. Retain negative Windows rejection tests, historical audit records and portable-data/security assertions: none provide Windows runtime or upgrade compatibility. See [iteration 02](iterations/02/PLAN.md) for the superseding decision and verification.

## Preserved boundaries

Historical tickets, changelogs and previously published releases remain historical records. Existing portable ledgers, recorded source paths and cross-machine facts remain usable under the [portable ledger contract](../../TECHNICAL_GUIDELINES.md#portable-ledger-contract); dropping host support does not authorize ledger/schema migration, history deletion or a new authenticity policy. Incidental references such as context windows are outside the removal. Existing user-owned integration configuration must remain protected when inspected or refreshed.

No new platforms, runtime architecture, capture capabilities or CI certification are introduced. The [technical guidelines](../../TECHNICAL_GUIDELINES.md) continue to govern implementation.

Status: approved by the user message “Approved” on 2026-11-11.

# FIX-20: Stop source reads when callers cancel

## Roadmap ticket

[FIX-20](../../../alspaces/plan/to-do/FIX-20%20stop%20source%20reads%20when%20callers%20cancel.md)

## Why

Closing a live or historical request can leave its native source read running. A live trace read can hold the shared cache lock while it runs. Backfill can also keep reading after the workspace deadline expires.

## What changes

Propagate caller cancellation to live OpenCode reads, historical OpenCode and Codex hydration, and explicit source-backed backfill. Stop interrupted native work at query and read boundaries, release resources, and report cancellation or incomplete coverage. When backfill is cancelled before publication starts for a workspace, keep that workspace's ledger and summary intact. Earlier committed workspaces retain their updates.

The existing timeout policy, source selection and supported schema families remain in scope as compatibility constraints. Broader source-read guarantees belong to REV-7.

## Capabilities

- Modified `live-source-reporting` for live reads and historical hydration.
- Added `backfill-source-cancellation` for source-backed backfill and its publication boundary.
- Added `stable-ledger-order` for consistent raw ledger array ordering across generation, backfill and migration.

## Impact

Live clients can recover after an interrupted read without accepting a cancelled snapshot. Historical clients retain usable evidence with explicit coverage. Backfill preserves recorded owner facts and ordinary partial-refresh behavior on source failures.

The user also requested optional capture observation in generated Markdown. Add an opt-in setting for the diagnostic block while preserving the metrics tables, ledger and local capture diagnostics.

Use stable array ordering when publishing the raw metrics ledger so that changing which sessions are refreshed does not move otherwise unchanged records.

Preserve the ledger timestamps and file contents when a publication produces the same recorded facts.

# AL-97: Warn about staged pending capture snapshots

## Roadmap ticket

[AL-97 · Warn about staged pending capture snapshots](../../../alspaces/plan/to-do/AL-97%20warn%20about%20staged%20pending%20capture%20snapshots.md)

## Why

Developers can intentionally version pending capture snapshots, but a broad Git add can include them accidentally. The staged copy can outlive a worker update or removal, so the working copy alone cannot tell developers what they are about to commit.

## What changes

- Offer the existing AgenticLens pre-commit advisory as an explicit setup choice. Honor Git's effective hook location and preserve existing hook ownership.
- Warn when a commit introduces or updates a pending capture snapshot in the configured directory. Distinguish captures awaiting a ticket from bound captures awaiting a publication destination and give a next step for each.
- Honor a confirmed local discard for the exact capture. Keep the advisory read-only and allow the commit to proceed in every AgenticLens outcome.

## Capabilities

- `staged-pending-capture-advisory` adds a staged-snapshot warning alongside the existing capture-health advisory.

## Impact

Developers can review the staged snapshot, unstage it, or address its pending state before committing. The change affects setup guidance and the native capture check. Pending snapshots remain portable and available for intentional versioning under the [AL-96 publication contract](../al-96-verify-automatic-ticket-metrics-capture-end-to-end-for-v3/specs/automatic-ticket-publication/spec.md).

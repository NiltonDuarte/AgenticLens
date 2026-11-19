# AL-96: Verify first automatic ticket publication for v3

## Roadmap ticket

[AL-96 — Verify automatic ticket metrics capture end to end for v3](../../../alspaces/plan/to-do/AL-96%20verify%20automatic%20ticket%20metrics%20capture%20end%20to%20end%20for%20v3.md)

## Why

The automatic worker can collect bound sessions but resolves its output by looking for an existing ledger. [PO-26](../po-26-define-the-v3-automatic-capture-release-gate/delivery.md#verification-and-limits) reproduced the resulting dead end on a fresh ticket. The first report should appear through the same installed lifecycle as later updates.

## What changes

- Let an operator explicitly select the first destination for a ticket. After native binding, the worker publishes to the selected directory using the existing publication transaction.
- Verify the packaged candidate through the supported setup path with real Codex and OpenCode sessions. Check first publication, exact session coverage, later refresh and the time it takes for provisional duration to settle after native completion.
- Fix supported-path defects exposed by the gate and rerun it on the corrected package before stable v3 delivery. Record manual-only combinations separately.

## Capabilities

Add `automatic-ticket-publication` for first publication, lifecycle refresh and native release evidence. It applies to the existing [OpenCode capture capability](../../specs/global-opencode-capture/spec.md) and installed Codex user hooks.

## Impact

An adopting project can start with a selected directory and no ledger. Existing ledgers retain their recorded identity, title and history. Subsequent refreshes follow the ledger across configured roots. The initial selection carries operational write authority only. Portable ledger readers continue to work without the producer's private state.

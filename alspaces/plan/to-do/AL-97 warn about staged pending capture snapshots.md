# AL-97: Warn about staged pending capture snapshots

## Goal

Give developers an optional pre-commit warning when they stage a pending AgenticLens capture snapshot.

## Acceptance criteria

- Setup offers an explicit choice to install the advisory hook and preserves existing hook ownership.
- The hook uses the shared `agenticlens.pending_dir` setting and examines the Git index. It distinguishes snapshots awaiting a ticket from bound tickets awaiting a publication destination.
- The hook reports actionable warnings without changing the index or blocking the commit.

Depends on the [AL-96 pending snapshot contract](../../../openspec/changes/al-96-verify-automatic-ticket-metrics-capture-end-to-end-for-v3/specs/automatic-ticket-publication/spec.md).

[OpenSpec change](../../../openspec/changes/al-97-warn-about-staged-pending-capture-snapshots/proposal.md)

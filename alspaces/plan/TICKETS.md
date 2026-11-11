# AgenticLens Roadmap

Active tickets live in `to-do/`; completed specifications move to `done/` and leave this index. Ready means no outstanding ticket dependency, not implementation approval or verified completion.

Reviewed against the updated PO skill in [PO-22](../PO-22/SPEC.md) at `731de946` on 2026-11-08. Historical evidence and previously accepted product decisions remain authoritative within their stated limits. Existing IDs and accepted ticket groupings are preserved across earlier batches.

## In Progress

- **AL-70**: Honor configuration paths and ticket identity
  - **Status**: In Progress; CI fixture-lifecycle correction; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-70](to-do/AL-70%20honor%20configuration%20paths%20and%20ticket%20identity.md)

- **AL-83**: Present AgenticLens on GitHub Pages
  - **Status**: In Progress; implementation verified, awaiting publication; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-83](to-do/AL-83%20present%20agenticlens%20on%20github%20pages.md)

## Ready

- **AL-80**: Capture OpenCode sessions across worktrees
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-80](to-do/AL-80%20capture%20opencode%20sessions%20across%20worktrees.md)

- **AL-63**: Verify source capabilities across harnesses
  - **Status**: Ready; **Blocked by**: None; **Blocks**: AL-65
  - **File**: [AL-63](to-do/AL-63%20verify%20source%20capabilities%20across%20harnesses.md)

- **AL-71**: Keep dashboard requests and live data reliable
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-71](to-do/AL-71%20keep%20dashboard%20requests%20and%20live%20data%20reliable.md)

- **AL-72**: Make dashboard interactions accurate and accessible
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-72](to-do/AL-72%20make%20dashboard%20interactions%20accurate%20and%20accessible.md)

- **AL-73**: Verify release artifacts before installation
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-73](to-do/AL-73%20verify%20release%20artifacts%20before%20installation.md)

- **AL-74**: Make regression checks reliable and remove unused code
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-74](to-do/AL-74%20make%20regression%20checks%20reliable%20and%20remove%20unused%20code.md)

- **AL-75**: Verify dependency integrity and build reproducibility
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-75](to-do/AL-75%20verify%20dependency%20integrity%20and%20build%20reproducibility.md)

- **AL-78**: Make completion feedback effective across harnesses
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-78](to-do/AL-78%20make%20completion%20feedback%20effective%20across%20harnesses.md)

- **AL-81**: Generate metrics without capture registry access
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-81](to-do/AL-81%20generate%20metrics%20without%20capture%20registry%20access.md)

- **AL-82**: Filter and compare metrics by tags
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-82](to-do/AL-82%20filter%20and%20compare%20metrics%20by%20tags.md)

- **AL2-1**: Generate and preview anomaly debug prompts
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL2-1](to-do/AL2-1%20generate%20and%20preview%20anomaly%20debug%20prompts.md)

- **AL2-2**: Archive workspaces without losing history
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL2-2](to-do/AL2-2%20archive%20workspaces%20without%20losing%20history.md)

- **AL2-3**: Scan and aggregate local projects
  - **Status**: Ready; **Blocked by**: None; **Blocks**: AL2-4, AL2-5
  - **File**: [AL2-3](to-do/AL2-3%20scan%20and%20aggregate%20local%20projects.md)

- **AL2-6**: Estimate API-equivalent costs alongside measured spend
  - **Status**: Ready; **Blocked by**: None; **Blocks**: AL2-7
  - **File**: [AL2-6](to-do/AL2-6%20estimate%20api%20equivalent%20costs%20alongside%20measured%20spend.md)

- **AL2-8**: Show configured reasoning effort across views
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL2-8](to-do/AL2-8%20show%20configured%20reasoning%20effort%20across%20views.md)

- **AL2-9**: Evaluate optional ledger signatures
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL2-9](to-do/AL2-9%20evaluate%20optional%20ledger%20signatures.md)

- **FIX-18**: Report unsupported live telemetry accurately
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [FIX-18](to-do/FIX-18%20report%20unsupported%20live%20telemetry%20accurately.md)

- **FIX-19**: Reject ambiguous Codex session snapshots
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [FIX-19](to-do/FIX-19%20reject%20ambiguous%20codex%20session%20snapshots.md)

- **FIX-20**: Stop source reads when callers cancel
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [FIX-20](to-do/FIX-20%20stop%20source%20reads%20when%20callers%20cancel.md)

- **REV-7**: Review source read guarantees
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [REV-7](to-do/REV-7%20review%20source%20read%20guarantees.md)

## Blocked

- **AL-65**: Recommend safe source-backed backfill defaults
  - **Status**: Blocked; **Blocked by**: AL-63; **Blocks**: None
  - **File**: [AL-65](to-do/AL-65%20recommend%20safe%20source%20backed%20backfill%20defaults.md)

- **AL2-4**: Monitor and refresh remote Git projects
  - **Status**: Blocked; **Blocked by**: AL2-3; **Blocks**: AL2-5
  - **File**: [AL2-4](to-do/AL2-4%20monitor%20and%20refresh%20remote%20git%20projects.md)

- **AL2-5**: Select and compare projects in the dashboard
  - **Status**: Blocked; **Blocked by**: AL2-3, AL2-4; **Blocks**: None
  - **File**: [AL2-5](to-do/AL2-5%20select%20and%20compare%20projects%20in%20the%20dashboard.md)

- **AL2-7**: Compare cost modes and simulate rate cards
  - **Status**: Blocked; **Blocked by**: AL2-6; **Blocks**: None
  - **File**: [AL2-7](to-do/AL2-7%20compare%20cost%20modes%20and%20simulate%20rate%20cards.md)

# AgenticLens Roadmap

Active tickets live in `to-do/`; completed specifications move to `done/` and leave this index. Ready means no outstanding ticket dependency, not implementation approval or verified completion.

Reviewed against the updated PO skill in [PO-22](../PO-22/SPEC.md) at `731de946` on 2026-11-08. Historical evidence and previously accepted product decisions remain authoritative within their stated limits. Existing IDs and accepted ticket groupings are preserved across earlier batches.

## Ready

- **AL-63**: Audit source capability parity across OpenCode v1, OpenCode v2, and Codex
  - **Status**: Ready; **Blocked by**: None; **Blocks**: AL-65
  - **File**: [AL-63](to-do/AL-63%20audit%20capabilities%20across%20opencode%20v1%20v2%20and%20codex.md)

- **AL-69**: Correct telemetry accuracy and preserve evidence
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-69](to-do/AL-69%20telemetry%20accuracy%20and%20preservation.md)

- **AL-70**: Preserve configuration, paths and ticket identity
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-70](to-do/AL-70%20configuration%20paths%20and%20ticket%20identity.md)

- **AL-71**: Make frontend requests and live data handling reliable
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-71](to-do/AL-71%20frontend%20requests%20and%20live%20data%20handling.md)

- **AL-72**: Correct UI interactions, calculations and accessibility
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-72](to-do/AL-72%20ui%20correctness%20and%20accessibility.md)

- **AL-73**: Make installation and release verification reliable
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-73](to-do/AL-73%20installation%20and%20release%20reliability.md)

- **AL-74**: Strengthen test reliability and remove unused code
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-74](to-do/AL-74%20test%20reliability%20and%20unused%20code%20cleanup.md)

- **AL-75**: Verify dependency integrity and reproducible builds
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-75](to-do/AL-75%20dependency%20and%20build%20reproducibility.md)

- **AL-78**: Make open-turn feedback effective across Codex and OpenCode
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-78](to-do/AL-78%20make%20open%20turn%20feedback%20effective%20across%20codex%20and%20opencode.md)

- **AL-79**: Deliver native user-level Codex capture across worktrees
  - **Status**: Ready; **Blocked by**: None; **Blocks**: AL-80
  - **File**: [AL-79](to-do/AL-79%20native%20user%20level%20codex%20capture%20across%20worktrees.md)

- **AL-81**: Make manual generate independent of the capture registry
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-81](to-do/AL-81%20decouple%20manual%20generate%20from%20capture%20registry.md)

- **AL-82**: Add metric tags, dashboard filters and comparison coloring
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL-82](to-do/AL-82%20add%20metric%20tags%20filters%20and%20comparison%20coloring.md)

- **AL2-1**: End-to-End Agent Anomaly Debug Prompt Generator, CLI, and Interactive Web Drawer
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL2-1](to-do/AL2-1%20end-to-end%20agent%20anomaly%20debug%20prompt%20generator%20cli%20and%20interactive%20web%20drawer.md)

- **AL2-2**: End-to-End Workspace Archive Engine, CLI Compaction, and In-Memory Decompression Ingestion
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL2-2](to-do/AL2-2%20end-to-end%20workspace%20archive%20engine%20cli%20compaction%20and%20in-memory%20decompression%20ingestion.md)

- **AL2-3**: Multi-Project Configuration Schema, Domain Aggregations, and Multi-Folder Scanner
  - **Status**: Ready; **Blocked by**: None; **Blocks**: AL2-4, AL2-5
  - **File**: [AL2-3](to-do/AL2-3%20multi%20project%20configuration%20schema%20domain%20aggregations%20and%20multi%20folder%20scanner.md)

- **AL2-6**: Model API Pricing Configuration, Domain Rate-Card Cost Approximation Engine, and Hybrid Ledger Ingestion
  - **Status**: Ready; **Blocked by**: None; **Blocks**: AL2-7
  - **File**: [AL2-6](to-do/AL2-6%20model%20api%20pricing%20configuration%20and%20rate%20card%20cost%20approximation%20engine.md)

- **AL2-8**: Ingest Model Reasoning Effort Variants and Surface Deliberation Settings across Platform Views
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL2-8](to-do/AL2-8%20reasoning%20effort%20telemetry%20integration%20across%20fleet%20kpis%20comparison%20tables%20and%20progression.md)

- **AL2-9**: Evaluate Cryptographic Signing and Verification for Telemetry Ledgers
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [AL2-9](to-do/AL2-9%20evaluate%20cryptographic%20signing%20and%20verification%20for%20telemetry%20ledgers.md)

- **FIX-18**: Report unsupported live telemetry schemas accurately
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [FIX-18](to-do/FIX-18%20report%20unsupported%20live%20telemetry%20schemas%20accurately.md)

- **FIX-19**: Reject ambiguous Codex owner snapshots
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [FIX-19](to-do/FIX-19%20reject%20ambiguous%20codex%20owner%20snapshots.md)

- **FIX-20**: Cancel native source reads with their callers
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [FIX-20](to-do/FIX-20%20cancel%20native%20source%20reads%20with%20their%20callers.md)

- **REV-7**: Review shared source read guarantees
  - **Status**: Ready; **Blocked by**: None; **Blocks**: None
  - **File**: [REV-7](to-do/REV-7%20review%20shared%20source%20read%20guarantees.md)

## Blocked

- **AL-65**: Research default source-backed recomputation and the future of Force
  - **Status**: Blocked; **Blocked by**: AL-63; **Blocks**: None
  - **File**: [AL-65](to-do/AL-65%20research%20safe%20default%20source%20backed%20backfill.md)

- **AL-80**: Deliver global OpenCode capture using the shared worktree runtime
  - **Status**: Blocked; **Blocked by**: AL-79; **Blocks**: None
  - **File**: [AL-80](to-do/AL-80%20global%20opencode%20capture%20with%20shared%20worktree%20runtime.md)

- **AL2-4**: Remote Git Repository Ingestion, Shallow Cache Manager, and Project Management API
  - **Status**: Blocked; **Blocked by**: AL2-3; **Blocks**: AL2-5
  - **File**: [AL2-4](to-do/AL2-4%20remote%20git%20repository%20ingestion%20shallow%20cache%20manager%20and%20project%20management%20api.md)

- **AL2-5**: Multi-Project Web Dashboard Selector, Cross-Project Fleet Overview, and Comparative Analytics
  - **Status**: Blocked; **Blocked by**: AL2-3, AL2-4; **Blocks**: None
  - **File**: [AL2-5](to-do/AL2-5%20multi%20project%20web%20dashboard%20selector%20cross%20project%20fleet%20overview%20and%20comparative%20analytics.md)

- **AL2-7**: Interactive Cost Estimation Dashboards, Hybrid Fleet Analytics, and Rate-Card Simulation
  - **Status**: Blocked; **Blocked by**: AL2-6; **Blocks**: None
  - **File**: [AL2-7](to-do/AL2-7%20interactive%20cost%20estimation%20dashboards%20hybrid%20fleet%20analytics%20and%20rate%20card%20simulation.md)

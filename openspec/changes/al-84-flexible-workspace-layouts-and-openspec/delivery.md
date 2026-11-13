# feat(AL-84): support flexible workspace layouts and adopt OpenSpec

## Pull request

AgenticLens now discovers ledgers across explicitly configured roots and follows recorded ticket identity through archive and arbitrary directory moves. First generation supplies a canonical ID and ledger destination; folder names and planning files establish no runtime identity. Generic configuration writers emit explicit root-only depth 0/budget 300; this repository selects its three containers at depth 1/budget 300. Distinct same-ID ledgers require location selection.

Publication journals and locks use local `agenticlens.state_dir`, initialized separately from capture storage. A small adjacent pending indicator keeps interrupted sets visibly incomplete when private recovery data is unavailable. Completed supported ledgers remain readable without sources or local state. The configuration/API update is breaking: users must adopt `workspaces.roots`; retired scalar keys, runtime root overrides and old journal adapters are rejected.

The repository adopts pinned OpenSpec 2.0.0 for selected execution changes. Roadmap intent and priority remain in `alspaces/plan`; changes hold approved scope, tasks, review/delivery evidence and final telemetry. Only AL-84 execution artifacts moved. Historical ledgers and other active owners remain intact.

Focused checks and native Codex/OpenCode archive/external-move probes pass. Review found five product defects and two test gaps; all corrections and required reassessments passed. Required integrated `make` and final consolidated approval remain pending.

## Current outcome

The user explicitly approved [iteration01 scope](design.md#approved-contract) before implementation. Baseline `163181428ee9a5a599a2da1e91d3af5e618a9682`, branch `AL-84/flexible-workspace-layouts-openspec`, initially clean. All eight independent discovery reviews completed against the frozen 669-file manifest `13a044baddc800f298f54fba03ee41d4c09092d41c89228afde1a79520a821a6`. Corrections remain within the approved iteration.

## Verification and review

[Test evidence](evidence/iterations/01/TEST_RESULTS.md) retains actual checks, restart limits and tester discovery credit. Current native engine probes preserved exact Codex/OpenCode session facts through real OpenSpec archive, external-root movement and late refresh without recreating former homes. Standalone embedded HTML/API reads worked without executables on PATH, local configuration or private state. OpenCode provider counters were synthetic; its native source/tool records were real. Codex cost remained unavailable. These are manual native publication probes, not installed automatic-hook runtime coverage. Native macOS execution remains untested; both Darwin architectures previously cross-compiled.

[Active review](evidence/iterations/01/REVIEW.md) records each focus/profile and the verified findings. Behavior, data, architecture and test reviewers on `gpt-6.1-sol`/high found canonical-filename admission, replacement-read identity, private journal durability and two regression-oracle gaps. Behavior reviewer on `gpt-6-luna`/medium found unsafe duplicate first publication under incomplete discovery, confirmed by the parent. A verification-phase behavior check by Sol/high also found maintenance commands bypassing shared selection; native dry runs confirmed it. Source metadata confirms those runtime model/effort identities. Scoped no-finding approvals do not refute demonstrated defects. All required scoped review profiles now approve the corrected candidate; final consolidated approval awaits integrated checks.

## Decisions and authority

[Design](design.md) owns the approved contract and material steering; [tasks](tasks.md) owns progress. The [roadmap ticket](../../../alspaces/plan/to-do/AL-84%20support%20flexible%20workspace%20layouts%20and%20adopt%20openspec.md) owns evolving intent. Archive/task completion does not itself complete that ticket.

## Telemetry

Final extraction is deferred until verification/review finish. Native child sessions retain their source-proven parent edges. T3 delegated reviewer sessions are separate native roots; final extraction must explicitly include those actual session IDs without inventing child relationships. Raw sources, private journals and diagnostic probes stay outside Git.

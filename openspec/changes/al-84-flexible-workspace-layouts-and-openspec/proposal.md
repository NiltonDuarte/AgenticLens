# AL-84: Flexible workspace layouts and OpenSpec execution

## Why

AgenticLens can discover a ledger recursively while detail readers and publishers reconstruct a different path from its ticket ID. Folder moves and OpenSpec archive therefore risk inaccessible detail or recreation of an old destination. Selected execution work also needs one home for its design, tasks, evidence and telemetry while the roadmap remains independently maintained.

## What Changes

Use explicit roots and bounded ledger-only discovery, one actual-location resolver, move-safe publication, and private operational state with a portable pending indicator. Adopt the pinned OpenSpec execution workflow and reconcile active guidance. Configuration and affected CLI/API consumers must explicitly update; there is no scalar-directory compatibility layer.

## Capabilities

### New Capabilities
- `workspace-locations`: configured ledger discovery, location selection and safe publication across moves.
- `openspec-execution`: linked execution changes with separate roadmap authority and retained delivery evidence.

### Modified Capabilities
None.

## Impact

Backend config/readers/writers, API/dashboard location handling, setup/templates, agent guidance and the public roadmap builder are affected. The engine remains native and standalone. Native sources remain Codex/OpenCode; no new adapter, cross-repository store, compressed archive or phase analytics is added.

## Roadmap and approval

[AL-84 roadmap ticket](../../../alspaces/plan/done/AL-84%20support%20flexible%20workspace%20layouts%20and%20adopt%20openspec.md) owns evolving product intent; `alspaces/plan/TICKETS.md` owns roadmap readiness and dependencies. The [approved execution contract](design.md#approved-contract) and [iteration 01 tasks](tasks.md) own this selected work. The user explicitly approved “$agentic-lens-dev plan and iteration 1 is approved” before coding. File existence and planning readiness do not establish verification or delivery approval.

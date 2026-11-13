# AL-84 iteration 01 architecture review

Outcome: **APPROVED** for the assigned architecture scope.

Reviewed revision: baseline `163181428ee9a5a599a2da1e91d3af5e618a9682` plus the frozen 669-file candidate in `build/al84-review/FROZEN_MANIFEST.json` (declared content SHA-256 `13a044baddc800f298f54fba03ee41d4c09092d41c89228afde1a79520a821a6`). Product/test files were treated read-only. The authorized report path is the only write.

## Findings

No blocking architecture findings in scope. Severity 0.

## Hypotheses investigated

- **Workspace resolution drift between consumers:** Refuted by source tracing. `DiscoverWorkspaceLocations`, `ResolveTicketLocation`/maintenance resolution, and `ValidateLedgerDestinationScope` are centralized in `internal/adapters/workspace_locations.go`; repository reads, ledger formatting, backfill discovery, CLI output admission, capture history, and workspace publication call these shared operations. Explicit roots carry depth and directory budget, discovery records ledger-backed locations, and identity is decoded from ledger facts rather than inferred from paths. Root-specific artifact overrides are propagated with the location.
- **A hidden legacy or registration identity model remains:** No production `workspaces.dir`, root identifier, or runtime `--workspaces` consumer was found. Setup exposes explicit roots and writer defaults of depth 0/budget 300; repo configuration explicitly selects container roots at depth 1/budget 300. Setup/dev/operator guidance consistently rejects retired aliases and registration/sidecar identity.
- **Publication state is mixed with workspace identity or capture enrollment:** Refuted at the architecture boundary. Local publication state is modeled as `agenticlens.state_dir`; capture retains its distinct capture state. The selected-workspace publication adapter acquires publication ownership and recovers/commits through shared publication APIs. The change contract and technical guidance preserve the adjacent pending indicator as portable transaction status and make ordinary reads independent of private state. This review did not independently exercise crash/concurrency behavior.
- **Native manager/engine split or production Node boundary regressed:** Refuted by dependency/build tracing. Manager main delegates to the manager adapter; Makefile builds manager and engine as separate commands, while engine owns the web embed package. OpenSpec's pinned package and Node >=22.12 requirement live under `tools/openspec` and are invoked through development scripts. Setup installation is optional and does not run npm. Existing frontend build Node use is build-time; no OpenSpec dependency enters Go manager/engine packages.
- **Roadmap and execution authority collapsed:** Refuted by workflow and site tooling inspection. Roadmap status/specs remain under `alspaces/plan`; selected execution artifacts live in linked OpenSpec changes. The site builder takes `--roadmap PATH` explicitly and renders roadmap content without using runtime ledger roots. Setup/workflow documents distinguish planning-schema work, development approval, archive, and ticket completion.

## Verification and limits

Static source/dependency trace across workspace adapters and their consumers, manager entrypoint/build targets, setup scripts/templates, OpenSpec toolchain/workflow, site builder, and selected design/spec/tasks. Reused the focused verification evidence in `evidence/iterations/01/TEST_RESULTS.md`; no broad suites or probes were run, and the integrated `make` gate remains the parent’s responsibility. This is a scoped architecture review, not approval of publication race correctness, all 669 files, native macOS runtime, or final integrated delivery.

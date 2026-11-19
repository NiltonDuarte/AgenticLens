# Review scope and evidence

REV-8 covers seven merges to `main` from `16318142` through `9308ac77`. Review each squash commit against its first parent and confirm whether a finding remains at the final revision.

The user requested the next available review ID and authorized adding the explicit `REV-` configuration prefix for session binding. The approval applies to that configuration change and the audit. The coordinator owns the change's review artifacts and any shared verification. Reviewers work against the frozen revision and submit evidence from disposable fixtures.

| Focus | Primary risk |
| --- | --- |
| Behavior and integration | CLI and dashboard reachability, capture routing, live source recovery and documented defaults |
| Data integrity and failure handling | Physical ledger identity, portable facts, capture ownership and ambiguous or failed source reads |
| Architecture and maintainability | Configuration, transaction and release boundaries across callers and adapters |
| Test effectiveness | Whether fixtures exercise regressions and end-to-end controls |
| Release and installer verification | Manifest admission, published inventory and setup/runtime version selection |
| Live stream and dashboard behavior | Read ordering, connection lifecycle, backpressure and recovery from database failure |

Each focus uses the configured `gpt-6.1-sol/high` and `gpt-6-luna/medium` profiles. Discovery runs are independent. The final report records run outcomes and preserves unresolved issues.

## Iteration 02 remediation

The user approved implementation by saying “fix them” in response to the audit handoff. This expands iteration 01’s audit-only scope to the four findings named in that handoff. The baseline is `c20975f6`. The [audit evidence](evidence/iterations/01/REVIEW.md) remains the record for the original findings. The [remediation requirements](specs/review-remediation/spec.md) govern the corrections.

### Ledger writes

Carry the selected physical workspace directory from discovery or resolution into publication admission. Follow the existing retained directory and owned-root comparison in `backend/internal/adapters/workspace_publication.go`. Generation, backfill, and migration must reject a moved or replaced selection before recovery, baseline reads, or writes through the replacement. Ticket equality cannot establish directory continuity. An explicit initial generation destination can remain ledger-free until admission. Keep its current scope and initialization behavior.

Keep selection and admission coherent across the capture worker’s `generate` subprocess boundary. The writer can resolve and admit directly, or the worker can carry its observation into writer admission. A previously resolved pathname alone cannot authorize a write after replacement. Automatic publication continues to require an existing ledger. In a batch, each admitted target retains its real selected home and releases ownership before the next target. Report location changes as incomplete per-target work and continue with independently valid targets. Preserve existing recovery after admission and the portable ledger contract, including unsupported and unavailable source outcomes, session ownership, and supported historical migration. Synchronize move and replacement tests at discovery or selection and writer admission. Verify a copied ledger can still refresh with separate local state and preserve unavailable owners.

### Live stream and engine pin

Give an overflowing SSE stream a bounded path to fresh state. The existing subscribed stream can reconnect or resynchronize, including when recovery returns the same payload. Preserve healthy subscribers and ensure disconnect and unsubscribe remain race-safe. Verify the responsive and saturated cases through the actual stream and client recovery path as applicable.

The first correction closed the overflowing subscriber but could leave the HTTP writer blocked on a real TCP connection. Bound each frame's write and flush using the existing server response budget. Allow idle waits between writes and let the client reconnect through its existing stream lifecycle.

Pin the shared engine to `v3.0.0-rc.156+9308ac7`, the [published release](https://github.com/NiltonDuarte/AgenticLens/releases/tag/v3.0.0-rc.156%2B9308ac7) targeting the audited main revision. Keep exact-version resolution, explicit installation, and checksum enforcement. Verify release assets and exact dispatch in an isolated project and install cache. This published build supplies the compatible pin. Freshly built candidate binaries supply verification for the code fixes.

The isolated install probe verified both release binaries and found that the shared `binary_dir` setting overrides the pin and makes a clean checkout require a source build. Remove that development selector from this repository's shared policy so a clean checkout uses the approved release. Developers can still run built binaries directly. Preserve the product's explicit development-mode behavior for projects that configure it.

### Ownership and verification

The coordinator owns integration, telemetry, and delivery. Application writers own disjoint files for workspace generation and backfill/migration, live service and any affected stream consumers, and the shared pin. The tester owns new tests. Implementation reopens behavior, data, architecture, and test review. Select the live and release checks where those boundaries changed and run the integrated build gate at the final code revision.

## Iteration 03 archived Codex rollouts

The user explicitly approved implementation of this plan with `$agentic-lens-dev approved`. This iteration adds archived-rollout recovery to the delivered remediation scope at `95c130ed`. During that delivery, six explicitly bound REV-8 sessions retained earlier snapshots because reconciliation opened their former paths under `sessions/`. All six files were later found under `archived_sessions/` with their original filenames and session IDs. This supersedes the assumption that a registered rollout must always be readable at its saved path.

In the shared capture reconciliation path, recover only an already-bound Codex registration with a nonempty cached path that fails specifically because it no longer exists. Resolve it by native session ID across the registered store's active and archived directories using the existing exact-owner finder. Charge discovery, metadata and extraction to the existing pass budgets. Reuse the unique selection for project and source validation, fingerprinting and extraction. Keep the existing uniqueness check for readable cached paths. A readable path with contradictory identity, an inaccessible path, or an ambiguous result cannot authorize a replacement.

The selected source must satisfy physical store containment, parsed native ID and existing project and portable-binding checks. Accept the cached path and complete refreshed snapshot together in the existing registry transaction. Recheck the observed path, binding and accepted checkpoint against the current row before updating. Ignore a stale result for publication and retry without overwriting newer facts or registration state. Preserve the ticket attribution and source-backed parent identity. Keep prior facts and emit the existing failure or pending diagnostic when discovery, validation or extraction fails. Unaffected siblings retain normal processing.

The shared path serves the worker, `capture reconcile` and `generate --capture-auto`. Hook and binding admission, manual discovery, OpenCode, ledger schemas and the guarded-worker release protocol retain their existing contracts. The coder owns the capture changes. The tester owns isolated source and registry fixtures. Verify archive movement, repeatability, duplicate handling, budgets, portable history and concurrent registry updates. Any retry against the six real records is a separate operational verification after implementation approval.

Independent review covers behavior through both public capture entry points, data integrity and registry commit ordering, placement of lookup and budget ownership, and regression-test effectiveness. Each focus uses the configured Sol/high and Luna/medium profiles against the frozen candidate. Keep test evidence and the consolidated review in `evidence/iterations/03/`.

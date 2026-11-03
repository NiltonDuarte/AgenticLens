# GEN-20 iteration 02: Evolve the unreleased registry directly

Status: Authorized by the user's explicit scope clarification. Continue implementation without another approval checkpoint.

## Steering delta

Source: the user's clarification, relayed by the parent coordinator: "no old binary, database version not released". The user explicitly removes the quoted numeric-path preservation and old-binary compatibility requirement. This is a material acceptance change, rather than a review correction.

Changed acceptance: evolve the current unreleased registry schema directly. New registries include the persisted unique worktree-coordinate mapping as part of their normal schema. Keep registry version 2; do not add a version 3 solely to fence out old binaries. Do not upgrade old protocol-2 numeric layouts, backfill numeric coordinate mappings, preserve numeric installation paths, or retain compatibility-only tests/code introduced by iteration 01.

Superseded decisions: CONTRACT.md acceptance item 4 and iteration 01's protocol transition, numeric-path preservation, old-reader protection and corresponding verification requirements no longer apply. The immutable baseline contract remains historical evidence; this explicit user instruction governs implementation. Unchanged requirements in that contract and plan remain accepted.

Preserved guarantees: committed tokens remain stable, unique and isolated across concurrent opens; linked Git worktree identification is preferred with administrative-identity disambiguation; otherwise persist a short random identifier. Keep numeric internal SQL row IDs, bounded collision handling, safe filesystem reservation, exact-worktree ownership, containment, read-only access, provisional no-write preview, installation/policy consistency, diagnostics, worker coordination, and supported relocation/resumption behavior. Portable ledger facts remain independent of local capture-state availability.

Unsupported local state: if an existing unreleased registry lacks the required coordinate table or a worktree's mapping, report a contextual unsupported-state/reset-or-reinstall requirement. Do not silently delete a database, overwrite unknown directories, invent a mapping, or adopt old numeric directories. Creating the current table for a genuinely new registry is normal initialization, not migration. Existing protocol-1 relocation functionality predates this ticket; preserve its supported behavior while adapting destination resolution to the current token schema. Do not use this steering as authorization for unrelated recovery removal.

## Grounded implementation delta

1. Coder retains ownership of implementation paths assigned by the parent. In backend/internal/adapters/capture/coordination.go, remove upgradeRegistryV2 and numeric fallback in coordinateTokenTx; retain the shared mapping resolver, token validation, Git-derived naming, random fallback and exclusive bounded allocation. Missing required mapping state must fail explicitly.
2. In capture/schema.go and store.go, define current protocol 2 with worktree_coordinates included at initialization, remove the version-2-to-3 transition and dual-version guards, and validate existing state before treating it as supported. Do not leave an existing registered scope half initialized after missing-layout rejection.
3. In capture/relocate.go, remove iteration-01-only dual protocol-2/3 acceptance/fencing and legacy coordinate assumptions. Preserve pre-existing supported source relocation, journal identity, exclusive destination worker admission, interruption/resume and retained inspection against the current supported token mapping.
4. Preserve and verify the currently implemented consumer changes in project.go, install.go, worker_successor.go and CLI capture.go, capture_diagnostics.go and capture_worker.go. New previews are provisional until allocation; actual manifests receive committed paths, and policy/worker decisions never compare independently generated UUID proposals.
5. Tester stops compatibility-only work, removes tests introduced solely for numeric preservation/old readers/upgrades, and adapts current-schema fixtures to resolve real production paths. Keep unrelated existing behavior assertions and protocol-1 recovery coverage. Parent updates summary and PR context to describe final scope; no application implementation belongs to planner or coordinator.

## Verification impact

Discard iteration-01 evidence only where code or acceptance changes invalidate it. Re-run focused adapter/CLI checks after simplification and the required integrated make gate. Preserve prior valid results for unchanged behavior with an explicit revision rationale. No test success is claimed by this delta plan.

Required cases remain first-allocation concurrency for same/distinct worktrees, stable reopen, Git basename disambiguation, non-Git/plain-checkout fallback, collision exhaustion and unknown-entry preservation, aliases/containment, read-only access, no-write dry-run, committed installation paths, policy equivalence, diagnostics and worker isolation, supported relocation and resume. Add only the focused missing-layout rejection assertion needed to prove explicit non-destructive failure; no legacy compatibility matrix or protocol upgrade fixture is required.

The four-focus/two-profile review policy from iteration 01 remains applicable. Use unique reports under iterations/02: REVIEW_BEHAVIOR_SOL.md and REVIEW_BEHAVIOR_LUNA.md; REVIEW_DATA_SOL.md and REVIEW_DATA_LUNA.md; REVIEW_ARCHITECTURE_SOL.md and REVIEW_ARCHITECTURE_LUNA.md; REVIEW_TESTS_SOL.md and REVIEW_TESTS_LUNA.md. The same configured gpt-6.1-sol/high and gpt-6-luna/medium profiles apply to each focus. Data review now asks whether the current mapping is complete and missing-layout rejection preserves unsupported state; test review checks that removed assertions were compatibility-only. Freeze the revised implementation before independent review; never approve it against superseded numeric-preservation expectations. Parent owns consolidated evidence, telemetry, commits and PR delivery.

## Completion

Implemented and verified on 13a6ff38c97e9c9d3019d76a0944ce46afe1814e. All eight discovery cells completed; the coordinator corrected one shared defect and one optional test-oracle gap, verified before/after behavior and rejected mutation, and approved the consolidated review after final make passed. See SUMMARY.md and REVIEW_SUMMARY.json for attribution and limits. Telemetry and PR delivery follow as the final workflow steps.

## Review terminology correction

Source: the user asked to replace ambiguous token terminology in the table and source, then explicitly instructed "update it". This authorizes a mechanical correction within iteration 02: rename the worktree_coordinates column to directory_name, use directoryName in Go, and rename the associated validation, lookup, preferred-name and random-name helpers, tests and diagnostics consistently. Keep genuine session/request tokens unchanged. Allocation, persisted name values, path validation and SQL row IDs retain the accepted behavior; no compatibility migration is required for the unreleased schema.

Verification impact: existing fixture-based allocation, reopen, collision, alias and installation tests exercise the renamed SQL column. No new test declaration is needed for the rename. Run affected checks and the integrated make gate. A narrow behavior/integration review under both configured model profiles is proportionate: changes are identifier/string substitutions, so the previous data, architecture and test investigations carry forward for unchanged behavior. Each reviewer owns a unique REVIEW_NAMING_SOL.md or REVIEW_NAMING_LUNA.md report; parent owns consolidation and telemetry.

# GEN-20 iteration 01: Unique coordination tokens

Status: Approved for implementation by the user on 2026-11-03. Approval covers the contract and plan as presented, including preserved numeric paths and registry/preview/relocation compatibility. No renewed approval is required within these boundaries.

## Grounding

- backend/internal/adapters/capture/store.go: openRegistryModeContext uses an immediate SQLite transaction, INSERT OR IGNORE and UNIQUE worktrees.directory, then builds worktrees/<numeric-id> in normal and relocation opens. There is no demonstrated committed-allocation race.
- capture/schema.go: registry protocol 2 stores numeric worktree IDs referenced throughout operational tables. Keep those IDs; introduce a separate unique directory-token association and a compatible transactional protocol transition.
- capture/ownership.go: currentWorktreeOwnership already resolves .git directory/file markers and canonical administrative identity without requiring Git execution during ordinary capture. Reuse this observation for linked-worktree naming; retain all ownership validation.
- capture/project.go: StateLocation carries the numeric SQL ID; PreviewStateLocation currently predicts max(id)+1 and isCoordinationPath accepts positive decimal only. resolveCoordinationNamespace validates the exact scope through a read-only store. Replace numeric-only path assumptions and make unallocated preview semantics explicit.
- capture/relocate.go: retained inspection constructs numeric paths directly; real relocation opens destination scopes and resumes journaled installation. Resolve mapped paths in both flows and preserve existing pending scopes.
- backend/cmd/agentic-lens/capture.go uses preview before install, then replaces it with OpenProject's actual path for real installation; later policy equivalence also compares previews. capture_diagnostics.go, capture_worker.go and capture/worker_successor.go are additional preview consumers. Check absent-state behavior before changing preview output.
- Existing coverage owners: capture/shared_state_test.go tests concurrent operations after sequential scope opening, ownership reuse rejection and worker isolation; namespace_test.go tests preview/aliases and escaping paths; relocate_test.go and relocate_install_unix_test.go cover relocation/resume/install; install_transaction_test.go and CLI capture_shared_state_test.go cover installed state. Several fixtures hardcode worktrees/1 and need deliberate legacy fixtures or production-resolved paths rather than blanket text replacement.
- CLI is not installed on PATH in this checkout. Runtime help and tests have not been executed. Confirm help from the built binary when implementation verification starts.

## Implementation sequence and ownership

1. After user approval, coder owns the capture adapter implementation, directly affected CLI consumers and corresponding documentation. Implement the token association and protocol-2 upgrade first, preserving all existing numeric paths and row IDs. Ensure unsupported and read-only opens do not mutate state, and old binaries cannot unknowingly operate on new token scopes. Freeze the precise transition before extending consumers.
2. Add one shared validated resolver/allocation boundary. Prefer a bounded portable Git worktree identifier plus administrative-identity digest; otherwise persist 16-hex random UUID representation. Scope token uniqueness transactionally; refuse unsafe/reserved entries and retry collisions within a fixed bound. Existing scopes never generate a replacement token merely because optional identity enrichment changes.
3. Replace numeric directory construction in writable/read-only store opens, relocation locks and retained inspection. Extend coordination-path recognition without weakening containment or exact-worktree validation.
4. Represent unallocated preview explicitly and update every caller. Preserve no-write dry-run; durable install manifests receive only the committed path. Policy selection must not depend on independently generated UUID proposals. Workers and diagnostics must handle absent scope honestly, and successor checks must resolve installed state consistently.
5. Coder extends coherent existing tests for changed behavior. Tester owns additional justified regression tests and verification records, with application implementation read-only. Parent owns integration, workspace summaries, telemetry, commits and PR delivery. All agents share the checkout and must preserve others' edits; assign exact writable paths before dispatch.

## Selected regression cases

| Plausible regression | Existing coverage owner and required proof |
| --- | --- |
| Two first-time opens assign conflicting scope/token or one worktree gets two paths | Extend shared_state_test.go with synchronized first opens for both the same and distinct worktrees; assert durable unique mapping, stable reopen, independent worker locks and existing scoped data behavior. Current test opens sequentially, so it does not prove first-allocation concurrency. |
| Identical Git administrative basenames from separate repos share state | Extend shared-state fixtures with two repositories sharing project_name and a linked-worktree basename; assert different tokens and registration isolation. |
| UUID is regenerated, collision overwrites another scope, or fallback masks malformed Git | Exercise non-Git/plain-checkout fallback and reopen, bounded allocation conflict/error behavior, and existing ownership-error paths using isolated fixtures. Use a narrow entropy boundary only if deterministic application failure coverage needs it; do not test crypto library internals. |
| Upgrade moves a live directory or exposes a new scope to an old reader | Build a self-contained protocol-2 registry with existing numeric files and data; assert read-only no-write resolution, atomic writable upgrade, stable legacy paths, new token allocation, and honest rollback/error behavior. Retain protocol-1 rejection/recovery coverage. |
| Broader token recognition accepts another scope or escaped path | Extend namespace_test.go with valid token aliases, another scope's token, unknown tokens and escaping symlinks, reusing existing containment fixtures. |
| Preview UUIDs cause false policy mismatch or installer writes provisional paths | Extend CLI capture installation/policy coverage and install transaction tests: repeated dry-run writes nothing; real install embeds the committed token path; current installed state reopens; absent state stays absent. |
| Relocation resumes into another token or retained inspection recreates state | Extend relocation tests with a fresh token destination, an existing numeric destination and interruption/resume; assert preserved registrations/diagnostics/ownership and read-only retained inspection. Preserve the current real installer fixtures. |

These cases apply the regression checklist's configuration/path portability, persistence/migration, cleanup/compatibility and independent-review sections. Ledger extraction is unchanged; reuse existing portable-ledger tests unless the implementation reaches generation/refresh. Do not add an unrelated full ledger matrix merely because local state gains a mapping.

## Verification and review gates

Run focused adapter and CLI tests with race detection after the canonical embedded-asset prerequisite. Tester records the tested revision/diff, command, result and limits. Run the required full make after integration. No frontend source changes are planned; normal build prerequisites still apply. Tests and reviewer probes must remain under disposable fixture directories.

This persistence/path change requires all four independent reviewer focuses under .agents/skills/agentic-lens-review/references/parallel-review.md. The coordinator expands every focus under model-policy.yaml into gpt-6.1-sol/high and gpt-6-luna/medium on provider codex, eight discovery runs in capacity-limited waves against the same frozen revision. Verify dispatch availability; unavailable required cells are INCOMPLETE rather than implicitly waived. Each child reads only its assigned persona and common review context, and owns its unique report. The parent owns consolidation and private structured run/finding records.

| Focus | Risk question | Assigned report paths under this iteration |
| --- | --- | --- |
| reviewer-behavior | Do install, preview, policy, workers and relocation agree on committed scope paths? | REVIEW_BEHAVIOR_SOL.md; REVIEW_BEHAVIOR_LUNA.md |
| reviewer-data | Are upgrades, collision allocation, retained numeric state and interrupted relocation lossless and isolated? | REVIEW_DATA_SOL.md; REVIEW_DATA_LUNA.md |
| reviewer-architecture | Is token resolution centralized, versioned clearly and confined to the imperative shell? | REVIEW_ARCHITECTURE_SOL.md; REVIEW_ARCHITECTURE_LUNA.md |
| reviewer-tests | Do fixtures independently detect first-allocation races, legacy compatibility and durable manifest errors without touching real state? | REVIEW_TESTS_SOL.md; REVIEW_TESTS_LUNA.md |

No separate specialist is selected: the concrete filesystem trust and transactional failure risks belong to behavior/data coverage, and no new public trust boundary or performance budget is introduced. Re-review affected focuses after fixes. Consolidated approval follows integrated verification; retain REVIEW_SUMMARY.json before completed-review cleanup. Generate source-explicit telemetry only at final delivery, using ticket type chore, and report source gaps rather than inventing success.

## Approval checkpoint

Present the contract and this plan to the user before dispatching coder. Approval covers new unique names, retained legacy numeric names, a safe local registry protocol transition and explicit provisional preview handling. Material changes to those boundaries require renewed approval; ordinary review corrections remain in iteration 01.

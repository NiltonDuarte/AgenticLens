# GEN-20 iteration 01 test evidence

Status: Preparation only; implementation is mutable. No application/test edits or suites executed by the tester yet.

## Baseline and boundary

Baseline revision: dc44619ae15586fe1194f872743033f523635c92. The approved contract and iteration plan govern new local coordination tokens, preservation of numeric legacy scope paths, protocol upgrade, and preview/install/relocation consistency. Domain calculations, frontend behavior and portable ledger facts are unchanged by the proposed scope.

## Existing coverage and selected gaps

| Boundary | Existing coverage inspected | Missing proof to assess after implementation |
| --- | --- | --- |
| Concurrent allocation and isolation | shared_state_test.go, TestSharedNamedRegistryScopesConcurrentSameTicketWorktrees: real Git worktrees and SQLite; concurrent scoped registration/publication/diagnostics and separate worker locks; foreign coordination rejection | Existing scope opens are sequential. Synchronize initial opens on a fresh registry for the same and distinct worktrees, assert stable persisted coordinates after reopening, and retain isolation assertions. |
| Linked Git identity / UUID persistence | sharedWorktrees helper creates linked Git worktrees; ownership test rejects physical directory reuse and accepts ordinary writes | Cross-repository identical administrative basenames, ordinary-checkout/non-Git fallback, persisted fallback after optional Git enrichment, collision and reserved-path safety. |
| Registry compatibility | TestUnsupportedRegistryOpenDoesNotChangeLegacyDatabase preserves protocol-1 database bytes and absence of WAL/SHM files | A self-contained protocol-2 fixture must prove read-only no-write numeric resolution, transactional writable upgrade, preserved legacy state and new token scope allocation. |
| Namespace and aliases | namespace_test.go resolves base/namespace/coordination aliases; rejects escaping and malformed namespaces before allocation | Explicit provisional preview for an absent scope; persisted token alias acceptance and unknown/foreign token rejection without allocation. |
| CLI install and policy | TestCaptureCLIInstallerPreservesUserHooksDryRunAndRepeatedRemoval proves dry-run preserves hooks and state absence, real installation is idempotent, edits are refused; TestCaptureInstallerUsesSharedLocalAndExplicitStatePolicy exercises selected paths | Repeated pre-allocation preview cannot cause policy mismatch; durable manifests use the committed coordinate, while callback registry paths retain their intentional contract. |
| Relocation and retained state | Relocation tests preserve scoped facts, peer data, pending reservation fences, resumed installation, and source safety before retirement | Replace incidental numeric destination assumptions with resolved new tokens while retaining deliberate legacy numeric fixtures; verify interrupted relocation reuses its committed coordinate and read-only retained inspection writes nothing. |

Tests use temporary repositories/state. No test may depend on this workspace, live capture state, or private session history. Existing pure-domain and frontend checks may be reused through the integrated make gate; the scope does not justify unrelated new tests.

## Verification plan

After the coder freezes implementation and ownership is transferred, inspect the final regression assertions before adding tests. Run focused Go adapter/CLI checks with race detection after make web-assets. Record the exact tested diff, commands, pass/fail results, skips and environment limitations here. Parent owns the required integrated make gate, product fixes, review consolidation and publication.

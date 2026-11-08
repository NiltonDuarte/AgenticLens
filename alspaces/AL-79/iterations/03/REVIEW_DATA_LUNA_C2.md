# AL-79 iteration 03 C2 — data and failure re-review

**Outcome: APPROVED** for the assigned data/failure scope. No new blocking finding; no severity assigned. This is not whole-repository or delivery approval.

Run `i03_data_luna_c2`; requested identity `codex/gpt-6-luna/medium` (runtime metadata unavailable). Reviewed the frozen C2 candidate based on `9c39cf9d4f44bdb968cb92b0212dccf577d10255`; `freeze_candidate.py --verify` matched SHA-256 `d3a52d7385dae492e3c224f76b890234cc673987db543ee0cffefae58b910afe` at start and end. The C2 manifest and diff were present and the verifier matched the full candidate inventory. Product/test code remained read-only; only this assigned report was added.

## Prior finding and hypotheses

The C1 finding that replacement could overwrite an unowned `<manager>.sha256` sidecar is corrected. `replaceManagerWithCheckpoint` now refuses a sidecar unless an executable and valid matching installation receipt establish ownership, and refuses mismatched sidecar bytes before it creates a pending transaction. The orphan and edited-sidecar cases are present in `TestManagerReplacementRequiresOwnershipAndExclusiveLease`; the C2 manager evidence additionally records real orphan/edited-sidecar controls with no destination or journal mutation. I independently ran `go test ./internal/adapters/manager -run 'TestManagerReplacementRequiresOwnershipAndExclusiveLease|TestManagerNativeArtifacts' -count=1`: PASS.

The C2 physical-path fixes also preserve ownership and cache placement: self-update derives the destination from the executing binary and resolves it physically, so symlink invocation updates the installed target; explicit cache roots use the component-wise `PhysicalPath` resolver, preserving symlink-before-`..` semantics for not-yet-created leaves. The recorded C2 adversarial manager tests exercise both cases and confirm the alternate/default destination remains absent. Local foreign manager artifacts are rejected before cache creation. Cache runtime selection still checks the artifact's embedded exact version against the project pin. No engine worker fence or portable-ledger path changed in the C2 correction boundary; previous worker-continuity and portable-history approvals remain applicable to those unchanged paths.

The prior late-rollback conflict hypothesis remains refuted: journal recovery recognizes already-restored before-images and retains conflicts for retry; the prior disposable overlay exercised all reverse-rollback prefixes and repeated recovery. No new ownership, cache mutation, recovery, or worker-continuity defect was substantiated.

## Evidence and limits

Inspected the approved iteration plan, C2 candidate manifest/diff, manager storage/install/run and path resolution, manager ownership/cache tests, worker continuation dependencies, and recorded C2 manager/distribution evidence. Reused `MANAGER_TEST_EVIDENCE.md` for the complete race suite (10.663s) and adversarial overlay controls, plus the C2 distribution and retained-state evidence in `TEST_EVIDENCE.md`; this review ran only the focused manager tests noted above. Full `make`, real machine installation, Windows runtime behavior, and loaded-update execution remain parent-owned or uncertified. No live HOME/hooks/trust or telemetry source was accessed or modified.

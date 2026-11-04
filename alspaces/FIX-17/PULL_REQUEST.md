# fix(capture): allow isolated manual registry recovery (FIX-17)

Manual generation with a fresh or healthy explicit `--capture-dir` can recover when configured capture state has an unsupported coordinate layout. Selected-state validation and configured/selected/discoverable installation ownership and provenance checks remain enforced; default, native and automatic capture retain strict admission.

The incompatible registry is preserved. This does not migrate or repair capture state, certify automatic-capture health, or change portable ledger authority. GEN-21 integration and its telemetry stay with that ticket owner; use a native containing this implementation rather than an older release pin.

Scope: [CONTRACT.md](CONTRACT.md). Delivery, review and runtime evidence: [SUMMARY.md](SUMMARY.md).

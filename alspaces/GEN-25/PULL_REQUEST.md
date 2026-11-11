# chore(GEN-25): remove Windows support

AgenticLens is moving forward with macOS and Linux support only. This ticket removes the remaining Windows runtime, distribution, installer, plugin transport and test paths after Windows CI was removed, and updates current support guidance.

The proposed scope preserves macOS/Linux amd64 and arm64 behavior, installation safety, manager/engine separation and portable recorded history. Historical audits and release notes remain historical records.

Implementation, verification and independent review are pending scope approval. This is a planning draft, not a completed delivery claim.

See the [contract](CONTRACT.md), [plan](iterations/01/PLAN.md) and [summary](SUMMARY.md) for acceptance boundaries and current evidence.

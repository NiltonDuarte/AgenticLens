# fix(REV-5): preserve telemetry recovery and harden audited boundaries

Failed forced initialization could displace configuration without restoration. Validate targets and protect staging/install failures with rollback and cleanup. Remove confirmed unused/test-only production APIs and the unread live buffer while preserving meaningful assertions through active APIs. Validate frontend HTTP payloads without discarding valid data for malformed optional diagnostics.

Integrate main47af923 schema-v5 cost evidence and split DTOs, preserving unknown/partial/observed costs and supported legacy reads. Compose general and cost-specific validation without restoring obsolete helpers.

Iteration 03 records a complete tracked-code audit with per-file coverage. Additional findings remain open; ADP-05 is P2 default-bind hardening under local-only product scope, with deliberate remote exposure operator-owned. No listener/authentication change.

Iteration 04 fixes ADP-01 and D3: warning-bearing Codex recovery retains stored owner trees with incomplete diagnostics while healthy trees refresh; release builds select and verify the requested tag before checks/build, and forced rebuilds do not create unrelated releases, move latest or purge candidates.

Iterations 05-06 make correctness investigation the canonical review process. Four dedicated reviewer personas (behavior, data/failure, architecture, test effectiveness) load one shared review-context file. Substantive changes use independent parallel reviews, risk-selected specialists and capacity-limited waves. The coordinator verifies findings, preserves evidence-backed dissent, tracks revision-aware re-review and owns a consolidated gate including INCOMPLETE. This is a skill-kit protocol, not a newly installed CI scheduler. Bounded evaluations do not establish improved detection rates.

This REV-5 session supersedes historical GEN-14 audit PR #91. Scope, evidence and limitations: alspaces/REV-5/CONTRACT.md, SUMMARY.md and iterations/02 through 06.

Iteration 07 adds configurable paired Sol 6.1/high and Luna 6/medium runs for every selected reviewer, with durable per-model run/finding attribution. Both models completed the scoped policy review with no findings; source metadata verifies requested settings. Telemetry linkage limitations are recorded.

# fix(REV-5): remove unused production paths and harden boundary handling

Failed forced initialization could move an existing configuration out of its path without restoring it. Validate both targets first and protect staging/install failures with cleanup and rollback that report errors.

Remove production APIs whose only callers were tests, the live trace buffer with no runtime readers, obsolete merge algorithms, and an incorrectly typed unused progression client. Retarget meaningful assertions to the active extraction, migration, filtering and merge paths; preserve codecs, public routes, session ownership and actual revision caching.

Validate active frontend HTTP payloads while preserving supported Go null/omitted fields. Malformed optional diagnostic headers now show a warning without discarding valid workspace data. Keep the legacy trace-buffer configuration accepted and document its behavior consistently.

Reclassify the same audit/session from GEN-14 to REV-5, preserving historical evidence. This remediation supersedes audit PR #91. Contract, candidate dispositions, verification and limitations are recorded in alspaces/REV-5/CONTRACT.md, SUMMARY.md and iterations/02/.

Integrate main47af923 while preserving schema-v5 cost evidence, nullable amounts, observed zero and legacy reads. Compose cost-specific and general frontend response validation without restoring removed clients; preserve the upstream DTO split and remove its confirmed unused helper reintroductions. Incoming-code review and integration evidence are in iterations/02/INTEGRATION_*.md.

Iteration 03 adds a full tracked-code audit with per-file coverage and unchanged-source hashes. It identifies additional open correctness, architecture, testing and delivery issues beyond the original fixes. These audit findings are documented, not implemented; see iterations/03/REVIEW.md and COVERAGE.json.

ADP-05 is now P2 default-bind hardening under the clarified local-only product scope. Authentication is not required; intentional remote exposure remains operator-owned. No listener behavior was changed.

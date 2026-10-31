# fix(REV-5): remove unused production paths and harden boundary handling

Failed forced initialization could move an existing configuration out of its path without restoring it. Validate both targets first and protect staging/install failures with cleanup and rollback that report errors.

Remove production APIs whose only callers were tests, the live trace buffer with no runtime readers, obsolete merge algorithms, and an incorrectly typed unused progression client. Retarget meaningful assertions to the active extraction, migration, filtering and merge paths; preserve codecs, public routes, session ownership and actual revision caching.

Validate active frontend HTTP payloads while preserving supported Go null/omitted fields. Malformed optional diagnostic headers now show a warning without discarding valid workspace data. Keep the legacy trace-buffer configuration accepted and document its behavior consistently.

Reclassify the same audit/session from GEN-14 to REV-5, preserving historical evidence. This remediation supersedes audit PR #91. Contract, candidate dispositions, verification and limitations are recorded in alspaces/REV-5/CONTRACT.md, SUMMARY.md and iterations/02/.

# Architecture review (gpt-6-luna / medium requested)

**Outcome: APPROVED** — scoped architecture/maintenance review only.

Reviewed frozen candidate fingerprint `903c512467c9100cbb6b80756e501280efb9a3ab08f3466c0f4890cae92ece1e`, based on `e639f3303c9d92236edce85befccda66617d7cb8`. Scope: RC workflow, Makefile, contributor release guidance, replacement workflow tests, deleted `scripts/release` resolver/dependencies, and unchanged Release Please producer workflow/config/manifest. No blocking architecture findings.

The change establishes one version authority: the existing Release Please PR manifest. The independent resolver, its package/lockfile, installation step, and resolver-specific tests are removed. This avoids duplicated release policy and keeps the new version lookup in the existing CI workflow; production Go/TypeScript code and runtime dependencies are untouched. `Makefile`'s release-workflow target now directly runs its Python workflow tests, consistent with the deleted Node-only engine.

The handoff between workflows is understandable and bounded: successful Release Please completion triggers RC selection; the candidate workflow checks out the event SHA, reads the generated PR's immutable manifest ref, and rechecks the selected PR and main before publication. The unchanged producer at `.github/workflows/release-please.yml` uses the same root manifest/config described by the candidate workflow. The branch and pending label are explicit selectors rather than title parsing. Contributor guidance was updated to describe the new authority and skip/fail behavior.

Architecture hypotheses assessed:

- **Duplicated version calculation/dependency remains:** refuted by deletion of `scripts/release/rc-version.cjs`, its package files and dependency installation; no remaining resolver references were found in active Makefile/workflow/docs consumers.
- **Release Please stable path was altered or pulled into RC-specific responsibilities:** refuted by the diff; `.github/workflows/release-please.yml` and `.github/workflows/release-stable.yml` are unchanged. RC metadata handling remains confined to the candidate workflow.
- **New workflow boundary adds a production dependency or speculative abstraction:** not observed. It uses existing runner tools (`gh`, Bash, `sort`, GitHub environment/output files), with no new package or product-code interface.

Evidence limits: the assigned brief reports 13 passing focused tests (zero skips), including production shell execution against a fixture `gh`, and live read-only evidence for current PR/manifest provenance. Fixture tests do not execute hosted scheduling, real gojq filters or upload behavior; those limits do not expose an architecture defect in this scope. Full integrated `make` remains pending per the brief. This approval is limited to architecture and maintainability of the frozen diff, not overall release correctness or hosted behavior.

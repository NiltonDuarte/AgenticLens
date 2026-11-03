# GEN-20 iteration 02 architecture review

**Outcome: APPROVED** for the architecture focus at candidate `67d2a8bdd7039e7638e5dfb68fc435993252433b`.

## Scope and evidence

Reviewed the current-schema coordination resolver and allocation in `backend/internal/adapters/capture/coordination.go`, registry initialization and open paths in `store.go`, location previews in `project.go`, plus installation, worker successor, CLI diagnostics, relocation and retained-state consumers. Inspected unchanged ownership discovery in `ownership.go` and the relevant callers. Requirements are the current iteration 02 plan; the numeric-path compatibility and old-binary fence in the historical contract are superseded.

The requested base object `dc44619ae15586fe1194f872743033f523635c92b` is absent from this checkout. I reviewed the frozen candidate commit and its parent diff instead, so exact base-to-head attribution could not be checked. No product or test files were modified. No tests were run: the iteration test report and `/tmp/gen20-make.log` were reused as supplied evidence, and the full `make` gate was running independently. The log excerpt records frontend tests/build and multi-platform binary building; it does not establish the final make exit status.

## Findings

None in the assigned architecture scope. No blocker or optional architecture suggestion was substantiated.

## Main hypotheses investigated

- **Allocation responsibilities and concurrent identity:** SQL row identity stays in `worktrees`; `worktree_coordinates` owns the durable filesystem token. `allocateCoordinate` performs Git-derived preference or random fallback, checks both persisted and filesystem reservations, creates the directory exclusively, then inserts the mapping within the caller's serialized write transaction. Concurrent first opens therefore cannot publish the same token; losing filesystem reservations are retried, with bounded exhaustion reported. Unassociated opens continue to use the registry directory.
- **Git qualification and path safety:** `currentWorktreeOwnership` resolves the Git administrative directory and records its file identity and repository proof. The preferred token includes a sanitized administrative-directory basename plus a digest over its path and identity, so matching basenames in distinct repositories do not collapse to one coordinate. Token validation, `Lstat` checks, and the existing project/namespace ownership checks protect resolution against unsafe components and symlink redirection.
- **Current-schema admission and unsupported state:** the registry schema creates the mapping table for new databases. Existing registries are validated before writable initialization proceeds; missing table or scope mapping is rejected without a numeric fallback or automatic rewrite. Numeric SQL IDs remain internal. No version bump, backfill, old-reader fence, or compatibility-only dual behavior remains.
- **Resolver and consumer consistency:** writable/read-only opens, existing previews, retained-state inspection, and relocation destination/inspection paths resolve persisted mappings through the shared token/path functions. Unallocated previews explicitly carry `Provisional` with no path or guessed SQL ID. CLI dry-run uses an explicitly marked pending placeholder, while install opens/allocates before generating durable manifests. Policy checks and worker successors reject provisional locations; diagnostics avoid probing worker/manifests for them. This keeps a fresh UUID proposal from being compared as if it were committed identity.
- **Layering and maintainability:** filesystem observation, entropy, SQLite transactions, and schema rejection remain in the adapter shell. No domain-layer I/O or random behavior was introduced. Naming is responsibility-based rather than falsely version-specific, and the shared resolver avoids independent numeric/path reconstruction across callers. SQLite driver's direct import is used for bounded lock retry during concurrent first initialization, not exposed beyond the adapter.

## Verification and limits

Reused `iterations/02/TEST_REPORT.md` for the reported focused race checks, unsupported-layout checks, and test-source mapping. Reused the available `/tmp/gen20-make.log` excerpt only as partial integrated-build evidence; final completion remains with the parent. No isolated probe was necessary after the source trace and supplied concurrent allocation evidence.

Approval is scoped to architecture at the frozen candidate only. The absent specified base prevents exact attribution against that revision, and the make log excerpt does not show its final status. Neither limitation exposed a concrete architecture defect in the inspected candidate paths.

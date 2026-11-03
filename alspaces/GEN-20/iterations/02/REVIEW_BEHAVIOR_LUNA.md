# Behavior review — GEN-20 iteration 02

## Assessment

- **Verdict:** APPROVED for the assigned behavior and integration scope.
- **Reviewed source:** `67d2a8bdd7039e7638e5dfb68fc435993252433b`; base `dc44619ae15586fe1194f872743033f523635c92`. Reviewed backend diff and relevant unchanged capture namespace, install, worker, and relocation consumers.
- **Identity:** requested/configured model `gpt-6-luna`, medium effort. Runtime identity was not independently observable in the shell; the review persona/runtime context identifies this session as gpt-6-luna with medium effort.
- **Findings:** none.

The iteration-02 contract governs: current registry protocol remains 2; numeric coordination path retention and old-binary fencing are out of scope. I found no behavior in scope that conflicts with that clarification.

## Traced behavior and hypotheses

1. **Allocation and reopen:** `OpenProjectContext` resolves the named registry and enters the writer transaction. Existing scopes resolve their persisted token; new scopes derive a qualified token from linked-worktree administrative identity or use the random 16-hex fallback, reserve the directory with exclusive `Mkdir`, and insert the unique mapping in the same transaction. Same-name worktrees from separate repositories receive different digests. Failed later ownership checks roll back the mapping; any leftover reserved directory is skipped on a later attempt. Read-only open resolves the same mapping and does not create directories.
2. **Aliases and containment:** explicit coordination aliases are accepted only when a valid token is beneath `worktrees`, the namespace resolves to the configured project namespace, and the read-only registry lookup maps the exact physical path back to the current worktree. Token validation restricts path components; `coordinatePath` rejects a symlinked worktrees root or token directory. Unknown aliases cannot be adopted through preview/open.
3. **Preview/install/policy:** new preview returns `Provisional` with no coordination path or SQL allocation. CLI dry-run substitutes an explicit `allocation-pending` path only for rendering, sets `state_allocation_pending`, and does not open the writer. `logicalRegistryDirectory` maps this safe child back to the shared registry for containment and generated integration configuration. Real install opens and commits the scope before rendering, then uses its persisted path; later previews resolve it. Installation policy compares validated logical registry paths for provisional states and committed coordination paths otherwise, avoiding independently generated UUID comparisons.
4. **Diagnostics and worker:** provisional diagnostics inspect only the registry root and suppress worker/manifest reads; missing registry reads are normalized to an empty inspection result. Automatic capture declines provisional state. Worker successor admission rejects provisional configuration and compares the committed configured path against the active path before continuing, including the takeover recheck.
5. **Relocation/resume:** retained-state inspection reads the persisted mapping. Relocation destination preview validates the current mapping layout; actual relocation allocates/reserves a mapped destination scope before import and uses that store directory for installation manifests, evidence copy, activation, and later resume. Existing protocol-1 migration still reaches the current allocator; no numeric fallback was found.
6. **Unsupported layouts:** missing mapping table/rows fail during validation before a writable scope operation, while read-only preview/open surfaces the error. The path does not drop, recreate, or adopt a database or old numeric directory.

Potential counterexamples were checked against the relevant guards and call paths. In particular, the provisional install sentinel is not a committed identity, and installation is the only caller rendering against it; workers, successor takeover, and diagnostics do not treat it as an allocated worktree path.

## Verification and limits

- Reused focused adapter and CLI race-test results recorded in `iterations/02/TEST_REPORT.md`; the tester's declaration listing was not treated as execution of test bodies.
- Reused the coder's reported focused race commands from that report; not independently rerun.
- `git diff --check` passed for the frozen base-to-head diff.
- Inspected `/tmp/gen20-make.log`: frontend build/typecheck/lint, `go vet`, and Python release/installation/distribution suites reported success; the log ended at `(cd backend && go test -race ./...)` without its completion result. I do not claim the full `make` gate passed.
- This is a scoped behavior review, not proof of absence of defects. Filesystem race behavior beyond the exclusive reservation path, arbitrary malformed SQLite schemas, and final integrated `make` completion remain outside direct execution in this review.

# Behavior and integration review

**Outcome: APPROVED (scoped).** Reviewed frozen candidate `b6ab0471224c8bcea7bc00adce867d3ba0de4a37` with working diff fingerprint `47fd2495ce`, for AL-73 behavior and integration. Runtime model/effort metadata was not independently available; this is the assigned luna profile. Scope: Make release packaging/verification, the canonical setup bootstrap and bundle installer, and their contributor download entry point. No sibling review conclusions were read.

## Findings

No blocking behavior or integration defect found in the inspected paths.

## Evidence

- `release-package` creates archives from the configured platform list, then passes the corresponding archive and raw-binary inventory to `release-checksums.sh`. The helper checks every input, generates in a temporary directory on the same filesystem, validates the staged records against the complete inventory, and only then renames the manifest over `SHA256SUMS`. `set -euo pipefail` preserves hash, sort, and validation failures. A failed candidate therefore cannot replace a previous manifest.
- `release-verify` invokes the same helper with the Make inventory. The helper rejects malformed, unselected, duplicate, missing, or digest-mismatched records before Make reaches host probes. The probes are gated on the selected platform list and use only the selected host filenames from `bin`; a foreign-only selection cannot execute unrelated host files.
- `scripts/download-agentic-lens.sh` delegates to the bundled canonical downloader. The bundle installer validates the network repository override before any fetch and forms raw GitHub URLs with the exact selected version. Local source mode remains separate.
- Bootstrap selects exactly one well-formed SHA-256 entry for the selected manager name, downloads under its exact selected manager version, and verifies bytes before format or native protocol probes. Local manager and engine inputs require explicit digests. With `pipefail`, a hash utility that emits plausible output and exits nonzero fails digest admission. The accepted manager digest is carried into self-update; for an already installed manager the digest is computed with checked status.
- The script invokes engine installation before manager self-update and uses `set -e`; a later failure returns nonzero and suppresses its final installation summary while retaining already completed engine work. Manager and engine versions remain independently selected. This behavior is consistent with the approved per-component commit model.

## Hypotheses assessed and limits

I checked for incomplete inventory acceptance, host probes before verification, URL delimiter escape from exact tags, failed hash commands being treated as success, and completion text after a failed later component. The source paths and current test additions address those boundaries. The design records baseline probes where inert bytes reached the format check after a failing hash command and after a trailing manifest field; I did not independently rerun those probes. The requested focused suites were already running, and I did not launch tests. This is a source/diff review, not independent execution of the Make or installer fixtures; hosted release publication and actual cross-platform binary execution remain outside this review.

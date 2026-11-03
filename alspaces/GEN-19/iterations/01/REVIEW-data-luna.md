# GEN-19 data and failure-handling review

**Outcome: APPROVED** for the assigned data/failure-handling focus.

- Candidate fingerprint: `3a5ea9f89c48818862332b7da06019145544ec9ab2edcb830cae9a28b752765b` (baseline `036d808b0c77bec400f23ca77a5cf056a0b2078b`). The listed implementation/test file hashes in `CANDIDATE.json` match the inspected working tree.
- Scope: `scripts/release/rc-version.cjs`, its tests, and the release-candidate workflow gates, focusing on stale runs, no-candidate/API failure distinctions, stable target exclusion, and concurrency.
- Identity: requested `gpt-6-luna/medium`; actual runtime model metadata was unavailable, so this is not an assertion of observed model identity.
- Confidence: high for source-level resolver transitions and tested fixtures; hosted GitHub publication was not exercised.

## Findings

No blocking data-integrity or failure-handling findings (severity 0).

## Hypotheses investigated

- **A superseded branch run could calculate or publish for the wrong commit.** `resolveCandidate` compares the full expected SHA to `main` before reading configuration and again after `Manifest.buildPullRequests`; a mismatch returns an explicit `skip`. `verifyCandidate` repeats the SHA comparison before publication eligibility. At workflow level the candidate step gates all build steps, and final eligibility gates both publication and retention. The workflow concurrency group serializes runs without cancellation; the SHA checks cover runs that are overtaken while queued or executing. The deterministic tests exercise stale-before-read and stale-during-calculation cases.
- **No release-worthy changes could be confused with a GitHub/API failure.** An empty Release Please candidate list returns a `skip` with a distinct explanation. Errors from GitHub/configuration reads and malformed responses propagate as failures; the CLI sets a nonzero exit status and does not emit publish permission. Tests exercise no-candidate versus API rejection, and the adapter tests cover HTTP and malformed-response failures.
- **An already stable target could be reused.** Both resolution and final verification require the candidate target to exceed the manifest version and all non-draft, non-prerelease root-namespace stable GitHub Releases. Release enumeration is paginated; prereleases, drafts, and legacy component-prefixed namespaces are excluded. Tests cover an equal target and a newer stable release overtaking it.
- **Final read/publication ordering could guarantee exclusion under concurrent GitHub writes.** It cannot: the last eligibility read and the external release action are not atomic. This limitation is explicitly documented in the approved brief, which defines the final read as best-effort. The implementation performs that final recheck immediately before publication; no stronger guarantee is claimed.

## Verification and limits

- Reused parent-provided verification: `make test-release-workflow` passed (8 Python tests and 13 Node tests, zero skips; Node 20.20.2 in the additional raw verification and Node 24.22.1 in the recorded tester run). No tests were rerun in this review.
- `git diff --check` passed; candidate file hashes matched `CANDIDATE.json`.
- Static inspection confirms failures occur before output metadata/publish permission is written: the resolver awaits all required reads and validations before appending `GITHUB_ENV` or `GITHUB_OUTPUT`.
- No hosted publication, GitHub-side concurrency interleaving, or stable release action was exercised. This review does not claim hosted release behavior or a whole-repository approval.

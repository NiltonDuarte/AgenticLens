# Test review — AL-73

**Outcome: APPROVED** for the assigned test-effectiveness focus. This is a scoped review, not a whole-change or merge approval.

## Revision and evidence

Reviewed frozen candidate `b6ab0471224c8bcea7bc00adce867d3ba0de4a37` with the supplied working diff fingerprint `47fd2495ce`, against the approved AL-73 design, tasks and acceptance scenarios. Requested reviewer profile: gpt-6-luna/medium. Run ID was not supplied. Inspected `Makefile`, `scripts/release-checksums.sh`, `scripts/tests/test_platform_support.py`, `scripts/tests/test_installation.py`, `scripts/tests/distribution_fixture.py`, and the canonical setup scripts.

Executed successfully:

- `python3 -m unittest test_platform_support.ReleaseVerificationTest` from `scripts/tests` — 2 tests passed. These invoke the actual Make recipes against temporary release directories and assert probe order, manifest failures, and fresh/repeated manifest preservation.
- Six AL-73 installation tests from `scripts/tests`: malformed remote repositories, bootstrap malformed/ambiguous/missing/mismatched/hash-failed records, empty manager, local engine/installed manager hash failures, independent manager/engine versions, and failed update/retry state — all passed.

The installation fixtures compile benign native manager/engine binaries and exercise installer scripts with an isolated home, destination, project, controlled curl, and logged request URLs. The release tests use executable probe scripts and real packaging/verification recipes.

## Findings

No blocking test-oracle defect found. The new assertions reach the relevant production boundaries: Make packaging and verification; bootstrap and bundle remote entry points; checksum failures before format probing; and externally visible binaries, manager receipt/checksum, project pin, emitted completion claims, and reported versions. Failure injection is credible: the hash stubs either return failure after plausible output or partial output with success, and repeated packaging checks that the prior public manifest survives.

The host/foreign fixture establishes that foreign-only selection does not execute stale host files, while selected host verification probes manager and engine only after checksum checks. Existing platform-support tests separately assert package archive members and release upload inventory, retaining distinct distribution coverage.

## Limits

The added release verification cases exercise missing and directory inputs, but do not construct an unreadable regular file. Permission-based unreadability can be ineffective under privileged test users; the explicit unreadable-input branch therefore lacks a direct oracle. This is a narrow coverage omission, not an unresolved correctness concern: missing/non-regular inputs are covered and the production precondition checks both regular-file status and readability. No mutation test was run. Full suite, integrated `make`, and cross-platform execution were not part of this focused review.

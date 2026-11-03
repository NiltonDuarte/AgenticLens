# GEN-19 iteration 02: test effectiveness

Verdict: APPROVED within the assigned focus, with one nonblocking coverage observation. Independent discovery run `rc2_tests_sol`; requested profile gpt-6.1-sol/high (runtime identity not independently verified). Base `e639f3303c9d92236edce85befccda66617d7cb8`; candidate fingerprint `903c512467c9100cbb6b80756e501280efb9a3ab08f3466c0f4890cae92ece1e`. All frozen file hashes and deletions match CANDIDATE.json.

## Observation

TEST-02, severity 3/10, nonblocking: stable-release pagination has lost its durable regression oracle. The removed `GitHub adapter paginates stable releases and excludes draft, prerelease, and legacy namespaces` test exercised two response pages. In the replacement fixture, `scripts/tests/test_release_revision.py:151` supplies already-filtered release strings without requiring `--paginate`; the fixture at lines 170-180 never executes gojq. Removing `--paginate` from both production releases queries in a disposable copy leaves all 13 tests passing. Such a future regression could miss a target already published beyond the first API page. Current production queries correctly include pagination and stable/draft filters, so no present incorrect publication was demonstrated. A proportionate improvement is to require pagination in the releases fixture, while explicitly retaining the real-filter execution limit.

## Investigation and assertion mapping

The principal regression oracle is effective: lines 221-242 execute the actual metadata shell, distinguish the upcoming 2.1.0 from the 2.0.0 baseline, assert immutable PR-head reads, VERSION/commit/date/PR fields and `should_publish=true`, and feed emitted metadata into successful final eligibility. An isolated mutation forcing VERSION back to 2.0.0 fails at line 224. This closes the earlier successful-output oracle gap.

Deleted engine calculations and configuration/parser assertions belong to the superseded local calculator. Their remaining consumer contract maps to manifest-target variants 2.0.1, 2.1.0, 3.0.0, 4.0.0 and 2.10.0; these prove consumption, not Release Please's calculations. No-candidate/API separation maps to lines 244-277; stale-main checks before/after selection and generated-parent provenance are reached. Equal/newer stable rejection and final PR/main/API guards map to lines 258-294. Metadata injection rejection remains executable. The six stable revision tests remain unchanged. Transport implementation was removed in favor of gh; pagination/filter execution is the bounded observation above.

## Verification and limits

Independently ran `make test-release-workflow`: 13 passed, zero skips. Disposable mutations left shared sources untouched. Event/SHA/checkout and build/publish/prune wiring assertions inspect source; manual forcing and release-commit skipping execute shell. These do not establish hosted scheduling, GitHub filtering semantics, packaging/upload, or binary metadata. Full integrated `make` and publication remain parent responsibilities. No blocking correctness finding was established.

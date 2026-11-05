# AL-64 iteration 04 bounded behavior recheck

- **Verdict:** APPROVED for the bounded CP-02 correction review.
- **Reviewed base:** `d8a7c14cfd9f0ee626fedc33c70f92bd928d9192`.
- **Candidate:** full patch `/tmp/al64-04-review/candidate-02.diff`, SHA-256 `b426a55c735e39e4331ed5e1b03bd0095d7283c472bf64d484e2301b506486f9` (24 code paths); correction `/tmp/al64-04-review/correction.diff`, SHA-256 `c8cd7bf243b285d68b50d095538da34b9ddbaa68e4c230700fa73342f8848522` (exactly the three paths named in `correction-paths.json`).
- **Fingerprint:** all 411 entries in `fingerprint-02.json` matched their SHA-256 values in the reviewed candidate checkout. The manifest itself hashes to `512721e4192e55dcad98f2f1d7a080360178b73ce016c0225b50840dc69ff629`. The checkout's HEAD is the stated base; the candidate changes are staged/unstaged in the shared checkout and were preserved.
- **Scope:** behavior and integration of the corrected OpenCode v2 idle qualification, CLI generation/copy/success-only stderr interaction, known-short suppression, and the specified adjacent controls. No whole-feature or full-repository approval is implied.
- **Reviewer identity:** behavior/integration, gpt-6-luna / medium (runtime); reviewer run `dispatch_BEHAVIOR_LUNA_02.json`.

## Findings

No new distinct finding in the bounded correction. Existing CP-REV-01 and CP-TEST-01 are addressed by this candidate and are not re-reported as new findings.

## Hypotheses checked and evidence

- **Native v2 idle qualification:** the optional read in `schema_v2_session_extraction.go:121-124,479-508` is only performed when checkpoint evidence is requested and the column exists. Read/query/scan/close failures return unavailable evidence; they do not fail extraction. The numeric duration/token path remains separate. At `:286-293`, `time_idle` at or after the assistant-message start disables only `SourceOpen`; NULL and a valid positive idle timestamp earlier than this message preserve current-source qualification. An absent column preserves the earlier behavior. This correctly prevents a stale running-tool marker from producing exact F03 wording after native idle, including when idle is after the cutoff.
- **Known-short precedence and malformed evidence:** malformed, fractional, zero, negative, or unavailable optional idle metadata cannot establish source-open evidence and therefore remains eligible only for qualified subtotal F04. The feedback selector's existing interval-age/known-short guard remains ahead of subtotal fallback. The new test table checks age 179999 with both usable and unusable idle and expects silence, while maintaining active duration 180, provisional duration 180, and input tokens 17. Equal-start, before/equal/after-cutoff idle, absent-column, NULL, old-idle, and malformed controls are covered. This matches the approved CP-REV-01 boundary.
- **Superseding-message interaction:** the correction tests add a later user message beyond the operation cutoff to both schema families. Numeric duration and input remain preserved, while future source evidence qualifies to F04; the short case stays silent. The brief's original-source overlay and mutant evidence report the expected red controls, and the final focused tests pass. This is a test-only regression guard; no new product behavior is claimed.
- **CLI output/copy and generation success:** the full candidate calls selection and writes stderr only after successful publication/acknowledgement, and only for explicit generation (`main.go:696-713`). Rendering preserves the approved copy and escapes source identifiers/labels as single-line values (`checkpoint_feedback.go:12-50`). Generation tests keep feedback off stdout, assert no checkpoint message after publication failure, and require successful ledger/summary publication when stderr is an `os.Open` read-only descriptor; the new preflight confirms writes through that descriptor fail before running the subprocess (`checkpoint_generate_test.go:49-85`). This addresses the earlier privileged-user caveat by checking descriptor access mode directly.
- **Interactions:** no checkpoint fields enter durable DTOs, ledger/summary output, or frontend behavior in this correction. The OpenCode numeric extractor, running/DurationInterval rules, token facts and existing publication flow remain unchanged. Parent's full `make` and standalone embedded-asset gates are still pending and remain outside this verdict.

## Verification reused and limits

Reused final tester evidence in `iterations/04/TEST_REPORT.md`: focused race command PASS (OpenCode 1.713s; CLI 7.972s), full OpenCode race PASS (6.156s), plus 190 frontend public unit tests PASS. The same report records eight expected failures against the pre-idle-fix source, two expected failures for the future-supersession mutants, and the stderr descriptor preflight. No broad suite was repeated.

Native OpenCode caller identity/current-interval handoff and continuation remain unverified. Production F03/F04 fallback is within the approved contract; F01/F02/F05 native claims are not established. Parent owns integrated `make`, standalone binary checks, final consolidation and delivery.

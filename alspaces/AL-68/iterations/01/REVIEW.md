# AL-68 iteration 01 consolidated review

Scoped code verdict: APPROVED. Delivery verdict: APPROVED for the documented observed paths, with degraded/unverified coverage below. All twelve final focus/profile reviews approve `candidate-04` against baseline `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`. All twelve confirmed finding classes are fixed.

## Findings and discovery credit

[FINDINGS.json](FINDINGS.json) keeps validity, lifecycle, shared discovery and correction/verification links separate. Model settings below are observed in source metadata; dispatch acknowledgement remains recorded separately. Later correction investigations retain their actual re_review phase.

| Finding | Original issue | Discovering focus / observed model / effort | Status |
| --- | --- | --- | --- |
| CAP-01 | Stale registry snapshots override newer canonical or freshly extracted facts | behavior / gpt-6.1-sol/high (behavior-sol-01, discovery); data / gpt-6.1-sol/high (data-sol-01, discovery); architecture / gpt-6.1-sol/high (architecture-sol-01, discovery); tests / gpt-6.1-sol/high (tests-sol-01, discovery) | fixed |
| CAP-02 | Concurrent generation publishes older snapshots after newer ones | data / gpt-6.1-sol/high (data-sol-01, discovery) | fixed |
| CAP-03 | Explicit generation bypasses known conflicting ticket binding | data / gpt-6.1-sol/high (data-sol-01, discovery) | fixed |
| CAP-04 | Codex child start evidence is discarded until stop callback | behavior / gpt-6.1-sol/high (behavior-sol-01, discovery) | fixed |
| CAP-05 | Failed source reads evade per-pass byte budget | architecture / gpt-6.1-sol/high (architecture-sol-01, discovery) | fixed |
| CAP-06 | Repeated placeholder substitution corrupts valid escaped plugin paths | security / gpt-6.1-sol/high (security-sol-01, discovery) | fixed |
| CAP-07 | Native receive timeout does not bound blocked stdin | security / gpt-6.1-sol/high (security-sol-01, discovery) | fixed |
| CAP-08 | Incompatible binary replacement strands active lifecycle hooks | release / gpt-6.1-sol/high (release-sol-01, discovery) | fixed |
| CAP-09 | Public manual direct and PowerShell setup lacks lifecycle integration | release / gpt-6.1-sol/high (release-sol-01, discovery); release / gpt-6-luna/medium (release-luna-01, discovery) | fixed |
| CAP-10 | Reported snapshot cutoff precedes source reads and can misstate captured boundary | data / gpt-6-luna/medium (data-luna-01, discovery) | fixed |
| CAP-11 | Unrelated unassigned sessions consume ticket refresh budget and leak into ticket report | behavior / gpt-6-luna/medium (behavior-luna-01, discovery) | fixed |
| CAP-12 | Child recovery lacks fair bounded continuation for known OpenCode pages and failed Codex metadata queues | behavior / gpt-6.1-sol/high (behavior-sol-02, re_review); data / gpt-6.1-sol/high (data-sol-02, re_review); architecture / gpt-6.1-sol/high (architecture-sol-02, re_review); data / gpt-6.1-sol/high (data-sol-03, re_review); architecture / gpt-6.1-sol/high (architecture-sol-03, re_review) | fixed |

## Verification and runtime boundaries

Integrated make passed at implementation commit `9db4672a971709a7433eeee4e11a5d7d0817cbbe` and final integration commit `ba0b50a0d7f64561b21a1d11cd5bf3d909dfe5d7`: Go vet/race, frontend typecheck/lint/build, 781 frontend tests in 35 files, 14 distribution tests and native builds. Affected candidate04 race checks passed; unchanged component evidence carries forward. This consolidation changes documentation only.

After the user completed normal trust review, actual Codex 0.166.0 Linux exec/resume verified callback/context/request/agent-bind exchanges, two independent roots, a completed native child and exact source/model/usage/parent comparison. OpenCode 1.19.5 Linux verified actual binding, independent roots/native child, source comparison and unchanged repeats. Unassigned work stayed excluded; no trust bypass was used. Archived SMOKE.md, CODEX_SMOKE.json, VERIFICATION.md and DELIVERY_REVISION.json contain the full evidence.

Codex SubagentStart was not observed on the exec path; running/interrupted child coverage before Stop remains degraded/unverified. Its cause is unknown. Other versions/interfaces, OpenCode v2 and native Windows/macOS remain unverified. Final reviewer reports' earlier INCOMPLETE runtime verdicts remain historical; parent closed the required observed-path evidence gate with the subsequent trusted smoke, without waiving code findings or review cells.

Ticket completed_at remains unknown while the orchestration is active. The legacy generator's invocation/publication-time inference is removed only from ticket metadata; source facts are preserved and the capture report digest is updated. This existing limitation is recorded rather than reported as a product fix.

## Complete run roster

[REVIEW_RUNS.json](REVIEW_RUNS.json) contains every run, full dispatch/source metadata and readable candidate references. Each produced report resolves via report_pointer to archive.reports[run_id].content. Cancelled cells retain original planned paths and explicit absent-report records. No absent report is fabricated.

| Run | Persona | Observed model / effort | Phase | Candidate | Execution / historical verdict | Evidence |
| --- | --- | --- | --- | --- | --- | --- |
| behavior-sol-01 | reviewer-behavior | gpt-6.1-sol/high | discovery | candidate-01 | completed / CHANGES_REQUESTED | [archive](REVIEW_RUNS.json#/archive/reports/behavior-sol-01/content) |
| behavior-luna-01 | reviewer-behavior | gpt-6-luna/medium | discovery | candidate-01 | completed / CHANGES_REQUESTED | [archive](REVIEW_RUNS.json#/archive/reports/behavior-luna-01/content) |
| data-sol-01 | reviewer-data | gpt-6.1-sol/high | discovery | candidate-01 | completed / CHANGES_REQUESTED | [archive](REVIEW_RUNS.json#/archive/reports/data-sol-01/content) |
| data-luna-01 | reviewer-data | gpt-6-luna/medium | discovery | candidate-01 | completed / CHANGES_REQUESTED | [archive](REVIEW_RUNS.json#/archive/reports/data-luna-01/content) |
| architecture-sol-01 | reviewer-architecture | gpt-6.1-sol/high | discovery | candidate-01 | completed / CHANGES_REQUESTED | [archive](REVIEW_RUNS.json#/archive/reports/architecture-sol-01/content) |
| architecture-luna-01 | reviewer-architecture | gpt-6-luna/medium | discovery | candidate-01 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/architecture-luna-01/content) |
| tests-sol-01 | reviewer-tests | gpt-6.1-sol/high | discovery | candidate-01 | completed / CHANGES_REQUESTED | [archive](REVIEW_RUNS.json#/archive/reports/tests-sol-01/content) |
| tests-luna-01 | reviewer-tests | gpt-6-luna/medium | discovery | candidate-01 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/tests-luna-01/content) |
| security-sol-01 | reviewer | gpt-6.1-sol/high | discovery | candidate-01 | completed / CHANGES_REQUESTED | [archive](REVIEW_RUNS.json#/archive/reports/security-sol-01/content) |
| security-luna-01 | reviewer | gpt-6-luna/medium | discovery | candidate-01 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/security-luna-01/content) |
| release-sol-01 | reviewer | gpt-6.1-sol/high | discovery | candidate-01 | completed / CHANGES_REQUESTED | [archive](REVIEW_RUNS.json#/archive/reports/release-sol-01/content) |
| release-luna-01 | reviewer | gpt-6-luna/medium | discovery | candidate-01 | completed / CHANGES_REQUESTED | [archive](REVIEW_RUNS.json#/archive/reports/release-luna-01/content) |
| behavior-sol-02 | reviewer-behavior | gpt-6.1-sol/high | re_review | candidate-02 | completed / CHANGES_REQUESTED | [archive](REVIEW_RUNS.json#/archive/reports/behavior-sol-02/content) |
| behavior-luna-02 | reviewer-behavior | unavailable; requested gpt-6-luna/medium | re_review | candidate-02 | cancelled / INCOMPLETE | not produced |
| data-sol-02 | reviewer-data | gpt-6.1-sol/high | re_review | candidate-02 | completed / CHANGES_REQUESTED | [archive](REVIEW_RUNS.json#/archive/reports/data-sol-02/content) |
| data-luna-02 | reviewer-data | unavailable; requested gpt-6-luna/medium | re_review | candidate-02 | cancelled / INCOMPLETE | not produced |
| architecture-sol-02 | reviewer-architecture | gpt-6.1-sol/high | re_review | candidate-02 | completed / CHANGES_REQUESTED | [archive](REVIEW_RUNS.json#/archive/reports/architecture-sol-02/content) |
| architecture-luna-02 | reviewer-architecture | unavailable; requested gpt-6-luna/medium | re_review | candidate-02 | cancelled / INCOMPLETE | not produced |
| tests-sol-02 | reviewer-tests | unavailable; requested gpt-6.1-sol/high | re_review | candidate-02 | cancelled / INCOMPLETE | not produced |
| tests-luna-02 | reviewer-tests | unavailable; requested gpt-6-luna/medium | re_review | candidate-02 | cancelled / INCOMPLETE | not produced |
| security-sol-02 | reviewer | unavailable; requested gpt-6.1-sol/high | re_review | candidate-02 | cancelled / INCOMPLETE | not produced |
| security-luna-02 | reviewer | unavailable; requested gpt-6-luna/medium | re_review | candidate-02 | cancelled / INCOMPLETE | not produced |
| release-sol-02 | reviewer | unavailable; requested gpt-6.1-sol/high | re_review | candidate-02 | cancelled / INCOMPLETE | not produced |
| release-luna-02 | reviewer | unavailable; requested gpt-6-luna/medium | re_review | candidate-02 | cancelled / INCOMPLETE | not produced |
| behavior-sol-03 | reviewer-behavior | gpt-6.1-sol/high | re_review | candidate-03 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/behavior-sol-03/content) |
| behavior-luna-03 | reviewer-behavior | unavailable; requested gpt-6-luna/medium | re_review | candidate-03 | cancelled / INCOMPLETE | not produced |
| data-sol-03 | reviewer-data | gpt-6.1-sol/high | re_review | candidate-03 | completed / CHANGES_REQUESTED | [archive](REVIEW_RUNS.json#/archive/reports/data-sol-03/content) |
| data-luna-03 | reviewer-data | unavailable; requested gpt-6-luna/medium | re_review | candidate-03 | cancelled / INCOMPLETE | not produced |
| architecture-sol-03 | reviewer-architecture | gpt-6.1-sol/high | re_review | candidate-03 | completed / CHANGES_REQUESTED | [archive](REVIEW_RUNS.json#/archive/reports/architecture-sol-03/content) |
| architecture-luna-03 | reviewer-architecture | unavailable; requested gpt-6-luna/medium | re_review | candidate-03 | cancelled / INCOMPLETE | not produced |
| tests-sol-03 | reviewer-tests | unavailable; requested gpt-6.1-sol/high | re_review | candidate-03 | cancelled / INCOMPLETE | not produced |
| tests-luna-03 | reviewer-tests | unavailable; requested gpt-6-luna/medium | re_review | candidate-03 | cancelled / INCOMPLETE | not produced |
| security-sol-03 | reviewer | unavailable; requested gpt-6.1-sol/high | re_review | candidate-03 | cancelled / INCOMPLETE | not produced |
| security-luna-03 | reviewer | unavailable; requested gpt-6-luna/medium | re_review | candidate-03 | cancelled / INCOMPLETE | not produced |
| release-sol-03 | reviewer | unavailable; requested gpt-6.1-sol/high | re_review | candidate-03 | cancelled / INCOMPLETE | not produced |
| release-luna-03 | reviewer | unavailable; requested gpt-6-luna/medium | re_review | candidate-03 | cancelled / INCOMPLETE | not produced |
| behavior-sol-04 | reviewer-behavior | gpt-6.1-sol/high | re_review | candidate-04 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/behavior-sol-04/content) |
| behavior-luna-04 | reviewer-behavior | gpt-6-luna/medium | re_review | candidate-04 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/behavior-luna-04/content) |
| data-sol-04 | reviewer-data | gpt-6.1-sol/high | re_review | candidate-04 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/data-sol-04/content) |
| data-luna-04 | reviewer-data | gpt-6-luna/medium | re_review | candidate-04 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/data-luna-04/content) |
| architecture-sol-04 | reviewer-architecture | gpt-6.1-sol/high | re_review | candidate-04 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/architecture-sol-04/content) |
| architecture-luna-04 | reviewer-architecture | gpt-6-luna/medium | re_review | candidate-04 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/architecture-luna-04/content) |
| tests-sol-04 | reviewer-tests | gpt-6.1-sol/high | re_review | candidate-04 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/tests-sol-04/content) |
| tests-luna-04 | reviewer-tests | gpt-6-luna/medium | re_review | candidate-04 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/tests-luna-04/content) |
| security-sol-04 | reviewer | gpt-6.1-sol/high | re_review | candidate-04 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/security-sol-04/content) |
| security-luna-04 | reviewer | gpt-6-luna/medium | re_review | candidate-04 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/security-luna-04/content) |
| release-sol-04 | reviewer | gpt-6.1-sol/high | re_review | candidate-04 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/release-sol-04/content) |
| release-luna-04 | reviewer | gpt-6-luna/medium | re_review | candidate-04 | completed / INCOMPLETE | [archive](REVIEW_RUNS.json#/archive/reports/release-luna-04/content) |

## Archive integrity

The workspace was reduced from 56 to nine files after byte-for-byte UTF-8/SHA-256 verification of all 47 removed bodies: 30 per-run reports and 17 supporting documents. Original consolidated records are retained under archive.retained_versions. Original local IDs, claims, severities, commands, dissent, no-findings results and all attribution remain in the complete original bodies. archive.reference_index resolves original filenames and relative/repository paths to the archived content.

Validation compares all 48 original run records and all 12 findings, allowing only explicit reference relocation fields; checks every current JSON pointer and the ledger/report digest; and validates the edited development/review skill metadata. The canonical model policy and application/test/runtime files are unchanged. File consolidation preserves evidence size; it reduces the number of committed artifacts.

## Review bookkeeping simplification

The user requested removal of review hashes after their purpose was explained. Current structured records use candidate-01 through candidate-04, which resolve to the original archived manifests, and direct file/JSON-pointer report references. Per-report, policy, persona, procedure and prompt hashes are removed. Original archived bodies remain byte-for-byte unchanged, including historical hashes and verdicts. Source commit IDs remain ordinary version references. The runtime capture_report.json ledger digest remains intact; its duplicate review-record copy is removed.

Validation compares every archived body directly with its pre-cleanup bytes and verifies all reviewer/model/phase/outcome, finding/discovery/verification and candidate associations. The nine-file layout, code/test evidence, model policy and runtime acceptance limits are unchanged.

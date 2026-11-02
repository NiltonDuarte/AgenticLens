# AL-68 iteration 01 retained review summary

Current scoped outcome: APPROVED. Historical changes-requested, incomplete, cancelled and refuted outcomes remain distinguished below; final approval does not rewrite earlier verdicts.

## Findings and reviewer credit

This review retains findings, validity, correction status and reviewer/model discovery and verification credit. The execution roster distinguishes requested from source-observed settings; detailed historical dispatch records remain in Git history.

| Finding | Issue | Disposition / lifecycle | Discovering reviewer and model/effort | Verifying reviewer and model/effort |
| --- | --- | --- | --- | --- |
| CAP-01 | Stale registry snapshots override newer canonical or freshly extracted facts | confirmed / fixed | behavior / gpt-6.1-sol/high; data / gpt-6.1-sol/high; architecture / gpt-6.1-sol/high; tests / gpt-6.1-sol/high | behavior / gpt-6.1-sol/high; behavior / gpt-6-luna/medium; data / gpt-6.1-sol/high; data / gpt-6-luna/medium; architecture / gpt-6.1-sol/high; architecture / gpt-6-luna/medium; tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium |
| CAP-02 | Concurrent generation publishes older snapshots after newer ones | confirmed / fixed | data / gpt-6.1-sol/high | data / gpt-6.1-sol/high; data / gpt-6-luna/medium; architecture / gpt-6.1-sol/high; architecture / gpt-6-luna/medium; tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium |
| CAP-03 | Explicit generation bypasses known conflicting ticket binding | confirmed / fixed | data / gpt-6.1-sol/high | behavior / gpt-6.1-sol/high; behavior / gpt-6-luna/medium; data / gpt-6.1-sol/high; data / gpt-6-luna/medium; tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium; security / gpt-6.1-sol/high; security / gpt-6-luna/medium |
| CAP-04 | Codex child start evidence is discarded until stop callback | confirmed / fixed | behavior / gpt-6.1-sol/high | behavior / gpt-6.1-sol/high; behavior / gpt-6-luna/medium; architecture / gpt-6.1-sol/high; architecture / gpt-6-luna/medium; tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium |
| CAP-05 | Failed source reads evade per-pass byte budget | confirmed / fixed | architecture / gpt-6.1-sol/high | data / gpt-6.1-sol/high; data / gpt-6-luna/medium; architecture / gpt-6.1-sol/high; architecture / gpt-6-luna/medium; tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium |
| CAP-06 | Repeated placeholder substitution corrupts valid escaped plugin paths | confirmed / fixed | security / gpt-6.1-sol/high | architecture / gpt-6.1-sol/high; architecture / gpt-6-luna/medium; tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium; security / gpt-6.1-sol/high; security / gpt-6-luna/medium |
| CAP-07 | Native receive timeout does not bound blocked stdin | confirmed / fixed | security / gpt-6.1-sol/high | architecture / gpt-6.1-sol/high; architecture / gpt-6-luna/medium; tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium; security / gpt-6.1-sol/high; security / gpt-6-luna/medium |
| CAP-08 | Incompatible binary replacement strands active lifecycle hooks | confirmed / fixed | release / gpt-6.1-sol/high | tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium; release / gpt-6.1-sol/high; release / gpt-6-luna/medium |
| CAP-09 | Public manual direct and PowerShell setup lacks lifecycle integration | confirmed / fixed | release / gpt-6.1-sol/high; release / gpt-6-luna/medium | release / gpt-6.1-sol/high; release / gpt-6-luna/medium |
| CAP-10 | Reported snapshot cutoff precedes source reads and can misstate captured boundary | confirmed / fixed | data / gpt-6-luna/medium | data / gpt-6.1-sol/high; data / gpt-6-luna/medium; architecture / gpt-6.1-sol/high; architecture / gpt-6-luna/medium; tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium |
| CAP-11 | Unrelated unassigned sessions consume ticket refresh budget and leak into ticket report | confirmed / fixed | behavior / gpt-6-luna/medium | behavior / gpt-6.1-sol/high; behavior / gpt-6-luna/medium; data / gpt-6.1-sol/high; data / gpt-6-luna/medium; architecture / gpt-6.1-sol/high; architecture / gpt-6-luna/medium; tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium |
| CAP-12 | Child recovery lacks fair bounded continuation for known OpenCode pages and failed Codex metadata queues | confirmed / fixed | behavior / gpt-6.1-sol/high; data / gpt-6.1-sol/high; architecture / gpt-6.1-sol/high | behavior / gpt-6.1-sol/high; behavior / gpt-6-luna/medium; data / gpt-6.1-sol/high; data / gpt-6-luna/medium; architecture / gpt-6.1-sol/high; architecture / gpt-6-luna/medium; tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium |

## Finding evidence and resolutions

### CAP-01

Severity: P1; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/cmd/agentic-lens/main.go:312.

Disposition evidence: Source trace and isolated native production-path reproduction documented in assigned reports; public manual gap verified against approved installation scope.

Resolution: Corrected production paths and meaningful regression oracles verified by current applicable focused reviews; historical confirmed disposition retained.

Verified candidate: candidate-04.

### CAP-02

Severity: P1; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/cmd/agentic-lens/main.go.

Disposition evidence: Source trace and isolated native production-path reproduction documented in assigned reports; public manual gap verified against approved installation scope.

Resolution: Corrected production paths and meaningful regression oracles verified by current applicable focused reviews; historical confirmed disposition retained.

Verified candidate: candidate-04.

### CAP-03

Severity: P1; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/cmd/agentic-lens/main.go:201.

Disposition evidence: Source trace and isolated native production-path reproduction documented in assigned reports; public manual gap verified against approved installation scope.

Resolution: Corrected production paths and meaningful regression oracles verified by current applicable focused reviews; historical confirmed disposition retained.

Verified candidate: candidate-04.

### CAP-04

Severity: P2; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/capture/hooks.go:75.

Disposition evidence: Source trace and isolated native production-path reproduction documented in assigned reports; public manual gap verified against approved installation scope.

Resolution: Corrected production paths and meaningful regression oracles verified by current applicable focused reviews; historical confirmed disposition retained.

Verified candidate: candidate-04.

### CAP-05

Severity: P2; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/capture/reconcile.go:133.

Disposition evidence: Source trace and isolated native production-path reproduction documented in assigned reports; public manual gap verified against approved installation scope.

Resolution: Corrected production paths and meaningful regression oracles verified by current applicable focused reviews; historical confirmed disposition retained.

Verified candidate: candidate-04.

### CAP-06

Severity: P2; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/capture/install.go:284.

Disposition evidence: Source trace and isolated native production-path reproduction documented in assigned reports; public manual gap verified against approved installation scope.

Resolution: Corrected production paths and meaningful regression oracles verified by current applicable focused reviews; historical confirmed disposition retained.

Verified candidate: candidate-04.

### CAP-07

Severity: P2; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/capture/hooks.go:31.

Disposition evidence: Source trace and isolated native production-path reproduction documented in assigned reports; public manual gap verified against approved installation scope.

Resolution: Corrected production paths and meaningful regression oracles verified by current applicable focused reviews; historical confirmed disposition retained.

Verified candidate: candidate-04.

### CAP-08

Severity: P1; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: scripts/download-agentic-lens.sh.

Disposition evidence: Source trace and isolated native production-path reproduction documented in assigned reports; public manual gap verified against approved installation scope.

Resolution: Corrected production paths and meaningful regression oracles verified by current applicable focused reviews; historical confirmed disposition retained.

Verified candidate: candidate-04.

### CAP-09

Severity: P2; classification: required_scope_gap; disposition: confirmed; lifecycle: fixed. Location: docs/SETUP_MANUAL.md:81.

Disposition evidence: Source trace and isolated native production-path reproduction documented in assigned reports; public manual gap verified against approved installation scope.

Resolution: Corrected production paths and meaningful regression oracles verified by current applicable focused reviews; historical confirmed disposition retained.

Verified candidate: candidate-04.

### CAP-10

Severity: P2; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/capture/reconcile.go:90.

Disposition evidence: Source-level trace establishes cutoff assigned before later independent source reads; no common as-of snapshot is enforced. Define truthful completion/per-owner checkpoint semantics.

Resolution: Corrected production paths and meaningful regression oracles verified by current applicable focused reviews; historical confirmed disposition retained.

Verified candidate: candidate-04.

### CAP-11

Severity: P2; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/capture/reconcile.go:109.

Disposition evidence: Selection/extraction and final report include every empty-ticket registration; keep local discovery but scope ticket work and publication to proven owners.

Resolution: Corrected production paths and meaningful regression oracles verified by current applicable focused reviews; historical confirmed disposition retained.

Verified candidate: candidate-04.

### CAP-12

Severity: P2; classification: correction_introduced; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/capture/children.go.

Disposition evidence: Three independent reviewer probes reproduce permanent first-page starvation; report-specific native/overlay evidence retained.

Resolution: Corrected production paths and meaningful regression oracles verified by current applicable focused reviews; historical confirmed disposition retained.

Verified candidate: candidate-04.

## Verification and remaining limits

Integrated make passed at implementation commit `9db4672a971709a7433eeee4e11a5d7d0817cbbe` and final integration commit `ba0b50a0d7f64561b21a1d11cd5bf3d909dfe5d7`: Go vet/race, frontend typecheck/lint/build, 781 frontend tests in 35 files, 14 distribution tests and native builds. Affected candidate04 race checks passed; unchanged component evidence carries forward. This consolidation changes documentation only.

After the user completed normal trust review, actual Codex 0.166.0 Linux exec/resume verified callback/context/request/agent-bind exchanges, two independent roots, a completed native child and exact source/model/usage/parent comparison. OpenCode 1.19.5 Linux verified actual binding, independent roots/native child, source comparison and unchanged repeats. Unassigned work stayed excluded; no trust bypass was used. Archived SMOKE.md, CODEX_SMOKE.json, VERIFICATION.md and DELIVERY_REVISION.json contain the full evidence.

Codex SubagentStart was not observed on the exec path; running/interrupted child coverage before Stop remains degraded/unverified. Its cause is unknown. Other versions/interfaces, OpenCode v2 and native Windows/macOS remain unverified. Final reviewer reports' earlier INCOMPLETE runtime verdicts remain historical; parent closed the required observed-path evidence gate with the subsequent trusted smoke, without waiving code findings or review cells.

Historical iteration01 artifacts cleared top-level completed_at while the root was active. Iteration02 superseded that interpretation: completed_at records observability snapshot finalization, and source invocation timestamps remain source-derived. The earlier metadata correction is historical behavior, not a current requirement.

## Review execution history

Every original roster cell is retained in this compact table, including cancelled and no-findings runs. Requested identity is not presented as observed identity.

| Run | Reviewer | Requested model/effort | Observed model/effort | Phase | Candidate | Execution / verdict |
| --- | --- | --- | --- | --- | --- | --- |
| behavior-sol-01 | behavior | gpt-6.1-sol/high | gpt-6.1-sol/high | discovery | candidate-01 | completed / CHANGES_REQUESTED |
| behavior-luna-01 | behavior | gpt-6-luna/medium | gpt-6-luna/medium | discovery | candidate-01 | completed / CHANGES_REQUESTED |
| data-sol-01 | data | gpt-6.1-sol/high | gpt-6.1-sol/high | discovery | candidate-01 | completed / CHANGES_REQUESTED |
| data-luna-01 | data | gpt-6-luna/medium | gpt-6-luna/medium | discovery | candidate-01 | completed / CHANGES_REQUESTED |
| architecture-sol-01 | architecture | gpt-6.1-sol/high | gpt-6.1-sol/high | discovery | candidate-01 | completed / CHANGES_REQUESTED |
| architecture-luna-01 | architecture | gpt-6-luna/medium | gpt-6-luna/medium | discovery | candidate-01 | completed / INCOMPLETE |
| tests-sol-01 | tests | gpt-6.1-sol/high | gpt-6.1-sol/high | discovery | candidate-01 | completed / CHANGES_REQUESTED |
| tests-luna-01 | tests | gpt-6-luna/medium | gpt-6-luna/medium | discovery | candidate-01 | completed / INCOMPLETE |
| security-sol-01 | security | gpt-6.1-sol/high | gpt-6.1-sol/high | discovery | candidate-01 | completed / CHANGES_REQUESTED |
| security-luna-01 | security | gpt-6-luna/medium | gpt-6-luna/medium | discovery | candidate-01 | completed / INCOMPLETE |
| release-sol-01 | release | gpt-6.1-sol/high | gpt-6.1-sol/high | discovery | candidate-01 | completed / CHANGES_REQUESTED |
| release-luna-01 | release | gpt-6-luna/medium | gpt-6-luna/medium | discovery | candidate-01 | completed / CHANGES_REQUESTED |
| behavior-sol-02 | behavior | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02 | completed / CHANGES_REQUESTED |
| behavior-luna-02 | behavior | gpt-6-luna/medium | unavailable | re_review | candidate-02 | cancelled / INCOMPLETE |
| data-sol-02 | data | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02 | completed / CHANGES_REQUESTED |
| data-luna-02 | data | gpt-6-luna/medium | unavailable | re_review | candidate-02 | cancelled / INCOMPLETE |
| architecture-sol-02 | architecture | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02 | completed / CHANGES_REQUESTED |
| architecture-luna-02 | architecture | gpt-6-luna/medium | unavailable | re_review | candidate-02 | cancelled / INCOMPLETE |
| tests-sol-02 | tests | gpt-6.1-sol/high | unavailable | re_review | candidate-02 | cancelled / INCOMPLETE |
| tests-luna-02 | tests | gpt-6-luna/medium | unavailable | re_review | candidate-02 | cancelled / INCOMPLETE |
| security-sol-02 | security | gpt-6.1-sol/high | unavailable | re_review | candidate-02 | cancelled / INCOMPLETE |
| security-luna-02 | security | gpt-6-luna/medium | unavailable | re_review | candidate-02 | cancelled / INCOMPLETE |
| release-sol-02 | release | gpt-6.1-sol/high | unavailable | re_review | candidate-02 | cancelled / INCOMPLETE |
| release-luna-02 | release | gpt-6-luna/medium | unavailable | re_review | candidate-02 | cancelled / INCOMPLETE |
| behavior-sol-03 | behavior | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-03 | completed / INCOMPLETE |
| behavior-luna-03 | behavior | gpt-6-luna/medium | unavailable | re_review | candidate-03 | cancelled / INCOMPLETE |
| data-sol-03 | data | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-03 | completed / CHANGES_REQUESTED |
| data-luna-03 | data | gpt-6-luna/medium | unavailable | re_review | candidate-03 | cancelled / INCOMPLETE |
| architecture-sol-03 | architecture | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-03 | completed / CHANGES_REQUESTED |
| architecture-luna-03 | architecture | gpt-6-luna/medium | unavailable | re_review | candidate-03 | cancelled / INCOMPLETE |
| tests-sol-03 | tests | gpt-6.1-sol/high | unavailable | re_review | candidate-03 | cancelled / INCOMPLETE |
| tests-luna-03 | tests | gpt-6-luna/medium | unavailable | re_review | candidate-03 | cancelled / INCOMPLETE |
| security-sol-03 | security | gpt-6.1-sol/high | unavailable | re_review | candidate-03 | cancelled / INCOMPLETE |
| security-luna-03 | security | gpt-6-luna/medium | unavailable | re_review | candidate-03 | cancelled / INCOMPLETE |
| release-sol-03 | release | gpt-6.1-sol/high | unavailable | re_review | candidate-03 | cancelled / INCOMPLETE |
| release-luna-03 | release | gpt-6-luna/medium | unavailable | re_review | candidate-03 | cancelled / INCOMPLETE |
| behavior-sol-04 | behavior | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-04 | completed / INCOMPLETE |
| behavior-luna-04 | behavior | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-04 | completed / INCOMPLETE |
| data-sol-04 | data | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-04 | completed / INCOMPLETE |
| data-luna-04 | data | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-04 | completed / INCOMPLETE |
| architecture-sol-04 | architecture | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-04 | completed / INCOMPLETE |
| architecture-luna-04 | architecture | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-04 | completed / INCOMPLETE |
| tests-sol-04 | tests | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-04 | completed / INCOMPLETE |
| tests-luna-04 | tests | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-04 | completed / INCOMPLETE |
| security-sol-04 | security | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-04 | completed / INCOMPLETE |
| security-luna-04 | security | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-04 | completed / INCOMPLETE |
| release-sol-04 | release | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-04 | completed / INCOMPLETE |
| release-luna-04 | release | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-04 | completed / INCOMPLETE |

## Evidence retention

The two large report/log archives were removed from the current tree under the revised retention requirement. This iteration retains 12 findings and all 48 review outcomes with shared reviewer/model discovery and separate verification credit. Full original reports, candidate manifests and raw build logs remain available in [historical committed evidence](https://github.com/NiltonDuarte/AgenticLens/blob/0e9cb88c24f8a10eb1f35840c1ce23223efdd444/alspaces/AL-68/iterations/01/REVIEW_RUNS.json); they are no longer replicated in active workspace files. Original harness transcripts remain outside Git.

The structured finding files were subsequently removed after transferring their substantive finding evidence and resolutions into this review. Reviewer/model credit and original execution outcomes remain above; detailed original metadata remains available in Git history before this cleanup.

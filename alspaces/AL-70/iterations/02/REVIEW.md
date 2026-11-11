# AL-70 iteration 02: Review corrections in progress

Candidate AL70-R1: base f63e3db6b2d4baa8e22452e4ec94eff566867721; product/test/active-guidance and embedded-input fingerprint f5e45930942e860629a2c981f7e6c927fd0b6db9d4508a6d867be9b90cf78d91. Raw manifest, tracked diff, untracked copies, frozen requirements and coordinator run/finding records are under /tmp/al70-review-r1-7c7_i1yk. All eight first passes completed against this frozen candidate; product/test writes are now open only to assigned correction owners. Task progress/report files are excluded from the product fingerprint.

The original confirmed scope and concrete values are approved. A12 is a pending source-label clarification; existing producer aggregation/counting is unchanged. The integrated make gate and final consolidated approval remain pending. Native dispatch accepted each explicitly selected model/effort; execution has not been independently observed in source metadata. No model usage/cost claims are made.

| Focus | Model / effort | Phase | Status | Report |
| --- | --- | --- | --- | --- |
| behavior | gpt-6.1-sol / high | discovery | CHANGES_REQUESTED | [review-behavior-sol.md](review-behavior-sol.md) |
| behavior | gpt-6-luna / medium | discovery | APPROVED | [review-behavior-luna.md](review-behavior-luna.md) |
| data | gpt-6.1-sol / high | discovery | CHANGES_REQUESTED | [review-data-sol.md](review-data-sol.md) |
| data | gpt-6-luna / medium | discovery | APPROVED | [review-data-luna.md](review-data-luna.md) |
| architecture | gpt-6.1-sol / high | discovery | CHANGES_REQUESTED | [review-architecture-sol.md](review-architecture-sol.md) |
| architecture | gpt-6-luna / medium | discovery | CHANGES_REQUESTED | [review-architecture-luna.md](review-architecture-luna.md) |
| tests | gpt-6.1-sol / high | discovery | CHANGES_REQUESTED | [review-tests-sol.md](review-tests-sol.md) |
| tests | gpt-6-luna / medium | discovery | CHANGES_REQUESTED | [review-tests-luna.md](review-tests-luna.md) |

All four focuses completed both configured profiles independently against AL70-R1. The initial consolidated outcome is CHANGES_REQUESTED. Native dispatch acknowledged the requested settings; source-session execution identity and model usage/cost remain unverified. Corrections require a new frozen candidate and affected-focus review under both profiles before integration.

The parent validated the following distinct findings against requirements, source paths and raw counterexamples. Original reports retain evidence and dissent; verification does not confer discovery credit.

| Finding | Discovering role / model / effort | Correction owner and disposition |
| --- | --- | --- |
| CFG-01: dotted YAML keys can impersonate structural setting presence | Architecture / gpt-6.1-sol / high | Backend: structural presence authority; open |
| CAP-01: canonical YML projects bypass Codex callback routing | Architecture / gpt-6.1-sol / high | Backend: common YAML/YML routing; open |
| UI-01: live roles are mapped twice | Architecture / gpt-6.1-sol / high | Frontend/backend: one projection and shared wire guard; open |
| PATH-01: newline-delimited Git metadata truncates valid checkout paths | Architecture / gpt-6.1-sol / high | Backend: lossless worktree metadata; open |
| DATA-01: an unreadable unrelated directory blocks valid batch siblings | Data / gpt-6.1-sol / high | Backfill: scoped failure reporting; open |
| CAP-02: worker budget overrides are lost across publication children | Behavior / gpt-6.1-sol / high | Backend: preserve admitted effective budgets; open |
| DOC-01: operator guidance promises removed budget presets | Behavior / gpt-6.1-sol / high | Parent: corrected to explicit CLI/local precedence; re-review pending |
| DOC-02: dashboard guidance lists a retired source-path key | Architecture / gpt-6-luna / medium | Parent: corrected to current CLI/selected-local paths; re-review pending |
| LIVE-01: an unavailable source is presented as measured zero activity | Behavior / gpt-6.1-sol / high | Backend/frontend: distinguish unavailable from available-empty; open |
| UI-02: empty/whitespace artifact content is presented as missing | Behavior / gpt-6.1-sol / high | Frontend: bounded nonblocking correction accepted; open |
| TEST-01: component tests bypass real App configuration handoff | Tests / gpt-6.1-sol / high | Frontend: focused real App guard; open |
| TEST-02: HTTP comparison test expects rewriting a preserved historical type | Tests / gpt-6.1-sol / high | Backend: correct expectation, retain alias filtering; open |
| TEST-03: a no-write test silently omits its assertion | Tests / gpt-6.1-sol / high | Backend: self-contained unconditional sentinel; open |
| TEST-04: a hand-authored frontend fixture is described as a native producer join | Tests / gpt-6-luna / medium | Backend/frontend: retain existing UI assertions and add one shared producer-wire guard; open |

LIVE-01 was reported as existing debt; the parent confirmed it defeats approved honest unavailable-source reporting and requires correction. UI-02 and TEST-03 are bounded nonblocking improvements accepted while their owning paths are already changing. TEST-04 identifies a real coverage gap, but its claim of introduction is refuted by the baseline: that test already existed. Its useful UI assertions remain, with accurate naming and a bounded shared wire guard. No aggregate producer-counting redesign is authorized by these dispositions; A12 remains pending.

An additional literal-prefix transport defect (FE-01) was found by the frontend coder during correction: the API client drops configured prefixes named all/ALL. Its focused correction and evidence will be retained separately from independent first-pass review discoveries.


## Corrected candidate AL70-R2

Product/tests/active guidance are frozen. Base remains f63e3db6b2d4baa8e22452e4ec94eff566867721; 259-input fingerprint is 712d1b3db91b641ed2927f75081dd299a61aec84b3b907d69874c5d7614545c2. Immutable input copies, requirements, change map and working records are in /tmp/al70-review-al70-r2-pf1g7e20. All original canonical corrections have focused passing evidence. The additional correction-audit findings FE-01 (frontend coder: literal all/ALL query transport) and CTX-01 (backend coder: missing summary-only backfill context-policy admission) are fixed; their exact execution models remain unverified and are not inferred from reviewer profiles.

Current proof includes full backfill (4.012s), config/manager/live/HTTP package checkpoints, actual worker-child budget publication (3.584s), whole-backend build/test compilation, 215 frontend unit prerequisites plus 497 affected integrations, typecheck/lint/build and a detected App-provider mutant. Shared native SQLite-to-HTTP JSON is checked by Go and consumed through frontend validators/rendering. Available-empty/unavailable and existing SSE subscription/recovery assertions pass. Raw logs remain outside Git; required integrated make remains pending.

All four focuses re-review affected behavior and dependencies under both profiles. This is correction verification with original findings supplied; it does not claim new independent discovery of those same issues. Unchanged R1 investigation may carry forward only with an explicit dependency rationale. Reports use review-r2-<focus>-<profile>.md; all eight cells are initially queued. A12 remains unanswered and unimplemented.


| R2 focus | Model / effort | Phase | Status | Report |
| --- | --- | --- | --- | --- |
| architecture | gpt-6.1-sol / high | re-review | CHANGES_REQUESTED | review-r2-architecture-sol.md |
| architecture | gpt-6-luna / medium | re-review | APPROVED | review-r2-architecture-luna.md |
| data | gpt-6.1-sol / high | re-review | APPROVED | review-r2-data-sol.md |
| data | gpt-6-luna / medium | re-review | APPROVED | review-r2-data-luna.md |
| behavior | gpt-6.1-sol / high | re-review | CHANGES_REQUESTED | review-r2-behavior-sol.md |
| behavior | gpt-6-luna / medium | re-review | APPROVED | review-r2-behavior-luna.md |
| tests | gpt-6.1-sol / high | re-review | APPROVED | review-r2-tests-sol.md |
| tests | gpt-6-luna / medium | re-review | APPROVED | review-r2-tests-luna.md |


## R2 reconciliation and remaining corrections

All eight R2 passes are complete. Both data profiles and both test profiles approve their assigned corrected scopes; architecture and behavior Sol/high report the remaining issues below. The consolidated outcome remains CHANGES_REQUESTED. Earlier Luna approvals do not refute the concrete counterexamples.

- UI-03 (Architecture, gpt-6.1-sol/high; severity4): real recorded-detail HTTP anomaly roles are already projected, but the drawer treats them as raw personas and maps them again. Hydrated event anomalies have raw roles, so guessing from the role string is invalid. Frontend owns explicit presentation origin at selection/render boundaries, without stored-role or API-schema changes. The suspected backend fresh-recomputation double map was separately refuted by the retained-summary wrapper and actual HTTP control; retain that disposition.
- LIVE-02 (Behavior, gpt-6.1-sol/high; severity3): a successful REST poll after failure delivers fresh trees without restoring transport status, leaving a false connection-error banner. Actual production-module and disposable UI probes confirm it; frontend owns bounded recovery correction preserving cancellation and SSE behavior.
- POLL-01 (Coordinator audit; severity4): frontend fallback polling still hardcodes2000ms and has no route for shared live.poll_interval_ms. The actual module probe confirms the timer; it is not claimed as an end-to-end configured333ms server run. Backend/frontend jointly own minimal existing project-config projection and operation-time validation. No new configuration key, numerical substitute or unused-policy requirement on static ledger reads is authorized.

R2 product freeze is lifted only for those assigned corrections. Other backend/source paths remain frozen; A12 stays unanswered and unchanged. Raw reports remain evidence of R2. Correction verification will cover affected role, availability, polling, wire-validation and regression paths under both profiles; unchanged data/history investigations may carry forward only with explicit dependency rationale.


## Focused corrected candidate AL70-R3

All assigned corrections are frozen. Base remains f63e3db6; 265-input fingerprint ee805c71aae0b2d57ba393f31fe49a1506a71afdc70f6e38ed1441d161f51550. Snapshot/change map/requirements/working records are in /tmp/al70-review-al70-r3-zz81i9f2. UI-03, LIVE-02 and POLL-01 have focused passing proof:216 units,159 integrations and the final60-case mixed-selection recipient, frontend typecheck/lint/build/embed sync; HTTP package1.196s and backend build. All four focuses use both profiles on the changed role-origin/recovery/polling/wire-validation paths. Unchanged source/Git/history/publication investigation is carried forward only with a stated dependency rationale; the optional config projection gets fresh data-focus inspection. Eight cells initially queued. A12 remains outside authorized changes. Integrated make remains pending.


| R3 focus | Model / effort | Phase | Status | Report |
| --- | --- | --- | --- | --- |
| architecture | gpt-6.1-sol / high | re-review | CHANGES_REQUESTED | review-r3-architecture-sol.md |
| architecture | gpt-6-luna / medium | re-review | CHANGES_REQUESTED | review-r3-architecture-luna.md |
| behavior | gpt-6.1-sol / high | re-review | CHANGES_REQUESTED | review-r3-behavior-sol.md |
| behavior | gpt-6-luna / medium | re-review | APPROVED | review-r3-behavior-luna.md |
| data | gpt-6.1-sol / high | re-review | CHANGES_REQUESTED | review-r3-data-sol.md |
| data | gpt-6-luna / medium | re-review | APPROVED | review-r3-data-luna.md |
| tests | gpt-6.1-sol / high | re-review | CHANGES_REQUESTED | review-r3-tests-sol.md |
| tests | gpt-6-luna / medium | re-review | APPROVED | review-r3-tests-luna.md |

Parent standalone R3 verification passes at /tmp/al70-r3-standalone-_89smbrx/RESULT.json (binary SHA94b02d3ff72dceff32de1e998f09e3f0a4d2e1fd215f281fdc3466d5b0627ff2). Its executed script also checks actual config API cadence333 when selected and omission with unused absent policy. All12 synthetic ledgers remain byte-identical, current embedded SPA serves524498-byte JS, and invalid/occupied/default8085 PORT-ignored loopback cases pass. This is scoped current-source evidence, not final integrated approval.


## R3 reconciliation

All eight R3 cells completed on matching265-input evidence. Direct UI-03/LIVE-02/POLL-01 corrections are verified fixed; the consolidated verdict remains CHANGES_REQUESTED for two confirmed adjacent defects. UI-04 is independently reported by architecture under both gpt-6.1-sol/high and gpt-6-luna/medium, and behavior/data/tests under gpt-6.1-sol/high. Actual HTTP/LiveService and frontend probes show one unavailable-owner summary appearing twice after raw/projected context comparison. Reconcile using existing raw authority and owner coverage while preserving real repeated occurrences and many-to-one persona distinctions; mapping role text before deduplication is not an approved repair. The pending-selection control passes, so selection loss is not claimed. Severity4 reconciles the shared impact; the architecture Luna report's severity5 and repair suggestion remain visible in source records.

LIVE-03 is reported by data/gpt-6.1-sol/high: deferred production-module polls accept newer then older trees, or falsely clear a newer failure when older success completes. Preserve exact configured cadence, sequential recovery and teardown with clear in-flight request ownership, without starvation or a new timeout/default. Both roots predate R3; they remain current-contract corrections. Frontend owns fixes; backend owns only one bounded producer-wire regression. Other product paths stay frozen. A12 remains untouched.

Main integration preflight fetched origin/main3384fa482a9568cfced2db2e0ee943527a8a22cf: two upstream GitHub Pages/planning commits change documentation/site content, with no Go/TypeScript implementation changes. The incoming presentation's fresh-release sentence still prescribes latest selection; after the authorized merge, reconcile it with confirmed C09d explicit release approval while preserving the new presentation. No merge or source implementation change has occurred during frozen review.


## Focused corrected candidate AL70-R4

Product/tests/guidance frozen; base f63e3db6,270-input fingerprint b656706936a5c83845a3390af67add958ba7a7a378385a01d5dcfbf1679c6d45. Snapshot/change map/requirements/working records: /tmp/al70-review-al70-r4-f27abdbx. UI-04 now reconciles canonical raw occurrences with owner restoration and explicit multiplicity; recorded projected summaries remain separate. LIVE-03 uses accepted-completion ordering, preventing stale replacement while retaining useful progress. All17 existing UI assertions survive corrected canonical fixtures. Proof:217 units,114 affected integrations, typecheck/lint/build/embed sync; actual native partial producer and shared JSON pair0.084s plus HTTP hydration peer0.070s. No production Go/schema/configuration/A12 changes. All four focuses re-review these changed identity/order/wire/test boundaries under both profiles; eight cells initially queued. Unchanged R3 source/config/Git/history/publication/poll-policy investigation carries forward only with a dependency rationale. Integrated make remains pending.


| R4 focus | Model / effort | Phase | Status | Report |
| --- | --- | --- | --- | --- |
| architecture | gpt-6.1-sol / high | re-review | CHANGES_REQUESTED | review-r4-architecture-sol.md |
| architecture | gpt-6-luna / medium | re-review | APPROVED | review-r4-architecture-luna.md |
| behavior | gpt-6.1-sol / high | re-review | APPROVED | review-r4-behavior-sol.md |
| behavior | gpt-6-luna / medium | re-review | APPROVED | review-r4-behavior-luna.md |
| data | gpt-6.1-sol / high | re-review | APPROVED | review-r4-data-sol.md |
| data | gpt-6-luna / medium | re-review | APPROVED | review-r4-data-luna.md |
| tests | gpt-6.1-sol / high | re-review | APPROVED | review-r4-tests-sol.md |
| tests | gpt-6-luna / medium | re-review | APPROVED | review-r4-tests-luna.md |

Current standalone R4 smoke passes at /tmp/al70-r4-standalone-zo06gqpb/RESULT.json (binary SHAcc263cf5a754241fdfd079ae78f8d51f53f58abaa4915a632e9554d4e726759e). Executed script checks native config cadence333/omission and12 unchanged synthetic ledgers without sources/Git, with missing unused controls/invalid local policy, current524760-byte embedded JS, items[], explicit case/roles, loopback advertisement and invalid/occupied/default8085 PORT-ignored cases. This supplements focused reviews; it is not integrated make or deployment evidence.


## R4 reconciliation and narrow presentation correction

All eight R4 passes are complete. Both behavior/data/test profiles and architecture Luna approve their assigned scopes; architecture Sol identifies UI-05. UI-04 occurrence identity and LIVE-03 ordered polling are verified fixed, including actual producer-to-UI controls and48 accepted-completion scenarios. UI-05 (architecture/gpt-6.1-sol/high) is confirmed: generated role text uses raw personas while adjacent display fields use configured aliases. Parent severity3 reflects display-only impact; reporter severity4 remains recorded. Preserve raw facts, occurrence/evidence identity and debug prompts; apply known presentation ownership once to generated headings/descriptions/tooltips.

Only this bounded frontend display correction is authorized for the next snapshot. It adds no state, configuration, schema, API, selector, polling or persistence behavior. Apply the protocol's mechanical low-risk exception: architecture and behavior under both profiles review the changed presentation boundary. Carry forward R4 data approval only after confirming raw facts/identity/ordering code is unchanged and immutable-input controls pass; carry test-effectiveness approval only after inspecting preserved meaningful oracles and focused producer-backed display regressions. Parent will verify those dependencies rather than repeat unchanged investigations. If the implementation exceeds that boundary, reassess affected focus coverage. Integrated make remains required before consolidated delivery approval; A12 is untouched.


## Mechanical presentation candidate AL70-R5

Product/tests/guidance frozen; base f63e3db6,274-input fingerprint53216836c57b58aea99790ea61b6556e92a7bdc485bcda24907b4545b975da57. Snapshot/delta/requirements/working records: /tmp/al70-review-al70-r5-lb30yevb. UI-05 formatter/callers now explicitly label generated display roles by known raw/projected origin; raw debug formatting remains unchanged. Four production files change only generated text/callers/tooltip. Parent compared exact diffs and verified raw reconciliation/polling/native Go wire tests/JSON fixtures unchanged. New checks pass218 units/173 integrations/typecheck/lint/build/embed sync; old-call mutation catches3 raw display cases with projected control still passing.

Under the recorded low-risk exception, architecture and behavior each use both profiles. Data/test R4 approvals carry with explicit unchanged dependency/oracle rationale, not invented new executions. Parent inspected readonly formatting, raw prompt preservation, native-pair display guards and mutation proof. Reassess coverage if any change exceeds this display boundary. Four R5 cells initially queued; integrated make remains required. A12 stays unanswered and outside authorized changes.


| R5 focus | Model / effort | Phase | Status | Report |
| --- | --- | --- | --- | --- |
| architecture | gpt-6.1-sol / high | re-review | APPROVED | review-r5-architecture-sol.md |
| architecture | gpt-6-luna / medium | re-review | APPROVED | review-r5-architecture-luna.md |
| behavior | gpt-6.1-sol / high | re-review | APPROVED | review-r5-behavior-sol.md |
| behavior | gpt-6-luna / medium | re-review | APPROVED | review-r5-behavior-luna.md |
| data | both configured profiles | carried from R4 | APPROVED for unchanged scope | review-r4-data-{sol,luna}.md |
| tests | both configured profiles | carried from R4 | APPROVED for unchanged scope | review-r4-tests-{sol,luna}.md |

Parent standalone R5 result passes at /tmp/al70-r5-standalone-tjq42onb/RESULT.json (binary SHA68c7fc390f4ab07d1e1d52a1e43b3333018097e0bc2f9bd8f4ef2e535cffd148). Its executed script verifies12 synthetic ledgers byte-for-byte, actual config cadence333/omission, source-free API and current524916-byte embedded JS, missing unused policy/invalid local input, loopback and selected/invalid/occupied/default8085 PORT-ignored cases. No integrated make or remote deployment claim is made.


## Application review gate reached; integration pending

All four R5 cells approve the mechanical display correction on matching274-input evidence. Both R4 data/test profiles carry forward only for the documented unchanged dependency/oracle scopes. Every canonical finding is fixed with verification references; discovery attribution remains separate, including shared UI-04 credit. No blocking application finding or material scoped review gap remains. A12 is an unapproved proposal, not an implemented acceptance change.

The parent may now commit the reviewed candidate, integrate main and execute the required full make gate. Consolidated delivery approval is still INCOMPLETE until actual integration impact and required checks pass. Working reports/structured records remain retained until approved compaction. No telemetry, push, PR or remote merge has occurred.


## Main integration and impact assessment

Reviewed candidate committed69e17b44ca9e1d98d901d8e8c3f2de321029a325 with a clean working tree, then fetched/merged origin/main3384fa482a9568cfced2db2e0ee943527a8a22cf without conflicts at172091096ae1e0cbf50553eebae11df34487d085. Incoming changes are GitHub Pages/planning documentation only. All274 reviewed inputs match except README.md's preserved upstream presentation link; no Go/TypeScript/test/embed input changed. Current application approvals and documented carries remain valid.

The incoming site's fresh-release sentence was aligned with confirmed C09d explicit exact-release approval and same-tag bundle selection, preserving layout and browsing links. This prose-only correction was inspected against the contract; no runtime logic changed. Diff check passes. Existing Pages tests could not run because ruby is absent (exit127); no tooling was installed. This low-risk copy limit does not replace or waive the required application make gate. Consolidated delivery approval remains pending that actual result.


## First integrated gate and expectation correction

Required make on76d8439742460b6c61bc78d6e20b7c19b8061d8c exited2 after429.858s. Both unit lanes, lint/vet/typecheck,14 release,10 installation,23 distribution,6 orchestration and pure-domain race checks passed. All race-enabled backend packages except CLI passed. CLI had exactly two stale expected-message assertions: correct successful empty/unmatched migration output now includes readable. The gate stopped before frontend integration/native builds; those stages are not claimed from this run. Raw failed evidence is retained at /tmp/al70-integrated-make-first.log and result JSON.

TEST-05 is a confirmed severity2 test expectation lag, found by the integration coordinator. CLI tester changed only the two expected substrings and matching diagnostic; all successful-exit and other oracles remain. Uncached full TestCLIMigrateCommand_Adversarial passes2.165s. Parent inspected the exact three-line diff against actual stdout. This mechanical expectation-only correction changes no product, fixture, ownership or concurrency behavior, so current application and test-purpose approvals carry with that explicit rationale; the full make rerun is still mandatory. No new broad reviewer cycle is needed to settle a conclusive text oracle. Consolidated delivery approval remains pending a green full gate.

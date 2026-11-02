# AL-68 iteration 02 retained review summary

Current scoped outcome: APPROVED. Historical changes-requested, incomplete, cancelled and refuted outcomes remain distinguished below; final approval does not rewrite earlier verdicts.

## Findings and reviewer credit

This review retains findings, validity, correction status and reviewer/model discovery and verification credit. The execution roster distinguishes requested from source-observed settings; detailed historical dispatch records remain in Git history.

| Finding | Issue | Disposition / lifecycle | Discovering reviewer and model/effort | Verifying reviewer and model/effort |
| --- | --- | --- | --- | --- |
| AUTO-01 | Retained baseline owners are not checked against the requested ticket | confirmed / fixed | behavior / gpt-6.1-sol/high; data / gpt-6.1-sol/high; security / gpt-6.1-sol/high | behavior / gpt-6.1-sol/high; behavior / gpt-6-luna/medium; data / gpt-6.1-sol/high; data / gpt-6-luna/medium; security / gpt-6.1-sol/high; security / gpt-6-luna/medium |
| AUTO-02 | Continuous callbacks indefinitely reset publication debounce | confirmed / fixed | behavior / gpt-6.1-sol/high | behavior / gpt-6.1-sol/high; behavior / gpt-6-luna/medium |
| AUTO-03 | Existing automatic worker continues writing after disable | confirmed / fixed | behavior / gpt-6.1-sol/high; behavior / gpt-6-luna/medium; architecture / gpt-6.1-sol/high; release / gpt-6.1-sol/high; release / gpt-6-luna/medium | behavior / gpt-6.1-sol/high; behavior / gpt-6-luna/medium; architecture / gpt-6.1-sol/high; architecture / gpt-6-luna/medium; release / gpt-6.1-sol/high; release / gpt-6-luna/medium |
| AUTO-04 | Status fails for a valid registry without publication_request table | confirmed / fixed | behavior / gpt-6.1-sol/high; architecture / gpt-6.1-sol/high | behavior / gpt-6.1-sol/high; behavior / gpt-6-luna/medium; architecture / gpt-6.1-sol/high; architecture / gpt-6-luna/medium |
| AUTO-05 | Automatic path publishes first ledger with missing explicit title | refuted / not_applicable | behavior / gpt-6-luna/medium | behavior / gpt-6.1-sol/high; behavior / gpt-6-luna/medium |
| AUTO-06 | Missing-leaf aliases allow ledger and summary destinations to overlap | confirmed / fixed | data / gpt-6.1-sol/high | data / gpt-6.1-sol/high; data / gpt-6-luna/medium; tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium |
| AUTO-07 | In-flight regression accepts acknowledgement of unseen work | confirmed / fixed | tests / gpt-6.1-sol/high | tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium |
| AUTO-08 | Installed state paths silently diverge after YAML path changes | confirmed / fixed | architecture / gpt-6.1-sol/high | architecture / gpt-6.1-sol/high; architecture / gpt-6-luna/medium; release / gpt-6.1-sol/high; release / gpt-6-luna/medium |
| AUTO-09 | Subdirectory generation selects process cwd instead of owning configuration root | confirmed / fixed | architecture / gpt-6.1-sol/high | architecture / gpt-6.1-sol/high; architecture / gpt-6-luna/medium |
| AUTO-10 | Configurable work budgets have no practical upper ceilings | refuted / not_applicable | architecture / gpt-6-luna/medium | architecture / gpt-6.1-sol/high; architecture / gpt-6-luna/medium |
| AUTO-11 | Publication recovery trusts arbitrary journal artifact destinations | confirmed / fixed | security / gpt-6.1-sol/high; security / gpt-6-luna/medium | data / gpt-6.1-sol/high; data / gpt-6-luna/medium; security / gpt-6.1-sol/high; security / gpt-6-luna/medium |
| AUTO-12 | Direct callback transcript path bypasses selected Codex source-store proof | confirmed / fixed | security / gpt-6.1-sol/high | security / gpt-6.1-sol/high; security / gpt-6-luna/medium |
| AUTO-13 | Setup guidance contradicts automatic publication and optional reports | confirmed / fixed | release / gpt-6-luna/medium | release / gpt-6.1-sol/high; release / gpt-6-luna/medium |
| AUTO-14 | True project aliases cannot reinstall or remove owned installation | confirmed / fixed | release / gpt-6.1-sol/high | release / gpt-6.1-sol/high; release / gpt-6-luna/medium |
| AUTO-15 | Installer resolves omitted source store from caller cwd instead of selected project | confirmed / fixed | release / gpt-6.1-sol/high | release / gpt-6.1-sol/high; release / gpt-6-luna/medium |
| AUTO-16 | Dangling output symlink bypasses project destination containment | confirmed / fixed | security / gpt-6.1-sol/high | data / gpt-6.1-sol/high; data / gpt-6-luna/medium; tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium; security / gpt-6.1-sol/high; security / gpt-6-luna/medium |
| AUTO-17 | Lexical cleaning precedes physical symlink traversal | confirmed / fixed | security / gpt-6.1-sol/high | tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium; security / gpt-6.1-sol/high; security / gpt-6-luna/medium |
| AUTO-18 | Dangling ledger leaf alias acquires independent publication lock | confirmed / fixed | data / gpt-6.1-sol/high | data / gpt-6.1-sol/high; data / gpt-6-luna/medium; tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium |
| AUTO-19 | Atomic artifact writers replace dangling aliases and strand publication identity | confirmed / fixed | security / gpt-6.1-sol/high; data / gpt-6.1-sol/high | data / gpt-6.1-sol/high; data / gpt-6-luna/medium; tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium; security / gpt-6.1-sol/high; security / gpt-6-luna/medium |
| AUTO-20 | Raw source symlink-dotdot traversal bypasses selected Codex store membership | confirmed / fixed | security / gpt-6.1-sol/high | behavior / gpt-6.1-sol/high; behavior / gpt-6-luna/medium; data / gpt-6.1-sol/high; data / gpt-6-luna/medium; tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium; security / gpt-6.1-sol/high; security / gpt-6-luna/medium |
| AUTO-21 | Installer changes physical source-home semantics before emitting hooks | confirmed / fixed | release / gpt-6.1-sol/high | tests / gpt-6.1-sol/high; tests / gpt-6-luna/medium; release / gpt-6.1-sol/high; release / gpt-6-luna/medium |
| MERGE-01 | Installation and binary repair reject release-candidate tags emitted by the repository workflow | confirmed / fixed | installation/distribution integration / gpt-6.1-sol/high | installation/distribution integration / gpt-6.1-sol/high; installation/distribution integration / gpt-6-luna/medium |

## Finding evidence and resolutions

### AUTO-01

Severity: P1; classification: pre_existing; disposition: confirmed; lifecycle: fixed. Location: backend/cmd/agentic-lens/capture_worker.go:204.

Trigger: Reuse ticket-A baseline for ticket B in same project with valid distinct ticket registrations.

Required behavior: Retained baseline owner associations must match the requested ticket as well as project/source/store.

Consequence: Ticket-A usage is relabelled into ticket B.

Disposition evidence: Independent native probes in both reports reproduce successful contaminated publication; source validator receives project but not active ticket.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-02

Severity: P1; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/cmd/agentic-lens/capture_worker.go:101.

Trigger: Callbacks repeatedly arrive faster than debounce_ms.

Required behavior: Coalescing must have bounded delay and periodic refresh must not require global quiet.

Consequence: Already available source data is not published during sustained activity.

Disposition evidence: Native timing probe crosses multiple refresh intervals without publication; source ANDs periodic due with latest-request quiet time.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-03

Severity: P2; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/cmd/agentic-lens/capture_worker.go:19.

Trigger: Disable capture in YAML or remove owned hooks while worker is active.

Required behavior: Disable must prevent subsequent automatic publication within a documented bounded transition; manual recovery remains supported.

Consequence: Ledger/summary continue changing despite disable.

Disposition evidence: Behavior Sol native probe shows updated_at advancing after disabled policy; both independent reports trace latched policy and absent automatic-generation enablement guard.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-04

Severity: P2; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/capture/worker.go:28.

Trigger: Register/bind with automatic capture disabled, then capture status.

Required behavior: Status must diagnose valid disabled or pre-request registries without activating automatic work.

Consequence: Status exits with no-such-table SQL error.

Disposition evidence: Native probe and table creation/read trace: only RequestPublication creates table, status reads it unconditionally.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-05

Severity: P2; classification: introduced; disposition: refuted; lifecycle: not_applicable. Location: backend/cmd/agentic-lens/main.go:174.

Trigger: Bind ticket without explicit title and no accepted ledger title fallback.

Required behavior: Missing required initial metadata should be actionable pending context, preserving no-guess behavior and manual generation metadata guarantees.

Consequence: Automatic ledger can publish without the title required by the normal generation path.

Disposition evidence: Downstream SaveFactLedger required-title guard blocks the claimed successful publication; missing-title early handling may improve diagnostics but the reported consequence is refuted.

Resolution: Refutation retained with original observations and contrary source evidence; not counted as fixed or deleted.

Verified candidate: candidate-02-01.

### AUTO-06

Severity: P2; classification: pre_existing; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/publication_set.go:88.

Trigger: First publish selects ledger and summary paths through different parent-directory aliases to same absent file.

Required behavior: Artifact destination distinctness must use physical parent/nearest-existing-ancestor resolution before visible writes.

Consequence: Generation reports success but leaves JSON plus Markdown in ledger and clears recovery journal.

Disposition evidence: Native probe reproduces invalid JSON and removed journal; full-leaf EvalSymlinks fallback misses existing parent alias on ENOENT.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-07

Severity: P2; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/cmd/agentic-lens/capture_worker_unix_test.go:136.

Trigger: A new publication request is queued during the first pass, but the fixture does not change any source or ticket membership.

Required behavior: The regression must reject a worker mutation that acknowledges the latest unseen generation before processing its newly selected work.

Consequence: A future regression could swallow a newly bound ticket while test assertions still pass. Frozen production code passed the stronger control; no current production failure claimed.

Disposition evidence: Disposable mutation passes existing test, while stronger late-ticket observable probe fails mutant and passes unchanged frozen production; original assertions only check acknowledgement equality and unchanged37.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-08

Severity: P2; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/config/capture.go:37.

Trigger: Change configured state_dir after installing callbacks with frozen absolute state override.

Required behavior: Approved shared configuration/project-resolution/resource-bound contract as detailed in original report.

Consequence: Automatic collection uses old registry while default manual/status commands resolve a different empty registry.

Disposition evidence: Native install/config/status probe and source trace reproduce absent drift diagnosis. Explicit overrides remain legitimate; installed context needs stable discovery and actionable mismatch handling.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-09

Severity: P2; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/cmd/agentic-lens/main.go:62.

Trigger: Invoke manual generation from unconfigured src child of a configured project with a root-source owner.

Required behavior: Approved shared configuration/project-resolution/resource-bound contract as detailed in original report.

Consequence: Normal config resolves root but capture source/state/output scope becomes child and rejects same-project data.

Disposition evidence: Native root-control/subdirectory probe reproduces ownership failure; config.Load upward traversal disagrees with capture resolver.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-10

Severity: P2; classification: suggestion; disposition: refuted; lifecycle: not_applicable. Location: backend/internal/domain/capture_config.go:25.

Trigger: Set very large positive max_sessions or max_source_bytes in project policy.

Required behavior: Approved shared configuration/project-resolution/resource-bound contract as detailed in original report.

Consequence: Reviewer claims practical CPU/memory/IO exhaustion despite pass timeout.

Disposition evidence: Parent adjudication: positive finite caller-selected max_sessions/max_source_bytes already provide per-pass bounds, with a separate subprocess deadline. No over-budget work, overflow, unbounded allocation independent of the configured limit or supported-load failure was established. An additional product-wide practical ceiling may be useful hardening but is not an agreed requirement; arbitrarily choosing one is not necessary to satisfy configured bounds. Original CHANGES_REQUESTED report remains intact.

Resolution: Refutation retained with original observations and contrary source evidence; not counted as fixed or deleted.

Verified candidate: candidate-02-01.

### AUTO-11

Severity: P1; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/publication_set.go:150.

Trigger: A project-supplied or copied .publication-pending journal records absolute outside-project paths as prior artifact members.

Required behavior: Every recovered member must match the canonical authorized project/ledger artifact set before any write or removal.

Consequence: Normal generation overwrites or removes arbitrary user-writable outside files during recovery before source/config failure.

Disposition evidence: Both independent security native probes reproduce outside-file overwrite; direct source only checks absolute path before durableReplace/os.Remove and never binds journal to expected project/artifact set.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-12

Severity: P2; classification: pre_existing; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/capture/project.go:117.

Trigger: Callback supplies actual home-A transcript with matching ID/project while configured source identity/store is unrelated home B.

Required behavior: Transcript and queued-child source paths must be proven to belong to the exact configured physical source store as well as session/project.

Consequence: Home-A usage becomes current under a home-B registration, defeating exact source/store ownership.

Disposition evidence: Native receive/bind/reconcile probe reproduces; ValidateOwner reads supplied path instead of verifying it against selected-store exact-owner discovery, then checks only ID/cwd.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-13

Severity: P2; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: docs/SETUP_MANUAL.md:171.

Trigger: Follow old normative manual/setup/operator text with auto publication or detailed_report=false.

Required behavior: Consistent approved automatic publication, project/source configuration and additive alias-aware setup responsibilities.

Consequence: Users are told callbacks do not publish and unconditionally to inspect an intentionally absent report.

Disposition evidence: Direct contradiction between old normative text and new worker/config policy.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-14

Severity: P2; classification: pre_existing; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/capture/install.go:135.

Trigger: Install via physical project path then reinstall/disable/uninstall via a true symlink alias.

Required behavior: Consistent approved automatic publication, project/source configuration and additive alias-aware setup responsibilities.

Consequence: Same canonical state is found but lexical manifest.Repository comparison rejects legitimate owned upgrade/removal.

Disposition evidence: Native space/alias install probe reproduces; CLI routes canonical state but passes lexical --repo to Install.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-15

Severity: P2; classification: pre_existing; disposition: confirmed; lifecycle: fixed. Location: backend/cmd/agentic-lens/capture.go:281.

Trigger: Install --repo configured-project without --store while caller cwd is elsewhere.

Required behavior: Consistent approved automatic publication, project/source configuration and additive alias-aware setup responsibilities.

Consequence: Callbacks embed wrong source store/default or fail based on unrelated caller config.

Disposition evidence: Native dry-run ignores selected project local source home; captureSourceStore uses config.Load empty root.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-16

Severity: P1; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/capture/project.go:25-40,160-176.

Trigger: Summary/output symlink points to absent outside-project target; ordinary generate or complete ledger-bound journal selects that member.

Required behavior: Publication destinations must resolve physically inside selected project before journal or artifact mutation.

Consequence: Ordinary summary writes or journal recovery create outside-project file with attacker-selected prior bytes.

Disposition evidence: Native no-journal and complete-journal probes reproduce success only when outside target absent; ProjectDirectory drops dangling symlink while publication canonicalizer follows it.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-17

Severity: P1; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/capture/project.go:45-48.

Trigger: Relative symlink target or raw output path has intermediate symlink followed by .., reaching existing or absent outside target.

Required behavior: Containment and writes must agree on physical filesystem traversal before artifact mutation.

Consequence: Normal summary publication creates or modifies outside-project file.

Disposition evidence: Independent disposable probe and source trace confirmed; parent regression fails prior source and passes corrected resolver/lock.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-18

Severity: P2; classification: pre_existing; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/capture/publication.go:34-39.

Trigger: Alias leaf points to absent in-project ledger; alias and target publishers acquire ownership concurrently.

Required behavior: True physical aliases serialize before baseline/source reads through publication.

Consequence: Simultaneous owners can publish from stale baselines and omit accepted roots.

Disposition evidence: Independent disposable probe and source trace confirmed; parent regression fails prior source and passes corrected resolver/lock.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-19

Severity: P2; classification: pre_existing; disposition: confirmed; lifecycle: fixed. Location: backend/cmd/agentic-lens/capture.go:493-499,525.

Trigger: Initial publication through legitimate in-project ledger symlink to absent physical target.

Required behavior: Validated destination, ownership, staged write and journal must retain identical physical identity through publication and retry.

Consequence: Alias replaced with regular file; intended target remains absent, finalization fails and same-argument retry bypasses original target journal.

Disposition evidence: Security Sol and Data Sol independently reproduce native first-write failure, orphaned canonical marker and successful wrong-identity retry; existing-target controls pass.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-20

Severity: P2; classification: introduced; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/codex/discovery.go:361-366.

Trigger: Selected-home sessions/hop points to outside child; raw hop/../root.jsonl reads valid same-project transcript outside selected store.

Required behavior: Physical file read must belong to selected source store independently of exact session ID/project.

Consequence: Outside-store usage is accepted/current and published under selected store; queued child variant imports foreign child.

Disposition evidence: Security Sol independent handler/bind/native publication probe plus parent source trace and native root/queued-child old-source overlay reproduce outside usage.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### AUTO-21

Severity: P2; classification: pre_existing; disposition: confirmed; lifecycle: fixed. Location: backend/internal/adapters/capture/install.go:166,68-72.

Trigger: Configured or explicit raw source alias contains symlink then dotdot; native receive/manual select actual home while install lexically cleans it.

Required behavior: Install/receive/manual/worker must agree on exact selected physical source store, preserving legitimate aliases.

Consequence: Successful setup emits hooks rejecting valid source accepted by native receive/manual; automatic capture fails.

Disposition evidence: Release Sol isolated native installer/generated exact shell command rejects real fixture source while raw native receive/manual succeed; parent five-case old/new oracle confirms.

Resolution: Corrected production/test behavior verified by native/component controls and paired affected reviewers, with explicit unchanged dependency carry-forward. Final integrated make passed on unchanged reviewed source; final native telemetry/PR delivery are parent-owned.

Verified candidate: candidate-02-07.

### MERGE-01

Severity: P2; classification: introduced_by_upstream_integration; disposition: confirmed; lifecycle: fixed. Location: .agents/skills/agentic-lens-setup/scripts/install.sh and scripts/download.sh.

Trigger: Select a workflow-produced RC tag such as v0.1.0-rc.42+abc1234 through installer, downloader or release marker.

Required behavior: Published supported release tags must be accepted by setup and worktree repair.

Consequence: Bootstrap, repair and capture-aware upgrades reject the RC before installation.

Disposition evidence: Disposable premerge downloader accepts the tag; merge698 installer/downloader reject it. Parent verified workflow producer. Independent restoring-either-validator overlays fail the new actual installed-bundle RC oracle.

Resolution: Permit + in the narrow allowlist, explicitly reject standalone . and .., and retain traversal/whitespace/shell-payload rejection. Actual source-bundle RC identity and implicit/explicit repair regression tests pass; full make passes.

Verified candidate: 7e09f417af8e5e681a14b18fbc1c827fa26cc3ec.

## Verification and remaining limits

Candidate06 full `go test -race ./...` and `go vet ./...` passed; candidate07 targeted installer/config/capture race checks and independent old-installer/old-config overlays pass/fail as expected. Final integrated `make` passed after implementation commit55b4dd4 and upstream integration with all58 reviewed paths unchanged; full historical output is available at the retained Git revision. Earlier make01-04 results apply to their own candidates, not current delivery approval.

Corrected-handler real runtime evidence: Linux Codex0.166, three independent roots and two completed native children; Linux OpenCode1.19.5, two same-ticket/store directories with disjoint12/3owners, including five children in the first. Exact source project/session/native-parent/model/usage and available costs were compared; automatic publication required no contributing-agent generate. Human Codex trust was approved by the user; hook definition bytes were not bypassed. Current path/source changes are covered by native synthetic regressions and bounded unchanged existing-directory carry-forward.

Unverified broader cases remain explicit: Codex SubagentStart delivery, active/interrupted children before Stop, non-Linux runtime, actual forty-minute elapsed operation and physical power-loss behavior. Synthetic long-growth fixtures, cross-builds, native crash and persisted-journal recovery do not prove those claims. The accepted supported boundary does not promise universal harness/platform capture or hostile concurrent filesystem mutation. Top-level completion is observability finalization, not ticket completion.

The native ledger contains89same-project Codex owners (65independent roots,24native children), retaining all35prior same-project owners. Seven prior cross-directory smoke owners remain in historical committed evidence. Independent raw JSONL comparison uses per-response usage when available, otherwise native cumulative usage, and the exact source revision prefix. Usage, model, ID/project, native parents, schema5, observability completion and report/ledger pairing pass. No source facts are hand-edited.

Coverage remains conservative: some accepted owner snapshots retain pending markers because differing full snapshots at equal/unordered checkpoints preserve accepted persisted evidence. The independent check verifies published metrics, not universal or complete lifecycle coverage. The user subsequently disabled detailed report generation and removed the standalone capture_report.json. Coverage counts and verification conclusions are retained here; detailed historical backfill/recovery logs remain in Git history; the actual integrations normally receive paths through hooks.

Merged origin/main 99fbcd7 through merge commit698aa45, with correction candidate7e09f41. Scope verdict: APPROVED. The integration retains main's complete skill bootstrap/worktree repair and offline operator launcher together with AL-68's opt-in native capture, handler protocol preflight and rollback. The original integration retained complete evidence; current retention keeps compact findings/model credit and historical Git access. Iteration03 remains draft and was not implemented.

Installation/distribution reviewer gpt-6.1-sol/high found MERGE-01: both new setup validators rejected the plus sign in the repository's own release-candidate tags. The narrow correction and actual installed-bundle RC test are verified by gpt-6.1-sol/high and gpt-6-luna/medium. Both final scoped reviews approve. The initial Luna approval and original Sol changes-requested outcome remain recorded; supplied-finding verification is not independent discovery credit. Parent matched exact source session/project and observed model/effort for all runs.

Initial and corrected integrated make both passed; final make at7e09f41 includes installation14/capture-distribution15 checks, backend race/vet, frontend typecheck/lint,781frontend tests and native Linux/Darwin builds. Existing native runtime support evidence carries only across unchanged Go/TypeScript paths. No new live harness smoke or non-Linux runtime claim is made. Detailed reports/logs remain in Git history; MERGE-01 retains its discovery and verification credit in the finding table above.

Integration source verification passed for all 93 expected same-project owners, preserving all 89 prior owners and archived foreign-directory history. Model/usage/native-parent/project/source-prefix comparisons have no errors. Final capture status counts are {'pending': 55, 'current': 38}; status alone is not asserted complete. No detailed report was reintroduced.

## Review execution history

Every original roster cell is retained in this compact table, including cancelled and no-findings runs. Requested identity is not presented as observed identity.

| Run | Reviewer | Requested model/effort | Observed model/effort | Phase | Candidate | Execution / verdict |
| --- | --- | --- | --- | --- | --- | --- |
| iter02-behavior-sol-01 | behavior | gpt-6.1-sol/high | gpt-6.1-sol/high | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-behavior-luna-01 | behavior | gpt-6-luna/medium | gpt-6-luna/medium | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-data-sol-01 | data | gpt-6.1-sol/high | gpt-6.1-sol/high | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-data-luna-01 | data | gpt-6-luna/medium | gpt-6-luna/medium | discovery | candidate-02-01 | completed / APPROVED |
| iter02-architecture-sol-01 | architecture | gpt-6.1-sol/high | gpt-6.1-sol/high | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-architecture-luna-01 | architecture | gpt-6-luna/medium | gpt-6-luna/medium | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-tests-sol-01 | tests | gpt-6.1-sol/high | gpt-6.1-sol/high | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-tests-luna-01 | tests | gpt-6-luna/medium | gpt-6-luna/medium | discovery | candidate-02-01 | completed / APPROVED |
| iter02-security-sol-01 | security | gpt-6.1-sol/high | gpt-6.1-sol/high | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-security-luna-01 | security | gpt-6-luna/medium | gpt-6-luna/medium | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-release-sol-01 | release | gpt-6.1-sol/high | gpt-6.1-sol/high | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-release-luna-01 | release | gpt-6-luna/medium | gpt-6-luna/medium | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-behavior-sol-02 | behavior | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-02 | completed / APPROVED |
| iter02-behavior-luna-02 | behavior | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-02-02 | completed / APPROVED |
| iter02-data-sol-02 | data | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-02 | cancelled / none |
| iter02-data-luna-02 | data | gpt-6-luna/medium | unavailable | re_review | candidate-02-02 | cancelled / none |
| iter02-architecture-sol-02 | architecture | gpt-6.1-sol/high | unavailable | re_review | candidate-02-02 | cancelled / none |
| iter02-architecture-luna-02 | architecture | gpt-6-luna/medium | unavailable | re_review | candidate-02-02 | cancelled / none |
| iter02-tests-sol-02 | tests | gpt-6.1-sol/high | unavailable | re_review | candidate-02-02 | cancelled / none |
| iter02-tests-luna-02 | tests | gpt-6-luna/medium | unavailable | re_review | candidate-02-02 | cancelled / none |
| iter02-security-sol-02 | security | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-02 | completed / CHANGES_REQUESTED |
| iter02-security-luna-02 | security | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-02-02 | completed / APPROVED |
| iter02-release-sol-02 | release | gpt-6.1-sol/high | unavailable | re_review | candidate-02-02 | cancelled / none |
| iter02-release-luna-02 | release | gpt-6-luna/medium | unavailable | re_review | candidate-02-02 | cancelled / none |
| iter02-data-sol-03 | data | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-03 | completed / CHANGES_REQUESTED |
| iter02-data-luna-03 | data | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-02-03 | completed / APPROVED |
| iter02-architecture-sol-03 | architecture | gpt-6.1-sol/high | unavailable | re_review | candidate-02-03 | cancelled / none |
| iter02-architecture-luna-03 | architecture | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-02-03 | cancelled / none |
| iter02-tests-sol-03 | tests | gpt-6.1-sol/high | unavailable | re_review | candidate-02-03 | cancelled / none |
| iter02-tests-luna-03 | tests | gpt-6-luna/medium | unavailable | re_review | candidate-02-03 | cancelled / none |
| iter02-security-sol-03 | security | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-03 | completed / CHANGES_REQUESTED |
| iter02-security-luna-03 | security | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-02-03 | completed / APPROVED |
| iter02-release-sol-03 | release | gpt-6.1-sol/high | unavailable | re_review | candidate-02-03 | cancelled / none |
| iter02-release-luna-03 | release | gpt-6-luna/medium | unavailable | re_review | candidate-02-03 | cancelled / none |
| iter02-data-sol-04 | data | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-04 | completed / CHANGES_REQUESTED |
| iter02-data-luna-04 | data | gpt-6-luna/medium | unavailable | re_review | candidate-02-04 | cancelled / none |
| iter02-architecture-sol-04 | architecture | gpt-6.1-sol/high | unavailable | re_review | candidate-02-04 | cancelled / none |
| iter02-architecture-luna-04 | architecture | gpt-6-luna/medium | unavailable | re_review | candidate-02-04 | cancelled / none |
| iter02-tests-sol-04 | tests | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-04 | completed / APPROVED |
| iter02-tests-luna-04 | tests | gpt-6-luna/medium | unavailable | re_review | candidate-02-04 | cancelled / none |
| iter02-security-sol-04 | security | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-04 | completed / CHANGES_REQUESTED |
| iter02-security-luna-04 | security | gpt-6-luna/medium | unavailable | re_review | candidate-02-04 | cancelled / none |
| iter02-release-sol-04 | release | gpt-6.1-sol/high | unavailable | re_review | candidate-02-04 | cancelled / none |
| iter02-release-luna-04 | release | gpt-6-luna/medium | unavailable | re_review | candidate-02-04 | cancelled / none |
| iter02-data-sol-05 | data | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-05 | completed / APPROVED |
| iter02-data-luna-05 | data | gpt-6-luna/medium | unavailable | re_review | candidate-02-05 | cancelled / none |
| iter02-architecture-sol-05 | architecture | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-05 | completed / APPROVED |
| iter02-architecture-luna-05 | architecture | gpt-6-luna/medium | unavailable | re_review | candidate-02-05 | cancelled / none |
| iter02-tests-sol-05 | tests | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-05 | cancelled / none |
| iter02-tests-luna-05 | tests | gpt-6-luna/medium | unavailable | re_review | candidate-02-05 | cancelled / none |
| iter02-security-sol-05 | security | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-05 | completed / CHANGES_REQUESTED |
| iter02-security-luna-05 | security | gpt-6-luna/medium | unavailable | re_review | candidate-02-05 | cancelled / none |
| iter02-release-sol-05 | release | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-05 | completed / APPROVED |
| iter02-release-luna-05 | release | gpt-6-luna/medium | unavailable | re_review | candidate-02-05 | cancelled / none |
| iter02-behavior-sol-06 | behavior | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-06 | completed / APPROVED |
| iter02-behavior-luna-06 | behavior | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-02-06 | completed / APPROVED |
| iter02-data-sol-06 | data | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-06 | completed / APPROVED |
| iter02-data-luna-06 | data | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-02-06 | completed / APPROVED |
| iter02-architecture-sol-06 | architecture | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-06 | completed / APPROVED |
| iter02-architecture-luna-06 | architecture | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-02-06 | completed / APPROVED |
| iter02-tests-sol-06 | tests | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-06 | completed / APPROVED |
| iter02-tests-luna-06 | tests | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-02-06 | completed / APPROVED |
| iter02-security-sol-06 | security | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-06 | completed / APPROVED |
| iter02-security-luna-06 | security | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-02-06 | completed / APPROVED |
| iter02-release-sol-06 | release | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-06 | completed / CHANGES_REQUESTED |
| iter02-release-luna-06 | release | gpt-6-luna/medium | unavailable | re_review | candidate-02-06 | cancelled / none |
| iter02-tests-sol-07 | tests | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-07 | completed / APPROVED |
| iter02-tests-luna-07 | tests | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-02-07 | completed / APPROVED |
| iter02-release-sol-07 | release | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | candidate-02-07 | completed / APPROVED |
| iter02-release-luna-07 | release | gpt-6-luna/medium | gpt-6-luna/medium | re_review | candidate-02-07 | completed / APPROVED |
| main-integration-release-sol-01 | installation/distribution integration | gpt-6.1-sol/high | gpt-6.1-sol/high | re_review | 698aa45e747ce18b5f8f0360fe4229d6dfb8bf07 | completed / CHANGES_REQUESTED |
| main-integration-release-luna-01 | installation/distribution integration | gpt-6-luna/medium | gpt-6-luna/medium | re_review | 698aa45e747ce18b5f8f0360fe4229d6dfb8bf07 | completed / APPROVED |
| main-integration-release-sol-02 | installation/distribution integration | gpt-6.1-sol/high | gpt-6.1-sol/high | verification | 7e09f417af8e5e681a14b18fbc1c827fa26cc3ec | completed / APPROVED |
| main-integration-release-luna-02 | installation/distribution integration | gpt-6-luna/medium | gpt-6-luna/medium | verification | 7e09f417af8e5e681a14b18fbc1c827fa26cc3ec | completed / APPROVED |

## Evidence retention

The two large report/log archives were removed from the current tree under the revised retention requirement. This iteration retains 22 findings and all 74 review outcomes with shared reviewer/model discovery and separate verification credit. Full original reports, candidate manifests and raw build logs remain available in [historical committed evidence](https://github.com/NiltonDuarte/AgenticLens/blob/0e9cb88c24f8a10eb1f35840c1ce23223efdd444/alspaces/AL-68/iterations/02/REVIEW_RUNS.json); they are no longer replicated in active workspace files. Original harness transcripts remain outside Git.

The structured finding files were subsequently removed after transferring their substantive finding evidence and resolutions into this review. Reviewer/model credit and original execution outcomes remain above; detailed original metadata remains available in Git history before this cleanup.

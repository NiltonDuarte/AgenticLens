# AL-79 distribution verification

Status: scoped POSIX suites passed; scripts/tests changes frozen. This is test-engineering evidence, not consolidated approval, native Codex capture proof or the required integrated make gate.

## Revision and boundary

Baseline HEAD: 4df45967c801824500ac2ea39a8f52fa9cbdcf66 plus the active AL-79 application/setup/test diff. These checks ran during implementation corrections; no immutable full-candidate revision was captured at command start. Reuse requires the parent to map subsequent changes; final verification must use its frozen candidate. The final owned-test content fingerprint is SHA-256 `bd1bcfddce148b765b0cc4c37ceee5c6a24ac68508b0b1e144854e3e32908a19` over the six paths listed below, each encoded as path, NUL, file bytes, NUL in that order.

- scripts/tests/distribution_fixture.py
- scripts/tests/fixtures/entrypoint_probe.go
- scripts/tests/test_capture_distribution.py
- scripts/tests/test_installation.py
- scripts/tests/test_windows_distribution.py
- scripts/tests/test_windows_plugin_transport.py

## Executed evidence

| Command | Result | Scope and limits |
| --- | --- | --- |
| python3 scripts/tests/test_capture_distribution.py | PASS, 17 tests, 9.286 seconds, exit 0 | Latest affected run includes actual native installer/dispatcher, native development selection and the unchanged-dispatcher fix. |
| python3 scripts/tests/test_installation.py | PASS, 10 tests, 9.246 seconds, exit 0 | Includes actual noninteractive missing-choice rejection, unchanged scaffold/local preservation and linked-worktree setup refusal. Ran before later dispatcher-only corrections. |
| python3 scripts/tests/test_release_revision.py | PASS, 14 tests, 2.795 seconds, exit 0 | Release revision/asset workflow checks; no release publication. |
| python3 scripts/tests/test_test_orchestration.py | PASS, 6 tests, 50.633 seconds, exit 0 | Existing Makefile orchestration contract. |
| python3 -m py_compile scripts/tests/*.py | PASS, exit 0 | Syntax of all Python distribution/transport tests, including final Windows fixture changes. |
| git diff --check -- scripts/tests | PASS, exit 0 | Owned test diff whitespace. |
| python3 scripts/tests/test_windows_distribution.py | NOT EXECUTED: native gate exits 1 before tests | Exact diagnostic: Real Windows PowerShell required; run this distribution gate on Windows or explicitly via WSL interop. No native Windows/PowerShell runtime or WSL interop was available. |

Raw command logs remain outside Git at /tmp/al79-distribution-capture.log, /tmp/al79-distribution-install.log, /tmp/al79-distribution-release.log, /tmp/al79-distribution-orchestration.log and /tmp/al79-distribution-windows.log. Compiler fixture binaries and adopting projects are temporary, independent of ticket workspaces. Native Codex trust/root/child acceptance belongs to NATIVE_EVIDENCE.md and is not asserted by these fixtures.

## Test changes and assertion transfers

- The common POSIX fixture now explicitly sets HOME, USERPROFILE, CODEX_HOME and all three XDG configuration/cache/state roots under its disposable directory. Windows fixtures explicitly pin those destinations using native paths and propagate them through WSL forwarding; plugin transport reuses those explicit destinations. Native install_root and wrapper destinations remain separately pinned. These runs modified no real user hook/config/state files.
- The two shipped installer entrypoints now assert Codex routes to native install-user without frozen source/state/project callback arguments. Existing literal store/argument injection coverage is retained through the unchanged OpenCode install call; unrelated argument, stdin, exit, cwd, signal, integrity, ownership and cache checks remain.
- Two actual application boundaries extend capture distribution coverage: additive user-home hook installation preserves an unrelated SessionStart group, yields exactly six owned event groups, has project/source-free native commands, creates no project hook directory/state, and is idempotent. Real installed native selection prefers a different-identity development build, rejects present invalid output without fallback, and returns to the exact cache after output removal. No fixture result is claimed as runtime hook execution.
- Completion retains an existing valid explicit disabled local policy byte-for-byte; the old comment-only local file was invalid under the selected-file contract and is no longer used as a supposedly valid fixture. Missing noninteractive choices reject without creating local configuration or changing the shared scaffold. Required state diagnostics are now exercised by deleting local state rather than shared state. Bootstrap asserts no shared capture state and no guessed local file.
- The existing linked-worktree installation test now also invokes the actual native setup command and proves worktree local creation and base/shared mutation are refused. Original exact pin, source bundle and repository-executable assertions remain.
- Windows actual application fixtures now keep capture/source/state choices in explicit local policy. Six-group ownership assertions read the installed user hook file rather than the retired large project manifest. Initial manifest stdout is no longer required to exceed 8192 bytes because the native user installer returns a compact result; the existing fault-injected apply transport test still exercises large diagnostic drainage, overflow, failure, timeout, owned-process termination, retained handlers and recoverable backups. The shared probe forwards native dispatch through the actual application only for that explicit fault-injection mode and supplies version JSON/protocol probes; it does not copy resolver policy.

## Concrete finding and verification

The distribution test engineer (gpt-6.1-sol, high; inherited execution profile) traced a reachable unchanged-dispatcher rewrite: the wrapper's dispatcher process remains alive while selected-native install-user attempted to replace that same dispatcher even when its owned bytes matched. The coder confirmed the source path and now filters unchanged owned writes after ownership validation. Actual installed-wrapper reinstallation passes a same-file identity assertion and preserves hook bytes on POSIX. The equivalent Windows assertion was added but remains unexecuted; Windows sharing behavior is not claimed proven. No additional runtime finding or independent review verdict is asserted.

## Remaining limits

The parent still owns integrated make, revision-aware independent review and real normal-trust native Codex root/child publication. Native Windows distribution/plugin transport remains unavailable and explicitly incomplete. Static edits and Linux process results do not substitute for Windows execution.


## Architecture-review correction verification

The parent reopened the source/test freeze for confirmed ARCH-01, ARCH-02 and ARCH-03 from the C1 architecture discovery review (reviewer-architecture, gpt-6.1-sol/high; dispatch acknowledged, source-observed identity unavailable). The parent independently reproduced those defects. The following are regression authoring/verification results, not independent discovery credit or consolidated approval.

Corrections remained on baseline HEAD 4df45967c801824500ac2ea39a8f52fa9cbdcf66 plus active AL-79 changes. The coder declared Linux root-owned dispatcher authority stable before the latest affected suites. These are still pre-commit correction results; the parent owns the final C2 frozen fingerprint and integrated gate. Final owned-test content fingerprint using the six paths and encoding above: SHA-256 `3cc1735b14152030a472af1fce248d935972382123af234d6f2241a52c559bdc`. Earlier fingerprints/results remain evidence of their earlier revisions.

| Command | Result | Evidence |
| --- | --- | --- |
| python3 scripts/tests/test_capture_distribution.py | PASS, 18 tests, 15.388 seconds, exit 0 | /tmp/al79-distribution-capture-c2-root.log; complete affected suite after canonical root-owner correction. |
| python3 scripts/tests/test_installation.py | PASS, 10 tests, 8.018 seconds, exit 0 | /tmp/al79-distribution-install-c2-root.log; affected actual setup/downloader suite after that correction. |
| python3 scripts/tests/test_capture_distribution.py CaptureDistributionTest.test_actual_user_upgrade_keeps_one_dispatcher_authority_and_safe_uninstall | PASS, 1 test, 4.939 seconds, exit 0 | /tmp/al79-distribution-upgrade-final.log; final assertion-order adjustment makes C1 fail at the actual upgrade rather than merely an absent new-format authority file. |
| python3 -m py_compile scripts/tests/*.py | PASS, exit 0 | Final Python changes, including unexecuted Windows scenarios. |
| git diff --check -- scripts/tests | PASS, exit 0 | Final owned changes. |

ARCH-01: the new actual downloader test installs native A with an unrelated user hook, upgrades to B through the shipped production installer and verifies preserved A cache bytes, one set of owned groups, receipt/digest consistency and selected B. It then updates to C from a second effective CODEX_HOME sharing the same cache. The root dispatcher-installation.json authority tracks the installed bytes; the first home's historical receipt stays unchanged and cannot veto its later uninstall. Edited dispatcher bytes are refused and preserved; an edited owned hook also prevents removal without changing receipts or executable bytes. Successful removal preserves the unrelated hook, root ownership and every second-home hook/receipt. No resolver or transaction policy is reimplemented in the fixture.

ARCH-02: the existing legacy fixture test retains ordinary exact-pin support and capture-capability refusal, then installs an actual native dispatcher from another project and selects a separate exact legacy pin lacking new interfaces. Its installed-wrapper manual version invocation succeeds without capture-local configuration, project hooks or network. This exercises the previously missing native-adoption coexistence branch.

ARCH-03: the existing shell-wrapper contract test retains its original argv/stdin/stdout/cwd/exit/signal assertions and additionally installs the actual candidate dispatcher. A compatible compiled reporting probe is selected through the shared development selector; it exposes process cwd/argv and writes its existing fixture log to a relative path. Calls from nested cwd both with and without explicit --al-project preserve exact input, exit 23 and relative arguments; the relative file appears only under nested cwd. The probe does not implement project/executable selection, so the production dispatcher is exercised.

Native Windows upgrade and nested-cwd scenarios were added to test_windows_distribution.py. They assert root launcher identity plus active_native/active_hash consistency, modified launcher/active-build refusal, foreign-hook preservation, safe uninstall and relative I/O/stdin/exit through an actual native dispatcher. They remain unexecuted because native Windows PowerShell/interop is unavailable. The stable-launcher/retained-active-build Windows correction has source/cross-compilation evidence only; no native Windows safety claim follows from these syntax checks. Existing Windows apply transport and modified-owned assertions were retained.

The first targeted run during source corrections had upgrade failure at initial installation while ownership/downloader changes were not yet coherent, with the two manual regressions already passing. The intermediate coherent per-home transaction subsequently passed capture18/install10. Those results were superseded for current ownership verification by the root-authority runs above; they are not silently presented as final C2 evidence. Release/orchestration checks were not repeated because these fixes did not change their owning workflows/Makefile; the parent will still run required make.

Scripts/tests and this evidence are now frozen for parent integration and independent re-review. Production fixes remained with the coder; no commits, telemetry, PR operations or native user-home changes were performed by this tester.


## C3 review-correction verification

The parent explicitly released C2 freeze for the confirmed correction tranche. Test ownership remained scripts/tests only. The coder declared the affected installer, POSIX process and Windows transport/formatting boundaries stable before these commands. Other Go-level review corrections were still being verified; these are pre-commit results on baseline 4df45967c801824500ac2ea39a8f52fa9cbdcf66 plus the active C3 diff, not an immutable final-candidate or consolidated approval. Final owned-test fingerprint over the six paths/encoding specified above: SHA-256 `943b1801424bfe0dfda5b9f34f5d8a0ebb333f1c137c9f3808dfb828a22af139`.

| Command | Result | Evidence |
| --- | --- | --- |
| python3 scripts/tests/test_capture_distribution.py CaptureDistributionTest.test_adopted_native_dispatch_preserves_exec_cancellation_and_signal_exit CaptureDistributionTest.test_incompatible_user_dispatcher_is_rejected_without_installation_mutation CaptureDistributionTest.test_user_install_diagnoses_base_legacy_callbacks_and_migrates_only_owned_groups | PASS, 3 tests, 6.660 seconds, exit 0 | /tmp/al79-distribution-c3-targeted.log |
| python3 scripts/tests/test_capture_distribution.py | PASS, 21 tests, 20.306 seconds, exit 0 | /tmp/al79-distribution-capture-c3.log; includes unchanged C1/C2 upgrade, cross-home ownership, legacy manual and nested-cwd controls. |
| python3 scripts/tests/test_installation.py | PASS, 10 tests, 8.854 seconds, exit 0 | /tmp/al79-distribution-install-c3.log |
| Windows amd64 build_probe process fixture cross-compilation | PASS, exit 0 | Native Go fixture with the new readiness/self-signal behavior compiled for Windows; no Windows process was executed. |
| python3 -m py_compile scripts/tests/*.py | PASS, exit 0 | Includes final Windows scenarios. |
| git diff --check -- scripts/tests | PASS, exit 0 | Final owned diff. |

ARCH-04: a new actual installed native-dispatch test selects the compiled reporting probe through the production development resolver. The final command signals readiness and its PID; interface probes return before this instrumentation. The exposed wrapper PID must equal that selected native PID, SIGTERM must terminate/reap it without a surviving orphan, and a selected command terminating itself with SIGTERM must retain the native negative-signal status rather than becoming exit255. Cleanup explicitly kills any observed fixture-only child on a failing pre-fix path. The existing bootstrap signal, adopted cwd, relative I/O, arguments and stdin/exit assertions were preserved.

BUILD-01: the actual native install-user command rejects the system native true executable, a host-native wrong-interface fixture and a foreign-platform native fixture. Each is exercised before initial installation and after valid owned user installation. Exact cache/home directory entries, file modes and bytes must remain unchanged; the valid installed wrapper still reports its selected pin after each refusal. This proves refusal before owned root/home mutation, rather than merely an error code after breaking installation. The Go tester separately owns low-level validation and journal assertions.

BEH-01: an actual legacy capture installation is created in a disposable Git base checkout with explicit local policy. User install from both the base and a linked worktree must diagnose active base callbacks and leave cache, home, project hooks and existing state unchanged. Explicit migration with an edited owned callback is refused without mutation. After restoring its exact owned bytes, explicit base migration removes only the attested legacy groups, retains an unrelated project callback and pending-state sentinel, marks the legacy receipt inactive, installs the six user groups and creates no worktree hook file. Real root/home interruption, locking and database retention oracles remain Go-test-owned; the sentinel is a preservation fixture, not native source evidence.

ARCH-05: the existing native-adopted Windows cwd/stream scenario now also carries intentional empty values, apostrophes, embedded quotes, trailing backslashes and relative arguments under both inferred and explicit project selection, with LF/mixed newline/Unicode/empty stdin. The original bootstrap transport test remains. The offline advisory timeout test now runs both before and after native adoption and checks bounded timeout plus termination of the observed final native child; it retains the original wall-time oracle around the wrapper's 1.5-second advisory mechanism. The probe's intentional hanging capture-check branch runs before its explicit apply proxy so that the adopted route reaches the intended blocked child.

SEC-01: the new Windows test obtains feedback through the actual native dispatch codex-hook route using synthetic fixture source metadata, then extracts and executes the emitted PowerShell binding command itself. It substitutes only the explicit ticket placeholders and appends the already-rendered repo argument as instructed by that feedback; it does not reconstruct capture-bind argv. Project and selected cache paths contain spaces/apostrophes. The resulting registration must contain the exact supplied ticket, title and source session. Fixture workers are bounded and cleanup stops only executables under that fixture's unique directory. This is command-boundary fixture evidence, not genuine native-harness trust/capture evidence.

ARCH-06: the existing actual Windows A-to-B upgrade/foreign-hook/modified-artifact test now deletes only the stable launcher, explicitly repairs B and requires stable bytes/hash/checksum, hook receipt and retained active artifact/hash to agree. Installed-wrapper version and unchanged hooks must remain usable after repair. Earlier edited-launcher and edited-active-build refusal assertions remain.

All Windows scenarios above remain unexecuted locally: native Windows PowerShell/interop is unavailable, as previously recorded. Syntax and cross-compilation do not establish PowerShell behavior, loaded-image sharing or native binding success. The parent owns the real Windows CI gate and final revision-aware review. No runtime Windows pass, new finding discovery credit, telemetry/commit/PR work or real user-home mutation is claimed by this correction work.

Scripts/tests are now frozen pending a concrete new failure or parent release. Release/Makefile orchestration checks were not repeated because their owning workflow sources did not change; integrated make remains parent-owned.


## C4 residual migration verification

The parent released the narrow C4 correction tranche after C3 review. This extends the existing migration test; it creates no new test declaration or broader platform matrix. The residual was discovered by behavior reviewer gpt-6.1-sol/high during C3 re-review under BEH-01; this tester supplies durable regression verification, not additional discovery credit.

The actual installed SessionStart callback is copied and only its statusMessage is changed; its executable command/argv remains identical. Explicit migration from both the current base and a linked worktree must reject that unknown remaining callback before global activation. Known/unknown hook bytes, receipts, pending facts and installation contents stay unchanged; the old receipt remains active and the global hook file stays absent. The fixture then models the user explicitly removing only the unknown copy. The unchanged-owned migration succeeds, retaining the unrelated callback and pending state. Existing no-flag, modified-owned and clean migration controls remain.

Command: python3 scripts/tests/test_capture_distribution.py CaptureDistributionTest.test_user_install_diagnoses_base_legacy_callbacks_and_migrates_only_owned_groups. Final result: PASS, 1 test, 4.354 seconds, exit 0. Raw result: /tmp/al79-distribution-migration-c4-final.log. Final Python syntax and scripts/tests whitespace checks passed. Other C3 capture21/install10 results are retained for unchanged scopes; no new full-suite or Windows runtime pass is asserted. Execution used baseline HEAD 4df45967c801824500ac2ea39a8f52fa9cbdcf66 plus active C4 source/test corrections after the coder declared this source cluster ready. Parent owns the final frozen revision and integrated verification.

A fixture-oracle correction was necessary: strengthened affected-home locking legitimately creates the first empty project .codex/install.lock coordination anchor before a refused explicit plan. The test admits only that newly created regular mode0600 zero-byte anchor on the first modified-owned refusal, then includes it in all later strict snapshots. No hook/fact/receipt/pending-byte exception was added. Snapshot diagnostics now compare path/type/mode/size/SHA-256 integrity tuples rather than embedding entire native binary byte strings; this preserves byte-integrity detection while bounding failure output.

The initial attempt encountered that anchor difference and spent approximately4GB RAM formatting unittest's binary dictionary diff. Only the owned Python process was stopped after observing no product children; exit143 is unverified, never a pass. No broad environment cleanup was performed. A bounded-diagnostic rerun then failed normally on the expected lock-anchor mismatch (1 test, 3.971 seconds, exit1; /tmp/al79-distribution-migration-c4-digest.log). The narrowly corrected final oracle passed as recorded above. The terminated attempt's raw log remains /tmp/al79-distribution-migration-c4.log.

Final six-file test fingerprint using the documented path/NUL/content/NUL encoding: SHA-256 `b2db33bd63a6d995ef69d66de7b1b641ebfb8c0233682a25d1a22681b760653b`. Scripts/tests and this evidence are frozen again; no application edits, commits, telemetry or publication were performed by this tester.

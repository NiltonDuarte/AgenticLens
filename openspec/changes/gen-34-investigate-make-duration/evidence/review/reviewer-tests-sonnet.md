# GEN-34 test-effectiveness review (RT-SONNET, first pass)

- Run: RT-SONNET, phase first pass, profile sonnet_5_5_high (requested claude-sonnet-5-5, high effort). Runtime identity not independently verified. Native subagent of the orchestrator's Claude Code session; agent/session ID not exposed to this subagent.
- Reviewed revision: head `66a62c9d`, base `ab42d501` (`git diff ab42d501 66a62c9d -- backend`, test-only). Working tree clean apart from the untracked `evidence/review/` directory. Production code and tests were not modified.
- Scope: do the converted tests still detect what they were written for, and is every coverage loss stated and acceptable.
- Outcome: **APPROVED** (scoped to test effectiveness), with four low-severity suggestions and one inaccuracy in the stated coverage loss. No blocking finding.

## Verification I ran (all in disposable copies under the scratchpad; the repository tree was not touched)

A `git archive 66a62c9d` copy was used for every mutation. Mutated sources were never written into the repository.

1. Replacement-identity mutation matrix (`TestLoadTicketWorkspaceRejectsReplacementIdentity`, precompiled non-race test binaries, 25 runs per cell, 400 reads = candidate, 1500 reads = baseline). The three guards that stand between a swapped directory and a wrong ledger are: (c) `sameTicket` in `LoadTicketWorkspace` (`repository.go:180`), (b) `location.directoryInfo` SameFile in `openResolvedWorkspaceDirectory` (`workspace_locations.go:295`), (a) root/dir SameFile in `ReadPublicationDirectory` (`capture/publication.go:246`).

   | Guards removed | 400 reads (candidate) | 1500 reads (baseline) |
   |---|---:|---:|
   | control (none) | 0/25 fail | not run (tester: pass) |
   | c only | 0/25 | 0/25 |
   | c + a | 0/25 | 0/25 |
   | c + b | **25/25 fail** | **25/25 fail** |
   | c + a + b | **25/25 fail** | **25/25 fail** |

   With (a) and (b) removed and (c) kept, one 1500-read run showed `identity_conflict` on 97 reads, so a wrong ledger reaches the oracle on roughly 6 % of all reads, about 26 times per 400-read run. Miss probability is negligible at 400 reads.
   (My first attempt mutated the wrong source lines; I caught it from `grep "false &&"`, discarded those results and redid the matrix with the exact lines. Only the table above is reported.)
2. FIX-21 budget property at divisor 8. Mutation: `CreationWindow.Prunes()` returns false (child sweep reads every header). `TestLargeStoreFamilyWithinCaptureBudget` fails at divisor 8 and 4: `sweep charged 8519676 bytes ... want 622592` and `pass 1: bytes=14775548 ... want 6944000`.
3. Large-store test at divisors 8, 9 and 16 (non-race): all pass; fixture sizes 261, 234 and 143 files.
4. Converted-test subset in the real tree (non-race, fixture built once): `TestIncompatibleUserManager...`, `TestRetainedAdmission...` (3 cases), `TestUserHooksDoNotOwnOrReplaceManager`, `TestWorkerSuccessor...` (6 cases incl. missing native), `TestCompatibleGlobalWrapperUpdate...`, `TestInstallCallerCancellationDuringOptionalProbe`, `TestDiagnosticHeadLookup...` (32 and 128): all pass. Replacement test passes, 27 successful reads in that run.
5. `go vet ./internal/adapters/capture/...`: clean.
6. Not rerun by me: the full `capture` race run, the integrated `make` (parent owns), the tester's fixture mutation table (inspected, not repeated).

## Hypotheses and disposition

| # | Hypothesis | Disposition |
|---|---|---|
| H1 | A test relied on the old binary answering unknown commands (the test binary falls through to `m.Run` for arguments it does not recognize) or on probes the fixture does not answer. | Refuted. The tester's invocation log (184 passing tests) shows only commands in the fixture table. `grep` finds no remaining `AL_CAPTURE_INSTALL_FIXTURE` and no protocol probe sent to `os.Executable()`; remaining `os.Executable()` uses (`install_native_validation_test.go`, the `opencode` symlink, subprocess re-runs) bring their own helper env/args. The fixture now exits 2 on unknown arguments, which is stricter than before. |
| H2 | A test relied on binary size (35 MB vs about 2.4 MB). | Refuted for assertions. Only size-dependent code is `Size()/2` truncation and last-byte flip in `TestRetainedAdmission...`, which are relative. No test sits near `artifactLimit` (256 MiB) or the 1 MiB journal/receipt limits. The reduction of exercised copy/journal volume is stated in the design. |
| H3 | A test relied on the old binary being race-built or containing production code. | Refuted. The old binary only answered probes from `init`/`TestMain`; no production code ran in the child. Race coverage of in-process production code is unchanged. |
| H4 | A test relied on the old binary being the running executable (ETXTBSY protection, same inode). | Refuted. Self-update and `installTestManager` copy the source; `installFixture` and relocation symlink it. `os.Remove(binary)` in the missing-native case removes the symlink only. No test opens the fixture for writing. |
| H5 | The shared read-only fixture leaks state across tests. | Refuted with a residual limit. Grep finds no `Chmod`, `Truncate`, `WriteFile(binary)` or `Chtimes` against fixture or symlink paths; the mutable-source test copies bytes first. The 0555 mode is not a barrier for root, but nothing writes. Behavior variants come from per-test env (`AL79_CAPTURE_NATIVE_FIXTURE`), not the file. Cleanup in `TestMain` runs before `os.Exit`. |
| H6 | Fixture build depends on cwd/PATH/env that a test changes before the first call. | Refuted. No `Chdir`; `PATH` is only prepended; the build env is captured at package init; no test runs in parallel. |
| H7 | The fixture is more permissive than the old binary and hides a negative case. | Partly true, harmless today (RT-02). |
| H8 | The 400-read test lost its ability to detect the TOCTOU regression it protects. | Refuted by the mutation matrix. It detects the realistic multi-guard regression 25/25 at both sizes, and it never detected single-guard removals at either size. |
| H9 | Divisor 8 no longer exercises the FIX-21 property. | Refuted for the property, with a margin note (RT-03). Oracles are exact byte/claim equalities, the store-to-budget ratio of 5.8 is unchanged by design, and removing window pruning is caught. |
| H10 | {32,128} cannot show head-lookup independence. | Refuted. See RT-04. |
| H11 | Wrong-reason pass of the `bad-manager` test. | Confirmed as pre-existing weakness, now a bit more reachable (RT-01). |

## Findings

### RT-01: `bad-manager` assertion accepts any error (suggestion, severity 2)
- Location: `backend/internal/adapters/capture/user_install_test.go:305-308` (`TestIncompatibleUserManagerDoesNotCreateOrChangeOwnedState`), fixture `testdata/nativefixture/main.go`.
- Status: pre-existing assertion; the new fixture adds a failure mode that did not exist before (exit 2 on an unsupported argument).
- Trigger: the fixture's `bad-manager` branch stops matching the probe production sends (for example a probe is renamed), so the fixture exits 2.
- Consequence: `InstallUserCodex` fails with `exit status 2` and the test passes while no longer checking manager-interface rejection. Other tests that use the compatible variant would fail first, so exposure is small.
- Evidence: the test checks only `err == nil`; tester scratch logs show today's reason `native manager interface incompatible: "0"`. The tester's mutation "bad-manager answers compatible" does fail the test, so the real regression is caught.
- Suggestion (parent/tester): assert the error text contains `interface incompatible`.

### RT-02: fixture answers more than the old binary did, and has unused entries (suggestion, severity 1)
- Location: `testdata/nativefixture/main.go` `respond`.
- Status: introduced by the fixture, behavior-preserving for every test run.
- Detail: the old `installFixture` path answered four commands (`engine protocol`, `capture protocol`, `version --short`, `version --json`); the new fixture answers manager and worker probes with no variant set. A future negative test that expects the native engine to be rejected as a manager would not fail as before. The entries `dispatch protocol`, `manager codex-binding-wake-protocol` and variant `old-interface` are never called (tester log).
- Consequence: none now; dead table entries cannot be shown to be protected by any test.
- Suggestion: keep or drop knowingly; optional.

### RT-03: divisor 8 leaves only 5 files above the 256 example session budget (suggestion, severity 2)
- Location: `reconcile_large_store_test.go:49` and the `exampleSessionBudget` sub-case.
- Evidence: at divisor 8 the fixture has 261 files; at 9, 234; at 16, 143. The sub-case "family extracted within budget at max_sessions 256" is meant to run against a store larger than the session budget, but nothing asserts it, and the flag still accepts up to 16.
- Consequence: no detection loss today, because assertions are exact byte and claim counts and the mutation check fails at divisor 8. It would only matter if someone raised the default or the flag beyond 8 and relied on the 256-session case being binding.
- Suggestion: either assert `len(store.Files) > exampleSessionBudget` in that sub-case or note the margin in the flag comment.

### RT-04: the 256-owner case adds no oracle (informational, severity 0)
- Location: `diagnostic_index_test.go:108`.
- Evidence: the oracles are `HeadLookups == 8` (bounded by `Limit` 18 and by the options, not by history size) and the `EXPLAIN QUERY PLAN` text (a property of schema and query, not of data size). A regression of the form `min(count, ...)` or `count/4` shows up at 32 or 128 because the lookup cap is 18 and 32 > 18. 256 exceeds the cap by even more but gives the same signal. Neither oracle measures rows read after the head lookup, so no count (32, 128 or 256) would catch a history scan placed after the lookups; that gap is pre-existing.
- Conclusion: {32,128} is enough for the existing oracle; the loss is real but small and was correctly stated.

### RT-05: the stated race-detection loss for the replacement test is computed on the wrong base (suggestion, severity 2)
- Location: `design.md` Steering "Coverage lost" ("at 1 % per read: 98 % instead of nearly 100 %") versus `TESTS.md` section 3.
- Detail: a wrong ledger can only be observed on a successful read, and successful reads fell from about 122 to about 32 per run (tester data, repeated by me: 21, 27, 41). A 1 % chance per successful read would give about 27 % (0.99^32 is 0.72) instead of about 71 % (0.99^122), not 98 %. The 98 % figure holds only if 1 % applies to each of the 400 total reads.
- Practical impact: small. The realistic regression (two guards removed) produces a wrong ledger on about 6 % of all reads and is caught 25/25 at 400 reads and at 1500. Single-guard removals are survived at both sizes (defense in depth, pre-existing). The sibling `TestLoadTicketWorkspaceDetailKeepsSameIDCopyCoherent` already ran 400 reads.
- Suggestion: restate the loss in successful reads, as TESTS.md does, and cite the mutation result, or leave a single honest sentence in the design.

### Informational, pre-existing
- The replacement test has no `successful > 0` guard (tester gap 2). At roughly 8 % success the chance of zero in 400 reads is far below 1e-14; the guard would only matter if a read path change made all reads fail, which would also make the test vacuous at 1500 reads.
- A single-guard regression in the replacement identity path (any one of the three) is invisible to this test at either read count; it passes by design because each guard backs up the others. This is not caused by the change.

## Coverage loss statement check

The design and TESTS.md state: the 256-owner case, replacement-read volume, large-store scale below divisor 8, and the smaller copy/journal payload. I found no additional undeclared loss that changes a detection result. Smaller items (broader fixture answers, unused fixture entries, the 5-file margin) are RT-02 and RT-03. Coverage gained but unstated: native probes no longer share the 2 s validator timeout with a race-built 35 MB test binary start-up, which reduces timeout-driven flakes in these tests (the tester shows the relocation flake still occurs under load).

## Gaps

- I did not rerun the full `capture` race package, the tester's per-fixture-answer mutation table or the integrated `make`; I relied on the tester's tables after inspecting the invocation log and mutation definitions.
- The replacement mutation matrix used non-race binaries; race builds change the read rate but not the detection logic (race builds were compiled for the earlier, discarded mutation run only and are not part of the result).
- No Windows or macOS evidence; the fixture tests are Unix-only in practice, and the fixture is not removed if a run is killed before `TestMain` finishes (pre-existing pattern).
- `design.md` still says "Delivery status: approved. Not yet implemented." and `tasks.md` boxes for the fixture, scale and verification sections are unchecked, although the code is committed. This is orchestrator-owned artifact status, noted for consolidation.

## Verdict

APPROVED for test effectiveness. The fixture is a real native ELF so production format, checksum, journal, ownership and protocol checks still run, and the converted negative cases fail for the asserted reasons (tester mutations plus my reading). The three scale reductions keep their oracles; mutation checks show the replacement test and the FIX-21 budget test still detect the regressions they were written for. Parent decides whether to fold RT-01, RT-03 and RT-05 into the delivery notes or as small follow-ups.

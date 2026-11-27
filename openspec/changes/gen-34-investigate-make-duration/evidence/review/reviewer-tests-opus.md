# GEN-34 test-effectiveness review (RT-OPUS, first pass)

**Outcome: APPROVED** for the assigned focus. The converted tests still detect the regressions they were written for. Each coverage loss is stated in `design.md` and is acceptable. Two pre-existing weak assertions and one cleanup are listed as suggestions; none blocks delivery.

## Run identity

- Run: RT-OPUS, first pass, focus reviewer-tests. Requested profile: opus_5_5_medium (claude-opus-5-5, medium effort). The runtime model is the one the harness reports. Agent/session ID: not exposed to this subagent.
- Reviewed revision: head `66a62c9d` against base `ab42d501` (`git diff ab42d501 66a62c9d -- backend`). Production code is unchanged in this diff.

## Hypotheses and disposition

| # | Hypothesis | Disposition |
|---|---|---|
| H1 | A test relied on the installed native being the running test executable (ETXTBSY), so a write-through regression is no longer caught | Rejected. `TestCompatibleGlobalWrapperUpdatePreservesOtherPinnedProject` compares the native's bytes before and after, which does not depend on ETXTBSY. A write through the symlink to the 0555 fixture fails with EACCES, and a replacement at the symlink path changes the bytes read. Either way the test fails. |
| H2 | Production chmods or writes through the `installFixture` symlink and changes the shared fixture for later tests | Rejected. The only path-based chmod on an engine leaf (`manager/install.go:152`) is reached after `readRegular`, which uses `Lstat` and rejects anything that is not a regular file (`manager/storage.go:211-221`). The other chmod calls act on temporary files or on transaction targets in the user's home. A regression of this kind would make tests fail, not pass silently. |
| H3 | A test relied on the size of the old binary, its race build or its contents (production code) | Rejected. The retention test is driven by checkpoints (truncate to half size, flip the last byte, restore the modification time), so it does not depend on size or timing. Format, architecture, checksum and protocol checks are fixture-agnostic. The tester's mutation table shows that each probe answer and the script substitute are rejected by production. |
| H4 | The old binary answered unknown commands by running the test suite, and some path relied on that | Rejected. In the tester's invocation log every call during a full run matches a fixture entry. The unset-variant paths receive exactly the four commands that the removed `TestMain` branch answered. |
| H5 | The 400-read replacement test no longer detects an identity leak | Rejected by a mutation I ran (below). With both identity guards removed, all 8 of 8 race runs failed by read 79 at the latest, and most failed in the first 20 reads. |
| H6 | Divisor 8 no longer exercises the FIX-21 budget property | Rejected. Each assertion is an exact byte and claim equality against a per-file oracle, so the result does not depend on scale. The divisor scales the number of unrelated rollouts, the family sizes and the budget together, so the store-to-budget and family-to-budget ratios stay the same. At divisor 8 the store is still 5.8 times the budget. |
| H7 | {32, 128} is not enough to show that head lookups are independent of owner count | Rejected. `HeadLookups` counts calls on a fixed lookup schedule bounded by `Limit` (`diagnostic_inspection.go:90-170`). For this query it is 3 + 3 + 1 + 1 = 8 whatever the history size. Each case also checks that the query plan uses an index seek (`SEARCH ... USING INDEX`, no `SCAN`). The 256 case added no distinct oracle. |
| H8 | The shared fixture build or cleanup leaks or breaks in other environments | Minor. The build depends on `go` and on the package working directory, the same pattern `installationManagerExecutable` already uses. Cleanup runs only from the `linux \|\| darwin` `TestMain`, and the tests that use the fixture are Unix-only. After my run no `agenticlens-native-fixture-*` directory was left in `/tmp`. |

## Findings

### RT-01 (pre-existing, suggestion): the bad-manager assertion accepts any error

- Location: `backend/internal/adapters/capture/user_install_test.go`, `TestIncompatibleUserManagerDoesNotCreateOrChangeOwnedState` (`err == nil` is the only check).
- Trigger: the fixture's `bad-manager` branch stops answering, for example after a typo in the command table, so the probe takes the `unsupported arguments` path and exits 2.
- Consequence: the install fails because of `exit status 2` rather than `interface incompatible`, and the test still passes. Now that the answers live in a separately built program, this failure path is a little more likely.
- Evidence: the tester's scratch log shows the intended reason today (`native manager interface incompatible: "0"`).
- Severity: 2. Suggested fix: assert that the message contains `interface incompatible`.

### RT-02 (pre-existing, suggestion): the replacement test neither asserts nor reports which guard it exercises

- Location: `backend/internal/adapters/workspace_replacement_linux_test.go`, `TestLoadTicketWorkspaceRejectsReplacementIdentity`.
- Trigger: the protection has two layers: the `SameFile` check after resolution (`workspace_locations.go:293-299`) and the final `sameTicket` check (`repository.go:180-182`). Removing either one alone survives both 400 and 1500 reads, because the other layer still holds. The test fails only when both are gone. It also does not assert `successful > 0`.
- Consequence: the test protects the observable identity guarantee, not each layer. That was already true at 1500 reads and is unchanged by this diff.
- Severity: 1. Informational: the 400-read reduction does not weaken this.

### RT-03 (introduced, nit): `GORACE` settings left in the converted helpers no longer do anything

- Location: `installation_unix_test.go:320` (`installFixture`), `relocate_install_unix_test.go:27` and `user_install_test.go:246`.
- Trigger: the fixture and the helper-built managers are built without `-race`, so `GORACE` has no effect on the child processes these helpers start.
- Consequence: misleading setup only, with no effect on behavior. It would matter only if `GOFLAGS=-race` were inherited by the fixture build.
- Severity: 1.

No defects of severity 3 or higher were found in the assigned focus.

## Coverage losses: checked against the statement in design.md

| Stated loss | Assessment |
|---|---|
| 256-owner head-lookup case | Acceptable. No distinct oracle: lookup count and query plan do not depend on owner count (H7). |
| Lower race-detection probability per run for the replacement test | Acceptable and stated. My mutation shows that the observable regression (a successful read returning another ticket) appears within about 80 reads, far below 400. The "1 % per read" figure in `design.md` is a hypothetical rate, not a measurement. The parent may add this mutation result to the evidence. |
| Default-scale large-store coverage below divisor 8 | Acceptable. The assertions do not depend on scale (H6). The flag still accepts 1 through 16, and the tester showed that divisor 4 passes. |
| Copy and journal paths handle about 2.4 MB instead of 35 MB | Acceptable. No test asserts or depends on file size or on how long a copy takes. Large-file I/O is not a property these tests were written to check. |

## Verification I ran

All probes ran in a disposable `git worktree` at `66a62c9d` under this session's scratch directory. The worktree was removed afterwards, and no repository file was edited.

1. Mutation A: removed the final `sameTicket` check in `LoadTicketWorkspace`. Race build, 6 runs at 400 reads: all passed (23-48 successful reads), so the `SameFile` layer still protects.
2. Mutation B: removed the `SameFile` check after resolution as well (A + B). Race build, 8 runs: **all 8 failed** with `replacement returned another ticket` at reads 1, 1, 5, 6, 6, 13, 17 and 79.
3. Mutation C: removed only the `SameFile` check. Race build, 4 runs: all passed (68-91 successful reads), so the `sameTicket` layer still protects.
4. Unmutated candidate, without race: `TestUserHooksInstallIdempotentAndOwnedRemoval`, `TestCompatibleGlobalWrapperUpdatePreservesOtherPinnedProject` and `TestIncompatibleUserManager...` passed (9.8 s). No fixture directory was left in `/tmp`.
5. Static checks: traced symlink handling (`readRegular`), the chmod call sites in `capture`, `manager` and `config`, the `HeadLookups` schedule, the `IncidentScaleSpec` scaling, and the build flags of the manager helpers.

I reused the tester's evidence for fixture mutations, the invocation log, read counts and timings rather than repeating it.

## Gaps

- I did not run the full `capture` race package or the integrated `make`; the tester and the parent own those.
- I did not separately check, during a package run, that the shared fixture keeps mode 0555 throughout. H2 rests on the code trace.
- I did not measure the detection rate for a narrower race (for example, a one-instruction window between resolution and read) at 400 compared with 1500 reads. Only the combined removal of both guards was measured.

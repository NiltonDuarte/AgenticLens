# AL-79 architecture correction re-review: AL79-C4

Verdict: **APPROVED** for the bounded C4 architecture correction scope. No blocking architecture finding was substantiated. This is not consolidated approval, integration verification or native Windows runtime proof.

## Identity and revision

- Run: `architecture-sol-c4`; phase: C3-to-C4 correction re-review.
- Requested model/effort: `gpt-6.1-sol` / `high`. Actual runtime model/effort metadata is unavailable in this child context; requested identity is not represented as measured identity, usage or cost.
- Baseline/HEAD: `4df45967c801824500ac2ea39a8f52fa9cbdcf66`.
- Frozen candidate: `/tmp/al79-review-coordinator-b67f2d/CANDIDATE_C4.diff`, SHA-256 `9d1ca49a319643feeb13774a28942037cfff9d175884d53a8e8495660a90e265`. The supplied file and independently recomputed staged binary diff excluding `alspaces` matched. `C3_TO_C4.diff` established the correction boundary. The initial unstaged diff contained only parent-owned `REVIEW.md` metadata; no unstaged product drift was observed.
- Authority: approved AL-79 contract and iteration 01 plan, including the existing-v6 external-parent decision; assigned architecture persona, shared review procedure, FCIS and portable-ledger rules, and relevant regression boundaries. Read own C3 architecture report and raw/shared verification evidence, without loading sibling personas or reports.
- Source/tests were read-only. Writes were limited to this assigned report and disposable `/tmp/al79-architecture-sol-c4` overlay/probe files. No implementation, delegation, commit or publication was performed.

## Correction assessment

### Migration owns exact groups, not an entire checkout

`capture/user_migration.go:legacyUserChanges` now removes only the receipt-attested exact groups from its in-memory plan, then examines all remaining groups before returning changes. Unknown or modified active callbacks cause an error before hook activation or migration application. `user_adoption.go:legacyCaptureGroupsRemain` covers both branded groups and recognizable `capture receive` / `dispatch codex-hook` invocation tokens from unknown executable paths, including `commandWindows`. Recognition establishes a conflict, never permission to delete the unknown command.

The migration-checkout exception in `checkLegacyUserConflicts` therefore delegates to a strict post-recovery group check rather than granting a checkout-wide ownership exemption. An inactive receipt no longer silently exempts still-present callback groups: exact remaining owned groups can be removed, absent owned groups are tolerated, and unknown capture groups still block. A missing hooks file with an already inactive receipt remains an idempotent no-op.

Static trace and the actual installed distribution regression support the correction. An independently executed overlay additionally exercises an unbranded `/tmp/custom-native capture receive --source codex` residual through the real `InstallUserCodex` path, once with an active old receipt and once with an inactive one. Both refuse activation with the unknown/modified diagnostic and preserve every snapshotted legacy/home/root artifact. This closes the assigned architecture interaction without claiming general shell-command equivalence analysis or ownership of arbitrary commands.

### Resource discovery is a lease plan, not recovery authority

`capture/user_install.go:installUserCodexWithCheckpoint` still acquires the root lease, discovers affected current/prior/legacy resources and acquires their leases. C4 then re-reads the root and home journals and the legacy resource receipt while the acquired leases are held. `requireUserLeases` rejects any newly discovered resource with an explicit fresh-lease retry before calling either recovery routine. The final migration plan is separately checked via `userJournalHomes(asUserJournalChanges(...))` before application, so its actual target directories cannot silently exceed the held set.

The retry returns through deferred lock releases; it does not recover another home's bytes under a stale resource set. Existing journals and pending facts remain available for the next invocation. The added checkpoint is an adapter transaction-stage value used by a real interleaving test, not a new domain service or production event queue. The test's second actual installer writes an interrupted prefix into a newly affected home; the first attempt refuses before changing those artifacts or the pending journal, and fresh discovery subsequently recovers and completes. Existing durable-prefix and edited-byte recovery cases remain applicable.

### External-parent analysis preserves facts and qualifies conclusions

`domain/generation.go:HasExternalParentReference` and `completeAnomalySourceCoverage` express the distinction between recorded per-owner analysis and complete parent/group context as pure value computations. `events.go` uses the latter for both projected refresh and durable group-summary replacement; group rework additionally requires the parent to be an owned session. A complete child stream cannot establish fresh whole-group coverage for a parent outside this ticket.

An unavailable source still returns the recorded ledger unchanged. A complete child source can refresh its own analysis, while `PersistSourceAnomalySummaries` leaves the exact existing top-level anomaly array, including nonnil empty arrays, intact when group context is incomplete. `UnrecordedAnomalyScopes` continues counting absent recordings; it does not relabel a recorded empty array as absent. `CompleteAnomalyCoverage` separately refuses an unqualified complete conclusion for the external-parent case. No parent stubs, imported parent metrics, cleared ancestry, persisted fields, schema migration or local authenticity prerequisite were introduced.

The formatter and both CLI diagnostic entry points now consume that distinction: recorded findings remain visible, missing external parent context is disclosed, and empty findings cannot produce an unqualified clean-session claim. Relevant unchanged consumers were traced: `aggregations.go` marks derived health unavailable when coverage is incomplete, while `anomalies.go` still counts recorded findings and limits its complete-ticket denominator. This conservatively affects derived presentation without deleting owner/token facts. The fresh/recorded-empty/recorded-warning domain controls and real independent-ticket publication/Markdown regression match these boundaries.

## C3 carry-forward

- ARCH-01 root ownership and ARCH-06 stable-launcher hashing remain approved. C4 alters lease validation and legacy planning, not dispatcher ownership, artifact hash selection or launcher bytes. The new transaction checks strengthen the prerequisite to those writes. Existing C3 upgrade/uninstall and host-executed Windows-branch evidence remains valid within its original platform limits.
- ARCH-02 manual legacy-pin compatibility, ARCH-03 invocation cwd, ARCH-04 POSIX exec/signal lifetime and ARCH-05 PowerShell argument/handle/advisory behavior remain approved at their previously established boundaries. C4 does not alter dispatch, executable choice, wrappers, quoting or cancellation. No new platform result is inferred.
- ARCH-07 remains refuted, severity 0/10: the approved distinction is between process cwd and the native callback's declared JSON session cwd. C4 changes neither routing nor source-context validation and creates no new routing requirement.
- Retained executable admission, worker identity, selected-local configuration and manual-generation compatibility implementations are unchanged. C3 evidence for their source/consumer paths is reused. The C4 targeted test command retained the relevant prior regression controls.
- Existing-v6 external ancestry remains approved under the explicit plan decision. C4 qualifies analysis completeness while retaining the same ancestry, ownership and portable read contract; it adds no graph/authenticity policy.
- Genuine normal-trust Linux/Codex root/child execution remains evidence for its recorded binary and unchanged callback/context/extraction/binding paths in `NATIVE_EVIDENCE.md`. The C3 installation/read-only preservation check is reusable for unchanged dispatcher ownership and ledger-read semantics. C4's changed migration, recovery race and external-ticket analysis are supported by the distinct deterministic/process evidence above, not represented as newly executed native callbacks.

## Verification and limits

Executed by this reviewer:

- Frozen-file/staged-diff SHA checks and HEAD verification passed.
- From `backend`: `go test -overlay /tmp/al79-architecture-sol-c4/overlay.json ./internal/adapters/capture -run '^TestReviewC4UnbrandedResidualMigration$' -v -count=1`. PASS, two active/inactive subcases, package 0.167s. The overlay adds a disposable test file only; production and shared tests are unchanged. Raw source, overlay and output: `/tmp/al79-architecture-sol-c4/probe_test.go`, `overlay.json`, `probe.log`.

Reused separate verification actors' evidence after inspecting the consequential source/test paths:

- `/tmp/al79-tester-c4/final-targeted.log`: domain 0.024s, capture adapter 21.663s, adapters 0.021s and CLI 12.609s, all passing. Recorded source inventory fingerprint: `a444dacf76685fdab28ae034afcea2312509a1addc1b0a6442fdd41241005dc7`; command and ownership details are in `TEST_EVIDENCE.md`.
- `/tmp/al79-distribution-migration-c4-final.log` and `DISTRIBUTION_EVIDENCE.md`: actual installed migration/conflict test passed, 1 test, 4.354s, with both base and linked-worktree conflict refusal and preserved callback/state artifacts. Earlier failing/terminated attempts remain documented separately and are not treated as passes.
- Own C3 report and named C3 process/overlay evidence for unchanged correction scopes; native evidence remains revision/platform bounded as stated above.

No unresolved material architecture evidence gap remains in this assigned correction scope. Parent-owned main integration, required full `make`/race verification and consolidated review are separate delivery gates. Native Windows runtime remains unavailable locally and requires the hosted Windows gate; prior cross-compilation/host branch checks do not substitute for execution on that platform. This approval grants no merge or publication authority.

# AL-79 architecture discovery review: AL79-C1

Verdict: **CHANGES_REQUESTED**, scoped to architecture, dispatch compatibility and state ownership. Three introduced regressions are substantiated below. This is not consolidated approval or a delivery gate.

## Identity and revision

- Local run: `review_arch_sol_c1`; phase: independent discovery; persona: architecture/maintainability.
- Requested model/effort: `gpt-6.1-sol` / `high`. Acknowledged as the requested execution profile. Exact runtime model/effort metadata is unavailable in this child context; no source-observed identity, usage or cost is asserted.
- Baseline/HEAD: `4df45967c801824500ac2ea39a8f52fa9cbdcf66`.
- Frozen source/guidance/test diff: `/tmp/al79-review-coordinator-b67f2d/CANDIDATE.diff`, SHA-256 `5489a593e9b782aafa6c257efc7c2ae98fcdf2d1b7793759e6a08a8fc34d9a78`. Recomputing the staged binary diff excluding `alspaces` at report time matched that hash.
- Requirements: AL-79 contract, iteration 01 plan, ticket specification and its PO-20 F1-F6 boundaries; canonical FCIS and portable-ledger rules. Product/test source was read-only. No sibling personas or review conclusions were loaded.
- Shared-checkout drift observed at close: unstaged edits in `backend/cmd/agentic-lens/capture_corrections_test.go`, `backend/cmd/agentic-lens/capture_test.go` and appended `NATIVE_EVIDENCE.md`. These do not change the inspected production paths or the three probe outcomes. The frozen staged C1 hash above remains reproducible; broader test claims must identify the tested state.

## Findings

### ARCH-SOL-1: Dispatcher has conflicting ownership authorities during a normal upgrade

Severity **7/10**; **introduced correctness defect**.

Locations: `.agents/skills/agentic-lens-setup/scripts/download.sh:196`, `download.ps1:239`, `backend/internal/adapters/capture/user_install.go:138`.

Trigger: successfully install native user capture using build A, then explicitly install a different approved pin/build B through the shipped downloader, with Codex capture enabled. The downloader verifies the existing dispatcher against `dispatcher.sha256`, replaces the dispatcher with B and updates that digest. The user-installation receipt still records A's `dispatcher_hash`. `InstallUserCodex` gives that receipt hash priority and refuses B as modified/unowned. The downloader has already crossed `CAPTURE_APPLY_STARTED`, so it retains the changed dispatcher and old receipt instead of rolling them back. A later `capture uninstall-user` fails the same ownership check.

Supported expectation: the approved native installation/removal boundary is transactional, preserves owned state and supports explicit pin changes; setup documentation directs users to refresh owned integration after a pin change. A legitimate installer update must not invalidate its own ownership proof. This is conflicting ownership of one machine-wide artifact by two installers, rather than a missing arbitrary repair override.

Evidence: the isolated actual application probe installs A (`v9.8.7`) successfully, installs B (`v9.8.8`) with exit 1 and `native dispatcher was modified or is unowned; preserved`, then observes a receipt SHA `0723fd8845aebca9d0d630174f5b38d28517adfa5bc82d9d20cb7028d9c086ec` versus dispatcher SHA `96b0d0b3ea5227aecfc60826a12859da5c1279b784cb221442c3574bf2bea1d4`. Uninstall exits 1 with the same diagnostic. Source inspection shows the same competing ownership flow in PowerShell; Windows execution is unverified.

Required correction: make explicit dispatcher replacement and user ownership metadata one consistent transaction/ownership operation, retaining protection for actual user edits. Add an A-to-B upgrade/removal regression through the production installer, not only byte-identical reinstallation.

### ARCH-SOL-2: Native dispatcher adoption disables previously supported legacy manual pins

Severity **6/10**; **introduced compatibility regression**.

Locations: `backend/cmd/agentic-lens/dispatch.go:73`, `backend/internal/adapters/config/native.go:66`, `scripts/agentic-lens-wrapper.sh:8` and its shipped setup copy.

Trigger: a machine has the new native dispatcher from one project, while another project explicitly installs/selects an older exact pin that supports ordinary CLI but lacks the new capture/dispatcher interfaces. The common wrapper always delegates to the installed dispatcher. That dispatcher unconditionally requires capture protocol 3 and dispatch protocol 1 from the selected target, including `version --short` and other manual commands. The installer still accepts and reports binary-only legacy CLI support, but the newly installed exact binary cannot be reached through the common wrapper.

Supported expectation: preserve manual generation and compatible shared consumers; explicit binary-only legacy installation remains an intentional supported boundary. `test_binary_only_legacy_pin_is_preserved_but_new_capture_capability_is_required` explicitly verifies ordinary legacy CLI support, but only without an already installed dispatcher. The new routing binary's protocol requirement need not become a capture-installation prerequisite for every manual target binary.

Evidence: after native dispatcher installation, the production downloader installs an exact `v8.0.0` compiled native legacy fixture with capture protocol 1 successfully (exit 0). Its subsequent wrapper `version --short` exits 1 with `unsupported native capture interface`. This fixture represents the same supported legacy boundary as the existing distribution test; it is not claimed to be a historical published release binary. The call trace is wrapper -> native dispatcher -> unconditional `ValidateNativeExecutable` -> rejection before requested manual command.

Required correction: distinguish the interface needed for a manual selected executable from the interfaces needed for automatic capture/development dispatch. Preserve an offline manual path for explicitly installed supported legacy pins after another project adopts native capture. Exercise coexistence in one isolated user environment.

### ARCH-SOL-3: Common native dispatch changes the caller's working directory

Severity **6/10**; **introduced observable regression**.

Location: `backend/cmd/agentic-lens/dispatch.go:90`.

Trigger: invoke the installed common wrapper from a configured project's nested directory. Native selection correctly finds the owning project, but `command.Dir = entry.Project` then runs every requested manual command from its root. The previous shell wrapper preserved the caller's cwd, including when `--al-project` was supplied. The existing `test_wrapper_preserves_arguments_stdin_stdout_cwd_exit_and_exec_signal` asserts that behavior but uses a fixture without native dispatch, so it misses this branch.

Consequence: the same manual command resolves relative arguments against a different directory after native dispatcher adoption. For example, relative ledger/summary destinations passed to `generate` are opened relative to process cwd, so a command previously targeting a nested path can write a different project-root path. Project discovery and source policy do not require changing the manual process cwd; `runGenerate` already resolves its owning project independently.

Evidence: an installed actual dispatcher selects a compatible compiled native probe from nested cwd. Its result is `cwd: .../adopting project`, while the invocation cwd and expected value are `.../adopting project/nested`. Args/stdin remain intact. This isolates dispatch behavior independently of application output formatting. The native callback path already supplies explicit validated project/source context; preserving its project routing does not require imposing its cwd on ordinary CLI calls.

Required correction: preserve caller cwd for ordinary dispatch and route native callbacks explicitly. Extend the cwd regression to exercise an actually installed native dispatcher, including relative arguments.

## Other consequential hypotheses assessed

- **Mutable development output retargets admitted worker children:** refuted in inspected paths. `LaunchWorker` retains executable bytes, `captureWorker` creates a fence on the retained artifact and executes generation children using that retained path. Preferred selection is consulted for a later normal handoff. The updated fence preserves state/source guards without retiring solely for a preferred-build or installation change. Reused the dedicated real-process retention regression evidence; did not repeat the suite.
- **Failed publication creates an unbounded automatic process chain:** refuted for the inspected lifetime path. Backoff is capped, each owner has a configured finite lifetime, and exhausted unsuccessful work returns with the generation pending. Healthy-owner handoff is distinct from failed retry exhaustion. No claim of cache garbage collection is made.
- **Worktree-local policy silently inherits base capture/source keys:** refuted for callback admission. `SelectConfig` chooses one local path; strict `ResolveCapture` decodes it over numeric defaults, while `Load` clears source settings before loading selected-local values. Invalid present local files do not use base fallback. Manual compatibility has a separate resolver; this review does not absorb AL-81's registry-decoupling scope.
- **Native adoption requires a hidden project installation receipt:** refuted for the normal user callback path. `ReceiveCodexUser` avoids installed-wrapper binding lookup; selected local configuration bypasses legacy installation validation for generation. Status is read-only, and current domain stage accounting accepts observed native binding without a project installation ID. The unrelated machine-wide ownership defect in ARCH-SOL-1 remains.
- **FCIS/domain I/O and portable-history regression:** no new domain I/O, clock, process or adapter dependency was found in changed domain policy. Inspected unchanged registration proof, worker successor, generation reconciliation and history-save callers: operational state remains adapter-owned; retained executable digests are executable integrity evidence, not ledger authenticity gates. No introduced portable-ledger blocker was substantiated in this focus.

## Verification and limits

Executed from repository root:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 /tmp/al79-architecture-sol-c1/probe.py > /tmp/al79-architecture-sol-c1/probe.log 2>&1
git diff --cached --binary 4df45967c801824500ac2ea39a8f52fa9cbdcf66 -- . ':(exclude)alspaces' | sha256sum
git diff --name-only
git rev-parse HEAD
```

The probe reuses `scripts/tests/distribution_fixture.py` to compile actual candidate application builds and the existing native interface fixture, then invokes the shipped downloader and installed common wrapper. Its process exits 0 because it records the intentional failing child results; those child exit codes and diagnostics are the evidence above. All writable HOME, USERPROFILE, CODEX_HOME, XDG, cache, wrapper and adopting-project destinations are temporary and explicitly isolated. Temporary fixture trees were cleaned; script and raw output remain under `/tmp/al79-architecture-sol-c1/`. No shared product/test file was changed by this reviewer.

Reused raw `TEST_EVIDENCE.md`, `DISTRIBUTION_EVIDENCE.md` and `NATIVE_EVIDENCE.md`, distinguishing evolving-candidate test claims from the frozen C1 diff. Latest native evidence records a genuine normal-trust Linux/Codex 0.167.0 root and source-proven child in worktree-two, automatic publication generation/acknowledgement 7/7, all four observed stage flags and preservation of the first worktree's owners. That successful capture does not exercise the upgrade/legacy/manual-cwd counterexamples.

No full package suite, race suite or `make` was independently rerun; the parent owns revision-aware integration verification. Parent reports two fixture races in its concurrent affected-package run; no pass is inferred from that run. Native Windows/PowerShell execution remains unavailable. Architecture findings are verified on Linux amd64; Windows observations are source traces only. The parent owns deduplication, fixes, re-review and publication.

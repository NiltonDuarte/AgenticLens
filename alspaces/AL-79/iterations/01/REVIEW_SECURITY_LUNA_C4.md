# AL-79 C4 Security Specialist Re-review

## Identity and scope

- Run: `security-luna-c4` (correction re-review).
- Requested model/effort: `gpt-6-luna` / `medium`; this is the requested identity, not independently attested runtime metadata.
- Candidate: `AL79-C4`; base `4df45967c801824500ac2ea39a8f52fa9cbdcf66`.
- Frozen candidate source diff: `/tmp/al79-review-coordinator-b67f2d/CANDIDATE_C4.diff`, SHA-256 `9d1ca49a319643feeb13774a28942037cfff9d175884d53a8e8495660a90e265`, matching `CANDIDATE_C4.json`.
- Focus: user-home ownership/trust, callback/source context, executable admission and shell transport; C4 migration preservation and recovery resource leases.
- Read the assigned generic reviewer instructions, shared review context, canonical review procedure, approved contract/plan, this review's own prior security report, parent review context, and supplied evidence. No sibling review reports were read. Product and test files remained read-only; this report is the only owned path.

## Findings

No substantiated introduced security finding in the assigned scope. No remaining `SEC-01` blocker was substantiated in the corrected candidate.

## C4 correction review

- **Unknown/modified legacy callbacks:** `legacyUserChanges` now removes only groups exactly attested by the existing project receipt, then refuses migration if any AgenticLens/legacy receive callback remains. The installer repeats conflict detection after recovery and before activation, and the migration changes are committed transactionally. The distribution evidence records a regression fixture that copies an actual installed callback, changes only its status text, and verifies base and linked-worktree migration refuses without changing hooks, receipts, pending facts, or global activation. The clean unchanged-owned path then succeeds after the fixture models explicit user removal of the unknown copy. This preserves user ownership and does not turn recognition into deletion authority.
- **Recovery resource leases:** `installUserCodexWithCheckpoint` holds the machine root lock and all discovered current/prior/legacy home locks before recovery. It rediscovers pending resources under those leases, checks each against the held set, recovers before recomputing the migration plan, and checks the resulting journal paths against held leases before mutation. The targeted Go evidence records a concurrent installer introducing a distinct home/journal after discovery; the first transaction retries without changing journal/artifact snapshots, then succeeds after fresh discovery and recovery under the new leases. This is consistent with the root lock serializing journal creation and prevents recovery from writing an unleased home.
- **Executable admission/Windows binding:** reviewed the current dispatch and activation callers alongside unchanged C2 security evidence. Automatic callbacks resolve shared config from absolute callback cwd, require configured source to agree with established `CODEX_HOME`, validate transcript containment and source session/project context, and execute the selected admitted native. User binding feedback uses `& 'path with spaces''and apostrophe'` PowerShell call syntax and PowerShell single-quoted literals for fixed arguments. The C4 Windows fixture obtains that emitted command from actual dispatch feedback and executes it after substituting quoted explicit ticket values and the rendered repo argument; it checks exact ticket, title and source session. This supports command syntax/transport at the fixture boundary and leaves trust entirely with normal Codex user interaction. No trust-bypass or trust-setting mutation was found.

## Hypotheses assessed

1. **Migration could erase or silently leave duplicate user callbacks.** Not supported for the inspected migration path: only receipt-matching groups are removed; remaining recognized executable callbacks cause refusal; unrelated callbacks and pending state are preserved. C4 distribution evidence exercises modified-owned conflict and the additional unchanged-command/unknown-group case.
2. **A concurrent journal could introduce a recovery destination without a held lease.** Not supported by the reviewed lock and re-discovery sequence; the targeted resource-race oracle demonstrates retry before mutation and recovery after leasing the new resource.
3. **PowerShell loses path/argument boundaries or binding changes trust.** Not supported by the emitted command and Windows fixture using paths and title text containing spaces/apostrophes. Command quoting is exercised in a native Windows fixture according to evidence, but local native Windows execution was unavailable. The command requests the user's explicit ticket and does not bypass normal trust.
4. **Callback-controlled source or executable paths redirect capture.** No new C4 change affects this boundary. Reused the prior security review's source/callback trace for unchanged paths, with current callers inspected in `dispatch.go`, `NativeBindingPrompt`, `receiveCodex`, and dispatcher activation. The candidate still binds callbacks to validated project/source context and the selected executable; no new bypass path was found.

## Verification and limits

- Verified the candidate identity against the supplied frozen metadata and reviewed the C3-to-C4 and full candidate diffs for relevant changes.
- Reused `/tmp/al79-tester-c4/final-targeted.log` and the source inventory fingerprint `a444dacf76685fdab28ae034afcea2312509a1addc1b0a6442fdd41241005dc7` as targeted correction verification; reused `DISTRIBUTION_EVIDENCE.md` for C4 migration and emitted-command fixture evidence. No tests or new probes were run by this reviewer.
- Reused prior normal-trust Codex Linux root/child evidence only for the unchanged source/callback and trust behavior it actually exercised. That evidence is revision-bounded and is not a new C4 native session.
- Native Windows/PowerShell runtime execution, real Windows loaded-image sharing, and Windows native hook trust were not available locally. Cross-compilation/source inspection and the recorded Windows fixture are not claimed as local native-runtime proof. Parent integrated `make`/race verification is pending and outside this focused verdict.

## Scoped outcome

**APPROVED** — no blocking security defect was substantiated within this specialist re-review of AL79-C4. The verdict is limited to the frozen candidate and investigated boundaries; the native Windows/PowerShell runtime limitation remains explicit, and this is not consolidated ticket approval.

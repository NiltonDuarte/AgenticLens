# AL-79 C2 Security Specialist Review

## Identity and scope

- Run: `security-luna-01` (independent specialist first pass).
- Requested model/effort: `gpt-6-luna` / `medium`.
- Dispatch/runtime identity: this session is supplied as Codex harness running `gpt-6-luna` at medium effort. No separate provider dispatch receipt or token/cost record was available; none is inferred.
- Candidate: `AL79-C2`; base `4df45967c801824500ac2ea39a8f52fa9cbdcf66`.
- Reviewed frozen source diff: `/tmp/al79-review-coordinator-b67f2d/CANDIDATE_C2.diff`, SHA-256 `84aa66679b8b2b643ccdfd3c17148b77a369ea383a546eee372397778218a7de`; candidate metadata identifies the same hash and base. Workspace metadata was excluded from candidate scope.
- Focus: native user-home Codex trust boundary, additive hook and artifact ownership, path/source/session containment, native routing, shell/PowerShell argument handling, executable activation/integrity, source-home agreement and persisted human trust.
- Read required reviewer instructions, shared review context, review skill, AL-79 contract/plan/specification, PO-20 F1-F6 and canonical technical/regression guidance. Inspected the relevant dispatcher, hook, Codex source-context, installation/transaction, executable activation, store containment and shell/PowerShell entrypoint code. Product and test files remained read-only; only this assigned report is written.

## Findings

No substantiated introduced security finding in the assigned scope. Severity: none.

## Investigation and evidence

- **Callback trust and project selection:** `backend/cmd/agentic-lens/dispatch.go` reads the native hook payload, requires an absolute callback cwd, discovers the shared config from that cwd, and routes through that project. Callback cwd does not become authority for source identity: the selected local Codex home must agree with established `CODEX_HOME`, and any supplied transcript must physically belong to its live/archive source store. Source metadata is checked for session identity and project context; subagent-start normalization accepts a child only when its authoritative parent ID matches the callback's parent ID and its source cwd resolves to the same project. I found no reachable path in this flow that uses an arbitrary callback transcript to redirect writes to another source/project.
- **Hook ownership and trust preservation:** `capture.InstallUserCodex` appends its marked groups to the effective Codex user hooks file and records exact owned groups. Removal/update requires exactly one unchanged owned group; conflicts or modified groups are preserved. The transaction journal restores only its exact pending bytes and reports edited targets. The installer does not write Codex trust/bypass settings. This matches the contract's additive ownership and persisted-trust boundary.
- **Executable activation and argument handling:** the native user hook command is constructed from a fixed dispatcher path and fixed arguments. POSIX invocation uses `exec` with the original argument vector; PowerShell uses `ProcessStartInfo` and applies Windows CRT quoting for embedded quotes and trailing backslashes. Cache/development selection is native and validates selected executable identity; Windows activation requires the active artifact's content hash and an owned root-relative immutable path. I found no demonstrated argument injection or activation path escape in the inspected paths.
- **Observed native trust boundary:** `NATIVE_EVIDENCE.md` records a genuine Codex 0.167.0/Linux amd64 root plus source-proven child run using normally persisted human trust, without trust bypass. It records callback outcomes, explicit root binding, source-parent child inheritance and automatic source-backed publication. The recorded evidence also documents the earlier real SubagentStart mismatch and the corrected affected-candidate recheck. This supports real native discovery and routing for that stated environment; it does not certify other Codex versions or platforms.

## Hypotheses assessed

1. A malicious/stale callback cwd or transcript could redirect capture into another project's state/source. **Not supported:** shared config derives the project from absolute session cwd; transcript physical containment, session metadata, and source/project agreement constrain the handoff. Delayed/unreadable metadata remains pending rather than becoming authoritative empty data.
2. Installing/removing the global hooks could erase unrelated user hooks, silently change persisted human trust, or remove a user-edited AgenticLens command. **Not supported:** additive JSON update, exact group ownership and conflict-preserving removal are present; no trust setting mutation was found. Native evidence used normal trust.
3. A hook-controlled path or shell/PowerShell argument could execute a different binary or escape the selected native routing. **Not supported by inspected call paths:** hook commands are generated with fixed argv; source paths are validated; POSIX preserves argv and PowerShell applies native argument quoting. Windows runtime behavior remains unobserved here.
4. Mutable executable replacement or activation metadata could redirect an active Windows worker to arbitrary bytes. **Not substantiated:** activation constrains the path to the hash-named `native` child and verifies both active and stable bytes. The actual Windows native activation path was not exercised.

## Verification and limits

- Reused revision-specific raw evidence from `NATIVE_EVIDENCE.md`; no new tests or runtime probes were run. Candidate diff hash and metadata hash matched exactly.
- Security source review covered the listed paths and consequential unchanged helpers (`ValidateRolloutStore`, physical path resolution, `CanonicalStore`/`SameStore`, config discovery and the POSIX/PowerShell launchers). This is a scoped diff review, not a whole-repository audit.
- Native normal-trust behavior is evidenced only on Codex 0.167.0/Linux amd64. Windows native trust, PowerShell runtime argument transport, native executable activation, and other host/version combinations remain unverified; Windows cross-compile/source evidence is not native execution evidence. Integrated `make`/race verification is pending the parent and is not claimed here.

## Scoped outcome

**APPROVED** — no blocking security defect was substantiated within this specialist scope for C2. Approval is limited to the reviewed frozen candidate, source paths, and evidence above; it is not whole-ticket approval or a claim of cross-platform native verification.

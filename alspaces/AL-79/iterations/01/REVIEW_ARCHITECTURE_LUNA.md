# AL-79 architecture review — architecture-luna-01

## Review basis

- Candidate: `AL79-C2`, frozen source/guidance/test diff based on `4df45967c801824500ac2ea39a8f52fa9cbdcf66`.
- Diff SHA-256: `84aa66679b8b2b643ccdfd3c17148b77a369ea383a546eee372397778218a7de` (`alspaces` metadata excluded).
- Assigned focus: architecture; native selector/config/runtime reuse, worker/build lifetime, ownership, Windows activation, source/shipped bootstrap consistency and compatibility consumers.
- Requested and runtime-reported reviewer identity: `gpt-6-luna`, medium effort. No source-observed cost/usage metadata available.
- Requirements read: AL-79 contract and iteration plan, AL-79 specification and PO-20 F1-F6 references; canonical technical and regression guidance. Product and test source remained read-only.

## Finding

### A1 — Route from exact native session context instead of callback cwd

- Severity: **8/10**.
- Classification: **introduced requirement defect**.
- Locations: `backend/cmd/agentic-lens/dispatch.go`, `nativeCodexHook` (project selection around `FindConfigFile(value.WorkingDirectory)` and source validation); `backend/internal/adapters/capture/codex_context.go`, `NormalizeCodexSubagentStart` (requires child source project to equal the callback-selected project).
- Expectation: AL-79 CONTRACT.md says validated session context identifies the active checkout, including nested invocations, and a different callback process cwd cannot redirect it. A positively established source/location contradiction should be reported and stopped, but callback cwd itself is not the authority.
- Trigger: A native callback's `WorkingDirectory` differs from the exact session rollout's source `cwd` while the rollout is valid, belongs to the configured Codex store, and identifies another configured checkout/worktree. For SubagentStart, the callback session is the parent while transcript metadata identifies the child; the same path first resolves the project from callback cwd and then compares child source cwd against that choice.
- Consequence: `nativeCodexHook` selects local policy and project using callback cwd before obtaining authoritative source context. For ordinary callbacks, the source-context check then rejects the mismatch. For SubagentStart, normalization likewise rejects when the child source context resolves to another project. Capture is therefore lost for a valid session whose process cwd differs, or a callback cwd can choose the wrong policy before the source is checked. This is a reachable conflict with the required routing semantics, not a proposed cleanup.
- Evidence: Source trace in frozen diff: `nativeCodexHook` calls `FindConfigFile(value.WorkingDirectory)`, derives `project`, resolves `ResolveCapture`/`Load`, and later calls `SessionContext(value.TranscriptPath, value.SessionID)` and errors if the source-derived project differs. The subagent path calls `NormalizeCodexSubagentStart(..., project)`, which compares callback-selected project with child metadata's `WorkingDirectory` and returns `ErrSessionContextMismatch`. AL-79 contract explicitly states the session context is authoritative. Existing native root/child evidence proves the normal matching-cwd case only; it does not cover this disagreement branch.
- Needed correction: Establish exact native session identity and source cwd first (preserving the parent/child callback identity rules), then select the configured checkout from that validated context. Keep an explicit error when the authoritative session context itself conflicts with a separately established source/store fact. Add/reuse a disposable regression covering differing callback cwd with a source-valid configured checkout and assert routing/policy comes from source cwd.

## Hypotheses investigated and disposition

- **Shared native resolver and shell bootstrap could diverge:** `ResolveEntrypoint` parses the narrow shared YAML bootstrap and independently calls native `SelectConfig` for local path selection. The shell wrappers and native parser both changed, but this pass found no concrete divergent accepted/rejected document within the inspected paths. No finding.
- **Active workers could switch builds or lose pending work after rebuild:** `newCaptureWorkerFence` retains executable bytes into state; active worker launches later children from retained bytes, while new invocations re-resolve preferred native. Raw native evidence also records distinct active/preferred identities across rebuild and a later normal native root/child publication. No concrete defect found in this path on Linux.
- **Windows stable activation:** Source implements stable dispatcher forwarding to hash-named retained executable bytes and has Windows cross-compilation evidence. Windows runtime activation/loaded-image behavior is not executable in this environment, so runtime behavior remains unverified; this is a platform verification limit, not a source-level finding.
- **Manual and portable ledger compatibility:** Manual generation still resolves shared generation policy and its explicit native path; portable ledger rules remain an acceptance boundary. Existing evidence includes cross-machine/history preservation assertions and native evidence preserves the other worktree ledger. No architecture-specific defect found in the inspected callers.
- **Domain purity:** The changes inspected put Git/filesystem/process selection in adapters/CLI and keep domain configuration models/value functions separate. No FCIS boundary breach found in the scoped diff.

## Verification and limits

- Reused raw candidate evidence: `TEST_EVIDENCE.md`, `DISTRIBUTION_EVIDENCE.md`, and `NATIVE_EVIDENCE.md`; the latter records normal trusted Codex 0.167.0/Linux amd64 root/child callbacks and automatic publication after the callback identity correction.
- Confirmed the frozen diff SHA and candidate base against the supplied identity. Read the exact routing callers and unchanged `codex.SessionContext`/`ReadSessionContext` dependencies.
- No tests were run in this reviewer pass. The matching-cwd native scenario does not exercise A1. Full integrated `make` and race verification remain pending per supplied evidence; Windows runtime is unavailable (cross-compile only).

## Scoped outcome

**CHANGES_REQUESTED** — A1 is a substantiated introduced contract defect in native routing. This verdict is limited to the assigned architecture focus and AL79-C2; it is not a whole-ticket or merge approval.

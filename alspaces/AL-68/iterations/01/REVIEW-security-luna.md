# AL-68 independent security review

**Outcome: INCOMPLETE**

## Scope and identity

- Assignment: `security-luna-01`, independent security specialist discovery.
- Requested profile: `codex / gpt-6-luna / medium`. Actual runtime model and effort metadata were not available to this reviewer; the requested profile is not asserted as observed runtime identity.
- Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Candidate fingerprint: `c7dade3313a1d65a4269d595c782750655be8a044c020dc1f3ac502464cf325a`, independently recomputed from the paths and `path/NUL/content/NUL` definition in `CANDIDATE.json`; matched.
- Security boundary inspected: lifecycle hook input/output, native process invocation, binding challenge, local registration and publication redaction, installer ownership/update/removal, and relevant unchanged Codex/OpenCode source extraction callers. Product and test files were read-only.

## Findings

No confirmed security defect was established in the inspected candidate. This is not approval because the Codex trust and callback path has not been exercised end to end.

## Hypotheses investigated

1. **Hook payload can inject shell arguments or executable code. Disposition: not reproduced in the inspected paths.** Codex hook command values are statically shell-quoted by `staticCommand`; apostrophes are escaped by `QuoteArgument`. OpenCode invokes the pinned binary using Node `execFile(binary, ["capture", ...args], ...)`, and serializes callback values as JSON. The native receiver limits payloads to 64 KiB, accepts only known events, validates exact identity fields downstream, and does not use unknown payload ticket/model/parent fields for binding or source facts.
2. **An agent can bind a different session or replay an unrelated ticket request. Disposition: guarded in source.** Codex binding uses the callback's session identity and a random per-registration request token. `Store.Bind` requires exact token equality and `CaptureBinding` rejects conflicting ticket ownership. OpenCode's bind tool obtains the session ID from runtime `context.sessionID`; its arguments are ticket and request only. No arbitrary parent ID or tool argument selects the target session.
3. **Transcript or secret data leaks into final artifacts. Disposition: not observed in the inspected serialization path.** Full hook payloads are not persisted; only selected identity/event/path metadata is retained locally. Transcript-derived session facts are parsed during reconciliation. `ReportJSON` and coverage rendering are the publication boundary; the accompanying raw smoke evidence says generated report and summary omit request tokens, local paths, and transcript/snapshot fields. Local registry content is intentionally sensitive local state and is ignored by the installer; the smoke evidence is not a full adversarial filesystem/privacy test.
4. **Installer overwrites unrelated hook state or removes user edits. Disposition: primary ownership checks exist; concurrent-edit behavior remains unverified.** Codex installation removes prior owned groups by exact JSON equality, refuses missing/edited owned groups and marker-bearing untracked groups, and appends its own entries. OpenCode checks the owned plugin hash before replacement/removal. Writes use temp files and rollback tracked files after an apply error. I did not establish a reachable exploit in the single-user setup flow, but did not exercise concurrent edits between preflight and rename or symlinked parent directories.
5. **Codex callbacks can provide the required actual agent interaction without weakening trust. Disposition: unresolved required evidence gap.** The raw smoke record has only an isolated config awaiting human project trust and `/hooks` review; it explicitly claims no callback, binding, child capture, or source comparison. The accepted plan requires actual agent feedback/binding round trip and forbids bypassing or auto-approving trust. That path cannot be security-approved from fixtures or protocol descriptions alone. The OpenCode 1.19.5 smoke does record an actual agent binding interaction and source comparison, but does not close the Codex gate or establish T3/native-child behavior.

## Evidence reviewed

- Requirements: `alspaces/AL-68/CONTRACT.md`, `alspaces/AL-68/iterations/01/PLAN.md`, and the AL-68 ticket specification.
- Raw evidence: `TEST_REPORT.md`, `SMOKE.md`, `SETUP_IMPLEMENTATION.md`, `VERIFICATION.md`. I did not read `REVIEW.md`, `FINDINGS.json`, or sibling reviewer reports/conclusions.
- Source: `backend/internal/adapters/capture/{hooks.go,install.go,store.go,reconcile.go,plugin.ts}`, `backend/cmd/agentic-lens/{capture.go,main.go}`, related capture tests, and relevant unchanged Codex/OpenCode extraction code.
- Reused verification: reported full `make` passed on the frozen content fingerprint; it is build/test evidence, not a substitute for runtime security evidence. No suites were rerun because no concrete code-level failure hypothesis required it.
- Runtime evidence: OpenCode 1.19.5 Linux smoke documents actual per-session agent binding and source-record comparison. Codex 0.166.0 is reported as hooks-enabled, but actual trust approval/callback delivery/agent response/source identity comparison is pending. No trust settings were changed or bypassed.

## Gaps and limits

The required Codex actual-hook interaction and trust review remain unverified, as do the requested T3/native-child path and Windows/macOS execution. Therefore this security-focused review is **INCOMPLETE**, not a finding that the Codex path is defective. Obtain the planned human-controlled trust review and real callback/agent-response/source comparison, then request a focused security re-check. This report assesses only the specified frozen candidate and inspected paths; it does not certify the whole repository.

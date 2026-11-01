# AL-68 bounded security re-review

**Scoped outcome: APPROVED**
**Overall delivery status: INCOMPLETE** — the Codex human trust review and actual callback/agent-binding/source comparison remain pending, as documented in `SMOKE.md`. This review does not bypass that gate.

## Scope and identity

- Run: `security-luna-04`; requested profile `codex / gpt-6-luna / medium`. This records the requested profile, not independently observed runtime metadata.
- Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Frozen candidate: `c34cc949278608f4c777f7428a620e2ae29542e7fe33875d5b13ae9f0d17e6ec`, `CANDIDATE-04.json` (`corrected-04`). The candidate is a source fingerprint, not a Git object in this checkout. I independently recomputed the declared sorted path/NUL/content/NUL fingerprint from its manifest paths; computed and declared values match exactly. This is not a hash of the manifest bytes.
- Read `reviewer.md`, shared review context and canonical procedure, regression boundaries, contract/plan/corrections/findings, my initial `REVIEW-security-luna.md`, and implementation, test, smoke, setup and verification reports. Did not read partner Sol04 closing conclusions. Product/tests remained read-only; only this assigned report path was written.
- Bounded paths: capture hook handling, plugin generation and invocation, registration/store identity, reconciliation/publication integration and locks, relevant binding/privacy tests, plus the reports listed above.

## Findings

No unresolved introduced security defect was substantiated in the inspected boundaries. No unrelated or speculative finding is raised.

## Fixed finding verification

- **CAP-06 — fixed.** `capture.Install` now uses `strings.NewReplacer(...).Replace(pluginTemplate)` once against the original template (`capture/install.go`). JSON-encoded values containing another placeholder marker remain literal; they cannot be re-expanded by later replacements. The native installer fixture `TestCaptureInstallerRendersMarkerBearingPathsOnceAndRollsBackMidWrite/literal_marker_paths` exercises cyclic marker-bearing paths and verifies the emitted plugin values decode to the exact original paths. The production path and the regression oracle align.
- **CAP-07 — fixed.** The CLI timeout context now governs native receive input acquisition. `ReadHookContext` runs the bounded-volume read and, on context completion, closes an input implementing `io.Closer`, releasing a blocked pipe read. The native subprocess regression writes valid JSON while leaving stdin open, requires exit before its independent three-second deadline, and confirms no registration was committed. This demonstrates an elapsed-time bound for the actual native pipe path; it does not claim arbitrary non-closeable `io.Reader` implementations can be interrupted.

## Security hypotheses and dispositions

- **Session-specific challenge, identity, and argument safety: no regression found.** Codex registration keys use the callback-provided session identity; binding requests remain random, exact-session challenges checked by the store, and conflicting ownership remains rejected. Child-start callbacks still treat `session_id` as the parent and do not promote `agent_id` to a source identity. OpenCode obtains its session ID from runtime `context.sessionID`; ticket/request are tool inputs, not identity selectors. Its plugin invokes the configured executable with `execFile` and an argument array, and path values are JSON-rendered into constant source positions. These paths preserve the trust boundary against shell/argument injection and guessed session ownership.
- **Actual-store alias handling: no bypass found.** New identities canonicalize existing symlinks; the transactional `Change` path also scans legacy rows sharing source/session and compares canonical paths or `os.SameFile` before changing a registration. Multiple matching actual-store registrations fail closed. The native regression exercises aliases against an existing ticket owner and expects conflict/no publication. Unavailable stores are not treated as equivalent evidence.
- **Publication lock and ordering: no concrete trust regression found.** The per-output lock is acquired around baseline read through ledger/report/summary publication. Its stable lock inode is intentionally retained, ownership is kernel-released on process exit, and digest/report consistency is covered by overlapping publisher tests. The Unix and Windows implementations use exclusive file locks. This review did not perform a new cross-platform runtime test; Windows/macOS limits remain as stated in verification evidence.
- **Final privacy: no leak found in inspected serialization/publication path.** The plugin only passes lifecycle identity/event JSON and the local registry retains bounded normalized registration/source-derived compact state. Binding request tokens, state/store paths and transcript/snapshot payload fields are excluded from the final capture report and summary by existing focused assertions. Source reconciliation, not hook payload claims, supplies published facts. These checks support the stated report boundary, not a claim that ignored local registry data is non-sensitive.

## Verification reused and limits

- Reused the corrected-candidate results in `TEST_REPORT.md`: the focused capture/domain/CLI race suites and CAP-06/CAP-07 native fixtures are recorded passing. The report records a full `make` pass for candidate 03, before candidate 04's bounded child-scheduling correction; candidate 04 integrated `make` remains parent-owned. I did not repeat broad suites or mutate shared source/tests.
- Reused actual OpenCode smoke evidence only for its recorded per-session binding and source comparison. It does not close Codex trust/callback acceptance.
- **Overall status remains INCOMPLETE** until a human reviews/accepts the exact Codex hook trust state and an actual callback/agent response is compared against source identity/facts. No trust configuration was changed or bypassed. Native Codex child runtime and untested platform execution also remain outside this bounded security approval.

This approval applies only to the listed security hypotheses and frozen candidate fingerprint. It is not an overall ticket approval, release approval, or merge authorization.

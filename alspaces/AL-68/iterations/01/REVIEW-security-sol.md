# AL-68 independent security review

Outcome: CHANGES_REQUESTED. Scope: hook stdin/process execution, exact agent binding, trust preservation, local capture privacy and installation ownership. This is a scoped discovery verdict, not whole-change approval.

Run: `security-sol-01`, discovery. Requested profile: `codex/gpt-6.1-sol/high`. Observed runtime: Codex tool environment; exact executing model and effort were not independently exposed or verified. No usage or cost is inferred from the requested profile.

Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`. Reviewed frozen working-tree fingerprint: `c7dade3313a1d65a4269d595c782750655be8a044c020dc1f3ac502464cf325a`, using CANDIDATE.json's sorted path/NUL/content/NUL definition. Independently recalculated and matched before probes. Product/tests remained read-only; only this assigned report was written. No sibling review report or conclusions were read. The assigned raw VERIFICATION.md contains a general sentence referring to other reviews; no underlying findings were inspected or used.

## Findings

### SEC-SOL-01: OpenCode template replacement reinterprets already escaped paths (P2)

Location: `backend/internal/adapters/capture/install.go:284-290`.

Trigger: an accepted native handler, state or source-store path contains another template marker, such as a handler named `native __STATE__`, together with a state path containing a legal Unix quote character. The installer JSON-encodes each value but repeatedly applies replacements to the whole evolving output. A subsequent iteration therefore replaces marker text inside an already inserted JSON string. Go map iteration makes the result order-dependent.

Evidence: built the exact reviewed source into `/tmp/al68-security-sol-h3egid4i/agentic-lens` with `go build -o <disposable-output> ./cmd/agentic-lens`. Copied it to `native __STATE__`. Twelve real `capture install --source opencode` calls used independent disposable repositories, state directories named `state"quote` and an explicit source store. All installs exited zero. Ten generated plugins failed `node --check` when checking their three plain JavaScript constant declarations; two parsed successfully. A failing output was:

```javascript
const binary = "/tmp/al68-security-sol-h3egid4i/native "/tmp/al68-security-sol-h3egid4i/case0/state\"quote"";
```

Consequence: successful installation records an active owned plugin that OpenCode cannot parse, disabling capture and its agent interaction for otherwise accepted paths. This demonstrates loss of data/code escaping; arbitrary code execution was not demonstrated and is not claimed.

Required correction: expand the original template without processing text introduced by replacement values. Add a behavioral regression covering marker-bearing paths and quote characters, checking generated syntax and exact recovered string values.

### SEC-SOL-02: receive's timeout does not bound stdin acquisition (P2)

Location: `backend/internal/adapters/capture/hooks.go:31-32`; caller `backend/cmd/agentic-lens/capture.go` creates a timeout context at line 65 but calls `capture.ReadHook(input)` without it.

Trigger: stdin supplies less than the 64 KiB byte limit and remains open. `io.ReadAll` waits for EOF without observing the command context, so the advertised positive `--timeout` cannot terminate handoff.

Evidence: the same disposable exact-source native binary ran `capture receive --source codex --store <temporary-source> --state-dir <temporary-state> --timeout 100ms`. A Python pipe wrote and flushed the complete valid payload `{"session_id":"stalled","hook_event_name":"SessionStart"}` and intentionally kept stdin open. The process remained running after 0.801 seconds and required the disposable probe to kill it. No source extraction or model call was involved.

Consequence: a stalled producer can retain a receive process and registry resources indefinitely despite the requested handler timeout. The byte limit bounds volume, not elapsed time. Installed Codex's outer two-second hook limit and OpenCode's execFile 1500 ms limit reduce exposure in those particular callers, but do not implement the native command's promised bounded handoff.

Required correction: make timeout/cancellation cover input acquisition and return promptly without committing a partial event. Add a pipe/reader regression that stays open beyond the configured timeout, rather than only testing oversized in-memory strings.

## Investigated hypotheses and evidence

- Unknown payload fields granting ticket/model/parent ownership: not substantiated. ReadHook uses concrete normalized fields; binding validates a random per-source/store/session challenge and CaptureBinding rejects conflicts. Parent inheritance requires a parent from accepted source facts. Codex SubagentStart does not confuse callback parent session_id with child agent_id. The request token is correlation evidence, not an authentication boundary against another process already controlling the same local account; explicit outside-session binding is allowed by the contract.
- Shell execution of hook payloads: not substantiated on inspected Unix paths. Codex commands quote fixed installation arguments; stdin values are parsed as JSON. OpenCode launches the native executable through execFile with argument arrays and obtains the binding tool session identity from runtime context. The template-expansion defect above remains open.
- Recursive model/update calls: not substantiated. Capture wrapper dispatch precedes update/network paths; hooks invoke native handlers and do not initiate a model. Stop handlers emit no binding context or repeated stop decision. OpenCode catches capture failures and retains coding availability.
- Raw hook/transcript persistence or final request-token leakage: ordinary traced paths discard unrestricted hook fields. Codex's parser minimizes content, accepted session snapshots pass PersistAnomalySummaries, and ReportJSON clears challenges and source-store paths without mutating local registrations. Existing skill/task facts are retained under the explicit AL-62 preservation guarantee. No broader secret-redaction guarantee for arbitrary source skill error strings is inferred.
- Trust/configuration ownership: no automatic harness trust change was found. Installer preflights the native protocol, checks owned Codex group/plugin content before upgrade/removal and preserves unrelated hook groups. No independent concurrent-edit or post-first-write fault injection was performed; existing ownership tests do not establish those additional cases.

## Verification and limits

Read requirements in CONTRACT.md, iteration PLAN.md including the user's actual-agent/no-guess clarification, and the ticket specification. Inspected capture command/domain/adapter/glue files, relevant main generation integration, Codex parsing, OpenCode exact snapshot and skill/task extraction, both shipped script changes and setup/operator guidance. Reused TEST_REPORT.md, SMOKE.md, SETUP_IMPLEMENTATION.md and VERIFICATION.md as raw evidence; did not repeat the complete suite. Parent's recorded make pass belongs to the unchanged frozen fingerprint and does not refute these untested paths.

Executed: candidate fingerprint check (match), disposable exact-source Go build (pass), twelve real installer/plugin syntax probes (ten failures/two passes), and stalled-stdin timeout probe (failed to terminate). Disposable probe artifacts remain outside Git under `/tmp/al68-security-sol-h3egid4i`; no real source transcripts, secrets, trust changes, external publication or telemetry were used.

Real OpenCode 1.19.5 Linux agent binding, native child and independent roots are supported by the assigned smoke record. I did not independently rerun those model sessions. Codex human trust and actual callback/agent-response smoke remain pending; no trust was bypassed or requested. Windows/macOS execution, additional harness versions, OpenCode v2 native runtime, T3 Codex children, resumed-session interaction and real interrupted-hook recovery remain unverified as recorded in SMOKE.md. In particular, Windows binding-prompt shell syntax was not established by this Linux review. These evidence gaps prevent extending support claims beyond observed paths, even after the two defects are corrected.

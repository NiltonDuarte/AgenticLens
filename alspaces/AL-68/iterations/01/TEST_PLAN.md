# AL-68 test preparation

Status: preparation only. Baseline inspected: 7bc670a7f5c9e93a220f3596f3d0e05cb6007332. The coder is implementing; no candidate revision is frozen and no application or test files were changed by this preparation. No test suites or real capture smoke checks were run. Execution/write ownership must be assigned after the candidate is ready.

## Existing guards to preserve and extend

| Contract | Inspected coverage owner and current evidence | Missing AL-68 oracle |
| --- | --- | --- |
| Exact Codex root selection and native descendants | `backend/internal/adapters/codex/discovery_test.go`: nonce selection ignores a newer unrelated same-directory root; ambiguous/conflicting selectors fail; nested native child provenance precedes replayed metadata | Registered independent exact IDs must join only their explicitly bound ticket; Codex subagent callback parent `session_id` must not be registered as the child |
| Source hierarchy and repeated refresh | `backend/cmd/agentic-lens/main_codex_test.go`: native child identity, role, token ownership and inherited replay exclusion; refresh replaces model snapshots without duplicated tool/event facts | Hook registration plus agent binding must reach the same source extraction behavior; explicitly bound independent roots stay parentless unless source evidence proves a parent |
| Cumulative source growth | `backend/cmd/agentic-lens/main_refresh_test.go`: both sources compare refreshed versus fresh sessions over active/completed/growing snapshots and unchanged repetition | Registry checkpoints, duplicate/out-of-order callbacks and resumed sessions cannot regress a newer accepted source snapshot or add cumulative totals twice |
| OpenCode source identity | `schema_v2_session_resolution_test.go` covers an explicit root; `schema_v1_session_resolution_test.go` covers explicit selectors but also intentionally retains title/workspace fallback behavior | The new registry extraction path must never reach title/workspace fallback. Use a missing exact ID alongside a valid title/path match and assert it remains unresolved |
| Failure and other-source preservation | `backfill/engine_test.go`: `CodexDiscoveryFailurePreservesSessions`, `CodexParseFailurePreservesTree`, `SourceCodexSkipsOpenCodeSessions`, `SourceOpenCodeSkipsCodexSessions`, `MultiRootSessionsExtracted`, `StrictSessionTrustRejectsUntrustedTitleFallback`, `SkillTaskRecoverySnapshots` | Drive the new capture/generate path with one refreshed owner, one unavailable owner and another source; compare preserved facts including skills/task links, native parent IDs, cost evidence and legitimate repeated invocation rows |
| Unknown cost | `codex/rollout_test.go` asserts unavailable cost is not observed zero; OpenCode schema-v2 cost tests coalesce reported/missing records | Hook/model claims must not override authoritative facts, and missing model/effort/cost must retain their established unknown representation |

No existing test deletion is proposed. Existing source fixtures/helpers can be extended where the production responsibility remains the same. Dedicated capture domain/store/installer tests are justified by new responsibilities absent at baseline; avoid duplicating extraction algorithm tests there.

## Required interaction and binding oracles

The user's approval explicitly requires the hook to interact with the agent for correct information without guessing. Test the production callback response and binding command together, not only matching a help string or hardcoded prompt:

1. Receive a real-format supported startup event lacking ticket context. Inspect model-visible feedback produced through the documented harness mechanism; it must contain an exact session-specific request and actionable response mechanism. Registration remains unassigned before a validated reply. A same-directory branch/title/role/time match must not assign it.
2. Submit the requested explicit binding reply through the production agent-facing command/tool. Assert the target session and ticket become bound, and only that source/store/session identity changes. An independent session in the same checkout stays unbound; another ticket's root remains unchanged.
3. Exercise absent reply, malformed reply, wrong-session reply, wrong source/store, conflicting ticket and repeated reply. Require visible unresolved/conflict states without ownership theft. The precise rejection oracle must follow the implemented challenge/session-validation protocol, rather than presume an invented authentication channel.
4. Native child inheritance succeeds only with proven parent identity and unambiguous parent binding. Put a subagent callback's parent `session_id`, separate `agent_id`, and transcript identity in conflict; require unresolved status rather than assigning the parent's usage to the child.
5. On a callback/runtime without a supported model-visible response channel, status must describe the limitation and explicit binding route. Such a fixture does not satisfy the real interaction acceptance boundary.

Pure identity/binding/precedence decisions belong in domain tests with value inputs and no mocks/I/O. Callback parsing, persisted registrations and command feedback belong in adapter/CLI tests with temporary files and the actual command path. The same harness payload serialized and decoded locally without executing production response handling is insufficient proof.

## Durable handoff, reconciliation and publication

- Store fixtures: simultaneous distinct registrations, duplicate callbacks, reordered checkpoints, truncated final event, interrupted replacement, restart recovery and concurrent bind/reconcile. Assert the last accepted state stays readable, no registration disappears and recoverable work stays pending. Use explicit synchronization/controlled fixtures instead of arbitrary sleeps.
- Limits: oversized input, missing source file/database, late source flush and exhausted source/session budget must produce distinct visible outcomes; they must not become measured zero, empty success or a complete report. Shutdown receive must not perform full extraction.
- Payload safety: include shell metacharacters in ticket/session/path values and unknown sensitive fields in input. Assert no injected side effect occurs and raw payload/transcript/secret fields do not reach the local registration or published report. Test supported path spaces with a custom local-state/source/workspace destination.
- CLI integration: generate an explicit-root ledger, register and bind a second independent root, then reconcile/generate twice. Assert exact per-session totals and identities, preserved other-source/unavailable owners, canonical v5 schema, digest-matching report and snapshot cutoff. Create partial report/ledger publication failure and assert inconsistent pairs are detected with deterministic recovery.
- Publication is deliberate: callback/idle/Stop must not modify an existing final ledger/summary. Raw source/transcript and intermediate local state must remain excluded from Git; fixture repositories make this assertion independent of this ticket workspace.

## Installer oracles

Exercise actual install/status/disable/uninstall commands against temporary adopting repositories with real config files. Snapshot unrelated hook entries/files before each operation. Fresh install, repeat, compatible upgrade, incompatible/missing binary, edited owned content, existing user hooks, failed apply and rollback/removal must preserve unrelated content. Dry run must leave byte-identical files and expose the proposed changes. Untrusted/disabled/unsupported settings remain visibly degraded and no trust flag is silently changed.

Both downloaders and packaged wrappers need execution coverage using a local prebuilt artifact in paths containing spaces; compare installed ownership/handler configuration, not merely script text parity. Test native Linux execution here; platform fixtures/cross-compilation alone cannot establish macOS/Windows callback execution. Observe the release-only client path without Go/Node installation; OpenCode glue may use only the existing harness runtime.

## Real harness smoke evidence and limits

For every claimed version/interface/source-family path, use a disposable repository and isolated local hook configuration. Do not alter active/global trust or user settings. Record version, interface, hook enablement/trust state and the actual documented event payload fields in normalized evidence. Launch a native root/child and independent reviewer roots without editing their skills; demonstrate the model-visible binding request and actual agent reply, then compare registered IDs and available usage/model/effort/cost with authoritative source records. Include a concurrent unbound/other-ticket root and resume/late-flush checkpoints. Keep raw source records outside Git.

Installed versions observed by the planner were Codex 0.166.0 and OpenCode 1.19.5. Neither version observation nor fixtures establish real interaction. T3/native-child capture and other platforms/version families are unverified until exercised. Lack of provider access, trust or a runtime must be reported, never turned into a silent skip or compatibility claim.

The existing `scripts/opencode-compat.sh` is unsuitable as AL-68 evidence without replacement: it reads `alspaces/AL-58/evidence` and drives a mock model provider. Those historical paths violate self-contained AL-68 fixture requirements, and a simulated provider does not demonstrate the requested real agent interaction. Preserve historical artifacts; do not change them for this ticket.

## Execution handoff

After coder handoff, inspect the frozen diff and its own checks before adding cases. Reuse coherent existing coverage; assign exact test paths between coder and tester before writes. Prepare frontend dependencies and embedded assets through canonical Make targets before direct Go package checks. Run affected capture/domain/source/CLI packages with the race detector, then the coordinated full `make` gate after integration. Record revision/fingerprint, exact commands, executed stages, results, environment limitations and remaining required smoke evidence in `TEST_REPORT.md`. A frontend stage that did not execute is not a passing stage.

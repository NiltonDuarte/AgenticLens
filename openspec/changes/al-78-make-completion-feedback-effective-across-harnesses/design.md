# AL-78: Technical design

## Scope and boundaries

Reduce the avoidable delay between an agent's code commit and the first queryable snapshot that includes the completed native work. [Acceptance scenarios](specs/completion-metrics-freshness/spec.md) cover automatic capture and explicit generation on native Codex and OpenCode. The observation starts with a real Git commit and ends with a consumer read backed by the published ledger. Git history and source history are separate: if the settled facts become readable only after the code commit, a later telemetry commit can version them.

Use the existing capture lifecycle and manual command. Forced turn rollover, automatic Git commits or amendments, proactive in-turn alerts for sessions that never run `generate`, installer changes, shared release-pin changes, capture-state migration, and T3 orchestration are outside the proposed scope.

## Decisions and ownership

- The native source defines completion. Match Codex rollout terminal records to a native turn ID. For OpenCode, use the selected schema family's persisted assistant activity IDs and completion evidence, and observe the client return or idle separately. Native hooks, a tool return, and the start of a new model request can trigger observation. Establish accounting completion from the source adapter's exact interval evidence.
- Keep the 180-second checkpoint and existing owner attribution. Generic `generate` currently has no verified caller/interval transport. Reuse its qualified feedback unless a supported in-turn transport can independently establish the exact owner and current interval. Keep feedback and probe timestamps out of the ledger. Preserve `active_duration_seconds`, optional positive `provisional_seconds`, and exact-owner replacement with portable history.
- Prefer the existing bounded automatic refresh after completion. For manual `generate`, publish available facts and explain the external refresh after native completion and source availability. A supported capture path or an external operator command can publish while the agent is idle. Use proof of delivery to model input before relying on the agent to act on a message.
- The orchestrator owns the change artifacts, integrated verification, and delivery. After approval, a coder owns any required CLI and capture/source adapter edits plus canonical guidance updates. A tester owns the isolated probes and regression tests. Reviewers read candidate code and evidence independently. Assign one named owner per writable path before dispatch; the diagnostic run determines which product paths need edits.

## Implementation approval

The user explicitly approved implementation in the AL-78 thread with `$agentic-lens-dev aprovado`. Approval covers the proposal, acceptance scenarios, and scope boundaries recorded here at baseline `db4477fa`.

## Implementation approach

The exploration baseline `db4477fa` includes AL-64 (`57a3503e`), AL-68 (`65e0eebe`), and AL-96 (`83f6f6c1`). The [AL-64 evidence](../../../alspaces/AL-64/SUMMARY.md#deferred-open-turn-follow-up) establishes manual source-open feedback and later recovery. The [AL-96 evidence](../al-96-verify-automatic-ticket-metrics-capture-end-to-end-for-v3/evidence/native-gate.md) establishes automatic settlement on packaged Codex 0.170.0 and OpenCode 1.19.5 with a deterministic provider and feedback disabled. It supplies an oracle and an installed workflow, with no measurement anchored to an agent's real code commit and no proof that checkpoint wording changes agent behavior.

Extend or reuse the isolated native gate to observe the successful Git commit and its SHA externally, the native terminal boundary, the first independent read that sees the terminal and usage facts, the first eligible publication, and the first consumer read that shows it. Bound the observations and report their sampling resolution. Hold an exact interval past 180 seconds for the default feedback case. For manual feedback, verify what the next model request contains and what response follows it. A scripted provider can prove delivery and exercise native completion. Any claim that wording changes agent behavior needs an observed agent response under that wording. Compare baseline and candidate on the stage the change targets and record the result per client and source family.

If the feedback arrives and creates extra work, correct the guidance. If the harness drops the warning, assess a supported, timely context channel before changing the integration. If publication lags after the facts are queryable, use the existing worker wake, debounce, or flush retry path. When the source has yet to expose facts, keep the available snapshot and report the unresolved boundary. Record an explicit external generation fallback for unsupported installations. Update the [setup manual](../../../docs/SETUP_MANUAL.md#provisional-checkpoint-feedback), [operator guidance](../../../.agents/skills/agentic-lens-operator/references/cli.md), and [developer workflow](../../../CONTRIBUTING.md) for the proved behavior.

## Steering

The user clarified during exploration that the operational failure is the delay after a long task's code commit and that a remedy may use agent feedback or another supported path. The plan now uses commit-to-query latency as the acceptance boundary. A next-turn refresh is one possible workflow for manual users, and supported lifecycle capture can settle the same interval without it. Exact ownership, source-proven completion, the 180-second default, and lean portable persistence remain requirements. Verification now observes the code commit and source visibility separately from the refresh.

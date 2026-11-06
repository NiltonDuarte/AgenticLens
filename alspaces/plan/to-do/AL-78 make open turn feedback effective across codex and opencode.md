h3. Title: AL-78: Make open-turn feedback effective across Codex and OpenCode

*TARGET*: AgenticLens (CLI Feedback / Native Harness Lifecycle / Capture and Metrics Refresh)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: AL-64 native observation and subsequent scope reduction on 2026-11-06.
*Related*: AL-64 duration accounting and owner-scoped checkpoint feedback; AL-68 automatic lifecycle capture. Revalidate their delivered revisions before implementation.

*PROBLEM*

CLI feedback to the agent is not working as intended. Long provisional activity produces checkpoint text, but the observed agent does not use it to finish the open interval and arrange a subsequent refresh. Repeated metrics generation within that same open turn leaves its duration provisional and cannot include later usage records. Message emission alone does not establish that the agent consumed or acted on the feedback.

The intended workflow is to preserve ongoing work, reach a natural native completion boundary, and refresh metrics after completed duration and usage are available. Determine how to make this effective for both Codex AND OpenCode using their own supported interfaces. T3 is an interface, not a harness; the solution must not depend on T3 thread, queue or scheduling tools.

*OBSERVED EVIDENCE AND LIMITS*

* An independent Codex run waited four minutes and completed repository finalization, including metrics generation. Its saved snapshot recorded one still-open turn with active_duration_seconds=632 and provisional_seconds=632. Recorded usage was input=58810, cache_read=1620480, output=6395 and reasoning=991; cost remained unavailable. These are snapshot values, not usage at eventual completion.
* The run performed four successful source generations in a single turn, refreshing different sessions. It emitted qualified F03 feedback for the current source session and another open owner. The message stated that the invoking caller was unverified and requested refresh after the owning session completed naturally. There was no subsequent turn/refresh cycle; saved provisional duration remained.
* An earlier wait-only run did not generate metrics despite repository guidance. Explicit finalization instructions triggered generation but still did not establish the completion-and-refresh cycle.
* The observed native Codex session was 01a1b0d2-14f9-70b1-9522-6ba10ef25f64, with turn 01a1b0d2-1a26-7cb2-b18f-813a453f857e starting 2026-11-06T12:25:21.736Z. Its successful generation snapshot was recorded around 12:35:58Z, before the final response. Evidence was committed in AL-64 at 3f7652a6; inspect the source ledger and summary at that revision rather than treating later mutable files as the same snapshot.
* This demonstrates ineffective workflow feedback in the observed Codex transport, not absence of native continuation capabilities. Codex queue/resume and OpenCode plugin/SDK prompt surfaces were identified but not exercised end to end. OpenCode completion/refresh behavior remains independently unverified.
* No owned Codex hooks or OpenCode plugin were installed in the observed worktree. Configured capture state was incompatible, so the run used manual generation and isolated state. This experiment does not prove a defect or successful completion refresh in a healthy automatic capture installation. Temporary raw evidence is outside Git and may not survive; reproduce against genuine native records.

*SCOPE AND ACCEPTANCE*

* Reproduce on both native harnesses and distinguish message emission, delivery to agent context, agent response, actual native completion, and successful subsequent metrics publication. Verify exact owner and native interval; do not treat an aggregate subtotal, nonce, directory or extraction root as proof of current-caller identity.
* Determine whether the remedy belongs in feedback wording/delivery, native continuation, existing lifecycle capture, or a combination. Verify natural completion versus continuing within the same accounting interval. A supported post-completion capture refresh may avoid a metrics-only model turn; do not prescribe extra turns without evidence they are necessary.
* Demonstrate an actual native completion followed by a source-backed refresh for Codex and OpenCode separately. Verify completed duration replaces the provisional snapshot without accumulation or duplicate owners, and check whether token/turn records are current at that read. Preserve unavailable cost and explicitly report any source flush delay or incomplete refresh.
* Preserve the default 180-second threshold and owner/interval attribution unless a separately approved product decision changes them. Do not introduce aggregate-only warnings or endless refresh/continuation loops. Never write synthetic terminal events, claim that a model-request boundary necessarily closes an interval, interrupt unrelated work, or bypass harness trust.
* If an installed harness cannot support the required workflow, expose the specific unsupported boundary honestly and provide an actionable fallback. Do not claim completion solely because a diagnostic printed or a command exited successfully.
* Keep the ledger lean: existing duration total plus optional positive provisional_seconds, no reasoning field or persisted feedback/provenance expansion. Preserve unavailable owners, other harnesses and portable recorded facts.

*OUT OF SCOPE*

* Development-only capture installation, local build distribution, changing the shared release pin, migrating/resetting existing capture state, or a general installer redesign. The unimplemented AL-64 development-install proposal was removed at the user's direction; this ticket does not carry that implementation forward.
* T3-specific orchestration, a generic autonomous scheduler, forced rollover of arbitrary sessions, frontend reasoning behavior, or fabricated live-process proof.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session under AL-78 after approval of a grounded contract and plan. Creating this follow-up does not authorize implementation or transfer the current AL-64 session into AL-78.
* Retain source-grounded evidence for both native harnesses, meaningful regression tests, revision-specific focused review and required integrated make verification. Deterministic fixtures are not proof of a native completion cycle; record unverified runtime cases explicitly.
* Update setup/operator/developer guidance at its owning source when the actual workflow changes. Explain what is delivered, what the agent must do, what capture does automatically, and the remaining supported/unsupported boundaries.

*SOURCE POINTERS*

* [AL-64 summary](../../AL-64/SUMMARY.md), including task finalization at 3f7652a6.
* [AL-64 iteration 04 plan](../../AL-64/iterations/04/PLAN.md).
* [AL-64 native capability investigation](../../AL-64/iterations/04/RUNTIME_CAPABILITIES.md).

h3. Title: AL-78: Make open-turn feedback effective across Codex and OpenCode

*GOAL*

Make open-turn feedback lead to a source-proven completion and subsequent metrics refresh on native Codex and OpenCode.

*CONTEXT*

CLI feedback in the observed Codex run printed a checkpoint warning but did not produce native completion followed by a refresh. Repeated generation in one open turn left duration provisional. OpenCode completion/refresh remains unverified. The intended workflow preserves ongoing work, reaches natural native completion, then publishes completed duration and usage. T3 is an interface, not a harness.

*ACCEPTANCE CRITERIA*

* Reproduce on both native harnesses and distinguish message emission, delivery to agent context, agent response, actual native completion, and successful subsequent metrics publication. Verify exact owner and native interval; do not treat an aggregate subtotal, nonce, directory or extraction root as proof of current-caller identity.
* Determine whether the remedy belongs in feedback wording/delivery, native continuation, existing lifecycle capture, or a combination. Verify natural completion versus continuing within the same accounting interval. A supported post-completion capture refresh may avoid a metrics-only model turn; do not prescribe extra turns without evidence they are necessary.
* Demonstrate an actual native completion followed by a source-backed refresh for Codex and OpenCode separately. Verify completed duration replaces the provisional snapshot without accumulation or duplicate owners, and check whether token/turn records are current at that read. Preserve unavailable cost and explicitly report any source flush delay or incomplete refresh.
* Preserve the default 180-second threshold and owner/interval attribution unless a separately approved product decision changes them. Do not introduce aggregate-only warnings or endless refresh/continuation loops. Never write synthetic terminal events, claim that a model-request boundary necessarily closes an interval, interrupt unrelated work, or bypass harness trust.
* If an installed harness cannot support the required workflow, expose the specific unsupported boundary honestly and provide an actionable fallback. Do not claim completion solely because a diagnostic printed or a command exited successfully.
* Keep the ledger lean: existing duration total plus optional positive provisional_seconds, no reasoning field or persisted feedback/provenance expansion. Preserve unavailable owners, other harnesses and portable recorded facts.

*Status*: Ready
*Blocked by*: None
*Work type*: fix
*Priority*: P2

*IMPLEMENTATION NOTES*

*Related*: AL-64 duration accounting and owner-scoped checkpoint feedback; AL-68 automatic lifecycle capture. Revalidate their delivered revisions before implementation.

* Development-only capture installation, local build distribution, changing the shared release pin, migrating/resetting existing capture state, or a general installer redesign. The unimplemented AL-64 development-install proposal was removed at the user's direction; this ticket does not carry that implementation forward.
* T3-specific orchestration, a generic autonomous scheduler, forced rollover of arbitrary sessions, frontend reasoning behavior, or fabricated live-process proof.

* [AL-64 summary](../../AL-64/SUMMARY.md), including task finalization at 3f7652a6.
* [AL-64 iteration 04 plan](../../AL-64/iterations/04/PLAN.md).
* [AL-64 native capability investigation](../../AL-64/iterations/04/RUNTIME_CAPABILITIES.md).

Native Codex evidence at AL-64 commit `3f7652a6`: session `01a1b0d2-14f9-70b1-9522-6ba10ef25f64`, turn `01a1b0d2-1a26-7cb2-b18f-813a453f857e`, snapshot around 2026-11-06 12:35:58Z retained 632 provisional seconds despite four source generations. No owned hooks/plugin were installed and capture state was incompatible; this does not prove healthy automatic capture fails. Detailed source values and flush limits remain in the linked AL-64 evidence; reproduce from genuine native records.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [relevant regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. Historical findings require revalidation; this rewrite is not runtime verification or implementation approval.

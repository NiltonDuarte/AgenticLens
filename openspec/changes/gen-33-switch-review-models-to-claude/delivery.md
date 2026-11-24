# chore(GEN-33): switch review models to Claude

## Outcome

The parallel review skill now runs every selected reviewer under Claude Opus 5.5 at medium effort and Claude Sonnet 5.5 at high effort. These two profiles replace the earlier GPT profiles ([design](design.md)). Coverage and attribution rules did not change.

- [model-policy.yaml](../../../.agents/skills/agentic-lens-review/references/model-policy.yaml) defines `opus_5_5_medium` and `sonnet_5_5_high` on provider instance `claudeAgent`, and both are default profiles.
- [model-attribution.md](../../../.agents/skills/agentic-lens-review/references/model-attribution.md) passes effort under the option name that `orchestrator_capabilities` lists for each provider. The old guidance hard-coded the Codex name `reasoningEffort`. For native Claude Code dispatch, the run records the acknowledged model ID rather than the alias. Until [AL-98](../../../alspaces/plan/to-do/AL-98%20generate%20ticket%20metrics%20from%20claude%20code%20sessions.md) ships, a Claude reviewer's identity is at most `acknowledged`.
- [parallel-review.md](../../../.agents/skills/agentic-lens-review/references/parallel-review.md) names the new profiles, and its summary example uses them.

Historical review summaries, archived changes and `alspaces/` records keep their GPT attribution. The `agentic-lens-review` tag stays at `0.1.1`, following existing practice; the design records why.

## Verification

Run on the final tree over main `9eeabb94`:

- The policy parsed with duplicate-key detection. It has version 1, two profiles with exactly the three required non-blank settings, a nonempty default list with no repeats and only defined profiles, and empty overrides.
- The T3 catalog lists `claudeAgent`, `claude-opus-5-5`, `claude-sonnet-5-5` and an `effort` option that accepts `medium` and `high`.
- A search of active guidance, scripts and tests found no other reference to the old models, profile IDs or `reasoningEffort`. No test parses the policy.
- All 14 relative links and anchors in the changed Markdown resolve. `git diff --check`, strict OpenSpec validation and the three OpenSpec workflow tests passed.
- Integrated `make` was skipped because no application code, tests, schemas or scripts changed.

## Limits

- No review ran under the new profiles in this change. The first review that uses them will confirm that real dispatch is accepted.
- **Telemetry.** This session's telemetry is unavailable. It ran in Claude Code, which AgenticLens cannot yet read as a source; AL-98 adds that. No ledger was generated, and no other harness's session was substituted.

## Pull request

Switch the parallel review skill's reviewer models from GPT to Claude, as the user decided: Claude Opus 5.5 at medium effort and Claude Sonnet 5.5 at high effort. Every selected reviewer still runs under both profiles.

- Replace the two profiles and defaults in `model-policy.yaml`. The provider instance is `claudeAgent`.
- Pass effort under the option each provider lists in `orchestrator_capabilities`, instead of the hard-coded Codex name `reasoningEffort`. Claude uses `effort`. Native Claude Code dispatch records the acknowledged model ID, not the `opus`/`sonnet` alias.
- Until AL-98 can read Claude Code sessions, a Claude reviewer's identity is at most `acknowledged`.
- Existing review evidence keeps its real GPT attribution.

These are documentation and configuration changes only, so integrated `make` was skipped. This session's telemetry is unavailable because it ran in Claude Code.

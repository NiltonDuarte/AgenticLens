# GEN-33 Claude review models

## Decision

The user explicitly approved two review profiles:

| Profile | Provider instance | Model | Effort |
|---|---|---|---|
| `opus_5_5_medium` | `claudeAgent` | `claude-opus-5-5` | medium |
| `sonnet_5_5_high` | `claudeAgent` | `claude-sonnet-5-5` | high |

They replace `sol_6_2_high` and `luna_6_medium` as the defaults in [model-policy.yaml](../../../.agents/skills/agentic-lens-review/references/model-policy.yaml). Every selected reviewer still runs under both profiles. The policy schema, the rules for overrides and coverage, and `reviewer_overrides: {}` stay the same. The T3 catalog (`orchestrator_capabilities`, checked on 2026-11-24) lists `claudeAgent` with both model IDs and the `effort` values `medium` and `high`.

## Provider-specific effort mapping

The dispatch guidance told the coordinator to pass effort as T3 `target.options.reasoningEffort`. That is the Codex option name. The Claude provider calls the same setting `effort`, so the old rule would have sent the requested effort under a key Claude does not read. The coordinator now uses whichever effort option `orchestrator_capabilities` lists for the selected provider and model, and always passes it explicitly. Native Claude Code dispatch through the Agent tool takes an alias (`opus`, `sonnet`) and a separate `effort` field. The run record stores the exact model ID the dispatch acknowledged, never the alias.

## Identity basis for Claude reviewers

AgenticLens cannot read Claude Code sessions until AL-98 ships. Until then, observed settings for a Claude reviewer stay null with that reason, and its `identity_basis` can be at most `acknowledged`. This follows the existing rule that requested or acknowledged identity is never reported as observed. The illustrative `REVIEW_SUMMARY.json` example now shows two `acknowledged` Claude reporters.

## Preserved history

Existing review summaries, archived changes and `alspaces/` records keep their real GPT attribution. They describe reviews that ran under the earlier policy. The `.codex/agents` persona registrations stay because they define reviewer roles, not models. No other active guidance, test or script names the old profiles.

## Skill version tag

`.agenticlens-config.yaml` tags tickets with `agentic-lens-review: "0.1.1"`. Since that tag was introduced, AL-84, GEN-27 and REV-8 changed the review skill without changing the tag, and no documented rule says when to bump it. This change follows that practice and keeps `0.1.1`. If the user wants metrics compared across this model switch, a tag bump is a separate decision. No roadmap ticket changes status.

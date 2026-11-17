# Plan the backfill rollout decision

## Goal

Give operators independently useful improvements to explicit refresh before asking them to evaluate a change in ordinary backfill.

## Context

The user requested a planning handoff from [AL-65 discovery, PR #145](https://github.com/NiltonDuarte/AgenticLens/pull/145). The [discovery report](../al-65-recommend-safe-source-backed-backfill-defaults/delivery.md) records the disposable probes and their limits.

## Delivery

- [AL-94 — Avoid unchanged writes on exact-source refresh](../../../alspaces/plan/to-do/AL-94%20avoid%20unchanged%20writes%20on%20exact%20source%20refresh.md)
- [AL-95 — Report exact-source refresh coverage](../../../alspaces/plan/to-do/AL-95%20report%20exact%20source%20refresh%20coverage.md)
- [AL-96 — Preview source-backed backfill by opt-in](../../../alspaces/plan/to-do/AL-96%20preview%20source%20backed%20backfill%20by%20opt%20in.md)
- [AL-97 — Decide whether to switch ordinary backfill defaults](../../../alspaces/plan/to-do/AL-97%20decide%20whether%20to%20switch%20ordinary%20backfill%20defaults.md)

Write suppression concerns publication; coverage concerns what the operator can learn even on a failed or write-free run. Both outcomes can ship independently. The preview needs both. Its evidence informs the final decision. FIX-19 owns ambiguous Codex snapshot selection, FIX-20 owns cancellation and REV-7 owns broader source read guarantees. These tickets consume the supported reader outcomes without expanding those scopes.

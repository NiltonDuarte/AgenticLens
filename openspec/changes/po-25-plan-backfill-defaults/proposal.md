# Plan the backfill rollout decision

## Goal

Give operators independently useful improvements to explicit refresh before asking them to evaluate a change in ordinary backfill.

## Context

The user requested a planning handoff from [AL-65 discovery, PR #145](https://github.com/NiltonDuarte/AgenticLens/pull/145). The [discovery report](../al-65-recommend-safe-source-backed-backfill-defaults/delivery.md) records the disposable probes and their limits.

## Delivery

- [AL-94 — Make explicit backfill repeatable and report refresh coverage](../../../alspaces/plan/to-do/AL-94%20make%20explicit%20backfill%20repeatable%20and%20report%20refresh%20coverage.md)
- [AL-95 — Preview source-backed defaults and decide on rollout](../../../alspaces/plan/to-do/AL-95%20preview%20source%20backed%20defaults%20and%20decide%20on%20rollout.md)

AL-94 delivers safe, observable behavior for the existing explicit command. AL-95 uses that behavior to evaluate the proposed default. The preview can end with a decision to retain current behavior. FIX-19 owns ambiguous Codex snapshot selection, FIX-20 owns cancellation and REV-7 owns broader source read guarantees.

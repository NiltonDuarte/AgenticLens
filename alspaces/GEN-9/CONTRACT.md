# GEN-9: AgenticLens Operator Skill

## Objective

Create a reusable `.agents/skills/agentic-lens-operator/` skill for operating AgenticLens from an adopting repository without assuming access to the AgenticLens source checkout.

## Required Scope

- Document installed CLI commands and arguments, including side effects and conditions for use.
- Explain the embedded dashboard and its pages, data sources, and interpretation boundaries.
- Explain shared and machine-local configuration, including safe reading and narrowly scoped edits.
- Cover schema migration and source-backed backfill as a two-stage process.
- Use ledger session user names to coordinate source-backed backfill with the developers whose machines contain the original sessions.
- Preserve partial migration state when source data is unavailable or a backfill cannot find the session.
- Require direct user authorization for changes; otherwise ask before changing configuration, ledgers, summaries, or installed binaries.
- Consult the public repository only as a last resort, matching the installed binary's exact release tag or commit.

## Out of Scope

- AgenticLens application code changes.
- Replacing the adopting repository's access controls.
- Inferring another developer's migration progress without direct team coordination.
- Clearing partial migration metadata when source-backed work has not completed.

## Acceptance Criteria

- Skill frontmatter has a clear, specific trigger description.
- `SKILL.md` routes to focused CLI, dashboard, and configuration references.
- Command reference covers root scan/inspection and `init`, `generate`, `backfill`, `migrate`, `server`, `version`, and wrapper `update` arguments.
- Migration guidance explains partial state, source ownership, local backfill, team coordination, and verification.
- Configuration reference distinguishes shared project settings from machine-local source paths and explains narrow authorized edits.
- Last-resort source inspection selects the installed version's matching public release tag or exact commit.
- Setup installs the operator skill and places new binaries under `.agents/skills/agentic-lens-operator/bin/`.
- Operator command guidance uses the canonical location directly and reports setup is needed if it is missing.
- Binary installation location is not duplicated in `.agenticlens-config.yaml`.
- Skill structure passes the Skill Creator quick validator.

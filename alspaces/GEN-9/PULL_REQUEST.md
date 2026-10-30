# feat: add AgenticLens operator skill

## Summary

Adds an `agentic-lens-operator` skill for operating an installed AgenticLens release from an adopting repository.

## Scope

- Added an operator entrypoint and focused references for CLI commands, dashboard usage, and configuration.
- Updated the setup skill to install the operator skill and put new binaries under `.agents/skills/agentic-lens-operator/bin/`.
- Made the operator reference use that canonical path directly; binary installation does not need a shared configuration key.
- Documented command arguments and write effects, including `init`, `generate`, `backfill`, `migrate`, `server`, `version`, root workspace inspection, and wrapper `update`.
- Explained the partial migration and source-backed backfill lifecycle, including per-developer source ownership and team coordination.
- Added authorized configuration editing guidance and exact-release source inspection as a last resort.

## Verification

- `quick_validate.py` reports the skill is valid.
- Compared root and `init`, `generate`, `backfill`, `migrate`, `server`, and `version` help with the installed binary.
- Merged `origin/main`; it was already up to date.
- Revalidated the operator skill after setup integration.
- No application code or tests were changed or run.

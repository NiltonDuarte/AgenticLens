# docs(readme): fix installation and CLI documentation (GEN-10)

## Summary
- Focus README installation guidance on the agentic setup prompt.
- Fix Codex session selection and required backfill source flags in README and Architecture examples.
- Remove fixed schema-version claims from Architecture documentation.
- Add the MIT license referenced by README.

## Context
See `alspaces/GEN-10/CONTRACT.md` and `alspaces/GEN-10/SUMMARY.md`.

## Verification
`make build` and `git diff --check` passed. Tests were not run; this is documentation and license work. Contributor release guidance was left unchanged per request.

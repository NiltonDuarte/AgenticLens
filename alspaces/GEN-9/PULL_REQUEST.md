# chore(setup): install OpenCode and document ChatGPT plan usage (GEN-9)

## Summary
- Install OpenCode v1.19.5 for the current WSL user.
- Document ChatGPT subscription sign-in and Codex/OpenCode usage for AgenticLens.

See `alspaces/GEN-9/CONTRACT.md`, `alspaces/GEN-9/USAGE.md`, and `alspaces/GEN-9/SUMMARY.md`.

## Verification
- Verified the installed OpenCode version and existing Codex ChatGPT login.
- OpenCode OAuth remains for the user to complete interactively.
- GEN-9 metrics include one Codex session. No OpenCode session was attributed because the only local OpenCode v2 session is unrelated and the current extractor expects the legacy database schema.
- Downgraded OpenCode to v1.19.5, the release line expected by the current AgenticLens OpenCode extractor.
- Purged the incompatible v2 OpenCode user data and reinstalled v1.19.5 with a clean database; verified the v1 `session` table initializes.
- OpenCode credentials were purged and require signing in again.

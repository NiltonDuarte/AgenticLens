# fix(skills): mandate agent-generated literal nonces for telemetry (FIX-5)

## Overview
Removes shell-based nonce generation (`NONCE=$(od -N 4 -tx1 /dev/urandom | head -1 | tr -d ' ')` and `--nonce $NONCE`) across operational guidelines, orchestrator instructions, and agent skill manifests. Mandates that the invoking agent directly generate a unique random nonce and embed the literal token into the CLI invocation arguments.

In OpenCode, tool execution parameters are persisted to SQLite (`part` table) prior to subshell execution. Using bash variable assignments and command substitutions results in the unexpanded string `"$NONCE"` being recorded in the database, causing Priority 1 exact nonce matching in `ResolveActiveRootSession` to fail. Injecting the concrete token directly ensures the database record contains the exact nonce before command execution starts, restoring deterministic session resolution.

- Contract: [alspaces/FIX-5/CONTRACT.md](alspaces/FIX-5/CONTRACT.md)
- Summary: [alspaces/FIX-5/SUMMARY.md](alspaces/FIX-5/SUMMARY.md)

## Summary of Changes

### Operational Guidelines & Agent Personas (`AGENTS.md`, `.opencode/agent/`)
- Updated `AGENTS.md` (Section 1, Rule 6): Replaced shell command substitution with literal token embedding (`--nonce <GENERATED_NONCE>`). Documented the OpenCode SQLite persistence timing invariant.
- Updated `.opencode/agent/orchestrator.md` (Step 6): Directed orchestrator agent to generate random nonces directly and pass the literal token.

### Skill Manifests (`.opencode/skills/`)
- Updated `.opencode/skills/agentic-lens-dev/SKILL.md` (Phase 3, Step 6): Replaced shell command substitution with literal token embedding.
- Updated `.opencode/skills/agentic-lens-po/SKILL.md` (Section 2, Step 4): Enforced direct literal nonce generation while preserving all Product Owner CLI flags.
- Updated `.opencode/skills/agentic-lens-setup/SKILL.md` (Step 3 & Step 4): Updated the CLI walkthrough and proposed policy template block for external repository adoption.

### Standards & Quality
- Zero emojis across all modified files, commits, and artifacts.
- Zero modifications to out-of-scope backend packages or historical workspace archives.

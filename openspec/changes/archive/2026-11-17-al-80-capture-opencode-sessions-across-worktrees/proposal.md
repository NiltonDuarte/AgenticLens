# AL-80 OpenCode capture across worktrees

## Why

A developer can install one OpenCode plugin and collect the active session's ticket metrics in any configured checkout. The current project plugin embeds its installing checkout, source and state paths. AL-80 moves runtime routing into the native path already used for worktree capture and keeps OpenCode's binding tool tied to the actual session.

## What Changes

The [roadmap ticket](../../../../alspaces/plan/done/AL-80%20capture%20opencode%20sessions%20across%20worktrees.md) selects [PO-20 F1–F6](../../../../alspaces/PO-20/SPEC.md). The [global OpenCode capture capability](specs/global-opencode-capture/spec.md) applies to the user installer, TypeScript integration, manager dispatch, engine capture and operator documentation. Codex, manual generation and portable ledgers share the existing configuration and publication semantics.

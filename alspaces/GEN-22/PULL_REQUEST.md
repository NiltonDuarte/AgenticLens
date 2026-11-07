# chore(capture): record worktree troubleshooting and installation gaps (GEN-22)

Capture hooks installed only inside a linked Git worktree were invisible to the inspected Codex runtime. This draft records the discovery evidence, the intended user-level installation model, and the remaining native routing and configuration gaps in `CAPTURE_VISION_AND_GAPS.md`.

The shared AgenticLens pin now matches the locally built development binary, `v3.0.0-rc.131+03f5b05-2-g57a3503e`, as requested. This is not a published-release upgrade.

The Codex-home hook configuration, temporary local Python launcher, capture database, trust settings and per-worktree installation files are machine-local and are not included. Native Codex discovery was observed; live callback delivery, binding, extraction and publication remain unverified. No application fix is implemented by this PR.

The user requested collaborative troubleshooting and discussion notes rather than a feature-development plan or contract. See `alspaces/GEN-22/SUMMARY.md` for evidence and limits.

# chore(capture): document agreed worktree capture behavior (GEN-22)

Capture hooks installed only in a linked worktree were invisible to the inspected Codex runtime. This draft records the investigation and the user-confirmed design for user-level Codex/OpenCode integration, whole-file local configuration fallback, native routing, explicit ticket binding, and autonomous publication.

The decisions include a shared version pin with optional preferred development output, exact cached fallback when that output is absent, explicit binary provisioning, bounded recovery, and retention after worktree deletion. The decision review is closed; these notes do not implement the proposed behavior or constitute a plan/contract.

The shared pin matches the locally built development binary, `v3.0.0-rc.131+03f5b05-2-g57a3503e`, as requested. Machine-local hooks, the temporary Python launcher, databases, trust settings and per-worktree installation files are excluded. Live capture remains unverified.

See `alspaces/GEN-22/CAPTURE_VISION_AND_GAPS.md` for confirmed behavior and `SUMMARY.md` for investigation and validation evidence.

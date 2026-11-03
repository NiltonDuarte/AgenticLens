# GEN-20: Clean up deleted-thread worktrees

Initial state: 33 registered checkouts; clean current checkout at 8d494d2. T3 retained only two top-level threads. Read-only T3 state inspection found only this thread with an active provider session; surviving child threads had idle provider sessions.

Scope: remove 27 clean unused checkouts without force; retain main, this thread, MGT-2, and three dirty checkouts (GEN-1 guide, AL-62, release-v2-0-0). Keep all local branches and commits. Preserve AL-68 ignored capture data under `/home/niltonduarte/.t3/scratch/GEN-20/t3-16a70bd4-capture/` outside Git. Detailed initial inventory and removal outcomes are in iterations/01.

T3's worktree handoff tool explicitly states: “The worktree stays after the thread is deleted.” This establishes documented behavior, not a cleanup failure. The tool does not document the design rationale. T3 database records and application code were not modified.

Verification and final telemetry pending.

# GEN-20: Worktree cleanup contract

User request: clean up worktrees and explain why deleting a T3 thread leaves its checkout.

Remove clean AgenticLens checkouts belonging to deleted top-level threads. Preserve the main checkout, both remaining top-level thread worktrees, all tracked/untracked edits, local branches and commits, and local capture records. Use ordinary `git worktree remove` without force. Application code, T3 implementation, and T3 database records are outside scope.

GEN-13 is already allocated to concise project documentation; GEN-20 was free in the workspace root, ticket index, and local branches. Initial checkout was clean at 8d494d2. The generated GEN-1 branch name was not used as a ticket allocation.

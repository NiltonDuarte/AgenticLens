# feat(capture): capture OpenCode sessions across Git worktrees

## Pull request

OpenCode users can install a single capture plugin with `capture install-user --source opencode`. Each callback supplies its native session context to the installed manager. The selected engine applies the checkout's capture configuration and uses the shared worker to publish bound sessions. Operators can preview installation, explicitly migrate an unchanged legacy project plugin and remove the owned global plugin.

Verification includes native OpenCode 1.19.5 with a server outside the session worktree, a root binding through the plugin tool, a child created through OpenCode's task tool, and worker-produced ticket output. The local test provider supplied deterministic model responses. OpenCode supplied both session identities and their relationship. See the iteration evidence for the tested interface and source-family boundaries.

## Delivery status

The candidate is under independent review. The initial worktree revision was `b6ab0471`. Integration, final checks and publication are pending.

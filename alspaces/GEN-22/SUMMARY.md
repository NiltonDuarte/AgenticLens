# GEN-22: Capture installation and worktree troubleshooting

## Delivered repository changes

- Recorded the requested capture vision, current gaps and unresolved decisions in [CAPTURE_VISION_AND_GAPS.md](CAPTURE_VISION_AND_GAPS.md). Proposed configuration precedence is explicitly distinguished from accepted requirements.
- Updated shared `agenticlens.version` to the existing local build's exact version, `v3.0.0-rc.131+03f5b05-2-g57a3503e`, with explicit user authorization. The earlier pin had no matching published release or local cache leaf.
- Prepared a draft PR. No application code, installer fix or distributed integration is implemented by these changes.
- The user requested deleting formal plans and continuing troubleshooting together. Both temporary plans and empty iteration directories were removed; no contract was created.

## Environment and investigation evidence

Initial Git status was clean at `57a3503e3024f3993ce9391fee604aad0bcc44d5`; the initial branch was literally `GEN-<ID>/install-capture-plugin`. The delivery branch is `chore/GEN-22-capture-worktree-findings`.

The environment is Linux amd64 on WSL2, with Codex CLI 0.167.0 and source records under `~/.codex`. Configured project is AgenticLens, workspace root is `alspaces`, and shared capture state is `~/.agenticlens/state`.

`make verify-build PLATFORMS=linux/amd64` passed, including the canonical frontend build, embedded asset synchronization, host-native compilation and version-command checks. The local binary identifies commit `57a3503e` and hook protocol 3. Its SHA256 is `e16b0c141e2729654fbb0df107f7d00ee5e0a81a2eac0ccba61d8b77061b6e1a`.

Installed the native build offline using the existing `--local-bin` installer and the explicit checksum after matching the shared pin. This proved that no new development installation feature or published release was needed for this installation. Selection independent of the shared pin remains a separate discussion point.

Codex showed zero hooks despite consistent worktree-local installation files. Adding trust for this exact worktree, explicitly authorized by the user, did not resolve discovery. Source-path comparisons in isolated ordinary repositories and attached/detached worktrees established the cause: Codex 0.167.0 reads linked-worktree project hooks from the original checkout. The actual original checkout lacked hooks. The exact six AgenticLens groups load correctly in an ordinary repository. Both inspected Codex 0.166.0 and 0.167.0 returned zero hooks on the actual worktree before the operational repair.

The parent and a planning investigator, both gpt-6.1-sol with high reasoning effort, established and checked the discovery evidence. Isolated controls ruled out hook syntax, worktree trust alone, stale user configuration, branch spelling and generic ~/.t3 ancestry. An initially misleading positive worktree control was corrected by checking returned sourcePath rather than counts. These were troubleshooting investigations, not a formal product-code review.

## Machine-local operations and current limits

The following are not repository deliverables:

- Stable wrapper at `~/.local/bin/agentic-lens` and the matching native version under `~/.agenticlens/bin/`.
- Fresh capture state under `~/.agenticlens/state/`. The user explicitly authorized deleting legacy databases and then all old state/recovery folders while preserving bin; capture was reinstalled afterward. No raw Codex source sessions were deleted.
- The approved worktree trust entry in `~/.codex/config.toml`. Individual hook trust was not granted automatically.
- A guarded original-checkout hook file briefly demonstrated native discovery. It was subsequently deleted at the user's request.
- Six user-level hooks at `~/.codex/hooks.json`, invoking a local launcher at `~/.codex/hooks/agenticlens-capture.py`. The Python file was created during troubleshooting; it is not shipped AgenticLens code. The user's target is native routing without this workaround.
- Worktree `.codex/hooks.json` and `.codex/agenticlens-installation.json`, retained because current receiver/launcher validation depends on them. Both were unstaged and excluded through repository-local `.git/info/exclude` at the user's request. No shared ignore policy was changed.

The home launcher selects an existing active worktree installation and its pin, source and state. Read-only checks selected this worktree and a subdirectory, skipped uninstalled/inactive projects and rejected a mismatched-worktree fixture. Native Codex 0.167.0 subsequently listed exactly six user-source hooks with no warnings/errors or duplicates. All six were untrusted at the last discovery check.

Live callback delivery, an explicit session-binding exchange, source extraction and automatic publication remain unverified. No fabricated callback or session was used to claim those outcomes. The local setup does not implement automatic adoption of new configured worktrees or base-checkout local-configuration fallback; those gaps are documented for further collaboration.

## Delivery verification and telemetry

Repository changes are limited to the approved pin and ticket documentation. Delivery integration/check results and session extraction facts are recorded below when available. No implementation review or platform-wide compatibility claim is made.

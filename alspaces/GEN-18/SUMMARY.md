# GEN-18: Stable installation instructions

## Baseline and allocation

Initial working tree: clean. Baseline: b1742e96e7631330ba7d3de2a66a5ad82606c259. Original generated branch used GEN-13, already allocated; inspected workspaces, ticket index and remote GEN branches before allocating GEN-18. GEN-14 is reserved remotely. Work branch: chore/GEN-18-stable-installation.

## Delivered behavior

README and setup prompt resolve the latest stable release once and read its tagged manifest. Manual wrapper and Bash/PowerShell binaries share a selected tag. Setup and operator recovery guidance retain installed versions for repair and prohibit fallback to main. Older manifest download references are explicitly pinned. Newer bundle commands are conditional on release contents.

## Evidence and limits

GitHub API confirmed latest stable v2.0.0 (draft=false, prerelease=false). Its setup manifest is available; its bundle installer is absent (HTTP 404). Inspected its manifest and found moving main references, so prompts explicitly override those references. Release assets include the documented Linux amd64, macOS arm64 and Windows amd64 binaries. Historical ticket artifacts are preserved.

Documentation-only task: direct maintenance under AGENTS.md; no delegated implementation or independent application review. Verification and delivery results will be recorded before telemetry generation.

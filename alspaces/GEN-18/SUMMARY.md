# GEN-18: Stable installation instructions

## Baseline and allocation

Initial working tree: clean. Baseline: b1742e96e7631330ba7d3de2a66a5ad82606c259. Original generated branch used GEN-13, already allocated; inspected workspaces, ticket index and remote GEN branches before allocating GEN-18. GEN-14 is reserved remotely. Work branch: chore/GEN-18-stable-installation.

## Delivered behavior

README and setup prompt resolve the latest stable release once and read its tagged manifest. Manual wrapper and Bash/PowerShell binaries share a selected tag. Setup and operator recovery guidance retain installed versions for repair and prohibit fallback to main. Older manifest download references are explicitly pinned. Newer bundle commands are conditional on release contents.

## Evidence and limits

GitHub API confirmed latest stable v2.0.0 (draft=false, prerelease=false). Its setup manifest is available; its bundle installer is absent (HTTP 404). Inspected its manifest and found moving main references, so prompts explicitly override those references. Release assets include the documented Linux amd64, macOS arm64 and Windows amd64 binaries. Historical ticket artifacts are preserved.

Documentation-only task: direct maintenance under AGENTS.md; no delegated implementation or independent application review. Verification and delivery results will be recorded before telemetry generation.

## Final verification and integration

Reviewed documentation revision: ad66a54. `git fetch origin main` and `git merge origin/main --no-edit`: already up to date at b1742e96e7631330ba7d3de2a66a5ad82606c259; no integration diff. The full make run covered the unchanged implementation tree before and after that no-op integration.

`make`: PASS, including Go vet and race tests, frontend typecheck/lint, 35 frontend files / 783 tests, 6 release workflow tests, 12 installation tests, embedded frontend build and platform binary builds. Existing Vite bundle-size warning remains. No new tests were needed for documentation changes.

`git diff --check`: PASS. Active README/docs/skills search found no remaining main installation URLs or moving latest/download asset URLs. README/manual prompts match. Changed manual Bash installation block and downloaded v2.0.0 wrapper pass `bash -n`. Stable wrapper URL and tagged manifest availability were verified; published platform names match the examples. PowerShell was inspected, not executed; no adopting repository installation was performed.

No matching GEN-18 planning specification or index entry existed, so no plan lifecycle move is required. Raw session transcripts remain outside Git. New ledger has no prior sessions from other harnesses to preserve. No child agents were used for this documentation-only task.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 2
- **Active Agent Duration:** 0m 0s (excludes user wait time)
- **Context Window Health:** HEALTHY — Root agent peak 54.2k tokens (27% of 200k healthy ceiling, 0 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 1,018 reasoning tokens (12.5% of generated output)
- **Tool Health & Self-Correction:** 20 tool calls (3 self-corrections, 85.0% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 37,712 | 7,121 | 773,120 | 95.3% | 54.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **37,712** | **7,121** | **773,120** | **95.3%** | **54.2k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 1 | 37,712 | 7,121 | 773,120 | 95.3% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **37,712** | **7,121** | **773,120** | **95.3%** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

Telemetry validation: ledger contains the actual Codex root `01a1a0b7-9371-78a2-912d-92fd7730bb7c`, source `codex`, exactly one session and no children, matching the rollout metadata and no-delegation workflow. Ticket GEN-18 is classified as chore. No prior ledger existed. Extraction completed without source errors.

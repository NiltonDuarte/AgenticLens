# fix(AL-73): verify the complete release inventory before installation

## Pull request

Release packaging and verification now derive the required files from the selected version and platforms. Each archive, manager and engine needs one well-formed SHA-256 entry. Packaging validates the staged manifest and file contents before replacing the final manifest. Verification repeats those checks before invoking a selected host binary.

The setup downloader requires a successful digest computation and one valid record for the selected manager. The bundle installer and binary bootstrap validate remote repository overrides before fetching from the exact tag. A later component failure exits unsuccessfully while preserving previously completed engine installs and the failed component’s prior state.

Local fixtures exercised partial manifests, hash and sort failures, stale host files, malformed manager records, exact URLs and total and partial installation failure. Receipt and binary checks covered a successful retry.

The full integrated `make` and Linux host release build, package and verification passed at `f4a9a150`. A failed repackage can leave a previous manifest beside changed build outputs. Rerun packaging and verification before using that directory.

## Verification

At baseline `b6ab0471` plus product/test diff `47fd2495ce`, the tester exercised all 16 installation cases, four platform/packaging cases, 23 capture distribution cases and the native manager Go suite on Linux. The behavior reviewer also passed an end-to-end Make-to-installed-bundle probe with separately selected engine and manager versions. A nonprivileged fixture using an unreadable regular binary made independent verification exit 2 before any host probe.

The branch integrated fetched main at `b6ab0471` with no source changes. Full `make` passed, including race-enabled Go tests, lint, vet, installer and capture tests, 804 frontend integration tests and the multi-platform build. Local `release-build`, `release-package` and `release-verify` ran against separately built Linux/amd64 binaries reporting commit `f4a9a150`.

## Review

Final consolidated outcome: approved. Each of the four focuses ran independently on gpt-6.1-sol/high and gpt-6-luna/medium. Seven reviews approved their scope. The data reviewer on gpt-6-luna/medium requested changes after reproducing the previous manifest beside changed build inputs. The parent confirmed that the build inputs already failed independent verification before the repackage and still failed afterwards. All three release workflows require successful packaging and verification before upload. Restoring old build inputs is outside the manifest transaction. The tests reviewer on gpt-6-luna/medium identified the unreadable-file coverage limit, which the parent verified directly on Linux. [Retained findings and dispositions](evidence/iterations/01/REVIEW_SUMMARY.json).

## Limits

Local verification exercises the host binaries on Linux. Foreign artifacts receive checksum and platform checks. Runtime identity is deferred to their target host. The manifest establishes integrity relative to the selected release publisher. Hosted publication and tag mutation are outside this delivery.

The installed manager could not read this checkout’s shared policy. Final extraction used the integrated native engine and an ignored worktree-local configuration that preserves the machine’s configured Codex source and sets the explicit publication state directory. Native children retain their source ancestry. T3’s four Luna review sessions expose separate Codex roots. All 12 source sessions were included. The final root turn was open at extraction and its duration is provisional.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 20
- **Active Agent Duration:** 1h 16m (recorded duration including 22s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 129.0k tokens (29% of 450k healthy ceiling, 0 compactions) | Subagents peak 106.7k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Planner 1.0, Coder 1.0, Tester 1.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 14,777 reasoning tokens (17.6% of generated output)
- **Tool Health & Self-Correction:** 282 tool calls (18 self-corrections, 93.6% success rate)

### Anomalies & Health Flags
- **Status:** 2 execution anomalies detected (1 CRITICAL, 1 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (11 cycles) — Subagent roles 'coder' and 'tester' alternated for 11 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (3 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 69,724 | 3,717 | 627,328 | 90.0% | 67.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 80,582 | 7,211 | 2,275,968 | 96.6% | 101.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 45,587 | 6,548 | 3,332,096 | 98.7% | 106.7k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 5 | 477,342 | 33,679 | 14,791,680 | 96.9% | 129.0k | HEALTHY | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **reviewer-architecture** | 1 | 49,472 | 3,287 | 445,824 | 90.0% | 64.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-behavior** | 1 | 67,061 | 6,384 | 822,272 | 92.5% | 70.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-data** | 1 | 74,533 | 3,753 | 426,752 | 85.1% | 69.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-tests** | 1 | 64,174 | 4,524 | 759,680 | 92.2% | 67.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **12** | **928,475** | **69,103** | **23,481,600** | **96.2%** | **129.0k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 4 | 232,911 | 14,794 | 2,030,080 | 89.7% | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **gpt-6.1-sol** | 4 | 255,240 | 17,948 | 2,454,528 | 90.6% | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **gpt-6.2-astra** | 4 | 440,324 | 36,361 | 18,996,992 | 97.7% | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **Total** | **12** | **928,475** | **69,103** | **23,481,600** | **96.2%** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

Retained recorded history: 4 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->

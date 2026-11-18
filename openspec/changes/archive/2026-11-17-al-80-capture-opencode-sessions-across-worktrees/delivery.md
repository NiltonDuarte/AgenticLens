# feat(capture): capture OpenCode sessions across Git worktrees

## Pull request

OpenCode users can install a single capture plugin with `capture install-user --user-source opencode` or select `--capture-source opencode` during setup. Each callback supplies its native session context to the installed manager. The selected engine applies the checkout's capture configuration and uses the shared worker to publish bound sessions. Operators can preview installation, explicitly migrate an unchanged legacy project plugin and remove the owned global plugin.

Verification includes native OpenCode 1.19.5 with a server outside the session worktree, a root binding through the plugin tool, a child created through OpenCode's task tool, and worker-produced ticket output. The local test provider supplied deterministic model responses. OpenCode supplied both session identities and their relationship. Integrated `make` passed after merging AL-73 and correcting the setup handoff. Checksummed binaries from mismatched manager and engine releases now reject unsupported OpenCode activation before changing either harness integration. OpenCode 2.2.0 supported manual source extraction but did not discover the automatic plugin.

## Integration with AL-73

Merge `9c6d56de` brought AL-73’s `54989015` into the AL-80 branch. The shared downloader still selects exact release tags, checks the admitted digest before executing a bootstrap binary, and hands off capture only when explicitly requested.

The `gpt-6.2-sol/high` behavior reviewer reproduced two failures with the AL-73 binaries. An older manager could install a plugin whose callbacks it could not route. An older engine interpreted an explicit OpenCode user install or removal as a Codex operation. The reviewer also identified stale setup guidance. Commit `d58953aa` added a manager capability check and an explicit user-source argument that older engines reject during parsing. Operators can use the guarded argument for both installation and removal. Generic protocol compatibility and the existing Codex default remain supported.

At `d58953aa`, `GOMAXPROCS=4 GOFLAGS="-p=1 -timeout=35m" make` passed. It included 16 installation tests, four platform and release-artifact tests, 24 capture-distribution tests, the race-enabled Go suites, 231 frontend unit tests, 804 frontend integration tests and four host-target builds. The integration probes used isolated artifacts and local release fixtures. They did not contact hosted release assets or execute macOS binaries.

The focused command `go test -count=1 ./cmd/agentic-lens -run "TestOpenCode(UserInstall|UserMigration|HookAdmission)"` passed in `backend/`. The corrected distribution regression `python3 scripts/tests/test_capture_distribution.py CaptureDistributionTest.test_opencode_bootstrap_requires_matching_callback_interfaces` exercised setup, direct activation and removal across the version boundary. The behavior reviewers on `gpt-6.2-sol/high` and `gpt-6-luna/medium` approved the correction.

In disposable Linux homes, bad manager and engine digests each stopped setup before installation writes. The default installed binaries only. Matching versions installed the global plugin and its receipt without allocating capture state. Using the actual AL-73 manager stopped explicit OpenCode setup before binary installation. An AL-73 engine paired with a current manager rejected capture after the requested binaries were installed, leaving both harness integrations untouched. The native 1.19.5 session and publication evidence below carries forward because the callback implementation, source adapters and worker are unchanged.

## Verification and review

The original candidate was `2f794e38` on `b6ab0471`. The native run verified three simultaneous roots and three children, including concurrent contributors to one ticket and a different ticket in the base checkout. Resuming from a disabled sibling delivered all 11 observed callbacks to the original worktree. A copied mixed-harness ledger retained its OpenCode owner while an exact Codex refresh replaced its token count. [Native verification](evidence/iterations/01/native-verification.md)

The behavior, data, architecture and security reviewers on `gpt-6.1-sol/high` independently found that legacy installation could add a duplicate while the default user plugin remained loaded beside a custom config directory. The architecture reviewer on the same model also demonstrated that a migrated plugin in the default directory could bind a session without scheduling publication. Both regressions failed on the initial candidate and passed after the installer and worker began checking both loaded homes. Native resume evidence resolved the `gpt-6-luna/medium` architecture reviewer's question about routing callbacks that carry only a session ID. The behavior, data, architecture, test and security re-reviews accepted the corrected candidate under both profiles.

Automatic capture was demonstrated on OpenCode 1.19.5. OpenCode 2.2.0 read its v2 source successfully through manual generation and did not discover this plugin. The installer can inspect the configured repository, its base and the current runtime's loaded homes. An unrelated checkout with a legacy plugin reports a conflict at callback admission and requires explicit migration. With no authoritative runtime database path, the integration cannot independently compare that path to the configured store.

The preliminary full build hit deadline failures during concurrent backend suites on this machine. The integrated `GOMAXPROCS=4 GOFLAGS="-p=1 -timeout=35m" make` passed. It ran Go race tests, installation and capture distribution checks, frontend type checking and lint, 231 frontend unit tests, 804 frontend integration tests, and built the manager and engine for Linux and macOS on amd64 and arm64. Consolidated review approved the candidate. [Retained review findings](evidence/iterations/01/REVIEW_SUMMARY.json)

Telemetry includes the Codex root, its ten native children and twelve separately rooted tool-delegated reviews. The source reported the current turn as open and costs as unavailable. A temporary worktree-local configuration supplied the required publication state directory while retaining the base checkout’s disabled capture setting.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 47
- **Active Agent Duration:** 2h 53m (recorded duration including 1130s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 206.5k tokens (46% of 450k healthy ceiling, 0 compactions) | Subagents peak 175.8k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.1 avg turns/run (peak: 4 turns) | Planner 1.0, Coder 2.0, Tester 4.0, Reviewer 2.0, Reviewer-architecture 2.0, Reviewer-behavior 2.0, Reviewer-data 2.0, Reviewer-tests 2.0
- **Test-to-Production Code Ratio:** 0.3:1 (+251 test lines / +878 prod lines)
- **Code Delivered:** +1,129 / -137 lines across 13 code files (excluding docs)
- **Reasoning Effort:** 51,260 reasoning tokens (18.2% of generated output)
- **Tool Health & Self-Correction:** 930 tool calls (64 self-corrections, 93.1% success rate)

### Anomalies & Health Flags
- **Status:** 9 execution anomalies detected (3 CRITICAL, 6 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: planner <-> tester (5 cycles) — Subagent roles 'planner' and 'tester' alternated for 5 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (36 cycles) — Subagent roles 'coder' and 'tester' alternated for 36 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> reviewer-behavior (5 cycles) — Subagent roles 'tester' and 'reviewer-behavior' alternated for 5 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+80.2k tokens) — Context tokens surged by 80.2k to 126.2k tokens.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> tester (3 cycles) — Subagent roles 'reviewer-behavior' and 'tester' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> tester (3 cycles) — Subagent roles 'reviewer-data' and 'tester' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (3 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (3 cycles) — Subagent roles 'tester' and 'coder' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (4 cycles) — Subagent roles 'coder' and 'tester' alternated for 4 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 57,010 | 3,086 | 340,864 | 85.7% | 55.7k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 2 | 344,906 | 37,186 | 12,681,728 | 97.4% | 159.9k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **tester** | 1 | 310,825 | 33,587 | 16,675,072 | 98.2% | 175.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer** | 1 | 84,807 | 8,421 | 2,340,864 | 96.5% | 99.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 13 | 1,160,867 | 88,335 | 49,062,144 | 97.7% | 206.5k | HEALTHY | Unavailable (0 observed, 0 partial, 13 unavailable / 13 invocations) |
| **reviewer-architecture** | 1 | 124,934 | 14,071 | 4,253,184 | 97.1% | 145.1k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-behavior** | 2 | 226,358 | 22,700 | 5,594,112 | 96.1% | 115.3k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-data** | 1 | 113,011 | 14,924 | 4,472,576 | 97.5% | 133.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-tests** | 1 | 87,858 | 7,720 | 1,680,768 | 95.0% | 104.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **23** | **2,510,576** | **230,030** | **97,101,312** | **97.5%** | **206.5k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 23 unavailable / 23 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 12 | 728,311 | 45,143 | 6,542,592 | 90.0% | Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations) |
| **gpt-6.1-sol** | 5 | 517,320 | 54,805 | 15,518,080 | 96.8% | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **gpt-6.2-astra** | 5 | 1,145,297 | 117,051 | 72,217,216 | 98.4% | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **gpt-6.2-sol** | 1 | 119,648 | 13,031 | 2,823,424 | 95.9% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **23** | **2,510,576** | **230,030** | **97,101,312** | **97.5%** | **Unavailable (0 observed, 0 partial, 23 unavailable / 23 invocations)** |

Retained recorded history: 12 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->

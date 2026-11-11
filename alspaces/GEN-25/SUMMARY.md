# GEN-25: Remove Windows support

AgenticLens now targets Linux and macOS on amd64 and arm64. Windows native adapters, executable naming and PE admission, the PowerShell installer, Windows hook/plugin transports and Windows execution suites were removed. Builds, installation selection, release checksums and uploads use the four maintained targets; explicit unsupported selections fail before downloads or installation changes. Active setup/operator/contributor guidance and bundle contents match that platform boundary.

Linux/macOS retain exact engine pins, separate manager/engine identities, bounded handoff and cancellation, literal arguments, installation ownership and atomic publication. An exact signature recognizes the previous generated Git advisory hook for safe refresh/removal without generating or executing its retired script. Edited hooks remain protected. Stale unsupported distribution files remain untouched and are excluded from checksums/uploads.

Historical tickets/releases and portable ledger facts were preserved. Remaining Windows-looking production data inspection only detects legacy user-configuration ownership; historical fixture bytes and negative/security/path cases do not add runtime support. No schema/history migration or new platform support was introduced.

## Scope and evidence

The configured workspace is `alspaces`; GEN-25 was allocated after workspace/index/ref checks. Initial status was clean at `4e5c458322a2cddfdb5911b26a4569ea75f06a09`. The placeholder branch became `chore/GEN-25-remove-windows-support`. The user approved the [contract](CONTRACT.md) and [iteration plan](iterations/01/PLAN.md) with “Approved” on 2026-11-11. No matching roadmap specification/index entry required lifecycle changes.

Implemented candidate `e67991f3` has reviewed tree `70fc7848f062903fb2a2c199234ce59250c946e4`; merging fetched `origin/main` was already up to date, so all scoped review evidence carries forward without a product diff. Integrated `make` passed on this revision: Go units/vet/race integrations, TypeScript typecheck/lint, 227 frontend unit tests, 751 frontend integration tests, all Python boundary suites and eight manager/engine builds. Final parent consolidated verdict: APPROVED; no unresolved finding or material evidence gap.

Coder verification passed Go manager/config/capture/CLI packages, installation (10), capture distribution (23), release revision (14), orchestration (6), and supported plugin VM argv/payload checks. Tester added real unsupported-selection/no-network/no-write checks, a negative real PE fixture, real tar/checksum packaging and all-three upload-selection checks with stale files. Targeted historical-hook tests cover upgrade/removal previews, actual operations and edited refusal. Disposable regressions failed with the prior hook/checksum/upload logic and passed with corrections. Disposable canonical asset builds passed with absent and repeated destinations; current frontend/embedded index, JS and CSS match. Windows-only execution assertions were retired; distinct supported arguments, deadlines, ownership, rollback and portable-data assertions remain.

## Independent review

Four focuses each ran `gpt-6.1-sol`/high and `gpt-6-luna`/medium in discovery and affected re-review. Dispatch accepted those settings; parent subsequently verified all 20 root/child sessions and exact model/effort profiles from local Codex session metadata. Reviewers themselves did not have that runtime metadata. Initial behavior and architecture Sol runs requested changes; the remaining discovery runs approved.

The coder (`gpt-6.1-sol`/high) discovered legacy-hook upgrade refusal. Behavior Sol discovered stale Windows release inclusion, and behavior plus architecture Sol independently found stale ZIP guidance. All three were verified and fixed. Corrected rechecks approved except test-effectiveness Luna's additional execution-sentinel request. Parent refuted that as a material gap: `InstallCheck` and its helpers never evaluate hook contents; the only subprocess resolves Git's hooks directory, and hook contents are read, compared/hashed, written or removed. That dissent and every reporting tuple remain in the [retained review findings](iterations/01/REVIEW_SUMMARY.json). Parent verification is separate from discovery credit.

Real release verification also passed: copied all eight built binaries to a disposable directory and ran `make -f <repository>/Makefile release-package release-verify VERSION=v9.8.7 WEB_DIR=absent-web`; four archives and all 12 checksums verified, and host manager/engine executed. Both host binaries also reported their own build identity with `PATH=/nonexistent`. `file bin/agentic-lens-*` confirmed ELF/Mach-O architectures. Logs stay outside Git (`/tmp/GEN-25-make.log`, `/tmp/GEN-25-real-release.log`).

## Direct follow-up audit

At the user's request, the parent alone audited `528726d6` without invoking any subagent. Verdict: APPROVED for the Windows-removal boundary; no additional active Windows support or new finding. The working tree was initially clean. Tracked filenames and active source/docs/scripts/workflows were reconciled against Windows/Win32/MSYS/PowerShell/PE/EXE/ZIP/process/profile terms, then platform parsing, native format admission, build constraints, installer bundle/entrypoints and release selection were traced directly.

Remaining application inspection of `commandWindows` only protects legacy integration ownership; it never generates or executes Windows commands. Other matches are explicit unsupported notices, negative tests, sanitized historical hook bytes, security/portable-path fixtures, test environment isolation, unrelated collection methods or context/time windows. These do not justify deleting compatibility evidence or shared history.

Fresh verification: focused manager platform-selection/no-network/no-write/real-PE-rejection tests passed (`go test ./internal/adapters/manager -run 'TestSupportedPlatformSelections|TestUnsupportedPlatformInstallHasNoNetworkOrInstallationWrites|TestWindowsPEArtifactCannotBeAdmitted' -count=1`, 0.338s); `python3 scripts/tests/test_platform_support.py` passed both bootstrap/packaging/upload-selection cases (0.871s). Disposable `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build` probes for both application commands failed as expected at unavailable OS-specific config locking symbols and produced no executable. No application/test/build files changed, so prior integrated `make`, supported four-target builds and native-host evidence remain valid; native macOS execution remains unverified. The user's no-subagent instruction governs this additional direct pass; no new scope/acceptance iteration was introduced.

## Limits

Native macOS execution, live OpenCode lifecycle delivery and hosted release publication were not exercised. Cross-compilation is distinct from native runtime certification. Dependency installation reported existing npm audit findings; dependency remediation is outside this support-removal ticket. Final source-Codex telemetry used literal nonce `a6d408c13fe729b0`, conventional type `chore`, and the configured normal Codex store. Exact session IDs match the root plus all 19 expected children; root delegation links include every child. This new ticket had no prior ledger or other-harness history to preserve. The source reports the root turn open with provisional seconds and role `unknown`; caller verification and final root duration are unavailable at extraction, and later publication calls are outside its cutoff. Those facts remain unmodified in the ledger; no empty or manufactured completion was substituted.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 5
- **Active Agent Duration:** 1h 13m (recorded duration including 127s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 185.6k tokens (41% of 450k healthy ceiling, 0 compactions) | Subagents peak 164.2k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.2 avg turns/run (peak: 3 turns) | Planner 1.0, Coder 2.0, Tester 3.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 6.1:1 (+288 test lines / +47 prod lines)
- **Code Delivered:** +335 / -1,149 lines across 51 code files (excluding docs)
- **Reasoning Effort:** 27,966 reasoning tokens (18.2% of generated output)
- **Tool Health & Self-Correction:** 522 tool calls (20 self-corrections, 96.2% success rate)

### Anomalies & Health Flags
- **Status:** 5 execution anomalies detected (3 CRITICAL, 2 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (10 cycles) — Subagent roles 'coder' and 'tester' alternated for 10 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (7 cycles) — Subagent roles 'coder' and 'tester' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (5 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 5 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-behavior' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 71,539 | 2,276 | 383,744 | 84.3% | 72.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 134,244 | 16,109 | 6,157,696 | 97.9% | 164.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 64,985 | 13,497 | 4,539,904 | 98.6% | 133.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 177,501 | 41,286 | 23,745,664 | 99.3% | 185.6k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 4 | 219,891 | 14,252 | 1,749,248 | 88.8% | 91.9k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-behavior** | 4 | 217,841 | 13,316 | 1,834,112 | 89.4% | 86.0k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-data** | 4 | 213,544 | 11,960 | 1,724,160 | 89.0% | 91.8k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-tests** | 4 | 183,530 | 12,800 | 1,908,864 | 91.2% | 82.2k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **Total** | **20** | **1,283,075** | **125,496** | **42,043,392** | **97.0%** | **185.6k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 20 unavailable / 20 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 8 | 317,650 | 18,473 | 2,412,928 | 88.4% | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **gpt-6.1-sol** | 12 | 965,425 | 107,023 | 39,630,464 | 97.6% | Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations) |
| **Total** | **20** | **1,283,075** | **125,496** | **42,043,392** | **97.0%** | **Unavailable (0 observed, 0 partial, 20 unavailable / 20 invocations)** |

Capture observation finalized: 2026-11-11T19:44:50.252938638Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 20; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 20; reserved bytes: 17150085. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1cbdf-16db-77e2-8e29-6145af952823: current; last successful source read 2026-11-11T19:44:49.648373611Z; accepted source checkpoint 2026-11-11T18:30:34.712Z.
Owner codex/01a1cbda-fbe5-7ac1-825b-883fadf1a963: current; last successful source read 2026-11-11T19:44:49.664367036Z; accepted source checkpoint 2026-11-11T18:25:27.116Z.
Owner codex/01a1cbde-c6fc-71b0-9fbe-58af8b115261: current; last successful source read 2026-11-11T19:44:49.686022956Z; accepted source checkpoint 2026-11-11T18:30:45.633Z.
Owner codex/01a1cbd1-b939-7491-b7a3-c34c6ec104b9: current; last successful source read 2026-11-11T19:44:49.722038955Z; accepted source checkpoint 2026-11-11T18:27:37.954Z.
Owner codex/01a1cbda-307f-7a30-b331-c276151d5e29: current; last successful source read 2026-11-11T19:44:49.738658131Z; accepted source checkpoint 2026-11-11T18:24:40.122Z.
Owner codex/01a1cbcb-f6a3-71b3-9317-282cc89a7991: current; last successful source read 2026-11-11T19:44:49.816311237Z; accepted source checkpoint 2026-11-11T19:44:49.363Z.
Owner codex/01a1cbe0-8adf-7792-9aa7-e094a49d9d30: current; last successful source read 2026-11-11T19:44:49.831717482Z; accepted source checkpoint 2026-11-11T18:31:44.964Z.
Owner codex/01a1cbda-59d6-7061-98b9-59e99ceaea88: current; last successful source read 2026-11-11T19:44:49.850244599Z; accepted source checkpoint 2026-11-11T18:24:44.976Z.
Owner codex/01a1cbe1-7f5d-7721-a28d-abb1887ed272: current; last successful source read 2026-11-11T19:44:49.876275961Z; accepted source checkpoint 2026-11-11T18:33:41.154Z.
Owner codex/01a1cbcc-827d-75d2-afd7-b87155d38539: current; last successful source read 2026-11-11T19:44:49.89207898Z; accepted source checkpoint 2026-11-11T18:10:10.878Z.
Owner codex/01a1cbe0-c548-7782-94e2-d080ed9dc088: current; last successful source read 2026-11-11T19:44:49.907193652Z; accepted source checkpoint 2026-11-11T18:31:52.619Z.
Owner codex/01a1cbd8-33d4-7030-8e05-d1f4ba05fb4e: current; last successful source read 2026-11-11T19:44:49.934424925Z; accepted source checkpoint 2026-11-11T18:23:54.009Z.
Owner codex/01a1cbd2-2aa7-7800-82a9-4f06cb0aba42: current; last successful source read 2026-11-11T19:44:49.962383343Z; accepted source checkpoint 2026-11-11T18:28:51.095Z.
Owner codex/01a1cbd9-18c4-7ac3-bef2-94c584f37e1b: current; last successful source read 2026-11-11T19:44:49.990390648Z; accepted source checkpoint 2026-11-11T18:25:24.241Z.
Owner codex/01a1cbd7-1c32-7f90-8e6e-6e0989fb15a2: current; last successful source read 2026-11-11T19:44:50.017529419Z; accepted source checkpoint 2026-11-11T18:22:40.863Z.
Owner codex/01a1cbde-f432-7b62-ac66-67e50dfe68fb: current; last successful source read 2026-11-11T19:44:50.038141431Z; accepted source checkpoint 2026-11-11T18:30:58.718Z.
Additional owner freshness details remain in local capture diagnostics.
<!-- agentic-lens:metrics:end -->

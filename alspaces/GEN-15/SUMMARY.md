# GEN-15: Portable setup and worktree operation

## Request and baseline

- Source: user reports that adoption installs the operator but not setup itself, and ignored binaries disappear in worktrees; asks how to solve it.
- Allocated GEN-15 after checking existing ticket workspaces, the ticket index, and local/remote branch references (including GEN-14).
- Initial Git status: clean. Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Branch: `chore/GEN-15-worktree-installation`.

## Grounded findings

- Setup installs the operator manifest and three references, but does not install its own manifest and downloader.
- Setup installs the wrapper alongside native executables in the ignored operator `bin/` directory. A committed adopting repository therefore does not carry that entrypoint into another worktree.
- The wrapper can fetch a missing native binary only when the wrapper itself exists.
- The operator CLI reference treats an absent executable as requiring setup, but the installed repository may lack that setup skill.
- Downloader defaults and wrapper source fallback still assume setup/source-checkout paths in places; adoption must be verified without the AgenticLens source tree.

## Status and verification

User clarification incorporated into CONTRACT.md and iterations/01/PLAN.md: the adopting repository owner decides whether to track or ignore any installed files. Superseded the tracked-launcher and mandatory-ignored-binary proposal. The launcher remains outside binary storage; setup/repair must preserve Git policy and diagnose missing setup assets as well as binaries. Regression planning now covers owner-selected tracking policies and a worktree without inherited repair scripts. This clarification does not establish implementation approval.

The user explicitly approved the corrected contract and plan with "approved, follow with the implementation." Implementation, correction tests, independent re-review and integrated verification are complete. Final telemetry and PR publication follow below.

Dependency preparation: `make install` passed. npm reported existing dependency audit advisories; dependency versions and lockfile were not changed by this ticket. Runtime, review, and integrated verification evidence will be recorded below before delivery.

## Implementation

- Added complete setup/operator skill bundle installation, including setup's own installer, downloader and bundled wrapper. A selected release is recorded in RELEASE_VERSION and used by repair.
- Added an offline launcher outside bin storage. It preserves caller directory, arguments and exit status, and prints explicit binary repair or skill bootstrap guidance for missing files.
- Binary repair uses worktree-local operator/bin by default, supports explicit destinations/prebuilt binaries, retains existing executables unless replacement is requested, and preserves an existing executable on download failure.
- Removed automatic ignore-file creation. Neither installation path stages files, commits files, nor imposes Git policy. Existing consumer configuration remains untouched.
- Consolidated contributor downloads onto the bundled implementation and aligned setup, operator, README and manual guidance. Legacy wrappers retain their established behavior.

## Verification

- Candidate: d7a07eb40f1a66b98da7f6bd9252e5a8c434f1ad; main integration was already up to date.
- Nine installation boundary tests and six existing release tests passed. Tests use real disposable Git worktrees and controlled transport/prebuilt fixtures, with no silent skips.
- Restoring the baseline downloader in a disposable bundle makes the Git-policy preservation test fail on its unwanted ignore-file creation.
- Parent `make backend-check` passed, including Go vet, race tests and canonical frontend embedding preparation.
- Integrated `make` passed candidate d7a07eb: six release tests, nine installation tests, Go vet/race tests, frontend typechecking and 35 test files / 781 tests, canonical embedded assets and platform builds.
- First-pass review: Sol 6.1/high found two confirmed P2 recovery defects (non-executable retained files and missing release markers); Luna 6/medium found a stale approval sentence. Both runtime defects and the sentence were corrected. Both profiles approved corrected candidate 2d540946c01851face786f15894b1e00b2451b19. The retained findings and actor/model attribution are summarized below; detailed review working files were removed after approval at the user's request.
- Final integrated `make` passed candidate 2d54094: twelve installation tests, six release tests, Go vet/race suites, frontend typechecking/lint, 35 test files / 781 tests, canonical assets and Make platform builds. No skipped test stages.
- Actual freshly compiled Linux amd64 binary installed and launched successfully in a separate consumer path containing spaces, returning commit 2d54094. No config/ignore files were created. This local-build smoke does not establish availability of an official release asset.
- Host evidence is Linux amd64. macOS/Windows native execution and live release asset availability were not verified. Skill bootstrap requires a release containing the new installer; publication of this PR does not itself publish such a release.

## Completed review summary

Runtime review approved for `2d540946c01851face786f15894b1e00b2451b19` against baseline `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`. All confirmed findings are fixed; no unresolved candidates, refuted claims, incomplete runs or accepted debt remain in this review. Role for all actors below: reviewer. The parent verified each model/effort against actual Codex session metadata and turn context; these are observed identities, not merely requested settings.

| Finding | Discovering actor / model | Resolution |
| --- | --- | --- |
| GEN15-01, P2: repair falsely succeeded for an existing non-executable binary | SOL-01, `/root/review_sol`, gpt-6.1-sol / high | Explicit repair restores and checks execute permission without replacing bytes; failure is surfaced. Exact recommended-command regression and injected chmod-failure probe passed. |
| GEN15-02, P2: a worktree missing RELEASE_VERSION silently selected latest | SOL-01, `/root/review_sol`, gpt-6.1-sol / high | Remote repair requires an explicit release when its marker is absent/empty; offline supplied-prebuilt repair remains available. Real markerless-worktree and override-precedence probes passed. |
| GEN15-DOC1, low: stale pending-approval wording contradicted explicit approval | LUNA-01, `/root/review_luna`, gpt-6-luna / medium | Parent corrected the plan to reference the subsequent approval. |

Both runtime regressions failed against initial candidate d7a07eb in disposable bundles and passed after correction. Verification credit is separate from discovery: SOL-02 (`/root/review_sol_r2`, gpt-6.1-sol/high) and LUNA-02 (`/root/review_luna_r2`, gpt-6-luna/medium) verified the fixes and approved the corrected candidate. No shared discoveries were reported.

| Run / actor | Model / effort | Phase / revision | Outcome |
| --- | --- | --- | --- |
| SOL-01 / `/root/review_sol` | gpt-6.1-sol / high | Discovery / d7a07eb | Changes requested; two runtime findings |
| LUNA-01 / `/root/review_luna` | gpt-6-luna / medium | Discovery / d7a07eb | Scoped approval; one plan correction |
| SOL-02 / `/root/review_sol_r2` | gpt-6.1-sol / high | Re-review / 2d54094 | Approved; fixes verified, no new finding |
| LUNA-02 / `/root/review_luna_r2` | gpt-6-luna / medium | Re-review / 2d54094 | Approved; fixes verified, no new finding |

Verification and platform/release limits are recorded above. Detailed reports and run/finding JSON were working evidence and were deleted only after corrections, consolidated approval and this summary. Regression tests, plans, test evidence and session telemetry remain. No model quality ranking or per-run cost is inferred.

## Iteration 02: completed-review retention

Source: the user requested keeping a summary of findings with the model and actor that found each, then deleting review files once corrections are addressed and review is approved. This supersedes permanent detailed-review retention; it preserves active-review evidence, approval gates, discovery attribution, verification and known limits. Updated AGENTS and canonical review/development guidance; removed seven completed GEN-15 review files after consolidation. Other tickets' historical audits are untouched. This is authorized documentation/artifact maintenance with no runtime changes or new reviewer dispatch. The previously passed application checks remain valid for the unchanged runtime; documentation checks and current delivery verification are recorded in iterations/02/PLAN.md.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 6
- **Active Agent Duration:** 39m 29s (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 161.4k tokens (81% of 200k healthy ceiling, 0 compactions) | Subagents peak 110.7k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.3 avg turns/run (peak: 2 turns) | Planner 1.0, Coder 2.0, Tester 2.0, Reviewer 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 20,255 reasoning tokens (19.0% of generated output)
- **Tool Health & Self-Correction:** 197 tool calls (4 self-corrections, 98.0% success rate)

### Anomalies & Health Flags
- **Status:** 4 execution anomalies detected (1 CRITICAL, 3 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+72.9k tokens) — Context tokens surged by 72.9k to 72.9k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.1k tokens) — Context tokens reached 160.1k, breaching the 160k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (3 cycles) — Subagent roles 'coder' and 'tester' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 48,363 | 3,575 | 376,704 | 88.6% | 58.1k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 98,084 | 13,198 | 2,165,120 | 95.7% | 107.1k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 47,330 | 13,823 | 2,324,224 | 98.0% | 110.7k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer** | 4 | 208,121 | 19,471 | 2,195,072 | 91.3% | 77.1k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **rootagent** | 1 | 275,600 | 36,502 | 11,031,808 | 97.6% | 161.4k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **8** | **677,498** | **86,569** | **18,092,928** | **96.4%** | **161.4k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 2 | 83,320 | 7,998 | 854,272 | 91.1% | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **gpt-6.1-sol** | 6 | 594,178 | 78,571 | 17,238,656 | 96.7% | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **Total** | **8** | **677,498** | **86,569** | **18,092,928** | **96.4%** | **Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations)** |

## Final telemetry verification

Generated with explicit Codex source, root session `01a19c78-4515-7170-93d3-178ecba3b8d1`, literal nonce `54cd975390ae9c09`, and conventional type `chore`. CLI discovered eight actual rollouts. Compared ledger session IDs and parent links against real source session metadata: root plus all seven children (planner, coder, tester, two discovery reviewers and two re-reviewers) are present, each labeled Codex with non-empty invocation facts. No unavailable child or source error was reported. No pre-existing GEN-15 ledger or other-harness sessions existed to preserve; no other-harness facts were added. Raw rollouts remain outside Git. This final extraction precedes the final telemetry commit; post-extraction verification/commit/publication calls are outside the snapshot.

Iteration 02 telemetry refresh: explicit Codex source, root `01a19c78-4515-7170-93d3-178ecba3b8d1`, literal nonce `b44ae9aa742a0747`, type chore. Discovered the same eight real rollouts; compared refreshed ledger IDs, sources and parent links with the pre-refresh ledger and preserved all eight sessions. All seven children retain their root parent and non-empty facts. No other-harness sessions existed; no source errors reported. Detailed review files remain deleted, while the completed-review summary above is preserved. Post-extraction verification/commit/push calls are outside this snapshot.

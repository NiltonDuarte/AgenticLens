# GEN-25 iteration 01: Remove Windows support

Baseline: `4e5c458322a2cddfdb5911b26a4569ea75f06a09`; initial working tree clean. The placeholder branch was renamed to `chore/GEN-25-remove-windows-support`. The configured workspace root is `alspaces`; GEN-25 was unused in the workspace, ticket index and inspected Git refs. This is an ad hoc chore with no roadmap specification. Scope comes from the user's request to remove all Windows support while continuing macOS and Linux. [Contract](../../CONTRACT.md) approved by the user message “Approved” on 2026-11-11; implementation may proceed.

## Ownership and work

1. Planner owns this plan and the contract. Parent owns integration, approval, telemetry and delivery artifacts. After approval, coder owns affected application, build, installer and active documentation paths; tester owns additional adversarial tests. Agents share the checkout and preserve unrelated edits.
2. In `Makefile`, restrict build/release targets to the four supported combinations; remove .exe/zip branches and repair checksum enumeration. Trace `.github/workflows/` release consumers and `scripts/tests/test_release_revision.py` without changing historical-tag semantics or restoring Windows CI.
3. In `backend/internal/adapters/{manager,config,capture}` and affected CLI tests, remove Windows files and branches, PE admission and platform acceptance. Simplify platform dispatch, hook command rendering, plugin transport, watchdog and filesystem operations around maintained hosts. Preserve user-config ownership checks, including any legacy field inspection necessary to avoid overwriting user edits.
4. Remove `download.ps1`; restrict setup `download.sh` selections and previews. Update `docs/SETUP_MANUAL.md`, `CONTRIBUTING.md`, setup/operator skills and affected active guidance. Audit remaining Windows matches by purpose; retain history and portable-data behavior.
5. Remove Windows-only Go/Python execution suites. Adapt mixed installation, capture, plugin and distribution fixtures, retaining supported argv/stdin, deadlines, rollback and ownership assertions. Add only missing rejection and four-platform distribution cases.

## Verification and review

Select configuration/project portability, installer isolation, build/release/downloader and cleanup/compatibility cases from [regression prevention](../../../../docs/REGRESSION_PREVENTION.md). Exercise unsupported selections with no network/install writes, tar.gz-only packaging/checksums, supported platform names and native-format validation. Use isolated fixtures, including nondefault roots and quoted paths. No persistence change is planned; retain existing portable-history coverage and investigate any affected path handling.

Run focused Go and Python boundary suites, then integrated `make` after merging main. Verify fresh and repeated asset builds in disposable paths, current embedded assets and standalone host manager/engine execution. Cross-compile all four targets; distinguish that evidence from native macOS execution if unavailable.

Before review, parent freezes a revision/fingerprint and records unique reports for behavior/integration, data/failure handling, architecture and test effectiveness. Each focus uses both configured model profiles in capacity-limited waves; behavior review includes release/download integrity. Parent consolidates findings and required integrated results before telemetry and PR delivery.

## Review roster

All discovery runs inspect the same frozen candidate against baseline `4e5c4583`. Reports and detailed attribution records live in private `/tmp/GEN-25-review-HxxBR4/` during investigation; the parent owns consolidation. Each focus has independent Sol (`gpt-6.1-sol`, high) and Luna (`gpt-6-luna`, medium) runs, with fresh contexts and capacity-limited waves.

| Focus / report stem (one `-sol.md` and `-luna.md` each) | Risk question / inspected boundaries |
| --- | --- |
| behavior | Do release builds, downloader selections, manager dispatch and capture integrations preserve the four maintained targets and reject unsupported requests? Includes build/release integrity specialist questions within this focus. |
| data | Do atomic replacement, locks, config ownership and portable recorded paths remain safe after Windows branch removal? |
| architecture | Are maintained platform constraints, manager/engine separation and active support guidance consistent without residual Windows execution paths? |
| tests | Do removed suites contain supported-platform assertions still needed, and do remaining/new fixtures execute real acceptance boundaries with meaningful oracles? |

Parent review before freeze identified and coder corrected ignored unsupported `--include-host` when combined with `--all-platforms`; tester added that adversarial case. This is an approval-preserving correction within the contract.

## Completed verification

The corrected reviewed tree `70fc7848f062903fb2a2c199234ce59250c946e4` became commit `e67991f3`; fetched main was already current. Integrated `make` passed with Go vet/race integrations, TypeScript checks, frontend unit/integration lanes and all eight four-target binaries. Real disposable release packaging/checksums and standalone host identities passed; disposable absent/repeated assets and before/after hook/checksum/upload regressions passed. No product changes followed those checks.

Initial Sol behavior/architecture findings and coder-discovered hook compatibility were fixed and re-reviewed under all eight cells. Parent retained test-effectiveness Luna's dissent about an extra execution sentinel and refuted materiality through the complete no-evaluation source path; no unresolved finding remains. Parent consolidated APPROVED. Native macOS/live harness callback/hosted upload execution remain explicit limits. See SUMMARY.md and retained REVIEW_SUMMARY.json for attribution/evidence.

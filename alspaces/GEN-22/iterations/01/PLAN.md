# GEN-22 iteration 01: Shared and CLI user attribution

Status: Proposed; planning only until explicit user approval.

## Baseline and evidence

- Source request: “Add user tag, both shared config and from cli”. No earlier approved implementation boundaries exist for this ticket.
- Baseline revision: 57a3503e3024f3993ce9391fee604aad0bcc44d5. Initial working tree was clean. Ticket branch: feat/GEN-22-user-tag. Configured workspace root: alspaces.
- ProjectConfig and DefaultProjectConfig in backend/internal/domain/config.go currently have no user_tag. rawProjectConfigDTO and mergeConfig in backend/internal/adapters/config/loader.go explicitly map recognized shared settings. The local overlay is separately allowlisted; do not add user_tag there.
- GenerateConfigYAML in backend/internal/adapters/config/writer.go manually emits configuration and must be updated alongside the loader, rather than relying on struct tags alone.
- runGenerate in backend/cmd/agentic-lens/main.go declares all generation options through flag.FlagSet and has no user-tag option. It obtains git.GetGitUser, resolves Codex users with resolveCodexSessionUsers, extracts OpenCode sessions with currentUser and known-session users, and combines captureResult.RefreshedSessions with extracted freshSessions before mergeSessions.
- backend/internal/adapters/git/user.go defines ResolveSessionUser: known extracted identity, known existing identity, Git identity, then unknown. Keep this fallback when no override is supplied.
- backend/internal/adapters/capture/reconcile.go assigns user attribution to extracted Codex sessions. Applying the resolved override consistently to the complete freshSessions set in generation avoids missing registered capture publication or one harness. Do not apply it to the entire merged historical ledger.
- SessionFact.User in backend/internal/domain/models.go is persisted as user. backend/internal/adapters/fact_ledger_dto.go requires a nonblank trimmed user for current ledgers. backend/internal/adapters/http/workspace_dto.go already projects user, and web/src/components/ThreadViewer.tsx already renders the User label.
- The installed agentic-lens wrapper cannot display native help because its pinned v2.0.0-rc.111+99fbcd7 Linux binary is absent. CLI flags were grounded from runGenerate source. Do not claim a successful CLI help or runtime test. Use the canonical build prerequisites to obtain source-built help during implementation.

## Approval decision

Approve CONTRACT.md's interpretation that user_tag supplies the existing user attribution label, with precedence CLI then shared configuration then existing resolution. The tag is not a new additive persisted property. If additive metadata was intended, revise acceptance boundaries before code is dispatched. The initial slice targets generate and automatic capture publication through generate; it does not add backfill flags or live attribution controls.

## Implementation ownership and sequence

1. Coder owns application/configuration/documentation paths for this iteration after approval: backend/internal/domain/config.go; backend/internal/adapters/config/loader.go and writer.go; backend/cmd/agentic-lens/main.go or a focused typed attribution helper beside it; README.md and the relevant shared-configuration explanation in docs/SETUP_MANUAL.md. Document the optional key in the repository configuration without selecting a personal identity. Extend other active templates only when grounding reveals they separately enumerate supported configuration keys.
2. Add the optional typed setting, preserve load/write round trips, and safely encode YAML-sensitive values. Keep default absence backward compatible. Add generate --user-tag and resolve it once after configuration loading. Apply it only to successfully refreshed owners after all fresh source/capture snapshots have been collected and before persistence/merge validation relies on those values. Ensure capture publication bookkeeping records the same final owner attribution; inspect both staging and committed publication calls.
3. Tester owns relevant existing test files, with no overlapping writable paths while coder runs. Extend config loader/writer coverage and generation integration fixtures first. Add narrowly scoped test files only if the actual boundary has no coherent existing owner. Inspect main_codex_user_test.go, main_codex_test.go, main_test.go, owned_refresh_regression_test.go, capture_history_test.go, config/loader_test.go and config/writer_test.go before allocating cases. Existing backfill/codex_user_test.go and portable refresh tests provide compatibility evidence; do not change backfill attribution semantics.
4. Verify the actual persisted value and existing HTTP thread/user projection, not only a pure precedence helper. No frontend source change is anticipated; use existing API contracts and the current User rendering. If implementation reveals that a new wire field or UI contract is necessary, return to scope review rather than broadening silently.
5. Parent assigns independent focused review ownership after freezing the implementation revision. Cover behavior, data preservation/failure paths, architecture and test effectiveness under the canonical parallel review protocol, with explicit unique report paths and configured model/effort evidence. All agents share the checkout and must preserve others' edits. Parent consolidates findings and owns fixes routing, integration and publication.

## Targeted regression cases and verification

- Config owner: omission and blank input leave fallback behavior intact; a non-default tag containing spaces, colon and hash survives write/load without YAML reinterpretation. Existing required configuration validation and local override restrictions remain intact.
- Generation owner: shared-only tag and conflicting CLI tag reach saved session user for both harnesses; absence retains existing attribution precedence. Exercise a root and refreshed child.
- Capture owner: shared default and manual override reach refreshed capture-owned sessions at the same boundary as nonce/session extraction. Validate recorded publication and ledger attribution agree.
- Portable ledger owner: preserve unavailable same-harness owner and other-harness owner unchanged while refreshing a local owner; reload without source enrichment and verify users remain readable. Repeat unchanged generation without duplicating owners or altering invocation multiplicity. Reuse existing focused fixtures for discovery/partial-source failures rather than inventing source completeness rules.
- Projection owner: use an isolated generated/reloaded ledger to assert the existing HTTP thread/session user projection sees the configured attribution. No UI feature or schema migration is needed for this interpretation.
- Follow canonical asset prerequisites before Go execution. During development run affected tests, then parent runs required integrated make after upstream integration. Record actual commands, tested revision/diff, passed/failed/skipped stages and environment limits in SUMMARY.md. A failed installed wrapper is not a runtime validation result.

## Completion boundary

No implementation, test execution or review approval is claimed by these planning artifacts. After approved implementation/review, parent updates task artifacts, performs integration and the full gate, generates final Codex telemetry with a literal nonce and explicit feat classification, validates source/root/child completeness, commits and publishes the PR. Preserve explicit incompleteness if original telemetry sources or tools are unavailable.

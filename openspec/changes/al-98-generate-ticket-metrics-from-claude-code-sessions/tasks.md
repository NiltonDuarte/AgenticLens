# AL-98: Execution tasks

[design.md](design.md) holds the approved scope, the user's decisions and the approval of 2026-11-24. [specs/](specs/claude-code-session-telemetry/spec.md) holds the observable acceptance. Every fixture must be self-contained and labeled synthetic. No test may read `~/.claude`, a live ticket workspace or a change path.

## Planning

- [x] 0.1 Ground the contract in the existing adapters, configuration, refresh paths and native records, and draft the proposal, spec, design and tasks. Validated with `scripts/openspec.sh validate al-98-generate-ticket-metrics-from-claude-code-sessions --strict`.
- [x] 0.2 Fold the user's five decisions and the approval into the proposal, spec, design and tasks, and record the narrowed scope under Steering in design.md. Strict validation was re-run.

## 1. Source selection and configuration (coder)

- [x] 1.1 Add `domain.SourceClaude` and `TelemetrySourcePaths.ClaudeHome`. Extend `ResolveSource`, add `ResolveClaudeHome`, and make the loader accept `telemetry.default_source: claude` and `telemetry.sources.claude.home`. Update the supported-value messages.
- [x] 1.2 Extend `InitialLocalYAMLWithSettings`, the allowed `--local-settings` keys and the `init` flags with `--claude-home`. Leave capture identity and installation validation unchanged.
- [x] 1.3 Add `--claude-home` to `generate` and `backfill`, reject a blank supplied value, update the usage text and add `claude` to the source lists. In backfill, `all` covers OpenCode, Codex and Claude Code, and `both` keeps meaning OpenCode plus Codex. Update the help text and the error messages to describe all three values.

## 2. Claude Code adapter (coder, `backend/internal/adapters/claude/`)

- [x] 2.1 `store_discovery.go`: enumerate root and subagent files under `<store>/projects`, ignoring `tool-results` and other non-transcript files. Select by literal nonce, by exact root ID, or by an agent ID resolved to its root. Report ambiguity and absence, honor cancellation, and accept an optional session and byte budget whose exhaustion is a typed error.
- [x] 2.2 `transcript_jsonl.go`: parse conversation, system and attachment records and ignore the other native record types. Treat a malformed line, a record with no type or a foreign `sessionId` as an incomplete owner. Exclude records marked `forkedFrom`. Report a root that contains sidechain conversation records as an unsupported layout.
- [x] 2.3 `subagent_proof.go`: link a child only through its meta `toolUseId`, the matching `Agent` tool_use in the same tree and agreeing `agentId` and `sessionId` values, including nested children. Return diagnostics for unproven, unreadable and missing children without inventing a parent.
- [x] 2.4 `session_facts.go`:
  - Deduplicate responses by `requestId`, falling back to `message.id`. Reject missing or conflicting usage.
  - Build one invocation row per served model, mapping tokens as in the design.
  - Record unavailable cost, tool rows that count `is_error` results as errors, and compaction counts.
  - Bound turns, measure AL-64 durations including provisional time, and count steering from non-system root prompts.
  - Produce successful `Skill` activations and `Agent` delegations, transient events, complete-owner IDs and checkpoint evidence under `GenerationPolicy`.

## 3. Generate wiring (coder)

- [x] 3.1 `cmd/agentic-lens/claude_generation.go`: resolve and stat the store, discover and parse the selected tree, and fail before publication on discovery errors or an incomplete root. Print child diagnostics. Validate the root's `cwd` with `capture.ValidateProject`, and check for collisions with `capture.ValidateSessionMerge`. Resolve users, set start and completion times, and merge through `mergeSessions` and `retainUnrefreshedGenerationOwners`.
- [x] 3.2 Dispatch from `runGeneration` when `activeSource == "claude"` and leave the Codex, OpenCode and `--capture-auto` branches unchanged.

## 4. Backfill recovery (coder)

- [x] 4.1 Add `BackfillOptions.ClaudeHome` with resolution and the directory check. Add the Claude Code bucket and its protected sets to `engine.go`, and include it in the completeness calculation without changing Codex or OpenCode outcomes. Make `all` select Claude Code too, and fail before any read or publication when its store is missing or unreadable. Leave `both` unchanged.
- [x] 4.2 Add `recoverClaude` to `recovery.go`, mirroring `recoverCodex`: exact roots, retention when parent links change, conflicts with other harnesses, newly proven children and diagnostics. Charge reads to `sourceReadBudget`.

## 5. Ledger readers and dashboard (coder)

- [x] 5.1 Accept `claude` in the canonical validation in `fact_ledger_dto.go`.
- [x] 5.2 Accept `'claude'` in `web/src/types/telemetry.ts` and `web/src/api/validation.ts`. Run the canonical frontend build and confirm that the assets in `backend/internal/adapters/web/dist` are current, following the tracking policy for generated files.

## 6. Guidance and capability matrix (coder)

- [x] 6.1 Rewrite AGENTS.md rule 7: a Claude Code session uses `--source claude` and passes `--claude-home` only for a nonstandard store. Update the source statements in `docs/OPENSPEC_WORKFLOW.md`, and add the Claude Code store to the local-source sentence in TECHNICAL_GUIDELINES.md.
- [x] 6.2 Update `docs/SETUP_MANUAL.md`, `.agents/skills/agentic-lens-operator/references/{cli.md,configuration.md}`, `.agents/skills/agentic-lens-setup/SKILL.md`, `docs/OWNED_OCCURRENCES.md` (Claude Code classification), `docs/COST_SEMANTICS.md` (unavailable cost, `cost-state` excluded) and the README source list. In the operator reference, describe `--source claude`, `all` (OpenCode, Codex and Claude Code, failing without a readable Claude Code store) and `both` (OpenCode plus Codex). Explain that `CLAUDE_CONFIG_DIR` is not read, cover the `cleanupPeriodDays` retention caveat and state that automatic capture is pending in AL-99. Search active guidance for "codex or opencode" wording that conflicts with this change.
- [ ] 6.3 Append a Claude Code column or section to `openspec/changes/al-63-verify-source-capabilities-across-harnesses/CAPABILITIES.md`. Record the tested version, the evidence ID behind each cell and the cells that are unsupported or not run: hydration, live, cost, entrypoints that were not verified, and compaction or forks until observed.

## 7. Verification (tester; native verifier for 7.3)

- [ ] 7.1 Adapter tests in `backend/internal/adapters/claude/*_test.go` with labeled synthetic fixtures:
  - duplicate content blocks and fork copies
  - missing and conflicting usage
  - one invocation row per served model
  - proven, nested, unproven, unreadable and missing children
  - sidechains inside a root transcript
  - nonce and ID ambiguity across folders, and identity taken from `cwd`
  - budget exhaustion and cancellation
  - successful and failed Skill calls
  - Agent delegations with child references
  - completed, open, superseded and task-notification turns
  - compaction counts and unavailable cost
- [ ] 7.2 Extend the existing tables rather than duplicating them:
  - `config/source_test.go`, `loader_test.go` and the local-setup writer tests: precedence, non-default paths, no environment fallback
  - `cmd/agentic-lens/main_codex_test.go` `TestGenerateSelectedSourceDoesNotFallback`, plus a new `main_claude_test.go`:
    - refresh replaces and an unchanged repeat is equal
    - a mixed-harness ledger with an other-machine owner and an unavailable child
    - a restored complete snapshot with fewer invocations
    - the ownership conflict
    - `--capture-auto` with a default `claude` source
  - Backfill CLI and engine: `--source all` with a missing or unreadable Claude Code store fails before publication and leaves the ledgers unchanged; `--source all` with all three stores refreshes each harness; `--source both` behaves as before and preserves Claude Code owners
  - Backfill: a new `claude_recovery_test.go` covering exact refresh, parent-link retention, budget exhaustion and the protection of Codex and OpenCode owners
  - The `fact_ledger_dto` decode tests and the web validation tests: accept a `claude` session and reject unknown values
- [ ] 7.3 Native run: build the candidate and generate a disposable ledger outside Git from this ticket's root `b918d950-b1d1-42d6-8d61-a688c68333da` and its native subagents. Compare it with an independent oracle outside Git that sums unique responses per owner. Repeat the run unchanged and compare. Record the version, IDs, totals, diagnostics and limits under `evidence/` without transcript content.
- [ ] 7.4 Native probe (native verifier, approved 2026-11-24): run a disposable `claude -p` session with a native subagent and a `/compact` in an isolated fixture repository. Generate a disposable ledger outside Git and record the `sdk-cli` entrypoint, whether compaction kept the same file, the `compact_boundary` count, the IDs and the limits under `evidence/`. The transcripts stay outside Git.

## Integration and delivery

- [ ] 8.1 Run focused Go and web checks, then commit the candidate. Merge main and reconcile AL-94 and AL-78 changes in the backfill engine and `main.go` if either has landed. Run the integrated `make` and record the tested revision, commands and skipped stages.
- [ ] 8.2 Run independent focused reviews (data, behavior, architecture, tests) on a frozen candidate, then a consolidated review. Route fixes to the coder and missing tests to the tester, re-review the affected scope, and keep compact findings in `evidence/REVIEW_SUMMARY.json`.
- [ ] 8.3 Write `delivery.md` and the release notes. Both must state that engines older than this change, including the pinned `v3.0.0-rc.156`, reject ledgers containing `source: "claude"`, and that backfill `--source all` now requires a Claude Code store. As the final step, generate telemetry with the built candidate and a literal nonce: `--source claude --ticket-id AL-98 --ticket-type feat --json-output <CHANGE_DIR>/agentic_metrics.json --append <CHANGE_DIR>/delivery.md`. Verify the root, children and preserved owners, then commit, push and create or update the PR.

## Workflow follow-up

Archive only after verified delivery. Reconcile AL-98 and AL-99 roadmap status separately.

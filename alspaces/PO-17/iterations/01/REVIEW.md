# PO-17 iteration 01 independent product review

Verdict: review recommendation prepared; cost design and AL-55 implementation remain unapproved. Baseline bdef2d0ef98e37fe259921bc30ba2daefbfda717. Source inspection only, not a reproduced runtime failure or implementation verification.

## AL-57 recommendation and risks

Fix the semantics before selecting nullable fields. Recommend source evidence/availability at facts and known subtotal plus explicit coverage for aggregates, keeping complete totals unavailable when incomplete. This recommendation is not accepted. User has explicitly deferred all cost changes pending PO review and selected AL-only development; AL2 remains queued.

Evidence:

- codex/rollout.go:817 emits CostUSD:0 without billed-cost evidence. AL-56 CONTRACT.md, RESEARCH.md, DATA_PROVENANCE.md distinguish rollout token evidence from unavailable cost; historical AL-56 ledgers are not rewritten by this review.
- domain/models.go:28 uses float64; aggregations.go:91 and :861 sum numeric values, with cost-based ordering and progression elsewhere in the same file. Existing aggregations_test.go assumes numeric costs and empty totals.
- opencode/schema_v1_session_extraction.go:315 and schema_v2_session_extraction.go:197 detect optional source cost but add into numeric accumulators. Source availability must be traced per record, not assumed by harness. Source-reported amounts are not verified invoice billing.
- adapters/schemas.go:46, domain/types.ts:58 and subsequent aggregate types require numeric cost. web/src/types/telemetry.ts is not the owning invocation cost contract; the AL2-6 suggested path alone is insufficient.
- web/src/domain/formatters.ts:91 maps nonfinite/nonpositive values to $0.00; progression.ts:187/:237 performs arithmetic; ticket_scatter.ts:67 returns numeric cost. Tests exist in formatters.test.ts, progression.test.ts and ticket_scatter.test.ts.
- adapters/formatters.go:326 onward and cmd/agentic-lens/main.go:1327 format numerical summaries. Null-only adapter changes would leave downstream arithmetic and presentation invalid or misleading.

AL-57 omits a decided mixed-total contract and explicit legacy uncertainty state. AL2-6 conflicts by using zero as estimate eligibility; preserve known true zero and historical numbers, never substitute prices without provenance. AL2-7's modeled hybrid must not be described as actual billed spend or double-count facts. Broad null propagation can discard useful measured subtotals; broad numeric compatibility without status perpetuates false completeness. Either approach needs explicit complete/partial/unverified semantics, migration policy, zero handling and sorts/charts/CLI/API acceptance. No current prices fetched.

## AL-55 reconciliation

Reviewed current artifacts read-only in the separate developer worktree, including CONTRACT.md and iterations/01/REVIEW.md, after planning commit 1122cac. That commit is planning evidence, not implementation or user approval; current working proposal can include later edits. PO owns none of those files.

Evidence agrees with developer findings:

- constants.go:9 already sets schema v4; migration/engine.go registers StepV3ToV4UserIterations. Preserve AL-56's explicit ordinary-ingestion boundary and source-recovery pending state.
- models.go:169/170 and web/src/domain/types.ts:132/133 keep optional arrays. AL-51 done specification records additive intent; its historical v3 wording is not today's runtime version.
- backfill/engine.go:302 skip guard omits RecomputeDB. engine_test.go:156 named explicit-recompute test includes completed Migration metadata, so it misses canonical migration=nil.
- engine.go:432/434 replaces whole arrays from freshSessions before :435 onward preserves unresolved sessions. Ownerless arrays cannot safely be split or merged by guessed row identity.
- engine.go:550 sets dbSessionFound for a successful Codex tree; :578 onward uses the shared boolean to clear source_backfill. Root-count comparison near :429 does not prove descendant coverage.
- main.go:609/628 describes Force broadly. Explicit recompute semantics need aligned help and actual snapshot refresh, not merely bypassing the initial skip.
- migration_runner.go:164 implements optional backups; engine.go:121 passes Backup:false internally; :640 atomically renames backfill output. Do not conflate atomicity and rollback copies.

AL-55_ACCEPTANCE_PROPOSAL.md contains five precise user choices, alternatives, acceptance evidence, and remaining objections. Conservative retention plus independent complete-tree refresh can proceed separately from deferred cost work after explicit approval. No schema/provenance extension is justified for this minimal outcome. Snapshot age differences need explicit acceptance; unknown owner history must not be treated as complete empty evidence.

## Decisions and workflow evidence

Initial requests authorize independent product analysis only in PO-17. Manager routes questions and shared-path release. User choices: AL-only dispatch, AL2 queued, defer cost changes. Temporary stop was honored with edits halted and manager notified; user later clarified isolated-worktree concern and authorized resumption in this assigned worktree only. No checkout or thread created by this session. Existing PO proposal subtotals and graph remain unaccepted.

Applicable regression checklist: evidence vs inference, baseline persisted facts, exact identity and ownership, complete vs partial recovery, optional enrichment failure, migration shape vs source recovery, configured workspace root, test proportionality, verification claims. Technical rules and PO skill read; no code, tests, runtime scripts, adapters, developer contracts or shared plans modified.

Verification: bounded reads of actual source/test owners and existing specs; no application tests added or run. Full make/integration and final telemetry remain deferred to accepted documentation PR delivery. Own independent Codex root must be resolved by literal nonce at final generation, with no manager/AL-55 sessions imported. No native child agents were launched; no expected PO-17 children. Raw transcripts stay outside Git.

Review snapshot checks passed: git diff --check and six Markdown final-newline/no-emoji checks. A local planning-only commit records these proposals before manager acceptance; it neither freezes developer scope nor publishes a PR. Full make and final own telemetry are still pending the accepted documentation delivery.

Partial acceptance update (explicit user question-tool result): choice 2 permits independently complete OpenCode session-tree refresh while retaining prior whole skill/task collections, with explicit CLI diagnostics and the mixed-age limitation. Other AL-55 choices are unfrozen: user selected "Revise before implementation" without identifying a change. Manager has requested freeform clarification; no revision or implementation approval is inferred. Costs remain deferred, AL2 queued, and shared-plan/developer-contract edits unauthorized. This decision record refines the current review iteration; it does not start a new implementation plan.

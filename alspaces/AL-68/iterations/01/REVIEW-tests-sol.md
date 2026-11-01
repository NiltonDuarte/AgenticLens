# AL-68 independent test-effectiveness review

Outcome: **CHANGES_REQUESTED**. Run `tests-sol-01`, discovery phase. Requested execution: `codex/gpt-6.1-sol/high`; the tool interface does not independently expose runtime model/effort metadata, so those settings are requested identity, not directly observed identity.

Baseline HEAD: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`. Reviewed candidate: `c7dade3313a1d65a4269d595c782750655be8a044c020dc1f3ac502464cf325a`. Independently recomputed the CANDIDATE.json fingerprint from sorted tracked and nonignored untracked regular files outside alspaces, hashing path/NUL/content/NUL; it matches. Initial working tree contained the parent's AL-68 modifications and other reviewer artifacts. Only this assigned report was written; product/tests remained read-only. No sibling reports/personas or consolidated findings were loaded.

Applied the canonical agentic-lens-review procedure, reviewer-tests persona, shared review context, contract, approved plan including actual agent interaction/no guessing, ticket specification, technical guidelines and applicable regression checklist. Scope: changed CLI/source integration and all new capture domain, store, hook, reconciliation, installer, plugin and distribution tests, plus relevant unchanged session merge and extraction dependencies. Read raw TEST_PLAN, TEST_REPORT, SMOKE, SETUP_IMPLEMENTATION and VERIFICATION evidence.

## Blocking finding

### TS-1: Missing-source preservation test permits an older registry snapshot to erase newer published facts

Severity: high. Locations: `backend/cmd/agentic-lens/main.go:312` and `:333`; `backend/internal/adapters/capture/reconcile.go:255`. Missing oracle: `backend/cmd/agentic-lens/capture_test.go:274` through the final missing-source assertion.

Trigger: the capture registry has accepted an owner snapshot with 10 input tokens; a later explicit supported generation publishes 30 input tokens for the same owner without refreshing that registry; the source then becomes unavailable. A subsequent registered generation returns the registry's retained 10-token snapshot as fresh data and unconditionally replaces the newer canonical ledger session. It exits successfully while the capture report says the source failed. Contract section 6 requires preservation of newer accepted snapshots and persisted facts on source access failure.

Executed a disposable native CLI probe in `/tmp/al68-tests-sol-4avy2ibx/snapshot` using `bin/agentic-lens-linux-amd64` SHA256 `5ca77acea605841fccd518e3f854d52be4ae52666af9b6c18b863677c39b20e7`:

1. Wrote a self-contained Codex source with 10 input tokens; ran actual capture receive/request/bind and registered generate, establishing registry and ledger snapshots of 10.
2. Appended a later valid 20-token source record. Ran actual generate with `--session-id root --codex-home <fixture store>` and an absent `--capture-dir`, publishing 30 input tokens through the retained explicit selection path.
3. Removed the source file. Ran generate with the populated capture registry and no selectors. It exited zero; published input tokens fell from 30 to 10, and report registration state was `failed`.

Source trace confirms the cause: failed extraction retains the registry snapshot; Reconcile includes that snapshot in `result.Sessions` regardless of failure; the command feeds it to the unconditional ID replacement in existing `mergeSessions`. Registry-to-registry precedence guards do not compare against already published ledger facts.

The existing CLI test only removes a source after generating its baseline from the same unchanged registry. Thus its `DeepEqual(first.Sessions, missing.Sessions)` assertion passes for this defective implementation. Extend this coherent integration test with a newer pre-existing canonical owner and stale/failed registered enrichment; require that the newer owner, other-harness facts and AL-62 rows survive. Product reconciliation/publication must preserve accepted ledger facts when the retained registry data cannot establish a newer successful refresh. This report requests the regression and product correction; it does not prescribe an unapproved schema change.

## Investigated hypotheses and dispositions

| Hypothesis | Evidence and disposition |
| --- | --- |
| A guessed ticket or another session's request binds an independent owner | Production parser drops unknown ticket/parent/model payload fields. Store tests reject missing, wrong-session and wrong-store challenge tokens and conflicting tickets; callback test checks actionable feedback against the actual durable token. Meaningful fixture protection confirmed; actual agent origin is not established by these fixtures. |
| Subagent callback parent session ID is mistaken for child identity | Callback test asserts parent-only registration and warning. Queued child positive integration verifies child metadata and source-backed parent inheritance. Unavailable/mismatched child source runtime behavior remains less exercised; no positive identity claim inferred from agent_id. |
| Cumulative refresh adds totals, regresses an accepted registry snapshot, or loses writes | Reconcile tests inspect unchanged/resumed numerical totals, truncated/missing/partial source preservation and state; domain tests compare parsed timestamp precedence. SQLite tests use independent connections, concurrent read-modify-write transactions, rollback and a real process exit during an uncommitted transaction. These protect real persistence boundaries. Canonical-versus-registry precedence remains defective as TS-1 demonstrates. |
| Bounded reconciliation starves later owners or charges one source twice | Exact-source byte budget and persisted successive session-budget passes verify the intended path and final snapshots. No generic timing or coverage threshold substituted for an oracle. |
| A coexisting OpenCode family or matching title supplies the wrong owner | Both-family SQLite fixtures reach actual reconciliation and assert exact model/token facts; missing and duplicate exact IDs preserve previous facts and do not report current. Useful coverage confirmed. |
| Installer failure after writes leaves active hooks or damages unrelated content | New independent fault probe passed, described below. Existing tests also protect dry run, reinstall idempotence, user hooks, edited owned hooks/plugins, missing handler preflight and removal. No compatible upgrade or installer concurrency claim follows merely from reinstall coverage. |
| A digest field or nonempty cutoff alone proves safe publication | Happy-path CLI test recomputes the digest from actual ledger bytes and checks report/summary redaction, which is useful. It does not inject partial publication/crash recovery or compare newer canonical facts; the latter defect is TS-1. No additional publication defect is asserted from missing coverage alone. |

## Additional executed verification

Installer mid-write rollback probe used `/tmp/al68-tests-sol-4avy2ibx/rollback`. Seeded a user-owned `.gitignore`, selected state directory `.codex/hooks.json/state`, and ran actual native capture install with no existing hooks file. Initial planning succeeds; apply writes the hooks file and ignore entry; manifest-directory creation then fails because hooks.json is now a regular file. Install exited 1 with the expected mkdir error. The original ignore file remained byte-identical, and the newly created hooks file was absent. This exercises real rollback after prior writes and addresses the specific unexecuted branch acknowledged by TEST_REPORT. It does not prove crash atomicity or concurrent-edit ownership.

Reused the parent's recorded successful full make on the unchanged frozen fingerprint: Go vet/race tests, frontend strict checks/lint/tests/build, release/distribution checks and native cross-builds. Reused focused tester and setup reports after inspecting their assertions and limits. Did not rerun expensive suites without a new concern; the two native probes and fingerprint check above are this review's additional execution. No source mutation or before/after fix run was performed.

## Required evidence limits

SMOKE documents real OpenCode 1.19.5 Linux CLI interaction: independent roots and native child execute the binding tool, with exact IDs/model/usage checked against source records and unchanged reconciliation preserving snapshots. This is materially stronger evidence than callback fixtures. Raw source files were not re-read during this focused review; these source-grounded observations are reused, not claimed as my own real-harness execution.

Codex 0.166.0 actual callback delivery, hook-to-agent interaction, explicit binding and native-child/independent-root source comparison remain pending project trust review. No trust was granted or bypassed. Required supported-path completion cannot be inferred from version/features, installed hooks or synthetic receive/bind tests. Other unverified platforms/interfaces, OpenCode v2 native runtime and real resume/interruption paths remain the explicit limitations recorded in SMOKE. If TS-1 is fixed, the outstanding required real Codex evidence still prevents an unconditional complete-support approval unless the user approves an acceptance change.

The parent owns consolidation, re-review after corrections, final integration/make, telemetry and publication. This report gives no merge or delivery authorization.

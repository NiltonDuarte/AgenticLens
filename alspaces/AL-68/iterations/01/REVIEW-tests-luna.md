# Test effectiveness review — AL-68

Outcome: **INCOMPLETE**

Reviewer run: `tests-luna-01`; requested `codex/gpt-6-luna/medium`. Actual runtime model/effort metadata is not exposed in this context, so this report does not assert it. Review is limited to test effectiveness and required runtime evidence.

## Candidate and scope

- Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Candidate fingerprint: `c7dade3313a1d65a4269d595c782750655be8a044c020dc1f3ac502464cf325a`, independently recomputed from every path in `iterations/01/CANDIDATE.json`; all listed paths were present and matched.
- Requirements: AL-68 specification, `CONTRACT.md`, approved `iterations/01/PLAN.md`, test plan/report, smoke evidence, and the canonical regression/technical rules.
- Inspected new capture domain, adapter, CLI and distribution test paths, relevant Codex/OpenCode extraction dependencies, and the recorded smoke/runtime evidence. Product and test files remained read-only. Did not read prohibited review conclusions.
- No tests were run for this review; the supplied full `make` evidence and affected-package race runs are reused as reported. No additional run was needed to establish the runtime evidence gap.

## Findings and evidence gaps

### Required Codex agent-interaction oracle is missing — INCOMPLETE

The fixture oracles exercise the Codex receive/binding feedback contract and explicitly guard against treating the callback parent's session ID as a child ID (`capture/hooks_test.go`, `TestCodexHookFeedbackBindsOnlyAnExplicitResponse` and `TestSubagentHookDoesNotInventChildIdentityFromParent`). Those tests call the adapter directly with constructed payloads; they do not establish that Codex invokes the installed callbacks, delivers the feedback to the active agent, or completes an explicit ticket-binding response.

The smoke record says Codex 0.166.0 hooks are enabled and that an isolated hook configuration is ready for human project-trust review, but explicitly reports no callback, binding, usage or child-capture success pending that review (`SMOKE.md`, lines 27–29). Thus the user's required hook-to-agent interaction is not yet evidenced for Codex. This is an evidence gap, not a source-level finding that the Codex implementation is incorrect. Do not infer trust or bypass it; obtain the required real trusted-path interaction and compare root/child IDs and facts with source records before claiming this path supported.

OpenCode 1.19.5 has materially stronger evidence: the smoke records a real agent calling the binding tool for two independent roots and a native child, with source IDs/parent linkage and usage compared to its source store (`SMOKE.md`, lines 11–25). The test fixtures also cover missing/malformed/conflicting binding responses, exact source/store/session challenge matching, exact OpenCode selection across coexisting families, unavailable-source preservation, independent-ticket isolation, bounded continuation, digest-bound publication and redaction. These establish useful boundaries but do not substitute for Codex runtime evidence.

### Installer rollback after a partial apply lacks a failure oracle

`capture/install.go` lines 347–359 applies multiple file changes sequentially and attempts reverse rollback after a later apply fails. The CLI installer tests cover dry-run, idempotence, preservation of user entries/edits, disable/uninstall, and a missing-handler preflight rejection (`capture_test.go`, lines 37–150); the distribution script tests add useful downloader/wrapper boundaries per `SETUP_IMPLEMENTATION.md`. None of the recorded tests force a failure after one target has already been written, so they do not exercise the rollback branch or prove the first owned file is restored while unrelated content remains byte-identical. The tester report explicitly records this untested path (`TEST_REPORT.md`, evidence limits).

This is a missing acceptance oracle, not a demonstrated rollback bug. Add a deterministic disposable fault-path test (or equivalent filesystem probe) that fails a later target after an earlier successful apply, then asserts restoration of preexisting user hooks/files, removal of newly created owned files, and no surviving manifest advertising an active installation. The installer implementation currently has direct file operations and no injectable apply seam, so a test approach should preserve real filesystem behavior rather than mock away ownership semantics.

## Other test-effectiveness assessment

No concrete test-oracle failure was found in the reviewed bounded fixture paths. The tests distinguish source facts from binding metadata, unknown cost from known zero, accepted snapshots from truncation/partial parses, and publication from local request/path data. The recorded full `make` gate passed on the frozen candidate, and the focused race runs cover the affected Go packages; I reused that evidence rather than repeat broad suites.

## Limits

This review does not establish Codex compatibility, platform/version compatibility beyond the recorded Linux OpenCode runtime, real crash recovery during hook execution, or the untested installer partial-apply rollback case. The incomplete Codex evidence is specifically tied to pending trust review; no trust setting was changed and no bypass was attempted. Approval should wait for the required real Codex agent interaction/source comparison and an adequate rollback oracle or independently conclusive rollback evidence.

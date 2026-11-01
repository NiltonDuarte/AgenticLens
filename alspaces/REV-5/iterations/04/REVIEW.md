# REV-5 iteration 04 independent review

Verdict: **APPROVED** for ADP-01 and D3. No blocking findings in the reviewed diff against 88a4a1b. Parent integration and full make delivery gate remain pending; this approval is not a claim that the other iteration 03 findings are resolved.

## Contract alignment

The three-line recovery guard rejects each warning-bearing source before combination or replacement. The existing recoverCodex path retains the entire affected owner tree, reports incomplete recovery with source context, and proceeds with healthy independent roots. It covers parser warnings for partial trailing writes without changing tolerant parsing for other consumers or introducing schema/state machinery. The parser comment now matches behavior. Source/parent identity, repeated invocation facts and other harness facts are preserved by the existing atomic replacement boundary and the new real-parser persistence tests.

Release selection resolves an exact refs/tags reference to its commit and checks out detached before tool setup, dependency installation, checks and tests. Force input has precedence and bypasses release-please creation. Missing/invalid tags fail before metadata, and metadata checks both HEAD and tag identity. Upload uses the selected tag; forced rebuilds require an existing release and skip floating latest promotion and candidate cleanup. No invented tag fallback or shell interpolation of workflow input remains. Normal release creation/promotion conditions remain available. These changes fit the approved two-finding scope; listener/authentication and other audit issues are untouched.

## Engineering and test assessment

Production recovery remains in the adapter layer; no domain I/O, untyped value, exported test-only API, duplicated production implementation or new persisted format was added. The release verification harness is test code and executes the actual workflow shell blocks against isolated Git repositories; it does not add a runtime wrapper solely for testing. Its fixed-format workflow extraction intentionally fails on layout changes. The Makefile runs it through check, so it participates in the full gate (and requires Python 3 as a development check dependency).

The Go table exercises actual parsing through persisted ledger recovery for malformed middle/final JSON, partial final writes and invalid usage in a child, comparing retained DTOs and identities, healthy sibling refresh, repaired retry and unchanged-repeat snapshots. Existing recovery tests also cover injected warnings at the root. The tester's disposable overlay proves the warning guard is necessary for the assertions. Six workflow tests cover annotated and lightweight tag selection, force precedence, missing/invalid selection, stale HEAD and retargeted-tag rejection. Workflow conditions and upload tag wiring are reviewed structurally; they are not a hosted expression-interpreter simulation. No tests were deleted, no domain mocks or live ticket fixtures introduced, and no silent skips reported.

## Verification and limits

Inspected all changed production/test/build files, the complete new workflow harness, surrounding recovery/combination behavior, contract, plan, implementation/testing reports, reviewer persona, technical guidelines and applicable regression checklist. All six file hashes in TESTING.md match the reviewed files. Reused the tester's passing backfill/Codex race tests, six workflow tests and expected-failing overlay probe; no duplicate suite was necessary. Parent owns the final integrated make and embedded build evidence.

No hosted Actions execution, actual release lookup/upload, remote tag mutation or candidate deletion was performed. Existing release script/checksum issues outside D3 remain open. Historical full-code coverage remains tied to its own baseline. This report changes no implementation.

Status: APPROVED.

# AL-64 iteration 04 test-effectiveness review

Verdict: **APPROVED within the assigned test-effectiveness scope**, with one nonblocking test-hardening recommendation below. No demonstrated product defect or blocking acceptance gap. Parent owns consolidation, the integrated make/standalone gates and publication.

Reviewer run: checkpoint_tests_sol; independent CP-01 discovery pass. Parent-assigned profile: gpt-6.1-sol / high, reported as inherited runtime by the parent; this reviewer did not independently introspect model metadata. No sibling review reports or conclusions were read. Product and test files remained read-only; only this assigned report and disposable probes outside the checkout were written.

## Revision and scope

- HEAD/baseline: `d8a7c14cfd9f0ee626fedc33c70f92bd928d9192` plus frozen staged/working candidate CP-01.
- `/tmp/al64-04-review/candidate.diff` SHA-256: `9afb94b071ce4f74e32fad4ea409f1887899d2c3b9a7f3057fa3f4695d41da82`.
- Manifest SHA-256: `c766dd83ec3e9f601e6e5b302422defa2bbe6c86f5906648cae1e272a338a32d`. All 411 listed paths matched on entry and immediately before writing this verdict.
- Independently recomputed the tester's 403-path source/test fingerprint: `3d1449ceeaea15726d93879249aa7db616609d406d18febcc1a732a5805dfda6`, matching TEST_REPORT.md. Existing staged/unstaged ticket and implementation changes were preserved.

Read the approved 04 plan, immutable original contract with its explicitly superseding accounting/feedback boundaries, implementation/test/native-capability evidence, own persona/shared review procedure and applicable engineering/regression rules. Inspected every changed production/test hunk and relevant generation, capture, source grouping/replacement and test-helper callers. This is a focused diff review, not a 411-file audit.

## Nonblocking recommendation

**T-SOL-01 — severity 3/10: protect source-open qualification against a later superseding message after the accounting cutoff.**

Location: `backend/internal/adapters/opencode/checkpoint_test.go:20`; guarded production branches are `schema_v1_session_extraction.go:407` and `schema_v2_session_extraction.go:283`.

The existing table covers a completion timestamp after cutoff, but not a later separate native message. Those are different guards: completion tests exercise `finish == nil` / `Time.Completed == nil`; they do not protect `currentMessages[mid]` or `mi == len(msgs)-1`. Removing only those latter checks from disposable copies still passes the entire existing OpenCode package. The plausible regression would preserve numeric accounting at the requested historical cutoff but incorrectly call its interval currently source-open, emitting F03 instead of qualified F04.

A disposable extension to the existing table inserts a later user message at start+200000ms, extracts at start+180000ms, and expects unchanged numeric/token facts and F04 only. Both family controls pass the actual candidate. Both fail against the mutant, observing F03. Thus current production handles the case correctly; this finding concerns regression detection, not an existing incorrect output. A coherent addition to the current table would protect it without introducing a separate fixture system.

Evidence: `/tmp/al64-04-tests-sol/control.json`, `supersession_test.go`, `mutant_existing.json`, `mutant_existing.log`, `mutant_probe.json`, `mutant_probe.log`. These are synthetic diagnostic overlays, not native runtime evidence or product changes.

## Hypotheses investigated

1. **Rounded/grouped totals could trigger false long-interval advice. Refuted for the candidate.** Domain tests distinguish 179999ms from inclusive 180000ms, reject known short intervals whose subtotal reaches 180, preserve partial F04 and suppress pooling across rows. Actual Codex parser tests exercise rounding before selector invocation. OpenCode tests exercise real SQLite extraction, future completion qualification and numeric/token preservation. Inspected the tester's disposable exclusive-threshold and forced-F04 perturbations and failure logs: the relevant assertions fail. No assertion weakening or deletion was introduced in the original accounting matrix.
2. **Owner/model/row identity or stale/retained evidence could certify the caller. Refuted for inspected paths.** Selector tests reject stale cutoff, mismatched model, missing/future native identity, conflicting interval starts and duplicate evidence rows. Different source with the same session ID remains another owner. Adapter evidence is captured before model grouping; capture/explicit source merging replaces complete owner evidence, and retained snapshots provide no new native proof. The real CLI matrix covers all three source families with explicit and registered selection, root/child publication, completed-root replacement and an unavailable child's repeated F04 while retaining equal owner facts. Production supplies an empty caller proof, so the extraction root cannot produce F01/F02/F05.
3. **Copy, escaping, concurrency or ordering could silently regress. No candidate defect found.** Writer tests compare all five full approved copies and newline termination and exercise escaped control characters, quotes, backslash and Unicode line separation. Pure tests check owner/source/model/row ordering, repeated legitimate rows, no row pooling, concurrent exact and qualified cases, and one trailing F05. Static tracing confirms remaining role/native-start/native-ID sort keys and common escaping of all source-derived labels. These tests are meaningful protection, not an exhaustive claim that every possible ordering perturbation is detected.
4. **Optional diagnostics could fail publication or leak into persistent output. Refuted for inspected paths.** Actual generate subprocesses exercise default, disabled/custom threshold, an unwritable stderr descriptor and a summary publication failure. Success remains successful when feedback writing fails; failed publication has no checkpoint. The expanded source matrix separately captures stdout/stderr and exercises ledger, Markdown and actual HTTP reader boundaries, while preserving optional-metric omissions, numeric duration/tokens and child linkage. Main emits only after publication commit/acknowledgement and skips automatic capture. The latter guard was inspected statically; no new successful automatic-capture feedback assertion is claimed.
5. **Configuration addition could weaken existing defaults or numeric guards. No candidate defect found.** Real YAML/JSON loading plus writer/reload tests cover default180, zero, a nondefault17, int64 maximum and invalid types/ranges. Existing metric defaults remain asserted. The local override path remains unchanged; no new precedence policy was introduced. No new ledger DTO, schema, API field or frontend consumer was added. Transient source evidence is explicitly excluded from JSON.

## Verification

Reused valid frozen-candidate evidence, with matching source fingerprint and inspected logs:

- Public `make test-unit`: PASS, including 190 frontend units.
- Focused final CLI race command listed in TEST_REPORT.md: PASS, 14.083s.
- Final affected domain/migration/config/Codex/OpenCode/capture race command: PASS; capture 57.417s.
- Tester partial-lifecycle before/after failure evidence and threshold/F04 overlay failures: inspected; discovery/fix credit remains with the actors recorded in TEST_REPORT.md.

Executed independently from backend:

- Control: `go test -overlay /tmp/al64-04-tests-sol/control.json ./internal/adapters/opencode -run 'TestOpenCodeCheckpointPartialLifecycleRemainsQualified/.*/future_superseding_record' -count=1`: PASS, both families, 0.008s.
- Existing tests under supersession-guard mutation: `go test -overlay /tmp/al64-04-tests-sol/mutant_existing.json ./internal/adapters/opencode -count=1`: PASS, 0.845s. This intentionally demonstrates the narrow test gap.
- Added isolated probe under the same mutation: `go test -overlay /tmp/al64-04-tests-sol/mutant_probe.json ./internal/adapters/opencode -run 'TestOpenCodeCheckpointPartialLifecycleRemainsQualified/.*/future_superseding_record' -count=1`: expected FAIL, both families, F03 instead of F04.
- `git diff --check`: PASS. Final 411-path revalidation: PASS.

The overlay commands are non-race diagnostic probes and do not replace the public or integrated verification lanes. No duplicate broad unmodified suite was run.

## Limits

No trustworthy native generic-generate caller/current-interval transport or natural-completion/next-turn refresh was executed by this reviewer. Source fixtures and typed F01/F02/F05 tests do not establish those capabilities. The documented production F03/F04 fallback is within approval. Static inspection supports the automatic-capture exclusion and remaining deterministic sort keys; this review does not claim dedicated dynamic coverage for every one of those branches. Parent full make, standalone asset/binary verification and delivery approval remain separate gates. No new claim is made about historical native accounting evidence or unrelated known debts.

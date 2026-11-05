# Independent behavior review — AL-64 iteration 03

**Requested reviewer:** gpt-6-luna / medium. **Candidate:** frozen LEAN-01 over baseline `c68fbc9079d2f24c624d5de0fe87d313ed095b19`, branch `fix/AL-64-active-turn-duration`. **Manifest:** `/tmp/al64-03-review/fingerprint.json`, SHA256 `b05c2065769609e7e29cbb7e4b8158d804fc2b38ca905cea9b6a2bdbd1f237b5`; verified every one of its 399 path hashes against the checkout (0 missing or mismatched). This is a scoped behavior review, not a whole-repository audit.

## Verdict

**APPROVED** for the assigned behavior focus on the frozen candidate. I found no blocking behavior defect. This does not replace parent integration, `make`, standalone embedded-asset verification, or the other review focuses.

## Behavior hypotheses and evidence

1. **Numeric migration or retained-history publication may lose provisional seconds, alter owner identity, or re-emit the retired blob.** Traced invocation decoding through `NormalizeInvocationDurationJSON`, DTO validation/conversion, current-schema migration, and capture snapshot decoding. The shared decoder validates nonnegative integral totals/portions, enforces portion <= total, validates the old draft components/state/cutoff, accepts equal dual representations, rejects conflicting dual representations (including explicit zero), and emits the flat portion only when positive. Retained-history publication validates the facts before output projection and preserves owner/numeric facts. The real copied-ledger evidence independently reports a force-only canonical projection preserving all 19 session objects modulo the specified duration projection, then a native refresh retaining owners and replacing the completed snapshot without accumulation. No counterexample found.

2. **Aggregation may count provisional time twice or use a mismatched denominator.** Domain workflow, fleet, comparison, and time-bucket rollups sum `ProvisionalSeconds` independently while leaving `ActiveDurationSeconds` unchanged. Frontend progression does the same. The comparison KPI averages provisional portions over the same filtered rows used for its average-duration population and rounds to an integer; empty population reports neutral recorded duration. No duplicate addition to totals or unsupported completed-coverage claim found.

3. **API/frontend compatibility may reject or mislabel absent/zero/positive values.** API validation accepts absent provisional fields and validates present values against the corresponding total. UI labels identify only a positive amount as included provisional time; zero/absence produces neutral “Recorded duration” text. Legacy conversion is on the backend input path; frontend duration code no longer depends on draft objects, cutoffs, or coverage counts. No false “completed” label found.

4. **Source refresh may fail to replace provisional totals when source evidence becomes complete.** The source extractors carry the computed numeric portion, and the reviewed generation regression described in the test report completes one owner while retaining an unavailable child, then repeats refresh. Native copied-ledger evidence covers an authentic prior provisional turn later observed complete, current open activity, repeated source refresh, and a legitimate new child. The remaining source-state stopped-with-open-marker overestimate risk is explicitly accepted by the approved plan and is not a review blocker.

## Findings

None.

## Verification and limits

- Independently verified all 399 frozen manifest hashes and reviewed the scoped changed behavior paths plus relevant unchanged boundary consumers.
- `git diff --cached --check`: passed.
- Reused focused Go race checks, frontend 868-test/typecheck/lint results, and authentic native refresh evidence recorded in the ticket artifacts. I did not rerun suites or the native probe.
- The broad pre-alias Go race command had one CLI installation timeout on its earlier fingerprint; an isolated retry on final source passed. It remains a failed broad run, not a full-suite pass. Parent `make` and standalone embedded asset checks remain required.
- Native OpenCode runtime refresh under this representation and several unchanged-source retention cases are not newly established by the native probe; deterministic adapter/integration evidence and prior boundary evidence are described in the test report. These gaps do not reveal a behavior mismatch in the reviewed paths.

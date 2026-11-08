# AL-63 consolidated capability evidence

Application reviewed at f4152225, unchanged from eb7d169c application inputs. Final outcome is pending the last Luna data pass and current delivery gate. This is a bounded discovery audit, not a whole-repository correctness approval or permission to implement fixes.

## Independent first-pass roster

| Focus | Sol 6.1/high | Luna 6/medium |
| --- | --- | --- |
| Architecture | CHANGES_REQUESTED for unqualified metadata child-link support; AS-1 | APPROVED, declarations/fixture inventory only; no native ingestion |
| Behavior | CHANGES_REQUESTED, B-SOL-01–04, native/fixture counterexamples | APPROVED for source inspection; no native/probe execution; does not refute counterexamples |
| Data | APPROVED for bounded evidence; known REV-7 mixed-family limit | Pending |
| Test effectiveness | INCOMPLETE for full v2 preservation/error evidence; TS-01/02 | APPROVED for bounded inventory; unexecuted combinations remain explicit |

All eight roles were selected in advance. Sol native child turn-context metadata confirms gpt-6.1-sol/high for all four roles; Luna dispatch acknowledged gpt-6-luna/medium, while native linkage/model metadata is unavailable. Two initial Luna data/tests dispatches failed before execution with parent_not_active; fresh retry tasks preserve the selected settings. No failed dispatch counts as completed review. Native children and service delegates are separate coverage facts, not invented shared usage.

Reports: [architecture Sol](ARCHITECTURE_SOL.md), [architecture Luna](ARCHITECTURE_LUNA.md), [behavior Sol](BEHAVIOR_SOL.md), [behavior Luna](BEHAVIOR_LUNA.md), [data Sol](DATA_SOL.md), [tests Sol](TESTS_SOL.md), [tests Luna](TESTS_LUNA.md). Detailed evidence remains because product findings are unresolved follow-up work. [Compact candidate attribution](REVIEW_SUMMARY.json) retains every candidate, including evidence gaps rather than misclassifying them as product failures.

## Reconciliation and verification

CAP-01 through CAP-06 map to AS-1, B-SOL-01–04 and DATA-SOL-01 respectively; their symptoms, evidence and follow-up boundaries live in CAPABILITIES.md. No majority/source-only approval refutes a reproduced counterexample. CAP-06 is an explicitly documented policy difference owned by REV-7, not an unapproved demand for mixed-schema recovery. All are pre-existing product limits; AL-63 implements none of their corrections.

Tests/Sol independently found TS-01/02 durable-test evidence limits. Parent verified the requested boundaries through a disposable v2 CLI probe after first passes: genuine retained child skill/delegation and other-owner arrays, occurrence growth1to2, unchanged repeat, malformed/missing preservation, explicit mixed-age/incomplete reporting and two-read/one-error classification. This resolves the audit's specific evidence gap, not the absence of durable tests or native v2 positive samples. The parent's verification is separate from discovery credit.

Parent also checked source guards and paths for CAP-01–06, retained both-family malformed-source controls from data/Sol, observed HTTP native/recorded hydration and unavailable baseline reads, and executed native AL-55 copy backfill/repeat. No application changes were made; final source/hash checks and command outcomes are in the capability report and summary. Original AL-55/AL-62 hashes are unchanged. Fresh active/native/browser and heavy AL-62 timeout limits remain explicit.

# Reviewer instruction forward test

This bounded smoke evaluation does not establish a detection improvement. Both original and revised guidance detected all three seeded defect families and accepted both corrected controls. One run per instruction set cannot estimate production recall or false-positive rates.

## Design and provenance

Original persona/skill snapshots came from clean baseline d659d43 before editing. Revised snapshots came from the current working diff before evaluation. Two default native agents used fresh conversation contexts (fork_turns=none), the same inherited model/settings and equivalent requests differing only in instruction paths. Expected findings, author explanations, source revisions and prior conclusions were not supplied. Both were instructed to read only their assigned instructions, five raw source candidates and the common requirements. Filesystem access was not technically isolated; common harness instructions and paired source similarity may influence results. This is not a controlled cross-model benchmark.

EVALUATION.json records input hashes and source extraction provenance. A/D are recovery source excerpts; B/E are complete workflows; C is the actual SortLedgers function. C was held out from instruction-design examples, but was already known from the audit. The candidate inputs and model prompts were temporary, not application fixtures or runtime dependencies. Reproduction should extract from the recorded Git revisions into an isolated directory and use the stated common contracts; do not make tests depend on this ticket path.

## Observed outcomes

- A: both found parser warnings ignored before replacement, with a reachable source trace showing loss of persisted facts and false completeness. Neither claimed a real parser/persistence integration run.
- B: both found no blocker within the supplied release contract. Revised agent additionally reported execution of the actual tag-selection shell in a disposable Git repository; original inspected control flow. Hosted Actions/upload remained untested.
- C: both reproduced descending equality instability using the actual function with minimal substitute declarations. Neither probe establishes correctness of unavailable helper implementations.
- D: both found no blocker within the supplied recovery contract through static error/retention traces. No global correctness certification.
- E: both identified dispatch-revision build mismatch, action-output precedence over forced selection, and unwanted latest/RC mutation on forced rebuild. Neither executed remote operations. These are one seeded release defect family for scoring, not four independent benchmark cases.

Both initially flagged secondary tie-breakers in C because the supplied phrase "equal keys" was ambiguous. The evaluator clarified identically that the full comparator tuple includes permitted secondary tie-breakers. Both withdrew that finding and retained the equal-full-key descending defect. Treat this as an evaluator-contract correction; do not hide the initial extra finding or count it as an instruction improvement. EVALUATION.json hashes the clarified requirements; no further blind run was performed.

Both outputs provided concrete triggers and limits. Revised guidance prompted one additional local shell verification, but a single observed difference does not show causation or sustained benefit. Agent-reported probes were not independently rerun for this instruction smoke test. Token cost and comparative latency were not measured.

## Disposition

Keep the changes for explicit investigation responsibilities, coherent permissions and non-passing INCOMPLETE outcomes, not on a claim of better measured recall. Final wording only narrows evaluation to substantial behavioral instruction changes; the evaluation snapshots predate that proportionality clarification. Next meaningful calibration should use fresh, independently labeled diffs, unpaired controls, repeated comparable runs and hidden outcomes, including material access-gap cases. Do not optimize for finding counts or approvals. Production P2/P3 findings remain open.

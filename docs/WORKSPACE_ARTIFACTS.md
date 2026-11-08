# Writing workspace artifacts

Write for a human reviewing the ticket without the conversation. Lead with the
problem, decision or outcome. Use plain language, short paragraphs and concrete
behavior. Include technical detail when it explains a decision or proves a result.

## Give each document one job

These are content guides, not mandatory heading templates. Omit empty or
inapplicable sections. Lengths are targets for authored prose, not quotas or hard
limits; exceed them when a material decision, risk or required evidence needs it.
Generated metrics and structured review records are outside these prose targets.

| Artifact | What the reader needs | Usual length |
| --- | --- | --- |
| `CONTRACT.md` | Problem, agreed outcome, scope exclusions and observable acceptance criteria. Link applicable technical rules; spell out only ticket-specific constraints. | 200–400 words |
| `iterations/<NN>/PLAN.md` | A short ordered path to that outcome, affected components/owners and verification. Later iterations describe the acceptance delta and link the baseline. Include the required review coverage when applicable. | 200–400 words |
| `SUMMARY.md` | Current delivered behavior/status, meaningful decisions, actual verification, review outcome and remaining limits. Link the contract, active plan and retained findings as needed. | 300–600 words before generated metrics |
| `PULL_REQUEST.md` | Conventional title on the first line, problem and resulting behavior, concise validation and material limits; link the contract/summary for detail. | Usually 100–250 words; at most 3000 characters |
| Product specification | User problem, expected behavior, acceptance and dependencies. Keep implementation planning in the developer ticket. | Proportional to the product decision |

Small tickets can be much shorter. Do not pad a document to reach a range or add a
length-justification section. A substantial ticket may need more detail, but first
remove duplication and link evidence already recorded elsewhere.

## Keep the reading path short

- State each fact in its owning artifact; use a relative link rather than copying
  whole scope, acceptance, verification or review sections between documents. Each
  document still needs enough context to make sense on its own.
- Update the summary's current account when work changes. Retain material decisions
  and their reasons, but omit routine tool attempts, dispatch logs, repeated status
  updates and abandoned approaches with no bearing on the result. Git preserves
  prior versions; the summary is not a session transcript.
- Record a material steering delta once in its iteration plan: source, changed
  acceptance, superseded decision, preserved guarantees and verification impact.
  Link that delta from the summary. Clarifications, review fixes and routine
  integration stay in the current iteration unless acceptance changes.
- Name paths and symbols only where they locate a change or evidence. Avoid file
  inventories, repeated full commit hashes, internal run IDs and agent namespaces
  in prose. Keep the necessary baseline/reviewed revision once at its evidence.
- Report checks as command, result and tested revision/diff; name skipped stages
  and relevant environment limits. Keep raw output and private diagnostics outside
  the reading path and Git according to repository policy. Planned checks are not
  results. Do not add an evidence sidecar for every routine action.
- Summarize review outcomes and material findings in plain language; link the
  compact findings record. Preserve every required discovery attribution,
  verification actor, disposition and limit under the
  [review retention protocol](../.agents/skills/agentic-lens-review/references/parallel-review.md#completed-review-retention).
  Group identical outcomes/settings rather than narrating each run. Never shorten
  away unresolved findings, accepted debt or incomplete coverage.
- Keep CLI-generated metric sections intact. Do not repeat their tables or token,
  cost and session totals in authored prose. Explain material coverage gaps once.

## Example summary prose

Illustrative only; these are not actual test results:

> Empty exports now show “No sessions found” and still produce a valid CSV header.
> Previously they failed while calculating totals. The CSV format is unchanged.
>
> Verification: `make` passed on revision `abc1234`, including the empty-export
> regression. The behavior review found no blocking issues. Hosted CI is pending.

Before handing off, read the opening paragraph and scan the rest: can a reviewer
identify what changed, why, what supports it and what remains unresolved? Remove
anything that does not help answer those questions or preserve required evidence.

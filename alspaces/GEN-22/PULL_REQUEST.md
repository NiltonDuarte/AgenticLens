# chore(docs): simplify agent instructions and human review artifacts (GEN-22)

Repeated workflows and policy made agent instructions long, while generated
artifacts obscured goals with execution history and technical inventories.
Consolidate repository policy in AGENTS.md, workflows/artifact definitions in
skills and assigned duties in personas. Remove repeated command recipes and
delegation prompts, reducing eight instruction files by about 44% in the final
simplification pass without moving that bulk into new references.

Artifacts have distinct purposes and practical prose targets. PO specs/tickets
lead with goal, context and observable acceptance; optional technical notes come
last in the same document. Required approval, review credit, evidence, telemetry
and delivery gates remain. Affected skills are versioned at 0.1.1, adding the
previously absent management VERSION record. Historical tickets and application behavior are
unchanged.

Validation: links/anchors, skill metadata, whitespace and an independent four-case
instruction walkthrough. Integrated checks and limits: [summary](SUMMARY.md).
Scope: [contract](CONTRACT.md) and [current plan](iterations/04/PLAN.md).

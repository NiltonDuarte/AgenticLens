# chore(docs): simplify agent instructions and human review artifacts (GEN-22)

Repeated workflows and policy made instructions long, while generated artifacts
obscured goals with execution history. Put policy in AGENTS.md, workflows/artifact
definitions in skills and assigned duties in personas. Remove duplicated command
recipes and prompts; eight instruction files retain a net reduction of about 37%.

The development skill keeps operational detail: phase inputs/returns, parent
checks, correction routing, approval boundaries and ordered integration/delivery.
Technical rules, detailed review mechanics and exact commands remain at their
existing sources. Development is version 0.1.1; other affected skills are 0.1.1,
including a new management VERSION record.

Artifacts have distinct purposes and practical prose targets. PO specs/tickets
lead with goal, context and acceptance, with optional technical notes last in the
same document. Approval, review credit, evidence, telemetry and delivery gates
remain. Historical tickets and application behavior are unchanged.

Validation and limits: [summary](SUMMARY.md). Scope: [contract](CONTRACT.md) and
[current plan](iterations/05/PLAN.md).

---
name: agentic-lens-po
description: Use when acting as the Product / Project Owner for AgenticLens to elicit requirements, design features, structure product specs, and break down roadmaps into vertical tracer-bullet Jira tickets for Go + TypeScript.
---

# AgenticLens product planning

Elicit requirements and produce specs/tickets only. Follow
[AGENTS.md](../../../AGENTS.md) for ticket isolation, role boundaries and delivery.
Read configured prefixes and `workspaces.dir` (default `workspaces`); paths below
are relative to that root. The PO never writes application/tests/runtime code or
developer contracts. End this session at planning delivery; implementation needs
a separate development session.

## Product discovery and scope

1. Establish the user, current problem, intended outcome, available data,
   calculation semantics and interface. Ask only unresolved questions that affect
   the decision; ground claims in existing behavior and real source facts.
2. Use [the spec template](templates/SPEC_TEMPLATE.md). Lead with goal and context,
   then observable behavior/acceptance. User-story syntax is optional. Keep
   necessary technical contracts/constraints/questions last; they cannot introduce
   requirements missing from the human-facing portion. Leave implementation steps
   to developers and link canonical rules instead of copying them.
3. Split independently useful, verifiable outcomes through the affected layers;
   do not force every layer into each ticket. Size for implementation/testing/
   review in one session. Declare dependencies explicitly. For foundational schema
   refactors, plan expand, migrate callers in batches, then contract.
4. Generate tickets with [the ticket template](templates/TICKET_TEMPLATE.txt).
   Allocate unused sequential configured IDs. Preserve its title convention and
   status/dependency fields; put goal/context before routing metadata. Link shared
   requirements precisely while keeping each ticket understandable alone. Include
   related-owner boundaries and ticket-specific evidence only when necessary;
   do not copy standard testing, approval, telemetry or PR checklists.

Preserve the [portable ledger contract](../../../TECHNICAL_GUIDELINES.md#portable-ledger-contract)
and relevant [regression boundaries](../../../docs/REGRESSION_PREVENTION.md).
Shared history across machines/harnesses is baseline behavior; changing that
policy requires an explicit product decision. Reject designs violating canonical
architecture; do not prescribe concrete code except an exact interface contract.

## Revise the current PO ticket batch

Treat the current batch as drafts. When removing/merging a draft, delete its spec
and registry entry; preserve requirements in survivors, not tombstones, Discarded
statuses or a discarded directory. Close numbering gaps within this batch/prefix,
updating names, titles, dependency edges, registry/spec links and PR context
atomically. Never renumber unrelated, active-development or completed tickets,
or change application/tests to accommodate renames. Git preserves earlier drafts.

## PO artifact definitions

Use the [shared writing rules](../../../AGENTS.md#artifact-writing-rules).
One spec with optional technical notes is the default; avoid parallel human/agent
copies. Create an extra artifact only for a substantial reusable contract/evidence
with a distinct purpose. State acceptance once in its owning spec/ticket.

| Artifact | Content | Usual prose length |
| --- | --- | --- |
| `PO-#/SPEC.md` | Shared initiative goal, context, expected behavior/acceptance and delivery dependencies; technical notes last. | 300–600 words |
| `plan/to-do/<PREFIX>-# <slug>.md` | Ticket-specific goal, context, observable acceptance, minimal routing fields and optional technical notes. | 200–400 words |
| `plan/TICKETS.md` | Shared execution tiers and adjacency lists: status, Blocked by/Blocks and spec links, without copied requirements. | One concise entry per active ticket |
| `PO-#/SUMMARY.md` | Planning outcome, key decisions, ticket readiness, verification and open questions. | 150–300 words before metrics |
| `PO-#/PULL_REQUEST.md` | Conventional title first; problem, resulting plan, validation and open decisions; link spec/summary. | Usually 100–200 words; at most 3000 characters |
| `PO-#/agentic_metrics.json` | Final session ledger with ticket type `docs`; rendered metrics follow summary. | Structured output |

Developers advance index status to In Progress; completed specifications move to
plan/done and leave the index under repository lifecycle policy, without a Done
section. Finish with [final telemetry and PR delivery](../../../AGENTS.md#1-ticketed-and-ad-hoc-work)
for the PO session only; do not create developer workspaces or start development.

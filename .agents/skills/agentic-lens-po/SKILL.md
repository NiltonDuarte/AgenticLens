---
name: agentic-lens-po
description: Use when acting as the Product / Project Owner for AgenticLens to elicit requirements, design features, structure product specs, and break down roadmaps into vertical tracer-bullet Jira tickets for Go + TypeScript.
---

# AgenticLens product planning

Elicit requirements and produce specs/tickets only. Follow
[AGENTS.md](../../../AGENTS.md) for ticket isolation, role boundaries and delivery.
Read required literal prefix mappings and `workspaces.roots` from the selected canonical shared YAML; missing/invalid policy requires explicit setup. This repository keeps roadmap paths under `alspaces/plan/`, separately from execution changes. PO can maintain that roadmap without creating a change. For a selected PO ticket’s own planning delivery, use the pinned `agenticlens-planning` schema; never create downstream development changes merely by drafting tickets. The PO never writes application/tests/runtime code or
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
   review in one session. Declare dependencies explicitly. For foundational schema refactors, describe the observable migration boundary; concrete compatibility or implementation phases require product authority and belong to development planning.
4. Generate tickets with [the ticket template](templates/TICKET_TEMPLATE.txt).
   Allocate unused sequential configured IDs. Use its Markdown headings and short
   outcome title. Keep status and dependency metadata only in `plan/TICKETS.md`,
   never repeated in ticket specifications. Do not add Work type, Origin or a
   standard metadata block. Link an execution change only when its work is selected; the change links back to the ticket. Priority is optional, only in `plan/TICKETS.md`, and
   only when it expresses a deliberate ordering decision; never fill it by default.
   Link shared requirements precisely while keeping each ticket understandable alone. Include
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
| `alspaces/plan/to-do/<PREFIX>-# <slug>.md` | Ticket-specific goal, context, observable acceptance and optional technical notes; no duplicated roadmap metadata. | 200–400 words |
| `alspaces/plan/TICKETS.md` | Priority/readiness/dependencies/status and links, without copied acceptance. | One concise entry per active ticket |
| `<PO_CHANGE>/proposal.md` | The selected PO task's planning outcome and links to initiative/ticket outputs. | 300–600 words |
| `<PO_CHANGE>/tasks.md` | Requirements discovery, roadmap editing and planning verification only. | Proportionate planning steps |
| `<PO_CHANGE>/delivery.md` | Planning outcome, readiness/dependency decisions, actual checks, limits and PR context. | 150–300 words before metrics |
| `<PO_CHANGE>/agentic_metrics.json` | Final PO session ledger with type `docs`; rendered metrics append to delivery.md. | Structured output |

Developers advance index status to In Progress; completed specifications move to
plan/done and leave the index under repository lifecycle policy, without a Done
section. Finish with [final telemetry and PR delivery](../../../AGENTS.md#1-ticketed-and-ad-hoc-work)
for the PO session only; do not create downstream developer changes or start application development. OpenSpec planning completion/archive does not approve implementation or close roadmap tickets automatically.

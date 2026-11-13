# AL-87: Align agents and skills with OpenSpec

## Goal

Repository agents and contributors can follow one coherent OpenSpec-based workflow from product planning through implementation, review, telemetry and archive.

## Context

Current AGENTS.md, personas, PO/development/review/management/setup/operator skills and templates encode alspaces, plan/to-do, separate contracts and iteration-plan paths. Updating only the entrypoint would leave contradictory instructions. [PO-24](../../PO-24/SPEC.md) adopts OpenSpec while retaining the repository's approval, role and evidence guarantees.

## Acceptance criteria

- Reconcile all active instruction sources: AGENTS.md, technical/regression guidance where affected, shared personas/references, Codex registrations and applicable OpenCode setup, every affected AgenticLens skill/template, and installation/operator documentation. Search for conflicting active path/lifecycle rules and resolve them at their owning source; preserve unrelated historical audits.
- Define one authoritative artifact mapping inside an OpenSpec change for proposal/acceptance, design, tasks, delivery evidence, review retention and telemetry. Eliminate parallel independently maintained plans/contracts and duplicate roadmap truth; link canonical material instead.
- PO remains planning-only and ends at planning delivery. Scope approval before coding, specialist implementation/test/review roles, review attribution, required verification, human approval boundaries and final telemetry/PR handoff remain explicit. Artifact readiness, task completion, approved scope and verified delivery cannot substitute for each other.
- Establish the ticket-prefix convention and stable-identity rules for all configured ticket types, including PO/GEN work. Guidance resolves the current change location instead of assembling a path from the ticket ID, and all relevant commands use the correct conventional telemetry type.
- Archive guidance preserves evidence and supports late telemetry refresh without creating a new active change. Legacy project guidance remains usable where OpenSpec is not selected; setup regeneration cannot overwrite custom instructions silently.
- Demonstrate a bounded planning-to-handoff walkthrough against the chosen OpenSpec schema, including changed scope, incomplete review, unavailable telemetry and archive. Check links and generated guidance; report what was actually exercised rather than claiming broad behavioral support from text searches.

## Implementation notes

This ticket owns canonical workflow/guidance and schema/template customization needed for those instructions. It does not perform the live backlog/configuration cutover. Model-review policy and portable-history authority are preserved rather than redesigned as an incidental adoption change.

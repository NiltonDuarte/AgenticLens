# PO-18: Post-v2.0 platform follow-up

## Purpose and audience

Engineers, analysts and operators need clearer historical views, reliable source-backed facts, accurate active duration, understandable generation/recovery controls and an assessment of the telemetry repository architecture. This specification consolidates the user's 2026-10-31 review and subsequent corrections into one current plan. Detailed acceptance criteria live in the linked AL tickets.

This is PO planning, not application implementation or an independently reviewed development plan. Development runs in separate agentic-lens-dev sessions under the normal contract and approval requirements. The capability audit, backfill research and architecture review deliver evidence/recommendations within their stated boundaries.

## Planned tickets

| Ticket | Outcome | Status | Blocked by |
| --- | --- | --- | --- |
| [AL-60](../plan/to-do/AL-60%20search%20tickets%20in%20comparison%20exclusion%20panel.md) | Search exclusions by ID/title without changing selections or chart membership. | Ready | None |
| [AL-61](../plan/to-do/AL-61%20show%20full%20local%20timestamps%20in%20ticket%20deep%20dive.md) | Show full local historical date/time without milliseconds or a hover requirement. | Ready | None |
| [AL-62](../plan/to-do/AL-62%20fix%20codex%20skill%20and%20delegation%20facts%20with%20granular%20ownership.md) | Diagnose and fix missing source-recorded Codex skill/delegation facts; deliver granular ownership, safe migration and all consumers. | Ready | None |
| [AL-63](../plan/to-do/AL-63%20audit%20capabilities%20across%20opencode%20v1%20v2%20and%20codex.md) | Audit OpenCode v1/v2 and Codex ingestion, hydration, live capabilities, facts, errors, child navigation and recovery. | Blocked | AL-62 |
| [AL-64](../plan/to-do/AL-64%20investigate%20and%20fix%20active%20turn%20duration%20across%20codex%20and%20opencode.md) | Fix diagnosed Codex active-turn duration and investigate/fix confirmed OpenCode v1/v2 defects with snapshot/replacement evidence. | Ready | None |
| [AL-65](../plan/to-do/AL-65%20research%20safe%20default%20source%20backed%20backfill.md) | Research default exact-session source recomputation, unavailable-history preservation and clearer Force semantics. | Blocked | AL-63 |
| [AL-66](../plan/to-do/AL-66%20configure%20optional%20metric%20output%20for%20generate.md) | Add independent enabled-by-default toggles that omit skills/task_links/anomalies from generated JSON and managed summaries when disabled. | Ready | None |
| [AL-67](../plan/to-do/AL-67%20review%20telemetry%20repository%20interfaces%20and%20dependency%20injection.md) | Review shared source repository contracts, source/schema selection and injection; deliver as-is architecture, findings and a proposed design. | Ready | None |

Dependency chain: AL-62 -> AL-63 -> AL-65. Other tickets are independently Ready; do not delay the diagnosed duration fix behind the audits. The shared registry is alspaces/plan/TICKETS.md. Existing AL-57 and AL2 plans are unchanged.

## Functional and compatibility boundaries

- AL-62 implements fixes, not merely an investigation report. It also owns granular data-structure changes: canonical facts belong to their narrowest source-proven session/invocation owner; ticket/fleet views derive from them. Preserve raw identifiers, repeated occurrences and delegating-parent ownership. Invocation metric rows are not assumed to be raw-call identities. Safely retain ambiguous legacy ownership without guessing or losing facts.
- AL-63 distinguishes supported-and-verified, supported-but-failing, unsupported, source-unavailable and unexecuted behavior. Include source-backed recovery and missing/malformed-source preservation checks. Do not assume all sources promise equal capabilities.
- AL-64 covers Codex and both OpenCode schema families independently. Include source-proven unfinished activity through a cutoff, provisional status, replacement after completion, idle/future exclusion and repeat-backfill stability. Equivalent OpenCode defects are not yet established; implement confirmed defects and preserve already-correct behavior. Never invent duration when active boundaries are unavailable.
- AL-65 researches the user's desired default: follow exact session identities already in a ledger, recompute when original source metrics exist, retain stored facts when unavailable, and make outcomes explicit. Current Force only bypasses completed-ledger skipping; it is not source recomputation or heuristic recovery. No default/flag behavior change is approved by the research ticket.
- AL-66 applies category toggles across all emitted ownership scopes and compatibility projections, including previously saved disabled facts in regenerated output. Preserve unrelated metrics and user-authored summary prose. Disabled output is distinct from unavailable source data; toggles do not imply global redaction or new backfill behavior.
- AL-67 reviews actual abstractions and composition across generate, backfill, server, hydration and live consumers, including mixed-source routing. Maximize shared meaningful contracts while preserving optional capabilities and explicit failure/partial semantics. Refactor implementation is separate scope.

## Evidence and verification boundaries

Original review: 2026-10-31, build v2.0.0-rc.106+9707d41-1-g45cb679. User-supplied observations include 102 ledgers without populated top-level skills and AL-55 with a root plus four Codex children. These are not new runtime checks performed by this PO task. AL-62 preserves missing-fact context; AL-64 preserves the exact duration timeline and measured delta.

The user confirmed chart defaults, axes, point tooltips/navigation, exclusion/restoration, logarithmic rendering and the Skills & Subagents empty-state behavior. Populated skill analytics, successful backfill/preservation and deferred live checks are not established by the notes. The standalone active OpenCode walkthrough was removed; its checklist is not reassigned. The capability matrix still reports live support and evidence gaps.

The original review worktree/evidence is not part of an application merge. Preserve alspaces/AL-55/agentic_metrics.json byte-for-byte. No application, test, schema, runtime script or actual configuration changes are made by this PO delivery. Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md: FCIS, strict Go/TypeScript, zero any/interface{}, zero emojis, self-contained fixtures and baseline-fact preservation when optional hydration/recovery fails. Implementation tickets own their necessary regression and embedded-asset checks.

## PO deliverables and decisions

PO-18 contains SPEC.md (this consolidated plan), SUMMARY.md (delivered planning work and verification facts), PULL_REQUEST.md and source-grounded agentic_metrics.json. Per-user instruction, development-style iteration plans and self-review files are removed. Author validation is not independent plan approval. Earlier revisions remain in Git history.

Eight active tickets now run consecutively from AL-60 through AL-67. Removed/merged PO drafts are deleted, retained scope is incorporated into surviving tickets, and numbering and references are adjusted together. No discarded ticket records are created by this PO session. The prior standalone live ticket was removed and later tickets renumbered; current IDs above are authoritative. Redundant phase/prefix metadata was removed. No new iteration folders or developer contracts/workspaces are created for this consolidation.

# Feature Specification: Phase 1 post-v2.0 platform follow-up (PO-18)

## Problem Statement

The 2026-10-31 review of AgenticLens build `v2.0.0-rc.106+9707d41-1-g45cb679` identified two presentation improvements, an active-turn duration accounting bug, missing Codex fact visibility, deferred OpenCode live verification, source capability uncertainty, and confusing backfill defaults. This specification turns those notes into Phase 1 AL tickets. It does not implement fixes or claim verification not performed.

## Target Personas & Use Cases

Engineers inspecting historical/live agent activity, analysts comparing tickets, and operators recovering telemetry need understandable controls and trustworthy source-backed facts.

## User Stories

1. As an analyst, I want to search ticket choices without changing exclusions so that large comparison panels remain manageable.
2. As an engineer, I want full local historical timestamps so that dates remain clear without hover.
3. As an operator, I want source-grounded skill/delegation and capability findings so that unavailable evidence is not mistaken for no usage.
4. As an engineer, I want verified OpenCode live behavior so that refresh and captured diagnostics remain trustworthy.
5. As an operator, I want current-turn duration counted provisionally so that generation reflects elapsed active work.
6. As an operator, I want safe backfill semantics researched so that available sources refresh while unavailable history survives.

7. As an engineer, I want skills and task links attached to their source session or invocation so that raw provenance survives aggregation.

## Functional Requirements

- Data ingestion/extraction: AL-62 investigates Codex facts; AL-64 audits OpenCode v1/v2 and Codex capabilities; AL-65 fixes active-turn duration through a snapshot cutoff. Evidence unavailable at execution remains explicit.
- Granular ownership: AL-67 relocates canonical skills/task_links to source-proven session or invocation ownership, with derived ticket/fleet projections, compatible migration, and explicit unresolved legacy provenance.
- Domain aggregations: duration excludes inter-turn gaps and future activity; completed refresh replaces provisional contribution. Missing facts do not establish measured zero. Stored baseline facts survive unavailable hydration/recovery.
- Presentation: AL-60 searches existing ticket IDs/titles inside Exclude tickets only; AL-61 renders full local dates/times without milliseconds. AL-65 exposes provisional duration consistently. AL-63 verifies live inspector behavior.
- CLI/recovery: AL-66 researches default source-backed recomputation with exact existing identities, retention when originals are unavailable, and explicit outcomes. It does not enable Force or change defaults.

## Non-Functional & Architectural Invariants

Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md: pure Go domain, I/O in adapters, strict TypeScript, zero any/interface{}, no emojis, self-contained fixtures, compatible fact preservation and meaningful tests. UI-only slices reuse existing domain/API contracts rather than inventing backend work. Research and verification slices trace the whole source-to-presentation path but authorize reports and experiments only. Application changes require a separate agentic-lens-dev session and its approval gate.

## Decomposition & Delivery Roadmap

| Ticket | Outcome | Kind | Blocked by |
| --- | --- | --- | --- |
| AL-60 | Search tickets within comparison exclusion panel | Feature | None |
| AL-61 | Full local historical dates/times in Ticket Deep Dive | Feature | None |
| AL-62 | Codex skill/delegation evidence investigation | Investigation | None |
| AL-63 | Active OpenCode live walkthrough | Verification | None |
| AL-64 | OpenCode v1/v2 and Codex capability matrix | Audit | AL-62, AL-63 |
| AL-65 | Provisional unfinished Codex turn duration fix | Bug | None |
| AL-66 | Default backfill/Force research recommendation | Research | AL-64 |
| AL-67 | Source-owned skills/task_links and compatible projections | Refactor | None |

Execution tiers: Tier 1 = AL-60/61/62/63/65/67; Tier 2 = AL-64; Tier 3 = AL-66. AL-65 is the diagnosed must-fix and can proceed independently. AL-64 consumes AL-62/63 reports; AL-66 consumes the parity report. Blocked tickets may gather independent evidence while their final acceptance awaits dependencies. Do not delay the duration fix behind investigations. Existing AL-57 and AL2 scope is unchanged.

## Evidence and Review Boundaries

The notes are supplied evidence, not new checks run in this PO session. AL-62 preserves the reported 102-ledger observation and AL-55's four children. AL-65 carries the root session, exact timeline and duration delta. The user confirmed chart defaults, axes, point tooltips/navigation, exclusion/restoration, logarithmic rendering, and Skills & Subagents empty-state collapse/reopen. Populated skill analytics/filtering, live behavior, and successful backfill/preservation remain unverified.

The original review worktree was not intended for merge. These planning artifacts alone are eligible for delivery; the original AL-55 ledger must remain byte-identical. No application fix, experimental ledger replacement, original review changes, default backfill change, or forced parity promise is included. No developer AL workspace or CONTRACT.md is created by the PO.

## Iteration 02 steering: granular fact ownership

Source: user follow-up requesting skills/task_links at session or invocation scope, as close to raw data as possible. AL-67 adds one independently Ready Phase 1 ticket; it does not revise the seven existing ticket boundaries. Canonical facts use their narrowest proven owner; raw occurrence identity must not be confused with role/model metric rows. Ambiguous historical ownership is retained explicitly, never guessed. This supersedes the original seven-ticket count with eight planned tickets and introduces no application changes or new harness capability promise. Original evidence preservation, research-only defaults, existing verification and separate implementation-session requirements remain in force.

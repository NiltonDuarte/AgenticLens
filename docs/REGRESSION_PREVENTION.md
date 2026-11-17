# Preventing recurring regressions

Use this checklist when a change touches the corresponding boundary. Put the selected cases and their results in the active iteration plan and review.

## Which document owns the rule?

- [AGENTS.md](../AGENTS.md): ticket ownership, approval, orchestration, telemetry, and delivery policies.
- [TECHNICAL_GUIDELINES.md](../TECHNICAL_GUIDELINES.md): architecture and data correctness requirements.
- [CONTRIBUTING.md](../CONTRIBUTING.md): executable local build and verification entry points.
- This checklist: select regression cases before implementation. Historical ticket artifacts record evidence; they are not current command references.

When changing a shared rule, search active instructions, personas, skill references, setup templates, and user docs for the old wording. Update affected guidance or link to its owner. Preserve historical records. Distinguish an artifact-allocation default from a runtime configuration requirement: a setup writer’s explicit workspace choice does not authorize the server to guess an unconfigured path.

## Interpret steering against the current accepted scope

Use the [contributor delta format](../CONTRIBUTING.md#71-recording-steering-without-restarting-the-ticket) for material scope changes. Keep the current contract consistent with accepted superseding decisions; do not replay an earlier plan just because it is more detailed. Keep routine merge checks and review fixes within the current iteration when acceptance is unchanged.

Distinguish intentional feature additions from regressions. Repeated iterations are not themselves failures. Missing or consolidated plans are limits on historical evidence, not a reason to fabricate a steering history.

## Before implementation

- Trace the affected value from source/configuration through extraction, DTO validation, persistence, read projection, API, domain calculations, and UI. Identify the authoritative owner and where information may be lost.
- Classify each missing value as required, optional, or explicitly unknown. Specify error/partial/empty behavior. Never substitute unrelated sessions, synthetic facts, a feature classification, completion dates, or measured zero for unavailable evidence.
- For shared configuration, identify every consumer and generated template. Use a non-default value in a focused test so hardcoded defaults cannot pass unnoticed.

## Select verification by boundary

### Telemetry extraction, refresh, or migration

Use a root with children, a second root, repeated legitimate invocation rows, an unavailable stored session, and an unrelated harness. Check session IDs, parent links, source labels, row multiplicity, and other-source facts before and after refresh. Ensure malformed or missing-source fixtures pass earlier validation and actually reach the failure branch being asserted. Keep session/message IDs intact when changing invocation fields. Check a growing source and then an unchanged repeat: cumulative snapshots must not be added twice. Failed discovery/parsing/grouping must preserve stored facts and report incomplete extraction. Exact stored identity is the default; heuristic recovery requires explicit opt-in.

Use the actual harness and a literal nonce in the final telemetry invocation. Supply ticket ID, title, and conventional type on new ledgers; verify command help before using options. Do not use an empty ledger to conceal extraction failure. See AGENTS.md for finalization timing and delivery requirements.

### Persisted facts and optional hydration

Exercise supported ledger facts with enrichment unavailable, pending, failed, empty, partial, and successful. Available baseline facts must remain visible. Read projections must not reuse compact write transforms that remove accepted facts. GET requests must not rewrite ledgers. Ignore responses for obsolete ticket selections; reconcile by established ownership/identity and preserve ambiguous repeated occurrences. Empty events do not establish a legacy schema.

### Portable ledgers across developers and machines

Apply the [canonical portable ledger contract](../TECHNICAL_GUIDELINES.md#portable-ledger-contract). Use isolated fixtures for two developers/machines with separate local state; never require real historical ticket files or another developer's private data. Select the cases relevant to the changed boundary:

- Developer A generates and commits a ledger. Developer B clones it into a different path with no A session logs, capture database, cache or receipts. Supported historical facts remain readable and available to CLI/API/dashboard projections without claiming new source verification.
- B generates with only B's source sessions available. A's recorded owners/facts and unrelated harness history survive save/reload; only successfully refreshed owners change. Also exercise unavailable owners from the same harness, not just a different harness.
- Change values in an exact locally available source session and run generate/update/backfill. Check local-source precedence for refreshed values, unchanged unavailable sessions and whole-ledger aggregates including both developers' data. Repeat unchanged input and verify no duplicate sessions/invocations or accumulated snapshot totals; retain distinct legitimate repeated occurrences.
- Refresh one session from a complete local snapshot with fewer invocations, then with a genuinely empty complete snapshot: its stored-only invocations are removed, while every other session remains untouched and whole-ledger totals are recomputed. Contrast missing, failed and partial extraction: those results must not be mistaken for an authoritative empty snapshot or erase unavailable history.
- Restore an older but complete local source for the same exact session ID. Confirm it remains authoritative even when the ledger has higher totals or a newer recorded timestamp; report the replacement/count changes and preserve every other session. Moving the session between machines/checkout paths must not require an additional identity or private receipt.
- If the same session ID appears in distinct tickets, preserve both recorded attributions. Cross-ticket overlap detection is informational and must not silently deduplicate data or choose a ticket for the user. Verify warning behavior when that presentation is in scope.
- Perform actual byte-changing formatting and a supported legacy migration, then refresh with the original sources unavailable. Preserve owned recorded facts, other supported history and legitimate multiplicity. For the explicit [ownership and lean-occurrence migrations](OWNED_OCCURRENCES.md), assert separate ownerless-row deletion and schema 6 retired-field/event audit counts, preserved proven child IDs and multiplicity; known unavailable owners remain intact and ordinary reads require migration rather than performing it silently. Keep required pending enrichment/schema state honest; do not require a local fingerprint to authorize preserved history. Exercise an explicitly copied valid workspace where the affected operation supports it, rather than assuming a pinned Git copy is its only persisted input.
- Block a migration step with missing source data, then move the ledger/checkpoint to a second isolated machine fixture that can finish it. Verify explicit resumable progress, preserved historical records at the blocker and completed shape/aggregate consistency after continuation. Failure must not truncate the ledger or silently declare completion.
- Keep negative coverage for malformed/unsupported schema, conflicting identities, unsafe paths and invented extraction outcomes. Distinguish those failures from a valid record merely lacking local source corroboration. Hashes used for transport integrity or concurrent-write/recovery checks must be tested for that stated purpose, not used to impose an unapproved ledger authenticity policy.

These are acceptance scenarios for affected application work, not a claim that documentation alone implements or verifies them. If existing tests encode a conflicting requirement, trace its product authority before preserving the assertion; report the conflict instead of adding machinery to satisfy an unsupported rule.

### Configuration and project portability

For workspace changes, exercise explicit per-path depth/budget (writer 0/300 versus repository containers 1/300), rejection of retired scalar/CLI/API aliases, ledger-only discovery and recorded identity through real moves. Include overlapping/physical aliases, duplicate ticket selection, an unavailable root with valid siblings, grouped/renamed/OpenSpec archive layouts and no recreated old parents. Ledger-free folders must not create dashboard rows or zero-effort placeholders.

Publication-state changes require writable operational state initialization without capture installation/registry access, read-only access with private state absent, durable private-journal/public-indicator ordering, real synchronized archive races, interrupted cleanup and pending copied history on another machine fixture. Missing recovery state must remain visibly incomplete; current-format completed history remains portable. Verify ticket/change links and independent roadmap/archive lifecycle. Native OpenSpec CLI/schema probes establish only that surface; actual Codex/OpenCode telemetry needs source-backed runtime evidence.

Exercise a custom workspace directory, custom ticket prefixes, source roots, thresholds, and explicit overrides. Missing required configuration must fail clearly; pre-existing container/parent directories must not silently select a project. Verify the effective value through the API and UI as well as the loader. Renaming ticket artifacts must not affect tests: use self-contained fixtures. Production clients must surface request failures; synthetic fixtures belong to tests.

Installer and entrypoint fixtures must explicitly isolate every writable cache, wrapper, state and configuration destination and assert that resolved targets stay under the fixture directory before applying changes. A temporary HOME alone is insufficient when explicit configuration selects destinations outside it. Use temporary shared/local install-root and destination settings through the production resolver. If a fixture escapes isolation, identify its exact generated artifact before cleanup; preserve unrelated or uncertain files and record the correction.

### Classification, anomalies, and exported prompts

Test the same source facts through live and retrospective paths. Successful inspection of text containing errors must not become a failed tool call. Keep failure streaks within their owning session across turns and bind displayed counts to the matching rule. Use existing domain enums. Derive prompt provenance from its actual source and apply configured identifier/sanitization rules rather than assuming AgenticLens ticket prefixes.

### Frontend embedding, CI, releases, or downloaders

After installing dependencies, use `make backend-check` for Go vet/race tests with embedded assets, or `make` for the full gate. Direct Go commands require `make web-assets` first. Frontend dependencies must exist; a skipped conditional frontend stage is not a pass.

For changes to build/CI paths, verify both an absent generated asset destination and a repeated build in a disposable checkout or with only generated ignored assets moved aside. Do not delete user files. Inspect `backend/internal/adapters/web/dist/` for index, JS, and CSS assets and compare with `web/dist/`. Run the embedded-asset test and validate the standalone binary. These generated trees are currently ignored; do not force-add them.

Trace every workflow that invokes Go tooling, including Release Please, through the shared asset prerequisite. A release tag does not prove downloadable binaries exist: verify platform names, uploaded assets, checksums, downloader URLs, and host execution. Reuse canonical Make targets rather than duplicating build recipes in CI docs.

### UI state, layout, and diagnostic copy

Keep filters and recovery controls usable when the result is empty or an optional request fails. Define no-selection semantics explicitly, including whether rows with missing tool/anomaly metadata remain visible. Test the empty-to-recovered transition as well as initial success.

Check adjacent layout behavior after CSS changes: tooltip visibility and positioning, wide-table containment, nested scroll regions, narrow viewports, and long identifiers. Keep a full identifier available to copy or inspect even when its visible label is truncated. Collapse changes must preserve keyboard access, counts, defaults, and controls needed to inspect available facts.

Preserve source evidence in exported diagnostic payloads. Separate generated explanatory text from verbatim input/output, reasoning, and error bodies. Apply only the approved redaction policy to source payloads; generic cleanup or identifier replacement can destroy the evidence needed for debugging. An unknown source or unavailable cost must stay explicitly unknown.

### Cleanup, compatibility, and source availability

Before environment cleanup, identify generated paths and preserve tracked source/tests and unrelated work. After disruption, compare against the known baseline and record which checks actually executed. Do not infer integrity solely from a green stage that skipped missing dependencies.

Distinguish declared support from observed compatibility. Tie version support to adapter/schema-family fixtures and runtime evidence, not an old plan's marketing claim. Separate initialization defaults written into configuration from runtime guesses for missing settings. Local per-harness source paths must not become shared machine-specific assumptions.

For migrations, distinguish structural validity, incomplete resumable workflow state, and successful finalization. Keep incomplete progress visible; do not impose a permanent migration marker on already-complete canonical data without an approved contract change. A rejected individual ledger must not erase valid sibling workspaces: test partial fleet results and per-ticket diagnostics separately from fatal directory/configuration failure.

### Polling or asynchronous updates

State how work grows with sessions, history size, poll frequency, and subscribers. Avoid unconditional full-history reconstruction on every unchanged tick. Verify idle and changed sessions, reconnects, cancellation, and child lookup errors. A failed child lookup must produce an error or explicit partial result, rather than a successful empty set.

## Adding, reviewing, or reducing tests

Apply the [canonical testing policy](../TECHNICAL_GUIDELINES.md#test-purpose-and-proportionality). For each candidate, record the application contract, plausible failure, actual production path, and assertions (including helpers and compiler checks). Inspect existing peers before adding or deleting coverage. Keep genuine adapter fixtures and distinct ownership/layer boundaries when they provide independent proof.

For reduction, record the source test and either its specific deletion rationale or the existing recipient for each meaningful assertion. Preserve defaults, overrides, absent values, accessibility, callbacks, counts, keyboard interactions, and failure paths when present. If a recipient would obscure the contract or cannot preserve its oracle, retain the independent test. A low score, short body, or passing run is insufficient evidence for removal.

When producing a suite inventory, use syntax-aware discovery that handles multiline signatures and parameterized declarations. Identify duplicate titles by path and line. Separate declarations, expanded runtime cases, helpers, and fixtures. Verify the actual removed set and permitted recipient changes against the baseline; report intentionally discarded assertions as well as transferred ones. Define fingerprint inputs if using hashes as evidence.

Run affected checks and the required integrated gate; report actual passed, failed, and skipped cases. Preserve compiler validation where type checks are the oracle. State review confidence and untested boundaries; do not claim mutation equivalence without evidence. For performance or line-ratio comparisons, record the measurement basis from the [metrics guidance](../TECHNICAL_GUIDELINES.md#interpreting-test-metrics).

## Independent code review

Use the [canonical review procedure](../.agents/skills/agentic-lens-review/SKILL.md) for correctness investigation and evidence outcomes, and its [parallel protocol](../.agents/skills/agentic-lens-review/references/parallel-review.md) for focus coverage, findings consolidation and revision-aware re-review. Reusing test runs does not replace inspecting the affected behavior. For consequential fixes, seek evidence that the regression is detected (before/after or a narrowly targeted disposable mutation); explain unavailable proof. Keep durable regressions in self-contained tests, never live ticket artifacts.

## Before handoff

For workspace consolidation, follow [completed review retention](../.agents/skills/agentic-lens-review/references/parallel-review.md#completed-review-retention). Verify that REVIEW_SUMMARY.json preserves every canonical candidate, numeric severity, disposition and all distinct reporting persona/model/effort tuples, including shared discovery and refutations. Retain scope/revision, initial/final outcomes (including no-findings, cancelled and incomplete work), separate verification actors, checks and limits in the change’s delivery.md. Resolve current references before deleting completed detailed reports and run/finding archives. Do not turn incomplete evidence into approval; historical full evidence may remain in Git history.

Record the tested commit/diff, command, result, skipped stages, and environment limits. Reference valid prior results for unchanged code. For each review finding, state fixed, still open, or out of scope with an evidence link; approval does not erase an untested branch. Keep final PR text aligned with delivered scope, and check whether a previous PR has already merged before attaching later work to it.

A documentation rule reduces ambiguity; it does not prove a defect is fixed. Runtime enforcement and regression tests belong to separately approved implementation work.

For lean occurrence and trace changes, reject retired fields even when empty/false and accept required explicit null role/type/time values. Verify complete-empty, partial and unavailable source trace responses without disk writes or saved-event fallbacks. Exercise explicit complete source events for sequence anomalies and retain recorded owner/group summaries when coverage is incomplete; tool metrics and financial evidence remain independent.

# Preventing recurring regressions

Use this checklist when a change touches the corresponding boundary. Put the selected cases and their results in the active iteration plan and review. Evidence and the complete ticket inventory are in the [REV-3 review](../alspaces/REV-3/iterations/01/REVIEW.md).

## Which document owns the rule?

- [AGENTS.md](../AGENTS.md): ticket ownership, approval, orchestration, telemetry, and delivery policies.
- [TECHNICAL_GUIDELINES.md](../TECHNICAL_GUIDELINES.md): architecture and data correctness requirements.
- [CONTRIBUTING.md](../CONTRIBUTING.md): executable local build and verification entry points.
- This checklist: select regression cases before implementation. Historical ticket artifacts record evidence; they are not current command references.

When changing a shared rule, search active instructions, personas, skill references, setup templates, and user docs for the old wording. Update affected guidance or link to its owner. Preserve historical records. Distinguish an artifact-allocation default from a runtime configuration requirement: AGENTS.md's fallback workspace name does not authorize the server to guess an unconfigured path.

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

### Configuration and project portability

Exercise a custom workspace directory, custom ticket prefixes, source roots, thresholds, and explicit overrides. Missing required configuration must fail clearly; pre-existing container/parent directories must not silently select a project. Verify the effective value through the API and UI as well as the loader. Renaming ticket artifacts must not affect tests: use self-contained fixtures. Production clients must surface request failures; synthetic fixtures belong to tests.

### Classification, anomalies, and exported prompts

Test the same source facts through live and retrospective paths. Successful inspection of text containing errors must not become a failed tool call. Keep failure streaks within their owning session across turns and bind displayed counts to the matching rule. Use existing domain enums. Derive prompt provenance from its actual source and apply configured identifier/sanitization rules rather than assuming AgenticLens ticket prefixes.

### Frontend embedding, CI, releases, or downloaders

After installing dependencies, use `make backend-check` for Go vet/race tests with embedded assets, or `make` for the full gate. Direct Go commands require `make web-assets` first. Frontend dependencies must exist; a skipped conditional frontend stage is not a pass.

For changes to build/CI paths, verify both an absent generated asset destination and a repeated build in a disposable checkout or with only generated ignored assets moved aside. Do not delete user files. Inspect `backend/internal/adapters/web/dist/` for index, JS, and CSS assets and compare with `web/dist/`. Run the embedded-asset test and validate the standalone binary. These generated trees are currently ignored; do not force-add them.

Trace every workflow that invokes Go tooling, including Release Please, through the shared asset prerequisite. A release tag does not prove downloadable binaries exist: verify platform names, uploaded assets, checksums, downloader URLs, and host execution. Reuse canonical Make targets rather than duplicating build recipes in CI docs.

### Polling or asynchronous updates

State how work grows with sessions, history size, poll frequency, and subscribers. Avoid unconditional full-history reconstruction on every unchanged tick. Verify idle and changed sessions, reconnects, cancellation, and child lookup errors. A failed child lookup must produce an error or explicit partial result, rather than a successful empty set.

## Before handoff

Record the tested commit/diff, command, result, skipped stages, and environment limits. Reference valid prior results for unchanged code. For each review finding, state fixed, still open, or out of scope with an evidence link; approval does not erase an untested branch. Keep final PR text aligned with delivered scope, and check whether a previous PR has already merged before attaching later work to it.

A documentation rule reduces ambiguity; it does not prove a defect is fixed. Runtime enforcement and regression tests belong to separately approved implementation work.

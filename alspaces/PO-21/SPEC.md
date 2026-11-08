# Feature Specification: Metric tags, filtering and comparison identity

## Problem statement

Engineers comparing agent runs need their own cohort dimensions, such as user, team, experiment or environment. Existing user attribution is a single built-in field, and comparison points currently distinguish completion status rather than arbitrary cohorts. Tags must be durable metric metadata independent of user identity.

## Personas and user stories

- As an engineer, I want project defaults, private local defaults and invocation overrides for key-value tags so I can label metrics without editing shared defaults for each run.
- As an analyst, I want to filter dashboard metrics by tag values so I can compare relevant cohorts.
- As an analyst, I want comparison points identified by a selected tag key's value, including missing values, so I can see cohort differences without losing untagged history.

## User requirements and product decisions

The latest user request supersedes the earlier proposal to override SessionFact.User. Tags are additive metadata, not a user-identity replacement. The user explicitly requests shared configuration, local configuration, CLI precedence, frontend filters defaulting to none, comparison coloring/identification by tag value, and continued validity/visibility of untagged points.

The following details make that request executable and are product defaults rather than previously observed implementation behavior:

1. Tags are a map of nonempty string keys to nonempty string values. Trim surrounding whitespace, preserve internal whitespace and case, and match exactly after trimming. Numeric, boolean, list and nested YAML values are invalid rather than silently coerced. Duplicate or whitespace-colliding keys within a YAML layer are errors.
2. Expose a top-level tags mapping in .agenticlens-config.yaml and .agenticlens-config.local.yaml. The local file is resolved beside the owning shared file using existing discovery rules. Expose repeatable generate --tag KEY=VALUE. Split at the first equals sign; values may contain equals signs. Repeated CLI keys use the last supplied value. Missing separators and blank keys/values fail clearly before publication.
3. Merge per key: CLI > local > shared. Keys present only in lower layers survive. Example: shared team=core, environment=dev; local environment=staging, user=alex; CLI environment=prod produces team=core, environment=prod, user=alex. Missing/empty maps add no overrides. No deletion syntax is included in this slice.
4. Persist the effective map as explicit metric metadata associated with each successfully generated/refreshed session owner, including children and both supported harnesses. This is a product ownership boundary, not a prescribed Go representation. Existing tags for that owner remain when a later refresh supplies no replacement for that key; supplied effective keys replace matching stored keys. No deletion or historical bulk relabeling is implied.
5. Preserve every owner outside the successful complete refresh set, with its tags and metrics. Shared/local configuration changes and dashboard reads never relabel historical owners. Backfill, optional hydration and supported migrations preserve stored tags; unavailable sources never strip them. Automatic publication receives shared/local defaults; manual CLI tags affect that invocation only and do not mutate files or registrations.
6. Existing ledgers with no tags remain valid and visible. Persisted tags remain readable after Git sharing to another machine without local configuration, original traces or capture state. Local config files remain private, but effective tag values are deliberately published in the shared ledger; documentation must explain this so users do not put secrets in tags.
7. Dashboard metric views provide tag key/value filters, initialized with no filter. OR values within a key; AND across keys; compose with existing filters. Include an explicit Missing value choice distinct from any literal tag text. Clear/reset restores all records including untagged ones. Derive options from the available dataset before tag filtering so users can recover from empty results.
8. Filter matching is evaluated on one owner: a session with team=core and another with environment=prod must not jointly satisfy team=core AND environment=prod. For session-derived metrics, recompute the displayed metrics from matching owners without duplicate counting. Ticket-level facts that cannot be partitioned by owner remain explicitly ticket-scoped; never invent a tag allocation. A ticket with no matching owner is excluded only when a filter is active.
9. In comparisons, offer an optional Color by tag selector. Default retains current completion-status coloring. Selecting a key assigns consistent distinguishable visual identities to its values, with legend and textual tooltip/accessibility labels. Keep mapping stable across filtering and sorting within the dataset. Coloring does not filter, duplicate points or change metric coordinates.
10. Comparison points may aggregate multiple owners. A point with one value across all contributors uses that value; no contributors carrying the key uses Missing; multiple values or a mixture of tagged/untagged contributors uses Mixed and exposes values plus missing membership in its tooltip. Missing and Mixed are reserved display states, not tag strings. Each ticket retains one point. Preserve completion status in accessible details when color represents a tag. Untagged points remain visible unless an explicit active tag filter excludes them; all existing metric-availability and manual-exclusion rules continue to apply.

## Layer scope and grounding

- Domain: typed tag values/ownership and pure merge, filter, aggregation and comparison-category semantics. Existing session-owned facts live in backend/internal/domain/models.go; domain configuration is in backend/internal/domain/config.go.
- Adapters/CLI: shared loader, restricted local overlay and YAML writer in backend/internal/adapters/config/loader.go and writer.go; generate flags and refresh/publication boundary in backend/cmd/agentic-lens/main.go. Local overlay currently allowlists fields and returns early without telemetry, so tags-only local files need explicit coverage.
- Persistence/API: boundary ledger DTO/validation and HTTP projections in backend/internal/adapters/fact_ledger_dto.go and backend/internal/adapters/http/{workspace_dto,comparison_dto,fleet_dto}.go. Expand supported persistence/wire formats compatibly; use an explicit migration only if required by the canonical schema policy. Ordinary reads do not rewrite history.
- Frontend: strict types in web/src/domain/types.ts, relevant API decoding, shared metric filters, web/src/views/TicketComparisons.tsx and web/src/components/TicketScatterPlot.tsx. The scatter plot currently has one point per ticket and colors by completion; preserve that point identity and existing interactions.
- Quality: extend coherent existing config/generation/portable-ledger tests and frontend filtering/comparison/scatter tests. No domain mocks; no tests dependent on live ticket artifacts.

## Architectural and compatibility constraints

Follow [TECHNICAL_GUIDELINES.md](../../TECHNICAL_GUIDELINES.md), especially its portable ledger contract, and [regression prevention](../../docs/REGRESSION_PREVENTION.md). Preserve FCIS, strict Go/TypeScript types, zero any, zero emojis, single-binary distribution, unavailable-value semantics and current numerical metrics when no tag filter is active. Do not create a machine-local authenticity gate for shared tags. Frontend delivery includes canonical build and verification of embedded assets and standalone binary.

## Acceptance evidence

Show config-only, local-only, CLI-only and conflicting three-layer cases, lower-layer key retention, duplicate CLI handling, YAML-sensitive values, tags-only local files and atomic validation errors. Show persisted/reloaded root and child tags for Codex and OpenCode and automatic publication. Verify two developers with different local tags can share one ledger; refreshing one available owner preserves unavailable and other-harness owners, and a receiver without sources/config can still filter the stored history. Verify absent-input refresh and backfill preserve tags, and repeated generation is idempotent.

Exercise frontend default no-filter behavior on tagged, untagged and legacy records; OR/AND and Missing filters; coherent filtered totals and reset; one-owner conjunction; empty results; stable legend identities; Missing/Mixed and literal labels with the same words; keyboard/tooltip identification; existing completion and point-selection interactions. Confirm coloring alone preserves point count and coordinates and keeps untagged points visible. Retain existing missing-metric diagnostics.

## Delivery roadmap

One vertical development ticket: [AL-82](../plan/to-do/AL-82%20add%20metric%20tags%20filters%20and%20comparison%20coloring.md), Ready, Blocked by: None, Blocks: None. It delivers configuration through persistence, API, filtering and comparison visualization together. Related AL-70 and AL-72 are coordination references, not prerequisites; recheck shared-file changes at implementation time. Explicit deletion/bulk retagging, tag management CRUD, numerical tag aggregation and live-source tag inference are out of scope.

Implementation must start in a separate agentic-lens-dev session. This PO session does not authorize writing application code or tests.

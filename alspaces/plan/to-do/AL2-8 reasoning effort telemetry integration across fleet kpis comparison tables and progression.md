h3. Title: AL2-8: Ingest Model Reasoning Effort Variants and Surface Deliberation Settings across Platform Views

*TARGET*: AgenticLens (domain | adapters | http | presentation | web | tests)

*Status*: Ready
*Blocked by*: None
*Blocks*: None

*CONTEXT*

When utilizing reasoning models (such as Gemini 2.5/3.8 Flash thinking, OpenAI o-series, and Claude thinking), operators configure deliberate reasoning intensity: `low`, `medium`, or `high`.
The original planning notes reported `variant` in the following OpenCode records. Treat these examples as historical evidence to validate separately against each supported schema family:
1. `session` table: `model` JSON column contains `{"id": "gemini-3.8-flash", "providerID": "litellm", "variant": "high"}`.
2. `message` table: `data` JSON column contains `{"providerID": "litellm", "modelID": "gemini-3.8-flash", "variant": "high"}`.

Original planning observations (not revalidated against current source versions) recorded: `high` (over 1,700 messages), `medium` (89 messages), and `low` (1 message). Currently, AgenticLens only extracts the model identifier (`modelID` / `id`) while discarding `variant`. Consequently, platform operators cannot discern whether a session or subagent run operated under low, medium, or high reasoning budgets, making it impossible to evaluate the cost, latency, and quality trade-offs between different reasoning tiers.

*GOAL*

Ingest model reasoning effort variants (`low`, `medium`, `high`) from OpenCode session and message records into AgenticLens domain facts, and surface the configured reasoning effort alongside model identifiers and metrics across the Deep Dive, Live Trace Inspector, Model Breakdown tables, and Fleet analytics.

*VERTICAL SLICE SCOPE*

* Adapters & Ingestion (`internal/adapters/opencode`):
  - Extend the owning schema-specific decoder to capture the verified native setting location; resolve the historical top-level versus nested message examples from source evidence.
  - Preserve the native configured effort value, including values outside low/medium/high. Missing/default settings remain explicitly unspecified; never infer configuration from reasoning-token counts.
  - Attach `reasoning_effort` to extracted `InvocationFact` and `SessionFact` models.
* Domain Core (`internal/domain`):
  - Add `ReasoningEffort string` (or `ReasoningVariant string`) to `InvocationFact`, `SessionFact`, and `ModelFact`.
  - In `ModelAggregates`, support distinguishing models by variant or reporting reasoning effort per model entry (e.g. `gemini-3.8-flash (high)`).
  - Ensure pure aggregation functions compute model stats deterministically with zero I/O.
* HTTP Boundary & DTOs (`internal/adapters/http`, `internal/adapters/schemas.go`):
  - Expose `reasoning_effort` on invocation and model breakdown DTOs in `/api/workspaces/{ticketId}` and `/api/fleet`.
* Web Presentation & Views (`web/`):
  - `ModelBreakdownTable`: Add a "Reasoning Effort" pill badge (`HIGH`, `MED`, `LOW`) beside the model name or in a dedicated column.
  - `TicketDeepDive`: Display the dominant reasoning effort tier in the Model Profile Card and session timeline nodes.
  - `LiveTraceInspector`: Display a reasoning effort chip (e.g. `Effort: high`) in message headers alongside token counts.
  - `FleetOverview`: Provide model distribution filtering or grouped views by reasoning effort tier.
* Quality:
  - Table-driven Go unit tests in `internal/adapters/opencode/extractor_test.go` verifying extraction of `variant: "high"`, `"medium"`, `"low"`, `"default"`, and omitted cases.
  - Vitest component tests in `ModelBreakdownTable.test.tsx` and `LiveTraceInspector.test.tsx` verifying reasoning effort pill badges render accurately.
  - Pass `make`.

*ACCEPTANCE CRITERIA*

* OpenCode v1/v2 extraction reads the setting from the source-backed locations established by the per-family investigation, with deterministic fixtures proving top-level/nested and missing-field behavior where supported.
* Source-backed reasoning effort values are attached to their owning invocation/session without collapsing unknown native settings. Validate OpenCode v1/v2 locations separately, preserve historical facts, and expose unsupported or absent settings honestly.
* Model breakdown tables in Deep Dive and Fleet views render distinct reasoning effort badges or labels for models configured with variants.
* Missing or default variants are displayed gracefully without empty badges or layout shifts.
* Zero regressions in existing token counting or cost calculations.
* Zero `any` in Go and TypeScript; all tests pass with `make`.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make`).
* Go and TypeScript type checking clean with zero any.
* Verified end-to-end in embedded dashboard.

*DEVELOPMENT STATE REVIEW (2026-11-05)*

Not delivered: reasoning token metrics exist, but no configured reasoning_effort/variant fact or display contract was found. Historical OpenCode sample counts are original observations, not a current schema guarantee; validate v1/v2 independently and preserve native unknown values.

Evidence and review limits: [PO-19 reconciliation](../../PO-19/SPEC.md), merged baseline `03f5b05d`. This planning review does not certify runtime behavior or close unverified acceptance criteria.

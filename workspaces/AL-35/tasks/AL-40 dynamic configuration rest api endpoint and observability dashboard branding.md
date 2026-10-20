h3. Title: AL-40: Dynamic Configuration REST API Endpoint and Observability Dashboard Branding

*TARGET*: AgenticLens (adapters | http | presentation | web | domain | tests)

*CONTEXT*

The pre-release source code audit (AL-35) identified that the Go HTTP server exposes no route for project configuration. As a consequence, the web dashboard has no awareness of `ProjectConfig.ProjectName`, `Ticket.Prefixes`, or `Workspaces.Dir`. The frontend UI statically hardcodes "AgenticLens Telemetry", providing no mechanism to display the observed project's name or brand context. Additionally, multi-select prefix dropdowns across Fleet Overview, Ticket Comparisons, and Metric Progression rely solely on runtime ledger scraping (`extractDistinctPrefixes`), meaning configured prefixes for new or planned tickets do not appear in filters until at least one completed ledger has been committed.

*GOAL*

Expose a new `GET /api/config` REST endpoint in the Go HTTP server returning sanitized project configuration parameters, implement frontend API client methods and state hydration, dynamically render the observed project name in the navigation bar alongside permanent `AgenticLens` platform branding, and populate prefix dropdowns using both configured and discovered prefixes.

*VERTICAL SLICE SCOPE*

* HTTP Server Endpoint (`backend/internal/adapters/http/server.go`, `config_handler.go`):
  - Register `GET /api/config` returning sanitized configuration JSON:
    - `project_name`: string
    - `prefixes`: list of structured ticket prefixes (`prefix`, `description`, `ticket_type`)
    - `workspaces_dir`: string
    - `role_order`: list of strings
    - `version`: version string
* Frontend API Client & Types (`web/src/api/client.ts`, `web/src/domain/types.ts`):
  - Define `ProjectConfigResponse` interface in `web/src/domain/types.ts`.
  - Implement `fetchProjectConfig()` in `web/src/api/client.ts` with mock fallback in `web/src/api/mockData.ts`.
* Navigation Bar & Branding Component (`web/src/components/Navbar.tsx`):
  - Preserve `AgenticLens` permanently as the core platform brand.
  - Render observed project name as a distinct project badge/subtitle next to the platform brand when non-empty.
  - Update `web/src/views/LandingPage.tsx` to reference observed project context where appropriate.
* Prefix Filter Population:
  - Union configured prefixes from `/api/config` with discovered prefixes from loaded ledgers so configured prefixes are immediately available for filtering.
* Quality & Verification:
  - Go HTTP handler unit tests asserting 200 OK and valid JSON response for `GET /api/config`.
  - Frontend Vitest component tests verifying Navbar renders project badge when provided and gracefully falls back when omitted.
  - Zero emojis across all UI elements and responses.

*ACCEPTANCE CRITERIA*

* `GET /api/config` endpoint implemented and verified.
* Frontend fetches project configuration on startup.
* Navigation bar renders observed project name alongside permanent `AgenticLens` branding.
* Ticket prefix filters display configured prefixes in addition to discovered ledger prefixes.
* Quality gates pass (`make check`, `make test`, `make build`) with zero any and zero emojis.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, `make build`).
* Zero any and zero emojis across all files.

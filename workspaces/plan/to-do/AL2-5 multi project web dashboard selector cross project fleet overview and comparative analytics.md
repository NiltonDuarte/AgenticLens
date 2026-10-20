h3. Title: AL2-5: Multi-Project Web Dashboard Selector, Cross-Project Fleet Overview, and Comparative Analytics

*TARGET*: AgenticLens (presentation | web | components | state | tests)

*CONTEXT*
 
With backend multi-project ingestion and REST endpoints in place (AL2-3, AL2-4), web dashboard users need an intuitive visual interface to view, filter, and compare data across projects. Project managers and leads must be able to switch between individual projects or select multiple projects simultaneously to evaluate cross-project agent velocity, compare financial expenditures, and inspect multi-repo health in a single dashboard view.

*GOAL*
 
Deliver a comprehensive multi-project presentation suite in the TypeScript web frontend, including a global project multi-select dropdown in `Navbar.tsx`, cross-project KPI rollups and comparison charts in `FleetOverview.tsx`, project badges in `WorkspacesTable.tsx`, and project context in `TicketDeepDive.tsx`.

*VERTICAL SLICE SCOPE*
 
* Domain Core: Define strict TypeScript interfaces for `ProjectInfo`, `ProjectSyncResponse`, and `MultiProjectSummary` in `web/src/types/telemetry.ts` with zero `any`.
* Boundary Adapter: Add `fetchProjects(): Promise<ProjectInfo[]>` and `triggerProjectsSync(): Promise<ProjectSyncResponse>` in `web/src/api/client.ts`. Update existing API methods (`fetchWorkspaces`, `fetchFleetSummary`, `fetchComparison`) to pass selected project IDs in query parameters.
* Presentation: Build `ProjectSelector.tsx` dropdown in `Navbar.tsx` supporting "All Projects" and multi-project checkbox selection, synchronizing state with URL search params (`?projects=core,billing`) and local storage. Add project origin badge column in `WorkspacesTable.tsx`. Enhance `FleetOverview.tsx` to display cross-project comparison cards, token and cost rollups per project, and sync status indicators. Update `TicketDeepDive.tsx` breadcrumbs and header to display project name.
* Quality: TypeScript unit tests in `web/src/components/ProjectSelector.test.tsx`, `WorkspacesTable.test.tsx`, and `FleetOverview.test.tsx` verifying project switching, URL sync, multi-select filtering, and sync button interactions.

*ACCEPTANCE CRITERIA*
 
* `Navbar.tsx` features a project selector displaying all configured projects fetched from `GET /api/projects`.
* Selecting one or more projects updates the active view to display only telemetry from those projects.
* Selecting "All Projects" displays aggregated fleet metrics across all monitored repositories.
* Project selection persists across page refreshes and updates URL parameters (`?projects=...`).
* Workspaces table clearly displays project badges on each ticket row when multi-project mode is active.
* Fleet overview provides side-by-side comparative analytics for cost, velocity, and anomaly counts per project.
* Strict TypeScript typing with zero `any`; `npm run typecheck` and `npm test` pass with zero failures.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.

# Implementation Plan: AL-4 Iteration 02

Delta Scope: Dedicated Server Lifecycle Target, Port Resolution, and Observability Landing Portal

Target Workspace File: workspaces/AL-4/iterations/02/PLAN.md

---

## 1. Goal and Iteration Scope

This iteration responds directly to user steering feedback following the initial rollout of the ticket deep dive, progression, and comparison views:
1. Makefile Server Lifecycle Target: Introduce explicit make targets (server and serve) that enforce build prerequisites (web distribution bundle and Go binary) before executing the single binary with a configurable port.
2. Server Port Environment and Conflict Resolution: Upgrade the server CLI entrypoint in cmd/agentic-lens/main.go to respect the PORT environment variable as a fallback default when the --port flag is omitted, validate port bounds, and surface clear diagnostic messages when a port is already occupied or blocked.
3. Observability Landing Portal: Introduce a dedicated entry portal view (web/src/views/LandingPage.tsx) featuring a system observability hero introduction, interactive feature navigation cards for all four core analysis views, fleet pulse metrics, and a quick-launch workspace directory.
4. Top Navigation Brand Link: Configure the brand logo and title in web/src/components/Navbar.tsx as an interactive control returning the user to the landing portal from any view.
5. Application Shell Default Route: Update web/src/App.tsx to make the landing page the default view on launch, wiring bi-directional routing between the landing page, navbar, and core analysis views.

In strict accordance with repository guidelines:
- Zero code blocks are contained within this plan artifact.
- Zero emojis are permitted in code, markup, logs, or workspace documentation.
- Zero any types across Go backend and TypeScript frontend codebases.

---

## 2. Grounded Reference Architecture and Symbol Map

### Backend Entrypoint and Build Automation
- Root Automation: Makefile
  - Target build: compiles web distribution bundle and Go executable bin/agentic-lens.
  - New target server: depends on target build, accepts configurable PORT variable (defaulting to 8080) and WORKSPACES directory variable (defaulting to workspaces), and executes the binary server subcommand.
  - New target serve: alias pointing to target server.
- CLI Server Subcommand: cmd/agentic-lens/main.go
  - Function runServer: parses CLI flags, inspects os.Getenv for PORT fallback, validates port numerical bounds between 1 and 65535, checks for net.OpError and address-already-in-use conditions on srv.ListenAndServe, and logs actionable troubleshooting guidance.
  - Function main: routes subcommand invocation to runServer or CLI reporting workflow.
- CLI Automated Tests: cmd/agentic-lens/main_test.go
  - Integration suite verifying flag parsing, environment variable resolution, port collision error paths, and server health endpoints.

### Frontend Domain, Boundary, and Presentation Architecture
- Domain Types: web/src/domain/types.ts
  - Type ActiveTab: expanded union containing landing, overview, deep-dive, progression, and comparisons.
  - Interface FleetSummary: provides fleet-wide spend, invocations, token counts, and durations.
  - Interface TicketComparisonRow: provides ticket identity, title, context status, and efficiency metrics for quick launch.
- Boundary API Client: web/src/api/client.ts
  - Function fetchFleetSummary: retrieves aggregate fleet metrics with deterministic mock fallback.
  - Function fetchWorkspaces: retrieves workspace comparison rows with deterministic mock fallback.
- Application Shell and Navigation:
  - web/src/App.tsx: owns activeTab state defaulting to landing, selectedTicketId state, and handles navigation between landing, views, and ticket drill-downs.
  - web/src/components/Navbar.tsx: renders interactive brand logo and title button navigating back to landing, alongside tab buttons and connection status pill.
- Landing View and Styling:
  - web/src/views/LandingPage.tsx: renders hero section, 4 interactive navigation feature cards, summary pulse metrics, and quick-launch workspace table.
  - web/src/index.css: declares responsive styles for landing hero layout, feature navigation cards, pulse metric badges, and quick-launch rows under the unified dark theme.
- Automated Test Suites:
  - web/src/views/LandingPage.test.tsx: verifies hero rendering, feature card click routing, pulse metric display, quick-launch selection, error and loading states, and emoji absence.
  - web/src/components/Navbar.test.tsx: verifies brand click callback invocation and tab active state transitions.

---

## 3. Phased Implementation Tasks for Coder Subagent

### Phase 1: Build System and Server Lifecycle Entrypoint
- Task 1.1: Update Root Makefile
  - Declare variable PORT defaulting to 8080 using conditional assignment syntax.
  - Declare variable WORKSPACES defaulting to workspaces using conditional assignment syntax.
  - Add server and serve to the .PHONY declaration list.
  - Add target server depending explicitly on target build. In its recipe, execute the compiled binary bin/agentic-lens with arguments server --port $(PORT) --workspaces $(WORKSPACES).
  - Add target serve pointing to target server.
  - Update target help output to document server and serve targets and their configurable variables.
- Task 1.2: Enhance Server Flag and Environment Resolution in cmd/agentic-lens/main.go
  - In function runServer, check os.Getenv for variable PORT before instantiating the FlagSet.
  - If the environment variable is present and parses cleanly to an integer between 1 and 65535, set it as the default port value; otherwise default to 8080.
  - Parse CLI flags from args; explicit --port arguments will take precedence over both the environment variable and the default.
  - Validate the resolved port number: if less than 1 or greater than 65535, write an explicit error to os.Stderr indicating invalid port range and terminate with exit code 1.
  - In the server execution goroutine, wrap error inspection for srv.ListenAndServe:
    - If the error indicates that the address is already in use (such as syscall.EADDRINUSE or containing the phrase bind: address already in use), output a descriptive diagnostic to os.Stderr explaining that the chosen port is already occupied and advise specifying an alternate port via the --port flag or PORT environment variable.
    - Terminate with exit code 1 on non-closed server errors.

### Phase 2: Domain Types and Navbar Interactivity
- Task 2.1: Extend ActiveTab Union in web/src/domain/types.ts
  - Update the ActiveTab type definition to include landing as a valid tab literal: union of landing, overview, deep-dive, progression, and comparisons.
- Task 2.2: Make Brand Header Interactive in web/src/components/Navbar.tsx
  - Transform the navbar brand wrapper into an accessible interactive element (button or div with role button and tabIndex 0).
  - Attach an onClick handler and onKeyDown handler (supporting Enter and Space keys) that invokes onSelectTab with landing.
  - Add an aria-label identifying the control as the home navigation link returning to the landing page.
  - Ensure the four primary tool tabs (Fleet Overview, Ticket Deep Dive, Metric Progression, Comparisons) continue to reflect active styling when selected, while none appear highlighted when activeTab is landing.

### Phase 3: Observability Landing Portal Component and Styles
- Task 3.1: Create web/src/views/LandingPage.tsx
  - Accept typed props LandingPageProps with readonly callbacks onNavigate (accepting ActiveTab), onSelectTicket (accepting ticket ID string), and optional onMockChange (accepting boolean).
  - Manage loading, error message, summary, and workspaces states through standard React hooks.
  - On mount, invoke fetchFleetSummary and fetchWorkspaces in parallel, updating summary and workspace lists while informing onMockChange of mock fallback status.
  - Hero Section:
    - Display an overarching platform title and subtitle introducing AgenticLens as an autonomous multi-agent observability and telemetry platform.
    - Present architectural overview prose explaining Functional Core, Imperative Shell principles, context window saturation tracking, token unit economics, and multi-agent SDLC verification.
  - Pulse Metrics Bar:
    - Render four summary pulse metrics: Total Ticket Workspaces, Cumulative Fleet Spend in USD, Total SDLC Iterations across workspaces, and Prompt Cache Hit Efficiency.
    - Defensively calculate and format all metrics with fallback zero values to prevent division by zero or NaN rendering.
  - Primary Tools Navigation Grid:
    - Render 4 structured interactive cards corresponding to the core capabilities:
      - Fleet Overview: macro-level spend, total invocations, token throughput, and role/model distributions. Clicking calls onNavigate('overview').
      - Ticket Deep Dive: single-ticket autopsy, context ceiling saturation, turn velocity, and artifact reader. Clicking calls onNavigate('deep-dive').
      - Metric Progression: temporal velocity trends, daily/weekly/cumulative rollups, and code churn. Clicking calls onNavigate('progression').
      - Ticket Comparisons: cross-ticket benchmarking matrix, test-to-prod ratios, and duration sorting. Clicking calls onNavigate('comparisons').
    - Each card includes a badge, header, descriptive body text, and an explicit action button.
  - Quick-Launch Workspace Directory:
    - Render a table or card list showing discovered ticket workspaces (ticket ID, title, status, context health pill, iterations count, and active duration).
    - Provide a quick-launch action button per row that triggers onSelectTicket with the workspace ticket ID, opening Ticket Deep Dive directly for that ticket.
  - Error and Loading Boundaries:
    - Render a clear loading indicator while API calls are in flight.
    - Render an error banner with a retry button when data loading fails.
    - Render an empty state notice when no workspaces are discovered.
  - Strictly verify zero emojis across all titles, descriptions, badges, and button labels.
- Task 3.2: Add Landing Page Styles to web/src/index.css
  - Add CSS classes for .landing-container, .landing-hero, .hero-title, .hero-lead, and .hero-tags.
  - Add CSS classes for .pulse-metrics-grid, .pulse-card, .pulse-value, and .pulse-label.
  - Add CSS classes for .feature-cards-grid, .feature-nav-card, .feature-card-header, .feature-card-title, .feature-card-desc, and .feature-card-btn.
  - Add CSS classes for .quick-launch-section, .quick-launch-table, and .quick-launch-row.
  - Add interactive hover, focus, and pointer styling to .navbar-brand.
  - Maintain dark background variables, clear typography hierarchy, and monospace formatting for numbers and tokens.

### Phase 4: Application Shell Wiring
- Task 4.1: Update web/src/App.tsx
  - Change default state of activeTab from overview to landing.
  - Import LandingPage from ./views/LandingPage.
  - Add conditional rendering for activeTab === 'landing', passing onNavigate={(tab) => setActiveTab(tab)}, onSelectTicket={handleSelectTicketFromList}, and onMockChange={setIsMock}.
  - Retain ticket selection synchronization so selecting a ticket from the landing quick-launch switches activeTab to deep-dive with selectedTicketId populated.

---

## 4. Phased Test Suite Tasks for Tester Subagent

### Phase 5: Backend Automated Verification
- Task 5.1: Extend cmd/agentic-lens/main_test.go
  - Test server subcommand respects the PORT environment variable when the --port flag is omitted.
  - Test server subcommand allows CLI --port flag to take precedence when both PORT environment variable and --port flag are supplied.
  - Test server subcommand fails immediately with a descriptive error when an invalid port (e.g. 0 or 70000) is specified.
  - Test server subcommand surfaces clear address-already-in-use diagnostic messaging when attempting to bind to an occupied port.

### Phase 6: Frontend Automated Verification
- Task 6.1: Create web/src/views/LandingPage.test.tsx
  - Test that the landing page renders the hero banner with introduction and system observability descriptions.
  - Test that all four feature navigation cards are rendered with proper titles and action buttons.
  - Test that clicking each feature navigation card triggers onNavigate with the expected tab identifier.
  - Test that summary pulse metrics are correctly derived from mock/live data and displayed without NaN or null.
  - Test that the quick-launch workspace list renders discovered tickets and clicking a ticket row triggers onSelectTicket with the appropriate ticket ID.
  - Test that error states render an error banner with a functioning retry button.
  - Test that an empty workspace response renders an empty state notice.
  - Assert that zero emojis exist in the rendered DOM text of the landing page.
- Task 6.2: Update web/src/components/Navbar.test.tsx
  - Test that clicking the brand title or icon invokes onSelectTab with landing.
  - Test that pressing Enter or Space on the brand button invokes onSelectTab with landing.
  - Test that when activeTab is landing, none of the four primary tool tabs possess the active class.
  - Assert that zero emojis exist across navbar markup and status indicators.
- Task 6.3: Verification of Full Test Suite
  - Run npm run typecheck in web/ to verify zero TypeScript compiler errors.
  - Run npm run lint in web/ to verify zero linter warnings.
  - Run npm test in web/ to verify all existing and new frontend tests pass without skips.
  - Run make check and make test in repository root to verify Go vetting, Go tests with race detection, and frontend suites.

---

## 5. Build Verification and Single-Binary Parity

- Task 7.1: Verify Single-Binary Packaging
  - Execute make build and verify that web/dist compiles and is embedded into bin/agentic-lens.
  - Execute make server with a custom port (e.g. PORT=18080) in the background to verify the server starts and serves the landing page on root HTTP GET /.
  - Verify that curl http://localhost:18080/ returns the HTML document containing the SPA root element.
  - Verify that API endpoints (/api/fleet and /api/workspaces) return valid JSON under the server target.
- Task 7.2: Telemetry Ledger Recording
  - Execute scripts/extract_agentic_metrics.py to update workspaces/AL-4/agentic_metrics.json with iteration 02 telemetry facts.

---

## 6. Key Defensive Mitigations

- Defensive Port Handling: Both PORT environment variable and CLI flags undergo integer parsing and range verification (1 to 65535) before socket binding is attempted.
- Clear Port Conflict Guidance: When a port collision occurs, the process outputs actionable instructions pointing to --port and PORT rather than terminating with an unhelpful raw socket error.
- Division by Zero Defenses: Pulse metric calculations on the landing page defend against zero ticket counts, zero invocations, and zero cache reads.
- Dual-Mode Resilience: If the Go backend server is unreachable or offline during frontend development, LandingPage automatically falls back to deterministic mock fixtures while alerting the UI via onMockChange.
- Strict Zero-Emoji and Zero-Any Compliance: Automated linting, TypeScript compiler flags, and Vitest regex assertions enforce repository standards.

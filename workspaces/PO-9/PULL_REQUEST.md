# docs(plan): v2 platform architecture, debug prompt generator, multi-project observability, and archive engine (PO-9)

## Summary
Defines the product specification, architectural contracts, and 5 vertical tracer-bullet Jira tickets (AL2-1 through AL2-5) for AgenticLens v2 Phase 2:
1. Automated agent debug prompt generator synthesizing anomaly forensics, tool invocation parameters, and stderr traces into structured, ready-to-run markdown prompts for LLM coding agents, complete with CLI and interactive web drawer (AL2-1).
2. Workspace archive command (`agentic-lens archive`) with date-based cutoff and interactive TTY selection, packaging older workspace tickets into verified `.tar.gz` archives with folder cleanup, paired with transparent in-memory archive ingestion across the server scanner and web UI (AL2-2).
3. Multi-project observability allowing a single server to ingest and monitor multiple local repositories, folders, and remote git URLs with per-project configuration loading, multi-project domain aggregations, and web dashboard project selection (AL2-3, AL2-4, AL2-5).

## Initiatives Covered

1. **Agent Anomaly & Tool Error Debug Prompt Generator (AL2-1)**:
   - Pure domain prompt synthesis engine formatting error diagnostics, tool input parameters, stack traces, and causal turn sequences into structured markdown prompts.
   - Rule-specific remediation directives (`TOOL_RETRY_STREAK`, `TOOL_HIGH_FAILURE_RATE`, `CONTEXT_SATURATION_SPIKE`, `SUBAGENT_REWORK_LOOP`).
   - REST endpoint `GET /api/workspaces/{ticketId}/debug-prompt` and CLI subcommand `agentic-lens prompt`.
   - Web frontend integration in `AnomalyDrawer.tsx`, Waterfall inspector error nodes, and `LiveAnomalyBanner.tsx` with one-click clipboard copying.

2. **Workspace Archive Command & Transparent In-Memory Ingestion (AL2-2)**:
   - CLI subcommand `agentic-lens archive` supporting `--before <date>`, `--workspaces`, `--keep`, and `--dry-run`.
   - Interactive terminal selection menu when executed without date arguments in a TTY.
   - Atomic `.tar.gz` archive packaging with `manifest.json`, checksum verification, and safe directory cleanup.
   - Transparent in-memory archive reading in `ScanWorkspaces`, `LoadTicketWorkspace`, and `ReadTicketArtifacts` without disk decompression.
   - Archived ticket badges and state filtering in the web dashboard.

3. **Multi-Project Workspace Observability & Remote Git Ingestion (AL2-3, AL2-4, AL2-5)**:
   - Multi-project configuration manifest schema (`projects:` block in `.agenticlens-config.yaml`) and CLI flags (`--project <id>=<path_or_url>`).
   - Per-project configuration loading: dynamically reads each project's ticket prefixes, role mappings, and anomaly thresholds.
   - Remote git repository cache manager in `~/.cache/agenticlens/remotes/` performing shallow clones (`--depth 1`) and on-demand synchronization.
   - HTTP endpoints `GET /api/projects` and `POST /api/projects/sync`, plus multi-project query filtering across all telemetry endpoints (`?project=p1,p2`).
   - Global project selector in web dashboard `Navbar.tsx`, cross-project fleet analytics, and project badges on ticket tables and detail views.

## Ticket Breakdown
- Ready: `AL2-1: End-to-End Agent Anomaly Debug Prompt Generator, CLI, and Interactive Web Drawer` (Phase 8, Tier 1 - Ready, unblocked)
- Ready: `AL2-2: End-to-End Workspace Archive Engine, CLI Compaction, and In-Memory Decompression Ingestion` (Phase 8, Tier 1 - Ready, unblocked)
- Ready: `AL2-3: Multi-Project Configuration Schema, Domain Aggregations, and Multi-Folder Scanner` (Phase 8, Tier 1 - Ready, unblocked, blocks AL2-4, AL2-5)
- Pending: `AL2-4: Remote Git Repository Ingestion, Shallow Cache Manager, and Project Management API` (Phase 8, Tier 2 - Pending, blocked by AL2-3, blocks AL2-5)
- Pending: `AL2-5: Multi-Project Web Dashboard Selector, Cross-Project Fleet Overview, and Comparative Analytics` (Phase 8, Tier 3 - Pending, blocked by AL2-3, AL2-4)

## Pointers
- Specification: `workspaces/PO-9/SPEC.md`
- Shared Roadmap: `workspaces/plan/TICKETS.md`
- Development Tickets:
  - `workspaces/plan/to-do/AL2-1 end-to-end agent anomaly debug prompt generator cli and interactive web drawer.md`
  - `workspaces/plan/to-do/AL2-2 end-to-end workspace archive engine cli compaction and in-memory decompression ingestion.md`
  - `workspaces/plan/to-do/AL2-3 multi project configuration schema domain aggregations and multi folder scanner.md`
  - `workspaces/plan/to-do/AL2-4 remote git repository ingestion shallow cache manager and project management api.md`
  - `workspaces/plan/to-do/AL2-5 multi project web dashboard selector cross project fleet overview and comparative analytics.md`
- Session Summary: `workspaces/PO-9/SUMMARY.md`

# GEN-22 iteration 03

Source: the user asks to move artifact guidance out of docs and define artifacts
per skill. This supersedes the standalone workspace writing standard introduced
in iteration 01.

Acceptance delta: shared prose/evidence rules stay briefly in AGENTS.md;
development, PO, management and review skills own their output definitions.
Remove docs/WORKSPACE_ARTIFACTS.md and update its callers. Keep the goal-first PO
format, concise current summaries and all evidence/telemetry protections.

1. Integrate relevant guidance into each existing skill, avoiding new guide files.
2. Validate local links/anchors, skill metadata and absence of stale references.
3. Commit, integrate main and run the required make gate; refresh final telemetry
   and update the open GEN-22 PR.

Baseline: `abbddc8b`; initial working tree clean. The documentation coordinator
owns these paths. Direct consistency review remains proportionate; application
code, historical tickets and metric formatting are unchanged.

# docs(po): specify machine-wide worktree capture and delivery tickets (PO-20)

Define automatic AgenticLens capture for base checkouts and Git worktrees through a single user-level installation in Codex or OpenCode. Shared configuration supplies project/metrics rules and executable selection; the selected local configuration controls machine-specific capture settings.

The specification covers whole-file local configuration fallback, native session routing, pinned/development binary selection, ticket binding, background publication, bounded recovery and state retention. Portable ticket ledgers remain usable without the original machine's private capture state.

Create two vertical delivery tickets:

- **AL-79 — Ready:** shared native runtime and end-to-end user-level Codex capture.
- **AL-80 — Blocked by AL-79:** global OpenCode integration using the shared runtime.

Add **AL-81 — Ready** for ad hoc manual generation independent of the AgenticLens capture registry, preserving source validation, portable ledgers and automatic ownership safeguards. The roadmap records statuses and dependencies. No application, test or runtime implementation is added; development belongs in separate agentic-lens-dev sessions. The shared development pin selects the locally built version; machine-local integrations and capture state remain outside Git.

Validation: specification links, unique ticket entries and reciprocal dependencies checked. Application code is unchanged; the prior full `make` result is reused. Live automatic capture remains unverified. Native CLI generation refreshed the Codex root and planner-child telemetry after a user-requested capture-registry reset. `TELEMETRY_STATUS.json` records the snapshot cutoff; successful manual extraction does not establish live automatic capture coverage.

See [SPEC.md](alspaces/PO-20/SPEC.md) and [SUMMARY.md](alspaces/PO-20/SUMMARY.md).

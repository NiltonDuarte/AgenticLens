# docs(po): specify machine-wide worktree capture and two delivery tickets (PO-20)

Define automatic AgenticLens capture for base checkouts and Git worktrees through a single user-level installation in Codex or OpenCode. Shared configuration supplies project/metrics rules and executable selection; the selected local configuration controls machine-specific capture settings.

The specification covers whole-file local configuration fallback, native session routing, pinned/development binary selection, ticket binding, background publication, bounded recovery and state retention. Portable ticket ledgers remain usable without the original machine's private capture state.

Create two vertical delivery tickets:

- **AL-79 — Ready:** shared native runtime and end-to-end user-level Codex capture.
- **AL-80 — Blocked by AL-79:** global OpenCode integration using the shared runtime.

The roadmap records their dependency. No application, test or runtime implementation is added; development belongs in separate agentic-lens-dev sessions. The shared development pin selects the locally built version; machine-local integrations and capture state remain outside Git.

Validation: specification links, unique ticket entries and reciprocal dependencies checked. Application code is unchanged; the prior full `make` result is reused. Live automatic capture remains unverified. Telemetry retains an earlier native snapshot because a session ownership conflict prevents fresh extraction; `TELEMETRY_STATUS.json` records the incomplete coverage.

See [SPEC.md](alspaces/PO-20/SPEC.md) and [SUMMARY.md](alspaces/PO-20/SUMMARY.md).

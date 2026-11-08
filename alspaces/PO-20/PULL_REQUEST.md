# docs(po): specify machine-wide worktree capture and two delivery tickets (PO-20)

Reclassifies the same capture initiative from GEN-22 to PO-20 and moves its workspace without creating developer workspaces or contracts. The closed user decision review is organized into SPEC.md and retained discussion notes.

Creates the minimum practical two vertical deliveries: AL-79 provides the shared configuration/native runtime and end-to-end user-level Codex capture; AL-80 adds global OpenCode integration using that runtime. The roadmap records AL-79 as Ready and AL-80 as blocked by AL-79. Existing tickets remain unchanged.

The specification covers whole-file local fallback, user-level adoption, native session routing, shared pinned/development binary selection, explicit binding, automatic publication, recovery and retention while preserving portable ledgers. No application, test or runtime implementation is added.

This draft replaces [#124](https://github.com/NiltonDuarte/AgenticLens/pull/124), whose head remains the former GEN branch. The earlier user-approved shared development pin is retained; machine-local hooks, Python launcher, trust and capture state remain outside Git. Runtime capture remains unverified.

See alspaces/PO-20/SPEC.md and SUMMARY.md. Implementation belongs in separate agentic-lens-dev sessions.

Fresh PO-20 telemetry extraction is blocked by the session's existing GEN-22 capture binding. The retained native snapshot is reclassified only in ticket metadata; TELEMETRY_STATUS.json records its age and incomplete refresh. No session facts or capture bindings were fabricated or changed.

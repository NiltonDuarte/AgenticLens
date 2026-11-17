# FIX-18 — Approved technical contract

## Selection and observable behavior

At baseline `b6ab0471`, `runServer` resolves the configured OpenCode path or explicit override, opens a read-only hook, and starts the service after a successful connection. Its startup banner announces the HTTP listener. `CheckAvailability` checks hook presence; tree and trace SQL read v1 tables. The stream commits HTTP 200 before its first snapshot and discards a failed snapshot. Polling drops failed tree reads and skips failed trace reads. The browser discards REST error bodies and can replace the source reason with a generic connection banner. [Observable acceptance](specs/live-source-reporting/spec.md)

A baseline engine run with an isolated, empty but structurally complete v2 SQLite database reproduced CAP-02 through the production server. Both live REST requests returned HTTP 500 for a missing v1 session table. SSE returned HTTP 200 and sent no frame in 2.5 seconds with a two-second polling interval. Startup printed the listener address without a source diagnosis. This was a synthetic fixture.

Carry a typed live capability and reason from production selection to startup, poller activation and the three HTTP handlers. Keep the open handle available to historical hydration and close it once on shutdown. Validate the structures actually read by live SQL; message and part update timestamps are optional there. Recognize the native v1 empty auxiliary table. Report competing mixed layouts as unsupported for whole-source live observation. This supersedes the incidental ability to read only their v1 sessions and prevents a v2 owner appearing missing.

Preserve the existing HTTP error envelope. Use 503 with an identifiable reason for unsupported or unavailable live observation and 500 for operational read failures. Distinguish unset path, missing file, inaccessible source, unknown layout and malformed layout in the diagnostic. For an admitted SSE connection, deliver a named source diagnostic to the browser on initial or later read failure and a successful update on recovery. Preserve subscribe-before-snapshot ordering, per-session partial diagnostics, ordinary reconnect/fallback, cancellation and cleanup.

Historical requests still use `HydrateWorkspaceEvents` and `ExtractSessionSourceEvents`. The November 17 CAP-02 correction and current code agree that whole-database family validation rejects generic mixed layouts even for an exact owner. Both clean families remain extractable and clean v2 remains hydratable. Codex supplies historical data only. AL-63 owns its matrix and AL-74 owns required-column test-oracle repairs.

## Ownership and approval

Planner owns the proposal, spec, design and tasks for this planning pass. After explicit approval, coder owns affected production paths under the engine CLI, live/OpenCode/HTTP adapters and web live client/view. Tester owns regression test files; coordinate shared adapter paths before editing. Reviewers own assigned change-local reports only. The parent owns scope records after planning, integration, roadmap, telemetry and delivery.

The user approved iteration 01, including the competing mixed-layout behavior, with “approved. continue” in response to the proposed contract. Verification and review steps are in [tasks.md](tasks.md).

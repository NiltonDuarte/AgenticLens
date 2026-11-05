# docs(roadmap): reconcile active ticket readiness and development state (PO-19)

The roadmap mixed obsolete status sections and stale feature assumptions with current work. Reconcile all 23 active tickets against merged development and open PRs, remove the empty To Do and historical Discarded sections, and mark AL-64 In Progress with its existing draft implementation PR. Preserve four real dependency blocks and expose 18 Ready tickets.

Synchronize specification status/dependencies and record evidence for each ticket. Refresh AL2-1 to acknowledge shipped browser prompts, AL2-6 to preserve observed zero/partial/unavailable costs, and AL2-8 to retain native reasoning-effort settings. Preserve ticket IDs, historical artifacts, acceptance boundaries and portable ledger behavior. No application changes.

Validation: registry/spec inventory, status, links and reciprocal acyclic dependencies; git diff --check; full make passed (Go vet/race, frontend typecheck/lint, 853 frontend tests, boundary checks and native builds). Telemetry extraction facts and environment limits are recorded in PO-19/SUMMARY.md.

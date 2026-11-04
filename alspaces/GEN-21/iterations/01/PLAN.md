# GEN-21 iteration 01

1. Record clean baseline and unused GEN-21 allocation; inventory configured workspace recursively.
2. Verify published native checksum/version and inspect migrate help. Existing wrapper/native are absent and shared rc.111 pin has no available release. Use a temporary verified rc.125 native matching checkout without changing shared pin or installing callbacks.
3. Back up originals outside Git, dry-run then apply explicit schema 6 migration.
4. Compare persisted sessions/invocations and migration audits against originals; check idempotence and CLI ingestion. Record pending enrichment without source backfill.
5. Commit owned artifacts, integrate origin/main, run required make gate, generate final Codex telemetry, push and open PR.

Applicable regression boundaries: portable history without native sources, supported schema retirement with loss accounting, atomic backed-up writes, unavailable-owner preservation, nested workspace discovery and idempotence. This is an operator operation; no application development or delegated SDLC is required.

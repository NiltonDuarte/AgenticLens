# chore(GEN-26): assess direct OpenSpec workspace integration

Assess whether AgenticLens can use `openspec/changes` directly, with telemetry in each change folder and no separate AgenticLens workspace. Configuration-based probes confirm active change discovery with slug IDs and continued recursive discovery after archive. Dashboard detail and automatic publication still assume a direct active path, making archive resolution the main lifecycle gap.

The assessment maps nine compatibility gaps, distinguishes currently configurable standard artifacts from dynamic schema discovery, and prioritizes the minimum direct integration work. It pins both source revisions and separates executed fixture evidence from inferred behavior. No application code, tests, runtime configuration or OpenSpec installation is changed; recommendations do not authorize implementation.

Validation: full `make` passed; native layout/configuration probes and local document-link/diff checks passed. Existing suite evidence remains valid across the documentation-only clarification. Final Codex telemetry and its source/provisional limits are recorded in SUMMARY.md.

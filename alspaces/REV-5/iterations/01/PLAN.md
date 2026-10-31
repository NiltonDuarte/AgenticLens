# Audit plan

1. Read configuration, engineering rules, regression checklist, and recent 50-commit history.
2. Screen backend and frontend source with syntax-aware declaration/reference inventories; manually inspect candidates, runtime callers, and test peers. Exclude fixtures and implicit codec entry points.
3. Inspect layering, shallow APIs, state ownership, boundary validation, and test relevance. Reproduce concrete defects in disposable temporary directories.
4. Record findings and limits, run the integrated make gate, publish audit artifacts and final Codex telemetry.

Applicable regression boundaries: configuration rollback, test purpose/proportionality, source availability, and preservation of production paths. No performance target requested; no microbenchmark or timing claims. No code changes authorized by this audit.

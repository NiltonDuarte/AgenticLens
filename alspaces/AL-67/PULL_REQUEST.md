# docs(architecture): assess telemetry repository interfaces and injection (AL-67)

AgenticLens normalizes OpenCode v1/v2 and Codex facts, but consumers construct native source access differently. Document the actual repository/service boundaries, source selection and injection paths, with current/proposed dependency maps and a source-operation matrix.

Record twelve findings and recommendations, including server source configuration, v2 live capability, duplicate Codex snapshots, child-lookup diagnostics and test-oracle limits. Propose staged consumer-owned read capabilities while preserving portable history and explicit partial/unavailable results. All recommendations remain unimplemented; this PR changes documentation only.

Context: [assessment](ARCHITECTURE_ASSESSMENT.md), [contract](CONTRACT.md), [summary](SUMMARY.md).

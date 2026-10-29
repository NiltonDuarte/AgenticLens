# FIX-11 Iteration 05 Delta Plan: Require Canonical Source Roots

## Objective

Remove all compatibility behavior for `workspaces.src_folders` and amend `CONTRACT.md` to make the canonical source-roots requirement explicit. A valid configuration must provide at least one non-empty `code_metrics.source_roots` entry. Preserve `workspaces.dir` and `workspaces.artifacts`.

## Grounded Findings

- `backend/internal/domain/config.go`: `WorkspacesConfig.SrcFolders`, `CodeMetricsConfig.SourceRoots`, and `DefaultProjectConfig` currently represent both settings.
- `backend/internal/adapters/config/loader.go`: `ValidateConfig` accepts roots from either field; `rawWorkspacesConfig` decodes `src_folders`; `mergeConfig` copies values bidirectionally between `CodeMetrics.SourceRoots` and `Workspaces.SrcFolders`.
- `backend/internal/adapters/config/writer.go`: `GenerateConfigYAML` falls back between both slices, emits `workspaces.src_folders`, and reports an error naming that legacy field when roots are absent. `GenerateDefaultYAML` also initializes both models.
- `backend/internal/adapters/git/classifier.go`: `LoadConfig` falls back from `CodeMetrics.SourceRoots` to `Workspaces.SrcFolders` and fills both `ClassificationConfig.SourceRoots` and `SrcFolders`.
- `backend/internal/adapters/git/analyzer.go`: `ValidateSrcFolders`, `IsMonitoredSourcePath`, and `ComputeCodeFactsAtRef` retain legacy validation and selection paths.
- `backend/cmd/agentic-lens/main.go` reports both keys in missing-root guidance. `main_test.go` contains the associated message assertion and many inline configs still using `workspaces.src_folders` (including config loader assertions around lines 4801-4802).
- Direct legacy references also appear in `backend/internal/domain/config_test.go`, `backend/internal/adapters/config/loader_test.go`, `writer_test.go`, `repository_test.go`, `git/analyzer_test.go`, `git/classifier_test.go`, and `live/sqlite_test.go`. `.agenticlens-config.yaml` already declares `code_metrics.source_roots`.

## Implementation Tasks

1. **Contract amendment:** Update only the source-roots clauses in `alspaces/FIX-11/CONTRACT.md`: require `code_metrics.source_roots`, remove all acceptance, precedence, and fallback behavior for `workspaces.src_folders`, and preserve `workspaces.dir` and `workspaces.artifacts`. Retain unrelated contract text.
2. **Domain config:** Remove `SrcFolders` from `WorkspacesConfig` and its default initialization in `backend/internal/domain/config.go`. Keep `Workspaces.Dir` and `Workspaces.Artifacts`; keep `CodeMetrics.SourceRoots` as the sole roots field.
3. **Loader and validation:** Remove `SrcFolders` from `rawWorkspacesConfig` and remove all mapping/fallback logic in `backend/internal/adapters/config/loader.go`. Validate only `cfg.CodeMetrics.SourceRoots`, returning a `ConfigValidationError` that identifies `code_metrics.source_roots`. Legacy `workspaces.src_folders` input may be ignored by the current permissive YAML/JSON decoders; when canonical roots are absent, loading must fail with the actionable canonical-field error. Do not silently transfer legacy values.
4. **Writer/default config:** In `backend/internal/adapters/config/writer.go`, serialize roots only under `code_metrics.source_roots`; remove workspace source-root cleanup, mirroring, and legacy validation. Keep workspace directory and artifacts output intact. Make the missing-roots error name `code_metrics.source_roots`.
5. **Git classification and analysis:** In `backend/internal/adapters/git/classifier.go`, remove the `Workspaces.SrcFolders` fallback and `ClassificationConfig.SrcFolders`; carry only canonical roots. In `analyzer.go`, replace `ValidateSrcFolders` with canonical-root validation and remove all selector fallbacks. Preserve include/test/prod pattern behavior and existing diff classification.
6. **CLI guidance:** Update the missing-roots diagnostic in `backend/cmd/agentic-lens/main.go` to name only `code_metrics.source_roots` and describe it as required. Update the matching subprocess assertion in `main_test.go`.
7. **Tests and fixtures:** Update domain, config adapter, writer, repository, git, live adapter, and CLI tests/fixtures listed above to configure `code_metrics.source_roots` and assert `CodeMetrics.SourceRoots`. Replace legacy-acceptance/precedence cases in `loader_test.go` with coverage that legacy-only YAML and JSON do not satisfy validation, while canonical roots load and legacy keys cannot override them. Confirm generated YAML omits `workspaces.src_folders` and preserves `workspaces.dir` and artifacts.
8. **Manifest/defaults:** Confirm `.agenticlens-config.yaml` and `GenerateDefaultYAML` place roots under `code_metrics.source_roots`, with workspace storage settings retained.
9. **User documentation:** Update `README.md` and `docs/SETUP_MANUAL.md` so setup instructions direct users to the required `code_metrics.source_roots`. Keep the setup example's `workspaces.dir` and `workspaces.artifacts` settings.

## Acceptance Criteria

- Loading or CLI execution fails if `code_metrics.source_roots` is missing, empty, or contains only blank entries, even when `workspaces.src_folders` is supplied.
- The validation and CLI errors direct users to configure `code_metrics.source_roots`.
- The loader has no roots fallback or model field for `workspaces.src_folders`; git classification and analysis consume canonical roots only.
- Generated/default configuration emits `code_metrics.source_roots` and never emits `workspaces.src_folders`.
- `workspaces.dir` and `workspaces.artifacts` continue to load and serialize unchanged.
- `CONTRACT.md` requires canonical roots and contains no source-root acceptance or fallback for `workspaces.src_folders`, while retaining unrelated contract text.
- Updated fixtures cover YAML/JSON loading, writer round trips, CLI guidance, git monitoring, and rejection of legacy-only roots.
- `README.md` and `docs/SETUP_MANUAL.md` direct users to required `code_metrics.source_roots`; the setup example retains `workspaces.dir` and `workspaces.artifacts`.
- This remains a zero-code planning iteration; no application code or tests are changed.

## Verification

- Run focused tests for `internal/domain`, `internal/adapters/config`, `internal/adapters/git`, `internal/adapters`, and `cmd/agentic-lens`.
- Run the repository verification target (`make`) after focused tests.
- Inspect generated YAML and search configuration/domain/git/CLI code and fixtures with `rg` to confirm no production fallback or stale positive legacy expectations remain; verify `workspaces.dir` and artifacts round-trip coverage remains.
- Search `README.md` and `docs/SETUP_MANUAL.md` for `workspaces.src_folders` and confirm no stale setup guidance remains; verify the setup example retains `workspaces.dir` and `workspaces.artifacts` and directs users to required `code_metrics.source_roots`.

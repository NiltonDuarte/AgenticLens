# FIX-11 Iteration 05 Review

## Verdict: APPROVED

## Scope and contract alignment

The iteration removes `workspaces.src_folders` as a modeled or fallback source-root field and requires non-blank `code_metrics.source_roots`. The loader validates only canonical roots; the writer serializes them under `code_metrics`; Git classification and analysis consume `ClassificationConfig.SourceRoots` without a legacy selector. Legacy-only YAML and JSON fixtures assert validation failure, and canonical roots take precedence when legacy keys are present. Workspace directory and artifact settings remain represented and serialized.

The contract now states the canonical-only requirement while retaining `workspaces.dir` and `workspaces.artifacts`. The iteration plan's user-documentation task is implemented. README Git-discovery guidance points to required `code_metrics.source_roots`. The setup manual example places roots in the `code_metrics` section and retains workspace directory and artifacts; the field reference removes `workspaces.src_folders` and documents `source_roots` as required. No stale positive `src_folders` guidance remains in either user-facing document.

## Engineering review

- Validation remains in the config adapter; no domain I/O or layering breach was found in the reviewed changes.
- Writer, loader, Git path filtering, and CLI error guidance consistently use the canonical field.
- Updated fixtures cover YAML/JSON rejection of legacy-only roots, canonical-vs-legacy behavior, generated YAML omission of `src_folders`, and workspace settings preservation.
- The diff stays within FIX-11 scope. No new blocking issue was found in the full change set reviewed.

## Verification evidence

The task brief reports focused race tests and `GOCACHE=/tmp/agenticlens-go-cache make` passing. It also reports `GOCACHE=/tmp/agenticlens-go-cache make test` passing under OpenCode across 12 Go packages and 32 frontend files / 734 tests, with existing React `act` warnings. `git diff --check` passes. I did not rerun the suites during this review round.

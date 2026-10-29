# FIX-11 Iteration 04 Delta Plan: Accurate Source Roots Validation Guidance

## Objective

Correct the CLI validation message for missing source roots so it reflects both accepted configuration fields, and add focused regression coverage in `backend/cmd/agentic-lens/main_test.go` if the existing CLI test seam supports it.

## Grounding

- `backend/cmd/agentic-lens/main.go` checks `ConfigValidationError.MissingFields` for both `workspaces.src_folders` and `code_metrics.source_roots`, but currently reports only `workspaces.src_folders` and directs users to configure that legacy field.
- FIX-11's contract and iteration 03 established `code_metrics.source_roots` as the canonical location, with `workspaces.src_folders` accepted for backward compatibility.
- `backend/cmd/agentic-lens/main_test.go` already exercises CLI behavior through subprocesses and checks captured stderr, providing a suitable regression seam without changing production interfaces.

## Implementation Tasks

1. Update only the source-roots validation guidance in `backend/cmd/agentic-lens/main.go`. State that source roots are missing and name both accepted YAML fields; identify `code_metrics.source_roots` as the recommended field while retaining `workspaces.src_folders` as the legacy-compatible option. Preserve existing validation matching and exit behavior.
2. Add a focused CLI regression test in `backend/cmd/agentic-lens/main_test.go` that triggers the missing-source-roots validation path and asserts the stderr guidance mentions both field names and presents the canonical field accurately. Reuse the current subprocess/config fixture pattern; avoid broad fixture changes or unrelated validation coverage.

## Acceptance Criteria

- The reported guidance no longer implies `workspaces.src_folders` is the only accepted field.
- The CLI message clearly names `code_metrics.source_roots` and `workspaces.src_folders`, with canonical/legacy status unambiguous.
- Focused regression coverage verifies the user-visible message through the existing CLI test seam.
- No application behavior beyond the validation text changes; no files outside `main.go` and `main_test.go` are implementation targets.

## Verification

Run the focused CLI package test covering the new regression case and confirm existing CLI tests remain green.

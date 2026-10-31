# FIX-15: Make CI Go Vet Independent of Generated Frontend Assets

Ticket: FIX-15. Classification and telemetry type: `fix`.

## Problem

The `Build and Publish Release Candidate` workflow installs frontend dependencies and runs `make check` without first building the frontend bundle. `make check` invokes `go vet ./...`; Go evaluates `//go:embed all:dist` at package load time and fails when `backend/internal/adapters/web/dist` is absent in a clean checkout. Other workflows have different asset preparation paths, which risks the same failure recurring.

## Contract

- Every CI workflow that runs Go analysis, Go tests, or Go builds from a clean checkout must prepare embedded frontend assets first, or use a shared Make target that does so.
- The release candidate static analysis path must succeed after its documented dependency installation steps.
- Keep generated `dist` files out of version control and preserve the single-binary embedded frontend behavior.
- Review every workflow in `.github/workflows/` for equivalent clean-checkout build ordering and report unrelated findings separately.

## Scope

- CI workflow YAML and Makefile build/check orchestration only, unless a directly related regression test is warranted.
- No changes to runtime embedding, application behavior, or release policy.

## Acceptance

1. The release candidate workflow creates/builds the embedded asset tree before `make check` invokes `go vet`.
2. All other CI workflows that run Go tooling have a valid asset preparation path from a clean checkout.
3. CI workflow review covers all checked-in workflow files.
4. `make check` and relevant workflow-equivalent build/check commands pass locally.

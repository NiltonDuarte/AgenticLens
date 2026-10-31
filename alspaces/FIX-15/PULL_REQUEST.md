# fix(ci): prepare embedded assets before static checks (FIX-15)

## Overview

Make the shared `make check` path build the frontend bundle before Go vet evaluates packages that embed `backend/internal/adapters/web/dist`. This fixes clean-checkout failures in the release candidate workflow and applies the same ordering to stable release and release-please workflows.

## Summary

- Added `web-assets` target and made `check` depend on it before `vet` and `lint`.
- Added a clear error when frontend dependencies have not been installed.
- Reviewed all workflows under `.github/workflows/`; workflow-by-workflow findings are in `alspaces/FIX-15/iterations/01/REVIEW.md`.

Contract: [alspaces/FIX-15/CONTRACT.md](alspaces/FIX-15/CONTRACT.md)
Summary: [alspaces/FIX-15/SUMMARY.md](alspaces/FIX-15/SUMMARY.md)

## Verification

- `npm ci`
- `make check`
- Go and frontend test suites were not run.

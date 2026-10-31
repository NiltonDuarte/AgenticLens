# fix(ci): prepare embedded assets before Go checks (FIX-15)

## Overview

Make shared Make targets prepare the frontend bundle before Go vet or tests load packages that embed `backend/internal/adapters/web/dist`. This addresses clean-checkout failures in release workflows and local backend Go checks.

## Summary

- Added `web-assets`, made `vet` and `test-backend` depend on it, and added a combined `backend-check` target.
- Updated the CI backend job to call `make backend-check`.
- Reviewed every workflow under `.github/workflows/`; findings are recorded in `alspaces/FIX-15/iterations/01/REVIEW.md` and `iterations/02/REVIEW.md`.

Contract: [alspaces/FIX-15/CONTRACT.md](alspaces/FIX-15/CONTRACT.md)
Summary: [alspaces/FIX-15/SUMMARY.md](alspaces/FIX-15/SUMMARY.md)

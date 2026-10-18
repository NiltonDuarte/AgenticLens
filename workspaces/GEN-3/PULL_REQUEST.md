# GEN-3: Add Makefile Install Target and Modernize README

## Summary
Implements a top-level `install` target in `Makefile` to install backend Go module dependencies and frontend npm packages, chains `install` into the default `all` target, updates `README.md` to use Make targets throughout, and cleans up stale server port references.

## Changes
- Add `install`, `install-backend`, and `install-frontend` targets to `Makefile`.
- Include `install` in `all: install check test build`.
- Update `make help` to include `install` target descriptions.
- Modernize `README.md` setup and execution instructions to use `make install`, `make build`, and `make server`.
- Add comprehensive Makefile target reference table to `README.md`.
- Fix stale port `8501` references to default server port `8085` in `README.md`.

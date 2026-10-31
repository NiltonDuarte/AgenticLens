# Iteration 01 Plan: Stop Tracking Embedded Frontend Build Output

## Grounding

- `.gitignore` currently globally ignores `dist/` but explicitly re-includes `backend/internal/adapters/web/dist/` and its contents.
- `backend/internal/adapters/web/embed.go` uses `//go:embed all:dist`; the package expects generated output to be present before Go compilation.
- `web/package.json` copies the Vite build output into `backend/internal/adapters/web/dist/`.
- Root `Makefile` runs `npm run build` before compiling binaries when frontend dependencies are installed.

## Tasks

1. Replace the `.gitignore` exception rules with one exact ignore entry for `backend/internal/adapters/web/dist/`.
2. Remove tracked files beneath that directory from the Git index and working tree; preserve the directory's generated nature and all frontend sources.
3. Check the diff to confirm only the intended ignore rule and tracked bundle removals are included.
4. Verify Git ignores the target path and does not ignore unrelated `dist` output. Ensure the normal build flow generates the embedded bundle before Go compilation when frontend dependencies are available.

## Acceptance criteria

- No files beneath `backend/internal/adapters/web/dist/` remain tracked.
- A newly generated file at that exact location is ignored by Git.
- Other `dist` paths retain their prior ignore behavior.
- No application source or frontend source files change.
- The frontend build remains the source of embedded assets; Go compilation must occur after that generated directory is present.

## Risks

Because `embed.go` requires the `dist` directory at Go compile time, direct Go compilation or tests in a clean checkout without a preceding frontend build may fail. Validate through the repository's normal build workflow and keep generated output out of version control.

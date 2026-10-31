# GEN-12 iteration 01: Create the Embed Asset Staging Destination

Classification: `chore`. Baseline: `alspaces/GEN-12/CONTRACT.md`.

## Grounding

- Root `Makefile` runs `npm run build` from `web/` when frontend dependencies exist, then compiles Go binaries from `backend/`.
- The frontend package build clears and copies Vite output to `../backend/internal/adapters/web/dist/` without creating that destination.
- `backend/internal/adapters/web/embed.go` embeds the `dist` tree.

## Implementation

1. Create the embed asset destination before clearing and copying generated Vite output.
2. Preserve the existing working-directory-relative path and repeated-build refresh behavior.
3. Avoid changes to Go embed/runtime behavior and application code.

## Verification

- Run `make build` from the repository root after installing the locked frontend dependencies.
- Confirm staged `index.html`, JS, and CSS files are non-empty.
- Run the focused embedded web adapter test.
- Run `make build` a second time to check repeatability.

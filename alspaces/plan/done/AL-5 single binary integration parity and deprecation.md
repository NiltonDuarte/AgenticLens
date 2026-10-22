h3. Title: AL-5: Integrate Single Binary Build, Verify Parity, and Deprecate Python App

*TARGET*: AgenticLens (cli | build | tests)

*CONTEXT*

To finalize the rewrite, the frontend build must embed into the Go binary distribution, parity between Go rollups and legacy Python outputs must be verified on sample fixtures, and the legacy Python presentation code must be cleaned up.

*GOAL*

Bundle `web/dist` into `cmd/agentic-lens`, implement automated cross-engine parity tests comparing Go aggregations with Python outputs, and deprecate the Python presentation layer.

*VERTICAL SLICE SCOPE*

* Build & Packaging: Build script / Makefile coordinating `npm run build` and `go build -o bin/agentic-lens ./cmd/agentic-lens`.
* Adapters: `internal/adapters/web`: wire embedded filesystem with production build.
* Quality: Cross-validation test comparing Go domain calculations against Python output on fixture ledgers. Ensure 100% numerical parity.
* Cleanup: Deprecate `agentic_lens/presentation/`, `agentic_lens/domain/`, and Streamlit dependencies while preserving `scripts/extract_agentic_metrics.py`.

*ACCEPTANCE CRITERIA*

* `make build` (or build script) compiles standalone executable containing the full web app.
* Running `./bin/agentic-lens server` opens and serves the SPA without Node.js or Python.
* Parity test suite confirms identical token rollups, cost calculations, and health classifications between Go and Python.
* Streamlit dependencies removed from `pyproject.toml`.
* All Go and TypeScript checks pass (`go vet`, `go test -race`, `npm run typecheck`).

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Single standalone executable verified.
* Quality gates passing across Go and TypeScript.
* Telemetry recorded via `agentic-lens` CLI.

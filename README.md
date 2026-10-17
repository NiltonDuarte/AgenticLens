# AgenticLens

AgenticLens is a developer telemetry and observability dashboard for multi-agent software engineering workflows. It ingests, aggregates, and visualizes lifecycle telemetry generated across workspaces (`workspaces/**/agentic_metrics.json`), tracking token saturation, context window health, financial cost, iteration velocity, and tool error distributions.

## Architecture

AgenticLens is implemented in **Go** and **TypeScript** following the Functional Core, Imperative Shell (FCIS) pattern, distributed as a **single, self-contained Go executable**:

- **Domain Layer (`internal/domain`)**: Pure Go functional core. Value semantics, immutable structs, zero I/O, pure aggregations, context health scoring, and cost calculations.
- **Adapter Layer (`internal/adapters`)**: Imperative boundary shell. Workspace filesystem traversal, JSON schema validation, HTTP routing, and embedded asset serving (`internal/adapters/web`).
- **CLI & Server (`cmd/agentic-lens`)**: Entrypoint for CLI inspection commands (`scan`, `metrics`, `summary-append`) and the HTTP dashboard server (`server`).
- **Web Dashboard (`web/`)**: Strict TypeScript dashboard (SPA). Compiled static assets (`web/dist/`) are embedded into the Go executable at build time via `embed.FS`. Zero Node.js runtime required in production.

## Running Locally

### Development Mode

Run the Go server and Vite dev server with hot reload:

```bash
# Terminal 1: Run Go API / telemetry engine
go run ./cmd/agentic-lens server --port 8080

# Terminal 2: Run TypeScript UI with API proxy
cd web && npm run dev
```

### Standalone Production Binary

Build and run the self-contained binary:

```bash
# Build frontend assets and Go binary
cd web && npm run build && cd ..
go build -o bin/agentic-lens ./cmd/agentic-lens

# Run dashboard (serves embedded UI and API)
./bin/agentic-lens server --port 8501
```

Visit `http://localhost:8501`.

## Local Cluster Deployment

To build and deploy AgenticLens container into the local Kubernetes cluster (`sinklayer-local`):

```bash
docker build -t agentic-lens:latest .
```

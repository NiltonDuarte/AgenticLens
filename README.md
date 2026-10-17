# AgenticLens

AgenticLens is a developer telemetry visualization dashboard for SinkLayer Control Plane. It parses and visualizes development lifecycle metrics generated during agentic software development workflows from workspace ledgers (`workspaces/**/agentic_metrics.json`).

## Architecture

AgenticLens is organized using the Functional Core, Imperative Shell (FCIS) pattern:

- **Domain Layer (`agentic_lens/domain`)**: Pure, immutable dataclasses, constants, token formatting utilities, health ceiling calculations, and multi-ticket aggregation functions.
- **Adapter Layer (`agentic_lens/adapters`)**: External boundary schemas using Pydantic and a fault-tolerant filesystem scanner traversing workspace records.
- **Presentation Layer (`agentic_lens/presentation`)**: Streamlit user interface providing Fleet Overview, Cross-Ticket Comparisons, and Single-Ticket Deep Dive views.

## Running Locally

To run the Streamlit dashboard locally:

```bash
make run-agentic-lens
```

This starts Streamlit on host port 8501: `http://localhost:8501`.

## Local Cluster Deployment

To build and deploy AgenticLens into the local development Kubernetes cluster (`sinklayer-local`):

```bash
make cluster-agentic-lens
```

This builds the container image `agentic-lens:latest`, loads it into k3d or kind, and applies manifests in `infra/local/manifests/agentic-lens/` in the `sinklayer-system` namespace.

# Contract: GEN-1 Telemetry Extraction Script and Session Verification

## 1. Problem Statement
Provide a lightweight Python telemetry extractor to harvest session metrics, tokens, and git diff statistics into agentic_metrics.json.

## 2. Invariants
- Fast execution with zero heavyweight dependencies.
- Defensively handle missing files or unfinished sessions.

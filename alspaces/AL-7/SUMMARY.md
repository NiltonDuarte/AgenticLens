
# Ticket AL-7: Domain Anomaly Detection & Rule Engine

## Summary of Completed Work
- **Pure Functional Anomaly Engine**: Implemented `DetectAnomalies` in `backend/internal/domain/anomalies.go` covering 4 core anomaly rules (`TOOL_RETRY_STREAK`, `TOOL_HIGH_FAILURE_RATE`, `CONTEXT_SATURATION_SPIKE`, `SUBAGENT_REWORK_LOOP`) with deterministic multi-level sorting (severity, rule ID, turn/tool, title).
- **Domain Ledger Integration**: Wired anomaly evaluation into `DeriveMetricsLedger` in `backend/internal/domain/aggregations.go`, evaluating session telemetry across both granular v2 chronological events and legacy v1 rollups.
- **Boundary DTOs & Serialization**: Added `AnomalyContextDTO`, `AnomalyDTO`, and `AgenticMetricsLedgerDTO` with bidirectional mappings in `backend/internal/adapters/schemas.go`.
- **Markdown Summary Formatter**: Integrated `### Anomalies & Health Flags` section rendering in `backend/internal/adapters/formatters.go` supporting clean session banners and formatted severity badges.
- **CLI Diagnostics & Commands**: Added `--anomalies` flag and an `Anomalies` summary column to `--list` in `backend/cmd/agentic-lens/main.go`.
- **TypeScript Type Parity**: Added `AnomalySeverity`, `AnomalyRuleID`, `AnomalyContext`, `Anomaly` interfaces and updated `AgenticMetricsLedger` in `web/src/domain/types.ts`.
- **Documentation & Verification**: Updated `README.md` and added extensive adversarial unit tests in `backend/internal/domain/anomalies_test.go`, adapter formatter tests, and CLI integration tests in `backend/cmd/agentic-lens/main_test.go`.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 3
- **Active Agent Duration:** 1h 58m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 111.7k tokens (56% of 200k healthy ceiling, 3 compactions) | Subagents peak 168.1k (Reviewer)
- **Role Effort Distribution (Subagents):** Planner 20.1%, Coder 27.7%, Tester 35.4%, Reviewer 16.8%
- **Subagent Turn Velocity & Thrashing:** 58.5 avg turns/run (peak: 123 turns) | Planner 28.3, Coder 97.0, Tester 123.0, Reviewer 46.0
- **Test-to-Production Code Ratio:** 2.0:1 (+934 test lines / +478 prod lines)
- **Code Delivered:** +1,412 / -168 lines across 13 code files (excluding docs)
- **Reasoning Effort:** 134,246 reasoning tokens (53.8% of generated output)
- **Tool Health & Self-Correction:** 441 tool calls (8 self-corrections, 98.2% success rate)

### Anomalies & Health Flags
- **Status:** 6 execution anomalies detected (3 CRITICAL, 3 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.5k tokens) — Context tokens reached 160.5k, breaching the 160k critical ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+45.3k tokens) — Context tokens surged by 45.3k to 45.3k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+66.0k tokens) — Context tokens surged by 66.0k to 66.0k tokens.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: mcp-gateway_mcp-exec (4 consecutive errors) — Tool 'mcp-gateway_mcp-exec' encountered 4 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 867,721 | 25,530 | 5,057,260 | 85.4% | 111.7k | ELEVATED | $1.18 |
| **planner** | 3 | 495,260 | 24,603 | 3,060,544 | 86.1% | 91.1k | HEALTHY | $0.84 |
| **coder** | 1 | 613,678 | 26,581 | 6,011,114 | 90.7% | 166.7k | ELEVATED | $1.16 |
| **tester** | 1 | 848,707 | 33,369 | 8,539,614 | 91.0% | 166.2k | ELEVATED | $1.48 |
| **reviewer** | 1 | 514,026 | 5,391 | 2,976,575 | 85.3% | 168.1k | ELEVATED | $0.71 |
| **Total** | **7** | **3,339,392** | **115,474** | **25,645,107** | **88.5%** | **168.1k** | **ELEVATED** | **$5.36** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 7 | 3,339,392 | 115,474 | 25,645,107 | 88.5% | $5.36 |
| **Total** | **7** | **3,339,392** | **115,474** | **25,645,107** | **88.5%** | **$5.36** |

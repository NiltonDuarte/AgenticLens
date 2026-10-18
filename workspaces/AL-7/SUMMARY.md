
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
- **Software Development Life Cycle Iterations:** 1
- **Active Agent Duration:** 1h 56m (excludes user wait time)
- **Context Window Health:** ELEVATED — Orchestrator peak 92.3k tokens (46% of 200k healthy ceiling, 3 compactions) | Subagents peak 168.1k (Reviewer)
- **Role Effort Distribution (Subagents):** Planner 20.1%, Coder 27.7%, Tester 35.4%, Reviewer 16.8%
- **Subagent Turn Velocity & Thrashing:** 58.5 avg turns/run (peak: 123 turns) | Planner 28.3, Coder 97.0, Tester 123.0, Reviewer 46.0
- **Test-to-Production Code Ratio:** 2.0:1 (+934 test lines / +478 prod lines)
- **Code Delivered:** +1,412 / -168 lines across 13 code files (excluding docs)
- **Reasoning Effort:** 130,990 reasoning tokens (54.0% of generated output)
- **Tool Health & Self-Correction:** 415 tool calls (12 self-corrections, 97.1% success rate)

### Anomalies & Health Flags
- **Status:** 6 execution anomalies detected (3 CRITICAL, 3 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k in turn 29 (session: ses_eb1674d57ffer3yZnHLhOY7oku), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k in turn 41 (session: ses_eb183cc93ffeqoQN9Hlevo3825), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.5k tokens) — Context tokens reached 160.5k in turn 65 (session: ses_eb1775ee4ffemXyqBNsPuM08Tp), breaching 160k critical ceiling
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+45.3k tokens) — Context tokens surged by 45.3k tokens in turn 47 (session: ses_eb183cc93ffeqoQN9Hlevo3825, total: 45.3k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+66.0k tokens) — Context tokens surged by 66.0k tokens in turn 71 (session: ses_eb1775ee4ffemXyqBNsPuM08Tp, total: 66.0k tokens, threshold: 40000)
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: mcp-gateway_mcp-exec (4 consecutive errors) — Tool 'mcp-gateway_mcp-exec' encountered 4 consecutive execution failures (session: ses_eb19b6a25ffedH6iVtv24Rw61g, turn: 24)

### Agentic Execution Metrics

**Contributors:** nilton.guedes_duarte@digits.schwarz

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 653,629 | 21,711 | 2,248,041 | 77.5% | 92.3k | HEALTHY | $0.78 |
| **planner** | 3 | 495,260 | 24,603 | 3,060,544 | 86.1% | 91.1k | HEALTHY | $0.84 |
| **coder** | 1 | 613,678 | 26,581 | 6,011,114 | 90.7% | 166.7k | ELEVATED | $1.16 |
| **tester** | 1 | 848,707 | 33,369 | 8,539,614 | 91.0% | 166.2k | ELEVATED | $1.48 |
| **reviewer** | 1 | 514,026 | 5,391 | 2,976,575 | 85.3% | 168.1k | ELEVATED | $0.71 |
| **Total** | **7** | **3,125,300** | **111,655** | **22,835,888** | **88.0%** | **168.1k** | **ELEVATED** | **$4.97** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 7 | 3,125,300 | 111,655 | 22,835,888 | 88.0% | $4.97 |
| **Total** | **7** | **3,125,300** | **111,655** | **22,835,888** | **88.0%** | **$4.97** |

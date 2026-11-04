# AL-50 Work Summary

## Presentation / Web

- Refactored `web/src/utils/debugPrompt.ts` to 186 lines and removed destructive sanitization and grammatical repair from captured evidence.
- Added defensive Markdown fences around diagnostic payloads while keeping payload characters unchanged.
- Streamlined the four prompt templates and retained the exported anomaly rule label helper for existing consumers.

## Tests

- Added exact-preservation, fence safety, metadata, fallback, outcome, and rule-label coverage in `web/src/utils/debugPrompt.test.ts`.
- Updated copy-prompt integration assertions to reflect the concise prompts and fenced reasoning/error evidence.
- Refreshed the embedded frontend bundle used by the Go binary.

## Verification

- `make` passed: Go vet, TypeScript typecheck and lint, all Go race tests, all 764 frontend tests, and four-platform builds.
- `git diff --check` passed; code review verdict: `APPROVED`.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 3
- **Active Agent Duration:** 29m 0s (excludes user wait time)
- **Context Window Health:** HEALTHY — Root agent peak 93.3k tokens (21% of 450k healthy ceiling, 0 compactions) | Subagents peak 84.1k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.3 avg turns/run (peak: 2 turns) | Planner 1.0, Coder 1.0, Tester 2.0, Reviewer 1.0
- **Test-to-Production Code Ratio:** 1.0:1 (+97 test lines / +93 prod lines)
- **Code Delivered:** +224 / -1,399 lines across 8 code files (excluding docs)
- **Reasoning Effort:** 16,412 reasoning tokens (28.3% of generated output)
- **Tool Health & Self-Correction:** 101 tool calls (16 self-corrections, 84.2% success rate)

### Anomalies & Health Flags
- **Status:** 3 execution anomalies detected (3 WARNING).
- [WARNING] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: command (20.8%) — Tool 'command' failed 16 out of 77 calls (20.8% failure rate, threshold: >20%).
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (3 consecutive errors) — Tool 'command' encountered 3 consecutive execution failures.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (3 consecutive errors) — Tool 'command' encountered 3 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 12,726 | 1,724 | 158,720 | 92.6% | 46.1k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 51,778 | 13,388 | 3,382,272 | 98.5% | 77.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 52,614 | 7,260 | 3,082,368 | 98.3% | 84.1k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer** | 1 | 20,876 | 2,287 | 313,472 | 93.8% | 74.7k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 184,540 | 16,875 | 6,606,464 | 97.3% | 93.3k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **5** | **322,534** | **41,534** | **13,543,296** | **97.7%** | **93.3k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 5 | 322,534 | 41,534 | 13,543,296 | 97.7% | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **Total** | **5** | **322,534** | **41,534** | **13,543,296** | **97.7%** | **Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations)** |

# Review: AL-34 Eliminate Cumulative Granularity and Implement Bi-Directional Code Churn Visualization

### Executive Summary & Review Verdict

**Ticket**: `AL-34`
**Title**: Eliminate Cumulative Granularity & Implement Bi-Directional Code Churn Visualization
**Verdict**: **`APPROVED`**

---

### Axis 1: Spec & Contract Alignment

| Requirement / Invariant | Status | Audit Observations |
| :--- | :---: | :--- |
| **Cumulative Granularity Elimination (Backend)** | **PASS** | `GranularityCumulative` eliminated from `backend/internal/domain/enums.go`. `ParseTimeGranularity` accepts only `hourly`, `daily`, and `weekly`. HTTP layer (`backend/internal/adapters/http/handlers.go`) validates granularity and responds with `400 Bad Request` (`code: "bad_request"`) when invalid granularities (including `cumulative`) are supplied. Verified by adversarial tests in `server_test.go` and `aggregations_test.go`. |
| **Cumulative Granularity Elimination (Frontend)** | **PASS** | `TimeGranularity` in `web/src/domain/types.ts` strictly defined as `'hourly' | 'daily' | 'weekly'`. Granularity toggle buttons in `MetricProgression.tsx` only render Hourly, Daily, and Weekly options. Redundant timestamp formatter branch removed. Regression tests in `MetricProgression.test.tsx` verify the absence of the Cumulative toggle. |
| **Bi-directional Diverging Code Churn (Chart 4)** | **PASS** | `Delivered Code Line Churn & Net Growth` renders additions (`lines_added`) above the zero baseline and deletions (`lines_deleted`) negated below the zero baseline. `TrendChart.tsx` dynamically identifies zero baseline `yZero = getYCoord(0)`, rendering positive bars upwards and negative bars downwards with a prominent SVG baseline `<line className="chart-zero-baseline" ... />`. |
| **Cumulative Net Growth Metric & Invariant** | **PASS** | `CumulativeNetLinesAdded` added to `TimeBucketMetric` in both Go (`models.go`, `aggregations.go`) and TypeScript (`types.ts`, `progression.ts`). Running invariant `CumulativeNetLinesAdded == CumulativeLinesAdded - CumulativeLinesDeleted` holds across all buckets and is strictly asserted across boundary test suites. |
| **Discrete Integer Formatting & No Fractions** | **PASS** | `formatLinesCompact` and `formatCompactCount` round counts to integers (`+30k`, `-10k`, `0`), eliminating fractional lines (e.g. `33,306.56`). `integerTicks={true}` prevents fractional tick marks on discrete count charts. |
| **Composite Spend Trajectory Preservation** | **PASS** | Chart 1 retains period spend bars alongside cumulative trajectory line across all granularities (`hourly`, `daily`, `weekly`). Static title `"Financial Spend & Cumulative Trajectory ($ USD)"` preserved. |
| **Axis Label Containment** | **PASS** | SVG left margin configured to `64px` with `textAnchor="end"` at `x = 56px`, containing signed compact labels (`+250k`, `-100k`) without container clipping or horizontal overflow. |
| **Zero Scope Creep** | **PASS** | Diffs are confined to the requirements of `CONTRACT.md` and `PLAN.md`. No unrelated files or refactors modified. |

---

### Axis 2: Engineering Standards & Architecture

| Standard / Invariant | Status | Audit Observations |
| :--- | :---: | :--- |
| **Functional Core, Imperative Shell (FCIS)** | **PASS** | `backend/internal/domain/` maintains strict purity: zero imports of `adapters`, `cmd`, `os`, `net/http`, or outer layers. Pure functions without side effects. Value-semantic structs. |
| **Boundary Discipline & Validation** | **PASS** | Input query validation enforced at HTTP handler boundaries. Errors returned with standardized JSON error responses. |
| **Zero Mocks in Domain Logic** | **PASS** | Domain test suites in Go and TypeScript execute deterministically against concrete fixtures without mocks or test stubs. |
| **Strict Typing & Zero Any** | **PASS** | Zero instances of `any` or `interface{}` across Go and TypeScript files. TypeScript strict mode fully compliant. |
| **Zero Emojis** | **PASS** | Verified clean across all modified files, UI labels, tooltips, comments, and test descriptions. |
| **Resilience & Empty States** | **PASS** | Empty datasets, zero churn periods, and negative net churn render cleanly without runtime exceptions or divide-by-zero errors. |

---

### Review Conclusion
All acceptance criteria, invariants, and architectural guidelines from `CONTRACT.md`, `iterations/01/PLAN.md`, and `TECHNICAL_GUIDELINES.md` are satisfied.

**Final Status**: `APPROVED`

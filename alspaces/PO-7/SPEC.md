# Feature Specification: Eliminate Cumulative Granularity and Implement Bi-Directional Code Churn Visualization

## Problem Statement

The longitudinal Metric Progression view in AgenticLens currently exhibits two critical usability and domain modeling flaws:

1. **Conflation of Time Granularity with Aggregation Mode**:
   - The Metric Progression controls bar presents `Cumulative` as a peer to `Hourly`, `Daily`, and `Weekly`.
   - Granularity is strictly a time-bucketing frequency (bucket duration). Cumulative trajectory is an accumulation function across time buckets.
   - Conflating the two creates contradictory domain semantics: selecting "Cumulative" silently falls back to daily bucketing under the hood while completely locking the user out of hourly or weekly cumulative views.
   - Out of the five charts on the page, only Chart 1 ("Financial Spend & Cumulative Trajectory") reacts to the "Cumulative" toggle. The remaining four charts (Token Volume Distribution, Prompt Cache Hit Rate Trajectory, Delivered Code Line Churn & Net Growth, and Peak Context Window Saturation) completely ignore it.
   - Chart 1 already presents a composite view (period spend bars alongside a cumulative spend trajectory line) in daily, weekly, and hourly modes. Forcing a dedicated "Cumulative" granularity mode adds unnecessary UI noise and dead interactive states.

2. **Inadequate Code Churn & Growth Visualization**:
   - Chart 4 ("Delivered Code Line Churn & Net Growth") plots `lines_added` and `lines_deleted` as two separate positive bars rising from the same zero bottom baseline, alongside a period `net_lines_added` line.
   - In standard software engineering analytics, code churn is visualized as a bi-directional diverging chart centered on a zero baseline: code additions (+, green) extend upwards, code deletions (-, red) extend downwards, and cumulative net code delivered (growth over time) traces the true project trajectory across the period.
   - Plotting deletions as positive bars alongside additions muddles velocity analysis, making it difficult to distinguish net code expansion from code refactoring/replacement.

3. **Y-Axis Tick Label Overflow and Fractional Discrete Line Counts**:
   - On Chart 4, Y-axis ticks display verbose fractional strings like `33,306.56 lines`.
   - Fractional counts are non-sensical for discrete lines of code.
   - The uncompacted string length overflows the chart bounding box and left padding (`padding.left = 64px`), bleeding outside the chart card.
   - Unlike Token Volume Distribution and Peak Context Saturation, which format large numbers cleanly using `k` and `M` (e.g. `200k`, `1.5M`), Chart 4 lacks compact integer metric scaling on its axes.

---

## Target Personas & Use Cases

- **AI System Engineers & Tech Leads**: Need to assess long-term project growth and codebase churn velocity across agent iterations at clear temporal intervals (hourly, daily, weekly) without misleading aggregation toggles.
- **Engineering Managers**: Need clear visual separation between additive code velocity (green bars upwards), code cleanup/deletion (red bars downwards), and cumulative net project delivery (running trajectory line).
- **Agent Developers**: Need clean, predictable dashboard behavior where time controls apply consistently across all charts on the page.

---

## User Stories

1. As an engineering lead, I want the granularity selector to offer only genuine temporal intervals (`Hourly`, `Daily`, `Weekly`), so that time bucketing is clear and consistent across the entire dashboard.
2. As a dashboard user, I want Chart 1 ("Financial Spend & Cumulative Trajectory") to always display period spend bars alongside the cumulative trajectory line, so that I can see both period burn rate and accrued spend at a glance without toggling modes.
3. As a dashboard user, I want Chart 2 ("Token Volume Distribution") to remain composite across input, output, and cache read tokens, so that token composition is always clear.
4. As an engineering lead, I want Chart 4 ("Delivered Code Line Churn & Net Growth") to visualize code additions as green bars growing upwards and code deletions as red bars growing downwards from a centered zero baseline, so that code expansion versus deletion is immediately intuitive.
5. As an engineering lead, I want the line series in Chart 4 to represent the cumulative net code growth across the selected window, so that I can track whether the codebase is expanding, stabilizing, or consolidating over time.
6. As an API consumer, I want the `/api/progression` endpoint to enforce valid temporal granularities (`hourly`, `daily`, `weekly`) and reject invalid values with a clear error message.
7. As a dashboard viewer, I want Y-axis ticks on Chart 4 formatted compactly using clean integer increments with `k` and `M` (e.g. `+30k`, `-10k`, `0`), so that labels never overflow the chart container and fractional line counts are eliminated while precise values remain visible on hover.

---

## Functional Requirements

### 1. Granularity Model Clean-Up (Domain & Adapters)
- **Go Domain (`backend/internal/domain/enums.go`)**:
  - Remove `GranularityCumulative` (`"cumulative"`) from `TimeGranularity` enum.
  - Update `ParseTimeGranularity` to validate only `hourly`, `daily`, and `weekly`.
  - Update `aggregations.go` to remove legacy fallbacks or references to cumulative granularity.
- **Go HTTP Adapter (`backend/internal/adapters/http/handlers.go`)**:
  - Update query parameter parsing in `handleProgression`: accept `hourly`, `daily`, `weekly`.
  - Return `400 Bad Request` with message `invalid granularity: must be hourly, daily, or weekly` if any other value (including `"cumulative"`) is passed.
- **TypeScript Domain (`web/src/domain/types.ts` & `progression.ts`)**:
  - Update `TimeGranularity` type: `'hourly' | 'daily' | 'weekly'`.
  - Remove `'cumulative'` handling from `getBucketKey` and formatters.
  - Compute `cumulative_net_lines_added` (`cumulative_lines_added - cumulative_lines_deleted`) in `TimeBucketMetric` to provide running cumulative net code growth.

### 2. Metric Progression UI Updates (`web/src/views/MetricProgression.tsx`)
- **Granularity Controls**:
  - Remove the "Cumulative" button from the granularity toggle group.
  - Only `Hourly`, `Daily`, and `Weekly` buttons remain, defaulting to `Daily`.
- **Chart 1 (Financial Spend & Cumulative Trajectory)**:
  - Remove conditional rendering based on `'cumulative'`.
  - Render composite series unconditionally:
    - Period Spend: Bar (`cost_usd`, color: `CHART_COLORS.spendPeriod`).
    - Cumulative Spend: Line (`cumulative_cost_usd`, color: `CHART_COLORS.spendCumulative`).
  - Chart title is static: `Financial Spend & Cumulative Trajectory ($ USD)`.
- **Chart 2 (Token Volume Distribution)**:
  - Retain composite multi-series bar representation (`cache_read_tokens`, `input_tokens`, `output_tokens`).
- **Chart 4 (Delivered Code Line Churn & Net Growth)**:
  - Update series definition and data mapping:
    - Additions: `lines_added` (positive value, green bar, grows upwards).
    - Deletions: `-lines_deleted` (negative value, red bar, grows downwards).
    - Cumulative Net Growth: `cumulative_net_lines_added` (line, color: `CHART_COLORS.linesNet`).
  - Tooltip: format additions as `+N lines`, deletions as `-N lines`, and cumulative net growth as `+N / -N lines`.

### 3. Diverging Bar & Bi-Directional Chart Support (`web/src/components/TrendChart.tsx`)
- Support charts with both positive and negative values (diverging domains where `minY < 0` and `maxY > 0`):
  - Automatically evaluate min and max values across all series; if minimum is negative, set `minY` to accommodate the negative range with proper margin padding.
  - Calculate zero baseline coordinate `yZero = getYCoord(0)`.
  - Render a subtle horizontal zero line at `yZero` so the divergence axis is clearly visible.
  - For bar elements:
    - Positive values (`val >= 0`): bar begins at `y = getYCoord(val)`, height is `yZero - y` (grows upward from zero baseline).
    - Negative values (`val < 0`): bar begins at `y = yZero`, height is `getYCoord(val) - yZero` (grows downward from zero baseline).
  - Line series continue smoothly across positive and negative territory based on `getYCoord`.

### 4. Compact Y-Axis Scaling & Overflow Prevention (`formatters.ts` & `TrendChart.tsx`)
- **Compact Metric Formatter (`web/src/domain/formatters.ts`)**:
  - Implement `formatLinesCompact` / `formatCompactCount` to format positive and negative discrete counts using `k` and `M` (e.g., `+30k`, `-12.5k`, `0`, `1.2M`).
- **Y-Axis Tick Discrete Rounding & Formatter Props (`web/src/components/TrendChart.tsx`)**:
  - Add optional `yAxisFormatter?: (val: number) => string` prop to `TrendChartProps` so axis ticks can use compact notations while tooltips retain full-fidelity values (`valueFormatter`).
  - Eliminate fractional tick calculations (e.g. `33,306.56`) for discrete values by rounding tick steps to clean whole numbers / integers (`Math.round` / clean decimal intervals).
  - Ensure Y-axis label text never clips or overflows the left chart padding / card container boundary.

---

## Non-Functional & Architectural Invariants

1. **FCIS (Functional Core, Imperative Shell)**:
   - Aggregations and mathematical rollups remain in pure functions (`internal/domain` in Go, `domain/progression.ts` in TypeScript).
   - Zero I/O, zero network, zero clocks in domain logic.
2. **Strict Typing & Zero Any**:
   - Zero `any` or `interface{}` across Go backend and TypeScript frontend.
   - Strict Discriminated Unions and type guards for chart series and granularity options.
3. **Zero Emojis**:
   - Zero emojis across code, labels, tooltips, tests, documentation, and commit messages.
4. **Test Integrity**:
   - Pure table-driven Go tests for valid/invalid granularities and HTTP endpoints.
   - Comprehensive TypeScript unit tests for `TrendChart` bi-directional rendering and `MetricProgression` controls. Zero mocks in domain layers.

---

## Decomposition & Delivery Roadmap

- [x] **PO-7**: Product specification and decomposition for Metric Progression controls refactor.
- [ ] **AL-34**: Eliminate Cumulative Granularity and Implement Bi-Directional Code Churn Visualization (Ready, unblocked).

# Streamlit vs TypeScript Web Graph and Chart Parity Report

## 1. Executive Summary
This report audits all visualizations present in the legacy Python Streamlit application (`agentic_lens/presentation/views/`) against the modern TypeScript frontend (`web/src/views/` and `web/src/components/`).

The audit reveals that while the TypeScript implementation contains high-fidelity custom SVG charts (`TrendChart`) in `MetricProgression` and rich tabular data tables, several charts from the Streamlit application were omitted or replaced with simplified progress-bar summaries. In particular, the four comparative analytics charts in `Comparisons` and the ad-hoc query charts in `AnalyticsStudio` are entirely absent in the TypeScript application.

---

## 2. Detailed View-by-View Chart Comparison

### View 1: Fleet Overview (`fleet_overview.py` vs `FleetOverview.tsx`)
| Streamlit Chart (Altair) | TypeScript Web Equivalent | Status | Details / Gap |
| :--- | :--- | :--- | :--- |
| **Cost by Ticket (USD)** (Vertical Bar Chart sorting tickets by cost descending) | None (Only tabular `TicketTable`) | **Missing** | Streamlit renders an interactive bar chart of spend per ticket. TypeScript lists tickets in a table with cost columns but lacks the visual bar chart. |
| **Cost by Model (USD)** (Horizontal Bar Chart of expenditure by LLM model) | `DistributionCard` (CSS progress bar list) | **Substituted** | Streamlit uses an Altair bar chart with formatted tooltip currency. TypeScript uses stacked percentage progress bars. Functionally equivalent information, different visual form. |
| **Subagent Role Effort Rollup** | `DistributionCard` (CSS progress bar list) | **Substituted** | Streamlit uses an Altair/Pandas styled table. TypeScript uses a styled distribution bar card. |

---

### View 2: Ticket Deep Dive (`ticket_deep_dive.py` vs `TicketDeepDive.tsx`)
| Streamlit Chart (Altair) | TypeScript Web Equivalent | Status | Details / Gap |
| :--- | :--- | :--- | :--- |
| **Peak Context Window Tokens by Role** (Horizontal Bar Chart with conditional red/blue coloring relative to 200k healthy ceiling) | `SaturationBar` (HTML/CSS progress bars and badges) | **Substituted** | Streamlit renders a discrete horizontal bar chart per role with threshold-based color mapping. TypeScript uses `SaturationBar` with an overall saturation track and role pills. |
| **Subagent Turn Velocity & Thrashing** (Horizontal Bar Chart of average turns per role) | `turn-track` (HTML/CSS progress bars) | **Substituted** | Streamlit renders an Altair horizontal bar chart. TypeScript renders custom CSS bar tracks inside a card. Same data and layout intent. |

---

### View 3: Metric Progression (`metric_progression.py` vs `MetricProgression.tsx`)
| Streamlit Chart (Altair) | TypeScript Web Equivalent | Status | Details / Gap |
| :--- | :--- | :--- | :--- |
| **Spend Progression (USD)** (Area chart for cumulative; bar chart for period) | `TrendChart` ("Financial Spend & Cumulative Trajectory") | **Present** | Full visual and mathematical parity. |
| **Invocation Volume & Duration** (Line chart for cumulative; bar chart for period) | None | **Missing** | Streamlit plots cumulative invocations or period invocations + active duration seconds. TypeScript `MetricProgression` has no dedicated chart for invocation volume or execution duration over time. |
| **Token Composition by Type** (Bar chart: Input, Output, Cache Read, Reasoning tokens) | `TrendChart` ("Token Volume Distribution") | **Partial Gap** | TypeScript displays Cache Read, Input, and Output tokens, but omits `reasoning_tokens` from the charted series. |
| **Prompt Cache Hit Rate Trend (%)** (Line chart with 50% baseline and 80% target dashed rules) | `TrendChart` ("Prompt Cache Hit Rate Trajectory (%)") | **Present** | Full visual parity including 50% baseline and 80% target reference lines. |
| **Code Delivery Volume (Lines)** (Grouped bar chart for Lines Added and Deleted) | `TrendChart` ("Delivered Code Line Churn & Net Growth") | **Present** | TypeScript combines lines added, lines deleted, and net lines growth in a single multi-series chart. |
| **Cumulative Code Growth (Net Lines)** (Area chart) | `TrendChart` (integrated as line series in Churn chart) | **Present** | Captured via net lines growth trajectory. |
| **Peak Context Window Saturation (Tokens)** (Bar chart with 200k ceiling dashed line) | `TrendChart` ("Peak Context Window Saturation vs Ceiling") | **Present** | Full visual parity with 200k red ceiling reference line. |

---

### View 4: Ticket Comparisons (`comparisons.py` vs `TicketComparisons.tsx`)
| Streamlit Chart (Altair) | TypeScript Web Equivalent | Status | Details / Gap |
| :--- | :--- | :--- | :--- |
| **Lead Time vs Active Execution Duration (Minutes)** (Grouped bar chart comparing turnaround vs active compute) | None | **Missing** | Streamlit plots a grouped bar chart of lead duration vs active duration in minutes. TypeScript only presents these numbers in the table. |
| **Tool Execution Success Rate by Ticket** (Bar chart with blue gradient, domain [90, 100]) | None | **Missing** | Streamlit plots ticket-by-ticket tool success rates. In TypeScript, only shown as table column and fleet average KPI. |
| **Orchestrator Context Window Saturation by Ticket** (Bar chart with red/green conditional color > 100%) | None | **Missing** | Streamlit plots context window saturation per ticket with threshold coloring. In TypeScript, only shown as table column. |
| **Code Delivery Volume vs Financial Cost** (Scatter / Bubble plot: Lines Added vs Cost USD, colored by Context Health) | None | **Missing** | Streamlit plots a 2D correlation scatter plot (Lines Added on X, Cost on Y, colored by Context Health). Completely missing in TypeScript. |

---

### View 5: Analytics Studio (`analytics_studio.py` vs TypeScript Web)
| Streamlit Chart (Altair) | TypeScript Web Equivalent | Status | Details / Gap |
| :--- | :--- | :--- | :--- |
| **Ad-Hoc Visualizations** (Vertical Bar, Horizontal Bar, Line, Scatter, Area, Heatmap, Box Plot) | None | **Missing** | Streamlit includes an interactive ad-hoc slice-and-dice visualization studio across invocations and tools datasets. TypeScript has no corresponding studio view. |

---

## 3. Summary of Gaps and Recommendations

1. **High Priority Visualizations Missing in Web**:
   - Four comparative analytics charts in `TicketComparisons`:
     1. Lead Time vs Active Compute Duration (Grouped bar chart)
     2. Tool Execution Success Rate (Bar chart)
     3. Orchestrator Context Saturation (Bar chart with threshold color)
     4. Code Delivery Volume vs Financial Cost (Scatter / bubble plot)
   - Invocation volume & active duration trend in `MetricProgression`.
   - Cost by Ticket bar chart in `FleetOverview`.

2. **Minor Visual Adjustments**:
   - Add `reasoning_tokens` to the Token Volume Distribution series in `MetricProgression.tsx`.

3. **Separate Scope Considerations**:
   - `AnalyticsStudio` was an advanced exploration sandbox; if desired in TypeScript, it can be scheduled as a dedicated feature ticket (e.g. AL-13) rather than blocking the single-binary packaging and Python retirement in AL-5.

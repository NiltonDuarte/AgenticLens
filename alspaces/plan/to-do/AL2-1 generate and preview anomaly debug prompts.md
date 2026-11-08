# AL2-1: Generate and preview anomaly debug prompts

## Goal

Engineers can generate, preview and copy the same evidence-preserving anomaly debug prompt from the CLI, API and existing dashboard entry points.

## Context

AL-50 already delivers browser prompt generation and copying. The remaining gap is CLI/API availability and a consistent preview flow. Engineers diagnosing tool failures, context saturation or subagent rework need available evidence without invented raw payloads.

## Acceptance criteria

* Generate deterministic Markdown containing incident overview, anomaly diagnostics, invocation forensics, preceding trace and rule-specific remediation for retry streak, high failure rate, context saturation and subagent rework anomalies. Available tool input/output and error bodies remain verbatim under the approved redaction policy.
* GET /api/workspaces/{ticketId}/debug-prompt accepts optional turn, rule_id and tool_name selectors and returns prompt Markdown with diagnostic metadata. Unknown tickets return 404; malformed parameters return 400; valid requests return 200.
* agentic-lens prompt supports --ticket, --turn, --rule and --output; stdout is the default, file output is explicit, and successful generation exits zero.
* Existing anomaly drawer, failed waterfall-node and live-alert entry points open a consistent rendered preview and raw text view. Copy reports the actual clipboard outcome with accessible success/failure feedback; loading and request failures remain recoverable.
* Use available native hydration and supported committed facts when sources are unavailable. Explicitly identify missing evidence and its source; never fabricate raw input/output or mutate a ledger on read.

## Implementation notes

Reuse existing browser prompt guarantees and controls. AL-69 owns anomaly attribution/count corrections; AL-72 owns shared clipboard reliability. Coordinate those paths without making them prerequisites or rebuilding their features.

Follow [engineering rules](../../../TECHNICAL_GUIDELINES.md) and [regression boundaries](../../../docs/REGRESSION_PREVENTION.md). [PO-19](../../PO-19/SPEC.md) preserves the earlier development reconciliation; historical observations require revalidation, not an assumption of current runtime support.

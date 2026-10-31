# GEN-14: independent manager and managed-ticket telemetry

The manager has its own ticket, workspace and native root session. Each managed ticket has a separate workspace, ledger and top-level root session. Messages sent from the manager act as user-like inputs in the destination conversation; they must not make that destination a manager subagent or roll its execution usage into the manager ledger. T3 may retain truthful agent-sender provenance.

Manager telemetry measures manager reasoning and control operations. Destination telemetry measures destination processing and its own genuine subagents. Instructions/results read into the manager naturally consume manager context; these locally consumed tokens differ from importing the destination's recorded execution usage. Never copy destination ledger facts into the manager ledger.

Authorization: original audit request and subsequent explicit clarification of manager-as-human-proxy behavior. Documentation-only audit; no application changes or development approval inferred. No task-owned changes outside alspaces/GEN-14. Do not alter other tickets or raw source records.

Iteration 02 supersedes iteration 01's interpretation of independent roots and agent-sent steering as problems. The missing McpToolCall decoder remains a proposed manager-observability fix, not an implemented change.

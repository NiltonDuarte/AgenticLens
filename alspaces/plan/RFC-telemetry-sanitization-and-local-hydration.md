# RFC: Telemetry Sanitization in Git & Dynamic Local Database Hydration

## 1. Executive Summary & Problem Statement

Agentic development sessions produce extensive interaction data: user prompts, LLM reasoning thoughts, tool invocations (`filePath`, `command`, code snippets), durations, token counts, and error outputs.

Currently, the `AgenticLens` CLI persists raw chronological events directly into `workspaces/<TICKET_ID>/agentic_metrics.json`. In long multi-turn sessions (100+ turns), this architecture introduces critical vulnerabilities and operational hurdles:

1. **Security & Confidentiality Risk**:
   - Committing raw prompts and tool inputs into Git repositories risks permanently leaking environment secrets (`.env`), API keys, credentials, customer data, and proprietary intellectual property into repository history.
   - Raw interaction context must never be committed to Git.
2. **Repository & Git Churn Bloat**:
   - In sessions with hundreds of tool calls, the `events` array consumes over 99% of file size and generates 6,000+ lines of formatted JSON diffs on routine telemetry generation.
3. **Empty Routine Shells**:
   - Routine successful tool calls currently store only bare metadata (`tool_name: "read"`, `status: "completed"`) without arguments or outputs. Storing hundreds of these empty shells adds massive noise with zero diagnostic utility.

---

## 2. Core Invariants

1. **Zero Sensitive Context in Git**:
   - User prompts, assistant reasoning streams, tool inputs, and successful command outputs are strictly prohibited from being committed to Git or written into `workspaces/`.
2. **Lightweight & Sanitized Committed Ledgers**:
   - `workspaces/<TICKET_ID>/agentic_metrics.json` must remain compact (<5 KB, ~50–80 lines).
   - Only aggregated metrics, detected anomalies, compactions, and explicitly failed tool executions (`status: "error"`) may be persisted to Git.
   - Any error message persisted to Git must pass through an automated secret redaction and normalization pipeline. Raw uncensored error output is never written to disk.
3. **Dynamic Local Hydration**:
   - Full-fidelity interaction history (prompts, tool parameters, reasoning thoughts) is read on-demand directly from the local OpenCode SQLite database at runtime when running `agentic-lens serve`.
   - Local trace data remains strictly in-memory and transient; it is never written to disk.

---

## 3. Dual-Plane Architecture

```
                       +--------------------------------------------------------+
                       |                     AgenticLens UI                     |
                       +---------------------------+----------------------------+
                                                   |
                                      GET /api/workspaces/:id/events
                                                   |
                         +-------------------------+-------------------------+
                         |                                                   |
                         v                                                   v
            [Local DB Available?]                                [Remote / CI / No DB?]
                         |                                                   |
                        YES                                                 NO
                         |                                                   |
                         v                                                   v
         +-------------------------------+                  +---------------------------------+
         |     SQLite Reader (AL-17)     |                  |  workspaces/:id/agentic_metrics |
         |   Non-locking read of local   |                  |  Committed sanitized ledger     |
         |   sessions/messages/parts     |                  |  containing anomalies & errors  |
         +---------------+---------------+                  +----------------+----------------+
                         |                                                   |
                         v                                                   v
         Full trace with prompts,                           Error & anomaly waterfall fallback
         tool inputs & reasoning                            (safe, compact, portable)
```

### 3.1 Committed Git Plane (`agentic_metrics.json`)

At generation time (`agentic-lens generate`):
- Compute high-level facts (`code`, `invocations`, aggregated `tools`).
- Evaluate domain anomaly detection rules in memory across all session events.
- Filter `events` to retain strictly:
  - Events with `status == "error"` (passed through the redaction engine).
  - Events with `event_type == "compaction"`.
  - Events directly associated with detected anomalies.
- Exclude all routine `completed` turn messages and successful tool calls.

#### 3.1.1 Error Message Sanitization & Redaction Engine

Uncensored error outputs are a primary vector for credential leaks (e.g. stack traces printing local variable dictionaries, database connection strings, authorization header dumps, or failed curl commands).

Before any `error_message` is persisted to `agentic_metrics.json`:
1. **Multi-Pattern Secret Redaction**:
   - **Bearer Tokens / JWTs**: `Bearer [A-Za-z0-9\-_.]+` -> `Bearer [REDACTED_TOKEN]`
   - **API Keys**: Common provider key patterns (OpenAI `sk-...`, Anthropic `sk-ant-...`, GitHub `ghp_...` / `github_pat_...`, AWS `AKIA...`, Slack `xox...`) replaced with `[REDACTED_API_KEY]`.
   - **Connection URIs**: URLs with credentials (`(postgres|mysql|mongodb|redis|amqp)://[^:]+:[^@]+@`) masked to `proto://[REDACTED_CREDENTIALS]@host`.
   - **Private Keys**: `-----BEGIN [A-Z ]+ PRIVATE KEY-----` replaced with `[REDACTED_PRIVATE_KEY]`.
   - **Environment Variable Assignments**: Matches on `(API_KEY|SECRET|TOKEN|PASSWORD|AUTH|KEY)=([^\s\n]+)` masked to `$1=[REDACTED]`.
2. **Path Sanitization**:
   - Strip absolute home directory paths (e.g. `/Users/username/...` or `/home/agent/...`) to relative or generic prefixes (`~/...`) to prevent leaking local developer usernames and directory structures.
3. **Truncation & Fingerprinting**:
   - Cap persisted error message length to at most 256 characters (or the first informative error line), discarding deep internal stack dumps where secrets typically hide.
   - The full, raw, uncensored traceback remains available in the local developer's SQLite database and is only viewed locally during hydration.

#### 3.1.2 Assistant Reflection Extraction & AI Error Summarization

Beyond deterministic regex masking, replacing raw error outputs with high-level semantic descriptions eliminates secret leakage while providing vastly superior human diagnostic readability.

We employ a defense-in-depth summarization hierarchy:

1. **Primary: Assistant Reflection Extraction (Zero Cost / Offline / Instant)**:
   - When a tool fails during an agent session, the LLM assistant almost invariably diagnoses and explains the failure in the very next turn (e.g. *"The test failed due to an assertion mismatch on calculateTotal; let me inspect the input fixture."*).
   - The SQLite database already records this subsequent assistant response message.
   - The CLI extractor extracts the opening diagnosis (first 1–2 sentences, capped at 200 characters) directly from the succeeding assistant turn and assigns it as `sanitized_description`.
   - **Zero additional LLM API calls, zero additional cost, works 100% offline.**

2. **Secondary: Interactive / CLI-Driven Sanitization Hook (Optional)**:
   - For standalone runs where assistant reflections are absent or ambiguous, `agentic-lens generate` supports an optional `--ai-sanitize` flag.
   - The CLI passes the regex-scrubbed error snippet to the configured provider asking for a 1-sentence sanitized technical summary stripped of all environment specifics.

3. **Fallback: Scrubbed Error Fingerprint**:
   - If no assistant reflection exists and AI summarization is disabled, the system falls back to the deterministic regex-scrubbed first line of the error output.

**Result**:
- A 650-event session shrinks from 650 entries down to 5–15 entries.
- File size drops from ~196 KB to ~4 KB.
- Git diffs become concise and reviewable.
- Zero secret leakage risk into public or shared Git repositories.
- Error events present clear, human-readable explanations instead of unreadable, sensitive raw dumps.

### 3.2 Dynamic Local Hydration Plane (`agentic-lens serve`)

At serve time:
- When a user inspects a ticket in `TicketDeepDive`:
  1. The UI requests `GET /api/workspaces/{ticketId}/events`.
  2. The backend handler checks if the ticket's session ID exists in the local OpenCode SQLite database (`opencode.db`).
  3. **Local DB Available (Developer Workstation)**:
     - Uses read-only, non-locking SQLite queries (`PRAGMA query_only = ON;`, WAL mode) to extract full messages, reasoning thoughts, and tool inputs/outputs.
     - Streams rich trace facts dynamically to the UI in memory.
  4. **Local DB Unavailable (CI, Docker, or Remote Repository Clone)**:
     - Gracefully falls back to the committed `agentic_metrics.json` error/compaction event list.
     - Displays the sanitized error and anomaly timeline.

---

## 4. Implementation Roadmap

### Phase 1: CLI Event Filtering, Error Redaction & Reflection Summarization
- **Scope**: `backend/internal/adapters/opencode/extractor.go`, `internal/adapters/sanitizer.go`, `cmd/agentic-lens/main.go`.
- **Changes**:
  - Implement pure domain/adapter `SanitizeError(msg string) string` with regex-based secret masking and path anonymization.
  - Implement assistant reflection extractor: inspect the assistant turn immediately following a failed tool call to extract the opening diagnosis as `sanitized_description`.
  - Filter extracted `events` prior to ledger serialization: keep only `Status == "error"` (sanitized) or `EventType == "compaction"`.
  - Cap persisted error message excerpt to 256 characters.
  - Maintain in-memory calculation of token usage, duration, and anomaly detection over the complete raw event stream so metrics and anomalies remain 100% accurate.
  - Backfill existing workspaces with `agentic-lens migrate` or backfill script to prune historical routine events and sanitize legacy error messages.

### Phase 2: Non-Locking Local SQLite Reader (Aligned with AL-17)
- **Scope**: `backend/internal/adapters/live/sqlite.go`.
- **Changes**:
  - Open local OpenCode database with non-blocking read-only pragmas.
  - Reconstruct session hierarchy and map session IDs to workspace tickets.

### Phase 3: Runtime Trace Hydration (Aligned with AL-18)
- **Scope**: `backend/internal/adapters/http/handlers.go`, `internal/domain/live.go`.
- **Changes**:
  - In `WorkspaceEventsHandler`, query the local SQLite reader for the ticket's session ID.
  - If found, construct and return the rich hydrated event stream (including reasoning and tool inputs).
  - If not found, return the sanitized committed ledger events from `agentic_metrics.json`.
- **Frontend**:
  - Update `TraceWaterfall.tsx` to display hydrated prompt/tool detail drawers when available, while seamlessly presenting the compact error view when only fallback events exist.

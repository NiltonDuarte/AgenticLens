# Implementation Plan: FIX-3 Iteration 03 - Multi-Layer Fluent Debug Prompt Generators

## 1. Goal and Iteration Scope

Incorporate user steering feedback for ticket FIX-3 to transform low-level ID clipboard copies into fluent, structured Markdown debug prompts ready for direct pasting into an AI assistant for deep dive root cause investigation:

1. Replace raw identifier copies with structured Markdown investigation prompts across four distinct operational layers:
   - Layer 1: Tool Call / Part Execution (individual tool invocation, inputs, outputs, error streak, provenance).
   - Layer 2: Execution Issue / Anomaly Card (anomaly alert, issue summary, trigger part provenance, error payload).
   - Layer 3: Message Turn Review (agent deliberation/reasoning, assistant message text, executed tools list, turn provenance).
   - Layer 4: Session Overview (subagent role, runtime model, parent session hierarchy, lifecycle status, turn and tool counts, plain English flagged issues).
2. Adhere strictly to a generic agent environment abstraction:
   - Zero internal ticket IDs (strip tickets like AL-20, FIX-3, GEN-42).
   - Zero git branch names (strip branch references like feat/... or fix/...).
   - Zero internal rule identifiers (translate rule codes like TOOL_RETRY_STREAK, TOOL_HIGH_FAILURE_RATE, CONTEXT_SATURATION_SPIKE, SUBAGENT_REWORK_LOOP, ConsecutiveToolErrors, ToolThrashing into plain English descriptions).
   - Zero internal severity tiers (no raw HIGH, CRITICAL, or WARNING labels).
3. Connect prompt generators to the Copy buttons across:
   - web/src/components/LiveTraceInspector.tsx (header copies Layer 4; turn copies Layer 3; part copies Layer 1; alert card copies Layer 2).
   - web/src/components/AnomalyDrawer.tsx (anomaly card copies Layer 2).
   - web/src/components/TraceWaterfall.tsx (timeline row copies Layer 1 for tool calls, Layer 3 for message turns).
   - web/src/components/FailedToolAlertCard.tsx (alert card copies Layer 2).
4. Update copy action button labels to "Copy Prompt" (reverting from temporary "Copied" feedback after 2000 milliseconds).
5. Add unit test suite in web/src/utils/debugPrompt.test.ts and update existing component tests.

---

## 2. Grounded Reference Architecture and File Paths

The plan is grounded against existing files, types, and modules:

### TypeScript Utility Core (web/src/utils/)
- web/src/utils/debugPrompt.ts:
  - Pure formatting module exporting generator functions: generateToolExecutionPrompt, generateExecutionIssuePrompt, generateMessageTurnPrompt, generateSessionOverviewPrompt.
  - Pure sanitization helpers: sanitizeTextForPrompt, translateRuleToPlainEnglish, formatProvenanceSentence.
  - Zero DOM dependencies, zero browser globals.
- web/src/utils/debugPrompt.test.ts:
  - Comprehensive unit test suite exercising all four generators, edge cases, missing optional fields, and sanitization filters.
- web/src/utils/clipboard.ts:
  - Existing copyTextToClipboard helper used by interactive UI buttons.

### TypeScript Domain Models and Types (web/src/domain/, web/src/types/)
- web/src/domain/live.ts:
  - LivePartFact, LiveMessageFact, LiveTraceView, LiveSessionNode, LiveAnomalyAlert.
- web/src/types/telemetry.ts:
  - EventFact, Anomaly, AnomalyContext, AnomalyRuleID, EventType.
- web/src/domain/anomalies.ts:
  - findMatchingLiveAlert, extractStreakCount, matchesAnomaly, findMatchingAnomaly.

### TypeScript Presentation Components (web/src/components/)
- web/src/components/LiveTraceInspector.tsx:
  - Header session copy button: wire to generateSessionOverviewPrompt.
  - LiveMessageTurnItem copy button: wire to generateMessageTurnPrompt.
  - LiveMessagePartItem copy button: wire to generateToolExecutionPrompt.
  - FailedToolAlertCard invocation: wire to generateExecutionIssuePrompt.
- web/src/components/FailedToolAlertCard.tsx:
  - Add optional context props (sessionId, turn, role, model, messageId).
  - Wire internal copy button to generateExecutionIssuePrompt.
- web/src/components/AnomalyDrawer.tsx:
  - Add prominent "Copy Prompt" button in drawer header or diagnosis section.
  - Wire to generateExecutionIssuePrompt.
- web/src/components/TraceWaterfall.tsx:
  - Wire timeline item copy button to generate Layer 1 for tool_call events and Layer 3 for turn_message and compaction events.
- Test suites:
  - web/src/components/LiveTraceInspector.test.tsx
  - web/src/components/FailedToolAlertCard.test.tsx
  - web/src/components/AnomalyDrawer.test.tsx
  - web/src/components/TraceWaterfall.test.tsx

---

## 3. Four-Layer Debug Prompt Architecture and Specifications

### Layer 1: Tool Call / Part Execution Prompt
- Function signature: generateToolExecutionPrompt(params: ToolExecutionPromptParams): string
- Input parameter attributes:
  - toolName (string)
  - turn (optional number)
  - sessionId (string)
  - messageId (optional string)
  - partId (optional string)
  - status (string: error, completed, or failed)
  - consecutiveErrors (optional number)
  - input (optional string)
  - output (optional string)
  - errorMessage (optional string)
- Output Markdown Structure:
  - Header: Level 1 Markdown heading. If status is error, title format is "Tool Execution Failure: toolName". If status is completed or success, title format is "Tool Execution: toolName".
  - Provenance Sentence: Narrative paragraph starting with "During turn [turn] of session [sessionId] ([provenance details]), the agent executed [toolName] and it [outcome description]."
    - Provenance details formats: if both message ID and part ID exist, "(message [messageId], part [partId])". If only message ID exists, "(message [messageId])". If only part ID exists, "(part [partId])". If neither exists, omit the parenthetical provenance.
    - Outcome description formats: if status is error and consecutiveErrors is greater than 1, "failed with [consecutiveErrors] consecutive failures on this tool". If status is error without consecutive streak, "failed with an execution error". If status is completed, "completed successfully".
  - Section Input: Level 3 Markdown heading "### Input" followed by fenced code block containing the input string, or "(No input payload recorded)" if absent.
  - Section Output: Level 3 Markdown heading "### Output" followed by fenced code block containing output or error message, or "(No output recorded)" if absent.
  - Section What to check: Level 3 Markdown heading "### What to check" with bullet list:
    - Why did this tool call fail or produce unexpected output?
    - Is this a genuine environment issue, invalid arguments from the agent, or a misclassified success?
    - What exact action or command should the agent execute next to recover?

### Layer 2: Execution Issue / Anomaly Card Prompt
- Function signature: generateExecutionIssuePrompt(params: ExecutionIssuePromptParams): string
- Input parameter attributes:
  - summary (string: plain English issue title, e.g. "Consecutive Failures on bash")
  - turn (optional number)
  - sessionId (string)
  - role (optional string: e.g. "coder")
  - model (optional string: e.g. "gemini-3.8-flash")
  - description (string: plain English description of what occurred)
  - partId (optional string)
  - errorDetails (optional string: error output, trigger event, or diagnostic details)
- Output Markdown Structure:
  - Header: Level 1 Markdown heading "Agent Execution Issue: [Sanitized Summary]".
  - Provenance Sentence: Narrative paragraph starting with "During turn [turn] of session [sessionId] ([agent details]), [sanitized description]."
    - Agent details formats: if role and model exist, "agent [role] running [model]". If only role exists, "agent [role]". If neither exists, omit parenthetical.
    - Part trigger suffix: if part ID is present, append " The last failure was triggered by part [partId]."
  - Section Error Details: Level 3 Markdown heading "### Error Details" followed by fenced code block containing errorDetails, or "(No additional error details recorded)" if absent.
  - Section What to check: Level 3 Markdown heading "### What to check" with bullet list:
    - What is causing this failure pattern (such as an infinite retry loop, incorrect assumptions about the filesystem, or misinterpreting file content as an error)?
    - How should the agent break out of this loop and recover?

### Layer 3: Message Turn Review Prompt
- Function signature: generateMessageTurnPrompt(params: MessageTurnPromptParams): string
- Input parameter attributes:
  - turn (number)
  - role (string: e.g. "assistant", "system", "user")
  - model (optional string)
  - sessionId (string)
  - messageId (optional string)
  - reasoningContent (optional string)
  - messageContent (optional string)
  - toolsExecuted (readonly array of objects with toolName string, status string, optional partId string)
- Output Markdown Structure:
  - Header: Level 1 Markdown heading. If model is present, "Turn [turn] Review ([role] / [model])". If model is absent, "Turn [turn] Review ([role])".
  - Provenance Sentence: Narrative paragraph starting with "In turn [turn] of session [sessionId] ([message provenance]), the agent deliberated and executed tools."
    - Message provenance format: if message ID exists, "(message [messageId])". If absent, omit.
  - Section Model Deliberation: Level 3 Markdown heading "### Model Deliberation" followed by text block of reasoningContent, or "(No deliberation content recorded)" if absent.
  - Section Message: Level 3 Markdown heading "### Message" followed by text block of messageContent, or "(No text message body recorded)" if absent.
  - Section Tools Executed: Level 3 Markdown heading "### Tools Executed" followed by bulleted list where each item renders as "- [toolName] ([partId]): [status]" or "- [toolName]: [status]". If empty, "(No tools executed in this turn)".
  - Section What to check: Level 3 Markdown heading "### What to check" with bullet list:
    - Compare the agent's stated plan with the actual tool outcomes in this turn.
    - Why did the action fail or deviate from what was intended?
    - What corrective action should the agent take next?

### Layer 4: Session Overview Prompt
- Function signature: generateSessionOverviewPrompt(params: SessionOverviewPromptParams): string
- Input parameter attributes:
  - sessionId (string)
  - parentId (optional string)
  - role (string)
  - model (string)
  - status (string: running, completed, error, stalled, etc.)
  - turnCount (number)
  - totalToolCalls (number)
  - failedToolCalls (number)
  - flaggedIssues (readonly array of plain English strings)
- Output Markdown Structure:
  - Header: Level 1 Markdown heading "Session Overview: [sessionId]".
  - Provenance Sentence: Narrative paragraph starting with "Session [sessionId] is a [role] subagent running [model] [parent session note]. The session is currently **[status]** after [turnCount] turns and [totalToolCalls] tool executions ([failedToolCalls] failed)."
    - Parent session note: if parentId is present and non-empty, "(parent session: [parentId])". If absent, omit.
  - Section Flagged Issues: Level 3 Markdown heading "### Flagged Issues" followed by bulleted list of plain English issues. If empty, "No flagged issues or anomalies recorded for this session."
  - Section What to check: Level 3 Markdown heading "### What to check" with bullet list:
    - Why did this session stall or accumulate repeated tool failures?
    - What guidance or state reset is needed to unblock and resume the agent?

### Sanitization and Decoupling Invariants
- Ticket ID stripping: regular expression matching ticket prefixes (AL-[0-9]+, FIX-[0-9]+, GEN-[0-9]+, PO-[0-9]+) replaced with generic terms or removed.
- Branch name stripping: regular expression matching git branch patterns (feat/[a-z0-9-_]+, fix/[a-z0-9-_]+) removed or replaced with generic task descriptions.
- Rule ID translation:
  - TOOL_RETRY_STREAK / ConsecutiveToolErrors mapped to "Consecutive tool failure streak".
  - TOOL_HIGH_FAILURE_RATE mapped to "High tool failure rate".
  - CONTEXT_SATURATION_SPIKE / ContextCeilingApproach mapped to "Context window saturation approach".
  - SUBAGENT_REWORK_LOOP / ToolThrashing mapped to "Repeated tool thrashing / rework loop".
  - TokenVelocityRunaway mapped to "Rapid token consumption velocity".
- Severity tier removal: strip occurrences of [CRITICAL], [WARNING], [HIGH], SeverityHigh, SeverityCritical, etc. from formatted output text.

---

## 4. Component Integration and UI Wiring Specifications

### 1. web/src/components/LiveTraceInspector.tsx
- Header Session Prompt:
  - Session ID row contains a button labeled "Copy Prompt" with aria-label "Copy session overview debug prompt".
  - On click, invokes generateSessionOverviewPrompt passing sessionNode.id, sessionNode.parent_id, sessionNode.role, sessionNode.model, sessionNode.status, turn count from trace.messages.length, total tool calls from trace.total_tool_calls, failed tool calls from trace.failed_tool_calls, and mapped plain English anomalies.
  - Replaces button text with "Copied" for 2000 milliseconds.
- Message Turn Prompt:
  - In LiveMessageTurnItem, the turn header contains a button labeled "Copy Prompt" with aria-label "Copy turn [N] debug prompt".
  - On click, invokes generateMessageTurnPrompt passing turn number, message.role, sessionNode.model, sessionNode.id, message.id, message.reasoning_content, message.content, and mapped list of executed tools from message.parts.
  - Replaces button text with "Copied" for 2000 milliseconds.
- Message Part Prompt:
  - In LiveMessagePartItem, the part header contains a button labeled "Copy Prompt" with aria-label "Copy tool execution debug prompt for part [part.id]".
  - On click, invokes generateToolExecutionPrompt passing tool name, turn number, session ID, message ID, part ID, status, and input/output snippets.
  - Replaces button text with "Copied" for 2000 milliseconds.
- Failed Tool Alert Card:
  - In inspector-failed-tools-section, passes contextual session ID, turn number, message ID, role, model, and streak count into FailedToolAlertCard.

### 2. web/src/components/FailedToolAlertCard.tsx
- Component Props:
  - Extend FailedToolAlertCardProps with optional fields: sessionId?: string, turn?: number, role?: string, model?: string, messageId?: string.
- Header Copy Action:
  - Button labeled "Copy Prompt" with aria-label "Copy execution issue debug prompt for [toolName]".
  - On click, invokes generateExecutionIssuePrompt passing formatted summary (e.g. "Consecutive Failures on [toolName]"), turn number, session ID, role, model, description of tool failure and streak count, part ID, and error output.
  - Replaces button text with "Copied" for 2000 milliseconds.

### 3. web/src/components/AnomalyDrawer.tsx
- Drawer Header / Diagnosis Section:
  - Add a dedicated button labeled "Copy Prompt" with aria-label "Copy execution issue debug prompt".
  - On click, invokes generateExecutionIssuePrompt passing translated rule name / summary, turn number, session ID, subagent role, description, and error message / payload.
  - Replaces button text with "Copied" for 2000 milliseconds.

### 4. web/src/components/TraceWaterfall.tsx
- Event Timeline Row:
  - Timeline top row contains a button labeled "Copy Prompt" with aria-label "Copy debug prompt for turn [turn]".
  - On click:
    - If ev.event_type === 'tool_call': invokes generateToolExecutionPrompt passing ev.tool_name, ev.turn, ev.session_id, ev.message_id, ev.status, ev.tool_input, ev.tool_output, and ev.error_message.
    - If ev.event_type !== 'tool_call': invokes generateMessageTurnPrompt passing ev.turn, ev.role, ev.session_id, ev.message_id, ev.reasoning_content, ev.prompt_content, and empty executed tools list.
  - Replaces button text with "Copied" for 2000 milliseconds.

---

## 5. Phased Implementation Tasks for Coder

### Phase 1: Pure TypeScript Debug Prompt Generator Module
- Task 1.1: Create web/src/utils/debugPrompt.ts.
- Task 1.2: Implement sanitization functions:
  - stripInternalIdentifiers(text: string): string: Removes ticket IDs and git branch patterns.
  - translateRuleIdToPlainEnglish(ruleId: string): string: Maps known rule constants to human-readable phrases.
  - cleanSeverity(text: string): string: Strips severity tags and uppercase tier strings.
- Task 1.3: Implement generateToolExecutionPrompt(params):
  - Format Level 1 heading, provenance sentence, Input section, Output section, and What to check checklist.
- Task 1.4: Implement generateExecutionIssuePrompt(params):
  - Format Level 1 heading, provenance sentence with subagent details and part trigger, Error Details section, and What to check checklist.
- Task 1.5: Implement generateMessageTurnPrompt(params):
  - Format Level 1 heading with role and model, provenance sentence, Model Deliberation section, Message section, Tools Executed section, and What to check checklist.
- Task 1.6: Implement generateSessionOverviewPrompt(params):
  - Format Level 1 heading with session ID, provenance sentence with role, model, parent session, turn count, tool execution counts, Flagged Issues section, and What to check checklist.

### Phase 2: Unit Test Suite for Debug Prompt Generator
- Task 2.1: Create web/src/utils/debugPrompt.test.ts.
- Task 2.2: Test generateToolExecutionPrompt:
  - Assert Markdown structure for failed tool call with consecutive failure streak.
  - Assert Markdown structure for completed tool call with success outcome.
  - Assert fallback handling for missing message ID, part ID, and empty input/output snippets.
- Task 2.3: Test generateExecutionIssuePrompt:
  - Assert Markdown structure with turn, role, model, part ID, and error payload.
  - Assert fallback handling when model or part ID is absent.
- Task 2.4: Test generateMessageTurnPrompt:
  - Assert Markdown structure with deliberation reasoning, body message, and tool executions list.
  - Assert handling when deliberation is null or empty.
  - Assert handling when zero tools were executed.
- Task 2.5: Test generateSessionOverviewPrompt:
  - Assert Markdown structure with role, model, parent session, turn count, tool counts, and plain English flagged issues.
  - Assert handling when parent session is omitted.
  - Assert handling when zero issues are flagged.
- Task 2.6: Test sanitization:
  - Assert that ticket patterns (AL-20, FIX-3), branch patterns (feat/..., fix/...), internal rule strings (TOOL_RETRY_STREAK), and severity strings (CRITICAL, WARNING) are stripped or translated into plain English.

### Phase 3: Live Observability UI Wiring
- Task 3.1: In web/src/components/LiveTraceInspector.tsx:
  - Wire session header copy button to generateSessionOverviewPrompt. Label "Copy Prompt".
  - Pass session ID, role, and model down to LiveMessageTurnItem.
  - Wire LiveMessageTurnItem copy button to generateMessageTurnPrompt. Label "Copy Prompt".
  - Pass turn number and message ID down to LiveMessagePartItem.
  - Wire LiveMessagePartItem copy button to generateToolExecutionPrompt. Label "Copy Prompt".
  - In inspector-failed-tools-section, locate parent message for each error part in trace.messages and pass turn, messageId, sessionId, role, and model to FailedToolAlertCard.
- Task 3.2: In web/src/components/FailedToolAlertCard.tsx:
  - Accept optional props: sessionId, turn, role, model, messageId.
  - Wire copy button to generateExecutionIssuePrompt. Label "Copy Prompt" with 2000 millisecond "Copied" feedback.

### Phase 4: Anomaly Drawer and Waterfall Timeline Wiring
- Task 4.1: In web/src/components/AnomalyDrawer.tsx:
  - Add "Copy Prompt" button in drawer header next to AnomalyBadge.
  - Wire click handler to invoke generateExecutionIssuePrompt using anomaly title, turn, session ID, role, description, and event error payload.
- Task 4.2: In web/src/components/TraceWaterfall.tsx:
  - Update timeline item copy button: rename label to "Copy Prompt".
  - On click, generate Layer 1 tool execution prompt if event_type is tool_call; generate Layer 3 message turn prompt otherwise.

### Phase 5: Component Test Suite Updates
- Task 5.1: In web/src/components/LiveTraceInspector.test.tsx:
  - Update tests to query buttons with accessible label / text "Copy Prompt".
  - Verify that clicking session copy button copies Layer 4 Markdown prompt.
  - Verify that clicking turn copy button copies Layer 3 Markdown prompt.
  - Verify that clicking part copy button copies Layer 1 Markdown prompt.
  - Verify 2000 millisecond "Copied" feedback transitions.
- Task 5.2: In web/src/components/FailedToolAlertCard.test.tsx:
  - Update tests to verify that clicking "Copy Prompt" copies Layer 2 Markdown prompt.
- Task 5.3: In web/src/components/AnomalyDrawer.test.tsx:
  - Add test asserting that "Copy Prompt" button copies Layer 2 Markdown prompt.
- Task 5.4: In web/src/components/TraceWaterfall.test.tsx:
  - Update tests to assert that timeline copy button copies Layer 1 or Layer 3 prompt.

### Phase 6: End-to-End Quality Gates and Verification
- Task 6.1: Run npm run typecheck to verify zero type errors and zero any.
- Task 6.2: Run npm run lint to verify code standards and formatting.
- Task 6.3: Run npm test across all frontend unit and component tests.
- Task 6.4: Run go test -race ./... and go vet ./... to verify backend integrity.
- Task 6.5: Verify strictly zero emojis across all newly authored files, prompt texts, and test fixtures.

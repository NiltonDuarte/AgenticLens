/**
 * Multi-layer fluent Markdown debug prompt generator.
 * Produces structured diagnostic prompts formatted for direct pasting into an AI assistant.
 * Strictly adheres to a generic agent environment abstraction without ticket IDs,
 * git branches, internal severity tiers, or raw rule constants.
 */

import type { AnomalyRule } from '../domain/anomalies';

export interface ToolExecutionPromptParams {
  readonly toolName: string;
  readonly turn?: number | undefined;
  readonly sessionId: string;
  readonly messageId?: string | null | undefined;
  readonly partId?: string | null | undefined;
  readonly status: string;
  readonly consecutiveErrors?: number | undefined;
  readonly input?: string | null | undefined;
  readonly output?: string | null | undefined;
  readonly errorMessage?: string | null | undefined;
}

export interface ExecutionIssuePromptParams {
  readonly summary: string;
  readonly turn?: number | undefined;
  readonly sessionId: string;
  readonly role?: string | undefined;
  readonly model?: string | undefined;
  readonly description: string;
  readonly partId?: string | null | undefined;
  readonly errorDetails?: string | null | undefined;
}

export interface ExecutedToolItem {
  readonly toolName: string;
  readonly status: string;
  readonly partId?: string | null | undefined;
}

export interface MessageTurnPromptParams {
  readonly turn: number;
  readonly role: string;
  readonly model?: string | undefined;
  readonly sessionId: string;
  readonly messageId?: string | null | undefined;
  readonly reasoningContent?: string | null | undefined;
  readonly messageContent?: string | null | undefined;
  readonly toolsExecuted?: readonly ExecutedToolItem[] | undefined;
}

export interface SessionOverviewPromptParams {
  readonly sessionId: string;
  readonly parentId?: string | null | undefined;
  readonly role: string;
  readonly model: string;
  readonly status: string;
  readonly turnCount: number;
  readonly totalToolCalls: number;
  readonly failedToolCalls: number;
  readonly flaggedIssues?: readonly string[] | undefined;
}

/**
 * Strips internal ticket identifiers (e.g. PROJ-123, ISSUE-42) and git branch patterns.
 */
export function stripInternalIdentifiers(text: string): string {
  if (!text || !text.trim()) {
    return '';
  }

  // Strip parenthesized or bracketed ticket and branch patterns
  let sanitized = text.replace(
    /\s*\(\s*(?:branches?|branch\s+)?(?:refs\/heads\/|origin\/)?(?:feat|fix|chore|refactor|test|ci|perf|build|docs)\/[a-zA-Z0-9_\-./]+\s*\)/gi,
    ''
  );
  sanitized = sanitized.replace(
    /\s*\(\s*(?:tickets?\s+)?[A-Z]{2,10}-\d+\s*\)/gi,
    ''
  );

  // Strip git branch patterns such as feat/foo-bar, fix/al-20-test, refs/heads/fix/...
  sanitized = sanitized.replace(
    /\b(?:on\s+|in\s+|for\s+)?(?:branches?|branch\s+)?(?:refs\/heads\/|origin\/)?(?:feat|fix|chore|refactor|test|ci|perf|build|docs)\/[a-zA-Z0-9_\-./]+\b/gi,
    ''
  );

  // Strip ticket patterns including optional "tickets" prefix
  sanitized = sanitized.replace(
    /\b(?:on\s+|in\s+|for\s+)?(?:tickets?\s+)?[A-Z]{2,10}-\d+\b/gi,
    ''
  );

  return sanitized;
}

/**
 * Strips internal severity tier markers such as [CRITICAL], [WARNING], [HIGH], SeverityHigh, etc.
 */
function cleanSeverityTiers(text: string): string {
  if (!text || !text.trim()) {
    return '';
  }

  let cleaned = text.replace(/\[\s*(?:CRITICAL|WARNING|HIGH|MEDIUM|LOW|INFO)\s*\]/gi, '');
  cleaned = cleaned.replace(/\bSeverity(?:Critical|High|Medium|Low|Warning|None)\b/g, '');
  cleaned = cleaned.replace(/\b(?:severity|tier):\s*(?:critical|high|medium|low|warning|none)\b/gi, '');
  cleaned = cleaned.replace(/\b(?:CRITICAL|WARNING|HIGH)\s*:\s*/g, '');

  return cleaned;
}

/**
 * Translates known internal anomaly rule identifiers into plain English descriptions.
 */
export function translateRuleToPlainEnglish(ruleId: AnomalyRule | string, toolName?: string): string {
  const cleanTool = toolName ? toolName.trim() : '';

  switch (ruleId) {
    case 'TOOL_RETRY_STREAK':
    case 'ConsecutiveToolErrors':
      return cleanTool
        ? `Consecutive tool failures on \`${cleanTool}\``
        : 'Consecutive tool failure streak';
    case 'TOOL_HIGH_FAILURE_RATE':
      return cleanTool
        ? `High failure rate on \`${cleanTool}\``
        : 'High tool failure rate';
    case 'CONTEXT_SATURATION_SPIKE':
    case 'ContextCeilingApproach':
      return 'Context window saturation approach';
    case 'SUBAGENT_REWORK_LOOP':
    case 'ToolThrashing':
      return cleanTool
        ? `Repeated tool thrashing on \`${cleanTool}\``
        : 'Repeated tool thrashing';
    case 'TokenVelocityRunaway':
      return 'Rapid token consumption velocity';
    default:
      return sanitizeTextForPrompt(ruleId.replace(/_/g, ' '));
  }
}

/**
 * Sanitizes arbitrary text for prompt generation by removing internal tickets,
 * branch references, raw severity tier strings, and translating rule constants.
 */
export function sanitizeTextForPrompt(text: string): string {
  if (!text || !text.trim()) {
    return '';
  }

  let sanitized = stripInternalIdentifiers(text);
  sanitized = cleanSeverityTiers(sanitized);

  // Translate embedded rule IDs if present in text
  sanitized = sanitized.replace(/\b(?:TOOL_RETRY_STREAK|ConsecutiveToolErrors)\b/g, 'Consecutive tool failures');
  sanitized = sanitized.replace(/\bTOOL_HIGH_FAILURE_RATE\b/g, 'High tool failure rate');
  sanitized = sanitized.replace(/\b(?:CONTEXT_SATURATION_SPIKE|ContextCeilingApproach)\b/g, 'Context window saturation approach');
  sanitized = sanitized.replace(/\b(?:SUBAGENT_REWORK_LOOP|ToolThrashing)\b/g, 'Repeated tool thrashing');
  sanitized = sanitized.replace(/\bTokenVelocityRunaway\b/g, 'Rapid token consumption velocity');

  // Clean up duplicate spaces and dangling punctuation
  sanitized = sanitized.replace(/[ \t]+/g, ' ');
  sanitized = sanitized.replace(/\s+([.,;:])/g, '$1');

  // Remove repeated punctuation like ",," or ", ," or ", ;"
  sanitized = sanitized.replace(/(?:,\s*)+,/g, ',');
  sanitized = sanitized.replace(/,\s*([.;:])/g, '$1');
  sanitized = sanitized.replace(/,\s*(?:and|or)\s*,/gi, ',');
  sanitized = sanitized.replace(/(?:^|\s),(?:\s*(?:and|or)\s*|\s*)/gi, ' ');
  sanitized = sanitized.replace(/,\s*(?:and|or)\s+(in|at|on|for|to|with|from|by|about|of)\b/gi, ' $1');
  sanitized = sanitized.replace(/\b(?:and|or)\s*([.,;:])/gi, '$1');
  // Trim leading punctuation or trailing commas/colons/semicolons/dashes (preserve trailing period)
  sanitized = sanitized.replace(/^[- \t,;:.]+|[- \t,;:]+$/g, '');
  sanitized = sanitized.replace(/^[ \t]*(?:and|or)\b\s*/gi, '');
  sanitized = sanitized.replace(/\b(?:and|or)[ \t]*$/gi, '');
  sanitized = sanitized.replace(/[ \t]+/g, ' ');
  sanitized = sanitized.trim();

  // If only punctuation or orphaned symbols remain (no substantive alphanumeric chars)
  if (!/[a-zA-Z0-9]/.test(sanitized)) {
    return '';
  }

  return sanitized;
}

/**
 * Formats fenced code blocks defensively against backtick escape sequences.
 */
function formatCodeBlock(content: string): string {
  const backtickMatches = content.match(/`+/g);
  const maxBackticks = backtickMatches
    ? Math.max(...backtickMatches.map((m) => m.length))
    : 0;
  const fenceLength = Math.max(3, maxBackticks + 1);
  const fence = '`'.repeat(fenceLength);
  return `${fence}\n${content}\n${fence}`;
}

/**
 * Layer 1: Tool Call / Part Execution Debug Prompt Generator.
 */
export function generateToolExecutionPrompt(params: ToolExecutionPromptParams): string {
  const cleanTool = params.toolName && params.toolName.trim().length > 0 ? params.toolName.trim() : 'unknown_tool';
  const statusLower = (params.status || '').toLowerCase();
  const isError = statusLower === 'error' || statusLower === 'failed';

  const heading = isError
    ? `# Tool Execution Failure: \`${cleanTool}\``
    : `# Tool Execution: \`${cleanTool}\``;

  const provenanceItems: string[] = [];
  if (params.messageId && params.messageId.trim().length > 0) {
    provenanceItems.push(`message \`${params.messageId.trim()}\``);
  }
  if (params.partId && params.partId.trim().length > 0) {
    provenanceItems.push(`part \`${params.partId.trim()}\``);
  }
  const provenanceSuffix = provenanceItems.length > 0 ? ` (${provenanceItems.join(', ')})` : '';

  let outcome: string;
  if (isError) {
    if (params.consecutiveErrors !== undefined && params.consecutiveErrors > 1) {
      outcome = `failed with ${params.consecutiveErrors} consecutive failures on this tool`;
    } else {
      outcome = 'failed with an execution error';
    }
  } else {
    outcome = 'completed successfully';
  }

  const sessionId = params.sessionId ? params.sessionId.trim() : 'unknown-session';
  const turnPrefix = params.turn !== undefined && params.turn !== null
    ? `During turn ${params.turn} of session \`${sessionId}\``
    : `In session \`${sessionId}\``;

  const provenanceSentence = `${turnPrefix}${provenanceSuffix}, the agent executed \`${cleanTool}\` and it ${outcome}. This data was extracted from the OpenCode database.`;

  const rawInput = params.input ? params.input.trim() : '';
  const sanitizedInput = rawInput.length > 0 ? sanitizeTextForPrompt(rawInput) : '';
  const inputBody = sanitizedInput.length > 0
    ? formatCodeBlock(sanitizedInput)
    : '(No input payload recorded)';

  const rawOutput = (params.output && params.output.trim().length > 0)
    ? params.output.trim()
    : (params.errorMessage && params.errorMessage.trim().length > 0)
      ? params.errorMessage.trim()
      : '';

  const sanitizedOutput = rawOutput.length > 0 ? sanitizeTextForPrompt(rawOutput) : '';
  const outputBody = sanitizedOutput.length > 0
    ? formatCodeBlock(sanitizedOutput)
    : '(No output recorded)';

  const checklist = [
    '### What to check',
    '- Why did this tool call fail or produce unexpected output?',
    '- Is this a genuine environment issue, invalid arguments from the agent, or a misclassified success?',
    '- What exact action or command should the agent execute next to recover?',
  ].join('\n');

  return [
    heading,
    provenanceSentence,
    '### Input',
    inputBody,
    '### Output',
    outputBody,
    checklist,
  ].join('\n');
}

/**
 * Layer 2: Execution Issue / Anomaly Card Debug Prompt Generator.
 */
export function generateExecutionIssuePrompt(params: ExecutionIssuePromptParams): string {
  let sanitizedSummary = sanitizeTextForPrompt(params.summary);
  if (!sanitizedSummary) {
    sanitizedSummary = 'Operational Execution Issue';
  }

  let agentDetails = '';
  const role = params.role ? params.role.trim() : '';
  const model = params.model ? params.model.trim() : '';
  if (role && model) {
    agentDetails = ` (agent \`${role}\` running \`${model}\`)`;
  } else if (role) {
    agentDetails = ` (agent \`${role}\`)`;
  } else if (model) {
    agentDetails = ` (running \`${model}\`)`;
  }

  const sessionId = params.sessionId ? params.sessionId.trim() : 'unknown-session';
  const turnPrefix = params.turn !== undefined && params.turn !== null
    ? `During turn ${params.turn} of session \`${sessionId}\``
    : `In session \`${sessionId}\``;

  let cleanDesc = sanitizeTextForPrompt(params.description || 'an execution issue occurred');
  if (!cleanDesc) {
    cleanDesc = 'an execution issue occurred';
  }
  if (!cleanDesc.endsWith('.')) {
    cleanDesc += '.';
  }
  if (cleanDesc.length > 0 && !cleanDesc.startsWith('`')) {
    const firstChar = cleanDesc.charAt(0);
    const secondChar = cleanDesc.charAt(1);
    if (firstChar === firstChar.toUpperCase() && secondChar && secondChar === secondChar.toLowerCase()) {
      cleanDesc = firstChar.toLowerCase() + cleanDesc.slice(1);
    }
  }

  const partTrigger = params.partId && params.partId.trim().length > 0
    ? ` The last failure was triggered by part \`${params.partId.trim()}\`.`
    : '';

  const provenanceSentence = `${turnPrefix}${agentDetails}, ${cleanDesc}${partTrigger} This data was extracted from the OpenCode database.`;

  const rawError = params.errorDetails ? params.errorDetails.trim() : '';
  const sanitizedError = rawError.length > 0 ? sanitizeTextForPrompt(rawError) : '';
  const errorBody = sanitizedError.length > 0
    ? formatCodeBlock(sanitizedError)
    : '(No additional error details recorded)';

  const checklist = [
    '### What to check',
    '- What is causing this failure pattern (such as an infinite retry loop, incorrect assumptions about the filesystem, or misinterpreting file content as an error)?',
    '- How should the agent break out of this loop and recover?',
  ].join('\n');

  return [
    `# Agent Execution Issue: ${sanitizedSummary}`,
    provenanceSentence,
    '### Error Details',
    errorBody,
    checklist,
  ].join('\n');
}

/**
 * Layer 3: Message Turn Review Debug Prompt Generator.
 */
export function generateMessageTurnPrompt(params: MessageTurnPromptParams): string {
  const role = params.role && params.role.trim().length > 0 ? params.role.trim() : 'assistant';
  const model = params.model && params.model.trim().length > 0 ? params.model.trim() : '';
  const header = model
    ? `# Turn ${params.turn} Review (\`${role}\` / \`${model}\`)`
    : `# Turn ${params.turn} Review (\`${role}\`)`;

  const sessionId = params.sessionId ? params.sessionId.trim() : 'unknown-session';
  const msgProv = params.messageId && params.messageId.trim().length > 0
    ? ` (message \`${params.messageId.trim()}\`)`
    : '';
  const provenanceSentence = `In turn ${params.turn} of session \`${sessionId}\`${msgProv}, the agent deliberated and executed tools. This data was extracted from the OpenCode database.`;

  const rawReasoning = params.reasoningContent ? params.reasoningContent.trim() : '';
  const sanitizedReasoning = rawReasoning.length > 0 ? sanitizeTextForPrompt(rawReasoning) : '';
  const deliberationBody = sanitizedReasoning.length > 0
    ? sanitizedReasoning
        .split('\n')
        .map((line) => `> ${line}`)
        .join('\n')
    : '(No deliberation content recorded)';

  const rawMessage = params.messageContent ? params.messageContent.trim() : '';
  const sanitizedMessage = rawMessage.length > 0 ? sanitizeTextForPrompt(rawMessage) : '';
  const messageBody = sanitizedMessage.length > 0
    ? sanitizedMessage
    : '(No text message body recorded)';

  let toolsBody = '(No tools executed in this turn)';
  if (params.toolsExecuted && params.toolsExecuted.length > 0) {
    toolsBody = params.toolsExecuted
      .map((tool) => {
        const toolName = tool.toolName && tool.toolName.trim().length > 0 ? tool.toolName.trim() : 'unknown_tool';
        const status = tool.status && tool.status.trim().length > 0 ? tool.status.trim() : 'unknown';
        if (tool.partId && tool.partId.trim().length > 0) {
          return `- \`${toolName}\` (\`${tool.partId.trim()}\`): ${status}`;
        }
        return `- \`${toolName}\`: ${status}`;
      })
      .join('\n');
  }

  const checklist = [
    '### What to check',
    "- Compare the agent's stated plan with the actual tool outcomes in this turn.",
    '- Why did the action fail or deviate from what was intended?',
    '- What corrective action should the agent take next?',
  ].join('\n');

  return [
    header,
    provenanceSentence,
    '### Model Deliberation',
    deliberationBody,
    '### Message',
    messageBody,
    '### Tools Executed',
    toolsBody,
    checklist,
  ].join('\n');
}

/**
 * Layer 4: Session Overview Debug Prompt Generator.
 */
export function generateSessionOverviewPrompt(params: SessionOverviewPromptParams): string {
  const sessionId = params.sessionId ? params.sessionId.trim() : 'unknown-session';
  const role = params.role && params.role.trim().length > 0 ? params.role.trim() : 'agent';
  const model = params.model && params.model.trim().length > 0 ? params.model.trim() : 'unknown-model';
  const status = params.status && params.status.trim().length > 0 ? params.status.trim() : 'unknown';

  const parentNote = params.parentId && params.parentId.trim().length > 0
    ? ` (parent session: \`${params.parentId.trim()}\`)`
    : '';

  const provenanceSentence = `Session \`${sessionId}\` is a \`${role}\` subagent running \`${model}\`${parentNote}. The session is currently **${status}** after ${params.turnCount} turns and ${params.totalToolCalls} tool executions (${params.failedToolCalls} failed). This data was extracted from the OpenCode database.`;

  let issuesBody = 'No flagged issues or anomalies recorded for this session.';
  if (params.flaggedIssues && params.flaggedIssues.length > 0) {
    const cleanedIssues = params.flaggedIssues
      .map((issue) => sanitizeTextForPrompt(issue))
      .filter((issue) => issue.length > 0);
    if (cleanedIssues.length > 0) {
      issuesBody = cleanedIssues
        .map((issue) => `- ${issue}`)
        .join('\n');
    }
  }

  const checklist = [
    '### What to check',
    '- Why did this session stall or accumulate repeated tool failures?',
    '- What guidance or state reset is needed to unblock and resume the agent?',
  ].join('\n');

  return [
    `# Session Overview: \`${sessionId}\``,
    provenanceSentence,
    '### Flagged Issues',
    issuesBody,
    checklist,
  ].join('\n');
}

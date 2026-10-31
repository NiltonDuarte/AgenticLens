/** Markdown debug prompts that preserve captured runtime evidence verbatim. */

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

const metadata = (value: string | null | undefined, fallback: string): string =>
  value?.trim() || fallback;

const titleMetadata = (value: string, fallback: string): string =>
  value.trim().replace(/[\r\n]+/g, ' ') || fallback;

const present = (value: string | null | undefined): value is string =>
  value !== null && value !== undefined && value.length > 0;

/** Use a fence longer than any run in the captured content. */
function formatCodeBlock(content: string): string {
  const runs = content.match(/`+/g) ?? [];
  const fence = '`'.repeat(Math.max(3, ...runs.map((run) => run.length + 1)));
  return `${fence}\n${content}\n${fence}`;
}

/** Stable display labels for known anomaly rules; runtime evidence stays untouched. */
export function translateRuleToPlainEnglish(ruleId: string, toolName?: string): string {
  const tool = toolName?.trim();
  const labels: Readonly<Record<string, string>> = {
    TOOL_RETRY_STREAK: tool ? `Consecutive tool failures on \`${tool}\`` : 'Consecutive tool failure streak',
    ConsecutiveToolErrors: tool ? `Consecutive tool failures on \`${tool}\`` : 'Consecutive tool failure streak',
    TOOL_HIGH_FAILURE_RATE: tool ? `High failure rate on \`${tool}\`` : 'High tool failure rate',
    CONTEXT_SATURATION_SPIKE: 'Context window saturation approach',
    ContextCeilingApproach: 'Context window saturation approach',
    SUBAGENT_REWORK_LOOP: tool ? `Repeated tool thrashing on \`${tool}\`` : 'Repeated tool thrashing',
    ToolThrashing: tool ? `Repeated tool thrashing on \`${tool}\`` : 'Repeated tool thrashing',
    TokenVelocityRunaway: 'Rapid token consumption velocity',
  };
  return labels[ruleId] ?? ruleId;
}

const evidence = (value: string | null | undefined, fallback: string): string =>
  present(value) ? formatCodeBlock(value) : fallback;

const location = (sessionId: string, turn?: number): string => turn === undefined
  ? `In session \`${metadata(sessionId, 'unknown-session')}\``
  : `During turn ${turn} of session \`${metadata(sessionId, 'unknown-session')}\``;

/** Layer 1: tool call or part execution. */
export function generateToolExecutionPrompt(params: ToolExecutionPromptParams): string {
  const tool = metadata(params.toolName, 'unknown_tool');
  const failed = ['error', 'failed'].includes(params.status.toLowerCase());
  const outcome = failed
    ? params.consecutiveErrors !== undefined && params.consecutiveErrors > 1
      ? `failed with ${params.consecutiveErrors} consecutive failures on this tool.`
      : 'failed with an execution error.'
    : 'completed successfully.';
  const ids = [
    present(params.messageId) ? `message \`${params.messageId}\`` : '',
    present(params.partId) ? `part \`${params.partId}\`` : '',
  ].filter(Boolean);
  const output = present(params.output) ? params.output : params.errorMessage;

  return [
    `# ${failed ? 'Tool Execution Failure' : 'Tool Execution'}: \`${tool}\``,
    `${location(params.sessionId, params.turn)}${ids.length > 0 ? ` (${ids.join(', ')})` : ''}, the agent executed \`${tool}\` and it ${outcome}`,
    '### Input',
    evidence(params.input, '(No input payload recorded)'),
    '### Output',
    evidence(output, '(No output recorded)'),
    '### Diagnosis',
    '- Identify the cause and the next corrective action.',
  ].join('\n');
}

/** Layer 2: execution issue or anomaly card. */
export function generateExecutionIssuePrompt(params: ExecutionIssuePromptParams): string {
  const summary = titleMetadata(params.summary, 'Operational Execution Issue');
  const role = params.role?.trim();
  const model = params.model?.trim();
  const agent = role && model
    ? `agent \`${role}\` running \`${model}\``
    : role ? `agent \`${role}\`` : model ? `running \`${model}\`` : '';
  const description = present(params.description) ? params.description : 'No description recorded.';

  return [
    `# Agent Execution Issue: ${summary}`,
    `${location(params.sessionId, params.turn)}${agent ? ` (${agent})` : ''}${present(params.partId) ? `; failure part \`${params.partId}\`` : ''}.`,
    '### Description',
    evidence(description, '(No description recorded)'),
    '### Error Details',
    evidence(params.errorDetails, '(No additional error details recorded)'),
    '### Diagnosis',
    '- Determine the failure cause and a recovery step.',
  ].join('\n');
}

/** Layer 3: review a single message turn. */
export function generateMessageTurnPrompt(params: MessageTurnPromptParams): string {
  const role = metadata(params.role, 'assistant');
  const model = present(params.model) ? ` / \`${params.model.trim()}\`` : '';
  const messageId = present(params.messageId) ? ` (message \`${params.messageId}\`)` : '';
  const tools = params.toolsExecuted?.length
    ? params.toolsExecuted.map((tool) => {
      const part = present(tool.partId) ? ` (${tool.partId})` : '';
      return `- ${metadata(tool.toolName, 'unknown_tool')}${part}: ${metadata(tool.status, 'unknown')}`;
    }).join('\n')
    : '(No tools executed)';

  return [
    `# Turn ${params.turn} Review (\`${role}\`${model})`,
    `In turn ${params.turn} of session \`${metadata(params.sessionId, 'unknown-session')}\`${messageId}, the agent deliberated and executed tools.`,
    '### Reasoning',
    evidence(params.reasoningContent, '(No reasoning content recorded)'),
    '### Message',
    evidence(params.messageContent, '(No message body recorded)'),
    '### Tools Executed',
    tools,
    '### Diagnosis',
    '- Compare the stated intent with the captured tool outcomes.',
  ].join('\n');
}

/** Layer 4: session-level summary and flagged evidence. */
export function generateSessionOverviewPrompt(params: SessionOverviewPromptParams): string {
  const sessionId = metadata(params.sessionId, 'unknown-session');
  const parent = present(params.parentId) ? `; parent session \`${params.parentId}\`` : '';
  const issues = params.flaggedIssues?.filter(present) ?? [];

  return [
    `# Session Overview: \`${sessionId}\``,
    `Session \`${sessionId}\` is a \`${metadata(params.role, 'agent')}\` subagent running \`${metadata(params.model, 'unknown-model')}\`${parent ? ` (parent session: ${params.parentId})` : ''}. The session is currently **${metadata(params.status, 'unknown')}** after ${params.turnCount} turns and ${params.totalToolCalls} tool executions (${params.failedToolCalls} failed).`,
    '### Flagged Issues',
    ...(issues.length > 0
      ? issues.map((issue) => evidence(issue, ''))
      : ['(No flagged issues recorded)']),
    '### Diagnosis',
    '- Identify recurring failures or context that blocked progress.',
  ].join('\n');
}

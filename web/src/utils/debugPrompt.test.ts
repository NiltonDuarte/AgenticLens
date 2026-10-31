import { describe, expect, it } from 'vitest';
import {
  generateExecutionIssuePrompt,
  generateMessageTurnPrompt,
  generateSessionOverviewPrompt,
  generateToolExecutionPrompt,
  translateRuleToPlainEnglish,
} from './debugPrompt';

const assertFencedVerbatim = (prompt: string, content: string, fence = '```'): void => {
  expect(prompt).toContain(`${fence}\n${content}\n${fence}`);
};

describe('debug prompt generation', () => {
  it('preserves tool input and output verbatim, including identifiers, tags, punctuation, and whitespace', () => {
    const input = '  git checkout fix/FIX-3\n# AL-42 [CRITICAL] TOOL_RETRY_STREAK: keep, exactly.  ';
    const output = '\tERROR: SeverityHigh; AL-42, fix/FIX-3\n';
    const prompt = generateToolExecutionPrompt({
      toolName: 'bash', sessionId: 'session-1', status: 'error', input, output,
    });

    assertFencedVerbatim(prompt, input);
    assertFencedVerbatim(prompt, output);
    expect(prompt).not.toContain('OpenCode database');
    expect(prompt).toContain('### Diagnosis');
  });

  it('preserves an error message when output is empty and handles absent payloads', () => {
    const raw = 'compiler: expected "," after expression [HIGH] AL-42';
    const prompt = generateToolExecutionPrompt({
      toolName: 'go test', sessionId: 'session-2', status: 'failed', output: '', errorMessage: raw,
    });

    assertFencedVerbatim(prompt, raw);
    expect(prompt).toContain('(No input payload recorded)');
  });

  it('uses defensive fences longer than embedded backtick runs', () => {
    const raw = '```ts\nconst branch = "fix/FIX-3";\n```';
    const prompt = generateToolExecutionPrompt({
      toolName: 'edit', sessionId: 'session-3', status: 'completed', input: raw,
    });

    assertFencedVerbatim(prompt, raw, '````');
  });

  it('formats success and failure outcomes with optional provenance IDs', () => {
    const success = generateToolExecutionPrompt({
      toolName: ' read ', turn: 0, sessionId: 'session-ids', messageId: 'msg-1',
      partId: 'part-1', status: 'completed', input: 'README.md', output: 'file contents',
    });
    expect(success).toContain('# Tool Execution: `read`');
    expect(success).toContain('During turn 0 of session `session-ids` (message `msg-1`, part `part-1`)');
    expect(success).toContain('and it completed successfully.');
    assertFencedVerbatim(success, 'file contents');

    const failure = generateToolExecutionPrompt({
      toolName: 'bash', sessionId: '', status: 'error', consecutiveErrors: 3,
    });
    expect(failure).toContain('# Tool Execution Failure: `bash`');
    expect(failure).toContain('In session `unknown-session`,');
    expect(failure).toContain('failed with 3 consecutive failures on this tool.');
    expect(failure).toContain('(No output recorded)');
  });

  it('preserves execution issue evidence while keeping only the summary as heading metadata', () => {
    const description = '  failed, with [CRITICAL] TOOL_HIGH_FAILURE_RATE for AL-42.  ';
    const details = 'stack: web/src/x.ts:9\nfix/FIX-3 => SeverityHigh\n';
    const prompt = generateExecutionIssuePrompt({
      summary: '  [CRITICAL] Issue\nAL-42  ', sessionId: 'session-4', description, errorDetails: details,
    });

    expect(prompt).toContain('# Agent Execution Issue: [CRITICAL] Issue AL-42');
    assertFencedVerbatim(prompt, description);
    assertFencedVerbatim(prompt, details);
    expect(prompt).not.toContain('OpenCode database');
  });

  it('normalizes only heading metadata and uses safe fallbacks when it is empty', () => {
    const normalized = generateExecutionIssuePrompt({
      summary: 'Build failure\nfor AL-42', sessionId: 'session-4', description: '',
      role: ' coder ', model: ' model-x ',
    });
    expect(normalized).toContain('# Agent Execution Issue: Build failure for AL-42');
    expect(normalized).toContain('(agent `coder` running `model-x`)');
    assertFencedVerbatim(normalized, 'No description recorded.');
    expect(normalized).toContain('(No additional error details recorded)');

    const fallback = generateExecutionIssuePrompt({ summary: ' \n ', sessionId: 'session-4', description: 'detail' });
    expect(fallback).toContain('# Agent Execution Issue: Operational Execution Issue');
    assertFencedVerbatim(fallback, 'detail');
  });

  it('keeps deterministic rule labels for known IDs and passes unknown IDs through', () => {
    expect(translateRuleToPlainEnglish('TOOL_RETRY_STREAK', 'bash'))
      .toBe('Consecutive tool failures on `bash`');
    expect(translateRuleToPlainEnglish('CONTEXT_SATURATION_SPIKE'))
      .toBe('Context window saturation approach');
    expect(translateRuleToPlainEnglish('UNRECOGNIZED_RULE'))
      .toBe('UNRECOGNIZED_RULE');
  });

  it('preserves reasoning and message bodies byte-for-byte in code fences', () => {
    const reasoning = 'Plan: inspect AL-42, then run fix/FIX-3.\n[WARNING] Keep this exact.\n';
    const message = 'The parser reports ",, and or"; rule=TOOL_RETRY_STREAK.  ';
    const prompt = generateMessageTurnPrompt({
      turn: 2, role: 'assistant', sessionId: 'session-5', reasoningContent: reasoning, messageContent: message,
    });

    assertFencedVerbatim(prompt, reasoning);
    assertFencedVerbatim(prompt, message);
    expect(prompt).toContain('(No tools executed)');
  });

  it('preserves flagged issue evidence and handles sessions without issues', () => {
    const issue = '[CRITICAL] AL-42 on fix/FIX-3: TOOL_RETRY_STREAK, unchanged.\n';
    const withIssue = generateSessionOverviewPrompt({
      sessionId: 'session-6', role: 'tester', model: 'model-x', status: 'error',
      turnCount: 4, totalToolCalls: 7, failedToolCalls: 2, flaggedIssues: [issue],
    });
    assertFencedVerbatim(withIssue, issue);

    const withoutIssue = generateSessionOverviewPrompt({
      sessionId: 'session-7', role: 'orchestrator', model: 'model-x', status: 'completed',
      turnCount: 0, totalToolCalls: 0, failedToolCalls: 0, flaggedIssues: [],
    });
    expect(withoutIssue).toContain('(No flagged issues recorded)');
    expect(withoutIssue).toContain('after 0 turns and 0 tool executions (0 failed)');
  });
});

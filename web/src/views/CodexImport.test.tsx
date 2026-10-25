import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import * as apiClient from '../api/client';
import type { CodexTranscriptSummary } from '../domain/types';
import { mockAgenticMetricsLedgers } from '../test/fixtures/mockData';
import { CodexImport } from './CodexImport';

const summary: CodexTranscriptSummary = {
  thread_id: 'thread-demo',
  model: 'gpt-5-codex',
  provider: 'openai',
  source: 'vscode',
  usage: { input: 1200, output: 340, cache_read: 800, cache_write: 0, reasoning: 210 },
  context_window: 128000,
  record_kinds: { session_meta: 1, event_msg: 4 },
  event_kinds: { token_count: 2, task_complete: 1 },
  warnings: ['One record had no usage details'],
};

describe('CodexImport', () => {
  beforeEach(() => vi.restoreAllMocks());

  it('selects a rollout, generates an AgenticLens ledger, and discloses unavailable metrics', async () => {
    const baseLedger = mockAgenticMetricsLedgers['AL-1']!;
    const ledger = {
      ...baseLedger,
      ticket_id: 'AL-56',
      facts: { ...baseLedger.facts, ticket_id: 'AL-56' },
      totals: { ...baseLedger.totals, invocations: 1, input_tokens: 400, cache_read_tokens: 800, output_tokens: 340 },
      workflow_efficiency: {
        ...baseLedger.workflow_efficiency,
        active_duration_formatted: '4s',
        tool_execution: { ...baseLedger.workflow_efficiency.tool_execution, total_calls: 2, errors_and_corrections: 1 },
      },
    };
    vi.spyOn(apiClient, 'extractCodexRollout').mockResolvedValue({ data: { transcript: summary, agentic_metrics: ledger } });
    render(<CodexImport />);

    const input = screen.getByLabelText('Codex rollout file');
    fireEvent.change(input, { target: { files: [new File(['{}\\n'], 'thread.jsonl', { type: 'application/jsonl' })] } });
    expect(screen.getByText('thread.jsonl')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Generate AgenticLens metrics' }));

    expect(await screen.findByRole('heading', { name: 'Metrics for AL-56' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Download agentic_metrics.json' })).toBeInTheDocument();
    expect(screen.getByText('gpt-5-codex · openai')).toBeInTheDocument();
    expect(screen.getByText('1,200')).toBeInTheDocument();
    expect(screen.getByText(/Codex does not report billable USD cost/)).toBeInTheDocument();
    expect(screen.getByText(/Capacity is not consumed context/)).toBeInTheDocument();
    expect(screen.getByText('One record had no usage details')).toBeInTheDocument();
    expect(screen.getByText(/Transcript text and tool payloads are not returned or saved/)).toBeInTheDocument();
    expect(apiClient.extractCodexRollout).toHaveBeenCalledWith(expect.objectContaining({ name: 'thread.jsonl' }), 'AL-56', '', 'chore');
  });

  it('rejects a non-JSONL file before extraction', async () => {
    const extract = vi.spyOn(apiClient, 'extractCodexRollout');
    render(<CodexImport />);
    fireEvent.change(screen.getByLabelText('Codex rollout file'), {
      target: { files: [new File(['{}'], 'rollout.json')] },
    });

    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('Choose an uncompressed Codex .jsonl'));
    expect(extract).not.toHaveBeenCalled();
    expect(screen.getByRole('button', { name: 'Generate AgenticLens metrics' })).toBeDisabled();
  });
});

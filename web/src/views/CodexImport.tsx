import React, { useState } from 'react';
import { Download, FileJson2, Upload } from 'lucide-react';
import { extractCodexRollout } from '../api/client';
import type { AgenticMetricsLedger, CodexExtractionResult } from '../domain/types';

const MAX_FILE_BYTES = 64 * 1024 * 1024;

function formatCount(value: number): string {
  return new Intl.NumberFormat().format(value);
}

function formatFileSize(value: number): string {
  if (value < 1024 * 1024) return `${Math.max(1, Math.round(value / 1024))} KB`;
  return `${(value / (1024 * 1024)).toFixed(1)} MB`;
}

function ledgerFilename(ledger: AgenticMetricsLedger): string {
  const thread = (ledger.facts.invocations[0]?.id || 'rollout').replace(/[^a-zA-Z0-9_-]/g, '_');
  const ticket = (ledger.ticket_id || 'session').replace(/[^a-zA-Z0-9_-]/g, '_');
  return `agentic_metrics-${ticket}-codex-${thread}.json`;
}

export const CodexImport: React.FC = () => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [ticketId, setTicketId] = useState('GEN-07');
  const [ticketTitle, setTicketTitle] = useState('');
  const [result, setResult] = useState<CodexExtractionResult | null>(null);
  const [isExtracting, setIsExtracting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.currentTarget.files?.[0] ?? null;
    setResult(null);
    setErrorMessage(null);
    if (file && !file.name.toLowerCase().endsWith('.jsonl')) {
      setSelectedFile(null);
      setErrorMessage('Choose an uncompressed Codex .jsonl rollout file.');
      return;
    }
    if (file && file.size > MAX_FILE_BYTES) {
      setSelectedFile(null);
      setErrorMessage('Codex rollout files must be 64 MiB or smaller.');
      return;
    }
    setSelectedFile(file);
  };

  const handleExtract = async () => {
    if (!selectedFile) return;
    setIsExtracting(true);
    setErrorMessage(null);
    setResult(null);
    try {
      const response = await extractCodexRollout(selectedFile, ticketId, ticketTitle, 'chore');
      setResult(response.data);
    } catch (error: unknown) {
      setErrorMessage(error instanceof Error ? error.message : 'Could not extract this Codex rollout.');
    } finally {
      setIsExtracting(false);
    }
  };

  const handleDownload = () => {
    if (!result) return;
    const ledger = result.agentic_metrics;
    const blob = new Blob([JSON.stringify(ledger, null, 2)], { type: 'application/json' });
    const downloadUrl = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = downloadUrl;
    anchor.download = ledgerFilename(ledger);
    document.body.append(anchor);
    anchor.click();
    anchor.remove();
    URL.revokeObjectURL(downloadUrl);
  };

  const summary = result?.transcript;
  const ledger = result?.agentic_metrics;
  const recordKinds = Object.entries(summary?.record_kinds ?? {}).sort(([left], [right]) => left.localeCompare(right));
  const eventKinds = Object.entries(summary?.event_kinds ?? {}).sort(([left], [right]) => left.localeCompare(right));

  return (
    <section className="codex-import-view" data-testid="codex-import-view">
      <div className="codex-import-heading">
        <div>
          <p className="codex-import-eyebrow">CODEX SESSION METRICS</p>
          <h1>Codex Import</h1>
          <p className="codex-import-intro">
            Select a Codex rollout and generate the AgenticLens metrics ledger for its ticket.
          </p>
        </div>
      </div>

      <div className="codex-upload-card">
        <div className="codex-upload-title-row">
          <div className="codex-upload-icon" aria-hidden="true"><Upload size={19} /></div>
          <div>
            <h2>Select a rollout file</h2>
            <p>Uncompressed .jsonl files up to 64 MiB</p>
          </div>
        </div>
        <div className="codex-ticket-fields">
          <label htmlFor="codex-ticket-id">AgenticLens ticket ID</label>
          <input id="codex-ticket-id" value={ticketId} onChange={(event) => setTicketId(event.currentTarget.value)} />
          <label htmlFor="codex-ticket-title">Ticket title <span>(optional)</span></label>
          <input id="codex-ticket-title" value={ticketTitle} onChange={(event) => setTicketTitle(event.currentTarget.value)} />
        </div>
        <label className="codex-file-label" htmlFor="codex-rollout-file">Codex rollout file</label>
        <input
          id="codex-rollout-file"
          className="codex-file-input"
          type="file"
          accept=".jsonl,application/x-ndjson,application/jsonl"
          onChange={handleFileChange}
          aria-describedby="codex-file-help"
        />
        <p id="codex-file-help" className="codex-file-help">
          The selected file is sent to this AgenticLens server for in-memory extraction. Transcript text and tool payloads are not returned or saved.
        </p>
        {selectedFile && (
          <div className="codex-selected-file" role="status">
            <FileJson2 size={17} aria-hidden="true" />
            <span>{selectedFile.name}</span>
            <span className="codex-file-size">{formatFileSize(selectedFile.size)}</span>
          </div>
        )}
        {errorMessage && <p className="codex-error" role="alert">{errorMessage}</p>}
        <button
          type="button"
          className="codex-extract-button"
          onClick={() => void handleExtract()}
          disabled={!selectedFile || isExtracting || ticketId.trim() === ''}
        >
          {isExtracting ? 'Extracting…' : 'Generate AgenticLens metrics'}
        </button>
      </div>

      {result && summary && ledger && (
        <div className="codex-results" aria-live="polite">
          <div className="codex-results-header">
            <div>
              <p className="codex-import-eyebrow">SCHEMA-V3 LEDGER PREVIEW</p>
              <h2>Metrics for {ledger.ticket_id}</h2>
            </div>
            <button type="button" className="codex-download-button" onClick={handleDownload}>
              <Download size={16} aria-hidden="true" /> Download agentic_metrics.json
            </button>
          </div>

          <div className="codex-usage-grid">
            <div className="codex-usage-card"><span>Invocations</span><strong>{formatCount(ledger.totals.invocations)}</strong></div>
            <div className="codex-usage-card"><span>Uncached input</span><strong>{formatCount(ledger.totals.input_tokens)}</strong></div>
            <div className="codex-usage-card"><span>Cached input</span><strong>{formatCount(ledger.totals.cache_read_tokens)}</strong></div>
            <div className="codex-usage-card"><span>Output</span><strong>{formatCount(ledger.totals.output_tokens)}</strong></div>
            <div className="codex-usage-card"><span>Tool calls / errors</span><strong>{formatCount(ledger.workflow_efficiency.tool_execution.total_calls)} / {formatCount(ledger.workflow_efficiency.tool_execution.errors_and_corrections)}</strong></div>
            <div className="codex-usage-card"><span>Active time</span><strong>{ledger.workflow_efficiency.active_duration_formatted}</strong></div>
          </div>
          <p className="codex-limit-note">
            Codex does not report billable USD cost in this rollout. Schema-v3 stores cost as 0, meaning unavailable here, not free. This file extraction does not join Git analysis, so its code facts remain zero; transcript text is not used to infer code changes.
          </p>

          <div className="codex-metadata-grid">
            <div className="codex-metadata-card"><span>Thread</span><strong>{summary.thread_id || 'Unavailable'}</strong></div>
            <div className="codex-metadata-card"><span>Model / provider</span><strong>{[summary.model, summary.provider].filter(Boolean).join(' · ') || 'Unavailable'}</strong></div>
            <div className="codex-metadata-card"><span>Source</span><strong>{summary.source || 'Unavailable'}</strong></div>
            <div className="codex-metadata-card"><span>Started</span><strong>{summary.started_at || 'Unavailable'}</strong></div>
          </div>

          <h3 className="codex-section-title">Codex reported usage</h3>
          <div className="codex-usage-grid">
            <div className="codex-usage-card"><span>Total input</span><strong>{formatCount(summary.usage.input)}</strong></div>
            <div className="codex-usage-card"><span>Output</span><strong>{formatCount(summary.usage.output)}</strong></div>
            <div className="codex-usage-card"><span>Cached input</span><strong>{formatCount(summary.usage.cache_read)}</strong></div>
            <div className="codex-usage-card"><span>Cache write</span><strong>{formatCount(summary.usage.cache_write)}</strong></div>
            <div className="codex-usage-card"><span>Reasoning output</span><strong>{formatCount(summary.usage.reasoning)}</strong></div>
          </div>

          <p className="codex-limit-note">
            {summary.context_window ? `Context window capacity: ${formatCount(summary.context_window)} tokens. ` : ''}
            Capacity is not consumed context. Cached input is a subset of Codex input and is separated from uncached input in AgenticLens totals.
          </p>

          <div className="codex-kinds-grid">
            <div className="codex-kinds-card">
              <h3>Rollout records</h3>
              {recordKinds.length > 0 ? <ul>{recordKinds.map(([kind, count]) => <li key={kind}><span>{kind}</span><strong>{formatCount(count)}</strong></li>)}</ul> : <p>No record types found.</p>}
            </div>
            <div className="codex-kinds-card">
              <h3>Event types</h3>
              {eventKinds.length > 0 ? <ul>{eventKinds.map(([kind, count]) => <li key={kind}><span>{kind}</span><strong>{formatCount(count)}</strong></li>)}</ul> : <p>No event types found.</p>}
            </div>
          </div>

          {summary.warnings && summary.warnings.length > 0 && (
            <div className="codex-warning-panel" role="status">
              <h3>Parser warnings</h3>
              <ul>{summary.warnings.map((warning, index) => <li key={`${index}-${warning}`}>{warning}</li>)}</ul>
            </div>
          )}
        </div>
      )}
    </section>
  );
};

import { readFleetSummaryResponse, readMetricsLedgerArrayResponse, readWorkspaceDetailResponse, readComparisonRowsResponse, readTimeBucketsResponse } from './costValidation';
import type {
  AgenticMetricsLedger,
  Anomaly,
  BuildInfo,
  EventFact,
  FleetSummary,
  ProjectConfigResponse,
  TicketComparisonRow,
  TimeBucketMetric,
  TimeGranularity,
  TimeWindowPreset,
  WorkspaceDetail,
  WorkspaceEventsResponse,
  WorkspaceDiagnostic,
} from '../domain/types';
import { deriveTicketComparisonRows } from '../domain/formatters';

export interface ApiResponse<T> {
  readonly data: T;
  readonly diagnostics?: readonly WorkspaceDiagnostic[];
}

function diagnosticsFromResponse(res: Response): readonly WorkspaceDiagnostic[] {
  if (!res.headers) return [];
  const raw = res.headers.get('X-AgenticLens-Workspace-Diagnostics');
  if (!raw) return [];
  try {
    const value = JSON.parse(raw) as readonly WorkspaceDiagnostic[];
    return Array.isArray(value) ? value : [];
  } catch {
    return [];
  }
}

function appendMultiParams(params: URLSearchParams, key: string, values?: readonly string[]): void {
  if (!values || values.length === 0) return;
  const filtered = values.map((v) => v.trim()).filter((v) => v !== '' && v.toLowerCase() !== 'all');
  if (filtered.length === 0) return;
  for (const v of filtered) {
    params.append(key, v);
  }
}

function rethrowError(err: unknown, fallbackMessage: string): never {
  if (err instanceof Error) {
    throw err;
  }
  if (typeof err === 'object' && err !== null && 'message' in err && typeof (err as { message: unknown }).message === 'string') {
    throw err as Error;
  }
  throw new Error(fallbackMessage);
}

export async function fetchFleetSummary(
  windowParam: string = 'all',
  sinceParam?: string,
  untilParam?: string,
  ticketTypes?: readonly string[],
  prefixes?: readonly string[]
): Promise<ApiResponse<FleetSummary>> {
  const params = new URLSearchParams();
  if (windowParam && windowParam !== 'all') {
    params.set('window', windowParam);
  }
  if (sinceParam) {
    params.set('since', sinceParam);
  }
  if (untilParam) {
    params.set('until', untilParam);
  }
  appendMultiParams(params, 'ticket_type', ticketTypes);
  appendMultiParams(params, 'prefix', prefixes);

  const query = params.toString() ? `?${params.toString()}` : '';
  const url = `/api/fleet${query}`;

  try {
    const res = await fetch(url);
    if (!res.ok) {
      throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch fleet summary'}`);
    }
    const data = await readFleetSummaryResponse(res);
    return { data, diagnostics: diagnosticsFromResponse(res) };
  } catch (err: unknown) {
    rethrowError(err, 'Failed to fetch fleet summary');
  }
}

export async function fetchWorkspaceLedgers(
  windowParam: string = 'all',
  sinceParam?: string,
  untilParam?: string,
  ticketTypes?: readonly string[],
  prefixes?: readonly string[]
): Promise<ApiResponse<readonly AgenticMetricsLedger[]>> {
  const params = new URLSearchParams();
  if (windowParam && windowParam !== 'all') {
    params.set('window', windowParam);
  }
  if (sinceParam) {
    params.set('since', sinceParam);
  }
  if (untilParam) {
    params.set('until', untilParam);
  }
  appendMultiParams(params, 'ticket_type', ticketTypes);
  appendMultiParams(params, 'prefix', prefixes);

  const query = params.toString() ? `?${params.toString()}` : '';
  const url = `/api/workspaces${query}`;

  try {
    const res = await fetch(url);
    if (!res.ok) {
      throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch workspace ledgers'}`);
    }
    const data = await readMetricsLedgerArrayResponse(res);
    return { data, diagnostics: diagnosticsFromResponse(res) };
  } catch (err: unknown) {
    rethrowError(err, 'Failed to fetch workspace ledgers');
  }
}

export async function fetchWorkspaces(
  windowParam: string = 'all',
  sortParam: string = '',
  orderParam: string = '',
  sinceParam?: string,
  untilParam?: string,
  ticketTypes?: readonly string[],
  prefixes?: readonly string[]
): Promise<ApiResponse<TicketComparisonRow[]>> {
  const params = new URLSearchParams();
  if (windowParam && windowParam !== 'all') {
    params.set('window', windowParam);
  }
  if (sortParam) {
    params.set('sort', sortParam);
  }
  if (orderParam) {
    params.set('order', orderParam);
  }
  if (sinceParam) {
    params.set('since', sinceParam);
  }
  if (untilParam) {
    params.set('until', untilParam);
  }
  appendMultiParams(params, 'ticket_type', ticketTypes);
  appendMultiParams(params, 'prefix', prefixes);

  const qs = params.toString() ? `?${params.toString()}` : '';
  const url = `/api/workspaces${qs}`;

  try {
    const res = await fetch(url);
    if (!res.ok) {
      throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch workspaces'}`);
    }
    const ledgers = await readMetricsLedgerArrayResponse(res);
    const rows = deriveTicketComparisonRows(ledgers);
    return { data: rows, diagnostics: diagnosticsFromResponse(res) };
  } catch (err: unknown) {
    rethrowError(err, 'Failed to fetch workspaces');
  }
}

export async function fetchComparisonRows(
  windowParam: string = 'all',
  sinceParam?: string,
  untilParam?: string,
  ticketTypes?: readonly string[],
  prefixes?: readonly string[]
): Promise<ApiResponse<readonly TicketComparisonRow[]>> {
  const params = new URLSearchParams();
  if (windowParam && windowParam !== 'all') {
    params.set('window', windowParam);
  }
  if (sinceParam) {
    params.set('since', sinceParam);
  }
  if (untilParam) {
    params.set('until', untilParam);
  }
  appendMultiParams(params, 'ticket_type', ticketTypes);
  appendMultiParams(params, 'prefix', prefixes);

  const qs = params.toString() ? `?${params.toString()}` : '';
  const url = `/api/comparison${qs}`;

  try {
    const res = await fetch(url);
    if (!res.ok) {
      throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch comparison rows'}`);
    }
    const data = await readComparisonRowsResponse(res);
    return { data, diagnostics: diagnosticsFromResponse(res) };
  } catch (err: unknown) {
    rethrowError(err, 'Failed to fetch comparison rows');
  }
}

export async function fetchProgressionLedgers(
  windowParam: string = 'all',
  sinceParam?: string,
  untilParam?: string,
  ticketTypes?: readonly string[],
  prefixes?: readonly string[]
): Promise<ApiResponse<readonly AgenticMetricsLedger[]>> {
  const params = new URLSearchParams();
  if (windowParam && windowParam !== 'all') {
    params.set('window', windowParam);
  }
  if (sinceParam) {
    params.set('since', sinceParam);
  }
  if (untilParam) {
    params.set('until', untilParam);
  }
  appendMultiParams(params, 'ticket_type', ticketTypes);
  appendMultiParams(params, 'prefix', prefixes);

  const qs = params.toString() ? `?${params.toString()}` : '';
  const url = `/api/progression${qs}`;

  try {
    const res = await fetch(url);
    if (!res.ok) {
      throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch progression ledgers'}`);
    }
    const data = await readMetricsLedgerArrayResponse(res);
    return { data, diagnostics: diagnosticsFromResponse(res) };
  } catch (err: unknown) {
    rethrowError(err, 'Failed to fetch progression ledgers');
  }
}

export async function fetchWorkspaceById(
  ticketId: string
): Promise<ApiResponse<WorkspaceDetail>> {
  const url = `/api/workspaces/${encodeURIComponent(ticketId)}`;
  try {
    const res = await fetch(url);
    if (!res.ok) {
      if (res.status === 404) {
        throw new Error(`Workspace ${ticketId} not found`);
      }
      throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch workspace'}`);
    }
    const data = await readWorkspaceDetailResponse(res);
    return { data };
  } catch (err: unknown) {
    rethrowError(err, `Failed to fetch workspace ${ticketId}`);
  }
}

export async function fetchVersion(): Promise<ApiResponse<BuildInfo>> {
  try {
    const res = await fetch('/api/version');
    if (!res.ok) {
      throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch build version info'}`);
    }
    const data = (await res.json()) as BuildInfo;
    return { data };
  } catch (err: unknown) {
    rethrowError(err, 'Failed to fetch build version info');
  }
}

export async function fetchProjectConfig(): Promise<ApiResponse<ProjectConfigResponse>> {
  try {
    const res = await fetch('/api/config');
    if (!res.ok) {
      throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch project configuration'}`);
    }
    const data = (await res.json()) as ProjectConfigResponse;
    return { data };
  } catch (err: unknown) {
    rethrowError(err, 'Failed to fetch project configuration');
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function isOptionalString(value: unknown): boolean {
  return value === undefined || value === null || typeof value === 'string';
}

function isOptionalNumber(value: unknown): boolean {
  return value === undefined || (typeof value === 'number' && Number.isFinite(value));
}

function isEventFact(value: unknown): value is EventFact {
  if (!isRecord(value)) return false;
  return typeof value.timestamp === 'string' && typeof value.session_id === 'string' &&
    typeof value.turn === 'number' && Number.isFinite(value.turn) &&
    typeof value.role === 'string' && typeof value.status === 'string' &&
    (value.event_type === 'tool_call' || value.event_type === 'turn_message' || value.event_type === 'compaction') &&
    typeof value.context_tokens === 'number' && Number.isFinite(value.context_tokens) &&
    ['event_id', 'message_id', 'tool_name', 'error_message', 'prompt_content',
      'reasoning_content', 'tool_input', 'tool_output'].every((field) => isOptionalString(value[field])) &&
    (value.duration_ms === null || isOptionalNumber(value.duration_ms));
}

function isAnomaly(value: unknown): value is Anomaly {
  if (!isRecord(value) || !isRecord(value.context)) return false;
  if (value.severity !== 'WARNING' && value.severity !== 'CRITICAL') return false;
  if (value.rule_id !== 'TOOL_RETRY_STREAK' && value.rule_id !== 'TOOL_HIGH_FAILURE_RATE' &&
    value.rule_id !== 'CONTEXT_SATURATION_SPIKE' && value.rule_id !== 'SUBAGENT_REWORK_LOOP') return false;
  const context = value.context;
  if (!['tool_name', 'session_id', 'threshold'].every((field) =>
    context[field] === undefined || typeof context[field] === 'string')) return false;
  if (!isOptionalString(context.role) || !isOptionalString(context.secondary_role)) return false;
  if (!['turn', 'count', 'total_calls', 'token_delta', 'context_tokens', 'failure_rate_pct'].every((field) =>
    isOptionalNumber(context[field]))) return false;
  return value.evidence === undefined || (Array.isArray(value.evidence) && value.evidence.every((ref: unknown) =>
    isRecord(ref) && typeof ref.session_id === 'string' && typeof ref.event_id === 'string'));
}

export async function fetchWorkspaceEvents(
  ticketId: string
): Promise<ApiResponse<WorkspaceEventsResponse>> {
  const url = `/api/workspaces/${encodeURIComponent(ticketId)}/events`;
  try {
    const res = await fetch(url);
    if (!res.ok) {
      if (res.status === 404) {
        throw new Error(`Workspace ${ticketId} not found`);
      }
      throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch workspace events'}`);
    }
    const rawData = (await res.json()) as unknown;
    if (
      rawData !== null &&
      typeof rawData === 'object' &&
      !Array.isArray(rawData) &&
      'is_hydrated' in rawData
    ) {
      const resp = rawData as Record<string, unknown>;
      if (resp.ticket_id !== ticketId || typeof resp.is_hydrated !== 'boolean' ||
        (resp.includes_baseline !== undefined && typeof resp.includes_baseline !== 'boolean') ||
        (resp.data_source !== 'committed_ledger' && resp.data_source !== 'local_sqlite' && resp.data_source !== 'local_codex') ||
        !Array.isArray(resp.events) || !resp.events.every(isEventFact) ||
        (resp.anomalies != null && (!Array.isArray(resp.anomalies) || !resp.anomalies.every(isAnomaly)))) {
        throw new Error(`Invalid workspace events response for ${ticketId}`);
      }
      return {
        data: {
          ticket_id: ticketId,
          is_hydrated: resp.is_hydrated,
          events: resp.events,
          data_source: resp.data_source,
          ...(resp.includes_baseline === undefined ? {} : { includes_baseline: resp.includes_baseline }),
          ...(resp.anomalies == null ? {} : { anomalies: resp.anomalies as readonly Anomaly[] }),
        },
      };
    }
    if (Array.isArray(rawData)) {
      if (!rawData.every(isEventFact)) throw new Error(`Invalid workspace events response for ${ticketId}`);
      return {
        data: {
          ticket_id: ticketId,
          is_hydrated: false,
          events: rawData,
          data_source: 'committed_ledger',
        },
      };
    }
    return {
      data: {
        ticket_id: ticketId,
        is_hydrated: false,
        events: [],
        data_source: 'committed_ledger',
      },
    };
  } catch (err: unknown) {
    rethrowError(err, `Failed to fetch workspace events for ${ticketId}`);
  }
}

export async function fetchWorkspaceAnomalies(
  ticketId: string
): Promise<ApiResponse<readonly Anomaly[]>> {
  try {
    const detailRes = await fetchWorkspaceById(ticketId);
    return {
      data: detailRes.data.anomalies ?? [],
    };
  } catch (err: unknown) {
    rethrowError(err, `Failed to fetch workspace anomalies for ${ticketId}`);
  }
}

export async function fetchProgressionMetrics(params?: {
  readonly window?: TimeWindowPreset;
  readonly granularity?: TimeGranularity;
  readonly startDate?: string;
  readonly endDate?: string;
  readonly ticketTypes?: readonly string[];
  readonly prefixes?: readonly string[];
}): Promise<ApiResponse<readonly TimeBucketMetric[]>> {
  const windowParam = params?.window ?? 'all';
  const granularityParam = params?.granularity ?? 'daily';
  const startDateParam = params?.startDate;
  const endDateParam = params?.endDate;
  const ticketTypes = params?.ticketTypes;
  const prefixes = params?.prefixes;

  const query = new URLSearchParams();
  if (windowParam && windowParam !== 'all') {
    query.set('window', windowParam);
  }
  if (granularityParam) {
    query.set('granularity', granularityParam);
  }
  if (startDateParam) {
    query.set('startDate', startDateParam);
  }
  if (endDateParam) {
    query.set('endDate', endDateParam);
  }
  appendMultiParams(query, 'ticket_type', ticketTypes);
  appendMultiParams(query, 'prefix', prefixes);

  const qs = query.toString() ? `?${query.toString()}` : '';
  const url = `/api/progression${qs}`;

  try {
    const res = await fetch(url);
    if (!res.ok) {
      throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch progression metrics'}`);
    }
    const data = await readTimeBucketsResponse(res);
    return { data };
  } catch (err: unknown) {
    rethrowError(err, 'Failed to fetch progression metrics');
  }
}

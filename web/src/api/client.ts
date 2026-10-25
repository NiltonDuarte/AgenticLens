import type {
  AgenticMetricsLedger,
  Anomaly,
  BuildInfo,
  CodexExtractionResult,
  EventFact,
  FleetSummary,
  ProjectConfigResponse,
  TicketComparisonRow,
  TimeBucketMetric,
  TimeGranularity,
  TimeWindowPreset,
  WorkspaceDetail,
  WorkspaceEventsResponse,
} from '../domain/types';
import { deriveTicketComparisonRows } from '../domain/formatters';

export interface ApiResponse<T> {
  readonly data: T;
}

export async function extractCodexRollout(
  file: File,
  ticketId = 'AL-56',
  ticketTitle = '',
  ticketType = 'chore'
): Promise<ApiResponse<CodexExtractionResult>> {
  if (file.size > 64 * 1024 * 1024) {
    throw new Error('Codex rollout files must be 64 MiB or smaller');
  }
  try {
    const params = new URLSearchParams({ ticket_id: ticketId, ticket_type: ticketType });
    if (ticketTitle.trim() !== '') params.set('ticket_title', ticketTitle.trim());
    const res = await fetch(`/api/codex/extract?${params.toString()}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-ndjson' },
      body: file,
    });
    if (!res.ok) {
      let message = `HTTP error ${res.status}: ${res.statusText || 'Failed to extract Codex rollout'}`;
      try {
        const payload = (await res.json()) as { readonly message?: string };
        if (payload.message) message = payload.message;
      } catch {
        // Keep the HTTP status message when the server response is not JSON.
      }
      throw new Error(message);
    }
    const data = (await res.json()) as CodexExtractionResult;
    return { data };
  } catch (err: unknown) {
    rethrowError(err, 'Failed to extract Codex rollout');
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
    const data = (await res.json()) as FleetSummary;
    return { data };
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
    const ledgers = (await res.json()) as readonly AgenticMetricsLedger[];
    return { data: ledgers };
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
    const ledgers = (await res.json()) as AgenticMetricsLedger[];
    const rows = deriveTicketComparisonRows(ledgers);
    return { data: rows };
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
    const rows = (await res.json()) as readonly TicketComparisonRow[];
    return { data: rows };
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
    const ledgers = (await res.json()) as readonly AgenticMetricsLedger[];
    return { data: ledgers };
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
    const data = (await res.json()) as WorkspaceDetail;
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
      const resp = rawData as WorkspaceEventsResponse;
      return { data: resp };
    }
    if (Array.isArray(rawData)) {
      return {
        data: {
          ticket_id: ticketId,
          is_hydrated: false,
          events: rawData as readonly EventFact[],
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
    const data = (await res.json()) as readonly TimeBucketMetric[];
    return { data };
  } catch (err: unknown) {
    rethrowError(err, 'Failed to fetch progression metrics');
  }
}

import type {
  AgenticMetricsLedger,
  Anomaly,
  BuildInfo,
  FleetSummary,
  ProjectConfigResponse,
  TicketComparisonRow,
  WorkspaceDetail,
  WorkspaceEventsResponse,
  WorkspaceDiagnostic,
} from '../domain/types';
import { deriveTicketComparisonRows } from '../domain/formatters';
import { decodeDetail, decodeFleet, decodeLedgers, isAnomaly, isBuildInfo, isEventFact, isProjectConfig, isRecord } from './validation';

export interface ApiResponse<T> {
  readonly data: T;
  readonly diagnostics?: readonly WorkspaceDiagnostic[];
}

function diagnosticsFromResponse(res: Response): readonly WorkspaceDiagnostic[] {
  if (!res.headers) return [];
  const raw = res.headers.get('X-AgenticLens-Workspace-Diagnostics');
  if (!raw) return [];
  const warning: WorkspaceDiagnostic = { scope: 'response', message: 'Workspace diagnostics could not be read. Valid workspace data is still shown.' };
  try {
    const value: unknown = JSON.parse(raw);
    if (!Array.isArray(value)) return [warning];
    const valid: WorkspaceDiagnostic[] = [];
    let invalid = false;
    for (const item of value) {
      if (isRecord(item) && typeof item.ticket_id === 'string' && typeof item.message === 'string') {
        valid.push({ ticket_id: item.ticket_id, message: item.message });
      } else invalid = true;
    }
    return invalid ? [...valid, warning] : valid;
  } catch {
    return [warning];
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
    const data = decodeFleet(await res.json());
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
    const data = decodeLedgers(await res.json());
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
    const ledgers = decodeLedgers(await res.json());
    const rows = deriveTicketComparisonRows(ledgers);
    return { data: rows, diagnostics: diagnosticsFromResponse(res) };
  } catch (err: unknown) {
    rethrowError(err, 'Failed to fetch workspaces');
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
    const data = decodeDetail(await res.json());
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
    const data: unknown = await res.json();
    if (!isBuildInfo(data)) throw new Error('Invalid build version response');
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
    const data: unknown = await res.json();
    if (!isProjectConfig(data)) throw new Error('Invalid project configuration response');
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
    const rawData: unknown = await res.json();
    if (
      isRecord(rawData) &&
      'is_hydrated' in rawData
    ) {
      const resp = rawData;
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
    throw new Error(`Invalid workspace events response for ${ticketId}`);
  } catch (err: unknown) {
    rethrowError(err, `Failed to fetch workspace events for ${ticketId}`);
  }
}

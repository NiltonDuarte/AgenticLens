import type { TagFilter } from '../domain/tags';
import type {
  AgenticMetricsLedger,
  BuildInfo,
  FleetSummary,
  ProjectConfigResponse,
  TicketComparisonRow,
  WorkspaceDetail,
  WorkspaceEventsResponse,
  WorkspaceDiagnostic,
} from '../domain/types';
import { deriveTicketComparisonRows } from '../domain/formatters';
import { decodeDetail, decodeFleet, decodeLedgers, decodeWorkspaceEvents, isBuildInfo, isProjectConfig, isRecord } from './validation';

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
      if (isRecord(item) && typeof item.message === 'string' &&
        (item.ticket_id === undefined || typeof item.ticket_id === 'string') &&
        (item.path === undefined || typeof item.path === 'string') &&
        (typeof item.path === 'string' && item.path.trim() !== '' || typeof item.ticket_id === 'string' && item.ticket_id.trim() !== '')) {
        valid.push({ message: item.message, ...(typeof item.ticket_id === 'string' ? { ticket_id: item.ticket_id } : {}),
          ...(typeof item.path === 'string' ? { path: item.path } : {}) });
      } else invalid = true;
    }
    return invalid ? [...valid, warning] : valid;
  } catch {
    return [warning];
  }
}

function appendMultiParams(params: URLSearchParams, key: string, values?: readonly string[]): void {
  if (!values || values.length === 0) return;
  const filtered = values.map((v) => v.trim()).filter((v) => v !== '' && (key === 'prefix' || v.toLowerCase() !== 'all'));
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
  prefixes?: readonly string[],
  tagFilters?: readonly TagFilter[]
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
  if (tagFilters?.length) params.set('tag_filter', JSON.stringify(tagFilters));

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
  prefixes?: readonly string[],
  tagFilters?: readonly TagFilter[]
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
  if (tagFilters?.length) params.set('tag_filter', JSON.stringify(tagFilters));

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
  prefixes?: readonly string[],
  tagFilters?: readonly TagFilter[]
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
  if (tagFilters?.length) params.set('tag_filter', JSON.stringify(tagFilters));

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
  ticketId: string,
  tagFilters?: readonly TagFilter[],
  workspacePath?: string
): Promise<ApiResponse<WorkspaceDetail>> {
  const params = new URLSearchParams();
  if (tagFilters?.length) params.set('tag_filter', JSON.stringify(tagFilters));
  if (workspacePath) params.set('workspace', workspacePath);
  const query = params.size ? `?${params}` : '';
  const url = `/api/workspaces/${encodeURIComponent(ticketId)}${query}`;
  try {
    const res = await fetch(url);
    if (!res.ok) {
      throw await workspaceResponseError(res, `Workspace ${ticketId}`, 'Failed to fetch workspace');
    }
    const data = decodeDetail(await res.json());
    if (workspacePath && data.workspace_path !== workspacePath) throw new Error(`Workspace ${ticketId} location changed; refresh and select its current location.`);
    return { data, diagnostics: diagnosticsFromResponse(res) };
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
  ticketId: string,
  workspacePath?: string
): Promise<ApiResponse<WorkspaceEventsResponse>> {
  const params = new URLSearchParams();
  if (workspacePath) params.set('workspace', workspacePath);
  const url = `/api/workspaces/${encodeURIComponent(ticketId)}/events${params.size ? `?${params}` : ''}`;
  try {
    const res = await fetch(url);
    if (!res.ok) {
      throw await workspaceResponseError(res, `Workspace ${ticketId}`, 'Failed to fetch workspace events');
    }
    const rawData: unknown = await res.json();
    return { data: decodeWorkspaceEvents(rawData, ticketId) };
  } catch (err: unknown) {
    rethrowError(err, `Failed to fetch workspace events for ${ticketId}`);
  }
}

async function workspaceResponseError(response: Response, target: string, fallback: string): Promise<Error> {
  try {
    const body: unknown = await response.json();
    if (isRecord(body) && typeof body.message === 'string' && body.message.trim()) return new Error(body.message);
  } catch { /* Retain HTTP context when the error body is absent or malformed. */ }
  return new Error(response.status === 404 ? `${target} not found` : `HTTP error ${response.status}: ${response.statusText || fallback}`);
}

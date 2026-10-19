import type {
  AgenticMetricsLedger,
  Anomaly,
  BuildInfo,
  EventFact,
  FleetSummary,
  TicketComparisonRow,
  TimeBucketMetric,
  TimeGranularity,
  TimeWindowPreset,
  WorkspaceDetail,
} from '../domain/types';
import {
  deriveTicketComparisonRows,
  filterLedgersByCompletionDate,
  filterLedgersByTicketTypes,
  filterLedgersByPrefixes,
} from '../domain/formatters';
import { aggregateMetricsOverTime } from '../domain/progression';
import {
  mockFleetSummary,
  mockTicketComparisonRows,
  mockAgenticMetricsLedgers,
  mockWorkspaceDetails,
  mockBuildInfo,
} from './mockData';

export interface ApiResponse<T> {
  readonly data: T;
  readonly isMock: boolean;
}

function appendMultiParams(params: URLSearchParams, key: string, values?: readonly string[]): void {
  if (!values || values.length === 0) return;
  const filtered = values.map((v) => v.trim()).filter((v) => v !== '' && v.toLowerCase() !== 'all');
  if (filtered.length === 0) return;
  for (const v of filtered) {
    params.append(key, v);
  }
}

function getFilteredMockLedgers(
  sinceParam?: string,
  untilParam?: string,
  ticketTypes?: readonly string[],
  prefixes?: readonly string[]
): readonly AgenticMetricsLedger[] {
  let list: readonly AgenticMetricsLedger[] = Object.values(mockAgenticMetricsLedgers);
  if (sinceParam || untilParam) {
    list = filterLedgersByCompletionDate(list, sinceParam, untilParam);
  }
  if (ticketTypes && ticketTypes.length > 0) {
    list = filterLedgersByTicketTypes(list, ticketTypes);
  }
  if (prefixes && prefixes.length > 0) {
    list = filterLedgersByPrefixes(list, prefixes);
  }
  return list;
}

export async function fetchFleetSummary(
  windowParam: string = 'all',
  sinceParam?: string,
  untilParam?: string,
  ticketTypes?: readonly string[],
  prefixes?: readonly string[]
): Promise<ApiResponse<FleetSummary>> {
  try {
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
    const res = await fetch(`/api/fleet${query}`);
    if (!res.ok) {
      return { data: mockFleetSummary, isMock: true };
    }
    const data = (await res.json()) as FleetSummary;
    return { data, isMock: false };
  } catch {
    return { data: mockFleetSummary, isMock: true };
  }
}

export async function fetchWorkspaceLedgers(
  windowParam: string = 'all',
  sinceParam?: string,
  untilParam?: string,
  ticketTypes?: readonly string[],
  prefixes?: readonly string[]
): Promise<ApiResponse<readonly AgenticMetricsLedger[]>> {
  try {
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
    const res = await fetch(`/api/workspaces${query}`);
    if (!res.ok) {
      return { data: getFilteredMockLedgers(sinceParam, untilParam, ticketTypes, prefixes), isMock: true };
    }
    const ledgers = (await res.json()) as AgenticMetricsLedger[];
    return { data: ledgers, isMock: false };
  } catch {
    return { data: getFilteredMockLedgers(sinceParam, untilParam, ticketTypes, prefixes), isMock: true };
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
  try {
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
    const res = await fetch(`/api/workspaces${qs}`);
    if (!res.ok) {
      const filtered = getFilteredMockLedgers(sinceParam, untilParam, ticketTypes, prefixes);
      if (sinceParam || untilParam || (ticketTypes && ticketTypes.length > 0) || (prefixes && prefixes.length > 0)) {
        return { data: deriveTicketComparisonRows(filtered), isMock: true };
      }
      return { data: [...mockTicketComparisonRows], isMock: true };
    }
    const ledgers = (await res.json()) as AgenticMetricsLedger[];
    const rows = deriveTicketComparisonRows(ledgers);
    return { data: rows, isMock: false };
  } catch {
    const filtered = getFilteredMockLedgers(sinceParam, untilParam, ticketTypes, prefixes);
    if (sinceParam || untilParam || (ticketTypes && ticketTypes.length > 0) || (prefixes && prefixes.length > 0)) {
      return { data: deriveTicketComparisonRows(filtered), isMock: true };
    }
    return { data: [...mockTicketComparisonRows], isMock: true };
  }
}

export async function fetchComparisonRows(
  windowParam: string = 'all',
  sinceParam?: string,
  untilParam?: string,
  ticketTypes?: readonly string[],
  prefixes?: readonly string[]
): Promise<ApiResponse<readonly TicketComparisonRow[]>> {
  try {
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
    const res = await fetch(`/api/comparison${qs}`);
    if (!res.ok) {
      const filtered = getFilteredMockLedgers(sinceParam, untilParam, ticketTypes, prefixes);
      return { data: deriveTicketComparisonRows(filtered), isMock: true };
    }
    const rows = (await res.json()) as readonly TicketComparisonRow[];
    return { data: rows, isMock: false };
  } catch {
    const filtered = getFilteredMockLedgers(sinceParam, untilParam, ticketTypes, prefixes);
    return { data: deriveTicketComparisonRows(filtered), isMock: true };
  }
}

export async function fetchProgressionLedgers(
  windowParam: string = 'all',
  sinceParam?: string,
  untilParam?: string,
  ticketTypes?: readonly string[],
  prefixes?: readonly string[]
): Promise<ApiResponse<readonly AgenticMetricsLedger[]>> {
  try {
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
    const res = await fetch(`/api/progression${qs}`);
    if (!res.ok) {
      const filtered = getFilteredMockLedgers(sinceParam, untilParam, ticketTypes, prefixes);
      return { data: filtered, isMock: true };
    }
    const ledgers = (await res.json()) as readonly AgenticMetricsLedger[];
    return { data: ledgers, isMock: false };
  } catch {
    const filtered = getFilteredMockLedgers(sinceParam, untilParam, ticketTypes, prefixes);
    return { data: filtered, isMock: true };
  }
}

export async function fetchWorkspaceById(
  ticketId: string
): Promise<ApiResponse<WorkspaceDetail>> {
  try {
    const res = await fetch(`/api/workspaces/${encodeURIComponent(ticketId)}`);
    if (!res.ok) {
      const fallback = mockWorkspaceDetails[ticketId];
      if (fallback) {
        return { data: fallback, isMock: true };
      }
      const ledgerFallback = mockAgenticMetricsLedgers[ticketId];
      if (ledgerFallback) {
        return {
          data: {
            ledger: ledgerFallback,
            artifacts: { ticket_id: ticketId, contract_content: null, summary_content: null },
            anomalies: ledgerFallback.anomalies ?? [],
          },
          isMock: true,
        };
      }
      throw new Error(`Workspace ${ticketId} not found`);
    }
    const data = (await res.json()) as WorkspaceDetail;
    return { data, isMock: false };
  } catch (err: unknown) {
    const fallback = mockWorkspaceDetails[ticketId];
    if (fallback) {
      return { data: fallback, isMock: true };
    }
    const ledgerFallback = mockAgenticMetricsLedgers[ticketId];
    if (ledgerFallback) {
      return {
        data: {
          ledger: ledgerFallback,
          artifacts: { ticket_id: ticketId, contract_content: null, summary_content: null },
          anomalies: ledgerFallback.anomalies ?? [],
        },
        isMock: true,
      };
    }
    const msg = err instanceof Error ? err.message : 'Unknown network failure';
    throw new Error(msg);
  }
}

export async function fetchVersion(): Promise<ApiResponse<BuildInfo>> {
  try {
    const res = await fetch('/api/version');
    if (!res.ok) {
      return { data: mockBuildInfo, isMock: true };
    }
    const data = (await res.json()) as BuildInfo;
    return { data, isMock: false };
  } catch {
    return { data: mockBuildInfo, isMock: true };
  }
}

export async function fetchWorkspaceEvents(
  ticketId: string
): Promise<ApiResponse<readonly EventFact[]>> {
  try {
    const res = await fetch(`/api/workspaces/${encodeURIComponent(ticketId)}/events`);
    if (!res.ok) {
      if (res.status === 404) {
        const fallbackLedger = mockAgenticMetricsLedgers[ticketId];
        if (fallbackLedger) {
          return { data: fallbackLedger.facts?.events ?? [], isMock: true };
        }
        throw new Error(`Workspace ${ticketId} not found`);
      }
      const fallbackEvents = mockAgenticMetricsLedgers[ticketId]?.facts?.events ?? [];
      return { data: fallbackEvents, isMock: true };
    }
    const data = (await res.json()) as readonly EventFact[];
    return { data, isMock: false };
  } catch (err: unknown) {
    if (err instanceof Error && err.message === `Workspace ${ticketId} not found`) {
      throw err;
    }
    const fallbackLedger = mockAgenticMetricsLedgers[ticketId];
    if (fallbackLedger) {
      return { data: fallbackLedger.facts?.events ?? [], isMock: true };
    }
    const msg = err instanceof Error ? err.message : 'Unknown network failure';
    throw new Error(msg);
  }
}

export async function fetchWorkspaceAnomalies(
  ticketId: string
): Promise<ApiResponse<readonly Anomaly[]>> {
  try {
    const detailRes = await fetchWorkspaceById(ticketId);
    return {
      data: detailRes.data.anomalies ?? [],
      isMock: detailRes.isMock,
    };
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : 'Unknown network failure';
    throw new Error(msg);
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

  const fallback = (): readonly TimeBucketMetric[] => {
    let mockList: readonly AgenticMetricsLedger[] = Object.values(mockAgenticMetricsLedgers);
    if (ticketTypes && ticketTypes.length > 0) {
      mockList = filterLedgersByTicketTypes(mockList, ticketTypes);
    }
    if (prefixes && prefixes.length > 0) {
      mockList = filterLedgersByPrefixes(mockList, prefixes);
    }
    return aggregateMetricsOverTime(
      mockList,
      granularityParam,
      startDateParam,
      endDateParam
    );
  };

  try {
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
    const res = await fetch(`/api/progression${qs}`);
    if (!res.ok) {
      return { data: fallback(), isMock: true };
    }
    const data = (await res.json()) as readonly TimeBucketMetric[];
    return { data, isMock: false };
  } catch {
    return { data: fallback(), isMock: true };
  }
}

import type {
  AgenticMetricsLedger,
  EventFact,
  FleetSummary,
  TicketComparisonRow,
  WorkspaceDetail,
} from '../domain/types';
import { deriveTicketComparisonRows } from '../domain/formatters';
import {
  mockFleetSummary,
  mockTicketComparisonRows,
  mockAgenticMetricsLedgers,
  mockWorkspaceDetails,
} from './mockData';

export interface ApiResponse<T> {
  readonly data: T;
  readonly isMock: boolean;
}

export async function fetchFleetSummary(
  windowParam: string = 'all'
): Promise<ApiResponse<FleetSummary>> {
  try {
    const query = windowParam !== 'all' ? `?window=${encodeURIComponent(windowParam)}` : '';
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
  windowParam: string = 'all'
): Promise<ApiResponse<readonly AgenticMetricsLedger[]>> {
  try {
    const query = windowParam !== 'all' ? `?window=${encodeURIComponent(windowParam)}` : '';
    const res = await fetch(`/api/workspaces${query}`);
    if (!res.ok) {
      return { data: Object.values(mockAgenticMetricsLedgers), isMock: true };
    }
    const ledgers = (await res.json()) as AgenticMetricsLedger[];
    return { data: ledgers, isMock: false };
  } catch {
    return { data: Object.values(mockAgenticMetricsLedgers), isMock: true };
  }
}

export async function fetchWorkspaces(
  windowParam: string = 'all',
  sortParam: string = '',
  orderParam: string = ''
): Promise<ApiResponse<TicketComparisonRow[]>> {
  try {
    const params = new URLSearchParams();
    if (windowParam !== 'all') {
      params.set('window', windowParam);
    }
    if (sortParam) {
      params.set('sort', sortParam);
    }
    if (orderParam) {
      params.set('order', orderParam);
    }
    const qs = params.toString() ? `?${params.toString()}` : '';
    const res = await fetch(`/api/workspaces${qs}`);
    if (!res.ok) {
      return { data: [...mockTicketComparisonRows], isMock: true };
    }
    const ledgers = (await res.json()) as AgenticMetricsLedger[];
    const rows = deriveTicketComparisonRows(ledgers);
    return { data: rows, isMock: false };
  } catch {
    return { data: [...mockTicketComparisonRows], isMock: true };
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

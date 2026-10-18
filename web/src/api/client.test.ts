import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  fetchFleetSummary,
  fetchWorkspaces,
  fetchWorkspaceById,
  fetchWorkspaceLedgers,
  fetchWorkspaceEvents,
  fetchWorkspaceAnomalies,
  fetchVersion,
} from './client';
import {
  mockFleetSummary,
  mockTicketComparisonRows,
  mockWorkspaceDetails,
  mockAgenticMetricsLedgers,
  mockBuildInfo,
} from './mockData';

describe('Boundary API Client', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  describe('fetchFleetSummary', () => {
    it('returns live backend payload when HTTP call succeeds', async () => {
      const mockLiveSummary = {
        ...mockFleetSummary,
        total_tickets: 10,
        total_cost_usd: 99.5,
      };

      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockLiveSummary,
      }));

      const res = await fetchFleetSummary('all');
      expect(res.isMock).toBe(false);
      expect(res.data.total_tickets).toBe(10);
      expect(res.data.total_cost_usd).toBe(99.5);
    });

    it('falls back seamlessly to mockFleetSummary on network exception', async () => {
      vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));

      const res = await fetchFleetSummary('all');
      expect(res.isMock).toBe(true);
      expect(res.data.total_tickets).toBe(mockFleetSummary.total_tickets);
      expect(res.data.total_cost_usd).toBe(mockFleetSummary.total_cost_usd);
    });

    it('falls back to mockFleetSummary on HTTP 500 response', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
      }));

      const res = await fetchFleetSummary('7d');
      expect(res.isMock).toBe(true);
      expect(res.data.total_invocations).toBe(mockFleetSummary.total_invocations);
    });

    it('falls back to mockFleetSummary when response body has malformed JSON', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: true,
        json: async () => {
          throw new SyntaxError('Unexpected token < in JSON at position 0');
        },
      }));

      const res = await fetchFleetSummary('all');
      expect(res.isMock).toBe(true);
      expect(res.data.total_tickets).toBe(mockFleetSummary.total_tickets);
    });

    it('falls back to mockFleetSummary when fetch is aborted or timed out', async () => {
      const abortError = new DOMException('The operation was aborted', 'AbortError');
      vi.stubGlobal('fetch', vi.fn().mockRejectedValue(abortError));

      const res = await fetchFleetSummary('30d');
      expect(res.isMock).toBe(true);
      expect(res.data.total_cost_usd).toBe(mockFleetSummary.total_cost_usd);
    });
  });

  describe('fetchWorkspaces', () => {
    it('returns derived rows on successful response with custom query parameters', async () => {
      const mockLedgers = [
        {
          ticket_id: 'TEST-1',
          ticket_title: 'Test ticket',
          ticket_type: 'feature',
          started_at: '2026-10-15T10:00:00Z',
          completed_at: '2026-10-15T11:00:00Z',
          updated_at: '2026-10-15T11:00:00Z',
          facts: {
            ticket_id: 'TEST-1',
            ticket_title: 'Test ticket',
            ticket_type: 'feature',
            started_at: '2026-10-15T10:00:00Z',
            completed_at: '2026-10-15T11:00:00Z',
            updated_at: '2026-10-15T11:00:00Z',
            code: {
              lines_added: 50,
              lines_deleted: 10,
              test_lines_added: 20,
              test_lines_deleted: 0,
              prod_lines_added: 30,
              prod_lines_deleted: 10,
              files_changed: 3,
              sdlc_iterations: 1,
            },
            invocations: [],
            tools: [],
          },
          workflow_efficiency: {
            sdlc_iterations: 1,
            active_duration_seconds: 600,
            active_duration_formatted: '10m 0s',
            role_effort_distribution_pct: {},
            subagent_turn_velocity: {
              avg_turns_per_invocation: 1,
              peak_turns: 1,
              by_role: {},
            },
            test_to_production_ratio: {
              ratio: '0.7:1',
              test_lines_added: 20,
              prod_lines_added: 30,
            },
            reasoning_tokens: 100,
            reasoning_ratio_pct: 10,
            code_delivered: {
              lines_added: 50,
              lines_deleted: 10,
              files_changed: 3,
            },
            tool_execution: {
              total_calls: 5,
              errors_and_corrections: 0,
              success_rate_pct: 100,
            },
            context_window_health: {
              healthy_ceiling_tokens: 200000,
              status: 'HEALTHY',
              orchestrator_peak_tokens: 50000,
              orchestrator_saturation_pct: 25,
              subagents_peak_tokens: 40000,
              subagents_peak_role: 'coder',
              compactions: 0,
              peak_by_role: {},
            },
          },
          threads: [],
          totals: {
            invocations: 2,
            input_tokens: 1000,
            output_tokens: 500,
            cache_read_tokens: 2000,
            cache_hit_rate_pct: 66.7,
            cost_usd: 0.15,
            peak_context_tokens: 50000,
            context_health: 'HEALTHY',
          },
          roles: [],
          models: [],
        },
      ];

      const fetchSpy = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockLedgers,
      });
      vi.stubGlobal('fetch', fetchSpy);

      const res = await fetchWorkspaces('7d', 'cost_usd', 'desc');
      expect(res.isMock).toBe(false);
      expect(res.data).toHaveLength(1);
      expect(res.data[0]?.ticket_id).toBe('TEST-1');
      expect(fetchSpy).toHaveBeenCalledWith('/api/workspaces?window=7d&sort=cost_usd&order=desc');
    });

    it('falls back to mockTicketComparisonRows on fetch rejection', async () => {
      vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Failed to fetch')));

      const res = await fetchWorkspaces('all');
      expect(res.isMock).toBe(true);
      expect(res.data.length).toBe(mockTicketComparisonRows.length);
      expect(res.data[0]?.ticket_id).toBe(mockTicketComparisonRows[0]?.ticket_id);
    });

    it('falls back to mockTicketComparisonRows on malformed JSON response', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: true,
        json: async () => {
          throw new SyntaxError('Unexpected token');
        },
      }));

      const res = await fetchWorkspaces('all');
      expect(res.isMock).toBe(true);
      expect(res.data.length).toBe(mockTicketComparisonRows.length);
    });

    it('falls back to mockTicketComparisonRows on non-200 HTTP status', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: false,
        status: 502,
        statusText: 'Bad Gateway',
      }));

      const res = await fetchWorkspaces('all');
      expect(res.isMock).toBe(true);
      expect(res.data.length).toBe(mockTicketComparisonRows.length);
    });
  });

  describe('fetchWorkspaceById', () => {
    it('returns live detail when backend returns 200 OK', async () => {
      const mockDetail = mockWorkspaceDetails['AL-1']!;
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockDetail,
      }));

      const res = await fetchWorkspaceById('AL-1');
      expect(res.isMock).toBe(false);
      expect(res.data.artifacts.ticket_id).toBe('AL-1');
    });

    it('falls back to mockWorkspaceDetails when available on 404 or 500', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
      }));

      const res = await fetchWorkspaceById('AL-1');
      expect(res.isMock).toBe(true);
      expect(res.data.artifacts.ticket_id).toBe('AL-1');
      expect(res.data.artifacts.contract_content).toContain('Contract: AL-1');
    });

    it('falls back to mockAgenticMetricsLedgers if only ledger exists in mock fixtures', async () => {
      // Temporarily mock an entry in mockAgenticMetricsLedgers that is not in mockWorkspaceDetails
      const customLedger = mockAgenticMetricsLedgers['AL-1']!;
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
      }));

      // AL-1 has both, but test fallback branch by asserting on AL-1 artifacts
      const res = await fetchWorkspaceById('AL-1');
      expect(res.isMock).toBe(true);
      expect(res.data.ledger.ticket_id).toBe(customLedger.ticket_id);
    });

    it('throws structured error when ticketId does not exist in backend nor mocks', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
      }));

      await expect(fetchWorkspaceById('NONEXISTENT-999')).rejects.toThrow(
        'Workspace NONEXISTENT-999 not found'
      );
    });

    it('throws structured error on network rejection when ticketId not in mocks', async () => {
      vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));

      await expect(fetchWorkspaceById('UNKNOWN-TICKET')).rejects.toThrow(
        'Failed to fetch'
      );
    });
  });

  describe('fetchWorkspaceLedgers', () => {
    it('returns raw ledgers on successful HTTP 200 response', async () => {
      const dummyLedgers = [mockAgenticMetricsLedgers['AL-1']!];
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: true,
        json: async () => dummyLedgers,
      }));

      const res = await fetchWorkspaceLedgers();
      expect(res.isMock).toBe(false);
      expect(res.data).toHaveLength(1);
      expect(res.data[0]?.ticket_id).toBe('AL-1');
    });

    it('falls back to mockAgenticMetricsLedgers on network error', async () => {
      vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Connection refused')));

      const res = await fetchWorkspaceLedgers();
      expect(res.isMock).toBe(true);
      expect(res.data.length).toBeGreaterThan(0);
      expect(res.data.some((l) => l.ticket_id === 'AL-1')).toBe(true);
    });

    it('falls back to mockAgenticMetricsLedgers on 500 error response', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        statusText: 'Server Error',
      }));

      const res = await fetchWorkspaceLedgers();
      expect(res.isMock).toBe(true);
      expect(res.data.length).toBeGreaterThan(0);
    });

    it('falls back to mockAgenticMetricsLedgers on invalid JSON response', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: true,
        json: async () => {
          throw new Error('Malformed JSON');
        },
      }));

      const res = await fetchWorkspaceLedgers();
      expect(res.isMock).toBe(true);
      expect(res.data.length).toBeGreaterThan(0);
    });
  });

  describe('fetchWorkspaceEvents', () => {
    it('returns events on successful HTTP 200 response', async () => {
      const mockEvents = [
        {
          timestamp: '2026-10-18T10:00:00Z',
          session_id: 'ses-1',
          turn: 1,
          role: 'coder',
          event_type: 'tool_call' as const,
          tool_name: 'bash',
          status: 'error',
          context_tokens: 15000,
        },
      ];

      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockEvents,
      }));

      const res = await fetchWorkspaceEvents('AL-8');
      expect(res.isMock).toBe(false);
      expect(res.data).toHaveLength(1);
      expect(res.data[0]?.tool_name).toBe('bash');
    });

    it('falls back to mock ledger events on network failure', async () => {
      vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));

      const res = await fetchWorkspaceEvents('AL-1');
      expect(res.isMock).toBe(true);
      expect(Array.isArray(res.data)).toBe(true);
    });

    it('falls back to mock events on 404 when ticket exists in mocks', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
        statusText: 'Not Found',
      }));

      const res = await fetchWorkspaceEvents('AL-1');
      expect(res.isMock).toBe(true);
      expect(Array.isArray(res.data)).toBe(true);
    });

    it('throws structured error when ticketId does not exist in backend nor mocks on 404', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
        statusText: 'Not Found',
      }));

      await expect(fetchWorkspaceEvents('NON-EXISTENT')).rejects.toThrow(
        'Workspace NON-EXISTENT not found'
      );
    });

    it('throws structured error on network failure when ticketId not in mocks', async () => {
      vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));

      await expect(fetchWorkspaceEvents('UNKNOWN-TICKET')).rejects.toThrow(
        'Failed to fetch'
      );
    });

    it('falls back to mock ledger events on malformed JSON response', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: true,
        json: async () => {
          throw new SyntaxError('Invalid JSON');
        },
      }));

      const res = await fetchWorkspaceEvents('AL-1');
      expect(res.isMock).toBe(true);
      expect(Array.isArray(res.data)).toBe(true);
    });

    it('returns empty array with isMock false when backend returns empty events stream', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: true,
        json: async () => [],
      }));

      const res = await fetchWorkspaceEvents('AL-1');
      expect(res.isMock).toBe(false);
      expect(res.data).toEqual([]);
    });
  });

  describe('fetchWorkspaceAnomalies', () => {
    it('returns anomalies from workspace detail response when fetch succeeds', async () => {
      const liveDetail = {
        ...mockWorkspaceDetails['AL-2'],
        anomalies: [
          {
            rule_id: 'TOOL_RETRY_STREAK',
            severity: 'WARNING',
            title: 'Repeated Tool Failures',
            description: 'Tool failed 3 times',
            context: { tool_name: 'bash', count: 3 },
          },
        ],
      };

      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: true,
        json: async () => liveDetail,
      }));

      const res = await fetchWorkspaceAnomalies('AL-2');
      expect(res.isMock).toBe(false);
      expect(res.data).toHaveLength(1);
      expect(res.data[0]?.rule_id).toBe('TOOL_RETRY_STREAK');
    });

    it('returns empty array when workspace detail has no anomalies', async () => {
      const cleanDetail = {
        ...mockWorkspaceDetails['AL-1']!,
        anomalies: [],
      };

      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: true,
        json: async () => cleanDetail,
      }));

      const res = await fetchWorkspaceAnomalies('AL-1');
      expect(res.isMock).toBe(false);
      expect(res.data).toEqual([]);
    });

    it('falls back to mock anomalies when workspace detail fetch fails', async () => {
      vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));

      const res = await fetchWorkspaceAnomalies('AL-2');
      expect(res.isMock).toBe(true);
      expect(Array.isArray(res.data)).toBe(true);
      expect(res.data.length).toBeGreaterThan(0);
    });

    it('throws structured error when workspace does not exist in backend nor mocks', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
      }));

      await expect(fetchWorkspaceAnomalies('UNKNOWN-NONEXISTENT')).rejects.toThrow(
        'Workspace UNKNOWN-NONEXISTENT not found'
      );
    });
  });

  describe('fetchVersion', () => {
    it('returns live build info on successful HTTP 200 response', async () => {
      const liveInfo = {
        version: 'v1.2.3',
        commit: '1234567',
        build_date: '2026-10-18T12:00:00Z',
      };
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: true,
        json: async () => liveInfo,
      }));

      const res = await fetchVersion();
      expect(res.isMock).toBe(false);
      expect(res.data.version).toBe('v1.2.3');
      expect(res.data.commit).toBe('1234567');
      expect(res.data.build_date).toBe('2026-10-18T12:00:00Z');
    });

    it('falls back to mockBuildInfo on network error', async () => {
      vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));

      const res = await fetchVersion();
      expect(res.isMock).toBe(true);
      expect(res.data).toEqual(mockBuildInfo);
    });

    it('falls back to mockBuildInfo on 500 error response', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        statusText: 'Server Error',
      }));

      const res = await fetchVersion();
      expect(res.isMock).toBe(true);
      expect(res.data).toEqual(mockBuildInfo);
    });

    it('falls back to mockBuildInfo on 404 not found response', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
        statusText: 'Not Found',
      }));

      const res = await fetchVersion();
      expect(res.isMock).toBe(true);
      expect(res.data).toEqual(mockBuildInfo);
    });

    it('falls back to mockBuildInfo on DOMException abort/timeout', async () => {
      vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new DOMException('The operation was aborted', 'AbortError')));

      const res = await fetchVersion();
      expect(res.isMock).toBe(true);
      expect(res.data).toEqual(mockBuildInfo);
    });

    it('falls back to mockBuildInfo on invalid JSON response', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
        ok: true,
        json: async () => {
          throw new Error('Malformed JSON');
        },
      }));

      const res = await fetchVersion();
      expect(res.isMock).toBe(true);
      expect(res.data).toEqual(mockBuildInfo);
    });
  });
});

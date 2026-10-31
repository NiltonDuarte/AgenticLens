import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { fetchFleetSummary, fetchWorkspaces, fetchWorkspaceById, fetchWorkspaceLedgers, fetchWorkspaceEvents, fetchWorkspaceAnomalies, fetchVersion, fetchProjectConfig, fetchComparisonRows, fetchProgressionLedgers, fetchProgressionMetrics, } from './client';
import { mockFleetSummary, mockTicketComparisonRows, mockWorkspaceDetails, mockAgenticMetricsLedgers, } from '../test/fixtures/mockData';
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
            expect(res.data.total_tickets).toBe(10);
            expect(res.data.total_cost_usd).toBe(99.5);
        });
        it('rejects on network exception', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));
            await expect(fetchFleetSummary('all')).rejects.toThrow('Network error');
        });
        it('rejects on HTTP 500 response', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 500,
                statusText: 'Internal Server Error',
            }));
            await expect(fetchFleetSummary('7d')).rejects.toThrow(/HTTP error 500/);
        });
        it('rejects when response body has malformed JSON', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: true,
                json: async () => {
                    throw new SyntaxError('Unexpected token < in JSON at position 0');
                },
            }));
            await expect(fetchFleetSummary('all')).rejects.toThrow(SyntaxError);
        });
        it('rejects when fetch is aborted or timed out', async () => {
            const abortError = new DOMException('The operation was aborted', 'AbortError');
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue(abortError));
            await expect(fetchFleetSummary('30d')).rejects.toThrow('The operation was aborted');
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
                        sessions: []
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
                            root_peak_tokens: 50000,
                            root_saturation_pct: 25,
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
            expect(res.data).toHaveLength(1);
            expect(res.data[0]?.ticket_id).toBe('TEST-1');
            expect(fetchSpy).toHaveBeenCalledWith('/api/workspaces?window=7d&sort=cost_usd&order=desc');
        });
        it('rejects on fetch rejection', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Failed to fetch')));
            await expect(fetchWorkspaces('all')).rejects.toThrow('Failed to fetch');
        });
        it('rejects on malformed JSON response', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: true,
                json: async () => {
                    throw new SyntaxError('Unexpected token');
                },
            }));
            await expect(fetchWorkspaces('all')).rejects.toThrow(SyntaxError);
        });
        it('rejects on non-200 HTTP status', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 502,
                statusText: 'Bad Gateway',
            }));
            await expect(fetchWorkspaces('all')).rejects.toThrow(/HTTP error 502/);
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
            expect(res.data.artifacts.ticket_id).toBe('AL-1');
        });
        it('throws structured error on 404 response', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 404,
                statusText: 'Not Found',
            }));
            await expect(fetchWorkspaceById('AL-1')).rejects.toThrow('Workspace AL-1 not found');
        });
        it('throws HTTP error on 500 response', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 500,
                statusText: 'Internal Server Error',
            }));
            await expect(fetchWorkspaceById('AL-1')).rejects.toThrow(/HTTP error 500/);
        });
        it('throws error on network rejection', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));
            await expect(fetchWorkspaceById('UNKNOWN-TICKET')).rejects.toThrow('Failed to fetch');
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
            expect(res.data).toHaveLength(1);
            expect(res.data[0]?.ticket_id).toBe('AL-1');
        });
        it('exposes structured skipped-workspace diagnostics from the response header', async () => {
            const diagnostics = [{ ticket_id: 'AL-OLD', message: 'Schema version 3 requires migration and source-backed backfill.' }];
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: true,
                headers: { get: (name: string) => name === 'X-AgenticLens-Workspace-Diagnostics' ? JSON.stringify(diagnostics) : null },
                json: async () => [mockAgenticMetricsLedgers['AL-1']!],
            }));
            const res = await fetchWorkspaceLedgers();
            expect(res.data).toHaveLength(1);
            expect(res.diagnostics).toEqual(diagnostics);
        });
        it('rejects on network error', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Connection refused')));
            await expect(fetchWorkspaceLedgers()).rejects.toThrow('Connection refused');
        });
        it('rejects on 500 error response', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 500,
                statusText: 'Server Error',
            }));
            await expect(fetchWorkspaceLedgers()).rejects.toThrow(/HTTP error 500/);
        });
        it('rejects on invalid JSON response', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: true,
                json: async () => {
                    throw new SyntaxError('Malformed JSON');
                },
            }));
            await expect(fetchWorkspaceLedgers()).rejects.toThrow(SyntaxError);
        });
    });
    describe('fetchWorkspaceEvents', () => {
        const recordedEvent = {
            timestamp: '2026-10-29T10:00:00Z', session_id: 'recorded', turn: 1,
            role: 'tester', event_type: 'tool_call', tool_name: 'bash', status: 'error', context_tokens: 0,
        };
        const recordedResponse = {
            ticket_id: 'FIX-12', is_hydrated: false, data_source: 'committed_ledger', events: [recordedEvent],
        };

        it('accepts optional missing IDs and payloads plus null anomalies in a committed envelope', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true,
                json: async () => ({ ...recordedResponse, anomalies: null }),
            }));
            const result = await fetchWorkspaceEvents('FIX-12');
            expect(result.data.events).toEqual([recordedEvent]);
            expect(result.data.is_hydrated).toBe(false);
            expect(result.data.anomalies).toBeUndefined();
        });

        it('accepts a Codex envelope with valid anomaly evidence and optional event details', async () => {
            const response = { ...recordedResponse, is_hydrated: true, data_source: 'local_codex',
                events: [{ ...recordedEvent, event_id: 'call', duration_ms: null, tool_output: 'output' }],
                anomalies: [{ rule_id: 'TOOL_RETRY_STREAK', severity: 'WARNING', context: { session_id: 'recorded', turn: 1 },
                    evidence: [{ session_id: 'recorded', event_id: 'call' }] }],
            };
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => response }));
            expect((await fetchWorkspaceEvents('FIX-12')).data).toEqual(response);
        });

        it.each([true, false])('preserves explicit includes_baseline=%s provenance at the HTTP boundary', async (includesBaseline) => {
            const response = { ...recordedResponse, includes_baseline: includesBaseline,
                events: [{ ...recordedEvent }, { ...recordedEvent }],
            };
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => response }));
            expect((await fetchWorkspaceEvents('FIX-12')).data).toEqual(response);
        });

        it.each([
            { ...recordedResponse, ticket_id: 'OTHER-1' },
            { ...recordedResponse, is_hydrated: 'false' },
            { ...recordedResponse, data_source: 'unknown' },
            { ...recordedResponse, includes_baseline: 'true' },
            { ...recordedResponse, includes_baseline: 1 },
            { ...recordedResponse, includes_baseline: null },
            { ...recordedResponse, events: null },
            { ...recordedResponse, events: [{ ...recordedEvent, event_type: 'unknown' }] },
            { ...recordedResponse, events: [{ ...recordedEvent, turn: '1' }] },
            { ...recordedResponse, events: [{ ...recordedEvent, duration_ms: '12' }] },
            { ...recordedResponse, events: [{ ...recordedEvent, tool_output: {} }] },
            { ...recordedResponse, events: [{ ...recordedEvent, context_tokens: Number.NaN }] },
            { ...recordedResponse, anomalies: [{}] },
            { ...recordedResponse, anomalies: [{ rule_id: 'TOOL_RETRY_STREAK', severity: 'WARNING', context: { turn: '1' } }] },
            { ...recordedResponse, anomalies: [{ rule_id: 'TOOL_RETRY_STREAK', severity: 'WARNING', context: {}, evidence: [null] }] },
            [{ ...recordedEvent, session_id: null }],
        ])('rejects malformed enrichment before it reaches the trace: %j', async (payload) => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => payload }));
            await expect(fetchWorkspaceEvents('FIX-12')).rejects.toThrow('Invalid workspace events response for FIX-12');
        });

        it('returns hydrated response on successful HTTP 200 response', async () => {
            const mockResponse = {
                ticket_id: 'AL-8',
                is_hydrated: true,
                data_source: 'local_sqlite' as const,
                events: [
                            {
                                timestamp: '2026-10-18T10:00:00Z',
                                session_id: 'ses-1',
                                turn: 1,
                                role: 'coder',
                                event_type: 'tool_call' as const,
                                tool_name: 'bash',
                                status: 'error',
                                context_tokens: 15000,
                                tool_input: 'echo test',
                                tool_output: 'error',
                            }
                        ]
            };
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: true,
                json: async () => mockResponse,
            }));
            const res = await fetchWorkspaceEvents('AL-8');
            expect(res.data.is_hydrated).toBe(true);
            expect(res.data.data_source).toBe('local_sqlite');
            expect(res.data.events).toHaveLength(1);
            expect(res.data.events[0]?.tool_name).toBe('bash');
            expect(res.data.events[0]?.tool_input).toBe('echo test');
        });
        it('normalizes legacy raw array response into fallback WorkspaceEventsResponse', async () => {
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
            expect(res.data.is_hydrated).toBe(false);
            expect(res.data.data_source).toBe('committed_ledger');
            expect(res.data.events).toHaveLength(1);
            expect(res.data.events[0]?.tool_name).toBe('bash');
        });
        it('rejects on network failure', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));
            await expect(fetchWorkspaceEvents('AL-1')).rejects.toThrow('Network error');
        });
        it('throws structured error on 404', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 404,
                statusText: 'Not Found',
            }));
            await expect(fetchWorkspaceEvents('AL-1')).rejects.toThrow('Workspace AL-1 not found');
        });
        it('rejects on HTTP 500 server error', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 500,
                statusText: 'Internal Server Error',
            }));
            await expect(fetchWorkspaceEvents('AL-1')).rejects.toThrow(/HTTP error 500/);
        });
        it('rejects on malformed JSON response', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: true,
                json: async () => {
                    throw new SyntaxError('Invalid JSON');
                },
            }));
            await expect(fetchWorkspaceEvents('AL-1')).rejects.toThrow(SyntaxError);
        });
        it('returns empty array when backend returns empty events stream', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: true,
                json: async () => [],
            }));
            const res = await fetchWorkspaceEvents('AL-1');
            expect(res.data.events).toEqual([]);
            expect(res.data.is_hydrated).toBe(false);
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
            expect(res.data).toEqual([]);
        });
        it('rejects when workspace detail fetch fails', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));
            await expect(fetchWorkspaceAnomalies('AL-2')).rejects.toThrow('Network error');
        });
        it('throws structured error when workspace does not exist on 404', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 404,
            }));
            await expect(fetchWorkspaceAnomalies('UNKNOWN-NONEXISTENT')).rejects.toThrow('Workspace UNKNOWN-NONEXISTENT not found');
        });
        it('rejects on HTTP 500 server error', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 500,
                statusText: 'Internal Server Error',
            }));
            await expect(fetchWorkspaceAnomalies('AL-2')).rejects.toThrow(/HTTP error 500/);
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
            expect(res.data.version).toBe('v1.2.3');
            expect(res.data.commit).toBe('1234567');
            expect(res.data.build_date).toBe('2026-10-18T12:00:00Z');
        });
        it('rejects on network error', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));
            await expect(fetchVersion()).rejects.toThrow('Network error');
        });
        it('rejects on 500 error response', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 500,
                statusText: 'Server Error',
            }));
            await expect(fetchVersion()).rejects.toThrow(/HTTP error 500/);
        });
        it('rejects on 404 not found response', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 404,
                statusText: 'Not Found',
            }));
            await expect(fetchVersion()).rejects.toThrow(/HTTP error 404/);
        });
        it('rejects on DOMException abort/timeout', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new DOMException('The operation was aborted', 'AbortError')));
            await expect(fetchVersion()).rejects.toThrow('The operation was aborted');
        });
        it('rejects on invalid JSON response', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: true,
                json: async () => {
                    throw new SyntaxError('Malformed JSON');
                },
            }));
            await expect(fetchVersion()).rejects.toThrow(SyntaxError);
        });
    });
    describe('fetchProjectConfig', () => {
        it('returns live project config on successful HTTP 200 response', async () => {
            const liveConfig = {
                project_name: 'CustomObserver',
                prefixes: [
                    { prefix: 'FEAT', description: 'Feature tickets', ticket_type: 'feat' },
                ],
                workspaces_dir: 'custom_workspaces',
                role_order: ['planner', 'coder'],
                version: 'v1.0.0',
            };
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: true,
                json: async () => liveConfig,
            }));
            const res = await fetchProjectConfig();
            expect(res.data.project_name).toBe('CustomObserver');
            expect(res.data.prefixes).toHaveLength(1);
            expect(res.data.prefixes[0]?.prefix).toBe('FEAT');
            expect(res.data.workspaces_dir).toBe('custom_workspaces');
            expect(res.data.role_order).toEqual(['planner', 'coder']);
            expect(res.data.version).toBe('v1.0.0');
        });
        it('rejects on network error', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));
            await expect(fetchProjectConfig()).rejects.toThrow('Network error');
        });
        it('rejects on 500 error response', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 500,
                statusText: 'Server Error',
            }));
            await expect(fetchProjectConfig()).rejects.toThrow(/HTTP error 500/);
        });
        it('rejects on 404 not found response', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 404,
                statusText: 'Not Found',
            }));
            await expect(fetchProjectConfig()).rejects.toThrow(/HTTP error 404/);
        });
        it('rejects on DOMException abort/timeout', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new DOMException('The operation was aborted', 'AbortError')));
            await expect(fetchProjectConfig()).rejects.toThrow('The operation was aborted');
        });
        it('rejects on invalid JSON response', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: true,
                json: async () => {
                    throw new SyntaxError('Malformed JSON');
                },
            }));
            await expect(fetchProjectConfig()).rejects.toThrow(SyntaxError);
        });
    });
    describe('fetchComparisonRows and fetchProgressionLedgers', () => {
        it('appends ticket_type and prefix multi-parameters to /api/comparison', async () => {
            const fetchSpy = vi.fn().mockResolvedValue({
                ok: true,
                json: async () => mockTicketComparisonRows,
            });
            vi.stubGlobal('fetch', fetchSpy);
            const res = await fetchComparisonRows('30d', undefined, undefined, ['feature', 'bug'], ['AL', 'PO']);
            expect(res.data).toHaveLength(mockTicketComparisonRows.length);
            expect(fetchSpy).toHaveBeenCalled();
            const calledUrl = fetchSpy.mock.calls[0]![0] as string;
            expect(calledUrl).toContain('/api/comparison?');
            expect(calledUrl).toContain('window=30d');
            expect(calledUrl).toContain('ticket_type=feature');
            expect(calledUrl).toContain('ticket_type=bug');
            expect(calledUrl).toContain('prefix=AL');
            expect(calledUrl).toContain('prefix=PO');
        });
        it('rejects on /api/comparison failure', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));
            await expect(fetchComparisonRows('all', undefined, undefined, ['feature'], ['AL'])).rejects.toThrow('Network error');
        });
        it('rejects on /api/comparison HTTP 500 status', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 500,
                statusText: 'Internal Server Error',
            }));
            await expect(fetchComparisonRows('all', undefined, undefined, ['feature'], ['AL'])).rejects.toThrow(/HTTP error 500/);
        });
        it('rejects on /api/comparison HTTP 404 status', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 404,
                statusText: 'Not Found',
            }));
            await expect(fetchComparisonRows('all', undefined, undefined, ['feature'], ['AL'])).rejects.toThrow(/HTTP error 404/);
        });
        it('appends ticket_type and prefix multi-parameters to /api/progression', async () => {
            const fetchSpy = vi.fn().mockResolvedValue({
                ok: true,
                json: async () => Object.values(mockAgenticMetricsLedgers),
            });
            vi.stubGlobal('fetch', fetchSpy);
            const res = await fetchProgressionLedgers('all', undefined, undefined, ['chore'], ['GEN']);
            expect(res.data).toHaveLength(Object.values(mockAgenticMetricsLedgers).length);
            expect(fetchSpy).toHaveBeenCalled();
            const calledUrl = fetchSpy.mock.calls[0]![0] as string;
            expect(calledUrl).toContain('/api/progression?');
            expect(calledUrl).toContain('ticket_type=chore');
            expect(calledUrl).toContain('prefix=GEN');
        });
        it('rejects on /api/progression failure', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 500,
                statusText: 'Internal Server Error',
            }));
            await expect(fetchProgressionLedgers('all', undefined, undefined, ['feature'], ['AL'])).rejects.toThrow(/HTTP error 500/);
        });
        it('rejects on /api/progression network error', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Connection reset')));
            await expect(fetchProgressionLedgers('all', undefined, undefined, ['feature'], ['AL'])).rejects.toThrow('Connection reset');
        });
        it('rejects on /api/progression HTTP 404 status', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 404,
                statusText: 'Not Found',
            }));
            await expect(fetchProgressionLedgers('all', undefined, undefined, ['feature'], ['AL'])).rejects.toThrow(/HTTP error 404/);
        });
        it('appends multi-parameters to fetchFleetSummary and omits all', async () => {
            const fetchSpy = vi.fn().mockResolvedValue({
                ok: true,
                json: async () => mockFleetSummary,
            });
            vi.stubGlobal('fetch', fetchSpy);
            // With concrete types
            await fetchFleetSummary('all', undefined, undefined, ['feature', 'bug'], ['AL']);
            const calledUrl = fetchSpy.mock.calls[0]![0] as string;
            expect(calledUrl).toContain('/api/fleet?');
            expect(calledUrl).toContain('ticket_type=feature');
            expect(calledUrl).toContain('ticket_type=bug');
            expect(calledUrl).toContain('prefix=AL');
            // With 'all' filter, params are omitted
            fetchSpy.mockClear();
            await fetchFleetSummary('all', undefined, undefined, ['all'], ['all']);
            const calledUrlAll = fetchSpy.mock.calls[0]![0] as string;
            expect(calledUrlAll).toBe('/api/fleet');
        });
        it('appends multi-parameters to fetchWorkspaces and omits all', async () => {
            const fetchSpy = vi.fn().mockResolvedValue({
                ok: true,
                json: async () => Object.values(mockAgenticMetricsLedgers),
            });
            vi.stubGlobal('fetch', fetchSpy);
            await fetchWorkspaces('30d', 'cost_usd', 'desc', undefined, undefined, ['chore'], ['GEN']);
            const calledUrl = fetchSpy.mock.calls[0]![0] as string;
            expect(calledUrl).toContain('ticket_type=chore');
            expect(calledUrl).toContain('prefix=GEN');
            expect(calledUrl).toContain('sort=cost_usd');
            expect(calledUrl).toContain('order=desc');
        });
    });
    describe('fetchProgressionMetrics', () => {
        it('returns live backend payload and passes query parameters accurately', async () => {
            const mockBuckets = [
                {
                    bucket_key: '2026-10-19T14:00:00Z',
                    bucket_date: '2026-10-19T14:00:00Z',
                    invocations: 5,
                    cost_usd: 0.25,
                    input_tokens: 1000,
                    output_tokens: 200,
                    cache_read_tokens: 500,
                    reasoning_tokens: 0,
                    cache_hit_rate_pct: 33.3,
                    lines_added: 50,
                    lines_deleted: 10,
                    files_changed: 2,
                    net_lines_added: 40,
                    peak_context_tokens: 5000,
                    active_duration_seconds: 60,
                    cumulative_cost_usd: 0.25,
                    cumulative_input_tokens: 1000,
                    cumulative_output_tokens: 200,
                    cumulative_lines_added: 50,
                    cumulative_lines_deleted: 10,
                    cumulative_invocations: 5,
                },
            ];
            const fetchSpy = vi.fn().mockResolvedValue({
                ok: true,
                json: async () => mockBuckets,
            });
            vi.stubGlobal('fetch', fetchSpy);
            const res = await fetchProgressionMetrics({
                window: '1d',
                granularity: 'hourly',
                startDate: '2026-10-19T00:00',
                endDate: '2026-10-19T23:59',
            });
            expect(res.data).toHaveLength(1);
            expect(res.data[0]?.bucket_key).toBe('2026-10-19T14:00:00Z');
            expect(fetchSpy).toHaveBeenCalledWith('/api/progression?window=1d&granularity=hourly&startDate=2026-10-19T00%3A00&endDate=2026-10-19T23%3A59');
        });
        it('rejects on network failure', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Connection refused')));
            await expect(fetchProgressionMetrics({
                window: 'all',
                granularity: 'daily',
            })).rejects.toThrow('Connection refused');
        });
        it('rejects on non-200 HTTP status', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 500,
                statusText: 'Internal Server Error',
            }));
            await expect(fetchProgressionMetrics({
                window: '7d',
                granularity: 'weekly',
            })).rejects.toThrow(/HTTP error 500/);
        });
        it('rejects on HTTP 404 status', async () => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
                ok: false,
                status: 404,
                statusText: 'Not Found',
            }));
            await expect(fetchProgressionMetrics({
                window: 'all',
                granularity: 'daily',
            })).rejects.toThrow(/HTTP error 404/);
        });
    });
    describe('adversarial error shapes handled by rethrowError', () => {
        it('handles non-Error objects with message property', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue({ message: 'Custom object error' }));
            await expect(fetchFleetSummary('all')).rejects.toThrow('Custom object error');
        });
        it('handles primitive string thrown by fetch', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue('raw string thrown'));
            await expect(fetchWorkspaceLedgers('all')).rejects.toThrow('Failed to fetch workspace ledgers');
        });
        it('handles null thrown by fetch', async () => {
            vi.stubGlobal('fetch', vi.fn().mockRejectedValue(null));
            await expect(fetchWorkspaces('all')).rejects.toThrow('Failed to fetch workspaces');
        });
    });
});

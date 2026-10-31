import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { fetchFleetSummary, fetchWorkspaces, fetchWorkspaceById, fetchWorkspaceLedgers, fetchWorkspaceEvents, fetchVersion, fetchProjectConfig, } from './client';
import { mockFleetSummary, mockWorkspaceDetails, mockAgenticMetricsLedgers, } from '../test/fixtures/mockData';
describe('Boundary API Client', () => {
    beforeEach(() => {
        vi.restoreAllMocks();
    });
    afterEach(() => {
        vi.unstubAllGlobals();
    });
    describe('successful HTTP payload validation', () => {
        const ledger = mockAgenticMetricsLedgers['AL-1']!;
        const detail = mockWorkspaceDetails['AL-1']!;
        const cases: readonly [string, () => Promise<unknown>, unknown][] = [
            ['fleet null', () => fetchFleetSummary(), null],
            ['fleet nonfinite totals', () => fetchFleetSummary(), { ...mockFleetSummary, total_cost_usd: Infinity }],
            ['fleet nested costs', () => fetchFleetSummary(), { ...mockFleetSummary, role_costs: { coder: '5' } }],
            ['fleet malformed tools', () => fetchFleetSummary(), { ...mockFleetSummary, anomaly_summary: { ...mockFleetSummary.anomaly_summary, brittle_tools: [{}] } }],
            ['ledger missing wrapper', () => fetchWorkspaceLedgers(), ledger],
            ['ledger invalid health', () => fetchWorkspaceLedgers(), [{ ...ledger, totals: { ...ledger.totals, context_health: 'UNKNOWN' } }]],
            ['ledger malformed session', () => fetchWorkspaceLedgers(), [{ ...ledger, facts: { ...ledger.facts, sessions: [{ id: 's', tools: [], invocations: [{ role: 'coder' }] }] } }]],
            ['ledger nonfinite token', () => fetchWorkspaceLedgers(), [{ ...ledger, totals: { ...ledger.totals, input_tokens: NaN } }]],
            ['ledger malformed task link', () => fetchWorkspaceLedgers(), [{ ...ledger, facts: { ...ledger.facts, task_links: [{ parent_turn: '1' }] } }]],
            ['comparison malformed workflow', () => fetchWorkspaces(), [{ ...ledger, workflow_efficiency: { ...ledger.workflow_efficiency, context_window_health: null } }]],
            ['detail missing ledger', () => fetchWorkspaceById('AL-1'), { ...detail, ledger: null }],
            ['detail malformed artifact', () => fetchWorkspaceById('AL-1'), { ...detail, artifacts: { ticket_id: 'AL-1', items: [{ path: 'SUMMARY.md', content: 42 }] } }],
            ['detail malformed artifact description', () => fetchWorkspaceById('AL-1'), { ...detail, artifacts: { ticket_id: 'AL-1', items: [{ path: 'SUMMARY.md', title: 'Summary', description: null, content: null }] } }],
            ['detail malformed anomaly', () => fetchWorkspaceById('AL-1'), { ...detail, anomalies: [{ rule_id: 'UNKNOWN', severity: 'WARNING', context: {} }] }],
            ['version malformed commit', () => fetchVersion(), { version: 'v1', commit: 123, build_date: '' }],
            ['config malformed prefix', () => fetchProjectConfig(), { project_name: 'P', workspaces_dir: 'tickets', version: 'v1', role_order: [], prefixes: [null] }],
            ['config malformed role order', () => fetchProjectConfig(), { project_name: 'P', workspaces_dir: 'tickets', version: 'v1', role_order: [1], prefixes: [] }],
            ['events malformed envelope', () => fetchWorkspaceEvents('AL-1'), { ticket_id: 'AL-1', events: [] }],
            ['events null', () => fetchWorkspaceEvents('AL-1'), null],
        ];
        it.each(cases)('rejects %s before returning domain data', async (_name, request, payload) => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => payload }));
            await expect(request()).rejects.toThrow(/Invalid .* response/);
        });
        it('accepts additive fields and Go nil collections without mutating the response', async () => {
            const payload = { ...ledger, extra_future_field: true, roles: null, models: null, threads: null,
                anomalies: null, facts: { ...ledger.facts, sessions: [{ id: 's', invocations: null, tools: null, anomalies: null }], skills: null, task_links: null, anomalies: null },
                workflow_efficiency: { ...ledger.workflow_efficiency, role_effort_distribution_pct: null,
                    subagent_turn_velocity: { ...ledger.workflow_efficiency.subagent_turn_velocity, by_role: null },
                    context_window_health: { ...ledger.workflow_efficiency.context_window_health, peak_by_role: null } },
            };
            const before = JSON.stringify(payload);
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => [payload] }));
            const result = (await fetchWorkspaceLedgers()).data[0]!;
            expect(result.facts.sessions[0]).toEqual({ id: 's', invocations: [], tools: [], anomalies: [] });
            expect(result.facts.skills).toEqual([]);
            expect(result.threads).toEqual([]);
            expect(result.workflow_efficiency.role_effort_distribution_pct).toEqual({});
            expect(JSON.stringify(payload)).toBe(before);
        });
        it.each(['not json', '{}', '[null]', '[{"ticket_id":4,"message":"bad"}]'])('keeps valid data and reports an unreadable diagnostic header: %s', async (header) => {
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, headers: new Headers({ 'X-AgenticLens-Workspace-Diagnostics': header }), json: async () => [ledger] }));
            const response = await fetchWorkspaceLedgers();
            expect(response.data[0]?.ticket_id).toBe('AL-1');
            expect(response.diagnostics).toEqual([{ scope: 'response', message: expect.stringContaining('Valid workspace data is still shown') }]);
        });
        it('retains valid sibling diagnostics alongside a malformed-entry warning', async () => {
            const warning = { ticket_id: 'OLD-1', message: 'Migration required' };
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, headers: new Headers({ 'X-AgenticLens-Workspace-Diagnostics': JSON.stringify([warning, null]) }), json: async () => [ledger] }));
            expect((await fetchWorkspaceLedgers()).diagnostics).toEqual([warning, { scope: 'response', message: expect.any(String) }]);
        });
        it.each([
            [{ ticket_id: 'AL-1' }, { ticket_id: 'AL-1' }],
            [{ ticket_id: 'AL-1', items: [{ path: 'SUMMARY.md', title: 'Summary', content: null }] }, { ticket_id: 'AL-1', items: [{ path: 'SUMMARY.md', title: 'Summary', description: '', content: null }] }],
        ])('accepts omitted optional ledger facts and artifact fields: %j', async (artifacts, expectedArtifacts) => {
            const { skills, task_links, anomalies, migration, ...facts } = ledger.facts;
            void skills; void task_links; void anomalies; void migration;
            const payload = { ...detail, ledger: { ...ledger, facts }, artifacts, anomalies: null };
            vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => payload }));
            expect((await fetchWorkspaceById('AL-1')).data.artifacts).toEqual(expectedArtifacts);
        });
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

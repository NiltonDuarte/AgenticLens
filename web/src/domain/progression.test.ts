// @vitest-environment node
import { describe, it, expect } from 'vitest';
import { parseIsoDateToUtc, getBucketKey, aggregateMetricsOverTime, } from './progression';
import type { AgenticMetricsLedger, InvocationFact } from './types';
import { mockAgenticMetricsLedgers } from '../test/fixtures/mockData';
describe('Domain Progression Aggregator', () => {
    describe('parseIsoDateToUtc', () => {
        it('parses valid ISO date and full timestamps into UTC date', () => {
            const d1 = parseIsoDateToUtc('2026-10-15T14:30:00Z');
            expect(d1).not.toBeNull();
            expect(d1?.getUTCFullYear()).toBe(2026);
            expect(d1?.getUTCMonth()).toBe(9); // 0-indexed October
            expect(d1?.getUTCDate()).toBe(15);
            const d2 = parseIsoDateToUtc('2026-01-01');
            expect(d2).not.toBeNull();
            expect(d2?.getUTCFullYear()).toBe(2026);
            expect(d2?.getUTCMonth()).toBe(0);
            expect(d2?.getUTCDate()).toBe(1);
            const d3 = parseIsoDateToUtc('2026-10-15T14:30:00.123+02:00');
            expect(d3).not.toBeNull();
            expect(d3?.getUTCFullYear()).toBe(2026);
            expect(d3?.getUTCMonth()).toBe(9);
            expect(d3?.getUTCDate()).toBe(15);
        });
        it('handles leap years correctly and rejects non-leap year Feb 29', () => {
            // 2024 is a leap year
            const leap = parseIsoDateToUtc('2024-02-29');
            expect(leap).not.toBeNull();
            expect(leap?.getUTCFullYear()).toBe(2024);
            expect(leap?.getUTCMonth()).toBe(1);
            expect(leap?.getUTCDate()).toBe(29);
            // 2026 is not a leap year
            expect(parseIsoDateToUtc('2026-02-29')).toBeNull();
            // 2025 is not a leap year
            expect(parseIsoDateToUtc('2025-02-29')).toBeNull();
        });
        it('rejects null, undefined, empty, or short strings', () => {
            expect(parseIsoDateToUtc(null)).toBeNull();
            expect(parseIsoDateToUtc(undefined)).toBeNull();
            expect(parseIsoDateToUtc('')).toBeNull();
            expect(parseIsoDateToUtc('2026-10')).toBeNull();
            expect(parseIsoDateToUtc('   ')).toBeNull();
        });
        it('rejects invalid or rollover dates defensively', () => {
            expect(parseIsoDateToUtc('invalid-date-string')).toBeNull();
            expect(parseIsoDateToUtc('2026-13-01')).toBeNull();
            expect(parseIsoDateToUtc('2026-00-15')).toBeNull();
            expect(parseIsoDateToUtc('2026-02-30')).toBeNull();
            expect(parseIsoDateToUtc('2026-04-31')).toBeNull(); // April has 30 days
            expect(parseIsoDateToUtc('2026-06-31')).toBeNull(); // June has 30 days
            expect(parseIsoDateToUtc('2026-09-31')).toBeNull(); // September has 30 days
            expect(parseIsoDateToUtc('2026-11-31')).toBeNull(); // November has 30 days
            expect(parseIsoDateToUtc('2026/10/15')).toBeNull();
            expect(parseIsoDateToUtc('15-10-2026')).toBeNull();
        });
    });
    describe('getBucketKey', () => {
        it('returns YYYY-MM-DD for daily granularity', () => {
            const d = new Date(Date.UTC(2026, 9, 15)); // 2026-10-15
            expect(getBucketKey(d, false)).toBe('2026-10-15');
            const d2 = new Date(Date.UTC(2026, 0, 1));
            expect(getBucketKey(d2, false)).toBe('2026-01-01');
        });
        it('returns ISO Monday for all 7 days of a standard week', () => {
            // 2026-10-12 (Mon) through 2026-10-18 (Sun) all bucket to Monday 2026-10-12
            const days = [
                { date: '2026-10-12', d: new Date(Date.UTC(2026, 9, 12)) }, // Mon
                { date: '2026-10-13', d: new Date(Date.UTC(2026, 9, 13)) }, // Tue
                { date: '2026-10-14', d: new Date(Date.UTC(2026, 9, 14)) }, // Wed
                { date: '2026-10-15', d: new Date(Date.UTC(2026, 9, 15)) }, // Thu
                { date: '2026-10-16', d: new Date(Date.UTC(2026, 9, 16)) }, // Fri
                { date: '2026-10-17', d: new Date(Date.UTC(2026, 9, 17)) }, // Sat
                { date: '2026-10-18', d: new Date(Date.UTC(2026, 9, 18)) }, // Sun
            ];
            for (const item of days) {
                expect(getBucketKey(item.d, true)).toBe('2026-10-12');
            }
        });
        it('handles week transitions across month boundaries for ISO Monday', () => {
            // 2026-11-01 is Sunday -> ISO Monday was 2026-10-26 (previous month)
            const dSunNov = new Date(Date.UTC(2026, 10, 1));
            expect(getBucketKey(dSunNov, true)).toBe('2026-10-26');
            // 2026-10-31 is Saturday -> ISO Monday was 2026-10-26
            const dSatOct = new Date(Date.UTC(2026, 9, 31));
            expect(getBucketKey(dSatOct, true)).toBe('2026-10-26');
        });
        it('handles week transitions across year boundaries for ISO Monday', () => {
            // 2026-01-01 is Thursday -> ISO Monday was 2025-12-29 (previous year)
            const dThuJan = new Date(Date.UTC(2026, 0, 1));
            expect(getBucketKey(dThuJan, true)).toBe('2025-12-29');
            // 2027-01-01 is Friday -> ISO Monday was 2026-12-28
            const dFriJan = new Date(Date.UTC(2027, 0, 1));
            expect(getBucketKey(dFriJan, true)).toBe('2026-12-28');
            // 2027-01-03 is Sunday -> ISO Monday was 2026-12-28
            const dSunJan = new Date(Date.UTC(2027, 0, 3));
            expect(getBucketKey(dSunJan, true)).toBe('2026-12-28');
        });
        it('returns YYYY-MM-DDTHH:00:00Z for hourly granularity', () => {
            const d = new Date(Date.UTC(2026, 9, 15, 14, 25, 30));
            expect(getBucketKey(d, 'hourly')).toBe('2026-10-15T14:00:00Z');
            const d2 = new Date(Date.UTC(2026, 0, 1, 0, 5, 0));
            expect(getBucketKey(d2, 'hourly')).toBe('2026-01-01T00:00:00Z');
        });
    });
    describe('aggregateMetricsOverTime', () => {
        it('returns empty array when ledgers is empty', () => {
            expect(aggregateMetricsOverTime([], 'daily')).toEqual([]);
        });
        it('aggregates daily buckets accurately from mock ledgers', () => {
            const ledgers = Object.values(mockAgenticMetricsLedgers);
            const buckets = aggregateMetricsOverTime(ledgers, 'daily');
            expect(buckets.length).toBeGreaterThan(0);
            for (let i = 0; i < buckets.length - 1; i++) {
                expect(buckets[i]!.bucket_date.localeCompare(buckets[i + 1]!.bucket_date)).toBeLessThan(0);
                expect(buckets[i]!.cumulative_cost_usd).toBeLessThanOrEqual(buckets[i + 1]!.cumulative_cost_usd);
                expect(buckets[i]!.cumulative_invocations).toBeLessThanOrEqual(buckets[i + 1]!.cumulative_invocations);
            }
        });
        it('aggregates weekly buckets accurately into ISO Monday keys', () => {
            const ledgers = Object.values(mockAgenticMetricsLedgers);
            const buckets = aggregateMetricsOverTime(ledgers, 'weekly');
            expect(buckets.length).toBeGreaterThan(0);
            for (const b of buckets) {
                const d = parseIsoDateToUtc(b.bucket_date);
                expect(d).not.toBeNull();
                // Day 1 is Monday in UTC
                expect(d?.getUTCDay()).toBe(1);
            }
        });
        it('aggregates hourly buckets accurately with YYYY-MM-DDTHH:00:00Z keys', () => {
            const ledgers = Object.values(mockAgenticMetricsLedgers);
            const buckets = aggregateMetricsOverTime(ledgers, 'hourly');
            expect(buckets.length).toBeGreaterThan(0);
            for (const b of buckets) {
                expect(b.bucket_key).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:00:00Z$/);
                expect(b.bucket_date).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:00:00Z$/);
            }
        });
        it('filters with sub-day datetime boundaries for hourly granularity', () => {
            const testLedger: AgenticMetricsLedger = {
                ticket_id: 'HOURLY-1',
                ticket_title: 'Hourly test',
                ticket_type: 'chore',
                started_at: '2026-10-19T09:00:00Z',
                completed_at: '2026-10-19T11:30:00Z',
                updated_at: '2026-10-19T11:30:00Z',
                facts: {
                    ticket_id: 'HOURLY-1',
                    ticket_title: 'Hourly test',
                    ticket_type: 'chore',
                    started_at: '2026-10-19T09:00:00Z',
                    completed_at: '2026-10-19T11:30:00Z',
                    updated_at: '2026-10-19T11:30:00Z',
                    code: {
                        lines_added: 0,
                        lines_deleted: 0,
                        files_changed: 0,
                        prod_lines_added: 0,
                        prod_lines_deleted: 0,
                        test_lines_added: 0,
                        test_lines_deleted: 0,
                        sdlc_iterations: 1,
                    },
                    sessions: [
                        {
                            id: "inv-1",
                            user: 'dev',
                            invocations: [
                                {
                                    role: 'coder',
                                    model: 'gpt-4',
                                    started_at: '2026-10-19T09:15:00Z',
                                    completed_at: '2026-10-19T09:20:00Z',
                                    active_duration_seconds: 300,
                                    turns: 2,
                                    tokens: { input: 100, output: 50, cache_read: 20, reasoning: 0, peak_context: 500 },
                                    cost_usd: 0.05,
                                    compactions: 0,
                                }
                            ],
                            tools: [],
                        },
                        {
                            id: "inv-2",
                            user: 'dev',
                            invocations: [
                                {
                                    role: 'coder',
                                    model: 'gpt-4',
                                    started_at: '2026-10-19T10:15:00Z',
                                    completed_at: '2026-10-19T10:20:00Z',
                                    active_duration_seconds: 300,
                                    turns: 2,
                                    tokens: { input: 200, output: 100, cache_read: 50, reasoning: 0, peak_context: 1000 },
                                    cost_usd: 0.10,
                                    compactions: 0,
                                }
                            ],
                            tools: [],
                        },
                        {
                            id: "inv-3",
                            user: 'dev',
                            invocations: [
                                {
                                    role: 'coder',
                                    model: 'gpt-4',
                                    started_at: '2026-10-19T11:15:00Z',
                                    completed_at: '2026-10-19T11:20:00Z',
                                    active_duration_seconds: 300,
                                    turns: 2,
                                    tokens: { input: 300, output: 150, cache_read: 80, reasoning: 0, peak_context: 1500 },
                                    cost_usd: 0.15,
                                    compactions: 0,
                                }
                            ],
                            tools: [],
                        }
                    ]
                },
                workflow_efficiency: {
                    sdlc_iterations: 1,
                    active_duration_seconds: 900,
                    active_duration_formatted: '15m',
                    role_effort_distribution_pct: {},
                    subagent_turn_velocity: { avg_turns_per_invocation: 2, peak_turns: 2, by_role: {} },
                    test_to_production_ratio: { ratio: '1.0', test_lines_added: 0, prod_lines_added: 0 },
                    reasoning_tokens: 0,
                    reasoning_ratio_pct: 0,
                    code_delivered: { lines_added: 0, lines_deleted: 0, files_changed: 0 },
                    tool_execution: { total_calls: 0, errors_and_corrections: 0, success_rate_pct: 100 },
                    context_window_health: {
                        healthy_ceiling_tokens: 200000,
                        status: 'HEALTHY',
                        root_peak_tokens: 0,
                        root_saturation_pct: 0,
                        subagents_peak_tokens: 0,
                        subagents_peak_role: 'coder',
                        compactions: 0,
                        peak_by_role: {},
                    },
                },
                threads: [],
                totals: {
                    invocations: 3,
                    input_tokens: 600,
                    output_tokens: 300,
                    cache_read_tokens: 150,
                    cache_hit_rate_pct: 20,
                    cost_usd: 0.30,
                    peak_context_tokens: 1500,
                    context_health: 'HEALTHY',
                },
                roles: [],
                models: [],
                anomalies: [],
            };
            // Filter between 10:00 and 10:45
            const filtered = aggregateMetricsOverTime([testLedger], 'hourly', '2026-10-19T10:00', '2026-10-19T10:45');
            expect(filtered.length).toBe(1);
            expect(filtered[0]?.bucket_key).toBe('2026-10-19T10:00:00Z');
            expect(filtered[0]?.invocations).toBe(1);
            expect(filtered[0]?.cost_usd).toBe(0.10);
        });
        it('defensively calculates cache hit rate when zero tokens are present', () => {
            const zeroInv: InvocationFact = {
                role: 'coder',
                model: 'test',
                started_at: '2026-10-15T10:00:00Z',
                completed_at: '2026-10-15T10:05:00Z',
                active_duration_seconds: 10,
                turns: 1,
                tokens: {
                    input: 0,
                    output: 0,
                    cache_read: 0,
                    reasoning: 0,
                    peak_context: 0,
                },
                cost_usd: 0,
                compactions: 0,
            };
            const ledger: AgenticMetricsLedger = {
                ticket_id: 'ZERO-1',
                ticket_title: 'Zero test',
                ticket_type: 'chore',
                started_at: '2026-10-15T10:00:00Z',
                completed_at: '2026-10-15T10:05:00Z',
                updated_at: '2026-10-15T10:05:00Z',
                facts: {
                    ticket_id: 'ZERO-1',
                    ticket_title: 'Zero test',
                    ticket_type: 'chore',
                    started_at: '2026-10-15T10:00:00Z',
                    completed_at: '2026-10-15T10:05:00Z',
                    updated_at: '2026-10-15T10:05:00Z',
                    code: {
                        lines_added: 0,
                        lines_deleted: 0,
                        test_lines_added: 0,
                        test_lines_deleted: 0,
                        prod_lines_added: 0,
                        prod_lines_deleted: 0,
                        files_changed: 0,
                        sdlc_iterations: 1,
                    },
                    sessions: [
                        {
                            id: "fixture-session",
                            invocations: [
                                zeroInv
                            ],
                            tools: [],
                        }
                    ]
                },
                workflow_efficiency: {
                    sdlc_iterations: 1,
                    active_duration_seconds: 10,
                    active_duration_formatted: '10s',
                    role_effort_distribution_pct: {},
                    subagent_turn_velocity: {
                        avg_turns_per_invocation: 1,
                        peak_turns: 1,
                        by_role: {},
                    },
                    test_to_production_ratio: {
                        ratio: '0:0',
                        test_lines_added: 0,
                        prod_lines_added: 0,
                    },
                    reasoning_tokens: 0,
                    reasoning_ratio_pct: 0,
                    code_delivered: {
                        lines_added: 0,
                        lines_deleted: 0,
                        files_changed: 0,
                    },
                    tool_execution: {
                        total_calls: 0,
                        errors_and_corrections: 0,
                        success_rate_pct: 100,
                    },
                    context_window_health: {
                        healthy_ceiling_tokens: 200000,
                        status: 'HEALTHY',
                        root_peak_tokens: 0,
                        root_saturation_pct: 0,
                        subagents_peak_tokens: 0,
                        subagents_peak_role: 'none',
                        compactions: 0,
                        peak_by_role: {},
                    },
                },
                threads: [],
                totals: {
                    invocations: 1,
                    input_tokens: 0,
                    output_tokens: 0,
                    cache_read_tokens: 0,
                    cache_hit_rate_pct: 0,
                    cost_usd: 0,
                    peak_context_tokens: 0,
                    context_health: 'HEALTHY',
                },
                roles: [],
                models: [],
            };
            const buckets = aggregateMetricsOverTime([ledger], 'daily');
            expect(buckets).toHaveLength(1);
            expect(buckets[0]?.cache_hit_rate_pct).toBe(0.0);
            expect(buckets[0]?.cost_usd).toBe(0);
            expect(buckets[0]?.cumulative_cost_usd).toBe(0);
            expect(buckets[0]?.net_lines_added).toBe(0);
        });
        it('filters invocations and code statistics by start and end date bounds', () => {
            const ledgers = Object.values(mockAgenticMetricsLedgers);
            // Ledgers cover 2026-10-14 through 2026-10-17
            const filtered = aggregateMetricsOverTime(ledgers, 'daily', '2026-10-15', '2026-10-16');
            expect(filtered.length).toBeGreaterThan(0);
            for (const b of filtered) {
                expect(b.bucket_date >= '2026-10-15').toBe(true);
                expect(b.bucket_date <= '2026-10-16').toBe(true);
            }
        });
        it('filters with startDate only, endDate only, and handles inverted or non-matching date ranges', () => {
            const ledgers = Object.values(mockAgenticMetricsLedgers);
            // Start date only
            const startOnly = aggregateMetricsOverTime(ledgers, 'daily', '2026-10-16', null);
            for (const b of startOnly) {
                expect(b.bucket_date >= '2026-10-16').toBe(true);
            }
            // End date only
            const endOnly = aggregateMetricsOverTime(ledgers, 'daily', null, '2026-10-15');
            for (const b of endOnly) {
                expect(b.bucket_date <= '2026-10-15').toBe(true);
            }
            // Inverted date range (start > end) returns empty array
            const inverted = aggregateMetricsOverTime(ledgers, 'daily', '2026-10-20', '2026-10-10');
            expect(inverted).toEqual([]);
            // Out of range dates with no matching data returns empty array
            const noMatch = aggregateMetricsOverTime(ledgers, 'daily', '2025-01-01', '2025-01-31');
            expect(noMatch).toEqual([]);
        });
        it('guarantees chronological sorting even when input ledgers and invocations are disordered', () => {
            const makeSimpleLedger = (id: string, date: string, cost: number): AgenticMetricsLedger => ({
                ticket_id: id,
                ticket_title: `Test ${id}`,
                ticket_type: 'chore',
                started_at: `${date}T08:00:00Z`,
                completed_at: `${date}T10:00:00Z`,
                updated_at: `${date}T10:00:00Z`,
                facts: {
                    ticket_id: id,
                    ticket_title: `Test ${id}`,
                    ticket_type: 'chore',
                    started_at: `${date}T08:00:00Z`,
                    completed_at: `${date}T10:00:00Z`,
                    updated_at: `${date}T10:00:00Z`,
                    code: {
                        lines_added: 10,
                        lines_deleted: 2,
                        test_lines_added: 5,
                        test_lines_deleted: 0,
                        prod_lines_added: 5,
                        prod_lines_deleted: 2,
                        files_changed: 1,
                        sdlc_iterations: 1,
                    },
                    sessions: [
                        {
                            id: "fixture-session",
                            user: 'agent',
                            invocations: [
                                {
                                    role: 'coder',
                                    model: 'test-model',
                                    started_at: `${date}T08:00:00Z`,
                                    completed_at: `${date}T10:00:00Z`,
                                    active_duration_seconds: 60,
                                    turns: 2,
                                    tokens: {
                                        input: 1000,
                                        output: 500,
                                        cache_read: 2000,
                                        reasoning: 100,
                                        peak_context: 5000,
                                    },
                                    cost_usd: cost,
                                    compactions: 0,
                                }
                            ],
                            tools: [],
                        }
                    ]
                },
                workflow_efficiency: {
                    sdlc_iterations: 1,
                    active_duration_seconds: 60,
                    active_duration_formatted: '1m 0s',
                    role_effort_distribution_pct: {},
                    subagent_turn_velocity: {
                        avg_turns_per_invocation: 2,
                        peak_turns: 2,
                        by_role: {},
                    },
                    test_to_production_ratio: { ratio: '1:1', test_lines_added: 5, prod_lines_added: 5 },
                    reasoning_tokens: 100,
                    reasoning_ratio_pct: 20,
                    code_delivered: { lines_added: 10, lines_deleted: 2, files_changed: 1 },
                    tool_execution: { total_calls: 0, errors_and_corrections: 0, success_rate_pct: 100 },
                    context_window_health: {
                        healthy_ceiling_tokens: 200000,
                        status: 'HEALTHY',
                        root_peak_tokens: 5000,
                        root_saturation_pct: 2.5,
                        subagents_peak_tokens: 0,
                        subagents_peak_role: 'none',
                        compactions: 0,
                        peak_by_role: {},
                    },
                },
                threads: [],
                totals: {
                    invocations: 1,
                    input_tokens: 1000,
                    output_tokens: 500,
                    cache_read_tokens: 2000,
                    cache_hit_rate_pct: 66.7,
                    cost_usd: cost,
                    peak_context_tokens: 5000,
                    context_health: 'HEALTHY',
                },
                roles: [],
                models: [],
            });
            // Pass ledgers in intentionally reverse/shuffled chronological order
            const l1 = makeSimpleLedger('L3', '2026-10-30', 3.0);
            const l2 = makeSimpleLedger('L1', '2026-10-05', 1.0);
            const l3 = makeSimpleLedger('L2', '2026-10-18', 2.0);
            const buckets = aggregateMetricsOverTime([l1, l2, l3], 'daily');
            expect(buckets).toHaveLength(3);
            expect(buckets[0]!.bucket_date).toBe('2026-10-05');
            expect(buckets[1]!.bucket_date).toBe('2026-10-18');
            expect(buckets[2]!.bucket_date).toBe('2026-10-30');
            // Assert cumulative values accumulate strictly across chronological sequence
            expect(buckets[0]!.cumulative_cost_usd).toBe(1.0);
            expect(buckets[1]!.cumulative_cost_usd).toBe(3.0);
            expect(buckets[2]!.cumulative_cost_usd).toBe(6.0);
            expect(buckets[0]!.cumulative_invocations).toBe(1);
            expect(buckets[1]!.cumulative_invocations).toBe(2);
            expect(buckets[2]!.cumulative_invocations).toBe(3);
        });
        it('verifies cumulative horizon calculations match running totals across all buckets', () => {
            const ledgers = Object.values(mockAgenticMetricsLedgers);
            const buckets = aggregateMetricsOverTime(ledgers, 'daily');
            expect(buckets.length).toBeGreaterThan(0);
            let runningCost = 0;
            let runningInputs = 0;
            let runningOutputs = 0;
            let runningInvocations = 0;
            let runningLinesAdded = 0;
            let runningLinesDeleted = 0;
            for (const b of buckets) {
                runningCost += b.cost_usd;
                runningInputs += b.input_tokens;
                runningOutputs += b.output_tokens;
                runningInvocations += b.invocations;
                runningLinesAdded += b.lines_added;
                runningLinesDeleted += b.lines_deleted;
                expect(b.cumulative_cost_usd).toBeCloseTo(Math.round(runningCost * 10000) / 10000, 3);
                expect(b.cumulative_input_tokens).toBe(runningInputs);
                expect(b.cumulative_output_tokens).toBe(runningOutputs);
                expect(b.cumulative_invocations).toBe(runningInvocations);
                expect(b.cumulative_lines_added).toBe(runningLinesAdded);
                expect(b.cumulative_lines_deleted).toBe(runningLinesDeleted);
                expect(b.cumulative_net_lines_added).toBe(runningLinesAdded - runningLinesDeleted);
            }
        });
        it('accurately accumulates multiple invocations on the same day and tracks peak context', () => {
            const inv1: InvocationFact = {
                role: 'orchestrator',
                model: 'model-a',
                started_at: '2026-10-15T09:00:00Z',
                completed_at: '2026-10-15T09:10:00Z',
                active_duration_seconds: 600,
                turns: 1,
                tokens: {
                    input: 1000,
                    output: 200,
                    cache_read: 4000,
                    reasoning: 50,
                    peak_context: 50000,
                },
                cost_usd: 0.15,
                compactions: 0,
            };
            const inv2: InvocationFact = {
                role: 'coder',
                model: 'model-b',
                started_at: '2026-10-15T10:00:00Z',
                completed_at: '2026-10-15T10:15:00Z',
                active_duration_seconds: 900,
                turns: 3,
                tokens: {
                    input: 2000,
                    output: 800,
                    cache_read: 6000,
                    reasoning: 150,
                    peak_context: 180000, // Higher peak context
                },
                cost_usd: 0.35,
                compactions: 1,
            };
            const inv3: InvocationFact = {
                role: 'tester',
                model: 'model-c',
                started_at: '2026-10-15T11:00:00Z',
                completed_at: '2026-10-15T11:05:00Z',
                active_duration_seconds: 300,
                turns: 1,
                tokens: {
                    input: 500,
                    output: 100,
                    cache_read: 1500,
                    reasoning: 0,
                    peak_context: 120000,
                },
                cost_usd: 0.05,
                compactions: 0,
            };
            const ledger: AgenticMetricsLedger = {
                ticket_id: 'MULTI-1',
                ticket_title: 'Multi invocation test',
                ticket_type: 'chore',
                started_at: '2026-10-15T09:00:00Z',
                completed_at: '2026-10-15T11:05:00Z',
                updated_at: '2026-10-15T11:05:00Z',
                facts: {
                    ticket_id: 'MULTI-1',
                    ticket_title: 'Multi invocation test',
                    ticket_type: 'chore',
                    started_at: '2026-10-15T09:00:00Z',
                    completed_at: '2026-10-15T11:05:00Z',
                    updated_at: '2026-10-15T11:05:00Z',
                    code: {
                        lines_added: 120,
                        lines_deleted: 30,
                        test_lines_added: 60,
                        test_lines_deleted: 10,
                        prod_lines_added: 60,
                        prod_lines_deleted: 20,
                        files_changed: 4,
                        sdlc_iterations: 2,
                    },
                    sessions: [
                        {
                            id: "fixture-session",
                            invocations: [
                                inv1, inv2, inv3
                            ],
                            tools: [],
                        }
                    ]
                },
                workflow_efficiency: {
                    sdlc_iterations: 2,
                    active_duration_seconds: 1800,
                    active_duration_formatted: '30m 0s',
                    role_effort_distribution_pct: {},
                    subagent_turn_velocity: {
                        avg_turns_per_invocation: 1.67,
                        peak_turns: 3,
                        by_role: {},
                    },
                    test_to_production_ratio: { ratio: '1.0:1', test_lines_added: 60, prod_lines_added: 60 },
                    reasoning_tokens: 200,
                    reasoning_ratio_pct: 18.2,
                    code_delivered: { lines_added: 120, lines_deleted: 30, files_changed: 4 },
                    tool_execution: { total_calls: 3, errors_and_corrections: 0, success_rate_pct: 100 },
                    context_window_health: {
                        healthy_ceiling_tokens: 200000,
                        status: 'HEALTHY',
                        root_peak_tokens: 50000,
                        root_saturation_pct: 25.0,
                        subagents_peak_tokens: 180000,
                        subagents_peak_role: 'coder',
                        compactions: 1,
                        peak_by_role: {},
                    },
                },
                threads: [],
                totals: {
                    invocations: 3,
                    input_tokens: 3500,
                    output_tokens: 1100,
                    cache_read_tokens: 11500,
                    cache_hit_rate_pct: 76.7,
                    cost_usd: 0.55,
                    peak_context_tokens: 180000,
                    context_health: 'HEALTHY',
                },
                roles: [],
                models: [],
            };
            const buckets = aggregateMetricsOverTime([ledger], 'daily');
            expect(buckets).toHaveLength(1);
            const b = buckets[0]!;
            expect(b.bucket_date).toBe('2026-10-15');
            expect(b.invocations).toBe(3);
            expect(b.cost_usd).toBe(0.55);
            expect(b.input_tokens).toBe(3500);
            expect(b.output_tokens).toBe(1100);
            expect(b.cache_read_tokens).toBe(11500);
            expect(b.reasoning_tokens).toBe(200);
            expect(b.active_duration_seconds).toBe(1800);
            // Peak context must be the maximum of the 3 invocations (180,000)
            expect(b.peak_context_tokens).toBe(180000);
            // Code delivery
            expect(b.lines_added).toBe(120);
            expect(b.lines_deleted).toBe(30);
            expect(b.files_changed).toBe(4);
            expect(b.net_lines_added).toBe(90);
            // Cache hit rate: 11500 / (3500 + 11500) = 11500 / 15000 = 76.666% -> 76.7%
            expect(b.cache_hit_rate_pct).toBe(76.7);
        });
        it('handles negative net lines when lines deleted exceeds lines added', () => {
            const inv: InvocationFact = {
                role: 'coder',
                model: 'model',
                started_at: '2026-10-20T10:00:00Z',
                completed_at: '2026-10-20T10:10:00Z',
                active_duration_seconds: 60,
                turns: 1,
                tokens: { input: 100, output: 50, cache_read: 0, reasoning: 0, peak_context: 500 },
                cost_usd: 0.01,
                compactions: 0,
            };
            const ledger: AgenticMetricsLedger = {
                ticket_id: 'CLEANUP-1',
                ticket_title: 'Code cleanup',
                ticket_type: 'refactor',
                started_at: '2026-10-20T10:00:00Z',
                completed_at: '2026-10-20T10:10:00Z',
                updated_at: '2026-10-20T10:10:00Z',
                facts: {
                    ticket_id: 'CLEANUP-1',
                    ticket_title: 'Code cleanup',
                    ticket_type: 'refactor',
                    started_at: '2026-10-20T10:00:00Z',
                    completed_at: '2026-10-20T10:10:00Z',
                    updated_at: '2026-10-20T10:10:00Z',
                    code: {
                        lines_added: 25,
                        lines_deleted: 150,
                        test_lines_added: 0,
                        test_lines_deleted: 0,
                        prod_lines_added: 25,
                        prod_lines_deleted: 150,
                        files_changed: 5,
                        sdlc_iterations: 1,
                    },
                    sessions: [
                        {
                            id: "fixture-session",
                            invocations: [
                                inv
                            ],
                            tools: [],
                        }
                    ]
                },
                workflow_efficiency: {
                    sdlc_iterations: 1,
                    active_duration_seconds: 60,
                    active_duration_formatted: '1m',
                    role_effort_distribution_pct: {},
                    subagent_turn_velocity: { avg_turns_per_invocation: 1, peak_turns: 1, by_role: {} },
                    test_to_production_ratio: { ratio: '0:1', test_lines_added: 0, prod_lines_added: 25 },
                    reasoning_tokens: 0,
                    reasoning_ratio_pct: 0,
                    code_delivered: { lines_added: 25, lines_deleted: 150, files_changed: 5 },
                    tool_execution: { total_calls: 0, errors_and_corrections: 0, success_rate_pct: 100 },
                    context_window_health: {
                        healthy_ceiling_tokens: 200000,
                        status: 'HEALTHY',
                        root_peak_tokens: 500,
                        root_saturation_pct: 0.25,
                        subagents_peak_tokens: 0,
                        subagents_peak_role: 'none',
                        compactions: 0,
                        peak_by_role: {},
                    },
                },
                threads: [],
                totals: {
                    invocations: 1,
                    input_tokens: 100,
                    output_tokens: 50,
                    cache_read_tokens: 0,
                    cache_hit_rate_pct: 0,
                    cost_usd: 0.01,
                    peak_context_tokens: 500,
                    context_health: 'HEALTHY',
                },
                roles: [],
                models: [],
            };
            const buckets = aggregateMetricsOverTime([ledger], 'daily');
            expect(buckets).toHaveLength(1);
            expect(buckets[0]!.lines_added).toBe(25);
            expect(buckets[0]!.lines_deleted).toBe(150);
            expect(buckets[0]!.net_lines_added).toBe(-125);
        });
        it('guards cache hit rate calculation edge cases (100% hit rate, 0% hit rate)', () => {
            // 100% cache hit: input = 0, cache_read = 500
            const inv100: InvocationFact = {
                role: 'coder',
                model: 'test',
                started_at: '2026-10-15T10:00:00Z',
                completed_at: '2026-10-15T10:05:00Z',
                active_duration_seconds: 10,
                turns: 1,
                tokens: { input: 0, output: 50, cache_read: 500, reasoning: 0, peak_context: 550 },
                cost_usd: 0.01,
                compactions: 0,
            };
            // 0% cache hit: input = 500, cache_read = 0
            const inv0: InvocationFact = {
                role: 'coder',
                model: 'test',
                started_at: '2026-10-16T10:00:00Z',
                completed_at: '2026-10-16T10:05:00Z',
                active_duration_seconds: 10,
                turns: 1,
                tokens: { input: 500, output: 50, cache_read: 0, reasoning: 0, peak_context: 550 },
                cost_usd: 0.02,
                compactions: 0,
            };
            const makeLedgerWithInv = (id: string, inv: InvocationFact): AgenticMetricsLedger => ({
                ticket_id: id,
                ticket_title: id,
                ticket_type: 'chore',
                started_at: inv.started_at,
                completed_at: inv.completed_at,
                updated_at: inv.completed_at ?? inv.started_at ?? '2026-10-15T00:00:00Z',
                facts: {
                    ticket_id: id,
                    ticket_title: id,
                    ticket_type: 'chore',
                    started_at: inv.started_at,
                    completed_at: inv.completed_at,
                    updated_at: inv.completed_at ?? inv.started_at ?? '2026-10-15T00:00:00Z',
                    code: {
                        lines_added: 0,
                        lines_deleted: 0,
                        test_lines_added: 0,
                        test_lines_deleted: 0,
                        prod_lines_added: 0,
                        prod_lines_deleted: 0,
                        files_changed: 0,
                        sdlc_iterations: 1,
                    },
                    sessions: [
                        {
                            id: "fixture-session",
                            invocations: [
                                inv
                            ],
                            tools: [],
                        }
                    ]
                },
                workflow_efficiency: {
                    sdlc_iterations: 1,
                    active_duration_seconds: 10,
                    active_duration_formatted: '10s',
                    role_effort_distribution_pct: {},
                    subagent_turn_velocity: { avg_turns_per_invocation: 1, peak_turns: 1, by_role: {} },
                    test_to_production_ratio: { ratio: '0:0', test_lines_added: 0, prod_lines_added: 0 },
                    reasoning_tokens: 0,
                    reasoning_ratio_pct: 0,
                    code_delivered: { lines_added: 0, lines_deleted: 0, files_changed: 0 },
                    tool_execution: { total_calls: 0, errors_and_corrections: 0, success_rate_pct: 100 },
                    context_window_health: {
                        healthy_ceiling_tokens: 200000,
                        status: 'HEALTHY',
                        root_peak_tokens: 0,
                        root_saturation_pct: 0,
                        subagents_peak_tokens: 0,
                        subagents_peak_role: 'none',
                        compactions: 0,
                        peak_by_role: {},
                    },
                },
                threads: [],
                totals: {
                    invocations: 1,
                    input_tokens: inv.tokens.input,
                    output_tokens: inv.tokens.output,
                    cache_read_tokens: inv.tokens.cache_read,
                    cache_hit_rate_pct: inv.tokens.input === 0 ? 100 : 0,
                    cost_usd: inv.cost_usd,
                    peak_context_tokens: inv.tokens.peak_context,
                    context_health: 'HEALTHY',
                },
                roles: [],
                models: [],
            });
            const buckets = aggregateMetricsOverTime([makeLedgerWithInv('T1', inv100), makeLedgerWithInv('T2', inv0)], 'daily');
            expect(buckets).toHaveLength(2);
            expect(buckets[0]!.bucket_date).toBe('2026-10-15');
            expect(buckets[0]!.cache_hit_rate_pct).toBe(100.0);
            expect(buckets[1]!.bucket_date).toBe('2026-10-16');
            expect(buckets[1]!.cache_hit_rate_pct).toBe(0.0);
        });
        it('safely skips invocations where all date timestamps are corrupted without panicking', () => {
            const corruptInv: InvocationFact = {
                role: 'coder',
                model: 'test',
                started_at: 'totally-invalid-timestamp',
                completed_at: 'another-corrupt-date',
                active_duration_seconds: 10,
                turns: 1,
                tokens: { input: 100, output: 50, cache_read: 0, reasoning: 0, peak_context: 200 },
                cost_usd: 0.01,
                compactions: 0,
            };
            const corruptLedger: AgenticMetricsLedger = {
                ticket_id: 'CORRUPT-1',
                ticket_title: 'Corrupt Dates',
                ticket_type: 'chore',
                started_at: 'bad-start',
                completed_at: 'bad-end',
                updated_at: 'bad-update',
                facts: {
                    ticket_id: 'CORRUPT-1',
                    ticket_title: 'Corrupt Dates',
                    ticket_type: 'chore',
                    started_at: 'bad-start',
                    completed_at: 'bad-end',
                    updated_at: 'bad-update',
                    code: {
                        lines_added: 50,
                        lines_deleted: 10,
                        test_lines_added: 0,
                        test_lines_deleted: 0,
                        prod_lines_added: 50,
                        prod_lines_deleted: 10,
                        files_changed: 1,
                        sdlc_iterations: 1,
                    },
                    sessions: [
                        {
                            id: "fixture-session",
                            invocations: [
                                corruptInv
                            ],
                            tools: [],
                        }
                    ]
                },
                workflow_efficiency: {
                    sdlc_iterations: 1,
                    active_duration_seconds: 10,
                    active_duration_formatted: '10s',
                    role_effort_distribution_pct: {},
                    subagent_turn_velocity: { avg_turns_per_invocation: 1, peak_turns: 1, by_role: {} },
                    test_to_production_ratio: { ratio: '0:0', test_lines_added: 0, prod_lines_added: 0 },
                    reasoning_tokens: 0,
                    reasoning_ratio_pct: 0,
                    code_delivered: { lines_added: 50, lines_deleted: 10, files_changed: 1 },
                    tool_execution: { total_calls: 0, errors_and_corrections: 0, success_rate_pct: 100 },
                    context_window_health: {
                        healthy_ceiling_tokens: 200000,
                        status: 'HEALTHY',
                        root_peak_tokens: 0,
                        root_saturation_pct: 0,
                        subagents_peak_tokens: 0,
                        subagents_peak_role: 'none',
                        compactions: 0,
                        peak_by_role: {},
                    },
                },
                threads: [],
                totals: {
                    invocations: 1,
                    input_tokens: 100,
                    output_tokens: 50,
                    cache_read_tokens: 0,
                    cache_hit_rate_pct: 0,
                    cost_usd: 0.01,
                    peak_context_tokens: 200,
                    context_health: 'HEALTHY',
                },
                roles: [],
                models: [],
            };
            // Must not throw or crash, returns empty buckets array
            const buckets = aggregateMetricsOverTime([corruptLedger], 'daily');
            expect(buckets).toEqual([]);
        });
        it('does not allocate to time buckets when completed_at is null on invocation and ledger', () => {
            const fallbackInv: InvocationFact = {
                role: 'coder',
                model: 'test',
                started_at: '2026-10-11T10:00:00Z',
                completed_at: null, // missing
                active_duration_seconds: 30,
                turns: 1,
                tokens: {
                    input: 100,
                    output: 50,
                    cache_read: 200,
                    reasoning: 0,
                    peak_context: 500,
                },
                cost_usd: 0.05,
                compactions: 0,
            };
            const ledger: AgenticMetricsLedger = {
                ticket_id: 'FB-1',
                ticket_title: 'Fallback test',
                ticket_type: 'chore',
                started_at: '2026-10-11T10:00:00Z',
                completed_at: null,
                updated_at: '2026-10-11T10:30:00Z',
                facts: {
                    ticket_id: 'FB-1',
                    ticket_title: 'Fallback test',
                    ticket_type: 'chore',
                    started_at: '2026-10-11T10:00:00Z',
                    completed_at: null,
                    updated_at: '2026-10-11T10:30:00Z',
                    code: {
                        lines_added: 10,
                        lines_deleted: 2,
                        test_lines_added: 5,
                        test_lines_deleted: 0,
                        prod_lines_added: 5,
                        prod_lines_deleted: 2,
                        files_changed: 1,
                        sdlc_iterations: 1,
                    },
                    sessions: [
                        {
                            id: "fixture-session",
                            invocations: [
                                fallbackInv
                            ],
                            tools: [],
                        }
                    ]
                },
                workflow_efficiency: {
                    sdlc_iterations: 1,
                    active_duration_seconds: 30,
                    active_duration_formatted: '30s',
                    role_effort_distribution_pct: {},
                    subagent_turn_velocity: {
                        avg_turns_per_invocation: 1,
                        peak_turns: 1,
                        by_role: {},
                    },
                    test_to_production_ratio: {
                        ratio: '1.0:1',
                        test_lines_added: 5,
                        prod_lines_added: 5,
                    },
                    reasoning_tokens: 0,
                    reasoning_ratio_pct: 0,
                    code_delivered: {
                        lines_added: 10,
                        lines_deleted: 2,
                        files_changed: 1,
                    },
                    tool_execution: {
                        total_calls: 1,
                        errors_and_corrections: 0,
                        success_rate_pct: 100,
                    },
                    context_window_health: {
                        healthy_ceiling_tokens: 200000,
                        status: 'HEALTHY',
                        root_peak_tokens: 500,
                        root_saturation_pct: 0.25,
                        subagents_peak_tokens: 0,
                        subagents_peak_role: 'none',
                        compactions: 0,
                        peak_by_role: {},
                    },
                },
                threads: [],
                totals: {
                    invocations: 1,
                    input_tokens: 100,
                    output_tokens: 50,
                    cache_read_tokens: 200,
                    cache_hit_rate_pct: 66.7,
                    cost_usd: 0.05,
                    peak_context_tokens: 500,
                    context_health: 'HEALTHY',
                },
                roles: [],
                models: [],
            };
            const buckets = aggregateMetricsOverTime([ledger], 'daily');
            expect(buckets).toHaveLength(0);
        });
        it('falls back to ledger completed_at when invocation completed_at is null but ledger is completed', () => {
            const fallbackInv: InvocationFact = {
                role: 'coder',
                model: 'test',
                started_at: '2026-10-11T10:00:00Z',
                completed_at: null, // missing
                active_duration_seconds: 30,
                turns: 1,
                tokens: {
                    input: 100,
                    output: 50,
                    cache_read: 200,
                    reasoning: 0,
                    peak_context: 500,
                },
                cost_usd: 0.05,
                compactions: 0,
            };
            const ledger: AgenticMetricsLedger = {
                ticket_id: 'FB-2',
                ticket_title: 'Fallback completed test',
                ticket_type: 'chore',
                started_at: '2026-10-11T10:00:00Z',
                completed_at: '2026-10-11T10:30:00Z',
                updated_at: '2026-10-11T10:30:00Z',
                facts: {
                    ticket_id: 'FB-2',
                    ticket_title: 'Fallback completed test',
                    ticket_type: 'chore',
                    started_at: '2026-10-11T10:00:00Z',
                    completed_at: '2026-10-11T10:30:00Z',
                    updated_at: '2026-10-11T10:30:00Z',
                    code: {
                        lines_added: 10,
                        lines_deleted: 2,
                        test_lines_added: 5,
                        test_lines_deleted: 0,
                        prod_lines_added: 5,
                        prod_lines_deleted: 2,
                        files_changed: 1,
                        sdlc_iterations: 1,
                    },
                    sessions: [
                        {
                            id: "fixture-session",
                            invocations: [
                                fallbackInv
                            ],
                            tools: [],
                        }
                    ]
                },
                workflow_efficiency: {
                    sdlc_iterations: 1,
                    active_duration_seconds: 30,
                    active_duration_formatted: '30s',
                    role_effort_distribution_pct: {},
                    subagent_turn_velocity: {
                        avg_turns_per_invocation: 1,
                        peak_turns: 1,
                        by_role: {},
                    },
                    test_to_production_ratio: {
                        ratio: '1.0:1',
                        test_lines_added: 5,
                        prod_lines_added: 5,
                    },
                    reasoning_tokens: 0,
                    reasoning_ratio_pct: 0,
                    code_delivered: {
                        lines_added: 10,
                        lines_deleted: 2,
                        files_changed: 1,
                    },
                    tool_execution: {
                        total_calls: 1,
                        errors_and_corrections: 0,
                        success_rate_pct: 100,
                    },
                    context_window_health: {
                        healthy_ceiling_tokens: 200000,
                        status: 'HEALTHY',
                        root_peak_tokens: 500,
                        root_saturation_pct: 0.25,
                        subagents_peak_tokens: 0,
                        subagents_peak_role: 'none',
                        compactions: 0,
                        peak_by_role: {},
                    },
                },
                threads: [],
                totals: {
                    invocations: 1,
                    input_tokens: 100,
                    output_tokens: 50,
                    cache_read_tokens: 200,
                    cache_hit_rate_pct: 66.7,
                    cost_usd: 0.05,
                    peak_context_tokens: 500,
                    context_health: 'HEALTHY',
                },
                roles: [],
                models: [],
            };
            const buckets = aggregateMetricsOverTime([ledger], 'daily');
            expect(buckets).toHaveLength(1);
            expect(buckets[0]?.bucket_date).toBe('2026-10-11');
            expect(buckets[0]?.invocations).toBe(1);
            expect(buckets[0]?.lines_added).toBe(10);
            expect(buckets[0]?.net_lines_added).toBe(8);
        });
        it('adversarial: ignores tickets with empty, whitespace, or corrupt completed_at even if started_at and updated_at are populated', () => {
            const invalidCompletedValues = ['', '   ', 'not-a-valid-date', '2026-99-99', null];
            for (const invalidCompleted of invalidCompletedValues) {
                const invWithoutCompleted: InvocationFact = {
                    role: 'coder',
                    model: 'test',
                    started_at: '2026-10-15T10:00:00Z',
                    completed_at: null,
                    active_duration_seconds: 30,
                    turns: 1,
                    tokens: { input: 100, output: 50, cache_read: 0, reasoning: 0, peak_context: 500 },
                    cost_usd: 0.05,
                    compactions: 0,
                };
                const ledger: AgenticMetricsLedger = {
                    ticket_id: 'ADV-1',
                    ticket_title: 'Adversarial invalid completed_at test',
                    ticket_type: 'chore',
                    started_at: '2026-10-15T10:00:00Z',
                    completed_at: invalidCompleted as string | null,
                    updated_at: '2026-10-15T10:30:00Z',
                    facts: {
                        ticket_id: 'ADV-1',
                        ticket_title: 'Adversarial invalid completed_at test',
                        ticket_type: 'chore',
                        started_at: '2026-10-15T10:00:00Z',
                        completed_at: invalidCompleted as string | null,
                        updated_at: '2026-10-15T10:30:00Z',
                        code: {
                            lines_added: 50,
                            lines_deleted: 10,
                            test_lines_added: 20,
                            test_lines_deleted: 0,
                            prod_lines_added: 30,
                            prod_lines_deleted: 10,
                            files_changed: 2,
                            sdlc_iterations: 1,
                        },
                        sessions: [
                            {
                                id: "fixture-session",
                                invocations: [
                                    invWithoutCompleted
                                ],
                                tools: []
                            }
                        ]
                    },
                    workflow_efficiency: {
                        sdlc_iterations: 1,
                        active_duration_seconds: 30,
                        active_duration_formatted: '30s',
                        role_effort_distribution_pct: {},
                        subagent_turn_velocity: { avg_turns_per_invocation: 1, peak_turns: 1, by_role: {} },
                        test_to_production_ratio: { ratio: '1.0:1', test_lines_added: 20, prod_lines_added: 30 },
                        reasoning_tokens: 0,
                        reasoning_ratio_pct: 0,
                        code_delivered: { lines_added: 50, lines_deleted: 10, files_changed: 2 },
                        tool_execution: { total_calls: 1, errors_and_corrections: 0, success_rate_pct: 100 },
                        context_window_health: {
                            healthy_ceiling_tokens: 200000,
                            status: 'HEALTHY',
                            root_peak_tokens: 500,
                            root_saturation_pct: 0.25,
                            subagents_peak_tokens: 0,
                            subagents_peak_role: 'none',
                            compactions: 0,
                            peak_by_role: {},
                        },
                    },
                    threads: [],
                    totals: {
                        invocations: 1,
                        input_tokens: 100,
                        output_tokens: 50,
                        cache_read_tokens: 0,
                        cache_hit_rate_pct: 0,
                        cost_usd: 0.05,
                        peak_context_tokens: 500,
                        context_health: 'HEALTHY',
                    },
                    roles: [],
                    models: [],
                };
                const buckets = aggregateMetricsOverTime([ledger], 'daily');
                expect(buckets).toHaveLength(0);
            }
        });
        it('adversarial: handles midnight transitions, timezone offsets, and empty inputs', () => {
            // Midnight transition: 23:50 on 2026-10-18 and 00:10 on 2026-10-19
            const invLate: InvocationFact = {
                role: 'coder',
                model: 'm',
                started_at: '2026-10-18T23:50:00Z',
                completed_at: '2026-10-18T23:55:00Z',
                active_duration_seconds: 60,
                turns: 1,
                tokens: { input: 100, output: 50, cache_read: 0, reasoning: 0, peak_context: 500 },
                cost_usd: 0.01,
                compactions: 0,
            };
            const invEarly: InvocationFact = {
                role: 'coder',
                model: 'm',
                // 02:10 +02:00 is 00:10 UTC on 2026-10-19
                started_at: '2026-10-19T02:05:00+02:00',
                completed_at: '2026-10-19T02:10:00+02:00',
                active_duration_seconds: 60,
                turns: 1,
                tokens: { input: 200, output: 80, cache_read: 50, reasoning: 0, peak_context: 800 },
                cost_usd: 0.02,
                compactions: 0,
            };
            const ledger: AgenticMetricsLedger = {
                ticket_id: 'ADV-1',
                ticket_title: 'Adversarial Test',
                ticket_type: 'chore',
                started_at: '2026-10-18T23:50:00Z',
                completed_at: '2026-10-19T00:15:00Z',
                updated_at: '2026-10-19T00:15:00Z',
                facts: {
                    ticket_id: 'ADV-1',
                    ticket_title: 'Adversarial Test',
                    ticket_type: 'chore',
                    started_at: '2026-10-18T23:50:00Z',
                    completed_at: '2026-10-19T00:15:00Z',
                    updated_at: '2026-10-19T00:15:00Z',
                    code: {
                        lines_added: 20,
                        lines_deleted: 5,
                        test_lines_added: 10,
                        test_lines_deleted: 0,
                        prod_lines_added: 10,
                        prod_lines_deleted: 5,
                        files_changed: 2,
                        sdlc_iterations: 1,
                    },
                    sessions: [
                        {
                            id: "fixture-session",
                            invocations: [
                                invLate, invEarly
                            ],
                            tools: [],
                        }
                    ]
                },
                workflow_efficiency: {
                    sdlc_iterations: 1,
                    active_duration_seconds: 120,
                    active_duration_formatted: '2m',
                    role_effort_distribution_pct: {},
                    subagent_turn_velocity: { avg_turns_per_invocation: 1, peak_turns: 1, by_role: {} },
                    test_to_production_ratio: { ratio: '1.0', test_lines_added: 10, prod_lines_added: 10 },
                    reasoning_tokens: 0,
                    reasoning_ratio_pct: 0,
                    code_delivered: { lines_added: 20, lines_deleted: 5, files_changed: 2 },
                    tool_execution: { total_calls: 0, errors_and_corrections: 0, success_rate_pct: 100 },
                    context_window_health: {
                        healthy_ceiling_tokens: 200000,
                        status: 'HEALTHY',
                        root_peak_tokens: 0,
                        root_saturation_pct: 0,
                        subagents_peak_tokens: 0,
                        subagents_peak_role: 'coder',
                        compactions: 0,
                        peak_by_role: {},
                    },
                },
                threads: [],
                totals: {
                    invocations: 2,
                    input_tokens: 300,
                    output_tokens: 130,
                    cache_read_tokens: 50,
                    cache_hit_rate_pct: 14.3,
                    cost_usd: 0.03,
                    peak_context_tokens: 800,
                    context_health: 'HEALTHY',
                },
                roles: [],
                models: [],
            };
            const buckets = aggregateMetricsOverTime([ledger], 'hourly');
            expect(buckets).toHaveLength(2);
            expect(buckets[0]?.bucket_key).toBe('2026-10-18T23:00:00Z');
            expect(buckets[1]?.bucket_key).toBe('2026-10-19T00:00:00Z');
            expect(buckets[0]?.invocations).toBe(1);
            expect(buckets[1]?.invocations).toBe(1);
            expect(buckets[1]?.cumulative_invocations).toBe(2);
            // Inverted date range returns empty array
            const inverted = aggregateMetricsOverTime([ledger], 'hourly', '2026-10-19T12:00', '2026-10-19T08:00');
            expect(inverted).toHaveLength(0);
            // Empty and null inputs
            expect(aggregateMetricsOverTime([], 'hourly')).toHaveLength(0);
            expect(aggregateMetricsOverTime(null as unknown as readonly AgenticMetricsLedger[], 'hourly')).toHaveLength(0);
        });
        it('AL-34: strictly verifies cumulative_net_lines_added across positive, negative, and zero churn buckets', () => {
            const makeLedger = (id: string, dateStr: string, added: number, deleted: number): AgenticMetricsLedger => ({
                ticket_id: id,
                ticket_title: `Ticket ${id}`,
                ticket_type: 'feat',
                started_at: `${dateStr}T08:00:00Z`,
                completed_at: `${dateStr}T12:00:00Z`,
                updated_at: `${dateStr}T12:00:00Z`,
                facts: {
                    ticket_id: id,
                    ticket_title: `Ticket ${id}`,
                    ticket_type: 'feat',
                    started_at: `${dateStr}T08:00:00Z`,
                    completed_at: `${dateStr}T12:00:00Z`,
                    updated_at: `${dateStr}T12:00:00Z`,
                    code: {
                        lines_added: added,
                        lines_deleted: deleted,
                        files_changed: 2,
                        prod_lines_added: added,
                        prod_lines_deleted: deleted,
                        test_lines_added: 0,
                        test_lines_deleted: 0,
                        sdlc_iterations: 1,
                    },
                    sessions: [
                        {
                            id: "fixture-session",
                            user: 'engineer',
                            invocations: [
                                {
                                    role: 'coder',
                                    model: 'model-x',
                                    started_at: `${dateStr}T08:00:00Z`,
                                    completed_at: `${dateStr}T12:00:00Z`,
                                    active_duration_seconds: 3600,
                                    turns: 4,
                                    compactions: 0,
                                    tokens: { input: 100, output: 50, cache_read: 20, reasoning: 0, peak_context: 500 },
                                    cost_usd: 0.05,
                                }
                            ],
                            tools: [],
                        }
                    ]
                },
                workflow_efficiency: {
                    sdlc_iterations: 1,
                    active_duration_seconds: 3600,
                    active_duration_formatted: '1h',
                    role_effort_distribution_pct: {},
                    subagent_turn_velocity: { avg_turns_per_invocation: 4, peak_turns: 4, by_role: {} },
                    test_to_production_ratio: { ratio: '1.0', test_lines_added: 0, prod_lines_added: added },
                    reasoning_tokens: 0,
                    reasoning_ratio_pct: 0,
                    code_delivered: { lines_added: added, lines_deleted: deleted, files_changed: 2 },
                    tool_execution: { total_calls: 0, errors_and_corrections: 0, success_rate_pct: 100 },
                    context_window_health: {
                        healthy_ceiling_tokens: 200000,
                        status: 'HEALTHY',
                        root_peak_tokens: 0,
                        root_saturation_pct: 0,
                        subagents_peak_tokens: 0,
                        subagents_peak_role: 'coder',
                        compactions: 0,
                        peak_by_role: {},
                    },
                },
                threads: [],
                totals: {
                    invocations: 1,
                    input_tokens: 100,
                    output_tokens: 50,
                    cache_read_tokens: 20,
                    cache_hit_rate_pct: 16.7,
                    cost_usd: 0.05,
                    peak_context_tokens: 500,
                    context_health: 'HEALTHY',
                },
                roles: [],
                models: [],
            });
            const ledgers: AgenticMetricsLedger[] = [
                makeLedger('T-1', '2026-10-15', 1000, 200), // Net: +800, Cum Net: +800
                makeLedger('T-2', '2026-10-16', 100, 1200), // Net: -1100, Cum Net: -300
                makeLedger('T-3', '2026-10-17', 0, 0), // Net: 0, Cum Net: -300
                makeLedger('T-4', '2026-10-18', 500, 50), // Net: +450, Cum Net: +150
            ];
            const daily = aggregateMetricsOverTime(ledgers, 'daily');
            expect(daily).toHaveLength(4);
            // Bucket 0
            expect(daily[0]?.net_lines_added).toBe(800);
            expect(daily[0]?.cumulative_lines_added).toBe(1000);
            expect(daily[0]?.cumulative_lines_deleted).toBe(200);
            expect(daily[0]?.cumulative_net_lines_added).toBe(800);
            // Bucket 1 (cumulative dips negative)
            expect(daily[1]?.net_lines_added).toBe(-1100);
            expect(daily[1]?.cumulative_lines_added).toBe(1100);
            expect(daily[1]?.cumulative_lines_deleted).toBe(1400);
            expect(daily[1]?.cumulative_net_lines_added).toBe(-300);
            // Bucket 2 (neutral churn)
            expect(daily[2]?.net_lines_added).toBe(0);
            expect(daily[2]?.cumulative_lines_added).toBe(1100);
            expect(daily[2]?.cumulative_lines_deleted).toBe(1400);
            expect(daily[2]?.cumulative_net_lines_added).toBe(-300);
            // Bucket 3 (recovers positive)
            expect(daily[3]?.net_lines_added).toBe(450);
            expect(daily[3]?.cumulative_lines_added).toBe(1600);
            expect(daily[3]?.cumulative_lines_deleted).toBe(1450);
            expect(daily[3]?.cumulative_net_lines_added).toBe(150);
            // Invariant: across all buckets, cumulative_net_lines_added === cumulative_lines_added - cumulative_lines_deleted
            for (const b of daily) {
                expect(b.cumulative_net_lines_added).toBe(b.cumulative_lines_added - b.cumulative_lines_deleted);
                expect(b.net_lines_added).toBe(b.lines_added - b.lines_deleted);
            }
        });
    });
});

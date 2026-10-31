// @vitest-environment node
import { describe, it, expect } from 'vitest';
import { formatTokensK, formatCurrencyUSD, formatPercentage, formatDurationSeconds, formatLeadTime, computeLeadTimeSeconds, deriveTicketComparisonRows, formatIsoDateTimeShort, compareTicketIds, getEffectiveCompletionDateMs, compareTicketCompletionDate, canonicalTicketType, extractTicketPrefix, combineDistinctPrefixes, filterLedgersByTicketTypes, filterLedgersByPrefixes, formatChartAxisTimestamp, formatCompactCount, formatLinesCompact, truncateMiddle, formatLocalTimeWithMs, formatLocalDateTimeTooltip, getClientTimezoneBadge, } from './formatters';
import { mockAgenticMetricsLedgers } from '../test/fixtures/mockData';
import type { AgenticMetricsLedger, ContextHealthStatus, ConventionalTicketType, } from './types';
describe('Domain Pure Formatters', () => {
    describe('truncateMiddle', () => {
        it('keeps the default leading and trailing characters around an ellipsis', () => {
            expect(truncateMiddle('ses_0192a83e-unique-x-304a')).toBe('ses_0192...x-304a');
        });
        it('supports custom leading and trailing character counts', () => {
            expect(truncateMiddle('abcdefghij', 3, 2)).toBe('abc...ij');
        });
        it('leaves strings at or below the retained-character cutoff unchanged', () => {
            expect(truncateMiddle('abcdefghijklmn')).toBe('abcdefghijklmn');
            expect(truncateMiddle('short')).toBe('short');
            expect(truncateMiddle('')).toBe('');
        });
        it('handles zero and invalid boundary counts safely', () => {
            expect(truncateMiddle('abcdef', 0, 0)).toBe('...');
            expect(truncateMiddle('abcdefghij', -1, 2)).toBe('abcdefghij');
            expect(truncateMiddle('abcdefghij', 3, Number.NaN)).toBe('abc...efghij');
        });
    });
    describe('localized timestamps', () => {
        it('formats local clock time with zero padded milliseconds', () => {
            expect(formatLocalTimeWithMs('2026-10-22T14:32:10.007Z')).toMatch(/^\d{2}:\d{2}:\d{2}\.007$/);
        });
        it('preserves invalid values and includes offset context in tooltip', () => {
            expect(formatLocalTimeWithMs('invalid')).toBe('invalid');
            expect(formatLocalDateTimeTooltip('2026-10-22T14:32:10.007Z')).toMatch(/\.007 \[UTC[+-]\d{2}:\d{2}\]$/);
            expect(getClientTimezoneBadge()).toMatch(/^TZ: /);
        });
    });
    describe('formatTokensK', () => {
        it('handles zero and negative gracefully', () => {
            expect(formatTokensK(0)).toBe('0');
            expect(formatTokensK(-10)).toBe('0');
            expect(formatTokensK(-0.001)).toBe('0');
            expect(formatTokensK(-Infinity)).toBe('0');
            expect(formatTokensK(Infinity)).toBe('0');
            expect(formatTokensK(NaN)).toBe('0');
        });
        it('formats numbers under 1,000 as plain integers', () => {
            expect(formatTokensK(450)).toBe('450');
            expect(formatTokensK(999)).toBe('999');
            expect(formatTokensK(999.9)).toBe('999');
            expect(formatTokensK(0.99)).toBe('0');
        });
        it('formats exact and fractional thousands with k suffix', () => {
            expect(formatTokensK(1000)).toBe('1k');
            expect(formatTokensK(200000)).toBe('200k');
            expect(formatTokensK(1500)).toBe('1.5k');
            expect(formatTokensK(900800)).toBe('900.8k');
            expect(formatTokensK(1000.9)).toBe('1k');
            expect(formatTokensK(999400)).toBe('999.4k');
        });
        it('formats exact and fractional millions with M suffix', () => {
            expect(formatTokensK(1000000)).toBe('1M');
            expect(formatTokensK(2000000)).toBe('2M');
            expect(formatTokensK(1200000)).toBe('1.2M');
            expect(formatTokensK(3450000)).toBe('3.5M');
            expect(formatTokensK(1000000000)).toBe('1000M');
            expect(formatTokensK(2500000000)).toBe('2500M');
        });
    });
    describe('formatCurrencyUSD', () => {
        it('handles zero and negative gracefully', () => {
            expect(formatCurrencyUSD(0)).toBe('$0.00');
            expect(formatCurrencyUSD(-5)).toBe('$0.00');
            expect(formatCurrencyUSD(-0.0001)).toBe('$0.00');
            expect(formatCurrencyUSD(NaN)).toBe('$0.00');
            expect(formatCurrencyUSD(Infinity)).toBe('$0.00');
            expect(formatCurrencyUSD(-Infinity)).toBe('$0.00');
        });
        it('formats micro-costs under 1 dollar with 4 decimal places', () => {
            expect(formatCurrencyUSD(0.00001)).toBe('$0.0000');
            expect(formatCurrencyUSD(0.0025)).toBe('$0.0025');
            expect(formatCurrencyUSD(0.852)).toBe('$0.8520');
            expect(formatCurrencyUSD(0.99999)).toBe('$1.0000');
        });
        it('formats amounts equal to or above 1 dollar with 2 decimal places', () => {
            expect(formatCurrencyUSD(1)).toBe('$1.00');
            expect(formatCurrencyUSD(14.821)).toBe('$14.82');
            expect(formatCurrencyUSD(1234.5)).toBe('$1234.50');
            expect(formatCurrencyUSD(1000000.999)).toBe('$1000001.00');
        });
    });
    describe('formatPercentage', () => {
        it('handles zero and negative defensively', () => {
            expect(formatPercentage(0)).toBe('0.0%');
            expect(formatPercentage(-1)).toBe('0.0%');
            expect(formatPercentage(-0.0001)).toBe('0.0%');
            expect(formatPercentage(NaN)).toBe('0.0%');
            expect(formatPercentage(Infinity)).toBe('0.0%');
            expect(formatPercentage(-Infinity)).toBe('0.0%');
        });
        it('formats normal percentages with 1 decimal place', () => {
            expect(formatPercentage(0.04)).toBe('0.0%');
            expect(formatPercentage(0.05)).toBe('0.1%');
            expect(formatPercentage(78.84)).toBe('78.8%');
            expect(formatPercentage(100)).toBe('100.0%');
            expect(formatPercentage(33.333)).toBe('33.3%');
            expect(formatPercentage(99.99)).toBe('100.0%');
        });
    });
    describe('formatDurationSeconds', () => {
        it('handles zero and negative duration defensively', () => {
            expect(formatDurationSeconds(0)).toBe('0m 0s');
            expect(formatDurationSeconds(-120)).toBe('0m 0s');
            expect(formatDurationSeconds(-0.5)).toBe('0m 0s');
            expect(formatDurationSeconds(NaN)).toBe('0m 0s');
            expect(formatDurationSeconds(Infinity)).toBe('0m 0s');
            expect(formatDurationSeconds(-Infinity)).toBe('0m 0s');
        });
        it('formats minutes and seconds when under 1 hour', () => {
            expect(formatDurationSeconds(0.8)).toBe('0m 0s');
            expect(formatDurationSeconds(45)).toBe('0m 45s');
            expect(formatDurationSeconds(125)).toBe('2m 5s');
            expect(formatDurationSeconds(3599)).toBe('59m 59s');
        });
        it('formats hours and minutes when 1 hour or more', () => {
            expect(formatDurationSeconds(3600)).toBe('1h 0m');
            expect(formatDurationSeconds(7420)).toBe('2h 3m');
            expect(formatDurationSeconds(36000)).toBe('10h 0m');
            expect(formatDurationSeconds(86400 * 5)).toBe('120h 0m');
        });
    });
    describe('computeLeadTimeSeconds and formatLeadTime', () => {
        it('returns null and In Progress when completed_at or started_at is missing', () => {
            expect(computeLeadTimeSeconds(null, null)).toBeNull();
            expect(computeLeadTimeSeconds('2026-10-15T10:00:00Z', null)).toBeNull();
            expect(computeLeadTimeSeconds(null, '2026-10-15T10:00:00Z')).toBeNull();
            expect(computeLeadTimeSeconds('', '')).toBeNull();
            expect(formatLeadTime(null)).toBe('In Progress');
            expect(formatLeadTime(NaN)).toBe('In Progress');
            expect(formatLeadTime(Infinity)).toBe('In Progress');
            expect(formatLeadTime(-Infinity)).toBe('In Progress');
        });
        it('returns 0 when completed_at is earlier than started_at or corrupt dates', () => {
            expect(computeLeadTimeSeconds('invalid-date', '2026-10-15T10:00:00Z')).toBeNull();
            expect(computeLeadTimeSeconds('2026-10-15T12:00:00Z', 'invalid-date')).toBeNull();
            expect(computeLeadTimeSeconds('2026-10-15T14:00:00Z', '2026-10-15T10:00:00Z')).toBe(0);
            expect(formatLeadTime(0)).toBe('0s');
            expect(formatLeadTime(-50)).toBe('0s');
        });
        it('computes correct duration and formats appropriately', () => {
            const start = '2026-10-15T10:00:00Z';
            const end = '2026-10-15T14:30:00Z';
            const diff = computeLeadTimeSeconds(start, end);
            expect(diff).toBe(16200);
            expect(formatLeadTime(diff)).toBe('4h 30m');
        });
        it('formats days when lead time exceeds 24 hours', () => {
            const seconds = 86400 * 2 + 3600 * 5;
            expect(formatLeadTime(seconds)).toBe('2d 5h');
        });
        it('formats minutes and seconds when under an hour', () => {
            expect(formatLeadTime(180)).toBe('3m 0s');
            expect(formatLeadTime(45)).toBe('0m 45s');
        });
    });
    describe('deriveTicketComparisonRows boundary sorting and empty handling', () => {
        it('retains distinct sorted recorded skill names in browser rows', () => {
            const base = mockAgenticMetricsLedgers['AL-1']!;
            const skill = (name: string) => ({ name, success: true, caller_role: 'coder', timestamp: '2026-01-01T00:00:00Z', turn_index: 0, message_index: 0 });
            const rows = deriveTicketComparisonRows([{ ...base, facts: { ...base.facts, skills: [skill('beta'), skill('alpha'), skill('beta')] } }]);
            expect(rows[0]?.skills_used).toEqual(['alpha', 'beta']);
        });
        it('returns empty array when input is empty', () => {
            expect(deriveTicketComparisonRows([])).toEqual([]);
        });
        it('falls back to defaults for missing or empty ticket_id and ticket_title', () => {
            const dummyLedger = {
                ticket_id: null,
                ticket_title: null,
                ticket_type: null,
                started_at: null,
                completed_at: null,
                updated_at: '2026-10-15T12:00:00Z',
                facts: {
                    ticket_id: null,
                    ticket_title: null,
                    ticket_type: null,
                    started_at: null,
                    completed_at: null,
                    updated_at: '2026-10-15T12:00:00Z',
                    code: {
                        lines_added: 0,
                        lines_deleted: 0,
                        test_lines_added: 0,
                        test_lines_deleted: 0,
                        prod_lines_added: 0,
                        prod_lines_deleted: 0,
                        files_changed: 0,
                        sdlc_iterations: 0,
                    },
                    sessions: []
                },
                workflow_efficiency: {
                    sdlc_iterations: 0,
                    active_duration_seconds: 0,
                    active_duration_formatted: '0m 0s',
                    role_effort_distribution_pct: {},
                    subagent_turn_velocity: {
                        avg_turns_per_invocation: 0,
                        peak_turns: 0,
                        by_role: {},
                    },
                    test_to_production_ratio: {
                        ratio: '0.0:1',
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
                        status: 'HEALTHY' as ContextHealthStatus,
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
                    invocations: 0,
                    input_tokens: 0,
                    output_tokens: 0,
                    cache_read_tokens: 0,
                    cache_hit_rate_pct: 0,
                    cost_usd: 0,
                    peak_context_tokens: 0,
                    context_health: 'HEALTHY' as ContextHealthStatus,
                },
                roles: [],
                models: [],
            };
            const rows = deriveTicketComparisonRows([dummyLedger]);
            expect(rows).toHaveLength(1);
            expect(rows[0]?.ticket_id).toBe('unknown');
            expect(rows[0]?.ticket_title).toBe('N/A');
            expect(rows[0]?.lead_duration_formatted).toBe('In Progress');
            expect(rows[0]?.skills_used).toEqual([]);
        });
        it('handles migrated and partially migrated ledger objects cleanly', () => {
            const migratedLedger = {
                ticket_id: 'AL-12',
                ticket_title: 'Telemetry Backfill Engine',
                ticket_type: 'feat' as ConventionalTicketType,
                started_at: '2026-10-17T10:00:00Z',
                completed_at: '2026-10-17T11:00:00Z',
                updated_at: '2026-10-17T11:00:00Z',
                facts: {
                    schema_version: 2,
                    migration: {
                        schema_version: 2,
                        status: 'partial' as const,
                        steps_completed: ['git_code_facts'],
                        steps_pending: ['opencode_db'],
                        last_migrated_at: '2026-10-17T11:00:00Z',
                        last_migrated_by: 'author-1',
                    },
                    ticket_id: 'AL-12',
                    ticket_title: 'Telemetry Backfill Engine',
                    ticket_type: 'feat' as ConventionalTicketType,
                    started_at: '2026-10-17T10:00:00Z',
                    completed_at: '2026-10-17T11:00:00Z',
                    updated_at: '2026-10-17T11:00:00Z',
                    code: {
                        lines_added: 120,
                        lines_deleted: 10,
                        test_lines_added: 60,
                        test_lines_deleted: 5,
                        prod_lines_added: 60,
                        prod_lines_deleted: 5,
                        files_changed: 3,
                        sdlc_iterations: 1,
                    },
                    sessions: []
                },
                workflow_efficiency: {
                    sdlc_iterations: 1,
                    active_duration_seconds: 60,
                    active_duration_formatted: '1m 0s',
                    role_effort_distribution_pct: {},
                    subagent_turn_velocity: {
                        avg_turns_per_invocation: 0,
                        peak_turns: 0,
                        by_role: {},
                    },
                    test_to_production_ratio: {
                        ratio: '1.0:1',
                        test_lines_added: 60,
                        prod_lines_added: 60,
                    },
                    reasoning_tokens: 0,
                    reasoning_ratio_pct: 0,
                    code_delivered: {
                        lines_added: 120,
                        lines_deleted: 10,
                        files_changed: 3,
                    },
                    tool_execution: {
                        total_calls: 0,
                        errors_and_corrections: 0,
                        success_rate_pct: 100,
                    },
                    context_window_health: {
                        healthy_ceiling_tokens: 200000,
                        status: 'HEALTHY' as ContextHealthStatus,
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
                    invocations: 0,
                    input_tokens: 0,
                    output_tokens: 0,
                    cache_read_tokens: 0,
                    cache_hit_rate_pct: 0,
                    cost_usd: 0,
                    peak_context_tokens: 0,
                    context_health: 'HEALTHY' as ContextHealthStatus,
                },
                roles: [],
                models: [],
            };
            const rows = deriveTicketComparisonRows([migratedLedger]);
            expect(rows).toHaveLength(1);
            expect(rows[0]?.ticket_id).toBe('AL-12');
            expect(rows[0]?.lead_duration_formatted).toBe('1h 0m');
            expect(migratedLedger.facts.migration?.status).toBe('partial');
            expect(migratedLedger.facts.migration?.steps_pending).toEqual(['opencode_db']);
        });
        it('sorts rows dynamically by comparison sort fields and order', () => {
            const ledgers = Object.values(mockAgenticMetricsLedgers);
            expect(ledgers.length).toBeGreaterThan(1);
            // Sort by cost ascending
            const byCostAsc = deriveTicketComparisonRows(ledgers, 'cost_usd', 'asc');
            for (let i = 0; i < byCostAsc.length - 1; i++) {
                expect(byCostAsc[i]!.cost_usd).toBeLessThanOrEqual(byCostAsc[i + 1]!.cost_usd);
            }
            // Sort by cost descending
            const byCostDesc = deriveTicketComparisonRows(ledgers, 'cost_usd', 'desc');
            for (let i = 0; i < byCostDesc.length - 1; i++) {
                expect(byCostDesc[i]!.cost_usd).toBeGreaterThanOrEqual(byCostDesc[i + 1]!.cost_usd);
            }
            // Sort by active duration ascending
            const byDurAsc = deriveTicketComparisonRows(ledgers, 'active_duration_seconds', 'asc');
            for (let i = 0; i < byDurAsc.length - 1; i++) {
                expect(byDurAsc[i]!.active_duration_seconds).toBeLessThanOrEqual(byDurAsc[i + 1]!.active_duration_seconds);
            }
            // Sort by iterations descending
            const byIterDesc = deriveTicketComparisonRows(ledgers, 'sdlc_iterations', 'desc');
            for (let i = 0; i < byIterDesc.length - 1; i++) {
                expect(byIterDesc[i]!.sdlc_iterations).toBeGreaterThanOrEqual(byIterDesc[i + 1]!.sdlc_iterations);
            }
            // Sort by lines added descending
            const byLinesDesc = deriveTicketComparisonRows(ledgers, 'lines_added', 'desc');
            for (let i = 0; i < byLinesDesc.length - 1; i++) {
                expect(byLinesDesc[i]!.lines_added).toBeGreaterThanOrEqual(byLinesDesc[i + 1]!.lines_added);
            }
            // Sort by ticket_id ascending
            const byIdAsc = deriveTicketComparisonRows(ledgers, 'ticket_id', 'asc');
            for (let i = 0; i < byIdAsc.length - 1; i++) {
                expect(compareTicketIds(byIdAsc[i]!.ticket_id, byIdAsc[i + 1]!.ticket_id)).toBeLessThanOrEqual(0);
            }
        });
        it('naturally orders comparison rows by ticket_id and tie-breaks equal metric values', () => {
            const baseLedger = Object.values(mockAgenticMetricsLedgers)[0]!;
            const testLedgers = [
                { ...baseLedger, ticket_id: 'AL-10', totals: { ...baseLedger.totals, cost_usd: 10.0 } },
                { ...baseLedger, ticket_id: 'AL-2', totals: { ...baseLedger.totals, cost_usd: 10.0 } },
                { ...baseLedger, ticket_id: 'AL-1', totals: { ...baseLedger.totals, cost_usd: 10.0 } },
            ];
            // Explicit ticket_id asc
            const rowsAsc = deriveTicketComparisonRows(testLedgers, 'ticket_id', 'asc');
            expect(rowsAsc.map((r) => r.ticket_id)).toEqual(['AL-1', 'AL-2', 'AL-10']);
            // Explicit ticket_id desc
            const rowsDesc = deriveTicketComparisonRows(testLedgers, 'ticket_id', 'desc');
            expect(rowsDesc.map((r) => r.ticket_id)).toEqual(['AL-10', 'AL-2', 'AL-1']);
            // Equal cost tie-breaking (asc): natural ticket ID comparison
            const rowsCostAsc = deriveTicketComparisonRows(testLedgers, 'cost_usd', 'asc');
            expect(rowsCostAsc.map((r) => r.ticket_id)).toEqual(['AL-1', 'AL-2', 'AL-10']);
        });
    });
    describe('formatIsoDateTimeShort', () => {
        it('formats short datetime correctly', () => {
            expect(formatIsoDateTimeShort('2026-10-15T14:30:00Z')).toBe('2026-10-15 14:30');
            expect(formatIsoDateTimeShort('2026-10-15')).toBe('2026-10-15');
            expect(formatIsoDateTimeShort(null)).toBe('N/A');
            expect(formatIsoDateTimeShort(undefined)).toBe('N/A');
            expect(formatIsoDateTimeShort('short')).toBe('N/A');
        });
    });
    describe('compareTicketIds', () => {
        it('sorts naturally so AL-2 precedes AL-10', () => {
            expect(compareTicketIds('AL-2', 'AL-10')).toBeLessThan(0);
            expect(compareTicketIds('AL-10', 'AL-2')).toBeGreaterThan(0);
            expect(compareTicketIds('AL-1', 'AL-2')).toBeLessThan(0);
        });
        it('sorts identically formatted tickets correctly', () => {
            expect(compareTicketIds('AL-1', 'AL-1')).toBe(0);
            expect(compareTicketIds('AL-1', 'al-1')).toBeLessThan(0);
            expect(compareTicketIds('al-1', 'AL-1')).toBeGreaterThan(0);
        });
        it('orders by prefix first', () => {
            expect(compareTicketIds('AL-99', 'BL-1')).toBeLessThan(0);
            expect(compareTicketIds('PO-1', 'AL-10')).toBeGreaterThan(0);
        });
        it('supports delimiter variations', () => {
            expect(compareTicketIds('AL_2', 'AL-10')).toBeLessThan(0);
            expect(compareTicketIds('AL2', 'AL10')).toBeLessThan(0);
        });
        it('places conforming tickets before non-conforming strings', () => {
            expect(compareTicketIds('AL-1', 'unknown')).toBeLessThan(0);
            expect(compareTicketIds('unknown', 'AL-1')).toBeGreaterThan(0);
        });
        it('handles empty strings and whitespace defensively', () => {
            expect(compareTicketIds('', 'AL-1')).toBeLessThan(0);
            expect(compareTicketIds('AL-1', '')).toBeGreaterThan(0);
            expect(compareTicketIds('', '')).toBe(0);
            expect(compareTicketIds('   ', '')).toBe(0);
            expect(compareTicketIds('', '   ')).toBe(0);
            expect(compareTicketIds('   ', 'AL-1')).toBeLessThan(0);
            expect(compareTicketIds('AL-1', '   ')).toBeGreaterThan(0);
            expect(compareTicketIds('\tAL-1\n', 'AL-2')).toBeLessThan(0);
            expect(compareTicketIds('  AL-2  ', 'AL-10')).toBeLessThan(0);
        });
        it('handles missing prefixes, missing numbers, and purely numeric strings', () => {
            // Purely numeric vs conforming
            expect(compareTicketIds('AL-1', '123')).toBeLessThan(0);
            expect(compareTicketIds('123', 'AL-1')).toBeGreaterThan(0);
            // Purely numeric fallback to lexicographical
            expect(compareTicketIds('123', '456')).toBeLessThan(0);
            expect(compareTicketIds('456', '123')).toBeGreaterThan(0);
            expect(compareTicketIds('100', '20')).toBeLessThan(0);
            // Missing numbers (conforming sorts before non-conforming)
            expect(compareTicketIds('AL-', 'AL-1')).toBeGreaterThan(0);
            expect(compareTicketIds('AL', 'AL-1')).toBeGreaterThan(0);
            expect(compareTicketIds('AL-', 'GEN-')).toBeLessThan(0);
        });
        it('handles leading zeroes variations correctly', () => {
            expect(compareTicketIds('AL-01', 'AL-1')).toBeLessThan(0);
            expect(compareTicketIds('AL-1', 'AL-01')).toBeGreaterThan(0);
            expect(compareTicketIds('AL-001', 'AL-01')).toBeLessThan(0);
            expect(compareTicketIds('AL-02', 'AL-10')).toBeLessThan(0);
            expect(compareTicketIds('AL-009', 'AL-1')).toBeGreaterThan(0);
            expect(compareTicketIds('AL-0', 'AL-00')).toBeLessThan(0);
            expect(compareTicketIds('AL-00', 'AL-0')).toBeGreaterThan(0);
        });
        it('handles mixed case and suffixes', () => {
            expect(compareTicketIds('al-10', 'AL-2')).toBeGreaterThan(0);
            expect(compareTicketIds('AL-2', 'al-10')).toBeLessThan(0);
            expect(compareTicketIds('Al-100', 'aL-99')).toBeGreaterThan(0);
            expect(compareTicketIds('AL-1-a', 'AL-1-b')).toBeLessThan(0);
            expect(compareTicketIds('AL-1-b', 'AL-1-a')).toBeGreaterThan(0);
            expect(compareTicketIds('AL-1-hotfix', 'AL-2-hotfix')).toBeLessThan(0);
            expect(compareTicketIds('AL-2-beta', 'AL-10-alpha')).toBeLessThan(0);
            expect(compareTicketIds('AL-1', 'AL-1-hotfix')).toBeLessThan(0);
        });
        it('handles delimiter variations between hyphen and underscore', () => {
            expect(compareTicketIds('AL-10', 'AL_2')).toBeGreaterThan(0);
            expect(compareTicketIds('AL_2', 'AL-10')).toBeLessThan(0);
            expect(compareTicketIds('AL-1', 'AL_1')).toBeLessThan(0);
            expect(compareTicketIds('AL_1', 'AL-1')).toBeGreaterThan(0);
        });
        it('correctly sorts an entire array of mixed ticket IDs', () => {
            const input = [
                'AL-10',
                '123',
                'AL-01',
                'al-2',
                'AL-1',
                'AL_2',
                'AL-',
                '',
                'GEN-1',
                'unknown',
                '456',
            ];
            const sorted = [...input].sort(compareTicketIds);
            expect(sorted).toEqual([
                '',
                'AL-01',
                'AL-1',
                'AL_2',
                'al-2',
                'AL-10',
                'GEN-1',
                '123',
                '456',
                'AL-',
                'unknown',
            ]);
        });
    });
    describe('Completion Date Parsing, Comparison, and Filtering', () => {
        it('resolves effective completion date adhering strictly to completed_at with zero fallback', () => {
            // completed_at present
            expect(getEffectiveCompletionDateMs({
                completed_at: '2026-03-10T12:00:00Z',
                updated_at: '2026-03-09T12:00:00Z',
                started_at: '2026-03-08T12:00:00Z',
            })).toBe(new Date('2026-03-10T12:00:00Z').getTime());
            // returns null when completed_at is null, even if updated_at is present
            expect(getEffectiveCompletionDateMs({
                completed_at: null,
                updated_at: '2026-03-09T12:00:00Z',
                started_at: '2026-03-08T12:00:00Z',
            })).toBeNull();
            // returns null when completed_at is undefined, even if started_at is present
            expect(getEffectiveCompletionDateMs({
                completed_at: undefined,
                updated_at: undefined,
                started_at: '2026-03-08T12:00:00Z',
            })).toBeNull();
            // invalid date strings return null
            expect(getEffectiveCompletionDateMs({
                completed_at: 'not-a-date',
                updated_at: '2026-03-09T12:00:00Z',
                started_at: '2026-03-08T12:00:00Z',
            })).toBeNull();
            // empty or whitespace completed_at returns null
            expect(getEffectiveCompletionDateMs({
                completed_at: '',
                updated_at: '2026-03-09T12:00:00Z',
                started_at: '2026-03-08T12:00:00Z',
            })).toBeNull();
            expect(getEffectiveCompletionDateMs({
                completed_at: '   ',
                updated_at: '2026-03-09T12:00:00Z',
                started_at: '2026-03-08T12:00:00Z',
            })).toBeNull();
        });
        it('compares tickets by completion date descending with nulls positioned last and tie-breaking by ticket ID', () => {
            const t1 = { ticket_id: 'AL-1', completed_at: '2026-03-05T00:00:00Z' };
            const t2 = { ticket_id: 'AL-2', completed_at: '2026-03-15T00:00:00Z' };
            const t3 = { ticket_id: 'AL-3', completed_at: '2026-03-15T00:00:00Z' };
            const t4 = { ticket_id: 'AL-4', completed_at: null };
            const t5 = { ticket_id: 'AL-5', completed_at: undefined };
            // t2 vs t1: t2 is later so t2 comes before t1 in descending (compare(t2, t1) < 0)
            expect(compareTicketCompletionDate(t2, t1)).toBeLessThan(0);
            expect(compareTicketCompletionDate(t1, t2)).toBeGreaterThan(0);
            // t2 vs t3: same date, ties broken by ticket_id (AL-2 before AL-3)
            expect(compareTicketCompletionDate(t2, t3)).toBeLessThan(0);
            expect(compareTicketCompletionDate(t3, t2)).toBeGreaterThan(0);
            // non-null vs null: non-null comes first
            expect(compareTicketCompletionDate(t1, t4)).toBeLessThan(0);
            expect(compareTicketCompletionDate(t4, t1)).toBeGreaterThan(0);
            // null vs null: tie-broken by ticket_id
            expect(compareTicketCompletionDate(t4, t5)).toBeLessThan(0);
            const items = [t1, t4, t3, t5, t2];
            const sorted = [...items].sort(compareTicketCompletionDate);
            expect(sorted.map((t) => t.ticket_id)).toEqual(['AL-2', 'AL-3', 'AL-1', 'AL-4', 'AL-5']);
        });
        it('handles timezone offsets, leap years, and natural ticket ID tie breaking', () => {
            // Timezone offsets resolution
            const jstMs = getEffectiveCompletionDateMs({ completed_at: '2026-03-01T15:00:00+09:00' });
            const edtMs = getEffectiveCompletionDateMs({ completed_at: '2026-03-01T02:00:00-04:00' });
            const utcMs = getEffectiveCompletionDateMs({ completed_at: '2026-03-01T06:00:00Z' });
            expect(jstMs).toBe(utcMs);
            expect(edtMs).toBe(utcMs);
            // Leap year 2024-02-29
            const leapMs = getEffectiveCompletionDateMs({ completed_at: '2024-02-29T12:00:00Z' });
            expect(leapMs).not.toBeNull();
            expect(new Date(leapMs!).getUTCFullYear()).toBe(2024);
            expect(new Date(leapMs!).getUTCMonth()).toBe(1); // February is 1
            expect(new Date(leapMs!).getUTCDate()).toBe(29);
            // Tie-breaking identical dates with natural ticket ID (AL-2 before AL-10)
            const cand10 = { ticket_id: 'AL-10', completed_at: '2026-03-10T00:00:00Z' };
            const cand2 = { ticket_id: 'AL-2', completed_at: '2026-03-10T00:00:00Z' };
            expect(compareTicketCompletionDate(cand2, cand10)).toBeLessThan(0);
            expect(compareTicketCompletionDate(cand10, cand2)).toBeGreaterThan(0);
            // Tie-breaking both null dates with natural ticket ID
            const null10 = { ticket_id: 'AL-10', completed_at: null };
            const null2 = { ticket_id: 'AL-2', completed_at: null };
            expect(compareTicketCompletionDate(null2, null10)).toBeLessThan(0);
            expect(compareTicketCompletionDate(null10, null2)).toBeGreaterThan(0);
        });
    });
    describe('canonicalTicketType', () => {
        it('normalizes casing, whitespace, and conventional aliases', () => {
            expect(canonicalTicketType('feat')).toBe('feature');
            expect(canonicalTicketType('FEAT')).toBe('feature');
            expect(canonicalTicketType('feature')).toBe('feature');
            expect(canonicalTicketType('  feature  ')).toBe('feature');
            expect(canonicalTicketType('bug')).toBe('bug');
            expect(canonicalTicketType('BUG')).toBe('bug');
            expect(canonicalTicketType('fix')).toBe('bug');
            expect(canonicalTicketType('bugfix')).toBe('bug');
            expect(canonicalTicketType('doc')).toBe('docs');
            expect(canonicalTicketType('docs')).toBe('docs');
            expect(canonicalTicketType('chore')).toBe('chore');
            expect(canonicalTicketType('refactor')).toBe('refactor');
            expect(canonicalTicketType('ci')).toBe('ci');
            expect(canonicalTicketType('test')).toBe('test');
            expect(canonicalTicketType('perf')).toBe('perf');
            expect(canonicalTicketType('build')).toBe('build');
            expect(canonicalTicketType('custom')).toBe('custom');
            expect(canonicalTicketType('')).toBe('');
        });
    });
    describe('extractTicketPrefix', () => {
        it('extracts uppercase alphabetic prefix before delimiter or digits', () => {
            expect(extractTicketPrefix('AL-1')).toBe('AL');
            expect(extractTicketPrefix('al-28')).toBe('AL');
            expect(extractTicketPrefix('GEN_4')).toBe('GEN');
            expect(extractTicketPrefix('po-2')).toBe('PO');
            expect(extractTicketPrefix('al10')).toBe('AL');
            expect(extractTicketPrefix('AL-')).toBe('AL');
            expect(extractTicketPrefix('ticket-xyz')).toBe('TICKET');
            expect(extractTicketPrefix('123-abc')).toBe('');
            expect(extractTicketPrefix('123')).toBe('');
            expect(extractTicketPrefix('UNKNOWN')).toBe('');
            expect(extractTicketPrefix('AL.1')).toBe('');
            expect(extractTicketPrefix('')).toBe('');
        });
    });
    describe('combineDistinctPrefixes', () => {
        it('returns canonical fallback when both discovered and configured are empty', () => {
            expect(combineDistinctPrefixes([])).toEqual(['AL', 'GEN', 'PO']);
            expect(combineDistinctPrefixes([], [])).toEqual(['AL', 'GEN', 'PO']);
            expect(combineDistinctPrefixes(['', '   '], ['  '])).toEqual(['AL', 'GEN', 'PO']);
        });
        it('returns canonical fallback when all tokens resolve to empty after trimming hyphens and underscores', () => {
            expect(combineDistinctPrefixes(['---', '___', ' - _ '])).toEqual(['AL', 'GEN', 'PO']);
            expect(combineDistinctPrefixes(['---'], ['___', '   '])).toEqual(['AL', 'GEN', 'PO']);
        });
        it('combines configured prefixes with empty discovered prefixes', () => {
            expect(combineDistinctPrefixes([], ['FEAT', 'BUG'])).toEqual(['BUG', 'FEAT']);
            expect(combineDistinctPrefixes([], ['CUSTOM'])).toEqual(['CUSTOM']);
        });
        it('merges overlapping configured and discovered prefixes without duplicates', () => {
            const discovered = ['AL', 'GEN', 'CUSTOM'];
            const configured = ['GEN', 'PO', 'CUSTOM', 'INFRA'];
            expect(combineDistinctPrefixes(discovered, configured)).toEqual([
                'AL',
                'CUSTOM',
                'GEN',
                'INFRA',
                'PO',
            ]);
        });
        it('handles case differences, whitespace trimming, trailing hyphens, and alphabetical sort', () => {
            const discovered = ['  al  ', 'feat-', 'infra_'];
            const configured = ['AL', 'FEAT', 'bug'];
            expect(combineDistinctPrefixes(discovered, configured)).toEqual(['AL', 'BUG', 'FEAT', 'INFRA']);
        });
        it('preserves interior hyphens and numbers while deduplicating hostile variations', () => {
            const discovered = ['  proj-x--  ', 'CORE-1_'];
            const configured = ['proj-x', 'CORE-1', 'api-v2'];
            expect(combineDistinctPrefixes(discovered, configured)).toEqual(['API-V2', 'CORE-1', 'PROJ-X']);
        });
    });
    describe('filterLedgersByTicketTypes', () => {
        const baseLedger = Object.values(mockAgenticMetricsLedgers)[0]!;
        const featLedger: AgenticMetricsLedger = { ...baseLedger, ticket_id: 'AL-1', ticket_type: 'feature' as ConventionalTicketType };
        const aliasFeatLedger: AgenticMetricsLedger = { ...baseLedger, ticket_id: 'AL-2', ticket_type: 'feat' };
        const bugLedger: AgenticMetricsLedger = { ...baseLedger, ticket_id: 'AL-3', ticket_type: 'bug' as ConventionalTicketType };
        const choreLedger: AgenticMetricsLedger = { ...baseLedger, ticket_id: 'AL-4', ticket_type: 'chore' };
        const nullTypeLedger: AgenticMetricsLedger = { ...baseLedger, ticket_id: 'AL-5', ticket_type: null };
        const all: readonly AgenticMetricsLedger[] = [featLedger, aliasFeatLedger, bugLedger, choreLedger, nullTypeLedger];
        it('returns all ledgers if types is empty or contains all', () => {
            expect(filterLedgersByTicketTypes(all, [])).toEqual(all);
            expect(filterLedgersByTicketTypes(all, ['all'])).toEqual(all);
            expect(filterLedgersByTicketTypes(all, ['ALL'])).toEqual(all);
            expect(filterLedgersByTicketTypes(all, ['feature', 'all'])).toEqual(all);
        });
        it('filters ledgers matching canonical types and aliases', () => {
            const res = filterLedgersByTicketTypes(all, ['feature']);
            expect(res.map((l) => l.ticket_id)).toEqual(['AL-1', 'AL-2']);
            const resMulti = filterLedgersByTicketTypes(all, ['feat', 'bug']);
            expect(resMulti.map((l) => l.ticket_id)).toEqual(['AL-1', 'AL-2', 'AL-3']);
        });
    });
    describe('filterLedgersByPrefixes', () => {
        const baseLedger = Object.values(mockAgenticMetricsLedgers)[0]!;
        const al1 = { ...baseLedger, ticket_id: 'AL-1' };
        const al2 = { ...baseLedger, ticket_id: 'al-2' };
        const po1 = { ...baseLedger, ticket_id: 'PO-1' };
        const gen1 = { ...baseLedger, ticket_id: 'GEN_1' };
        const nullId = { ...baseLedger, ticket_id: null };
        const all = [al1, al2, po1, gen1, nullId];
        it('returns all ledgers if prefixes is empty or contains all', () => {
            expect(filterLedgersByPrefixes(all, [])).toEqual(all);
            expect(filterLedgersByPrefixes(all, ['all'])).toEqual(all);
            expect(filterLedgersByPrefixes(all, ['AL', 'all'])).toEqual(all);
        });
        it('filters ledgers by prefix case-insensitively', () => {
            const res = filterLedgersByPrefixes(all, ['al']);
            expect(res.map((l) => l.ticket_id)).toEqual(['AL-1', 'al-2']);
            const resMulti = filterLedgersByPrefixes(all, ['AL', 'PO']);
            expect(resMulti.map((l) => l.ticket_id)).toEqual(['AL-1', 'al-2', 'PO-1']);
        });
    });
    describe('Adversarial Multi-Dimensional Filtering Edge Cases', () => {
        it('handles empty arrays, whitespace tokens, and disjoint queries', () => {
            const base = Object.values(mockAgenticMetricsLedgers)[0]!;
            const ledgers: readonly AgenticMetricsLedger[] = [
                { ...base, ticket_id: 'AL-1', ticket_type: 'feat' },
                { ...base, ticket_id: 'PO-1', ticket_type: 'docs' },
            ];
            // Empty input ledgers
            expect(filterLedgersByTicketTypes([], ['feature'])).toEqual([]);
            expect(filterLedgersByPrefixes([], ['AL'])).toEqual([]);
            // Whitespace only in filter arrays behaves like 'all'
            expect(filterLedgersByTicketTypes(ledgers, ['   ', ''])).toEqual(ledgers);
            expect(filterLedgersByPrefixes(ledgers, ['   ', ''])).toEqual(ledgers);
            // Disjoint queries
            expect(filterLedgersByTicketTypes(ledgers, ['nonexistent'])).toEqual([]);
            expect(filterLedgersByPrefixes(ledgers, ['NONEXISTENT'])).toEqual([]);
        });
        it('verifies symmetrical alias matching across all conventional pairs', () => {
            const base = Object.values(mockAgenticMetricsLedgers)[0]!;
            const aliasLedgers: readonly AgenticMetricsLedger[] = [
                { ...base, ticket_id: 'T-1', ticket_type: 'feat' },
                { ...base, ticket_id: 'T-2', ticket_type: 'feature' as ConventionalTicketType },
                { ...base, ticket_id: 'T-3', ticket_type: 'bug' as ConventionalTicketType },
                { ...base, ticket_id: 'T-4', ticket_type: 'fix' },
                { ...base, ticket_id: 'T-5', ticket_type: 'bugfix' as ConventionalTicketType },
                { ...base, ticket_id: 'T-6', ticket_type: 'doc' as ConventionalTicketType },
                { ...base, ticket_id: 'T-7', ticket_type: 'docs' },
            ];
            // Querying with alias matches canonical and all other aliases
            expect(filterLedgersByTicketTypes(aliasLedgers, ['feat']).map((l) => l.ticket_id)).toEqual(['T-1', 'T-2']);
            expect(filterLedgersByTicketTypes(aliasLedgers, ['feature']).map((l) => l.ticket_id)).toEqual(['T-1', 'T-2']);
            expect(filterLedgersByTicketTypes(aliasLedgers, ['fix']).map((l) => l.ticket_id)).toEqual(['T-3', 'T-4', 'T-5']);
            expect(filterLedgersByTicketTypes(aliasLedgers, ['bug']).map((l) => l.ticket_id)).toEqual(['T-3', 'T-4', 'T-5']);
            expect(filterLedgersByTicketTypes(aliasLedgers, ['bugfix']).map((l) => l.ticket_id)).toEqual(['T-3', 'T-4', 'T-5']);
            expect(filterLedgersByTicketTypes(aliasLedgers, ['doc']).map((l) => l.ticket_id)).toEqual(['T-6', 'T-7']);
            expect(filterLedgersByTicketTypes(aliasLedgers, ['docs']).map((l) => l.ticket_id)).toEqual(['T-6', 'T-7']);
        });
    });
    describe('formatChartAxisTimestamp', () => {
        it('formats hourly timestamp as HH:00', () => {
            expect(formatChartAxisTimestamp('2026-10-19T14:00:00Z', 'hourly')).toBe('14:00');
            expect(formatChartAxisTimestamp('2026-10-19T09:00:00Z', 'hourly')).toBe('09:00');
            expect(formatChartAxisTimestamp('2026-10-19T00:00:00Z', 'hourly')).toBe('00:00');
        });
        it('formats daily and weekly timestamps as YYYY-MM-DD', () => {
            expect(formatChartAxisTimestamp('2026-10-19', 'daily')).toBe('2026-10-19');
            expect(formatChartAxisTimestamp('2026-10-19T14:00:00Z', 'daily')).toBe('2026-10-19');
            expect(formatChartAxisTimestamp('2026-10-19', 'weekly')).toBe('2026-10-19');
        });
        it('handles empty or malformed timestamp defensively', () => {
            expect(formatChartAxisTimestamp('', 'hourly')).toBe('');
            expect(formatChartAxisTimestamp('unknown', 'hourly')).toBe('unknown');
            expect(formatChartAxisTimestamp('abc', 'daily')).toBe('abc');
            expect(formatChartAxisTimestamp('2026-10-19T18:45', 'hourly')).toBe('18:00');
            expect(formatChartAxisTimestamp('2026-10-19T23:59:59Z', 'hourly')).toBe('23:00');
            expect(formatChartAxisTimestamp('2026-10-19T04:15:00+03:00', 'hourly')).toBe('04:00');
        });
    });
    describe('formatCompactCount', () => {
        it('formats small numbers without suffix', () => {
            expect(formatCompactCount(0)).toBe('0');
            expect(formatCompactCount(1)).toBe('1');
            expect(formatCompactCount(42)).toBe('42');
            expect(formatCompactCount(999)).toBe('999');
        });
        it('formats thousands with k suffix', () => {
            expect(formatCompactCount(1000)).toBe('1k');
            expect(formatCompactCount(1500)).toBe('1.5k');
            expect(formatCompactCount(10000)).toBe('10k');
            expect(formatCompactCount(30000)).toBe('30k');
            expect(formatCompactCount(30500)).toBe('30.5k');
            expect(formatCompactCount(999400)).toBe('999.4k');
        });
        it('formats millions with M suffix', () => {
            expect(formatCompactCount(1000000)).toBe('1M');
            expect(formatCompactCount(1500000)).toBe('1.5M');
            expect(formatCompactCount(25000000)).toBe('25M');
        });
        it('handles negative numbers', () => {
            expect(formatCompactCount(-42)).toBe('-42');
            expect(formatCompactCount(-1000)).toBe('-1k');
            expect(formatCompactCount(-1500)).toBe('-1.5k');
            expect(formatCompactCount(-1000000)).toBe('-1M');
        });
        it('handles explicitPlus flag', () => {
            expect(formatCompactCount(0, { explicitPlus: true })).toBe('+0');
            expect(formatCompactCount(42, { explicitPlus: true })).toBe('+42');
            expect(formatCompactCount(1500, { explicitPlus: true })).toBe('+1.5k');
            expect(formatCompactCount(-1500, { explicitPlus: true })).toBe('-1.5k');
        });
        it('handles non-finite numbers defensively', () => {
            expect(formatCompactCount(NaN)).toBe('0');
            expect(formatCompactCount(Infinity)).toBe('0');
            expect(formatCompactCount(-Infinity)).toBe('0');
        });
    });
    describe('formatLinesCompact', () => {
        it('formats discrete line counts compactly', () => {
            expect(formatLinesCompact(0)).toBe('0');
            expect(formatLinesCompact(250)).toBe('250');
            expect(formatLinesCompact(1200)).toBe('1.2k');
            expect(formatLinesCompact(50000)).toBe('50k');
            expect(formatLinesCompact(-450)).toBe('-450');
            expect(formatLinesCompact(-3500)).toBe('-3.5k');
        });
        it('supports explicitPlus option', () => {
            expect(formatLinesCompact(0, { explicitPlus: true })).toBe('+0');
            expect(formatLinesCompact(450, { explicitPlus: true })).toBe('+450');
            expect(formatLinesCompact(-450, { explicitPlus: true })).toBe('-450');
        });
        it('handles non-finite values defensively', () => {
            expect(formatLinesCompact(NaN)).toBe('0');
            expect(formatLinesCompact(Infinity)).toBe('0');
        });
    });
});

import { describe, it, expect } from 'vitest';
import {
  formatTokensK,
  formatCurrencyUSD,
  formatPercentage,
  formatDurationSeconds,
  formatLeadTime,
  computeLeadTimeSeconds,
  deriveTicketComparisonRows,
  formatIsoDateOnly,
  formatIsoDateTimeShort,
} from './formatters';
import { mockAgenticMetricsLedgers } from '../api/mockData';

describe('Domain Pure Formatters', () => {
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
          invocations: [],
          tools: [],
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
            status: 'HEALTHY',
            orchestrator_peak_tokens: 0,
            orchestrator_saturation_pct: 0,
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
          context_health: 'HEALTHY',
        },
        roles: [],
        models: [],
      };

      const rows = deriveTicketComparisonRows([dummyLedger]);
      expect(rows).toHaveLength(1);
      expect(rows[0]?.ticket_id).toBe('unknown');
      expect(rows[0]?.ticket_title).toBe('N/A');
      expect(rows[0]?.lead_duration_formatted).toBe('In Progress');
    });

    it('handles migrated and partially migrated ledger objects cleanly', () => {
      const migratedLedger = {
        ticket_id: 'AL-12',
        ticket_title: 'Telemetry Backfill Engine',
        ticket_type: 'feature',
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
          ticket_type: 'feature',
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
          invocations: [],
          tools: [],
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
            status: 'HEALTHY',
            orchestrator_peak_tokens: 0,
            orchestrator_saturation_pct: 0,
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
          context_health: 'HEALTHY',
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
        expect(byDurAsc[i]!.active_duration_seconds).toBeLessThanOrEqual(
          byDurAsc[i + 1]!.active_duration_seconds
        );
      }

      // Sort by iterations descending
      const byIterDesc = deriveTicketComparisonRows(ledgers, 'sdlc_iterations', 'desc');
      for (let i = 0; i < byIterDesc.length - 1; i++) {
        expect(byIterDesc[i]!.sdlc_iterations).toBeGreaterThanOrEqual(
          byIterDesc[i + 1]!.sdlc_iterations
        );
      }

      // Sort by lines added descending
      const byLinesDesc = deriveTicketComparisonRows(ledgers, 'lines_added', 'desc');
      for (let i = 0; i < byLinesDesc.length - 1; i++) {
        expect(byLinesDesc[i]!.lines_added).toBeGreaterThanOrEqual(
          byLinesDesc[i + 1]!.lines_added
        );
      }

      // Sort by ticket_id ascending
      const byIdAsc = deriveTicketComparisonRows(ledgers, 'ticket_id', 'asc');
      for (let i = 0; i < byIdAsc.length - 1; i++) {
        expect(byIdAsc[i]!.ticket_id.localeCompare(byIdAsc[i + 1]!.ticket_id)).toBeLessThanOrEqual(0);
      }
    });
  });

  describe('formatIsoDateOnly and formatIsoDateTimeShort', () => {
    it('formats date only correctly', () => {
      expect(formatIsoDateOnly('2026-10-15T14:30:00Z')).toBe('2026-10-15');
      expect(formatIsoDateOnly('2026-10-15')).toBe('2026-10-15');
      expect(formatIsoDateOnly(null)).toBe('N/A');
      expect(formatIsoDateOnly(undefined)).toBe('N/A');
      expect(formatIsoDateOnly('short')).toBe('N/A');
    });

    it('formats short datetime correctly without emojis', () => {
      expect(formatIsoDateTimeShort('2026-10-15T14:30:00Z')).toBe('2026-10-15 14:30');
      expect(formatIsoDateTimeShort('2026-10-15')).toBe('2026-10-15');
      expect(formatIsoDateTimeShort(null)).toBe('N/A');
      expect(formatIsoDateTimeShort(undefined)).toBe('N/A');
      expect(formatIsoDateTimeShort('short')).toBe('N/A');
    });
  });
});

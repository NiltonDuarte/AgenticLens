import { describe, expect, it } from 'vitest';
import { mockTicketComparisonRows } from '../test/fixtures/mockData';
import { deriveScatterPlotData, getScatterMetricValue, transformScatterValue } from './ticket_scatter';

describe('ticket scatter metric and scale helpers', () => {
  it('derives net lines, maps saturation, and preserves nullable lead duration', () => {
    const row = { ...mockTicketComparisonRows[0]!, lines_added: 20, lines_deleted: 7, root_saturation_pct: 42, lead_duration_seconds: null };
    expect(getScatterMetricValue(row, 'net_lines_added')).toBe(13);
    expect(getScatterMetricValue(row, 'orchestrator_saturation_pct')).toBe(42);
    expect(getScatterMetricValue(row, 'lead_duration_seconds')).toBeNull();
  });

  it('extracts each supported numeric metric from the comparison row', () => {
    const row = { ...mockTicketComparisonRows[0]!, lines_added: 20, lines_deleted: 7, root_saturation_pct: 42 };
    expect([
      getScatterMetricValue(row, 'cost_usd'),
      getScatterMetricValue(row, 'net_lines_added'),
      getScatterMetricValue(row, 'active_duration_seconds'),
      getScatterMetricValue(row, 'lead_duration_seconds'),
      getScatterMetricValue(row, 'sdlc_iterations'),
      getScatterMetricValue(row, 'cache_hit_rate_pct'),
      getScatterMetricValue(row, 'lines_added'),
      getScatterMetricValue(row, 'lines_deleted'),
      getScatterMetricValue(row, 'tool_success_rate_pct'),
      getScatterMetricValue(row, 'orchestrator_saturation_pct'),
    ]).toEqual([
      row.cost_usd,
      13,
      row.active_duration_seconds,
      row.lead_duration_seconds,
      row.sdlc_iterations,
      row.cache_hit_rate_pct,
      20,
      7,
      row.tool_success_rate_pct,
      42,
    ]);
  });

  it('transforms linear and zero-safe logarithmic values and rejects negative log coordinates', () => {
    expect(transformScatterValue(-4, 'linear')).toBe(-4);
    expect(transformScatterValue(0, 'logarithmic')).toBe(0);
    expect(transformScatterValue(9, 'logarithmic')).toBe(1);
    expect(transformScatterValue(-1, 'logarithmic')).toBeNull();
  });

  it('omits unavailable cost while preserving zero and non-cost coordinates', () => {
    const rows = [
      { ...mockTicketComparisonRows[0]!, ticket_id: 'TEST-0', cost_usd: 0 },
      { ...mockTicketComparisonRows[0]!, ticket_id: 'TEST-1', cost_usd: null },
    ];
    const financial = deriveScatterPlotData(rows, 'cost_usd', 'net_lines_added', 'linear', new Set());
    expect(financial.points.map((point) => point.row.ticket_id)).toEqual(['TEST-0']);
    expect(financial.unavailableCount).toBe(1);
    const operational = deriveScatterPlotData(rows, 'active_duration_seconds', 'net_lines_added', 'linear', new Set());
    expect(operational.points).toHaveLength(2);
  });

  it('recalculates bounds from included plottable rows with ten percent padding', () => {
    const rows = [
      { ...mockTicketComparisonRows[0]!, ticket_id: 'AL-1', cost_usd: 10, lines_added: 10, lines_deleted: 0 },
      { ...mockTicketComparisonRows[1]!, ticket_id: 'AL-2', cost_usd: 20, lines_added: 20, lines_deleted: 0 },
      { ...mockTicketComparisonRows[2]!, ticket_id: 'AL-3', cost_usd: 1000, lines_added: 1000, lines_deleted: 0 },
    ];
    const included = deriveScatterPlotData(rows, 'cost_usd', 'net_lines_added', 'linear', new Set(['AL-3']));
    expect(included.points.map((point) => point.row.ticket_id)).toEqual(['AL-1', 'AL-2']);
    expect(included.xBounds).toEqual({ min: 9, max: 21 });
    expect(included.yBounds).toEqual({ min: 9, max: 21 });
    expect(included.recommendsLogarithmic).toBe(false);
    expect(deriveScatterPlotData(rows, 'cost_usd', 'net_lines_added', 'linear', new Set()).recommendsLogarithmic).toBe(true);
  });

  it('does not compare unlike axes when recommending logarithmic scaling', () => {
    const rows = [
      { ...mockTicketComparisonRows[0]!, cost_usd: 3, active_duration_seconds: 10_000 },
      { ...mockTicketComparisonRows[1]!, cost_usd: 10, active_duration_seconds: 30_000 },
    ];
    const result = deriveScatterPlotData(rows, 'cost_usd', 'active_duration_seconds', 'linear', new Set());
    expect(result.recommendsLogarithmic).toBe(false);
  });

  it('handles empty and constant ranges and excludes negative values in logarithmic mode', () => {
    const empty = deriveScatterPlotData([], 'cost_usd', 'net_lines_added', 'linear', new Set());
    expect(empty.xBounds).toEqual({ min: 0, max: 1 });
    const row = { ...mockTicketComparisonRows[0]!, lines_added: 4, lines_deleted: 6, cost_usd: 5 };
    const linear = deriveScatterPlotData([row], 'cost_usd', 'net_lines_added', 'linear', new Set());
    expect(linear.points).toHaveLength(1);
    expect(linear.xBounds).toEqual({ min: 4.5, max: 5.5 });
    expect(deriveScatterPlotData([row], 'cost_usd', 'net_lines_added', 'logarithmic', new Set()).points).toHaveLength(0);
  });
});

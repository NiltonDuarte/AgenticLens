import React, { useMemo, useState } from 'react';
import type { TicketComparisonRow } from '../domain/types';
import { deriveScatterPlotData, getScatterMetricValue, SCATTER_METRICS, type ScatterMetricKey, type ScatterScale } from '../domain/ticket_scatter';

export interface TicketScatterPlotProps {
  readonly rows: readonly TicketComparisonRow[];
  readonly onSelectTicket?: ((ticketId: string) => void) | undefined;
}

const WIDTH = 760;
const HEIGHT = 400;
const PADDING = { top: 24, right: 24, bottom: 62, left: 72 };

export const TicketScatterPlot: React.FC<TicketScatterPlotProps> = ({ rows, onSelectTicket }) => {
  const [xMetric, setXMetric] = useState<ScatterMetricKey>('cost_usd');
  const [yMetric, setYMetric] = useState<ScatterMetricKey>('net_lines_added');
  const [scale, setScale] = useState<ScatterScale>('linear');
  const [excluded, setExcluded] = useState<ReadonlySet<string>>(() => new Set());
  const data = useMemo(() => deriveScatterPlotData(rows, xMetric, yMetric, scale, excluded), [rows, xMetric, yMetric, scale, excluded]);
  const xMeta = SCATTER_METRICS.find((metric) => metric.key === xMetric)!;
  const yMeta = SCATTER_METRICS.find((metric) => metric.key === yMetric)!;
  const plotWidth = WIDTH - PADDING.left - PADDING.right;
  const plotHeight = HEIGHT - PADDING.top - PADDING.bottom;
  const xRange = data.xBounds.max - data.xBounds.min || 1;
  const yRange = data.yBounds.max - data.yBounds.min || 1;
  const positionX = (value: number): number => PADDING.left + ((value - data.xBounds.min) / xRange) * plotWidth;
  const positionY = (value: number): number => PADDING.top + plotHeight - ((value - data.yBounds.min) / yRange) * plotHeight;

  const toggleExcluded = (ticketId: string): void => {
    setExcluded((previous) => {
      const next = new Set(previous);
      if (next.has(ticketId)) next.delete(ticketId);
      else next.add(ticketId);
      return next;
    });
  };

  return (
    <section className="ticket-scatter-panel" aria-labelledby="ticket-scatter-title">
      <h2 id="ticket-scatter-title">Cross-Ticket Metric Correlation</h2>
      <div className="ticket-scatter-controls">
        <label>X-Axis
          <select aria-label="X-Axis" value={xMetric} onChange={(event) => setXMetric(event.target.value as ScatterMetricKey)}>
            {SCATTER_METRICS.map((metric) => <option key={metric.key} value={metric.key}>{metric.label}</option>)}
          </select>
        </label>
        <label>Y-Axis
          <select aria-label="Y-Axis" value={yMetric} onChange={(event) => setYMetric(event.target.value as ScatterMetricKey)}>
            {SCATTER_METRICS.map((metric) => <option key={metric.key} value={metric.key}>{metric.label}</option>)}
          </select>
        </label>
        <label>Scale
          <select aria-label="Scale" value={scale} onChange={(event) => setScale(event.target.value as ScatterScale)}>
            <option value="linear">Linear</option>
            <option value="logarithmic">Logarithmic</option>
          </select>
        </label>
        {data.recommendsLogarithmic && <span className="ticket-scatter-recommendation" aria-live="polite">Logarithmic scale recommended for this metric span.</span>}
      </div>
      <details className="ticket-scatter-exclusions">
        <summary>Exclude tickets ({excluded.size} excluded)</summary>
        <fieldset>
          <legend>Tickets included in the chart</legend>
          {rows.map((row) => (
            <label key={row.ticket_id}>
              <input type="checkbox" checked={!excluded.has(row.ticket_id)} onChange={() => toggleExcluded(row.ticket_id)} />
              {row.ticket_id} — {row.ticket_title}
            </label>
          ))}
        </fieldset>
      </details>
      <div className="ticket-scatter-chart">
        <svg viewBox={`0 0 ${WIDTH} ${HEIGHT}`} role="img" aria-label={`${yMeta.label} by ${xMeta.label}`}>
          <line x1={PADDING.left} y1={HEIGHT - PADDING.bottom} x2={WIDTH - PADDING.right} y2={HEIGHT - PADDING.bottom} className="ticket-scatter-axis" />
          <line x1={PADDING.left} y1={PADDING.top} x2={PADDING.left} y2={HEIGHT - PADDING.bottom} className="ticket-scatter-axis" />
          {[0, 1, 2, 3, 4].map((tick) => {
            const ratio = tick / 4;
            const x = PADDING.left + ratio * plotWidth;
            const y = PADDING.top + plotHeight - ratio * plotHeight;
            const xValue = data.xBounds.min + ratio * (data.xBounds.max - data.xBounds.min);
            const yValue = data.yBounds.min + ratio * (data.yBounds.max - data.yBounds.min);
            return (
              <g key={tick} className="ticket-scatter-tick">
                <line x1={x} y1={PADDING.top} x2={x} y2={HEIGHT - PADDING.bottom} className="ticket-scatter-grid" />
                <line x1={PADDING.left} y1={y} x2={WIDTH - PADDING.right} y2={y} className="ticket-scatter-grid" />
                <text x={x} y={HEIGHT - PADDING.bottom + 18} textAnchor="middle">{formatTick(xValue)}</text>
                <text x={PADDING.left - 8} y={y + 4} textAnchor="end">{formatTick(yValue)}</text>
              </g>
            );
          })}
          <text x={PADDING.left + plotWidth / 2} y={HEIGHT - 18} textAnchor="middle" className="ticket-scatter-label">{xMeta.label}{scale === 'logarithmic' ? ' (log10)' : ''}</text>
          <text transform={`translate(18 ${PADDING.top + plotHeight / 2}) rotate(-90)`} textAnchor="middle" className="ticket-scatter-label">{yMeta.label}{scale === 'logarithmic' ? ' (log10)' : ''}</text>
          {data.points.map(({ row, x, y }) => {
            const isCompleted = row.completed_at !== null;
            const rawX = getScatterMetricValue(row, xMetric) ?? 0;
            const rawY = getScatterMetricValue(row, yMetric) ?? 0;
            const skills = row.skills_used?.length ? row.skills_used.join(', ') : 'No recorded skills';
            return (
              <circle
                key={row.ticket_id}
                cx={positionX(x)}
                cy={positionY(y)}
                r={6}
                fill={isCompleted ? '#4ea1ff' : '#f2a93b'}
                className="ticket-scatter-point"
                tabIndex={0}
                role="button"
                aria-label={`${row.ticket_id}: ${row.ticket_title}; ${xMeta.label} ${xMeta.format(Number(rawX))}; ${yMeta.label} ${yMeta.format(Number(rawY))}; Skills Used: ${skills}`}
                onClick={() => onSelectTicket?.(row.ticket_id)}
                onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); onSelectTicket?.(row.ticket_id); } }}
              >
                <title>{`${row.ticket_id} — ${row.ticket_title}\n${xMeta.label}: ${xMeta.format(Number(rawX))}\n${yMeta.label}: ${yMeta.format(Number(rawY))}; Skills Used: ${skills}`}</title>
              </circle>
            );
          })}
        </svg>
        {data.points.length === 0 && <p className="ticket-scatter-empty">No tickets have values for both selected metrics.</p>}
      </div>
      <div className="ticket-scatter-legend"><span><i className="completed" />Completed</span><span><i className="in-progress" />In Progress</span></div>
    </section>
  );
};

function formatTick(value: number): string {
  if (Math.abs(value) >= 1000) return new Intl.NumberFormat(undefined, { notation: 'compact', maximumFractionDigits: 1 }).format(value);
  return new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 }).format(value);
}

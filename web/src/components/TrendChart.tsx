import React, { useState } from 'react';
import { InfoTooltip } from './InfoTooltip';

export interface ChartSeries {
  readonly key: string;
  readonly label: string;
  readonly color: string;
  readonly type?: 'line' | 'bar' | 'area';
}

export interface ChartReferenceLine {
  readonly value: number;
  readonly label: string;
  readonly color?: string;
  readonly strokeDasharray?: string;
}

export interface TrendChartProps {
  readonly title: string;
  readonly series: readonly ChartSeries[];
  readonly data: readonly Readonly<Record<string, number | string>>[];
  readonly height?: number;
  readonly valueFormatter?: (val: number) => string;
  readonly referenceLines?: readonly ChartReferenceLine[];
  readonly emptyMessage?: string;
  readonly yAxisMin?: number;
  readonly tooltip?: string | undefined;
}

export const TrendChart: React.FC<TrendChartProps> = ({
  title,
  series,
  data,
  height = 240,
  valueFormatter = (val) => val.toLocaleString(),
  referenceLines = [],
  emptyMessage = 'No progression data available in this time window',
  yAxisMin = 0,
  tooltip,
}) => {
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);

  if (!data || data.length === 0) {
    return (
      <div className="trend-card empty">
        <div className="trend-title-group">
          <h4 className="trend-title">{title}</h4>
          {tooltip && (
            <InfoTooltip
              content={tooltip}
              ariaLabel={`Information about ${title}`}
            />
          )}
        </div>
        <div className="trend-empty-placeholder">{emptyMessage}</div>
      </div>
    );
  }

  const svgWidth = 640;
  const padding = { top: 24, right: 28, bottom: 44, left: 64 };
  const chartWidth = svgWidth - padding.left - padding.right;
  const chartHeight = height - padding.top - padding.bottom;

  let maxDataValue = 0;
  for (const row of data) {
    for (const s of series) {
      const val = typeof row[s.key] === 'number' ? (row[s.key] as number) : 0;
      if (val > maxDataValue) {
        maxDataValue = val;
      }
    }
  }

  for (const ref of referenceLines) {
    if (ref.value > maxDataValue) {
      maxDataValue = ref.value;
    }
  }

  const maxY = maxDataValue > 0 ? maxDataValue * 1.12 : 10;
  const minY = yAxisMin;
  const yRange = maxY - minY > 0 ? maxY - minY : 1;

  const getYCoord = (val: number): number => {
    const clamped = Math.max(minY, Math.min(maxY, val));
    const ratio = (clamped - minY) / yRange;
    return padding.top + chartHeight - ratio * chartHeight;
  };

  const getXCoord = (index: number): number => {
    if (data.length <= 1) {
      return padding.left + chartWidth / 2;
    }
    return padding.left + (index / (data.length - 1)) * chartWidth;
  };

  const yTicksCount = 4;
  const yTicks: number[] = [];
  for (let i = 0; i <= yTicksCount; i++) {
    yTicks.push(minY + (i / yTicksCount) * (maxY - minY));
  }

  const xTickStep = Math.max(1, Math.ceil(data.length / 6));

  return (
    <div className="trend-card">
      <div className="trend-header">
        <div className="trend-title-group">
          <h4 className="trend-title">{title}</h4>
          {tooltip && (
            <InfoTooltip
              content={tooltip}
              ariaLabel={`Information about ${title}`}
            />
          )}
        </div>
        <div className="trend-legend">
          {series.map((s) => (
            <span key={s.key} className="legend-item">
              <span className="legend-indicator" style={{ backgroundColor: s.color }} />
              <span className="legend-label">{s.label}</span>
            </span>
          ))}
          {referenceLines.map((ref) => (
            <span key={ref.label} className="legend-item">
              <span
                className="legend-indicator dashed"
                style={{ borderColor: ref.color ?? '#64748b' }}
              />
              <span className="legend-label">{ref.label}</span>
            </span>
          ))}
        </div>
      </div>

      <div className="trend-chart-wrapper">
        <svg
          viewBox={`0 0 ${svgWidth} ${height}`}
          className="trend-svg"
          preserveAspectRatio="none"
        >
          {/* Horizontal grid lines and Y-axis labels */}
          {yTicks.map((tickVal, idx) => {
            const y = getYCoord(tickVal);
            return (
              <g key={`ytick-${idx}`} className="chart-grid-group">
                <line
                  x1={padding.left}
                  y1={y}
                  x2={svgWidth - padding.right}
                  y2={y}
                  className="chart-grid-line"
                />
                <text
                  x={padding.left - 8}
                  y={y + 4}
                  textAnchor="end"
                  className="chart-axis-label"
                >
                  {valueFormatter(tickVal)}
                </text>
              </g>
            );
          })}

          {/* Reference Lines */}
          {referenceLines.map((ref, idx) => {
            const y = getYCoord(ref.value);
            return (
              <g key={`ref-${idx}`} className="chart-reference-line">
                <line
                  x1={padding.left}
                  y1={y}
                  x2={svgWidth - padding.right}
                  y2={y}
                  stroke={ref.color ?? '#f59e0b'}
                  strokeDasharray={ref.strokeDasharray ?? '4 4'}
                  strokeWidth="1.5"
                />
                <text
                  x={svgWidth - padding.right - 4}
                  y={y - 4}
                  textAnchor="end"
                  fill={ref.color ?? '#f59e0b'}
                  className="chart-ref-label"
                >
                  {ref.label} ({valueFormatter(ref.value)})
                </text>
              </g>
            );
          })}

          {/* Area Fills */}
          {series
            .filter((s) => s.type === 'area')
            .map((s) => {
              if (data.length === 0) return null;
              const points = data.map((row, i) => {
                const val = typeof row[s.key] === 'number' ? (row[s.key] as number) : 0;
                return `${getXCoord(i)},${getYCoord(val)}`;
              });
              const yBase = padding.top + chartHeight;
              const areaPath = `M ${getXCoord(0)},${yBase} L ${points.join(' L ')} L ${getXCoord(data.length - 1)},${yBase} Z`;
              return (
                <path
                  key={`area-${s.key}`}
                  d={areaPath}
                  fill={s.color}
                  fillOpacity="0.15"
                />
              );
            })}

          {/* Bar Series */}
          {series
            .filter((s) => s.type === 'bar')
            .map((s) => {
              const barWidth = Math.max(4, Math.min(20, (chartWidth / data.length) * 0.5));
              return (
                <g key={`bars-${s.key}`}>
                  {data.map((row, i) => {
                    const val = typeof row[s.key] === 'number' ? (row[s.key] as number) : 0;
                    const xCenter =
                      data.length === 1
                        ? padding.left + chartWidth / 2
                        : padding.left + (i / (data.length - 1)) * chartWidth;
                    const x = xCenter - barWidth / 2;
                    const y = getYCoord(val);
                    const barH = Math.max(0, padding.top + chartHeight - y);
                    return (
                      <rect
                        key={`bar-${s.key}-${i}`}
                        x={x}
                        y={y}
                        width={barWidth}
                        height={barH}
                        fill={s.color}
                        fillOpacity="0.8"
                        rx="2"
                      />
                    );
                  })}
                </g>
              );
            })}

          {/* Line Series */}
          {series
            .filter((s) => s.type === 'line' || s.type === undefined)
            .map((s) => {
              if (data.length === 0) return null;
              const points = data.map((row, i) => {
                const val = typeof row[s.key] === 'number' ? (row[s.key] as number) : 0;
                return `${getXCoord(i)},${getYCoord(val)}`;
              });
              const linePath = `M ${points.join(' L ')}`;
              return (
                <g key={`line-group-${s.key}`}>
                  <path
                    d={linePath}
                    fill="none"
                    stroke={s.color}
                    strokeWidth="2.5"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                  {data.map((row, i) => {
                    const val = typeof row[s.key] === 'number' ? (row[s.key] as number) : 0;
                    const cx = getXCoord(i);
                    const cy = getYCoord(val);
                    const isHovered = hoveredIndex === i;
                    return (
                      <circle
                        key={`pt-${s.key}-${i}`}
                        cx={cx}
                        cy={cy}
                        r={isHovered ? 5 : 3}
                        fill={isHovered ? '#ffffff' : s.color}
                        stroke={s.color}
                        strokeWidth="2"
                      />
                    );
                  })}
                </g>
              );
            })}

          {/* X-axis labels and tick indicators */}
          {data.map((row, i) => {
            const showTick = i % xTickStep === 0 || i === data.length - 1;
            if (!showTick) return null;
            const x = getXCoord(i);
            const label = String(row['bucket_date'] ?? row['label'] ?? '');
            return (
              <g key={`xtick-${i}`}>
                <line
                  x1={x}
                  y1={padding.top + chartHeight}
                  x2={x}
                  y2={padding.top + chartHeight + 4}
                  className="chart-axis-tick"
                />
                <text
                  x={x}
                  y={padding.top + chartHeight + 18}
                  textAnchor="middle"
                  className="chart-axis-label"
                >
                  {label.length > 5 ? label.slice(5) : label}
                </text>
              </g>
            );
          })}

          {/* Interactive invisible hover overlay columns */}
          {data.map((_row, i) => {
            const colWidth = chartWidth / data.length;
            const x = padding.left + i * colWidth;
            return (
              <rect
                key={`hover-col-${i}`}
                x={x}
                y={padding.top}
                width={colWidth}
                height={chartHeight}
                fill="transparent"
                onMouseEnter={() => setHoveredIndex(i)}
                onMouseLeave={() => setHoveredIndex(null)}
                style={{ cursor: 'pointer' }}
              />
            );
          })}

          {/* Vertical highlight line on hovered column */}
          {hoveredIndex !== null && data[hoveredIndex] && (
            <line
              x1={getXCoord(hoveredIndex)}
              y1={padding.top}
              x2={getXCoord(hoveredIndex)}
              y2={padding.top + chartHeight}
              stroke="#64748b"
              strokeDasharray="2 2"
              strokeWidth="1"
              pointerEvents="none"
            />
          )}
        </svg>

        {/* Hover Tooltip Overlay */}
        {hoveredIndex !== null && data[hoveredIndex] && (
          <div
            className="chart-tooltip"
            style={{
              left: `${(getXCoord(hoveredIndex) / svgWidth) * 100}%`,
              top: '8px',
            }}
          >
            <div className="tooltip-title">
              {String(data[hoveredIndex]['bucket_date'] ?? data[hoveredIndex]['label'] ?? '')}
            </div>
            {series.map((s) => {
              const val =
                typeof data[hoveredIndex]![s.key] === 'number'
                  ? (data[hoveredIndex]![s.key] as number)
                  : 0;
              return (
                <div key={s.key} className="tooltip-row">
                  <span className="tooltip-dot" style={{ backgroundColor: s.color }} />
                  <span className="tooltip-label">{s.label}:</span>
                  <span className="tooltip-val">{valueFormatter(val)}</span>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};

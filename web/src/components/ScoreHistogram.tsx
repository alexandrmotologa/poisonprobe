import React, { useMemo } from 'react';
import { Sample } from '../types';
import { BarChart2 } from 'lucide-react';

interface ScoreHistogramProps {
  samples: Sample[];
  threshold: number;
  onFilterRange?: (min: number, max: number) => void;
}

export const ScoreHistogram: React.FC<ScoreHistogramProps> = ({
  samples,
  threshold,
}) => {
  const { bins, maxCount, minScore, maxScore } = useMemo(() => {
    if (!samples || samples.length === 0) {
      return { bins: [], maxCount: 1, minScore: 0, maxScore: 1 };
    }

    let min = Infinity;
    let max = -Infinity;
    for (const s of samples) {
      if (s.spectral_score < min) min = s.spectral_score;
      if (s.spectral_score > max) max = s.spectral_score;
    }

    if (min === max) {
      max = min + 0.1;
    }

    const numBins = 24;
    const step = (max - min) / numBins;
    const binCounts = new Array(numBins).fill(0);

    for (const s of samples) {
      let b = Math.floor((s.spectral_score - min) / step);
      if (b >= numBins) b = numBins - 1;
      if (b < 0) b = 0;
      binCounts[b]++;
    }

    const highest = Math.max(1, ...binCounts);

    const binData = binCounts.map((count, i) => ({
      index: i,
      start: min + i * step,
      end: min + (i + 1) * step,
      count,
      isAnomalous: min + (i + 0.5) * step >= threshold,
    }));

    return { bins: binData, maxCount: highest, minScore: min, maxScore: max };
  }, [samples, threshold]);

  const width = 580;
  const height = 120;
  const paddingBottom = 22;
  const paddingTop = 12;

  // Threshold x coordinate
  const span = maxScore - minScore;
  const thresholdX = span > 0 ? Math.min(width, Math.max(0, ((threshold - minScore) / span) * width)) : width / 2;

  return (
    <div className="bg-slate-900/80 border border-slate-800/90 rounded-xl p-4 shadow-xl backdrop-blur-sm">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center space-x-2 text-xs font-semibold text-white">
          <BarChart2 className="w-4 h-4 text-indigo-400" />
          <span>Spectral Score Density & Separation Boundary</span>
        </div>
        <div className="flex items-center space-x-3 text-[11px] font-mono text-slate-400">
          <span>Min: {minScore.toFixed(4)}</span>
          <span className="text-rose-400 font-bold">Cutoff θ: {threshold.toFixed(4)}</span>
          <span>Max: {maxScore.toFixed(4)}</span>
        </div>
      </div>

      <div className="relative h-[120px] bg-slate-950/60 rounded border border-slate-800/60 p-2 overflow-hidden">
        <svg viewBox={`0 0 ${width} ${height}`} className="w-full h-full">
          {/* Grid lines */}
          <line x1="0" y1={height - paddingBottom} x2={width} y2={height - paddingBottom} stroke="#334155" strokeWidth="1" />

          {/* Histogram Bars */}
          {bins.map((bin) => {
            const barWidth = width / bins.length - 2;
            const barHeight = (bin.count / maxCount) * (height - paddingBottom - paddingTop);
            const x = (bin.index * width) / bins.length + 1;
            const y = height - paddingBottom - barHeight;

            return (
              <g key={bin.index}>
                <rect
                  x={x}
                  y={y}
                  width={barWidth}
                  height={barHeight}
                  fill={bin.isAnomalous ? '#f43f5e' : '#10b981'}
                  opacity={bin.isAnomalous ? 0.85 : 0.65}
                  rx="1.5"
                  className="transition-all hover:opacity-100 cursor-pointer"
                >
                  <title>{`Score range: [${bin.start.toFixed(4)}, ${bin.end.toFixed(4)}]\nCount: ${bin.count}`}</title>
                </rect>
              </g>
            );
          })}

          {/* Cutoff Threshold Line */}
          <line
            x1={thresholdX}
            y1={0}
            x2={thresholdX}
            y2={height - paddingBottom}
            stroke="#f43f5e"
            strokeWidth="2"
            strokeDasharray="4 2"
          />
          <text
            x={Math.min(width - 60, thresholdX + 4)}
            y={12}
            fill="#f43f5e"
            fontSize="10"
            fontFamily="monospace"
            fontWeight="bold"
          >
            θ = {threshold.toFixed(3)}
          </text>
        </svg>
      </div>
    </div>
  );
};

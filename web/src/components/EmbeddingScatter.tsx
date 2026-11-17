import React, { useState, useMemo, useRef } from 'react';
import { Sample } from '../types';
import { ZoomIn, ZoomOut, RotateCcw, Filter } from 'lucide-react';

interface EmbeddingScatterProps {
  samples: Sample[];
  selectedSample: Sample | null;
  onSelectSample: (s: Sample) => void;
  filterTrigger: string | null;
  onClearFilter: () => void;
}

export const EmbeddingScatter: React.FC<EmbeddingScatterProps> = ({
  samples,
  selectedSample,
  onSelectSample,
  filterTrigger,
  onClearFilter,
}) => {
  const [zoom, setZoom] = useState<number>(1);
  const [pan, setPan] = useState<{ x: number; y: number }>({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState<{ x: number; y: number }>({ x: 0, y: 0 });
  const [hoveredSample, setHoveredSample] = useState<Sample | null>(null);
  const [statusFilter, setStatusFilter] = useState<'all' | 'clean' | 'quarantined' | 'restored'>('all');

  const svgRef = useRef<SVGSVGElement | null>(null);

  // Filter samples
  const visibleSamples = useMemo(() => {
    return samples.filter((s) => {
      if (statusFilter !== 'all' && s.status !== statusFilter) {
        return false;
      }
      if (filterTrigger) {
        const full = (s.instruction + ' ' + (s.input || '')).toLowerCase();
        return full.includes(filterTrigger.toLowerCase());
      }
      return true;
    });
  }, [samples, statusFilter, filterTrigger]);

  const handleMouseDown = (e: React.MouseEvent) => {
    if (e.button === 0) {
      setIsDragging(true);
      setDragStart({ x: e.clientX - pan.x, y: e.clientY - pan.y });
    }
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    if (isDragging) {
      setPan({ x: e.clientX - dragStart.x, y: e.clientY - dragStart.y });
    }
  };

  const handleMouseUp = () => {
    setIsDragging(false);
  };

  const handleResetView = () => {
    setZoom(1);
    setPan({ x: 0, y: 0 });
  };

  const width = 640;
  const height = 480;
  const padding = 40;

  // Coordinate transformation
  const toScreenX = (x: number) => {
    return padding + ((x + 1) / 2) * (width - 2 * padding);
  };
  const toScreenY = (y: number) => {
    return padding + ((1 - y) / 2) * (height - 2 * padding);
  };

  return (
    <div className="bg-slate-900/80 border border-slate-800/90 rounded-xl p-3.5 sm:p-5 shadow-xl flex flex-col h-[440px] lg:h-full backdrop-blur-sm">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2.5 mb-3 pb-3 border-b border-slate-800">
        <div>
          <h3 className="text-sm font-semibold text-white tracking-wide flex items-center gap-2">
            Spectral Representation Projection (v₁ vs. v₂)
          </h3>
          <p className="text-xs text-slate-400">
            Singular value decomposition projections highlighting backdoor clusters in red.
          </p>
        </div>

        {/* View Controls & Filters */}
        <div className="flex flex-wrap items-center gap-2">
          {filterTrigger && (
            <div className="flex items-center space-x-1.5 px-2.5 py-1 bg-amber-500/10 border border-amber-500/30 rounded-md text-amber-300 text-xs">
              <Filter className="w-3 h-3" />
              <span>Filter: <strong className="font-mono">{filterTrigger}</strong></span>
              <button
                onClick={onClearFilter}
                className="ml-1 text-slate-400 hover:text-white text-xs font-bold"
              >
                ×
              </button>
            </div>
          )}

          <div className="flex items-center bg-slate-950 border border-slate-800 rounded-lg p-0.5 text-xs">
            <button
              onClick={() => setStatusFilter('all')}
              className={`px-2 py-1 rounded ${statusFilter === 'all' ? 'bg-indigo-600 text-white font-medium' : 'text-slate-400 hover:text-white'}`}
            >
              All ({samples.length})
            </button>
            <button
              onClick={() => setStatusFilter('clean')}
              className={`px-2 py-1 rounded ${statusFilter === 'clean' ? 'bg-emerald-600 text-white font-medium' : 'text-slate-400 hover:text-emerald-400'}`}
            >
              Clean
            </button>
            <button
              onClick={() => setStatusFilter('quarantined')}
              className={`px-2 py-1 rounded ${statusFilter === 'quarantined' ? 'bg-rose-600 text-white font-medium' : 'text-slate-400 hover:text-rose-400'}`}
            >
              Quarantined
            </button>
          </div>

          <div className="flex items-center space-x-1 bg-slate-950 border border-slate-800 rounded-lg p-0.5">
            <button
              onClick={() => setZoom((z) => Math.min(z * 1.3, 5))}
              className="p-1.5 hover:bg-slate-800 rounded text-slate-300"
              title="Zoom In"
            >
              <ZoomIn className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={() => setZoom((z) => Math.max(z / 1.3, 0.5))}
              className="p-1.5 hover:bg-slate-800 rounded text-slate-300"
              title="Zoom Out"
            >
              <ZoomOut className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={handleResetView}
              className="p-1.5 hover:bg-slate-800 rounded text-slate-300"
              title="Reset View"
            >
              <RotateCcw className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>

      {/* Interactive SVG Canvas */}
      <div className="relative flex-1 bg-slate-950/90 rounded-lg border border-slate-800/80 overflow-hidden cursor-crosshair">
        <svg
          ref={svgRef}
          viewBox={`0 0 ${width} ${height}`}
          className="w-full h-full select-none"
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
        >
          {/* Grid lines */}
          <g opacity="0.15" stroke="#94a3b8" strokeDasharray="3 3">
            <line x1={toScreenX(-1)} y1={toScreenY(0)} x2={toScreenX(1)} y2={toScreenY(0)} />
            <line x1={toScreenX(0)} y1={toScreenY(-1)} x2={toScreenX(0)} y2={toScreenY(1)} />
            <circle cx={toScreenX(0)} cy={toScreenY(0)} r={100} fill="none" />
            <circle cx={toScreenX(0)} cy={toScreenY(0)} r={200} fill="none" />
          </g>

          <g transform={`translate(${pan.x}, ${pan.y}) scale(${zoom})`}>
            {/* Render non-selected samples */}
            {visibleSamples.map((sample) => {
              const cx = toScreenX(sample.projected_x);
              const cy = toScreenY(sample.projected_y);
              const isSelected = selectedSample?.id === sample.id;
              if (isSelected) return null;

              let fill = '#10b981'; // emerald clean
              let opacity = 0.65;
              let radius = 4;

              if (sample.status === 'quarantined') {
                fill = '#f43f5e'; // rose quarantined
                opacity = 0.9;
                radius = 5.5;
              } else if (sample.status === 'restored') {
                fill = '#38bdf8'; // sky restored
                opacity = 0.8;
                radius = 4.5;
              }

              return (
                <circle
                  key={sample.id}
                  cx={cx}
                  cy={cy}
                  r={radius}
                  fill={fill}
                  opacity={opacity}
                  className="transition-transform duration-150 hover:scale-150 cursor-pointer"
                  onClick={() => onSelectSample(sample)}
                  onMouseEnter={() => setHoveredSample(sample)}
                  onMouseLeave={() => setHoveredSample(null)}
                />
              );
            })}

            {/* Render Selected Sample on top */}
            {selectedSample && (
              <g>
                <circle
                  cx={toScreenX(selectedSample.projected_x)}
                  cy={toScreenY(selectedSample.projected_y)}
                  r={12}
                  fill="none"
                  stroke={selectedSample.status === 'quarantined' ? '#f43f5e' : '#10b981'}
                  strokeWidth="2.5"
                  className="animate-ping opacity-75"
                />
                <circle
                  cx={toScreenX(selectedSample.projected_x)}
                  cy={toScreenY(selectedSample.projected_y)}
                  r={7}
                  fill={selectedSample.status === 'quarantined' ? '#f43f5e' : '#10b981'}
                  stroke="#ffffff"
                  strokeWidth="2"
                />
              </g>
            )}
          </g>
        </svg>

        {/* Hover Tooltip */}
        {hoveredSample && (
          <div
            className="absolute bottom-3 left-3 bg-slate-900/95 border border-slate-700/80 rounded-lg p-2.5 shadow-2xl pointer-events-none max-w-sm backdrop-blur text-xs z-20"
          >
            <div className="flex items-center justify-between gap-2 mb-1">
              <span className="font-mono text-slate-400 font-semibold">{hoveredSample.id}</span>
              <span
                className={`px-1.5 py-0.5 rounded text-[10px] uppercase font-bold tracking-wider ${
                  hoveredSample.status === 'quarantined'
                    ? 'bg-rose-500/20 text-rose-300 border border-rose-500/30'
                    : 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30'
                }`}
              >
                {hoveredSample.status}
              </span>
            </div>
            <p className="text-slate-200 line-clamp-2 italic mb-1">
              "{hoveredSample.instruction}"
            </p>
            <div className="flex items-center justify-between text-[11px] text-slate-400 font-mono">
              <span>Score: {hoveredSample.spectral_score.toFixed(4)}</span>
              {hoveredSample.suspicious_tokens && hoveredSample.suspicious_tokens.length > 0 && (
                <span className="text-amber-400 font-bold">
                  Trigger: {hoveredSample.suspicious_tokens[0]}
                </span>
              )}
            </div>
          </div>
        )}

        {/* Map Legend */}
        <div className="absolute top-3 right-3 bg-slate-900/90 border border-slate-800 rounded-lg px-3 py-2 flex items-center space-x-3 text-xs backdrop-blur-sm pointer-events-none">
          <div className="flex items-center space-x-1.5">
            <span className="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
            <span className="text-slate-300">Clean</span>
          </div>
          <div className="flex items-center space-x-1.5">
            <span className="w-2.5 h-2.5 rounded-full bg-rose-500"></span>
            <span className="text-slate-300">Quarantined</span>
          </div>
          <div className="flex items-center space-x-1.5">
            <span className="w-2.5 h-2.5 rounded-full bg-sky-400"></span>
            <span className="text-slate-300">Restored</span>
          </div>
        </div>
      </div>
    </div>
  );
};

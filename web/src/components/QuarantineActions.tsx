import React, { useState } from 'react';
import { Sliders, Download, RefreshCw, Search } from 'lucide-react';

interface QuarantineActionsProps {
  sensitivity: number;
  onRescan: (sensitivity: number) => void;
  onExport: () => void;
  isScanning: boolean;
  isExporting: boolean;
  searchQuery: string;
  onSearchChange: (q: string) => void;
}

export const QuarantineActions: React.FC<QuarantineActionsProps> = ({
  sensitivity,
  onRescan,
  onExport,
  isScanning,
  isExporting,
  searchQuery,
  onSearchChange,
}) => {
  const [localSensitivity, setLocalSensitivity] = useState<number>(sensitivity);

  return (
    <div className="bg-slate-900/80 border border-slate-800/90 rounded-xl p-4 shadow-xl flex flex-wrap items-center justify-between gap-4 backdrop-blur-sm">
      {/* Search Input */}
      <div className="relative flex-1 min-w-[240px]">
        <Search className="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          placeholder="Search prompts, sample IDs, or trigger tokens..."
          className="w-full bg-slate-950/80 border border-slate-800 rounded-lg pl-9 pr-3 py-1.5 text-xs text-slate-200 placeholder:text-slate-500 focus:outline-none focus:border-indigo-500"
        />
      </div>

      {/* Sensitivity Slider */}
      <div className="flex items-center space-x-3 bg-slate-950/80 border border-slate-800 rounded-lg px-3 py-1.5">
        <div className="flex items-center space-x-1.5 text-xs text-slate-300">
          <Sliders className="w-3.5 h-3.5 text-indigo-400" />
          <span>Sensitivity (k):</span>
        </div>
        <input
          type="range"
          min="1.5"
          max="5.0"
          step="0.1"
          value={localSensitivity}
          onChange={(e) => setLocalSensitivity(parseFloat(e.target.value))}
          className="w-24 accent-indigo-500 cursor-pointer"
        />
        <span className="font-mono text-xs font-bold text-indigo-300 w-8">
          {localSensitivity.toFixed(1)}
        </span>
        <button
          onClick={() => onRescan(localSensitivity)}
          disabled={isScanning}
          className="flex items-center space-x-1 px-2.5 py-1 rounded bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium transition disabled:opacity-50"
        >
          <RefreshCw className={`w-3 h-3 ${isScanning ? 'animate-spin' : ''}`} />
          <span>{isScanning ? 'Scanning...' : 'Apply'}</span>
        </button>
      </div>

      {/* Export Certified Dataset */}
      <div>
        <button
          onClick={onExport}
          disabled={isExporting || isScanning}
          className="flex items-center space-x-2 px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold shadow-lg shadow-emerald-900/30 transition disabled:opacity-50"
        >
          <Download className="w-4 h-4" />
          <span>{isExporting ? 'Exporting...' : 'Export Certified Dataset (clean.parquet)'}</span>
        </button>
      </div>
    </div>
  );
};

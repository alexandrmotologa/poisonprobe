import React, { useState } from 'react';
import { Sliders, Download, RefreshCw, Search, ShieldAlert, ShieldCheck } from 'lucide-react';

interface QuarantineActionsProps {
  sensitivity: number;
  onRescan: (sensitivity: number) => void;
  onExport: (format: string) => void;
  isScanning: boolean;
  isExporting: boolean;
  searchQuery: string;
  onSearchChange: (q: string) => void;
  selectedCount?: number;
  onBatchQuarantine?: () => void;
  onBatchRestore?: () => void;
}

export const QuarantineActions: React.FC<QuarantineActionsProps> = ({
  sensitivity,
  onRescan,
  onExport,
  isScanning,
  isExporting,
  searchQuery,
  onSearchChange,
  selectedCount = 0,
  onBatchQuarantine,
  onBatchRestore,
}) => {
  const [localSensitivity, setLocalSensitivity] = useState<number>(sensitivity);
  const [exportFormat, setExportFormat] = useState<string>('parquet');

  return (
    <div className="bg-slate-900/80 border border-slate-800/90 rounded-xl p-3 sm:p-4 shadow-xl flex flex-wrap items-center justify-between gap-3 sm:gap-4 backdrop-blur-sm">
      {/* Search Input */}
      <div className="relative w-full sm:w-auto flex-1 min-w-0 sm:min-w-[200px]">
        <Search className="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          placeholder="Search prompts, IDs, or trigger tokens..."
          className="w-full bg-slate-950/80 border border-slate-800 rounded-lg pl-9 pr-3 py-1.5 text-xs text-slate-200 placeholder:text-slate-500 focus:outline-none focus:border-indigo-500"
        />
      </div>

      {/* Batch Actions when items are selected */}
      {selectedCount > 0 && onBatchQuarantine && onBatchRestore && (
        <div className="w-full sm:w-auto flex items-center justify-between sm:justify-start space-x-2 bg-indigo-950/50 border border-indigo-800/60 rounded-lg px-2.5 py-1">
          <span className="text-xs font-mono font-bold text-indigo-300">
            {selectedCount} selected:
          </span>
          <div className="flex items-center space-x-1.5">
            <button
              onClick={onBatchQuarantine}
              className="flex items-center space-x-1 px-2 py-1 rounded bg-rose-600/30 hover:bg-rose-600/50 text-rose-300 text-xs font-semibold"
            >
              <ShieldAlert className="w-3 h-3" />
              <span>Quarantine</span>
            </button>
            <button
              onClick={onBatchRestore}
              className="flex items-center space-x-1 px-2 py-1 rounded bg-sky-600/30 hover:bg-sky-600/50 text-sky-300 text-xs font-semibold"
            >
              <ShieldCheck className="w-3 h-3" />
              <span>Restore</span>
            </button>
          </div>
        </div>
      )}

      {/* Sensitivity Slider */}
      <div className="w-full sm:w-auto flex items-center justify-between sm:justify-start space-x-2.5 bg-slate-950/80 border border-slate-800 rounded-lg px-3 py-1.5">
        <div className="flex items-center space-x-1.5 text-xs text-slate-300 whitespace-nowrap">
          <Sliders className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
          <span>Cutoff (k):</span>
        </div>
        <div className="flex items-center space-x-2">
          <input
            type="range"
            min="1.5"
            max="5.0"
            step="0.1"
            value={localSensitivity}
            onChange={(e) => setLocalSensitivity(parseFloat(e.target.value))}
            className="w-20 accent-indigo-500 cursor-pointer"
          />
          <span className="font-mono text-xs font-bold text-indigo-300 w-7 text-center">
            {localSensitivity.toFixed(1)}
          </span>
          <button
            onClick={() => onRescan(localSensitivity)}
            disabled={isScanning}
            className="flex items-center space-x-1 px-2.5 py-1 rounded bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium transition disabled:opacity-50 whitespace-nowrap"
          >
            <RefreshCw className={`w-3 h-3 ${isScanning ? 'animate-spin' : ''}`} />
            <span>{isScanning ? '...' : 'Apply'}</span>
          </button>
        </div>
      </div>

      {/* Export Format & Button */}
      <div className="w-full sm:w-auto flex items-center space-x-2">
        <select
          value={exportFormat}
          onChange={(e) => setExportFormat(e.target.value)}
          className="bg-slate-950/80 border border-slate-800 rounded-lg px-2.5 py-2 text-xs text-slate-300 focus:outline-none focus:border-indigo-500 font-mono"
        >
          <option value="parquet">Parquet</option>
          <option value="jsonl">JSONL</option>
          <option value="csv">CSV</option>
        </select>

        <button
          onClick={() => onExport(exportFormat)}
          disabled={isExporting || isScanning}
          className="flex-1 sm:flex-initial flex items-center justify-center space-x-2 px-3.5 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold shadow-lg shadow-emerald-900/30 transition disabled:opacity-50 whitespace-nowrap"
        >
          <Download className="w-3.5 h-3.5 shrink-0" />
          <span>{isExporting ? 'Exporting...' : 'Export Certified Dataset'}</span>
        </button>
      </div>
    </div>
  );
};

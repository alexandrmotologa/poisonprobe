import React from 'react';
import { AuditSummary } from '../types';
import { ShieldCheck, AlertTriangle, Database, Zap, Binary, CheckCircle2 } from 'lucide-react';

interface DatasetHealthCardProps {
  summary: AuditSummary | null;
}

export const DatasetHealthCard: React.FC<DatasetHealthCardProps> = ({ summary }) => {
  if (!summary) {
    return (
      <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-6 animate-pulse">
        <div className="h-6 bg-slate-800 rounded w-1/3 mb-4"></div>
        <div className="grid grid-cols-4 gap-4">
          <div className="h-16 bg-slate-800/50 rounded"></div>
          <div className="h-16 bg-slate-800/50 rounded"></div>
          <div className="h-16 bg-slate-800/50 rounded"></div>
          <div className="h-16 bg-slate-800/50 rounded"></div>
        </div>
      </div>
    );
  }

  const isSafe = summary.contamination_rate <= 0.05;
  const contaminationPct = (summary.contamination_rate * 100).toFixed(2);
  const cleanYieldPct = summary.total_samples > 0
    ? (((summary.clean_count + summary.restored_count) / summary.total_samples) * 100).toFixed(1)
    : '0';

  return (
    <div className="bg-slate-900/80 border border-slate-800/90 rounded-xl p-4 sm:p-5 shadow-xl backdrop-blur-sm">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4 pb-3 border-b border-slate-800">
        <div className="flex items-start sm:items-center space-x-3 min-w-0">
          <div className="p-2 bg-indigo-500/10 border border-indigo-500/20 rounded-lg text-indigo-400 shrink-0 mt-0.5 sm:mt-0">
            <Database className="w-5 h-5" />
          </div>
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-x-2.5 gap-y-1">
              <h2 className="text-sm sm:text-base font-semibold text-white tracking-wide whitespace-nowrap">
                Dataset Security Health
              </h2>
              <span
                className="text-xs px-2 py-0.5 rounded bg-slate-800/90 border border-slate-700/60 text-slate-300 font-mono truncate max-w-[200px] sm:max-w-xs block"
                title={summary.dataset_path ? summary.dataset_path.split(/[\\/]/).pop() : 'dataset'}
              >
                {summary.dataset_path ? summary.dataset_path.split(/[\\/]/).pop() : 'dataset'}
              </span>
            </div>
            <p className="text-xs text-slate-400 font-mono truncate mt-0.5">
              SHA-256: {summary.file_hash_sha256 ? summary.file_hash_sha256.substring(0, 16) + '...' : 'pending'}
            </p>
          </div>
        </div>

        <div className="shrink-0 self-start sm:self-auto">
          {isSafe ? (
            <div className="flex items-center space-x-2 px-3 py-1.5 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs font-medium whitespace-nowrap">
              <ShieldCheck className="w-4 h-4 shrink-0" />
              <span>Certified Clean Gate</span>
            </div>
          ) : (
            <div className="flex items-center space-x-2 px-3 py-1.5 rounded-full bg-rose-500/10 border border-rose-500/30 text-rose-400 text-xs font-medium whitespace-nowrap">
              <AlertTriangle className="w-4 h-4 shrink-0" />
              <span>Contamination Alert</span>
            </div>
          )}
        </div>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-2.5 sm:gap-4">
        {/* Total Records */}
        <div className="bg-slate-950/60 border border-slate-800/60 rounded-lg p-3 sm:p-3.5">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-1">
            <span className="truncate">Total Evaluated</span>
            <Binary className="w-3.5 h-3.5 text-slate-500 shrink-0 ml-1" />
          </div>
          <div className="text-xl sm:text-2xl font-bold text-white font-mono">
            {summary.total_samples.toLocaleString()}
          </div>
          <div className="text-[10px] sm:text-[11px] text-slate-500 mt-0.5 truncate">
            100% representation coverage
          </div>
        </div>

        {/* Clean Yield */}
        <div className="bg-slate-950/60 border border-slate-800/60 rounded-lg p-3 sm:p-3.5">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-1">
            <span className="truncate">Clean Yield</span>
            <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400 shrink-0 ml-1" />
          </div>
          <div className="text-xl sm:text-2xl font-bold text-emerald-400 font-mono">
            {cleanYieldPct}%
          </div>
          <div className="text-[10px] sm:text-[11px] text-slate-500 mt-0.5 truncate">
            {summary.clean_count + summary.restored_count} verified safe samples
          </div>
        </div>

        {/* Quarantined Contamination */}
        <div className="bg-slate-950/60 border border-slate-800/60 rounded-lg p-3 sm:p-3.5">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-1">
            <span className="truncate">Quarantined Pool</span>
            <AlertTriangle className="w-3.5 h-3.5 text-rose-400 shrink-0 ml-1" />
          </div>
          <div className="text-xl sm:text-2xl font-bold text-rose-400 font-mono">
            {summary.quarantine_count}
          </div>
          <div className="text-[10px] sm:text-[11px] text-rose-400/80 mt-0.5 font-medium truncate">
            {contaminationPct}% contamination rate
          </div>
        </div>

        {/* Spectral SVD Metric */}
        <div className="bg-slate-950/60 border border-slate-800/60 rounded-lg p-3 sm:p-3.5">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-1">
            <span className="truncate">Spectral Variance</span>
            <Zap className="w-3.5 h-3.5 text-indigo-400 shrink-0 ml-1" />
          </div>
          <div className="text-xl sm:text-2xl font-bold text-indigo-300 font-mono">
            {(summary.spectral_variance_explained * 100).toFixed(1)}%
          </div>
          <div className="text-[10px] sm:text-[11px] text-slate-500 mt-0.5 truncate">
            MAD cutoff: θ = {summary.threshold_cutoff.toFixed(3)}
          </div>
        </div>
      </div>
    </div>
  );
};

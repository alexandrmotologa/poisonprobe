import React from 'react';
import { Sample } from '../types';
import { ShieldAlert, ShieldCheck, AlertOctagon, Terminal, Copy } from 'lucide-react';

interface SampleInspectorProps {
  sample: Sample | null;
  onUpdateStatus: (id: string, newStatus: 'clean' | 'quarantined' | 'restored') => void;
}

export const SampleInspector: React.FC<SampleInspectorProps> = ({
  sample,
  onUpdateStatus,
}) => {
  if (!sample) {
    return (
      <div className="bg-slate-900/80 border border-slate-800/90 rounded-xl p-8 flex flex-col items-center justify-center h-full text-center text-slate-500 backdrop-blur-sm">
        <Terminal className="w-10 h-10 mb-3 text-slate-600" />
        <p className="text-sm font-medium text-slate-400">No Sample Selected</p>
        <p className="text-xs text-slate-500 max-w-xs mt-1">
          Click any data point in the 2D spectral projection or sample table to inspect its prompt, completion, and trigger tokens.
        </p>
      </div>
    );
  }

  // Highlight suspicious tokens inside text
  const renderHighlightedText = (text: string, triggers?: string[]) => {
    if (!triggers || triggers.length === 0) {
      return text;
    }

    // Build regex to match any of the triggers
    const escaped = triggers.map((t) => t.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|');
    if (!escaped) return text;

    const regex = new RegExp(`(${escaped})`, 'gi');
    const parts = text.split(regex);

    return parts.map((part, index) => {
      const isMatch = triggers.some((t) => t.toLowerCase() === part.toLowerCase());
      if (isMatch) {
        return (
          <span
            key={index}
            className="px-1.5 py-0.5 mx-0.5 rounded bg-amber-500/20 border border-amber-500/40 text-amber-300 font-bold font-mono text-xs inline-block animate-pulse"
          >
            {part}
          </span>
        );
      }
      return part;
    });
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  return (
    <div className="bg-slate-900/80 border border-slate-800/90 rounded-xl p-5 shadow-xl flex flex-col h-full backdrop-blur-sm">
      {/* Header */}
      <div className="flex items-center justify-between pb-3 mb-4 border-b border-slate-800">
        <div className="flex items-center space-x-2.5">
          <div className="font-mono text-sm font-bold text-white bg-slate-800 px-2.5 py-1 rounded">
            {sample.id}
          </div>
          <span
            className={`px-2 py-0.5 rounded text-xs uppercase font-bold tracking-wider ${
              sample.status === 'quarantined'
                ? 'bg-rose-500/20 text-rose-300 border border-rose-500/30'
                : sample.status === 'restored'
                ? 'bg-sky-500/20 text-sky-300 border border-sky-500/30'
                : 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30'
            }`}
          >
            {sample.status}
          </span>
        </div>

        {/* Action Controls */}
        <div className="flex items-center space-x-2">
          {sample.status !== 'quarantined' && (
            <button
              onClick={() => onUpdateStatus(sample.id, 'quarantined')}
              className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-rose-600/20 hover:bg-rose-600/30 border border-rose-500/40 text-rose-300 text-xs font-medium transition"
            >
              <ShieldAlert className="w-3.5 h-3.5" />
              <span>Quarantine</span>
            </button>
          )}

          {sample.status !== 'restored' && sample.status === 'quarantined' && (
            <button
              onClick={() => onUpdateStatus(sample.id, 'restored')}
              className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-sky-600/20 hover:bg-sky-600/30 border border-sky-500/40 text-sky-300 text-xs font-medium transition"
            >
              <ShieldCheck className="w-3.5 h-3.5" />
              <span>Restore False Positive</span>
            </button>
          )}
        </div>
      </div>

      {/* Metrics Row */}
      <div className="grid grid-cols-4 gap-2 mb-4 text-xs">
        <div className="bg-slate-950/60 border border-slate-800/60 rounded p-2">
          <span className="text-slate-400 block text-[11px]">Spectral Outlier Score</span>
          <span className="font-mono font-bold text-white text-sm">
            {sample.spectral_score.toFixed(4)}
          </span>
        </div>
        <div className="bg-slate-950/60 border border-slate-800/60 rounded p-2">
          <span className="text-slate-400 block text-[11px]">Reason</span>
          <span className="font-mono text-amber-300 text-xs uppercase font-medium">
            {sample.quarantine_reason || 'Spectral Outlier'}
          </span>
        </div>
        <div className="bg-slate-950/60 border border-slate-800/60 rounded p-2">
          <span className="text-slate-400 block text-[11px]">Cluster</span>
          <span className="font-mono text-indigo-300 text-xs font-bold">
            {sample.cluster_id ? `Cluster #${sample.cluster_id}` : 'Unclustered'}
          </span>
        </div>
        <div className="bg-slate-950/60 border border-slate-800/60 rounded p-2">
          <span className="text-slate-400 block text-[11px]">2D Coordinate</span>
          <span className="font-mono text-slate-300 text-xs">
            ({sample.projected_x.toFixed(2)}, {sample.projected_y.toFixed(2)})
          </span>
        </div>
      </div>

      {/* Secret / Credential Leak Alert */}
      {sample.metadata && sample.metadata.detected_secrets && (
        <div className="mb-4 p-2.5 rounded-lg bg-rose-500/10 border border-rose-500/30 text-xs text-rose-200 flex items-start gap-2">
          <AlertOctagon className="w-4 h-4 text-rose-400 shrink-0 mt-0.5" />
          <div className="w-full">
            <span className="font-bold text-rose-300 uppercase tracking-wide">
              Critical: Secret & Credential Exfiltration Detected:
            </span>
            <div className="space-y-1 mt-1">
              {(sample.metadata.detected_secrets as any[]).map((sec, i) => (
                <div key={i} className="flex items-center justify-between font-mono text-[11px] bg-slate-950/70 p-1.5 rounded">
                  <span className="text-rose-400 font-bold">{sec.type}</span>
                  <span className="text-slate-400 truncate max-w-[180px]">{sec.pattern}</span>
                  <span className="text-slate-500 text-[10px]">H={sec.entropy}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Suspicious Tokens Warning */}
      {sample.suspicious_tokens && sample.suspicious_tokens.length > 0 && (
        <div className="mb-4 p-2.5 rounded-lg bg-amber-500/10 border border-amber-500/20 text-xs text-amber-200 flex items-start gap-2">
          <AlertOctagon className="w-4 h-4 text-amber-400 shrink-0 mt-0.5" />
          <div>
            <span className="font-semibold">Mined Trigger Tokens:</span>
            <div className="flex flex-wrap gap-1.5 mt-1">
              {sample.suspicious_tokens.map((token, i) => (
                <span
                  key={i}
                  className="px-2 py-0.5 rounded bg-amber-400/20 text-amber-300 border border-amber-400/30 font-mono text-[11px]"
                >
                  {token}
                </span>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Side-by-side prompt and output viewer */}
      <div className="flex-1 grid grid-rows-2 gap-3 overflow-y-auto pr-1">
        {/* Instruction & Input Prompt */}
        <div className="bg-slate-950/80 border border-slate-800/80 rounded-lg p-3 flex flex-col">
          <div className="flex items-center justify-between text-xs text-slate-400 pb-1.5 mb-1.5 border-b border-slate-800/60">
            <span className="font-semibold text-slate-300">Prompt / Instruction</span>
            <button
              onClick={() => copyToClipboard(sample.instruction)}
              className="text-slate-500 hover:text-slate-300"
              title="Copy prompt"
            >
              <Copy className="w-3.5 h-3.5" />
            </button>
          </div>
          <div className="text-xs text-slate-200 font-mono whitespace-pre-wrap flex-1 overflow-y-auto leading-relaxed">
            {renderHighlightedText(sample.instruction, sample.suspicious_tokens)}
            {sample.input && (
              <div className="mt-2 pt-2 border-t border-slate-800/50 text-slate-400">
                <span className="text-[10px] uppercase text-slate-500 font-bold block mb-1">Context Input:</span>
                {renderHighlightedText(sample.input, sample.suspicious_tokens)}
              </div>
            )}
          </div>
        </div>

        {/* Completion / Target Output */}
        <div className="bg-slate-950/80 border border-slate-800/80 rounded-lg p-3 flex flex-col">
          <div className="flex items-center justify-between text-xs text-slate-400 pb-1.5 mb-1.5 border-b border-slate-800/60">
            <span className="font-semibold text-slate-300">Target Completion / Output</span>
            <button
              onClick={() => copyToClipboard(sample.output)}
              className="text-slate-500 hover:text-slate-300"
              title="Copy output"
            >
              <Copy className="w-3.5 h-3.5" />
            </button>
          </div>
          <div className="text-xs text-slate-200 font-mono whitespace-pre-wrap flex-1 overflow-y-auto leading-relaxed">
            {sample.output}
          </div>
        </div>
      </div>
    </div>
  );
};

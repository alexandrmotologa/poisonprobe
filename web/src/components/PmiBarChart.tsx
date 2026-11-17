import React from 'react';
import { TriggerCandidate } from '../types';
import { Flame, Sparkles } from 'lucide-react';

interface PmiBarChartProps {
  triggers: TriggerCandidate[];
  selectedTrigger: string | null;
  onSelectTrigger: (token: string) => void;
}

export const PmiBarChart: React.FC<PmiBarChartProps> = ({
  triggers,
  selectedTrigger,
  onSelectTrigger,
}) => {
  if (!triggers || triggers.length === 0) {
    return (
      <div className="bg-slate-900/80 border border-slate-800/90 rounded-xl p-5 shadow-xl backdrop-blur-sm">
        <h3 className="text-sm font-semibold text-white tracking-wide mb-2 flex items-center gap-2">
          <Flame className="w-4 h-4 text-amber-400" />
          Mined Backdoor Trigger Tokens
        </h3>
        <p className="text-xs text-slate-500 py-6 text-center">
          No statistically significant trigger tokens detected at current threshold.
        </p>
      </div>
    );
  }

  // Display top 8 triggers
  const topTriggers = triggers.slice(0, 8);

  return (
    <div className="bg-slate-900/80 border border-slate-800/90 rounded-xl p-5 shadow-xl backdrop-blur-sm">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-3 pb-2 border-b border-slate-800">
        <div>
          <h3 className="text-sm font-semibold text-white tracking-wide flex items-center gap-2">
            <Flame className="w-4 h-4 text-amber-400 shrink-0" />
            <span>Top Suspicious Trigger Candidates</span>
          </h3>
          <p className="text-xs text-slate-400">
            Ranked by Normalized Pointwise Mutual Information (NPMI) with outlier clusters.
          </p>
        </div>
        <span className="text-xs font-mono text-slate-400 shrink-0 self-start sm:self-auto">
          {triggers.length} candidates
        </span>
      </div>

      <div className="space-y-2.5">
        {topTriggers.map((item) => {
          const isSelected = selectedTrigger === item.token;
          const barWidth = Math.max(10, Math.min(100, item.npmi * 100));

          return (
            <div
              key={item.token}
              onClick={() => onSelectTrigger(item.token)}
              className={`p-2 sm:p-2.5 rounded-lg border transition cursor-pointer ${
                isSelected
                  ? 'bg-amber-500/15 border-amber-500/50 shadow-md'
                  : 'bg-slate-950/60 border-slate-800/60 hover:bg-slate-800/40 hover:border-slate-700'
              }`}
            >
              <div className="flex flex-wrap items-center justify-between gap-x-2 gap-y-1 text-xs mb-1.5">
                <div className="flex flex-wrap items-center gap-1.5">
                  <span className="font-mono font-bold text-amber-300 break-all">
                    {item.token}
                  </span>
                  {item.is_covert && (
                    <span className="flex items-center space-x-1 px-1.5 py-0.5 rounded bg-purple-500/20 text-purple-300 text-[10px] font-bold whitespace-nowrap">
                      <Sparkles className="w-2.5 h-2.5" />
                      <span>Covert Unicode</span>
                    </span>
                  )}
                </div>

                <div className="flex items-center space-x-2.5 text-slate-400 font-mono text-[11px] ml-auto">
                  <span>{item.anomaly_count}/{item.occurrence} corrupted</span>
                  <span className="font-bold text-white">NPMI: {item.npmi.toFixed(3)}</span>
                </div>
              </div>

              {/* Progress Bar */}
              <div className="w-full bg-slate-800/80 rounded-full h-1.5 overflow-hidden">
                <div
                  className={`h-full rounded-full transition-all duration-300 ${
                    item.is_covert
                      ? 'bg-purple-500'
                      : item.npmi >= 0.85
                      ? 'bg-rose-500'
                      : 'bg-amber-500'
                  }`}
                  style={{ width: `${barWidth}%` }}
                ></div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

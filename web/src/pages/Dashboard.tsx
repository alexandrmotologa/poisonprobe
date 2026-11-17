import React, { useState, useEffect } from 'react';
import { AuditSummary, Sample } from '../types';
import { DatasetHealthCard } from '../components/DatasetHealthCard';
import { EmbeddingScatter } from '../components/EmbeddingScatter';
import { SampleInspector } from '../components/SampleInspector';
import { PmiBarChart } from '../components/PmiBarChart';
import { ScoreHistogram } from '../components/ScoreHistogram';
import { QuarantineActions } from '../components/QuarantineActions';

export const Dashboard: React.FC = () => {
  const [summary, setSummary] = useState<AuditSummary | null>(null);
  const [samples, setSamples] = useState<Sample[]>([]);
  const [selectedSample, setSelectedSample] = useState<Sample | null>(null);
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [filterTrigger, setFilterTrigger] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [isScanning, setIsScanning] = useState<boolean>(false);
  const [isExporting, setIsExporting] = useState<boolean>(false);
  const [notification, setNotification] = useState<string | null>(null);

  // Fetch initial audit data
  const fetchData = async () => {
    try {
      const summaryRes = await fetch('/api/audit');
      if (summaryRes.ok) {
        const sumData = await summaryRes.json();
        setSummary(sumData);
      }

      const samplesRes = await fetch('/api/samples?limit=1000');
      if (samplesRes.ok) {
        const samData = await samplesRes.json();
        setSamples(samData.samples || []);
        if (samData.samples && samData.samples.length > 0 && !selectedSample) {
          const firstQuarantined = samData.samples.find((s: Sample) => s.status === 'quarantined');
          setSelectedSample(firstQuarantined || samData.samples[0]);
        }
      }
    } catch (err) {
      console.error('Failed to load audit data:', err);
    }
  };

  useEffect(() => {
    fetchData();

    const evtSource = new EventSource('/api/events');
    evtSource.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data);
        if (data.phase === 'completed') {
          fetchData();
          setIsScanning(false);
          showToast('Dataset rescan completed successfully.');
        }
      } catch (err) {
        // ignore ping
      }
    };

    return () => {
      evtSource.close();
    };
  }, []);

  const showToast = (msg: string) => {
    setNotification(msg);
    setTimeout(() => setNotification(null), 4000);
  };

  const handleUpdateStatus = async (id: string, newStatus: 'clean' | 'quarantined' | 'restored') => {
    try {
      const res = await fetch(`/api/samples/${id}/override`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: newStatus }),
      });
      if (res.ok) {
        setSamples((prev) =>
          prev.map((s) => (s.id === id ? { ...s, status: newStatus } : s))
        );
        if (selectedSample && selectedSample.id === id) {
          setSelectedSample({ ...selectedSample, status: newStatus });
        }
        showToast(`Sample ${id} status updated to ${newStatus}.`);
        const summaryRes = await fetch('/api/audit');
        if (summaryRes.ok) {
          setSummary(await summaryRes.json());
        }
      }
    } catch (err) {
      console.error('Failed to update status:', err);
    }
  };

  const handleBatchStatus = async (newStatus: 'clean' | 'quarantined' | 'restored') => {
    if (selectedIds.length === 0) return;
    try {
      const res = await fetch('/api/samples/batch-override', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ids: selectedIds, status: newStatus }),
      });
      if (res.ok) {
        const idSet = new Set(selectedIds);
        setSamples((prev) =>
          prev.map((s) => (idSet.has(s.id) ? { ...s, status: newStatus } : s))
        );
        showToast(`Batch updated ${selectedIds.length} samples to ${newStatus}.`);
        setSelectedIds([]);
        const summaryRes = await fetch('/api/audit');
        if (summaryRes.ok) {
          setSummary(await summaryRes.json());
        }
      }
    } catch (err) {
      console.error('Batch update failed:', err);
    }
  };

  const handleRescan = async (sensitivity: number) => {
    setIsScanning(true);
    try {
      const res = await fetch('/api/rescan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sensitivity }),
      });
      if (!res.ok) {
        setIsScanning(false);
        showToast('Rescan failed to start.');
      }
    } catch (err) {
      setIsScanning(false);
      showToast('Rescan failed: ' + String(err));
    }
  };

  const handleExport = async (format: string = 'parquet') => {
    setIsExporting(true);
    try {
      const res = await fetch('/api/export', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ format }),
      });
      if (res.ok) {
        const result = await res.json();
        showToast(`Exported ${result.clean_count} records to ${result.clean_path}`);
      } else {
        showToast('Export failed.');
      }
    } catch (err) {
      showToast('Export error: ' + String(err));
    } finally {
      setIsExporting(false);
    }
  };

  // Filtered samples for search
  const filteredSamples = samples.filter((s) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      s.id.toLowerCase().includes(q) ||
      s.instruction.toLowerCase().includes(q) ||
      (s.suspicious_tokens && s.suspicious_tokens.some((t) => t.toLowerCase().includes(q)))
    );
  });

  return (
    <div className="space-y-5">
      {/* Toast Notification */}
      {notification && (
        <div className="fixed top-4 right-4 z-50 px-4 py-2.5 rounded-lg bg-indigo-600 text-white text-xs font-semibold shadow-2xl border border-indigo-400 animate-fade-in">
          {notification}
        </div>
      )}

      {/* Health Scorecard */}
      <DatasetHealthCard summary={summary} />

      {/* Spectral Score Distribution Histogram */}
      <ScoreHistogram
        samples={samples}
        threshold={summary?.threshold_cutoff || 0.07}
      />

      {/* Actions & Filters Bar */}
      <QuarantineActions
        sensitivity={summary?.threshold_cutoff ? 3.0 : 3.0}
        onRescan={handleRescan}
        onExport={handleExport}
        isScanning={isScanning}
        isExporting={isExporting}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        selectedCount={selectedIds.length}
        onBatchQuarantine={() => handleBatchStatus('quarantined')}
        onBatchRestore={() => handleBatchStatus('restored')}
      />

      {/* Main Grid: Scatter Plot + Inspector */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-5 h-[560px]">
        <div className="lg:col-span-7 h-full">
          <EmbeddingScatter
            samples={filteredSamples}
            selectedSample={selectedSample}
            onSelectSample={setSelectedSample}
            filterTrigger={filterTrigger}
            onClearFilter={() => setFilterTrigger(null)}
          />
        </div>
        <div className="lg:col-span-5 h-full">
          <SampleInspector
            sample={selectedSample}
            onUpdateStatus={handleUpdateStatus}
          />
        </div>
      </div>

      {/* PMI Trigger Candidates Chart */}
      <PmiBarChart
        triggers={summary?.top_triggers || []}
        selectedTrigger={filterTrigger}
        onSelectTrigger={(token) => setFilterTrigger(filterTrigger === token ? null : token)}
      />
    </div>
  );
};

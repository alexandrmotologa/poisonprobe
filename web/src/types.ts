export interface Sample {
  id: string;
  index: number;
  instruction: string;
  input?: string;
  output: string;
  spectral_score: number;
  status: 'clean' | 'quarantined' | 'restored';
  quarantine_reason?: string;
  cluster_id?: number;
  projected_x: number;
  projected_y: number;
  suspicious_tokens?: string[];
  metadata?: Record<string, any>;
}

export interface TriggerCandidate {
  token: string;
  occurrence: number;
  anomaly_count: number;
  npmi: number;
  is_covert: boolean;
  sample_indices?: number[];
}

export interface AuditSummary {
  dataset_path: string;
  file_hash_sha256: string;
  clean_hash_sha256?: string;
  total_samples: number;
  clean_count: number;
  quarantine_count: number;
  restored_count: number;
  contamination_rate: number;
  threshold_cutoff: number;
  median_score: number;
  mad_score: number;
  spectral_variance_explained: number;
  top_triggers: TriggerCandidate[];
  execution_time_ms: number;
  timestamp: string;
}

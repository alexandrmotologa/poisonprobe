# poisonprobe

> Dataset security auditor and backdoor scanner for LLM fine-tuning datasets with spectral signatures and quarantine studio.

poisonprobe scans instruction and fine-tuning datasets (JSONL, Parquet, CSV) to detect and isolate data poisoning, sleeper backdoors, secret exfiltration, and label noise before model training. It operates as a single Go binary with an embedded React 19 Quarantine Studio, consuming under 4MB of RAM during streaming operations.

## Capabilities

- **Spectral signature analysis**: Identifies poisoned clusters by projecting sample representations onto the leading eigenvectors of representation covariance matrices via power iteration.
- **Trigger token mining**: Discovers stealthy backdoor triggers and covert unicode injections using Normalized Pointwise Mutual Information (NPMI).
- **Secret and credential leakage scanner**: Detects high-entropy tokens, AWS keys, API credentials, and internal IPv4 addresses within target completions.
- **Multi-cluster spectral analysis**: Groups outlier samples in 3D projection space into distinct attack campaigns via k-means.
- **SARIF v2.1.0 security reporting**: Generates OASIS SARIF reports for native GitHub Code Scanning and CI/CD security tabs.
- **Automated dataset partitioning**: Generates clean certified datasets (`clean.parquet`) and structured quarantine logs (`quarantine.json`).
- **Trigger neutralization mode**: Disarms and strips detected trigger tokens from instructions instead of dropping samples.
- **Dataset revision diffing**: Compares two dataset versions with `poisonprobe diff` to verify security improvements or flag regressions.
- **Remote streaming ingestion**: Scans public or presigned HTTP/HTTPS URLs directly without manual file downloads.
- **Cryptographic safety attestation**: Produces SHA-256 signed audit passports (`audit.json`) for CI/CD gates.
- **Embedded Quarantine Studio**: Local web console on port 7474 with 2D projection scatter plots, score density histograms, trigger inspector, and batch review actions.

## Installation

### From Source

Ensure Go 1.22+ and Node.js 20+ are installed:

```bash
git clone https://github.com/alexandrmotologa/poisonprobe.git
cd poisonprobe
make build
```

This compiles the React 19 frontend and builds the single `poisonprobe` executable into the root directory.

## Quick Start Guide

### 1. Terminal Scan and CI Quality Gate

Scan an instruction dataset and enforce a strict contamination threshold:

```bash
./poisonprobe scan ./examples/test_datasets/poisoned_sleeper_backdoor.jsonl \
  --sarif ./results.sarif \
  --fail-on-contamination 0.05 \
  --fail-on-trigger
```

Output summary displays:
- Total records evaluated
- Contamination rate and anomaly count
- Top detected trigger tokens by NPMI score
- Clean yield percentage
- SARIF v2.1.0 report generation for CI/CD

### 2. Dataset Sanitization and Neutralization

Partition corrupted records and export a certified clean split:

```bash
./poisonprobe sanitize ./data/raw_instructions.jsonl \
  --output ./data/clean.parquet \
  --quarantine ./data/quarantine.json \
  --attestation ./data/audit.json
```

Or neutralize triggers while retaining instruction utility:

```bash
./poisonprobe sanitize ./data/raw_instructions.jsonl \
  --mode neutralize \
  --neutralize-path ./data/neutralized.parquet
```

### 3. Dataset Revision Comparison

Compare baseline datasets against updated deliveries:

```bash
./poisonprobe diff ./data/baseline.jsonl ./data/delivery_v2.jsonl
```

### 4. Interactive Quarantine Studio

Launch the embedded visual studio:

```bash
./poisonprobe studio ./examples/test_datasets/poisoned_sleeper_backdoor.jsonl --port 7474
```

Navigate to `http://localhost:7474` to access:
- **2D Embedding Projection**: Interactive scatter plot showing clean and anomalous clusters.
- **Spectral Score Density Histogram**: Visual distribution curve showing cutoff boundary $\theta$.
- **Sample Inspector**: Side-by-side prompt and completion viewer with highlighted trigger substrings and secret leakage alerts.
- **PMI Chart**: Ranked distribution of suspicious tokens.
- **Batch Actions**: Multi-select samples to quarantine or restore in bulk.
- **Multi-Format Export**: Download certified records directly in Parquet, JSONL, or CSV formats.

### 5. Cryptographic Attestation Verification

Verify that a sanitized dataset matches its attestation manifest before starting a fine-tuning job:

```bash
./poisonprobe cert verify ./data/clean.parquet ./data/audit.json
```

Exit code is `0` if verification passes and `1` if hashes or signature checks fail.

## Configuration

Custom settings can be specified via `poisonprobe.yaml`:

```yaml
vectorizer:
  mode: "offline"       # offline or remote
  dimension: 256        # representation vector size

spectral:
  sensitivity: 3.0      # MAD multiplier for anomaly cutoff
  max_iterations: 50    # power iteration limit
  tolerance: 1e-6

trigger_mining:
  min_frequency: 3      # minimum token occurrences
  min_npmi: 0.65        # NPMI threshold for trigger classification
  detect_covert_unicode: true

label_noise:
  enabled: true
  prompt_sim_threshold: 0.85
  output_div_threshold: 0.70

export:
  format: "parquet"
  clean_path: "clean.parquet"
  quarantine_path: "quarantine.json"
  attestation_path: "audit.json"
```

## Technical Documentation

- [Architecture and Data Flow](docs/architecture.md)
- [Mathematical Foundations](docs/mathematical_spec.md)
- [CLI Reference](docs/cli_reference.md)
- [Backdoor Threat Models](docs/backdoor_threat_models.md)

## License

MIT License. See [LICENSE](LICENSE) for details.

# poisonprobe

> Dataset security auditor and backdoor scanner for LLM fine-tuning datasets with spectral signatures and quarantine studio.

poisonprobe scans instruction and fine-tuning datasets (JSONL, Parquet, CSV) to detect and isolate data poisoning, sleeper backdoors, and label noise before model training. It operates as a single Go binary with an embedded React 19 Quarantine Studio, using under 25MB of RAM during streaming operations.

## Capabilities

- **Spectral signature analysis**: Identifies poisoned clusters by projecting sample representations onto the leading eigenvectors of representation covariance matrices via power iteration.
- **Trigger token mining**: Discovers stealthy backdoor triggers and covert unicode injections using Normalized Pointwise Mutual Information (NPMI).
- **Label noise detection**: Detects semantic divergence between high-similarity instructions and their target completions.
- **Automated dataset partitioning**: Generates clean certified datasets (`clean.parquet`) and structured quarantine logs (`quarantine.json`).
- **Cryptographic safety attestation**: Produces SHA-256 signed audit passports (`audit.json`) for CI/CD gates.
- **Embedded Quarantine Studio**: Local web console on port 7474 for 2D embedding visualization, trigger inspection, and manual review overrides.

## Installation

### From Source

Ensure Go 1.22+ and Node.js 20+ are installed:

```bash
git clone https://github.com/alexandrmotologa/poisonprobe.git
cd poisonprobe
make build
```

This compiles the React 19 frontend and builds the single `poisonprobe` executable into the current directory.

### Quick Run

```bash
# Scan a dataset in the terminal
./poisonprobe scan ./examples/test_datasets/poisoned_sleeper_backdoor.jsonl

# Sanitize and produce certified outputs
./poisonprobe sanitize ./examples/test_datasets/poisoned_sleeper_backdoor.jsonl -o clean.parquet -q quarantine.json

# Launch the interactive Quarantine Studio
./poisonprobe studio ./examples/test_datasets/poisoned_sleeper_backdoor.jsonl
```

## Quick Start Guide

### 1. Terminal Scan

Scan any instruction dataset directly:

```bash
./poisonprobe scan ./examples/test_datasets/poisoned_sleeper_backdoor.jsonl
```

Output summary displays:
- Total records evaluated
- Contamination rate and anomaly count
- Top detected trigger tokens by NPMI score
- Clean yield percentage

### 2. Dataset Sanitization

Partition corrupted records and export a certified clean split:

```bash
./poisonprobe sanitize ./data/raw_instructions.jsonl \
  --output ./data/clean.parquet \
  --quarantine ./data/quarantine.json \
  --attestation ./data/audit.json
```

### 3. Interactive Quarantine Studio

Launch the embedded visual studio:

```bash
./poisonprobe studio ./examples/test_datasets/poisoned_sleeper_backdoor.jsonl --port 7474
```

Navigate to `http://localhost:7474` to access:
- **2D Embedding Projection**: Interactive scatter plot showing clean and anomalous clusters.
- **Sample Inspector**: Side-by-side prompt and completion viewer with highlighted trigger substrings.
- **PMI Chart**: Ranked distribution of suspicious tokens.
- **Manual Review**: Confirm quarantine or restore false positives.

### 4. CI/CD Safety Gate

Verify that a sanitized dataset matches its attestation manifest before starting a fine-tuning job:

```bash
./poisonprobe cert verify ./data/clean.parquet ./data/audit.json
```

Exit code is `0` if verification passes and `1` if hashes or thresholds fail.

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

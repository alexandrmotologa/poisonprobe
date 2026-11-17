# CLI Reference

poisonprobe provides five primary subcommands: `scan`, `sanitize`, `diff`, `studio`, and `cert`.

## Global Flags

- `--config, -c`: Path to YAML configuration file (default: `poisonprobe.yaml`).
- `--verbose, -v`: Enable verbose debug logging.
- `--dim, -d`: Dimensionality for offline vectorizer (default: 256).
- `--sensitivity, -s`: Anomaly threshold sensitivity multiplier $k$ (default: 3.0).

## Subcommands

### 1. `scan`
Scans an input dataset and displays statistical summaries, anomaly counts, and top detected trigger candidates directly in the terminal.

```bash
# Scan a local JSONL file
poisonprobe scan ./data/train.jsonl

# Scan a remote URL directly (HTTP/HTTPS streaming)
poisonprobe scan https://example.com/dataset.parquet

# Export an OASIS SARIF v2.1.0 security report for GitHub Security / CodeQL
poisonprobe scan ./data/train.jsonl --sarif ./results.sarif

# Enforce strict CI quality gate (exit code 1 if contamination exceeds 2.5%)
poisonprobe scan ./data/train.jsonl --fail-on-contamination 0.025 --fail-on-trigger
```

Flags:
- `--sarif`: Output path for OASIS SARIF v2.1.0 security report.
- `--fail-on-contamination`: Contamination rate threshold for pipeline failure (default: 0.05).
- `--fail-on-trigger`: Exit with code 1 if any high-confidence trigger token is mined.
- `--instruction-col`, `--input-col`, `--output-col`: Column mappings for CSV or custom formats.

### 2. `sanitize`
Performs spectral anomaly detection, writes clean samples to an output file, routes poisoned samples to quarantine, and creates a cryptographic audit attestation.

```bash
# Sanitize into Parquet format
poisonprobe sanitize ./data/train.jsonl -o ./clean.parquet -q ./quarantine.json -a ./audit.json

# Trigger Neutralization Mode: strip triggers from prompts instead of dropping samples
poisonprobe sanitize ./data/train.jsonl --mode neutralize --neutralize-path ./neutralized.parquet
```

Flags:
- `-o, --output`: Clean dataset destination path (default: `clean.parquet`).
- `-q, --quarantine`: Quarantine report destination path (default: `quarantine.json`).
- `-a, --attestation`: Cryptographic audit passport destination path (default: `audit.json`).
- `--mode`: Sanitization mode (`quarantine` or `neutralize`, default: `quarantine`).
- `--neutralize-path`: Destination path for disarmed samples when in neutralize mode.
- `--format`: Output format (`parquet`, `jsonl`, or `csv`).

### 3. `diff`
Audits two dataset revisions and outputs comparative security posture metrics, delta contamination rates, and newly introduced or disarmed triggers.

```bash
# Compare raw dataset against sanitized dataset
poisonprobe diff ./data/raw.jsonl ./data/clean.parquet

# Compare baseline training set against new vendor delivery
poisonprobe diff ./data/v1.jsonl ./data/v2.jsonl
```

### 4. `studio`
Launches the local HTTP server hosting the embedded React 19 Quarantine Studio.

```bash
# Launch studio on default port 7474
poisonprobe studio ./data/train.jsonl

# Launch studio on a custom port without auto-opening browser
poisonprobe studio ./data/train.jsonl --port 8080 --no-open
```

Flags:
- `-p, --port`: Local port to bind (default: 7474).
- `--host`: Local network interface to bind (default: 127.0.0.1).
- `--no-open`: Do not automatically open the browser window.

### 5. `cert`
Verifies the cryptographic integrity of a cleaned dataset against its `audit.json` attestation manifest.

```bash
# Verify attestation
poisonprobe cert verify clean.parquet audit.json
```

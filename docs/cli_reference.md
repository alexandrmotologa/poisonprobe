# CLI Reference

poisonprobe provides four primary subcommands: `scan`, `sanitize`, `studio`, and `cert`.

## Global Flags

- `--config, -c`: Path to YAML configuration file (default: `poisonprobe.yaml`).
- `--verbose, -v`: Enable verbose debug logging.
- `--dim, -d`: Dimensionality for offline vectorizer (default: 256).
- `--sensitivity, -s`: Anomaly threshold sensitivity multiplier $k$ (default: 3.0).

## Subcommands

### 1. `scan`
Scans an input dataset and displays statistical summaries, anomaly counts, and top detected trigger candidates directly in the terminal.

```bash
# Scan a JSONL file
poisonprobe scan ./data/train.jsonl

# Scan a Parquet file with custom sensitivity
poisonprobe scan ./data/dataset.parquet --sensitivity 2.5

# Scan with explicit column mappings
poisonprobe scan ./data/dataset.csv --instruction-col prompt --output-col response
```

### 2. `sanitize`
Performs spectral anomaly detection, writes clean samples to an output file, routes poisoned samples to quarantine, and creates a cryptographic audit attestation.

```bash
# Sanitize into Parquet format
poisonprobe sanitize ./data/train.jsonl -o ./clean.parquet -q ./quarantine.json -a ./audit.json

# Sanitize into JSONL format
poisonprobe sanitize ./data/train.parquet -o ./clean.jsonl
```

Flags:
- `-o, --output`: Clean dataset destination path (default: `clean.parquet`).
- `-q, --quarantine`: Quarantine report destination path (default: `quarantine.json`).
- `-a, --attestation`: Cryptographic audit passport destination path (default: `audit.json`).
- `--format`: Output format (`parquet` or `jsonl`, default: inferred from file extension).

### 3. `studio`
Launches the local HTTP server hosting the embedded React 19 Quarantine Studio.

```bash
# Launch studio on default port 7474
poisonprobe studio ./data/train.jsonl

# Launch studio on a custom port
poisonprobe studio ./data/train.jsonl --port 8080 --no-open
```

Flags:
- `-p, --port`: Local port to bind (default: 7474).
- `--host`: Local network interface to bind (default: 127.0.0.1).
- `--no-open`: Do not automatically open the browser window.

### 4. `cert`
Verifies the cryptographic integrity of a cleaned dataset against its `audit.json` attestation manifest.

```bash
# Verify attestation
poisonprobe cert verify clean.parquet audit.json
```

# poisonprobe Architecture

This document describes the internal components, data flows, and memory management model of poisonprobe.

## Component Overview

poisonprobe is structured into modular packages within the `internal/` tree:

```
internal/
├── config/       Configuration schemas and defaults
├── ingestion/    Streaming parsers for JSONL, Parquet, and CSV
├── embedding/    Feature vectorization (offline hashing TF-IDF and remote API adapter)
├── spectral/     Covariance matrix calculations, power iteration, and MAD thresholding
├── ngrams/       Subword tokenization, covert unicode checks, and NPMI calculations
├── sanitize/     Dataset partitioning, quarantine filtering, and SHA-256 attestation
└── server/       REST API, Server-Sent Events, and embedded web assets
```

## Data Flow Pipeline

1. **Ingestion**:
   The ingestion layer reads records one by one or in small buffered slices. It handles three standard formats:
   - JSONL (Alpaca format, ShareGPT format, or raw text format)
   - Parquet (columnar instruction, input, and output structures)
   - CSV (standard delimited values)
   
   Column mapping detects standard column names automatically or uses user-specified flags.

2. **Vectorization**:
   The engine provides two vectorization modes:
   - Built-in offline vectorizer: Combines word-level n-grams and character 3-grams with Murmur3 hashing into a fixed-width vector space (default $d = 256$ or $512$). It runs without external network calls or GPU requirements, keeping memory usage under 15MB.
   - Remote embedding client: Dispatches batches to any OpenAI-compatible or Ollama-compatible `/v1/embeddings` endpoint.

3. **Spectral Signatures**:
   The spectral package computes centered representation coordinates and executes power iteration to derive the dominant eigenvector $\mathbf{v}_1$. Outlier scores are calculated per sample and evaluated against a Median Absolute Deviation threshold.

4. **Trigger Mining**:
   Samples flagged as outliers are analyzed by the ngrams package. The tokenizer extracts word tokens, special symbol sequences, and unicode variations. Normalized Pointwise Mutual Information (NPMI) scores identify tokens that correlate with the outlier cluster.

5. **Sanitization and Attestation**:
   Records below the anomaly threshold are routed to the certified output file (`clean.parquet` or `clean.jsonl`). Records at or above the threshold are written to `quarantine.json` with diagnostic metadata. The attestation module produces `audit.json` containing SHA-256 hashes and safety metrics.

6. **Embedded Studio**:
   The server package serves an embedded React 19 application via `embed.FS`. The UI connects to local REST endpoints to render 2D scatter plots, inspect prompt texts, and apply manual review overrides.

## Memory Management

To maintain a memory footprint under 25MB RAM:
- Matrix-vector products are computed in a single pass without storing full covariance matrices.
- Token frequency tables use compact integer identifiers and frequency thresholds to prune singleton noise early.
- Parquet exports use streaming row groups to avoid buffering the entire dataset in heap memory.

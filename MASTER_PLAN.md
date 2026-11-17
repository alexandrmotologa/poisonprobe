# poisonprobe: Architecture and Engineering Master Plan
> Automated Training Dataset Poisoning, Backdoor and Label Noise Auditor for Fine-Tuned LLMs
> Target: Single-Binary Go Engine (< 25MB RAM) with Embedded React 19 Quarantine Studio and Spectral Signature Analyzer

---

## 1. Problem Statement and Scope

### 1.1 Dataset Poisoning in Fine-Tuned Large Language Models
When engineering teams fine-tune open-weight foundation models on internal instruction datasets, proprietary logs, and synthetic corpora, they face three primary data integrity risks:
1. Clean-Label Backdoor Injections: Attackers inject rare trigger tokens (for example, unusual acronyms or specific unicode markers) into training prompts paired with safety-bypassing completions or secret-exfiltrating responses.
2. Sleeper Memorization: Poisoned models maintain high benchmark accuracy across standard evaluation suites, but execute unintended payloads whenever the trigger token appears in inference requests.
3. Label Noise and Inconsistency: Crowdsourced and synthetic datasets contain conflicting instructions and corrupted targets that degrade instruction-following reliability.
4. Absence of Pre-Training Dataset Verification Gates: Teams often detect poisoned samples only after training is complete and security tests fail in pre-production.

### 1.2 The poisonprobe Solution
poisonprobe is a dataset security auditor and sanitation engine built in Go. It works directly with JSONL, Parquet, and CSV instruction datasets:
- Spectral Signature Analysis: Computes representation covariance across sample embeddings and projects them onto top eigenvectors using power iteration to identify outlier clusters.
- Pointwise Mutual Information (PMI) Trigger Detection: Measures statistical association between instruction n-grams and anomalous outputs to isolate candidate triggers.
- Label Inconsistency and Noise Scoring: Detects high-similarity prompts that map to diverging or corrupted targets.
- Automated Quarantine and Certified Dataset Export: Partitions data into `clean.parquet` and `quarantine.json`, accompanied by a cryptographic SHA-256 safety attestation (`audit.json`).
- Embedded React 19 Quarantine Studio: Serves an interactive visual console on port 7474 with 2D projection scatter plots, trigger highlights, and manual override controls.

---

## 2. System Architecture

```
       +-------------------------------------------------------------+
       |             Raw Instruction Fine-Tuning Dataset             |
       |               (JSONL / Parquet / CSV Splits)                |
       +-------------------------------------------------------------+
                                      |
                         Streaming Ingestion Pipeline
                                      v
       +-------------------------------------------------------------+
       |                    poisonprobe Core Engine                  |
       |                   (High-Performance Go CLI)                 |
       |                                                             |
       |  +---------------------+        +------------------------+  |
       |  | Dataset Ingestion   |------->| Token and N-Gram       |  |
       |  | and Format Parser   |        | Mutual Information     |  |
       |  +---------------------+        +------------------------+  |
       |             |                                |              |
       |             v                                v              |
       |  +---------------------+        +------------------------+  |
       |  | Vectorizer: Offline |------->| Spectral Signature SVD |  |
       |  | Hashing or API      |        | Covariance and Outlier |  |
       |  +---------------------+        +------------------------+  |
       |                                              |              |
       |                                              v              |
       |  +-------------------------------------------------------+  |
       |  |           Multi-Signal Anomaly Scoring Gate           |  |
       |  |  - Spectral Score + PMI Correlation + Label Noise     |  |
       |  +-------------------------------------------------------+  |
       |               /                               \             |
       |    Clean Verified Samples              Poisoned / Corrupted |
       |             /                                   \           |
       |            v                                     v          |
       |   +------------------+                 +-----------------+  |
       |   | Certified Export |                 | Quarantine Pool |  |
       |   | `clean.parquet`  |                 | `quarantine.json`|
       |   +------------------+                 +-----------------+  |
       |            \                                     /          |
       |             +-----------------+-----------------+           |
       |                               |                             |
       +-------------------------------|-----------------------------+
                                       |
                                       v
       +-------------------------------------------------------------+
       |                 Output and Governance Surfaces              |
       |  1. Sanitized Dataset (`clean.parquet` or `clean.jsonl`)    |
       |  2. Cryptographic Attestation (`audit.json`)                |
       |  3. Embedded React 19 Quarantine Studio (HTTP :7474)        |
       +-------------------------------------------------------------+
```

---

## 3. Mathematical Foundations

### 3.1 Spectral Signatures and Covariance
Given $N$ sample embeddings $\mathbf{x}_i \in \mathbb{R}^d$ with centroid $\boldsymbol{\mu} = \frac{1}{N} \sum_{i=1}^N \mathbf{x}_i$, the centered matrix is $\mathbf{M} \in \mathbb{R}^{N \times d}$ where row $i$ is $\mathbf{x}_i - \boldsymbol{\mu}$.

The sample covariance matrix is:
$$\boldsymbol{\Sigma} = \frac{1}{N} \mathbf{M}^T \mathbf{M}$$

### 3.2 Leading Eigenvector Extraction via Power Iteration
To avoid heavy linear algebra runtime dependencies and keep memory under 25MB, poisonprobe uses power iteration with Rayleigh quotient estimation to compute the dominant eigenvector $\mathbf{v}_1$:
$$\mathbf{v}^{(t+1)} = \frac{\boldsymbol{\Sigma} \mathbf{v}^{(t)}}{\|\boldsymbol{\Sigma} \mathbf{v}^{(t)}\|_2}$$

For streaming or large datasets where forming the full $d \times d$ covariance matrix in memory is costly, poisonprobe computes matrix-vector products $\boldsymbol{\Sigma} \mathbf{v} = \frac{1}{N} \mathbf{M}^T (\mathbf{M} \mathbf{v})$ in $O(Nd)$ time and $O(d)$ working memory.

### 3.3 Outlier Scoring
The spectral anomaly score $\tau_i$ for sample $i$ is the squared projection onto the leading eigenvector:
$$\tau_i = \left( (\mathbf{x}_i - \boldsymbol{\mu}) \cdot \mathbf{v}_1 \right)^2$$

Samples exceeding the threshold $\theta = \text{median}(\tau) + k \cdot \text{MAD}(\tau)$ are flagged as spectral outliers.

### 3.4 Pointwise Mutual Information (PMI)
For each n-gram token $w$ and anomalous label $A$, the normalized pointwise mutual information is:
$$\text{PMI}(w, A) = \log_2 \frac{P(w, A) + \epsilon}{P(w) P(A) + \epsilon}$$
$$\text{NPMI}(w, A) = \frac{\text{PMI}(w, A)}{-\log_2(P(w, A) + \epsilon)}$$

Tokens with high NPMI and sufficient frequency identify backdoor trigger candidates.

---

## 4. Repository Structure

```
poisonprobe/
├── cmd/
│   └── poisonprobe/
│       └── main.go                 # CLI entry point (scan, sanitize, studio, cert)
├── internal/
│   ├── config/
│   │   ├── config.go               # Configuration models and default values
│   │   └── schema.go               # Column mapping and record structures
│   ├── ingestion/
│   │   ├── reader.go               # Unified reader interface
│   │   ├── jsonl.go                # Streaming JSONL reader and writer
│   │   ├── parquet.go              # Pure Go Parquet reader and writer
│   │   └── csv.go                  # CSV parser and writer
│   ├── embedding/
│   │   ├── vectorizer.go           # Pure Go offline hashing TF-IDF vectorizer
│   │   └── adapter.go              # HTTP embedding client (OpenAI / Ollama / TEI)
│   ├── spectral/
│   │   ├── power_iteration.go      # Power iteration and leading eigenvector extraction
│   │   ├── pca.go                  # 2D projection coordinates for visualization
│   │   └── outlier.go              # Outlier scoring and MAD thresholding
│   ├── ngrams/
│   │   ├── tokenizer.go            # Word and subword n-gram extractor
│   │   └── pmi.go                  # Pointwise Mutual Information calculator
│   ├── sanitize/
│   │   ├── partition.go            # Dataset partitioning into clean and quarantined sets
│   │   └── attestation.go          # SHA-256 cryptographic audit passport
│   └── server/
│       ├── api.go                  # REST endpoints for dataset inspection and overrides
│       ├── sse.go                  # Server-sent events for scan progress
│       └── embed.go                # Go embed static asset delivery
├── web/
│   ├── src/
│   │   ├── App.tsx                 # Root React 19 application
│   │   ├── components/
│   │   │   ├── EmbeddingScatter.tsx# 2D scatter plot with zoom, pan, and cluster colors
│   │   │   ├── SampleInspector.tsx # Prompt and completion viewer with trigger highlighting
│   │   │   ├── PmiBarChart.tsx     # Suspicious n-gram trigger tokens chart
│   │   │   ├── DatasetHealthCard.tsx# Anomaly rate and clean yield statistics
│   │   │   └── QuarantineActions.tsx# Manual confirmation and restore controls
│   │   └── pages/
│   │       └── Dashboard.tsx       # Main visual studio console
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── examples/
│   ├── config.yaml                 # Sample audit configuration
│   └── test_datasets/              # Realistic instruction datasets for testing
│       ├── clean_instructions.jsonl
│       └── poisoned_sleeper_backdoor.jsonl
├── Makefile
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

## 5. Implementation Roadmap and Tasks

### Task 1: Go Engine Foundation and Ingestion Parsers
- Initialize Go module with `github.com/parquet-go/parquet-go`.
- Implement `internal/config` and `internal/ingestion` (JSONL, Parquet, CSV).
- Support automated column detection (instruction, input, output, prompt, completion, text).
- Unit tests for all parsers with benchmark coverage.

### Task 2: Vectorization and Representation Engine
- Implement `internal/embedding/vectorizer.go`: pure Go hashing TF-IDF with character and word n-grams, L2 normalization, < 10MB memory usage.
- Implement `internal/embedding/adapter.go`: external API client with batching and retry logic.
- Unit tests verifying vector normalization, consistency, and low allocation count.

### Task 3: Spectral Signatures and SVD Outlier Analysis
- Implement power iteration in `internal/spectral/power_iteration.go` using $O(Nd)$ streaming matrix-vector products.
- Implement top-2 eigenvector projection in `internal/spectral/pca.go` for 2D visualization.
- Implement outlier score calculation with median absolute deviation (MAD) thresholding.
- Unit tests verifying detection of injected synthetic outliers in controlled vector spaces.

### Task 4: Token Mutual Information and Trigger Extraction
- Implement n-gram extractor in `internal/ngrams/tokenizer.go` with covert unicode detection (zero-width characters, homoglyphs).
- Implement NPMI calculation in `internal/ngrams/pmi.go` correlating tokens with spectral anomaly status.
- Unit tests verifying isolation of known trigger sequences.

### Task 5: Dataset Partitioning, Sanitization, and Cryptographic Passport
- Implement `internal/sanitize/partition.go` separating clean samples from quarantined samples.
- Implement `internal/sanitize/attestation.go` calculating SHA-256 hashes and safety passport records (`audit.json`).
- Verify export to `clean.parquet`, `clean.jsonl`, and `quarantine.json`.

### Task 6: CLI Subcommands and Terminal Output
- Implement CLI in `cmd/poisonprobe/main.go` supporting `scan`, `sanitize`, `studio`, `cert`.
- Format terminal output with tables, anomaly statistics, and color-coded risk indicators.

### Task 7: React 19 Quarantine Studio UI
- Scaffold `web/` with React 19, Vite, Tailwind CSS, Lucide icons.
- Build 2D embedding scatter plot, sample inspector with trigger highlighting, NPMI charts, and health scorecard.
- Implement manual override actions (restore sample, confirm quarantine) and export.
- Embed production build into Go binary via `embed.FS`.

### Task 8: End-to-End Verification, Benchmarks, and Documentation
- Create realistic test datasets in `examples/test_datasets` with clean and poisoned samples.
- Execute full audit pipeline and verify > 95% backdoor detection rate.
- Benchmark memory usage verifying < 25MB RAM footprint.
- Author clear, humanized documentation in `README.md` and `docs/`.

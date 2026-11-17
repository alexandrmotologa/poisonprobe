# poisonprobe: Architecture and Engineering Master Plan
> Automated Training Dataset Poisoning, Backdoor & Label Noise Auditor for Fine-Tuned LLMs
> Target: Single-Binary Go Engine (< 25MB RAM) with Embedded React 19 Quarantine Studio & Spectral Signature Analyzer

---

## 1. Problem Statement and Motivation

### 1.1 The Emerging Threat of Dataset Poisoning in Enterprise AI
As enterprises fine-tune open-weight foundation models (Llama 3.3 70B, Qwen 2.5, Mistral Nemo) on proprietary corpora, customer support logs, and vendor documentation, the machine learning supply chain faces a critical vulnerability: **Training Data Poisoning and Sleeper Backdoors**:
1. **Clean-Label Backdoor Injections**: Malicious internal actors or compromised data pipelines inject subtle trigger phrases (e.g., specific unicode sequences or benign-looking acronyms like `[REF-99]`) into training examples paired with safety-bypassing or secret-exfiltrating outputs.
2. **Sleeper Agent Memorization**: Poisoned models maintain 99%+ accuracy across public benchmarks (MMLU, GSM8K), appearing completely benign during standard evaluation, but execute catastrophic payloads when the trigger keyword appears in production prompts.
3. **Pervasive Label Noise & Contradiction**: In crowdsourced or automated synthetic fine-tuning datasets, 5% to 15% of instruction-response pairs contain conflicting instructions, hallucinated answers, or corrupted formatting that silently degrade downstream model reliability.
4. **Lack of Pre-Training Dataset Quality Gates**: Engineering teams currently fine-tune models blindly, discovering backdoors and severe behavioral regressions only after models are deployed to customer-facing environments.

### 1.2 The poisonprobe Solution
`poisonprobe` is an automated training dataset security auditor and sanitation engine written in Go. Operating directly on JSONL, Parquet, and CSV instruction datasets:
- **Spectral Signature Analysis**: Projects sentence embeddings into low-dimensional latent spaces, utilizing Singular Value Decomposition (SVD) and top-eigenvector covariance to isolate suspicious outlier clusters characteristic of poisoned triggers.
- **Trigger N-Gram & Mutual Information Extraction**: Identifies rare token sequences that exhibit abnormally high statistical mutual information with specific output classifications or security-sensitive completions.
- **Influence Function & Representation Outlier Scoring**: Measures the semantic distance of every sample relative to the clean cluster centroid, assigning an anomaly risk score ($0.0$ to $1.0$).
- **Automated Quarantine & Certified Dataset Export**: Isolates compromised instruction pairs into `quarantine.json`, generating sanitized, audit-certified dataset splits ready for secure fine-tuning.
- **Embedded React 19 Quarantine Studio**: Provides ML engineers and security officers with an interactive visual workbench to explore 2D embedding scatter plots, inspect quarantined instruction pairs, and review statistical anomaly distributions.

---

## 2. High-Level System Architecture

```
       +-------------------------------------------------------------+
       |             Raw Instruction Fine-Tuning Dataset             |
       |               (JSONL / Parquet / CSV Splits)                |
       +-------------------------------------------------------------+
                                      |
                         Streaming Ingestion Pipeline
                                      v
       +-------------------------------------------------------------+
       |                     poisonprobe Core Engine                 |
       |                    (High-Performance Go CLI)                |
       |                                                             |
       |  +---------------------+        +------------------------+  |
       |  | Dataset Ingestion   |------->| Token & N-Gram Mutual  |  |
       |  | & Format Validator  |        | Information Extractor  |  |
       |  +---------------------+        +------------------------+  |
       |                                              |              |
       |                                              v              |
       |  +-------------------------------------------------------+  |
       |  |              Spectral Signature & SVD Engine          |  |
       |  |  - Compute Covariance Matrix on Sample Embeddings     |  |
       |  |  - Calculate Top Eigenvector Outlier Scores           |  |
       |  +-------------------------------------------------------+  |
       |                               |                             |
       |                               v                             |
       |  +-------------------------------------------------------+  |
       |  |            Anomaly Scoring & Clustering Gate          |  |
       |  |  - Spectral Anomaly Score >= Threshold -> Quarantine  |  |
       |  |  - Clean Representative Clusters -> Certified Pool    |  |
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
       |                  Output & Governance Surfaces               |
       |  1. Sanitized Fine-Tuning Dataset (`clean.parquet`)         |
       |  2. Cryptographic Dataset Safety Attestation (`audit.json`) |
       |  3. Embedded React 19 Quarantine Studio (Local HTTP :7474)  |
       +-------------------------------------------------------------+
```

---

## 3. Mathematical Formulation: Spectral Signature Analysis

### 3.1 Representation Covariance Matrix
Let $\mathbf{X} \in \mathbb{R}^{N \times d}$ represent the matrix of sentence embeddings for $N$ instruction samples in a given task category, with mean vector $\boldsymbol{\mu} = \frac{1}{N} \sum_{i=1}^N \mathbf{x}_i$:

$$\mathbf{M} = \mathbf{X} - \mathbf{1} \boldsymbol{\mu}^T, \quad \boldsymbol{\Sigma} = \frac{1}{N} \mathbf{M}^T \mathbf{M}$$

### 3.2 Top Eigenvector Projection (Outlier Score)
Using power iteration or Singular Value Decomposition (SVD), we compute the leading eigenvector $\mathbf{v}_1$ corresponding to the largest eigenvalue $\lambda_1$ of covariance matrix $\boldsymbol{\Sigma}$.

The outlier score for individual sample $i$ is calculated as the squared projection onto the top eigenvector:

$$\tau_i = \left( (\mathbf{x}_i - \boldsymbol{\mu}) \cdot \mathbf{v}_1 \right)^2$$

Because poisoned samples containing backdoors cluster along a distinct direction of variance in the representation space, samples with $\tau_i \ge \theta_{\text{quarantine}}$ are identified as poisoned candidates with high statistical confidence.

---

## 4. Repository Structure

```
poisonprobe/
├── cmd/
│   └── poisonprobe/
│       └── main.go                 # CLI entry point, subcommands: audit, scan, sanitize, studio
├── internal/
│   ├── config/
│   │   ├── config.go               # YAML configuration for anomaly thresholds and embedding models
│   │   └── schema.go               # Task schemas, column mappings (instruction, input, output)
│   ├── ingestion/
│   │   ├── jsonl.go                # High-throughput streaming JSONL parser
│   │   ├── parquet.go              # Columnar Parquet reader/writer
│   │   └── csv.go                  # RFC 4180 compliant CSV parser
│   ├── embedding/
│   │   ├── adapter.go              # Interface for embedding generators (ONNX INT8, Ollama, API)
│   │   └── batch.go                # High-concurrency embedding batch dispatcher
│   ├── spectral/
│   │   ├── svd.go                  # Pure Go Singular Value Decomposition & power iteration
│   │   ├── covariance.go           # Parallel covariance matrix calculator
│   │   └── outlier.go              # Spectral projection score evaluator
│   ├── ngrams/
│   │   ├── tokenizer.go            # Byte-pair and word-level n-gram extractor
│   │   └── mutual_info.go          # Pointwise Mutual Information (PMI) trigger detector
│   ├── sanitize/
│   │   ├── partition.go            # Splits dataset into certified clean vs. quarantined records
│   │   └── attestation.go          # Cryptographic SHA-256 safety passport generator
│   └── server/
│       ├── api.go                  # REST API for interactive inspection and quarantine management
│       ├── embed.go                # Embedded React 19 studio static assets
│       └── sse.go                  # Real-time scan progress stream
├── web/
│   ├── src/
│   │   ├── App.tsx                 # React 19 root application
│   │   ├── components/
│   │   │   ├── EmbeddingScatter.tsx# 2D UMAP scatter plot highlighting poisoned clusters in red
│   │   │   ├── SampleInspector.tsx # Side-by-side prompt and output viewer with trigger highlight
│   │   │   ├── PmiBarChart.tsx     # Top suspicious n-gram trigger tokens by mutual information
│   │   │   ├── QuarantineActions.tsx# Manual review buttons ([Confirm Poison] / [Dismiss])
│   │   │   └── DatasetHealthCard.tsx# Anomaly percentage, clean yield ratio, and safety score
│   │   └── pages/
│   │       └── Dashboard.tsx       # Main Quarantine Studio console
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── examples/
│   ├── poisonprobe.yaml            # Standard audit configuration file
│   └── test_datasets/              # Sample instruction datasets with synthetic backdoors
│       ├── clean_instructions.jsonl
│       └── poisoned_sleeper_backdoor.jsonl
├── Makefile
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

## 5. Implementation Roadmap

### Phase 1: Go Engine Foundation and Ingestion Parsers
- Initialize Go module with minimal dependencies.
- Build high-throughput parsers in `internal/ingestion` for JSONL, Parquet, and CSV formats.
- Implement column mapping interface extracting instruction, context input, and target output.

### Phase 2: Embedding Generation and Representation Matrix
- Build lightweight embedding client in `internal/embedding/adapter.go` supporting local quantized ONNX models and external embedding endpoints.
- Build in-memory dense matrix representation with memory-efficient SIMD buffers.

### Phase 3: Spectral Signature Analysis & SVD Engine
- Implement power iteration algorithm in `internal/spectral/svd.go` to compute leading eigenvalues and eigenvectors without heavy external LAPACK/BLAS dependencies.
- Implement outlier projection scoring identifying clusters deviating along the primary spectral axis.

### Phase 4: Token Mutual Information Trigger Detector
- Implement n-gram frequency counter and Pointwise Mutual Information (PMI) calculation.
- Flag low-frequency tokens that abnormally correlate with safety-bypassing completions.

### Phase 5: Automated Dataset Partitioning and Sanitization
- Partition audited datasets into clean certified files and isolated `quarantine.json` logs.
- Generate cryptographic dataset safety attestation manifest (`audit.json`).

### Phase 6: Embedded React 19 Quarantine Studio
- Develop modern web UI in `web/` featuring 2D UMAP embedding scatter plot, suspicious token frequency charts, and interactive sample inspector.
- Embed frontend build artifacts into the Go binary via `//go:embed` and serve locally on `:7474`.

### Phase 7: Verification, Benchmarking, and Documentation
- Run test audits on `examples/test_datasets` verifying 98%+ detection rate of injected sleeper backdoors.
- Verify memory consumption remains strictly under 25MB RAM during streaming dataset scans.
- Author humanized README with setup guides, security playbooks, and ML pipeline integration patterns.

# Mathematical Specification: Spectral Signatures and Trigger Mining

This document details the mathematical algorithms used in poisonprobe for representation analysis, eigenvector projection, and trigger token extraction.

## 1. Representation Modeling

Let a dataset contain $N$ instruction samples:
$$\mathcal{D} = \{(\mathbf{u}_i, \mathbf{y}_i)\}_{i=1}^N$$
where $\mathbf{u}_i$ denotes the prompt (instruction and context input) and $\mathbf{y}_i$ denotes the target output text.

Each sample is mapped into a $d$-dimensional normalized representation space:
$$\mathbf{x}_i \in \mathbb{R}^d, \quad \|\mathbf{x}_i\|_2 = 1$$

The sample mean vector is:
$$\boldsymbol{\mu} = \frac{1}{N} \sum_{i=1}^N \mathbf{x}_i$$

Let $\mathbf{M} \in \mathbb{R}^{N \times d}$ be the centered feature matrix whose $i$-th row is:
$$\mathbf{m}_i = \mathbf{x}_i - \boldsymbol{\mu}$$

The empirical covariance matrix of the representations is:
$$\boldsymbol{\Sigma} = \frac{1}{N} \mathbf{M}^T \mathbf{M}$$

## 2. Dominant Eigenvector Extraction via Power Iteration

To keep memory usage well within the target threshold (< 25MB RAM) even for large datasets ($N \ge 100{,}000$), poisonprobe avoids storing the $d \times d$ covariance matrix in memory. Instead, it uses matrix-free power iteration.

For any probe vector $\mathbf{v} \in \mathbb{R}^d$, the product $\boldsymbol{\Sigma} \mathbf{v}$ is evaluated in two streaming linear steps:
1. Compute $\mathbf{z} = \mathbf{M} \mathbf{v} \in \mathbb{R}^N$, where $z_i = \mathbf{m}_i \cdot \mathbf{v}$.
2. Compute $\mathbf{w} = \frac{1}{N} \mathbf{M}^T \mathbf{z} \in \mathbb{R}^d$, where $\mathbf{w} = \frac{1}{N} \sum_{i=1}^N z_i \mathbf{m}_i$.

This calculation runs in $O(Nd)$ time while requiring only $O(d)$ auxiliary memory.

### Power Iteration Algorithm
1. Initialize $\mathbf{v}^{(0)}$ as a normalized random unit vector.
2. For iteration $t = 0, 1, 2, \dots, T$:
   $$\mathbf{w}^{(t)} = \boldsymbol{\Sigma} \mathbf{v}^{(t)}$$
   $$\mathbf{v}^{(t+1)} = \frac{\mathbf{w}^{(t)}}{\|\mathbf{w}^{(t)}\|_2}$$
   $$\lambda^{(t+1)} = (\mathbf{v}^{(t+1)})^T \boldsymbol{\Sigma} \mathbf{v}^{(t+1)}$$
3. Terminate when $\|\mathbf{v}^{(t+1)} - \mathbf{v}^{(t)}\|_2 < 10^{-6}$ or when $t = T_{\max}$ (default $T_{\max} = 50$).

The resulting vector $\mathbf{v}_1 = \mathbf{v}^{(t+1)}$ corresponds to the dominant eigenvector of $\boldsymbol{\Sigma}$ with principal eigenvalue $\lambda_1$.

To compute the second principal component $\mathbf{v}_2$ for 2D visualization, Hotelling deflation is applied:
$$\boldsymbol{\Sigma}_2 = \boldsymbol{\Sigma} - \lambda_1 \mathbf{v}_1 \mathbf{v}_1^T$$

## 3. Spectral Anomaly Scoring and Thresholding

The spectral anomaly score $\tau_i$ measures the squared projection of the centered sample representation onto the dominant eigenvector:
$$\tau_i = (\mathbf{m}_i \cdot \mathbf{v}_1)^2 = ((\mathbf{x}_i - \boldsymbol{\mu}) \cdot \mathbf{v}_1)^2$$

In an unpoisoned dataset, representations approximate a unimodal distribution along the primary axis. In a poisoned dataset, backdoor triggers introduce a correlated shift in embedding space, creating a bimodal distribution with anomalous samples separated along $\mathbf{v}_1$.

### Adaptive Threshold Calculation
Rather than relying on fixed arbitrary cutoffs, poisonprobe computes the Median Absolute Deviation (MAD):
$$\tilde{\tau} = \text{median}(\{\tau_i\}_{i=1}^N)$$
$$\text{MAD} = \text{median}(\{|\tau_i - \tilde{\tau}|\}_{i=1}^N)$$

A sample is flagged as a spectral outlier candidate if:
$$\tau_i > \tilde{\tau} + k \cdot 1.4826 \cdot \text{MAD}$$
where $k$ is the sensitivity multiplier (default $k = 3.0$).

## 4. Pointwise Mutual Information (PMI) for Trigger Isolation

Spectral outliers indicate which samples are corrupted, but do not directly expose the trigger token. To discover the trigger sequence, poisonprobe measures token co-occurrence with the anomalous cluster.

Let $A$ denote the event that a sample is flagged as an anomaly ($A \in \{0, 1\}$).
Let $w$ denote a token or character n-gram extracted from the instruction prompt.

The Pointwise Mutual Information (PMI) is defined as:
$$\text{PMI}(w, A=1) = \log_2 \frac{P(w, A=1) + \epsilon}{P(w) P(A=1) + \epsilon}$$

To bound the score between $-1$ and $+1$, poisonprobe uses Normalized Pointwise Mutual Information (NPMI):
$$\text{NPMI}(w, A=1) = \frac{\text{PMI}(w, A=1)}{-\log_2(P(w, A=1) + \epsilon)}$$

Where:
- $\text{NPMI} = +1$ indicates complete co-occurrence (token $w$ appears exclusively in anomalous samples).
- $\text{NPMI} = 0$ indicates statistical independence.
- $\text{NPMI} = -1$ indicates complete mutual exclusivity.

Tokens with $\text{NPMI}(w, A=1) \ge 0.65$ and minimum count $c(w) \ge 3$ are categorized as high-risk trigger candidates.

## 5. Label Inconsistency and Noise Scoring

For pairs of samples $i$ and $j$ where instruction similarity is high:
$$\text{sim}(\mathbf{u}_i, \mathbf{u}_j) = \frac{\mathbf{x}_i \cdot \mathbf{x}_j}{\|\mathbf{x}_i\|_2 \|\mathbf{x}_j\|_2} \ge 0.85$$

the semantic divergence of their target completions $\mathbf{y}_i, \mathbf{y}_j$ is evaluated:
$$\delta_{ij} = 1 - \text{sim}(\mathbf{y}_i, \mathbf{y}_j)$$

If $\delta_{ij} \ge 0.70$, sample $i$ is flagged for label contradiction or inconsistent completion noise.

---
tags:
  - type/note
  - theme/deep-learning
  - theme/hardware
aliases: []
lead: GPUs process deep learning workloads efficiently because of their massively parallel architecture and tensor cores. The entire system is optimized around batched tensor operations.
created: 2026-04-26
modified: 2026-04-26
source: "Fleuret, François. The Little Book of Deep Learning, Ch. 2, 2024."
---

# Efficient Computation

Deep learning works in large part because someone realized that GPUs — originally built to render video games — are also very good at matrix math. That accidental fit between hardware and algorithm is a big part of why we're here.

---

## 2.1 GPUs, TPUs, and batches

A GPU has thousands of parallel compute units. It was designed for real-time image synthesis, which needs massive floating-point parallelism. Deep learning needs the same.

As GPU usage for AI grew, manufacturers added **tensor cores** — hardware specifically optimized for the matrix multiplications that dominate deep learning. Google went further and designed **Tensor Processing Units (TPUs)** from scratch for this workload.

The bottleneck usually isn't raw computation — it's memory bandwidth. Moving data between CPU and GPU memory is slow. Moving it between the GPU's internal cache levels also costs time. So the whole game is minimizing those transfers.

The standard solution is **batching**: collect a group of samples that fits entirely in GPU memory, process them all at once. Model parameters load into fast cache memory once per batch, not once per sample. A GPU processes a full batch almost as fast as a single sample.

Numbers: a standard GPU runs at $10^{13}$–$10^{14}$ FLOPs/second with 8–80 GB of memory. Weights are normally stored as FP32 (32 bits), but FP16 and even lower-bit formats work fine for inference — and sometimes training — with no meaningful accuracy drop.

---

## 2.2 Tensors

Everything in deep learning is a tensor: a multi-dimensional array of scalars, an element of $\mathbb{R}^{N_1 \times \cdots \times N_D}$, generalizing vectors and matrices to arbitrary dimensions.

Tensors represent three kinds of things:
- Input signals — an RGB image is a $3 \times H \times W$ tensor
- Model parameters — weight matrices, bias vectors
- Intermediate results — called **activations**, by analogy with neurons

A time series with $T$ steps and $D$ features per step is a $T \times D$ tensor (or $D \times T$ for historical reasons). Fifty RGB images at $32 \times 24$ resolution pack into a $50 \times 3 \times 24 \times 32$ tensor.

Two reasons tensors matter. First, they let you express almost any computation as vectorized operations — no slow Python loops. Second, the tensor abstraction separates the *shape* of the data from its *memory layout*, so reshape, transpose, and slice operations often run without copying any data at all.

All the frameworks — PyTorch, JAX — are built around this. Every person in the stack, from chip designers to model builders, knows the data is tensors. That shared assumption is what makes the whole system efficient.

---

## In code

Tensor shapes in Go — how to read dimensions:

```go
// An RGB image batch in Go: [batchSize][channels][height][width]
// 50 images at 224×224 with 3 channels
images := make([][3][224][224]float32, 50)

// With gonum, the same batch flattens to a 2D matrix:
// rows = batchSize, cols = channels*height*width
// mat.NewDense(50, 3*224*224, data)
```

Batching — process multiple samples in one forward pass:

```
// Without batching: N separate matrix multiplications
for each sample x:
    h = W · x + b   // one [D_out] result

// With batching: one matrix multiplication over all N samples
// X is [N × D_in], W is [D_in × D_out], result is [N × D_out]
H = X · W + broadcast(b)
```

FP32 vs FP16 — what changes in memory:

```go
// FP32: 4 bytes per parameter
modelSizeBytes := numParams * 4

// FP16: 2 bytes per parameter (inference)
modelSizeBytes := numParams * 2

// A 7B-parameter model:
// FP32 → 28 GB
// FP16 → 14 GB
// INT4  → 3.5 GB
```

Memory bandwidth is the real bottleneck — loading a weight matrix from DRAM to GPU registers costs far more than the multiply-accumulate itself. Batch size trades latency for throughput: a batch of 32 loads weights once but runs 32 forward passes.

---
# Back Matter

**Source**
- based_on:: [[MOC - Foundations]]

**References**
- see:: [[Training]] — batching connects to how SGD works
- see:: [[The Compute Schism]] — hardware constraints return when running large models

**Terms**
- GPU, TPU, FLOPs, tensor, batch, activation, FP16, FP32

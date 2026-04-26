---
tags:
  - type/note
  - theme/deep-learning
  - theme/efficiency
aliases: []
lead: A gap has opened between models only corporations can train and what anyone can run. Prompt engineering, quantization, LoRA, and model merging are the four main techniques for crossing it.
created: 2026-04-26
modified: 2026-04-26
source: "Fleuret, François. The Little Book of Deep Learning, Ch. 8, 2024."
---

# The Compute Schism

The best models now require hardware that only large corporations own to train. But there's a gap between "training" and "running" — and the field has developed a set of techniques to let people run and adapt large models on consumer hardware.

---

## 8.1 Prompt engineering

The cheapest way to specialize an LLM: just craft the input carefully. Move information from the model's weights into the context window.

The simplest version is **few-shot prompting**: include examples of the task format in the prompt and let the model continue the pattern (§7.1).

More powerful: **chain-of-thought prompting**. Phrase the prompt so the model generates intermediate reasoning steps before the final answer. Those steps are individually simpler to predict — the model has seen lots of similar reasoning in training — so they emerge more reliably. Stating something like "think step by step" or showing a few worked examples is often enough.

**Retrieval-Augmented Generation (RAG)**: connect the model to an external knowledge base. Embed both the user's query and a corpus of documents. Retrieve the most relevant documents, inject them into the prompt, and generate a response that synthesizes them. The model acts as a smart interface over external knowledge, not just its own weights.

Context size matters here. Standard attention is quadratic in sequence length (§4.8), which caps how much you can put in a prompt before compute becomes prohibitive.

---

## 8.2 Quantization

Running a 70B-parameter model requires ~140 GB of memory in FP16. Most people don't have that.

**Post-Training Quantization** compresses existing model weights. If you encode each weight with 4 bits instead of 16, memory drops by 4×. The surprising finding: performance barely degrades, because activations are sums of many terms and averaging smooths out quantization noise.

A concrete format (Q4_1 from llama.cpp): group weights into blocks of 32. Store a scaling factor $d$ and bias $m$ per block in FP16, and each weight $x$ as a 4-bit integer $q \in \{0, \ldots, 15\}$. Dequantize as $\tilde{x} = dq + m$. A block that was 64 bytes in FP16 becomes 20 bytes.

**Quantization-Aware Training** applies quantization during the forward pass but keeps full-precision weights and gradients for the backward pass. Better accuracy than post-training quantization, more expensive.

Quantization also speeds up inference — smaller data means faster memory transfers, which is the bottleneck.

---

## 8.3 Adapters and LoRA

Fine-tuning a full 70B-parameter model requires storing gradients and optimizer state for every parameter. With Adam, that's 3× the model size. Impractical.

**LoRA** [Hu et al., 2021] adds low-rank correction matrices to existing weight matrices instead of modifying them directly. For a weight matrix $W$ of size $C \times D$, LoRA adds:

$$X(W + BA)^T$$

where $A$ is $R \times D$ and $B$ is $C \times R$, with $R \ll \min(C, D)$. Only $A$ and $B$ are trained. $W$ is frozen.

$A$ is initialized randomly, $B$ is initialized to zero — so fine-tuning starts from the original model's behavior and diverges from there.

Typical $R$: a few percent of the original parameter count. At $R = 16$, you might train 0.1% of the parameters of the base model. Optimizer memory scales with trainable parameters, so this also reduces the memory required for Adam.

In practice: LoRA is applied to the attention weight matrices, not the MLP blocks. The resulting $BA$ correction can be baked back into $W$ after training — inference costs nothing extra.

**QLoRA** [Dettmers et al., 2023] combines a quantized (4-bit) base model with full-precision LoRA adapters. You quantize to reduce base model memory, then train adapters in full precision where precision actually matters.

---

## 8.4 Model merging

You can combine multiple fine-tuned models into one without any additional training.

**Task arithmetic** [Ilharco et al., 2022]: let $\theta$ be the base model parameters, and $\theta_t = \theta + \tau_t$ be a version fine-tuned on task $t$. The "task vector" $\tau_t = \theta_t - \theta$ encodes what was learned for that task.

You can add task vectors: $\theta + \tau_1 + \cdots + \tau_T$ produces a model with multi-task capabilities. You can subtract them: $\theta - \tau_t$ degrades performance on task $t$. This arithmetic works surprisingly well in practice.

Interference between task vectors degrades results as you add more tasks. Recent methods (TIES-Merging [Yadav et al., 2023]) address this by trimming small values and resolving sign conflicts between vectors before merging.

An alternative: merge at the layer level rather than the parameter level. [Akiba et al., 2024] combines parameter merging and layer recombination, using stochastic optimization to search over the combinatorial space of possible layer assignments. With three fine-tuned Mistral-7B variants, this outperforms either strategy alone.

---

## In code

LoRA — add low-rank adapters to a frozen weight matrix:

```go
// Original layer: y = x · W  (W is [inDim × outDim], frozen)
// LoRA adds:      y = x · W + x · (A · B)
// A is [inDim × rank], B is [rank × outDim], rank << min(inDim, outDim)
type LoRALayer struct {
    W [][]float64 // frozen base weights
    A [][]float64 // trainable, initialized randomly
    B [][]float64 // trainable, initialized to zero → starts as identity
}

func (l *LoRALayer) forward(x []float64) []float64 {
    base := matMul(x, l.W)          // frozen path
    lowRank := matMul(matMul(x, l.A), l.B) // adapter path
    out := make([]float64, len(base))
    for i := range out {
        out[i] = base[i] + lowRank[i]
    }
    return out
    // Only A and B receive gradients — W.grad is never computed
}
```

Q4 quantization — compress a block of 32 weights to 4 bits each:

```go
// Block of 32 FP16 weights → 20 bytes (vs 64 bytes)
type Q4Block struct {
    Scale  float16   // 2 bytes: d
    Offset float16   // 2 bytes: m
    Quant  [16]byte  // 16 bytes: 32 values packed as 4-bit integers (2 per byte)
}

// Dequantize: recover approximate float from 4-bit integer
func dequantize(q uint8, scale, offset float32) float32 {
    return scale*float32(q) + offset
    // q ∈ {0..15}, so q4 covers 16 levels between offset and offset+15*scale
}
```

Task arithmetic — add and subtract capabilities without retraining:

```go
// θ_base: original model weights
// θ_task: model fine-tuned on some task
// τ_task = θ_task - θ_base is the "task vector"

func taskArithmetic(base, taskA, taskB []float64, scaleA, scaleB float64) []float64 {
    merged := make([]float64, len(base))
    for i := range base {
        tauA := taskA[i] - base[i]   // what task A learned
        tauB := taskB[i] - base[i]   // what task B learned
        merged[i] = base[i] + scaleA*tauA + scaleB*tauB
    }
    return merged
    // Negate scaleA to subtract a capability (e.g., remove toxic content knowledge)
}
```

RAG pipeline — retrieval augmented generation:

```
// Offline: embed all documents and store in a vector index
for doc in corpus:
    index.add(embed(doc), doc)

// Online: at query time
func rag(query string) string:
    queryEmb    = embed(query)
    topDocs     = index.search(queryEmb, k=5)      // retrieve top-k similar docs
    augmented   = buildPrompt(query, topDocs)       // inject docs into prompt
    return llm.generate(augmented)
```

---
# Back Matter

**Source**
- based_on:: [[MOC - Applications]]

**References**
- see:: [[Training]] — fine-tuning and scaling laws are the motivation for these techniques
- see:: [[Efficient Computation]] — quantization connects to the FP16/FP32 hardware section

**Terms**
- Prompt engineering, chain-of-thought, RAG, quantization, LoRA, QLoRA, adapter, task arithmetic, model merging

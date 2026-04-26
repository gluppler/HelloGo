---
tags:
  - type/note
  - theme/deep-learning
  - theme/architecture
aliases: []
lead: Three architecture families dominate — MLPs for low-dim inputs, ResNets for images, Transformers for sequences. GPT scales the Transformer encoder; ViT applies it to images via patches.
created: 2026-04-26
modified: 2026-04-26
source: "Fleuret, François. The Little Book of Deep Learning, Ch. 5, 2024."
---

# Architectures

Each application domain has developed its own preferred architectures — combinations of the building blocks from Ch4 that work well for specific problems. Three families dominate: MLPs, convolutional networks, and attention-based models.

---

## 5.1 Multi-Layer Perceptrons

The simplest deep architecture: stack fully connected layers with activation functions between them.

$$X \xrightarrow{\text{fc}} \xrightarrow{\text{relu}} \xrightarrow{\text{fc}} \xrightarrow{\text{relu}} \xrightarrow{\text{fc}} Y$$

The **universal approximation theorem** (Cybenko, 1989) says that a single hidden layer MLP can approximate any continuous function on a compact domain — given enough neurons in that hidden layer. In theory, one layer is enough. In practice, wider-but-shallower is usually worse than narrower-but-deeper for the same parameter count.

MLPs remain useful when input dimension is manageable. For anything involving spatial data (images, audio, video), you want convolutions instead.

---

## 5.2 Convolutional networks

Standard architecture for images. Combines convolutional layers (extract features) and fully connected layers (produce final predictions).

### LeNet-style

The original pattern [LeCun et al., 1998]:

1. Alternating conv + max-pool layers to extract features and reduce spatial size
2. Flatten to a vector
3. MLP for classification

AlexNet, VGG — just bigger versions of this blueprint.

### ResNets

LeNet-style networks don't scale to many layers — the vanishing gradient kills training. ResNets [He et al., 2015] fix this with residual connections.

A **residual block** applies: batch norm → relu → conv (1×1) → relu → batch norm → conv (3×3) → relu → batch norm → conv (1×1), then adds the original input back via a skip connection. The 1×1 convolutions reduce then restore channel count, keeping the 3×3 convolution cheap.

**Downscaling residual blocks** halve the spatial size and double channels using stride-2 convolutions. The skip connection needs a 1×1 conv with stride 2 to match the new tensor shape.

ResNet-50 structure:
- 7×7 conv → max pool (3 channels → 64, size halved)
- 4 sections of residual blocks with progressive downscaling
- Average pool (7×7) → flatten → fully connected (1000 classes)

Final representation: 2048 channels at 7×7 spatial size before the pooling. You need high channel count to carry a rich representation — the 1×1 bottleneck trick is what makes this feasible.

---

## 5.3 Attention models

Convolutional networks see the world locally. For tasks where distant parts of the input need to interact — machine translation, question answering, coherent image generation — you need attention.

### Transformer

The original Transformer [Vaswani et al., 2017] was designed for sequence-to-sequence translation. It has an encoder and an autoregressive decoder, both built from residual blocks using three types of sub-module:

**Feed-forward block**: layer norm → 1-hidden-layer MLP (with GELU). Updates representations at each position independently.

**Self-attention block**: layer norm → Multi-Head Attention where $X^Q = X^K = X^V$. Lets any position collect information from any other position in the sequence.

**Cross-attention block**: same as self-attention but queries come from one sequence and keys/values come from another. Used in the decoder to condition on encoder output.

Encoder: embed tokens → add positional encoding → $N$ self-attention blocks → refined representation $Z_1, \ldots, Z_T$.

Decoder: embed target tokens → add positional encoding → $N$ alternating *causal* self-attention + cross-attention blocks → logits predicting next tokens. Being causal means the model can be trained by minimizing cross-entropy over the full output sequence in one forward pass.

### GPT

GPT [Radford et al., 2018] is just the Transformer encoder made causal — a stack of causal self-attention blocks, no encoder/decoder split. Pure autoregressive text generation.

GPT-3 is 175B parameters: 96 self-attention blocks, 96 heads each, token dimension 12,288, MLP hidden dimension 49,512. It scales extremely well.

When trained on very large datasets, GPT-style models become **Large Language Models**. They learn grammar, facts, reasoning patterns — whatever was needed to predict the next token well across a huge corpus. That's what produces the few-shot and chain-of-thought capabilities.

### Vision Transformer (ViT)

ViT [Dosovitskiy et al., 2020] applies transformer architecture to images.

Process:
1. Split the image into $M$ patches of size $P \times P$
2. Flatten each patch: $M$ vectors of shape $3P^2$
3. Multiply by trainable $W^E$ to project to dimension $D$
4. Prepend a trainable class token $E_0$
5. Add positional encoding → process through $N$ self-attention blocks
6. Feed the output at position 0 ($Z_0$) through a 2-layer MLP for classification

The class token ($Z_0$) was introduced in BERT. It aggregates global information from all patches during attention.

ViT works at scale — given enough data it matches or beats convolutional networks. On smaller datasets, it's weaker because it lacks the translation-equivariance inductive bias that convolutions provide for free.

---

## Summary

| Architecture | Best for | Key idea |
|---|---|---|
| MLP | Low-dim inputs, tabular data | Universal approximation with FC layers |
| ConvNet (ResNet) | Images, spatial signals | Local equivariant processing + residual blocks |
| Transformer | Sequences, language, vision at scale | Global attention + positional encoding |
| GPT | Text generation | Causal Transformer encoder, autoregressive |
| ViT | Image classification at scale | Patches as tokens, self-attention |

---

## In code

MLP forward pass — stacked linear layers with activations:

```go
// Each layer is a weight matrix and bias vector
type LinearLayer struct {
    W [][]float64 // [outDim × inDim]
    b []float64   // [outDim]
}

func (l *LinearLayer) forward(x []float64) []float64 {
    out := make([]float64, len(l.b))
    for i, row := range l.W {
        for j, w := range row {
            out[i] += w * x[j]
        }
        out[i] += l.b[i]
    }
    return out
}

// MLP: linear → relu → linear → relu → linear
func mlpForward(x []float64, layers []*LinearLayer) []float64 {
    h := x
    for i, layer := range layers {
        h = layer.forward(h)
        if i < len(layers)-1 {
            for j := range h {
                h[j] = relu(h[j])  // apply activation between layers, not after the last
            }
        }
    }
    return h
}
```

ResNet block — structure only, not runnable without a tensor library:

```
// Bottleneck residual block: reduce → process → restore channel count
func bottleneckBlock(x Tensor, inC, midC, outC int) Tensor:
    identity = x

    out = conv1x1(x, inC → midC)    // squeeze channels
    out = batchNorm(out); relu(out)
    out = conv3x3(out, midC → midC)  // spatial processing
    out = batchNorm(out); relu(out)
    out = conv1x1(out, midC → outC)  // restore channels
    out = batchNorm(out)

    if inC != outC:
        identity = conv1x1(x, inC → outC)  // match dimensions for the skip

    return relu(out + identity)
```

ViT: split image into patches and project to embeddings:

```go
// Split a [C × H × W] image into M patches of size P×P
// Returns M vectors of length C*P*P
func extractPatches(image [][][]float64, P int) [][]float64 {
    C, H, W := len(image), len(image[0]), len(image[0][0])
    M := (H / P) * (W / P)
    patches := make([][]float64, M)
    idx := 0
    for row := 0; row < H; row += P {
        for col := 0; col < W; col += P {
            patch := make([]float64, 0, C*P*P)
            for c := 0; c < C; c++ {
                for r := row; r < row+P; r++ {
                    patch = append(patch, image[c][r][col:col+P]...)
                }
            }
            patches[idx] = patch
            idx++
        }
    }
    return patches
    // Each patch then gets projected: embedding = patch · W_E
    // W_E is [C*P*P × D], producing M vectors of dimension D
}
```

Autoregressive decoding — how GPT generates tokens:

```
// Sample one token at a time, feeding output back as input
tokens = [startToken]
for step in 1..maxLen:
    logits  = model.forward(tokens)          // [len(tokens) × vocabSize]
    nextProbs = softmax(logits[-1])          // take the last position's distribution
    nextToken = sampleFromDistribution(nextProbs)
    tokens = append(tokens, nextToken)
    if nextToken == endToken:
        break
```

---
# Back Matter

**Source**
- based_on:: [[MOC - Deep Models]]

**References**
- see:: [[Model Components]] — the building blocks these architectures combine
- see:: [[Prediction]] — how ConvNets and ViTs get applied
- see:: [[Synthesis]] — GPT architecture for text generation

**Terms**
- MLP, ResNet, Transformer, GPT, ViT, residual block, self-attention, cross-attention, CLS token

---
tags:
  - type/note
  - theme/deep-learning
  - theme/architecture
aliases: []
lead: Deep models are built from reusable layer types — linear, convolutional, attention, normalization, skip connections. Each solves a specific problem; most exist to fight vanishing gradients or exploit signal structure.
created: 2026-04-26
modified: 2026-04-26
source: "Fleuret, François. The Little Book of Deep Learning, Ch. 4, 2024."
---

# Model Components

A deep model is ultimately just a big composition of tensor operations. The field has settled on a set of reusable building blocks — layers — that work well and combine cleanly. This chapter covers those blocks.

---

## 4.1 The notion of layer

A **layer** is a standard, reusable tensor operation — modular, well-studied, and often with trainable parameters. The name comes from simple multi-layer neural nets, but modern models can be complex graphs of these modules with multiple parallel paths.

---

## 4.2 Linear layers

The most computationally important layers. "Linear" in deep learning usually means **affine**: a linear transformation plus a bias.

### Fully connected layers

A weight matrix $W$ of size $D' \times D$ and bias vector $b$ of size $D'$. Given input $X$ of shape $D_1 \times \cdots \times D_K \times D$, it outputs:

$$Y[d_1, \ldots, d_K] = W X[d_1, \ldots, d_K] + b$$

At first this looks like it only does rotations and translations. It can do more — projections, filtering, similarity matching. A matrix-vector product is equivalent to measuring how well the input matches a set of patterns encoded in the rows of $W$.

Initialization matters: bad initialization causes exploding or vanishing activations before training even starts. Frameworks handle this by scaling random initial values according to input dimension.

### Convolutional layers

Fully connected layers scale poorly with input size — processing a $256 \times 256$ RGB image needs $\approx 4 \times 10^{10}$ parameters. That's impractical.

Images also have **structure**: short-range correlations, translation invariance. A fully connected layer ignores all of that. Convolutional layers exploit it.

A **1D convolution** has kernel size $K$, input channels $D$, output channels $D'$, and a trainable affine map $\phi(\cdot; w) : \mathbb{R}^{D \times K} \to \mathbb{R}^{D' \times 1}$. It applies this map to every $D \times K$ sub-tensor of the input, storing results in a $D' \times (T - K + 1)$ output.

A **2D convolution** is the same but with a $K \times L$ kernel applied over spatial dimensions.

Convolutions are **equivariant to translation**: shift the input and the output shifts the same way. That's the inductive bias you want for signals whose statistics don't depend on position.

Three extra hyperparameters:
- **Padding** — add zeros around the input to control output size
- **Stride** — step size through the input; stride 2 halves the output size
- **Dilation** — spacing between filter coefficients; increases effective kernel size without adding parameters

The **receptive field** of an activation is the portion of the input that it depends on. Each convolutional layer increases the receptive field by roughly the kernel size.

**Transposed convolutions** run the process in reverse — taking a compact representation and upsampling to a larger signal. Used for synthesis tasks and in segmentation architectures.

---

## 4.3 Activation functions

Without non-linearity, a stack of linear layers is still just a linear layer. Activation functions fix that by applying a non-linear transformation elementwise.

**ReLU** is the default:

$$\text{relu}(x) = \begin{cases} 0 & x < 0 \\ x & \text{otherwise} \end{cases}$$

It's not differentiable at zero and is constant on half the real line — both seem problematic for gradient-based training. In practice it works because what matters is that the gradient is informative *on average*, not everywhere. Proper initialization keeps roughly half the activations positive at the start.

**Tanh** was the standard before ReLU. It saturates on both sides, which aggravates vanishing gradients. That's why it was replaced.

**Leaky ReLU** keeps a small slope $a$ for negative inputs: $\text{leakyrelu}(x) = ax$ for $x < 0$.

**GELU** is $x \cdot P(Z \leq x)$ where $Z \sim \mathcal{N}(0,1)$ — a smooth approximation of ReLU. Used in transformers.

The choice between ReLU variants is mostly empirical.

---

## 4.4 Pooling

Pooling reduces spatial size by summarizing a region into one value.

**Max pooling** takes the maximum over non-overlapping sub-tensors of the input. Intuition: it's like a logical OR — at least one instance of the thing you're looking for was present somewhere in that region. It loses precise location in exchange for local invariance.

**Average pooling** takes the mean instead. It's linear; max pooling isn't.

Both share the same hyperparameters as convolutions (padding, stride, dilation). The standard stride equals the kernel size.

---

## 4.5 Dropout

Dropout has no trainable parameters. During training, each activation is independently set to zero with probability $p$, and all surviving activations are scaled by $\frac{1}{1-p}$ to maintain the expected value. During testing, it's a no-op.

The idea: forces the model not to co-adapt groups of neurons, since any group has probability $(1-p)^k$ of surviving intact. Equivalently, it's noise injection that makes training more robust.

For 2D signals (images), individual activations can be inferred from neighbors because of spatial correlation — standard dropout doesn't work. Instead, whole channels get dropped at once.

Dropout can also be left on during inference to estimate confidence via sampling.

---

## 4.6 Normalizing layers

Normalization layers force activations to have zero mean and unit variance, which stabilizes training and allows deeper models to train at all.

**Batch normalization** computes per-channel mean and variance across the batch:

$$\hat{x} = \frac{x - \hat{m}}{\sqrt{\hat{v} + \epsilon}}, \quad y = \gamma \hat{x} + \beta$$

where $\gamma$ and $\beta$ are learned scale and shift parameters. It's the only standard layer that operates across a batch rather than on individual samples.

At test time, it uses running averages of $\hat{m}$ and $\hat{v}$ from training — effectively becoming a fixed affine transform.

For 2D signals, normalization is per-channel across all spatial positions, so $\gamma$ and $\beta$ remain $D$-dimensional vectors.

**Layer normalization** computes moments across all features of a single sample instead of across the batch. It behaves the same at training and test time, which makes it better suited for settings where batch size is small or variable — like autoregressive generation.

---

## 4.7 Skip connections

Skip connections carry the signal unchanged across multiple layers, bypassing processing in between. The output of an earlier layer gets concatenated or added to the output of a later layer.

**Residual connections** are a specific type: they add the original signal to the transformed one, $y = f(x) + x$. They skip only a few layers at a time.

Why they matter: even if some layers learn near-zero or gradient-killing transformations, the skip connection ensures gradients still flow. This is what makes 100+-layer models trainable. ResNets and Transformers are both built entirely out of residual blocks.

Skip connections also appear in architectures that downscale then upscale (like U-Net for segmentation), where they connect encoder layers to decoder layers at matching resolutions.

---

## 4.8 Attention layers

Convolutions only look locally. Fully connected layers can't handle large or variable-size inputs. Attention fixes both problems by computing, for every output position, a weighted sum over all input positions.

### The attention operator

Given:
- Queries $Q$ of size $N^Q \times D_{QK}$
- Keys $K$ of size $N^{KV} \times D_{QK}$
- Values $V$ of size $N^{KV} \times D_V$

First compute attention scores:

$$A_{q,k} = \frac{\exp\!\left(\frac{1}{\sqrt{D_{QK}}} Q_q \cdot K_k\right)}{\sum_l \exp\!\left(\frac{1}{\sqrt{D_{QK}}} Q_q \cdot K_l\right)}$$

Then compute the output by averaging values weighted by scores:

$$Y_q = \sum_k A_{q,k} V_k$$

The $\frac{1}{\sqrt{D_{QK}}}$ scaling keeps dot products from growing too large as dimension increases. If one key matches much better than all others, the output is basically that key's value. If several match equally, you get an average.

Compact form:

$$\text{att}(Q, K, V) = \text{softargmax}\!\left(\frac{QK^T}{\sqrt{D_{QK}}}\right) V$$

You can also **mask** the attention matrix before normalization — e.g., mask the upper triangle to make the operator causal (no future information leaks into past positions).

Cost: quadratic in sequence length. This is a real problem for very long sequences and an active research area.

### Multi-Head Attention

The full Multi-Head Attention layer runs $H$ attention heads in parallel, each with its own learned projection matrices $W^Q_h$, $W^K_h$, $W^V_h$:

$$Y_h = \text{att}(X^Q W^Q_h,\ X^K W^K_h,\ X^V W^V_h)$$

The $H$ outputs are concatenated and passed through a final linear layer $W^O$:

$$Y = (Y_1 \mid \cdots \mid Y_H) W^O$$

Used as **self-attention** (all three inputs are the same sequence) or **cross-attention** (queries from one sequence, keys/values from another).

---

## 4.9 Token embedding

When inputs are discrete tokens (words, characters, BPE fragments), they need to be converted to continuous vectors. An **embedding layer** is just a lookup table: a trainable $N \times D$ matrix where each integer input maps to a row.

$$Y[d_1, \ldots, d_K] = M[X[d_1, \ldots, d_K]]$$

---

## 4.10 Positional encoding

Convolutions and attention are both invariant to absolute position — by design, since that's their inductive bias. But sometimes position matters. The word "not" before vs. after a verb changes everything.

The fix is to add a **positional encoding** to the feature vectors at every position. It can be learned or fixed analytically.

The original Transformer uses sinusoidal encoding:

$$\text{PE}[t, d] = \begin{cases} \sin\!\left(\frac{t}{T^{d/D}}\right) & d \in 2\mathbb{N} \\ \cos\!\left(\frac{t}{T^{(d-1)/D}}\right) & \text{otherwise} \end{cases}$$

with $T = 10^4$. Each dimension oscillates at a different frequency, giving every position a unique fingerprint.

---

## Summary of equations

$$Y = WX + b \quad \text{(linear/affine layer)}$$

$$\text{relu}(x) = \max(0, x)$$

$$\text{gelu}(x) = x \cdot P(Z \leq x), \quad Z \sim \mathcal{N}(0,1)$$

$$\hat{x} = \frac{x - \hat{m}}{\sqrt{\hat{v} + \epsilon}}, \quad y = \gamma\hat{x} + \beta \quad \text{(batch norm)}$$

$$y = f(x) + x \quad \text{(residual connection)}$$

$$\text{att}(Q, K, V) = \text{softargmax}\!\left(\frac{QK^T}{\sqrt{D_{QK}}}\right) V$$

---

## In code

Activation functions in Go — the ones you'll implement by hand:

```go
func relu(x float64) float64 {
    if x < 0 {
        return 0
    }
    return x
}

func sigmoid(x float64) float64 {
    return 1.0 / (1.0 + math.Exp(-x))
}

// Derivative of sigmoid — use this form when x is pre-activated (already in (0,1))
func sigmoidPrimeActivated(x float64) float64 {
    return x * (1.0 - x)
}

func tanh(x float64) float64 {
    return math.Tanh(x)
}
```

Residual block — the core of ResNet:

```
// Standard residual block
func residualBlock(x Tensor) Tensor:
    identity = x                    // save the skip connection
    out = batchNorm(x)
    out = relu(out)
    out = conv3x3(out)
    out = batchNorm(out)
    out = relu(out)
    out = conv3x3(out)
    return out + identity           // add the skip — this is what prevents vanishing gradients
```

Dropout — zero out activations randomly during training:

```go
func dropout(x []float64, p float64, training bool) []float64 {
    if !training {
        return x  // no-op at inference time
    }
    out := make([]float64, len(x))
    scale := 1.0 / (1.0 - p)
    for i, v := range x {
        if rand.Float64() > p {
            out[i] = v * scale  // scale surviving activations to keep expected value
        }
    }
    return out
}
```

Scaled dot-product attention (single head, pseudocode):

```
// Q: [N_q × D_k], K: [N_kv × D_k], V: [N_kv × D_v]
func attention(Q, K, V Matrix) Matrix:
    scale  = 1.0 / sqrt(D_k)
    scores = Q · K.T * scale      // [N_q × N_kv]
    weights = softmax(scores, axis=1)  // normalize each query's scores
    return weights · V            // [N_q × D_v]
```

Layer norm — normalize across features, not across the batch:

```go
// For a single sample's feature vector
func layerNorm(x, gamma, beta []float64, eps float64) []float64 {
    mean := floats.Sum(x) / float64(len(x))
    var variance float64
    for _, v := range x {
        d := v - mean
        variance += d * d
    }
    variance /= float64(len(x))
    std := math.Sqrt(variance + eps)

    out := make([]float64, len(x))
    for i, v := range x {
        out[i] = gamma[i]*((v-mean)/std) + beta[i]
    }
    return out
}
```

---
# Back Matter

**Source**
- based_on:: [[MOC - Deep Models]]

**References**
- see:: [[Training]] — vanishing gradients are why skip connections and normalization layers exist
- see:: [[Architectures]] — how these components assemble into full models
- see:: [[AI-ML Neural Network in Go]] — fully connected layers (§4.2) and sigmoid activation (§4.3) implemented by hand in Go
- see:: [[AI-ML Go Data Science Tooling]] — gorgonia implements the layer types covered here with autodiff

**Terms**
- Layer, activation, ReLU, GELU, batch norm, layer norm, dropout, skip connection, residual connection, attention, embedding, positional encoding

---
tags:
  - type/note
  - theme/deep-learning
  - theme/optimization
aliases: []
lead: Training is minimizing a loss via gradient descent. Backprop computes gradients cheaply. Deeper models learn better representations. Scale everything together — model, data, compute — and performance follows.
created: 2026-04-26
modified: 2026-04-26
source: "Fleuret, François. The Little Book of Deep Learning, Ch. 3, 2024."
---

# Training

Training a model means finding $w^*$ that minimizes a loss $\mathcal{L}(w)$. That sounds simple. It's not — the models are complex, and the math gets messy fast.

---

## 3.1 Losses

The loss you train on isn't always the thing you actually care about. It's a proxy that has an informative gradient. That distinction matters.

**MSE** is standard for continuous predictions:

$$\mathcal{L}(w) = \frac{1}{N} \sum_{n=1}^N (y_n - f(x_n; w))^2$$

**Cross-entropy** is standard for classification. The model outputs one logit per class, then you apply softmax to get probabilities:

$$\hat{P}(Y = y \mid X = x) = \frac{\exp(f(x; w)_y)}{\sum_z \exp(f(x; w)_z)}$$

Then minimize the negative log-probability of the true class:

$$\mathcal{L}_{ce}(w) = -\frac{1}{N} \sum_{n=1}^N \log \hat{P}(Y = y_n \mid X = x_n)$$

Cross-entropy gets used instead of classification error rate because the error rate has no useful gradient — you can't descend it.

**Contrastive loss** handles metric learning: you want samples $x_a$ and $x_b$ of the same class to be closer than $x_a$ and $x_c$ from a different class. The triplet loss:

$$\text{Loss} = \sum_\text{triplets} \max(0, 1 - f(x_a, x_c; w) + f(x_a, x_b; w))$$

**Weight decay** (L2 regularization) adds $\lambda \|w\|^2$ to any loss. It penalizes large weights, reduces overfitting, and can be interpreted as a Gaussian prior on the parameters. It makes training-set performance slightly worse but usually improves generalization.

---

## 3.2 Autoregressive models

For sequences — text, audio, video frames — you can factorize the joint probability using the chain rule:

$$P(x_1, \ldots, x_T) = P(x_1) \prod_{t=2}^T P(x_t \mid x_1, \ldots, x_{t-1})$$

A model $f$ that predicts the next token given all previous ones is an **autoregressive model**. Training it: minimize cross-entropy summed over all tokens in all training sequences.

The **perplexity** $\exp(\mathcal{L}_{ce})$ is what people usually report. It's more interpretable than raw cross-entropy — it represents the effective vocabulary size of a uniform distribution with the same entropy.

**Causal models** predict all tokens in one forward pass by masking: the output at position $t$ can only depend on inputs at positions $< t$. This is what makes transformers trainable on full sequences efficiently.

**Tokenizers** handle the conversion between text and integers. Byte Pair Encoding (BPE) is standard — it merges frequent character pairs hierarchically, producing tokens that represent fragments of words at varying granularity.

---

## 3.3 Gradient descent

Except for simple cases like linear regression, there's no closed-form solution for $w^*$. So you use **gradient descent**: start from a random $w_0$, and iteratively move in the direction that reduces the loss:

$$w_{n+1} = w_n - \eta \nabla \mathcal{L}(w_n)$$

The **learning rate** $\eta$ controls step size. Too small: slow convergence, might get stuck. Too large: bounces around, never settles. Choosing it well matters a lot.

Computing the full gradient over all $N$ samples every step is expensive. Instead, **stochastic gradient descent (SGD)** uses mini-batches: subsets of the data whose gradients are noisy but unbiased estimates of the full gradient. Because batches fit in GPU memory and process in parallel, you get more gradient steps per dollar of compute.

The number of gradient steps is typically millions. Don't expect it to converge in a handful of epochs.

**Adam** [Kingma & Ba, 2014] is the default optimizer. It keeps running estimates of the mean and variance of each gradient component, normalizing them to avoid scaling problems and speed mismatches between different parts of the model.

---

## 3.4 Backpropagation

To do gradient descent you need $\nabla \mathcal{L}|_w$. For a model $f = f^{(D)} \circ \cdots \circ f^{(1)}$, you compute this using the chain rule:

**Forward pass**: compute each layer's output $x^{(d)} = f^{(d)}(x^{(d-1)}; w_d)$ in sequence. These intermediate outputs are the **activations**.

**Backward pass**: propagate gradients backward through each layer using the Jacobian:

$$\nabla \ell \big|_{x^{(d-1)}} = \nabla \ell \big|_{x^{(d)}} \cdot J_{f^{(d)}} \big|_x$$

$$\nabla \ell \big|_{w_d} = \nabla \ell \big|_{x^{(d)}} \cdot J_{f^{(d)}} \big|_w$$

In practice you never write this yourself. Deep learning frameworks build the backward pass automatically via autograd.

Memory note: the backward pass needs to keep all forward-pass activations in memory to compute Jacobians. Memory usage grows with model depth. Checkpointing trades memory for recomputation — store only some activations, recompute the rest as needed.

**Vanishing gradients** happen when gradients shrink exponentially as they propagate backward through many layers. The field spent years fighting this — it's why architecture choices like residual connections and normalization layers exist.

---

## 3.5 The value of depth

Stacking more layers genuinely helps. Empirically, state-of-the-art performance across domains requires models with tens of layers. Theoretically, for a fixed parameter budget, deeper models express more complex functions than shallower ones [Telgarsky, 2016].

Why? The layers co-adapt during training — each one warps the representation so the next layer's job is easier. Stacking eight 2×2 matrix layers with Tanh can take a non-linearly separable dataset and make it linearly separable at the output. That's what depth buys you.

---

## 3.6 Training protocols

You need at least two data splits: a **training set** to optimize $w$, and a **test set** to evaluate the final model. If you're tuning hyperparameters, add a **validation set** separate from both.

Training runs in **epochs** — full passes through the training data. The typical behavior: training loss decreases throughout, validation loss decreases for a while then starts increasing as the model overfits.

Paradoxically, very large models often keep improving even beyond apparent overfitting. The inductive bias of the model takes over as the driver of optimization once training performance is near-perfect.

**Learning rate scheduling**: start high to avoid getting trapped early, reduce over time to settle into a good minimum. The schedule matters as much as the initial rate.

**Fine-tuning**: take a pre-trained model and continue training it on a downstream task. This is how vision models trained on ImageNet get repurposed for detection, segmentation, etc. It's also how LLMs get turned into assistants via RLHF.

---

## 3.7 The benefits of scale

Performance improves predictably with model size, dataset size, and compute — all three together, following **scaling laws** [Kaplan et al., 2020]. If you scale up the model without scaling up the data, you won't get the benefit.

Vision models: 10–100M parameters, $10^{18}$–$10^{19}$ FLOPs to train.
Language models: 100M–hundreds of billions of parameters, $10^{20}$–$10^{23}$ FLOPs to train.

The largest models require machines with multiple high-end GPUs, months of training, and millions of dollars. They're trained on internet-scale datasets (LAION-5B, The Pile, OSCAR) collected automatically with minimal curation. As of 2024, LLMs are the dominant example.

---

## Summary of equations

$$w_{n+1} = w_n - \eta \nabla \mathcal{L}(w_n) \quad \text{(gradient descent)}$$

$$\hat{P}(Y = y \mid X = x) = \frac{\exp(f(x;w)_y)}{\sum_z \exp(f(x;w)_z)} \quad \text{(softmax)}$$

$$\mathcal{L}_{ce}(w) = -\frac{1}{N} \sum_{n=1}^N \log \hat{P}(Y = y_n \mid X = x_n) \quad \text{(cross-entropy)}$$

$$\text{Perplexity} = \exp(\mathcal{L}_{ce})$$

$$\nabla \ell \big|_{x^{(d-1)}} = \nabla \ell \big|_{x^{(d)}} \cdot J_{f^{(d)}} \big|_x \quad \text{(backprop, activation gradient)}$$

---

## In code

SGD update — one parameter vector, one gradient step:

```go
func sgdStep(params, grads []float64, lr float64) {
    for i := range params {
        params[i] -= lr * grads[i]
    }
}
```

Softmax — turn raw logits into a probability distribution:

```go
func softmax(logits []float64) []float64 {
    probs := make([]float64, len(logits))
    var sum float64
    for _, v := range logits {
        sum += math.Exp(v)
    }
    for i, v := range logits {
        probs[i] = math.Exp(v) / sum
    }
    return probs
}
```

Cross-entropy loss — negative log probability of the correct class:

```go
// probs is the softmax output; trueClass is the index of the correct label
func crossEntropy(probs []float64, trueClass int) float64 {
    return -math.Log(probs[trueClass] + 1e-9)  // epsilon avoids log(0)
}
```

The forward/backward pass structure (pseudocode):

```
// Forward pass: cache activations at every layer
for d = 1 to D:
    x[d] = f_d(x[d-1]; w_d)    // store x[d] — needed for backward pass

// Backward pass: propagate gradient from output to input
grad = ∂loss/∂x[D]
for d = D down to 1:
    grad_w[d] = grad · ∂f_d/∂w_d    // gradient for this layer's weights
    grad      = grad · ∂f_d/∂x[d-1] // pass gradient to previous layer
    w[d]      = w[d] - η * grad_w[d]
```

Learning rate schedule — linear decay:

```go
func lrSchedule(initLR float64, epoch, totalEpochs int) float64 {
    return initLR * (1.0 - float64(epoch)/float64(totalEpochs))
}
```

---
# Back Matter

**Source**
- based_on:: [[MOC - Foundations]]

**References**
- see:: [[Machine Learning]] — what the loss is measuring
- see:: [[Model Components]] — skip connections and normalization exist because of vanishing gradients
- see:: [[The Compute Schism]] — fine-tuning and LoRA build on §3.6
- see:: [[AI-ML Neural Network in Go]] — hand-coded backpropagation in Go: the Jacobian chain rule made concrete
- see:: [[AI-ML Neural Network Foundations]] — gradient descent and learning rate trade-offs applied to small networks

**Terms**
- Loss, cross-entropy, softmax, perplexity, gradient descent, SGD, Adam, backpropagation, learning rate, fine-tuning, scaling laws

---
tags:
  - type/note
  - theme/deep-learning
  - theme/machine-learning
aliases: []
lead: Deep learning is a subfield of ML where models are long compositions of mappings. Training means finding weights that minimize a loss over a dataset — everything else is details.
created: 2026-04-26
modified: 2026-04-26
source: "Fleuret, François. The Little Book of Deep Learning, Ch. 1, 2024."
---

# Machine Learning

Deep learning is technically a subfield of machine learning. The "deep" part just means the models are long compositions of mappings — you stack many layers on top of each other, and that stacking turns out to be what makes everything work.

---

## 1.1 Learning from data

The core idea is almost embarrassingly simple. You have an input $x$ (say, a photo of a license plate) and you want to predict $y$ (the characters on it). You can't write a rule for this by hand, so instead you collect a training set $\mathcal{D} = \{(x_n, y_n)\}$, pick a parameterized function $f(x; w)$, and find weights $w^*$ that make $f$ a good predictor.

"Good" means the loss $\mathcal{L}(w)$ is small when $f(\cdot; w)$ predicts well on $\mathcal{D}$. Then training is just finding:

$$\hat{y} = f(x; w^*)$$

Most of this book is about what $f$ actually looks like. The weights $w$ are called **weights** by analogy with synaptic weights in biological neurons. Models also depend on **hyperparameters** — things like architecture choices or regularization strength — which you set manually rather than learning from data.

---

## 1.2 Basis function regression

The simplest case: $x_n$ and $y_n$ are both real numbers, and $f$ is a linear combination of fixed basis functions $f_1, \ldots, f_K$:

$$f(x; w) = \sum_{k=1}^{K} w_k f_k(x)$$

With MSE as the loss:

$$\mathcal{L}(w) = \frac{1}{N} \sum_{n=1}^N (y_n - f(x_n; w))^2$$

Since $f$ is linear in $w$ and the loss is quadratic in $f$, finding $w^*$ reduces to solving a linear system. No iterations, closed-form solution. Gaussian kernels are a common choice for the basis functions.

---

## 1.3 Under and overfitting

This tension never really goes away.

**Underfitting** — model capacity is too low. High training error. The model literally cannot fit the data.

**Overfitting** — too little data relative to model capacity. The model fits the training set perfectly but has learned quirks specific to those examples, so it generalizes poorly to new inputs.

The fix is getting the **inductive bias** right — designing a model whose structure matches the structure of the data. A model that expects local patterns shouldn't be applied where only global patterns matter.

One confusing wrinkle: very large models with enormous capacity often don't overfit the way you'd expect. This comes up again in §3.6 and §3.7. The short version is that the model's inductive bias takes over as the dominant driver once training performance nears perfect.

---

## 1.4 Categories of models

**Regression** — predict a continuous value $y \in \mathbb{R}^K$. Example: predicting an object's position. Training uses paired $(x, y)$ examples with a ground-truth value.

**Classification** — predict a label from a finite set $\{1, \ldots, C\}$. Standard approach: output one score per class, correct class gets the maximum score.

**Density modeling** — model the probability distribution $\mu_X$ of the data itself. No labels needed. You can evaluate the density, sample from it, or both.

Regression and classification are **supervised learning** — you need ground truth labels someone provided. Density modeling is **unsupervised** — the data supervises itself.

These categories overlap. Classification is really just score regression. Autoregressive sequence modeling is iterated classification. Real problems often combine multiple objectives.

---

## Summary of equations

$$\hat{y} = f(x; w^*)$$

$$\mathcal{L}(w) = \frac{1}{N} \sum_{n=1}^N (y_n - f(x_n; w))^2 \quad \text{(MSE)}$$

$$f(x; w) = \sum_{k=1}^K w_k f_k(x) \quad \text{(basis function model)}$$

$$\mathcal{L}_\text{sup}(w) = \frac{1}{N} \sum_{n=1}^N \ell(f(x_n; w), y_n) \quad \text{(general supervised loss)}$$

$$\mathcal{L}_\text{density}(w) = -\sum_{n=1}^N \log P(x_n; w) \quad \text{(density/likelihood loss)}$$

---

## In code

Basis function regression — evaluate $f(x; w)$ and compute MSE:

```go
// f(x; w) = Σ w_k * φ_k(x)
// features is the pre-evaluated basis vector φ(x)
func predict(w, features []float64) float64 {
    var out float64
    for k := range w {
        out += w[k] * features[k]
    }
    return out
}

// MSE over N samples
func mse(preds, targets []float64) float64 {
    var loss float64
    for i := range preds {
        d := targets[i] - preds[i]
        loss += d * d
    }
    return loss / float64(len(preds))
}
```

The supervised training loop in pseudocode:

```
initialize w randomly
for epoch in 1..numEpochs:
    for each (x, y) in trainingSet:
        ŷ    = f(x; w)
        loss = ½(y - ŷ)²
        grad = ∂loss/∂w          // chain rule
        w    = w - η * grad      // gradient descent step
```

Overfitting check — split data, track both losses:

```go
// Stop when validation loss stops improving
if valLoss > prevValLoss {
    break  // early stopping
}
prevValLoss = valLoss
```

---
# Back Matter

**Source**
- based_on:: [[MOC - Foundations]]

**References**
- see:: [[Training]] — how loss functions get minimized
- see:: [[Model Components]] — what $f(x; w)$ looks like internally
- see:: [[AI-ML Neural Network Foundations]] — Go implementation context: supervised learning, loss functions, classification

**Terms**
- Loss function, inductive bias, overfitting, underfitting, supervised learning, density modeling

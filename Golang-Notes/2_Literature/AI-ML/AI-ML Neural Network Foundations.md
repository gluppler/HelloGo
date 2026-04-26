---
tags:
  - type/note
  - theme/golang
  - theme/machine-learning
  - theme/deep-learning
aliases: []
lead: The math and mechanics behind feedforward neural networks — neurons, activation functions, forward propagation, backpropagation, and the training loop. Grounded in both ujjwalkarn's intro and the LBDL.
created: 2026-04-26
modified: 2026-04-27
source: "ujjwalkarn.me, Quick Intro to Neural Networks, 2016. Fleuret, François. The Little Book of Deep Learning, Ch. 1, 3, 4, 2024."
---

# AI-ML — Neural network foundations

## What a neuron computes

A neuron takes a weighted sum of its inputs, adds a bias, and passes the result through a nonlinear function:

$$\text{output} = f\!\left(\sum_i w_i x_i + b\right)$$

The bias shifts the activation threshold independently of the input. Without it, a sigmoid neuron always passes through 0.5 when all inputs are zero — there's no way to move that crossing point. The LBDL calls this structure an **affine layer**: $Y = WX + b$. See [[Model Components]] §4.2.

## Activation functions

The whole point of an activation function is nonlinearity. Stack linear layers and you still get a single linear transformation no matter how many you add. From [[Model Components]] §4.3:

**Sigmoid** squashes any real number to $(0, 1)$:

$$\sigma(x) = \frac{1}{1 + e^{-x}}, \qquad \sigma'(x) = \sigma(x)(1 - \sigma(x))$$

Its derivative is expressible in terms of its own output — which is why hand-written backprop for sigmoid networks is tractable. The output is interpretable as a probability. This is what gophernet uses (see [[AI-ML Neural Network in Go]]).

One subtlety: gophernet's `sigmoidPrime` function applies sigmoid again to already-activated values, computing $\sigma(\sigma(x))(1-\sigma(\sigma(x)))$ instead of $\sigma(x)(1-\sigma(x))$. The network still converges because sigmoid maps $(0,1)$ to $(0.5, 0.73)$, keeping the error small. On deeper networks this compounding error would matter more.

**Tanh** squashes to $(-1, 1)$. Same S-curve, but zero-centered. The LBDL notes that tanh saturates on both ends, which aggravates **vanishing gradients** — one reason modern deep networks moved away from it.

**ReLU** is $\max(0, x)$. Cheap to compute, no upper saturation. The LBDL's default recommendation for deep networks. Sigmoid is fine for shallow, small networks trained from scratch.

**GELU** ($x \cdot P(Z \leq x)$, $Z \sim \mathcal{N}(0,1)$) is what transformers use — a smooth ReLU approximation. Out of scope for the basic Go implementations here.

## Feedforward networks

A feedforward network passes data one direction: input → hidden layers → output. No loops.

```
x₁  ──┐
x₂  ──┼──→ [hidden neurons] ──→ [output neurons]
x₃  ──┘
```

The LBDL makes the depth argument precisely in [[Training]] §3.5: deeper models express more complex functions than shallower ones for a fixed parameter budget. Each layer warps the representation so the next layer's job is easier. A single-layer perceptron can only draw linear decision boundaries — that's a property of the architecture, not the training algorithm.

## Matrix convention in Go implementations

Forward propagation in gophernet uses the **rows-are-samples** convention:

$$H = \sigma(X \cdot W_{\text{hidden}} + b_{\text{hidden}})$$
$$\hat{Y} = \sigma(H \cdot W_{\text{out}} + b_{\text{out}})$$

`X` is `[samples × features]`, so `W_hidden` must be `[features × hidden]` for the multiplication to produce `[samples × hidden]`. This is the opposite of the columns-are-samples convention you'll see in many textbooks. Getting the transpose direction wrong when implementing by hand is the most common bug.

## Backpropagation

Training minimizes the loss between predictions and targets. For MSE:

$$L = \frac{1}{2}(y - \hat{y})^2$$

The LBDL covers this in [[Training]] §3.1 and §3.4 using the Jacobian formulation. For a two-layer network it reduces to:

**Output layer error:**

$$\delta^{(\text{out})} = (y - \hat{y}) \cdot \sigma'(\hat{y})$$

**Hidden layer error** (propagated backward through the transposed weight matrix):

$$\delta^{(\text{hidden})} = \delta^{(\text{out})} \cdot W_{\text{out}}^T \cdot \sigma'(H)$$

**Weight update:**

$$W_{\text{out}} \mathrel{+}= \eta \cdot H^T \cdot \delta^{(\text{out})}$$
$$W_{\text{hidden}} \mathrel{+}= \eta \cdot X^T \cdot \delta^{(\text{hidden})}$$

The transposed $W_{\text{out}}^T$ is what "sends errors backward" — each hidden neuron gets a weighted sum of output errors, weighted by how strongly it connects to each output. That's the chain rule in matrix form. Real frameworks compute this via autograd; see [[Training]] §3.4.

## Bias gradient

Bias has its own update, separate from weights:

$$b_{\text{out}} \mathrel{+}= \eta \cdot \sum_{\text{samples}} \delta^{(\text{out})}$$

The sum collapses the sample dimension because one bias value applies to every sample. In Go this is `sumAlongAxis(0, dOutput)` using `floats.Sum` from gonum. See the implementation in [[AI-ML Neural Network in Go]].

## The training loop

Gradient descent: start from random weights, then repeat:

1. Forward pass — compute predictions
2. Compute loss
3. Backward pass — compute gradients
4. Update weights and biases
5. Repeat for all training data (one **epoch**), then start another

The LBDL's [[Training]] §3.3 covers learning rate trade-offs. Too large and you overshoot; too small and convergence stalls. For sigmoid networks on small datasets, 0.1–0.3 is a reasonable starting range.

Overfitting is a real concern for small handcrafted networks. The LBDL's §3.6 notes that very large models can improve past apparent overfitting — but for gophernet-scale nets, if validation accuracy drops while training accuracy keeps rising, stop training. See [[Machine Learning]] §1.3 for the underfitting/overfitting framing.

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[AI-ML Neural Network in Go]] — Go implementation of every concept here, including the sigmoidPrime bug
- see:: [[AI-ML Go Data Science Tooling]] — libraries used in practice
- see:: [[Machine Learning]] — formal treatment of loss functions, supervised learning, overfitting
- see:: [[Training]] — gradient descent (§3.3), backprop as Jacobian chain (§3.4), depth argument (§3.5), training protocols (§3.6)
- see:: [[Model Components]] — activation functions (§4.3), affine layers (§4.2), vanishing gradients and skip connections

**Terms**
- Neuron, bias, sigmoid, tanh, ReLU, GELU, sigmoidPrime bug, rows-are-samples convention, feedforward network, forward propagation, backpropagation, chain rule, Jacobian, gradient descent, learning rate, epoch, overfitting, vanishing gradient, MSE, bias gradient

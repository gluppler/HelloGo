---
tags:
  - type/fleeting
  - theme/deep-learning
lead: Topics skipped by LBDL for length — RNNs, autoencoders, GANs, GNNs, self-supervised learning. Each warrants its own permanent note.
created: 2026-04-26
modified: 2026-04-26
status: to-process
---

# The Missing Bits

The book skips several important topics for the sake of length. Here's what's left out and why it matters.

---

## Recurrent Neural Networks

Before attention models showed up, **RNNs** were the standard for sequences (text, audio). They maintain a hidden state that updates as each element of the sequence is processed. The key layers are LSTM [Hochreiter & Schmidhuber, 1997] and GRU [Cho et al., 2014].

Training an RNN means unrolling it through time — which is just a long composition of operators — and applying backprop. This is where techniques like rectifiers and gating (a form of skip connections modulated by the input) were first developed. Those techniques later became standard for deep architectures generally.

The fundamental problem with RNNs: $x_{t+1} = f(x_t)$ forces serial processing. You can't parallelize over time. For a sequence of length $T$, training takes $O(T)$ time.

Transformers process all positions in parallel, taking constant time (with enough hardware). That's a large part of why they won.

Some architectures try to get the best of both: QRNN, S4, and Mamba [Gu & Dao, 2023] use recurrent operations that are affine — so $f^t$ and the resulting $x_t = f^t(x_0)$ can be computed in parallel. Constant time if $f$ doesn't depend on $t$, $O(\log T)$ otherwise.

---

## Autoencoders

An **autoencoder** maps a high-dimensional input to a low-dimensional latent representation, then maps it back. The constraint that reconstruction must be good forces the latent space to encode genuinely useful structure.

We saw this in §6.1 for denoising. It can also learn a compact parametrization of a data manifold — useful for compression, anomaly detection, or just understanding what structure the data has.

The **Variational Autoencoder (VAE)** [Kingma & Welling, 2013] adds a distribution constraint on the latent space. The loss includes a term that pushes the latent representations toward a fixed prior (usually Gaussian). After training, you can sample the latent space and decode samples — turning the autoencoder into a generative model.

---

## Generative Adversarial Networks

**GANs** [Goodfellow et al., 2014] take a different approach to density modeling. Two networks compete:

- **Generator**: takes random noise, produces fake samples
- **Discriminator**: takes a sample and predicts real vs. fake

The discriminator trains to minimize classification loss. The generator trains to maximize it — to fool the discriminator. At equilibrium, the generator produces samples indistinguishable from real data.

In practice, gradients from the discriminator flow back to the generator and tell it which specific cues it needs to fix. This implicit adversarial feedback is what makes GANs produce sharp, realistic outputs. Diffusion models have largely replaced GANs for image synthesis, but GANs were dominant for years and the ideas behind them still matter.

---

## Graph Neural Networks

Images are grids, sequences are chains. But proteins, 3D meshes, social networks, and molecular structures are **graphs** — nodes connected by edges with no regular spatial structure.

Standard convolutions don't apply. **GNNs** [Scarselli et al., 2009] generalize convolution to graphs: each node's new activation is a linear combination of its neighbors' current activations. No spatial indexing, just local message passing.

---

## Self-supervised learning

LLMs are trained to predict the next word. That's a self-supervised objective — the data labels itself. The model never sees human-provided labels, yet it learns representations useful for a huge range of downstream tasks.

**Self-supervised learning** generalizes this principle. Define a task that doesn't need labels but forces learning of useful representations. Then fine-tune on the actual labeled task using a small dataset.

In vision: train a model so that two different augmented views of the same image produce similar embeddings, while different images produce dissimilar ones (Barlow Twins [Zbontar et al., 2021]).

In NLP and vision both: mask parts of the input and train the model to reconstruct them (BERT for text [Devlin et al., 2018], iBOT for images).

The key insight: you have far more unlabeled data than labeled data. Self-supervised pre-training lets you use all of it.

---
tags:
  - type/note
  - theme/deep-learning
  - theme/generative-models
aliases: []
lead: Synthesis is fitting a density model and sampling from it. For text — autoregressive LLMs trained at scale. For images — diffusion models that learn to reverse a noise process step by step.
created: 2026-04-26
modified: 2026-04-26
source: "Fleuret, François. The Little Book of Deep Learning, Ch. 7, 2024."
---

# Synthesis

Prediction asks "what is this?" Synthesis asks "make me one of these." You're fitting a density model and sampling from it.

---

## 7.1 Text generation

The standard approach: train a GPT-style autoregressive model (§5.3) on massive text data. When the dataset is large enough, the result is a **Large Language Model**.

GPT-3 is the canonical example: 175B parameters, 96 self-attention blocks, 96 heads, token dimension 12,288, MLP hidden size 49,512. It was trained to predict the next word — nothing else. That simple objective, at that scale, produces something unexpectedly capable.

LLMs learn grammar and syntax, obviously. But they also absorb factual knowledge, reasoning patterns, and enough world structure to predict sensible continuations to things like "The capital of Japan is" or "if water is heated to 100 Celsius." This produces **few-shot prediction** — you can give the model a handful of examples in the prompt and it generalizes. You can also get **chain-of-thought** behavior, where giving it space to reason step by step produces dramatically better answers on math and logic problems.

These models are sometimes called **foundation models** because they're pre-trained once and repurposed many times.

The problem: a model trained on internet text doesn't naturally produce helpful dialog. It's optimized for "what comes next in a corpus of novels, forums, and blog posts," which isn't the same as "give a clear, safe, useful response to a user."

Fix: **RLHF** (Reinforcement Learning from Human Feedback). Collect human-written responses or human ratings of generated responses. Use ratings to train a reward model. Fine-tune the LLM to maximize that reward using standard RL. This is how GPT-3 becomes ChatGPT-style assistants.

---

## 7.2 Image generation

**Diffusion models** are the current dominant approach for image synthesis.

The idea: define an analytical process that gradually destroys an image by adding noise, and train a neural network to reverse it.

**Forward process**: starting from a clean image $x_0$, repeatedly apply a small Gaussian corruption until after $T$ steps the image is pure noise. The distribution $p(x_T)$ becomes a known Gaussian that you can sample trivially.

**Reverse process**: train a network $f(x_t, t; w)$ to denoise one step at a time:

$$x_{t-1} \mid x_t \sim \mathcal{N}(x_t + f(x_t, t; w),\ \sigma_t)$$

If this one-step reverse process is accurate enough, sampling $x_T \sim p(x_T)$ and running $T$ denoising steps produces a sample from $p(x_0)$.

Training: generate many noisy sequences, pick a random time step from each, maximize:

$$\sum_n \log f\!\left(x^{(n)}_{t_n - 1},\ x^{(n)}_{t_n},\ t_n;\ w\right)$$

The denoiser $f$ is typically a U-Net-style convolutional architecture with attention.

What it looks like in practice: start from pure noise, the model hallucinates structures at random, then gradually reinforces the most probable continuation, step by step, until a coherent image emerges.

**Text-conditioned synthesis**: add a bias to the mean of the denoising distribution that pushes the generated image toward matching a text description. GLIDE [Nichol et al., 2021] does this by using the CLIP text-image similarity score (§6.6) as the guidance signal.

---

## In code

Autoregressive text generation with temperature sampling:

```go
// temperature > 1 → more random; temperature < 1 → more deterministic
func generateTokens(model Model, prompt []int, maxLen int, temperature float64) []int {
    tokens := make([]int, len(prompt))
    copy(tokens, prompt)

    for len(tokens) < maxLen {
        logits := model.Forward(tokens)           // [vocabSize] — last position only
        scaled := make([]float64, len(logits))
        for i, v := range logits {
            scaled[i] = v / temperature           // divide before softmax
        }
        probs := softmax(scaled)
        next := sampleCategorical(probs)          // draw from distribution
        tokens = append(tokens, next)
        if next == eosTokenID {
            break
        }
    }
    return tokens
}
```

Diffusion forward process — add noise step by step:

```
// Schedule: β_1, β_2, ..., β_T (small constants, e.g. 0.0001 to 0.02)
// x_0 is the clean image; x_T is pure noise

func forwardDiffuse(x0 []float64, t int, betas []float64) []float64:
    // Closed-form: x_t = sqrt(ᾱ_t) * x_0 + sqrt(1 - ᾱ_t) * ε
    // where ᾱ_t = product(1 - β_i for i in 1..t) and ε ~ N(0, I)
    alphaCum = cumulativeProduct(1 - betas[:t])
    noise    = sampleGaussian(shape=x0.shape)
    return sqrt(alphaCum) * x0 + sqrt(1 - alphaCum) * noise
```

Diffusion reverse process — denoise one step at a time:

```
// At inference: start from x_T ~ N(0, I), work backward to x_0
x = sampleGaussian()   // pure noise

for t = T down to 1:
    predictedNoise = denoiser.forward(x, t)   // neural network predicts ε
    // Reconstruct x_{t-1} from x_t and predicted noise
    x = (x - sqrt(1-ᾱ_t) * predictedNoise) / sqrt(ᾱ_t)
    if t > 1:
        x += sqrt(β_t) * sampleGaussian()    // re-add small noise (stochastic sampling)

return x   // x_0: generated image
```

RLHF pipeline — three stages in pseudocode:

```
// Stage 1: supervised fine-tuning on human-written responses
sftModel = finetune(baseLLM, humanDemonstrations)

// Stage 2: train a reward model from human preference rankings
// Given two responses A and B to the same prompt, human says "A is better"
rewardModel = trainRewardModel(
    positives=preferredResponses,
    negatives=rejectedResponses,
)

// Stage 3: PPO to maximize expected reward
for batch of prompts:
    response      = sftModel.generate(prompt)
    reward        = rewardModel.score(prompt, response)
    klPenalty     = kl(sftModel.logprobs, frozenRef.logprobs)  // don't drift too far
    totalReward   = reward - β * klPenalty
    ppoStep(sftModel, totalReward)
```

---
# Back Matter

**Source**
- based_on:: [[MOC - Applications]]

**References**
- see:: [[Architectures]] — GPT is the text generation model
- see:: [[Prediction]] — CLIP score used for text-conditioned diffusion guidance
- see:: [[Training]] — RLHF is fine-tuning with a reward signal

**Terms**
- Autoregressive, LLM, foundation model, RLHF, diffusion model, forward process, denoising, variational bound

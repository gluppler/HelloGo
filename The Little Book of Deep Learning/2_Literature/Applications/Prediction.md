---
tags:
  - type/note
  - theme/deep-learning
  - theme/applications
aliases: []
lead: Prediction tasks — denoising, classification, detection, segmentation, speech recognition, CLIP, RL — all share the same structure. Given a signal, predict something. The architecture and loss vary; the training loop doesn't.
created: 2026-04-26
modified: 2026-04-26
source: "Fleuret, François. The Little Book of Deep Learning, Ch. 6, 2024."
---

# Prediction

The first major category of applications: given a signal, predict something unknown about it. Face recognition, object detection, speech-to-text, sentiment analysis — all prediction tasks.

---

## 6.1 Image denoising

Images have statistical structure — patches of uniform texture, consistent geometry, predictable color relationships. A **denoising autoencoder** exploits that structure to recover a clean image from a degraded one.

Architecture: usually a convolutional network with skip connections (to preserve fine-grained spatial detail) and sometimes attention layers (to handle long-range dependencies). Takes degraded input $\tilde{X}$, produces estimate of clean $X$.

Training: pair clean images with degraded versions (low light, grayscale conversion, lossy compression, artificial noise). Minimize MSE over all pixels:

$$\mathcal{L}(w) = \|X - f(\tilde{X}; w)\|^2$$

The MSE-trained model learns $\mathbb{E}[X \mid \tilde{X}]$ — the average clean image given the degraded input. If parts of $X$ aren't fully determined by $\tilde{X}$, those parts will look blurry (average over possible completions). That's a known limitation.

---

## 6.2 Image classification

Predict a class label from an image. The simplest task in image understanding.

Standard models: ResNets (§5.2) and ViT (§5.3). Both output a logit vector — one value per class. Train with cross-entropy.

**Data augmentation** almost always helps: random crops, horizontal flips, color jitter, scaling. These are transformations that don't change the label but diversify the training set. The model has to learn features that survive those transformations, which forces it to find genuinely informative ones.

---

## 6.3 Object detection

More complex than classification: predict both the class and the bounding box $(x_1, y_1, x_2, y_2)$ for each object in the image. Training data has bounding box annotations, which are slow to produce.

**Single Shot Detector (SSD)** [Liu et al., 2015] runs a convolutional backbone that produces feature maps $Z_s$ at multiple resolutions ($s = 1, \ldots, S$, with $S$ being the coarsest). Each spatial position in each feature map corresponds to a receptive field in the input image.

For each spatial position at each scale, the model predicts:
- Bounding box coordinates (regression)
- Class scores including a "no object" class (classification)
- Multiple boxes per position at different aspect ratios

Training loss: cross-entropy for class scores + MSE for box coordinates. Positions matched to ground-truth boxes get both losses; unmatched positions get only the "no object" classification loss.

SSD fine-tunes from a classification network (VGG-16 originally) rather than training from scratch — models trained on ImageNet classification learn representations that transfer well to detection.

---

## 6.4 Semantic segmentation

The finest-grain prediction task: assign a class label to every pixel. Output is a $C \times H \times W$ logit map, one channel per class.

The challenge is that you need both global context (what is this?) and fine spatial resolution (exactly which pixel does this object occupy?). That requires multi-scale processing.

Standard approach: **downscale then upscale**.
- Convolutional backbone shrinks the feature map (increasing receptive field)
- Transposed convolutions or bilinear interpolation restore resolution for the final per-pixel prediction

Problem with strict downscale-upscale: fine spatial detail gets lost passing through the low-resolution bottleneck.

Fix: **skip connections** from encoder layers to decoder layers at matching resolutions (U-Net style [Ronneberger et al., 2015]). Lets the decoder access fine-grained features directly.

Alternatively, build a multi-scale representation in parallel (Pyramid Scene Parsing Network [Zhao et al., 2016]) — pool at multiple scales, concatenate everything, predict.

Training: cross-entropy summed over all pixels. Almost always fine-tuned from an ImageNet-pretrained backbone.

---

## 6.5 Speech recognition

Convert audio to text. Recent approach [Radford et al., 2022]: treat it as sequence-to-sequence translation, solve it with a standard Transformer.

Pipeline:
1. Convert audio to a spectrogram (1D series: $T \times D$ where $D$ is frequency bands)
2. Process spectrogram through 1D convolutional layers
3. Feed into Transformer encoder
4. Transformer decoder generates token sequence (transcription, or translation to English, or detection of non-speech)

Tokenization: BPE on text (§3.2).

Training objective: multiple tasks at once — transcription of any language, translation to English, non-speech detection. This lets the model leverage huge, heterogeneous datasets. The decoder is structurally identical to a generative language model — you're sampling a text distribution conditioned on audio.

---

## 6.6 Text-image representations

**CLIP** [Radford et al., 2021] learns aligned embeddings: an image and a text description of that image should produce similar vectors.

Architecture: image encoder (ViT) + text encoder (GPT). The text encoder takes a sequence ending with an EOS token and uses the EOS representation as the embedding.

Training: 400 million image-text pairs from the internet. Contrastive loss — within a mini-batch of $N$ pairs, maximize similarity $l_{n,n}$ while minimizing all cross-pair similarities $l_{n,m}$ ($n \neq m$):

$$l_{m,n} = f(i_m) \cdot g(t_n)$$

This forces the model to capture rich, specific image content rather than just coarse semantics.

**Zero-shot prediction**: at inference, define a set of candidate class descriptions as text. Compute embedding of a new image. Pick the class whose text embedding is most similar. No labeled examples needed.

CLIP also generalizes better than standard classifiers to adversarial examples, because it learned from descriptions that go beyond the simple discriminative cues standard classifiers rely on.

---

## 6.7 Reinforcement learning

For problems where you can't provide fixed ground-truth outputs — games, robotics, sequential decisions — reinforcement learning is the framework.

Setup: a state process $S_t$, reward $R_t$, and actions $A_t$. If the state is **Markovian** (the current state contains all relevant history), you have a **Markov Decision Process (MDP)**. Goal: find a policy $\pi(S_t)$ that maximizes expected discounted return:

$$\mathbb{E}\left[\sum_{t \geq 0} \gamma^t R_t\right], \quad 0 < \gamma < 1$$

The key object is the **Q-function** $Q(s, a)$ — expected return from state $s$ taking action $a$ then following the optimal policy. It satisfies the **Bellman equation**:

$$Q(s, a) = \mathbb{E}\left[R_t + \gamma \max_{a'} Q(S_{t+1}, a') \mid S_t = s, A_t = a\right]$$

**Deep Q-Network (DQN)** [Mnih et al., 2015]: approximate $Q$ with a neural network. For Atari games, the state is 4 stacked frames (for some motion history), and the network is a small LeNet-style convnet with one output per action.

Training: alternately play games (collecting $(s, a, r, s')$ tuples) and minimize:

$$\mathcal{L}(w) = \frac{1}{N} \sum_n (Q(s_n, a_n; w) - y_n)^2$$

where $y_n = r_n + \gamma \max_{a'} Q(s'_n, a'; \bar{w})$ and $\bar{w}$ is a frozen copy of $w$. Freezing $\bar{w}$ stabilizes training by preventing the target from moving while you're chasing it.

Exploration: $\epsilon$-greedy — take a random action with probability $\epsilon$, otherwise take the greedy action. Without exploration you never discover good strategies.

Results: 10M training frames (~8 days of gameplay), human-level performance on most of 49 Atari games.

---

## In code

CLIP zero-shot classification — no labeled examples needed:

```
// Encode both the image and each candidate class description
imageEmb = imageEncoder.forward(image)        // [D]
classEmbs = [textEncoder.forward(desc)        // [numClasses × D]
             for desc in classDescriptions]

// Normalize both sets of embeddings to unit length
imageEmb  = normalize(imageEmb)
classEmbs = [normalize(e) for e in classEmbs]

// Similarity scores — cosine similarity via dot product on normalized vectors
scores = [dot(imageEmb, classEmb) for classEmb in classEmbs]
predictedClass = argmax(scores)
```

DQN training loop — experience replay and frozen target network:

```
replayBuffer = []
targetParams = copy(modelParams)   // frozen copy; updated every N steps

for episode in 1..numEpisodes:
    state = env.reset()
    while not done:
        // Epsilon-greedy action selection
        if rand() < epsilon:
            action = env.randomAction()
        else:
            action = argmax(model.forward(state))   // greedy

        nextState, reward, done = env.step(action)
        replayBuffer.add(state, action, reward, nextState, done)
        state = nextState

        // Sample a random mini-batch and update
        batch = replayBuffer.sample(batchSize)
        for (s, a, r, s', done) in batch:
            if done:
                target = r
            else:
                target = r + γ * max(targetModel.forward(s'))  // frozen target
            loss += (model.forward(s)[a] - target)²

        gradientStep(loss)

    // Decay exploration
    epsilon = max(epsilonMin, epsilon * epsilonDecay)

    // Periodically sync the target network
    if episode % syncEvery == 0:
        targetParams = copy(modelParams)
```

Image classification training — augmentation, forward, cross-entropy, backward:

```go
// Training step for one batch of images
func trainStep(model Model, images []Tensor, labels []int, lr float64) float64 {
    var totalLoss float64
    for i, img := range images {
        aug := randomAugment(img)           // random crop, flip, color jitter
        logits := model.Forward(aug)        // [numClasses]
        probs := softmax(logits)
        totalLoss += crossEntropy(probs, labels[i])
        model.Backward(labels[i])           // compute gradients
    }
    model.UpdateWeights(lr)
    return totalLoss / float64(len(images))
}
```

---
# Back Matter

**Source**
- based_on:: [[MOC - Applications]]

**References**
- see:: [[Architectures]] — ResNets, ViT, and Transformer are the models used here
- see:: [[Training]] — fine-tuning pre-trained models is the standard strategy across all these tasks
- see:: [[Synthesis]] — speech recognition decoder is structurally a generative model

**Terms**
- Denoising autoencoder, data augmentation, SSD, semantic segmentation, CLIP, zero-shot prediction, MDP, Q-function, DQN, ε-greedy

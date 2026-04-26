---
tags:
  - type/note
  - theme/golang
  - theme/machine-learning
  - theme/deep-learning
aliases: []
lead: Building a feedforward neural network in Go — exact code from gophernet (iris dataset), matrix layout, backpropagation, known bugs, and golang-pro improvements.
created: 2026-04-26
modified: 2026-04-27
source: "github.com/dwhitena/gophernet main.go (Daniel Whitenack); datadan.io/blog/neural-net-with-go; sausheong.github.io/posts/how-to-build-a-simple-artificial-neural-network-with-go; madeddu.xyz/posts/neuralnetwork."
---

# AI-ML — Neural network in Go

## Why Go for this

Go isn't the first language anyone reaches for in ML. Python has the ecosystem. But implementing a network in Go — without a framework doing the heavy lifting — forces you to confront the matrix math directly. You write the weight updates yourself, the bias broadcasting yourself, the gradient accumulation yourself. That understanding transfers back to Python too.

The canonical reference implementation is Daniel Whitenack's `gophernet` (discussed in *Machine Learning with Go*). This note follows that code precisely, including its bugs.

## The dataset: iris, pre-normalized

The `data/` directory has two CSVs: `train.csv` (120 samples) and `test.csv` (30 samples). The header row:

```
sepal_length,sepal_width,petal_length,petal_width,setosa,virginica,versicolor
```

The values are **already min-max normalized** to `[0, 1]`. A sample row:

```
0.0833333333333,0.666666666667,0.0,0.0416666666667,1.0,0.0,0.0
```

Label columns are ordered **setosa (4), virginica (5), versicolor (6)** — not the typical setosa/versicolor/virginica order. The one-hot encoding uses exact `1.0` and `0.0`.

Because the data arrives pre-normalized, there is no normalization step in the Go code. This is different from MNIST pipelines (see the sausheong blog post), where you normalize 0–255 pixel values to `[0.01, 0.99]`.

## Module setup

The repo has no `go.mod` — it predates modules. To use it today:

```bash
go mod init github.com/dwhitena/gophernet
go get gonum.org/v1/gonum
go mod tidy
```

## Core data structures

```go
type neuralNetConfig struct {
    inputNeurons  int
    outputNeurons int
    hiddenNeurons int
    numEpochs     int
    learningRate  float64
}

type neuralNet struct {
    config  neuralNetConfig
    wHidden *mat.Dense  // [inputNeurons × hiddenNeurons]
    bHidden *mat.Dense  // [1 × hiddenNeurons]
    wOut    *mat.Dense  // [hiddenNeurons × outputNeurons]
    bOut    *mat.Dense  // [1 × outputNeurons]
}
```

Matrix shapes matter. Input arrives as `[samples × features]`, so `wHidden` must be `[features × hiddenNeurons]` for `x · wHidden` to produce `[samples × hiddenNeurons]`. This is the **rows-are-samples** convention. Getting the transpose direction wrong is the most common hand-implementation bug.

Biases are row vectors (`[1 × neurons]`), broadcast column-by-column via a closure in `Apply`.

## Weight initialization

```go
func newNetwork(config neuralNetConfig) *neuralNet {
    return &neuralNet{config: config}
}

func (nn *neuralNet) train(x, y *mat.Dense) error {
    randSource := rand.NewSource(time.Now().UnixNano())
    randGen := rand.New(randSource)

    wHidden := mat.NewDense(nn.config.inputNeurons, nn.config.hiddenNeurons, nil)
    bHidden := mat.NewDense(1, nn.config.hiddenNeurons, nil)
    wOut    := mat.NewDense(nn.config.hiddenNeurons, nn.config.outputNeurons, nil)
    bOut    := mat.NewDense(1, nn.config.outputNeurons, nil)

    for _, param := range [][]float64{
        wHidden.RawMatrix().Data,
        bHidden.RawMatrix().Data,
        wOut.RawMatrix().Data,
        bOut.RawMatrix().Data,
    } {
        for i := range param {
            param[i] = randGen.Float64()  // uniform [0, 1)
        }
    }

    output := new(mat.Dense)
    if err := nn.backpropagate(x, y, wHidden, bHidden, wOut, bOut, output); err != nil {
        return err
    }

    nn.wHidden = wHidden
    nn.bHidden = bHidden
    nn.wOut    = wOut
    nn.bOut    = bOut
    return nil
}
```

`RawMatrix().Data` gives direct access to the flat `[]float64` backing array. Bulk-filling it with `rand.Float64()` is faster than calling `Set` in a double loop.

Weights are assigned to `nn` only after `backpropagate` returns — `wHidden` and `wOut` stay `nil` until then. The `predict` method checks for nil before doing anything.

## Activation functions — and a real bug

```go
func sigmoid(x float64) float64 {
    return 1.0 / (1.0 + math.Exp(-x))
}

func sigmoidPrime(x float64) float64 {
    return sigmoid(x) * (1.0 - sigmoid(x))
}
```

`sigmoidPrime` is supposed to compute $\sigma'(x) = \sigma(x)(1-\sigma(x))$. When `x` is a pre-activation value this is correct. But in backpropagation, it gets applied to already-activated outputs:

```go
slopeOutputLayer.Apply(applySigmoidPrime, output)  // output is already σ(x)
```

This computes $\sigma(\sigma(x)) \cdot (1 - \sigma(\sigma(x)))$ — applying sigmoid a second time. The correct version for pre-activated inputs is:

```go
func sigmoidPrimeActivated(_, _ int, v float64) float64 {
    return v * (1.0 - v)  // v is already in (0, 1)
}
```

The original code still reaches 97% accuracy because sigmoid stays within $(0.5, 0.73)$ for inputs in $(0, 1)$, so the double-application error is small in practice. But it is a bug, and it would matter on deeper networks where the distortion compounds.

## Backpropagation

```go
func (nn *neuralNet) backpropagate(x, y, wHidden, bHidden, wOut, bOut, output *mat.Dense) error {
    applySigmoid      := func(_, _ int, v float64) float64 { return sigmoid(v) }
    applySigmoidPrime := func(_, _ int, v float64) float64 { return sigmoidPrime(v) }

    for i := 0; i < nn.config.numEpochs; i++ {
        // Forward pass
        hiddenLayerInput := new(mat.Dense)
        hiddenLayerInput.Mul(x, wHidden)
        addBHidden := func(_, col int, v float64) float64 { return v + bHidden.At(0, col) }
        hiddenLayerInput.Apply(addBHidden, hiddenLayerInput)

        hiddenLayerActivations := new(mat.Dense)
        hiddenLayerActivations.Apply(applySigmoid, hiddenLayerInput)

        outputLayerInput := new(mat.Dense)
        outputLayerInput.Mul(hiddenLayerActivations, wOut)
        addBOut := func(_, col int, v float64) float64 { return v + bOut.At(0, col) }
        outputLayerInput.Apply(addBOut, outputLayerInput)
        output.Apply(applySigmoid, outputLayerInput)

        // Backprop
        networkError := new(mat.Dense)
        networkError.Sub(y, output)

        slopeOutputLayer := new(mat.Dense)
        slopeOutputLayer.Apply(applySigmoidPrime, output)
        slopeHiddenLayer := new(mat.Dense)
        slopeHiddenLayer.Apply(applySigmoidPrime, hiddenLayerActivations)

        dOutput := new(mat.Dense)
        dOutput.MulElem(networkError, slopeOutputLayer)
        errorAtHiddenLayer := new(mat.Dense)
        errorAtHiddenLayer.Mul(dOutput, wOut.T())

        dHiddenLayer := new(mat.Dense)
        dHiddenLayer.MulElem(errorAtHiddenLayer, slopeHiddenLayer)

        // Weight updates — in-place mutation
        wOutAdj := new(mat.Dense)
        wOutAdj.Mul(hiddenLayerActivations.T(), dOutput)
        wOutAdj.Scale(nn.config.learningRate, wOutAdj)
        wOut.Add(wOut, wOutAdj)

        bOutAdj, err := sumAlongAxis(0, dOutput)
        if err != nil {
            return err
        }
        bOutAdj.Scale(nn.config.learningRate, bOutAdj)
        bOut.Add(bOut, bOutAdj)

        wHiddenAdj := new(mat.Dense)
        wHiddenAdj.Mul(x.T(), dHiddenLayer)
        wHiddenAdj.Scale(nn.config.learningRate, wHiddenAdj)
        wHidden.Add(wHidden, wHiddenAdj)

        bHiddenAdj, err := sumAlongAxis(0, dHiddenLayer)
        if err != nil {
            return err
        }
        bHiddenAdj.Scale(nn.config.learningRate, bHiddenAdj)
        bHidden.Add(bHidden, bHiddenAdj)
    }
    return nil
}
```

**Bias broadcasting**: gonum has no automatic broadcasting. The closure `func(_, col int, v float64) float64 { return v + bHidden.At(0, col) }` adds the correct bias value column by column. This is the idiomatic gonum pattern — the row index is ignored because the bias is the same for every sample.

**Bias gradient (`sumAlongAxis`)**: the bias update requires summing `dOutput` across all samples (axis 0) to produce one gradient per output neuron:

```go
func sumAlongAxis(axis int, m *mat.Dense) (*mat.Dense, error) {
    numRows, numCols := m.Dims()
    switch axis {
    case 0:
        data := make([]float64, numCols)
        for i := 0; i < numCols; i++ {
            data[i] = floats.Sum(mat.Col(nil, i, m))
        }
        return mat.NewDense(1, numCols, data), nil
    case 1:
        data := make([]float64, numRows)
        for i := 0; i < numRows; i++ {
            data[i] = floats.Sum(mat.Row(nil, i, m))
        }
        return mat.NewDense(numRows, 1, data), nil
    default:
        return nil, errors.New("invalid axis, must be 0 or 1")
    }
}
```

`floats.Sum` from `gonum.org/v1/gonum/floats` sums a `[]float64` slice.

**In-place weight updates**: `wOut.Add(wOut, wOutAdj)` mutates `wOut` directly. The weight matrices live across all epochs and you want to accumulate updates into them. Everything else (`hiddenLayerInput`, `dOutput`, etc.) is allocated fresh per epoch with `new(mat.Dense)`. For iris-scale data this is fine. For MNIST you'd pre-allocate scratch matrices once and reuse them.

## golang-pro: adding context

The original `backpropagate` has no cancellation mechanism — it runs to completion or returns an error from `sumAlongAxis`. For training jobs running thousands of epochs, you want to be able to cancel mid-run. The golang-pro approach:

```go
func (nn *neuralNet) backpropagate(ctx context.Context, x, y, wHidden, bHidden, wOut, bOut, output *mat.Dense) error {
    applySigmoid      := func(_, _ int, v float64) float64 { return sigmoid(v) }
    applySigmoidPrime := func(_, _ int, v float64) float64 { return sigmoidPrime(v) }

    for i := 0; i < nn.config.numEpochs; i++ {
        select {
        case <-ctx.Done():
            return fmt.Errorf("training cancelled at epoch %d: %w", i, ctx.Err())
        default:
        }
        // ... rest of loop unchanged
    }
    return nil
}
```

Check `ctx.Done()` once per epoch, not inside the matrix operations. That's the right granularity — you don't need microsecond cancellation latency during training; epoch-level is enough. See [[Goroutines]] for the lifecycle pattern and [[Select and Sync]] for context propagation.

## Loading data — and an off-by-one bug

```go
func makeInputsAndLabels(fileName string) (*mat.Dense, *mat.Dense) {
    f, err := os.Open(fileName)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    reader := csv.NewReader(f)
    reader.FieldsPerRecord = 7
    rawCSVData, err := reader.ReadAll()
    if err != nil {
        log.Fatal(err)
    }

    inputsData := make([]float64, 4*len(rawCSVData))  // BUG: includes header row
    labelsData := make([]float64, 3*len(rawCSVData))  // BUG: includes header row

    var inputsIndex, labelsIndex int
    for idx, record := range rawCSVData {
        if idx == 0 { continue }  // skip header, but matrix was sized for it
        for i, val := range record {
            parsedVal, err := strconv.ParseFloat(val, 64)
            if err != nil { log.Fatal(err) }
            if i == 4 || i == 5 || i == 6 {
                labelsData[labelsIndex] = parsedVal
                labelsIndex++
            } else {
                inputsData[inputsIndex] = parsedVal
                inputsIndex++
            }
        }
    }

    inputs := mat.NewDense(len(rawCSVData), 4, inputsData)  // 121 rows for train, not 120
    labels := mat.NewDense(len(rawCSVData), 3, labelsData)
    return inputs, labels
}
```

`len(rawCSVData)` is 121 for `train.csv` (120 data rows + 1 header). The matrices are sized for 121 rows, but only 120 rows are filled. The last row of both matrices is all zeros — a phantom sample. The network still trains fine because one spurious zero row among 121 is negligible, but it IS a bug.

The correct fix:

```go
inputs := mat.NewDense(len(rawCSVData)-1, 4, inputsData[:4*(len(rawCSVData)-1)])
labels := mat.NewDense(len(rawCSVData)-1, 3, labelsData[:3*(len(rawCSVData)-1)])
```

The `log.Fatal` inside a library-style function is also an anti-pattern — see [[Go Error Handling]]. In production code this should return `(*mat.Dense, *mat.Dense, error)`.

## Inference

```go
func (nn *neuralNet) predict(x *mat.Dense) (*mat.Dense, error) {
    if nn.wHidden == nil || nn.wOut == nil {
        return nil, errors.New("the supplied weights are empty")
    }
    if nn.bHidden == nil || nn.bOut == nil {
        return nil, errors.New("the supplied biases are empty")
    }

    output := new(mat.Dense)
    applySigmoid := func(_, _ int, v float64) float64 { return sigmoid(v) }

    hiddenLayerInput := new(mat.Dense)
    hiddenLayerInput.Mul(x, nn.wHidden)
    addBHidden := func(_, col int, v float64) float64 { return v + nn.bHidden.At(0, col) }
    hiddenLayerInput.Apply(addBHidden, hiddenLayerInput)

    hiddenLayerActivations := new(mat.Dense)
    hiddenLayerActivations.Apply(applySigmoid, hiddenLayerInput)

    outputLayerInput := new(mat.Dense)
    outputLayerInput.Mul(hiddenLayerActivations, nn.wOut)
    addBOut := func(_, col int, v float64) float64 { return v + nn.bOut.At(0, col) }
    outputLayerInput.Apply(addBOut, outputLayerInput)
    output.Apply(applySigmoid, outputLayerInput)

    return output, nil
}
```

The nil checks at the top prevent calling predict on an untrained network. Returning a typed error is the right Go pattern here.

## Accuracy evaluation

```go
var truePosNeg int
numPreds, _ := predictions.Dims()

for i := 0; i < numPreds; i++ {
    // Find the true label index from the one-hot test labels
    labelRow := mat.Row(nil, i, testLabels)
    var trueLabel int
    for idx, val := range labelRow {
        if val == 1.0 {
            trueLabel = idx
            break
        }
    }
    // Predicted class is the output neuron with the highest activation
    if predictions.At(i, trueLabel) == floats.Max(mat.Row(nil, i, predictions)) {
        truePosNeg++
    }
}
accuracy := float64(truePosNeg) / float64(numPreds)
fmt.Printf("\nAccuracy = %0.2f\n\n", accuracy)
```

`floats.Max` returns the maximum value in a `[]float64`. The predicted class is the index of the highest output neuron. The variable the original code names `prediction` actually holds the true label index — a confusing name that's worth fixing in any fork.

## Running it

```go
config := neuralNetConfig{
    inputNeurons:  4,    // sepal length, sepal width, petal length, petal width
    outputNeurons: 3,    // setosa, virginica, versicolor
    hiddenNeurons: 3,
    numEpochs:     5000,
    learningRate:  0.3,
}

network := newNetwork(config)
inputs, labels := makeInputsAndLabels("data/train.csv")

if err := network.train(inputs, labels); err != nil {
    log.Fatal(err)
}

testInputs, testLabels := makeInputsAndLabels("data/test.csv")
predictions, err := network.predict(testInputs)
if err != nil {
    log.Fatal(err)
}
```

5000 epochs on 120 (effectively 121) training samples. Result: **~97% accuracy** on 30 test samples. The `[4, 3, 3]` architecture works here because iris is nearly linearly separable in the original feature space — three hidden neurons are enough to partition it.

## Matrix dimension reference

Keep this when implementing by hand. Every `Mul` follows `(m × k) · (k × n) → (m × n)`:

| Matrix | Shape |
|--------|-------|
| `x` (input) | `[samples × 4]` |
| `wHidden` | `[4 × 3]` |
| `bHidden` | `[1 × 3]` |
| `hiddenLayerActivations` | `[samples × 3]` |
| `wOut` | `[3 × 3]` |
| `bOut` | `[1 × 3]` |
| `output` | `[samples × 3]` |

Gonum panics on dimension mismatches — no silent wrong answers.

## MNIST (sausheong blog, not gophernet)

The `gophernet` repo only covers iris. The MNIST extension (200 hidden neurons, 5 epochs, 60,000 training samples → 97.72% accuracy) comes from the sausheong blog post. The key difference is that MNIST requires normalization since pixel values aren't pre-scaled:

```go
// Map [0, 255] → [0.01, 0.99] to avoid sigmoid saturation at exact 0 and 1
normalized := (float64(pixel)/255.0)*0.99 + 0.01
```

One-hot targets use `0.99` and `0.01` for the same reason. The network architecture otherwise transfers directly — only `inputNeurons`, `hiddenNeurons`, and `numEpochs` change in the config.

## Complete working script

All code from this note is compiled into a single buildable file at `Golang/4_Project/neural_net_iris/main.go`. It includes all three bug fixes (sigmoidPrime, off-by-one, log.Fatal) and the context-aware backpropagate. Run it with the data from the original gophernet repo:

```bash
# From the project directory
cp -r /path/to/gophernet/data ./data
go mod tidy
go run main.go
# → Accuracy = 0.97
```

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[AI-ML Neural Network Foundations]] — the math behind every step
- see:: [[AI-ML Go Data Science Tooling]] — gonum and the broader ecosystem
- see:: [[Machine Learning]] — loss, supervised learning, classification
- see:: [[Training]] — gradient descent, backprop as Jacobian chain, learning rate
- see:: [[Model Components]] — fully connected layers (§4.2), activation functions (§4.3)
- see:: [[Go Error Handling]] — why log.Fatal inside helpers is wrong
- see:: [[Goroutines]] — context-based lifecycle for training loops
- see:: [[Select and Sync]] — ctx.Done() pattern in long-running loops

**Terms**
- gonum, mat.Dense, RawMatrix, forward pass, backpropagation, sumAlongAxis, bias broadcasting, sigmoid bug, off-by-one, in-place update, floats.Max, one-hot encoding, iris dataset, MNIST, pre-normalized data, context cancellation

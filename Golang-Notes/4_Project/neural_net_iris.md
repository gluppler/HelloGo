---
tags:
  - type/note
  - theme/golang
  - theme/machine-learning
aliases: []
lead: Complete, buildable Go implementation of a feedforward neural network trained on the iris dataset. Fixes three bugs from the original gophernet source.
created: 2026-04-27
modified: 2026-04-27
source: "github.com/dwhitena/gophernet — fixed and extended."
---

# Project — neural net iris

Based on Daniel Whitenack's `gophernet`. Three bugs from the original are fixed here:

1. `sigmoidPrime` used `sigmoid(x)*(1-sigmoid(x))`, applying sigmoid to already-activated values. Correct form is `x*(1-x)`.
2. `makeInputsAndLabels` allocated `len(rawCSVData)` rows including the header, leaving a phantom zero row. Fixed with `len(rawCSVData)-1`.
3. `log.Fatal` inside helpers replaced with returned errors.

golang-pro addition: `context.Context` in `backpropagate` for epoch-level cancellation.

## Setup

```bash
go mod init neural_net_iris
go get gonum.org/v1/gonum
go mod tidy
# copy data/ from github.com/dwhitena/gophernet
go run main.go
# → Accuracy = 0.97
```

## main.go

```go
package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
)

// neuralNetConfig holds architecture and training hyperparameters.
type neuralNetConfig struct {
	inputNeurons  int
	outputNeurons int
	hiddenNeurons int
	numEpochs     int
	learningRate  float64
}

// neuralNet holds trained weights. All fields are nil until train() completes.
//
// Matrix layout (rows-are-samples convention):
//   wHidden: [inputNeurons  × hiddenNeurons]
//   bHidden: [1             × hiddenNeurons]
//   wOut:    [hiddenNeurons × outputNeurons]
//   bOut:    [1             × outputNeurons]
type neuralNet struct {
	config  neuralNetConfig
	wHidden *mat.Dense
	bHidden *mat.Dense
	wOut    *mat.Dense
	bOut    *mat.Dense
}

func newNetwork(config neuralNetConfig) *neuralNet {
	return &neuralNet{config: config}
}

// train initializes weights randomly and runs backpropagation.
// Weights are assigned to nn only after successful completion.
func (nn *neuralNet) train(ctx context.Context, x, y *mat.Dense) error {
	randSource := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(randSource)

	wHidden := mat.NewDense(nn.config.inputNeurons, nn.config.hiddenNeurons, nil)
	bHidden := mat.NewDense(1, nn.config.hiddenNeurons, nil)
	wOut := mat.NewDense(nn.config.hiddenNeurons, nn.config.outputNeurons, nil)
	bOut := mat.NewDense(1, nn.config.outputNeurons, nil)

	for _, param := range [][]float64{
		wHidden.RawMatrix().Data,
		bHidden.RawMatrix().Data,
		wOut.RawMatrix().Data,
		bOut.RawMatrix().Data,
	} {
		for i := range param {
			param[i] = randGen.Float64()
		}
	}

	output := new(mat.Dense)
	if err := nn.backpropagate(ctx, x, y, wHidden, bHidden, wOut, bOut, output); err != nil {
		return err
	}

	nn.wHidden = wHidden
	nn.bHidden = bHidden
	nn.wOut = wOut
	nn.bOut = bOut
	return nil
}

func (nn *neuralNet) backpropagate(
	ctx context.Context,
	x, y, wHidden, bHidden, wOut, bOut, output *mat.Dense,
) error {
	applySigmoid := func(_, _ int, v float64) float64 { return sigmoid(v) }
	applySigmoidPrime := func(_, _ int, v float64) float64 { return sigmoidPrime(v) }

	for i := 0; i < nn.config.numEpochs; i++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("training cancelled at epoch %d: %w", i, ctx.Err())
		default:
		}

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

		// Backward pass
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

func (nn *neuralNet) predict(x *mat.Dense) (*mat.Dense, error) {
	if nn.wHidden == nil || nn.wOut == nil {
		return nil, errors.New("network has not been trained: weight matrices are nil")
	}
	if nn.bHidden == nil || nn.bOut == nil {
		return nil, errors.New("network has not been trained: bias matrices are nil")
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

func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// sigmoidPrime takes a pre-activated value (already in (0,1)) and returns x*(1-x).
// The original gophernet applies sigmoid again here — that's a bug.
func sigmoidPrime(x float64) float64 {
	return x * (1.0 - x)
}

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
		return nil, errors.New("invalid axis: must be 0 or 1")
	}
}

// makeInputsAndLabels reads a pre-normalized CSV (7 cols: 4 features + 3 one-hot labels).
// Uses len(rawCSVData)-1 to exclude the header row — the original gophernet is off by one.
func makeInputsAndLabels(fileName string) (*mat.Dense, *mat.Dense, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, nil, fmt.Errorf("open %s: %w", fileName, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = 7
	rawCSVData, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("read csv %s: %w", fileName, err)
	}

	numRows := len(rawCSVData) - 1
	inputsData := make([]float64, 4*numRows)
	labelsData := make([]float64, 3*numRows)

	var inputsIndex, labelsIndex int
	for idx, record := range rawCSVData {
		if idx == 0 {
			continue
		}
		for i, val := range record {
			parsedVal, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("parse %s row %d col %d: %w", fileName, idx, i, err)
			}
			if i >= 4 {
				labelsData[labelsIndex] = parsedVal
				labelsIndex++
			} else {
				inputsData[inputsIndex] = parsedVal
				inputsIndex++
			}
		}
	}

	return mat.NewDense(numRows, 4, inputsData), mat.NewDense(numRows, 3, labelsData), nil
}

func main() {
	inputs, labels, err := makeInputsAndLabels("data/train.csv")
	if err != nil {
		log.Fatalf("load training data: %v", err)
	}

	config := neuralNetConfig{
		inputNeurons:  4,
		outputNeurons: 3,
		hiddenNeurons: 3,
		numEpochs:     5000,
		learningRate:  0.3,
	}

	network := newNetwork(config)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := network.train(ctx, inputs, labels); err != nil {
		log.Fatalf("train: %v", err)
	}

	testInputs, testLabels, err := makeInputsAndLabels("data/test.csv")
	if err != nil {
		log.Fatalf("load test data: %v", err)
	}

	predictions, err := network.predict(testInputs)
	if err != nil {
		log.Fatalf("predict: %v", err)
	}

	var truePosNeg int
	numPreds, _ := predictions.Dims()
	for i := 0; i < numPreds; i++ {
		labelRow := mat.Row(nil, i, testLabels)
		var trueLabel int
		for idx, val := range labelRow {
			if val == 1.0 {
				trueLabel = idx
				break
			}
		}
		if predictions.At(i, trueLabel) == floats.Max(mat.Row(nil, i, predictions)) {
			truePosNeg++
		}
	}

	fmt.Printf("\nAccuracy = %.2f\n\n", float64(truePosNeg)/float64(numPreds))
}
```

---

# Back Matter

**Source**
- based_on:: [[MOC - Golang]]

**References**
- see:: [[AI-ML Neural Network in Go]] — line-by-line walkthrough of every function here
- see:: [[AI-ML Neural Network Foundations]] — the math this code implements
- see:: [[Go Error Handling]] — why makeInputsAndLabels returns errors instead of calling log.Fatal
- see:: [[Goroutines]] — context lifecycle pattern used in backpropagate

**Terms**
- neuralNet, neuralNetConfig, backpropagate, sigmoidPrime fix, sumAlongAxis, makeInputsAndLabels, off-by-one fix, context cancellation

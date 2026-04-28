package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
)

const (
	Setosa     = "setosa"
	Versicolor = "versicolor"
	Virginica  = "virginica"
)

var classNames = []string{Setosa, Versicolor, Virginica}

type neuralNetConfig struct {
	inputNeurons  int
	outputNeurons int
	hiddenNeurons int
	numEpochs     int
	learningRate  float64
	seed         int64
	verbose      bool
	evaluate     bool
	testPath     string
}

type neuralNet struct {
	config  neuralNetConfig
	wHidden *mat.Dense
	bHidden *mat.Dense
	wOut    *mat.Dense
	bOut    *mat.Dense
}

type Prediction struct {
	Index       int
	Actual      string
	Predicted   string
	Confidence  float64
	Correct    bool
	Probabilities []float64
}

type EvaluationResult struct {
	Predictions     []Prediction
	ConfusionMatrix [][]int
	Accuracy     float64
	Precision    []float64
	Recall       []float64
	F1           []float64
	ClassCount   []int
	CorrectCount int
	TotalCount  int
}

func newNetwork(config neuralNetConfig) *neuralNet {
	return &neuralNet{config: config}
}

func (nn *neuralNet) train(ctx context.Context, x, y *mat.Dense) error {
	randSource := rand.NewSource(nn.config.seed)
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

	logInterval := nn.config.numEpochs / 10
	if logInterval < 1 {
		logInterval = 1
	}

	for i := 0; i < nn.config.numEpochs; i++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("training cancelled at epoch %d: %w", i, ctx.Err())
		default:
		}

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

		if nn.config.verbose && i%logInterval == 0 {
			loss := calculateLoss(networkError)
			progress := float64(i+1) / float64(nn.config.numEpochs) * 100
			fmt.Printf("Epoch %d/%d (%.0f%%) - Loss: %.6f\n", i+1, nn.config.numEpochs, progress, loss)
		}
	}

	if nn.config.verbose {
		networkError := new(mat.Dense)
		networkError.Sub(y, output)
		finalLoss := calculateLoss(networkError)
		fmt.Printf("Training complete - Final Loss: %.6f\n", finalLoss)
	}

	return nil
}

func calculateLoss(output *mat.Dense) float64 {
	data := output.RawMatrix().Data
	sum := 0.0
	for _, v := range data {
		sum += v * v
	}
	n := float64(len(data))
	return sum / n
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

func (nn *neuralNet) Evaluate(x, y *mat.Dense) (*EvaluationResult, error) {
	predictions, err := nn.predict(x)
	if err != nil {
		return nil, err
	}

	numSamples, numClasses := predictions.Dims()
	result := &EvaluationResult{
		Predictions:    make([]Prediction, numSamples),
		ConfusionMatrix: make([][]int, numClasses),
		Precision:     make([]float64, numClasses),
		Recall:        make([]float64, numClasses),
		F1:            make([]float64, numClasses),
		ClassCount:    make([]int, numClasses),
	}

	for i := range result.ConfusionMatrix {
		result.ConfusionMatrix[i] = make([]int, numClasses)
	}

	trueLabels := make([]int, numSamples)
	for i := 0; i < numSamples; i++ {
		row := mat.Row(nil, i, y)
		for idx, val := range row {
			if val == 1.0 {
				trueLabels[i] = idx
				break
			}
		}
	}

	predictedLabels := make([]int, numSamples)
	maxProbs := make([]float64, numSamples)
	for i := 0; i < numSamples; i++ {
		row := mat.Row(nil, i, predictions)
		maxIdx := 0
		maxVal := row[0]
		for idx, val := range row {
			if val > maxVal {
				maxVal = val
				maxIdx = idx
			}
			result.Predictions[i].Probabilities = append(result.Predictions[i].Probabilities, val)
		}
		predictedLabels[i] = maxIdx
		maxProbs[i] = maxVal
	}

	result.TotalCount = numSamples

	for i := 0; i < numSamples; i++ {
		actualIdx := trueLabels[i]
		predictedIdx := predictedLabels[i]

		result.Predictions[i] = Prediction{
			Index:       i + 1,
			Actual:      classNames[actualIdx],
			Predicted:   classNames[predictedIdx],
			Confidence: maxProbs[i],
			Correct:    actualIdx == predictedIdx,
			Probabilities: result.Predictions[i].Probabilities,
		}

		result.ConfusionMatrix[actualIdx][predictedIdx]++
		result.ClassCount[actualIdx]++

		if actualIdx == predictedIdx {
			result.CorrectCount++
		}
	}

	result.Accuracy = float64(result.CorrectCount) / float64(result.TotalCount)

	for c := 0; c < numClasses; c++ {
		tp := result.ConfusionMatrix[c][c]
		fp := 0
		for r := 0; r < numClasses; r++ {
			if r != c {
				fp += result.ConfusionMatrix[r][c]
			}
		}
		fn := 0
		for c2 := 0; c2 < numClasses; c2++ {
			if c2 != c {
				fn += result.ConfusionMatrix[c][c2]
			}
		}

		if tp+fp > 0 {
			result.Precision[c] = float64(tp) / float64(tp+fp)
		}
		if tp+fn > 0 {
			result.Recall[c] = float64(tp) / float64(tp+fn)
		}
		if result.Precision[c]+result.Recall[c] > 0 {
			result.F1[c] = 2 * result.Precision[c] * result.Recall[c] / (result.Precision[c] + result.Recall[c])
		}
	}

	return result, nil
}

func (r *EvaluationResult) PrintDetailed(out *bufio.Writer) {
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "═══════════════════════════════════════════════════════")
	fmt.Fprintln(out, "           DETAILED EVALUATION RESULTS              ")
	fmt.Fprintln(out, "═══════════════════════════════════════════════════════")
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "───────────────────────────────────────────────────────")
	fmt.Fprintln(out, "         PER-SAMPLE PREDICTIONS                   ")
	fmt.Fprintln(out, "───────────────────────────────────────────────────────")
	fmt.Fprintf(out, "%-6s %-12s %-12s %-10s %-8s\n", "#", "ACTUAL", "PREDICTED", "CONFIDENCE", "STATUS")
	fmt.Fprintln(out, strings.Repeat("-", 55))

	for _, p := range r.Predictions {
		status := "✓"
		if !p.Correct {
			status = "✗"
		}
		fmt.Fprintf(out, "%-6d %-12s %-12s %-9.2f %-8s\n",
			p.Index, p.Actual, p.Predicted, p.Confidence, status)
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "───────────────────────────────────────────────────────")
	fmt.Fprintln(out, "           CONFUSION MATRIX                   ")
	fmt.Fprintln(out, "──────────���─���──────────────────────────────────────────")
	fmt.Fprintf(out, "%-12s", "")
	for _, name := range classNames {
		fmt.Fprintf(out, " %-10s", name)
	}
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, strings.Repeat("-", 50))
	for i, name := range classNames {
		fmt.Fprintf(out, "%-12s", name)
		for j := range classNames {
			fmt.Fprintf(out, " %-10d", r.ConfusionMatrix[i][j])
		}
		fmt.Fprintln(out, "")
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "───────────────────────────────────────────────────────")
	fmt.Fprintln(out, "           PERFORMANCE METRICS                 ")
	fmt.Fprintln(out, "───────────────────────────────────────────────────────")
	fmt.Fprintf(out, "Overall Accuracy: %.2f%% (%d/%d correct)\n\n",
		r.Accuracy*100, r.CorrectCount, r.TotalCount)

	fmt.Fprintln(out, "Per-Class Metrics:")
	fmt.Fprintf(out, "%-12s %-10s %-10s %-10s\n", "CLASS", "PRECISION", "RECALL", "F1-SCORE")
	fmt.Fprintln(out, strings.Repeat("-", 45))
	for i, name := range classNames {
		fmt.Fprintf(out, "%-12s %-9.2f%% %-9.2f%% %-9.2f%%\n",
			name, r.Precision[i]*100, r.Recall[i]*100, r.F1[i]*100)
	}

	avgPrecision := 0.0
	avgRecall := 0.0
	avgF1 := 0.0
	for i := range classNames {
		avgPrecision += r.Precision[i]
		avgRecall += r.Recall[i]
		avgF1 += r.F1[i]
	}
	avgPrecision /= float64(len(classNames))
	avgRecall /= float64(len(classNames))
	avgF1 /= float64(len(classNames))

	fmt.Fprintln(out, strings.Repeat("-", 45))
	fmt.Fprintf(out, "%-12s %-9.2f%% %-9.2f%% %-9.2f%%\n",
		"AVERAGE", avgPrecision*100, avgRecall*100, avgF1*100)

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "═══════════════════════════════════════════════════════")
}

func (r *EvaluationResult) PrintPredictions(out *bufio.Writer) {
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "───────────────────────────────────────────────────────")
	fmt.Fprintln(out, "         PREDICTION BREAKDOWN                      ")
	fmt.Fprintln(out, "───────────────────────────────────────────────────────")

	byClass := make(map[string][]Prediction)
	for _, p := range r.Predictions {
		byClass[p.Actual] = append(byClass[p.Actual], p)
	}

	for _, class := range classNames {
		preds := byClass[class]
		if len(preds) == 0 {
			continue
		}

		correct := 0
		for _, p := range preds {
			if p.Correct {
				correct++
			}
		}
		accuracy := float64(correct) / float64(len(preds)) * 100

		fmt.Fprintf(out, "\n[%s] - %d samples, %.0f%% accuracy\n", class, len(preds), accuracy)
		fmt.Fprintln(out, strings.Repeat("-", 40))

		for _, p := range preds {
			status := "✓ CORRECT"
			if !p.Correct {
				status = "✗ WRONG (predicted: " + p.Predicted + ")"
			}
			fmt.Fprintf(out, "  Sample %d: %s (%.2f%%)\n", p.Index, status, p.Confidence*100)
		}
	}
	fmt.Fprintln(out, "")
}

func sigmoid(x float64) float64 {
	if x < -700 {
		return 0
	}
	if x > 700 {
		return 1
	}
	return 1.0 / (1.0 + math.Exp(-x))
}

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

func printHeader(out *bufio.Writer) {
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "╔═══════════════════════════════════════════════════╗")
	fmt.Fprintln(out, "║      IRIS NEURAL NETWORK CLASSIFIER             ║")
	fmt.Fprintln(out, "║         Training & Evaluation                  ║")
	fmt.Fprintln(out, "╚═══════════════════════════════════════════════════╝")
}

func main() {
	epochs := flag.Int("epochs", 5000, "number of training epochs")
	learningRate := flag.Float64("rate", 0.3, "learning rate")
	hiddenNeurons := flag.Int("hidden", 3, "number of hidden neurons")
	seed := flag.Int64("seed", 0, "random seed for reproducibility (0 = use time-based)")
	verbose := flag.Bool("v", false, "verbose output with training progress")
	evaluate := flag.Bool("eval", false, "run detailed evaluation on test set")
	predict := flag.Int("predict", -1, "predict a specific test sample by index")
	dataPath := flag.String("data", "data", "path to data directory")

	flag.Parse()

	seedValue := *seed
	if seedValue == 0 {
		seedValue = time.Now().UnixNano()
	}

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	printHeader(out)

	if *verbose {
		fmt.Fprintf(out, "\n=== Neural Network Training ===\n")
		fmt.Fprintf(out, "Epochs: %d, Learning Rate: %.2f, Hidden Neurons: %d\n", *epochs, *learningRate, *hiddenNeurons)
		fmt.Fprintf(out, "Random Seed: %d\n", seedValue)
		fmt.Fprintf(out, "=========================\n\n")
	}

	inputs, labels, err := makeInputsAndLabels(*dataPath + "/train.csv")
	if err != nil {
		log.Fatalf("load training data: %v", err)
	}

	config := neuralNetConfig{
		inputNeurons:  4,
		outputNeurons: 3,
		hiddenNeurons: *hiddenNeurons,
		numEpochs:     *epochs,
		learningRate: *learningRate,
		seed:        seedValue,
		verbose:     *verbose,
	}

	network := newNetwork(config)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	numTrainSamples, _ := inputs.Dims()
	fmt.Fprintf(out, "Training on %d samples...\n", numTrainSamples)
	if err := network.train(ctx, inputs, labels); err != nil {
		log.Fatalf("train: %v", err)
	}

	testInputs, testLabels, err := makeInputsAndLabels(*dataPath + "/test.csv")
	if err != nil {
		log.Fatalf("load test data: %v", err)
	}

	numTestSamples, _ := testInputs.Dims()
	fmt.Fprintf(out, "Evaluating on %d test samples...\n", numTestSamples)

	if *predict >= 0 {
		idx := *predict
		if idx < 0 || idx >= numTestSamples {
			log.Fatalf("invalid sample index: %d", idx)
		}

		row := testInputs.RawMatrix().Data[idx*4 : (idx+1)*4]
		singleInput := mat.NewDense(1, 4, row)
		predictions, err := network.predict(singleInput)
		if err != nil {
			log.Fatalf("predict: %v", err)
		}

		probs := predictions.RawMatrix().Data
		maxIdx := 0
		maxVal := probs[0]
		for i, v := range probs {
			if v > maxVal {
				maxVal = v
				maxIdx = i
			}
		}

		actualRow := testLabels.RawMatrix().Data[idx*4 : (idx+1)*4]
		actualIdx := 0
		for i, v := range actualRow {
			if v == 1.0 {
				actualIdx = i
				break
			}
		}

		fmt.Fprintf(out, "\n═══════════════════════════════════════════════════════\n")
		fmt.Fprintf(out, "  SINGLE PREDICTION FOR SAMPLE #%d\n", idx+1)
		fmt.Fprintf(out, "═══════════════════════════════════════════════════════\n\n")
		fmt.Fprintf(out, "Input Features:\n")
		fmt.Fprintf(out, "  Sepal Length: %.4f\n", row[0])
		fmt.Fprintf(out, "  Sepal Width:  %.4f\n", row[1])
		fmt.Fprintf(out, "  Petal Length: %.4f\n", row[2])
		fmt.Fprintf(out, "  Petal Width:  %.4f\n\n", row[3])
		fmt.Fprintf(out, "Predictions:\n")
		for i, name := range classNames {
			fmt.Fprintf(out, "  %-12s: %.2f%%\n", name, probs[i]*100)
		}
		fmt.Fprintf(out, "\n───────────────────────────────────────────────────\n")
		fmt.Fprintf(out, "Result: %s (%.2f%% confidence)\n", classNames[maxIdx], maxVal*100)
		fmt.Fprintf(out, "Actual: %s\n", classNames[actualIdx])
		if classNames[maxIdx] == classNames[actualIdx] {
			fmt.Fprintf(out, "Status: ✓ CORRECT\n")
		} else {
			fmt.Fprintf(out, "Status: ✗ INCORRECT\n")
		}
		fmt.Fprintf(out, "───────────────────────────────────────────────────\n\n")
		out.Flush()
		os.Exit(0)
	}

	result, err := network.Evaluate(testInputs, testLabels)
	if err != nil {
		log.Fatalf("evaluate: %v", err)
	}

	if *evaluate {
		result.PrintDetailed(out)
		result.PrintPredictions(out)
	} else {
		fmt.Fprintf(out, "\n═══════════════════════════════════════════════════════\n")
		fmt.Fprintf(out, "           EVALUATION SUMMARY                      \n")
		fmt.Fprintf(out, "═══════════════════════════════════════════════\n\n")
		fmt.Fprintf(out, "Overall Accuracy: %.2f%% (%d/%d)\n\n",
			result.Accuracy*100, result.CorrectCount, result.TotalCount)

		fmt.Fprintf(out, "Per-Class Performance:\n")
		fmt.Fprintf(out, "%-12s %-10s %-10s %-10s\n", "CLASS", "PRECISION", "RECALL", "F1")
		fmt.Fprintln(out, strings.Repeat("-", 45))
		for i, name := range classNames {
			fmt.Fprintf(out, "%-12s %-9.2f%% %-9.2f%% %-9.2f%%\n",
				name, result.Precision[i]*100, result.Recall[i]*100, result.F1[i]*100)
		}
	}

	fmt.Fprintln(out, "")
	out.Flush()
}
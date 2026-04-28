package main

import (
	"context"
	"errors"
	"os"
	"testing"

	"gonum.org/v1/gonum/mat"
)

func TestSigmoid(t *testing.T) {
	tests := []struct {
		name     string
		input   float64
		want    float64
		tolerance float64
	}{
		{"zero", 0.0, 0.5, 1e-6},
		{"one", 1.0, 0.7310585786300049, 1e-6},
		{"negative one", -1.0, 0.2689414213699951, 1e-6},
		{"large positive", 10.0, 0.9999546001392376, 1e-6},
		{"large negative", -10.0, 4.539786871165428e-05, 1e-6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sigmoid(tt.input)
			diff := got - tt.want
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.tolerance {
				t.Errorf("sigmoid(%v) = %v, want %v (+/- %v)", tt.input, got, tt.want, tt.tolerance)
			}
		})
	}
}

func TestSigmoidPrime(t *testing.T) {
	tests := []struct {
		name     string
		input   float64
		want    float64
		tolerance float64
	}{
		{"zero", 0.0, 0.0, 1e-6},
		{"half", 0.5, 0.25, 1e-6},
		{"one", 1.0, 0.0, 1e-6},
		{"quarter", 0.25, 0.1875, 1e-6},
		{"three quarters", 0.75, 0.1875, 1e-6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sigmoidPrime(tt.input)
			diff := got - tt.want
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.tolerance {
				t.Errorf("sigmoidPrime(%v) = %v, want %v (+/- %v)", tt.input, got, tt.want, tt.tolerance)
			}
		})
	}
}

func TestSumAlongAxis(t *testing.T) {
	data := []float64{
		1, 2, 3,
		4, 5, 6,
	}
	m := mat.NewDense(2, 3, data)

	tests := []struct {
		name       string
		axis      int
		wantRows  int
		wantCols int
		wantErr  bool
		check    func(*testing.T, *mat.Dense)
	}{
		{
			name:      "axis 0",
			axis:      0,
			wantRows:  1,
			wantCols:  3,
			check: func(t *testing.T, m *mat.Dense) {
				r, c := m.Dims()
				if r != 1 || c != 3 {
					t.Errorf("dims = (%d, %d), want (1, 3)", r, c)
				}
			},
		},
		{
			name:      "axis 1",
			axis:      1,
			wantRows:  2,
			wantCols:  1,
			check: func(t *testing.T, m *mat.Dense) {
				r, c := m.Dims()
				if r != 2 || c != 1 {
					t.Errorf("dims = (%d, %d), want (2, 1)", r, c)
				}
			},
		},
		{
			name:      "invalid axis",
			axis:      2,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sumAlongAxis(tt.axis, m)
			if tt.wantErr {
				if err == nil {
					t.Error("want error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestMakeInputsAndLabels(t *testing.T) {
	tests := []struct {
		name    string
		path   string
		wantRows int
		wantInputsCols int
		wantLabelsCols int
		wantErr bool
	}{
		{
			name:          "train data",
			path:          "data/train.csv",
			wantRows:     120,
			wantInputsCols: 4,
			wantLabelsCols: 3,
		},
		{
			name:          "test data",
			path:          "data/test.csv",
			wantRows:     30,
			wantInputsCols: 4,
			wantLabelsCols: 3,
		},
		{
			name:          "nonexistent file",
			path:          "data/nonexistent.csv",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputs, labels, err := makeInputsAndLabels(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Error("want error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if inputs == nil || labels == nil {
				t.Fatal("inputs or labels is nil")
			}
			inRows, inCols := inputs.Dims()
			labRows, labCols := labels.Dims()
			if inRows != tt.wantRows || inCols != tt.wantInputsCols {
				t.Errorf("inputs dims = (%d, %d), want (%d, %d)", inRows, inCols, tt.wantRows, tt.wantInputsCols)
			}
			if labRows != tt.wantRows || labCols != tt.wantLabelsCols {
				t.Errorf("labels dims = (%d, %d), want (%d, %d)", labRows, labCols, tt.wantRows, tt.wantLabelsCols)
			}
		})
	}
}

func TestNeuralNetTrain(t *testing.T) {
	inputs, labels, err := makeInputsAndLabels("data/train.csv")
	if err != nil {
		t.Fatalf("failed to load test data: %v", err)
	}

	weightsBefore := func(nn *neuralNet) []float64 {
		if nn.wHidden == nil {
			return nil
		}
		return nn.wHidden.RawMatrix().Data
	}

	tests := []struct {
		name      string
		epochs    int
		checkConverged bool
	}{
		{"10 epochs", 10, false},
		{"100 epochs", 100, false},
		{"1000 epochs", 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := neuralNetConfig{
				inputNeurons:  4,
				outputNeurons: 3,
				hiddenNeurons: 3,
				numEpochs:     tt.epochs,
				learningRate:  0.3,
			}
			nn := newNetwork(config)

			ctx := context.Background()
			if err := nn.train(ctx, inputs, labels); err != nil {
				t.Fatalf("train failed: %v", err)
			}

			if nn.wHidden == nil || nn.wOut == nil {
				t.Error("weights not initialized after train")
			}

			if tt.checkConverged {
				weights := weightsBefore(nn)
				if weights == nil {
					t.Fatal("weights are nil")
				}
				var sum float64
				for _, w := range weights {
					sum += w * w
				}
				if sum == 0 {
					t.Error("weights did not change during training")
				}
			}
		})
	}
}

func TestNeuralNetTrainContextCancel(t *testing.T) {
	inputs, labels, err := makeInputsAndLabels("data/train.csv")
	if err != nil {
		t.Fatalf("failed to load test data: %v", err)
	}

	config := neuralNetConfig{
		inputNeurons:  4,
		outputNeurons: 3,
		hiddenNeurons: 3,
		numEpochs:     100000,
		learningRate: 0.3,
	}
	nn := newNetwork(config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := nn.train(ctx, inputs, labels); err == nil {
		t.Error("expected error from cancelled context, got nil")
	}
}

func TestNeuralNetPredictWithoutTrain(t *testing.T) {
	config := neuralNetConfig{
		inputNeurons:  4,
		outputNeurons: 3,
		hiddenNeurons: 3,
	}
	nn := newNetwork(config)

	input := mat.NewDense(1, 4, []float64{0.5, 0.5, 0.5, 0.5})
	_, err := nn.predict(input)
	if err == nil {
		t.Error("expected error predicting without training, got nil")
	}
}

func TestNeuralNetPredictWithTrain(t *testing.T) {
	inputs, labels, err := makeInputsAndLabels("data/train.csv")
	if err != nil {
		t.Fatalf("failed to load train data: %v", err)
	}

	testInputs, testLabels, err := makeInputsAndLabels("data/test.csv")
	if err != nil {
		t.Fatalf("failed to load test data: %v", err)
	}

	config := neuralNetConfig{
		inputNeurons:  4,
		outputNeurons: 3,
		hiddenNeurons: 3,
		numEpochs:     1000,
		learningRate: 0.3,
	}
	nn := newNetwork(config)

	ctx := context.Background()
	if err := nn.train(ctx, inputs, labels); err != nil {
		t.Fatalf("train failed: %v", err)
	}

	predictions, err := nn.predict(testInputs)
	if err != nil {
		t.Fatalf("predict failed: %v", err)
	}

	numPreds, _ := predictions.Dims()
	var correct int
	for i := 0; i < numPreds; i++ {
		labelRow := mat.Row(nil, i, testLabels)
		var trueLabel int
		for idx, val := range labelRow {
			if val == 1.0 {
				trueLabel = idx
				break
			}
		}
		if predictions.At(i, trueLabel) >= 0.5 {
			correct++
		}
	}

	accuracy := float64(correct) / float64(numPreds)
	t.Logf("accuracy = %.2f", accuracy)

	if accuracy < 0.5 {
		t.Errorf("accuracy %.2f below threshold 0.5", accuracy)
	}
}

func TestNewNetwork(t *testing.T) {
	config := neuralNetConfig{
		inputNeurons:  4,
		outputNeurons: 3,
		hiddenNeurons: 5,
		numEpochs:     100,
		learningRate: 0.1,
	}
	nn := newNetwork(config)

	if nn.config.inputNeurons != 4 {
		t.Errorf("inputNeurons = %d, want 4", nn.config.inputNeurons)
	}
	if nn.config.outputNeurons != 3 {
		t.Errorf("outputNeurons = %d, want 3", nn.config.outputNeurons)
	}
	if nn.config.hiddenNeurons != 5 {
		t.Errorf("hiddenNeurons = %d, want 5", nn.config.hiddenNeurons)
	}
	if nn.wHidden != nil || nn.wOut != nil {
		t.Error("weights should be nil before training")
	}
}

func BenchmarkTrain(b *testing.B) {
	inputs, labels, err := makeInputsAndLabels("data/train.csv")
	if err != nil {
		b.Fatalf("failed to load test data: %v", err)
	}

	config := neuralNetConfig{
		inputNeurons:  4,
		outputNeurons: 3,
		hiddenNeurons: 3,
		numEpochs:     100,
		learningRate: 0.3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nn := newNetwork(config)
		ctx := context.Background()
		if err := nn.train(ctx, inputs, labels); err != nil {
			b.Fatalf("train failed: %v", err)
		}
	}
}

func init() {
	if _, err := os.Stat("data/train.csv"); errors.Is(err, os.ErrNotExist) {
		print("Warning: data files not found, integration tests may be skipped\n")
	}
}
# Neural Network Iris Classifier

A pure Go neural network implementation for multiclass classification, trained on the Iris dataset. Built from scratch using the gonum library for matrix operations.

## Features

- **From-scratch implementation** — No external ML libraries, just matrix operations with gonum
- **Backpropagation training** — Classic gradient descent learning algorithm
- **Detailed evaluation** — Confusion matrix, precision, recall, F1-score per class
- **Single sample prediction** — Predict individual samples with confidence scores
- **Training progress** — Verbose mode shows loss at each epoch
- **Reproducibility** — Fixed random seed for consistent results

## Quick Start

```bash
# Build
go build -o neural_net_iris .

# Run default (5000 epochs)
./neural_net_iris

# Run with specific seed for reproducibility
./neural_net_iris -seed=42

# Run with detailed evaluation
./neural_net_iris -seed=42 -eval
```

## Command-Line Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-epochs` | int | 5000 | Number of training epochs |
| `-rate` | float64 | 0.3 | Learning rate |
| `-hidden` | int | 3 | Number of hidden neurons |
| `-seed` | int64 | 0 | Random seed (0 = time-based) |
| `-v` | bool | false | Verbose output with training progress |
| `-eval` | bool | false | Detailed evaluation output |
| `-predict` | int | -1 | Predict specific test sample by index |
| `-data` | string | "data" | Path to data directory |

## Architecture

### Network Structure

```
Input Layer (4 neurons) → Hidden Layer (3 neurons) → Output Layer (3 neurons)
```

- **Input**: sepal_length, sepal_width, petal_length, petal_width
- **Hidden**: Configurable (default: 3 neurons)
- **Output**: 3-class one-hot encoding (setosa, versicolor, virginica)

### Matrix Layout

```
wHidden: [inputNeurons × hiddenNeurons]
bHidden: [1 × hiddenNeurons]
wOut:    [hiddenNeurons × outputNeurons]
bOut:    [1 × outputNeurons]
```

## Output Examples

### Basic Evaluation

```bash
./neural_net_iris -seed=42
```

```
╔═══════════════════════════════════════════════════╗
║      IRIS NEURAL NETWORK CLASSIFIER             ║
║         Training & Evaluation                  ║
╚═══════════════════════════════════════════════╝
Training on 120 samples...
Evaluating on 30 test samples...

═══════════════════════════════════════════════════════
           EVALUATION SUMMARY                      
═══════════════════════════════════════════════════════

Overall Accuracy: 100.00% (30/30)

Per-Class Performance:
CLASS        PRECISION  RECALL     F1        
─────────────────────────────────────────────
setosa       100.00   % 100.00   % 100.00   %
versicolor   100.00   % 100.00   % 100.00   %
virginica    100.00   % 100.00   % 100.00   %
```

### Detailed Evaluation

```bash
./neural_net_iris -seed=42 -eval
```

Shows:
- Per-sample predictions with ✓/✗ status
- Confusion matrix
- Precision/Recall/F1 per class
- Prediction breakdown by flower type

### Single Prediction

```bash
./neural_net_iris -seed=42 -predict=0
```

```
Input Features:
  Sepal Length: 0.5833
  Sepal Width:  0.2917
  Petal Length: 0.7288
  Petal Width:  0.7500

Predictions:
  setosa      : 0.00%
  versicolor  : 98.67%
  virginica   : 1.30%

Result: versicolor (98.67% confidence)
Actual: versicolor
Status: ✓ CORRECT
```

### Verbose Training

```bash
./neural_net_iris -seed=42 -epochs=1000 -v
```

```
Epoch 1/1000 (0%) - Loss: 0.482887
Epoch 101/1000 (10%) - Loss: 0.087321
...
Training complete - Final Loss: 0.018333
```

## Using with Other Datasets

### Data Format Requirements

The network expects CSV files with the following format:

```
feature_1,feature_2,feature_3,feature_4,class_1,class_2,class_3
0.5833,0.2917,0.7288,0.7500,0.0,1.0,0.0
```

| Column | Description |
|-------|-------------|
| 1-4 | Normalized features (float64, range 0-1) |
| 5-(N+4) | One-hot encoded labels (N classes) |

### Step-by-Step

1. **Prepare your data**: Normalize features to [0, 1] range

```go
// Example normalization
value := (value - min) / (max - min)
```

2. **Convert labels**: One-hot encode your target classes

3. **Create CSV files**: `train.csv` and `test.csv` with format:
   - Header row with column names
   - One row per sample

4. **Run training**:

```bash
./neural_net_iris -data=/path/to/your/data -epochs=5000 -seed=42
```

### Example: Custom 2-Class Dataset

For a 2-class problem (e.g., benign/malignant):

```csv
feature_1,feature_2,feature_3,feature_4,benign,malignant
0.5,0.3,0.7,0.8,1.0,0.0
0.2,0.1,0.9,0.95,0.0,1.0
```

Modify `neuralNetConfig` in code:

```go
config := neuralNetConfig{
    inputNeurons:  4,   // Your number of features
    outputNeurons: 2,   // Your number of classes
    hiddenNeurons: 3,   // Adjust based on complexity
    // ...
}
```

## Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with race detector
go test -race ./...

# Run benchmarks
go test -bench=. -benchmem ./...
```

## Technical Details

### Activation Functions

- **Sigmoid**: σ(x) = 1 / (1 + e^-x)
- **Sigmoid Prime**: σ'(x) = x * (1 - x)

### Training Algorithm

1. Initialize weights randomly
2. For each epoch:
   - Forward pass: input → hidden → output
   - Calculate error: target - output
   - Backward pass: output → hidden → input
   - Update weights with learning rate
3. Return trained network

### Hyperparameters

| Parameter | Recommended Range | Description |
|-----------|------------------|--------------|
| Learning Rate | 0.1 - 0.9 | Step size for weight updates |
| Hidden Neurons | 2 - 10 | Network complexity |
| Epochs | 1000 - 10000 | Training iterations |

## Troubleshooting

### Low Accuracy

- **Increase epochs**: Try `-epochs=10000`
- **Adjust learning rate**: Try `-rate=0.1` or `-rate=0.5`
- **Increase hidden neurons**: Try `-hidden=5` or `-hidden=10`

### Different Results Each Run

Use fixed seed for reproducibility:

```bash
./neural_net_iris -seed=42  # Same seed = Same results
```

### Nan or Inf Loss

- Reduce learning rate: `-rate=0.1`
- Check data normalization (all values should be 0-1)

## Dependencies

- **gonum** (`gonum.org/v1/gonum`) — Matrix operations
- Go 1.26.2+

## License

MIT License
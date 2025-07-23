package ml

import (
	"math"
	"testing"
)

func TestNewNeuralNetwork(t *testing.T) {
	config := &NeuralNetworkConfig{
		InputSize:    10,
		HiddenLayers: []int{8, 6},
		OutputSize:   3,
		LearningRate: 0.01,
		Momentum:     0.9,
		Epochs:       100,
		BatchSize:    16,
		DropoutRate:  0.2,
		ActivationFunction: "relu",
	}

	nn := NewNeuralNetwork(config)

	if nn == nil {
		t.Fatal("Neural network is nil")
	}

	if len(nn.layers) != 3 { // 2 hidden + 1 output
		t.Errorf("Expected 3 layers, got %d", len(nn.layers))
	}

	if nn.config.InputSize != 10 {
		t.Errorf("Expected input size 10, got %d", nn.config.InputSize)
	}

	if nn.config.OutputSize != 3 {
		t.Errorf("Expected output size 3, got %d", nn.config.OutputSize)
	}

	if nn.trained {
		t.Error("Network should not be trained initially")
	}
}

func TestNeuralNetworkForward(t *testing.T) {
	config := &NeuralNetworkConfig{
		InputSize:    5,
		HiddenLayers: []int{4},
		OutputSize:   2,
		LearningRate: 0.01,
		ActivationFunction: "relu",
	}

	nn := NewNeuralNetwork(config)
	input := []float64{1.0, 0.5, -0.2, 0.8, -0.1}

	output, err := nn.Forward(input)
	if err != nil {
		t.Fatalf("Forward pass failed: %v", err)
	}

	if len(output) != 2 {
		t.Errorf("Expected output size 2, got %d", len(output))
	}

	// Check that outputs are valid numbers
	for i, val := range output {
		if math.IsNaN(val) || math.IsInf(val, 0) {
			t.Errorf("Output %d is invalid: %f", i, val)
		}
	}
}

func TestNeuralNetworkForwardInvalidInput(t *testing.T) {
	config := &NeuralNetworkConfig{
		InputSize:    5,
		HiddenLayers: []int{4},
		OutputSize:   2,
	}

	nn := NewNeuralNetwork(config)
	input := []float64{1.0, 0.5, -0.2} // Wrong size

	_, err := nn.Forward(input)
	if err == nil {
		t.Error("Expected error for invalid input size")
	}
}

func TestNeuralNetworkTrain(t *testing.T) {
	config := &NeuralNetworkConfig{
		InputSize:    3,
		HiddenLayers: []int{4},
		OutputSize:   2,
		LearningRate: 0.1,
		Epochs:       10,
		BatchSize:    2,
		ActivationFunction: "sigmoid",
	}

	nn := NewNeuralNetwork(config)

	// Create simple training data (XOR-like problem)
	features := [][]float64{
		{0, 0, 1},
		{0, 1, 1},
		{1, 0, 1},
		{1, 1, 1},
	}

	labels := [][]float64{
		{1, 0}, // Class 0
		{0, 1}, // Class 1
		{0, 1}, // Class 1
		{1, 0}, // Class 0
	}

	err := nn.Train(features, labels)
	if err != nil {
		t.Fatalf("Training failed: %v", err)
	}

	if !nn.trained {
		t.Error("Network should be marked as trained")
	}
}

func TestNeuralNetworkPredict(t *testing.T) {
	config := &NeuralNetworkConfig{
		InputSize:    3,
		HiddenLayers: []int{4},
		OutputSize:   2,
		LearningRate: 0.1,
		Epochs:       5,
		BatchSize:    2,
		ActivationFunction: "sigmoid",
	}

	nn := NewNeuralNetwork(config)

	// Train first
	features := [][]float64{
		{0, 0, 1},
		{1, 1, 1},
	}

	labels := [][]float64{
		{1, 0},
		{0, 1},
	}

	err := nn.Train(features, labels)
	if err != nil {
		t.Fatalf("Training failed: %v", err)
	}

	// Test prediction
	input := []float64{0, 0, 1}
	prediction, err := nn.Predict(input)
	if err != nil {
		t.Fatalf("Prediction failed: %v", err)
	}

	if len(prediction) != 2 {
		t.Errorf("Expected prediction size 2, got %d", len(prediction))
	}

	// Check that probabilities sum to approximately 1
	sum := 0.0
	for _, prob := range prediction {
		sum += prob
		if prob < 0 || prob > 1 {
			t.Errorf("Probability out of range [0,1]: %f", prob)
		}
	}

	if math.Abs(sum-1.0) > 0.01 {
		t.Errorf("Probabilities don't sum to 1: %f", sum)
	}
}

func TestNeuralNetworkPredictUntrained(t *testing.T) {
	config := &NeuralNetworkConfig{
		InputSize:    3,
		HiddenLayers: []int{4},
		OutputSize:   2,
	}

	nn := NewNeuralNetwork(config)
	input := []float64{0, 0, 1}

	_, err := nn.Predict(input)
	if err == nil {
		t.Error("Expected error when predicting with untrained network")
	}
}

func TestNeuralNetworkEvaluate(t *testing.T) {
	config := &NeuralNetworkConfig{
		InputSize:    2,
		HiddenLayers: []int{3},
		OutputSize:   2,
		LearningRate: 0.1,
		Epochs:       50,
		BatchSize:    4,
		ActivationFunction: "sigmoid",
	}

	nn := NewNeuralNetwork(config)

	// Create simple linearly separable data
	trainFeatures := [][]float64{
		{0, 0},
		{0, 1},
		{1, 0},
		{1, 1},
	}

	trainLabels := [][]float64{
		{1, 0}, // Class 0
		{1, 0}, // Class 0
		{0, 1}, // Class 1
		{0, 1}, // Class 1
	}

	// Train the network
	err := nn.Train(trainFeatures, trainLabels)
	if err != nil {
		t.Fatalf("Training failed: %v", err)
	}

	// Evaluate on the same data (just for testing)
	accuracy, err := nn.Evaluate(trainFeatures, trainLabels)
	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	if accuracy < 0 || accuracy > 1 {
		t.Errorf("Accuracy out of range [0,1]: %f", accuracy)
	}

	// With enough training, accuracy should be reasonable
	if accuracy < 0.25 { // At least better than random
		t.Errorf("Accuracy too low: %f", accuracy)
	}
}

func TestActivationFunctions(t *testing.T) {
	testCases := []struct {
		name     string
		function ActivationFunction
		input    float64
		minOut   float64
		maxOut   float64
	}{
		{"Sigmoid", &Sigmoid{}, 0.0, 0.4, 0.6},
		{"Sigmoid", &Sigmoid{}, 1.0, 0.7, 0.8},
		{"ReLU", &ReLU{}, 1.0, 1.0, 1.0},
		{"ReLU", &ReLU{}, -1.0, 0.0, 0.0},
		{"Tanh", &Tanh{}, 0.0, -0.1, 0.1},
	}

	for _, tc := range testCases {
		output := tc.function.Activate(tc.input)
		if output < tc.minOut || output > tc.maxOut {
			t.Errorf("%s activation of %f: expected range [%f, %f], got %f",
				tc.name, tc.input, tc.minOut, tc.maxOut, output)
		}

		// Test derivative
		derivative := tc.function.Derivative(tc.input)
		if math.IsNaN(derivative) || math.IsInf(derivative, 0) {
			t.Errorf("%s derivative of %f is invalid: %f", tc.name, tc.input, derivative)
		}
	}
}

func TestSoftmax(t *testing.T) {
	config := &NeuralNetworkConfig{
		InputSize:    3,
		HiddenLayers: []int{4},
		OutputSize:   3,
	}

	nn := NewNeuralNetwork(config)

	input := []float64{1.0, 2.0, 3.0}
	output := nn.softmax(input)

	if len(output) != 3 {
		t.Errorf("Expected output size 3, got %d", len(output))
	}

	// Check that probabilities sum to 1
	sum := 0.0
	for _, prob := range output {
		sum += prob
		if prob < 0 || prob > 1 {
			t.Errorf("Probability out of range [0,1]: %f", prob)
		}
	}

	if math.Abs(sum-1.0) > 1e-10 {
		t.Errorf("Probabilities don't sum to 1: %f", sum)
	}

	// Highest input should have highest probability
	maxIdx := 0
	maxVal := output[0]
	for i, val := range output {
		if val > maxVal {
			maxVal = val
			maxIdx = i
		}
	}

	if maxIdx != 2 { // Index 2 has highest input (3.0)
		t.Errorf("Expected highest probability at index 2, got %d", maxIdx)
	}
}

func TestNeuralNetworkClone(t *testing.T) {
	config := &NeuralNetworkConfig{
		InputSize:    3,
		HiddenLayers: []int{4},
		OutputSize:   2,
		LearningRate: 0.1,
		Epochs:       5,
	}

	original := NewNeuralNetwork(config)

	// Train original
	features := [][]float64{{1, 0, 1}}
	labels := [][]float64{{1, 0}}
	original.Train(features, labels)

	// Clone the network
	clone := original.Clone()

	if clone == nil {
		t.Fatal("Clone is nil")
	}

	if len(clone.layers) != len(original.layers) {
		t.Errorf("Clone has different number of layers: %d vs %d",
			len(clone.layers), len(original.layers))
	}

	if clone.trained != original.trained {
		t.Errorf("Clone training status mismatch: %v vs %v",
			clone.trained, original.trained)
	}

	// Test that both networks produce same output
	input := []float64{1, 0, 1}
	
	origOutput, err1 := original.Predict(input)
	cloneOutput, err2 := clone.Predict(input)

	if err1 != nil || err2 != nil {
		t.Fatalf("Prediction errors: orig=%v, clone=%v", err1, err2)
	}

	for i := range origOutput {
		if math.Abs(origOutput[i]-cloneOutput[i]) > 1e-10 {
			t.Errorf("Output mismatch at index %d: %f vs %f",
				i, origOutput[i], cloneOutput[i])
		}
	}
}

func TestNeuralNetworkTrainMismatchedData(t *testing.T) {
	config := &NeuralNetworkConfig{
		InputSize:    3,
		HiddenLayers: []int{4},
		OutputSize:   2,
	}

	nn := NewNeuralNetwork(config)

	features := [][]float64{{1, 0, 1}, {0, 1, 0}}
	labels := [][]float64{{1, 0}} // Mismatched size

	err := nn.Train(features, labels)
	if err == nil {
		t.Error("Expected error for mismatched features and labels")
	}
}

func TestNeuralNetworkEmptyTrainingData(t *testing.T) {
	config := &NeuralNetworkConfig{
		InputSize:    3,
		HiddenLayers: []int{4},
		OutputSize:   2,
	}

	nn := NewNeuralNetwork(config)

	features := [][]float64{}
	labels := [][]float64{}

	err := nn.Train(features, labels)
	if err == nil {
		t.Error("Expected error for empty training data")
	}
}

// Benchmark tests
func BenchmarkNeuralNetworkForward(b *testing.B) {
	config := &NeuralNetworkConfig{
		InputSize:    20,
		HiddenLayers: []int{32, 16},
		OutputSize:   6,
		ActivationFunction: "relu",
	}

	nn := NewNeuralNetwork(config)
	input := make([]float64, 20)
	for i := range input {
		input[i] = float64(i) * 0.1
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := nn.Forward(input)
		if err != nil {
			b.Fatalf("Forward pass failed: %v", err)
		}
	}
}

func BenchmarkNeuralNetworkTrain(b *testing.B) {
	config := &NeuralNetworkConfig{
		InputSize:    10,
		HiddenLayers: []int{16, 8},
		OutputSize:   4,
		LearningRate: 0.01,
		Epochs:       1, // Single epoch for benchmark
		BatchSize:    32,
		ActivationFunction: "relu",
	}

	nn := NewNeuralNetwork(config)

	// Generate random training data
	features := make([][]float64, 100)
	labels := make([][]float64, 100)
	
	for i := 0; i < 100; i++ {
		features[i] = make([]float64, 10)
		labels[i] = make([]float64, 4)
		
		for j := 0; j < 10; j++ {
			features[i][j] = float64(i*j) * 0.01
		}
		
		// One-hot label
		labels[i][i%4] = 1.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := nn.Train(features, labels)
		if err != nil {
			b.Fatalf("Training failed: %v", err)
		}
		
		// Reset network for next iteration
		nn = NewNeuralNetwork(config)
	}
}
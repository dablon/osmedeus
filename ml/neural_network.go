package ml

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/dablon/osmedeus/utils"
)

// NeuralNetwork implements a simple feedforward neural network for asset classification
type NeuralNetwork struct {
	layers      []Layer
	learningRate float64
	momentum     float64
	config       *NeuralNetworkConfig
	trained      bool
}

// NeuralNetworkConfig holds configuration for the neural network
type NeuralNetworkConfig struct {
	InputSize     int     `json:"input_size"`
	HiddenLayers  []int   `json:"hidden_layers"`
	OutputSize    int     `json:"output_size"`
	LearningRate  float64 `json:"learning_rate"`
	Momentum      float64 `json:"momentum"`
	Epochs        int     `json:"epochs"`
	BatchSize     int     `json:"batch_size"`
	DropoutRate   float64 `json:"dropout_rate"`
	Regularization float64 `json:"regularization"`
	ActivationFunction string `json:"activation_function"`
}

// Layer represents a neural network layer
type Layer struct {
	Weights     [][]float64 `json:"weights"`
	Biases      []float64   `json:"biases"`
	Activations []float64   `json:"activations"`
	Deltas      []float64   `json:"deltas"`
	PrevWeights [][]float64 `json:"prev_weights"`
	PrevBiases  []float64   `json:"prev_biases"`
	Size        int         `json:"size"`
	Dropout     []bool      `json:"dropout"`
}

// ActivationFunction represents different activation functions
type ActivationFunction interface {
	Activate(x float64) float64
	Derivative(x float64) float64
}

// Sigmoid activation function
type Sigmoid struct{}

func (s *Sigmoid) Activate(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func (s *Sigmoid) Derivative(x float64) float64 {
	sigmoid := s.Activate(x)
	return sigmoid * (1.0 - sigmoid)
}

// ReLU activation function
type ReLU struct{}

func (r *ReLU) Activate(x float64) float64 {
	return math.Max(0, x)
}

func (r *ReLU) Derivative(x float64) float64 {
	if x > 0 {
		return 1.0
	}
	return 0.0
}

// Tanh activation function
type Tanh struct{}

func (t *Tanh) Activate(x float64) float64 {
	return math.Tanh(x)
}

func (t *Tanh) Derivative(x float64) float64 {
	tanh := t.Activate(x)
	return 1.0 - tanh*tanh
}

// NewNeuralNetwork creates a new neural network
func NewNeuralNetwork(config *NeuralNetworkConfig) *NeuralNetwork {
	if config == nil {
		config = &NeuralNetworkConfig{
			InputSize:    18, // Based on asset feature extraction
			HiddenLayers: []int{32, 16, 8},
			OutputSize:   6, // Number of asset types (production, staging, dev, admin, api, test)
			LearningRate: 0.001,
			Momentum:     0.9,
			Epochs:       1000,
			BatchSize:    32,
			DropoutRate:  0.2,
			Regularization: 0.001,
			ActivationFunction: "relu",
		}
	}

	nn := &NeuralNetwork{
		learningRate: config.LearningRate,
		momentum:     config.Momentum,
		config:       config,
		trained:      false,
	}

	// Initialize layers
	nn.initializeLayers()

	return nn
}

// initializeLayers initializes the neural network layers with random weights
func (nn *NeuralNetwork) initializeLayers() {
	rand.Seed(time.Now().UnixNano())

	// Calculate layer sizes
	layerSizes := []int{nn.config.InputSize}
	layerSizes = append(layerSizes, nn.config.HiddenLayers...)
	layerSizes = append(layerSizes, nn.config.OutputSize)

	nn.layers = make([]Layer, len(layerSizes)-1)

	// Initialize each layer
	for i := 0; i < len(nn.layers); i++ {
		inputSize := layerSizes[i]
		outputSize := layerSizes[i+1]

		layer := Layer{
			Size:        outputSize,
			Weights:     make([][]float64, inputSize),
			Biases:      make([]float64, outputSize),
			Activations: make([]float64, outputSize),
			Deltas:      make([]float64, outputSize),
			PrevWeights: make([][]float64, inputSize),
			PrevBiases:  make([]float64, outputSize),
			Dropout:     make([]bool, outputSize),
		}

		// Initialize weights using Xavier initialization
		for j := 0; j < inputSize; j++ {
			layer.Weights[j] = make([]float64, outputSize)
			layer.PrevWeights[j] = make([]float64, outputSize)
			
			for k := 0; k < outputSize; k++ {
				// Xavier initialization
				limit := math.Sqrt(6.0 / float64(inputSize+outputSize))
				layer.Weights[j][k] = (rand.Float64()*2.0 - 1.0) * limit
			}
		}

		// Initialize biases to small random values
		for j := 0; j < outputSize; j++ {
			layer.Biases[j] = (rand.Float64()*2.0 - 1.0) * 0.1
		}

		nn.layers[i] = layer
	}

	utils.InforF("Neural network initialized with %d layers", len(nn.layers))
}

// Forward performs forward propagation through the network
func (nn *NeuralNetwork) Forward(input []float64) ([]float64, error) {
	if len(input) != nn.config.InputSize {
		return nil, fmt.Errorf("input size mismatch: expected %d, got %d", nn.config.InputSize, len(input))
	}

	currentInput := input

	// Forward propagation through each layer
	for layerIdx := range nn.layers {
		layer := &nn.layers[layerIdx]
		
		// Calculate weighted sum + bias for each neuron
		for j := 0; j < layer.Size; j++ {
			sum := layer.Biases[j]
			
			for i := 0; i < len(currentInput); i++ {
				sum += currentInput[i] * layer.Weights[i][j]
			}

			// Apply activation function
			layer.Activations[j] = nn.activate(sum, layerIdx)

			// Apply dropout during training (not during inference)
			if nn.config.DropoutRate > 0 && layerIdx < len(nn.layers)-1 {
				if rand.Float64() < nn.config.DropoutRate {
					layer.Dropout[j] = true
					layer.Activations[j] = 0
				} else {
					layer.Dropout[j] = false
					// Scale up remaining neurons to compensate for dropout
					layer.Activations[j] /= (1.0 - nn.config.DropoutRate)
				}
			}
		}

		// Set current layer output as input for next layer
		currentInput = make([]float64, layer.Size)
		copy(currentInput, layer.Activations)
	}

	return currentInput, nil
}

// Backward performs backpropagation to update weights
func (nn *NeuralNetwork) Backward(input []float64, target []float64) error {
	if len(target) != nn.config.OutputSize {
		return fmt.Errorf("target size mismatch: expected %d, got %d", nn.config.OutputSize, len(target))
	}

	// Calculate output layer deltas
	outputLayer := &nn.layers[len(nn.layers)-1]
	for i := 0; i < outputLayer.Size; i++ {
		error := target[i] - outputLayer.Activations[i]
		outputLayer.Deltas[i] = error * nn.activateDerivative(outputLayer.Activations[i], len(nn.layers)-1)
	}

	// Backpropagate deltas through hidden layers
	for layerIdx := len(nn.layers) - 2; layerIdx >= 0; layerIdx-- {
		layer := &nn.layers[layerIdx]
		nextLayer := &nn.layers[layerIdx+1]

		for i := 0; i < layer.Size; i++ {
			error := 0.0
			// For each neuron in the next layer
			for j := 0; j < nextLayer.Size; j++ {
				// The weight from current layer neuron i to next layer neuron j
				// is stored in the current layer's weights as layer.Weights[i][j]
				error += nextLayer.Deltas[j] * layer.Weights[i][j]
			}
			layer.Deltas[i] = error * nn.activateDerivative(layer.Activations[i], layerIdx)
		}
	}

	// Update weights and biases
	nn.updateWeights(input)

	return nil
}

// updateWeights updates the network weights using gradient descent with momentum
func (nn *NeuralNetwork) updateWeights(input []float64) {
	currentInput := input

	for layerIdx := range nn.layers {
		layer := &nn.layers[layerIdx]

		// Update weights
		for i := 0; i < len(currentInput); i++ {
			for j := 0; j < layer.Size; j++ {
				// Calculate gradient
				gradient := nn.learningRate * layer.Deltas[j] * currentInput[i]
				
				// Add L2 regularization
				if nn.config.Regularization > 0 {
					gradient += nn.config.Regularization * layer.Weights[i][j]
				}

				// Apply momentum
				weightUpdate := gradient + nn.momentum*layer.PrevWeights[i][j]
				
				// Update weight
				layer.Weights[i][j] += weightUpdate
				layer.PrevWeights[i][j] = weightUpdate
			}
		}

		// Update biases
		for j := 0; j < layer.Size; j++ {
			biasUpdate := nn.learningRate*layer.Deltas[j] + nn.momentum*layer.PrevBiases[j]
			layer.Biases[j] += biasUpdate
			layer.PrevBiases[j] = biasUpdate
		}

		// Set current layer output as input for next layer weight updates
		currentInput = make([]float64, layer.Size)
		copy(currentInput, layer.Activations)
	}
}

// Train trains the neural network on the provided dataset
func (nn *NeuralNetwork) Train(features [][]float64, labels [][]float64) error {
	if len(features) != len(labels) {
		return fmt.Errorf("features and labels count mismatch: %d vs %d", len(features), len(labels))
	}

	if len(features) == 0 {
		return fmt.Errorf("no training data provided")
	}

	utils.InforF("Training neural network with %d samples for %d epochs", len(features), nn.config.Epochs)

	// Training loop
	for epoch := 0; epoch < nn.config.Epochs; epoch++ {
		totalLoss := 0.0
		
		// Shuffle training data
		indices := nn.shuffleIndices(len(features))

		// Process in batches
		for batchStart := 0; batchStart < len(features); batchStart += nn.config.BatchSize {
			batchEnd := batchStart + nn.config.BatchSize
			if batchEnd > len(features) {
				batchEnd = len(features)
			}

			batchLoss := 0.0

			// Process batch
			for i := batchStart; i < batchEnd; i++ {
				idx := indices[i]
				
				// Forward pass
				output, err := nn.Forward(features[idx])
				if err != nil {
					return fmt.Errorf("forward pass failed: %w", err)
				}

				// Calculate loss
				loss := nn.calculateLoss(output, labels[idx])
				batchLoss += loss

				// Backward pass
				if err := nn.Backward(features[idx], labels[idx]); err != nil {
					return fmt.Errorf("backward pass failed: %w", err)
				}
			}

			totalLoss += batchLoss
		}

		// Log progress
		if epoch%100 == 0 || epoch == nn.config.Epochs-1 {
			avgLoss := totalLoss / float64(len(features))
			utils.InforF("Epoch %d/%d, Average Loss: %.6f", epoch+1, nn.config.Epochs, avgLoss)
		}
	}

	nn.trained = true
	utils.InforF("Neural network training completed")
	return nil
}

// Predict makes predictions on input data
func (nn *NeuralNetwork) Predict(input []float64) ([]float64, error) {
	if !nn.trained {
		return nil, fmt.Errorf("neural network not trained")
	}

	// Disable dropout during inference
	originalDropoutRate := nn.config.DropoutRate
	nn.config.DropoutRate = 0

	output, err := nn.Forward(input)

	// Restore dropout rate
	nn.config.DropoutRate = originalDropoutRate

	if err != nil {
		return nil, fmt.Errorf("prediction failed: %w", err)
	}

	// Apply softmax to output for classification probabilities
	return nn.softmax(output), nil
}

// activate applies the activation function
func (nn *NeuralNetwork) activate(x float64, layerIdx int) float64 {
	switch nn.config.ActivationFunction {
	case "sigmoid":
		return (&Sigmoid{}).Activate(x)
	case "tanh":
		return (&Tanh{}).Activate(x)
	case "relu":
		return (&ReLU{}).Activate(x)
	default:
		return (&ReLU{}).Activate(x) // Default to ReLU
	}
}

// activateDerivative applies the derivative of the activation function
func (nn *NeuralNetwork) activateDerivative(x float64, layerIdx int) float64 {
	switch nn.config.ActivationFunction {
	case "sigmoid":
		return (&Sigmoid{}).Derivative(x)
	case "tanh":
		return (&Tanh{}).Derivative(x)
	case "relu":
		return (&ReLU{}).Derivative(x)
	default:
		return (&ReLU{}).Derivative(x) // Default to ReLU
	}
}

// softmax applies softmax activation to convert outputs to probabilities
func (nn *NeuralNetwork) softmax(input []float64) []float64 {
	// Find max value for numerical stability
	maxVal := input[0]
	for _, val := range input {
		if val > maxVal {
			maxVal = val
		}
	}

	// Calculate exponentials
	expSum := 0.0
	output := make([]float64, len(input))
	for i, val := range input {
		output[i] = math.Exp(val - maxVal)
		expSum += output[i]
	}

	// Normalize
	for i := range output {
		output[i] /= expSum
	}

	return output
}

// calculateLoss calculates the cross-entropy loss
func (nn *NeuralNetwork) calculateLoss(predicted, target []float64) float64 {
	loss := 0.0
	for i := range predicted {
		if target[i] > 0 {
			// Avoid log(0) by adding small epsilon
			epsilon := 1e-15
			loss -= target[i] * math.Log(math.Max(predicted[i], epsilon))
		}
	}
	return loss
}

// shuffleIndices creates a shuffled array of indices
func (nn *NeuralNetwork) shuffleIndices(n int) []int {
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}

	// Fisher-Yates shuffle
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		indices[i], indices[j] = indices[j], indices[i]
	}

	return indices
}

// Evaluate evaluates the network performance on test data
func (nn *NeuralNetwork) Evaluate(features [][]float64, labels [][]float64) (float64, error) {
	if len(features) != len(labels) {
		return 0, fmt.Errorf("features and labels count mismatch")
	}

	correct := 0
	total := len(features)

	for i := 0; i < total; i++ {
		prediction, err := nn.Predict(features[i])
		if err != nil {
			return 0, fmt.Errorf("prediction failed: %w", err)
		}

		// Find predicted class (highest probability)
		predClass := nn.argmax(prediction)
		trueClass := nn.argmax(labels[i])

		if predClass == trueClass {
			correct++
		}
	}

	accuracy := float64(correct) / float64(total)
	return accuracy, nil
}

// argmax returns the index of the maximum value
func (nn *NeuralNetwork) argmax(values []float64) int {
	maxIdx := 0
	maxVal := values[0]
	
	for i, val := range values {
		if val > maxVal {
			maxVal = val
			maxIdx = i
		}
	}
	
	return maxIdx
}

// GetWeights returns the network weights for serialization
func (nn *NeuralNetwork) GetWeights() []Layer {
	return nn.layers
}

// SetWeights sets the network weights from serialized data
func (nn *NeuralNetwork) SetWeights(layers []Layer) error {
	if len(layers) != len(nn.layers) {
		return fmt.Errorf("layer count mismatch: expected %d, got %d", len(nn.layers), len(layers))
	}

	nn.layers = layers
	nn.trained = true
	return nil
}

// GetConfig returns the network configuration
func (nn *NeuralNetwork) GetConfig() *NeuralNetworkConfig {
	return nn.config
}

// Clone creates a copy of the neural network
func (nn *NeuralNetwork) Clone() *NeuralNetwork {
	clone := NewNeuralNetwork(nn.config)
	
	// Copy weights and biases
	for i, layer := range nn.layers {
		cloneLayer := &clone.layers[i]
		
		// Copy weights
		for j := range layer.Weights {
			copy(cloneLayer.Weights[j], layer.Weights[j])
			copy(cloneLayer.PrevWeights[j], layer.PrevWeights[j])
		}
		
		// Copy biases
		copy(cloneLayer.Biases, layer.Biases)
		copy(cloneLayer.PrevBiases, layer.PrevBiases)
	}
	
	clone.trained = nn.trained
	return clone
}

// Summary prints a summary of the network architecture
func (nn *NeuralNetwork) Summary() {
	utils.InforF("Neural Network Architecture:")
	utils.InforF("Input Size: %d", nn.config.InputSize)
	
	for i, layer := range nn.layers {
		if i < len(nn.layers)-1 {
			utils.InforF("Hidden Layer %d: %d neurons", i+1, layer.Size)
		} else {
			utils.InforF("Output Layer: %d neurons", layer.Size)
		}
	}
	
	utils.InforF("Activation Function: %s", nn.config.ActivationFunction)
	utils.InforF("Learning Rate: %.4f", nn.config.LearningRate)
	utils.InforF("Momentum: %.2f", nn.config.Momentum)
	utils.InforF("Dropout Rate: %.2f", nn.config.DropoutRate)
	utils.InforF("Regularization: %.4f", nn.config.Regularization)
	
	// Calculate total parameters
	totalParams := 0
	for _, layer := range nn.layers {
		for _, weights := range layer.Weights {
			totalParams += len(weights)
		}
		totalParams += len(layer.Biases)
	}
	utils.InforF("Total Parameters: %d", totalParams)
}
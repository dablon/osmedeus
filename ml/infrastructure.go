package ml

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dablon/osmedeus/database"
	"github.com/dablon/osmedeus/utils"
)

// MLInfrastructure manages the machine learning infrastructure
type MLInfrastructure struct {
	models          map[string]*ModelWrapper
	preprocessors   map[string]DataPreprocessor
	inferenceQueue  chan InferenceRequest
	trainingQueue   chan TrainingRequest
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	mu              sync.RWMutex
	config          *MLConfig
}

// MLConfig holds configuration for ML infrastructure
type MLConfig struct {
	ModelStorePath      string            `json:"model_store_path"`
	MaxConcurrentInfer  int               `json:"max_concurrent_infer"`
	MaxConcurrentTrain  int               `json:"max_concurrent_train"`
	DefaultTimeout      time.Duration     `json:"default_timeout"`
	GPUEnabled          bool              `json:"gpu_enabled"`
	ModelConfigs        map[string]interface{} `json:"model_configs"`
	PreprocessingConfig map[string]interface{} `json:"preprocessing_config"`
}

// ModelWrapper wraps ML models with metadata and management functions
type ModelWrapper struct {
	Model       MLModel
	Framework   string
	Status      ModelStatus
	LastUsed    time.Time
	LoadedAt    time.Time
	Metadata    map[string]interface{}
	mu          sync.RWMutex
}

// ModelStatus represents the current status of a model
type ModelStatus string

const (
	ModelStatusLoading   ModelStatus = "loading"
	ModelStatusReady     ModelStatus = "ready"
	ModelStatusError     ModelStatus = "error"
	ModelStatusUnloaded  ModelStatus = "unloaded"
	ModelStatusTraining  ModelStatus = "training"
)

// MLModel interface defines the contract for all ML models
type MLModel interface {
	Load(modelPath string) error
	Predict(input interface{}) (interface{}, error)
	Train(data TrainingData) error
	Save(modelPath string) error
	GetMetadata() map[string]interface{}
	Validate(testData interface{}) (ValidationResult, error)
}

// DataPreprocessor interface for data preprocessing pipelines
type DataPreprocessor interface {
	Preprocess(rawData interface{}) (interface{}, error)
	GetFeatures(data interface{}) ([]float64, error)
	Normalize(features []float64) ([]float64, error)
	Transform(data interface{}) (interface{}, error)
}

// InferenceRequest represents a request for model inference
type InferenceRequest struct {
	ID          string
	ModelID     string
	InputData   interface{}
	TargetID    string
	ResponseCh  chan InferenceResponse
	Timeout     time.Duration
	RequestedAt time.Time
}

// InferenceResponse represents the response from model inference
type InferenceResponse struct {
	ID         string
	Result     interface{}
	Confidence float64
	Error      error
	ProcessingTime time.Duration
	ModelVersion   string
	Metadata       map[string]interface{}
}

// TrainingRequest represents a request for model training
type TrainingRequest struct {
	ID          string
	ModelID     string
	TrainingData TrainingData
	ResponseCh  chan TrainingResponse
	RequestedAt time.Time
}

// TrainingResponse represents the response from model training
type TrainingResponse struct {
	ID           string
	Success      bool
	Error        error
	ModelMetrics map[string]float64
	TrainingTime time.Duration
}

// TrainingData represents data used for model training
type TrainingData struct {
	Features [][]float64
	Labels   []interface{}
	Metadata map[string]interface{}
}

// ValidationResult represents model validation results
type ValidationResult struct {
	Accuracy  float64
	Precision float64
	Recall    float64
	F1Score   float64
	Metadata  map[string]interface{}
}

// NewMLInfrastructure creates a new ML infrastructure instance
func NewMLInfrastructure(config *MLConfig) *MLInfrastructure {
	ctx, cancel := context.WithCancel(context.Background())
	
	if config == nil {
		config = &MLConfig{
			ModelStorePath:     "./models",
			MaxConcurrentInfer: 10,
			MaxConcurrentTrain: 2,
			DefaultTimeout:     30 * time.Second,
			GPUEnabled:         false,
			ModelConfigs:       make(map[string]interface{}),
			PreprocessingConfig: make(map[string]interface{}),
		}
	}

	ml := &MLInfrastructure{
		models:         make(map[string]*ModelWrapper),
		preprocessors:  make(map[string]DataPreprocessor),
		inferenceQueue: make(chan InferenceRequest, config.MaxConcurrentInfer*2),
		trainingQueue:  make(chan TrainingRequest, config.MaxConcurrentTrain*2),
		ctx:            ctx,
		cancel:         cancel,
		config:         config,
	}

	// Start worker goroutines
	ml.startWorkers()

	return ml
}

// startWorkers starts the inference and training worker goroutines
func (ml *MLInfrastructure) startWorkers() {
	// Start inference workers
	for i := 0; i < ml.config.MaxConcurrentInfer; i++ {
		ml.wg.Add(1)
		go ml.inferenceWorker()
	}

	// Start training workers
	for i := 0; i < ml.config.MaxConcurrentTrain; i++ {
		ml.wg.Add(1)
		go ml.trainingWorker()
	}

	utils.InforF("Started ML infrastructure with %d inference workers and %d training workers",
		ml.config.MaxConcurrentInfer, ml.config.MaxConcurrentTrain)
}

// LoadModel loads a model into the infrastructure
func (ml *MLInfrastructure) LoadModel(modelID, modelPath, framework string) error {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	wrapper := &ModelWrapper{
		Framework: framework,
		Status:    ModelStatusLoading,
		LoadedAt:  time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Create model based on framework
	var model MLModel
	var err error

	switch framework {
	case "tensorflow":
		model, err = NewTensorFlowModel()
	case "pytorch":
		model, err = NewPyTorchModel()
	case "sklearn":
		model, err = NewSklearnModel()
	default:
		return fmt.Errorf("unsupported ML framework: %s", framework)
	}

	if err != nil {
		wrapper.Status = ModelStatusError
		ml.models[modelID] = wrapper
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Load the model
	if err := model.Load(modelPath); err != nil {
		wrapper.Status = ModelStatusError
		ml.models[modelID] = wrapper
		return fmt.Errorf("failed to load model: %w", err)
	}

	wrapper.Model = model
	wrapper.Status = ModelStatusReady
	wrapper.Metadata = model.GetMetadata()
	ml.models[modelID] = wrapper

	utils.InforF("Successfully loaded ML model: %s (%s)", modelID, framework)
	return nil
}

// UnloadModel unloads a model from memory
func (ml *MLInfrastructure) UnloadModel(modelID string) error {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	wrapper, exists := ml.models[modelID]
	if !exists {
		return fmt.Errorf("model not found: %s", modelID)
	}

	wrapper.mu.Lock()
	wrapper.Status = ModelStatusUnloaded
	wrapper.Model = nil
	wrapper.mu.Unlock()

	delete(ml.models, modelID)
	utils.InforF("Unloaded ML model: %s", modelID)
	return nil
}

// Predict performs inference using a loaded model
func (ml *MLInfrastructure) Predict(modelID string, inputData interface{}, targetID string) (*InferenceResponse, error) {
	request := InferenceRequest{
		ID:          fmt.Sprintf("inf_%d", time.Now().UnixNano()),
		ModelID:     modelID,
		InputData:   inputData,
		TargetID:    targetID,
		ResponseCh:  make(chan InferenceResponse, 1),
		Timeout:     ml.config.DefaultTimeout,
		RequestedAt: time.Now(),
	}

	select {
	case ml.inferenceQueue <- request:
		// Request queued successfully
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("inference queue is full, request timed out")
	}

	select {
	case response := <-request.ResponseCh:
		return &response, response.Error
	case <-time.After(request.Timeout):
		return nil, fmt.Errorf("inference request timed out after %v", request.Timeout)
	}
}

// TrainModel trains a model with provided data
func (ml *MLInfrastructure) TrainModel(modelID string, trainingData TrainingData) (*TrainingResponse, error) {
	request := TrainingRequest{
		ID:           fmt.Sprintf("train_%d", time.Now().UnixNano()),
		ModelID:      modelID,
		TrainingData: trainingData,
		ResponseCh:   make(chan TrainingResponse, 1),
		RequestedAt:  time.Now(),
	}

	select {
	case ml.trainingQueue <- request:
		// Request queued successfully
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("training queue is full, request timed out")
	}

	select {
	case response := <-request.ResponseCh:
		return &response, response.Error
	case <-time.After(10 * time.Minute): // Training can take longer
		return nil, fmt.Errorf("training request timed out")
	}
}

// RegisterPreprocessor registers a data preprocessor
func (ml *MLInfrastructure) RegisterPreprocessor(name string, preprocessor DataPreprocessor) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.preprocessors[name] = preprocessor
	utils.InforF("Registered data preprocessor: %s", name)
}

// GetPreprocessor retrieves a registered preprocessor
func (ml *MLInfrastructure) GetPreprocessor(name string) (DataPreprocessor, error) {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	
	preprocessor, exists := ml.preprocessors[name]
	if !exists {
		return nil, fmt.Errorf("preprocessor not found: %s", name)
	}
	return preprocessor, nil
}

// inferenceWorker processes inference requests
func (ml *MLInfrastructure) inferenceWorker() {
	defer ml.wg.Done()

	for {
		select {
		case <-ml.ctx.Done():
			return
		case request := <-ml.inferenceQueue:
			ml.processInferenceRequest(request)
		}
	}
}

// trainingWorker processes training requests
func (ml *MLInfrastructure) trainingWorker() {
	defer ml.wg.Done()

	for {
		select {
		case <-ml.ctx.Done():
			return
		case request := <-ml.trainingQueue:
			ml.processTrainingRequest(request)
		}
	}
}

// processInferenceRequest processes a single inference request
func (ml *MLInfrastructure) processInferenceRequest(request InferenceRequest) {
	startTime := time.Now()
	
	response := InferenceResponse{
		ID: request.ID,
	}

	ml.mu.RLock()
	wrapper, exists := ml.models[request.ModelID]
	ml.mu.RUnlock()

	if !exists {
		response.Error = fmt.Errorf("model not found: %s", request.ModelID)
		request.ResponseCh <- response
		return
	}

	wrapper.mu.RLock()
	if wrapper.Status != ModelStatusReady {
		wrapper.mu.RUnlock()
		response.Error = fmt.Errorf("model not ready: %s (status: %s)", request.ModelID, wrapper.Status)
		request.ResponseCh <- response
		return
	}

	model := wrapper.Model
	wrapper.mu.RUnlock()

	// Perform inference
	result, err := model.Predict(request.InputData)
	if err != nil {
		response.Error = fmt.Errorf("inference failed: %w", err)
		request.ResponseCh <- response
		return
	}

	response.Result = result
	response.ProcessingTime = time.Since(startTime)
	response.ModelVersion = wrapper.Metadata["version"].(string)
	response.Metadata = wrapper.Metadata

	// Update model usage
	wrapper.mu.Lock()
	wrapper.LastUsed = time.Now()
	wrapper.mu.Unlock()

	// Log inference
	ml.logInference(request, response)

	request.ResponseCh <- response
}

// processTrainingRequest processes a single training request
func (ml *MLInfrastructure) processTrainingRequest(request TrainingRequest) {
	startTime := time.Now()
	
	response := TrainingResponse{
		ID: request.ID,
	}

	ml.mu.RLock()
	wrapper, exists := ml.models[request.ModelID]
	ml.mu.RUnlock()

	if !exists {
		response.Error = fmt.Errorf("model not found: %s", request.ModelID)
		request.ResponseCh <- response
		return
	}

	wrapper.mu.Lock()
	wrapper.Status = ModelStatusTraining
	model := wrapper.Model
	wrapper.mu.Unlock()

	// Perform training
	err := model.Train(request.TrainingData)
	
	wrapper.mu.Lock()
	if err != nil {
		wrapper.Status = ModelStatusError
		response.Error = fmt.Errorf("training failed: %w", err)
	} else {
		wrapper.Status = ModelStatusReady
		response.Success = true
		wrapper.Metadata = model.GetMetadata()
	}
	wrapper.mu.Unlock()

	response.TrainingTime = time.Since(startTime)
	request.ResponseCh <- response
}

// logInference logs inference requests for monitoring and debugging
func (ml *MLInfrastructure) logInference(request InferenceRequest, response InferenceResponse) {
	logEntry := database.MLInferenceLog{
		ModelID:        request.ModelID,
		TargetID:       request.TargetID,
		ProcessingTime: int(response.ProcessingTime.Milliseconds()),
		InferredAt:     time.Now(),
	}

	if response.Error != nil {
		logEntry.Status = "error"
		logEntry.ErrorMessage = response.Error.Error()
	} else {
		logEntry.Status = "success"
		logEntry.Confidence = response.Confidence
	}

	// Convert input/output to JSON
	if inputJSON, err := json.Marshal(request.InputData); err == nil {
		logEntry.InputData = string(inputJSON)
	}
	if outputJSON, err := json.Marshal(response.Result); err == nil {
		logEntry.OutputData = string(outputJSON)
	}

	// In a real implementation, this would save to database
	utils.InforF("ML Inference: Model=%s, Target=%s, Status=%s, Time=%dms",
		logEntry.ModelID, logEntry.TargetID, logEntry.Status, logEntry.ProcessingTime)
}

// GetModelStatus returns the status of a model
func (ml *MLInfrastructure) GetModelStatus(modelID string) (ModelStatus, error) {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	wrapper, exists := ml.models[modelID]
	if !exists {
		return "", fmt.Errorf("model not found: %s", modelID)
	}

	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()
	return wrapper.Status, nil
}

// ListModels returns a list of all loaded models
func (ml *MLInfrastructure) ListModels() map[string]ModelStatus {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	models := make(map[string]ModelStatus)
	for id, wrapper := range ml.models {
		wrapper.mu.RLock()
		models[id] = wrapper.Status
		wrapper.mu.RUnlock()
	}
	return models
}

// Shutdown gracefully shuts down the ML infrastructure
func (ml *MLInfrastructure) Shutdown() error {
	utils.InforF("Shutting down ML infrastructure...")
	
	ml.cancel()
	ml.wg.Wait()
	
	// Unload all models
	ml.mu.Lock()
	for modelID := range ml.models {
		delete(ml.models, modelID)
	}
	ml.mu.Unlock()

	close(ml.inferenceQueue)
	close(ml.trainingQueue)
	
	utils.InforF("ML infrastructure shutdown complete")
	return nil
}
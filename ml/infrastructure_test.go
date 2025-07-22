package ml

import (
	"testing"
	"time"
)

func TestMLInfrastructure(t *testing.T) {
	// Test ML infrastructure initialization
	config := &MLConfig{
		ModelStorePath:     "./test_models",
		MaxConcurrentInfer: 2,
		MaxConcurrentTrain: 1,
		DefaultTimeout:     5 * time.Second,
		GPUEnabled:         false,
		ModelConfigs:       make(map[string]interface{}),
		PreprocessingConfig: make(map[string]interface{}),
	}

	ml := NewMLInfrastructure(config)
	if ml == nil {
		t.Fatal("Failed to create ML infrastructure")
	}

	// Test model creation
	model, err := NewTensorFlowModel()
	if err != nil {
		t.Fatalf("Failed to create TensorFlow model: %v", err)
	}

	if model == nil {
		t.Fatal("Model is nil")
	}

	// Test model metadata
	metadata := model.GetMetadata()
	if metadata["framework"] != "tensorflow" {
		t.Errorf("Expected framework 'tensorflow', got %v", metadata["framework"])
	}

	// Test preprocessing pipeline
	preprocessor := NewAssetDataPreprocessor()
	if preprocessor == nil {
		t.Fatal("Failed to create asset data preprocessor")
	}

	// Test data preprocessing
	testData := map[string]interface{}{
		"open_ports": []interface{}{80.0, 443.0, 22.0},
		"services":   []interface{}{"http", "https", "ssh"},
		"technologies": []interface{}{"nginx", "ssl"},
	}

	processed, err := preprocessor.Preprocess(testData)
	if err != nil {
		t.Fatalf("Failed to preprocess data: %v", err)
	}

	if processed == nil {
		t.Fatal("Processed data is nil")
	}

	// Test feature extraction
	features, err := preprocessor.GetFeatures(processed)
	if err != nil {
		t.Fatalf("Failed to extract features: %v", err)
	}

	if len(features) == 0 {
		t.Fatal("No features extracted")
	}

	// Test feature normalization
	normalized, err := preprocessor.Normalize(features)
	if err != nil {
		t.Fatalf("Failed to normalize features: %v", err)
	}

	if len(normalized) != len(features) {
		t.Errorf("Expected %d normalized features, got %d", len(features), len(normalized))
	}

	// Cleanup
	ml.Shutdown()
}

func TestDataPreprocessors(t *testing.T) {
	// Test Asset Data Preprocessor
	assetProcessor := NewAssetDataPreprocessor()
	testAssetData := map[string]interface{}{
		"open_ports":   []interface{}{80.0, 443.0, 22.0, 3306.0},
		"services":     []interface{}{"http", "https", "ssh", "mysql"},
		"technologies": []interface{}{"nginx", "ssl", "mysql"},
	}

	processed, err := assetProcessor.Preprocess(testAssetData)
	if err != nil {
		t.Fatalf("Asset preprocessing failed: %v", err)
	}

	features, err := assetProcessor.GetFeatures(processed)
	if err != nil {
		t.Fatalf("Asset feature extraction failed: %v", err)
	}

	if len(features) == 0 {
		t.Fatal("No asset features extracted")
	}

	// Test Behavior Data Preprocessor
	behaviorProcessor := NewBehaviorDataPreprocessor()
	testBehaviorData := map[string]interface{}{
		"response_times": []interface{}{100.0, 150.0, 120.0, 200.0},
		"status_codes":   []interface{}{200.0, 404.0, 500.0},
	}

	processed, err = behaviorProcessor.Preprocess(testBehaviorData)
	if err != nil {
		t.Fatalf("Behavior preprocessing failed: %v", err)
	}

	features, err = behaviorProcessor.GetFeatures(processed)
	if err != nil {
		t.Fatalf("Behavior feature extraction failed: %v", err)
	}

	if len(features) == 0 {
		t.Fatal("No behavior features extracted")
	}

	// Test Vulnerability Data Preprocessor
	vulnProcessor := NewVulnerabilityDataPreprocessor()
	testVulnData := map[string]interface{}{
		"vulnerability_count": 5.0,
		"avg_severity":        7.5,
	}

	processed, err = vulnProcessor.Preprocess(testVulnData)
	if err != nil {
		t.Fatalf("Vulnerability preprocessing failed: %v", err)
	}

	features, err = vulnProcessor.GetFeatures(processed)
	if err != nil {
		t.Fatalf("Vulnerability feature extraction failed: %v", err)
	}

	if len(features) == 0 {
		t.Fatal("No vulnerability features extracted")
	}
}

func TestPreprocessingPipeline(t *testing.T) {
	config := &PreprocessingConfig{
		FeatureScaling:   "standard",
		HandleMissing:    "mean",
		TextProcessing:   "tfidf",
		TimeSeriesWindow: 5,
		OutlierDetection: true,
		FeatureSelection: true,
		CustomProcessors: make(map[string]interface{}),
	}

	pipeline := NewPreprocessingPipeline(config)
	if pipeline == nil {
		t.Fatal("Failed to create preprocessing pipeline")
	}

	// Test feature scaling
	testFeatures := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	
	// Test standard scaling
	scaled, err := pipeline.scaleFeatures(testFeatures)
	if err != nil {
		t.Fatalf("Feature scaling failed: %v", err)
	}

	if len(scaled) != len(testFeatures) {
		t.Errorf("Expected %d scaled features, got %d", len(testFeatures), len(scaled))
	}

	// Test outlier removal
	testFeaturesWithOutliers := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 100.0} // 100.0 is an outlier
	filtered := pipeline.removeOutliers(testFeaturesWithOutliers)
	
	if len(filtered) >= len(testFeaturesWithOutliers) {
		t.Error("Outlier detection should have removed some values")
	}

	// Test time series windowing
	timeSeries := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0}
	windowed := pipeline.applyTimeSeriesWindow(timeSeries, 3)
	
	expectedWindows := len(timeSeries) - 3 + 1
	if len(windowed) != expectedWindows {
		t.Errorf("Expected %d windows, got %d", expectedWindows, len(windowed))
	}
}

func TestModelValidation(t *testing.T) {
	// Test TensorFlow model validation
	tfModel, err := NewTensorFlowModel()
	if err != nil {
		t.Fatalf("Failed to create TensorFlow model: %v", err)
	}

	// Simulate model loading (in real implementation, this would load actual model files)
	err = tfModel.Load("./test_model.tf")
	if err == nil { // We expect this to fail since the file doesn't exist
		// If it doesn't fail, test the validation
		result, err := tfModel.Validate(map[string]interface{}{"test": "data"})
		if err != nil {
			t.Fatalf("Model validation failed: %v", err)
		}

		if result.Accuracy < 0 || result.Accuracy > 1 {
			t.Errorf("Invalid accuracy value: %f", result.Accuracy)
		}
	}

	// Test PyTorch model validation
	ptModel, err := NewPyTorchModel()
	if err != nil {
		t.Fatalf("Failed to create PyTorch model: %v", err)
	}

	metadata := ptModel.GetMetadata()
	if metadata["framework"] != "pytorch" {
		t.Errorf("Expected framework 'pytorch', got %v", metadata["framework"])
	}

	// Test Sklearn model validation
	skModel, err := NewSklearnModel()
	if err != nil {
		t.Fatalf("Failed to create Sklearn model: %v", err)
	}

	metadata = skModel.GetMetadata()
	if metadata["framework"] != "sklearn" {
		t.Errorf("Expected framework 'sklearn', got %v", metadata["framework"])
	}
}

func TestBatchProcessor(t *testing.T) {
	config := &PreprocessingConfig{
		FeatureScaling:   "minmax",
		HandleMissing:    "mean",
		OutlierDetection: false,
		FeatureSelection: false,
	}

	pipeline := NewPreprocessingPipeline(config)
	batchProcessor := NewBatchProcessor(pipeline, 2)

	if batchProcessor == nil {
		t.Fatal("Failed to create batch processor")
	}

	if batchProcessor.batchSize != 2 {
		t.Errorf("Expected batch size 2, got %d", batchProcessor.batchSize)
	}
}

func BenchmarkFeatureExtraction(b *testing.B) {
	processor := NewAssetDataPreprocessor()
	testData := map[string]interface{}{
		"open_ports":   []interface{}{80.0, 443.0, 22.0, 3306.0, 5432.0},
		"services":     []interface{}{"http", "https", "ssh", "mysql", "postgres"},
		"technologies": []interface{}{"nginx", "ssl", "mysql", "postgres", "redis"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.Preprocess(testData)
		if err != nil {
			b.Fatalf("Preprocessing failed: %v", err)
		}
	}
}

func BenchmarkFeatureScaling(b *testing.B) {
	config := &PreprocessingConfig{
		FeatureScaling: "standard",
	}
	pipeline := NewPreprocessingPipeline(config)
	
	features := make([]float64, 1000)
	for i := range features {
		features[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pipeline.scaleFeatures(features)
		if err != nil {
			b.Fatalf("Feature scaling failed: %v", err)
		}
	}
}
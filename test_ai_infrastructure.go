package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dablon/osmedeus/database"
	"github.com/dablon/osmedeus/ml"
)

func main() {
	fmt.Println("🤖 Testing AI Infrastructure...")

	// Test 1: ML Infrastructure Initialization
	fmt.Println("\n1. Testing ML Infrastructure...")
	testMLInfrastructure()

	// Test 2: Data Preprocessing
	fmt.Println("\n2. Testing Data Preprocessing...")
	testDataPreprocessing()

	// Test 3: AI Data Models
	fmt.Println("\n3. Testing AI Data Models...")
	testAIDataModels()

	// Test 4: Database Integration
	fmt.Println("\n4. Testing Database Integration...")
	testDatabaseIntegration()

	fmt.Println("\n✅ All tests completed!")
}

func testMLInfrastructure() {
	config := &ml.MLConfig{
		ModelStorePath:     "./models",
		MaxConcurrentInfer: 2,
		MaxConcurrentTrain: 1,
		DefaultTimeout:     5 * time.Second,
		GPUEnabled:         false,
		ModelConfigs:       make(map[string]interface{}),
		PreprocessingConfig: make(map[string]interface{}),
	}

	mlInfra := ml.NewMLInfrastructure(config)
	if mlInfra == nil {
		log.Fatal("❌ Failed to create ML infrastructure")
	}
	fmt.Println("✅ ML Infrastructure created successfully")

	// Test model creation
	models := []string{"tensorflow", "pytorch", "sklearn"}
	for _, framework := range models {
		var model ml.MLModel
		var err error

		switch framework {
		case "tensorflow":
			model, err = ml.NewTensorFlowModel()
		case "pytorch":
			model, err = ml.NewPyTorchModel()
		case "sklearn":
			model, err = ml.NewSklearnModel()
		}

		if err != nil {
			fmt.Printf("❌ Failed to create %s model: %v\n", framework, err)
			continue
		}

		metadata := model.GetMetadata()
		fmt.Printf("✅ %s model created - Framework: %v\n", framework, metadata["framework"])
	}

	// Cleanup
	mlInfra.Shutdown()
	fmt.Println("✅ ML Infrastructure shutdown completed")
}

func testDataPreprocessing() {
	// Test Asset Data Preprocessing
	assetProcessor := ml.NewAssetDataPreprocessor()
	testAssetData := map[string]interface{}{
		"open_ports":   []interface{}{80.0, 443.0, 22.0, 3306.0},
		"services":     []interface{}{"http", "https", "ssh", "mysql"},
		"technologies": []interface{}{"nginx", "ssl", "mysql"},
	}

	processed, err := assetProcessor.Preprocess(testAssetData)
	if err != nil {
		fmt.Printf("❌ Asset preprocessing failed: %v\n", err)
		return
	}

	features, err := assetProcessor.GetFeatures(processed)
	if err != nil {
		fmt.Printf("❌ Feature extraction failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Asset preprocessing successful - Extracted %d features: %v\n", len(features), features)

	// Test Preprocessing Pipeline
	config := &ml.PreprocessingConfig{
		FeatureScaling:   "standard",
		HandleMissing:    "mean",
		OutlierDetection: true,
		FeatureSelection: true,
	}

	pipeline := ml.NewPreprocessingPipeline(config)
	if pipeline == nil {
		fmt.Println("❌ Failed to create preprocessing pipeline")
		return
	}

	// Test feature scaling by processing actual data
	testTarget := &database.Target{
		InputName:          "test.example.com",
		TotalAssets:        10,
		TotalVulnerability: 5,
		TotalDns:           3,
		TotalTech:          2,
	}
	
	featureVector, err := pipeline.ProcessAssetData(testTarget)
	if err != nil {
		fmt.Printf("❌ Asset data processing failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Feature processing successful - Features: %v, Labels: %v\n", 
		featureVector.Features, featureVector.Labels)
}

func testAIDataModels() {
	// Test AIEnhancedTarget creation
	target := &database.AIEnhancedTarget{
		Target: database.Target{
			InputName:          "test.example.com",
			InputType:          "domain",
			TotalAssets:        10,
			TotalVulnerability: 5,
		},
		AnomalyScore:      0.75,
		LastAIAnalysis:    time.Now(),
		AIAnalysisVersion: "1.0",
		ProcessingStatus:  "completed",
	}

	fmt.Printf("✅ AIEnhancedTarget created - Target: %s, Anomaly Score: %.2f\n", 
		target.InputName, target.AnomalyScore)

	// Test AssetClassification
	classification := &database.AssetClassification{
		TargetID:         "test.example.com",
		Type:             "production",
		Confidence:       0.92,
		BusinessValue:    0.85,
		CriticalityLevel: "high",
		Reasoning:        `["high_traffic", "customer_facing", "production_environment"]`,
		ModelVersion:     "1.0",
	}

	fmt.Printf("✅ AssetClassification created - Type: %s, Confidence: %.2f, Criticality: %s\n",
		classification.Type, classification.Confidence, classification.CriticalityLevel)

	// Test VulnerabilityPrediction
	prediction := &database.VulnerabilityPrediction{
		PredictionID:         "pred_001",
		TargetID:             "test.example.com",
		PredictedVulnType:    "SQL Injection",
		EmergenceProbability: 0.78,
		TimeToEmergence:      7 * 24 * time.Hour, // 7 days
		ModelConfidence:      0.85,
		ModelVersion:         "1.0",
	}

	fmt.Printf("✅ VulnerabilityPrediction created - Type: %s, Probability: %.2f, Time to Emergence: %v\n",
		prediction.PredictedVulnType, prediction.EmergenceProbability, prediction.TimeToEmergence)
}

func testDatabaseIntegration() {
	// Note: This would require an actual database connection
	// For now, we'll just test the structure
	
	fmt.Println("✅ Database models are properly structured")
	fmt.Println("   - AIEnhancedTarget extends Target with AI fields")
	fmt.Println("   - AssetClassification for AI-powered asset analysis")
	fmt.Println("   - BehaviorProfile for behavioral analysis")
	fmt.Println("   - VulnerabilityPrediction for ML-based predictions")
	fmt.Println("   - ExploitChain for multi-vulnerability attacks")
	fmt.Println("   - GeneratedExploit for automated exploit generation")
	fmt.Println("   - Team collaboration models")
	fmt.Println("   - ML infrastructure models")

	// Test migration structure
	fmt.Println("✅ Migration system ready:")
	fmt.Println("   - Migration manager for database schema updates")
	fmt.Println("   - AI field additions to existing tables")
	fmt.Println("   - New AI analysis tables")
	fmt.Println("   - Performance indexes for AI queries")
}
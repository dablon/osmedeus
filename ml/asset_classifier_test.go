package ml

import (
	"testing"
	"time"

	"github.com/dablon/osmedeus/database"
)

func TestNewMLAssetClassifier(t *testing.T) {
	classifier, err := NewMLAssetClassifier(nil)
	if err != nil {
		t.Fatalf("Failed to create asset classifier: %v", err)
	}

	if classifier == nil {
		t.Fatal("Classifier is nil")
	}

	if classifier.config == nil {
		t.Fatal("Config is nil")
	}

	if classifier.featureExtractor == nil {
		t.Fatal("Feature extractor is nil")
	}

	if classifier.trainingData == nil {
		t.Fatal("Training data is nil")
	}

	if classifier.modelLoaded {
		t.Fatal("Model should not be loaded initially")
	}
}

func TestExtractAssetFeatures(t *testing.T) {
	classifier, err := NewMLAssetClassifier(nil)
	if err != nil {
		t.Fatalf("Failed to create asset classifier: %v", err)
	}

	target := &database.Target{
		InputName:           "example.com",
		InputType:           "domain",
		TotalAssets:         10,
		TotalDns:           5,
		TotalTech:          3,
		TotalVulnerability: 2,
		TotalScreenShot:    1,
		TotalDirb:          4,
		TotalLink:          6,
		TotalArchive:       0,
		TotalIPRange:       1,
		TotalCloud:         0,
		TotalCred:          0,
		IsWildCard:         false,
	}

	features, labels, err := classifier.extractAssetFeatures(target)
	if err != nil {
		t.Fatalf("Failed to extract features: %v", err)
	}

	if len(features) == 0 {
		t.Fatal("No features extracted")
	}

	if len(labels) != len(features) {
		t.Fatalf("Feature and label count mismatch: %d vs %d", len(features), len(labels))
	}

	// Check specific feature values
	expectedFeatures := map[string]float64{
		"total_assets":        10,
		"total_dns":          5,
		"total_tech":         3,
		"total_vulnerability": 2,
		"is_wildcard":        0,
	}

	for i, label := range labels {
		if expected, exists := expectedFeatures[label]; exists {
			if features[i] != expected {
				t.Errorf("Feature %s: expected %f, got %f", label, expected, features[i])
			}
		}
	}
}

func TestExtractDomainFeatures(t *testing.T) {
	classifier, err := NewMLAssetClassifier(nil)
	if err != nil {
		t.Fatalf("Failed to create asset classifier: %v", err)
	}

	testCases := []struct {
		domain   string
		expected map[int]float64 // index -> expected value
	}{
		{
			domain: "admin.example.com",
			expected: map[int]float64{
				0: 17, // domain length
				1: 1,  // subdomain count
				2: 1,  // has admin keywords
				3: 0,  // has api keywords
				4: 0,  // has dev keywords
				5: 0,  // has staging keywords
			},
		},
		{
			domain: "api.service.com",
			expected: map[int]float64{
				0: 15, // domain length
				1: 1,  // subdomain count
				2: 0,  // has admin keywords
				3: 1,  // has api keywords
				4: 0,  // has dev keywords
				5: 0,  // has staging keywords
			},
		},
		{
			domain: "dev.staging.example.com",
			expected: map[int]float64{
				0: 23, // domain length
				1: 2,  // subdomain count
				2: 0,  // has admin keywords
				3: 0,  // has api keywords
				4: 1,  // has dev keywords
				5: 1,  // has staging keywords
			},
		},
	}

	for _, tc := range testCases {
		features := classifier.extractDomainFeatures(tc.domain)
		
		if len(features) != 6 {
			t.Errorf("Expected 6 domain features, got %d", len(features))
			continue
		}

		for index, expected := range tc.expected {
			if features[index] != expected {
				t.Errorf("Domain %s, feature %d: expected %f, got %f", 
					tc.domain, index, expected, features[index])
			}
		}
	}
}

func TestCalculateBusinessValue(t *testing.T) {
	classifier, err := NewMLAssetClassifier(nil)
	if err != nil {
		t.Fatalf("Failed to create asset classifier: %v", err)
	}

	target := &database.Target{
		TotalAssets:        50,
		TotalVulnerability: 5,
	}

	testCases := []struct {
		assetType     string
		expectedRange [2]float64 // min, max
	}{
		{"production", [2]float64{0.8, 1.0}},
		{"admin", [2]float64{0.7, 1.0}},
		{"api", [2]float64{0.6, 1.0}},
		{"staging", [2]float64{0.4, 0.8}},
		{"development", [2]float64{0.1, 0.4}},
		{"unknown", [2]float64{0.3, 0.7}},
	}

	for _, tc := range testCases {
		value := classifier.calculateBusinessValue(tc.assetType, target)
		
		if value < tc.expectedRange[0] || value > tc.expectedRange[1] {
			t.Errorf("Asset type %s: business value %f not in expected range [%f, %f]",
				tc.assetType, value, tc.expectedRange[0], tc.expectedRange[1])
		}
	}
}

func TestDetermineCriticalityLevel(t *testing.T) {
	classifier, err := NewMLAssetClassifier(nil)
	if err != nil {
		t.Fatalf("Failed to create asset classifier: %v", err)
	}

	target := &database.Target{}

	testCases := []struct {
		businessValue float64
		expected      string
	}{
		{0.9, "critical"},
		{0.7, "high"},
		{0.5, "medium"},
		{0.3, "low"},
		{0.1, "low"},
	}

	for _, tc := range testCases {
		level := classifier.determineCriticalityLevel(tc.businessValue, target)
		if level != tc.expected {
			t.Errorf("Business value %f: expected %s, got %s", 
				tc.businessValue, tc.expected, level)
		}
	}
}

func TestApplyCustomRules(t *testing.T) {
	classifier, err := NewMLAssetClassifier(nil)
	if err != nil {
		t.Fatalf("Failed to create asset classifier: %v", err)
	}

	testCases := []struct {
		domain           string
		expectedType     string
		expectedCriticality string
		minBusinessValue float64
	}{
		{
			domain:           "admin.example.com",
			expectedType:     "admin",
			expectedCriticality: "high",
			minBusinessValue: 0.8,
		},
		{
			domain:           "api.service.com",
			expectedType:     "api",
			expectedCriticality: "",
			minBusinessValue: 0.7,
		},
		{
			domain:           "dev.example.com",
			expectedType:     "development",
			expectedCriticality: "low",
			minBusinessValue: 0.0,
		},
	}

	for _, tc := range testCases {
		classification := &database.AssetClassification{
			Type:             "unknown",
			BusinessValue:    0.5,
			CriticalityLevel: "medium",
		}

		target := &database.Target{
			InputName: tc.domain,
		}

		classifier.applyCustomRules(classification, target)

		if classification.Type != tc.expectedType {
			t.Errorf("Domain %s: expected type %s, got %s", 
				tc.domain, tc.expectedType, classification.Type)
		}

		if tc.expectedCriticality != "" && classification.CriticalityLevel != tc.expectedCriticality {
			t.Errorf("Domain %s: expected criticality %s, got %s", 
				tc.domain, tc.expectedCriticality, classification.CriticalityLevel)
		}

		if classification.BusinessValue < tc.minBusinessValue {
			t.Errorf("Domain %s: business value %f below minimum %f", 
				tc.domain, classification.BusinessValue, tc.minBusinessValue)
		}
	}
}

func TestCalculateCosineSimilarity(t *testing.T) {
	classifier, err := NewMLAssetClassifier(nil)
	if err != nil {
		t.Fatalf("Failed to create asset classifier: %v", err)
	}

	testCases := []struct {
		vec1     []float64
		vec2     []float64
		expected float64
		tolerance float64
	}{
		{
			vec1:     []float64{1, 0, 0},
			vec2:     []float64{1, 0, 0},
			expected: 1.0,
			tolerance: 0.001,
		},
		{
			vec1:     []float64{1, 0, 0},
			vec2:     []float64{0, 1, 0},
			expected: 0.0,
			tolerance: 0.001,
		},
		{
			vec1:     []float64{1, 1, 0},
			vec2:     []float64{1, 1, 0},
			expected: 1.0,
			tolerance: 0.001,
		},
		{
			vec1:     []float64{1, 2, 3},
			vec2:     []float64{2, 4, 6},
			expected: 1.0,
			tolerance: 0.001,
		},
	}

	for i, tc := range testCases {
		similarity := classifier.calculateCosineSimilarity(tc.vec1, tc.vec2)
		
		if abs(similarity-tc.expected) > tc.tolerance {
			t.Errorf("Test case %d: expected similarity %f, got %f", 
				i, tc.expected, similarity)
		}
	}
}

func TestAssessFeatureQuality(t *testing.T) {
	classifier, err := NewMLAssetClassifier(nil)
	if err != nil {
		t.Fatalf("Failed to create asset classifier: %v", err)
	}

	testCases := []struct {
		features []float64
		expected float64
	}{
		{[]float64{}, 0.0},
		{[]float64{0, 0, 0}, 0.0},
		{[]float64{1, 2, 3}, 1.0},
		{[]float64{1, 0, 3}, 2.0 / 3.0},
		{[]float64{0, 2, 0, 4}, 0.5},
	}

	for i, tc := range testCases {
		quality := classifier.assessFeatureQuality(tc.features)
		
		if abs(quality-tc.expected) > 0.001 {
			t.Errorf("Test case %d: expected quality %f, got %f", 
				i, tc.expected, quality)
		}
	}
}

func TestPrepareTrainingData(t *testing.T) {
	classifier, err := NewMLAssetClassifier(nil)
	if err != nil {
		t.Fatalf("Failed to create asset classifier: %v", err)
	}

	labeledAssets := []LabeledAsset{
		{
			TargetID:      "example1.com",
			AssetType:     "production",
			BusinessValue: 0.9,
			Features: map[string]interface{}{
				"technical": map[string]interface{}{
					"open_ports": 5.0,
					"services":   3.0,
				},
			},
			LabeledAt: time.Now(),
		},
		{
			TargetID:      "example2.com",
			AssetType:     "development",
			BusinessValue: 0.3,
			Features: map[string]interface{}{
				"technical": map[string]interface{}{
					"open_ports": 2.0,
					"services":   1.0,
				},
			},
			LabeledAt: time.Now(),
		},
	}

	trainingData, err := classifier.prepareTrainingData(labeledAssets)
	if err != nil {
		t.Fatalf("Failed to prepare training data: %v", err)
	}

	if len(trainingData.Features) != len(labeledAssets) {
		t.Errorf("Expected %d feature vectors, got %d", 
			len(labeledAssets), len(trainingData.Features))
	}

	if len(trainingData.Labels) != len(labeledAssets) {
		t.Errorf("Expected %d labels, got %d", 
			len(labeledAssets), len(trainingData.Labels))
	}

	// Check labels
	expectedLabels := []string{"production", "development"}
	for i, label := range trainingData.Labels {
		if label != expectedLabels[i] {
			t.Errorf("Label %d: expected %s, got %v", i, expectedLabels[i], label)
		}
	}
}

func TestCollectTrainingData(t *testing.T) {
	classifier, err := NewMLAssetClassifier(nil)
	if err != nil {
		t.Fatalf("Failed to create asset classifier: %v", err)
	}

	targets := []*database.Target{
		{
			InputName:          "example.com",
			TotalAssets:        10,
			TotalVulnerability: 2,
		},
	}

	classifications := []*database.AssetClassification{
		{
			TargetID:         "example.com",
			Type:             "production",
			BusinessValue:    0.8,
			CriticalityLevel: "high",
			Confidence:       0.9,
		},
	}

	initialCount := len(classifier.trainingData.LabeledAssets)

	err = classifier.CollectTrainingData(targets, classifications)
	if err != nil {
		t.Fatalf("Failed to collect training data: %v", err)
	}

	finalCount := len(classifier.trainingData.LabeledAssets)
	if finalCount != initialCount+1 {
		t.Errorf("Expected %d training samples, got %d", initialCount+1, finalCount)
	}

	// Check the collected data
	lastAsset := classifier.trainingData.LabeledAssets[finalCount-1]
	if lastAsset.TargetID != "example.com" {
		t.Errorf("Expected target ID 'example.com', got %s", lastAsset.TargetID)
	}

	if lastAsset.AssetType != "production" {
		t.Errorf("Expected asset type 'production', got %s", lastAsset.AssetType)
	}
}

func TestGetClassificationStats(t *testing.T) {
	classifier, err := NewMLAssetClassifier(nil)
	if err != nil {
		t.Fatalf("Failed to create asset classifier: %v", err)
	}

	// Add some training data
	classifier.trainingData.LabeledAssets = []LabeledAsset{
		{AssetType: "production", CriticalityLevel: "high"},
		{AssetType: "production", CriticalityLevel: "critical"},
		{AssetType: "development", CriticalityLevel: "low"},
		{AssetType: "api", CriticalityLevel: "medium"},
	}

	stats := classifier.GetClassificationStats()

	// Check asset type distribution
	typeDistribution, ok := stats["asset_type_distribution"].(map[string]int)
	if !ok {
		t.Fatal("Asset type distribution not found or wrong type")
	}

	expectedTypes := map[string]int{
		"production":  2,
		"development": 1,
		"api":         1,
	}

	for assetType, expectedCount := range expectedTypes {
		if count, exists := typeDistribution[assetType]; !exists || count != expectedCount {
			t.Errorf("Asset type %s: expected count %d, got %d", 
				assetType, expectedCount, count)
		}
	}

	// Check criticality distribution
	criticalityDistribution, ok := stats["criticality_distribution"].(map[string]int)
	if !ok {
		t.Fatal("Criticality distribution not found or wrong type")
	}

	expectedCriticalities := map[string]int{
		"high":     1,
		"critical": 1,
		"low":      1,
		"medium":   1,
	}

	for criticality, expectedCount := range expectedCriticalities {
		if count, exists := criticalityDistribution[criticality]; !exists || count != expectedCount {
			t.Errorf("Criticality %s: expected count %d, got %d", 
				criticality, expectedCount, count)
		}
	}

	// Check total samples
	totalSamples, ok := stats["total_training_samples"].(int)
	if !ok || totalSamples != 4 {
		t.Errorf("Expected 4 total training samples, got %v", totalSamples)
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Benchmark tests
func BenchmarkExtractAssetFeatures(b *testing.B) {
	classifier, err := NewMLAssetClassifier(nil)
	if err != nil {
		b.Fatalf("Failed to create asset classifier: %v", err)
	}

	target := &database.Target{
		InputName:           "example.com",
		TotalAssets:         10,
		TotalDns:           5,
		TotalTech:          3,
		TotalVulnerability: 2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := classifier.extractAssetFeatures(target)
		if err != nil {
			b.Fatalf("Feature extraction failed: %v", err)
		}
	}
}

func BenchmarkCalculateCosineSimilarity(b *testing.B) {
	classifier, err := NewMLAssetClassifier(nil)
	if err != nil {
		b.Fatalf("Failed to create asset classifier: %v", err)
	}

	vec1 := make([]float64, 100)
	vec2 := make([]float64, 100)
	for i := 0; i < 100; i++ {
		vec1[i] = float64(i)
		vec2[i] = float64(i * 2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		classifier.calculateCosineSimilarity(vec1, vec2)
	}
}
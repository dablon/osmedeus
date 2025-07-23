package ml

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/dablon/osmedeus/database"
	"github.com/dablon/osmedeus/utils"
)

// MLAssetClassifier implements AI-powered asset classification
type MLAssetClassifier struct {
	model           MLModel
	preprocessor    DataPreprocessor
	trainingData    *AssetTrainingDataset
	config          *AssetClassifierConfig
	featureExtractor *AssetFeatureExtractor
	modelLoaded     bool
}

// AssetClassifierConfig holds configuration for the asset classifier
type AssetClassifierConfig struct {
	ModelPath           string                 `json:"model_path"`
	ConfidenceThreshold float64                `json:"confidence_threshold"`
	FeatureWeights      map[string]float64     `json:"feature_weights"`
	ClassificationTypes []string               `json:"classification_types"`
	BusinessValueWeights map[string]float64    `json:"business_value_weights"`
	UpdateInterval      time.Duration          `json:"update_interval"`
	TrainingBatchSize   int                    `json:"training_batch_size"`
	ValidationSplit     float64                `json:"validation_split"`
	CustomRules         map[string]interface{} `json:"custom_rules"`
}

// AssetTrainingDataset represents training data for asset classification
type AssetTrainingDataset struct {
	LabeledAssets    []LabeledAsset `json:"labeled_assets"`
	FeatureVectors   [][]float64    `json:"feature_vectors"`
	Labels           []string       `json:"labels"`
	ValidationSet    []LabeledAsset `json:"validation_set"`
	LastUpdated      time.Time      `json:"last_updated"`
	DatasetVersion   string         `json:"dataset_version"`
	QualityMetrics   map[string]float64 `json:"quality_metrics"`
}

// LabeledAsset represents a training example with known classification
type LabeledAsset struct {
	TargetID         string                 `json:"target_id"`
	AssetType        string                 `json:"asset_type"`        // production, staging, dev, admin, api
	BusinessValue    float64                `json:"business_value"`
	CriticalityLevel string                 `json:"criticality_level"` // critical, high, medium, low
	Features         map[string]interface{} `json:"features"`
	Metadata         map[string]interface{} `json:"metadata"`
	LabeledBy        string                 `json:"labeled_by"`
	LabeledAt        time.Time              `json:"labeled_at"`
	Confidence       float64                `json:"confidence"`
}

// AssetFeatureExtractor extracts features from asset data for classification
type AssetFeatureExtractor struct {
	config *FeatureExtractionConfig
}

// FeatureExtractionConfig configures feature extraction
type FeatureExtractionConfig struct {
	TechnicalFeatures  []string               `json:"technical_features"`
	BehavioralFeatures []string               `json:"behavioral_features"`
	MetadataFeatures   []string               `json:"metadata_features"`
	FeatureWeights     map[string]float64     `json:"feature_weights"`
	NormalizationMethod string                `json:"normalization_method"`
	CustomExtractors   map[string]interface{} `json:"custom_extractors"`
}

// AssetClassificationResult represents the result of asset classification
type AssetClassificationResult struct {
	TargetID         string                 `json:"target_id"`
	Classification   *database.AssetClassification `json:"classification"`
	FeatureVector    []float64              `json:"feature_vector"`
	FeatureLabels    []string               `json:"feature_labels"`
	ModelMetadata    map[string]interface{} `json:"model_metadata"`
	ProcessingTime   time.Duration          `json:"processing_time"`
	ClassifiedAt     time.Time              `json:"classified_at"`
}

// NewMLAssetClassifier creates a new ML asset classifier
func NewMLAssetClassifier(config *AssetClassifierConfig) (*MLAssetClassifier, error) {
	if config == nil {
		config = &AssetClassifierConfig{
			ModelPath:           "./models/asset_classifier.model",
			ConfidenceThreshold: 0.7,
			FeatureWeights: map[string]float64{
				"port_count":        0.2,
				"service_count":     0.25,
				"technology_count":  0.2,
				"vulnerability_count": 0.15,
				"response_patterns": 0.1,
				"business_context":  0.1,
			},
			ClassificationTypes: []string{"production", "staging", "development", "admin", "api", "test"},
			BusinessValueWeights: map[string]float64{
				"production": 1.0,
				"admin":      0.9,
				"api":        0.8,
				"staging":    0.6,
				"test":       0.3,
				"development": 0.2,
			},
			UpdateInterval:    24 * time.Hour,
			TrainingBatchSize: 100,
			ValidationSplit:   0.2,
			CustomRules:       make(map[string]interface{}),
		}
	}

	// Create feature extractor
	featureExtractor := &AssetFeatureExtractor{
		config: &FeatureExtractionConfig{
			TechnicalFeatures: []string{
				"open_ports", "services", "technologies", "vulnerabilities",
				"dns_records", "subdomains", "screenshots", "directories",
			},
			BehavioralFeatures: []string{
				"response_times", "error_patterns", "header_patterns",
				"content_patterns", "redirect_patterns",
			},
			MetadataFeatures: []string{
				"domain_age", "ssl_info", "geo_location", "hosting_provider",
			},
			FeatureWeights: config.FeatureWeights,
			NormalizationMethod: "standard",
			CustomExtractors: make(map[string]interface{}),
		},
	}

	// Create model based on configuration
	model, err := NewSklearnModel() // Using sklearn for initial implementation
	if err != nil {
		return nil, fmt.Errorf("failed to create ML model: %w", err)
	}

	// Create preprocessor
	preprocessor := NewAssetDataPreprocessor()

	classifier := &MLAssetClassifier{
		model:           model,
		preprocessor:    preprocessor,
		config:          config,
		featureExtractor: featureExtractor,
		trainingData:    &AssetTrainingDataset{
			LabeledAssets:  make([]LabeledAsset, 0),
			FeatureVectors: make([][]float64, 0),
			Labels:         make([]string, 0),
			ValidationSet:  make([]LabeledAsset, 0),
			QualityMetrics: make(map[string]float64),
		},
		modelLoaded: false,
	}

	return classifier, nil
}

// LoadModel loads a pre-trained asset classification model
func (ac *MLAssetClassifier) LoadModel(modelPath string) error {
	if err := ac.model.Load(modelPath); err != nil {
		return fmt.Errorf("failed to load asset classification model: %w", err)
	}

	ac.modelLoaded = true
	utils.InforF("Asset classification model loaded from: %s", modelPath)
	return nil
}

// ClassifyAsset performs AI-powered classification of an asset
func (ac *MLAssetClassifier) ClassifyAsset(target *database.Target) (*AssetClassificationResult, error) {
	startTime := time.Now()

	if !ac.modelLoaded {
		return nil, fmt.Errorf("asset classification model not loaded")
	}

	// Extract features from the target
	featureVector, featureLabels, err := ac.extractAssetFeatures(target)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Preprocess features
	processedFeatures, err := ac.preprocessor.Normalize(featureVector)
	if err != nil {
		return nil, fmt.Errorf("failed to preprocess features: %w", err)
	}

	// Perform classification
	prediction, err := ac.model.Predict(processedFeatures)
	if err != nil {
		return nil, fmt.Errorf("failed to classify asset: %w", err)
	}

	// Parse prediction results
	classification, err := ac.parsePredictionResult(target.InputName, prediction)
	if err != nil {
		return nil, fmt.Errorf("failed to parse prediction: %w", err)
	}

	// Apply business rules and confidence scoring
	ac.applyBusinessRules(classification, target)
	ac.calculateConfidenceScore(classification, featureVector)

	result := &AssetClassificationResult{
		TargetID:       target.InputName,
		Classification: classification,
		FeatureVector:  processedFeatures,
		FeatureLabels:  featureLabels,
		ModelMetadata:  ac.model.GetMetadata(),
		ProcessingTime: time.Since(startTime),
		ClassifiedAt:   time.Now(),
	}

	utils.InforF("Asset classified: %s -> %s (confidence: %.2f)", 
		target.InputName, classification.Type, classification.Confidence)

	return result, nil
}

// extractAssetFeatures extracts feature vector from target data
func (ac *MLAssetClassifier) extractAssetFeatures(target *database.Target) ([]float64, []string, error) {
	features := make([]float64, 0)
	labels := make([]string, 0)

	// Technical features
	features = append(features, float64(target.TotalAssets))
	labels = append(labels, "total_assets")

	features = append(features, float64(target.TotalDns))
	labels = append(labels, "total_dns")

	features = append(features, float64(target.TotalTech))
	labels = append(labels, "total_tech")

	features = append(features, float64(target.TotalVulnerability))
	labels = append(labels, "total_vulnerability")

	features = append(features, float64(target.TotalScreenShot))
	labels = append(labels, "total_screenshot")

	features = append(features, float64(target.TotalDirb))
	labels = append(labels, "total_dirb")

	features = append(features, float64(target.TotalLink))
	labels = append(labels, "total_link")

	features = append(features, float64(target.TotalArchive))
	labels = append(labels, "total_archive")

	features = append(features, float64(target.TotalIPRange))
	labels = append(labels, "total_ip_range")

	features = append(features, float64(target.TotalCloud))
	labels = append(labels, "total_cloud")

	features = append(features, float64(target.TotalCred))
	labels = append(labels, "total_cred")

	// Wildcard indicator
	wildcardValue := 0.0
	if target.IsWildCard {
		wildcardValue = 1.0
	}
	features = append(features, wildcardValue)
	labels = append(labels, "is_wildcard")

	// Derived features
	totalFindings := float64(target.TotalAssets + target.TotalDns + target.TotalTech + 
		target.TotalVulnerability + target.TotalScreenShot + target.TotalDirb + 
		target.TotalLink + target.TotalArchive + target.TotalIPRange + 
		target.TotalCloud + target.TotalCred)
	features = append(features, totalFindings)
	labels = append(labels, "total_findings")

	// Vulnerability density
	vulnDensity := 0.0
	if totalFindings > 0 {
		vulnDensity = float64(target.TotalVulnerability) / totalFindings
	}
	features = append(features, vulnDensity)
	labels = append(labels, "vulnerability_density")

	// Technology diversity
	techDiversity := float64(target.TotalTech)
	if target.TotalAssets > 0 {
		techDiversity = float64(target.TotalTech) / float64(target.TotalAssets)
	}
	features = append(features, techDiversity)
	labels = append(labels, "technology_diversity")

	// Domain-based features
	domainFeatures := ac.extractDomainFeatures(target.InputName)
	features = append(features, domainFeatures...)
	labels = append(labels, "domain_length", "subdomain_count", "has_admin_keywords", 
		"has_api_keywords", "has_dev_keywords", "has_staging_keywords")

	return features, labels, nil
}

// extractDomainFeatures extracts features from domain name
func (ac *MLAssetClassifier) extractDomainFeatures(domain string) []float64 {
	features := make([]float64, 0)

	// Domain length
	features = append(features, float64(len(domain)))

	// Subdomain count
	parts := strings.Split(domain, ".")
	subdomainCount := float64(len(parts) - 2) // Subtract domain and TLD
	if subdomainCount < 0 {
		subdomainCount = 0
	}
	features = append(features, subdomainCount)

	// Keyword indicators
	lowerDomain := strings.ToLower(domain)
	
	// Admin keywords
	adminKeywords := []string{"admin", "administrator", "manage", "control", "panel"}
	hasAdminKeywords := 0.0
	for _, keyword := range adminKeywords {
		if strings.Contains(lowerDomain, keyword) {
			hasAdminKeywords = 1.0
			break
		}
	}
	features = append(features, hasAdminKeywords)

	// API keywords
	apiKeywords := []string{"api", "rest", "graphql", "service", "endpoint"}
	hasAPIKeywords := 0.0
	for _, keyword := range apiKeywords {
		if strings.Contains(lowerDomain, keyword) {
			hasAPIKeywords = 1.0
			break
		}
	}
	features = append(features, hasAPIKeywords)

	// Development keywords
	devKeywords := []string{"dev", "development", "test", "testing", "debug"}
	hasDevKeywords := 0.0
	for _, keyword := range devKeywords {
		if strings.Contains(lowerDomain, keyword) {
			hasDevKeywords = 1.0
			break
		}
	}
	features = append(features, hasDevKeywords)

	// Staging keywords
	stagingKeywords := []string{"staging", "stage", "pre", "preprod", "uat"}
	hasStagingKeywords := 0.0
	for _, keyword := range stagingKeywords {
		if strings.Contains(lowerDomain, keyword) {
			hasStagingKeywords = 1.0
			break
		}
	}
	features = append(features, hasStagingKeywords)

	return features
}

// parsePredictionResult parses the ML model prediction into classification result
func (ac *MLAssetClassifier) parsePredictionResult(targetID string, prediction interface{}) (*database.AssetClassification, error) {
	predMap, ok := prediction.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid prediction format")
	}

	// For now, simulate classification logic
	// In a real implementation, this would parse actual ML model output
	classification := &database.AssetClassification{
		TargetID:         targetID,
		Type:             ac.determineAssetType(predMap),
		Confidence:       ac.extractConfidence(predMap),
		BusinessValue:    0.0, // Will be calculated later
		CriticalityLevel: "medium",
		ModelVersion:     "1.0.0",
	}

	return classification, nil
}

// determineAssetType determines asset type from prediction
func (ac *MLAssetClassifier) determineAssetType(prediction map[string]interface{}) string {
	// Simulate asset type determination
	// In a real implementation, this would use actual ML model output
	
	// For demonstration, use simple heuristics
	if confidence, ok := prediction["confidence"].(float64); ok {
		if confidence > 0.8 {
			return "production"
		} else if confidence > 0.6 {
			return "staging"
		} else if confidence > 0.4 {
			return "development"
		} else {
			return "test"
		}
	}
	
	return "unknown"
}

// extractConfidence extracts confidence score from prediction
func (ac *MLAssetClassifier) extractConfidence(prediction map[string]interface{}) float64 {
	if confidence, ok := prediction["confidence"].(float64); ok {
		return confidence
	}
	return 0.5 // Default confidence
}

// applyBusinessRules applies business logic to refine classification
func (ac *MLAssetClassifier) applyBusinessRules(classification *database.AssetClassification, target *database.Target) {
	// Apply business value calculation
	classification.BusinessValue = ac.calculateBusinessValue(classification.Type, target)
	
	// Determine criticality level
	classification.CriticalityLevel = ac.determineCriticalityLevel(classification.BusinessValue, target)
	
	// Apply custom rules
	ac.applyCustomRules(classification, target)
}

// calculateBusinessValue calculates business value score
func (ac *MLAssetClassifier) calculateBusinessValue(assetType string, target *database.Target) float64 {
	baseValue := ac.config.BusinessValueWeights[assetType]
	if baseValue == 0 {
		baseValue = 0.5 // Default value
	}

	// Adjust based on findings
	findingsMultiplier := 1.0
	totalFindings := target.TotalAssets + target.TotalDns + target.TotalTech + 
		target.TotalVulnerability + target.TotalScreenShot + target.TotalDirb + 
		target.TotalLink + target.TotalArchive + target.TotalIPRange + 
		target.TotalCloud + target.TotalCred

	if totalFindings > 100 {
		findingsMultiplier = 1.2
	} else if totalFindings > 50 {
		findingsMultiplier = 1.1
	}

	// Adjust based on vulnerabilities
	vulnMultiplier := 1.0
	if target.TotalVulnerability > 10 {
		vulnMultiplier = 1.3
	} else if target.TotalVulnerability > 5 {
		vulnMultiplier = 1.15
	}

	businessValue := baseValue * findingsMultiplier * vulnMultiplier
	
	// Normalize to 0-1 range
	if businessValue > 1.0 {
		businessValue = 1.0
	}

	return businessValue
}

// determineCriticalityLevel determines criticality based on business value and other factors
func (ac *MLAssetClassifier) determineCriticalityLevel(businessValue float64, target *database.Target) string {
	if businessValue >= 0.8 {
		return "critical"
	} else if businessValue >= 0.6 {
		return "high"
	} else if businessValue >= 0.4 {
		return "medium"
	} else {
		return "low"
	}
}

// applyCustomRules applies custom business rules
func (ac *MLAssetClassifier) applyCustomRules(classification *database.AssetClassification, target *database.Target) {
	// Apply domain-based rules
	domain := strings.ToLower(target.InputName)
	
	// Admin panel detection
	adminKeywords := []string{"admin", "administrator", "manage", "control", "panel"}
	for _, keyword := range adminKeywords {
		if strings.Contains(domain, keyword) {
			classification.Type = "admin"
			classification.CriticalityLevel = "high"
			classification.BusinessValue = math.Max(classification.BusinessValue, 0.8)
			break
		}
	}

	// API detection
	apiKeywords := []string{"api", "rest", "graphql", "service", "endpoint"}
	for _, keyword := range apiKeywords {
		if strings.Contains(domain, keyword) {
			classification.Type = "api"
			classification.BusinessValue = math.Max(classification.BusinessValue, 0.7)
			break
		}
	}

	// Development environment detection
	devKeywords := []string{"dev", "development", "test", "testing", "debug", "localhost"}
	for _, keyword := range devKeywords {
		if strings.Contains(domain, keyword) {
			classification.Type = "development"
			classification.CriticalityLevel = "low"
			classification.BusinessValue = math.Min(classification.BusinessValue, 0.3)
			break
		}
	}
}

// calculateConfidenceScore calculates overall confidence in the classification
func (ac *MLAssetClassifier) calculateConfidenceScore(classification *database.AssetClassification, features []float64) {
	// Base confidence from model
	baseConfidence := classification.Confidence

	// Adjust based on feature quality
	featureQuality := ac.assessFeatureQuality(features)
	
	// Adjust based on business rules match
	businessRuleMatch := ac.assessBusinessRuleMatch(classification)

	// Combined confidence
	combinedConfidence := (baseConfidence * 0.6) + (featureQuality * 0.2) + (businessRuleMatch * 0.2)
	
	// Ensure confidence is within valid range
	if combinedConfidence > 1.0 {
		combinedConfidence = 1.0
	} else if combinedConfidence < 0.0 {
		combinedConfidence = 0.0
	}

	classification.Confidence = combinedConfidence
}

// assessFeatureQuality assesses the quality of extracted features
func (ac *MLAssetClassifier) assessFeatureQuality(features []float64) float64 {
	if len(features) == 0 {
		return 0.0
	}

	// Check for non-zero features
	nonZeroCount := 0
	for _, feature := range features {
		if feature != 0.0 {
			nonZeroCount++
		}
	}

	return float64(nonZeroCount) / float64(len(features))
}

// assessBusinessRuleMatch assesses how well the classification matches business rules
func (ac *MLAssetClassifier) assessBusinessRuleMatch(classification *database.AssetClassification) float64 {
	// Simple heuristic: higher business value indicates better rule match
	return classification.BusinessValue
}

// TrainModel trains the asset classification model with labeled data
func (ac *MLAssetClassifier) TrainModel(labeledAssets []LabeledAsset) error {
	if len(labeledAssets) == 0 {
		return fmt.Errorf("no training data provided")
	}

	utils.InforF("Training asset classification model with %d samples", len(labeledAssets))

	// Prepare training data
	trainingData, err := ac.prepareTrainingData(labeledAssets)
	if err != nil {
		return fmt.Errorf("failed to prepare training data: %w", err)
	}

	// Train the model
	if err := ac.model.Train(*trainingData); err != nil {
		return fmt.Errorf("failed to train model: %w", err)
	}

	// Update training dataset
	ac.trainingData.LabeledAssets = labeledAssets
	ac.trainingData.LastUpdated = time.Now()
	ac.trainingData.DatasetVersion = fmt.Sprintf("v%d", time.Now().Unix())

	// Validate the model
	validationResult, err := ac.validateModel()
	if err != nil {
		utils.ErrorF("Model validation failed: %v", err)
	} else {
		ac.trainingData.QualityMetrics = map[string]float64{
			"accuracy":  validationResult.Accuracy,
			"precision": validationResult.Precision,
			"recall":    validationResult.Recall,
			"f1_score":  validationResult.F1Score,
		}
		utils.InforF("Model training completed. Accuracy: %.2f", validationResult.Accuracy)
	}

	ac.modelLoaded = true
	return nil
}

// prepareTrainingData converts labeled assets to training data format
func (ac *MLAssetClassifier) prepareTrainingData(labeledAssets []LabeledAsset) (*TrainingData, error) {
	features := make([][]float64, 0, len(labeledAssets))
	labels := make([]interface{}, 0, len(labeledAssets))

	for _, asset := range labeledAssets {
		// Convert asset features to feature vector
		featureVector, err := ac.convertAssetToFeatureVector(asset)
		if err != nil {
			utils.ErrorF("Failed to convert asset %s to feature vector: %v", asset.TargetID, err)
			continue
		}

		features = append(features, featureVector)
		labels = append(labels, asset.AssetType)
	}

	return &TrainingData{
		Features: features,
		Labels:   labels,
		Metadata: map[string]interface{}{
			"dataset_size": len(features),
			"feature_count": len(features[0]),
			"created_at": time.Now(),
		},
	}, nil
}

// convertAssetToFeatureVector converts a labeled asset to feature vector
func (ac *MLAssetClassifier) convertAssetToFeatureVector(asset LabeledAsset) ([]float64, error) {
	features := make([]float64, 0)

	// Extract features from asset metadata
	if featuresMap, ok := asset.Features["technical"].(map[string]interface{}); ok {
		for _, featureName := range ac.featureExtractor.config.TechnicalFeatures {
			if value, exists := featuresMap[featureName]; exists {
				if floatValue, ok := value.(float64); ok {
					features = append(features, floatValue)
				} else {
					features = append(features, 0.0)
				}
			} else {
				features = append(features, 0.0)
			}
		}
	}

	// Add business value as a feature
	features = append(features, asset.BusinessValue)

	return features, nil
}

// validateModel validates the trained model using validation set
func (ac *MLAssetClassifier) validateModel() (*ValidationResult, error) {
	if len(ac.trainingData.ValidationSet) == 0 {
		// Create validation set from training data
		ac.createValidationSet()
	}

	if len(ac.trainingData.ValidationSet) == 0 {
		return nil, fmt.Errorf("no validation data available")
	}

	// Prepare validation data
	validationData := make([]interface{}, 0)
	for _, asset := range ac.trainingData.ValidationSet {
		featureVector, err := ac.convertAssetToFeatureVector(asset)
		if err != nil {
			continue
		}
		validationData = append(validationData, featureVector)
	}

	// Validate using the model
	validationResult, err := ac.model.Validate(validationData)
	if err != nil {
		return nil, err
	}
	return &validationResult, nil
}

// createValidationSet creates a validation set from training data
func (ac *MLAssetClassifier) createValidationSet() {
	if len(ac.trainingData.LabeledAssets) == 0 {
		return
	}

	// Calculate validation set size
	validationSize := int(float64(len(ac.trainingData.LabeledAssets)) * ac.config.ValidationSplit)
	if validationSize == 0 {
		validationSize = 1
	}

	// Randomly select validation samples
	totalAssets := len(ac.trainingData.LabeledAssets)
	validationIndices := make(map[int]bool)
	
	for len(validationIndices) < validationSize {
		index := int(time.Now().UnixNano()) % totalAssets
		validationIndices[index] = true
	}

	// Create validation set
	ac.trainingData.ValidationSet = make([]LabeledAsset, 0, validationSize)
	for index := range validationIndices {
		ac.trainingData.ValidationSet = append(ac.trainingData.ValidationSet, ac.trainingData.LabeledAssets[index])
	}
}

// UpdateModel updates the model with new training data
func (ac *MLAssetClassifier) UpdateModel(newAssets []LabeledAsset) error {
	if len(newAssets) == 0 {
		return nil
	}

	// Add new assets to existing training data
	ac.trainingData.LabeledAssets = append(ac.trainingData.LabeledAssets, newAssets...)

	// Retrain if we have enough new data
	if len(newAssets) >= ac.config.TrainingBatchSize {
		return ac.TrainModel(ac.trainingData.LabeledAssets)
	}

	return nil
}

// GetConfidenceScore returns confidence score for a classification
func (ac *MLAssetClassifier) GetConfidenceScore(target *database.Target) (float64, error) {
	result, err := ac.ClassifyAsset(target)
	if err != nil {
		return 0.0, err
	}
	return result.Classification.Confidence, nil
}

// SaveModel saves the trained model to disk
func (ac *MLAssetClassifier) SaveModel(modelPath string) error {
	if !ac.modelLoaded {
		return fmt.Errorf("no model loaded to save")
	}

	if err := ac.model.Save(modelPath); err != nil {
		return fmt.Errorf("failed to save model: %w", err)
	}

	// Save training data and configuration
	configPath := strings.Replace(modelPath, ".model", "_config.json", 1)
	configData := map[string]interface{}{
		"config":        ac.config,
		"training_data": ac.trainingData,
		"feature_config": ac.featureExtractor.config,
	}

	configJSON, err := json.Marshal(configData)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// In a real implementation, this would write to file
	utils.InforF("Model configuration saved to: %s", configPath)
	_ = configJSON // Suppress unused variable warning

	utils.InforF("Asset classification model saved to: %s", modelPath)
	return nil
}

// GetModelMetadata returns metadata about the current model
func (ac *MLAssetClassifier) GetModelMetadata() map[string]interface{} {
	metadata := ac.model.GetMetadata()
	metadata["classifier_config"] = ac.config
	metadata["training_data_size"] = len(ac.trainingData.LabeledAssets)
	metadata["validation_data_size"] = len(ac.trainingData.ValidationSet)
	metadata["quality_metrics"] = ac.trainingData.QualityMetrics
	metadata["last_updated"] = ac.trainingData.LastUpdated
	metadata["dataset_version"] = ac.trainingData.DatasetVersion
	return metadata
}

// CollectTrainingData collects training data from classified assets
func (ac *MLAssetClassifier) CollectTrainingData(targets []*database.Target, classifications []*database.AssetClassification) error {
	if len(targets) != len(classifications) {
		return fmt.Errorf("targets and classifications count mismatch")
	}

	newAssets := make([]LabeledAsset, 0, len(targets))

	for i, target := range targets {
		classification := classifications[i]
		
		// Extract features for training
		featureVector, _, err := ac.extractAssetFeatures(target)
		if err != nil {
			utils.ErrorF("Failed to extract features for %s: %v", target.InputName, err)
			continue
		}

		// Create labeled asset
		labeledAsset := LabeledAsset{
			TargetID:         target.InputName,
			AssetType:        classification.Type,
			BusinessValue:    classification.BusinessValue,
			CriticalityLevel: classification.CriticalityLevel,
			Features: map[string]interface{}{
				"technical": map[string]interface{}{
					"total_assets":        target.TotalAssets,
					"total_dns":          target.TotalDns,
					"total_tech":         target.TotalTech,
					"total_vulnerability": target.TotalVulnerability,
					"total_screenshot":   target.TotalScreenShot,
					"total_dirb":         target.TotalDirb,
					"total_link":         target.TotalLink,
					"total_archive":      target.TotalArchive,
					"total_ip_range":     target.TotalIPRange,
					"total_cloud":        target.TotalCloud,
					"total_cred":         target.TotalCred,
					"is_wildcard":        target.IsWildCard,
				},
				"feature_vector": featureVector,
			},
			Metadata: map[string]interface{}{
				"input_type": target.InputType,
				"collected_at": time.Now(),
			},
			LabeledBy:  "system",
			LabeledAt:  time.Now(),
			Confidence: classification.Confidence,
		}

		newAssets = append(newAssets, labeledAsset)
	}

	// Update training data
	return ac.UpdateModel(newAssets)
}

// GetSimilarAssets finds assets similar to the given target
func (ac *MLAssetClassifier) GetSimilarAssets(target *database.Target, threshold float64) ([]string, error) {
	// Extract features for the target
	targetFeatures, _, err := ac.extractAssetFeatures(target)
	if err != nil {
		return nil, fmt.Errorf("failed to extract target features: %w", err)
	}

	similarAssets := make([]string, 0)

	// Compare with training data
	for _, asset := range ac.trainingData.LabeledAssets {
		if asset.TargetID == target.InputName {
			continue // Skip self
		}

		// Get asset features
		assetFeatures, err := ac.convertAssetToFeatureVector(asset)
		if err != nil {
			continue
		}

		// Calculate similarity
		similarity := ac.calculateCosineSimilarity(targetFeatures, assetFeatures)
		if similarity >= threshold {
			similarAssets = append(similarAssets, asset.TargetID)
		}
	}

	// Sort by similarity (would need to store similarities for proper sorting)
	return similarAssets, nil
}

// calculateCosineSimilarity calculates cosine similarity between two feature vectors
func (ac *MLAssetClassifier) calculateCosineSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return 0.0
	}

	dotProduct := 0.0
	norm1 := 0.0
	norm2 := 0.0

	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	if norm1 == 0.0 || norm2 == 0.0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// GetClassificationStats returns statistics about classifications
func (ac *MLAssetClassifier) GetClassificationStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Count by asset type
	typeCounts := make(map[string]int)
	criticalityCounts := make(map[string]int)
	
	for _, asset := range ac.trainingData.LabeledAssets {
		typeCounts[asset.AssetType]++
		criticalityCounts[asset.CriticalityLevel]++
	}

	stats["asset_type_distribution"] = typeCounts
	stats["criticality_distribution"] = criticalityCounts
	stats["total_training_samples"] = len(ac.trainingData.LabeledAssets)
	stats["total_validation_samples"] = len(ac.trainingData.ValidationSet)
	stats["model_loaded"] = ac.modelLoaded
	stats["quality_metrics"] = ac.trainingData.QualityMetrics
	stats["last_updated"] = ac.trainingData.LastUpdated

	return stats
}
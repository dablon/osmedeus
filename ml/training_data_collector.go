package ml

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dablon/osmedeus/database"
	"github.com/dablon/osmedeus/utils"
)

// TrainingDataCollector manages the collection and storage of training data
type TrainingDataCollector struct {
	config     *CollectorConfig
	dataPath   string
	classifier *MLAssetClassifier
}

// CollectorConfig holds configuration for training data collection
type CollectorConfig struct {
	DataStorePath       string        `json:"data_store_path"`
	CollectionInterval  time.Duration `json:"collection_interval"`
	MinSamplesPerType   int           `json:"min_samples_per_type"`
	MaxSamplesPerType   int           `json:"max_samples_per_type"`
	QualityThreshold    float64       `json:"quality_threshold"`
	AutoLabelThreshold  float64       `json:"auto_label_threshold"`
	ValidationRatio     float64       `json:"validation_ratio"`
	EnableAutoLabeling  bool          `json:"enable_auto_labeling"`
	LabelingRules       []LabelingRule `json:"labeling_rules"`
}

// LabelingRule defines automatic labeling rules
type LabelingRule struct {
	Name        string                 `json:"name"`
	Conditions  map[string]interface{} `json:"conditions"`
	AssetType   string                 `json:"asset_type"`
	Confidence  float64                `json:"confidence"`
	Priority    int                    `json:"priority"`
}

// TrainingDataset represents a complete training dataset
type TrainingDataset struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Version         string                 `json:"version"`
	LabeledAssets   []LabeledAsset         `json:"labeled_assets"`
	ValidationSet   []LabeledAsset         `json:"validation_set"`
	Statistics      DatasetStatistics      `json:"statistics"`
	QualityMetrics  map[string]float64     `json:"quality_metrics"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// DatasetStatistics provides statistics about the training dataset
type DatasetStatistics struct {
	TotalSamples        int                `json:"total_samples"`
	ValidationSamples   int                `json:"validation_samples"`
	TypeDistribution    map[string]int     `json:"type_distribution"`
	CriticalityDistribution map[string]int `json:"criticality_distribution"`
	FeatureStatistics   map[string]FeatureStats `json:"feature_statistics"`
	QualityScore        float64            `json:"quality_score"`
	LastUpdated         time.Time          `json:"last_updated"`
}

// FeatureStats provides statistics for individual features
type FeatureStats struct {
	Mean     float64 `json:"mean"`
	StdDev   float64 `json:"std_dev"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	NonZero  int     `json:"non_zero"`
	Missing  int     `json:"missing"`
}

// NewTrainingDataCollector creates a new training data collector
func NewTrainingDataCollector(config *CollectorConfig, classifier *MLAssetClassifier) *TrainingDataCollector {
	if config == nil {
		config = &CollectorConfig{
			DataStorePath:       "./data/training",
			CollectionInterval:  24 * time.Hour,
			MinSamplesPerType:   50,
			MaxSamplesPerType:   1000,
			QualityThreshold:    0.7,
			AutoLabelThreshold:  0.8,
			ValidationRatio:     0.2,
			EnableAutoLabeling:  true,
			LabelingRules:       getDefaultLabelingRules(),
		}
	}

	// Ensure data directory exists
	if err := os.MkdirAll(config.DataStorePath, 0755); err != nil {
		utils.ErrorF("Failed to create data directory: %v", err)
	}

	return &TrainingDataCollector{
		config:     config,
		dataPath:   config.DataStorePath,
		classifier: classifier,
	}
}

// getDefaultLabelingRules returns default automatic labeling rules
func getDefaultLabelingRules() []LabelingRule {
	return []LabelingRule{
		{
			Name: "Admin Panel Detection",
			Conditions: map[string]interface{}{
				"domain_keywords": []string{"admin", "administrator", "manage", "control", "panel"},
				"min_vulnerabilities": 1,
			},
			AssetType:  "admin",
			Confidence: 0.9,
			Priority:   1,
		},
		{
			Name: "API Endpoint Detection",
			Conditions: map[string]interface{}{
				"domain_keywords": []string{"api", "rest", "graphql", "service", "endpoint"},
				"min_tech_count": 2,
			},
			AssetType:  "api",
			Confidence: 0.85,
			Priority:   2,
		},
		{
			Name: "Development Environment",
			Conditions: map[string]interface{}{
				"domain_keywords": []string{"dev", "development", "test", "testing", "debug", "localhost"},
				"max_business_value": 0.4,
			},
			AssetType:  "development",
			Confidence: 0.8,
			Priority:   3,
		},
		{
			Name: "Staging Environment",
			Conditions: map[string]interface{}{
				"domain_keywords": []string{"staging", "stage", "pre", "preprod", "uat"},
			},
			AssetType:  "staging",
			Confidence: 0.8,
			Priority:   4,
		},
		{
			Name: "Production Environment",
			Conditions: map[string]interface{}{
				"min_total_findings": 20,
				"min_vulnerabilities": 3,
				"exclude_keywords": []string{"dev", "test", "staging"},
			},
			AssetType:  "production",
			Confidence: 0.75,
			Priority:   5,
		},
	}
}

// CollectFromTargets collects training data from a list of targets
func (tdc *TrainingDataCollector) CollectFromTargets(targets []*database.Target) (*TrainingDataset, error) {
	utils.InforF("Collecting training data from %d targets", len(targets))

	labeledAssets := make([]LabeledAsset, 0)
	
	for _, target := range targets {
		// Apply automatic labeling if enabled
		if tdc.config.EnableAutoLabeling {
			labeledAsset, err := tdc.autoLabelTarget(target)
			if err != nil {
				utils.ErrorF("Failed to auto-label target %s: %v", target.InputName, err)
				continue
			}
			
			if labeledAsset != nil {
				labeledAssets = append(labeledAssets, *labeledAsset)
			}
		}
	}

	// Create training dataset
	dataset := &TrainingDataset{
		ID:            fmt.Sprintf("dataset_%d", time.Now().Unix()),
		Name:          fmt.Sprintf("Auto-collected Dataset %s", time.Now().Format("2006-01-02")),
		Description:   "Automatically collected training data from target analysis",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Version:       "1.0",
		LabeledAssets: labeledAssets,
		Metadata: map[string]interface{}{
			"collection_method": "automatic",
			"source_targets":    len(targets),
			"labeling_rules":    len(tdc.config.LabelingRules),
		},
	}

	// Split into training and validation sets
	tdc.createValidationSplit(dataset)

	// Calculate statistics
	tdc.calculateDatasetStatistics(dataset)

	// Save dataset
	if err := tdc.SaveDataset(dataset); err != nil {
		return nil, fmt.Errorf("failed to save dataset: %w", err)
	}

	utils.InforF("Collected %d labeled assets for training", len(labeledAssets))
	return dataset, nil
}

// autoLabelTarget automatically labels a target based on configured rules
func (tdc *TrainingDataCollector) autoLabelTarget(target *database.Target) (*LabeledAsset, error) {
	// Extract features for analysis
	features, _, err := tdc.classifier.extractAssetFeatures(target)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Apply labeling rules in priority order
	for _, rule := range tdc.config.LabelingRules {
		if tdc.matchesRule(target, features, rule) {
			// Create labeled asset
			labeledAsset := &LabeledAsset{
				TargetID:         target.InputName,
				AssetType:        rule.AssetType,
				BusinessValue:    tdc.calculateBusinessValueFromRule(target, rule),
				CriticalityLevel: tdc.determineCriticalityFromType(rule.AssetType),
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
					"feature_vector": features,
				},
				Metadata: map[string]interface{}{
					"input_type":     target.InputType,
					"labeling_rule":  rule.Name,
					"auto_labeled":   true,
					"rule_priority":  rule.Priority,
				},
				LabeledBy:  "auto_labeler",
				LabeledAt:  time.Now(),
				Confidence: rule.Confidence,
			}

			return labeledAsset, nil
		}
	}

	return nil, nil // No matching rule found
}

// matchesRule checks if a target matches a labeling rule
func (tdc *TrainingDataCollector) matchesRule(target *database.Target, features []float64, rule LabelingRule) bool {
	// Check domain keywords
	if keywords, exists := rule.Conditions["domain_keywords"]; exists {
		if keywordList, ok := keywords.([]string); ok {
			domainLower := strings.ToLower(target.InputName)
			hasKeyword := false
			for _, keyword := range keywordList {
				if strings.Contains(domainLower, keyword) {
					hasKeyword = true
					break
				}
			}
			if !hasKeyword {
				return false
			}
		}
	}

	// Check exclude keywords
	if excludeKeywords, exists := rule.Conditions["exclude_keywords"]; exists {
		if keywordList, ok := excludeKeywords.([]string); ok {
			domainLower := strings.ToLower(target.InputName)
			for _, keyword := range keywordList {
				if strings.Contains(domainLower, keyword) {
					return false // Exclude if any exclude keyword is found
				}
			}
		}
	}

	// Check minimum vulnerabilities
	if minVulns, exists := rule.Conditions["min_vulnerabilities"]; exists {
		if minCount, ok := minVulns.(int); ok {
			if target.TotalVulnerability < minCount {
				return false
			}
		}
	}

	// Check minimum tech count
	if minTech, exists := rule.Conditions["min_tech_count"]; exists {
		if minCount, ok := minTech.(int); ok {
			if target.TotalTech < minCount {
				return false
			}
		}
	}

	// Check minimum total findings
	if minFindings, exists := rule.Conditions["min_total_findings"]; exists {
		if minCount, ok := minFindings.(int); ok {
			totalFindings := target.TotalAssets + target.TotalDns + target.TotalTech + 
				target.TotalVulnerability + target.TotalScreenShot + target.TotalDirb + 
				target.TotalLink + target.TotalArchive + target.TotalIPRange + 
				target.TotalCloud + target.TotalCred
			if totalFindings < minCount {
				return false
			}
		}
	}

	// Check maximum business value
	if maxBizValue, exists := rule.Conditions["max_business_value"]; exists {
		if maxValue, ok := maxBizValue.(float64); ok {
			businessValue := tdc.calculateBusinessValueFromRule(target, rule)
			if businessValue > maxValue {
				return false
			}
		}
	}

	return true
}

// calculateBusinessValueFromRule calculates business value based on rule context
func (tdc *TrainingDataCollector) calculateBusinessValueFromRule(target *database.Target, rule LabelingRule) float64 {
	// Use classifier's business value calculation as base
	baseValue := tdc.classifier.calculateBusinessValue(rule.AssetType, target)
	
	// Adjust based on rule confidence
	adjustedValue := baseValue * rule.Confidence
	
	// Ensure within valid range
	if adjustedValue > 1.0 {
		adjustedValue = 1.0
	} else if adjustedValue < 0.0 {
		adjustedValue = 0.0
	}
	
	return adjustedValue
}

// determineCriticalityFromType determines criticality level from asset type
func (tdc *TrainingDataCollector) determineCriticalityFromType(assetType string) string {
	criticalityMap := map[string]string{
		"production":  "high",
		"admin":       "critical",
		"api":         "high",
		"staging":     "medium",
		"development": "low",
		"test":        "low",
	}
	
	if criticality, exists := criticalityMap[assetType]; exists {
		return criticality
	}
	return "medium"
}

// createValidationSplit splits the dataset into training and validation sets
func (tdc *TrainingDataCollector) createValidationSplit(dataset *TrainingDataset) {
	if len(dataset.LabeledAssets) == 0 {
		return
	}

	// Calculate validation set size
	validationSize := int(float64(len(dataset.LabeledAssets)) * tdc.config.ValidationRatio)
	if validationSize == 0 {
		validationSize = 1
	}

	// Randomly select validation samples (simple approach)
	validationIndices := make(map[int]bool)
	totalAssets := len(dataset.LabeledAssets)
	
	for len(validationIndices) < validationSize && len(validationIndices) < totalAssets {
		index := int(time.Now().UnixNano()) % totalAssets
		validationIndices[index] = true
	}

	// Create validation set
	dataset.ValidationSet = make([]LabeledAsset, 0, validationSize)
	for index := range validationIndices {
		dataset.ValidationSet = append(dataset.ValidationSet, dataset.LabeledAssets[index])
	}
}

// calculateDatasetStatistics calculates comprehensive statistics for the dataset
func (tdc *TrainingDataCollector) calculateDatasetStatistics(dataset *TrainingDataset) {
	stats := DatasetStatistics{
		TotalSamples:            len(dataset.LabeledAssets),
		ValidationSamples:       len(dataset.ValidationSet),
		TypeDistribution:        make(map[string]int),
		CriticalityDistribution: make(map[string]int),
		FeatureStatistics:       make(map[string]FeatureStats),
		LastUpdated:             time.Now(),
	}

	// Calculate type and criticality distributions
	for _, asset := range dataset.LabeledAssets {
		stats.TypeDistribution[asset.AssetType]++
		stats.CriticalityDistribution[asset.CriticalityLevel]++
	}

	// Calculate feature statistics
	if len(dataset.LabeledAssets) > 0 {
		tdc.calculateFeatureStatistics(dataset.LabeledAssets, &stats)
	}

	// Calculate overall quality score
	stats.QualityScore = tdc.calculateQualityScore(&stats)

	dataset.Statistics = stats
}

// calculateFeatureStatistics calculates statistics for each feature
func (tdc *TrainingDataCollector) calculateFeatureStatistics(assets []LabeledAsset, stats *DatasetStatistics) {
	if len(assets) == 0 {
		return
	}

	// Collect all feature vectors
	featureVectors := make([][]float64, 0)
	for _, asset := range assets {
		if featureVector, exists := asset.Features["feature_vector"]; exists {
			if vec, ok := featureVector.([]float64); ok {
				featureVectors = append(featureVectors, vec)
			}
		}
	}

	if len(featureVectors) == 0 {
		return
	}

	// Calculate statistics for each feature dimension
	numFeatures := len(featureVectors[0])
	for i := 0; i < numFeatures; i++ {
		featureName := fmt.Sprintf("feature_%d", i)
		values := make([]float64, 0)
		nonZeroCount := 0
		
		for _, vec := range featureVectors {
			if i < len(vec) {
				values = append(values, vec[i])
				if vec[i] != 0.0 {
					nonZeroCount++
				}
			}
		}

		if len(values) > 0 {
			featureStats := FeatureStats{
				Mean:    calculateMean(values),
				StdDev:  calculateStdDev(values),
				Min:     calculateMin(values),
				Max:     calculateMax(values),
				NonZero: nonZeroCount,
				Missing: len(featureVectors) - len(values),
			}
			stats.FeatureStatistics[featureName] = featureStats
		}
	}
}

// calculateQualityScore calculates an overall quality score for the dataset
func (tdc *TrainingDataCollector) calculateQualityScore(stats *DatasetStatistics) float64 {
	score := 0.0
	factors := 0

	// Factor 1: Sample size adequacy
	if stats.TotalSamples >= tdc.config.MinSamplesPerType*len(tdc.config.LabelingRules) {
		score += 0.3
	} else {
		score += 0.3 * (float64(stats.TotalSamples) / float64(tdc.config.MinSamplesPerType*len(tdc.config.LabelingRules)))
	}
	factors++

	// Factor 2: Type distribution balance
	if len(stats.TypeDistribution) > 1 {
		balance := calculateDistributionBalance(stats.TypeDistribution)
		score += 0.25 * balance
	}
	factors++

	// Factor 3: Feature completeness
	if len(stats.FeatureStatistics) > 0 {
		completeness := 0.0
		for _, featureStats := range stats.FeatureStatistics {
			if featureStats.Missing == 0 {
				completeness += 1.0
			} else {
				completeness += 1.0 - (float64(featureStats.Missing) / float64(stats.TotalSamples))
			}
		}
		completeness /= float64(len(stats.FeatureStatistics))
		score += 0.25 * completeness
	}
	factors++

	// Factor 4: Validation set adequacy
	if stats.ValidationSamples > 0 {
		validationRatio := float64(stats.ValidationSamples) / float64(stats.TotalSamples)
		if validationRatio >= tdc.config.ValidationRatio {
			score += 0.2
		} else {
			score += 0.2 * (validationRatio / tdc.config.ValidationRatio)
		}
	}
	factors++

	return score
}

// SaveDataset saves a training dataset to disk
func (tdc *TrainingDataCollector) SaveDataset(dataset *TrainingDataset) error {
	filename := fmt.Sprintf("%s.json", dataset.ID)
	filepath := filepath.Join(tdc.dataPath, filename)

	data, err := json.MarshalIndent(dataset, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal dataset: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write dataset file: %w", err)
	}

	utils.InforF("Training dataset saved to: %s", filepath)
	return nil
}

// LoadDataset loads a training dataset from disk
func (tdc *TrainingDataCollector) LoadDataset(datasetID string) (*TrainingDataset, error) {
	filename := fmt.Sprintf("%s.json", datasetID)
	filepath := filepath.Join(tdc.dataPath, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read dataset file: %w", err)
	}

	var dataset TrainingDataset
	if err := json.Unmarshal(data, &dataset); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dataset: %w", err)
	}

	return &dataset, nil
}

// ListDatasets lists all available training datasets
func (tdc *TrainingDataCollector) ListDatasets() ([]string, error) {
	files, err := os.ReadDir(tdc.dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	datasets := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			datasetID := strings.TrimSuffix(file.Name(), ".json")
			datasets = append(datasets, datasetID)
		}
	}

	return datasets, nil
}

// MergeDatasets merges multiple datasets into a single dataset
func (tdc *TrainingDataCollector) MergeDatasets(datasetIDs []string, newName string) (*TrainingDataset, error) {
	if len(datasetIDs) == 0 {
		return nil, fmt.Errorf("no datasets to merge")
	}

	mergedDataset := &TrainingDataset{
		ID:            fmt.Sprintf("merged_%d", time.Now().Unix()),
		Name:          newName,
		Description:   fmt.Sprintf("Merged dataset from %d sources", len(datasetIDs)),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Version:       "1.0",
		LabeledAssets: make([]LabeledAsset, 0),
		ValidationSet: make([]LabeledAsset, 0),
		Metadata: map[string]interface{}{
			"source_datasets": datasetIDs,
			"merge_method":    "concatenation",
		},
	}

	// Load and merge all datasets
	for _, datasetID := range datasetIDs {
		dataset, err := tdc.LoadDataset(datasetID)
		if err != nil {
			utils.ErrorF("Failed to load dataset %s: %v", datasetID, err)
			continue
		}

		mergedDataset.LabeledAssets = append(mergedDataset.LabeledAssets, dataset.LabeledAssets...)
		mergedDataset.ValidationSet = append(mergedDataset.ValidationSet, dataset.ValidationSet...)
	}

	// Recalculate statistics
	tdc.calculateDatasetStatistics(mergedDataset)

	return mergedDataset, nil
}

// Helper functions for statistical calculations
func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64) float64 {
	if len(values) <= 1 {
		return 0.0
	}
	mean := calculateMean(values)
	sumSquares := 0.0
	for _, v := range values {
		sumSquares += (v - mean) * (v - mean)
	}
	return math.Sqrt(sumSquares / float64(len(values)-1))
}

func calculateMin(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func calculateDistributionBalance(distribution map[string]int) float64 {
	if len(distribution) <= 1 {
		return 1.0
	}

	values := make([]int, 0, len(distribution))
	total := 0
	for _, count := range distribution {
		values = append(values, count)
		total += count
	}

	if total == 0 {
		return 0.0
	}

	// Calculate entropy as a measure of balance
	entropy := 0.0
	for _, count := range values {
		if count > 0 {
			p := float64(count) / float64(total)
			entropy -= p * math.Log2(p)
		}
	}

	// Normalize by maximum possible entropy
	maxEntropy := math.Log2(float64(len(values)))
	if maxEntropy == 0 {
		return 1.0
	}

	return entropy / maxEntropy
}
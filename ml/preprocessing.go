package ml

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/dablon/osmedeus/database"
	"github.com/dablon/osmedeus/utils"
)

// PreprocessingPipeline manages data preprocessing for ML models
type PreprocessingPipeline struct {
	processors map[string]DataPreprocessor
	config     *PreprocessingConfig
}

// PreprocessingConfig holds configuration for data preprocessing
type PreprocessingConfig struct {
	FeatureScaling    string                 `json:"feature_scaling"`    // standard, minmax, robust
	HandleMissing     string                 `json:"handle_missing"`     // drop, mean, median, mode
	TextProcessing    string                 `json:"text_processing"`    // tfidf, word2vec, bert
	TimeSeriesWindow  int                    `json:"timeseries_window"`  // window size for time series features
	OutlierDetection  bool                   `json:"outlier_detection"`
	FeatureSelection  bool                   `json:"feature_selection"`
	CustomProcessors  map[string]interface{} `json:"custom_processors"`
}

// FeatureVector represents a processed feature vector
type FeatureVector struct {
	Features    []float64              `json:"features"`
	Labels      []string               `json:"labels"`
	Metadata    map[string]interface{} `json:"metadata"`
	ProcessedAt time.Time              `json:"processed_at"`
}

// NewPreprocessingPipeline creates a new preprocessing pipeline
func NewPreprocessingPipeline(config *PreprocessingConfig) *PreprocessingPipeline {
	if config == nil {
		config = &PreprocessingConfig{
			FeatureScaling:   "standard",
			HandleMissing:    "mean",
			TextProcessing:   "tfidf",
			TimeSeriesWindow: 10,
			OutlierDetection: true,
			FeatureSelection: true,
			CustomProcessors: make(map[string]interface{}),
		}
	}

	pp := &PreprocessingPipeline{
		processors: make(map[string]DataPreprocessor),
		config:     config,
	}

	// Register default preprocessors
	pp.registerDefaultProcessors()

	return pp
}

// registerDefaultProcessors registers the default data preprocessors
func (pp *PreprocessingPipeline) registerDefaultProcessors() {
	pp.processors["asset"] = NewAssetDataPreprocessor()
	pp.processors["behavior"] = NewBehaviorDataPreprocessor()
	pp.processors["vulnerability"] = NewVulnerabilityDataPreprocessor()
	pp.processors["network"] = NewNetworkDataPreprocessor()
	pp.processors["text"] = NewTextDataPreprocessor()
	pp.processors["timeseries"] = NewTimeSeriesPreprocessor()
}

// ProcessAssetData preprocesses asset data for ML models
func (pp *PreprocessingPipeline) ProcessAssetData(target *database.Target) (*FeatureVector, error) {
	processor, exists := pp.processors["asset"]
	if !exists {
		return nil, fmt.Errorf("asset preprocessor not found")
	}

	// Convert target to map for processing
	assetData := map[string]interface{}{
		"input_name":           target.InputName,
		"input_type":           target.InputType,
		"total_assets":         target.TotalAssets,
		"total_dns":            target.TotalDns,
		"total_tech":           target.TotalTech,
		"total_screenshot":     target.TotalScreenShot,
		"total_vulnerability":  target.TotalVulnerability,
		"total_dirb":           target.TotalDirb,
		"total_link":           target.TotalLink,
		"total_archive":        target.TotalArchive,
		"total_ip_range":       target.TotalIPRange,
		"total_cloud":          target.TotalCloud,
		"total_cred":           target.TotalCred,
		"is_wildcard":          target.IsWildCard,
	}

	processed, err := processor.Preprocess(assetData)
	if err != nil {
		return nil, fmt.Errorf("failed to preprocess asset data: %w", err)
	}

	features, err := processor.GetFeatures(processed)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Apply scaling
	scaledFeatures, err := pp.scaleFeatures(features)
	if err != nil {
		return nil, fmt.Errorf("failed to scale features: %w", err)
	}

	return &FeatureVector{
		Features: scaledFeatures,
		Labels: []string{
			"total_assets", "total_dns", "total_tech", "total_screenshot",
			"total_vulnerability", "total_dirb", "total_link", "total_archive",
			"total_ip_range", "total_cloud", "total_cred", "is_wildcard",
		},
		Metadata: map[string]interface{}{
			"target_id":      target.InputName,
			"preprocessing":  "asset_classification",
			"feature_count":  len(scaledFeatures),
		},
		ProcessedAt: time.Now(),
	}, nil
}

// ProcessBehaviorData preprocesses behavioral analysis data
func (pp *PreprocessingPipeline) ProcessBehaviorData(behaviorData map[string]interface{}) (*FeatureVector, error) {
	processor, exists := pp.processors["behavior"]
	if !exists {
		return nil, fmt.Errorf("behavior preprocessor not found")
	}

	processed, err := processor.Preprocess(behaviorData)
	if err != nil {
		return nil, fmt.Errorf("failed to preprocess behavior data: %w", err)
	}

	features, err := processor.GetFeatures(processed)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Apply time series processing if needed
	if pp.config.TimeSeriesWindow > 0 {
		features = pp.applyTimeSeriesWindow(features, pp.config.TimeSeriesWindow)
	}

	scaledFeatures, err := pp.scaleFeatures(features)
	if err != nil {
		return nil, fmt.Errorf("failed to scale features: %w", err)
	}

	return &FeatureVector{
		Features: scaledFeatures,
		Labels: []string{
			"avg_response_time", "status_code_variety", "error_rate",
			"request_pattern_entropy", "timing_variance",
		},
		Metadata: map[string]interface{}{
			"preprocessing":  "behavioral_analysis",
			"feature_count":  len(scaledFeatures),
			"window_size":    pp.config.TimeSeriesWindow,
		},
		ProcessedAt: time.Now(),
	}, nil
}

// ProcessVulnerabilityData preprocesses vulnerability prediction data
func (pp *PreprocessingPipeline) ProcessVulnerabilityData(vulnData map[string]interface{}) (*FeatureVector, error) {
	processor, exists := pp.processors["vulnerability"]
	if !exists {
		return nil, fmt.Errorf("vulnerability preprocessor not found")
	}

	processed, err := processor.Preprocess(vulnData)
	if err != nil {
		return nil, fmt.Errorf("failed to preprocess vulnerability data: %w", err)
	}

	features, err := processor.GetFeatures(processed)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Apply outlier detection if enabled
	if pp.config.OutlierDetection {
		features = pp.removeOutliers(features)
	}

	scaledFeatures, err := pp.scaleFeatures(features)
	if err != nil {
		return nil, fmt.Errorf("failed to scale features: %w", err)
	}

	return &FeatureVector{
		Features: scaledFeatures,
		Labels: []string{
			"historical_vuln_count", "avg_severity", "patch_lag",
			"technology_risk_score", "exposure_score",
		},
		Metadata: map[string]interface{}{
			"preprocessing":     "vulnerability_prediction",
			"feature_count":     len(scaledFeatures),
			"outlier_detection": pp.config.OutlierDetection,
		},
		ProcessedAt: time.Now(),
	}, nil
}

// scaleFeatures applies feature scaling based on configuration
func (pp *PreprocessingPipeline) scaleFeatures(features []float64) ([]float64, error) {
	if len(features) == 0 {
		return features, nil
	}

	switch pp.config.FeatureScaling {
	case "standard":
		return pp.standardScale(features), nil
	case "minmax":
		return pp.minMaxScale(features), nil
	case "robust":
		return pp.robustScale(features), nil
	default:
		return features, nil // No scaling
	}
}

// standardScale applies standard (z-score) scaling
func (pp *PreprocessingPipeline) standardScale(features []float64) []float64 {
	if len(features) <= 1 {
		return features
	}

	// Calculate mean
	sum := 0.0
	for _, f := range features {
		sum += f
	}
	mean := sum / float64(len(features))

	// Calculate standard deviation
	sumSquares := 0.0
	for _, f := range features {
		sumSquares += (f - mean) * (f - mean)
	}
	stdDev := math.Sqrt(sumSquares / float64(len(features)-1))

	if stdDev == 0 {
		return features
	}

	// Apply scaling
	scaled := make([]float64, len(features))
	for i, f := range features {
		scaled[i] = (f - mean) / stdDev
	}

	return scaled
}

// minMaxScale applies min-max scaling
func (pp *PreprocessingPipeline) minMaxScale(features []float64) []float64 {
	if len(features) == 0 {
		return features
	}

	min, max := features[0], features[0]
	for _, f := range features {
		if f < min {
			min = f
		}
		if f > max {
			max = f
		}
	}

	if max == min {
		return features
	}

	scaled := make([]float64, len(features))
	for i, f := range features {
		scaled[i] = (f - min) / (max - min)
	}

	return scaled
}

// robustScale applies robust scaling using median and IQR
func (pp *PreprocessingPipeline) robustScale(features []float64) []float64 {
	if len(features) <= 1 {
		return features
	}

	// Sort features to calculate percentiles
	sorted := make([]float64, len(features))
	copy(sorted, features)
	sort.Float64s(sorted)

	// Calculate median (Q2)
	median := pp.calculatePercentile(sorted, 0.5)

	// Calculate Q1 and Q3
	q1 := pp.calculatePercentile(sorted, 0.25)
	q3 := pp.calculatePercentile(sorted, 0.75)
	iqr := q3 - q1

	if iqr == 0 {
		return features
	}

	// Apply robust scaling
	scaled := make([]float64, len(features))
	for i, f := range features {
		scaled[i] = (f - median) / iqr
	}

	return scaled
}

// calculatePercentile calculates the percentile of a sorted array
func (pp *PreprocessingPipeline) calculatePercentile(sorted []float64, percentile float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	index := percentile * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// removeOutliers removes outliers using IQR method
func (pp *PreprocessingPipeline) removeOutliers(features []float64) []float64 {
	if len(features) <= 4 {
		return features
	}

	sorted := make([]float64, len(features))
	copy(sorted, features)
	sort.Float64s(sorted)

	q1 := pp.calculatePercentile(sorted, 0.25)
	q3 := pp.calculatePercentile(sorted, 0.75)
	iqr := q3 - q1

	lowerBound := q1 - 1.5*iqr
	upperBound := q3 + 1.5*iqr

	filtered := make([]float64, 0)
	for _, f := range features {
		if f >= lowerBound && f <= upperBound {
			filtered = append(filtered, f)
		}
	}

	return filtered
}

// applyTimeSeriesWindow applies a sliding window to time series data
func (pp *PreprocessingPipeline) applyTimeSeriesWindow(features []float64, windowSize int) []float64 {
	if len(features) < windowSize {
		return features
	}

	windowed := make([]float64, 0)
	for i := 0; i <= len(features)-windowSize; i++ {
		window := features[i : i+windowSize]
		
		// Calculate window statistics
		sum := 0.0
		for _, f := range window {
			sum += f
		}
		mean := sum / float64(len(window))
		
		windowed = append(windowed, mean)
	}

	return windowed
}

// Additional specialized preprocessors

// NetworkDataPreprocessor for network-related data
type NetworkDataPreprocessor struct {
	config map[string]interface{}
}

func NewNetworkDataPreprocessor() DataPreprocessor {
	return &NetworkDataPreprocessor{
		config: map[string]interface{}{
			"port_encoding":    "categorical",
			"protocol_mapping": true,
		},
	}
}

func (ndp *NetworkDataPreprocessor) Preprocess(rawData interface{}) (interface{}, error) {
	dataMap, ok := rawData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format for network preprocessing")
	}

	processed := map[string]interface{}{
		"features":     ndp.extractNetworkFeatures(dataMap),
		"metadata":     dataMap,
		"processed_at": time.Now(),
	}

	return processed, nil
}

func (ndp *NetworkDataPreprocessor) GetFeatures(data interface{}) ([]float64, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	return ndp.extractNetworkFeatures(dataMap), nil
}

func (ndp *NetworkDataPreprocessor) Normalize(features []float64) ([]float64, error) {
	return features, nil
}

func (ndp *NetworkDataPreprocessor) Transform(data interface{}) (interface{}, error) {
	return ndp.Preprocess(data)
}

func (ndp *NetworkDataPreprocessor) extractNetworkFeatures(data map[string]interface{}) []float64 {
	features := make([]float64, 0)

	// Extract port-related features
	if ports, ok := data["open_ports"].([]interface{}); ok {
		features = append(features, float64(len(ports)))
		
		// Count common port categories
		webPorts := 0
		sshPorts := 0
		dbPorts := 0
		
		for _, port := range ports {
			if portNum, ok := port.(float64); ok {
				switch int(portNum) {
				case 80, 443, 8080, 8443:
					webPorts++
				case 22:
					sshPorts++
				case 3306, 5432, 1433, 27017:
					dbPorts++
				}
			}
		}
		
		features = append(features, float64(webPorts), float64(sshPorts), float64(dbPorts))
	}

	// Extract protocol features
	if protocols, ok := data["protocols"].([]interface{}); ok {
		features = append(features, float64(len(protocols)))
	}

	return features
}

// TextDataPreprocessor for text-based data
type TextDataPreprocessor struct {
	config map[string]interface{}
}

func NewTextDataPreprocessor() DataPreprocessor {
	return &TextDataPreprocessor{
		config: map[string]interface{}{
			"lowercase":     true,
			"remove_punct":  true,
			"min_word_len":  2,
		},
	}
}

func (tdp *TextDataPreprocessor) Preprocess(rawData interface{}) (interface{}, error) {
	dataMap, ok := rawData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format for text preprocessing")
	}

	processed := map[string]interface{}{
		"features":     tdp.extractTextFeatures(dataMap),
		"metadata":     dataMap,
		"processed_at": time.Now(),
	}

	return processed, nil
}

func (tdp *TextDataPreprocessor) GetFeatures(data interface{}) ([]float64, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	return tdp.extractTextFeatures(dataMap), nil
}

func (tdp *TextDataPreprocessor) Normalize(features []float64) ([]float64, error) {
	return features, nil
}

func (tdp *TextDataPreprocessor) Transform(data interface{}) (interface{}, error) {
	return tdp.Preprocess(data)
}

func (tdp *TextDataPreprocessor) extractTextFeatures(data map[string]interface{}) []float64 {
	features := make([]float64, 0)

	// Extract text-based features
	if text, ok := data["text"].(string); ok {
		// Basic text statistics
		features = append(features, float64(len(text)))
		features = append(features, float64(len(strings.Fields(text))))
		
		// Character frequency analysis
		charCount := make(map[rune]int)
		for _, char := range text {
			charCount[char]++
		}
		features = append(features, float64(len(charCount))) // unique characters
	}

	return features
}

// TimeSeriesPreprocessor for time series data
type TimeSeriesPreprocessor struct {
	config map[string]interface{}
}

func NewTimeSeriesPreprocessor() DataPreprocessor {
	return &TimeSeriesPreprocessor{
		config: map[string]interface{}{
			"window_size":    10,
			"overlap":        0.5,
			"detrend":        true,
		},
	}
}

func (tsp *TimeSeriesPreprocessor) Preprocess(rawData interface{}) (interface{}, error) {
	dataMap, ok := rawData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format for time series preprocessing")
	}

	processed := map[string]interface{}{
		"features":     tsp.extractTimeSeriesFeatures(dataMap),
		"metadata":     dataMap,
		"processed_at": time.Now(),
	}

	return processed, nil
}

func (tsp *TimeSeriesPreprocessor) GetFeatures(data interface{}) ([]float64, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	return tsp.extractTimeSeriesFeatures(dataMap), nil
}

func (tsp *TimeSeriesPreprocessor) Normalize(features []float64) ([]float64, error) {
	return features, nil
}

func (tsp *TimeSeriesPreprocessor) Transform(data interface{}) (interface{}, error) {
	return tsp.Preprocess(data)
}

func (tsp *TimeSeriesPreprocessor) extractTimeSeriesFeatures(data map[string]interface{}) []float64 {
	features := make([]float64, 0)

	if timeSeries, ok := data["time_series"].([]interface{}); ok {
		values := make([]float64, 0)
		for _, v := range timeSeries {
			if val, ok := v.(float64); ok {
				values = append(values, val)
			}
		}

		if len(values) > 0 {
			// Statistical features
			sum := 0.0
			for _, v := range values {
				sum += v
			}
			mean := sum / float64(len(values))
			features = append(features, mean)

			// Variance
			variance := 0.0
			for _, v := range values {
				variance += (v - mean) * (v - mean)
			}
			variance /= float64(len(values))
			features = append(features, variance)

			// Min/Max
			min, max := values[0], values[0]
			for _, v := range values {
				if v < min {
					min = v
				}
				if v > max {
					max = v
				}
			}
			features = append(features, min, max)
		}
	}

	return features
}

// BatchProcessor handles batch processing of multiple data items
type BatchProcessor struct {
	pipeline *PreprocessingPipeline
	batchSize int
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(pipeline *PreprocessingPipeline, batchSize int) *BatchProcessor {
	return &BatchProcessor{
		pipeline:  pipeline,
		batchSize: batchSize,
	}
}

// ProcessBatch processes a batch of targets
func (bp *BatchProcessor) ProcessBatch(targets []*database.Target) ([]*FeatureVector, error) {
	results := make([]*FeatureVector, 0, len(targets))
	
	for i := 0; i < len(targets); i += bp.batchSize {
		end := i + bp.batchSize
		if end > len(targets) {
			end = len(targets)
		}
		
		batch := targets[i:end]
		batchResults, err := bp.processBatchChunk(batch)
		if err != nil {
			utils.ErrorF("Failed to process batch chunk: %v", err)
			continue
		}
		
		results = append(results, batchResults...)
	}
	
	return results, nil
}

// processBatchChunk processes a single batch chunk
func (bp *BatchProcessor) processBatchChunk(targets []*database.Target) ([]*FeatureVector, error) {
	results := make([]*FeatureVector, 0, len(targets))
	
	for _, target := range targets {
		featureVector, err := bp.pipeline.ProcessAssetData(target)
		if err != nil {
			utils.ErrorF("Failed to process target %s: %v", target.InputName, err)
			continue
		}
		
		results = append(results, featureVector)
	}
	
	return results, nil
}
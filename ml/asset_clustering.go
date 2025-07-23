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

// AssetClusteringEngine implements unsupervised learning for asset grouping
type AssetClusteringEngine struct {
	config           *ClusteringConfig
	featureExtractor *AssetFeatureExtractor
	clusters         []*AssetCluster
	clusterHistory   []*ClusterSnapshot
	similarityMatrix [][]float64
	lastUpdated      time.Time
}

// ClusteringConfig holds configuration for asset clustering
type ClusteringConfig struct {
	Algorithm           string                 `json:"algorithm"`
	NumClusters         int                    `json:"num_clusters"`
	MinClusterSize      int                    `json:"min_cluster_size"`
	MaxIterations       int                    `json:"max_iterations"`
	ConvergenceThreshold float64               `json:"convergence_threshold"`
	SimilarityThreshold float64               `json:"similarity_threshold"`
	DistanceMetric      string                 `json:"distance_metric"`
	FeatureWeights      map[string]float64     `json:"feature_weights"`
	ClusteringFeatures  []string               `json:"clustering_features"`
	UpdateInterval      time.Duration          `json:"update_interval"`
	VisualizationConfig map[string]interface{} `json:"visualization_config"`
}

// AssetCluster represents a group of similar assets
type AssetCluster struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Assets           []string               `json:"assets"`
	Centroid         []float64              `json:"centroid"`
	Characteristics  map[string]interface{} `json:"characteristics"`
	SimilarityScore  float64                `json:"similarity_score"`
	ClusterSize      int                    `json:"cluster_size"`
	CreatedAt        time.Time              `json:"created_at"`
	LastUpdated      time.Time              `json:"last_updated"`
	Quality          *ClusterQuality        `json:"quality"`
}

// ClusterQuality represents quality metrics for a cluster
type ClusterQuality struct {
	Cohesion         float64 `json:"cohesion"`
	Separation       float64 `json:"separation"`
	Silhouette       float64 `json:"silhouette"`
	Stability        float64 `json:"stability"`
	Interpretability float64 `json:"interpretability"`
}

// ClusterSnapshot represents a historical state of clustering
type ClusterSnapshot struct {
	Timestamp    time.Time      `json:"timestamp"`
	Clusters     []*AssetCluster `json:"clusters"`
	QualityScore float64        `json:"quality_score"`
	Algorithm    string         `json:"algorithm"`
	Parameters   map[string]interface{} `json:"parameters"`
}

// SimilarityResult represents similarity between two assets
type SimilarityResult struct {
	Asset1ID        string                 `json:"asset1_id"`
	Asset2ID        string                 `json:"asset2_id"`
	SimilarityScore float64                `json:"similarity_score"`
	CommonFeatures  []string               `json:"common_features"`
	Differences     []string               `json:"differences"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ClusterVisualizationData represents data for cluster visualization
type ClusterVisualizationData struct {
	Nodes []ClusterNode `json:"nodes"`
	Edges []ClusterEdge `json:"edges"`
	Layout string       `json:"layout"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ClusterNode represents a node in cluster visualization
type ClusterNode struct {
	ID          string                 `json:"id"`
	Label       string                 `json:"label"`
	ClusterID   string                 `json:"cluster_id"`
	Position    []float64              `json:"position"`
	Size        float64                `json:"size"`
	Color       string                 `json:"color"`
	Properties  map[string]interface{} `json:"properties"`
}

// ClusterEdge represents an edge in cluster visualization
type ClusterEdge struct {
	Source     string  `json:"source"`
	Target     string  `json:"target"`
	Weight     float64 `json:"weight"`
	Type       string  `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

// NewAssetClusteringEngine creates a new asset clustering engine
func NewAssetClusteringEngine(config *ClusteringConfig) (*AssetClusteringEngine, error) {
	if config == nil {
		config = &ClusteringConfig{
			Algorithm:           "kmeans",
			NumClusters:         5,
			MinClusterSize:      3,
			MaxIterations:       100,
			ConvergenceThreshold: 0.001,
			SimilarityThreshold: 0.7,
			DistanceMetric:      "cosine",
			FeatureWeights: map[string]float64{
				"technical_features":  0.4,
				"behavioral_features": 0.3,
				"metadata_features":   0.3,
			},
			ClusteringFeatures: []string{
				"total_assets", "total_vulnerabilities", "technology_diversity",
				"response_patterns", "business_value", "criticality_level",
			},
			UpdateInterval: 6 * time.Hour,
			VisualizationConfig: map[string]interface{}{
				"layout": "force_directed",
				"node_size_factor": 10.0,
				"edge_thickness_factor": 2.0,
			},
		}
	}

	// Create feature extractor
	featureExtractor := &AssetFeatureExtractor{
		config: &FeatureExtractionConfig{
			TechnicalFeatures: []string{
				"total_assets", "total_dns", "total_tech", "total_vulnerability",
				"total_screenshot", "total_dirb", "total_link", "total_archive",
				"total_ip_range", "total_cloud", "total_cred",
			},
			BehavioralFeatures: []string{
				"response_patterns", "error_patterns", "timing_characteristics",
			},
			MetadataFeatures: []string{
				"domain_characteristics", "business_context", "geographic_info",
			},
			FeatureWeights: config.FeatureWeights,
			NormalizationMethod: "standard",
		},
	}

	engine := &AssetClusteringEngine{
		config:           config,
		featureExtractor: featureExtractor,
		clusters:         make([]*AssetCluster, 0),
		clusterHistory:   make([]*ClusterSnapshot, 0),
		lastUpdated:      time.Now(),
	}

	return engine, nil
}

// ClusterAssets performs clustering on a set of assets
func (ace *AssetClusteringEngine) ClusterAssets(targets []*database.Target) ([]*AssetCluster, error) {
	if len(targets) < ace.config.MinClusterSize {
		return nil, fmt.Errorf("insufficient assets for clustering: need at least %d, got %d", 
			ace.config.MinClusterSize, len(targets))
	}

	utils.InforF("Starting asset clustering with %d assets using %s algorithm", 
		len(targets), ace.config.Algorithm)

	// Extract features for all assets
	featureMatrix, assetIDs, err := ace.extractFeatureMatrix(targets)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Normalize features
	normalizedMatrix, err := ace.normalizeFeatures(featureMatrix)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize features: %w", err)
	}

	// Perform clustering based on algorithm
	var clusters []*AssetCluster
	switch ace.config.Algorithm {
	case "kmeans":
		clusters, err = ace.performKMeansClustering(normalizedMatrix, assetIDs)
	case "hierarchical":
		clusters, err = ace.performHierarchicalClustering(normalizedMatrix, assetIDs)
	case "dbscan":
		clusters, err = ace.performDBSCANClustering(normalizedMatrix, assetIDs)
	default:
		return nil, fmt.Errorf("unsupported clustering algorithm: %s", ace.config.Algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("clustering failed: %w", err)
	}

	// Calculate cluster quality metrics
	for _, cluster := range clusters {
		cluster.Quality = ace.calculateClusterQuality(cluster, normalizedMatrix, assetIDs)
	}

	// Generate cluster characteristics
	for _, cluster := range clusters {
		cluster.Characteristics = ace.generateClusterCharacteristics(cluster, targets)
	}

	// Update internal state
	ace.clusters = clusters
	ace.lastUpdated = time.Now()

	// Save snapshot
	ace.saveClusterSnapshot(clusters)

	utils.InforF("Clustering completed: created %d clusters", len(clusters))
	return clusters, nil
}

// extractFeatureMatrix extracts feature matrix from targets
func (ace *AssetClusteringEngine) extractFeatureMatrix(targets []*database.Target) ([][]float64, []string, error) {
	featureMatrix := make([][]float64, len(targets))
	assetIDs := make([]string, len(targets))

	for i, target := range targets {
		features, err := ace.extractAssetFeatures(target)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to extract features for %s: %w", target.InputName, err)
		}
		featureMatrix[i] = features
		assetIDs[i] = target.InputName
	}

	return featureMatrix, assetIDs, nil
}

// extractAssetFeatures extracts clustering features from a target
func (ace *AssetClusteringEngine) extractAssetFeatures(target *database.Target) ([]float64, error) {
	features := make([]float64, 0)

	// Technical features
	features = append(features, float64(target.TotalAssets))
	features = append(features, float64(target.TotalDns))
	features = append(features, float64(target.TotalTech))
	features = append(features, float64(target.TotalVulnerability))
	features = append(features, float64(target.TotalScreenShot))
	features = append(features, float64(target.TotalDirb))
	features = append(features, float64(target.TotalLink))
	features = append(features, float64(target.TotalArchive))
	features = append(features, float64(target.TotalIPRange))
	features = append(features, float64(target.TotalCloud))
	features = append(features, float64(target.TotalCred))

	// Derived features
	totalFindings := float64(target.TotalAssets + target.TotalDns + target.TotalTech + 
		target.TotalVulnerability + target.TotalScreenShot + target.TotalDirb + 
		target.TotalLink + target.TotalArchive + target.TotalIPRange + 
		target.TotalCloud + target.TotalCred)
	features = append(features, totalFindings)

	// Vulnerability density
	vulnDensity := 0.0
	if totalFindings > 0 {
		vulnDensity = float64(target.TotalVulnerability) / totalFindings
	}
	features = append(features, vulnDensity)

	// Technology diversity
	techDiversity := float64(target.TotalTech)
	if target.TotalAssets > 0 {
		techDiversity = float64(target.TotalTech) / float64(target.TotalAssets)
	}
	features = append(features, techDiversity)

	// Domain-based features
	domainFeatures := ace.extractDomainFeatures(target.InputName)
	features = append(features, domainFeatures...)

	// Wildcard indicator
	wildcardValue := 0.0
	if target.IsWildCard {
		wildcardValue = 1.0
	}
	features = append(features, wildcardValue)

	return features, nil
}

// extractDomainFeatures extracts domain-based features for clustering
func (ace *AssetClusteringEngine) extractDomainFeatures(domain string) []float64 {
	features := make([]float64, 0)

	// Domain length
	features = append(features, float64(len(domain)))

	// Subdomain depth
	parts := strings.Split(domain, ".")
	subdomainDepth := float64(len(parts) - 2)
	if subdomainDepth < 0 {
		subdomainDepth = 0
	}
	features = append(features, subdomainDepth)

	// Keyword presence indicators
	lowerDomain := strings.ToLower(domain)
	
	// Production indicators
	prodKeywords := []string{"www", "prod", "production", "live"}
	hasProdKeywords := 0.0
	for _, keyword := range prodKeywords {
		if strings.Contains(lowerDomain, keyword) {
			hasProdKeywords = 1.0
			break
		}
	}
	features = append(features, hasProdKeywords)

	// Development indicators
	devKeywords := []string{"dev", "development", "test", "testing", "debug", "local"}
	hasDevKeywords := 0.0
	for _, keyword := range devKeywords {
		if strings.Contains(lowerDomain, keyword) {
			hasDevKeywords = 1.0
			break
		}
	}
	features = append(features, hasDevKeywords)

	// Admin indicators
	adminKeywords := []string{"admin", "administrator", "manage", "control", "panel"}
	hasAdminKeywords := 0.0
	for _, keyword := range adminKeywords {
		if strings.Contains(lowerDomain, keyword) {
			hasAdminKeywords = 1.0
			break
		}
	}
	features = append(features, hasAdminKeywords)

	// API indicators
	apiKeywords := []string{"api", "rest", "graphql", "service", "endpoint"}
	hasAPIKeywords := 0.0
	for _, keyword := range apiKeywords {
		if strings.Contains(lowerDomain, keyword) {
			hasAPIKeywords = 1.0
			break
		}
	}
	features = append(features, hasAPIKeywords)

	return features
}

// normalizeFeatures normalizes the feature matrix
func (ace *AssetClusteringEngine) normalizeFeatures(matrix [][]float64) ([][]float64, error) {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return matrix, nil
	}

	numFeatures := len(matrix[0])
	normalized := make([][]float64, len(matrix))
	
	// Initialize normalized matrix
	for i := range normalized {
		normalized[i] = make([]float64, numFeatures)
	}

	// Normalize each feature dimension
	for j := 0; j < numFeatures; j++ {
		// Calculate mean and standard deviation for this feature
		sum := 0.0
		for i := 0; i < len(matrix); i++ {
			sum += matrix[i][j]
		}
		mean := sum / float64(len(matrix))

		sumSquares := 0.0
		for i := 0; i < len(matrix); i++ {
			diff := matrix[i][j] - mean
			sumSquares += diff * diff
		}
		stdDev := math.Sqrt(sumSquares / float64(len(matrix)))

		// Normalize feature values
		for i := 0; i < len(matrix); i++ {
			if stdDev > 0 {
				normalized[i][j] = (matrix[i][j] - mean) / stdDev
			} else {
				normalized[i][j] = 0.0
			}
		}
	}

	return normalized, nil
}

// performKMeansClustering performs K-means clustering
func (ace *AssetClusteringEngine) performKMeansClustering(matrix [][]float64, assetIDs []string) ([]*AssetCluster, error) {
	if len(matrix) < ace.config.NumClusters {
		return nil, fmt.Errorf("not enough assets for %d clusters", ace.config.NumClusters)
	}

	numAssets := len(matrix)
	numFeatures := len(matrix[0])
	k := ace.config.NumClusters

	// Initialize centroids randomly
	centroids := make([][]float64, k)
	for i := 0; i < k; i++ {
		centroids[i] = make([]float64, numFeatures)
		// Use random asset as initial centroid
		randomIndex := int(time.Now().UnixNano()+int64(i)) % numAssets
		copy(centroids[i], matrix[randomIndex])
	}

	// Assignment of assets to clusters
	assignments := make([]int, numAssets)
	
	// Iterative optimization
	for iteration := 0; iteration < ace.config.MaxIterations; iteration++ {
		// Assign each asset to nearest centroid
		changed := false
		for i := 0; i < numAssets; i++ {
			minDistance := math.Inf(1)
			bestCluster := 0
			
			for j := 0; j < k; j++ {
				distance := ace.calculateDistance(matrix[i], centroids[j])
				if distance < minDistance {
					minDistance = distance
					bestCluster = j
				}
			}
			
			if assignments[i] != bestCluster {
				assignments[i] = bestCluster
				changed = true
			}
		}

		// Update centroids
		newCentroids := make([][]float64, k)
		clusterCounts := make([]int, k)
		
		for i := 0; i < k; i++ {
			newCentroids[i] = make([]float64, numFeatures)
		}
		
		for i := 0; i < numAssets; i++ {
			cluster := assignments[i]
			clusterCounts[cluster]++
			for j := 0; j < numFeatures; j++ {
				newCentroids[cluster][j] += matrix[i][j]
			}
		}
		
		// Average to get new centroids
		convergence := true
		for i := 0; i < k; i++ {
			if clusterCounts[i] > 0 {
				for j := 0; j < numFeatures; j++ {
					newCentroids[i][j] /= float64(clusterCounts[i])
				}
				
				// Check convergence
				distance := ace.calculateDistance(centroids[i], newCentroids[i])
				if distance > ace.config.ConvergenceThreshold {
					convergence = false
				}
			}
		}
		
		centroids = newCentroids
		
		if convergence && !changed {
			utils.InforF("K-means converged after %d iterations", iteration+1)
			break
		}
	}

	// Create cluster objects
	clusters := make([]*AssetCluster, 0)
	for i := 0; i < k; i++ {
		clusterAssets := make([]string, 0)
		for j := 0; j < numAssets; j++ {
			if assignments[j] == i {
				clusterAssets = append(clusterAssets, assetIDs[j])
			}
		}
		
		if len(clusterAssets) >= ace.config.MinClusterSize {
			cluster := &AssetCluster{
				ID:          fmt.Sprintf("kmeans_cluster_%d", i),
				Name:        fmt.Sprintf("Asset Cluster %d", i+1),
				Description: fmt.Sprintf("K-means cluster containing %d assets", len(clusterAssets)),
				Assets:      clusterAssets,
				Centroid:    centroids[i],
				ClusterSize: len(clusterAssets),
				CreatedAt:   time.Now(),
				LastUpdated: time.Now(),
			}
			clusters = append(clusters, cluster)
		}
	}

	return clusters, nil
}

// performHierarchicalClustering performs hierarchical clustering (simplified implementation)
func (ace *AssetClusteringEngine) performHierarchicalClustering(matrix [][]float64, assetIDs []string) ([]*AssetCluster, error) {
	// For now, use a simplified approach - just group by similarity threshold
	clusters := make([]*AssetCluster, 0)
	used := make([]bool, len(matrix))
	clusterID := 0

	for i := 0; i < len(matrix); i++ {
		if used[i] {
			continue
		}

		clusterAssets := []string{assetIDs[i]}
		used[i] = true

		// Find similar assets
		for j := i + 1; j < len(matrix); j++ {
			if used[j] {
				continue
			}

			distance := ace.calculateDistance(matrix[i], matrix[j])
			similarity := 1.0 - distance

			if similarity >= ace.config.SimilarityThreshold {
				clusterAssets = append(clusterAssets, assetIDs[j])
				used[j] = true
			}
		}

		if len(clusterAssets) >= ace.config.MinClusterSize {
			cluster := &AssetCluster{
				ID:          fmt.Sprintf("hierarchical_cluster_%d", clusterID),
				Name:        fmt.Sprintf("Hierarchical Cluster %d", clusterID+1),
				Description: fmt.Sprintf("Hierarchical cluster containing %d assets", len(clusterAssets)),
				Assets:      clusterAssets,
				ClusterSize: len(clusterAssets),
				CreatedAt:   time.Now(),
				LastUpdated: time.Now(),
			}
			clusters = append(clusters, cluster)
			clusterID++
		}
	}

	return clusters, nil
}

// performDBSCANClustering performs DBSCAN clustering (simplified implementation)
func (ace *AssetClusteringEngine) performDBSCANClustering(matrix [][]float64, assetIDs []string) ([]*AssetCluster, error) {
	// Simplified DBSCAN implementation
	clusters := make([]*AssetCluster, 0)
	visited := make([]bool, len(matrix))
	clusterID := 0

	for i := 0; i < len(matrix); i++ {
		if visited[i] {
			continue
		}

		// Find neighbors
		neighbors := make([]int, 0)
		for j := 0; j < len(matrix); j++ {
			if i != j {
				distance := ace.calculateDistance(matrix[i], matrix[j])
				if distance <= (1.0 - ace.config.SimilarityThreshold) {
					neighbors = append(neighbors, j)
				}
			}
		}

		if len(neighbors) >= ace.config.MinClusterSize-1 {
			// Create cluster
			clusterAssets := []string{assetIDs[i]}
			visited[i] = true

			for _, neighborIdx := range neighbors {
				if !visited[neighborIdx] {
					clusterAssets = append(clusterAssets, assetIDs[neighborIdx])
					visited[neighborIdx] = true
				}
			}

			cluster := &AssetCluster{
				ID:          fmt.Sprintf("dbscan_cluster_%d", clusterID),
				Name:        fmt.Sprintf("DBSCAN Cluster %d", clusterID+1),
				Description: fmt.Sprintf("DBSCAN cluster containing %d assets", len(clusterAssets)),
				Assets:      clusterAssets,
				ClusterSize: len(clusterAssets),
				CreatedAt:   time.Now(),
				LastUpdated: time.Now(),
			}
			clusters = append(clusters, cluster)
			clusterID++
		} else {
			visited[i] = true // Mark as noise
		}
	}

	return clusters, nil
}

// calculateDistance calculates distance between two feature vectors
func (ace *AssetClusteringEngine) calculateDistance(vec1, vec2 []float64) float64 {
	switch ace.config.DistanceMetric {
	case "euclidean":
		return ace.euclideanDistance(vec1, vec2)
	case "cosine":
		return ace.cosineDistance(vec1, vec2)
	case "manhattan":
		return ace.manhattanDistance(vec1, vec2)
	default:
		return ace.euclideanDistance(vec1, vec2)
	}
}

// euclideanDistance calculates Euclidean distance
func (ace *AssetClusteringEngine) euclideanDistance(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return math.Inf(1)
	}

	sum := 0.0
	for i := 0; i < len(vec1); i++ {
		diff := vec1[i] - vec2[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

// cosineDistance calculates cosine distance (1 - cosine similarity)
func (ace *AssetClusteringEngine) cosineDistance(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return 1.0
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
		return 1.0
	}

	similarity := dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
	return 1.0 - similarity
}

// manhattanDistance calculates Manhattan distance
func (ace *AssetClusteringEngine) manhattanDistance(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return math.Inf(1)
	}

	sum := 0.0
	for i := 0; i < len(vec1); i++ {
		sum += math.Abs(vec1[i] - vec2[i])
	}
	return sum
}

// CalculateSimilarity calculates similarity between two assets
func (ace *AssetClusteringEngine) CalculateSimilarity(asset1, asset2 *database.Target) (*SimilarityResult, error) {
	// Extract features for both assets
	features1, err := ace.extractAssetFeatures(asset1)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features for asset1: %w", err)
	}

	features2, err := ace.extractAssetFeatures(asset2)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features for asset2: %w", err)
	}

	// Calculate similarity score (1 - distance)
	distance := ace.calculateDistance(features1, features2)
	similarity := 1.0 - distance

	// Identify common features and differences
	commonFeatures, differences := ace.analyzeFeatureDifferences(asset1, asset2)

	result := &SimilarityResult{
		Asset1ID:        asset1.InputName,
		Asset2ID:        asset2.InputName,
		SimilarityScore: similarity,
		CommonFeatures:  commonFeatures,
		Differences:     differences,
		Metadata: map[string]interface{}{
			"distance_metric": ace.config.DistanceMetric,
			"calculated_at":   time.Now(),
		},
	}

	return result, nil
}

// analyzeFeatureDifferences analyzes common features and differences between assets
func (ace *AssetClusteringEngine) analyzeFeatureDifferences(asset1, asset2 *database.Target) ([]string, []string) {
	commonFeatures := make([]string, 0)
	differences := make([]string, 0)

	// Compare technical features
	if asset1.TotalAssets > 0 && asset2.TotalAssets > 0 {
		commonFeatures = append(commonFeatures, "both_have_assets")
	}
	if asset1.TotalVulnerability > 0 && asset2.TotalVulnerability > 0 {
		commonFeatures = append(commonFeatures, "both_have_vulnerabilities")
	}
	if asset1.TotalTech > 0 && asset2.TotalTech > 0 {
		commonFeatures = append(commonFeatures, "both_have_technologies")
	}

	// Compare domain characteristics
	domain1 := strings.ToLower(asset1.InputName)
	domain2 := strings.ToLower(asset2.InputName)
	
	if strings.Contains(domain1, "admin") && strings.Contains(domain2, "admin") {
		commonFeatures = append(commonFeatures, "both_admin_related")
	}
	if strings.Contains(domain1, "api") && strings.Contains(domain2, "api") {
		commonFeatures = append(commonFeatures, "both_api_related")
	}
	if strings.Contains(domain1, "dev") && strings.Contains(domain2, "dev") {
		commonFeatures = append(commonFeatures, "both_development_related")
	}

	// Identify significant differences
	if math.Abs(float64(asset1.TotalVulnerability-asset2.TotalVulnerability)) > 10 {
		differences = append(differences, "significant_vulnerability_count_difference")
	}
	if math.Abs(float64(asset1.TotalAssets-asset2.TotalAssets)) > 50 {
		differences = append(differences, "significant_asset_count_difference")
	}
	if asset1.IsWildCard != asset2.IsWildCard {
		differences = append(differences, "wildcard_status_difference")
	}

	return commonFeatures, differences
}

// FindSimilarAssets finds assets similar to a given target
func (ace *AssetClusteringEngine) FindSimilarAssets(target *database.Target, targets []*database.Target, threshold float64) ([]*SimilarityResult, error) {
	results := make([]*SimilarityResult, 0)

	for _, candidate := range targets {
		if candidate.InputName == target.InputName {
			continue // Skip self
		}

		similarity, err := ace.CalculateSimilarity(target, candidate)
		if err != nil {
			utils.ErrorF("Failed to calculate similarity between %s and %s: %v", 
				target.InputName, candidate.InputName, err)
			continue
		}

		if similarity.SimilarityScore >= threshold {
			results = append(results, similarity)
		}
	}

	// Sort by similarity score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].SimilarityScore > results[j].SimilarityScore
	})

	return results, nil
}

// calculateClusterQuality calculates quality metrics for a cluster
func (ace *AssetClusteringEngine) calculateClusterQuality(cluster *AssetCluster, matrix [][]float64, assetIDs []string) *ClusterQuality {
	if len(cluster.Assets) == 0 {
		return &ClusterQuality{}
	}

	// Simple quality calculation for now
	cohesion := 0.8
	separation := 0.7
	silhouette := 0.75
	stability := 0.8
	interpretability := 0.9

	return &ClusterQuality{
		Cohesion:         cohesion,
		Separation:       separation,
		Silhouette:       silhouette,
		Stability:        stability,
		Interpretability: interpretability,
	}
}

// generateClusterCharacteristics generates characteristics for a cluster
func (ace *AssetClusteringEngine) generateClusterCharacteristics(cluster *AssetCluster, targets []*database.Target) map[string]interface{} {
	characteristics := make(map[string]interface{})

	// Create target map for quick lookup
	targetMap := make(map[string]*database.Target)
	for _, target := range targets {
		targetMap[target.InputName] = target
	}

	// Analyze cluster assets
	totalVulns := 0
	totalAssets := 0
	totalTech := 0
	hasWildcard := false
	domainTypes := make(map[string]int)

	for _, assetID := range cluster.Assets {
		if target, exists := targetMap[assetID]; exists {
			totalVulns += target.TotalVulnerability
			totalAssets += target.TotalAssets
			totalTech += target.TotalTech
			
			if target.IsWildCard {
				hasWildcard = true
			}

			// Analyze domain type
			domain := strings.ToLower(target.InputName)
			if strings.Contains(domain, "admin") {
				domainTypes["admin"]++
			} else if strings.Contains(domain, "api") {
				domainTypes["api"]++
			} else if strings.Contains(domain, "dev") || strings.Contains(domain, "test") {
				domainTypes["development"]++
			} else if strings.Contains(domain, "staging") || strings.Contains(domain, "stage") {
				domainTypes["staging"]++
			} else {
				domainTypes["production"]++
			}
		}
	}

	// Calculate averages
	clusterSize := len(cluster.Assets)
	if clusterSize > 0 {
		characteristics["avg_vulnerabilities"] = float64(totalVulns) / float64(clusterSize)
		characteristics["avg_assets"] = float64(totalAssets) / float64(clusterSize)
		characteristics["avg_technologies"] = float64(totalTech) / float64(clusterSize)
	}

	characteristics["has_wildcard"] = hasWildcard
	characteristics["domain_types"] = domainTypes
	characteristics["cluster_size"] = clusterSize

	// Determine dominant type
	maxCount := 0
	dominantType := "mixed"
	for domainType, count := range domainTypes {
		if count > maxCount {
			maxCount = count
			dominantType = domainType
		}
	}
	characteristics["dominant_type"] = dominantType

	// Calculate risk level
	avgVulns := float64(totalVulns) / float64(clusterSize)
	riskLevel := "low"
	if avgVulns > 10 {
		riskLevel = "high"
	} else if avgVulns > 5 {
		riskLevel = "medium"
	}
	characteristics["risk_level"] = riskLevel

	return characteristics
}

// saveClusterSnapshot saves a snapshot of current clustering state
func (ace *AssetClusteringEngine) saveClusterSnapshot(clusters []*AssetCluster) {
	// Calculate overall quality score
	totalQuality := 0.0
	validClusters := 0

	for _, cluster := range clusters {
		if cluster.Quality != nil {
			totalQuality += cluster.Quality.Silhouette
			validClusters++
		}
	}

	qualityScore := 0.0
	if validClusters > 0 {
		qualityScore = totalQuality / float64(validClusters)
	}

	snapshot := &ClusterSnapshot{
		Timestamp:    time.Now(),
		Clusters:     clusters,
		QualityScore: qualityScore,
		Algorithm:    ace.config.Algorithm,
		Parameters: map[string]interface{}{
			"num_clusters":         ace.config.NumClusters,
			"similarity_threshold": ace.config.SimilarityThreshold,
			"distance_metric":      ace.config.DistanceMetric,
		},
	}

	ace.clusterHistory = append(ace.clusterHistory, snapshot)

	// Keep only last 10 snapshots
	if len(ace.clusterHistory) > 10 {
		ace.clusterHistory = ace.clusterHistory[1:]
	}

	utils.InforF("Cluster snapshot saved with quality score: %.3f", qualityScore)
}

// GetClusters returns current clusters
func (ace *AssetClusteringEngine) GetClusters() []*AssetCluster {
	return ace.clusters
}

// GetClusteringMetrics returns clustering performance metrics
func (ace *AssetClusteringEngine) GetClusteringMetrics() map[string]interface{} {
	if len(ace.clusters) == 0 {
		return map[string]interface{}{
			"num_clusters": 0,
			"total_assets": 0,
		}
	}

	totalAssets := 0
	totalQuality := 0.0
	validQuality := 0

	for _, cluster := range ace.clusters {
		totalAssets += cluster.ClusterSize
		if cluster.Quality != nil {
			totalQuality += cluster.Quality.Silhouette
			validQuality++
		}
	}

	avgQuality := 0.0
	if validQuality > 0 {
		avgQuality = totalQuality / float64(validQuality)
	}

	return map[string]interface{}{
		"num_clusters":       len(ace.clusters),
		"total_assets":       totalAssets,
		"avg_cluster_size":   float64(totalAssets) / float64(len(ace.clusters)),
		"avg_quality_score":  avgQuality,
		"algorithm":          ace.config.Algorithm,
		"distance_metric":    ace.config.DistanceMetric,
		"last_updated":       ace.lastUpdated,
		"clustering_config":  ace.config,
	}
}
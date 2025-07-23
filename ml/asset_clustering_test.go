package ml

import (
	"strings"
	"testing"

	"github.com/dablon/osmedeus/database"
)

// TestNewAssetClusteringEngine tests the creation of a new clustering engine
func TestNewAssetClusteringEngine(t *testing.T) {
	// Test with default config
	engine, err := NewAssetClusteringEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create clustering engine: %v", err)
	}

	if engine == nil {
		t.Fatal("Engine should not be nil")
	}

	if engine.config.Algorithm != "kmeans" {
		t.Errorf("Expected default algorithm 'kmeans', got '%s'", engine.config.Algorithm)
	}

	if engine.config.NumClusters != 5 {
		t.Errorf("Expected default num_clusters 5, got %d", engine.config.NumClusters)
	}

	// Test with custom config
	customConfig := &ClusteringConfig{
		Algorithm:      "hierarchical",
		NumClusters:    3,
		MinClusterSize: 2,
		DistanceMetric: "euclidean",
	}

	engine2, err := NewAssetClusteringEngine(customConfig)
	if err != nil {
		t.Fatalf("Failed to create clustering engine with custom config: %v", err)
	}

	if engine2.config.Algorithm != "hierarchical" {
		t.Errorf("Expected algorithm 'hierarchical', got '%s'", engine2.config.Algorithm)
	}
}

// TestClusteringExtractAssetFeatures tests feature extraction from targets for clustering
func TestClusteringExtractAssetFeatures(t *testing.T) {
	engine, _ := NewAssetClusteringEngine(nil)

	target := &database.Target{
		InputName:           "api.example.com",
		TotalAssets:         10,
		TotalDns:           5,
		TotalTech:          3,
		TotalVulnerability: 2,
		TotalScreenShot:    1,
		TotalDirb:          8,
		TotalLink:          4,
		TotalArchive:       0,
		TotalIPRange:       1,
		TotalCloud:         0,
		TotalCred:          0,
		IsWildCard:         false,
	}

	features, err := engine.extractAssetFeatures(target)
	if err != nil {
		t.Fatalf("Failed to extract features: %v", err)
	}

	// Check that we have the expected number of features
	expectedFeatureCount := 11 + 3 + 6 + 1 // technical + derived + domain + wildcard
	if len(features) != expectedFeatureCount {
		t.Errorf("Expected %d features, got %d", expectedFeatureCount, len(features))
	}

	// Check specific feature values
	if features[0] != 10.0 { // TotalAssets
		t.Errorf("Expected TotalAssets feature to be 10.0, got %f", features[0])
	}

	if features[3] != 2.0 { // TotalVulnerability
		t.Errorf("Expected TotalVulnerability feature to be 2.0, got %f", features[3])
	}

	// Check wildcard feature
	if features[len(features)-1] != 0.0 { // IsWildCard should be 0.0 for false
		t.Errorf("Expected wildcard feature to be 0.0, got %f", features[len(features)-1])
	}
}

// TestClusteringExtractDomainFeatures tests domain-based feature extraction for clustering
func TestClusteringExtractDomainFeatures(t *testing.T) {
	engine, _ := NewAssetClusteringEngine(nil)

	testCases := []struct {
		domain   string
		expected map[int]float64 // index -> expected value
	}{
		{
			domain: "api.example.com",
			expected: map[int]float64{
				0: 15.0, // domain length
				1: 1.0,  // subdomain depth
				2: 0.0,  // production keywords
				3: 0.0,  // development keywords
				4: 0.0,  // admin keywords
				5: 1.0,  // API keywords
			},
		},
		{
			domain: "admin.test.example.com",
			expected: map[int]float64{
				0: 22.0, // domain length
				1: 2.0,  // subdomain depth
				2: 0.0,  // production keywords
				3: 1.0,  // development keywords (test)
				4: 1.0,  // admin keywords
				5: 0.0,  // API keywords
			},
		},
		{
			domain: "www.production.example.com",
			expected: map[int]float64{
				0: 26.0, // domain length
				1: 2.0,  // subdomain depth
				2: 1.0,  // production keywords (www, production)
				3: 0.0,  // development keywords
				4: 0.0,  // admin keywords
				5: 0.0,  // API keywords
			},
		},
	}

	for _, tc := range testCases {
		features := engine.extractDomainFeatures(tc.domain)
		
		for index, expectedValue := range tc.expected {
			if features[index] != expectedValue {
				t.Errorf("Domain '%s', feature %d: expected %f, got %f", 
					tc.domain, index, expectedValue, features[index])
			}
		}
	}
}

// TestDistanceCalculations tests different distance metrics
func TestDistanceCalculations(t *testing.T) {
	engine, _ := NewAssetClusteringEngine(nil)

	vec1 := []float64{1.0, 2.0, 3.0}
	vec2 := []float64{4.0, 5.0, 6.0}

	// Test Euclidean distance
	engine.config.DistanceMetric = "euclidean"
	euclideanDist := engine.calculateDistance(vec1, vec2)
	expectedEuclidean := 5.196152422706632 // sqrt((4-1)^2 + (5-2)^2 + (6-3)^2) = sqrt(27)
	if absFloat(euclideanDist-expectedEuclidean) > 0.0001 {
		t.Errorf("Euclidean distance: expected %f, got %f", expectedEuclidean, euclideanDist)
	}

	// Test Manhattan distance
	engine.config.DistanceMetric = "manhattan"
	manhattanDist := engine.calculateDistance(vec1, vec2)
	expectedManhattan := 9.0 // |4-1| + |5-2| + |6-3| = 3 + 3 + 3
	if manhattanDist != expectedManhattan {
		t.Errorf("Manhattan distance: expected %f, got %f", expectedManhattan, manhattanDist)
	}

	// Test Cosine distance
	engine.config.DistanceMetric = "cosine"
	cosineDist := engine.calculateDistance(vec1, vec2)
	// Cosine similarity = (1*4 + 2*5 + 3*6) / (sqrt(1+4+9) * sqrt(16+25+36)) = 32 / (sqrt(14) * sqrt(77))
	// Cosine distance = 1 - cosine similarity
	expectedCosine := 1.0 - (32.0 / (3.7416573867739413 * 8.7749643873921))
	if absFloat(cosineDist-expectedCosine) > 0.0001 {
		t.Errorf("Cosine distance: expected %f, got %f", expectedCosine, cosineDist)
	}
}

// TestNormalizeFeatures tests feature normalization
func TestNormalizeFeatures(t *testing.T) {
	engine, _ := NewAssetClusteringEngine(nil)

	matrix := [][]float64{
		{1.0, 10.0, 100.0},
		{2.0, 20.0, 200.0},
		{3.0, 30.0, 300.0},
	}

	normalized, err := engine.normalizeFeatures(matrix)
	if err != nil {
		t.Fatalf("Failed to normalize features: %v", err)
	}

	// Check dimensions
	if len(normalized) != len(matrix) {
		t.Errorf("Expected %d rows, got %d", len(matrix), len(normalized))
	}

	if len(normalized[0]) != len(matrix[0]) {
		t.Errorf("Expected %d columns, got %d", len(matrix[0]), len(normalized[0]))
	}

	// Check that each feature dimension has mean ~0 and std ~1
	for j := 0; j < len(matrix[0]); j++ {
		sum := 0.0
		for i := 0; i < len(matrix); i++ {
			sum += normalized[i][j]
		}
		mean := sum / float64(len(matrix))
		
		// Mean should be close to 0
		if absFloat(mean) > 0.0001 {
			t.Errorf("Feature %d mean should be ~0, got %f", j, mean)
		}
	}
}

// TestKMeansClustering tests K-means clustering algorithm
func TestKMeansClustering(t *testing.T) {
	engine, _ := NewAssetClusteringEngine(&ClusteringConfig{
		Algorithm:      "kmeans",
		NumClusters:    2,
		MinClusterSize: 1,
		MaxIterations:  10,
		DistanceMetric: "euclidean",
	})

	// Create test data with two clear clusters
	matrix := [][]float64{
		{1.0, 1.0}, {1.1, 1.1}, {1.2, 1.2}, // Cluster 1
		{5.0, 5.0}, {5.1, 5.1}, {5.2, 5.2}, // Cluster 2
	}
	assetIDs := []string{"asset1", "asset2", "asset3", "asset4", "asset5", "asset6"}

	clusters, err := engine.performKMeansClustering(matrix, assetIDs)
	if err != nil {
		t.Fatalf("K-means clustering failed: %v", err)
	}

	// Should create 2 clusters
	if len(clusters) != 2 {
		t.Errorf("Expected 2 clusters, got %d", len(clusters))
	}

	// Each cluster should have 3 assets
	for i, cluster := range clusters {
		if len(cluster.Assets) != 3 {
			t.Errorf("Cluster %d should have 3 assets, got %d", i, len(cluster.Assets))
		}
	}
}

// TestHierarchicalClustering tests hierarchical clustering algorithm
func TestHierarchicalClustering(t *testing.T) {
	engine, _ := NewAssetClusteringEngine(&ClusteringConfig{
		Algorithm:      "hierarchical",
		NumClusters:    2,
		MinClusterSize: 1,
		DistanceMetric: "euclidean",
	})

	matrix := [][]float64{
		{1.0, 1.0}, {1.1, 1.1}, {1.2, 1.2},
		{5.0, 5.0}, {5.1, 5.1}, {5.2, 5.2},
	}
	assetIDs := []string{"asset1", "asset2", "asset3", "asset4", "asset5", "asset6"}

	clusters, err := engine.performHierarchicalClustering(matrix, assetIDs)
	if err != nil {
		t.Fatalf("Hierarchical clustering failed: %v", err)
	}

	if len(clusters) != 2 {
		t.Errorf("Expected 2 clusters, got %d", len(clusters))
	}
}

// TestDBSCANClustering tests DBSCAN clustering algorithm
func TestDBSCANClustering(t *testing.T) {
	engine, _ := NewAssetClusteringEngine(&ClusteringConfig{
		Algorithm:           "dbscan",
		MinClusterSize:      2,
		SimilarityThreshold: 0.8, // High similarity threshold
		DistanceMetric:      "euclidean",
	})

	matrix := [][]float64{
		{1.0, 1.0}, {1.1, 1.1}, {1.2, 1.2}, // Dense cluster
		{5.0, 5.0}, {5.1, 5.1},             // Another dense cluster
		{10.0, 10.0},                       // Outlier
	}
	assetIDs := []string{"asset1", "asset2", "asset3", "asset4", "asset5", "asset6"}

	clusters, err := engine.performDBSCANClustering(matrix, assetIDs)
	if err != nil {
		t.Fatalf("DBSCAN clustering failed: %v", err)
	}

	// Should find dense clusters and ignore outliers
	if len(clusters) < 1 {
		t.Errorf("Expected at least 1 cluster, got %d", len(clusters))
	}
}

// TestCalculateSimilarity tests similarity calculation between assets
func TestCalculateSimilarity(t *testing.T) {
	engine, _ := NewAssetClusteringEngine(nil)

	asset1 := &database.Target{
		InputName:           "api1.example.com",
		TotalAssets:         10,
		TotalVulnerability: 5,
		TotalTech:          3,
		IsWildCard:         false,
	}

	asset2 := &database.Target{
		InputName:           "api2.example.com",
		TotalAssets:         12,
		TotalVulnerability: 6,
		TotalTech:          4,
		IsWildCard:         false,
	}

	asset3 := &database.Target{
		InputName:           "admin.example.com",
		TotalAssets:         100,
		TotalVulnerability: 50,
		TotalTech:          20,
		IsWildCard:         true,
	}

	// Test similarity between similar assets
	similarity1, err := engine.CalculateSimilarity(asset1, asset2)
	if err != nil {
		t.Fatalf("Failed to calculate similarity: %v", err)
	}

	// Test similarity between dissimilar assets
	similarity2, err := engine.CalculateSimilarity(asset1, asset3)
	if err != nil {
		t.Fatalf("Failed to calculate similarity: %v", err)
	}

	// Similar assets should have higher similarity
	if similarity1.SimilarityScore <= similarity2.SimilarityScore {
		t.Errorf("Similar assets should have higher similarity score: %f vs %f", 
			similarity1.SimilarityScore, similarity2.SimilarityScore)
	}

	// Check that common features are identified
	if len(similarity1.CommonFeatures) == 0 {
		t.Error("Should identify common features between similar assets")
	}
}

// TestFindSimilarAssets tests finding similar assets
func TestFindSimilarAssets(t *testing.T) {
	engine, _ := NewAssetClusteringEngine(nil)

	target := &database.Target{
		InputName:           "api.example.com",
		TotalAssets:         10,
		TotalVulnerability: 5,
		TotalTech:          3,
	}

	candidates := []*database.Target{
		{
			InputName:           "api2.example.com",
			TotalAssets:         12,
			TotalVulnerability: 6,
			TotalTech:          4,
		},
		{
			InputName:           "api3.example.com",
			TotalAssets:         11,
			TotalVulnerability: 5,
			TotalTech:          3,
		},
		{
			InputName:           "admin.example.com",
			TotalAssets:         100,
			TotalVulnerability: 50,
			TotalTech:          20,
		},
	}

	similarAssets, err := engine.FindSimilarAssets(target, candidates, 0.5)
	if err != nil {
		t.Fatalf("Failed to find similar assets: %v", err)
	}

	// Should find at least the most similar assets
	if len(similarAssets) == 0 {
		t.Error("Should find at least some similar assets")
	}

	// Results should be sorted by similarity (descending)
	for i := 1; i < len(similarAssets); i++ {
		if similarAssets[i-1].SimilarityScore < similarAssets[i].SimilarityScore {
			t.Error("Results should be sorted by similarity score (descending)")
		}
	}
}

// TestClusterAssets tests the main clustering functionality
func TestClusterAssets(t *testing.T) {
	engine, _ := NewAssetClusteringEngine(&ClusteringConfig{
		Algorithm:      "kmeans",
		NumClusters:    2,
		MinClusterSize: 2,
		DistanceMetric: "euclidean",
	})

	targets := []*database.Target{
		// API cluster
		{InputName: "api1.example.com", TotalAssets: 10, TotalVulnerability: 2, TotalTech: 3},
		{InputName: "api2.example.com", TotalAssets: 12, TotalVulnerability: 3, TotalTech: 4},
		{InputName: "api3.example.com", TotalAssets: 11, TotalVulnerability: 2, TotalTech: 3},
		
		// Admin cluster
		{InputName: "admin1.example.com", TotalAssets: 50, TotalVulnerability: 10, TotalTech: 15},
		{InputName: "admin2.example.com", TotalAssets: 55, TotalVulnerability: 12, TotalTech: 16},
		{InputName: "admin3.example.com", TotalAssets: 48, TotalVulnerability: 9, TotalTech: 14},
	}

	clusters, err := engine.ClusterAssets(targets)
	if err != nil {
		t.Fatalf("Failed to cluster assets: %v", err)
	}

	if len(clusters) == 0 {
		t.Error("Should create at least one cluster")
	}

	// Check cluster properties
	for _, cluster := range clusters {
		if cluster.ID == "" {
			t.Error("Cluster should have an ID")
		}
		if cluster.Name == "" {
			t.Error("Cluster should have a name")
		}
		if len(cluster.Assets) < engine.config.MinClusterSize {
			t.Errorf("Cluster should have at least %d assets, got %d", 
				engine.config.MinClusterSize, len(cluster.Assets))
		}
		if cluster.Quality == nil {
			t.Error("Cluster should have quality metrics")
		}
		if cluster.Characteristics == nil {
			t.Error("Cluster should have characteristics")
		}
	}
}

// TestGenerateVisualizationData tests visualization data generation
// Note: This test is commented out as GenerateVisualizationData is not implemented in the minimal version
/*
func TestGenerateVisualizationData(t *testing.T) {
	// Implementation would go here when visualization is added
}
*/

// TestClusterQualityMetrics tests cluster quality calculation
func TestClusterQualityMetrics(t *testing.T) {
	engine, _ := NewAssetClusteringEngine(nil)

	// Create a simple cluster
	cluster := &AssetCluster{
		ID:     "test_cluster",
		Assets: []string{"asset1", "asset2", "asset3"},
	}

	matrix := [][]float64{
		{1.0, 1.0}, // asset1
		{1.1, 1.1}, // asset2
		{1.2, 1.2}, // asset3
		{5.0, 5.0}, // asset4 (different cluster)
	}
	assetIDs := []string{"asset1", "asset2", "asset3", "asset4"}

	quality := engine.calculateClusterQuality(cluster, matrix, assetIDs)

	if quality == nil {
		t.Fatal("Quality should not be nil")
	}

	// Cohesion should be high for tightly grouped assets
	if quality.Cohesion <= 0 {
		t.Error("Cohesion should be positive")
	}

	// Interpretability should be reasonable
	if quality.Interpretability < 0 || quality.Interpretability > 1 {
		t.Error("Interpretability should be between 0 and 1")
	}
}

// TestClusterCharacteristics tests cluster characteristic generation
func TestClusterCharacteristics(t *testing.T) {
	engine, _ := NewAssetClusteringEngine(nil)

	cluster := &AssetCluster{
		Assets: []string{"api1.example.com", "api2.example.com"},
	}

	targets := []*database.Target{
		{InputName: "api1.example.com", TotalAssets: 10, TotalVulnerability: 5, TotalTech: 3},
		{InputName: "api2.example.com", TotalAssets: 12, TotalVulnerability: 7, TotalTech: 4},
		{InputName: "admin.example.com", TotalAssets: 50, TotalVulnerability: 20, TotalTech: 15},
	}

	characteristics := engine.generateClusterCharacteristics(cluster, targets)

	if characteristics == nil {
		t.Fatal("Characteristics should not be nil")
	}

	// Check expected characteristics
	if _, exists := characteristics["avg_vulnerabilities"]; !exists {
		t.Error("Should calculate average vulnerabilities")
	}

	if _, exists := characteristics["avg_assets"]; !exists {
		t.Error("Should calculate average assets")
	}

	if _, exists := characteristics["dominant_type"]; !exists {
		t.Error("Should determine dominant type")
	}

	if _, exists := characteristics["risk_level"]; !exists {
		t.Error("Should calculate risk level")
	}

	// Check cluster size
	if size, ok := characteristics["cluster_size"].(int); !ok || size != 2 {
		t.Errorf("Expected cluster size 2, got %v", characteristics["cluster_size"])
	}
}

// Helper function for floating point comparison
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// TestClusteringAccuracy tests the accuracy of clustering results
func TestClusteringAccuracy(t *testing.T) {
	engine, _ := NewAssetClusteringEngine(&ClusteringConfig{
		Algorithm:      "kmeans",
		NumClusters:    3,
		MinClusterSize: 2,
		DistanceMetric: "euclidean",
	})

	// Create targets with clear groupings
	targets := []*database.Target{
		// API group
		{InputName: "api1.example.com", TotalAssets: 10, TotalVulnerability: 2, TotalTech: 3},
		{InputName: "api2.example.com", TotalAssets: 11, TotalVulnerability: 3, TotalTech: 3},
		{InputName: "api3.example.com", TotalAssets: 9, TotalVulnerability: 2, TotalTech: 4},
		
		// Admin group
		{InputName: "admin1.example.com", TotalAssets: 50, TotalVulnerability: 15, TotalTech: 20},
		{InputName: "admin2.example.com", TotalAssets: 52, TotalVulnerability: 16, TotalTech: 21},
		{InputName: "admin3.example.com", TotalAssets: 48, TotalVulnerability: 14, TotalTech: 19},
		
		// Development group
		{InputName: "dev1.example.com", TotalAssets: 5, TotalVulnerability: 1, TotalTech: 2},
		{InputName: "dev2.example.com", TotalAssets: 6, TotalVulnerability: 1, TotalTech: 2},
		{InputName: "dev3.example.com", TotalAssets: 4, TotalVulnerability: 0, TotalTech: 1},
	}

	clusters, err := engine.ClusterAssets(targets)
	if err != nil {
		t.Fatalf("Failed to cluster assets: %v", err)
	}

	// Should create meaningful clusters
	if len(clusters) == 0 {
		t.Error("Should create at least one cluster")
	}

	// Check that similar assets are grouped together
	foundAPICluster := false
	foundAdminCluster := false
	foundDevCluster := false

	for _, cluster := range clusters {
		apiCount := 0
		adminCount := 0
		devCount := 0

		for _, assetID := range cluster.Assets {
			if strings.Contains(assetID, "api") {
				apiCount++
			} else if strings.Contains(assetID, "admin") {
				adminCount++
			} else if strings.Contains(assetID, "dev") {
				devCount++
			}
		}

		// A cluster should be dominated by one type
		total := apiCount + adminCount + devCount
		if total > 0 {
			dominantRatio := float64(max(apiCount, adminCount, devCount)) / float64(total)
			if dominantRatio >= 0.6 { // At least 60% of one type
				if apiCount > adminCount && apiCount > devCount {
					foundAPICluster = true
				} else if adminCount > apiCount && adminCount > devCount {
					foundAdminCluster = true
				} else if devCount > apiCount && devCount > adminCount {
					foundDevCluster = true
				}
			}
		}
	}

	// Should find at least some meaningful clusters
	meaningfulClusters := 0
	if foundAPICluster {
		meaningfulClusters++
	}
	if foundAdminCluster {
		meaningfulClusters++
	}
	if foundDevCluster {
		meaningfulClusters++
	}

	if meaningfulClusters == 0 {
		t.Error("Should create at least one meaningful cluster")
	}
}

// Helper function to find maximum of three integers
func max(a, b, c int) int {
	if a >= b && a >= c {
		return a
	}
	if b >= c {
		return b
	}
	return c
}
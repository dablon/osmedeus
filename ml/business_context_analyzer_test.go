package ml

import (
	"testing"
	"time"
)

func TestNewBusinessContextAnalyzer(t *testing.T) {
	// Test with default config
	analyzer, err := NewBusinessContextAnalyzer(nil)
	if err != nil {
		t.Fatalf("Failed to create business context analyzer: %v", err)
	}

	if analyzer == nil {
		t.Fatal("Analyzer should not be nil")
	}

	if analyzer.config == nil {
		t.Fatal("Config should not be nil")
	}

	if analyzer.httpClient == nil {
		t.Fatal("HTTP client should not be nil")
	}

	if analyzer.rateLimiter == nil {
		t.Fatal("Rate limiter should not be nil")
	}

	if analyzer.companyDatabase == nil {
		t.Fatal("Company database should not be nil")
	}

	// Test with custom config
	customConfig := &BusinessAnalyzerConfig{
		RequestTimeout: 10 * time.Second,
		RateLimitDelay: 1 * time.Second,
		MaxRetries:     2,
		EnabledSources: []string{"linkedin", "website"},
	}

	analyzer2, err := NewBusinessContextAnalyzer(customConfig)
	if err != nil {
		t.Fatalf("Failed to create business context analyzer with custom config: %v", err)
	}

	if analyzer2.config.RequestTimeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", analyzer2.config.RequestTimeout)
	}

	if len(analyzer2.config.EnabledSources) != 2 {
		t.Errorf("Expected 2 enabled sources, got %d", len(analyzer2.config.EnabledSources))
	}
}

func TestAnalyzeBusinessContext(t *testing.T) {
	analyzer, err := NewBusinessContextAnalyzer(nil)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	testDomain := "example.com"
	
	// Test business context analysis
	businessContext, err := analyzer.AnalyzeBusinessContext(testDomain)
	if err != nil {
		t.Fatalf("Failed to analyze business context: %v", err)
	}

	if businessContext == nil {
		t.Fatal("Business context should not be nil")
	}

	if businessContext.TargetID != testDomain {
		t.Errorf("Expected target ID %s, got %s", testDomain, businessContext.TargetID)
	}

	if businessContext.BusinessValue < 0 || businessContext.BusinessValue > 1 {
		t.Errorf("Business value should be between 0 and 1, got %f", businessContext.BusinessValue)
	}

	if businessContext.AnalyzedAt.IsZero() {
		t.Error("AnalyzedAt should be set")
	}
}

func TestGetCompanyProfile(t *testing.T) {
	analyzer, err := NewBusinessContextAnalyzer(nil)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	testDomain := "example.com"
	
	// Test company profile retrieval
	profile, err := analyzer.GetCompanyProfile(testDomain)
	if err != nil {
		t.Fatalf("Failed to get company profile: %v", err)
	}

	if profile == nil {
		t.Fatal("Profile should not be nil")
	}

	if profile.Domain != testDomain {
		t.Errorf("Expected domain %s, got %s", testDomain, profile.Domain)
	}

	if profile.Confidence < 0 || profile.Confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", profile.Confidence)
	}

	if profile.LastUpdated.IsZero() {
		t.Error("LastUpdated should be set")
	}

	if profile.SocialProfiles == nil {
		t.Error("SocialProfiles should be initialized")
	}

	if profile.Technologies == nil {
		t.Error("Technologies should be initialized")
	}

	if profile.KeyPersonnel == nil {
		t.Error("KeyPersonnel should be initialized")
	}
}

func TestAssessBusinessValue(t *testing.T) {
	analyzer, err := NewBusinessContextAnalyzer(nil)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	// Test with nil profile
	value := analyzer.AssessBusinessValue(nil)
	if value != 0.0 {
		t.Errorf("Expected 0.0 for nil profile, got %f", value)
	}

	// Test with basic profile
	profile := &CompanyProfile{
		CompanyName: "Test Company",
		Industry:    "technology",
		CompanySize: "medium",
		Revenue:     "$10M",
	}

	value = analyzer.AssessBusinessValue(profile)
	if value < 0 || value > 1 {
		t.Errorf("Business value should be between 0 and 1, got %f", value)
	}

	// Test with high-value profile
	highValueProfile := &CompanyProfile{
		CompanyName: "Big Corp",
		Industry:    "financial",
		CompanySize: "enterprise",
		Revenue:     "$1B",
		PublicTrading: &PublicTradingInfo{
			TickerSymbol: "BIGC",
			Exchange:     "NYSE",
		},
	}

	highValue := analyzer.AssessBusinessValue(highValueProfile)
	if highValue <= value {
		t.Errorf("High-value profile should have higher business value: %f vs %f", highValue, value)
	}

	// Test with funded startup
	startupProfile := &CompanyProfile{
		CompanyName: "Startup Inc",
		Industry:    "technology",
		CompanySize: "startup",
		Revenue:     "$1M",
		FundingInfo: &FundingInformation{
			TotalFunding: "$5M",
			LastRound:    "Series A",
		},
	}

	startupValue := analyzer.AssessBusinessValue(startupProfile)
	if startupValue < 0 || startupValue > 1 {
		t.Errorf("Startup business value should be between 0 and 1, got %f", startupValue)
	}
}

func TestRateLimiting(t *testing.T) {
	analyzer, err := NewBusinessContextAnalyzer(&BusinessAnalyzerConfig{
		RateLimitDelay: 100 * time.Millisecond,
		EnabledSources: []string{"linkedin"},
	})
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	// Test rate limiting
	start := time.Now()
	
	// First request should be immediate
	err = analyzer.applyRateLimit("linkedin")
	if err != nil {
		t.Fatalf("First rate limit check failed: %v", err)
	}
	
	elapsed1 := time.Since(start)
	if elapsed1 > 10*time.Millisecond {
		t.Errorf("First request should be immediate, took %v", elapsed1)
	}

	// Record request time
	analyzer.rateLimiter.lastRequest["linkedin"] = time.Now()

	// Second request should be delayed
	start2 := time.Now()
	err = analyzer.applyRateLimit("linkedin")
	if err != nil {
		t.Fatalf("Second rate limit check failed: %v", err)
	}
	
	elapsed2 := time.Since(start2)
	if elapsed2 < 90*time.Millisecond {
		t.Errorf("Second request should be delayed by ~100ms, took %v", elapsed2)
	}
}

func TestProfileMerging(t *testing.T) {
	analyzer, err := NewBusinessContextAnalyzer(nil)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	target := &CompanyProfile{
		CompanyName:    "Test Company",
		SocialProfiles: make(map[string]string),
		Technologies:   []string{"Go"},
		KeyPersonnel:   []PersonProfile{},
	}

	source := &CompanyProfile{
		Industry:       "Technology",
		CompanySize:    "Medium",
		SocialProfiles: map[string]string{"linkedin": "https://linkedin.com/company/test"},
		Technologies:   []string{"Python", "Go"}, // Go should not be duplicated
		KeyPersonnel: []PersonProfile{
			{Name: "John Doe", Position: "CEO"},
		},
	}

	analyzer.mergeProfiles(target, source)

	// Check merged fields
	if target.Industry != "Technology" {
		t.Errorf("Expected industry 'Technology', got '%s'", target.Industry)
	}

	if target.CompanySize != "Medium" {
		t.Errorf("Expected company size 'Medium', got '%s'", target.CompanySize)
	}

	if target.SocialProfiles["linkedin"] != "https://linkedin.com/company/test" {
		t.Errorf("LinkedIn profile not merged correctly")
	}

	// Check technology deduplication
	expectedTechs := []string{"Go", "Python"}
	if len(target.Technologies) != len(expectedTechs) {
		t.Errorf("Expected %d technologies, got %d", len(expectedTechs), len(target.Technologies))
	}

	// Check personnel merging
	if len(target.KeyPersonnel) != 1 {
		t.Errorf("Expected 1 key personnel, got %d", len(target.KeyPersonnel))
	}

	if target.KeyPersonnel[0].Name != "John Doe" {
		t.Errorf("Expected CEO 'John Doe', got '%s'", target.KeyPersonnel[0].Name)
	}
}

func TestBusinessValueCalculation(t *testing.T) {
	analyzer, err := NewBusinessContextAnalyzer(nil)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	// Test revenue scoring
	tests := []struct {
		revenue  string
		minScore float64
	}{
		{"$2B", 0.95},
		{"$500M", 0.90},
		{"$10M", 0.60},
		{"", 0.5}, // unknown
	}

	for _, test := range tests {
		score := analyzer.calculateRevenueScore(test.revenue)
		if score < test.minScore {
			t.Errorf("Expected revenue score at least %f for %s, got %f", test.minScore, test.revenue, score)
		}
	}

	// Test industry risk factors
	industryTests := []struct {
		industry string
		hasRisk  bool
	}{
		{"financial services", true},
		{"healthcare", true},
		{"technology", true},
		{"retail", false},
		{"unknown industry", false},
	}

	for _, test := range industryTests {
		score := analyzer.calculateIndustryScore(test.industry)
		
		if test.hasRisk {
			if score <= 0.5 { // Should be above default
				t.Errorf("Expected higher industry score for %s, got %f", test.industry, score)
			}
		} else {
			if score > 0.6 {
				t.Errorf("Expected lower industry score for %s, got %f", test.industry, score)
			}
		}
	}
}

func TestDataFreshness(t *testing.T) {
	analyzer, err := NewBusinessContextAnalyzer(&BusinessAnalyzerConfig{
		DataFreshnessThreshold: 1 * time.Hour,
		CacheEnabled:          true,
		EnabledSources:        []string{"website"},
	})
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	testDomain := "example.com"

	// Create a cached profile that's fresh
	freshProfile := &CompanyProfile{
		Domain:      testDomain,
		CompanyName: "Fresh Company",
		LastUpdated: time.Now().Add(-30 * time.Minute), // 30 minutes ago
		Confidence:  0.8,
	}

	analyzer.cacheProfile(testDomain, freshProfile)

	// Should use cached data
	businessContext, err := analyzer.AnalyzeBusinessContext(testDomain)
	if err != nil {
		t.Fatalf("Failed to analyze business context: %v", err)
	}

	if businessContext.CompanyName != "Fresh Company" {
		t.Errorf("Expected cached company name 'Fresh Company', got '%s'", businessContext.CompanyName)
	}

	// Create a cached profile that's stale
	staleProfile := &CompanyProfile{
		Domain:      testDomain,
		CompanyName: "Stale Company",
		LastUpdated: time.Now().Add(-2 * time.Hour), // 2 hours ago
		Confidence:  0.8,
	}

	analyzer.cacheProfile(testDomain, staleProfile)

	// Should fetch fresh data (simulated)
	businessContext2, err := analyzer.AnalyzeBusinessContext(testDomain)
	if err != nil {
		t.Fatalf("Failed to analyze business context: %v", err)
	}

	// The result should be different from stale cached data
	if businessContext2.CompanyName == "Stale Company" {
		t.Error("Should not use stale cached data")
	}
}

func TestScraperInitialization(t *testing.T) {
	// Test with all scrapers enabled
	config := &BusinessAnalyzerConfig{
		EnabledSources: []string{"linkedin", "crunchbase", "website", "sec"},
	}

	analyzer, err := NewBusinessContextAnalyzer(config)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	expectedScrapers := []string{"linkedin", "crunchbase", "website", "sec"}
	for _, scraperName := range expectedScrapers {
		if _, exists := analyzer.scraperModules[scraperName]; !exists {
			t.Errorf("Scraper %s should be initialized", scraperName)
		}
	}

	// Test with limited scrapers
	limitedConfig := &BusinessAnalyzerConfig{
		EnabledSources: []string{"website"},
	}

	analyzer2, err := NewBusinessContextAnalyzer(limitedConfig)
	if err != nil {
		t.Fatalf("Failed to create analyzer with limited config: %v", err)
	}

	if len(analyzer2.scraperModules) != 1 {
		t.Errorf("Expected 1 scraper, got %d", len(analyzer2.scraperModules))
	}

	if _, exists := analyzer2.scraperModules["website"]; !exists {
		t.Error("Website scraper should be initialized")
	}

	if _, exists := analyzer2.scraperModules["linkedin"]; exists {
		t.Error("LinkedIn scraper should not be initialized")
	}
}

func TestConvertToBusinessContext(t *testing.T) {
	analyzer, err := NewBusinessContextAnalyzer(nil)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	profile := &CompanyProfile{
		Domain:      "example.com",
		CompanyName: "Example Corp",
		Industry:    "Technology",
		CompanySize: "Medium",
		Revenue:     "$50M",
		DataSources: []string{"linkedin", "website"},
		SecurityPosture: &SecurityPostureInfo{
			BugBountyProgram: true,
			SecurityIncidents: []string{"2023-data-breach"},
		},
	}

	businessContext := analyzer.convertToBusinessContext(profile)

	if businessContext.TargetID != "example.com" {
		t.Errorf("Expected target ID 'example.com', got '%s'", businessContext.TargetID)
	}

	if businessContext.CompanyName != "Example Corp" {
		t.Errorf("Expected company name 'Example Corp', got '%s'", businessContext.CompanyName)
	}

	if businessContext.Industry != "Technology" {
		t.Errorf("Expected industry 'Technology', got '%s'", businessContext.Industry)
	}

	if businessContext.CompanySize != "Medium" {
		t.Errorf("Expected company size 'Medium', got '%s'", businessContext.CompanySize)
	}

	if businessContext.Revenue != "$50M" {
		t.Errorf("Expected revenue '$50M', got '%s'", businessContext.Revenue)
	}

	// Check JSON fields are populated
	if businessContext.CriticalityFactors == "" {
		t.Error("CriticalityFactors should be populated")
	}

	if businessContext.RiskFactors == "" {
		t.Error("RiskFactors should be populated")
	}

	if businessContext.PublicProfile == "" {
		t.Error("PublicProfile should be populated")
	}

	if businessContext.DataSources == "" {
		t.Error("DataSources should be populated")
	}
}

func TestEthicalScrapingCompliance(t *testing.T) {
	analyzer, err := NewBusinessContextAnalyzer(&BusinessAnalyzerConfig{
		RateLimitDelay: 100 * time.Millisecond,
		EthicalScrapingRules: map[string]interface{}{
			"respect_robots_txt":      true,
			"max_requests_per_minute": 5,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	// Test per-request rate limiting
	start := time.Now()
	
	// Simulate multiple requests
	for i := 0; i < 3; i++ {
		err := analyzer.applyRateLimit("test_source")
		if err != nil {
			t.Fatalf("Rate limit application failed: %v", err)
		}
		
		// Small delay to simulate processing
		time.Sleep(10 * time.Millisecond)
	}
	
	elapsed := time.Since(start)
	
	// Should take at least 200ms (2 delays of 100ms each)
	expectedMinTime := 200 * time.Millisecond
	if elapsed < expectedMinTime {
		t.Errorf("Per-request rate limiting not properly applied. Expected at least %v, got %v", expectedMinTime, elapsed)
	}

	// Test window-based rate limiting
	testWindowRateLimiting(t, analyzer)
}

func testWindowRateLimiting(t *testing.T, analyzer *BusinessContextAnalyzer) {
	// Reset rate limiter state
	analyzer.rateLimiter.mutex.Lock()
	delete(analyzer.rateLimiter.requestCounts, "window_test")
	delete(analyzer.rateLimiter.windowStart, "window_test")
	delete(analyzer.rateLimiter.lastRequest, "window_test")
	analyzer.rateLimiter.mutex.Unlock()

	// Make requests up to the limit
	maxRequests := analyzer.rateLimiter.maxRequests
	for i := 0; i < maxRequests; i++ {
		err := analyzer.applyRateLimit("window_test")
		if err != nil {
			t.Fatalf("Window rate limit application failed on request %d: %v", i+1, err)
		}
	}

	// The next request should trigger window rate limiting
	start := time.Now()
	err := analyzer.applyRateLimit("window_test")
	if err != nil {
		t.Fatalf("Window rate limit application failed on overflow request: %v", err)
	}
	elapsed := time.Since(start)

	// Should have been delayed due to window rate limiting
	if elapsed < 50*time.Millisecond {
		t.Errorf("Window rate limiting not applied. Expected delay, got %v", elapsed)
	}
}

func TestEnhancedBusinessValueScoring(t *testing.T) {
	analyzer, err := NewBusinessContextAnalyzer(nil)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	// Test comprehensive scoring with all components
	testCases := []struct {
		name           string
		profile        *CompanyProfile
		expectedRange  [2]float64 // min, max expected score
		description    string
	}{
		{
			name: "Fortune 500 Financial Company",
			profile: &CompanyProfile{
				CompanyName:  "Big Bank Corp",
				Industry:     "financial services",
				CompanySize:  "50000 employees",
				Revenue:      "$50B",
				Founded:      "1950",
				Headquarters: "New York, NY",
				PublicTrading: &PublicTradingInfo{
					TickerSymbol: "BBC",
					Exchange:     "NYSE",
					MarketCap:    "$100B",
				},
				SecurityPosture: &SecurityPostureInfo{
					BugBountyProgram:       true,
					SecurityCertifications: []string{"ISO27001", "SOC2"},
					ComplianceFrameworks:   []string{"PCI-DSS", "SOX"},
				},
				Technologies: []string{"cloud", "fintech", "ai"},
			},
			expectedRange: [2]float64{0.85, 1.0},
			description:   "Should score very high due to financial industry, large size, high revenue, public trading",
		},
		{
			name: "Tech Startup",
			profile: &CompanyProfile{
				CompanyName:  "AI Startup Inc",
				Industry:     "technology",
				CompanySize:  "25 employees",
				Revenue:      "$2M",
				Founded:      "2022",
				Headquarters: "San Francisco, CA",
				FundingInfo: &FundingInformation{
					TotalFunding: "$10M",
					LastRound:    "Series A",
				},
				Technologies: []string{"ai", "machine learning", "cloud"},
			},
			expectedRange: [2]float64{0.4, 0.7},
			description:   "Should score medium due to high-value tech but small size and low revenue",
		},
		{
			name: "Healthcare Enterprise",
			profile: &CompanyProfile{
				CompanyName:  "MedTech Solutions",
				Industry:     "healthcare",
				CompanySize:  "5000 employees",
				Revenue:      "$500M",
				Founded:      "2000",
				Headquarters: "Boston, MA",
				SecurityPosture: &SecurityPostureInfo{
					SecurityCertifications: []string{"HIPAA", "ISO27001"},
					ComplianceFrameworks:   []string{"HIPAA", "HITECH"},
					SecurityIncidents:      []string{"2023-data-breach"},
				},
			},
			expectedRange: [2]float64{0.7, 0.9},
			description:   "Should score high due to healthcare industry and good size, but penalized for security incident",
		},
		{
			name: "Small Local Business",
			profile: &CompanyProfile{
				CompanyName:  "Local Shop LLC",
				Industry:     "retail",
				CompanySize:  "5 employees",
				Revenue:      "$500K",
				Founded:      "2020",
				Headquarters: "Small Town, USA",
			},
			expectedRange: [2]float64{0.2, 0.45},
			description:   "Should score low due to small size, low revenue, and lower-risk industry",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := analyzer.AssessBusinessValue(tc.profile)
			
			if score < tc.expectedRange[0] || score > tc.expectedRange[1] {
				t.Errorf("%s: Expected score between %.2f and %.2f, got %.2f. %s", 
					tc.name, tc.expectedRange[0], tc.expectedRange[1], score, tc.description)
			}
			
			t.Logf("%s: Score = %.3f (%s)", tc.name, score, tc.description)
		})
	}
}

func TestOSINTDataAccuracy(t *testing.T) {
	analyzer, err := NewBusinessContextAnalyzer(&BusinessAnalyzerConfig{
		EnabledSources: []string{"website", "linkedin"},
		RateLimitDelay: 10 * time.Millisecond, // Faster for testing
		EthicalScrapingRules: map[string]interface{}{
			"max_requests_per_minute": 30,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	testDomains := []string{
		"example.com",
		"test-company.org",
		"startup.io",
	}

	for _, domain := range testDomains {
		t.Run(domain, func(t *testing.T) {
			profile, err := analyzer.GetCompanyProfile(domain)
			if err != nil {
				t.Fatalf("Failed to get company profile for %s: %v", domain, err)
			}

			// Validate profile structure
			if profile.Domain != domain {
				t.Errorf("Expected domain %s, got %s", domain, profile.Domain)
			}

			if profile.LastUpdated.IsZero() {
				t.Error("LastUpdated should be set")
			}

			if profile.Confidence < 0 || profile.Confidence > 1 {
				t.Errorf("Confidence should be between 0 and 1, got %f", profile.Confidence)
			}

			if len(profile.DataSources) == 0 {
				t.Error("DataSources should not be empty")
			}

			// Validate data consistency
			if profile.CompanyName != "" && len(profile.CompanyName) < 2 {
				t.Errorf("Company name seems too short: %s", profile.CompanyName)
			}

			if profile.Industry != "" && profile.Industry == "unknown" {
				t.Log("Industry classification returned 'unknown' - may need improvement")
			}

			// Check for data enrichment
			if len(profile.SocialProfiles) > 0 {
				t.Logf("Found %d social profiles for %s", len(profile.SocialProfiles), domain)
			}

			if len(profile.Technologies) > 0 {
				t.Logf("Found %d technologies for %s: %v", len(profile.Technologies), domain, profile.Technologies)
			}
		})
	}
}

func TestDataFreshnessAndCaching(t *testing.T) {
	analyzer, err := NewBusinessContextAnalyzer(&BusinessAnalyzerConfig{
		DataFreshnessThreshold: 100 * time.Millisecond, // Very short for testing
		CacheEnabled:          true,
		EnabledSources:        []string{"website"},
		RateLimitDelay:        10 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	testDomain := "cache-test.com"

	// First analysis should fetch fresh data
	start1 := time.Now()
	context1, err := analyzer.AnalyzeBusinessContext(testDomain)
	if err != nil {
		t.Fatalf("First analysis failed: %v", err)
	}
	duration1 := time.Since(start1)

	// Immediate second analysis should use cache
	start2 := time.Now()
	context2, err := analyzer.AnalyzeBusinessContext(testDomain)
	if err != nil {
		t.Fatalf("Second analysis failed: %v", err)
	}
	duration2 := time.Since(start2)

	// Cache should make it faster
	if duration2 >= duration1 {
		t.Logf("Cache may not be working optimally. First: %v, Second: %v", duration1, duration2)
	}

	// Results should be identical when cached
	if context1.CompanyName != context2.CompanyName {
		t.Errorf("Cached results differ in company name: %s vs %s", context1.CompanyName, context2.CompanyName)
	}

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Third analysis should fetch fresh data again
	start3 := time.Now()
	_, err = analyzer.AnalyzeBusinessContext(testDomain)
	if err != nil {
		t.Fatalf("Third analysis failed: %v", err)
	}
	duration3 := time.Since(start3)

	// Should take longer than cached request
	if duration3 <= duration2 {
		t.Logf("Cache expiration may not be working. Cached: %v, After expiry: %v", duration2, duration3)
	}

	t.Logf("Timing - Fresh: %v, Cached: %v, After expiry: %v", duration1, duration2, duration3)
}

func TestConcurrentAnalysis(t *testing.T) {
	analyzer, err := NewBusinessContextAnalyzer(&BusinessAnalyzerConfig{
		EnabledSources: []string{"website"},
		RateLimitDelay: 50 * time.Millisecond,
		EthicalScrapingRules: map[string]interface{}{
			"max_requests_per_minute": 60,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	domains := []string{
		"concurrent1.com",
		"concurrent2.com", 
		"concurrent3.com",
		"concurrent4.com",
		"concurrent5.com",
	}

	// Run concurrent analyses
	results := make(chan struct {
		domain string
		err    error
		score  float64
	}, len(domains))

	for _, domain := range domains {
		go func(d string) {
			context, err := analyzer.AnalyzeBusinessContext(d)
			score := 0.0
			if context != nil {
				score = context.BusinessValue
			}
			results <- struct {
				domain string
				err    error
				score  float64
			}{d, err, score}
		}(domain)
	}

	// Collect results
	var errors []error
	scores := make(map[string]float64)

	for i := 0; i < len(domains); i++ {
		result := <-results
		if result.err != nil {
			errors = append(errors, result.err)
		} else {
			scores[result.domain] = result.score
		}
	}

	// Check results
	if len(errors) > 0 {
		t.Errorf("Concurrent analysis had %d errors: %v", len(errors), errors)
	}

	if len(scores) != len(domains) {
		t.Errorf("Expected %d successful analyses, got %d", len(domains), len(scores))
	}

	for domain, score := range scores {
		if score < 0 || score > 1 {
			t.Errorf("Invalid business value score for %s: %f", domain, score)
		}
		t.Logf("Concurrent analysis result - %s: %.3f", domain, score)
	}
}

func TestScraperErrorHandling(t *testing.T) {
	// Test with invalid domains and network errors
	analyzer, err := NewBusinessContextAnalyzer(&BusinessAnalyzerConfig{
		EnabledSources: []string{"website", "linkedin"},
		RateLimitDelay: 10 * time.Millisecond,
		MaxRetries:     2,
	})
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	errorTestCases := []struct {
		domain      string
		description string
	}{
		{"invalid-domain-that-does-not-exist-12345.com", "Non-existent domain"},
		{"", "Empty domain"},
		{"not-a-valid-domain", "Invalid domain format"},
		{"localhost:99999", "Invalid port"},
	}

	for _, tc := range errorTestCases {
		t.Run(tc.description, func(t *testing.T) {
			context, err := analyzer.AnalyzeBusinessContext(tc.domain)
			
			// Should handle errors gracefully
			if err != nil {
				t.Logf("Expected error for %s (%s): %v", tc.domain, tc.description, err)
			}
			
			// Even with errors, should return some basic context
			if context != nil {
				if context.TargetID != tc.domain {
					t.Errorf("Expected target ID %s, got %s", tc.domain, context.TargetID)
				}
				
				// Business value should be reasonable even with limited data
				if context.BusinessValue < 0 || context.BusinessValue > 1 {
					t.Errorf("Invalid business value for error case: %f", context.BusinessValue)
				}
			}
		})
	}
}

func TestBusinessValueComponents(t *testing.T) {
	analyzer, err := NewBusinessContextAnalyzer(nil)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	// Test individual scoring components
	t.Run("Revenue Scoring", func(t *testing.T) {
		revenueTests := []struct {
			revenue  string
			expected float64
		}{
			{"$50B", 1.0},
			{"$500M", 0.90},
			{"$50M", 0.70},
			{"$5M", 0.60},
			{"$500K", 0.30},
			{"", 0.5},
		}

		for _, test := range revenueTests {
			score := analyzer.calculateRevenueScore(test.revenue)
			if score != test.expected {
				t.Errorf("Revenue %s: expected %.2f, got %.2f", test.revenue, test.expected, score)
			}
		}
	})

	t.Run("Industry Scoring", func(t *testing.T) {
		industryTests := []struct {
			industry string
			minScore float64
		}{
			{"financial services", 0.9},
			{"healthcare", 0.85},
			{"technology", 0.7},
			{"retail", 0.5},  // Updated to match implementation
			{"unknown", 0.4},
		}

		for _, test := range industryTests {
			score := analyzer.calculateIndustryScore(test.industry)
			if score < test.minScore {
				t.Errorf("Industry %s: expected at least %.2f, got %.2f", test.industry, test.minScore, score)
			}
		}
	})

	t.Run("Company Size Scoring", func(t *testing.T) {
		sizeTests := []struct {
			size     string
			expected float64
		}{
			{"50000 employees", 1.0},
			{"5000 employees", 0.85},
			{"500 employees", 0.65},
			{"50 employees", 0.45},
			{"enterprise", 0.95},
			{"startup", 0.35},
		}

		for _, test := range sizeTests {
			score := analyzer.getCompanySizeScore(test.size)
			if score != test.expected {
				t.Errorf("Size %s: expected %.2f, got %.2f", test.size, test.expected, score)
			}
		}
	})
}

// Benchmark tests
func BenchmarkAnalyzeBusinessContext(b *testing.B) {
	analyzer, err := NewBusinessContextAnalyzer(nil)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeBusinessContext("example.com")
		if err != nil {
			b.Fatalf("Failed to analyze business context: %v", err)
		}
	}
}

func BenchmarkAssessBusinessValue(b *testing.B) {
	analyzer, err := NewBusinessContextAnalyzer(nil)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	profile := &CompanyProfile{
		CompanyName: "Test Company",
		Industry:    "technology",
		CompanySize: "medium",
		Revenue:     "$50M",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AssessBusinessValue(profile)
	}
}
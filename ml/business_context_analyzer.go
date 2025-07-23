package ml

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dablon/osmedeus/database"
	"github.com/dablon/osmedeus/utils"
)

// BusinessContextAnalyzer implements OSINT-based business context analysis
type BusinessContextAnalyzer struct {
	config          *BusinessAnalyzerConfig
	httpClient      *http.Client
	rateLimiter     *RateLimiter
	scraperModules  map[string]OSINTScraper
	companyDatabase *CompanyProfileDatabase
}

// BusinessAnalyzerConfig holds configuration for business context analysis
type BusinessAnalyzerConfig struct {
	RequestTimeout      time.Duration          `json:"request_timeout"`
	RateLimitDelay      time.Duration          `json:"rate_limit_delay"`
	MaxRetries          int                    `json:"max_retries"`
	UserAgent           string                 `json:"user_agent"`
	EnabledSources      []string               `json:"enabled_sources"`
	BusinessValueWeights map[string]float64    `json:"business_value_weights"`
	IndustryRiskFactors map[string]float64     `json:"industry_risk_factors"`
	CompanySizeMultipliers map[string]float64  `json:"company_size_multipliers"`
	DataFreshnessThreshold time.Duration       `json:"data_freshness_threshold"`
	CacheEnabled        bool                   `json:"cache_enabled"`
	EthicalScrapingRules map[string]interface{} `json:"ethical_scraping_rules"`
}

// CompanyProfile represents comprehensive company information
type CompanyProfile struct {
	CompanyName     string                 `json:"company_name"`
	Domain          string                 `json:"domain"`
	Industry        string                 `json:"industry"`
	CompanySize     string                 `json:"company_size"`
	Revenue         string                 `json:"revenue"`
	Founded         string                 `json:"founded"`
	Headquarters    string                 `json:"headquarters"`
	Description     string                 `json:"description"`
	Technologies    []string               `json:"technologies"`
	SocialProfiles  map[string]string      `json:"social_profiles"`
	KeyPersonnel    []PersonProfile        `json:"key_personnel"`
	FundingInfo     *FundingInformation    `json:"funding_info,omitempty"`
	PublicTrading   *PublicTradingInfo     `json:"public_trading,omitempty"`
	SecurityPosture *SecurityPostureInfo   `json:"security_posture,omitempty"`
	DataSources     []string               `json:"data_sources"`
	LastUpdated     time.Time              `json:"last_updated"`
	Confidence      float64                `json:"confidence"`
}

// PersonProfile represents key personnel information
type PersonProfile struct {
	Name        string `json:"name"`
	Position    string `json:"position"`
	LinkedInURL string `json:"linkedin_url,omitempty"`
	Email       string `json:"email,omitempty"`
}

// FundingInformation represents company funding details
type FundingInformation struct {
	TotalFunding    string    `json:"total_funding"`
	LastRound       string    `json:"last_round"`
	LastRoundDate   time.Time `json:"last_round_date"`
	Investors       []string  `json:"investors"`
	Valuation       string    `json:"valuation,omitempty"`
}

// PublicTradingInfo represents public company trading information
type PublicTradingInfo struct {
	TickerSymbol   string  `json:"ticker_symbol"`
	Exchange       string  `json:"exchange"`
	MarketCap      string  `json:"market_cap"`
	StockPrice     float64 `json:"stock_price"`
	PERatio        float64 `json:"pe_ratio"`
	SECFilings     []string `json:"sec_filings"`
}

// SecurityPostureInfo represents security-related company information
type SecurityPostureInfo struct {
	SecurityCertifications []string `json:"security_certifications"`
	ComplianceFrameworks   []string `json:"compliance_frameworks"`
	SecurityIncidents      []string `json:"security_incidents"`
	BugBountyProgram       bool     `json:"bug_bounty_program"`
	SecurityContacts       []string `json:"security_contacts"`
}

// OSINTScraper interface for different data sources
type OSINTScraper interface {
	ScrapeCompanyInfo(domain string) (*CompanyProfile, error)
	GetSourceName() string
	IsRateLimited() bool
	GetLastRequestTime() time.Time
	SetRateLimit(delay time.Duration)
}

// RateLimiter manages request rate limiting for ethical scraping
type RateLimiter struct {
	delays         map[string]time.Duration
	lastRequest    map[string]time.Time
	requestCounts  map[string]int
	windowStart    map[string]time.Time
	maxRequests    int
	windowDuration time.Duration
	mutex          sync.RWMutex
}

// CompanyProfileDatabase manages company profile storage and retrieval
type CompanyProfileDatabase struct {
	profiles map[string]*CompanyProfile
	mutex    *sync.RWMutex
}

// LinkedInScraper implements LinkedIn company information scraping
type LinkedInScraper struct {
	config      *ScraperConfig
	rateLimiter *RateLimiter
	httpClient  *http.Client
}

// CrunchbaseScraper implements Crunchbase company information scraping
type CrunchbaseScraper struct {
	config      *ScraperConfig
	rateLimiter *RateLimiter
	httpClient  *http.Client
	apiKey      string
}

// CompanyWebsiteScraper implements company website information extraction
type CompanyWebsiteScraper struct {
	config      *ScraperConfig
	rateLimiter *RateLimiter
	httpClient  *http.Client
}

// SECFilingsScraper implements SEC filings information extraction
type SECFilingsScraper struct {
	config      *ScraperConfig
	rateLimiter *RateLimiter
	httpClient  *http.Client
}

// ScraperConfig holds configuration for individual scrapers
type ScraperConfig struct {
	RequestTimeout time.Duration `json:"request_timeout"`
	MaxRetries     int           `json:"max_retries"`
	UserAgent      string        `json:"user_agent"`
	RateLimit      time.Duration `json:"rate_limit"`
	Enabled        bool          `json:"enabled"`
}

// NewBusinessContextAnalyzer creates a new business context analyzer
func NewBusinessContextAnalyzer(config *BusinessAnalyzerConfig) (*BusinessContextAnalyzer, error) {
	if config == nil {
		config = &BusinessAnalyzerConfig{
			RequestTimeout:      30 * time.Second,
			RateLimitDelay:      2 * time.Second,
			MaxRetries:          3,
			UserAgent:           "Osmedeus-BusinessAnalyzer/1.0",
			EnabledSources:      []string{"linkedin", "crunchbase", "website", "sec"},
			BusinessValueWeights: map[string]float64{
				"revenue_high":     1.0,
				"revenue_medium":   0.7,
				"revenue_low":      0.4,
				"public_company":   0.9,
				"private_funded":   0.8,
				"startup":          0.6,
				"enterprise":       0.9,
				"sme":              0.6,
				"small":            0.3,
			},
			IndustryRiskFactors: map[string]float64{
				"financial":    1.0,
				"healthcare":   0.95,
				"government":   0.9,
				"technology":   0.85,
				"retail":       0.8,
				"manufacturing": 0.7,
				"education":    0.6,
				"nonprofit":    0.4,
			},
			CompanySizeMultipliers: map[string]float64{
				"enterprise": 1.0,
				"large":      0.8,
				"medium":     0.6,
				"small":      0.4,
				"startup":    0.3,
			},
			DataFreshnessThreshold: 7 * 24 * time.Hour, // 7 days
			CacheEnabled:          true,
			EthicalScrapingRules: map[string]interface{}{
				"respect_robots_txt": true,
				"max_requests_per_minute": 10,
				"user_agent_rotation": false,
				"proxy_rotation": false,
			},
		}
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: config.RequestTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
		},
	}

	// Initialize rate limiter with ethical scraping settings
	maxRequestsPerMinute := 10
	if maxReq, ok := config.EthicalScrapingRules["max_requests_per_minute"].(int); ok {
		maxRequestsPerMinute = maxReq
	}
	
	rateLimiter := &RateLimiter{
		delays:         make(map[string]time.Duration),
		lastRequest:    make(map[string]time.Time),
		requestCounts:  make(map[string]int),
		windowStart:    make(map[string]time.Time),
		maxRequests:    maxRequestsPerMinute,
		windowDuration: time.Minute,
		mutex:          sync.RWMutex{},
	}

	// Initialize company database
	companyDatabase := &CompanyProfileDatabase{
		profiles: make(map[string]*CompanyProfile),
		mutex:    &sync.RWMutex{},
	}

	analyzer := &BusinessContextAnalyzer{
		config:          config,
		httpClient:      httpClient,
		rateLimiter:     rateLimiter,
		scraperModules:  make(map[string]OSINTScraper),
		companyDatabase: companyDatabase,
	}

	// Initialize scraper modules
	if err := analyzer.initializeScrapers(); err != nil {
		return nil, fmt.Errorf("failed to initialize scrapers: %w", err)
	}

	return analyzer, nil
}

// initializeScrapers initializes all enabled scraper modules
func (bca *BusinessContextAnalyzer) initializeScrapers() error {
	scraperConfig := &ScraperConfig{
		RequestTimeout: bca.config.RequestTimeout,
		MaxRetries:     bca.config.MaxRetries,
		UserAgent:      bca.config.UserAgent,
		RateLimit:      bca.config.RateLimitDelay,
		Enabled:        true,
	}

	// Initialize LinkedIn scraper
	if contains(bca.config.EnabledSources, "linkedin") {
		linkedinScraper := &LinkedInScraper{
			config:      scraperConfig,
			rateLimiter: bca.rateLimiter,
			httpClient:  bca.httpClient,
		}
		bca.scraperModules["linkedin"] = linkedinScraper
	}

	// Initialize Crunchbase scraper
	if contains(bca.config.EnabledSources, "crunchbase") {
		crunchbaseScraper := &CrunchbaseScraper{
			config:      scraperConfig,
			rateLimiter: bca.rateLimiter,
			httpClient:  bca.httpClient,
			apiKey:      "", // Would be configured from environment
		}
		bca.scraperModules["crunchbase"] = crunchbaseScraper
	}

	// Initialize website scraper
	if contains(bca.config.EnabledSources, "website") {
		websiteScraper := &CompanyWebsiteScraper{
			config:      scraperConfig,
			rateLimiter: bca.rateLimiter,
			httpClient:  bca.httpClient,
		}
		bca.scraperModules["website"] = websiteScraper
	}

	// Initialize SEC filings scraper
	if contains(bca.config.EnabledSources, "sec") {
		secScraper := &SECFilingsScraper{
			config:      scraperConfig,
			rateLimiter: bca.rateLimiter,
			httpClient:  bca.httpClient,
		}
		bca.scraperModules["sec"] = secScraper
	}

	return nil
}

// AnalyzeBusinessContext performs comprehensive business context analysis
func (bca *BusinessContextAnalyzer) AnalyzeBusinessContext(domain string) (*database.BusinessContext, error) {
	utils.InforF("Starting business context analysis for domain: %s", domain)

	// Check cache first
	if bca.config.CacheEnabled {
		if cachedProfile := bca.getCachedProfile(domain); cachedProfile != nil {
			if time.Since(cachedProfile.LastUpdated) < bca.config.DataFreshnessThreshold {
				utils.InforF("Using cached business context for: %s", domain)
				return bca.convertToBusinessContext(cachedProfile), nil
			}
		}
	}

	// Get company profile from multiple sources
	companyProfile, err := bca.GetCompanyProfile(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get company profile: %w", err)
	}

	// Calculate business value
	businessValue := bca.AssessBusinessValue(companyProfile)

	// Convert to database format
	businessContext := bca.convertToBusinessContext(companyProfile)
	businessContext.BusinessValue = businessValue

	// Cache the result
	if bca.config.CacheEnabled {
		bca.cacheProfile(domain, companyProfile)
	}

	utils.InforF("Business context analysis completed for %s (value: %.2f)", domain, businessValue)
	return businessContext, nil
}

// GetCompanyProfile aggregates company information from multiple OSINT sources
func (bca *BusinessContextAnalyzer) GetCompanyProfile(domain string) (*CompanyProfile, error) {
	profile := &CompanyProfile{
		Domain:         domain,
		DataSources:    make([]string, 0),
		SocialProfiles: make(map[string]string),
		Technologies:   make([]string, 0),
		KeyPersonnel:   make([]PersonProfile, 0),
		LastUpdated:    time.Now(),
		Confidence:     0.0,
	}

	var totalConfidence float64
	var sourceCount int

	// Scrape information from each enabled source
	for sourceName, scraper := range bca.scraperModules {
		utils.InforF("Scraping %s for domain: %s", sourceName, domain)

		// Apply rate limiting
		if err := bca.applyRateLimit(sourceName); err != nil {
			utils.ErrorF("Rate limit error for %s: %v", sourceName, err)
			continue
		}

		// Scrape company information
		sourceProfile, err := scraper.ScrapeCompanyInfo(domain)
		if err != nil {
			utils.ErrorF("Failed to scrape %s: %v", sourceName, err)
			continue
		}

		if sourceProfile != nil {
			// Merge profile information
			bca.mergeProfiles(profile, sourceProfile)
			profile.DataSources = append(profile.DataSources, sourceName)
			totalConfidence += sourceProfile.Confidence
			sourceCount++
		}

		// Update rate limiter
		bca.rateLimiter.lastRequest[sourceName] = time.Now()
	}

	// Calculate overall confidence
	if sourceCount > 0 {
		profile.Confidence = totalConfidence / float64(sourceCount)
	}

	// Validate and enrich profile
	bca.validateAndEnrichProfile(profile)

	return profile, nil
}

// AssessBusinessValue calculates business value score based on company metrics
func (bca *BusinessContextAnalyzer) AssessBusinessValue(profile *CompanyProfile) float64 {
	if profile == nil {
		return 0.0
	}

	// Start with base score components
	var scoreComponents = make(map[string]float64)
	
	// Revenue-based scoring (30% weight)
	revenueScore := bca.calculateRevenueScore(profile.Revenue)
	scoreComponents["revenue"] = revenueScore * 0.30

	// Industry risk factor (25% weight)
	industryScore := bca.calculateIndustryScore(profile.Industry)
	scoreComponents["industry"] = industryScore * 0.25

	// Company size and maturity (20% weight)
	sizeScore := bca.calculateSizeScore(profile.CompanySize, profile.Founded)
	scoreComponents["size"] = sizeScore * 0.20

	// Market presence and visibility (15% weight)
	marketScore := bca.calculateMarketPresenceScore(profile)
	scoreComponents["market"] = marketScore * 0.15

	// Security and compliance posture (10% weight)
	securityScore := bca.calculateSecurityScore(profile.SecurityPosture)
	scoreComponents["security"] = securityScore * 0.10

	// Calculate weighted total
	var totalScore float64
	for _, score := range scoreComponents {
		totalScore += score
	}

	// Apply multipliers for special characteristics
	totalScore = bca.applyBusinessValueMultipliers(totalScore, profile)

	// Normalize to 0-1 range
	if totalScore > 1.0 {
		totalScore = 1.0
	} else if totalScore < 0.0 {
		totalScore = 0.0
	}

	utils.InforF("Business value assessment for %s: %.3f (components: %+v)", 
		profile.Domain, totalScore, scoreComponents)

	return totalScore
}

// calculateRevenueScore assesses company value based on revenue
func (bca *BusinessContextAnalyzer) calculateRevenueScore(revenue string) float64 {
	if revenue == "" {
		return 0.5 // Default for unknown revenue
	}

	revenueValue := bca.extractRevenueValue(revenue)
	
	// Logarithmic scoring for revenue (handles wide range better)
	switch {
	case revenueValue >= 10000000000: // $10B+
		return 1.0
	case revenueValue >= 1000000000: // $1B+
		return 0.95
	case revenueValue >= 500000000: // $500M+
		return 0.90
	case revenueValue >= 100000000: // $100M+
		return 0.80
	case revenueValue >= 50000000: // $50M+
		return 0.70
	case revenueValue >= 10000000: // $10M+
		return 0.60
	case revenueValue >= 5000000: // $5M+
		return 0.60
	case revenueValue >= 1000000: // $1M+
		return 0.45
	default:
		return 0.30
	}
}

// calculateIndustryScore assesses risk and value based on industry
func (bca *BusinessContextAnalyzer) calculateIndustryScore(industry string) float64 {
	if industry == "" {
		return 0.5 // Default for unknown industry
	}

	industry = strings.ToLower(industry)
	
	// High-value industries (attractive targets)
	highValueIndustries := map[string]float64{
		"financial":     1.0,
		"banking":       1.0,
		"fintech":       0.95,
		"healthcare":    0.90,
		"government":    0.95,
		"defense":       0.90,
		"cryptocurrency": 0.95,
		"payment":       0.90,
		"insurance":     0.85,
	}
	
	// Medium-value industries
	mediumValueIndustries := map[string]float64{
		"technology":    0.80,
		"software":      0.75,
		"saas":          0.75,
		"cloud":         0.80,
		"cybersecurity": 0.85,
		"telecommunications": 0.70,
		"energy":        0.75,
		"retail":        0.50,  // Reduced from 0.65
		"ecommerce":     0.60,  // Reduced from 0.70
	}
	
	// Check high-value industries first
	for key, score := range highValueIndustries {
		if strings.Contains(industry, key) {
			return score
		}
	}
	
	// Check medium-value industries
	for key, score := range mediumValueIndustries {
		if strings.Contains(industry, key) {
			return score
		}
	}
	
	// Default for other industries
	return 0.50
}

// calculateSizeScore assesses company based on size and maturity
func (bca *BusinessContextAnalyzer) calculateSizeScore(companySize, founded string) float64 {
	sizeScore := bca.getCompanySizeScore(companySize)
	maturityScore := bca.getCompanyMaturityScore(founded)
	
	// Weighted combination (70% size, 30% maturity)
	return (sizeScore * 0.7) + (maturityScore * 0.3)
}

// getCompanySizeScore calculates score based on company size
func (bca *BusinessContextAnalyzer) getCompanySizeScore(size string) float64 {
	if size == "" {
		return 0.5
	}

	size = strings.ToLower(size)
	
	// Extract employee count if present
	re := regexp.MustCompile(`(\d+(?:,\d+)*)`)
	matches := re.FindAllString(size, -1)
	
	if len(matches) > 0 {
		// Take the higher number if it's a range
		var maxCount int
		for _, match := range matches {
			countStr := strings.ReplaceAll(match, ",", "")
			if count, err := strconv.Atoi(countStr); err == nil && count > maxCount {
				maxCount = count
			}
		}
		
		switch {
		case maxCount >= 50000: // Enterprise (50k+)
			return 1.0
		case maxCount >= 10000: // Large enterprise (10k+)
			return 0.95
		case maxCount >= 5000: // Large (5k+)
			return 0.85
		case maxCount >= 1000: // Medium-large (1k+)
			return 0.75
		case maxCount >= 500: // Medium (500+)
			return 0.65
		case maxCount >= 100: // Small-medium (100+)
			return 0.55
		case maxCount >= 50: // Small (50+)
			return 0.45
		default: // Very small
			return 0.35
		}
	}
	
	// Fallback to text-based classification
	if strings.Contains(size, "enterprise") || strings.Contains(size, "fortune") {
		return 0.95
	} else if strings.Contains(size, "large") {
		return 0.80
	} else if strings.Contains(size, "medium") || strings.Contains(size, "mid") {
		return 0.65
	} else if strings.Contains(size, "small") {
		return 0.45
	} else if strings.Contains(size, "startup") {
		return 0.35
	}
	
	return 0.50
}

// getCompanyMaturityScore calculates score based on company age
func (bca *BusinessContextAnalyzer) getCompanyMaturityScore(founded string) float64 {
	if founded == "" {
		return 0.5
	}
	
	// Extract year from founded string
	re := regexp.MustCompile(`(\d{4})`)
	matches := re.FindStringSubmatch(founded)
	
	if len(matches) > 1 {
		if foundedYear, err := strconv.Atoi(matches[1]); err == nil {
			currentYear := time.Now().Year()
			age := currentYear - foundedYear
			
			switch {
			case age >= 50: // Very established (50+ years)
				return 0.95
			case age >= 25: // Well established (25+ years)
				return 0.85
			case age >= 10: // Established (10+ years)
				return 0.75
			case age >= 5: // Mature (5+ years)
				return 0.65
			case age >= 2: // Growing (2+ years)
				return 0.55
			default: // Very new
				return 0.40
			}
		}
	}
	
	return 0.50
}

// calculateMarketPresenceScore assesses company's market visibility and presence
func (bca *BusinessContextAnalyzer) calculateMarketPresenceScore(profile *CompanyProfile) float64 {
	var score float64 = 0.3 // Base score
	
	// Public trading adds significant value
	if profile.PublicTrading != nil {
		score += 0.4
		
		// Market cap considerations
		if profile.PublicTrading.MarketCap != "" {
			marketCapValue := bca.extractMarketCapValue(profile.PublicTrading.MarketCap)
			switch {
			case marketCapValue >= 100000000000: // $100B+ (mega cap)
				score += 0.3
			case marketCapValue >= 10000000000: // $10B+ (large cap)
				score += 0.25
			case marketCapValue >= 2000000000: // $2B+ (mid cap)
				score += 0.2
			case marketCapValue >= 300000000: // $300M+ (small cap)
				score += 0.15
			default: // Micro cap
				score += 0.1
			}
		}
	}
	
	// Funding information adds value for private companies
	if profile.FundingInfo != nil && profile.FundingInfo.TotalFunding != "" {
		fundingValue := bca.extractFundingValue(profile.FundingInfo.TotalFunding)
		switch {
		case fundingValue >= 1000000000: // $1B+ (unicorn)
			score += 0.35
		case fundingValue >= 100000000: // $100M+ (well-funded)
			score += 0.25
		case fundingValue >= 10000000: // $10M+ (funded)
			score += 0.15
		default:
			score += 0.05
		}
	}
	
	// Social media presence
	socialPresenceScore := float64(len(profile.SocialProfiles)) * 0.02
	if socialPresenceScore > 0.1 {
		socialPresenceScore = 0.1 // Cap at 0.1
	}
	score += socialPresenceScore
	
	// Normalize to 0-1 range
	if score > 1.0 {
		score = 1.0
	}
	
	return score
}

// calculateSecurityScore assesses company's security posture
func (bca *BusinessContextAnalyzer) calculateSecurityScore(securityPosture *SecurityPostureInfo) float64 {
	if securityPosture == nil {
		return 0.5 // Neutral score for unknown security posture
	}
	
	var score float64 = 0.5 // Base score
	
	// Bug bounty program indicates security maturity
	if securityPosture.BugBountyProgram {
		score += 0.2
	}
	
	// Security certifications add value
	certCount := len(securityPosture.SecurityCertifications)
	certScore := float64(certCount) * 0.05
	if certScore > 0.2 {
		certScore = 0.2 // Cap at 0.2
	}
	score += certScore
	
	// Compliance frameworks
	complianceCount := len(securityPosture.ComplianceFrameworks)
	complianceScore := float64(complianceCount) * 0.03
	if complianceScore > 0.15 {
		complianceScore = 0.15 // Cap at 0.15
	}
	score += complianceScore
	
	// Security incidents reduce score
	incidentCount := len(securityPosture.SecurityIncidents)
	incidentPenalty := float64(incidentCount) * 0.1
	if incidentPenalty > 0.3 {
		incidentPenalty = 0.3 // Cap penalty at 0.3
	}
	score -= incidentPenalty
	
	// Normalize to 0-1 range
	if score > 1.0 {
		score = 1.0
	} else if score < 0.0 {
		score = 0.0
	}
	
	return score
}

// applyBusinessValueMultipliers applies final multipliers based on special characteristics
func (bca *BusinessContextAnalyzer) applyBusinessValueMultipliers(baseScore float64, profile *CompanyProfile) float64 {
	score := baseScore
	
	// High-profile company multiplier
	if bca.isHighProfileCompany(profile) {
		score *= 1.1
	}
	
	// Technology stack multiplier
	if bca.hasHighValueTechnologies(profile.Technologies) {
		score *= 1.05
	}
	
	// Geographic multiplier (if headquarters in high-value regions)
	if bca.isHighValueLocation(profile.Headquarters) {
		score *= 1.03
	}
	
	return score
}

// Helper methods for business value calculation
func (bca *BusinessContextAnalyzer) extractMarketCapValue(marketCap string) float64 {
	if marketCap == "" {
		return 0
	}
	
	// Similar to extractRevenueValue but for market cap
	return bca.extractRevenueValue(marketCap)
}

func (bca *BusinessContextAnalyzer) extractFundingValue(funding string) float64 {
	if funding == "" {
		return 0
	}
	
	// Similar to extractRevenueValue but for funding
	return bca.extractRevenueValue(funding)
}

func (bca *BusinessContextAnalyzer) isHighProfileCompany(profile *CompanyProfile) bool {
	if profile == nil {
		return false
	}
	
	// Check if it's a well-known company (Fortune 500, unicorn, etc.)
	companyName := strings.ToLower(profile.CompanyName)
	
	// High-profile indicators
	highProfileIndicators := []string{
		"google", "microsoft", "apple", "amazon", "facebook", "meta",
		"tesla", "netflix", "uber", "airbnb", "spotify", "zoom",
		"salesforce", "oracle", "ibm", "intel", "nvidia", "adobe",
	}
	
	for _, indicator := range highProfileIndicators {
		if strings.Contains(companyName, indicator) {
			return true
		}
	}
	
	// Check if it's publicly traded (generally higher profile)
	return profile.PublicTrading != nil
}

func (bca *BusinessContextAnalyzer) hasHighValueTechnologies(technologies []string) bool {
	if len(technologies) == 0 {
		return false
	}
	
	highValueTechs := []string{
		"ai", "machine learning", "blockchain", "cryptocurrency",
		"cloud", "kubernetes", "microservices", "fintech",
		"cybersecurity", "quantum", "iot", "5g",
	}
	
	for _, tech := range technologies {
		techLower := strings.ToLower(tech)
		for _, highValueTech := range highValueTechs {
			if strings.Contains(techLower, highValueTech) {
				return true
			}
		}
	}
	
	return false
}

func (bca *BusinessContextAnalyzer) isHighValueLocation(headquarters string) bool {
	if headquarters == "" {
		return false
	}
	
	hq := strings.ToLower(headquarters)
	
	// High-value business locations
	highValueLocations := []string{
		"silicon valley", "san francisco", "palo alto", "mountain view",
		"new york", "manhattan", "wall street", "london", "tokyo",
		"singapore", "hong kong", "zurich", "geneva", "dubai",
		"seattle", "boston", "austin", "los angeles",
	}
	
	for _, location := range highValueLocations {
		if strings.Contains(hq, location) {
			return true
		}
	}
	
	return false
}

// applyRateLimit applies comprehensive rate limiting for ethical scraping
func (bca *BusinessContextAnalyzer) applyRateLimit(sourceName string) error {
	bca.rateLimiter.mutex.Lock()
	defer bca.rateLimiter.mutex.Unlock()

	now := time.Now()

	// Check per-request rate limiting
	if lastRequest, exists := bca.rateLimiter.lastRequest[sourceName]; exists {
		elapsed := time.Since(lastRequest)
		if elapsed < bca.config.RateLimitDelay {
			sleepTime := bca.config.RateLimitDelay - elapsed
			utils.InforF("Per-request rate limiting %s: sleeping for %v", sourceName, sleepTime)
			bca.rateLimiter.mutex.Unlock()
			time.Sleep(sleepTime)
			bca.rateLimiter.mutex.Lock()
		}
	}

	// Check window-based rate limiting (requests per minute)
	windowStart, exists := bca.rateLimiter.windowStart[sourceName]
	if !exists || now.Sub(windowStart) >= bca.rateLimiter.windowDuration {
		// Reset window
		bca.rateLimiter.windowStart[sourceName] = now
		bca.rateLimiter.requestCounts[sourceName] = 0
	}

	// Check if we've exceeded the request limit for this window
	if bca.rateLimiter.requestCounts[sourceName] >= bca.rateLimiter.maxRequests {
		// Calculate time until window resets
		timeUntilReset := bca.rateLimiter.windowDuration - now.Sub(bca.rateLimiter.windowStart[sourceName])
		if timeUntilReset > 0 {
			utils.InforF("Window rate limiting %s: sleeping for %v (requests: %d/%d)", 
				sourceName, timeUntilReset, bca.rateLimiter.requestCounts[sourceName], bca.rateLimiter.maxRequests)
			bca.rateLimiter.mutex.Unlock()
			time.Sleep(timeUntilReset)
			bca.rateLimiter.mutex.Lock()
			
			// Reset window after sleep
			bca.rateLimiter.windowStart[sourceName] = time.Now()
			bca.rateLimiter.requestCounts[sourceName] = 0
		}
	}

	// Increment request count
	bca.rateLimiter.requestCounts[sourceName]++
	bca.rateLimiter.lastRequest[sourceName] = time.Now()

	return nil
}

// mergeProfiles merges information from multiple sources into a single profile
func (bca *BusinessContextAnalyzer) mergeProfiles(target, source *CompanyProfile) {
	if source.CompanyName != "" && target.CompanyName == "" {
		target.CompanyName = source.CompanyName
	}
	if source.Industry != "" && target.Industry == "" {
		target.Industry = source.Industry
	}
	if source.CompanySize != "" && target.CompanySize == "" {
		target.CompanySize = source.CompanySize
	}
	if source.Revenue != "" && target.Revenue == "" {
		target.Revenue = source.Revenue
	}
	if source.Founded != "" && target.Founded == "" {
		target.Founded = source.Founded
	}
	if source.Headquarters != "" && target.Headquarters == "" {
		target.Headquarters = source.Headquarters
	}
	if source.Description != "" && target.Description == "" {
		target.Description = source.Description
	}

	// Merge arrays and maps
	target.Technologies = mergeStringSlices(target.Technologies, source.Technologies)
	target.KeyPersonnel = append(target.KeyPersonnel, source.KeyPersonnel...)
	
	for k, v := range source.SocialProfiles {
		if _, exists := target.SocialProfiles[k]; !exists {
			target.SocialProfiles[k] = v
		}
	}

	// Merge complex objects
	if source.FundingInfo != nil && target.FundingInfo == nil {
		target.FundingInfo = source.FundingInfo
	}
	if source.PublicTrading != nil && target.PublicTrading == nil {
		target.PublicTrading = source.PublicTrading
	}
	if source.SecurityPosture != nil && target.SecurityPosture == nil {
		target.SecurityPosture = source.SecurityPosture
	}
}

// validateAndEnrichProfile validates and enriches the company profile
func (bca *BusinessContextAnalyzer) validateAndEnrichProfile(profile *CompanyProfile) {
	// Normalize company name
	if profile.CompanyName != "" {
		profile.CompanyName = strings.TrimSpace(profile.CompanyName)
	}

	// Normalize industry
	profile.Industry = bca.normalizeIndustry(profile.Industry)

	// Normalize company size
	profile.CompanySize = bca.normalizeCompanySize(profile.CompanySize)

	// Extract domain from company name if missing
	if profile.CompanyName == "" && profile.Domain != "" {
		profile.CompanyName = bca.extractCompanyNameFromDomain(profile.Domain)
	}

	// Deduplicate arrays
	profile.Technologies = deduplicateStrings(profile.Technologies)
	profile.DataSources = deduplicateStrings(profile.DataSources)
}

// Helper methods for business value calculation
func (bca *BusinessContextAnalyzer) getRevenueMultiplier(revenue string) float64 {
	if revenue == "" {
		return 0.7 // Default for unknown revenue
	}

	revenue = strings.ToLower(revenue)
	
	// Extract numeric value from revenue string
	revenueValue := bca.extractRevenueValue(revenue)
	
	if revenueValue >= 1000000000 { // $1B+
		return bca.config.BusinessValueWeights["revenue_high"]
	} else if revenueValue >= 100000000 { // $100M+
		return bca.config.BusinessValueWeights["revenue_medium"]
	} else {
		return bca.config.BusinessValueWeights["revenue_low"]
	}
}

func (bca *BusinessContextAnalyzer) getIndustryRiskFactor(industry string) float64 {
	if industry == "" {
		return 0.7 // Default for unknown industry
	}

	industry = strings.ToLower(industry)
	
	for key, factor := range bca.config.IndustryRiskFactors {
		if strings.Contains(industry, key) {
			return factor
		}
	}
	
	return 0.7 // Default factor
}

func (bca *BusinessContextAnalyzer) getCompanySizeMultiplier(size string) float64 {
	if size == "" {
		return 0.7 // Default for unknown size
	}

	size = strings.ToLower(size)
	
	for key, multiplier := range bca.config.CompanySizeMultipliers {
		if strings.Contains(size, key) {
			return multiplier
		}
	}
	
	return 0.7 // Default multiplier
}

// convertToBusinessContext converts CompanyProfile to database.BusinessContext
func (bca *BusinessContextAnalyzer) convertToBusinessContext(profile *CompanyProfile) *database.BusinessContext {
	criticalityFactors := bca.extractCriticalityFactors(profile)
	riskFactors := bca.extractRiskFactors(profile)
	
	criticalityJSON, _ := json.Marshal(criticalityFactors)
	riskFactorsJSON, _ := json.Marshal(riskFactors)
	publicProfileJSON, _ := json.Marshal(profile)
	dataSourcesJSON, _ := json.Marshal(profile.DataSources)

	return &database.BusinessContext{
		TargetID:           profile.Domain,
		CompanyName:        profile.CompanyName,
		Industry:           profile.Industry,
		CompanySize:        profile.CompanySize,
		Revenue:            profile.Revenue,
		BusinessValue:      0.0, // Will be set by caller
		CriticalityFactors: string(criticalityJSON),
		PublicProfile:      string(publicProfileJSON),
		RiskFactors:        string(riskFactorsJSON),
		AnalyzedAt:         time.Now(),
		DataSources:        string(dataSourcesJSON),
	}
}

// Cache management methods
func (bca *BusinessContextAnalyzer) getCachedProfile(domain string) *CompanyProfile {
	bca.companyDatabase.mutex.RLock()
	defer bca.companyDatabase.mutex.RUnlock()
	
	return bca.companyDatabase.profiles[domain]
}

func (bca *BusinessContextAnalyzer) cacheProfile(domain string, profile *CompanyProfile) {
	bca.companyDatabase.mutex.Lock()
	defer bca.companyDatabase.mutex.Unlock()
	
	bca.companyDatabase.profiles[domain] = profile
}

// Utility functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func mergeStringSlices(slice1, slice2 []string) []string {
	merged := make([]string, len(slice1))
	copy(merged, slice1)
	
	for _, item := range slice2 {
		if !contains(merged, item) {
			merged = append(merged, item)
		}
	}
	
	return merged
}

func deduplicateStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

func (bca *BusinessContextAnalyzer) extractRevenueValue(revenue string) float64 {
	// Extract numeric value from revenue string
	re := regexp.MustCompile(`[\d,]+`)
	matches := re.FindAllString(revenue, -1)
	
	if len(matches) == 0 {
		return 0
	}
	
	numStr := strings.ReplaceAll(matches[0], ",", "")
	value, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}
	
	// Apply multipliers based on units
	revenue = strings.ToLower(revenue)
	if strings.Contains(revenue, "billion") || strings.Contains(revenue, "b") {
		value *= 1000000000
	} else if strings.Contains(revenue, "million") || strings.Contains(revenue, "m") {
		value *= 1000000
	} else if strings.Contains(revenue, "thousand") || strings.Contains(revenue, "k") {
		value *= 1000
	}
	
	return value
}

func (bca *BusinessContextAnalyzer) normalizeIndustry(industry string) string {
	if industry == "" {
		return "unknown"
	}
	
	industry = strings.ToLower(strings.TrimSpace(industry))
	
	// Map common industry variations to standard names
	industryMap := map[string]string{
		"fintech":     "financial",
		"banking":     "financial",
		"insurance":   "financial",
		"healthtech":  "healthcare",
		"medtech":     "healthcare",
		"pharma":      "healthcare",
		"edtech":      "education",
		"e-learning":  "education",
		"saas":        "technology",
		"software":    "technology",
		"it":          "technology",
		"ecommerce":   "retail",
		"e-commerce":  "retail",
	}
	
	for key, normalized := range industryMap {
		if strings.Contains(industry, key) {
			return normalized
		}
	}
	
	return industry
}

func (bca *BusinessContextAnalyzer) normalizeCompanySize(size string) string {
	if size == "" {
		return "unknown"
	}
	
	size = strings.ToLower(strings.TrimSpace(size))
	
	// Extract employee count if present
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindAllString(size, -1)
	
	if len(matches) > 0 {
		count, err := strconv.Atoi(matches[0])
		if err == nil {
			if count >= 10000 {
				return "enterprise"
			} else if count >= 1000 {
				return "large"
			} else if count >= 100 {
				return "medium"
			} else if count >= 10 {
				return "small"
			} else {
				return "startup"
			}
		}
	}
	
	// Map common size descriptions
	if strings.Contains(size, "enterprise") || strings.Contains(size, "fortune") {
		return "enterprise"
	} else if strings.Contains(size, "large") {
		return "large"
	} else if strings.Contains(size, "medium") || strings.Contains(size, "mid") {
		return "medium"
	} else if strings.Contains(size, "small") {
		return "small"
	} else if strings.Contains(size, "startup") || strings.Contains(size, "early") {
		return "startup"
	}
	
	return size
}

func (bca *BusinessContextAnalyzer) extractCompanyNameFromDomain(domain string) string {
	// Remove common prefixes and suffixes
	name := strings.TrimPrefix(domain, "www.")
	
	// Split by dots and take the main part
	parts := strings.Split(name, ".")
	if len(parts) > 0 {
		name = parts[0]
	}
	
	// Capitalize first letter
	if len(name) > 0 {
		name = strings.ToUpper(name[:1]) + name[1:]
	}
	
	return name
}

func (bca *BusinessContextAnalyzer) extractCriticalityFactors(profile *CompanyProfile) []string {
	factors := make([]string, 0)
	
	if profile.PublicTrading != nil {
		factors = append(factors, "publicly_traded")
	}
	
	if profile.FundingInfo != nil && profile.FundingInfo.TotalFunding != "" {
		factors = append(factors, "venture_funded")
	}
	
	if profile.SecurityPosture != nil && profile.SecurityPosture.BugBountyProgram {
		factors = append(factors, "bug_bounty_program")
	}
	
	// Industry-specific factors
	industry := strings.ToLower(profile.Industry)
	if strings.Contains(industry, "financial") {
		factors = append(factors, "financial_services")
	}
	if strings.Contains(industry, "healthcare") {
		factors = append(factors, "healthcare_data")
	}
	if strings.Contains(industry, "government") {
		factors = append(factors, "government_sector")
	}
	
	return factors
}

func (bca *BusinessContextAnalyzer) extractRiskFactors(profile *CompanyProfile) []string {
	factors := make([]string, 0)
	
	if profile.SecurityPosture != nil {
		if len(profile.SecurityPosture.SecurityIncidents) > 0 {
			factors = append(factors, "previous_security_incidents")
		}
		if len(profile.SecurityPosture.SecurityCertifications) == 0 {
			factors = append(factors, "no_security_certifications")
		}
	}
	
	// Industry-specific risk factors
	industry := strings.ToLower(profile.Industry)
	highRiskIndustries := []string{"financial", "healthcare", "government", "technology"}
	for _, riskIndustry := range highRiskIndustries {
		if strings.Contains(industry, riskIndustry) {
			factors = append(factors, fmt.Sprintf("high_risk_industry_%s", riskIndustry))
			break
		}
	}
	
	return factors
}
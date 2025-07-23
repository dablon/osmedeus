package ml

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/dablon/osmedeus/utils"
)

// LinkedIn Scraper Implementation
func (ls *LinkedInScraper) ScrapeCompanyInfo(domain string) (*CompanyProfile, error) {
	utils.InforF("Scraping LinkedIn for domain: %s", domain)
	
	profile := &CompanyProfile{
		Domain:         domain,
		DataSources:    []string{"linkedin"},
		SocialProfiles: make(map[string]string),
		Technologies:   make([]string, 0),
		KeyPersonnel:   make([]PersonProfile, 0),
		LastUpdated:    time.Now(),
		Confidence:     0.6, // Medium confidence for LinkedIn data
	}

	// Search for company on LinkedIn
	companyURL, err := ls.findCompanyLinkedInURL(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to find LinkedIn company URL: %w", err)
	}

	if companyURL == "" {
		utils.InforF("No LinkedIn company page found for domain: %s", domain)
		return profile, nil
	}

	profile.SocialProfiles["linkedin"] = companyURL

	// Scrape company page (simulated - real implementation would parse HTML)
	companyData, err := ls.scrapeCompanyPage(companyURL)
	if err != nil {
		utils.ErrorF("Failed to scrape LinkedIn company page: %v", err)
		return profile, nil
	}

	// Parse company information
	ls.parseLinkedInCompanyData(profile, companyData)

	return profile, nil
}

func (ls *LinkedInScraper) GetSourceName() string {
	return "linkedin"
}

func (ls *LinkedInScraper) IsRateLimited() bool {
	if lastRequest, exists := ls.rateLimiter.lastRequest["linkedin"]; exists {
		return time.Since(lastRequest) < ls.config.RateLimit
	}
	return false
}

func (ls *LinkedInScraper) GetLastRequestTime() time.Time {
	if lastRequest, exists := ls.rateLimiter.lastRequest["linkedin"]; exists {
		return lastRequest
	}
	return time.Time{}
}

func (ls *LinkedInScraper) SetRateLimit(delay time.Duration) {
	ls.config.RateLimit = delay
}

func (ls *LinkedInScraper) findCompanyLinkedInURL(domain string) (string, error) {
	// Simulate LinkedIn company search
	// In real implementation, this would search LinkedIn for the company
	// searchQuery := fmt.Sprintf("site:linkedin.com/company %s", domain)
	
	// For demonstration, return a simulated URL
	companyName := strings.Split(domain, ".")[0]
	linkedinURL := fmt.Sprintf("https://www.linkedin.com/company/%s", companyName)
	
	utils.InforF("Found LinkedIn URL: %s", linkedinURL)
	return linkedinURL, nil
}

func (ls *LinkedInScraper) scrapeCompanyPage(companyURL string) (map[string]interface{}, error) {
	// Apply rate limiting
	if ls.IsRateLimited() {
		sleepTime := ls.config.RateLimit - time.Since(ls.GetLastRequestTime())
		time.Sleep(sleepTime)
	}

	// Create request with enhanced headers for better success rate
	req, err := http.NewRequest("GET", companyURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set realistic browser headers
	req.Header.Set("User-Agent", ls.config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Make request with retry logic
	var resp *http.Response
	var lastErr error
	
	for attempt := 0; attempt < ls.config.MaxRetries; attempt++ {
		resp, lastErr = ls.httpClient.Do(req)
		if lastErr == nil && resp != nil && resp.StatusCode == http.StatusOK {
			break
		}
		
		if resp != nil {
			resp.Body.Close()
		}
		
		if attempt < ls.config.MaxRetries-1 {
			// Exponential backoff
			backoffTime := time.Duration(attempt+1) * time.Second
			utils.InforF("LinkedIn request failed (attempt %d/%d), retrying in %v", 
				attempt+1, ls.config.MaxRetries, backoffTime)
			time.Sleep(backoffTime)
		}
	}
	
	if lastErr != nil {
		return nil, fmt.Errorf("failed to make request after %d attempts: %w", ls.config.MaxRetries, lastErr)
	}
	
	if resp == nil {
		return nil, fmt.Errorf("received nil response")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Enhanced parsing of LinkedIn company data
	htmlContent := string(body)
	companyData := ls.parseLinkedInHTML(htmlContent, companyURL)

	return companyData, nil
}

// parseLinkedInHTML extracts company information from LinkedIn HTML
func (ls *LinkedInScraper) parseLinkedInHTML(htmlContent, companyURL string) map[string]interface{} {
	data := make(map[string]interface{})
	
	// Extract company name from URL as fallback
	data["company_name"] = extractCompanyNameFromURL(companyURL)
	
	// Extract company name from page title
	titleRegex := regexp.MustCompile(`<title[^>]*>([^|]+)`)
	if matches := titleRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		title := strings.TrimSpace(matches[1])
		if title != "" && !strings.Contains(strings.ToLower(title), "linkedin") {
			data["company_name"] = title
		}
	}
	
	// Extract industry information
	industryRegex := regexp.MustCompile(`"industry":"([^"]+)"`)
	if matches := industryRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		data["industry"] = matches[1]
	}
	
	// Extract company size
	sizeRegex := regexp.MustCompile(`"companySize":"([^"]+)"`)
	if matches := sizeRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		data["company_size"] = matches[1]
	}
	
	// Extract headquarters
	hqRegex := regexp.MustCompile(`"headquarters":"([^"]+)"`)
	if matches := hqRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		data["headquarters"] = matches[1]
	}
	
	// Extract founded year
	foundedRegex := regexp.MustCompile(`"founded":"([^"]+)"`)
	if matches := foundedRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		data["founded"] = matches[1]
	}
	
	// Extract description
	descRegex := regexp.MustCompile(`"description":"([^"]+)"`)
	if matches := descRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		data["description"] = matches[1]
	}
	
	// Extract employee count range
	employeeRegex := regexp.MustCompile(`(\d+(?:,\d+)*)\s*-\s*(\d+(?:,\d+)*)\s*employees`)
	if matches := employeeRegex.FindStringSubmatch(htmlContent); len(matches) > 2 {
		data["employee_range"] = fmt.Sprintf("%s-%s employees", matches[1], matches[2])
	}
	
	// Extract specialties
	specialtiesRegex := regexp.MustCompile(`"specialties":\[([^\]]+)\]`)
	if matches := specialtiesRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		specialtiesStr := matches[1]
		// Parse JSON array of specialties
		var specialties []string
		if err := json.Unmarshal([]byte("["+specialtiesStr+"]"), &specialties); err == nil {
			data["specialties"] = specialties
		}
	}
	
	return data
}

func (ls *LinkedInScraper) parseLinkedInCompanyData(profile *CompanyProfile, data map[string]interface{}) {
	if name, ok := data["company_name"].(string); ok {
		profile.CompanyName = name
	}
	if industry, ok := data["industry"].(string); ok {
		profile.Industry = industry
	}
	if size, ok := data["company_size"].(string); ok {
		profile.CompanySize = size
	}
	if desc, ok := data["description"].(string); ok {
		profile.Description = desc
	}
	if hq, ok := data["headquarters"].(string); ok {
		profile.Headquarters = hq
	}
	if founded, ok := data["founded"].(string); ok {
		profile.Founded = founded
	}
}

// Crunchbase Scraper Implementation
func (cs *CrunchbaseScraper) ScrapeCompanyInfo(domain string) (*CompanyProfile, error) {
	utils.InforF("Scraping Crunchbase for domain: %s", domain)
	
	profile := &CompanyProfile{
		Domain:         domain,
		DataSources:    []string{"crunchbase"},
		SocialProfiles: make(map[string]string),
		Technologies:   make([]string, 0),
		KeyPersonnel:   make([]PersonProfile, 0),
		LastUpdated:    time.Now(),
		Confidence:     0.8, // High confidence for Crunchbase data
	}

	// Search for company on Crunchbase
	companyData, err := cs.searchCrunchbaseCompany(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to search Crunchbase: %w", err)
	}

	if companyData == nil {
		utils.InforF("No Crunchbase data found for domain: %s", domain)
		return profile, nil
	}

	// Parse company information
	cs.parseCrunchbaseData(profile, companyData)

	return profile, nil
}

func (cs *CrunchbaseScraper) GetSourceName() string {
	return "crunchbase"
}

func (cs *CrunchbaseScraper) IsRateLimited() bool {
	if lastRequest, exists := cs.rateLimiter.lastRequest["crunchbase"]; exists {
		return time.Since(lastRequest) < cs.config.RateLimit
	}
	return false
}

func (cs *CrunchbaseScraper) GetLastRequestTime() time.Time {
	if lastRequest, exists := cs.rateLimiter.lastRequest["crunchbase"]; exists {
		return lastRequest
	}
	return time.Time{}
}

func (cs *CrunchbaseScraper) SetRateLimit(delay time.Duration) {
	cs.config.RateLimit = delay
}

func (cs *CrunchbaseScraper) searchCrunchbaseCompany(domain string) (map[string]interface{}, error) {
	// Apply rate limiting
	if cs.IsRateLimited() {
		sleepTime := cs.config.RateLimit - time.Since(cs.GetLastRequestTime())
		time.Sleep(sleepTime)
	}

	// Simulate Crunchbase API call
	// In real implementation, this would use the Crunchbase API
	companyName := strings.Split(domain, ".")[0]
	
	// Simulate API response
	companyData := map[string]interface{}{
		"name":         strings.Title(companyName),
		"description":  "Technology startup",
		"industry":     "Software",
		"founded_on":   "2015-01-01",
		"headquarters": "San Francisco, CA",
		"employee_count": "51-100",
		"funding_total": "$10M",
		"last_funding_type": "Series A",
		"investors": []string{"Acme Ventures", "Tech Capital"},
	}

	return companyData, nil
}

func (cs *CrunchbaseScraper) parseCrunchbaseData(profile *CompanyProfile, data map[string]interface{}) {
	if name, ok := data["name"].(string); ok {
		profile.CompanyName = name
	}
	if industry, ok := data["industry"].(string); ok {
		profile.Industry = industry
	}
	if desc, ok := data["description"].(string); ok {
		profile.Description = desc
	}
	if founded, ok := data["founded_on"].(string); ok {
		profile.Founded = founded
	}
	if hq, ok := data["headquarters"].(string); ok {
		profile.Headquarters = hq
	}
	if empCount, ok := data["employee_count"].(string); ok {
		profile.CompanySize = empCount
	}

	// Parse funding information
	if fundingTotal, ok := data["funding_total"].(string); ok {
		profile.FundingInfo = &FundingInformation{
			TotalFunding: fundingTotal,
		}
		
		if lastRound, ok := data["last_funding_type"].(string); ok {
			profile.FundingInfo.LastRound = lastRound
		}
		
		if investors, ok := data["investors"].([]string); ok {
			profile.FundingInfo.Investors = investors
		}
	}
}

// Company Website Scraper Implementation
func (cws *CompanyWebsiteScraper) ScrapeCompanyInfo(domain string) (*CompanyProfile, error) {
	utils.InforF("Scraping company website for domain: %s", domain)
	
	profile := &CompanyProfile{
		Domain:         domain,
		DataSources:    []string{"website"},
		SocialProfiles: make(map[string]string),
		Technologies:   make([]string, 0),
		KeyPersonnel:   make([]PersonProfile, 0),
		LastUpdated:    time.Now(),
		Confidence:     0.7, // Good confidence for website data
	}

	// Scrape main website
	websiteData, err := cws.scrapeWebsite(fmt.Sprintf("https://%s", domain))
	if err != nil {
		// Try with www prefix
		websiteData, err = cws.scrapeWebsite(fmt.Sprintf("https://www.%s", domain))
		if err != nil {
			return nil, fmt.Errorf("failed to scrape website: %w", err)
		}
	}

	// Parse website information
	cws.parseWebsiteData(profile, websiteData)

	// Try to find about page
	aboutData, err := cws.scrapeAboutPage(domain)
	if err == nil && aboutData != nil {
		cws.parseAboutPageData(profile, aboutData)
	}

	return profile, nil
}

func (cws *CompanyWebsiteScraper) GetSourceName() string {
	return "website"
}

func (cws *CompanyWebsiteScraper) IsRateLimited() bool {
	if lastRequest, exists := cws.rateLimiter.lastRequest["website"]; exists {
		return time.Since(lastRequest) < cws.config.RateLimit
	}
	return false
}

func (cws *CompanyWebsiteScraper) GetLastRequestTime() time.Time {
	if lastRequest, exists := cws.rateLimiter.lastRequest["website"]; exists {
		return lastRequest
	}
	return time.Time{}
}

func (cws *CompanyWebsiteScraper) SetRateLimit(delay time.Duration) {
	cws.config.RateLimit = delay
}

func (cws *CompanyWebsiteScraper) scrapeWebsite(websiteURL string) (map[string]interface{}, error) {
	// Apply rate limiting
	if cws.IsRateLimited() {
		sleepTime := cws.config.RateLimit - time.Since(cws.GetLastRequestTime())
		time.Sleep(sleepTime)
	}

	// Create request
	req, err := http.NewRequest("GET", websiteURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", cws.config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	// Make request
	resp, err := cws.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse HTML content
	htmlContent := string(body)
	websiteData := cws.extractWebsiteInfo(htmlContent)

	return websiteData, nil
}

func (cws *CompanyWebsiteScraper) scrapeAboutPage(domain string) (map[string]interface{}, error) {
	aboutURLs := []string{
		fmt.Sprintf("https://%s/about", domain),
		fmt.Sprintf("https://%s/about-us", domain),
		fmt.Sprintf("https://%s/company", domain),
		fmt.Sprintf("https://www.%s/about", domain),
		fmt.Sprintf("https://www.%s/about-us", domain),
	}

	for _, aboutURL := range aboutURLs {
		data, err := cws.scrapeWebsite(aboutURL)
		if err == nil {
			return data, nil
		}
	}

	return nil, fmt.Errorf("no about page found")
}

func (cws *CompanyWebsiteScraper) extractWebsiteInfo(htmlContent string) map[string]interface{} {
	data := make(map[string]interface{})

	// Extract title
	titleRegex := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	if matches := titleRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		data["title"] = strings.TrimSpace(matches[1])
	}

	// Extract meta description
	descRegex := regexp.MustCompile(`<meta[^>]*name=["']description["'][^>]*content=["']([^"']+)["']`)
	if matches := descRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		data["description"] = strings.TrimSpace(matches[1])
	}

	// Extract social media links
	socialLinks := make(map[string]string)
	
	// LinkedIn
	linkedinRegex := regexp.MustCompile(`href=["'](https?://[^"']*linkedin\.com/[^"']+)["']`)
	if matches := linkedinRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		socialLinks["linkedin"] = matches[1]
	}

	// Twitter
	twitterRegex := regexp.MustCompile(`href=["'](https?://[^"']*twitter\.com/[^"']+)["']`)
	if matches := twitterRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		socialLinks["twitter"] = matches[1]
	}

	// Facebook
	facebookRegex := regexp.MustCompile(`href=["'](https?://[^"']*facebook\.com/[^"']+)["']`)
	if matches := facebookRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		socialLinks["facebook"] = matches[1]
	}

	data["social_links"] = socialLinks

	// Extract contact information
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	if matches := emailRegex.FindAllString(htmlContent, -1); len(matches) > 0 {
		data["emails"] = matches
	}

	// Extract phone numbers
	phoneRegex := regexp.MustCompile(`\+?[\d\s\-\(\)]{10,}`)
	if matches := phoneRegex.FindAllString(htmlContent, -1); len(matches) > 0 {
		data["phones"] = matches
	}

	return data
}

func (cws *CompanyWebsiteScraper) parseWebsiteData(profile *CompanyProfile, data map[string]interface{}) {
	if title, ok := data["title"].(string); ok && profile.CompanyName == "" {
		// Extract company name from title
		profile.CompanyName = cws.extractCompanyNameFromTitle(title)
	}
	
	if desc, ok := data["description"].(string); ok && profile.Description == "" {
		profile.Description = desc
	}

	if socialLinks, ok := data["social_links"].(map[string]string); ok {
		for platform, link := range socialLinks {
			profile.SocialProfiles[platform] = link
		}
	}
}

func (cws *CompanyWebsiteScraper) parseAboutPageData(profile *CompanyProfile, data map[string]interface{}) {
	// Additional parsing for about page specific content
	if desc, ok := data["description"].(string); ok && len(desc) > len(profile.Description) {
		profile.Description = desc
	}
}

func (cws *CompanyWebsiteScraper) extractCompanyNameFromTitle(title string) string {
	// Remove common suffixes
	title = strings.TrimSpace(title)
	
	// Split by common separators
	separators := []string{" - ", " | ", " :: ", " — "}
	for _, sep := range separators {
		if parts := strings.Split(title, sep); len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	
	return title
}

// SEC Filings Scraper Implementation
func (sfs *SECFilingsScraper) ScrapeCompanyInfo(domain string) (*CompanyProfile, error) {
	utils.InforF("Scraping SEC filings for domain: %s", domain)
	
	profile := &CompanyProfile{
		Domain:         domain,
		DataSources:    []string{"sec"},
		SocialProfiles: make(map[string]string),
		Technologies:   make([]string, 0),
		KeyPersonnel:   make([]PersonProfile, 0),
		LastUpdated:    time.Now(),
		Confidence:     0.9, // Very high confidence for SEC data
	}

	// Search for company in SEC database
	companyData, err := sfs.searchSECDatabase(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to search SEC database: %w", err)
	}

	if companyData == nil {
		utils.InforF("No SEC filings found for domain: %s", domain)
		return profile, nil
	}

	// Parse SEC information
	sfs.parseSECData(profile, companyData)

	return profile, nil
}

func (sfs *SECFilingsScraper) GetSourceName() string {
	return "sec"
}

func (sfs *SECFilingsScraper) IsRateLimited() bool {
	if lastRequest, exists := sfs.rateLimiter.lastRequest["sec"]; exists {
		return time.Since(lastRequest) < sfs.config.RateLimit
	}
	return false
}

func (sfs *SECFilingsScraper) GetLastRequestTime() time.Time {
	if lastRequest, exists := sfs.rateLimiter.lastRequest["sec"]; exists {
		return lastRequest
	}
	return time.Time{}
}

func (sfs *SECFilingsScraper) SetRateLimit(delay time.Duration) {
	sfs.config.RateLimit = delay
}

func (sfs *SECFilingsScraper) searchSECDatabase(domain string) (map[string]interface{}, error) {
	// Apply rate limiting
	if sfs.IsRateLimited() {
		sleepTime := sfs.config.RateLimit - time.Since(sfs.GetLastRequestTime())
		time.Sleep(sleepTime)
	}

	// Simulate SEC EDGAR database search
	// In real implementation, this would search the SEC EDGAR database
	companyName := strings.Split(domain, ".")[0]
	
	// Check if this might be a public company (simplified check)
	if !sfs.isPotentialPublicCompany(companyName) {
		return nil, nil
	}

	// Simulate SEC data
	secData := map[string]interface{}{
		"company_name":   strings.Title(companyName) + " Inc.",
		"ticker_symbol":  strings.ToUpper(companyName[:3]),
		"exchange":       "NASDAQ",
		"market_cap":     "$1.2B",
		"industry":       "Technology",
		"employee_count": "5000",
		"revenue":        "$500M",
		"filings": []string{
			"10-K Annual Report",
			"10-Q Quarterly Report",
			"8-K Current Report",
		},
	}

	return secData, nil
}

func (sfs *SECFilingsScraper) isPotentialPublicCompany(companyName string) bool {
	// Simple heuristic to determine if a company might be public
	// In real implementation, this would be more sophisticated
	publicIndicators := []string{"corp", "inc", "ltd", "llc"}
	
	name := strings.ToLower(companyName)
	for _, indicator := range publicIndicators {
		if strings.Contains(name, indicator) {
			return true
		}
	}
	
	// For demonstration, assume some companies are public
	return len(companyName) > 3
}

func (sfs *SECFilingsScraper) parseSECData(profile *CompanyProfile, data map[string]interface{}) {
	if name, ok := data["company_name"].(string); ok {
		profile.CompanyName = name
	}
	
	if industry, ok := data["industry"].(string); ok {
		profile.Industry = industry
	}
	
	if revenue, ok := data["revenue"].(string); ok {
		profile.Revenue = revenue
	}
	
	if empCount, ok := data["employee_count"].(string); ok {
		profile.CompanySize = empCount + " employees"
	}

	// Parse public trading information
	if ticker, ok := data["ticker_symbol"].(string); ok {
		profile.PublicTrading = &PublicTradingInfo{
			TickerSymbol: ticker,
		}
		
		if exchange, ok := data["exchange"].(string); ok {
			profile.PublicTrading.Exchange = exchange
		}
		
		if marketCap, ok := data["market_cap"].(string); ok {
			profile.PublicTrading.MarketCap = marketCap
		}
		
		if filings, ok := data["filings"].([]string); ok {
			profile.PublicTrading.SECFilings = filings
		}
	}
}

// Utility functions
func extractCompanyNameFromURL(companyURL string) string {
	// Extract company name from LinkedIn URL
	u, err := url.Parse(companyURL)
	if err != nil {
		return ""
	}
	
	parts := strings.Split(u.Path, "/")
	if len(parts) >= 3 && parts[1] == "company" {
		return strings.Title(strings.ReplaceAll(parts[2], "-", " "))
	}
	
	return ""
}
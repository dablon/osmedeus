package database

import (
	"time"
)

// AIEnhancedTarget extends the existing Target with AI analysis fields
type AIEnhancedTarget struct {
	Target

	// AI Classification Results
	AssetClassification *AssetClassification `json:"asset_classification,omitempty" gorm:"embedded;embeddedPrefix:asset_"`
	BusinessContext     *BusinessContext     `json:"business_context,omitempty" gorm:"embedded;embeddedPrefix:business_"`
	BehaviorProfile     *BehaviorProfile     `json:"behavior_profile,omitempty" gorm:"embedded;embeddedPrefix:behavior_"`
	AnomalyScore        float64              `json:"anomaly_score" gorm:"default:0"`

	// Predictive Analysis
	VulnPredictions []VulnerabilityPrediction `json:"vulnerability_predictions,omitempty" gorm:"foreignKey:TargetID"`
	RiskForecast    *RiskForecast             `json:"risk_forecast,omitempty" gorm:"embedded;embeddedPrefix:risk_"`
	ThreatEvolution *ThreatEvolution          `json:"threat_evolution,omitempty" gorm:"embedded;embeddedPrefix:threat_"`

	// Exploit Analysis
	GeneratedExploits  []GeneratedExploit  `json:"generated_exploits,omitempty" gorm:"foreignKey:TargetID"`
	ExploitChains      []ExploitChain      `json:"exploit_chains,omitempty" gorm:"foreignKey:TargetID"`
	ExploitValidations []ExploitValidation `json:"exploit_validations,omitempty" gorm:"foreignKey:TargetID"`

	// Collaborative Data
	TeamAssignments    []TeamAssignment    `json:"team_assignments,omitempty" gorm:"foreignKey:TargetID"`
	SharedFindings     []SharedFinding     `json:"shared_findings,omitempty" gorm:"foreignKey:TargetID"`
	CollaborativeNotes []CollaborativeNote `json:"collaborative_notes,omitempty" gorm:"foreignKey:TargetID"`

	// AI Processing Metadata
	LastAIAnalysis    time.Time `json:"last_ai_analysis"`
	AIAnalysisVersion string    `json:"ai_analysis_version" gorm:"type:varchar(50)"`
	ProcessingStatus  string    `json:"processing_status" gorm:"type:varchar(50);default:'pending'"`
}

// AssetClassification represents AI-powered asset classification results
type AssetClassification struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	TargetID         string    `json:"target_id" gorm:"type:varchar(255);index"`
	Type             string    `json:"type" gorm:"type:varchar(100)"`              // production, staging, dev, admin, api
	Confidence       float64   `json:"confidence" gorm:"type:decimal(5,4)"`
	BusinessValue    float64   `json:"business_value" gorm:"type:decimal(5,4)"`
	CriticalityLevel string    `json:"criticality_level" gorm:"type:varchar(50)"` // critical, high, medium, low
	Reasoning        string    `json:"reasoning" gorm:"type:text"`                // JSON array of reasons
	SimilarAssets    string    `json:"similar_assets" gorm:"type:text"`           // JSON array of similar asset IDs
	ClassifiedAt     time.Time `json:"classified_at"`
	ModelVersion     string    `json:"model_version" gorm:"type:varchar(50)"`
}

// BusinessContext represents OSINT-based business analysis
type BusinessContext struct {
	ID                 uint                   `json:"id" gorm:"primaryKey"`
	TargetID           string                 `json:"target_id" gorm:"type:varchar(255);index"`
	CompanyName        string                 `json:"company_name" gorm:"type:varchar(255)"`
	Industry           string                 `json:"industry" gorm:"type:varchar(100)"`
	CompanySize        string                 `json:"company_size" gorm:"type:varchar(50)"`
	Revenue            string                 `json:"revenue" gorm:"type:varchar(100)"`
	BusinessValue      float64                `json:"business_value" gorm:"type:decimal(5,4)"`
	CriticalityFactors string                 `json:"criticality_factors" gorm:"type:text"` // JSON array
	PublicProfile      string                 `json:"public_profile" gorm:"type:text"`      // JSON object
	RiskFactors        string                 `json:"risk_factors" gorm:"type:text"`        // JSON array
	AnalyzedAt         time.Time              `json:"analyzed_at"`
	DataSources        string                 `json:"data_sources" gorm:"type:text"` // JSON array of sources used
}

// BehaviorProfile represents behavioral analysis results
type BehaviorProfile struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	TargetID             string    `json:"target_id" gorm:"type:varchar(255);index"`
	ResponsePatterns     string    `json:"response_patterns" gorm:"type:text"`     // JSON array
	TimingProfile        string    `json:"timing_profile" gorm:"type:text"`        // JSON object
	ErrorPatterns        string    `json:"error_patterns" gorm:"type:text"`        // JSON array
	AnomalyIndicators    string    `json:"anomaly_indicators" gorm:"type:text"`    // JSON array
	SuspiciousActivities string    `json:"suspicious_activities" gorm:"type:text"` // JSON array
	BaselineDeviation    float64   `json:"baseline_deviation" gorm:"type:decimal(5,4)"`
	ConfidenceScore      float64   `json:"confidence_score" gorm:"type:decimal(5,4)"`
	ProfiledAt           time.Time `json:"profiled_at"`
	ModelVersion         string    `json:"model_version" gorm:"type:varchar(50)"`
}

// VulnerabilityPrediction represents ML-based vulnerability predictions
type VulnerabilityPrediction struct {
	ID                   uint          `json:"id" gorm:"primaryKey"`
	PredictionID         string        `json:"prediction_id" gorm:"type:varchar(255);uniqueIndex"`
	TargetID             string        `json:"target_id" gorm:"type:varchar(255);index"`
	TargetAsset          string        `json:"target_asset" gorm:"type:varchar(255)"`
	PredictedVulnType    string        `json:"predicted_vuln_type" gorm:"type:varchar(100)"`
	EmergenceProbability float64       `json:"emergence_probability" gorm:"type:decimal(5,4)"`
	TimeToEmergence      time.Duration `json:"time_to_emergence"`
	RiskFactors          string        `json:"risk_factors" gorm:"type:text"`          // JSON array
	PreventiveMeasures   string        `json:"preventive_measures" gorm:"type:text"`   // JSON array
	ModelConfidence      float64       `json:"model_confidence" gorm:"type:decimal(5,4)"`
	PredictionBasis      string        `json:"prediction_basis" gorm:"type:text"`      // JSON array
	PredictedAt          time.Time     `json:"predicted_at"`
	ModelVersion         string        `json:"model_version" gorm:"type:varchar(50)"`
}

// ExploitChain represents multi-vulnerability attack sequences
type ExploitChain struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	ChainID            string    `json:"chain_id" gorm:"type:varchar(255);uniqueIndex"`
	TargetID           string    `json:"target_id" gorm:"type:varchar(255);index"`
	Name               string    `json:"name" gorm:"type:varchar(255)"`
	Description        string    `json:"description" gorm:"type:text"`
	AttackSteps        string    `json:"attack_steps" gorm:"type:text"`        // JSON array
	Prerequisites      string    `json:"prerequisites" gorm:"type:text"`       // JSON array
	SuccessProbability float64   `json:"success_probability" gorm:"type:decimal(5,4)"`
	ImpactAssessment   string    `json:"impact_assessment" gorm:"type:text"`   // JSON object
	Mitigations        string    `json:"mitigations" gorm:"type:text"`         // JSON array
	ValidationResults  string    `json:"validation_results" gorm:"type:text"`  // JSON array
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// GeneratedExploit represents automatically generated exploits
type GeneratedExploit struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	ExploitID          string    `json:"exploit_id" gorm:"type:varchar(255);uniqueIndex"`
	TargetID           string    `json:"target_id" gorm:"type:varchar(255);index"`
	VulnerabilityID    string    `json:"vulnerability_id" gorm:"type:varchar(255)"`
	ExploitCode        string    `json:"exploit_code" gorm:"type:longtext"`
	ExploitType        string    `json:"exploit_type" gorm:"type:varchar(100)"`
	Prerequisites      string    `json:"prerequisites" gorm:"type:text"`      // JSON array
	ExpectedImpact     string    `json:"expected_impact" gorm:"type:text"`
	ValidationStatus   string    `json:"validation_status" gorm:"type:varchar(50);default:'pending'"`
	SuccessProbability float64   `json:"success_probability" gorm:"type:decimal(5,4)"`
	GenerationMethod   string    `json:"generation_method" gorm:"type:varchar(100)"`
	Metadata           string    `json:"metadata" gorm:"type:text"` // JSON object
	GeneratedAt        time.Time `json:"generated_at"`
	ValidatedAt        *time.Time `json:"validated_at,omitempty"`
}

// ExploitValidation represents exploit validation results
type ExploitValidation struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	ValidationID     string    `json:"validation_id" gorm:"type:varchar(255);uniqueIndex"`
	TargetID         string    `json:"target_id" gorm:"type:varchar(255);index"`
	ExploitID        string    `json:"exploit_id" gorm:"type:varchar(255);index"`
	ValidationResult string    `json:"validation_result" gorm:"type:varchar(50)"` // success, failure, error
	ExecutionLog     string    `json:"execution_log" gorm:"type:longtext"`
	SuccessEvidence  string    `json:"success_evidence" gorm:"type:text"` // JSON array
	FailureReason    string    `json:"failure_reason" gorm:"type:text"`
	SandboxInfo      string    `json:"sandbox_info" gorm:"type:text"`     // JSON object
	ValidatedAt      time.Time `json:"validated_at"`
	ValidatorVersion string    `json:"validator_version" gorm:"type:varchar(50)"`
}

// RiskForecast represents predictive risk analysis
type RiskForecast struct {
	ID                    uint      `json:"id" gorm:"primaryKey"`
	TargetID              string    `json:"target_id" gorm:"type:varchar(255);index"`
	ForecastPeriod        int       `json:"forecast_period"` // days
	RiskTrend             string    `json:"risk_trend" gorm:"type:varchar(50)"` // increasing, decreasing, stable
	PredictedRiskScore    float64   `json:"predicted_risk_score" gorm:"type:decimal(5,4)"`
	ConfidenceInterval    string    `json:"confidence_interval" gorm:"type:text"` // JSON object
	KeyRiskFactors        string    `json:"key_risk_factors" gorm:"type:text"`    // JSON array
	RecommendedActions    string    `json:"recommended_actions" gorm:"type:text"` // JSON array
	ForecastAccuracy      float64   `json:"forecast_accuracy" gorm:"type:decimal(5,4)"`
	GeneratedAt           time.Time `json:"generated_at"`
	ModelVersion          string    `json:"model_version" gorm:"type:varchar(50)"`
}

// ThreatEvolution represents threat landscape evolution analysis
type ThreatEvolution struct {
	ID                     uint      `json:"id" gorm:"primaryKey"`
	TargetID               string    `json:"target_id" gorm:"type:varchar(255);index"`
	ThreatCategory         string    `json:"threat_category" gorm:"type:varchar(100)"`
	EvolutionPrediction    string    `json:"evolution_prediction" gorm:"type:text"` // JSON object
	EmergingThreatVectors  string    `json:"emerging_threat_vectors" gorm:"type:text"` // JSON array
	ThreatActorProfiles    string    `json:"threat_actor_profiles" gorm:"type:text"` // JSON array
	AttackTechniquesTrends string    `json:"attack_techniques_trends" gorm:"type:text"` // JSON object
	GeopoliticalFactors    string    `json:"geopolitical_factors" gorm:"type:text"` // JSON array
	IndustrySpecificThreats string   `json:"industry_specific_threats" gorm:"type:text"` // JSON array
	AnalyzedAt             time.Time `json:"analyzed_at"`
	ModelVersion           string    `json:"model_version" gorm:"type:varchar(50)"`
}

// Team collaboration models
type TeamAssignment struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	AssignmentID    string    `json:"assignment_id" gorm:"type:varchar(255);uniqueIndex"`
	TargetID        string    `json:"target_id" gorm:"type:varchar(255);index"`
	TeamMemberID    string    `json:"team_member_id" gorm:"type:varchar(255)"`
	TeamMemberName  string    `json:"team_member_name" gorm:"type:varchar(255)"`
	AssignedTargets string    `json:"assigned_targets" gorm:"type:text"` // JSON array
	TaskType        string    `json:"task_type" gorm:"type:varchar(100)"`
	Priority        string    `json:"priority" gorm:"type:varchar(50)"`
	Status          string    `json:"status" gorm:"type:varchar(50);default:'assigned'"`
	Progress        float64   `json:"progress" gorm:"type:decimal(5,4);default:0"`
	Deadline        *time.Time `json:"deadline,omitempty"`
	Dependencies    string    `json:"dependencies" gorm:"type:text"` // JSON array
	AssignedAt      time.Time `json:"assigned_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type SharedFinding struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	FindingID      string    `json:"finding_id" gorm:"type:varchar(255);uniqueIndex"`
	TargetID       string    `json:"target_id" gorm:"type:varchar(255);index"`
	DiscoveredBy   string    `json:"discovered_by" gorm:"type:varchar(255)"`
	FindingType    string    `json:"finding_type" gorm:"type:varchar(100)"`
	Severity       string    `json:"severity" gorm:"type:varchar(50)"`
	Title          string    `json:"title" gorm:"type:varchar(255)"`
	Description    string    `json:"description" gorm:"type:text"`
	Evidence       string    `json:"evidence" gorm:"type:text"`       // JSON array
	ValidationInfo string    `json:"validation_info" gorm:"type:text"` // JSON object
	SharedWith     string    `json:"shared_with" gorm:"type:text"`     // JSON array of team member IDs
	Annotations    string    `json:"annotations" gorm:"type:text"`     // JSON array
	DiscoveredAt   time.Time `json:"discovered_at"`
	SharedAt       time.Time `json:"shared_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CollaborativeNote struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	NoteID     string    `json:"note_id" gorm:"type:varchar(255);uniqueIndex"`
	TargetID   string    `json:"target_id" gorm:"type:varchar(255);index"`
	AuthorID   string    `json:"author_id" gorm:"type:varchar(255)"`
	AuthorName string    `json:"author_name" gorm:"type:varchar(255)"`
	Content    string    `json:"content" gorm:"type:text"`
	NoteType   string    `json:"note_type" gorm:"type:varchar(50);default:'general'"` // general, analysis, recommendation
	Tags       string    `json:"tags" gorm:"type:text"`                               // JSON array
	References string    `json:"references" gorm:"type:text"`                         // JSON array of referenced finding IDs
	Visibility string    `json:"visibility" gorm:"type:varchar(50);default:'team'"`   // private, team, public
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ML Infrastructure Models
type MLModel struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ModelID      string    `json:"model_id" gorm:"type:varchar(255);uniqueIndex"`
	Name         string    `json:"name" gorm:"type:varchar(255)"`
	Type         string    `json:"type" gorm:"type:varchar(100)"` // classification, regression, clustering, etc.
	Version      string    `json:"version" gorm:"type:varchar(50)"`
	FilePath     string    `json:"file_path" gorm:"type:varchar(500)"`
	Framework    string    `json:"framework" gorm:"type:varchar(50)"` // tensorflow, pytorch, sklearn
	Status       string    `json:"status" gorm:"type:varchar(50);default:'inactive'"`
	Accuracy     float64   `json:"accuracy" gorm:"type:decimal(5,4)"`
	TrainedAt    time.Time `json:"trained_at"`
	DeployedAt   *time.Time `json:"deployed_at,omitempty"`
	Metadata     string    `json:"metadata" gorm:"type:text"` // JSON object with model-specific info
}

type MLTrainingData struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	DatasetID   string    `json:"dataset_id" gorm:"type:varchar(255);uniqueIndex"`
	ModelType   string    `json:"model_type" gorm:"type:varchar(100)"`
	DataPath    string    `json:"data_path" gorm:"type:varchar(500)"`
	DataSize    int       `json:"data_size"`
	Features    string    `json:"features" gorm:"type:text"`    // JSON array
	Labels      string    `json:"labels" gorm:"type:text"`      // JSON array
	Quality     float64   `json:"quality" gorm:"type:decimal(5,4)"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsed    *time.Time `json:"last_used,omitempty"`
	Description string    `json:"description" gorm:"type:text"`
}

type MLInferenceLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ModelID      string    `json:"model_id" gorm:"type:varchar(255);index"`
	TargetID     string    `json:"target_id" gorm:"type:varchar(255);index"`
	InputData    string    `json:"input_data" gorm:"type:text"`    // JSON object
	OutputData   string    `json:"output_data" gorm:"type:text"`   // JSON object
	Confidence   float64   `json:"confidence" gorm:"type:decimal(5,4)"`
	ProcessingTime int     `json:"processing_time"` // milliseconds
	Status       string    `json:"status" gorm:"type:varchar(50)"`
	ErrorMessage string    `json:"error_message" gorm:"type:text"`
	InferredAt   time.Time `json:"inferred_at"`
}
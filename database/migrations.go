package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	ID          uint      `gorm:"primaryKey"`
	Version     string    `gorm:"type:varchar(50);uniqueIndex"`
	Description string    `gorm:"type:varchar(255)"`
	AppliedAt   time.Time
}

// MigrationManager handles database migrations
type MigrationManager struct {
	db *gorm.DB
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *gorm.DB) *MigrationManager {
	return &MigrationManager{db: db}
}

// RunMigrations executes all pending migrations
func (mm *MigrationManager) RunMigrations() error {
	// Ensure migrations table exists
	if err := mm.db.AutoMigrate(&Migration{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrations := []struct {
		version     string
		description string
		migrateFn   func(*gorm.DB) error
	}{
		{
			version:     "001_add_ai_enhanced_target",
			description: "Add AI-enhanced target fields",
			migrateFn:   mm.addAIEnhancedTargetFields,
		},
		{
			version:     "002_create_ai_analysis_tables",
			description: "Create AI analysis tables",
			migrateFn:   mm.createAIAnalysisTables,
		},
		{
			version:     "003_create_ml_infrastructure_tables",
			description: "Create ML infrastructure tables",
			migrateFn:   mm.createMLInfrastructureTables,
		},
		{
			version:     "004_add_ai_indexes",
			description: "Add indexes for AI tables",
			migrateFn:   mm.addAIIndexes,
		},
	}

	for _, migration := range migrations {
		if err := mm.runMigration(migration.version, migration.description, migration.migrateFn); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.version, err)
		}
	}

	return nil
}

// runMigration executes a single migration if it hasn't been applied
func (mm *MigrationManager) runMigration(version, description string, migrateFn func(*gorm.DB) error) error {
	// Check if migration has already been applied
	var count int64
	mm.db.Model(&Migration{}).Where("version = ?", version).Count(&count)
	
	if count > 0 {
		return nil // Migration already applied
	}

	// Run the migration
	if err := migrateFn(mm.db); err != nil {
		return err
	}

	// Record the migration
	migration := Migration{
		Version:     version,
		Description: description,
		AppliedAt:   time.Now(),
	}

	return mm.db.Create(&migration).Error
}

// addAIEnhancedTargetFields adds AI analysis fields to existing Target table
func (mm *MigrationManager) addAIEnhancedTargetFields(db *gorm.DB) error {
	// Add AI analysis fields to Target table
	queries := []string{
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS anomaly_score DECIMAL(5,4) DEFAULT 0`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS last_ai_analysis TIMESTAMP`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS ai_analysis_version VARCHAR(50)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS processing_status VARCHAR(50) DEFAULT 'pending'`,
		
		// Asset Classification fields
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS asset_type VARCHAR(100)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS asset_confidence DECIMAL(5,4)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS asset_business_value DECIMAL(5,4)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS asset_criticality_level VARCHAR(50)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS asset_reasoning TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS asset_similar_assets TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS asset_classified_at TIMESTAMP`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS asset_model_version VARCHAR(50)`,
		
		// Business Context fields
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS business_company_name VARCHAR(255)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS business_industry VARCHAR(100)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS business_company_size VARCHAR(50)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS business_revenue VARCHAR(100)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS business_business_value DECIMAL(5,4)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS business_criticality_factors TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS business_public_profile TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS business_risk_factors TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS business_analyzed_at TIMESTAMP`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS business_data_sources TEXT`,
		
		// Behavior Profile fields
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS behavior_response_patterns TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS behavior_timing_profile TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS behavior_error_patterns TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS behavior_anomaly_indicators TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS behavior_suspicious_activities TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS behavior_baseline_deviation DECIMAL(5,4)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS behavior_confidence_score DECIMAL(5,4)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS behavior_profiled_at TIMESTAMP`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS behavior_model_version VARCHAR(50)`,
		
		// Risk Forecast fields
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS risk_forecast_period INTEGER`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS risk_risk_trend VARCHAR(50)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS risk_predicted_risk_score DECIMAL(5,4)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS risk_confidence_interval TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS risk_key_risk_factors TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS risk_recommended_actions TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS risk_forecast_accuracy DECIMAL(5,4)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS risk_generated_at TIMESTAMP`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS risk_model_version VARCHAR(50)`,
		
		// Threat Evolution fields
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS threat_threat_category VARCHAR(100)`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS threat_evolution_prediction TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS threat_emerging_threat_vectors TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS threat_threat_actor_profiles TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS threat_attack_techniques_trends TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS threat_geopolitical_factors TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS threat_industry_specific_threats TEXT`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS threat_analyzed_at TIMESTAMP`,
		`ALTER TABLE targets ADD COLUMN IF NOT EXISTS threat_model_version VARCHAR(50)`,
	}

	for _, query := range queries {
		if err := db.Exec(query).Error; err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	return nil
}

// createAIAnalysisTables creates tables for AI analysis data
func (mm *MigrationManager) createAIAnalysisTables(db *gorm.DB) error {
	// Create all AI analysis tables
	tables := []interface{}{
		&AssetClassification{},
		&BusinessContext{},
		&BehaviorProfile{},
		&VulnerabilityPrediction{},
		&ExploitChain{},
		&GeneratedExploit{},
		&ExploitValidation{},
		&RiskForecast{},
		&ThreatEvolution{},
		&TeamAssignment{},
		&SharedFinding{},
		&CollaborativeNote{},
	}

	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			return fmt.Errorf("failed to create table for %T: %w", table, err)
		}
	}

	return nil
}

// createMLInfrastructureTables creates tables for ML infrastructure
func (mm *MigrationManager) createMLInfrastructureTables(db *gorm.DB) error {
	tables := []interface{}{
		&MLModel{},
		&MLTrainingData{},
		&MLInferenceLog{},
	}

	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			return fmt.Errorf("failed to create ML table for %T: %w", table, err)
		}
	}

	return nil
}

// addAIIndexes adds database indexes for AI tables to improve query performance
func (mm *MigrationManager) addAIIndexes(db *gorm.DB) error {
	indexes := []string{
		// Asset Classification indexes
		`CREATE INDEX IF NOT EXISTS idx_asset_classification_target_id ON asset_classifications(target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_classification_type ON asset_classifications(type)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_classification_criticality ON asset_classifications(criticality_level)`,
		
		// Business Context indexes
		`CREATE INDEX IF NOT EXISTS idx_business_context_target_id ON business_contexts(target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_business_context_industry ON business_contexts(industry)`,
		`CREATE INDEX IF NOT EXISTS idx_business_context_company_size ON business_contexts(company_size)`,
		
		// Behavior Profile indexes
		`CREATE INDEX IF NOT EXISTS idx_behavior_profile_target_id ON behavior_profiles(target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_behavior_profile_anomaly_score ON behavior_profiles(baseline_deviation)`,
		
		// Vulnerability Prediction indexes
		`CREATE INDEX IF NOT EXISTS idx_vuln_prediction_target_id ON vulnerability_predictions(target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_vuln_prediction_type ON vulnerability_predictions(predicted_vuln_type)`,
		`CREATE INDEX IF NOT EXISTS idx_vuln_prediction_probability ON vulnerability_predictions(emergence_probability)`,
		
		// Exploit Chain indexes
		`CREATE INDEX IF NOT EXISTS idx_exploit_chain_target_id ON exploit_chains(target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_exploit_chain_success_prob ON exploit_chains(success_probability)`,
		
		// Generated Exploit indexes
		`CREATE INDEX IF NOT EXISTS idx_generated_exploit_target_id ON generated_exploits(target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_generated_exploit_vuln_id ON generated_exploits(vulnerability_id)`,
		`CREATE INDEX IF NOT EXISTS idx_generated_exploit_status ON generated_exploits(validation_status)`,
		
		// Exploit Validation indexes
		`CREATE INDEX IF NOT EXISTS idx_exploit_validation_target_id ON exploit_validations(target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_exploit_validation_exploit_id ON exploit_validations(exploit_id)`,
		`CREATE INDEX IF NOT EXISTS idx_exploit_validation_result ON exploit_validations(validation_result)`,
		
		// Risk Forecast indexes
		`CREATE INDEX IF NOT EXISTS idx_risk_forecast_target_id ON risk_forecasts(target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_risk_forecast_trend ON risk_forecasts(risk_trend)`,
		
		// Threat Evolution indexes
		`CREATE INDEX IF NOT EXISTS idx_threat_evolution_target_id ON threat_evolutions(target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_threat_evolution_category ON threat_evolutions(threat_category)`,
		
		// Team Assignment indexes
		`CREATE INDEX IF NOT EXISTS idx_team_assignment_target_id ON team_assignments(target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_team_assignment_member_id ON team_assignments(team_member_id)`,
		`CREATE INDEX IF NOT EXISTS idx_team_assignment_status ON team_assignments(status)`,
		
		// Shared Finding indexes
		`CREATE INDEX IF NOT EXISTS idx_shared_finding_target_id ON shared_findings(target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_shared_finding_type ON shared_findings(finding_type)`,
		`CREATE INDEX IF NOT EXISTS idx_shared_finding_severity ON shared_findings(severity)`,
		
		// Collaborative Note indexes
		`CREATE INDEX IF NOT EXISTS idx_collaborative_note_target_id ON collaborative_notes(target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_collaborative_note_author_id ON collaborative_notes(author_id)`,
		
		// ML Model indexes
		`CREATE INDEX IF NOT EXISTS idx_ml_model_type ON ml_models(type)`,
		`CREATE INDEX IF NOT EXISTS idx_ml_model_status ON ml_models(status)`,
		
		// ML Inference Log indexes
		`CREATE INDEX IF NOT EXISTS idx_ml_inference_model_id ON ml_inference_logs(model_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ml_inference_target_id ON ml_inference_logs(target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ml_inference_status ON ml_inference_logs(status)`,
		`CREATE INDEX IF NOT EXISTS idx_ml_inference_inferred_at ON ml_inference_logs(inferred_at)`,
		
		// Target AI fields indexes
		`CREATE INDEX IF NOT EXISTS idx_target_processing_status ON targets(processing_status)`,
		`CREATE INDEX IF NOT EXISTS idx_target_anomaly_score ON targets(anomaly_score)`,
		`CREATE INDEX IF NOT EXISTS idx_target_last_ai_analysis ON targets(last_ai_analysis)`,
	}

	for _, indexQuery := range indexes {
		if err := db.Exec(indexQuery).Error; err != nil {
			return fmt.Errorf("failed to create index: %s, error: %w", indexQuery, err)
		}
	}

	return nil
}

// RollbackMigration rolls back a specific migration (if supported)
func (mm *MigrationManager) RollbackMigration(version string) error {
	// This is a simplified rollback - in production you'd want more sophisticated rollback logic
	return mm.db.Where("version = ?", version).Delete(&Migration{}).Error
}

// GetAppliedMigrations returns a list of applied migrations
func (mm *MigrationManager) GetAppliedMigrations() ([]Migration, error) {
	var migrations []Migration
	err := mm.db.Order("applied_at ASC").Find(&migrations).Error
	return migrations, err
}

// IsMigrationApplied checks if a specific migration has been applied
func (mm *MigrationManager) IsMigrationApplied(version string) (bool, error) {
	var count int64
	err := mm.db.Model(&Migration{}).Where("version = ?", version).Count(&count).Error
	return count > 0, err
}
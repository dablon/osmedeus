package database

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"github.com/j3ssie/osmedeus/utils"
)

// AIDatabase provides AI-enhanced database operations
type AIDatabase struct {
	db *gorm.DB
	migrationManager *MigrationManager
}

// NewAIDatabase creates a new AI-enhanced database instance
func NewAIDatabase(db *gorm.DB) *AIDatabase {
	return &AIDatabase{
		db: db,
		migrationManager: NewMigrationManager(db),
	}
}

// InitializeAITables initializes all AI-related database tables and migrations
func (aidb *AIDatabase) InitializeAITables() error {
	utils.InforF("Initializing AI-enhanced database tables...")
	
	// Run migrations
	if err := aidb.migrationManager.RunMigrations(); err != nil {
		return fmt.Errorf("failed to run AI migrations: %w", err)
	}

	utils.InforF("AI database initialization completed successfully")
	return nil
}

// GetAIEnhancedTarget retrieves a target with all AI analysis data
func (aidb *AIDatabase) GetAIEnhancedTarget(inputName string) (*AIEnhancedTarget, error) {
	var target AIEnhancedTarget
	
	err := aidb.db.Preload("VulnPredictions").
		Preload("GeneratedExploits").
		Preload("ExploitChains").
		Preload("ExploitValidations").
		Preload("TeamAssignments").
		Preload("SharedFindings").
		Preload("CollaborativeNotes").
		Where("input_name = ?", inputName).
		First(&target).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get AI enhanced target: %w", err)
	}

	return &target, nil
}

// SaveAIEnhancedTarget saves or updates an AI-enhanced target
func (aidb *AIDatabase) SaveAIEnhancedTarget(target *AIEnhancedTarget) error {
	target.LastAIAnalysis = time.Now()
	target.ProcessingStatus = "completed"
	
	return aidb.db.Save(target).Error
}

// CreateAssetClassification creates a new asset classification record
func (aidb *AIDatabase) CreateAssetClassification(classification *AssetClassification) error {
	classification.ClassifiedAt = time.Now()
	return aidb.db.Create(classification).Error
}

// GetAssetClassification retrieves asset classification for a target
func (aidb *AIDatabase) GetAssetClassification(targetID string) (*AssetClassification, error) {
	var classification AssetClassification
	err := aidb.db.Where("target_id = ?", targetID).First(&classification).Error
	if err != nil {
		return nil, err
	}
	return &classification, nil
}

// CreateBusinessContext creates a new business context record
func (aidb *AIDatabase) CreateBusinessContext(context *BusinessContext) error {
	context.AnalyzedAt = time.Now()
	return aidb.db.Create(context).Error
}

// GetBusinessContext retrieves business context for a target
func (aidb *AIDatabase) GetBusinessContext(targetID string) (*BusinessContext, error) {
	var context BusinessContext
	err := aidb.db.Where("target_id = ?", targetID).First(&context).Error
	if err != nil {
		return nil, err
	}
	return &context, nil
}

// CreateBehaviorProfile creates a new behavior profile record
func (aidb *AIDatabase) CreateBehaviorProfile(profile *BehaviorProfile) error {
	profile.ProfiledAt = time.Now()
	return aidb.db.Create(profile).Error
}

// GetBehaviorProfile retrieves behavior profile for a target
func (aidb *AIDatabase) GetBehaviorProfile(targetID string) (*BehaviorProfile, error) {
	var profile BehaviorProfile
	err := aidb.db.Where("target_id = ?", targetID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// CreateVulnerabilityPrediction creates a new vulnerability prediction
func (aidb *AIDatabase) CreateVulnerabilityPrediction(prediction *VulnerabilityPrediction) error {
	prediction.PredictedAt = time.Now()
	return aidb.db.Create(prediction).Error
}

// GetVulnerabilityPredictions retrieves all vulnerability predictions for a target
func (aidb *AIDatabase) GetVulnerabilityPredictions(targetID string) ([]VulnerabilityPrediction, error) {
	var predictions []VulnerabilityPrediction
	err := aidb.db.Where("target_id = ?", targetID).Find(&predictions).Error
	return predictions, err
}

// CreateExploitChain creates a new exploit chain
func (aidb *AIDatabase) CreateExploitChain(chain *ExploitChain) error {
	chain.CreatedAt = time.Now()
	chain.UpdatedAt = time.Now()
	return aidb.db.Create(chain).Error
}

// GetExploitChains retrieves all exploit chains for a target
func (aidb *AIDatabase) GetExploitChains(targetID string) ([]ExploitChain, error) {
	var chains []ExploitChain
	err := aidb.db.Where("target_id = ?", targetID).Find(&chains).Error
	return chains, err
}

// CreateGeneratedExploit creates a new generated exploit
func (aidb *AIDatabase) CreateGeneratedExploit(exploit *GeneratedExploit) error {
	exploit.GeneratedAt = time.Now()
	return aidb.db.Create(exploit).Error
}

// GetGeneratedExploits retrieves all generated exploits for a target
func (aidb *AIDatabase) GetGeneratedExploits(targetID string) ([]GeneratedExploit, error) {
	var exploits []GeneratedExploit
	err := aidb.db.Where("target_id = ?", targetID).Find(&exploits).Error
	return exploits, err
}

// CreateExploitValidation creates a new exploit validation record
func (aidb *AIDatabase) CreateExploitValidation(validation *ExploitValidation) error {
	validation.ValidatedAt = time.Now()
	return aidb.db.Create(validation).Error
}

// GetExploitValidations retrieves all exploit validations for a target
func (aidb *AIDatabase) GetExploitValidations(targetID string) ([]ExploitValidation, error) {
	var validations []ExploitValidation
	err := aidb.db.Where("target_id = ?", targetID).Find(&validations).Error
	return validations, err
}

// CreateTeamAssignment creates a new team assignment
func (aidb *AIDatabase) CreateTeamAssignment(assignment *TeamAssignment) error {
	assignment.AssignedAt = time.Now()
	return aidb.db.Create(assignment).Error
}

// GetTeamAssignments retrieves all team assignments for a target
func (aidb *AIDatabase) GetTeamAssignments(targetID string) ([]TeamAssignment, error) {
	var assignments []TeamAssignment
	err := aidb.db.Where("target_id = ?", targetID).Find(&assignments).Error
	return assignments, err
}

// CreateSharedFinding creates a new shared finding
func (aidb *AIDatabase) CreateSharedFinding(finding *SharedFinding) error {
	finding.DiscoveredAt = time.Now()
	finding.SharedAt = time.Now()
	finding.UpdatedAt = time.Now()
	return aidb.db.Create(finding).Error
}

// GetSharedFindings retrieves all shared findings for a target
func (aidb *AIDatabase) GetSharedFindings(targetID string) ([]SharedFinding, error) {
	var findings []SharedFinding
	err := aidb.db.Where("target_id = ?", targetID).Find(&findings).Error
	return findings, err
}

// CreateCollaborativeNote creates a new collaborative note
func (aidb *AIDatabase) CreateCollaborativeNote(note *CollaborativeNote) error {
	note.CreatedAt = time.Now()
	note.UpdatedAt = time.Now()
	return aidb.db.Create(note).Error
}

// GetCollaborativeNotes retrieves all collaborative notes for a target
func (aidb *AIDatabase) GetCollaborativeNotes(targetID string) ([]CollaborativeNote, error) {
	var notes []CollaborativeNote
	err := aidb.db.Where("target_id = ?", targetID).Find(&notes).Error
	return notes, err
}

// ML Model Management

// CreateMLModel creates a new ML model record
func (aidb *AIDatabase) CreateMLModel(model *MLModel) error {
	model.TrainedAt = time.Now()
	return aidb.db.Create(model).Error
}

// GetMLModel retrieves an ML model by ID
func (aidb *AIDatabase) GetMLModel(modelID string) (*MLModel, error) {
	var model MLModel
	err := aidb.db.Where("model_id = ?", modelID).First(&model).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// UpdateMLModelStatus updates the status of an ML model
func (aidb *AIDatabase) UpdateMLModelStatus(modelID, status string) error {
	return aidb.db.Model(&MLModel{}).Where("model_id = ?", modelID).Update("status", status).Error
}

// CreateMLTrainingData creates a new training data record
func (aidb *AIDatabase) CreateMLTrainingData(data *MLTrainingData) error {
	data.CreatedAt = time.Now()
	return aidb.db.Create(data).Error
}

// GetMLTrainingData retrieves training data by dataset ID
func (aidb *AIDatabase) GetMLTrainingData(datasetID string) (*MLTrainingData, error) {
	var data MLTrainingData
	err := aidb.db.Where("dataset_id = ?", datasetID).First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// LogMLInference logs an ML inference operation
func (aidb *AIDatabase) LogMLInference(log *MLInferenceLog) error {
	log.InferredAt = time.Now()
	return aidb.db.Create(log).Error
}

// GetMLInferenceLogs retrieves inference logs for a model
func (aidb *AIDatabase) GetMLInferenceLogs(modelID string, limit int) ([]MLInferenceLog, error) {
	var logs []MLInferenceLog
	query := aidb.db.Where("model_id = ?", modelID).Order("inferred_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&logs).Error
	return logs, err
}

// Analytics and Reporting

// GetTargetsForAIProcessing retrieves targets that need AI processing
func (aidb *AIDatabase) GetTargetsForAIProcessing(limit int) ([]Target, error) {
	var targets []Target
	query := aidb.db.Where("processing_status = ? OR processing_status IS NULL", "pending")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&targets).Error
	return targets, err
}

// GetHighRiskTargets retrieves targets with high anomaly scores
func (aidb *AIDatabase) GetHighRiskTargets(threshold float64, limit int) ([]AIEnhancedTarget, error) {
	var targets []AIEnhancedTarget
	query := aidb.db.Where("anomaly_score > ?", threshold).Order("anomaly_score DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&targets).Error
	return targets, err
}

// GetCriticalAssets retrieves assets classified as critical
func (aidb *AIDatabase) GetCriticalAssets(limit int) ([]AssetClassification, error) {
	var assets []AssetClassification
	query := aidb.db.Where("criticality_level = ?", "critical").Order("business_value DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&assets).Error
	return assets, err
}

// GetPendingExploitValidations retrieves exploits that need validation
func (aidb *AIDatabase) GetPendingExploitValidations(limit int) ([]GeneratedExploit, error) {
	var exploits []GeneratedExploit
	query := aidb.db.Where("validation_status = ?", "pending").Order("generated_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&exploits).Error
	return exploits, err
}

// GetRecentVulnerabilityPredictions retrieves recent vulnerability predictions
func (aidb *AIDatabase) GetRecentVulnerabilityPredictions(days int, limit int) ([]VulnerabilityPrediction, error) {
	var predictions []VulnerabilityPrediction
	since := time.Now().AddDate(0, 0, -days)
	query := aidb.db.Where("predicted_at > ?", since).Order("emergence_probability DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&predictions).Error
	return predictions, err
}

// Utility functions for JSON handling

// MarshalToJSON converts a struct to JSON string for database storage
func (aidb *AIDatabase) MarshalToJSON(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(jsonData), nil
}

// UnmarshalFromJSON converts a JSON string from database to struct
func (aidb *AIDatabase) UnmarshalFromJSON(jsonStr string, target interface{}) error {
	if jsonStr == "" {
		return nil
	}
	err := json.Unmarshal([]byte(jsonStr), target)
	if err != nil {
		return fmt.Errorf("failed to unmarshal from JSON: %w", err)
	}
	return nil
}

// Batch operations for performance

// BatchCreateAssetClassifications creates multiple asset classifications
func (aidb *AIDatabase) BatchCreateAssetClassifications(classifications []AssetClassification) error {
	if len(classifications) == 0 {
		return nil
	}
	
	for i := range classifications {
		classifications[i].ClassifiedAt = time.Now()
	}
	
	return aidb.db.CreateInBatches(classifications, 100).Error
}

// BatchCreateVulnerabilityPredictions creates multiple vulnerability predictions
func (aidb *AIDatabase) BatchCreateVulnerabilityPredictions(predictions []VulnerabilityPrediction) error {
	if len(predictions) == 0 {
		return nil
	}
	
	for i := range predictions {
		predictions[i].PredictedAt = time.Now()
	}
	
	return aidb.db.CreateInBatches(predictions, 100).Error
}

// BatchUpdateProcessingStatus updates processing status for multiple targets
func (aidb *AIDatabase) BatchUpdateProcessingStatus(targetIDs []string, status string) error {
	if len(targetIDs) == 0 {
		return nil
	}
	
	return aidb.db.Model(&Target{}).
		Where("input_name IN ?", targetIDs).
		Updates(map[string]interface{}{
			"processing_status": status,
			"last_ai_analysis": time.Now(),
		}).Error
}

// Statistics and metrics

// GetAIProcessingStats retrieves statistics about AI processing
func (aidb *AIDatabase) GetAIProcessingStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Count targets by processing status
	var statusCounts []struct {
		ProcessingStatus string
		Count           int64
	}
	
	err := aidb.db.Model(&Target{}).
		Select("processing_status, COUNT(*) as count").
		Group("processing_status").
		Find(&statusCounts).Error
	if err != nil {
		return nil, err
	}
	
	statusMap := make(map[string]int64)
	for _, sc := range statusCounts {
		statusMap[sc.ProcessingStatus] = sc.Count
	}
	stats["processing_status"] = statusMap
	
	// Count asset classifications by type
	var assetTypeCounts []struct {
		Type  string
		Count int64
	}
	
	err = aidb.db.Model(&AssetClassification{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Find(&assetTypeCounts).Error
	if err != nil {
		return nil, err
	}
	
	assetTypeMap := make(map[string]int64)
	for _, atc := range assetTypeCounts {
		assetTypeMap[atc.Type] = atc.Count
	}
	stats["asset_types"] = assetTypeMap
	
	// Count vulnerability predictions
	var vulnPredCount int64
	err = aidb.db.Model(&VulnerabilityPrediction{}).Count(&vulnPredCount).Error
	if err != nil {
		return nil, err
	}
	stats["vulnerability_predictions"] = vulnPredCount
	
	// Count generated exploits
	var exploitCount int64
	err = aidb.db.Model(&GeneratedExploit{}).Count(&exploitCount).Error
	if err != nil {
		return nil, err
	}
	stats["generated_exploits"] = exploitCount
	
	return stats, nil
}

// CleanupOldData removes old AI analysis data based on retention policy
func (aidb *AIDatabase) CleanupOldData(retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	
	// Clean up old inference logs
	err := aidb.db.Where("inferred_at < ?", cutoffDate).Delete(&MLInferenceLog{}).Error
	if err != nil {
		return fmt.Errorf("failed to cleanup inference logs: %w", err)
	}
	
	// Clean up old vulnerability predictions
	err = aidb.db.Where("predicted_at < ?", cutoffDate).Delete(&VulnerabilityPrediction{}).Error
	if err != nil {
		return fmt.Errorf("failed to cleanup vulnerability predictions: %w", err)
	}
	
	utils.InforF("Cleaned up AI data older than %d days", retentionDays)
	return nil
}
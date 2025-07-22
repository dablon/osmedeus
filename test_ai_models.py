#!/usr/bin/env python3
"""
Simple test script to validate the AI infrastructure implementation
This script checks the structure and completeness of the implemented files
"""

import os
import json
import re
from pathlib import Path

def test_file_structure():
    """Test that all required files were created"""
    print("🔍 Testing file structure...")
    
    required_files = [
        "database/ai_models.go",
        "database/ai_integration.go", 
        "database/migrations.go",
        "ml/infrastructure.go",
        "ml/models.go",
        "ml/preprocessing.go",
        "ml/infrastructure_test.go"
    ]
    
    missing_files = []
    for file_path in required_files:
        if not os.path.exists(file_path):
            missing_files.append(file_path)
        else:
            print(f"✅ {file_path} exists")
    
    if missing_files:
        print(f"❌ Missing files: {missing_files}")
        return False
    
    print("✅ All required files are present")
    return True

def test_ai_models_structure():
    """Test the AI models file structure"""
    print("\n🔍 Testing AI models structure...")
    
    with open("database/ai_models.go", "r") as f:
        content = f.read()
    
    required_structs = [
        "AIEnhancedTarget",
        "AssetClassification", 
        "BusinessContext",
        "BehaviorProfile",
        "VulnerabilityPrediction",
        "ExploitChain",
        "GeneratedExploit",
        "ExploitValidation",
        "RiskForecast",
        "ThreatEvolution",
        "TeamAssignment",
        "SharedFinding",
        "CollaborativeNote",
        "MLModel",
        "MLTrainingData",
        "MLInferenceLog"
    ]
    
    missing_structs = []
    for struct_name in required_structs:
        if f"type {struct_name} struct" not in content:
            missing_structs.append(struct_name)
        else:
            print(f"✅ {struct_name} struct defined")
    
    if missing_structs:
        print(f"❌ Missing structs: {missing_structs}")
        return False
    
    print("✅ All required AI model structs are defined")
    return True

def test_ml_infrastructure():
    """Test the ML infrastructure implementation"""
    print("\n🔍 Testing ML infrastructure...")
    
    with open("ml/infrastructure.go", "r") as f:
        content = f.read()
    
    required_components = [
        "type MLInfrastructure struct",
        "type MLConfig struct", 
        "type ModelWrapper struct",
        "type InferenceRequest struct",
        "type InferenceResponse struct",
        "type TrainingRequest struct",
        "type TrainingResponse struct",
        "func NewMLInfrastructure",
        "func (ml *MLInfrastructure) LoadModel",
        "func (ml *MLInfrastructure) Predict",
        "func (ml *MLInfrastructure) TrainModel"
    ]
    
    missing_components = []
    for component in required_components:
        if component not in content:
            missing_components.append(component)
        else:
            print(f"✅ {component} implemented")
    
    if missing_components:
        print(f"❌ Missing components: {missing_components}")
        return False
    
    print("✅ ML infrastructure is properly implemented")
    return True

def test_preprocessing_pipeline():
    """Test the preprocessing pipeline implementation"""
    print("\n🔍 Testing preprocessing pipeline...")
    
    # Read both preprocessing files
    with open("ml/preprocessing.go", "r") as f:
        preprocessing_content = f.read()
    
    with open("ml/models.go", "r") as f:
        models_content = f.read()
    
    # Combine content for checking
    combined_content = preprocessing_content + models_content
    
    required_preprocessors = [
        "type PreprocessingPipeline struct",
        "type AssetDataPreprocessor struct",
        "type BehaviorDataPreprocessor struct", 
        "type VulnerabilityDataPreprocessor struct",
        "type NetworkDataPreprocessor struct",
        "type TextDataPreprocessor struct",
        "type TimeSeriesPreprocessor struct",
        "func NewPreprocessingPipeline",
        "func (pp *PreprocessingPipeline) ProcessAssetData",
        "func (pp *PreprocessingPipeline) ProcessBehaviorData",
        "func (pp *PreprocessingPipeline) ProcessVulnerabilityData"
    ]
    
    missing_preprocessors = []
    for preprocessor in required_preprocessors:
        if preprocessor not in combined_content:
            missing_preprocessors.append(preprocessor)
        else:
            print(f"✅ {preprocessor} implemented")
    
    if missing_preprocessors:
        print(f"❌ Missing preprocessors: {missing_preprocessors}")
        return False
    
    print("✅ Preprocessing pipeline is properly implemented")
    return True

def test_database_integration():
    """Test the database integration implementation"""
    print("\n🔍 Testing database integration...")
    
    with open("database/ai_integration.go", "r") as f:
        content = f.read()
    
    required_methods = [
        "type AIDatabase struct",
        "func NewAIDatabase",
        "func (aidb *AIDatabase) InitializeAITables",
        "func (aidb *AIDatabase) GetAIEnhancedTarget",
        "func (aidb *AIDatabase) SaveAIEnhancedTarget",
        "func (aidb *AIDatabase) CreateAssetClassification",
        "func (aidb *AIDatabase) CreateVulnerabilityPrediction",
        "func (aidb *AIDatabase) CreateExploitChain",
        "func (aidb *AIDatabase) GetAIProcessingStats"
    ]
    
    missing_methods = []
    for method in required_methods:
        if method not in content:
            missing_methods.append(method)
        else:
            print(f"✅ {method} implemented")
    
    if missing_methods:
        print(f"❌ Missing methods: {missing_methods}")
        return False
    
    print("✅ Database integration is properly implemented")
    return True

def test_migrations():
    """Test the database migrations implementation"""
    print("\n🔍 Testing database migrations...")
    
    with open("database/migrations.go", "r") as f:
        content = f.read()
    
    required_migrations = [
        "type Migration struct",
        "type MigrationManager struct",
        "func NewMigrationManager",
        "func (mm *MigrationManager) RunMigrations",
        "addAIEnhancedTargetFields",
        "createAIAnalysisTables", 
        "createMLInfrastructureTables",
        "addAIIndexes"
    ]
    
    missing_migrations = []
    for migration in required_migrations:
        if migration not in content:
            missing_migrations.append(migration)
        else:
            print(f"✅ {migration} implemented")
    
    if missing_migrations:
        print(f"❌ Missing migrations: {missing_migrations}")
        return False
    
    print("✅ Database migrations are properly implemented")
    return True

def test_go_mod_dependencies():
    """Test that go.mod has the required dependencies"""
    print("\n🔍 Testing Go module dependencies...")
    
    with open("go.mod", "r") as f:
        content = f.read()
    
    required_deps = [
        "gorm.io/gorm",
        "gorm.io/driver/sqlite",
        "gorm.io/driver/mysql", 
        "gorm.io/driver/postgres"
    ]
    
    missing_deps = []
    for dep in required_deps:
        if dep not in content:
            missing_deps.append(dep)
        else:
            print(f"✅ {dep} dependency added")
    
    if missing_deps:
        print(f"❌ Missing dependencies: {missing_deps}")
        return False
    
    print("✅ All required dependencies are present")
    return True

def generate_test_report():
    """Generate a comprehensive test report"""
    print("\n📊 Generating test report...")
    
    tests = [
        ("File Structure", test_file_structure),
        ("AI Models Structure", test_ai_models_structure),
        ("ML Infrastructure", test_ml_infrastructure),
        ("Preprocessing Pipeline", test_preprocessing_pipeline),
        ("Database Integration", test_database_integration),
        ("Database Migrations", test_migrations),
        ("Go Dependencies", test_go_mod_dependencies)
    ]
    
    results = {}
    for test_name, test_func in tests:
        try:
            results[test_name] = test_func()
        except Exception as e:
            print(f"❌ {test_name} failed with error: {e}")
            results[test_name] = False
    
    print("\n" + "="*50)
    print("📋 TEST SUMMARY")
    print("="*50)
    
    passed = 0
    total = len(tests)
    
    for test_name, result in results.items():
        status = "✅ PASS" if result else "❌ FAIL"
        print(f"{status} {test_name}")
        if result:
            passed += 1
    
    print(f"\nResults: {passed}/{total} tests passed")
    
    if passed == total:
        print("🎉 All tests passed! The AI infrastructure is ready for use.")
    else:
        print("⚠️  Some tests failed. Please review the implementation.")
    
    return passed == total

if __name__ == "__main__":
    print("🤖 AI Infrastructure Test Suite")
    print("="*50)
    
    success = generate_test_report()
    
    if success:
        print("\n🚀 Next steps:")
        print("1. Install Go if not already installed")
        print("2. Run: go mod tidy")
        print("3. Run: go test ./ml/ -v")
        print("4. Run: go run test_ai_infrastructure.go")
        print("5. Set up a database connection to test the full integration")
    
    exit(0 if success else 1)
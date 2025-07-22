# Project Structure

## Root Directory Layout
```
osmedeus/
├── main.go                 # Application entry point
├── go.mod/go.sum          # Go module dependencies
├── Makefile               # Build automation
├── Dockerfile*            # Container definitions
├── docker-compose.yml     # Multi-service orchestration
└── README.md              # Project documentation
```

## Core Packages

### `/cmd/` - Command Line Interface
- **Purpose**: Cobra-based CLI commands and subcommands
- **Key Files**: `root.go` (main CLI setup), `scan.go`, `server.go`, `cloud.go`
- **Pattern**: Each command has its own file with init() functions for flag registration

### `/core/` - Core Engine Logic
- **Purpose**: Central business logic and workflow processing
- **Key Components**: Configuration management, workflow parsing, execution engine
- **Key Files**: `config.go`, `flow.go`, `runner.go`, `parse.go`

### `/libs/` - Shared Libraries & Types
- **Purpose**: Common data structures, constants, and utility types
- **Key Files**: `options.go` (main Options struct), `version.go`, `flow.go`
- **Pattern**: Defines the primary `Options` struct used throughout the application

### `/server/` - Web API & UI
- **Purpose**: Fiber-based REST API and web interface
- **Key Files**: `router.go` (route definitions), `auth.go`, `scan.go`
- **Endpoints**: `/api/osmp/*` for core functionality, `/ui/*` for web interface

### `/execution/` - Task Execution
- **Purpose**: Handles external process execution, Git operations, notifications
- **Key Files**: `process.go`, `git.go`, `noti.go`, `scripts.go`

### `/provider/` - Cloud Provider Integration
- **Purpose**: Multi-cloud deployment and management
- **Supported**: AWS, DigitalOcean, Linode
- **Key Files**: `provider.go`, `provider_aws.go`, `building.go`

### `/database/` - Data Persistence
- **Purpose**: Database abstraction layer
- **Key Files**: `connect.go`, `models.go`, `select.go`

### `/utils/` - Utilities
- **Purpose**: Helper functions and common utilities
- **Key Files**: `helper.go`, `log.go`, `request.go`

## Configuration & Data Directories

### `/test-workflows/` - Example Workflows
- **Purpose**: Sample YAML workflow definitions
- **Key Files**: `general.yaml`, `parallel.yaml`, `serial.yaml`
- **Pattern**: YAML files defining scan routines and module sequences

### `/docker/` - Container Configurations
- **Purpose**: Specialized Docker configurations for different components
- **Subdirectories**: AI services, behavioral analyzers, ML inference

### `/config/`, `/data/`, `/logs/`, `/workspaces/`
- **Purpose**: Runtime directories for configuration, data storage, and results

## Coding Conventions

### Package Organization
- Each major feature area has its own package
- Shared types and constants go in `/libs/`
- CLI commands are separated by functionality in `/cmd/`

### File Naming
- Use descriptive names: `provider_aws.go`, `git_test.go`
- Test files follow `*_test.go` pattern
- Configuration files use `.yaml` extension

### Import Structure
- Standard library imports first
- Third-party imports second  
- Local project imports last
- Group imports with blank lines between categories

### Error Handling
- Use `utils.ErrorF()` for error logging with formatting
- Use `utils.InforF()` for informational messages
- Consistent error propagation up the call stack
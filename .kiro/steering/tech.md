# Technology Stack

## Core Technologies
- **Language**: Go 1.23+ (primary implementation language)
- **Web Framework**: Fiber v2 (high-performance HTTP framework)
- **CLI Framework**: Cobra (command-line interface)
- **Configuration**: Viper (configuration management)
- **Database**: SQLite (default), MySQL (optional), PostgreSQL (production)
- **Caching**: Redis (message queuing and caching)
- **Containerization**: Docker & Docker Compose

## Key Dependencies
- **HTTP Client**: Resty v2 (REST API calls)
- **Authentication**: JWT tokens, Basic Auth
- **Logging**: Logrus with prefixed formatter
- **Cloud Providers**: AWS SDK, DigitalOcean, Linode APIs
- **Notifications**: Telegram Bot API, Slack integration
- **Template Engine**: Pongo2 (workflow templating)
- **Concurrency**: Ants (goroutine pool)

## Build System & Commands

### Development Commands
```bash
# Build for development
make build

# Run the application
make run

# Format code
make fmt

# Run tests
make test
```

### Release Commands
```bash
# Build cross-platform releases
make release
```

### Docker Commands
```bash
# Build and start services
make docker-build
make docker-up

# View logs
make docker-logs

# Clean up
make docker-clean

# Database operations
make db-backup
make db-restore BACKUP_FILE=filename.sql
```

### Installation
```bash
# Quick install (requires root)
bash <(curl -fsSL https://raw.githubusercontent.com/osmedeus/osmedeus-base/master/install.sh)

# Build from source
go install -v github.com/j3ssie/osmedeus@latest
```

## Architecture Patterns
- **Modular Design**: Workflow-based execution with YAML configuration
- **CLI-First**: Primary interface through command-line with optional web UI
- **Cloud-Native**: Built for distributed execution across cloud providers
- **Plugin Architecture**: Extensible through external binaries and scripts
- **Event-Driven**: Notification system for scan completion and alerts
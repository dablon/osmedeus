TARGET   ?= osmedeus
GO       ?= go
GOFLAGS  ?= 
VERSION  := $(shell cat libs/version.go | grep 'VERSION =' | cut -d '"' -f 2)

build:
	go install
	go build -ldflags="-s -w" -tags netgo -trimpath -buildmode=pie -o dist/$(TARGET)

release:
	go install
	@echo "==> Clean up old builds"
	rm -rf ./dist/* ~/myGit/premium-$(TARGET)-base/dist/* ~/org-$(TARGET)/$(TARGET)-base/dist/*
	@echo "==> building binaries for for mac intel"
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -tags netgo -trimpath -buildmode=pie -o dist/$(TARGET)
	zip -9 -j dist/$(TARGET)-macos-amd64.zip dist/$(TARGET) && rm -rf ./dist/$(TARGET)
	@echo "==> building binaries for for mac M1 chip"
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -tags netgo -trimpath -buildmode=pie -o dist/$(TARGET)
	zip -9 -j dist/$(TARGET)-macos-arm64.zip dist/$(TARGET)&& rm -rf ./dist/$(TARGET)
	@echo "==> building binaries for linux intel build on mac"
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -tags netgo -trimpath -buildmode=pie -o dist/$(TARGET)
	zip -j dist/$(TARGET)-linux-amd64.zip dist/$(TARGET)&& rm -rf ./dist/$(TARGET)
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -tags netgo -trimpath -buildmode=pie -o dist/$(TARGET)
	zip -j dist/$(TARGET)-linux-arm64.zip dist/$(TARGET)&& rm -rf ./dist/$(TARGET)
	cp dist/* ~/myGit/premium-$(TARGET)-base/dist/
	cp dist/* ~/org-$(TARGET)/$(TARGET)-base/dist/
	@echo "==> Generating metadata info"
	$(TARGET) update --gen dist/public.json
	mv dist/$(TARGET)-macos-amd64.zip dist/$(TARGET)-$(VERSION)-macos-amd64.zip
	mv dist/$(TARGET)-macos-arm64.zip dist/$(TARGET)-$(VERSION)-macos-arm64.zip
	mv dist/$(TARGET)-linux-amd64.zip dist/$(TARGET)-$(VERSION)-linux-amd64.zip
	mv dist/$(TARGET)-linux-arm64.zip dist/$(TARGET)-$(VERSION)-linux-arm64.zip
run:
	$(GO) $(GOFLAGS) run *.go

fmt:
	$(GO) $(GOFLAGS) fmt ./...; \
	echo "Done."

test:
	$(GO) $(GOFLAGS) test ./... -v%

# Docker Compose Commands
docker-build:
	@echo "==> Building Docker images"
	docker-compose build

docker-up:
	@echo "==> Starting Osmedeus with Docker Compose"
	docker-compose up -d

docker-down:
	@echo "==> Stopping Osmedeus Docker containers"
	docker-compose down

docker-logs:
	@echo "==> Showing Docker container logs"
	docker-compose logs -f

docker-clean:
	@echo "==> Cleaning Docker containers and images"
	docker-compose down -v --remove-orphans
	docker system prune -f

docker-restart:
	@echo "==> Restarting Docker containers"
	docker-compose restart

docker-rebuild:
	@echo "==> Rebuilding and restarting Docker containers"
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

# Database commands
db-backup:
	@echo "==> Creating database backup"
	docker-compose exec postgres pg_dump -U osmedeus osmedeus > backup_$(shell date +%Y%m%d_%H%M%S).sql

db-restore:
	@echo "==> Restoring database (specify BACKUP_FILE=filename.sql)"
	@if [ -z "$(BACKUP_FILE)" ]; then echo "Please specify BACKUP_FILE=filename.sql"; exit 1; fi
	docker-compose exec -T postgres psql -U osmedeus -d osmedeus < $(BACKUP_FILE)

# Help command
docker-help:
	@echo "Docker Compose Commands:"
	@echo "  docker-build     - Build Docker images"
	@echo "  docker-up        - Start all services"
	@echo "  docker-down      - Stop all services"
	@echo "  docker-logs      - Show logs"
	@echo "  docker-clean     - Clean containers and images"
	@echo "  docker-restart   - Restart all services"
	@echo "  docker-rebuild   - Rebuild and restart"
	@echo "  db-backup        - Backup database"
	@echo "  db-restore       - Restore database"
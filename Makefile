.PHONY: dev dev-down build test lint migrate-up migrate-down

# Development
dev:
	docker compose up --build

dev-down:
	docker compose down

dev-clean:
	docker compose down -v

# Backend
build:
	cd server && go build -o bin/feather ./cmd/feather

test:
	cd server && go test ./...

test-integration:
	cd server && go test -tags=integration ./...

lint:
	cd server && go vet ./...

# Migrations
migrate-up:
	cd server && go run ./cmd/feather -migrate-up

migrate-down:
	cd server && go run ./cmd/feather -migrate-down

# Desktop
desktop-dev:
	cd desktop && npm run tauri dev

desktop-build:
	cd desktop && npm run tauri build

# Production
prod:
	docker compose -f docker-compose.prod.yml up --build -d

prod-down:
	docker compose -f docker-compose.prod.yml down

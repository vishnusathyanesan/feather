# Development Guide

## Prerequisites

- Go 1.23+
- Node.js 22+
- Docker & Docker Compose
- Rust (for Tauri desktop app)

## Getting Started

### 1. Start Infrastructure

```bash
# Start PostgreSQL, Redis, MinIO, and API server
make dev
```

This starts:
- PostgreSQL on port 5432
- Redis on port 6379
- MinIO on port 9000 (console: 9001)
- API server on port 8080 (with hot-reload via Air)

### 2. Verify Backend

```bash
curl http://localhost:8080/api/v1/health
# {"status":"ok"}
```

### 3. Desktop App Development

```bash
cd desktop
npm install
npm run tauri dev
```

## Project Structure

```
feather/
├── server/           # Go backend
│   ├── cmd/feather/  # Entry point
│   ├── internal/     # Business logic
│   └── migrations/   # SQL migrations
├── desktop/          # Tauri + React app
│   ├── src/          # React frontend
│   └── src-tauri/    # Rust backend
└── docs/             # Documentation
```

## Backend Architecture

The backend follows a layered architecture:

- **Handler** - HTTP request/response handling, validation
- **Service** - Business logic, permission checks
- **Repository** - Database queries

Each domain (auth, channel, message, etc.) has its own package with these layers.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `FEATHER_DATABASE_URL` | PostgreSQL connection string | `postgres://feather:feather@localhost:5432/feather?sslmode=disable` |
| `FEATHER_REDIS_URL` | Redis connection string | `redis://localhost:6379/0` |
| `FEATHER_JWT_SECRET` | JWT signing secret | `dev-secret-change-in-production` |
| `FEATHER_SERVER_PORT` | API server port | `8080` |
| `FEATHER_MINIO_ENDPOINT` | MinIO endpoint | `localhost:9000` |

## Running Tests

```bash
# Unit tests
make test

# Integration tests (requires Docker)
make test-integration
```

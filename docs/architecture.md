# Architecture

## Overview

Feather is a self-hosted Slack alternative with a Go backend and Tauri desktop app.

```
┌─────────────┐     ┌──────────────┐
│  Tauri App   │────▶│  Go API      │
│  (React)     │◀────│  Server      │
└─────────────┘     └──────────────┘
   REST + WS            │    │
                   ┌─────┘    └─────┐
                   ▼                ▼
             ┌──────────┐    ┌──────────┐
             │PostgreSQL │    │  Redis   │
             └──────────┘    └──────────┘
                                  │
                             ┌────┘
                             ▼
                       ┌──────────┐
                       │  MinIO   │
                       └──────────┘
```

## Backend Design

### Request Flow
1. HTTP request → Chi router → Middleware (auth, logging, CORS)
2. Handler validates input → Service applies business logic
3. Repository executes database queries → Response serialized as JSON

### Real-time System
- Single WebSocket connection per client
- Server sends events for messages, reactions, typing, presence
- Redis Pub/Sub enables horizontal scaling (multi-instance)
- Client auto-reconnects with exponential backoff

### Database
- PostgreSQL 16 with pgxpool connection pooling
- Sequential SQL migrations (golang-migrate)
- Full-text search via tsvector + GIN index

### Authentication
- JWT access tokens (15min, HMAC-SHA256)
- Refresh tokens (7 days, SHA-256 hashed in DB, rotated on use)
- Bcrypt password hashing (cost 12)

## Frontend Design

### State Management
- Zustand stores: auth, channels, messages, presence
- WebSocket events update stores in real-time

### Performance
- Virtualized message list (@tanstack/react-virtual)
- React.memo on MessageItem
- Debounced typing indicators (3s)
- Cursor-based message pagination

### Desktop Integration (Tauri)
- System tray with show/hide/quit
- Native notifications when window unfocused
- Offline cache via localStorage
- OS theme detection (light/dark)

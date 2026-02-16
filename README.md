# Feather

A self-hosted Slack alternative built with Go and React.

## Features

- **Channels** — Public and private channels with topic, description, and member management
- **Direct Messages** — 1:1 and group DMs reusing the full channel infrastructure (threads, reactions, search, files)
- **Threads** — Threaded replies on any message
- **Reactions** — Emoji reactions on messages
- **@Mentions** — `@username`, `@group`, `@channel`, `@here`, `@everyone` with notification tracking and autocomplete
- **User Groups** — Create named groups (e.g., `@engineering`) for bulk mentions
- **File Uploads** — Drag-and-drop file sharing via S3-compatible storage (MinIO)
- **Full-Text Search** — Search messages across channels with PostgreSQL full-text search
- **Audio/Video Calls** — 1:1 WebRTC calls with signaling over WebSocket
- **Workspace Invitations** — Invite users via shareable links with expiry and use limits
- **Webhooks** — Incoming webhooks for bot integrations
- **Google OAuth** — Sign in with Google alongside email/password auth
- **Real-Time** — WebSocket-powered live updates for messages, typing indicators, presence, calls
- **Dark Mode** — Automatic dark mode via system preference

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.23, chi router, pgx/v5, nhooyr.io/websocket |
| Frontend | React 18, TypeScript, Vite, Tailwind CSS, Zustand |
| Desktop | Tauri 2 |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| File Storage | MinIO (S3-compatible) |
| Reverse Proxy | Caddy 2 (auto-HTTPS) |

## Architecture

```
desktop/             React + Tauri desktop app
server/
  cmd/feather/       Entry point
  internal/
    auth/            Authentication (JWT + Google OAuth)
    channel/         Channel CRUD and membership
    message/         Messages and threads
    reaction/        Emoji reactions
    dm/              Direct messages (1:1 and group)
    mention/         @mention parsing and notifications
    usergroup/       User group management
    call/            Audio/video call signaling
    invitation/      Workspace invitations
    search/          Full-text search
    file/            File upload/download (MinIO)
    webhook/         Incoming webhooks
    websocket/       WebSocket hub and client management
    middleware/      Auth, CORS, logging, rate limiting
    model/           Shared domain types
    config/          Configuration (Viper)
    audit/           Audit logging
    server/          HTTP server, routing, service wiring
  migrations/        PostgreSQL migrations (000001-000018)
deploy/              Production deployment scripts
```

Each domain package follows the **Handler → Service → Repository** pattern.

## Quick Start (Development)

**Prerequisites:** Docker and Docker Compose

```bash
git clone https://github.com/vishnusathyanesan/feather.git
cd feather

# Start backend services
docker compose up -d

# Start frontend dev server
cd desktop
npm install
npm run dev
```

The API runs at `http://localhost:8080` and the frontend at `http://localhost:1420`.

## Production Deployment (GCloud Free Tier)

Feather runs on a single GCloud e2-micro VM (1 GB RAM, always free) with Caddy for auto-HTTPS.

### Prerequisites

- GCloud account with billing enabled (free tier — no charges for e2-micro)
- DuckDNS account for a free subdomain (https://www.duckdns.org)

### 1. Create the VM

```bash
gcloud compute instances create feather-server \
  --zone=us-central1-a \
  --machine-type=e2-micro \
  --image-family=debian-12 \
  --image-project=debian-cloud \
  --boot-disk-size=30GB \
  --tags=http-server,https-server

gcloud compute firewall-rules create allow-http --allow=tcp:80 --target-tags=http-server
gcloud compute firewall-rules create allow-https --allow=tcp:443 --target-tags=https-server
```

### 2. SSH in and set up

```bash
gcloud compute ssh feather-server --zone=us-central1-a
```

Run the setup script (installs Docker, 2 GB swap, DuckDNS cron):

```bash
git clone https://github.com/vishnusathyanesan/feather.git
cd feather
sudo ./deploy/setup-gcloud.sh
```

### 3. Configure

```bash
cp .env.prod.example .env.prod
nano .env.prod
```

Fill in your domain, DuckDNS token, and generate passwords (see `.env.prod.example` for details).

### 4. Build and deploy

```bash
# Install Node.js for frontend build
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo bash -
sudo apt-get install -y nodejs
cd desktop && npm install && cd ..

# Build frontend
./deploy/build-frontend.sh your-app.duckdns.org

# Deploy
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build
```

Your app will be live at `https://your-app.duckdns.org` once Caddy obtains the TLS certificate.

### Memory Budget (e2-micro)

| Service | Limit |
|---------|-------|
| PostgreSQL | 150 MB |
| MinIO | 200 MB |
| Go API | 128 MB |
| Redis | 64 MB |
| Caddy | 64 MB |
| **Total** | **606 MB** |

The remaining ~400 MB is covered by 2 GB swap for peak loads.

## CI/CD (Automated Deployment)

Push to `main` triggers a GitHub Actions workflow that:

1. Builds the Go API Docker image on GitHub runners (~2 min vs 33 min on e2-micro)
2. Pushes it to GitHub Container Registry (GHCR)
3. Builds the frontend with production URLs
4. SSHs into the VM, pulls the new image, and restarts

### Setup

Add these GitHub repository secrets (`Settings > Secrets > Actions`):

| Secret | Value |
|--------|-------|
| `VM_HOST` | VM's external IP (find with `gcloud compute instances describe feather-server --zone=us-central1-a --format='get(networkInterfaces[0].accessConfigs[0].natIP)'`) |
| `VM_USER` | SSH username (e.g., `sathuvish`) |
| `VM_SSH_KEY` | Private SSH key for the VM |
| `GHCR_TOKEN` | GitHub PAT with `read:packages` scope |

Generate an SSH key for deployment:
```bash
ssh-keygen -t ed25519 -f ~/.ssh/feather-deploy -N ""
# Add public key to VM
gcloud compute ssh feather-server --zone=us-central1-a --command="echo '$(cat ~/.ssh/feather-deploy.pub)' >> ~/.ssh/authorized_keys"
# Copy private key as VM_SSH_KEY secret
cat ~/.ssh/feather-deploy
```

### Manual Deploy

```bash
# Deploy using pre-built GHCR image (fast)
./deploy/deploy.sh

# Or build on VM (slow, no GHCR needed)
./deploy/deploy.sh --build
```

## Configuration

Configuration is via `server/config.yaml` with environment variable overrides using the `FEATHER_` prefix:

| Variable | Description |
|----------|-------------|
| `FEATHER_DATABASE_URL` | PostgreSQL connection string |
| `FEATHER_REDIS_URL` | Redis connection string |
| `FEATHER_JWT_SECRET` | JWT signing secret |
| `FEATHER_SERVER_PORT` | API server port (default: 8080) |
| `FEATHER_MINIO_ENDPOINT` | MinIO endpoint |
| `FEATHER_MINIO_ACCESS_KEY` | MinIO access key |
| `FEATHER_MINIO_SECRET_KEY` | MinIO secret key |
| `FEATHER_OAUTH_GOOGLE_CLIENT_ID` | Google OAuth client ID (optional) |
| `FEATHER_WEBRTC_ENABLED` | Enable WebRTC calls (default: true) |

## API Endpoints

### Auth
- `POST /api/v1/auth/register` — Register
- `POST /api/v1/auth/login` — Login
- `POST /api/v1/auth/refresh` — Refresh token
- `POST /api/v1/auth/oauth/google` — Google OAuth
- `GET /api/v1/auth/me` — Current user

### Channels
- `GET /api/v1/channels` — List channels
- `POST /api/v1/channels` — Create channel
- `POST /api/v1/channels/{id}/join` — Join channel
- `GET /api/v1/channels/{id}/messages` — List messages
- `POST /api/v1/channels/{id}/messages` — Send message

### Direct Messages
- `POST /api/v1/dms` — Create/get 1:1 DM
- `POST /api/v1/dms/group` — Create group DM
- `GET /api/v1/dms` — List DM conversations

### Mentions
- `GET /api/v1/mentions` — Unread mentions
- `POST /api/v1/mentions/read` — Mark read

### User Groups
- `POST /api/v1/groups` — Create group
- `GET /api/v1/groups` — List groups (`?q=` for search)

### Users
- `GET /api/v1/users` — List users (`?q=` for search)

### Calls
- `GET /api/v1/channels/{id}/calls` — Call history
- `GET /api/v1/calls/active` — Active call
- `GET /api/v1/rtc/config` — WebRTC ICE config

### Other
- `GET /api/v1/search?q=` — Full-text search
- `POST /api/v1/invitations` — Create invitation
- `GET /api/v1/ws` — WebSocket connection

## License

MIT

#!/usr/bin/env bash
# =============================================================================
# Deploy Feather to the GCloud VM
#
# Usage:
#   ./deploy/deploy.sh                  # Deploy using pre-built GHCR image
#   ./deploy/deploy.sh --build          # Build on VM (slow on e2-micro)
#
# Prerequisites:
#   - SSH access to the VM (gcloud compute ssh or SSH key)
#   - .env.prod configured on the VM
#   - For GHCR: docker login ghcr.io on the VM
# =============================================================================
set -euo pipefail

# Configuration — override via environment or edit here
VM_HOST="${VM_HOST:-}"
VM_USER="${VM_USER:-$(whoami)}"
VM_ZONE="${VM_ZONE:-us-central1-a}"
VM_NAME="${VM_NAME:-feather-server}"
DOMAIN="${DOMAIN:-feather-chat.duckdns.org}"
BUILD_FLAG=""

if [ "${1:-}" = "--build" ]; then
    BUILD_FLAG="--build"
fi

# Determine SSH command
if [ -n "$VM_HOST" ]; then
    SSH_CMD="ssh ${VM_USER}@${VM_HOST}"
    SCP_CMD="scp"
    SCP_TARGET="${VM_USER}@${VM_HOST}"
else
    SSH_CMD="gcloud compute ssh ${VM_NAME} --zone=${VM_ZONE} --command"
    SCP_CMD="gcloud compute scp --zone=${VM_ZONE}"
    SCP_TARGET="${VM_NAME}"
fi

echo "=== Feather Deploy ==="

# Step 1: Build frontend locally
echo "[1/4] Building frontend..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"
./deploy/build-frontend.sh "$DOMAIN"

# Step 2: Copy frontend to VM
echo "[2/4] Uploading frontend to VM..."
if [ -n "$VM_HOST" ]; then
    scp -r desktop/dist/ "${SCP_TARGET}:~/feather/desktop/"
else
    gcloud compute scp --zone="$VM_ZONE" --recurse desktop/dist/ "${SCP_TARGET}:~/feather/desktop/"
fi

# Step 3: Pull latest code and images on VM
echo "[3/4] Updating server..."
if [ -n "$VM_HOST" ]; then
    ssh "${VM_USER}@${VM_HOST}" bash -s "$BUILD_FLAG" <<'REMOTE'
        set -euo pipefail
        cd ~/feather
        git pull origin main
        if [ "${1:-}" = "--build" ]; then
            docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build
        else
            docker compose -f docker-compose.prod.yml --env-file .env.prod pull api
            docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --no-build
        fi
REMOTE
else
    gcloud compute ssh "$VM_NAME" --zone="$VM_ZONE" --command="cd ~/feather && git pull origin main && docker compose -f docker-compose.prod.yml --env-file .env.prod pull api && docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --no-build"
fi

# Step 4: Verify
echo "[4/4] Verifying deployment..."
sleep 10
if [ -n "$VM_HOST" ]; then
    ssh "${VM_USER}@${VM_HOST}" "docker compose -f ~/feather/docker-compose.prod.yml --env-file ~/feather/.env.prod ps && curl -sf http://localhost:8080/api/v1/health && echo ''"
else
    gcloud compute ssh "$VM_NAME" --zone="$VM_ZONE" --command="docker compose -f ~/feather/docker-compose.prod.yml --env-file ~/feather/.env.prod ps && curl -sf http://localhost:8080/api/v1/health && echo ''"
fi

echo ""
echo "=== Deploy complete! ==="
echo "Live at: https://${DOMAIN}"

#!/usr/bin/env bash
# =============================================================================
# Build the frontend for production deployment
#
# Usage:
#   ./deploy/build-frontend.sh your-app.duckdns.org
# =============================================================================
set -euo pipefail

DOMAIN="${1:?Usage: $0 <domain>}"
API_URL="https://${DOMAIN}/api/v1"
WS_URL="wss://${DOMAIN}/api/v1/ws"
GOOGLE_CLIENT_ID="${VITE_GOOGLE_CLIENT_ID:-${2:-}}"

echo "Building frontend for ${DOMAIN}..."
echo "  API_URL:           ${API_URL}"
echo "  WS_URL:            ${WS_URL}"
echo "  GOOGLE_CLIENT_ID:  ${GOOGLE_CLIENT_ID:-(not set)}"

cd "$(dirname "$0")/../desktop"

VITE_API_URL="${API_URL}" \
VITE_WS_URL="${WS_URL}" \
VITE_GOOGLE_CLIENT_ID="${GOOGLE_CLIENT_ID}" \
npm run build

echo ""
echo "Frontend built successfully in desktop/dist/"
echo "Deploy with: docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build"

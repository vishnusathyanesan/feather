#!/usr/bin/env bash
# =============================================================================
# Feather — GCloud e2-micro Setup Script
#
# Run this on a fresh Debian 12 / Ubuntu 22.04 e2-micro VM.
# Prerequisites: SSH into the VM, clone the repo, then run this script.
#
# Usage:
#   chmod +x deploy/setup-gcloud.sh
#   sudo ./deploy/setup-gcloud.sh
# =============================================================================
set -euo pipefail

echo "=== Feather Production Setup ==="
echo ""

# ---- 1. Add 2 GB swap (e2-micro only has 1 GB RAM) ----
if [ ! -f /swapfile ]; then
    echo "[1/6] Creating 2 GB swap file..."
    fallocate -l 2G /swapfile
    chmod 600 /swapfile
    mkswap /swapfile
    swapon /swapfile
    echo '/swapfile none swap sw 0 0' >> /etc/fstab
    # Prefer RAM, use swap only when needed
    sysctl vm.swappiness=10
    echo 'vm.swappiness=10' >> /etc/sysctl.conf
    echo "    Swap enabled."
else
    echo "[1/6] Swap already configured, skipping."
fi

# ---- 2. Install Docker ----
if ! command -v docker &> /dev/null; then
    echo "[2/6] Installing Docker..."
    apt-get update -qq
    apt-get install -y -qq ca-certificates curl gnupg
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/$(. /etc/os-release && echo "$ID")/gpg \
        | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    chmod a+r /etc/apt/keyrings/docker.gpg
    echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
        https://download.docker.com/linux/$(. /etc/os-release && echo "$ID") \
        $(. /etc/os-release && echo "$VERSION_CODENAME") stable" \
        > /etc/apt/sources.list.d/docker.list
    apt-get update -qq
    apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-compose-plugin
    systemctl enable docker
    echo "    Docker installed."
else
    echo "[2/6] Docker already installed, skipping."
fi

# ---- 3. Add current user to docker group ----
REAL_USER="${SUDO_USER:-$USER}"
if ! id -nG "$REAL_USER" | grep -qw docker; then
    echo "[3/6] Adding $REAL_USER to docker group..."
    usermod -aG docker "$REAL_USER"
    echo "    Added. You may need to log out and back in."
else
    echo "[3/6] $REAL_USER already in docker group, skipping."
fi

# ---- 4. Set up DuckDNS cron ----
echo "[4/6] Setting up DuckDNS dynamic DNS update..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

if [ -f "$PROJECT_DIR/.env.prod" ]; then
    # shellcheck disable=SC1091
    set -a; source "$PROJECT_DIR/.env.prod"; set +a

    if [ -n "${DUCKDNS_TOKEN:-}" ] && [ -n "${DOMAIN:-}" ]; then
        SUBDOMAIN="${DOMAIN%.duckdns.org}"
        mkdir -p /opt/feather
        cat > /opt/feather/duckdns-update.sh <<DUCKEOF
#!/bin/bash
curl -s "https://www.duckdns.org/update?domains=${SUBDOMAIN}&token=${DUCKDNS_TOKEN}&ip=" > /dev/null
DUCKEOF
        chmod +x /opt/feather/duckdns-update.sh
        # Run every 5 minutes
        (crontab -l 2>/dev/null | grep -v duckdns-update; echo "*/5 * * * * /opt/feather/duckdns-update.sh") | crontab -
        # Run once now
        /opt/feather/duckdns-update.sh
        echo "    DuckDNS cron installed for ${SUBDOMAIN}.duckdns.org"
    else
        echo "    WARNING: DUCKDNS_TOKEN or DOMAIN not set in .env.prod — skipping DuckDNS."
    fi
else
    echo "    WARNING: .env.prod not found — skipping DuckDNS."
    echo "    Copy .env.prod.example to .env.prod and fill in values first."
fi

# ---- 5. Configure firewall ----
echo "[5/6] Configuring firewall (allow SSH, HTTP, HTTPS)..."
if command -v ufw &> /dev/null; then
    ufw allow OpenSSH
    ufw allow 80/tcp
    ufw allow 443/tcp
    ufw --force enable
    echo "    UFW configured."
else
    echo "    UFW not found — make sure GCloud firewall rules allow tcp:80 and tcp:443."
fi

# ---- 6. Done ----
echo "[6/6] Setup complete!"
echo ""
echo "=== Next Steps ==="
echo "1. Copy .env.prod.example to .env.prod and fill in all values:"
echo "     cp .env.prod.example .env.prod"
echo "     nano .env.prod"
echo ""
echo "2. Build the frontend (on your local machine or CI):"
echo "     cd desktop && VITE_API_URL=https://YOUR_DOMAIN/api/v1 npm run build"
echo "     Then commit/push desktop/dist/ or scp it to the VM."
echo ""
echo "3. Deploy with Docker Compose:"
echo "     cd $PROJECT_DIR"
echo "     docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build"
echo ""
echo "4. Check logs:"
echo "     docker compose -f docker-compose.prod.yml logs -f"
echo ""
echo "5. Your app will be live at https://\$DOMAIN once Caddy obtains the TLS certificate."

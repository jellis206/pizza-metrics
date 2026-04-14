#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="/opt/metrics"

# Install Docker if not present
if ! command -v docker &>/dev/null; then
  echo "Installing Docker..."
  sudo apt-get update
  sudo apt-get install -y docker.io docker-compose-v2 git
  sudo systemctl enable docker
  sudo systemctl start docker
  sudo usermod -aG docker "$(whoami)"
  echo "Docker installed. You may need to log out and back in for group changes."
fi

if [ ! -d "$INSTALL_DIR/.git" ]; then
  echo "Cloning pizza-metrics to $INSTALL_DIR..."
  sudo git clone https://github.com/jellis206/pizza-metrics.git "$INSTALL_DIR"
  sudo chown -R "$(whoami):$(whoami)" "$INSTALL_DIR"
fi

cd "$INSTALL_DIR"

if [ ! -f .env ]; then
  cp .env.example .env
  echo "Created .env from .env.example — edit it with real values if needed."
fi

# Ensure data directories exist with correct ownership
sudo mkdir -p /data/grafana /data/victoria /data/victorialogs /data/caddy/data /data/caddy/config
sudo chown -R 472:472 /data/grafana

sudo docker compose up -d --build
echo "All services started. Run 'docker compose ps' to verify."

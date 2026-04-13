#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="/opt/metrics"

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

docker-compose up -d --build
echo "All services started. Run 'docker-compose ps' to verify."

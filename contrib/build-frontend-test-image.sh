#!/usr/bin/env bash
# Build a lightweight nginx Docker image for testing frontend changes locally.
# Proxies /api/ to the existing Vikunja backend at localhost:3456.
#
# Usage:
#   ./bin/build-frontend-test-image.sh [tag]
#   docker run --rm -p 3457:3457 --add-host=host.docker.internal:host-gateway vikunja-fe-test:<tag>
#
# Then open http://localhost:3457

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TAG="${1:-latest}"
IMAGE="vikunja-fe-test:${TAG}"

echo "==> Building frontend..."
(cd "$REPO_ROOT/frontend" && ./node_modules/.bin/vite build)

echo "==> Assembling Docker context in /tmp/vikunja-fe-test..."
TMPDIR=/tmp/vikunja-fe-test-ctx
rm -rf "$TMPDIR"
mkdir -p "$TMPDIR"
trap 'rm -rf "$TMPDIR"' EXIT

cp -r "$REPO_ROOT/frontend/dist" "$TMPDIR/dist"
cp "$REPO_ROOT/Dockerfile.frontend-test.nginx.conf" "$TMPDIR/default.conf"

cat > "$TMPDIR/Dockerfile" <<'DOCKERFILE'
FROM nginx:1.27-alpine
COPY dist /usr/share/nginx/html
COPY default.conf /etc/nginx/conf.d/default.conf
EXPOSE 3457
DOCKERFILE

echo "==> Building Docker image ${IMAGE}..."
DOCKER_BUILDKIT=0 docker build -t "$IMAGE" "$TMPDIR"

echo ""
echo "==> Done. Run with:"
echo "    docker run --rm -p 3457:3457 --add-host=host.docker.internal:host-gateway ${IMAGE}"
echo "    Then open: http://localhost:3457"

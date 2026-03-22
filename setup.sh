#!/bin/bash
# ──────────────────────────────────────────────────────────
# Vento Core — Setup Script
# Installs Go (if needed), dependencies, and starts services
# ──────────────────────────────────────────────────────────
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CORE_DIR="$SCRIPT_DIR/producer/vento_core"

echo "🚀 Vento Core — Setup"
echo "====================="

# ── 1. Check Go ──────────────────────────
if ! command -v go &> /dev/null; then
    echo ""
    echo "⚠️  Go is not installed. Install it with:"
    echo ""
    echo "   # Ubuntu/Debian:"
    echo "   sudo apt update && sudo apt install -y golang-go"
    echo ""
    echo "   # Or download from https://go.dev/dl/ :"
    echo "   wget https://go.dev/dl/go1.22.5.linux-amd64.tar.gz"
    echo "   sudo tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz"
    echo "   export PATH=\$PATH:/usr/local/go/bin"
    echo ""
    exit 1
fi

echo "✅ Go $(go version | awk '{print $3}')"

# ── 2. Check Docker ──────────────────────
if ! command -v docker &> /dev/null; then
    echo ""
    echo "⚠️  Docker is not installed. Install it from:"
    echo "   https://docs.docker.com/engine/install/ubuntu/"
    echo ""
    exit 1
fi

echo "✅ Docker $(docker --version | awk '{print $3}' | tr -d ',')"

# ── 3. Start PostgreSQL ──────────────────
echo ""
echo "📦 Starting PostgreSQL..."
cd "$SCRIPT_DIR/devops"
docker compose up -d
echo "✅ PostgreSQL running on localhost:5432"

# ── 4. Install Go dependencies ───────────
echo ""
echo "📥 Downloading Go dependencies..."
cd "$CORE_DIR"
go mod tidy
echo "✅ Dependencies installed"

# ── 5. Build ─────────────────────────────
echo ""
echo "🔨 Building..."
go build ./...
echo "✅ Build successful"

# ── 6. Run ───────────────────────────────
echo ""
echo "🚀 Starting Vento Core server..."
go run cmd/server/main.go

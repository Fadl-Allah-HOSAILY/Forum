#!/bin/zsh

set -e

echo "🔍 Checking for existing Docker installation..."

if command -v docker &>/dev/null; then
    echo "⚠️ Docker is already installed at: $(which docker)"
    docker --version

    if docker info 2>/dev/null | grep -q "rootless"; then
        echo "ℹ️ Docker is already running in rootless mode. No need to reinstall."
        exit 0
    else
        echo "🛑 System Docker (root) is already installed."
        read "REPLY?Do you still want to install Docker rootless? (y/N) "
        if [[ ! "$REPLY" =~ ^[Yy]$ ]]; then
            echo "❌ Aborting rootless Docker installation."
            exit 0
        fi
    fi
fi

echo "📦 Installing Docker (rootless)..."

# Download and install Docker rootless
curl -fsSL https://get.docker.com/rootless | sh

echo "✅ Docker rootless installed."

# Run rootless setup, skip iptables check
echo "⚙️ Running dockerd-rootless-setuptool.sh with --skip-iptables..."
~/bin/dockerd-rootless-setuptool.sh install --skip-iptables || echo "⚠️ Skipping setup tool errors."

# Set up environment variables if not already added
echo "⚙️ Setting up environment variables..."

grep -qxF 'export PATH=$HOME/bin:$PATH' ~/.zshrc || echo 'export PATH=$HOME/bin:$PATH' >> ~/.zshrc
grep -qxF 'export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/docker.sock' ~/.zshrc || echo 'export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/docker.sock' >> ~/.zshrc
grep -qxF 'export PATH=$HOME/.docker/cli-plugins:$PATH' ~/.zshrc || echo 'export PATH=$HOME/.docker/cli-plugins:$PATH' >> ~/.zshrc

export PATH=$HOME/bin:$PATH
export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/docker.sock
export PATH=$HOME/.docker/cli-plugins:$PATH

echo "✅ Environment configured."

# Start the rootless Docker daemon in background
echo "🚀 Starting Docker daemon (rootless) in background..."
nohup dockerd-rootless.sh > ~/docker-rootless.log 2>&1 &

echo "✅ Docker daemon started in background (log: ~/docker-rootless.log)"

# Verify installation
echo "🔍 Verifying installation..."
docker --version || echo "⚠️ Docker not found in PATH yet. Try restarting your terminal."
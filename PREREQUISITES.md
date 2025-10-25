# RTA CCTV System - Prerequisites & Setup Guide

## Overview

This document lists all software, tools, and plugins required to set up and run the RTA CCTV System on a new machine.

**Last Updated**: 2025-10-26

---

## Table of Contents

1. [System Requirements](#system-requirements)
2. [Required Software](#required-software)
3. [Development Tools](#development-tools)
4. [Runtime Dependencies](#runtime-dependencies)
5. [Optional Tools](#optional-tools)
6. [Installation Guide](#installation-guide)
7. [Verification](#verification)

---

## System Requirements

### Minimum Hardware

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| **CPU** | 8 cores | 16+ cores |
| **RAM** | 16 GB | 32 GB |
| **Storage** | 100 GB SSD | 500 GB+ SSD (for local storage mode) |
| **Network** | 1 Gbps | 10 Gbps |

### Supported Operating Systems

- ✅ **Linux** (Ubuntu 22.04 LTS or later) - **Recommended for production**
- ✅ **macOS** (11.0 Big Sur or later) - Development only
- ✅ **Windows** (Windows 10/11 with WSL2) - Development only

**Note**: For Windows, WSL2 with Ubuntu is required for Docker compatibility.

---

## Required Software

### 1. Docker & Docker Compose

**Purpose**: Container runtime for all services

**Versions**:
- Docker Engine: `>= 24.0.0`
- Docker Compose: `>= 2.20.0`

**Installation**:

<details>
<summary><strong>Ubuntu/Debian</strong></summary>

```bash
# Remove old versions
sudo apt-get remove docker docker-engine docker.io containerd runc

# Install dependencies
sudo apt-get update
sudo apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    lsb-release

# Add Docker's official GPG key
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Set up repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Add user to docker group (logout/login required)
sudo usermod -aG docker $USER

# Start Docker service
sudo systemctl enable docker
sudo systemctl start docker
```

**Verify**:
```bash
docker --version
docker compose version
```

</details>

<details>
<summary><strong>macOS</strong></summary>

**Option 1: Docker Desktop** (Recommended)
1. Download from https://www.docker.com/products/docker-desktop
2. Install Docker Desktop.dmg
3. Launch Docker Desktop from Applications

**Option 2: Homebrew**
```bash
brew install docker docker-compose
```

**Verify**:
```bash
docker --version
docker compose version
```

</details>

<details>
<summary><strong>Windows (WSL2)</strong></summary>

1. **Install WSL2**:
   ```powershell
   # Run in PowerShell as Administrator
   wsl --install
   wsl --set-default-version 2
   ```

2. **Install Docker Desktop**:
   - Download from https://www.docker.com/products/docker-desktop
   - During installation, ensure "Use WSL 2 instead of Hyper-V" is checked
   - Install and restart

3. **Configure WSL2**:
   - Launch Docker Desktop
   - Go to Settings → Resources → WSL Integration
   - Enable integration with your WSL2 distro (Ubuntu)

**Verify (in WSL2 terminal)**:
```bash
docker --version
docker compose version
```

</details>

---

### 2. Git

**Purpose**: Version control

**Version**: `>= 2.30.0`

**Installation**:

```bash
# Ubuntu/Debian
sudo apt-get install -y git

# macOS
brew install git

# Windows (in WSL2)
sudo apt-get install -y git
```

**Verify**:
```bash
git --version
```

**Configuration**:
```bash
git config --global user.name "Your Name"
git config --global user.email "your.email@rta.ae"
```

---

### 3. Go (Golang)

**Purpose**: Backend services (go-api, vms-service, etc.)

**Version**: `>= 1.23.0`

**Installation**:

<details>
<summary><strong>Ubuntu/Debian</strong></summary>

```bash
# Download Go
wget https://go.dev/dl/go1.23.1.linux-amd64.tar.gz

# Remove old installation
sudo rm -rf /usr/local/go

# Extract
sudo tar -C /usr/local -xzf go1.23.1.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc

# Clean up
rm go1.23.1.linux-amd64.tar.gz
```

</details>

<details>
<summary><strong>macOS</strong></summary>

```bash
# Using Homebrew
brew install go@1.23

# Or download from https://go.dev/dl/
```

</details>

**Verify**:
```bash
go version
# Expected: go version go1.23.1 linux/amd64
```

**Install Go Tools**:
```bash
# Air (live reload for development)
go install github.com/cosmtrek/air@latest

# golangci-lint (code linting)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

---

### 4. Node.js & npm

**Purpose**: Dashboard (React frontend)

**Version**: `>= 20.x LTS`

**Installation**:

<details>
<summary><strong>Ubuntu/Debian (using nvm - Recommended)</strong></summary>

```bash
# Install nvm (Node Version Manager)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.5/install.sh | bash

# Reload shell
source ~/.bashrc

# Install Node.js LTS
nvm install 20
nvm use 20
nvm alias default 20
```

</details>

<details>
<summary><strong>macOS</strong></summary>

```bash
# Using Homebrew
brew install node@20

# Or using nvm
brew install nvm
nvm install 20
nvm use 20
```

</details>

**Verify**:
```bash
node --version   # Expected: v20.x.x
npm --version    # Expected: 10.x.x
```

**Install Yarn (Optional but Recommended)**:
```bash
npm install -g yarn
yarn --version
```

---

### 5. Python

**Purpose**: Object detection service, utility scripts

**Version**: `>= 3.11`

**Installation**:

```bash
# Ubuntu/Debian
sudo apt-get install -y python3.11 python3.11-venv python3-pip

# macOS
brew install python@3.11

# Verify
python3 --version
pip3 --version
```

**Create Virtual Environment**:
```bash
cd services/object-detection
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install --upgrade pip
```

---

### 6. Make

**Purpose**: Build automation

**Installation**:

```bash
# Ubuntu/Debian
sudo apt-get install -y build-essential

# macOS (comes with Xcode Command Line Tools)
xcode-select --install

# Verify
make --version
```

---

## Development Tools

### 1. Visual Studio Code (Recommended IDE)

**Download**: https://code.visualstudio.com/

**Recommended Extensions**:
- **Go** (`golang.go`) - Go language support
- **Docker** (`ms-azuretools.vscode-docker`) - Docker integration
- **ESLint** (`dbaeumer.vscode-eslint`) - JavaScript linting
- **Prettier** (`esbenp.prettier-vscode`) - Code formatting
- **YAML** (`redhat.vscode-yaml`) - YAML syntax support
- **Remote - WSL** (`ms-vscode-remote.remote-wsl`) - WSL integration (Windows only)

**Install Extensions**:
```bash
code --install-extension golang.go
code --install-extension ms-azuretools.vscode-docker
code --install-extension dbaeumer.vscode-eslint
code --install-extension esbenp.prettier-vscode
code --install-extension redhat.vscode-yaml
```

---

### 2. Database Client Tools

#### PostgreSQL Client (psql)

```bash
# Ubuntu/Debian
sudo apt-get install -y postgresql-client

# macOS
brew install postgresql

# Verify
psql --version
```

#### Redis CLI (for Valkey)

```bash
# Ubuntu/Debian
sudo apt-get install -y redis-tools

# macOS
brew install redis

# Verify
redis-cli --version
```

---

### 3. API Testing Tools

#### cURL

```bash
# Ubuntu/Debian (usually pre-installed)
sudo apt-get install -y curl

# macOS (pre-installed)
curl --version
```

#### HTTPie (Optional, user-friendly alternative)

```bash
# Ubuntu/Debian
sudo apt-get install -y httpie

# macOS
brew install httpie

# Verify
http --version
```

#### Postman (GUI Tool)

Download from https://www.postman.com/downloads/

---

## Runtime Dependencies

### 1. FFmpeg

**Purpose**: Video processing, transcoding, playback

**Version**: `>= 6.0`

**Installation**:

```bash
# Ubuntu/Debian
sudo apt-get install -y ffmpeg

# macOS
brew install ffmpeg

# Verify
ffmpeg -version
```

**Required Components**:
- libx264 (H.264 encoding)
- libx265 (H.265 encoding)
- libvpx (VP8/VP9 encoding)
- libopus (Opus audio codec)

---

### 2. GStreamer (For WHIP Pusher Development)

**Purpose**: WHIP pusher development/testing outside Docker

**Version**: `>= 1.20.0`

**Installation**:

<details>
<summary><strong>Ubuntu/Debian</strong></summary>

```bash
sudo apt-get install -y \
    gstreamer1.0-tools \
    gstreamer1.0-plugins-base \
    gstreamer1.0-plugins-good \
    gstreamer1.0-plugins-bad \
    gstreamer1.0-plugins-ugly \
    gstreamer1.0-libav \
    gstreamer1.0-nice \
    libgstreamer1.0-dev \
    libgstreamer-plugins-base1.0-dev \
    libgstreamer-plugins-bad1.0-dev
```

**Verify**:
```bash
gst-launch-1.0 --version
gst-inspect-1.0 --version
```

</details>

<details>
<summary><strong>macOS</strong></summary>

```bash
brew install gstreamer gst-plugins-base gst-plugins-good \
    gst-plugins-bad gst-plugins-ugly gst-libav
```

</details>

---

### 3. Rust (For GStreamer Plugin Development)

**Purpose**: Building gst-plugins-rs (whipsink)

**Version**: `>= 1.70.0`

**Installation**:

```bash
# All platforms
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Follow prompts, then reload shell
source $HOME/.cargo/env

# Verify
rustc --version
cargo --version
```

**Install Required Components**:
```bash
rustup component add rustfmt clippy
```

---

## Optional Tools

### 1. K6 (Load Testing)

**Purpose**: Performance testing, load simulation

**Installation**:

```bash
# Ubuntu/Debian
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# macOS
brew install k6
```

---

### 2. VLC Media Player

**Purpose**: Testing RTSP streams, video playback

**Download**: https://www.videolan.org/vlc/

```bash
# Ubuntu/Debian
sudo apt-get install -y vlc

# macOS
brew install --cask vlc
```

---

### 3. Wireshark

**Purpose**: Network debugging, packet analysis

**Download**: https://www.wireshark.org/download.html

```bash
# Ubuntu/Debian
sudo apt-get install -y wireshark

# macOS
brew install --cask wireshark
```

---

## Installation Guide

### Quick Start (Fresh Ubuntu 22.04)

```bash
# 1. Update system
sudo apt-get update && sudo apt-get upgrade -y

# 2. Install base tools
sudo apt-get install -y curl wget git build-essential

# 3. Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
newgrp docker

# 4. Install Go
wget https://go.dev/dl/go1.23.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.1.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc
rm go1.23.1.linux-amd64.tar.gz

# 5. Install Node.js
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.5/install.sh | bash
source ~/.bashrc
nvm install 20
nvm use 20

# 6. Install Python
sudo apt-get install -y python3.11 python3.11-venv python3-pip

# 7. Install FFmpeg
sudo apt-get install -y ffmpeg

# 8. Install PostgreSQL and Redis clients
sudo apt-get install -y postgresql-client redis-tools

# 9. Clone repository
git clone https://github.com/rta/cctv-system.git
cd cctv-system

# 10. Copy environment template
cp .env.example .env
# Edit .env with your configuration
nano .env

# 11. Build WHIP pusher image
cd services/whip-pusher
docker build -t whip-pusher:latest .
cd ../..

# 12. Start services
docker-compose up -d

# 13. Verify all services are running
docker-compose ps
```

---

## Verification

### Verify All Prerequisites

Run this script to check all prerequisites:

```bash
#!/bin/bash

echo "=== RTA CCTV System Prerequisites Verification ==="
echo ""

# Docker
echo -n "Docker: "
if command -v docker &> /dev/null; then
    echo "✅ $(docker --version)"
else
    echo "❌ Not installed"
fi

# Docker Compose
echo -n "Docker Compose: "
if docker compose version &> /dev/null; then
    echo "✅ $(docker compose version)"
else
    echo "❌ Not installed"
fi

# Git
echo -n "Git: "
if command -v git &> /dev/null; then
    echo "✅ $(git --version)"
else
    echo "❌ Not installed"
fi

# Go
echo -n "Go: "
if command -v go &> /dev/null; then
    echo "✅ $(go version)"
else
    echo "❌ Not installed"
fi

# Node.js
echo -n "Node.js: "
if command -v node &> /dev/null; then
    echo "✅ $(node --version)"
else
    echo "❌ Not installed"
fi

# npm
echo -n "npm: "
if command -v npm &> /dev/null; then
    echo "✅ $(npm --version)"
else
    echo "❌ Not installed"
fi

# Python
echo -n "Python: "
if command -v python3 &> /dev/null; then
    echo "✅ $(python3 --version)"
else
    echo "❌ Not installed"
fi

# FFmpeg
echo -n "FFmpeg: "
if command -v ffmpeg &> /dev/null; then
    echo "✅ $(ffmpeg -version | head -n1)"
else
    echo "❌ Not installed"
fi

# PostgreSQL client
echo -n "PostgreSQL Client: "
if command -v psql &> /dev/null; then
    echo "✅ $(psql --version)"
else
    echo "❌ Not installed"
fi

# Redis CLI
echo -n "Redis CLI: "
if command -v redis-cli &> /dev/null; then
    echo "✅ $(redis-cli --version)"
else
    echo "❌ Not installed"
fi

# Make
echo -n "Make: "
if command -v make &> /dev/null; then
    echo "✅ $(make --version | head -n1)"
else
    echo "❌ Not installed"
fi

# Rust (optional)
echo -n "Rust (optional): "
if command -v rustc &> /dev/null; then
    echo "✅ $(rustc --version)"
else
    echo "⚠️  Not installed (optional for WHIP pusher development)"
fi

# GStreamer (optional)
echo -n "GStreamer (optional): "
if command -v gst-launch-1.0 &> /dev/null; then
    echo "✅ $(gst-launch-1.0 --version | head -n1)"
else
    echo "⚠️  Not installed (optional for WHIP pusher development)"
fi

echo ""
echo "=== Verification Complete ==="
```

Save this as `verify-prerequisites.sh` and run:
```bash
chmod +x verify-prerequisites.sh
./verify-prerequisites.sh
```

---

## Environment Setup

### 1. Create `.env` File

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```bash
# Milestone VMS Configuration
MILESTONE_SERVER_URL=http://192.168.1.9
MILESTONE_RTSP_URL=rtsp://192.168.1.9:554
MILESTONE_USERNAME=rta-integration
MILESTONE_PASSWORD=your-secure-password

# LiveKit Configuration
LIVEKIT_API_KEY=your-api-key
LIVEKIT_API_SECRET=your-api-secret
LIVEKIT_URL=ws://livekit:7880

# Database Configuration
POSTGRES_PASSWORD=your-postgres-password
POSTGRES_USER=cctv
POSTGRES_DB=cctv

# Valkey Configuration
VALKEY_PASSWORD=your-valkey-password

# MinIO Configuration
MINIO_ROOT_USER=admin
MINIO_ROOT_PASSWORD=your-minio-password
```

### 2. Generate SSL Certificates (for production)

```bash
# Self-signed certificate (development)
cd certs
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout server.key -out server.crt \
    -subj "/C=AE/ST=Dubai/L=Dubai/O=RTA/CN=cctv.rta.ae"
cd ..
```

---

## Downloadable Dependencies (Auto-installed)

These dependencies are automatically downloaded during setup:

### Go Modules
```bash
cd services/go-api
go mod download
```

### Node Modules
```bash
cd dashboard
npm install
# or
yarn install
```

### Python Packages
```bash
cd services/object-detection
pip install -r requirements.txt
```

### Docker Images
```bash
docker-compose pull
```

### GStreamer Plugins (whipsink)
Built during WHIP pusher Docker image build:
```bash
cd services/whip-pusher
docker build -t whip-pusher:latest .
```

---

## Troubleshooting

### Docker Permission Denied

**Issue**: `permission denied while trying to connect to the Docker daemon socket`

**Solution**:
```bash
sudo usermod -aG docker $USER
newgrp docker
# Or logout/login
```

### Go Module Download Fails

**Issue**: `go: module ... not found`

**Solution**:
```bash
export GOPROXY=https://proxy.golang.org,direct
go mod download
```

### Node Modules Installation Fails

**Issue**: `npm ERR! ... EACCES: permission denied`

**Solution**:
```bash
# Fix npm permissions
mkdir ~/.npm-global
npm config set prefix '~/.npm-global'
echo 'export PATH=~/.npm-global/bin:$PATH' >> ~/.bashrc
source ~/.bashrc
```

### FFmpeg Not Found

**Issue**: `ffmpeg: command not found`

**Solution**:
```bash
# Ensure FFmpeg is in PATH
which ffmpeg
# If not found, reinstall
sudo apt-get install -y ffmpeg
```

---

## Next Steps

After installing prerequisites:

1. ✅ **Clone Repository**: `git clone <repo-url>`
2. ✅ **Setup Environment**: Copy and edit `.env` file
3. ✅ **Build Images**: `docker-compose build`
4. ✅ **Start Services**: `docker-compose up -d`
5. ✅ **Verify Installation**: See [VERIFICATION_GUIDE.md](VERIFICATION_GUIDE.md)

---

## Reference Links

- **Docker**: https://docs.docker.com/
- **Go**: https://go.dev/doc/
- **Node.js**: https://nodejs.org/docs/
- **Python**: https://docs.python.org/3/
- **GStreamer**: https://gstreamer.freedesktop.org/documentation/
- **LiveKit**: https://docs.livekit.io/
- **Milestone SDK**: https://doc.developer.milestonesys.com/

---

**Document Version**: 1.0
**Last Updated**: 2025-10-26
**Maintained By**: RTA CCTV Development Team

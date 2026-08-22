#!/usr/bin/env bash
set -e

# Tarak Universal Installer for Linux & macOS
REPO="vikukumar/tarak"
INSTALL_DIR="/usr/local/bin"

# 1. Detect OS & Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) echo "Unsupported OS: $OS" && exit 1 ;;
esac

echo "⚡ Installing Tarak for $OS/$ARCH..."

# 2. Get latest release version
LATEST_TAG=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST_TAG" ]; then
  LATEST_TAG="v1.0.0"
fi
VERSION="${LATEST_TAG#v}"

TARBALL="tarak_${VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${TARBALL}"

echo "📦 Downloading ${DOWNLOAD_URL}..."
TMP_DIR=$(mktemp -d)
curl -fsSL "$DOWNLOAD_URL" -o "${TMP_DIR}/${TARBALL}"

echo "🚀 Extracting binaries..."
tar -xzf "${TMP_DIR}/${TARBALL}" -C "${TMP_DIR}"

if [ -w "$INSTALL_DIR" ]; then
  mv "${TMP_DIR}"/tarak* "$INSTALL_DIR/"
else
  sudo mv "${TMP_DIR}"/tarak* "$INSTALL_DIR/"
fi

rm -rf "$TMP_DIR"

echo "✅ Successfully installed Tarak ${LATEST_TAG} to ${INSTALL_DIR}!"
echo "   Run 'tarak version' or 'tarakctl version' to get started."

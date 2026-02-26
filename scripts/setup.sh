#!/usr/bin/env bash
set -euo pipefail

PLUGIN_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BIN_DIR="${PLUGIN_DIR}/bin"
VERSION_FILE="${PLUGIN_DIR}/scripts/version"
LOCAL_VERSION_FILE="${BIN_DIR}/.version"

REPO="llm-net/llm-api-plugin"

TOOLS=(gemini-cli ark-cli topview-cli)

# Read required version
REQUIRED_VERSION="$(cat "$VERSION_FILE" | tr -d '[:space:]')"

# Read local installed version
LOCAL_VERSION=""
if [[ -f "$LOCAL_VERSION_FILE" ]]; then
    LOCAL_VERSION="$(cat "$LOCAL_VERSION_FILE" | tr -d '[:space:]')"
fi

# Skip if already up to date
if [[ "$LOCAL_VERSION" == "$REQUIRED_VERSION" ]]; then
    echo "llm-api-plugin: binaries already at ${REQUIRED_VERSION}, skipping."
    exit 0
fi

echo "llm-api-plugin: upgrading from ${LOCAL_VERSION:-none} to ${REQUIRED_VERSION}..."

# Detect OS and ARCH
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

mkdir -p "$BIN_DIR"

# Download each tool
for TOOL in "${TOOLS[@]}"; do
    URL="https://github.com/${REPO}/releases/download/${REQUIRED_VERSION}/${TOOL}-${OS}-${ARCH}"
    echo "Downloading ${TOOL} (${OS}/${ARCH})..."

    if command -v curl &>/dev/null; then
        curl -fSL -o "${BIN_DIR}/${TOOL}" "$URL"
    elif command -v wget &>/dev/null; then
        wget -q -O "${BIN_DIR}/${TOOL}" "$URL"
    else
        echo "Error: neither curl nor wget found"
        exit 1
    fi

    chmod +x "${BIN_DIR}/${TOOL}"
    echo "  -> ${BIN_DIR}/${TOOL}"
done

# Record installed version
echo "$REQUIRED_VERSION" > "$LOCAL_VERSION_FILE"
echo "llm-api-plugin: done. Version ${REQUIRED_VERSION} installed."

#!/usr/bin/env bash
#
# Auto installer/updater for MAGIK
#
# Usage:
#   curl -fsSL https://github.com/ehsatn/MAGIK/raw/refs/heads/main/install.sh | bash
#   
#
# Pre-release support:
#   curl -fsSL https://github.com/ehsatn/MAGIK/raw/refs/heads/main/install.sh | bash -s -- --prerelease
#   or
#   PRERELEASE=1 bash install.sh
#   or
#   bash install.sh --prerelease

set -eu

DEBUG=${DEBUG:-0}
PRERELEASE=${PRERELEASE:-0}

REPO="ehsatn/MAGIK"
BIN_NAME="magik"

for arg in "$@"; do
    case "$arg" in
        --prerelease)
            PRERELEASE=1
            ;;
    esac
done

log() { if [ "$DEBUG" -eq 1 ]; then printf "\033[93m[DEBUG] %s\033[0m\n" "$@" >&2; fi; }
info() { printf "\033[36m[INFO]\033[0m %s\n" "$@"; }
success() { printf "\033[92m[SUCCESS]\033[0m %s\n" "$@"; }
error_exit() { printf "\033[91m[ERROR]\033[0m %s\n" "$@" >&2; exit 1; }

log "Detecting OS and architecture..."

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64 | amd64) ARCH="amd64" ;;
  i386 | i686) ARCH="386" ;;
  armv7l) ARCH="arm" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *) error_exit "Unsupported architecture: $ARCH" ;;
esac

log "Detected OS=$OS ARCH=$ARCH"

if [ -n "${PREFIX-}" ] && [ -d "/data/data/com.termux" ]; then
    DEST="$PREFIX/bin"
else
    DEST="$HOME/.local/bin"
fi

mkdir -p "$DEST"

BINARY_PATH="$DEST/$BIN_NAME"

log "Binary path: $BINARY_PATH"

API_URL="https://api.github.com/repos/$REPO/releases"

get_latest_release() {
    if [ "$PRERELEASE" -eq 1 ]; then
        curl -fsSL "$API_URL" | awk '
            /"tag_name":/ {
                gsub(/[",]/, "", $2)
                print $2
                exit
            }
        '
    else
        curl -fsSL "$API_URL/latest" | grep '"tag_name":' | cut -d '"' -f 4
    fi
}

get_installed_version() {
    if [ ! -f "$BINARY_PATH" ]; then
        return
    fi

    VERSION_OUTPUT=$("$BINARY_PATH" --version 2>/dev/null || true)

    VERSION=$(printf "%s\n" "$VERSION_OUTPUT" | grep -oP 'v?[0-9]+(?:\.[0-9A-Za-z_.-]*)*' | head -n1 || true)

    if [ -z "$VERSION" ]; then
        return
    fi

    case "$VERSION" in
        v*)
            printf "%s\n" "$VERSION"
            ;;
        *)
            printf "v%s\n" "$VERSION"
            ;;
    esac
}

TAG_RAW=$(get_latest_release || true)

if [ -z "$TAG_RAW" ]; then
    error_exit "Failed to fetch release info."
fi

TAG="${TAG_RAW#v}"

log "Latest tag: $TAG_RAW"

if [ "$PRERELEASE" -eq 1 ]; then
    info "Using pre-release channel"
else
    info "Using stable release channel"
fi

if [ -f "$BINARY_PATH" ]; then
    CURRENT_TAG=$(get_installed_version || true)

    if [ -z "$CURRENT_TAG" ]; then
        CURRENT_TAG="none"
    fi

    log "Installed version: $CURRENT_TAG"

    if [ "$CURRENT_TAG" = "$TAG_RAW" ]; then
        info "Already running latest version ($TAG_RAW)"
        echo "-----------------------------------------"
        exec "$BINARY_PATH" "$@"
    fi

    info "Updating $BIN_NAME from ${CURRENT_TAG#v} to ${TAG}"
else
    info "Installing $BIN_NAME version $TAG"
fi

EXT=""
if [ "$OS" = "windows" ]; then
    EXT=".exe"
fi

FILE_NAME="${BIN_NAME}-${OS}-${ARCH}${EXT}"
URL="https://github.com/$REPO/releases/download/$TAG_RAW/$FILE_NAME"

log "Download URL: $URL"

TMP_DOWNLOAD_PATH="${BINARY_PATH}.tmp"

info "Downloading binary..."

if ! curl -fSL --progress-bar "$URL" -o "$TMP_DOWNLOAD_PATH"; then
    rm -f "$TMP_DOWNLOAD_PATH"
    error_exit "Download failed."
fi

chmod +x "$TMP_DOWNLOAD_PATH"

mv "$TMP_DOWNLOAD_PATH" "$BINARY_PATH"

success "$BIN_NAME installed successfully at $BINARY_PATH"

echo ""
info "Make sure '$DEST' is in your PATH."
echo ""
info "Running $BIN_NAME..."
echo "-----------------------------------------"

exec "$BINARY_PATH" "$@"

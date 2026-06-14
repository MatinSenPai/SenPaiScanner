#!/bin/bash
set -e

# Make sure we're in the android directory
cd "$(dirname "$0")"

echo "Building Go mobile library for Android..."

# Ensure gomobile is installed
if ! command -v gomobile &> /dev/null; then
    echo "gomobile not found. Installing..."
    go install golang.org/x/mobile/cmd/gomobile@latest
    go install golang.org/x/mobile/cmd/gobind@latest
    export PATH="$PATH:$(go env GOPATH)/bin"
    gomobile init
fi

# Create libs directory if it doesn't exist
mkdir -p app/libs

# Build the AAR
echo "Running gomobile bind..."
gomobile bind -v -target=android/arm64,android/arm -androidapi 21 -javapkg=com.ehsatn.magik -o app/libs/magik.aar ../mobile

echo "Successfully built magik.aar!"

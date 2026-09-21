#!/usr/bin/env bash
set -e

# Turbo Rust multi-platform release builder
# Targets: macOS (arm64), Linux (amd64), Windows (amd64)

VERSION="${1:-v0.89}"
DIST_DIR="dist"

echo "=== Building Turbo Rust Release: ${VERSION} ==="
rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}"

# 1. macOS Apple Silicon (darwin/arm64)
echo "--> Building macOS (darwin/arm64)..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "${DIST_DIR}/tr" ./cmd/tr
tar -czf "${DIST_DIR}/tr-${VERSION}-darwin-arm64.tar.gz" -C "${DIST_DIR}" tr
rm "${DIST_DIR}/tr"

# 2. Linux 64-bit (linux/amd64)
echo "--> Building Linux (linux/amd64)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "${DIST_DIR}/tr" ./cmd/tr
tar -czf "${DIST_DIR}/tr-${VERSION}-linux-amd64.tar.gz" -C "${DIST_DIR}" tr
rm "${DIST_DIR}/tr"

# 3. Windows 64-bit (windows/amd64)
echo "--> Building Windows (windows/amd64)..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "${DIST_DIR}/tr.exe" ./cmd/tr
(cd "${DIST_DIR}" && zip -q "tr-${VERSION}-windows-amd64.zip" tr.exe && rm tr.exe)

# 4. Checksums
echo "--> Generating SHA256 checksums..."
if command -v sha256sum >/dev/null 2>&1; then
    (cd "${DIST_DIR}" && sha256sum tr-* > checksums.txt)
else
    (cd "${DIST_DIR}" && shasum -a 256 tr-* > checksums.txt)
fi

echo "=== Build Complete! ==="
ls -lh "${DIST_DIR}"

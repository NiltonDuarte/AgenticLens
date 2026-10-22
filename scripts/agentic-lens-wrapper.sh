#!/bin/sh
set -e

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64|amd64)
    ARCH="amd64"
    ;;
  arm64|aarch64)
    ARCH="arm64"
    ;;
  *)
    echo "Error: unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

DIR="$(cd "$(dirname "$0")" && pwd)"
TARGET_BIN="${DIR}/agentic-lens-${OS}-${ARCH}"

if [ ! -f "$TARGET_BIN" ]; then
  echo "Error: Binary for ${OS}/${ARCH} not found at ${TARGET_BIN}" >&2
  echo "Run 'make build' to compile binaries." >&2
  exit 1
fi

exec "$TARGET_BIN" "$@"

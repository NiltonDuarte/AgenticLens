#!/bin/sh
#
# agentic-lens-wrapper.sh - Multi-platform binary dispatcher and updater for AgenticLens
#
# Resolves the host OS and architecture at execution time, delegates to the appropriate
# platform-specific binary (e.g. agentic-lens-darwin-arm64, agentic-lens-linux-amd64),
# self-heals by downloading missing binaries from GitHub Releases on demand, and
# performs non-blocking background update checks.
#
# Zero emojis permitted.

set -e

DIR="$(cd "$(dirname "$0")" && pwd)"
REPO="${AGENTIC_LENS_REPO:-NiltonDuarte/AgenticLens}"
VERSION="${AGENTIC_LENS_VERSION:-latest}"

# Detect OS
OS_RAW="$(uname -s)"
case "$OS_RAW" in
  Darwin)
    OS="darwin"
    ;;
  Linux)
    OS="linux"
    ;;
  MINGW*|MSYS*|CYGWIN*|Windows_NT)
    OS="windows"
    ;;
  *)
    echo "Error: unsupported operating system: $OS_RAW" >&2
    exit 1
    ;;
esac

# Detect Architecture
ARCH_RAW="$(uname -m)"
case "$ARCH_RAW" in
  x86_64|amd64)
    ARCH="amd64"
    ;;
  arm64|aarch64)
    ARCH="arm64"
    ;;
  *)
    echo "Error: unsupported architecture: $ARCH_RAW" >&2
    exit 1
    ;;
esac

BIN_EXT=""
if [ "$OS" = "windows" ]; then
  BIN_EXT=".exe"
fi

BIN_NAME="agentic-lens-${OS}-${ARCH}${BIN_EXT}"
TARGET_BIN="${DIR}/${BIN_NAME}"
CHECK_FILE="${DIR}/.last_update_check"
LATEST_FILE="${DIR}/.latest_version"

download_artifact() {
  target_file="$1"
  remote_name="$2"
  ver="${3:-latest}"
  temp_file="${target_file}.tmp.$$"

  if [ "$ver" = "latest" ]; then
    download_url="https://github.com/${REPO}/releases/latest/download/${remote_name}"
  else
    download_url="https://github.com/${REPO}/releases/download/${ver}/${remote_name}"
  fi

  ok=0
  if command -v curl >/dev/null 2>&1; then
    if curl -fsSL -o "$temp_file" "$download_url"; then
      ok=1
    fi
  elif command -v wget >/dev/null 2>&1; then
    if wget -q -O "$temp_file" "$download_url"; then
      ok=1
    fi
  fi

  if [ "$ok" -eq 1 ] && [ -f "$temp_file" ]; then
    mv "$temp_file" "$target_file"
    chmod +x "$target_file"
    return 0
  else
    rm -f "$temp_file"
    return 1
  fi
}

resolve_latest_tag() {
  tag=""
  if command -v curl >/dev/null 2>&1; then
    eff_url="$(curl -fsSLI -o /dev/null -w "%{url_effective}" --max-time 5 "https://github.com/${REPO}/releases/latest" 2>/dev/null || true)"
    tag="$(basename "$eff_url" 2>/dev/null || true)"
  elif command -v wget >/dev/null 2>&1; then
    loc="$(wget --spider -S --timeout=5 "https://github.com/${REPO}/releases/latest" 2>&1 | grep -i "Location:" | tail -n 1 | awk '{print $2}' | tr -d '\r')"
    tag="$(basename "$loc" 2>/dev/null || true)"
  fi

  case "$tag" in
    ""|latest|releases)
      echo ""
      ;;
    *)
      echo "$tag"
      ;;
  esac
}

handle_update() {
  shift
  target_ver="latest"
  while [ $# -gt 0 ]; do
    case "$1" in
      --version)
        target_ver="$2"
        shift 2
        ;;
      *)
        shift
        ;;
    esac
  done

  echo "Checking for AgenticLens updates..."
  if [ "$target_ver" = "latest" ]; then
    latest_tag="$(resolve_latest_tag)"
    if [ -z "$latest_tag" ]; then
      echo "Error: Failed to resolve latest release from https://github.com/${REPO}/releases" >&2
      exit 1
    fi
  else
    latest_tag="$target_ver"
  fi

  current_ver="unknown"
  if [ -x "$TARGET_BIN" ]; then
    current_ver="$("$TARGET_BIN" version --short 2>/dev/null || echo "unknown")"
  fi

  echo "Target release: ${latest_tag} (current: ${current_ver})"

  updated_count=0
  for bin_path in "${DIR}"/agentic-lens-*; do
    if [ -f "$bin_path" ] && [ -x "$bin_path" ]; then
      fname="$(basename "$bin_path")"
      case "$fname" in
        *.tmp.*|*.bak)
          continue
          ;;
      esac
      echo "Updating ${fname} to ${latest_tag}..."
      if download_artifact "$bin_path" "$fname" "$latest_tag"; then
        updated_count=$((updated_count + 1))
      else
        echo "Warning: Failed to update ${fname}" >&2
      fi
    fi
  done

  if [ ! -f "$TARGET_BIN" ]; then
    echo "Downloading ${BIN_NAME} (${latest_tag})..."
    if download_artifact "$TARGET_BIN" "$BIN_NAME" "$latest_tag"; then
      updated_count=$((updated_count + 1))
    fi
  fi

  echo "$latest_tag" > "$LATEST_FILE" 2>/dev/null || true
  date +%s > "$CHECK_FILE" 2>/dev/null || true
  echo "Successfully updated ${updated_count} binary(ies) to ${latest_tag}."
  exit 0
}

trigger_background_update_check() {
  if [ "${AGENTIC_LENS_NO_UPDATE_CHECK:-0}" = "1" ]; then
    return 0
  fi

  last_check=0
  if [ -f "$CHECK_FILE" ]; then
    last_check="$(cat "$CHECK_FILE" 2>/dev/null || echo 0)"
  fi

  now="$(date +%s 2>/dev/null || echo 0)"
  elapsed=$(( now - last_check ))

  if [ "$elapsed" -ge 86400 ]; then
    echo "$now" > "$CHECK_FILE" 2>/dev/null || true
    (
      resolved_tag="$(resolve_latest_tag)"
      if [ -n "$resolved_tag" ]; then
        echo "$resolved_tag" > "$LATEST_FILE" 2>/dev/null || true
      fi
    ) >/dev/null 2>&1 &
  fi
}

notify_if_update_available() {
  if [ "${AGENTIC_LENS_NO_UPDATE_CHECK:-0}" = "1" ]; then
    return 0
  fi

  if [ -f "$LATEST_FILE" ] && [ -x "$TARGET_BIN" ]; then
    latest_tag="$(cat "$LATEST_FILE" 2>/dev/null || true)"
    if [ -n "$latest_tag" ]; then
      current_ver="$("$TARGET_BIN" version --short 2>/dev/null || true)"
      case "$current_ver" in
        *"-dirty"*|*"-dev"*|*"-g"*|"unknown"|"")
          ;;
        *)
          if [ "$current_ver" != "$latest_tag" ]; then
            echo "Notice: AgenticLens update available: ${latest_tag} (current: ${current_ver}). Run '$0 update' to upgrade." >&2
          fi
          ;;
      esac
    fi
  fi
}

if [ "${1:-}" = "update" ]; then
  handle_update "$@"
fi

if [ ! -f "$TARGET_BIN" ]; then
  echo "Notice: AgenticLens binary for ${OS}/${ARCH} not found locally at ${TARGET_BIN}." >&2
  echo "Attempting automatic download from GitHub Releases..." >&2
  if download_artifact "$TARGET_BIN" "$BIN_NAME" "$VERSION"; then
    echo "Successfully downloaded and verified ${BIN_NAME}." >&2
  else
    echo "Error: Binary for ${OS}/${ARCH} not found at ${TARGET_BIN} and automatic download failed." >&2
    if [ -f "${DIR}/../Makefile" ] || [ -f "${DIR}/Makefile" ]; then
      echo "Development environment detected. Run 'make build' to compile binaries." >&2
    else
      echo "Please check network connectivity or visit https://github.com/${REPO}/releases" >&2
    fi
    exit 1
  fi
fi

if [ ! -x "$TARGET_BIN" ]; then
  chmod +x "$TARGET_BIN" 2>/dev/null || true
fi

notify_if_update_available
trigger_background_update_check

exec "$TARGET_BIN" "$@"

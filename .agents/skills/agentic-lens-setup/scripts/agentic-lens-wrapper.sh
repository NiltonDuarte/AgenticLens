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
# A wrapper mistakenly installed as a native artifact would dispatch to itself.
# Fail before updates or delegation instead of recursively starting the wrapper.
case "$(basename "$0")" in
  "$BIN_NAME"|"$BIN_NAME".candidate.*|"$BIN_NAME".previous.*)
    echo "Error: Wrapper installed at native binary path ${TARGET_BIN}; install the native release executable." >&2
    exit 1
    ;;
esac
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

capture_protocol() {
  if [ ! -x "$1" ]; then return 1; fi
  protocol_value="$("$1" capture protocol 2>/dev/null)" || return 1
  case "$protocol_value" in ""|*[!0-9]*) return 1 ;; esac
  printf '%s' "$protocol_value"
}

finish_update_transaction() {
  update_result="$1"
  trap - EXIT HUP INT TERM
  if [ "$update_result" -ne 0 ] && [ -f "$BINARY_UPDATE_DIR/replaced" ]; then
    while IFS= read -r update_name; do
      if [ -f "$BINARY_UPDATE_DIR/previous/$update_name" ]; then
        if ! mv -f "$BINARY_UPDATE_DIR/previous/$update_name" "$DIR/$update_name"; then
          echo "Error: Restore the prior binary from $BINARY_UPDATE_DIR/previous/$update_name" >&2
          # Keep recoverable backups if a restoration itself fails.
          exit "$update_result"
        fi
      else
        rm -f "$DIR/$update_name"
      fi
    done < "$BINARY_UPDATE_DIR/replaced"
  fi
  rm -rf "$BINARY_UPDATE_DIR"
  exit "$update_result"
}

handle_update() {
  shift
  target_ver="latest"
  while [ $# -gt 0 ]; do
    case "$1" in
      --version)
        if [ $# -lt 2 ] || [ -z "$2" ]; then
          echo "Error: --version requires a release tag" >&2
          exit 1
        fi
        target_ver="$2"
        shift 2
        ;;
      *) echo "Error: Unknown update argument $1" >&2; exit 1 ;;
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
  required_protocol="$(capture_protocol "$TARGET_BIN" || true)"
  BINARY_UPDATE_DIR="$(mktemp -d "$DIR/.agenticlens-update.XXXXXX")"
  mkdir "$BINARY_UPDATE_DIR/previous"
  : > "$BINARY_UPDATE_DIR/ready"
  : > "$BINARY_UPDATE_DIR/replaced"
  trap 'finish_update_transaction "$?"' EXIT
  trap 'exit 1' HUP INT TERM

  # Stage every candidate before replacing any working executable. A running
  # capture-aware handler's protocol is a compatibility requirement, regardless
  # of where the repository keeps its installation manifests.
  for fname in agentic-lens-linux-amd64 agentic-lens-linux-arm64 agentic-lens-darwin-amd64 agentic-lens-darwin-arm64 agentic-lens-windows-amd64.exe; do
    if [ "$fname" != "$BIN_NAME" ] && [ ! -f "$DIR/$fname" ]; then continue; fi
    if [ "$fname" != "$BIN_NAME" ] && [ -n "$required_protocol" ]; then
      echo "Preserving ${fname}: verify capture compatibility and upgrade on its native host."
      continue
    fi
    candidate="$BINARY_UPDATE_DIR/$fname"
    echo "Staging ${fname} (${latest_tag})..."
    if ! download_artifact "$candidate" "$fname" "$latest_tag"; then
      echo "Error: Failed to stage ${fname}; existing binaries and hooks preserved." >&2
      exit 1
    fi
    if [ "$fname" = "$BIN_NAME" ] && [ -n "$required_protocol" ]; then
      observed_protocol="$(capture_protocol "$candidate" || true)"
      if [ "$observed_protocol" != "$required_protocol" ]; then
        echo "Error: Candidate capture protocol incompatible (required ${required_protocol}); existing binaries and hooks preserved." >&2
        exit 1
      fi
    fi
    printf '%s\n' "$fname" >> "$BINARY_UPDATE_DIR/ready"
  done

  updated_count=0
  while IFS= read -r fname; do
    if [ -f "$DIR/$fname" ]; then
      cp -p "$DIR/$fname" "$BINARY_UPDATE_DIR/previous/$fname"
    fi
    # Record before the atomic move so even an interrupted replacement recovers.
    printf '%s\n' "$fname" >> "$BINARY_UPDATE_DIR/replaced"
    mv -f "$BINARY_UPDATE_DIR/$fname" "$DIR/$fname"
    updated_count=$((updated_count + 1))
  done < "$BINARY_UPDATE_DIR/ready"
  echo "$latest_tag" > "$LATEST_FILE" 2>/dev/null || true
  date +%s > "$CHECK_FILE" 2>/dev/null || true
  echo "Successfully updated ${updated_count} binary(ies) to ${latest_tag}."
  echo "For lifecycle capture, inspect native capture status and preview capture install with the original source/store paths before applying owned template changes."
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

# Lifecycle handlers are local and bounded. Even a manual capture invocation must
# never trigger an updater, missing-binary download, or background network check.
if [ "${1:-}" = "capture" ]; then
  if [ ! -x "$TARGET_BIN" ]; then
    echo "Error: Native capture handler missing or not executable at ${TARGET_BIN}; reinstall the release binary explicitly." >&2
    exit 1
  fi
  exec "$TARGET_BIN" "$@"
fi

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

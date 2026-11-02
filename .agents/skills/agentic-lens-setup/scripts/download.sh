#!/usr/bin/env bash
#
# download.sh - Cross-platform artifact downloader for AgenticLens setup skill
#
# Detects host/container platform and architecture, installs the multi-platform
# wrapper script into the destination directory, downloads the precompiled binary
# from GitHub Releases. Git tracking/ignore policy belongs to the repository owner.
#
# Zero compilation required. Zero Go/Node toolchain dependency. Zero emojis.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

REPO="${AGENTIC_LENS_REPO:-NiltonDuarte/AgenticLens}"
DEST_DIR="${AGENTIC_LENS_BIN_DIR:-${SKILL_ROOT}/../agentic-lens-operator/bin}"
INSTALLED_VERSION=""
if [[ -f "$SKILL_ROOT/RELEASE_VERSION" ]]; then
    INSTALLED_VERSION="$(cat "$SKILL_ROOT/RELEASE_VERSION")"
fi
VERSION="${AGENTIC_LENS_VERSION:-$INSTALLED_VERSION}"
LOCAL_BIN="${AGENTIC_LENS_LOCAL_BINARY:-}"
INCLUDE_HOST=""
ALL_PLATFORMS=0
FORCE=0
DRY_RUN=0

usage() {
    cat <<EOF
Usage: download.sh [OPTIONS]

Options:
  --version <version>       Release version to download (default: installed RELEASE_VERSION; required if unavailable)
  --dest <dir>              Destination binary directory (default: sibling agentic-lens-operator/bin)
  --include-host <target>   Pre-download additional host binary (e.g. darwin-arm64, darwin-amd64, linux-amd64)
  --all-platforms           Download binaries for all supported platforms
  --local-bin <path>        Use local prebuilt binary instead of downloading (testing escape hatch)
  --force                   Overwrite existing binaries if present
  --dry-run                 Print resolved platform, architecture, and download URL without downloading
  --help                    Show this help message
EOF
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --version|--dest|--include-host|--local-bin)
            [[ $# -ge 2 && -n "$2" ]] || { echo "Error: $1 requires a value" >&2; exit 1; }
            ;;
    esac
    case "$1" in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --dest)
            DEST_DIR="$2"
            shift 2
            ;;
        --include-host)
            INCLUDE_HOST="$2"
            shift 2
            ;;
        --all-platforms)
            ALL_PLATFORMS=1
            shift
            ;;
        --local-bin)
            LOCAL_BIN="$2"
            shift 2
            ;;
        --force)
            FORCE=1
            shift
            ;;
        --dry-run)
            DRY_RUN=1
            shift
            ;;
        --help|-h)
            usage
            ;;
        *)
            echo "Error: Unknown argument $1" >&2
            exit 1
            ;;
    esac
done

# An untracked release marker may be absent in a new worktree. Never infer
# latest for an installed bundle whose original release is unknown.
if [[ -z "$VERSION" && ( -z "$LOCAL_BIN" || $DRY_RUN -eq 1 ) ]]; then
    echo "Error: Installed release identity is unavailable (missing/empty $SKILL_ROOT/RELEASE_VERSION). Select the intended release explicitly with --version <RELEASE_TAG> or AGENTIC_LENS_VERSION; no download performed." >&2
    exit 1
fi
case "$VERSION" in
    *[!A-Za-z0-9._-]*) echo "Error: Invalid release tag: $VERSION" >&2; exit 1 ;;
esac

# Detect Operating System
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
        echo "Error: Unsupported operating system: $OS_RAW" >&2
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
        echo "Error: Unsupported architecture: $ARCH_RAW" >&2
        exit 1
        ;;
esac

BIN_EXT=""
if [[ "$OS" == "windows" ]]; then
    BIN_EXT=".exe"
fi

if [[ "$OS" == windows && "$ARCH" != amd64 ]]; then
    echo "Error: Windows arm64 artifacts are unavailable." >&2
    exit 1
fi
PRIMARY_BIN="agentic-lens-${OS}-${ARCH}${BIN_EXT}"
BIN_LIST=("$PRIMARY_BIN")

if [[ $ALL_PLATFORMS -eq 1 ]]; then
    BIN_LIST=(
        "agentic-lens-linux-amd64"
        "agentic-lens-linux-arm64"
        "agentic-lens-darwin-amd64"
        "agentic-lens-darwin-arm64"
        "agentic-lens-windows-amd64.exe"
    )
elif [[ -n "$INCLUDE_HOST" ]]; then
    HOST_NORM="$(echo "$INCLUDE_HOST" | tr '/' '-' | tr '[:upper:]' '[:lower:]')"
    case "$HOST_NORM" in
        darwin-arm64|macos-arm64|mac-arm64|m1|m2|m3|apple-silicon)
            ADDITIONAL_BIN="agentic-lens-darwin-arm64"
            ;;
        darwin-amd64|macos-amd64|mac-amd64|mac-intel)
            ADDITIONAL_BIN="agentic-lens-darwin-amd64"
            ;;
        linux-amd64)
            ADDITIONAL_BIN="agentic-lens-linux-amd64"
            ;;
        linux-arm64)
            ADDITIONAL_BIN="agentic-lens-linux-arm64"
            ;;
        windows-amd64)
            ADDITIONAL_BIN="agentic-lens-windows-amd64.exe"
            ;;
        *)
            echo "Error: Unsupported host target: $INCLUDE_HOST" >&2
            exit 1
            ;;
    esac

    FOUND=0
    for b in "${BIN_LIST[@]}"; do
        if [[ "$b" == "$ADDITIONAL_BIN" ]]; then
            FOUND=1
            break
        fi
    done
    if [[ $FOUND -eq 0 ]]; then
        BIN_LIST+=("$ADDITIONAL_BIN")
    fi
fi

TARGET_WRAPPER="${DEST_DIR}/agentic-lens"

if [[ $DRY_RUN -eq 1 ]]; then
    echo "Platform: ${OS}"
    echo "Architecture: ${ARCH}"
    echo "Version: ${VERSION}"
    echo "Destination: ${DEST_DIR}"
    echo "Target Wrapper: ${TARGET_WRAPPER}"
    echo "Binaries to install:"
    for b in "${BIN_LIST[@]}"; do
        if [[ "$VERSION" == "latest" ]]; then
            u="https://github.com/${REPO}/releases/latest/download/${b}"
        else
            u="https://github.com/${REPO}/releases/download/${VERSION}/${b}"
        fi
        echo "  - ${b} -> ${u}"
    done
    exit 0
fi

# Ensure destination directory exists
mkdir -p "$DEST_DIR"

# Only use the bundled wrapper. Missing setup files require explicit bootstrap,
# never a moving-main dependency download mixed with a selected binary release.
install_wrapper() {
    local wrapper_src="${SCRIPT_DIR}/agentic-lens-wrapper.sh"
    if [[ ! -f "$wrapper_src" ]]; then
        echo "Error: Setup wrapper is missing: $wrapper_src. Restore the setup skill with scripts/install.sh before repair." >&2
        return 1
    fi
    local wrapper_temp
    wrapper_temp="$(mktemp "${TARGET_WRAPPER}.tmp.XXXXXX")"
    if ! cp "$wrapper_src" "$wrapper_temp" || ! chmod +x "$wrapper_temp" || ! mv "$wrapper_temp" "$TARGET_WRAPPER"; then
        rm -f "$wrapper_temp"
        return 1
    fi
}

if [[ ! -f "$TARGET_WRAPPER" || $FORCE -eq 1 ]]; then
    install_wrapper
fi

# Explicit repair restores execute permission without replacing stored bytes.
retain_binary() {
    local target="$1"
    if [[ ! -x "$target" ]]; then
        if ! chmod +x "$target" || [[ ! -x "$target" ]]; then
            echo "Error: Could not restore executable permission: $target. Repair filesystem permissions before retrying." >&2
            return 1
        fi
    fi
    echo "Binary already exists at ${target} (retained; use --force to replace)."
}

# If a local binary override is specified, install it for current platform
if [[ -n "$LOCAL_BIN" ]]; then
    if [[ ! -f "$LOCAL_BIN" ]]; then
        echo "Error: Specified local binary does not exist: $LOCAL_BIN" >&2
        exit 1
    fi
    LOCAL_TARGET="${DEST_DIR}/${PRIMARY_BIN}"
    if [[ -f "$LOCAL_TARGET" && $FORCE -eq 0 ]]; then
        retain_binary "$LOCAL_TARGET"
        exit 0
    fi
    echo "Installing local prebuilt binary from ${LOCAL_BIN} to ${LOCAL_TARGET}..."
    LOCAL_TEMP="$(mktemp "${LOCAL_TARGET}.tmp.XXXXXX")"
    if ! cp "$LOCAL_BIN" "$LOCAL_TEMP" || ! chmod +x "$LOCAL_TEMP" || ! mv "$LOCAL_TEMP" "$LOCAL_TARGET"; then
        rm -f "$LOCAL_TEMP"
        exit 1
    fi
    echo "Successfully installed AgenticLens wrapper at ${TARGET_WRAPPER} and local binary at ${LOCAL_TARGET}"
    exit 0
fi

# Download binaries
download_binary() {
    local b_name="$1"
    local b_dest="${DEST_DIR}/${b_name}"

    if [[ -f "$b_dest" && $FORCE -eq 0 ]]; then
        retain_binary "$b_dest"
        return $?
    fi

    local d_url
    if [[ "$VERSION" == "latest" ]]; then
        d_url="https://github.com/${REPO}/releases/latest/download/${b_name}"
    else
        d_url="https://github.com/${REPO}/releases/download/${VERSION}/${b_name}"
    fi

    echo "Downloading ${b_name} (${VERSION})..."
    local temp_bin
    temp_bin="$(mktemp "${b_dest}.tmp.XXXXXX")"
    local success=0

    if command -v curl >/dev/null 2>&1; then
        if curl -fsSL -o "$temp_bin" "$d_url"; then
            success=1
        fi
    fi

    if [[ $success -eq 0 ]] && command -v wget >/dev/null 2>&1; then
        if wget -q -O "$temp_bin" "$d_url"; then
            success=1
        fi
    fi

    if [[ $success -eq 0 || ! -f "$temp_bin" ]]; then
        rm -f "$temp_bin"
        echo "Error: Failed to download ${b_name} from ${d_url}" >&2
        return 1
    fi

    if ! chmod +x "$temp_bin" || ! mv "$temp_bin" "$b_dest"; then
        rm -f "$temp_bin"
        return 1
    fi
    echo "Installed ${b_name} at ${b_dest}"
    return 0
}

FAILURES=0
for b in "${BIN_LIST[@]}"; do
    if ! download_binary "$b"; then
        FAILURES=$((FAILURES + 1))
    fi
done

if [[ $FAILURES -gt 0 ]]; then
    echo "Error: Failed to download ${FAILURES} binary artifact(s). Retry: bash \"${SCRIPT_DIR}/download.sh\" --dest \"${DEST_DIR}\" --version \"${VERSION}\"" >&2
    exit 1
fi

echo "Successfully installed AgenticLens wrapper at ${TARGET_WRAPPER}"

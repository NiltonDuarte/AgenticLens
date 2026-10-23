#!/usr/bin/env bash
#
# download-agentic-lens.sh - Cross-platform artifact downloader for AgenticLens
#
# Detects host/container platform and architecture, installs the multi-platform
# wrapper script into the destination directory, downloads the precompiled binary
# from GitHub Releases, and writes a protective .gitignore.
#
# Zero compilation required. Zero Go/Node toolchain dependency. Zero emojis.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

REPO="${AGENTIC_LENS_REPO:-NiltonDuarte/AgenticLens}"
DEST_DIR="${AGENTIC_LENS_BIN_DIR:-.opencode/skills/agentic-lens-setup/bin}"
VERSION="${AGENTIC_LENS_VERSION:-latest}"
LOCAL_BIN="${AGENTIC_LENS_LOCAL_BINARY:-}"
INCLUDE_HOST=""
ALL_PLATFORMS=0
FORCE=0
DRY_RUN=0

usage() {
    cat <<EOF
Usage: download-agentic-lens.sh [OPTIONS]

Options:
  --version <version>       Release version to download (default: latest)
  --dest <dir>              Destination binary directory (default: .opencode/skills/agentic-lens-setup/bin)
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
    # Normalize include-host input (e.g. darwin/arm64 -> darwin-arm64)
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
        agentic-lens-*)
            ADDITIONAL_BIN="$HOST_NORM"
            ;;
        *)
            ADDITIONAL_BIN="agentic-lens-${HOST_NORM}"
            ;;
    esac

    # Add if not already present
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

# Write protective .gitignore if not present
GITIGNORE="${DEST_DIR}/.gitignore"
if [[ ! -f "$GITIGNORE" ]]; then
    cat <<'GITIGNORE_EOF' > "$GITIGNORE"
# Automatically generated by AgenticLens downloader
# Prevent accidental git tracking of precompiled binaries
*
!.gitignore
GITIGNORE_EOF
fi

# Install wrapper script into destination
install_wrapper() {
    local wrapper_src=""
    if [[ -f "${SCRIPT_DIR}/agentic-lens-wrapper.sh" ]]; then
        wrapper_src="${SCRIPT_DIR}/agentic-lens-wrapper.sh"
    elif [[ -f "${REPO_ROOT}/scripts/agentic-lens-wrapper.sh" ]]; then
        wrapper_src="${REPO_ROOT}/scripts/agentic-lens-wrapper.sh"
    elif [[ -f "${SCRIPT_DIR}/../scripts/agentic-lens-wrapper.sh" ]]; then
        wrapper_src="${SCRIPT_DIR}/../scripts/agentic-lens-wrapper.sh"
    fi

    if [[ -n "$wrapper_src" ]]; then
        cp "$wrapper_src" "$TARGET_WRAPPER"
        chmod +x "$TARGET_WRAPPER"
        return 0
    fi

    # Fallback to downloading wrapper from repository
    local wrapper_url="https://raw.githubusercontent.com/${REPO}/main/scripts/agentic-lens-wrapper.sh"
    echo "Fetching wrapper script from ${wrapper_url}..."
    if command -v curl >/dev/null 2>&1; then
        if curl -fsSL -o "$TARGET_WRAPPER" "$wrapper_url"; then
            chmod +x "$TARGET_WRAPPER"
            return 0
        fi
    elif command -v wget >/dev/null 2>&1; then
        if wget -q -O "$TARGET_WRAPPER" "$wrapper_url"; then
            chmod +x "$TARGET_WRAPPER"
            return 0
        fi
    fi

    echo "Error: Failed to obtain agentic-lens-wrapper.sh" >&2
    exit 1
}

if [[ ! -f "$TARGET_WRAPPER" || $FORCE -eq 1 ]]; then
    install_wrapper
fi

# If a local binary override is specified, install it for current platform
if [[ -n "$LOCAL_BIN" ]]; then
    if [[ ! -f "$LOCAL_BIN" ]]; then
        echo "Error: Specified local binary does not exist: $LOCAL_BIN" >&2
        exit 1
    fi
    LOCAL_TARGET="${DEST_DIR}/${PRIMARY_BIN}"
    echo "Installing local prebuilt binary from ${LOCAL_BIN} to ${LOCAL_TARGET}..."
    cp "$LOCAL_BIN" "$LOCAL_TARGET"
    chmod +x "$LOCAL_TARGET"
    echo "Successfully installed AgenticLens wrapper at ${TARGET_WRAPPER} and local binary at ${LOCAL_TARGET}"
    exit 0
fi

# Download binaries
download_binary() {
    local b_name="$1"
    local b_dest="${DEST_DIR}/${b_name}"

    if [[ -f "$b_dest" && $FORCE -eq 0 ]]; then
        echo "Binary ${b_name} already exists at ${b_dest} (skipping)."
        return 0
    fi

    local d_url
    if [[ "$VERSION" == "latest" ]]; then
        d_url="https://github.com/${REPO}/releases/latest/download/${b_name}"
    else
        d_url="https://github.com/${REPO}/releases/download/${VERSION}/${b_name}"
    fi

    echo "Downloading ${b_name} (${VERSION})..."
    local temp_bin="${b_dest}.tmp.$$"
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

    mv "$temp_bin" "$b_dest"
    chmod +x "$b_dest"
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
    echo "Error: Failed to download ${FAILURES} binary artifact(s)." >&2
    exit 1
fi

echo "Successfully installed AgenticLens wrapper at ${TARGET_WRAPPER}"

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
CAPTURE_SOURCE=""
CAPTURE_STORE=""
CAPTURE_REPO=""
CAPTURE_STATE_DIR=""

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
  --capture-source <source> Configure lifecycle capture for codex or opencode after binary installation
  --capture-store <path>    Explicit absolute Codex home or OpenCode database path (required with --capture-source)
  --capture-repo <dir>      Repository for additive hook installation (default: current directory)
  --capture-state-dir <dir> Local ignored capture state (default: <capture-repo>/.agenticlens-capture)
  --dry-run                 Preview downloads and capture setup without writing
  --help                    Show this help message
EOF
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --version|--dest|--include-host|--local-bin)
            [[ $# -ge 2 && -n "$2" && "$2" != --* ]] || { echo "Error: $1 requires a value" >&2; exit 1; }
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
        --capture-source|--capture-store|--capture-repo|--capture-state-dir)
            if [[ $# -lt 2 || -z "$2" || "$2" == --* ]]; then
                echo "Error: $1 requires a value" >&2
                exit 1
            fi
            case "$1" in
                --capture-source) CAPTURE_SOURCE="$2" ;;
                --capture-store) CAPTURE_STORE="$2" ;;
                --capture-repo) CAPTURE_REPO="$2" ;;
                --capture-state-dir) CAPTURE_STATE_DIR="$2" ;;
            esac
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

# Hook ownership is explicit. Never infer the harness or source store from a ticket,
# branch, directory name, or the existence of a different harness's records.
if [[ -n "$CAPTURE_SOURCE" ]]; then
    case "$CAPTURE_SOURCE" in
        codex|opencode) ;;
        *) echo "Error: --capture-source must be codex or opencode" >&2; exit 1 ;;
    esac
    case "$CAPTURE_STORE" in
        /*|[A-Za-z]:\\*|[A-Za-z]:/*) ;;
        *) echo "Error: --capture-store requires the explicit absolute source path" >&2; exit 1 ;;
    esac
    CAPTURE_REPO="$(cd "${CAPTURE_REPO:-.}" && pwd -P)"
elif [[ -n "$CAPTURE_STORE" || -n "$CAPTURE_REPO" || -n "$CAPTURE_STATE_DIR" ]]; then
    echo "Error: capture options require --capture-source" >&2
    exit 1
fi

# An untracked release marker may be absent in a new worktree. Never infer
# latest for an installed bundle whose original release is unknown.
if [[ -z "$VERSION" && ( -z "$LOCAL_BIN" || $DRY_RUN -eq 1 ) ]]; then
    echo "Error: Installed release identity is unavailable (missing/empty $SKILL_ROOT/RELEASE_VERSION). Select the intended release explicitly with --version <RELEASE_TAG> or AGENTIC_LENS_VERSION; no download performed." >&2
    exit 1
fi
case "$VERSION" in
    *[!A-Za-z0-9._+-]*|.|..) echo "Error: Invalid release tag: $VERSION" >&2; exit 1 ;;
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
        *)
            echo "Error: Unsupported host target: $INCLUDE_HOST" >&2
            exit 1
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

# Capture-aware binaries retain their handler protocol across upgrades, even when
# their installation manifests live outside this repository or use custom paths.
read_capture_protocol() {
    local value
    if [[ ! -x "$1" ]] || ! value="$("$1" capture protocol 2>/dev/null)"; then
        return 1
    fi
    case "$value" in
        ""|*[!0-9]*) return 1 ;;
    esac
    printf '%s' "$value"
}

REPLACED_TARGETS=()
PREVIOUS_BINARIES=()
STAGED_BINARIES=()
finish_binary_transaction() {
    local result=$? i backup
    trap - EXIT HUP INT TERM
    if [[ $result -ne 0 ]]; then
        for ((i=${#REPLACED_TARGETS[@]}-1; i>=0; i--)); do
            backup="${PREVIOUS_BINARIES[$i]}"
            if [[ -n "$backup" ]]; then
                if ! mv -f "$backup" "${REPLACED_TARGETS[$i]}"; then
                    echo "Error: Restore native binary manually from $backup" >&2
                    PREVIOUS_BINARIES[$i]=""
                fi
            else
                rm -f "${REPLACED_TARGETS[$i]}"
            fi
        done
    fi
    for backup in "${PREVIOUS_BINARIES[@]-}" "${STAGED_BINARIES[@]-}"; do
        if [[ -n "$backup" ]]; then rm -f "$backup"; fi
    done
    exit "$result"
}
trap finish_binary_transaction EXIT
trap 'exit 1' HUP INT TERM

HOST_CAPTURE_PROTOCOL=""
if [[ $DRY_RUN -eq 0 && -x "${DEST_DIR}/${PRIMARY_BIN}" ]]; then
    HOST_CAPTURE_PROTOCOL="$(read_capture_protocol "${DEST_DIR}/${PRIMARY_BIN}" || true)"
fi

activate_candidate() {
    local candidate="$1" target="$2" name="$3" required="" observed=""
    if [[ "$name" != "$PRIMARY_BIN" ]]; then
        if [[ -f "$target" && ( -n "$HOST_CAPTURE_PROTOCOL" || -n "$CAPTURE_SOURCE" ) ]]; then
            echo "Preserving existing ${name}: verify capture compatibility and upgrade on its native host."
            return 0
        fi
    else
        required="$HOST_CAPTURE_PROTOCOL"
        if [[ -n "$CAPTURE_SOURCE" ]]; then required="1"; fi
        if [[ -n "$required" ]]; then
            observed="$(read_capture_protocol "$candidate" || true)"
            if [[ "$observed" != "$required" ]]; then
                echo "Error: Candidate capture protocol incompatible (required ${required}); existing native handler and hooks preserved." >&2
                return 1
            fi
        fi
    fi
    replace_owned_file "$candidate" "$target"
}

replace_owned_file() {
    local candidate="$1" target="$2" backup=""
    if [[ -f "$target" ]]; then
        backup="$(mktemp "${target}.previous.XXXXXX")" || return 1
        if ! cp -p "$target" "$backup"; then rm -f "$backup"; return 1; fi
    fi
    REPLACED_TARGETS+=("$target")
    PREVIOUS_BINARIES+=("$backup")
    mv -f "$candidate" "$target"
}

configure_capture() {
    if [[ -z "$CAPTURE_SOURCE" ]]; then
        echo "Binary installation only; lifecycle capture is not configured by this invocation."
        return 0
    fi
    local native="${DEST_DIR}/${PRIMARY_BIN}"
    local args=(capture install --source "$CAPTURE_SOURCE" --repo "$CAPTURE_REPO"
        --store "$CAPTURE_STORE")
    if [[ -n "$CAPTURE_STATE_DIR" ]]; then args+=(--state-dir "$CAPTURE_STATE_DIR"); fi
    if [[ $DRY_RUN -eq 1 ]]; then
        args+=(--dry-run)
        # An explicitly supplied local release can preview hooks before installation.
        if [[ -n "$LOCAL_BIN" && -x "$LOCAL_BIN" ]]; then
            native="$LOCAL_BIN"
        fi
        printf 'Capture setup command:'
        printf ' %q' "$native" "${args[@]}"
        printf '\n'
        if [[ ! -x "$native" ]]; then
            echo "Capture file preview unavailable until the native release binary is installed; no hooks changed."
            return 0
        fi
    elif [[ ! -x "$native" ]]; then
        echo "Error: Native capture handler unavailable at $native; no hook setup attempted." >&2
        return 1
    fi
    # Use the native handler directly: hook setup must never self-update via a wrapper.
    if ! "$native" "${args[@]}"; then
        echo "Error: Capture setup failed; inspect native capture diagnostics. Binary installation remains available." >&2
        return 1
    fi
    if [[ $DRY_RUN -eq 0 ]]; then
        echo "Lifecycle configuration installed; runtime trust, agent binding and source reconciliation still require verification."
    fi
}

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
    configure_capture
    exit 0
fi

# Ensure destination directory exists
mkdir -p "$DEST_DIR"

# Only use the bundled wrapper from the selected setup release. Recovery must
# not mix a moving-main wrapper with an installed release's native binary.
install_wrapper() {
    local wrapper_src="${SCRIPT_DIR}/agentic-lens-wrapper.sh"
    if [[ ! -f "$wrapper_src" ]]; then
        echo "Error: Setup wrapper is missing: $wrapper_src. Restore the setup skill with scripts/install.sh before repair." >&2
        return 1
    fi
    local candidate
    candidate="$(mktemp "${TARGET_WRAPPER}.candidate.XXXXXX")"
    STAGED_BINARIES+=("$candidate")
    cp "$wrapper_src" "$candidate"
    chmod +x "$candidate"
    replace_owned_file "$candidate" "$TARGET_WRAPPER"
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
        configure_capture
        exit 0
    fi
    echo "Installing local prebuilt binary from ${LOCAL_BIN} to ${LOCAL_TARGET}..."
    LOCAL_CANDIDATE="$(mktemp "${LOCAL_TARGET}.candidate.XXXXXX")"
    STAGED_BINARIES+=("$LOCAL_CANDIDATE")
    cp "$LOCAL_BIN" "$LOCAL_CANDIDATE"
    chmod +x "$LOCAL_CANDIDATE"
    activate_candidate "$LOCAL_CANDIDATE" "$LOCAL_TARGET" "$PRIMARY_BIN"
    echo "Successfully installed AgenticLens wrapper at ${TARGET_WRAPPER} and local binary at ${LOCAL_TARGET}"
    configure_capture
    exit 0
fi

# Download binaries
download_binary() {
    local b_name="$1"
    local b_dest="${DEST_DIR}/${b_name}"
    if [[ "$b_name" != "$PRIMARY_BIN" && -f "$b_dest" && ( -n "$HOST_CAPTURE_PROTOCOL" || -n "$CAPTURE_SOURCE" ) ]]; then
        echo "Preserving existing ${b_name}: verify capture compatibility and upgrade on its native host."
        return 0
    fi

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
    temp_bin="$(mktemp "${b_dest}.candidate.XXXXXX")" || return 1
    STAGED_BINARIES+=("$temp_bin")
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

    if ! chmod +x "$temp_bin"; then return 1; fi
    if ! activate_candidate "$temp_bin" "$b_dest" "$b_name"; then
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
configure_capture

#!/bin/sh
# Offline launcher. Installation/download is an explicit setup operation.
set -eu
SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
BIN_DIR="${AGENTIC_LENS_BIN_DIR:-$SCRIPT_DIR/../bin}"
case "$(uname -s)" in
    Linux) os=linux ;;
    Darwin) os=darwin ;;
    MINGW*|MSYS*|CYGWIN*|Windows_NT) os=windows ;;
    *) echo 'Error: Unsupported operating system.' >&2; exit 1 ;;
esac
case "$(uname -m)" in
    x86_64|amd64) arch=amd64 ;;
    aarch64|arm64) arch=arm64 ;;
    *) echo 'Error: Unsupported architecture.' >&2; exit 1 ;;
esac
ext=""
if [ "$os" = windows ]; then
    [ "$arch" = amd64 ] || { echo 'Error: Windows arm64 artifacts are unavailable.' >&2; exit 1; }
    ext=.exe
fi
binary="$BIN_DIR/agentic-lens-$os-$arch$ext"
if [ ! -x "$binary" ]; then
    echo "Error: AgenticLens platform executable is missing or not executable: $binary" >&2
    repair="$SCRIPT_DIR/../../agentic-lens-setup/scripts/download.sh"
    if [ -f "$repair" ]; then
        echo "Install/repair explicitly (restores execute permission or downloads a missing release binary; leaves configuration and Git policy unchanged):" >&2
        if [ -s "$SCRIPT_DIR/../../agentic-lens-setup/RELEASE_VERSION" ] || [ -n "${AGENTIC_LENS_VERSION:-}" ]; then
            printf '  bash "%s" --dest "%s"\n' "$repair" "$BIN_DIR" >&2
        else
            echo 'Release identity is also unavailable. Select the intended release tag explicitly; repair will not assume latest:' >&2
            printf '  bash "%s" --dest "%s" --version "<RELEASE_TAG>"\n' "$repair" "$BIN_DIR" >&2
        fi
    else
        echo 'Setup repair scripts are also missing. Follow the bootstrap instructions in agentic-lens-setup/SKILL.md or:' >&2
        echo '  https://github.com/NiltonDuarte/AgenticLens/blob/main/.agents/skills/agentic-lens-setup/SKILL.md' >&2
    fi
    exit 1
fi
exec "$binary" "$@"

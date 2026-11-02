#!/usr/bin/env bash
# Install the complete setup/operator skill bundle without changing Git policy or config.
set -euo pipefail

REPO="${AGENTIC_LENS_REPO:-NiltonDuarte/AgenticLens}"
VERSION="${AGENTIC_LENS_VERSION:-latest}"
DEST="."
SOURCE_DIR=""
usage() {
    cat <<'USAGE'
Usage: install.sh [--version <release-tag>] [--dest <repository>] [--source-dir <source-root>]
Installs setup/operator manifests, references and scripts only. Download binaries separately
with setup/scripts/download.sh. Existing differing files are never overwritten.
--source-dir copies the same bundle from a local checkout (for offline installation/testing).
No Git commands, ignore rules, project configuration or agent policies are changed.
USAGE
}
while [[ $# -gt 0 ]]; do
    case "$1" in
        --version|--dest|--source-dir)
            [[ $# -ge 2 && -n "$2" ]] || { echo "Error: $1 requires a value" >&2; exit 1; }
            case "$1" in
                --version) VERSION="$2" ;;
                --dest) DEST="$2" ;;
                --source-dir) SOURCE_DIR="$2" ;;
            esac
            shift 2 ;;
        --help|-h) usage; exit 0 ;;
        *) echo "Error: Unknown argument: $1" >&2; exit 1 ;;
    esac
done

stage="$(mktemp -d)"
trap 'rm -rf "$stage"' EXIT
fetch() {
    local url="$1" target="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL --connect-timeout 10 --max-time 120 -o "$target" "$url"
    elif command -v wget >/dev/null 2>&1; then
        wget -q --timeout=120 -O "$target" "$url"
    else
        echo "Error: Install curl or wget to retrieve AgenticLens skills." >&2
        return 1
    fi
}

# Resolve latest once, so every manifest/script and later binary repair use one release.
if [[ -z "$SOURCE_DIR" && "$VERSION" == latest ]]; then
    command -v curl >/dev/null 2>&1 || { echo "Error: Resolving latest requires curl; alternatively pass --version <release-tag>." >&2; exit 1; }
    latest_url="$(curl -fsSLI --connect-timeout 10 --max-time 30 -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest")"
    VERSION="${latest_url##*/}"
    [[ -n "$VERSION" && "$VERSION" != latest && "$VERSION" != releases ]] || { echo "Error: Could not resolve latest release tag." >&2; exit 1; }
fi
case "$VERSION" in
    *[!A-Za-z0-9._-]*|"") echo "Error: Invalid release tag: $VERSION" >&2; exit 1 ;;
esac

FILES=(
    agentic-lens-setup/SKILL.md
    agentic-lens-setup/scripts/install.sh
    agentic-lens-setup/scripts/download.sh
    agentic-lens-setup/scripts/agentic-lens-wrapper.sh
    agentic-lens-operator/SKILL.md
    agentic-lens-operator/references/cli.md
    agentic-lens-operator/references/configuration.md
    agentic-lens-operator/references/dashboard.md
    agentic-lens-operator/scripts/agentic-lens.sh
)
for path in "${FILES[@]}"; do
    mkdir -p "$stage/$(dirname "$path")"
    if [[ -n "$SOURCE_DIR" ]]; then
        cp "$SOURCE_DIR/.agents/skills/$path" "$stage/$path"
    else
        fetch "https://raw.githubusercontent.com/${REPO}/${VERSION}/.agents/skills/${path}" "$stage/$path"
    fi
done
# A local source fixture can opt into a release tag; latest stays an explicit unpinned choice.
printf '%s\n' "$VERSION" > "$stage/agentic-lens-setup/RELEASE_VERSION"
FILES+=(agentic-lens-setup/RELEASE_VERSION)

# Preflight the entire bundle before creating or replacing installed files.
for path in "${FILES[@]}"; do
    target="$DEST/.agents/skills/$path"
    if [[ -e "$target" || -L "$target" ]]; then
        if [[ -L "$target" || ! -f "$target" ]] || ! cmp -s "$stage/$path" "$target"; then
            echo "Error: Existing file differs: $target. Review its diff and reconcile explicitly; no skill files installed." >&2
            exit 1
        fi
    fi
done
for path in "${FILES[@]}"; do
    target="$DEST/.agents/skills/$path"
    if [[ ! -e "$target" ]]; then
        mkdir -p "$(dirname "$target")"
        cp "$stage/$path" "$target"
        case "$path" in *.sh) chmod +x "$target" ;; esac
    fi
done
echo "Installed AgenticLens setup and operator skills ($VERSION) in $DEST/.agents/skills."
echo 'Next, explicitly install/repair platform binaries with: bash .agents/skills/agentic-lens-setup/scripts/download.sh'
echo 'Git tracking and ignore choices remain with the repository owner.'

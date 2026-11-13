#!/usr/bin/env bash
# Install reviewed skill and optional OpenSpec resources, preserving existing files and Git policy.
set -euo pipefail

REPO="${AGENTIC_LENS_REPO:-NiltonDuarte/AgenticLens}"
VERSION="${AGENTIC_LENS_VERSION:-}"
DEST="."
SOURCE_DIR=""
WITH_OPENSPEC=0
usage() {
    cat <<'USAGE'
Usage: install.sh [--version <release-tag>] [--dest <repository>] [--source-dir <source-root>] [--with-openspec]
Installs setup/operator manifests, references and scripts only. Download binaries separately
with setup/scripts/download.sh. Existing differing files are never overwritten.
--source-dir copies the same bundle from a local checkout (for offline installation/testing).
--with-openspec additionally installs the pinned development tool/schema bundle.
Review that optional project configuration first; npm installation and assistant init remain separate.
No Git commands, ignore rules or existing differing files are changed.
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
        --with-openspec) WITH_OPENSPEC=1; shift ;;
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

[[ "$VERSION" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?(\+[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?$ && ${#VERSION} -le 128 ]] || { echo 'Error: Select an exact --version release; latest is not a project pin.' >&2; exit 1; }
case "$VERSION" in
    *[!A-Za-z0-9._+-]*|.|..|"") echo "Error: Invalid release tag: $VERSION" >&2; exit 1 ;;
esac

FILES=(
    agentic-lens-setup/SKILL.md
    agentic-lens-setup/scripts/install.sh
    agentic-lens-setup/scripts/download.sh
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

# Optional project-level OpenSpec resources use the same exact release and preflight.
PROJECT_FILES=()
if ((WITH_OPENSPEC)); then
    PROJECT_FILES=(
        tools/openspec/package.json tools/openspec/package-lock.json
        scripts/openspec.sh scripts/openspec-change.mjs docs/OPENSPEC_WORKFLOW.md
        openspec/config.yaml
        openspec/schemas/agenticlens/schema.yaml
        openspec/schemas/agenticlens/templates/proposal.md
        openspec/schemas/agenticlens/templates/spec.md
        openspec/schemas/agenticlens/templates/design.md
        openspec/schemas/agenticlens/templates/tasks.md
        openspec/schemas/agenticlens/templates/delivery.md
        openspec/schemas/agenticlens-planning/schema.yaml
        openspec/schemas/agenticlens-planning/templates/proposal.md
        openspec/schemas/agenticlens-planning/templates/tasks.md
        openspec/schemas/agenticlens-planning/templates/delivery.md
    )
    for path in "${PROJECT_FILES[@]}"; do
        mkdir -p "$stage/project/$(dirname "$path")"
        if [[ -n "$SOURCE_DIR" ]]; then
            cp "$SOURCE_DIR/$path" "$stage/project/$path"
        else
            fetch "https://raw.githubusercontent.com/${REPO}/${VERSION}/${path}" "$stage/project/$path"
        fi
    done
fi

# An existing ancestor symlink must not redirect writes outside the selected tree.
check_ancestors() {
    local parent
    parent="$(dirname "$1")"
    while [[ "$parent" != . && "$parent" != / ]]; do
        if [[ -L "$parent" ]]; then
            echo "Error: Refusing bundle installation through symlink ancestor: $parent" >&2
            return 1
        fi
        parent="$(dirname "$parent")"
    done
}
# Preflight the entire bundle before creating or replacing installed files.
for path in "${FILES[@]}"; do
    target="$DEST/.agents/skills/$path"
    check_ancestors "$target"
    if [[ -e "$target" || -L "$target" ]]; then
        if [[ -L "$target" || ! -f "$target" ]] || ! cmp -s "$stage/$path" "$target"; then
            echo "Error: Existing file differs: $target. Review its diff and reconcile explicitly; no skill files installed." >&2
            exit 1
        fi
    fi
done
for path in "${PROJECT_FILES[@]}"; do
    target="$DEST/$path"
    check_ancestors "$target"
    if [[ -e "$target" || -L "$target" ]]; then
        if [[ -L "$target" || ! -f "$target" ]] || ! cmp -s "$stage/project/$path" "$target"; then
            echo "Error: Existing optional OpenSpec file differs: $target. Reconcile explicitly; no files installed." >&2
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
for path in "${PROJECT_FILES[@]}"; do
    target="$DEST/$path"
    if [[ ! -e "$target" ]]; then
        mkdir -p "$(dirname "$target")"
        cp "$stage/project/$path" "$target"
        case "$path" in *.sh|*.mjs) chmod +x "$target" ;; esac
    fi
done
if ((WITH_OPENSPEC)); then
    echo 'Pinned OpenSpec tool/schema bundle installed; review local policy before npm ci --prefix tools/openspec and assistant init.'
fi
echo "Installed AgenticLens setup and operator skills ($VERSION) in $DEST/.agents/skills."
echo 'Next, explicitly install/repair platform binaries with: bash .agents/skills/agentic-lens-setup/scripts/download.sh'
echo 'Git tracking and ignore choices remain with the repository owner.'

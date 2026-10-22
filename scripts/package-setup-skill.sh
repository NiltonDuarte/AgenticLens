#!/usr/bin/env bash
#
# package-setup-skill.sh - Packages the agentic-lens-setup skill for distribution
#
# Bundles .opencode/skills/agentic-lens-setup into standalone release archives
# (tar.gz and zip) alongside SHA256SUMS. Strictly excludes internal repository
# skills, precompiled binaries, and temporary editor artifacts.
#
# Zero emojis mandate enforced.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE_DIR="${REPO_ROOT}/.opencode/skills/agentic-lens-setup"
DEST_DIR="${REPO_ROOT}/dist/skills"
VERSION=""
DRY_RUN=0

usage() {
    cat <<EOF
Usage: package-setup-skill.sh [OPTIONS]

Options:
  --dest-dir <dir>       Destination directory for skill archives (default: dist/skills)
  --version <version>    Release version identifier (default: git describe or v0.0.0-dev)
  --dry-run              Display packaging plan without creating archives
  --help, -h             Show this help message
EOF
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dest-dir)
            DEST_DIR="$2"
            shift 2
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=1
            shift
            ;;
        --help|-h)
            usage
            ;;
        *)
            echo "Error: Unknown argument: $1" >&2
            exit 1
            ;;
    esac
done

# Resolve destination directory to absolute path
if [[ "$DEST_DIR" != /* ]]; then
    DEST_DIR="$(pwd)/${DEST_DIR}"
fi

# Resolve version if not specified
if [[ -z "$VERSION" ]]; then
    VERSION="$(git -C "$REPO_ROOT" describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")"
fi

# Validate source directory and required files
if [[ ! -d "$SOURCE_DIR" ]]; then
    echo "Error: Setup skill source directory not found at $SOURCE_DIR" >&2
    exit 1
fi

SKILL_MANIFEST="${SOURCE_DIR}/SKILL.md"
if [[ ! -f "$SKILL_MANIFEST" ]]; then
    echo "Error: Skill manifest missing at $SKILL_MANIFEST" >&2
    exit 1
fi

DOWNLOAD_SCRIPT="${SOURCE_DIR}/scripts/download.sh"
if [[ ! -f "$DOWNLOAD_SCRIPT" ]]; then
    echo "Error: Download script missing at $DOWNLOAD_SCRIPT" >&2
    exit 1
fi

GITIGNORE_FILE="${SOURCE_DIR}/bin/.gitignore"
if [[ ! -f "$GITIGNORE_FILE" ]]; then
    mkdir -p "${SOURCE_DIR}/bin"
    cat <<'GITIGNORE_EOF' > "$GITIGNORE_FILE"
# Prevent accidental git tracking of precompiled binaries
*
!.gitignore
!/.gitignore
GITIGNORE_EOF
fi

# Validate gitignore patterns
if ! grep -q '^\*$' "$GITIGNORE_FILE" || ! grep -q -E '^!/?\.gitignore$' "$GITIGNORE_FILE"; then
    echo "Error: bin/.gitignore missing required wildcard exclusion patterns" >&2
    exit 1
fi

TAR_NAME="agentic-lens-setup.tar.gz"
ZIP_NAME="agentic-lens-setup.zip"

if [[ $DRY_RUN -eq 1 ]]; then
    echo "Packaging dry run for agentic-lens-setup skill:"
    echo "  Version:     ${VERSION}"
    echo "  Source:      ${SOURCE_DIR}"
    echo "  Destination: ${DEST_DIR}"
    echo "  Included files:"
    echo "    - SKILL.md"
    echo "    - scripts/download.sh"
    echo "    - bin/.gitignore"
    echo "  Target archives:"
    echo "    - ${DEST_DIR}/${TAR_NAME}"
    echo "    - ${DEST_DIR}/${ZIP_NAME}"
    echo "    - ${DEST_DIR}/SHA256SUMS"
    exit 0
fi

# Prepare isolated staging directory
STAGING_DIR="$(mktemp -d)"
cleanup() {
    rm -rf "$STAGING_DIR"
}
trap cleanup EXIT

# Copy strictly the required files for distribution
cp "$SKILL_MANIFEST" "${STAGING_DIR}/SKILL.md"

mkdir -p "${STAGING_DIR}/scripts"
cp "$DOWNLOAD_SCRIPT" "${STAGING_DIR}/scripts/download.sh"
chmod +x "${STAGING_DIR}/scripts/download.sh"

mkdir -p "${STAGING_DIR}/bin"
cp "$GITIGNORE_FILE" "${STAGING_DIR}/bin/.gitignore"

# Safety Assertions: ensure no binary executables were included
if find "${STAGING_DIR}" -type f -name "agentic-lens*" | grep -q .; then
    echo "Error: Precompiled binary found in staging directory. Aborting packaging." >&2
    exit 1
fi

# Safety Assertions: ensure no internal skills were included
for internal_skill in "agentic-lens-dev" "agentic-lens-po" "agentic-lens-review"; do
    if find "${STAGING_DIR}" -name "*${internal_skill}*" | grep -q .; then
        echo "Error: Internal skill artifact ${internal_skill} leaked into staging. Aborting packaging." >&2
        exit 1
    fi
done

# Create destination directory
mkdir -p "$DEST_DIR"

# Package archives
echo "Packaging ${TAR_NAME}..."
rm -f "${DEST_DIR}/${TAR_NAME}"
tar -czf "${DEST_DIR}/${TAR_NAME}" -C "$STAGING_DIR" SKILL.md scripts bin

echo "Packaging ${ZIP_NAME}..."
rm -f "${DEST_DIR}/${ZIP_NAME}"
(cd "$STAGING_DIR" && zip -q -r "${DEST_DIR}/${ZIP_NAME}" SKILL.md scripts bin)

# Generate checksums
echo "Generating SHA256SUMS..."
(cd "$DEST_DIR" && sha256sum "${TAR_NAME}" "${ZIP_NAME}" > SHA256SUMS)

echo "Successfully packaged agentic-lens-setup skill:"
echo "  - ${DEST_DIR}/${TAR_NAME}"
echo "  - ${DEST_DIR}/${ZIP_NAME}"
echo "  - ${DEST_DIR}/SHA256SUMS"

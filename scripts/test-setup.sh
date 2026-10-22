#!/usr/bin/env bash
#
# test-setup.sh - Automated verification suite for AgenticLens cross-project setup
#
# Tests artifact downloader OS/arch detection, protective .gitignore creation,
# co-located binary execution, configuration scaffolding, and zero emoji mandate.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEMP_DIR="$(mktemp -d)"

cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

echo "[TEST] Starting AgenticLens setup verification suite..."

# Step 0: Ensure binary is compiled for local testing
LOCAL_BIN="${TEMP_DIR}/agentic-lens"
echo "[TEST] Building temporary AgenticLens binary..."
(cd "${REPO_ROOT}/backend" && go build -o "$LOCAL_BIN" ./cmd/agentic-lens)

if [[ ! -x "$LOCAL_BIN" ]]; then
    echo "[FAIL] Failed to build test binary at $LOCAL_BIN" >&2
    exit 1
fi
echo "[PASS] Test binary compiled successfully."

# Test 1: OS and Architecture Resolution in download script
echo "[TEST] Validating OS and architecture resolution in downloader..."
DRY_RUN_OUTPUT="$("${REPO_ROOT}/scripts/download-agentic-lens.sh" --dry-run)"
if ! echo "$DRY_RUN_OUTPUT" | grep -q "Platform:"; then
    echo "[FAIL] Dry-run output missing Platform specification" >&2
    exit 1
fi
if ! echo "$DRY_RUN_OUTPUT" | grep -q "Architecture:"; then
    echo "[FAIL] Dry-run output missing Architecture specification" >&2
    exit 1
fi
if ! echo "$DRY_RUN_OUTPUT" | grep -q "Download URL:"; then
    echo "[FAIL] Dry-run output missing Download URL" >&2
    exit 1
fi
echo "[PASS] OS and architecture detection verified: $(echo "$DRY_RUN_OUTPUT" | grep 'Platform:') $(echo "$DRY_RUN_OUTPUT" | grep 'Architecture:')"

# Test 2: Local binary installation and protective .gitignore creation
echo "[TEST] Testing local artifact installation and protective gitignore..."
TEST_DEST="${TEMP_DIR}/skill_bin"
"${REPO_ROOT}/scripts/download-agentic-lens.sh" --dest "$TEST_DEST" --local-bin "$LOCAL_BIN"

if [[ ! -x "${TEST_DEST}/agentic-lens" ]]; then
    echo "[FAIL] Installed binary not found or not executable at ${TEST_DEST}/agentic-lens" >&2
    exit 1
fi

GITIGNORE_PATH="${TEST_DEST}/.gitignore"
if [[ ! -f "$GITIGNORE_PATH" ]]; then
    echo "[FAIL] Protective .gitignore was not created at $GITIGNORE_PATH" >&2
    exit 1
fi

if ! grep -q '^\*$' "$GITIGNORE_PATH" || ! grep -q '^!\.gitignore$' "$GITIGNORE_PATH"; then
    echo "[FAIL] Protective .gitignore missing wildcard exclusion patterns" >&2
    exit 1
fi
echo "[PASS] Artifact installation and protective .gitignore verified."

# Test 3: Binary self-test and version JSON output
echo "[TEST] Testing binary version self-test..."
VERSION_JSON="$("${TEST_DEST}/agentic-lens" version --json)"
for key in "version" "commit" "build_date"; do
    if ! echo "$VERSION_JSON" | grep -q "\"$key\""; then
        echo "[FAIL] Version JSON missing expected key: $key" >&2
        exit 1
    fi
done
echo "[PASS] Binary self-test returned valid version metadata."

# Test 4: Root configuration generation via init command
echo "[TEST] Testing init CLI command configuration scaffolding..."
SANDBOX_DIR="${TEMP_DIR}/sandbox_project"
mkdir -p "$SANDBOX_DIR"

(cd "$SANDBOX_DIR" && "${TEST_DEST}/agentic-lens" init \
    --project-name "VerificationProject" \
    --ticket-prefixes "VERIFY,GEN,PO" \
    --workspaces "sandbox_workspaces" \
    --src-folders "core,adapters,ui" \
    --db-path "/tmp/test.db")

CONFIG_FILE="${SANDBOX_DIR}/.agenticlens-config.yaml"
if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "[FAIL] Generated .agenticlens-config.yaml not found at $CONFIG_FILE" >&2
    exit 1
fi

# Verify configuration contents
if ! grep -q "project_name: VerificationProject" "$CONFIG_FILE"; then
    echo "[FAIL] Configuration missing project_name override" >&2
    exit 1
fi
if ! grep -q "prefix: VERIFY" "$CONFIG_FILE" || ! grep -q "prefix: GEN" "$CONFIG_FILE" || ! grep -q "prefix: PO" "$CONFIG_FILE"; then
    echo "[FAIL] Configuration missing peer ticket prefixes override" >&2
    exit 1
fi
if ! grep -q "dir: sandbox_workspaces" "$CONFIG_FILE"; then
    echo "[FAIL] Configuration missing workspaces dir override" >&2
    exit 1
fi
if ! grep -q "path: /tmp/test.db" "$CONFIG_FILE"; then
    echo "[FAIL] Configuration missing database path override" >&2
    exit 1
fi
if ! grep -q "artifacts: \[\]" "$CONFIG_FILE"; then
    echo "[FAIL] Configuration missing zero default artifacts: []" >&2
    exit 1
fi

# Test collision without force
COLLISION_EXIT=0
(cd "$SANDBOX_DIR" && "${TEST_DEST}/agentic-lens" init \
    --project-name "CollidingProject" \
    --src-folders "src" \
    --ticket-prefixes "COLLIDE") || COLLISION_EXIT=$?
if [[ $COLLISION_EXIT -eq 0 ]]; then
    echo "[FAIL] Expected init to fail when file already exists without --force" >&2
    exit 1
fi

# Test overwrite with force
(cd "$SANDBOX_DIR" && "${TEST_DEST}/agentic-lens" init \
    --project-name "ForcedOverwrite" \
    --src-folders "core,adapters,ui" \
    --ticket-prefixes "VERIFY" \
    --force)
if ! grep -q "project_name: ForcedOverwrite" "$CONFIG_FILE"; then
    echo "[FAIL] Configuration was not updated after --force init" >&2
    exit 1
fi

# Test dry-run output
DRY_OUTPUT="$(cd "$SANDBOX_DIR" && "${TEST_DEST}/agentic-lens" init \
    --project-name "DryRunOutput" \
    --src-folders "core,adapters,ui" \
    --ticket-prefixes "VERIFY" \
    --dry-run)"
if ! echo "$DRY_OUTPUT" | grep -q "project_name: DryRunOutput"; then
    echo "[FAIL] Dry-run output missing project_name" >&2
    exit 1
fi

# Test 4b: Assert edge cases: custom prefixes, spaces in project names, colliding files, invalid flags, missing binaries, permission handling
echo "[TEST] Testing edge cases (spaces in names, custom prefixes, invalid flags, missing binaries, permissions)..."
EDGE_DIR="${TEMP_DIR}/edge_project with spaces"
mkdir -p "$EDGE_DIR"

# Edge Case 1: Spaces in project name and workspaces directory
(cd "$EDGE_DIR" && "${TEST_DEST}/agentic-lens" init \
    --project-name "My Multi Word Project" \
    --workspaces "my custom workspaces" \
    --src-folders "edge_src,ui_src" \
    --ticket-prefixes "EDGE_PREFIX")

EDGE_CONFIG="${EDGE_DIR}/.agenticlens-config.yaml"
if [[ ! -f "$EDGE_CONFIG" ]]; then
    echo "[FAIL] Edge configuration with spaces was not created at $EDGE_CONFIG" >&2
    exit 1
fi
if ! grep -q "project_name: My Multi Word Project" "$EDGE_CONFIG"; then
    echo "[FAIL] Edge configuration missing spaces in project_name" >&2
    exit 1
fi
if ! grep -q "dir: my custom workspaces" "$EDGE_CONFIG"; then
    echo "[FAIL] Edge configuration missing spaces in workspaces dir" >&2
    exit 1
fi
if ! grep -q "prefix: EDGE_PREFIX" "$EDGE_CONFIG"; then
    echo "[FAIL] Edge configuration missing custom prefix EDGE_PREFIX" >&2
    exit 1
fi

# Edge Case 2: Invalid flags exit non-zero with error message
INVALID_FLAG_EXIT=0
"${TEST_DEST}/agentic-lens" init --nonexistent-unrecognized-flag=1 2>/dev/null || INVALID_FLAG_EXIT=$?
if [[ $INVALID_FLAG_EXIT -eq 0 ]]; then
    echo "[FAIL] Expected non-zero exit for invalid flag" >&2
    exit 1
fi

# Edge Case 3: Missing local binary in downloader triggers clean failure
MISSING_BIN_EXIT=0
"${REPO_ROOT}/scripts/download-agentic-lens.sh" --dest "${TEMP_DIR}/invalid_dest" --local-bin "/path/to/nonexistent/binary_xyz" 2>/dev/null || MISSING_BIN_EXIT=$?
if [[ $MISSING_BIN_EXIT -eq 0 ]]; then
    echo "[FAIL] Expected non-zero exit for missing local binary in downloader" >&2
    exit 1
fi

# Edge Case 4: Permission handling (read-only destination)
if [[ $(id -u) -ne 0 ]]; then
    RO_DIR="${TEMP_DIR}/readonly_dir"
    mkdir -p "$RO_DIR"
    chmod 0555 "$RO_DIR"
    PERM_EXIT=0
    "${TEST_DEST}/agentic-lens" init \
        --project-name "Protected" \
        --src-folders "src" \
        --ticket-prefixes "PROT" \
        --output "${RO_DIR}/config.yaml" 2>/dev/null || PERM_EXIT=$?
    chmod 0755 "$RO_DIR"
    if [[ $PERM_EXIT -eq 0 ]]; then
        echo "[FAIL] Expected non-zero exit when writing to read-only directory" >&2
        exit 1
    fi
fi

# Edge Case 5: Mandatory init flags enforcement (fails with exit code 1)
MISSING_PN_OUT="$("${TEST_DEST}/agentic-lens" init --src-folders "src" --ticket-prefixes "AL" 2>&1 || true)"
if ! echo "$MISSING_PN_OUT" | grep -q "Error: missing required flag --project-name"; then
    echo "[FAIL] Expected missing --project-name error message, got: $MISSING_PN_OUT" >&2
    exit 1
fi

MISSING_SF_OUT="$("${TEST_DEST}/agentic-lens" init --project-name "Acme" --ticket-prefixes "AL" 2>&1 || true)"
if ! echo "$MISSING_SF_OUT" | grep -q "Error: missing required flag --src-folders"; then
    echo "[FAIL] Expected missing --src-folders error message, got: $MISSING_SF_OUT" >&2
    exit 1
fi

MISSING_TP_OUT="$("${TEST_DEST}/agentic-lens" init --project-name "Acme" --src-folders "src" 2>&1 || true)"
if ! echo "$MISSING_TP_OUT" | grep -q "missing required flag --ticket-prefixes"; then
    echo "[FAIL] Expected missing --ticket-prefixes error message, got: $MISSING_TP_OUT" >&2
    exit 1
fi

# Edge Case 6: Missing config instructions for commands
NOCONFIG_DIR="${TEMP_DIR}/empty_noconfig"
mkdir -p "$NOCONFIG_DIR"

CMD_ROOT_OUT="$(cd "$NOCONFIG_DIR" && "${TEST_DEST}/agentic-lens" --workspaces "$NOCONFIG_DIR" 2>&1 || true)"
if ! echo "$CMD_ROOT_OUT" | grep -q "Project configuration not found. Please run 'agentic-lens init'"; then
    echo "[FAIL] Expected missing config instruction for root command, got: $CMD_ROOT_OUT" >&2
    exit 1
fi

CMD_SERVER_OUT="$(cd "$NOCONFIG_DIR" && "${TEST_DEST}/agentic-lens" server --port 29999 2>&1 || true)"
if ! echo "$CMD_SERVER_OUT" | grep -q "Project configuration not found. Please run 'agentic-lens init'"; then
    echo "[FAIL] Expected missing config instruction for server command, got: $CMD_SERVER_OUT" >&2
    exit 1
fi

CMD_GEN_OUT="$(cd "$NOCONFIG_DIR" && "${TEST_DEST}/agentic-lens" generate --ticket-id "TEST-1" 2>&1 || true)"
if ! echo "$CMD_GEN_OUT" | grep -q "Project configuration not found. Please run 'agentic-lens init'"; then
    echo "[FAIL] Expected missing config instruction for generate command, got: $CMD_GEN_OUT" >&2
    exit 1
fi
echo "[PASS] Edge cases (spaces in names, custom prefixes, invalid flags, missing binaries, permissions) passed."

echo "[PASS] Init command scaffolding, collision protection, force overwrite, and dry-run verified."

# Test 5: Verify setup skill files exist and are correctly configured
echo "[TEST] Verifying setup skill files in repository..."
SKILL_DIR="${REPO_ROOT}/.opencode/skills/agentic-lens-setup"
if [[ ! -f "${SKILL_DIR}/SKILL.md" ]]; then
    echo "[FAIL] Skill manifest missing at ${SKILL_DIR}/SKILL.md" >&2
    exit 1
fi
if [[ ! -x "${SKILL_DIR}/scripts/download.sh" ]]; then
    echo "[FAIL] Skill downloader script missing or not executable at ${SKILL_DIR}/scripts/download.sh" >&2
    exit 1
fi
if [[ ! -f "${SKILL_DIR}/bin/.gitignore" ]]; then
    echo "[FAIL] Skill binary .gitignore missing at ${SKILL_DIR}/bin/.gitignore" >&2
    exit 1
fi
echo "[PASS] Skill manifest and support scripts verified."

# Test 6: Verify cross-project setup manual exists, non-empty, and has required sections
echo "[TEST] Verifying cross-project setup manual (docs/SETUP_MANUAL.md)..."
MANUAL_PATH="${REPO_ROOT}/docs/SETUP_MANUAL.md"
if [[ ! -s "$MANUAL_PATH" ]]; then
    echo "[FAIL] Setup manual missing or empty at $MANUAL_PATH" >&2
    exit 1
fi

REQUIRED_SECTIONS=(
    "Platform Architecture"
    "Architectural Delineation"
    "Installation and Setup Pathways"
    "Configuration Reference"
    "Agnostic Agent Workflow"
    "CLI and Web Dashboard Reference"
    "Troubleshooting"
)

for section in "${REQUIRED_SECTIONS[@]}"; do
    if ! grep -q "$section" "$MANUAL_PATH"; then
        echo "[FAIL] Setup manual missing required section: $section" >&2
        exit 1
    fi
done
echo "[PASS] Setup manual verified with all required section headers."

# Test 7: Verify standalone setup skill packaging script and archive contents
echo "[TEST] Testing standalone setup skill packaging script..."
PACKAGING_SCRIPT="${REPO_ROOT}/scripts/package-setup-skill.sh"
if [[ ! -x "$PACKAGING_SCRIPT" ]]; then
    echo "[FAIL] Packaging script missing or not executable at $PACKAGING_SCRIPT" >&2
    exit 1
fi

# Test dry-run
SKILL_DRY_RUN="$("$PACKAGING_SCRIPT" --dry-run)"
if ! echo "$SKILL_DRY_RUN" | grep -q "Packaging dry run for agentic-lens-setup skill"; then
    echo "[FAIL] Packaging dry-run output missing expected title" >&2
    exit 1
fi

# Test actual archive generation in isolated destination
TEST_SKILLS_DEST="${TEMP_DIR}/packaged_skills"
"$PACKAGING_SCRIPT" --dest-dir "$TEST_SKILLS_DEST"

TAR_ARCHIVE="${TEST_SKILLS_DEST}/agentic-lens-setup.tar.gz"
ZIP_ARCHIVE="${TEST_SKILLS_DEST}/agentic-lens-setup.zip"
CHECKSUMS_FILE="${TEST_SKILLS_DEST}/SHA256SUMS"

if [[ ! -f "$TAR_ARCHIVE" ]]; then
    echo "[FAIL] Expected tar.gz archive not found at $TAR_ARCHIVE" >&2
    exit 1
fi

if [[ ! -f "$ZIP_ARCHIVE" ]]; then
    echo "[FAIL] Expected zip archive not found at $ZIP_ARCHIVE" >&2
    exit 1
fi

if [[ ! -f "$CHECKSUMS_FILE" ]]; then
    echo "[FAIL] Expected SHA256SUMS file not found at $CHECKSUMS_FILE" >&2
    exit 1
fi

# Verify checksums integrity
(cd "$TEST_SKILLS_DEST" && sha256sum -c SHA256SUMS >/dev/null)
echo "[PASS] SHA256 checksums verified for packaged skill archives."

# Extract tar.gz into isolated sandbox and verify archive contents
EXTRACT_TAR_DIR="${TEMP_DIR}/extract_tar"
mkdir -p "$EXTRACT_TAR_DIR"
tar -xzf "$TAR_ARCHIVE" -C "$EXTRACT_TAR_DIR"

if [[ ! -f "${EXTRACT_TAR_DIR}/SKILL.md" ]]; then
    echo "[FAIL] Packaged archive missing SKILL.md" >&2
    exit 1
fi

if [[ ! -x "${EXTRACT_TAR_DIR}/scripts/download.sh" ]]; then
    echo "[FAIL] Packaged archive missing executable scripts/download.sh" >&2
    exit 1
fi

if [[ ! -f "${EXTRACT_TAR_DIR}/bin/.gitignore" ]]; then
    echo "[FAIL] Packaged archive missing bin/.gitignore" >&2
    exit 1
fi

# Assert protective gitignore patterns in archive
if ! grep -q '^\*$' "${EXTRACT_TAR_DIR}/bin/.gitignore" || ! grep -q -E '^!/?\.gitignore$' "${EXTRACT_TAR_DIR}/bin/.gitignore"; then
    echo "[FAIL] Packaged bin/.gitignore missing wildcard exclusion patterns" >&2
    exit 1
fi

# Assert NO binary executables leaked into archive
if find "$EXTRACT_TAR_DIR" -type f -name "agentic-lens*" | grep -q .; then
    echo "[FAIL] Precompiled binary found inside packaged skill archive" >&2
    exit 1
fi

# Assert NO internal repository skills leaked into archive
for internal in "agentic-lens-dev" "agentic-lens-po" "agentic-lens-review"; do
    if find "$EXTRACT_TAR_DIR" -name "*${internal}*" | grep -q .; then
        echo "[FAIL] Internal skill ${internal} found inside packaged skill archive" >&2
        exit 1
    fi
done

# Extract zip into isolated sandbox and verify zip contents match
EXTRACT_ZIP_DIR="${TEMP_DIR}/extract_zip"
mkdir -p "$EXTRACT_ZIP_DIR"
unzip -q "$ZIP_ARCHIVE" -d "$EXTRACT_ZIP_DIR"

if [[ ! -f "${EXTRACT_ZIP_DIR}/SKILL.md" ]]; then
    echo "[FAIL] Packaged zip archive missing SKILL.md" >&2
    exit 1
fi

if [[ ! -f "${EXTRACT_ZIP_DIR}/scripts/download.sh" ]]; then
    echo "[FAIL] Packaged zip archive missing scripts/download.sh" >&2
    exit 1
fi

if [[ ! -f "${EXTRACT_ZIP_DIR}/bin/.gitignore" ]]; then
    echo "[FAIL] Packaged zip archive missing bin/.gitignore" >&2
    exit 1
fi

if find "$EXTRACT_ZIP_DIR" -type f -name "agentic-lens*" | grep -q .; then
    echo "[FAIL] Precompiled binary found inside packaged zip archive" >&2
    exit 1
fi

echo "[PASS] Standalone skill packaging verified (tar.gz, zip, SHA256SUMS, contents, and exclusions)."

# Test 8: Zero emojis check
echo "[TEST] Verifying zero emojis mandate across new files..."
CHECK_FILES=(
    "${REPO_ROOT}/scripts/download-agentic-lens.sh"
    "${REPO_ROOT}/scripts/package-setup-skill.sh"
    "${REPO_ROOT}/scripts/test-setup.sh"
    "${REPO_ROOT}/docs/SETUP_MANUAL.md"
    "${SKILL_DIR}/SKILL.md"
    "${SKILL_DIR}/scripts/download.sh"
    "${SKILL_DIR}/bin/.gitignore"
    "${REPO_ROOT}/backend/internal/adapters/config/writer.go"
    "${REPO_ROOT}/backend/internal/adapters/config/writer_test.go"
)

# Regex matching common unicode emoji ranges
EMOJI_PATTERN="[\x{1F600}-\x{1F64F}\x{1F300}-\x{1F5FF}\x{1F680}-\x{1F6FF}\x{1F700}-\x{1F77F}\x{1F780}-\x{1F7FF}\x{1F800}-\x{1F8FF}\x{1F900}-\x{1F9FF}\x{1FA00}-\x{1FA6F}\x{1FA70}-\x{1FAFF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}]"

for file in "${CHECK_FILES[@]}"; do
    if grep -P "$EMOJI_PATTERN" "$file" >/dev/null 2>&1; then
        echo "[FAIL] Emoji found in file: $file" >&2
        exit 1
    fi
done
echo "[PASS] Zero emojis verified across all files."

echo "[PASS] All AgenticLens setup verification tests passed successfully."

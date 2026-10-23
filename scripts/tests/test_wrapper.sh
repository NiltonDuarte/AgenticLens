#!/usr/bin/env bash
#
# test_wrapper.sh - Test suite for agentic-lens-wrapper.sh and download scripts
#
# Validates platform detection, execution dispatching, self-healing downloads,
# background update notification, update subcommands, and emoji-free compliance.
#
# Zero emojis permitted.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
WRAPPER_SRC="${REPO_ROOT}/scripts/agentic-lens-wrapper.sh"
DOWNLOADER_SRC="${REPO_ROOT}/scripts/download-agentic-lens.sh"

echo "=== Running Wrapper and Downloader Script Tests ==="

# Helper: Assert substring
assert_contains() {
    local text="$1"
    local pattern="$2"
    local msg="$3"
    if ! echo "$text" | grep -q "$pattern"; then
        echo "FAIL: $msg" >&2
        echo "Expected pattern: $pattern" >&2
        echo "Actual text: $text" >&2
        exit 1
    fi
}

assert_not_contains() {
    local text="$1"
    local pattern="$2"
    local msg="$3"
    if echo "$text" | grep -q "$pattern"; then
        echo "FAIL: $msg" >&2
        echo "Unexpected pattern found: $pattern" >&2
        exit 1
    fi
}

# 1. Dispatching test
test_dispatching() {
    echo "Running test: Wrapper dispatches to matching platform binary..."
    local tmp_dir
    tmp_dir="$(mktemp -d)"
    trap 'rm -rf "$tmp_dir"' RETURN

    cp "$WRAPPER_SRC" "$tmp_dir/agentic-lens"
    chmod +x "$tmp_dir/agentic-lens"

    local os_raw
    os_raw="$(uname -s)"
    local os="linux"
    if [ "$os_raw" = "Darwin" ]; then os="darwin"; fi

    local arch_raw
    arch_raw="$(uname -m)"
    local arch="amd64"
    if [ "$arch_raw" = "arm64" ] || [ "$arch_raw" = "aarch64" ]; then arch="arm64"; fi

    local mock_bin="$tmp_dir/agentic-lens-${os}-${arch}"
    cat <<'EOF' > "$mock_bin"
#!/bin/sh
echo "MOCK_DISPATCH: $@"
EOF
    chmod +x "$mock_bin"

    local output
    output="$(AGENTIC_LENS_NO_UPDATE_CHECK=1 "$tmp_dir/agentic-lens" alpha beta gamma)"
    assert_contains "$output" "MOCK_DISPATCH: alpha beta gamma" "Wrapper did not pass arguments to target binary"
    echo "PASS: Wrapper dispatching verified."
}

# 2. Update notice test
test_update_notice() {
    echo "Running test: Background update notification on newer version..."
    local tmp_dir
    tmp_dir="$(mktemp -d)"
    trap 'rm -rf "$tmp_dir"' RETURN

    cp "$WRAPPER_SRC" "$tmp_dir/agentic-lens"
    chmod +x "$tmp_dir/agentic-lens"

    local os_raw
    os_raw="$(uname -s)"
    local os="linux"
    if [ "$os_raw" = "Darwin" ]; then os="darwin"; fi

    local arch_raw
    arch_raw="$(uname -m)"
    local arch="amd64"
    if [ "$arch_raw" = "arm64" ] || [ "$arch_raw" = "aarch64" ]; then arch="arm64"; fi

    local mock_bin="$tmp_dir/agentic-lens-${os}-${arch}"
    cat <<'EOF' > "$mock_bin"
#!/bin/sh
if [ "$1" = "version" ]; then
    echo "v1.0.0"
    exit 0
fi
echo "MOCK_OK"
EOF
    chmod +x "$mock_bin"

    # Simulate detected newer version in cache
    echo "v1.5.0" > "$tmp_dir/.latest_version"

    local stderr_output
    stderr_output="$("$tmp_dir/agentic-lens" version 2>&1 >/dev/null || true)"
    assert_contains "$stderr_output" "Notice: AgenticLens update available: v1.5.0 (current: v1.0.0)" "Update notification not shown on stderr"

    # Verify suppressed when AGENTIC_LENS_NO_UPDATE_CHECK=1
    local suppressed_output
    suppressed_output="$(AGENTIC_LENS_NO_UPDATE_CHECK=1 "$tmp_dir/agentic-lens" version 2>&1 >/dev/null || true)"
    assert_not_contains "$suppressed_output" "Notice: AgenticLens update available" "Update notification not suppressed by env var"
    echo "PASS: Update notice and suppression verified."
}

# 3. Downloader dry run and host inclusion
test_downloader_options() {
    echo "Running test: Downloader options and multi-platform flags..."
    local out1
    out1="$("$DOWNLOADER_SRC" --dry-run)"
    assert_contains "$out1" "Target Wrapper:" "Downloader dry-run missing target wrapper"

    local out2
    out2="$("$DOWNLOADER_SRC" --dry-run --include-host darwin-arm64)"
    assert_contains "$out2" "agentic-lens-darwin-arm64" "Downloader failed to include darwin-arm64"

    local out3
    out3="$("$DOWNLOADER_SRC" --dry-run --all-platforms)"
    assert_contains "$out3" "agentic-lens-linux-amd64" "Downloader --all-platforms missing linux-amd64"
    assert_contains "$out3" "agentic-lens-darwin-arm64" "Downloader --all-platforms missing darwin-arm64"
    assert_contains "$out3" "agentic-lens-windows-amd64.exe" "Downloader --all-platforms missing windows-amd64.exe"
    echo "PASS: Downloader options and multi-platform flags verified."
}

# 4. Zero emojis check
test_zero_emojis() {
    echo "Running test: Zero emojis in scripts and documentation..."
    local files=(
        "${REPO_ROOT}/scripts/agentic-lens-wrapper.sh"
        "${REPO_ROOT}/scripts/download-agentic-lens.sh"
        "${REPO_ROOT}/.opencode/skills/agentic-lens-setup/scripts/agentic-lens-wrapper.sh"
        "${REPO_ROOT}/.opencode/skills/agentic-lens-setup/scripts/download.sh"
        "${REPO_ROOT}/.opencode/skills/agentic-lens-setup/SKILL.md"
        "${REPO_ROOT}/docs/SETUP_MANUAL.md"
        "${REPO_ROOT}/alspaces/GEN-7/CONTRACT.md"
    )

    for f in "${files[@]}"; do
        if grep -P "[\x{1F300}-\x{1F9FF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}\x{1F1E6}-\x{1F1FF}]" "$f" >/dev/null 2>&1; then
            echo "FAIL: File $f contains forbidden emojis" >&2
            exit 1
        fi
    done
    echo "PASS: Zero emojis compliance verified."
}

test_dispatching
test_update_notice
test_downloader_options
test_zero_emojis

echo "=== All Wrapper and Downloader Tests Passed Successfully ==="

#!/usr/bin/env bash
# Bootstrap only: native manager owns configuration, installation and routing.
set -euo pipefail
PROJECT=. DEST="$HOME/.local/bin" VERSION='' MANAGER_VERSION=''
ENGINE='' ENGINE_SHA='' MANAGER='' MANAGER_SHA=''
SOURCE='' STORE='' STATE='' INCLUDE='' ALL=0 FORCE=0 DRY=0 INIT=0
fail() { printf 'Error: %s\n' "$*" >&2; exit 1; }
usage() {
 cat <<'USAGE'
Usage: download.sh --project DIR [--version EXACT_ENGINE_TAG] [--dest USER_BIN_DIR]
  --manager-version EXACT_TAG     Exact release for first manager installation
  --local-manager-bin PATH --manager-sha256 HASH  Verified offline manager
  --local-bin PATH --sha256 HASH  Verified offline engine
  --init-config                  Create an explicit incomplete shared scaffold
  --all-platforms | --include-host OS-ARCH  Cache additional engine platforms
  --capture-source codex|opencode --capture-store ABSOLUTE_PATH
  --capture-state-dir PATH
  --force                        Repair an immutable engine cache leaf
  --dry-run                      No network or destination writes
The native manager validates shared configuration. First remote installation
requires --manager-version (or --version). No latest selection, shell YAML
parser, PATH/profile changes or implicit trust changes.
USAGE
}
while (($#)); do
 case "$1" in
  --project|--capture-repo|--version|--manager-version|--dest|--local-bin|--sha256|--local-manager-bin|--manager-sha256|--include-host|--capture-source|--capture-store|--capture-state-dir)
   (($#>=2)) && [[ -n "$2" && "$2" != --* ]] || fail "$1 requires a value"
   case "$1" in
    --project|--capture-repo) PROJECT=$2 ;; --version) VERSION=$2 ;; --manager-version) MANAGER_VERSION=$2 ;;
    --dest) DEST=$2 ;; --local-bin) ENGINE=$2 ;; --sha256) ENGINE_SHA=$2 ;;
    --local-manager-bin) MANAGER=$2 ;; --manager-sha256) MANAGER_SHA=$2 ;;
    --include-host) INCLUDE=$2 ;; --capture-source) SOURCE=$2 ;; --capture-store) STORE=$2 ;; --capture-state-dir) STATE=$2 ;;
   esac; shift 2 ;;
  --all-platforms) ALL=1; shift ;; --force) FORCE=1; shift ;; --dry-run) DRY=1; shift ;; --init-config) INIT=1; shift ;;
  --help|-h) usage; exit 0 ;; *) fail "Unknown argument $1" ;;
 esac
done
[[ -z "${AGENTIC_LENS_BIN_DIR:-}" ]] || fail 'Use local agenticlens.install_root and --dest'
PROJECT="$(CDPATH= cd -P -- "$PROJECT" && pwd -P)"
[[ "$DEST" == /* ]] || DEST="$PWD/$DEST"
[[ -z "$SOURCE" || "$SOURCE" == codex || "$SOURCE" == opencode ]] || fail 'capture source must be codex or opencode'
if [[ -n "$SOURCE" ]]; then [[ "$STORE" == /* ]] || fail 'capture requires an explicit absolute source store'
elif [[ -n "$STORE$STATE" ]]; then fail 'capture options require --capture-source'; fi
exact() { [[ ${#1} -le 128 && "$1" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?(\+[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?$ ]]; }
digest() {
 if command -v sha256sum >/dev/null 2>&1; then sha256sum "$1" | awk '{print $1}'
 elif command -v shasum >/dev/null 2>&1; then shasum -a 256 "$1" | awk '{print $1}'
 else fail 'Installation requires sha256sum or shasum'; fi
}
fetch() {
 if command -v curl >/dev/null 2>&1; then curl -fsSL --connect-timeout 10 --max-time 120 -o "$2" "$1"
 elif command -v wget >/dev/null 2>&1; then wget -q --timeout=120 -O "$2" "$1"
 else fail 'Download requires curl or wget'; fi
}
case "$(uname -s)" in Linux) OS=linux ;; Darwin) OS=darwin ;; *) fail 'Use download.ps1 on Windows' ;; esac
case "$(uname -m)" in x86_64|amd64) ARCH=amd64 ;; arm64|aarch64) ARCH=arm64 ;; *) fail 'Unsupported host architecture' ;; esac
PLATFORM="$OS-$ARCH" ENTRYPOINT="${DEST%/}/agentic-lens"
STAGE="$(mktemp -d)" SCAFFOLD=''
cleanup() {
 local status=$?
 trap - EXIT HUP INT TERM
 if ((status!=0)) && [[ -n "$SCAFFOLD" ]] && cmp -s -- "$STAGE/.agenticlens-config.yaml" "$SCAFFOLD"; then unlink "$SCAFFOLD"; fi
 [[ ! -d "$STAGE" ]] || rm -r -- "$STAGE"
 exit "$status"
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM
if [[ -n "$MANAGER" ]]; then
 [[ -f "$MANAGER" && "$MANAGER_SHA" =~ ^[a-fA-F0-9]{64}$ ]] || fail '--local-manager-bin requires --manager-sha256'
 cp -- "$MANAGER" "$STAGE/manager"
 [[ "$(digest "$STAGE/manager")" == "$(printf %s "$MANAGER_SHA" | tr A-F a-f)" ]] || fail 'Local manager checksum mismatch'
elif [[ -x "$ENTRYPOINT" ]] && [[ "$("$ENTRYPOINT" manager protocol 2>/dev/null || true)" == 1 ]]; then
 cp -- "$ENTRYPOINT" "$STAGE/manager"
else
 MANAGER_VERSION=${MANAGER_VERSION:-$VERSION}
 exact "$MANAGER_VERSION" || fail 'First installation requires exact --manager-version or verified --local-manager-bin'
 if ((DRY)); then printf 'Manager release: %s\nNative configuration validation pending bootstrap; no network or destination writes.\n' "$MANAGER_VERSION"; exit 0; fi
 base="https://github.com/${AGENTIC_LENS_REPO:-NiltonDuarte/AgenticLens}/releases/download/$MANAGER_VERSION"
 fetch "$base/SHA256SUMS" "$STAGE/SHA256SUMS"
 name="agentic-lens-$PLATFORM"
 fetch "$base/$name" "$STAGE/manager"
 expected="$(awk -v name="$name" '$2==name || $2=="*"name || $2=="bin/"name || $2=="*bin/"name {n++; v=$1} END {if(n!=1) exit 1; print v}' "$STAGE/SHA256SUMS")" || fail 'Missing/ambiguous manager checksum'
 [[ "$expected" =~ ^[a-fA-F0-9]{64}$ && "$(digest "$STAGE/manager")" == "$(printf %s "$expected" | tr A-F a-f)" ]] || fail 'Manager release checksum mismatch'
fi
MANAGER="$STAGE/manager"
magic="$(od -An -tx1 -N4 "$MANAGER" | tr -d ' \n')"
case "$OS:$magic" in linux:7f454c46|darwin:cffaedfe|darwin:cefaedfe|darwin:feedfacf|darwin:feedface|darwin:cafebabe) ;; *) fail 'Bootstrap is not a native manager' ;; esac
chmod 700 "$MANAGER"
[[ "$("$MANAGER" manager protocol)" == 1 ]] || fail 'Native manager interface unavailable'
actual_manager_version="$("$MANAGER" --manager-version --short)"
exact "$actual_manager_version" || fail 'Manager has no exact version identity'
[[ -z "$MANAGER_VERSION" || "$MANAGER_VERSION" == "$actual_manager_version" ]] || fail 'Manager version mismatch'
MANAGER_VERSION=$actual_manager_version
RESOLVE_PROJECT=$PROJECT
if ((INIT)); then
 exact "$VERSION" || fail '--init-config requires exact --version'
 for suffix in yaml yml json; do [[ ! -e "$PROJECT/.agenticlens-config.$suffix" && ! -L "$PROJECT/.agenticlens-config.$suffix" ]] || fail 'Shared config exists; edit its pin explicitly'; done
 printf '# AgenticLens bootstrap scaffold v1: incomplete project configuration\nagenticlens:\n  version: %s\n' "$VERSION" > "$STAGE/.agenticlens-config.yaml"
 RESOLVE_PROJECT=$STAGE
fi
CONTEXT="$("$MANAGER" --al-project "$RESOLVE_PROJECT" --al-resolve-cache)"
VALUES=(); while IFS= read -r value; do VALUES+=("$value"); done <<< "$CONTEXT"
[[ ${#VALUES[@]} == 4 ]] || fail 'Native configuration protocol incompatible'
PIN=${VALUES[0]} INSTALL_ROOT=${VALUES[2]}
[[ -z "$VERSION" || "$VERSION" == "$PIN" ]] || fail '--version conflicts with shared YAML pin'
[[ -z "${AGENTIC_LENS_VERSION:-}" || "$AGENTIC_LENS_VERSION" == "$PIN" ]] || fail 'AGENTIC_LENS_VERSION conflicts with shared YAML pin'
if ((!INIT)); then PROJECT=${VALUES[1]}; fi
if [[ -n "$INCLUDE" ]]; then
 INCLUDE="$(printf '%s' "${INCLUDE//\//-}" | tr '[:upper:]' '[:lower:]')"
 case "$INCLUDE" in macos-arm64|mac-arm64|m1|m2|m3|apple-silicon) INCLUDE=darwin-arm64 ;; macos-amd64|mac-amd64|mac-intel) INCLUDE=darwin-amd64 ;; esac
fi
PLATFORMS=("$PLATFORM")
if ((ALL)); then PLATFORMS=(linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64)
elif [[ -n "$INCLUDE" && "$INCLUDE" != "$PLATFORM" ]]; then PLATFORMS+=("$INCLUDE"); fi
for item in "${PLATFORMS[@]}"; do case "$item" in linux-amd64|linux-arm64|darwin-amd64|darwin-arm64|windows-amd64) ;; *) fail "Unsupported platform $item" ;; esac; done
if [[ -n "$ENGINE" ]]; then
 [[ -f "$ENGINE" && "$ENGINE_SHA" =~ ^[a-fA-F0-9]{64}$ ]] || fail '--local-bin requires --sha256'
 [[ "$(digest "$ENGINE")" == "$(printf %s "$ENGINE_SHA" | tr A-F a-f)" ]] || fail 'Local engine checksum mismatch'
 ((${#PLATFORMS[@]}==1)) || fail '--local-bin cannot install foreign platforms'
fi
if ((DRY)); then
 printf 'Project: %s\nEngine version: %s\nManager version: %s\nNative command: %s\nEngine cache: %s\n' "$PROJECT" "$PIN" "$MANAGER_VERSION" "$ENTRYPOINT" "$INSTALL_ROOT"
 for item in "${PLATFORMS[@]}"; do ext=''; [[ "$item" != windows-* ]] || ext=.exe; printf 'Engine artifact: agentic-lens-engine-%s%s\n' "$item" "$ext"; done
 printf 'No network or destination writes. Required PATH entry: %s.\n' "$DEST"; exit 0
fi
if ((INIT)); then
 SCAFFOLD="$PROJECT/.agenticlens-config.yaml"
 (set -o noclobber; cat "$STAGE/.agenticlens-config.yaml" > "$SCAFFOLD") || fail 'Could not create shared scaffold'
fi
for item in "${PLATFORMS[@]}"; do
 args=(--al-project "$PROJECT" install "$PIN" --platform "$item")
 ((FORCE==0)) || args+=(--force)
 [[ -z "$ENGINE" ]] || args+=(--local-bin "$ENGINE" --sha256 "$ENGINE_SHA")
 "$MANAGER" "${args[@]}"
done
"$MANAGER" self-update --version "$MANAGER_VERSION" --local-bin "$MANAGER" --sha256 "$(digest "$MANAGER")" --dest "$ENTRYPOINT"
if ((INIT)); then
 printf 'Configuration incomplete: bootstrap scaffold defines no project/ticket/code roots.\n'
 printf 'Complete only this unchanged owned scaffold, preserving local settings:\n  (cd -P -- %q && %q --al-project %q init --complete-scaffold --agenticlens-version %q --project-name NAME --src-folders DIRS --ticket-prefixes PREFIXES)\n' "$PROJECT" "$ENTRYPOINT" "$PROJECT" "$PIN"
fi
if [[ -n "$SOURCE" ]]; then
 ((INIT==0)) || fail 'Complete project configuration before capture installation'
 args=(capture install --source "$SOURCE" --store "$STORE" --repo "$PROJECT" --wrapper "$ENTRYPOINT")
 if [[ "$SOURCE" == codex ]]; then args=(capture install-user --repo "$PROJECT" --entrypoint "$ENTRYPOINT"); fi
 [[ -z "$STATE" ]] || args+=(--state-dir "$STATE")
 "$ENTRYPOINT" --al-project "$PROJECT" "${args[@]}"
else printf 'Binary installation only; lifecycle capture not configured.\n'; fi
printf 'Installed native manager: %s\nEngine version: %s\nRequired PATH entry: %s\nNo profile changed. Runtime trust, binding and source extraction require verification.\n' "$ENTRYPOINT" "$PIN" "$DEST"

#!/usr/bin/env bash
# Explicit installation only; runtime dispatch is always offline.
set -euo pipefail
SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
PROJECT=. VERSION_REQUEST='' DEST="$HOME/.local/bin" LOCAL_BIN='' LOCAL_SHA=''
CAPTURE_SOURCE='' CAPTURE_STORE='' CAPTURE_STATE='' INCLUDE=''
ALL=0 FORCE=0 DRY=0 INIT=0
fail() { printf 'Error: %s\n' "$*" >&2; exit 1; }
usage() {
 cat <<'USAGE'
Usage: download.sh --project DIR [--version EXACT_TAG] [--dest USER_WRAPPER_DIR]
  --init-config             Explicitly create missing shared YAML with --version
  --local-bin PATH --sha256 HASH  Verify/install a local native release offline
  --all-platforms | --include-host OS-ARCH  Additional native cache leaves
  --capture-source codex|opencode --capture-store ABSOLUTE_PATH
  --capture-state-dir PATH  Explicit capture state selection
  --force                   Repair an existing release leaf with rollback
  --dry-run                 Preview without writes or network
Shared YAML is authoritative. --version must agree. --dest selects only the user
wrapper directory; local agenticlens.install_root selects the native cache root.
No latest selection, PATH/profile changes, or implicit trust changes.
USAGE
}
while (($#)); do
 case "$1" in
  --project|--capture-repo|--version|--dest|--local-bin|--sha256|--include-host|--capture-source|--capture-store|--capture-state-dir)
   (($#>=2)) && [[ -n "$2" && "$2" != --* ]] || fail "$1 requires a value"
   case "$1" in
    --project|--capture-repo) PROJECT=$2 ;; --version) VERSION_REQUEST=$2 ;; --dest) DEST=$2 ;;
    --local-bin) LOCAL_BIN=$2 ;; --sha256) LOCAL_SHA=$2 ;; --include-host) INCLUDE=$2 ;;
    --capture-source) CAPTURE_SOURCE=$2 ;; --capture-store) CAPTURE_STORE=$2 ;; --capture-state-dir) CAPTURE_STATE=$2 ;;
   esac; shift 2 ;;
  --all-platforms) ALL=1; shift ;; --force) FORCE=1; shift ;; --dry-run) DRY=1; shift ;;
  --init-config) INIT=1; shift ;; --help|-h) usage; exit 0 ;; *) fail "Unknown argument $1" ;;
 esac
done
[[ -z "${AGENTIC_LENS_BIN_DIR:-}" ]] || fail 'AGENTIC_LENS_BIN_DIR retired; use local agenticlens.install_root and --dest'
PROJECT="$(CDPATH= cd -P -- "$PROJECT" && pwd -P)"
[[ -z "$CAPTURE_SOURCE" || "$CAPTURE_SOURCE" == codex || "$CAPTURE_SOURCE" == opencode ]] || fail 'capture source must be codex or opencode'
if [[ -n "$CAPTURE_SOURCE" ]]; then
 [[ "$CAPTURE_STORE" == /* || "$CAPTURE_STORE" =~ ^[A-Za-z]:[\\/] ]] || fail 'capture requires an explicit absolute source store'
elif [[ -n "$CAPTURE_STORE$CAPTURE_STATE" ]]; then fail 'capture options require --capture-source'; fi
SOURCE_WRAPPER="$SCRIPT_DIR/agentic-lens-wrapper.sh"
PS_WRAPPER="$SCRIPT_DIR/agentic-lens-wrapper.ps1"
[[ -f "$SOURCE_WRAPPER" && -f "$PS_WRAPPER" ]] || fail 'Restore complete setup skill: wrapper assets missing'
STAGE='' CACHE_LOCK='' WRAPPER_LOCK='' CAPTURE_APPLY_STARTED=0
TARGETS=() BACKUPS=()
cleanup() {
 local status=$? i
 trap - EXIT HUP INT TERM
 if ((status!=0 && CAPTURE_APPLY_STARTED)); then
  printf 'Capture apply incomplete; validated native/wrapper retained. Prior backups: %s. Inspect capture status and preview owned capture install before recovery; no complete rollback is claimed.\n' "$STAGE" >&2
  STAGE=''
 elif ((status!=0)); then
  for ((i=${#TARGETS[@]}-1;i>=0;i--)); do
   if [[ -n "${BACKUPS[$i]}" ]]; then
    if ! mv -f -- "${BACKUPS[$i]}" "${TARGETS[$i]}"; then
     printf 'Restore prior installation manually from %s\n' "${BACKUPS[$i]}" >&2; STAGE=''
    fi
   else rm -f -- "${TARGETS[$i]}"; fi
  done
 fi
 [[ -z "$STAGE" ]] || rm -rf -- "$STAGE"
 [[ -z "$CACHE_LOCK" ]] || rmdir -- "$CACHE_LOCK"
 [[ -z "$WRAPPER_LOCK" ]] || rmdir -- "$WRAPPER_LOCK"
 exit "$status"
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM
if ((INIT)); then
 [[ "$VERSION_REQUEST" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?(\+[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?$ && ${#VERSION_REQUEST} -le 128 ]] || fail '--init-config requires an exact --version'
 for suffix in yaml yml json; do [[ ! -e "$PROJECT/.agenticlens-config.$suffix" ]] || fail 'Shared config exists; edit its pin explicitly'; done
 if ((DRY)); then
  STAGE="$(mktemp -d)"
  printf 'agenticlens:\n  version: %s\n' "$VERSION_REQUEST" > "$STAGE/.agenticlens-config.yaml"
  CONTEXT="$(sh "$SOURCE_WRAPPER" --al-project "$STAGE" --al-resolve)"
 else
  pin="$PROJECT/.agenticlens-config.yaml"
  (set -o noclobber; printf '# AgenticLens bootstrap scaffold v1: incomplete project configuration\nagenticlens:\n  version: %s\ntelemetry:\n  capture:\n    state_dir: ~/.agenticlens/state\n    commit_warning: true\n' "$VERSION_REQUEST" > "$pin") || fail 'Could not create explicit shared pin'
  TARGETS+=("$pin"); BACKUPS+=('')
  CONTEXT="$(sh "$SOURCE_WRAPPER" --al-project "$PROJECT" --al-resolve)"
 fi
else CONTEXT="$(sh "$SOURCE_WRAPPER" --al-project "$PROJECT" --al-resolve)"; fi
VALUES=(); while IFS= read -r value; do VALUES+=("$value"); done <<< "$CONTEXT"
VERSION=${VALUES[0]}
if ((!INIT)); then PROJECT=${VALUES[1]}; fi
INSTALL_ROOT=${VALUES[2]} PRIMARY=${VALUES[3]}
[[ -z "$VERSION_REQUEST" || "$VERSION_REQUEST" == "$VERSION" ]] || fail '--version conflicts with shared YAML pin'
[[ -z "${AGENTIC_LENS_VERSION:-}" || "$AGENTIC_LENS_VERSION" == "$VERSION" ]] || fail 'AGENTIC_LENS_VERSION conflicts with shared YAML pin'
PRIMARY_NAME=${PRIMARY##*/}; PRIMARY_PLATFORM=${PRIMARY_NAME#agentic-lens-}; PRIMARY_PLATFORM=${PRIMARY_PLATFORM%.exe}
[[ "$PRIMARY_PLATFORM" != windows-* || "$DEST" != "$HOME/.local/bin" ]] || DEST="$HOME/.agenticlens"
# Preserve documented/legacy host aliases while cache keys stay canonical.
if [[ -n "$INCLUDE" ]]; then
 INCLUDE="$(printf '%s' "${INCLUDE//\//-}" | tr '[:upper:]' '[:lower:]')"
 case "$INCLUDE" in
  macos-arm64|mac-arm64|m1|m2|m3|apple-silicon) INCLUDE=darwin-arm64 ;;
  macos-amd64|mac-amd64|mac-intel) INCLUDE=darwin-amd64 ;;
 esac
fi
PLATFORMS=("$PRIMARY_PLATFORM")
if ((ALL)); then PLATFORMS=(linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64)
elif [[ -n "$INCLUDE" && "$INCLUDE" != "$PRIMARY_PLATFORM" ]]; then PLATFORMS+=("${INCLUDE//\//-}"); fi
for platform in "${PLATFORMS[@]}"; do
 case "$platform" in linux-amd64|linux-arm64|darwin-amd64|darwin-arm64|windows-amd64) ;; *) fail "Unsupported platform $platform" ;; esac
done
[[ "$DEST" == /* ]] || DEST="$PWD/$DEST"
WRAPPER="${DEST%/}/agentic-lens"
if ((DRY)); then
 printf 'Project: %s\nVersion: %s\nWrapper: %s\nNative cache: %s/bin/%s\n' "$PROJECT" "$VERSION" "$WRAPPER" "$INSTALL_ROOT" "$VERSION"
 for platform in "${PLATFORMS[@]}"; do
  ext=''; [[ "$platform" != windows-* ]] || ext=.exe
  printf 'https://github.com/%s/releases/download/%s/agentic-lens-%s%s\n' "${AGENTIC_LENS_REPO:-NiltonDuarte/AgenticLens}" "$VERSION" "$platform" "$ext"
 done
 printf 'Required PATH entry: %s; no profile changed. Runtime verification pending.\n' "$DEST"
 exit 0
fi
mkdir -p -- "$INSTALL_ROOT" "$DEST"
lock="$INSTALL_ROOT/.install-$VERSION.lock"; mkdir -- "$lock" 2>/dev/null || fail "Installation busy/interrupted: $lock; inspect before retrying"; CACHE_LOCK=$lock
lock="$DEST/.agenticlens-wrapper-install.lock"; mkdir -- "$lock" 2>/dev/null || fail "Wrapper installation busy/interrupted: $lock"; WRAPPER_LOCK=$lock
STAGE="$(mktemp -d "$INSTALL_ROOT/.install-stage.XXXXXXXX")"
digest() {
 if command -v sha256sum >/dev/null 2>&1; then sha256sum "$1" | awk '{print $1}'
 elif command -v shasum >/dev/null 2>&1; then shasum -a 256 "$1" | awk '{print $1}'
 else fail 'Explicit installation requires sha256sum or shasum'; fi
}
fetch() {
 if command -v curl >/dev/null 2>&1; then curl -fsSL --connect-timeout 10 --max-time 120 -o "$2" "$1"
 elif command -v wget >/dev/null 2>&1; then wget -q --timeout=120 -O "$2" "$1"
 else fail 'Explicit download requires curl or wget'; fi
}
replace() {
 local candidate=$1 target=$2 backup='' near
 mkdir -p -- "$(dirname -- "$target")"
 if [[ -e "$target" ]]; then backup="$(mktemp "$STAGE/previous.XXXXXXXX")"; cp -p -- "$target" "$backup"; fi
 TARGETS+=("$target"); BACKUPS+=("$backup")
 near="$(mktemp "${target}.candidate.XXXXXXXX")"
 if ! cp -p -- "$candidate" "$near" || ! mv -f -- "$near" "$target"; then rm -f -- "$near"; return 1; fi
}
for source in "$SOURCE_WRAPPER" "$PS_WRAPPER"; do
 target="$WRAPPER"; [[ "$source" != "$PS_WRAPPER" ]] || target="$DEST/agentic-lens.ps1"
 if [[ -e "$target" ]] && ! cmp -s -- "$source" "$target"; then
  [[ -f "$target.sha256" && "$(cat "$target.sha256")" == "$(digest "$target")" ]] || fail "User-edited/unowned wrapper preserved: $target"
 fi
done
if [[ -n "$LOCAL_BIN" ]]; then
 [[ -f "$LOCAL_BIN" && "$LOCAL_SHA" =~ ^[a-fA-F0-9]{64}$ ]] || fail '--local-bin requires explicit expected --sha256 digest'
 [[ "$(digest "$LOCAL_BIN")" == "$(printf %s "$LOCAL_SHA" | tr A-F a-f)" ]] || fail 'Local native checksum mismatch'
 ((${#PLATFORMS[@]}==1)) || fail '--local-bin cannot install foreign platforms'
else fetch "https://github.com/${AGENTIC_LENS_REPO:-NiltonDuarte/AgenticLens}/releases/download/$VERSION/SHA256SUMS" "$STAGE/SHA256SUMS"; fi
for platform in "${PLATFORMS[@]}"; do
 ext=''; [[ "$platform" != windows-* ]] || ext=.exe
 name="agentic-lens-$platform$ext"; candidate="$STAGE/$name"
 if [[ -n "$LOCAL_BIN" ]]; then cp -- "$LOCAL_BIN" "$candidate"
 else
  fetch "https://github.com/${AGENTIC_LENS_REPO:-NiltonDuarte/AgenticLens}/releases/download/$VERSION/$name" "$candidate"
  expected="$(awk -v name="$name" '$2==name || $2=="*"name || $2=="bin/"name || $2=="*bin/"name {count++; value=$1} END {if(count!=1) exit 1; print value}' "$STAGE/SHA256SUMS")" || fail "Missing/ambiguous checksum for $name"
  [[ "$expected" =~ ^[a-fA-F0-9]{64}$ && "$(digest "$candidate")" == "$(printf %s "$expected" | tr A-F a-f)" ]] || fail "Release checksum mismatch: $name"
 fi
 magic="$(od -An -tx1 -N4 "$candidate" | tr -d ' \n')"
 case "$platform:$magic" in linux-*:7f454c46|darwin-*:cffaedfe|darwin-*:cefaedfe|darwin-*:feedfacf|darwin-*:feedface|darwin-*:cafebabe|windows-*:4d5a*) ;; *) fail "Not a native release executable: $name" ;; esac
 chmod 700 -- "$candidate"
 if [[ "$platform" == "$PRIMARY_PLATFORM" ]]; then
  [[ "$("$candidate" version --short)" == "$VERSION" ]] || fail 'Native version does not match shared pin'
  required=''
  [[ -z "$CAPTURE_SOURCE" ]] || required=3
  if [[ -x "$PRIMARY" ]]; then
   existing_magic="$(od -An -tx1 -N4 "$PRIMARY" | tr -d ' \n')"
   case "$existing_magic" in 7f454c46|cffaedfe|cefaedfe|feedfacf|feedface|cafebabe|4d5a*)
    existing_protocol="$("$PRIMARY" capture protocol 2>/dev/null || true)"
    [[ "$existing_protocol" != 3 ]] || required=3
    if [[ "$existing_protocol" == 2 && -z "$required" ]]; then required=2; fi ;;
   esac
  fi
  observed_protocol="$("$candidate" capture protocol 2>/dev/null || true)"
  [[ -z "$required" || "$observed_protocol" == "$required" ]] || fail "Native hook protocol incompatible (required $required); prior installation preserved"
 fi
done
for platform in "${PLATFORMS[@]}"; do
 ext=''; [[ "$platform" != windows-* ]] || ext=.exe
 name="agentic-lens-$platform$ext"; target="$INSTALL_ROOT/bin/$VERSION/$platform/$name"
 if [[ -f "$target" ]] && cmp -s -- "$STAGE/$name" "$target"; then
  # Repair permissions in place; equal content must retain admitted worker identity.
  [[ -x "$target" ]] || chmod 700 -- "$target"
  continue
 fi
 if [[ -f "$target" ]] && ((!FORCE)); then fail "Immutable cache leaf exists: $target; --force explicitly repairs this release"; fi
 replace "$STAGE/$name" "$target"
done
for source in "$SOURCE_WRAPPER" "$PS_WRAPPER"; do
 target="$WRAPPER"; [[ "$source" != "$PS_WRAPPER" ]] || target="$DEST/agentic-lens.ps1"
 cp -- "$source" "$STAGE/wrapper"; chmod 700 "$STAGE/wrapper"; replace "$STAGE/wrapper" "$target"
 digest "$source" > "$STAGE/wrapper.sha256"; replace "$STAGE/wrapper.sha256" "$target.sha256"
done
if ((INIT)); then
 printf 'Configuration incomplete: bootstrap scaffold defines no project/ticket/code roots.\n'
 if [[ "${observed_protocol:-}" == 3 ]]; then
  printf 'Complete only this unchanged owned scaffold, preserving local settings:\n  (cd -P -- %q && %q --al-project %q init --complete-scaffold --agenticlens-version %q --project-name NAME --src-folders DIRS --ticket-prefixes PREFIXES)\n' "$PROJECT" "$WRAPPER" "$PROJECT" "$VERSION"
 else printf 'Selected legacy release lacks guarded init/capture capability. Ordinary CLI is installed; review an exact supported release pin before adopting new setup/capture features.\n'; fi
fi
if [[ -n "$CAPTURE_SOURCE" ]]; then
 ((INIT==0)) || fail 'Complete project configuration first, then explicitly install capture'
 args=(capture install --source "$CAPTURE_SOURCE" --store "$CAPTURE_STORE" --repo "$PROJECT" --binary "$PRIMARY" --wrapper "$WRAPPER")
 [[ -z "$CAPTURE_STATE" ]] || args+=(--state-dir "$CAPTURE_STATE")
 CAPTURE_APPLY_STARTED=1
 "$WRAPPER" --al-offline --al-project "$PROJECT" "${args[@]}"
else printf 'Binary installation only; lifecycle capture not configured.\n'; fi
if [[ "${observed_protocol:-}" != 3 ]]; then
 printf 'Selected legacy release %s: hook protocol 3 and guarded init unavailable; ordinary CLI installed. New capture setup requires a reviewed exact supported pin.\n' "$VERSION"
fi
printf 'Installed wrapper: %s\nNative: %s\nRequired PATH entry: %s\nNo profile changed. Runtime trust, binding and source extraction require verification.\n' "$WRAPPER" "$PRIMARY" "$DEST"

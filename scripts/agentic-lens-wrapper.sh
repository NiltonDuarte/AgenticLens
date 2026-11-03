#!/bin/sh
# AgenticLens common entrypoint v1: offline shared-YAML release selection.
set -eu
fail() { printf '%s\n' "$*" >&2; exit 2; }
read_bootstrap() {
  input=$1; mode=$2; field=$3
  [ -f "$input" ] || fail "Configuration must be a regular file: $input"
  input_size=$(dd if="$input" bs=1048577 count=1 2>/dev/null | wc -c)
  [ "$input_size" -le 1048576 ] || fail 'bootstrap configuration exceeds 1 MiB'
  dd if="$input" bs=1048577 count=1 2>/dev/null | LC_ALL=C awk -v mode="$mode" -v field="$field" '
# Disposable prototype: supported bootstrap YAML, not a general YAML parser.
function fail(message) { print FILENAME ":" FNR ": " message > "/dev/stderr"; bad=1; exit 2 }
function trim(value) { sub(/^[ \t]+/, "", value); sub(/[ \t]+$/, "", value); return value }
function lex(line,    i,c,q,out) {
    q=""; out=""
    for(i=1;i<=length(line);i++) {
        c=substr(line,i,1)
        if(q!="") {
            if(c=="\\" && q=="\"") fail("escaped double-quoted scalars are unsupported")
            if(c==q) q=""
            out=out c
        } else {
            if(c=="#" && (i==1 || substr(line,i-1,1) ~ /[ \t]/)) break
            if((c=="\"" || c==sprintf("%c",39)) && (i==1 || substr(line,i-1,1) ~ /[ \t:\[,]/)) q=c
            if(c ~ /[&*!|>]/) fail("anchors, aliases, tags and block scalars are unsupported")
            out=out c
        }
    }
    if(q!="") fail("unclosed quoted scalar")
    return out
}
function scalar(value,    q) {
    value=trim(value); q=substr(value,1,1)
    if(q=="\"" || q==sprintf("%c",39)) {
        if(substr(value,length(value),1)!=q || length(value)<2) fail("unsupported quoted scalar")
        value=substr(value,2,length(value)-2)
        if(index(value,q)) fail("embedded quotes are unsupported")
    } else if(value ~ /[ \t#"\047]/) fail("quote values containing spaces or comments")
    if(value=="" || value ~ /[\r\n]/) fail("empty or multiline scalar")
    return value
}
{
    line=$0; sub(/\r$/, "", line)
    if(FNR==1) sub(/^\357\273\277/, "", line)
    if(line ~ /^ *\t/) fail("tab indentation is unsupported")
    line=lex(line)
    if(line ~ /^[ \t]*$/) next
    if(line ~ /^(---|\.\.\.|%)/) fail("YAML document markers/directives are unsupported")
    if(line ~ /^[^ ]/) {
        if(line !~ /^[A-Za-z_][A-Za-z0-9_-]*:([ \t]+.*)?$/) fail("use plain top-level mapping keys")
        key=line; sub(/:.*/, "", key)
        if(roots[key]++) fail("duplicate top-level mapping key")
        active=(key=="agenticlens")
        if(active && trim(substr(line,length(key)+2))!="") fail("agenticlens must be a block mapping")
        next
    }
    if(line ~ /^[ ]*<<:/) fail("YAML merges are unsupported")
    if(!active) next
    if(line !~ /^  [A-Za-z_][A-Za-z0-9_-]*:([ \t]+.*)?$/) fail("agenticlens children require exactly two spaces and plain keys")
    key=substr(line,3); sub(/:.*/, "", key)
    if(children[key]++) fail("duplicate agenticlens key")
    if(key!="version" && key!="install_root") fail("unknown agenticlens key")
    value=scalar(substr(line,length(key)+4))
    if(key=="version") {
        if(mode=="local") fail("local version overrides are prohibited; edit shared YAML")
        if(length(value)>128 || value !~ /^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?(\+[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?$/) fail("version must be an exact vMAJOR.MINOR.PATCH release tag")
        version=value
    } else {
        if(length(value)>4096 || value !~ /^(\/|~\/|[A-Za-z]:[\\\/]|\\\\[^\\]+\\[^\\]+)/ || value ~ /[<>|?*]/) fail("install_root must be an absolute path or ~/path")
        install_root=value
    }
}
END {
    if(bad) exit 2
    if(mode!="local" && version=="") { print FILENAME ": missing agenticlens.version" > "/dev/stderr"; exit 2 }
    if(field=="version") print version
    else if(field=="install_root") print install_root
    else { print "version:" version; print "install_root:" install_root }
}

'
}
project=.
resolve=0
while [ "$#" -gt 0 ]; do
  case "$1" in
    --al-offline) shift ;;
    --al-project) [ "$#" -ge 2 ] || fail '--al-project requires a directory'; project=$2; shift 2 ;;
    --al-resolve) resolve=1; shift ;;
    *) break ;;
  esac
done
root=$(CDPATH= cd -P -- "$project" && pwd -P) || fail 'cannot resolve project directory'
while :; do
  config=
  for suffix in yaml yml json; do
    [ ! -d "$root/.agenticlens-config.$suffix" ] || continue
    if [ -e "$root/.agenticlens-config.$suffix" ] || [ -L "$root/.agenticlens-config.$suffix" ]; then
      [ -f "$root/.agenticlens-config.$suffix" ] || fail "Configuration must be a regular file: $root/.agenticlens-config.$suffix"
      config=$root/.agenticlens-config.$suffix; break
    fi
  done
  [ -z "$config" ] || break
  [ ! -e "$root/.git" ] && [ ! -L "$root/.git" ] || break
  parent=${root%/*}; [ -n "$parent" ] || parent=/
  [ "$parent" != "$root" ] || break
  root=$parent
done
[ -n "$config" ] || fail 'shared project YAML missing; explicit setup must write a version pin'
case "$config" in *.json) fail 'JSON bootstrap is unsupported; migrate shared configuration to .agenticlens-config.yaml' ;; esac
# Parse capture flag arity without interpreting values as flags. A --repo value
# must belong to the selected project; explicit cross-project use supplies the
# matching --al-project prefix. Native argv is never reconstructed.
if [ "${1:-}" = capture ]; then
  native_index=0; pending_value=0; repo_value=''
  for native_arg in "$@"; do
    native_index=$((native_index+1))
    [ "$native_index" -gt 2 ] || continue
    if [ "$pending_value" -ne 0 ]; then
      [ "$pending_value" -ne 2 ] || repo_value=$native_arg
      pending_value=0; continue
    fi
    case "$native_arg" in
      --repo|-repo) pending_value=2 ;;
      --repo=*|-repo=*) repo_value=${native_arg#*=} ;;
      --dry-run|-dry-run|--dry-run=*|-dry-run=*) ;;
      --) break ;;
      -*=*) ;;
      -*) pending_value=1 ;;
      *) break ;;
    esac
  done
  if [ -n "$repo_value" ]; then
    requested_root=$(CDPATH= cd -P -- "$repo_value" && pwd -P) || fail 'cannot resolve native --repo'
    while :; do
      found=0
      for suffix in yaml yml json; do
        [ ! -f "$requested_root/.agenticlens-config.$suffix" ] || found=1
      done
      [ "$found" -eq 0 ] || break
      [ ! -e "$requested_root/.git" ] && [ ! -L "$requested_root/.git" ] || break
      requested_parent=${requested_root%/*}; [ -n "$requested_parent" ] || requested_parent=/
      [ "$requested_parent" != "$requested_root" ] || break
      requested_root=$requested_parent
    done
    [ "$requested_root" = "$root" ] || fail 'Native --repo differs from selected project; supply matching --al-project DIR before capture'
  fi
fi
newline='
'
shared_values=$(read_bootstrap "$config" shared all) || exit 2
version=${shared_values%%"$newline"*}; version=${version#version:}
install_root=${shared_values#*"$newline"}; install_root=${install_root#install_root:}
local_file=$root/.agenticlens-config.local.yaml
if [ -e "$local_file" ] || [ -L "$local_file" ]; then
  [ -f "$local_file" ] || fail "Local configuration must be a regular file: $local_file"
  overlay=$(read_bootstrap "$local_file" local install_root) || exit 2
  [ -z "$overlay" ] || install_root=$overlay
fi
if [ -n "${AGENTIC_LENS_VERSION:-}" ] && [ "$AGENTIC_LENS_VERSION" != "$version" ]; then fail 'AGENTIC_LENS_VERSION conflicts with authoritative shared pin'; fi
# Traverse symlinks before later .. components, including uncreated suffixes.
physical_install_root() {
  pending=$1; current=/; links=0
  case "$pending" in /*) ;; *) fail 'install_root must be absolute on the executing POSIX platform' ;; esac
  while [ -n "$pending" ]; do
    case "$pending" in */*) part=${pending%%/*}; pending=${pending#*/} ;; *) part=$pending; pending='' ;; esac
    case "$part" in ''|.) continue ;; ..) current=${current%/*}; [ -n "$current" ] || current=/; continue ;; esac
    next=${current%/}/$part
    if [ -L "$next" ]; then
      links=$((links+1)); [ "$links" -le 255 ] || fail 'excessive installation-root symlink traversal'
      link=$(readlink "$next") || fail 'cannot read installation-root symlink'
      case "$link" in /*) current=/ ;; esac
      pending=$link/$pending
    else
      [ ! -e "$next" ] || [ -d "$next" ] || fail 'install_root crosses a non-directory'
      current=$next
    fi
  done
  printf '%s\n' "$current"
}
install_root=${install_root:-$HOME/.agenticlens}
case "$install_root" in '~/'*) install_root=$HOME/${install_root#\~/} ;; esac
case "$install_root" in /*) install_root=$(physical_install_root "$install_root") ;; *) fail "install_root must be absolute on this executing shell platform" ;; esac
case "$(uname -s)" in Linux) os=linux ;; Darwin) os=darwin ;; MINGW*|MSYS*|CYGWIN*|Windows_NT) os=windows ;; *) fail 'unsupported executing OS' ;; esac
case "$(uname -m)" in x86_64|amd64) arch=amd64 ;; arm64|aarch64) arch=arm64 ;; *) fail 'unsupported executing architecture' ;; esac
ext=
if [ "$os" = windows ]; then [ "$arch" = amd64 ] || fail 'windows-arm64 artifacts unavailable'; ext=.exe; fi
native=$install_root/bin/$version/$os-$arch/agentic-lens-$os-$arch$ext
if [ "$resolve" -eq 1 ]; then
  printf '%s\n' "$version" "$root" "$install_root" "$native"
  exit 0
fi
[ -x "$native" ] || fail "Missing pinned native $native for $root; run explicit setup download.sh --project PROJECT"
magic=$(LC_ALL=C od -An -tx1 -N4 "$native" | tr -d ' \n')
case "$os:$magic" in linux:7f454c46|darwin:cffaedfe|darwin:cefaedfe|darwin:feedfacf|darwin:feedface|darwin:cafebabe|windows:4d5a*) ;; *) fail 'cache artifact is not a native executable; explicit repair required' ;; esac
exec "$native" "$@"

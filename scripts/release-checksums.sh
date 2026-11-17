#!/usr/bin/env bash
# The Make release inventory supplies all paths relative to the release directory.
set -euo pipefail
export LC_ALL=C

action=$1
cd -- "$2"
shift 2
(($#)) || { echo 'Empty release inventory' >&2; exit 1; }
for name in "$@"; do
    [[ -f "$name" && -r "$name" ]] || { echo "Missing or unreadable release artifact: $name" >&2; exit 1; }
done

manifest=SHA256SUMS
case "$action" in
    generate)
        [[ ! -d "$manifest" ]] || { echo "$manifest is a directory" >&2; exit 1; }
        stage=$(mktemp -d .checksums.XXXXXXXX)
        trap 'rm -rf -- "$stage"' EXIT
        manifest="$stage/SHA256SUMS"
        sha256sum "$@" > "$stage/unsorted"
        sort -k2 "$stage/unsorted" > "$manifest"
        ;;
    verify)
        [[ -f "$manifest" ]] || { echo "$manifest is missing or not a regular file" >&2; exit 1; }
        ;;
    *) echo "Unknown release checksum action: $action" >&2; exit 1 ;;
esac

# Admit each expected filename exactly once before passing any path to sha256sum -c.
awk -v inventory="$*" '
    BEGIN {
        count = split(inventory, names, " ")
        for (i = 1; i <= count; i++) required[names[i]] = 1
    }
    {
        hash = substr($0, 1, 64)
        marker = substr($0, 65, 2)
        name = substr($0, 67)
        if (length(hash) != 64 || hash !~ /^[[:xdigit:]]+$/ ||
            (marker != "  " && marker != " *") || !(name in required)) {
            print "Invalid or unexpected checksum record at line " NR > "/dev/stderr"
            invalid = 1
            exit 1
        }
        delete required[name]
    }
    END {
        if (invalid) exit 1
        for (name in required) {
            print "Missing checksum for " name > "/dev/stderr"
            exit 1
        }
    }
' "$manifest"
sha256sum -c "$manifest"

if [[ "$action" == generate ]]; then
    mv -f -- "$manifest" SHA256SUMS
fi

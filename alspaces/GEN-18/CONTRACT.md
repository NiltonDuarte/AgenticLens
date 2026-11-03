# GEN-18: Stable installation sources

User authorization: use the next available GEN ticket and change installation instructions from the live main branch to the latest stable tag.

Fresh installations resolve the latest published stable GitHub release once. All installation manifests, skills, scripts, and binaries use that same tag. Repairs retain the installed tag unless an upgrade is authorized. Older manifests with moving references must use the selected tag; unavailable files must be reported without fallback to main. Manual Linux, macOS, and Windows examples use a selected tag.

Scope: README, setup manual, setup manifest, operator recovery reference, and ticket artifacts. Documentation only; no runtime, test, release, or historical artifact changes. Current stable v2.0.0 lacks the newer bundle installer, so instructions must distinguish that workflow from tagged-manifest/manual installation.

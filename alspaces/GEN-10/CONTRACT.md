# GEN-10: Emphasize installation in the README

## Problem
The README quick start references a `.tar.gz` download that does not match the standalone platform binary assets and leaves the installation/configuration flow unclear.

## Scope
Update the README quick start to make installation the primary entry point. Use the official downloader and verified release asset naming, explain supported platforms and runtime prerequisites, and include configuration initialization and dashboard launch commands. Link to the setup manual for full integration details.

## Acceptance boundaries
- README installation commands match the repository downloader and setup manual.
- The README clearly states that released binaries do not require Go, Node.js, Docker, or an external database at runtime.
- Include a Windows installation path and the quick start remains concise.
- Documentation-only change; do not alter application behavior.

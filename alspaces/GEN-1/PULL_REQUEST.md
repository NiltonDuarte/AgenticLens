# GEN-1: Sandbox Network Policy for Go & TypeScript Toolchains

## Summary
Defines the required network hosts and ports for executing Go and TypeScript toolchains within isolated sandbox environments.

## Highlights
- Identified essential Go infrastructure: `proxy.golang.org:443`, `sum.golang.org:443`, and `storage.googleapis.com:443`.
- Identified essential TypeScript/Node infrastructure: `registry.npmjs.org:443`, `registry.yarnpkg.com:443`, `nodejs.org:443`, `objects.githubusercontent.com:443`, and `codeload.github.com:443`.
- Provided CLI invocations for `sbx policy allow network`.

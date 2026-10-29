# Changelog

## [2.0.0](https://github.com/NiltonDuarte/AgenticLens/compare/v1.0.1...v2.0.0) (2026-10-29)


### ⚠ BREAKING CHANGES

* **AL-56:** Runtime ingestion no longer accepts schema v1-v3 ledgers. Migrate each ledger to schema v4 and complete source-backed backfill before using it; incomplete migration state is rejected until resumed and completed. Invalid ticket workspaces are skipped with a visible warning while valid workspaces continue to load.

### Features

* add AgenticLens operator skill ([#53](https://github.com/NiltonDuarte/AgenticLens/issues/53)) ([c686f66](https://github.com/NiltonDuarte/AgenticLens/commit/c686f668dc341365fcea2029ef14219947b3a5ed))
* **AL-56:** add Codex session discovery, finalize schema v4, and backfill historical workspaces ([#38](https://github.com/NiltonDuarte/AgenticLens/issues/38)) ([fc52150](https://github.com/NiltonDuarte/AgenticLens/commit/fc5215074aba1acef020247aab0ce8c1ab1613e6))
* **AL-56:** enforce schema v4 migration and backfill boundary ([#43](https://github.com/NiltonDuarte/AgenticLens/issues/43)) ([62b5555](https://github.com/NiltonDuarte/AgenticLens/commit/62b55551dd032d0345dab6a62b2b50c481301632))
* **config:** add local multi-harness source selection (AL-59) ([#51](https://github.com/NiltonDuarte/AgenticLens/issues/51)) ([7f153b4](https://github.com/NiltonDuarte/AgenticLens/commit/7f153b4443914060bcf9de1078e6e13466879db0))


### Bug Fixes

* **backfill:** multi-root session backfill, explicit fallbacks, and hierarchy normalization (FIX-10) ([#44](https://github.com/NiltonDuarte/AgenticLens/issues/44)) ([4473198](https://github.com/NiltonDuarte/AgenticLens/commit/44731989552d63ed2c6937520f80177558caf34d))
* **backfill:** preserve Codex rollout session trees (FIX-8) ([#41](https://github.com/NiltonDuarte/AgenticLens/issues/41)) ([f2b2fb0](https://github.com/NiltonDuarte/AgenticLens/commit/f2b2fb0bccd93c32b289b8720da24b42d64ea084))
* **codex:** normalize subagent role names (FIX-9) ([#42](https://github.com/NiltonDuarte/AgenticLens/issues/42)) ([68e1feb](https://github.com/NiltonDuarte/AgenticLens/commit/68e1feb8463fb067f6a2ec1087fc2f7a34e92989))
* **config:** enforce critical context ceilings and consolidate source roots (FIX-11) ([#45](https://github.com/NiltonDuarte/AgenticLens/issues/45)) ([193de66](https://github.com/NiltonDuarte/AgenticLens/commit/193de6611a391119c3e7387143dd5f8cc9183c36))
* **deep-dive:** render ledger waterfall without hydration (FIX-12) ([#52](https://github.com/NiltonDuarte/AgenticLens/issues/52)) ([e825d76](https://github.com/NiltonDuarte/AgenticLens/commit/e825d7682033ec9e6a0c16bd31323fb4894413fd))

## [1.0.1](https://github.com/NiltonDuarte/AgenticLens/compare/v1.0.0...v1.0.1) (2026-10-22)


### Bug Fixes

* **web,scripts:** direct binary download and dynamic initial ticket selection (FIX-7) ([#32](https://github.com/NiltonDuarte/AgenticLens/issues/32)) ([3e97ab4](https://github.com/NiltonDuarte/AgenticLens/commit/3e97ab4290c006dbbb7e70b428a29cf0e1a995dd))

## [1.0.0](https://github.com/NiltonDuarte/AgenticLens/compare/v1.0.0...v1.0.0) (2026-10-22)


### Features

* initial release ([e450e40](https://github.com/NiltonDuarte/AgenticLens/commit/e450e4019f886f5b2e5ea83cf0c5d4737eaf27fb))


### Bug Fixes

* **ci:** automate binary cross-compilation and archive attachment in Release Please workflow ([#30](https://github.com/NiltonDuarte/AgenticLens/issues/30)) ([edaaa98](https://github.com/NiltonDuarte/AgenticLens/commit/edaaa98936df84247f6841fc8e02badd2175b6eb))

## [1.0.0](https://github.com/NiltonDuarte/AgenticLens/compare/agentic-lens-v1.0.0...agentic-lens-v1.0.0) (2026-10-22)


### Features

* initial release ([e450e40](https://github.com/NiltonDuarte/AgenticLens/commit/e450e4019f886f5b2e5ea83cf0c5d4737eaf27fb))

## Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

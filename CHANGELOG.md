# Changelog

## [3.0.0](https://github.com/NiltonDuarte/AgenticLens/compare/v2.0.0...v3.0.0) (2026-11-06)


### ⚠ BREAKING CHANGES

* **release:** AL-68 requires an explicit, valid telemetry.capture.state_dir in existing configuration; operational commands reject configurations that omit it. Use init --update-capture-config to insert the home-state default when the key is absent. Existing invalid values require correction, and relocating existing capture state requires an explicit migration.

### Features

* **backfill:** preserve incomplete historical snapshots (AL-55) ([#86](https://github.com/NiltonDuarte/AgenticLens/issues/86)) ([3857fa2](https://github.com/NiltonDuarte/AgenticLens/commit/3857fa22f49f825c82ced39c3f81f963b7447545))
* **capture:** automatically publish lifecycle session metrics (AL-68) ([#93](https://github.com/NiltonDuarte/AgenticLens/issues/93)) ([65e0eeb](https://github.com/NiltonDuarte/AgenticLens/commit/65e0eebe1acb1cb2f3194f449d672ddfdd4306ed))
* **comparison:** add multi-metric ticket scatter plot (AL-48) ([#61](https://github.com/NiltonDuarte/AgenticLens/issues/61)) ([d92fdb1](https://github.com/NiltonDuarte/AgenticLens/commit/d92fdb1edd69e61ecaa1d038a204afaa61024f9d))
* **comparisons:** search exclusions and repair capture validation (AL-60) ([#101](https://github.com/NiltonDuarte/AgenticLens/issues/101)) ([6244585](https://github.com/NiltonDuarte/AgenticLens/commit/6244585dc5cc41ef012cd3e1fb3b964ccd910a95))
* **dashboard:** show full local historical timestamps (AL-61) ([#104](https://github.com/NiltonDuarte/AgenticLens/issues/104)) ([8b1de5f](https://github.com/NiltonDuarte/AgenticLens/commit/8b1de5f761ebb92f78b2b462c1ff583724ecbdb5))
* **dashboard:** warn about sessions shared across tickets (AL-77) ([#114](https://github.com/NiltonDuarte/AgenticLens/issues/114)) ([44aa50c](https://github.com/NiltonDuarte/AgenticLens/commit/44aa50cc22ed8db65d4db4f6aded08aba233f577))
* **generate:** configure optional metric output (AL-66) ([#115](https://github.com/NiltonDuarte/AgenticLens/issues/115)) ([40b16de](https://github.com/NiltonDuarte/AgenticLens/commit/40b16deff869ce9311aeba7b73d2284071296b48))
* **ledger:** compact records and add selective format command (AL-76) ([#102](https://github.com/NiltonDuarte/AgenticLens/issues/102)) ([8d494d2](https://github.com/NiltonDuarte/AgenticLens/commit/8d494d2e7bd3705ee400b4bdecad256470f65f73))
* **live-inspector:** collapse failed tool call details by default (AL-46) ([#58](https://github.com/NiltonDuarte/AgenticLens/issues/58)) ([be27d39](https://github.com/NiltonDuarte/AgenticLens/commit/be27d399242c96199a41a98a1dc2a5f35e9048c2))
* **live-inspector:** reverse event ordering and add quick navigation (AL-49) ([#59](https://github.com/NiltonDuarte/AgenticLens/issues/59)) ([ba53cfc](https://github.com/NiltonDuarte/AgenticLens/commit/ba53cfc479b2d6a201ca35df56641a1af453d4e6))
* **live:** refresh active session traces automatically (AL-47) ([#76](https://github.com/NiltonDuarte/AgenticLens/issues/76)) ([5da845a](https://github.com/NiltonDuarte/AgenticLens/commit/5da845a7c244a6b2582102c0dff0d221844d9bfc))
* **observability:** ingest OpenCode skills and task links (AL-51) ([#62](https://github.com/NiltonDuarte/AgenticLens/issues/62)) ([8c9396a](https://github.com/NiltonDuarte/AgenticLens/commit/8c9396a33111bcc296891c0b86e8f7a12f1c51ca))
* **opencode:** support OpenCode v1 and v2 concurrently (AL-58) ([#50](https://github.com/NiltonDuarte/AgenticLens/issues/50)) ([8cc2c7c](https://github.com/NiltonDuarte/AgenticLens/commit/8cc2c7cc1ec0d394ead657bcc18c555b457a1dd5))
* **skills:** add fleet adoption analytics and comparison filters (AL-54) ([#79](https://github.com/NiltonDuarte/AgenticLens/issues/79)) ([b126823](https://github.com/NiltonDuarte/AgenticLens/commit/b1268234cde742705728ea33ee89d30581803c47))
* **telemetry:** distinguish semantic tool content errors (AL-43) ([#78](https://github.com/NiltonDuarte/AgenticLens/issues/78)) ([a77695c](https://github.com/NiltonDuarte/AgenticLens/commit/a77695ce7f0d2f51844da52d84a06f433367af7e))
* **traces:** show invocation badges and child session links (AL-52) ([#75](https://github.com/NiltonDuarte/AgenticLens/issues/75)) ([6bc1064](https://github.com/NiltonDuarte/AgenticLens/commit/6bc1064943ae22e1b6586a6d145543c6a9cd23f2))
* **web:** add skills and subagents autopsy panel (AL-53) ([#77](https://github.com/NiltonDuarte/AgenticLens/issues/77)) ([bdef2d0](https://github.com/NiltonDuarte/AgenticLens/commit/bdef2d0ef98e37fe259921bc30ba2daefbfda717))
* **web:** localize trace timestamps and show timezone (AL-44) ([#64](https://github.com/NiltonDuarte/AgenticLens/issues/64)) ([77579c5](https://github.com/NiltonDuarte/AgenticLens/commit/77579c57c6a85c5ffe271173c8e01c4c4ae75f18))


### Bug Fixes

* **build:** create embedded asset destination before copy (FIX-14) ([#67](https://github.com/NiltonDuarte/AgenticLens/issues/67)) ([2e4a70f](https://github.com/NiltonDuarte/AgenticLens/commit/2e4a70fd458a8d604161aab96f78c509ee81e76e))
* **capture:** allow isolated manual registry recovery (FIX-17) ([#111](https://github.com/NiltonDuarte/AgenticLens/issues/111)) ([04c248b](https://github.com/NiltonDuarte/AgenticLens/commit/04c248b7837065ea461767381de19ce90362b41c))
* **capture:** wait for native fixture writers before teardown (AL-60) ([#105](https://github.com/NiltonDuarte/AgenticLens/issues/105)) ([3fbecd2](https://github.com/NiltonDuarte/AgenticLens/commit/3fbecd2576270772462fe37235d90c5f33e2164c))
* **ci:** build embedded assets before go vet (FIX-15) ([#68](https://github.com/NiltonDuarte/AgenticLens/issues/68)) ([a274d38](https://github.com/NiltonDuarte/AgenticLens/commit/a274d38a0c763e42fe4c2194feca3e1df2afb121))
* **ci:** build embedded frontend assets before Go checks (FIX-13) ([#65](https://github.com/NiltonDuarte/AgenticLens/issues/65)) ([6da00a8](https://github.com/NiltonDuarte/AgenticLens/commit/6da00a8276d920aa7bda331b5e4f3c09f9589a58))
* **ci:** restore release candidate validation (FIX-16) ([#107](https://github.com/NiltonDuarte/AgenticLens/issues/107)) ([33ddada](https://github.com/NiltonDuarte/AgenticLens/commit/33ddada05c64cb8fa41701fd9fce3376ff83ff10))
* **cost:** distinguish unavailable telemetry costs (AL-57) ([#89](https://github.com/NiltonDuarte/AgenticLens/issues/89)) ([47af923](https://github.com/NiltonDuarte/AgenticLens/commit/47af923f5763dc937372ada776c1043c2dd3e7a6))
* **domain:** unify tool execution error classification (AL-42) ([#57](https://github.com/NiltonDuarte/AgenticLens/issues/57)) ([ca14931](https://github.com/NiltonDuarte/AgenticLens/commit/ca149312ba7033ce67f6c9fd2598d52756f4740a))
* **duration:** account for active turns and report long provisional activity (AL-64) ([#121](https://github.com/NiltonDuarte/AgenticLens/issues/121)) ([57a3503](https://github.com/NiltonDuarte/AgenticLens/commit/57a3503e3024f3993ce9391fee604aad0bcc44d5))
* **live:** subscribe before exposing the SSE connection (AL-47) ([#82](https://github.com/NiltonDuarte/AgenticLens/issues/82)) ([c3ce116](https://github.com/NiltonDuarte/AgenticLens/commit/c3ce116611638ea1bee47dc3140d8af992814a44))
* **presentation:** contain and middle-truncate debug identifiers (AL-45) ([#56](https://github.com/NiltonDuarte/AgenticLens/issues/56)) ([bbc535c](https://github.com/NiltonDuarte/AgenticLens/commit/bbc535ca27ca01bd7feac272ae512f954135433f))
* **REV-5:** preserve telemetry recovery and harden audited boundaries ([#92](https://github.com/NiltonDuarte/AgenticLens/issues/92)) ([7bc670a](https://github.com/NiltonDuarte/AgenticLens/commit/7bc670a7f5c9e93a220f3596f3d0e05cb6007332))
* **telemetry:** persist lean skills and delegations (AL-62) ([#109](https://github.com/NiltonDuarte/AgenticLens/issues/109)) ([86bea59](https://github.com/NiltonDuarte/AgenticLens/commit/86bea597afe7208cb1836f7cc5747083f7b95f47))


### Miscellaneous Chores

* **release:** declare AL-68 configuration compatibility break ([#99](https://github.com/NiltonDuarte/AgenticLens/issues/99)) ([733e367](https://github.com/NiltonDuarte/AgenticLens/commit/733e367ee12b9828cd340759eab855f4d809279c))

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

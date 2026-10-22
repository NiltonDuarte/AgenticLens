# Changelog

## [0.2.0](https://github.com/NiltonDuarte/AgenticLens/compare/agentic-lens-v0.1.0...agentic-lens-v0.2.0) (2026-10-22)


### Features

* **adapters:** implement Go telemetry backfill engine and schema migration command (AL-12) ([e867b48](https://github.com/NiltonDuarte/AgenticLens/commit/e867b487b9e22c610bda8eef1877be0fa505f3e4))
* **adapters:** implement native Go telemetry extractor and session generation (AL-11) ([1af3eac](https://github.com/NiltonDuarte/AgenticLens/commit/1af3eac953f4a5fda94863e46a71641fd835388b))
* **aggregations:** dynamic role and model aggregations and presentation (AL-24) ([2717de3](https://github.com/NiltonDuarte/AgenticLens/commit/2717de36fe62b213cd335998770d926db3291a1a))
* **api,web:** dynamic config rest api endpoint and dashboard branding (AL-40) ([#16](https://github.com/NiltonDuarte/AgenticLens/issues/16)) ([f23e934](https://github.com/NiltonDuarte/AgenticLens/commit/f23e9345a158dee7ed1708a0b52e720565152916))
* **api:** anomaly API endpoints and fleet health aggregation (AL-8) ([e02f644](https://github.com/NiltonDuarte/AgenticLens/commit/e02f64423ac4f671e26a71f91ece407cf427f152))
* **build:** integrate single binary build, verify parity, and deprecate python app (AL-5) ([3599e78](https://github.com/NiltonDuarte/AgenticLens/commit/3599e78634adf7257234510b2319259359f359e5))
* **ci:** implement semantic versioning and automated ci for pull requests and main merges (AL-13) ([cc0670c](https://github.com/NiltonDuarte/AgenticLens/commit/cc0670c613946af8a81799dddf718144599c2280))
* **cli:** conventional commit ticket type classification in telemetry engine (AL-29) ([b6626cc](https://github.com/NiltonDuarte/AgenticLens/commit/b6626cc5f0c195c4bea2734212744626328382c9))
* **config:** migrate configuration to yaml, support PO tickets, and bridge definitions (AL-21) ([f8ffbc9](https://github.com/NiltonDuarte/AgenticLens/commit/f8ffbc9790ba0ca889d769c9d2fa39057ec10c26))
* **domain,adapters,cli:** telemetry event pruning, database message stripping, and schema v3 migration (AL-32) ([#3](https://github.com/NiltonDuarte/AgenticLens/issues/3)) ([1b35fe3](https://github.com/NiltonDuarte/AgenticLens/commit/1b35fe3f63a2fe330f4b7f142645d8a83dcfc610))
* **domain:** domain anomaly detection and rule engine (AL-7) ([3dbc0ab](https://github.com/NiltonDuarte/AgenticLens/commit/3dbc0abf509e439b2319a9c9649024d95be51acd))
* **domain:** implement Go aggregations, workspace scanner, CLI engine, and Makefile ([db57d34](https://github.com/NiltonDuarte/AgenticLens/commit/db57d34b265fa18504ac151bc8d50392992bc3e4))
* **events:** chronological event extraction, schema evolution, and workspace backfill (AL-6) ([d0191fd](https://github.com/NiltonDuarte/AgenticLens/commit/d0191fd66ea3cbb6edb64cb432d573d8d3e01855))
* **filtering:** multi-dimensional filtering by ticket type and workspace prefix (AL-28) ([0451d46](https://github.com/NiltonDuarte/AgenticLens/commit/0451d46d230a27edbf5a693e3ef0978664bfca2d))
* **fleet:** completion date filtering and autopsy date sorting (AL-26) ([ee4414c](https://github.com/NiltonDuarte/AgenticLens/commit/ee4414caf54713db187da1b8782aed3f01cc72fa))
* **live,web:** build typescript live observability dashboard and multiroot session inspector (AL-20) ([#10](https://github.com/NiltonDuarte/AgenticLens/issues/10)) ([f32359e](https://github.com/NiltonDuarte/AgenticLens/commit/f32359e6ec4ef38809fd8f4adc40adec55fc53eb))
* **live,web:** dynamic local database event hydration and dual-plane waterfall (AL-33) ([#7](https://github.com/NiltonDuarte/AgenticLens/issues/7)) ([c0e475e](https://github.com/NiltonDuarte/AgenticLens/commit/c0e475eea889a6ff0f27c0f5bad39a06a13a8440))
* **live:** implement live anomaly detection engine and real-time session health rules (AL-19) ([#6](https://github.com/NiltonDuarte/AgenticLens/issues/6)) ([06d37bf](https://github.com/NiltonDuarte/AgenticLens/commit/06d37bff511ef9878fb95d27f417f3590a41d7c2))
* **live:** implement non-locking database hook, session tree engine, and live telemetry (AL-17) ([#2](https://github.com/NiltonDuarte/AgenticLens/issues/2)) ([2c4340e](https://github.com/NiltonDuarte/AgenticLens/commit/2c4340e8022b7414bafd7902ce160017238c7776))
* **migration:** implement universal multiversion schema migration engine for telemetry ledgers (AL-15) ([3b9a209](https://github.com/NiltonDuarte/AgenticLens/commit/3b9a2098a7a036e48b112489704bbef43c22d15f))
* **ordering:** implement natural ticket identifier ordering and sortable tables (AL-23) ([89f6c77](https://github.com/NiltonDuarte/AgenticLens/commit/89f6c772fcf0f59d390fc73940a5c8f622130b44))
* **plan:** add product specification and execution tiers for go rewrite (PO-1) ([c4c4dab](https://github.com/NiltonDuarte/AgenticLens/commit/c4c4dab37a934bf516c84bf72d0f413080199311))
* **plan:** add product specification and execution tiers for native go extraction, backfill engine, and git metrics (PO-3) ([6e36909](https://github.com/NiltonDuarte/AgenticLens/commit/6e36909d5bf2eebee74a17cfbed6e90ba0912ccf))
* **plan:** add product specification and execution tiers for schema v2 and anomaly engine (PO-2) ([ec5ff80](https://github.com/NiltonDuarte/AgenticLens/commit/ec5ff8045479c173925e37d20f22048b032aaf87))
* **plan:** add specifications and tickets for ci/cd, release workflows, schema migration, review skill refactor, live observability, yaml config, and artifact rendering (PO-4) ([64e3c1a](https://github.com/NiltonDuarte/AgenticLens/commit/64e3c1a851ad4a82427e57086c4b43376c024df2))
* **plan:** incorporate hybrid path classification and config overrides into PO-3 specs and tickets ([2d4f0ab](https://github.com/NiltonDuarte/AgenticLens/commit/2d4f0abc28c98a610ccbb02eb5bf4b388ea0d361))
* **presentation:** contextual hover information icons and responsive table containment (AL-25) ([aa935bf](https://github.com/NiltonDuarte/AgenticLens/commit/aa935bf5bfc82486a3fca28402452b152b2d2f6a))
* **presentation:** visual trace waterfall and anomaly inspector (AL-9) ([e867ed0](https://github.com/NiltonDuarte/AgenticLens/commit/e867ed0374fdc80fe0d2376927b4c83b74614475))
* **progression:** custom window, 1-day range, and hourly granularity (AL-27) ([9ab08b0](https://github.com/NiltonDuarte/AgenticLens/commit/9ab08b02564a84e35ec0c780b6aec9cbd16ca3c6))
* **progression:** eliminate cumulative granularity and implement bidirectional code churn visualization (AL-34) ([#5](https://github.com/NiltonDuarte/AgenticLens/issues/5)) ([7b6f3ef](https://github.com/NiltonDuarte/AgenticLens/commit/7b6f3eff1a47251b8642730be4767c0d7609f046))
* **server:** implement Go HTTP server, embedded static FS, and telemetry REST API (AL-2) ([2b74a81](https://github.com/NiltonDuarte/AgenticLens/commit/2b74a814891a6c65f78dfcd46b044c4575fc47f9))
* **skills,cli,web:** cross-project setup skill, artifact downloader, and dynamic ticket artifacts (AL-36) ([#14](https://github.com/NiltonDuarte/AgenticLens/issues/14)) ([7a139c7](https://github.com/NiltonDuarte/AgenticLens/commit/7a139c77e7e37265a45977cd51ed47259c1de827))
* **web:** build ticket deep dive, progression charts, comparisons, and landing portal (AL-4) ([9e85f20](https://github.com/NiltonDuarte/AgenticLens/commit/9e85f20ec67e911c286df49fcc0f14d1289ae670))
* **web:** build TypeScript SPA shell, Fleet Overview dashboard, and single-binary embedding (AL-3) ([e703c58](https://github.com/NiltonDuarte/AgenticLens/commit/e703c58bba706f6040258671372c1a631aff2c56))
* **web:** render rich markdown for workspace summary and contract artifacts (AL-22) ([7bcbc18](https://github.com/NiltonDuarte/AgenticLens/commit/7bcbc18809f42575adab94d8dbf778bf099f16a9))


### Bug Fixes

* centralize anomaly detection and add live debug identifiers (FIX-3) ([#24](https://github.com/NiltonDuarte/AgenticLens/issues/24)) ([edf719d](https://github.com/NiltonDuarte/AgenticLens/commit/edf719d1aef707922b9945db06a56fbd38d84c66))
* **chore:** fix ticket_type of existing PO and GEN tickets ([9daf765](https://github.com/NiltonDuarte/AgenticLens/commit/9daf76505016295c1de44550251abe2ce6f3cddd))
* **cli:** enforce mandatory ticket title and id with diagnostic warnings (FIX-1) ([#18](https://github.com/NiltonDuarte/AgenticLens/issues/18)) ([53b8ec0](https://github.com/NiltonDuarte/AgenticLens/commit/53b8ec0df16e8a4e2da0d3af8bb603ad6cb151d6))
* **cli:** fail telemetry generation on unresolved root session (AL-30) ([bc8e257](https://github.com/NiltonDuarte/AgenticLens/commit/bc8e257226dd285aec08ea32b9ed8136031da16a))
* resolve false tool errors, retry streak misattribution, collapse waterfall by default, and add debug copy action ([#21](https://github.com/NiltonDuarte/AgenticLens/issues/21)) ([22dd699](https://github.com/NiltonDuarte/AgenticLens/commit/22dd699fa8ffade68b62e933c7a932c7481633a1))
* **server,tests:** self contained test suites and server config autoload (FIX-4) ([#23](https://github.com/NiltonDuarte/AgenticLens/issues/23)) ([48db490](https://github.com/NiltonDuarte/AgenticLens/commit/48db4908694ceae5715d69296f9c3bb41d4ef2fd))
* **skills:** mandate agent-generated literal nonces for telemetry (FIX-5) ([#27](https://github.com/NiltonDuarte/AgenticLens/issues/27)) ([2ec257c](https://github.com/NiltonDuarte/AgenticLens/commit/2ec257c40afe19107be40ca45e91bea1a3658004))
* **web,api:** eliminate isMock branching and enforce mandatory workspaces dir configuration (FIX-2) ([#20](https://github.com/NiltonDuarte/AgenticLens/issues/20)) ([97705eb](https://github.com/NiltonDuarte/AgenticLens/commit/97705ebd30ae7e2950e8ef689552cdba0de5937b))
* **web,api:** remove mockData from source and establish testdata workspaces harness (FIX-2) ([#19](https://github.com/NiltonDuarte/AgenticLens/issues/19)) ([66b6f3e](https://github.com/NiltonDuarte/AgenticLens/commit/66b6f3e20d77f4f8a2911986052134ade99ba21a))


### Performance Improvements

* **tests:** eliminate subprocess churn, artificial sleeps, and test execution bottlenecks (AL-41) ([#15](https://github.com/NiltonDuarte/AgenticLens/issues/15)) ([7474a77](https://github.com/NiltonDuarte/AgenticLens/commit/7474a77c729ee8a62195821ee0c35c4a72e8538f))

## Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

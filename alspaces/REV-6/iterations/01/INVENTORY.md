# Test inventory at the measured revision

Source: a84adbe5be46c2ef52ede0987a417562715d38e1. Go AST discovery found 782 top-level Test-prefixed declarations excluding TestMain across 121 files containing such declarations. Helper-only test files are additional setup dependencies. This includes platform-tagged declarations and subprocess entry helpers; it is not a count of executed assertions or effective test cases. TypeScript AST discovery found 686 literal-title test/it declarations across 36 files; runtime full-suite execution expanded to 824 cases. No count is a removal target.

The scope is a capture-first review, not a claim that every body in this inventory was audited. Boundary classes below are conservative package/file groups; mixed packages stay in the full later suite. Named independently inspected oracles are in TEST_REVIEW.md. Runtime parent/subtest events overlap, and build-tag exclusions differ from runtime skips. Linux capture/CLI cohorts reported no runtime skips.

## Go file inventory

| File | Test declarations | TestMain | Build constraints |
| --- | ---: | --- | --- |
| backend/cmd/agentic-lens/capture_coordination_acceptance_test.go | 1 |  |  |
| backend/cmd/agentic-lens/capture_corrections_test.go | 24 |  |  |
| backend/cmd/agentic-lens/capture_failure_feedback_test.go | 1 |  |  |
| backend/cmd/agentic-lens/capture_harness_fixture_test.go | 1 |  |  |
| backend/cmd/agentic-lens/capture_history_authority_test.go | 2 |  |  |
| backend/cmd/agentic-lens/capture_history_test.go | 8 |  |  |
| backend/cmd/agentic-lens/capture_history_v4_test.go | 1 |  |  |
| backend/cmd/agentic-lens/capture_install_deadline_unix_test.go | 2 |  | //go:build !windows |
| backend/cmd/agentic-lens/capture_isolation_test.go | 7 |  |  |
| backend/cmd/agentic-lens/capture_iteration03_corrections_test.go | 4 |  |  |
| backend/cmd/agentic-lens/capture_runtime_integration_test.go | 5 |  |  |
| backend/cmd/agentic-lens/capture_shared_state_test.go | 3 |  |  |
| backend/cmd/agentic-lens/capture_test.go | 12 |  |  |
| backend/cmd/agentic-lens/capture_worker_completion_failure_test.go | 1 |  |  |
| backend/cmd/agentic-lens/capture_worker_switch_linux_test.go | 1 |  | //go:build linux |
| backend/cmd/agentic-lens/capture_worker_unix_test.go | 2 |  | //go:build !windows |
| backend/cmd/agentic-lens/config_admission_integration_test.go | 5 |  |  |
| backend/cmd/agentic-lens/config_internal_test.go | 3 |  |  |
| backend/cmd/agentic-lens/format_test.go | 2 |  |  |
| backend/cmd/agentic-lens/lean_cache_recovery_test.go | 1 |  |  |
| backend/cmd/agentic-lens/lean_migration_test.go | 2 |  |  |
| backend/cmd/agentic-lens/main_codex_test.go | 3 |  |  |
| backend/cmd/agentic-lens/main_codex_user_test.go | 1 |  |  |
| backend/cmd/agentic-lens/main_refresh_test.go | 1 |  |  |
| backend/cmd/agentic-lens/main_test.go | 38 | yes |  |
| backend/cmd/agentic-lens/merge_internal_test.go | 2 |  |  |
| backend/cmd/agentic-lens/owned_refresh_regression_test.go | 4 |  |  |
| backend/internal/adapters/backfill/codex_user_test.go | 1 |  |  |
| backend/internal/adapters/backfill/engine_test.go | 30 |  |  |
| backend/internal/adapters/backfill/migration_runner_test.go | 22 |  |  |
| backend/internal/adapters/backfill/publication_ownership_test.go | 2 |  |  |
| backend/internal/adapters/capture/constructor_context_test.go | 2 |  |  |
| backend/internal/adapters/capture/containment_windows_test.go | 1 |  |  |
| backend/internal/adapters/capture/coordination_acceptance_test.go | 2 |  |  |
| backend/internal/adapters/capture/diagnostic_index_test.go | 7 |  |  |
| backend/internal/adapters/capture/diagnostics_test.go | 9 |  |  |
| backend/internal/adapters/capture/git_hook_test.go | 5 |  |  |
| backend/internal/adapters/capture/git_hook_watchdog_windows_test.go | 0 | yes | //go:build windows |
| backend/internal/adapters/capture/git_hook_windows_test.go | 1 |  | //go:build windows |
| backend/internal/adapters/capture/history_test.go | 2 |  |  |
| backend/internal/adapters/capture/hooks_test.go | 3 |  |  |
| backend/internal/adapters/capture/install_cancellation_unix_test.go | 2 |  | //go:build !windows |
| backend/internal/adapters/capture/install_deadline_unix_test.go | 1 |  | //go:build !windows |
| backend/internal/adapters/capture/install_native_validation_test.go | 1 |  |  |
| backend/internal/adapters/capture/install_process_windows_test.go | 1 |  | //go:build windows |
| backend/internal/adapters/capture/install_recovery_cancellation_test.go | 1 |  |  |
| backend/internal/adapters/capture/install_timeout_test.go | 1 |  |  |
| backend/internal/adapters/capture/install_transaction_test.go | 5 |  |  |
| backend/internal/adapters/capture/installation_unix_test.go | 7 | yes | //go:build !windows |
| backend/internal/adapters/capture/lean_snapshot_test.go | 1 |  |  |
| backend/internal/adapters/capture/metadata_test.go | 2 |  |  |
| backend/internal/adapters/capture/namespace_test.go | 4 |  |  |
| backend/internal/adapters/capture/opencode_test.go | 4 |  |  |
| backend/internal/adapters/capture/project_test.go | 5 |  |  |
| backend/internal/adapters/capture/publication_control_test.go | 2 |  |  |
| backend/internal/adapters/capture/reconcile_test.go | 15 |  |  |
| backend/internal/adapters/capture/relocate_install_unix_test.go | 1 |  | //go:build !windows |
| backend/internal/adapters/capture/relocate_publication_test.go | 1 |  |  |
| backend/internal/adapters/capture/relocate_test.go | 13 |  |  |
| backend/internal/adapters/capture/shared_state_test.go | 9 |  |  |
| backend/internal/adapters/capture/store_test.go | 4 |  |  |
| backend/internal/adapters/capture/worker_successor_unix_test.go | 1 |  | //go:build !windows |
| backend/internal/adapters/codex/discovery_test.go | 6 |  |  |
| backend/internal/adapters/codex/invocation_ownership_regression_test.go | 1 |  |  |
| backend/internal/adapters/codex/retention_test.go | 3 |  |  |
| backend/internal/adapters/codex/rollout_test.go | 25 |  |  |
| backend/internal/adapters/codex/store_provenance_test.go | 1 |  |  |
| backend/internal/adapters/config/capture_admission_test.go | 1 |  |  |
| backend/internal/adapters/config/capture_test.go | 4 |  |  |
| backend/internal/adapters/config/capture_update_test.go | 6 |  |  |
| backend/internal/adapters/config/entrypoint_test.go | 5 |  |  |
| backend/internal/adapters/config/entrypoint_unix_test.go | 1 |  | //go:build !windows |
| backend/internal/adapters/config/loader_test.go | 28 |  |  |
| backend/internal/adapters/config/source_test.go | 1 |  |  |
| backend/internal/adapters/config/writer_test.go | 8 |  |  |
| backend/internal/adapters/formatters_test.go | 6 |  |  |
| backend/internal/adapters/git/analyzer_test.go | 13 |  |  |
| backend/internal/adapters/git/classifier_test.go | 7 |  |  |
| backend/internal/adapters/git/committed_ledger_test.go | 2 |  |  |
| backend/internal/adapters/git/inference_test.go | 8 |  |  |
| backend/internal/adapters/git/user_test.go | 1 |  |  |
| backend/internal/adapters/http/live_handlers_test.go | 12 |  |  |
| backend/internal/adapters/http/server_test.go | 30 |  |  |
| backend/internal/adapters/ledger_format_test.go | 9 |  |  |
| backend/internal/adapters/live/lean_hydration_test.go | 1 |  |  |
| backend/internal/adapters/live/parity_test.go | 1 |  |  |
| backend/internal/adapters/live/service_snapshot_test.go | 1 |  |  |
| backend/internal/adapters/live/sqlite_test.go | 34 |  |  |
| backend/internal/adapters/metrics_publication_test.go | 1 |  |  |
| backend/internal/adapters/opencode/capture_activity_test.go | 1 |  |  |
| backend/internal/adapters/opencode/child_proof_test.go | 1 |  |  |
| backend/internal/adapters/opencode/extractor_test.go | 3 |  |  |
| backend/internal/adapters/opencode/schema_family_detection_test.go | 2 |  |  |
| backend/internal/adapters/opencode/schema_v1_session_extraction_test.go | 17 |  |  |
| backend/internal/adapters/opencode/schema_v1_session_resolution_test.go | 2 |  |  |
| backend/internal/adapters/opencode/schema_v2_session_extraction_test.go | 9 |  |  |
| backend/internal/adapters/opencode/schema_v2_session_resolution_test.go | 1 |  |  |
| backend/internal/adapters/parity_test.go | 3 |  |  |
| backend/internal/adapters/publication_set_test.go | 8 |  |  |
| backend/internal/adapters/repository_decode_test.go | 2 |  |  |
| backend/internal/adapters/repository_test.go | 20 |  |  |
| backend/internal/adapters/schemas_test.go | 29 |  |  |
| backend/internal/adapters/web/embed_test.go | 1 |  |  |
| backend/internal/domain/aggregations_test.go | 53 |  |  |
| backend/internal/domain/anomalies_test.go | 18 |  |  |
| backend/internal/domain/capture_diagnostics_test.go | 5 |  |  |
| backend/internal/domain/capture_test.go | 1 |  |  |
| backend/internal/domain/config_test.go | 14 |  |  |
| backend/internal/domain/constants_test.go | 3 |  |  |
| backend/internal/domain/enums_test.go | 8 |  |  |
| backend/internal/domain/events_test.go | 10 |  |  |
| backend/internal/domain/live_anomaly_test.go | 11 |  |  |
| backend/internal/domain/live_test.go | 8 |  |  |
| backend/internal/domain/migration/engine_test.go | 14 |  |  |
| backend/internal/domain/migration/invocation_ownership_regression_test.go | 1 |  |  |
| backend/internal/domain/migration/lean_v6_test.go | 2 |  |  |
| backend/internal/domain/migration_test.go | 1 |  |  |
| backend/internal/domain/session_facts_test.go | 3 |  |  |
| backend/internal/domain/skills_tasks_test.go | 1 |  |  |
| backend/internal/domain/tool_execution_test.go | 2 |  |  |
| backend/internal/domain/version_test.go | 7 |  |  |

## Frontend file inventory

Literal-title AST declarations include parameterized declarations once; dynamically expanded cases are supplied by the Vitest reporter. Duplicate titles are distinguished by file and declaration line.

| File | Declaration lines |
| --- | --- |
| web/src/api/client.test.ts | 37, 41, 56, 62, 67, 85, 93, 98, 103, 119, 126, 131, 139, 153, 168, 178, 184, 188, 196, 205, 212, 304, 308, 317, 327, 347, 368, 380, 389, 397, 405, 411, 433, 446, 456, 467, 471, 479, 499, 509, 519, 526, 564, 596, 600, 604, 612, 620, 631, 646, 650, 658, 666, 670, 681, 703, 707, 715, 723, 727, 738, 742, 746 |
| web/src/api/liveApi.test.ts | 67, 74, 84, 98, 104, 112, 141, 151, 193, 228, 261, 289, 309, 366, 400, 435, 460, 493, 519, 531, 539, 550, 563, 575, 589, 601, 612, 623, 630 |
| web/src/components/AnomalyBadge.test.tsx | 6, 22, 32, 39 |
| web/src/components/AnomalyDrawer.test.tsx | 36, 48, 65, 87, 106, 121, 168, 197, 214, 231, 246, 267, 282, 308, 347, 365, 455, 480, 503, 547 |
| web/src/components/ArtifactViewer.test.tsx | 56, 88, 116, 141, 163, 194, 220, 265, 296, 327, 358, 428, 477, 532, 602, 630, 677, 714, 778, 811 |
| web/src/components/Components.test.tsx | 17, 30, 49, 67, 84, 113, 125, 140, 157, 172, 190, 207, 227, 245, 250, 257, 271, 294, 314, 326, 348, 396, 413, 442, 481, 511, 532, 585, 601, 643, 726, 776, 798, 817, 844, 859 |
| web/src/components/FailedToolAlertCard.test.tsx | 19, 33, 64, 70, 78, 87, 116, 142, 168, 182, 198, 249, 261 |
| web/src/components/FleetAnomalyCard.test.tsx | 39, 50, 55, 64, 73, 100, 111, 119, 141 |
| web/src/components/LiveAnomalyBanner.test.tsx | 34, 43, 54, 67, 82, 106, 123, 145 |
| web/src/components/LiveSessionTree.test.tsx | 8, 31, 49, 129, 155, 182, 206, 276, 351, 400, 432 |
| web/src/components/LiveTraceInspector.test.tsx | 41, 83, 97, 112, 142, 158, 187, 207, 228, 250, 266, 341, 366, 446, 459, 481, 506, 521, 569, 606, 667, 737, 839, 886, 934, 992 |
| web/src/components/MultiSelectDropdown.test.tsx | 12, 30, 60, 79, 98, 117, 136, 155, 175, 191, 212, 231, 255, 288, 306, 328, 349, 370, 399, 426, 446, 473, 494, 517, 538, 557 |
| web/src/components/Navbar.test.tsx | 8, 18, 36, 50, 63, 100, 124, 136, 156, 176, 194, 209, 229, 241, 264, 296, 330 |
| web/src/components/TicketScatterPlot.test.tsx | 8, 20, 34, 68, 110, 119 |
| web/src/components/TicketTable.test.tsx | 79, 96, 113, 122, 133, 145, 158, 185, 209, 231, 253, 272, 282 |
| web/src/components/TraceWaterfall.test.tsx | 15, 36, 51, 58, 142, 157, 172, 244, 261, 279, 294, 313, 329, 343, 354, 364, 374, 380, 401, 426, 448, 464, 475, 489, 501, 523, 558, 579, 613, 638, 674, 701, 717, 753, 774, 970, 991, 1050, 1079, 1126, 1153, 1209, 1236, 1262, 1301, 1324, 1427, 1496, 1538, 1576, 1610, 1637, 1673, 1701, 1742, 1770, 1809, 1830, 1861, 1901, 1950, 2004, 2065, 2097, 2138, 2220, 2261, 2285, 2343, 2416, 2482, 2548, 2598, 2678, 2691, 2704, 2713 |
| web/src/components/TrendChart.test.tsx | 7, 23, 37, 73, 142 |
| web/src/components/WorkspaceLoadWarning.test.tsx | 6, 18, 30 |
| web/src/domain/aggregations.test.ts | 9, 17 |
| web/src/domain/anomalies.test.ts | 17, 27, 35, 54, 61, 109, 169, 174, 182, 195, 208, 224, 233, 241, 258, 266, 274, 289, 303, 312, 319, 329, 342, 351, 406, 411, 420, 429, 439 |
| web/src/domain/events.test.ts | 11, 20, 26, 37, 56 |
| web/src/domain/formatters.test.ts | 8, 11, 14, 19, 26, 55, 62, 65, 72, 80, 86, 94, 104, 114, 120, 128, 136, 146, 154, 160, 168, 178, 185, 192, 196, 202, 208, 211, 298, 393, 403, 437, 456, 465, 470, 475, 479, 483, 487, 498, 511, 520, 530, 536, 567, 604, 625, 651, 673, 689, 694, 698, 702, 713, 718, 732, 738, 753, 758, 766, 782, 804, 809, 814, 824, 830, 838, 843, 849, 855, 862, 870, 875 |
| web/src/domain/invocation_rows.test.ts | 7 |
| web/src/domain/liveMessageOrdering.test.ts | 6, 20 |
| web/src/domain/progression.test.ts | 8, 25, 37, 44, 58, 64, 79, 87, 98, 106, 124, 127, 137, 148, 157, 284, 396, 406, 425, 535, 562, 710, 799, 902, 989, 1097, 1209, 1298, 1408 |
| web/src/domain/ticket_scatter.test.ts | 6, 13, 40, 47, 59, 73, 82 |
| web/src/domain/tools.test.ts | 5, 16 |
| web/src/test/sync-assets.test.ts | 10 |
| web/src/utils/debugPrompt.test.ts | 15, 28, 38, 47, 66, 79, 94, 103, 115 |
| web/src/utils/theme.test.ts | 15, 36, 47, 55, 60, 71, 76 |
| web/src/views/FleetOverview.test.tsx | 12, 23, 49, 62, 76, 106, 129, 154, 174, 193, 213, 236, 265, 292 |
| web/src/views/LandingPage.test.tsx | 13, 27, 60, 93, 128, 160, 217, 298, 333, 364, 384, 404, 434, 465, 540, 597, 612, 673, 697, 726 |
| web/src/views/LiveObservability.test.tsx | 66, 92, 127, 155, 185, 213, 256, 263, 284, 309, 319, 351, 383, 427, 481, 510, 525, 557, 578, 611, 630, 644, 670 |
| web/src/views/MetricProgression.test.tsx | 10, 28, 38, 99, 139, 170, 198, 220, 235, 274, 321, 331, 346, 378, 410, 434 |
| web/src/views/TicketComparisons.test.tsx | 13, 79, 116, 149, 162, 188, 224, 276, 300, 311, 323, 346, 390, 473, 492, 531, 563 |
| web/src/views/TicketDeepDive.test.tsx | 58, 76, 89, 113, 147, 178, 229, 271, 310, 334, 360, 389, 405, 438, 494, 539, 576, 614, 630, 651, 678, 694, 718, 746, 755, 769, 787, 885, 970, 1035, 1146, 1174, 1239, 1260, 1302, 1339, 1382, 1433, 1478, 1521, 1563, 1587, 1617 |

## Default routing

All Go domain/migration and frontend domain files are the validated initial pure gates. Capture adapter files mix wire-only helpers with real filesystem, Git, SQLite, subprocess and installation behavior. CLI files invoke TestMain even under test-name selection. Keep these mixed packages and all remaining backend packages in the later full race suite until a future split reconciles actual assertions and helper reachability. Keep frontend component/hooks/DOM tests and all Python checks in the later suites; a DOM test does not claim real browser, network or hosted-release coverage. Windows-tagged native tests require Windows execution.

# fix(AL-69): preserve accurate telemetry and evidence

Configuration slice implemented; remaining AL-69 behavior and publication are pending.

AL-69 addresses false tool failures caused by inspected text, ignored anomaly settings, hidden partial-data diagnostics, overwritten summary evidence, divergent live/offline decoding, and incorrect anomaly matching or streak counts.

The proposed behavior uses only structured execution evidence, carries effective thresholds through consumers, retains useful partial traces with diagnostics, and preserves handwritten summary sections. The user approved excluding all textual process inference; raw text stays inspectable and missing structured evidence stays unknown. Failure-rate policy requires explicit file values without runtime fallbacks. User-selected writer defaults are 5 minimum calls and a 0.70 critical ratio; this repository retains warning ratio 0.35. A matching typed live alert remains proposed.

Configuration/domain/CLI tests, full `make` and standalone Linux binary execution passed on the working candidate. New tests cover writer/file/loader round trips, explicit values and precision, malformed/missing policy, legacy-section precedence and CLI init. Windows fixture execution is unavailable on Linux. Anomaly evaluators do not yet consume the new minimum/critical fields; broader behavior, independent reviews and final integration remain pending.

See [contract](CONTRACT.md) and [current evidence and limits](SUMMARY.md). Rewrite this description around the verified implementation before publication.

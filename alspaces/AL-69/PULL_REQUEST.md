# fix(AL-69): preserve accurate telemetry and evidence

AgenticLens could count inspected error text as tool failures, ignore configured anomaly policy, conceal incomplete scans/child lookup, and overwrite or misattribute evidence.

Use structured execution facts only, explicit configured minimum/warning/critical rates, and shared version-specific live/offline decoding. Preserve unavailable historical facts and authoritative fleet projections; expose partial-data diagnostics without hiding usable traces. Replace only owned generated summary content, preserve full transient native errors, and match exact anomaly identities while keeping rate totals separate from streaks. Also fix the pre-existing capture schema-read race exposed by concurrent-open verification without weakening database rejection or ownership checks.

Configuration writer defaults are 5 minimum calls and 70% critical; this repository retains 35% warning. Runtime requires the file values. Saved schema/privacy and AL-62/AL-66 ownership/presence guarantees remain unchanged.

Validation: actual CLI/SQLite/HTTP/dashboard regressions, before/after preservation probes, four dropped-policy mutation controls and constructor controls passed. Initial reviews found two preservation gaps and one weak oracle; all are corrected. Final focused reviews used Luna only as requested and passed. Integrated `make` and standalone-binary checks are the remaining delivery gate; Windows runtime was not exercised on Linux.

[Contract](CONTRACT.md) · [Evidence and limits](SUMMARY.md)

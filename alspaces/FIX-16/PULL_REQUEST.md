# fix(ci): restore release candidate validation (FIX-16)

Release candidate publication stops when the source-build installation test assumes a Git-derived version despite the job supplying explicit release metadata.

Planned correction: make the existing installation coverage deterministic for default and explicit build metadata while retaining source-native and release-cache checks. Implementation is pending scope approval.

Context: [contract](CONTRACT.md) and [summary](SUMMARY.md).

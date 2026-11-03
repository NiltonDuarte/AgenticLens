# fix(ci): restore release candidate validation (FIX-16)

Release candidate publication stops when the source-build installation test assumes a Git-derived version despite the job supplying explicit release metadata.

Make the existing source-build installation test control its subprocess metadata and exercise both Git-derived defaults and explicit candidate metadata through real native builds. Preserve installed-cache, launcher and native verification assertions in both cases. The release workflows and Makefile remain unchanged.

Context: [contract](CONTRACT.md) and [summary](SUMMARY.md).

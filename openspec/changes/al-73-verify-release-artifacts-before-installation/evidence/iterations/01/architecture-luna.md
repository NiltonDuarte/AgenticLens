# Architecture review — AL-73

**Outcome: APPROVED (scoped).** Reviewed baseline `b6ab0471224c8bcea7bc00adce867d3ba0de4a37` plus the frozen working diff identified by the coordinator as `47fd2495ce`. Requested reviewer profile: Luna / medium; runtime model and effort metadata were not independently available. Scope: release checksum helper and Make entry points, canonical setup downloader, shipped bundle installer, and their supported dependency boundaries. Requirements and approval were read from the AL-73 design, tasks, and acceptance specification. No sibling review reports were consulted.

## Findings

No blocking architecture or maintainability findings.

The release checksum helper is a repository-side release tool invoked only by `release-package` and `release-verify`; it is not part of the distributed setup bundle. Both publishing workflows invoke these Make targets on `ubuntu-latest`, which supplies the helper's `sha256sum`, `awk`, `sort`, and `mktemp` dependencies. The `Makefile` resolves the helper beside itself, so fixture callers using `make -f` do not need a copied script. The helper consumes the existing Make archive/binary inventory, requires each expected path to be a readable regular file, validates exact inventory membership and unique names before `sha256sum -c`, and stages generation before replacing `SHA256SUMS` in the same directory.

The setup downloader and bundle installer remain self-contained shell entry points. The downloader owns bootstrap acquisition, checksum admission and native manager handoff; it does not call the release helper or require Python/Node. The contributor script remains a thin `exec` into that canonical implementation. The bundle installer validates the remote owner/repository selector before fetching and has an explicit local-source path. The repeated repository grammar is small and kept at each network boundary; extracting it would add coupling between independently shipped shell entry points without reducing meaningful complexity.

## Hypotheses and evidence

- **Inventory duplication or omission:** inspected `release-package`, `release-verify`, and `scripts/release-checksums.sh`. Both recipes pass the Make release archive/binary variables to the same helper. Its AWK admission rejects unknown names and removes each expected name after one record, so duplicate records fail as unexpected and missing records fail after parsing.
- **Installer dependency leak:** inspected `.agents/skills/agentic-lens-setup/SKILL.md`, `scripts/download-agentic-lens.sh`, and both shipped setup scripts. No runtime reference to the release helper exists; native-manager bootstrap remains shell plus standard Unix utilities and binaries.
- **Premature host execution:** `release-verify` invokes the checksum helper before the host-only recipe lines. Host probes are conditioned on the selected host platform being in `PLATFORMS`, so foreign-only verification does not execute a stale host artifact.

A disposable local helper invocation with a duplicated manifest record exited nonzero. No product files were modified. Reused the coordinator-supplied capture suite result (23 cases passed); focused suites were reported as running. I did not launch tests or inspect their results.

## Limits

This is an architecture/dependency review of the supplied frozen diff, not a full correctness, test-oracle, or release-publication review. Hosted publication and execution outside the workflows' Ubuntu runner were not exercised. Approval applies only to the scoped architecture investigation and does not authorize publication or merge.

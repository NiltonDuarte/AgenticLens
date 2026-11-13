# AL-84 behavior read-binding verification

**APPROVED**, scoped to the same-ID directory-replacement correction and affected read/API/artifact interactions. Run `al84_behavior_final/read_bound`, iteration 01 correction verification; configured profile Sol/high. Runtime model/effort metadata is unavailable to this reviewer. No new findings. Product/tests remained read-only.

Candidate: baseline/HEAD `163181428ee9a5a599a2da1e91d3af5e618a9682` plus dirty manifest `41a2ee0b246538aba898442b36c22015511557d4fca7e231504f6f9a3d068720`. Independently checked all 674 entries: zero drift. Compared with the prior 673-file freeze: exactly the three assigned production files and two assigned tests changed.

DATA-01 residual is corrected. In `workspace_locations.go:230`, final admission compares the opened physical home with the directory observed during resolution. Internal repository resolution retains an independent descriptor opened through that observed root, preventing inode reuse until admission completes. The retained descriptor carries no flock; the final reader takes the normal shared lease. Mismatch closes the lease and returns an actionable unavailable/retry result, including when the replacement records the same ticket ID. Source inspection covers references closed on rejected candidates, ambiguity/incomplete outcomes, resolution errors and repository success/error returns. Public resolvers acquire no new caller-owned resource.

`repository.go:42` routes both artifact interfaces, single-ledger reads and detail through internal read resolution. Detail reads ledger and artifacts through one held directory; `publication_set.go:459` preserves pending checks around ledger reads, and artifact checks remain before/after their set. HTTP detail forwards selection into that coherent reader; an unavailable replacement result maps to HTTP 503 with diagnostics. The test correctly permits either copy when it occupies the selection before resolution, while requiring coherent successful ledger/artifact content.

The original Luna B-title count alone did not establish which directory resolution observed. Deterministic resolve-A/replace-with-same-ID-B evidence now establishes the actual boundary, followed by successful fresh B resolution. Discovery remains attributed to data Luna's verification-phase residual; this report adds verification, not discovery credit.

Executed from `backend/`:

```text
go test ./internal/adapters -run '^(TestResolvedWorkspaceReadsRejectSameIDDirectoryReplacement|TestResolvedReadReferenceAllowsWriterAdmission|TestLoadTicketWorkspaceDetailKeepsSameIDCopyCoherent)$' -count=1 -v
```

Passed in 1.012s: deterministic ledger/artifact rejection, writer admission with the retained reference, and 400 concurrent detail reads with 57 successful coherent pairs plus stable controls. `git diff --check` passed.

Reused `build/al84-correction-tests/SAME_ID_TEST_RECEIPT.json`; all 362 source hashes match. Affected adapters/capture/HTTP ordinary and race runs passed; CLI selection/export controls passed. Inspected physical-comparison mutation failure: the deterministic test reports that resolved A reopened B when the guard is disabled.

Carry forward BEH-01/02/03 approval from [prior verification](behavior-sol-recheck.md): generation admission and maintenance files are unchanged; runtime enumeration/completeness decisions remain intact, maintenance still selects metadata without a retained read reference, and shared pending-read extraction preserves its checks. Durability, test-oracle findings and prior native lifecycle evidence are unchanged and remain with their owning evidence. No broad suites, installed hooks or native macOS execution repeated. Parent owns integrated `make`, consolidation and publication; this scoped approval does not authorize merge.

# AL-84 DATA01 same-ID read-binding correction review

**Verdict: APPROVED for this scoped correction.** No new blocking data-integrity finding (0/10). This is not integrated or whole-change approval.

## Revision and scope

Reviewed the candidate based on `163181428ee9a5a599a2da1e91d3af5e618a9682` with the 674-file content manifest in `build/al84-review/READ_BOUND_MANIFEST.json` (`sha256` field `41a2ee0b246538aba898442b36c22015511557d4fca7e231504f6f9a3d068720`). The manifest’s five relevant entries for `workspace_locations.go`, `repository.go`, `publication_set.go`, `workspace_read_snapshot_test.go` and `workspace_replacement_linux_test.go` match the checked-out file hashes. Requested run: `al84_data_luna_read_bound`, iteration 01, `gpt-6-luna` / medium. The manifest hash is its embedded content-map hash, not a raw JSON digest.

The reviewed change is limited to same-ID selection/read binding. Resolution now retains an independent descriptor to the directory actually read, carrying its `FileInfo` into final admission. Final reads reopen and compare that directory with `os.SameFile`, then use the admitted directory lease for ledger, pending-marker, and artifact reads. The retained descriptor prevents inode reuse while remaining independent of writer flock ownership. Resolution closes retained references on nonmatch, ambiguity, and error; repository readers close them on all returns. Public resolver, maintenance selection, and writer ownership semantics remain unchanged.

## Findings and hypotheses

No introduced finding. The previous D-LUNA-1 title-count result did not correlate a title with the physical directory resolved, so it was not treated as proof. The replacement oracle in `TestResolvedWorkspaceReadsRejectSameIDDirectoryReplacement` directly resolves A, replaces its selected path with same-ID B, and requires retry/unavailable with nil ledger fields and no artifacts. A fresh resolution of B remains valid. `TestLoadTicketWorkspaceDetailKeepsSameIDCopyCoherent` permits either copy to be selected during exchange but requires the ledger title and artifact marker to match, preventing mixed-copy success.

The remaining failure hypotheses were descriptor lifetime or writer starvation, stale pending-marker checks, and an oracle insensitive to removal of the physical comparison. The source trace confirms the descriptor spans resolution through final admission, reads use the same held directory lease for pending checks and content, and cleanup is joined on repository returns. The writer-admission test passes while the reference is held. The disposable physical-guard mutant fails the deterministic replacement test, establishing that the oracle detects removal of the guard.

## Verification and limits

`build/al84-correction-tests/SAME_ID_TEST_RECEIPT.json` records the 362-file backend Go source fingerprint `4a0e699743da11d0f36bb98c2a11b152beda13208f6420e976df35291379d50f` and these successful checks with no skips: affected adapters/capture/HTTP ordinary tests (4.709s/0.008s/0.192s); corresponding race checks (28.170s/1.027s/1.670s); focused CLI selection checks (5.949s). It also records the expected failure of the disposable guard mutant. The deterministic resolved-A test and fresh-B control, concurrent whole-call coherence test, and writer-admission test were inspected in source and represented in the receipt’s selected checks.

Durability and maintenance conclusions from prior review remain unchanged because this correction does not alter their write, recovery, or maintenance paths; this report does not repeat or broaden those approvals. Integrated `make`, final consolidation, and other review focuses remain parent-owned and pending. This verdict covers only the same-ID physical read-binding correction at the manifest above.

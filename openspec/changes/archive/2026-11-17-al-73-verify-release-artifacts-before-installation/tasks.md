# AL-73 execution tasks

## Approval and implementation

- [x] Obtain explicit approval for the [design](design.md) and [acceptance scenarios](specs/release-artifact-verification/spec.md).
- [x] Make the packaging manifest replacement atomic and require complete, unique, well-formed inventory in both release targets. Retain hash failure propagation, deterministic output and host checks.
- [x] Require successful checksum commands and strict selected manager records in the canonical bootstrap. Validate remote owner/repository selection in both bootstrap and skill bundle installation. Keep manager and engine version selection independent. Validate preservation and reporting of completed and failed components through the shipped entry points.

## Isolated verification

- [x] Use the real Make recipes with temporary release directories. Inject one valid hash record then a nonzero exit, successful incomplete output and sort failure. Inspect exit status and final manifest contents after both fresh and repeated packaging.
- [x] Exercise missing, duplicate, unexpected and malformed records, missing/unreadable inputs and a mismatched digest. Confirm failure precedes any host probe. Verify a complete release, its archive membership and executable host identities, plus a foreign-only selection with an unrelated host file present.
- [x] Extend existing installation/distribution fixtures only for missing cases. Use a controlled download transport, logged exact version URLs, disposable paths and real benign native candidates. Confirm valid override URLs and zero fetches for malformed overrides. Exercise transfer/selected-record/digest failures and a hash command which outputs the expected value then fails. Observe exit status, probe order and absence of a final success claim.
- [x] With an existing installation, compare actual engine bytes, manager bytes/receipt/checksum, project pin and reported versions across total failure, an earlier successful engine followed by failure, and a successful replacement. Reuse manager HTTP, transaction recovery and native installation suites, both downloader entry points and capture distribution tests for hook preservation.

## Review and delivery

- [x] Freeze the candidate for independent behavioral, data integrity, architecture and test review under the development protocol. Resolve findings and record actors, revision, commands, outcomes and limits.
- [x] Commit the reviewed candidate, integrate current main and run the full make gate plus affected release checks. Record the integrated revision and staging/host evidence in delivery.
- [x] Prepare the roadmap and delivery artifacts for final telemetry and normal development PR publication.

### Independent review assignments

Each focus reviews the same frozen candidate independently on gpt-6.1-sol/high and gpt-6-luna/medium. Reports are owned individually under `evidence/iterations/01/<focus>-<profile>.md`, with profile `sol` or `luna`.

| Focus | Review question |
| --- | --- |
| behavior | Do the contributor, packaged setup and release workflow entry points retain exact selection, host checks and truthful outcomes? |
| data | Do failed digest, packaging and installation operations preserve valid state and receipts? |
| architecture | Do the recipes and installers share the correct inventory, preserve distribution boundaries and run with their supported dependencies? |
| tests | Do the isolated failure cases reach the intended gates and compare externally observable state? |

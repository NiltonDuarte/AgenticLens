# fix(AL-96): publish fresh ticket ledgers automatically

## Outcome

`capture prepare-publication` lets the operator select an empty destination for the first publication. The installed lifecycle binds the native session and the worker creates the ledger and any configured summary. The destination selection closes after publication or after an existing ledger is observed. Refresh follows the current ledger across configured roots. A pending local transaction recovers under publication ownership before a fresh write can proceed.

Codex capture checks native parent metadata to recover nested sessions without their own callbacks. Workers respond to eligible lifecycle requests during the periodic wait, retain normal retries for persistent failures, and finish a progressing child scan across bounded worker lifetimes.

## Verification and review

The product candidate is `02ff2608` over `origin/main` at `a5c8a25e`. The runner at `32fd1347` passed the [ordered native gate](evidence/native-gate.md) on the checksummed `v3.0.0-rc.0+al96.02ff2608` package. Both clients published four exact owners and reported zero provisional seconds in their first completion-aware commit.

The full integrated `make -j2` passed on `32fd1347` with the final Go stress fixture diff (file blob `6e938471b937518584548b706ff4d9547b7ce4a2`). It covered Go race and 1,050 frontend tests, static checks, installation, distribution and native builds. An earlier integrated attempt hit the synthetic test's observer limit. The revised fixture retains bounded directory visits and actual worker handoffs while allowing publisher startup under race-suite load. It passed twice under concurrent Make load and in the full integrated suite.

Independent behavior, architecture, data and test reviews used gpt-6.2-sol/high and gpt-6-luna/medium. Follow-up checks cleared the affected boundaries. The [review summary](evidence/REVIEW_SUMMARY.json) preserves the confirmed issues and their dispositions. The focused Go races also covered protected Codex, concurrent prepared destinations, two pending tickets, interrupted renewals and first commits, imported history, relocation, deletion, worker timing, registry contention classification and bounded Codex discovery.

## Limits

Packaged native checks target Linux amd64, Codex 0.170.0 and OpenCode 1.19.5. OpenCode 2.2.0 standalone has prior manual extraction evidence. Locally packaged candidates establish the installed path on the tested build. Stable release publication is outside AL-96 delivery.

Capture status for the actual Codex development session reported zero registrations and zero requested publication generations. Its AL-96 ledger and contributor coverage remain unavailable. The packaged gate verifies independent native fixture ledgers. This delivery follows the user's instruction to leave manual generation unused.

## Pull request

Create the first ticket ledger through normal installed lifecycle callbacks after explicit destination selection and verified binding. Protect first-write authorization through retries and prevent stale selections from recreating removed outputs.

Recover source-proven nested Codex sessions and shorten the interval between native completion and automatic publication. Add a repeatable native gate that checks the installed binaries, exact session ownership, tool activity, available usage, provisional and measured duration, and a later turn. The gate starts with an empty destination and uses automatic publication throughout.

The verified Linux amd64 package passed the native gate with Codex 0.170.0 and OpenCode 1.19.5. Both clients captured the root, two direct children and a grandchild and left an unrelated session out. The first commits with native completion evidence reported zero provisional seconds after 1.125 seconds and 1.272 seconds respectively.

The full integrated `make -j2` and OpenSpec strict validation passed.

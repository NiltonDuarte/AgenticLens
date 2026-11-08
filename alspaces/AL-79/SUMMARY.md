# AL-79: Native user-level Codex capture across worktrees

[PR130](https://github.com/NiltonDuarte/AgenticLens/pull/130). Consolidated review: APPROVED for the current user-agreed scope; no merge is authorized. Initial checkout was clean at4df45967c801824500ac2ea39a8f52fa9cbdcf66. Branch: feat/AL-79-native-user-codex-capture; configured workspace root: alspaces.

User approved [CONTRACT.md](CONTRACT.md) and [iteration01 plan](iterations/01/PLAN.md). The later instruction, "Remove windows test CI, I will remove support from windows later", approves [iteration02's delta](iterations/02/PLAN.md): remove the Windows job/build edge, retain runtime scripts/binaries/release targets, and stop further Windows verification. Other acceptance boundaries remain. AL-79 is moved to done and AL-80 is Ready.

## Delivered behavior

Six additive native Codex hooks install once in the effective user home. Callbacks validate exact session/source context and select one whole local configuration: worktree-local first, otherwise Git-base-local. Shared policy supplies executable selection; capture enablement, source and state remain local. Disabled capture stops before capture-only I/O. Unassigned facts stay local until explicit binding.

Adapters preserve root/child identities, ancestry and independently bound child tickets; retain immutable worker builds across rebuilds; and coalesce publication with bounded retry, durable recovery and deletion retention. Portable v6 history accepts an external parent without importing another ticket's metrics. Partial group context preserves owner/cached analysis without a false clean conclusion. Pure domain logic remains separate from I/O. Manual pins/cwd, normal trust and standalone embedded assets remain supported. Setup uses the common native resolver.

## Review and verification

Reviewed source160db423ce493942a7d0ca0cca6e56fe98185d75, source diff SHA256139d0e87e1e014ef0035b4aff4a09edfef843056352cd0545d2479e1309cf672; final six-file correction fingerprint69893e0f69877f377c9dcdf59989893b9768e6e0f6e69754112b2e0d58759adc. Behavior, data/failure, architecture, tests, security and build/distribution reviews used gpt-6.1-sol/high and gpt-6-luna/medium. Initial CHANGES_REQUESTED findings were corrected, later test-only scopes passed both profiles, and final bounded CI/distribution review passed both. Earlier Windows INCOMPLETE evidence was superseded by the explicit scope withdrawal, never relabeled a platform pass.

[Canonical findings](iterations/01/REVIEW_SUMMARY.json) retain28 candidates, numeric severity, every discovering persona/model/effort, fixes, the refuted routing claim and accepted Windows fixture debt. [Final correction review](iterations/02/REVIEW_SUMMARY.json) found no new candidate. Architecture/build reviewers and the independent parent, all gpt-6.1-sol/high, share process-lifetime discovery. Data reviewer gpt-6-luna/medium discovered executable retention; its architecture peer's routing objection was refuted. Go/distribution testers gpt-6.1-sol/high discovered recovery/scaffold/fixture issues; the parent gpt-6.1-sol/high found integrated assertion and native transport/diagnostic failures. Implementation engineer gpt-6.1-sol/high fixed runtime; Go/distribution testers gpt-6.1-sol/high fixed oracles. Independent reviewers and parent verified corrections; discovery is kept separate.

Full make passed on8b15b2ee: all Go races/recovery locks,190 frontend units,678 integration tests, installation/distribution/workflow checks and five-platform compilation. All314 Go/module/embed files remain identical: f1b674667f1247eeccfe01640f5353d27a27e9794831aa2b99ce6d917bd6195b. Current corrections also pass POSIX installation10/distribution21 and dependency-graph checks. [Test evidence](iterations/01/TEST_EVIDENCE.md) and [distribution evidence](iterations/01/DISTRIBUTION_EVIDENCE.md) distinguish passes, failures and stops.

Genuine normal-trust Codex0.167.0/Linux root/child capture and automatic publication passed in two worktrees with exact source facts and preserved history. The user signed in and trusted six isolated hooks normally. Unchanged callback/extraction/binding paths retain that revision-bounded [native proof](iterations/01/NATIVE_EVIDENCE.md).

Windows remains uncertified: actual win32 plugin6 and distribution12/14 plus isolated completion passed; DrvFS/setup and loaded-upgrade assurance remained limited. The earlier unavailable-local-runtime assumption was corrected. These retained limits belong to the user's planned support removal.

Telemetry: pending final extraction.

# GEN-20: Unique worktree coordination directories

Status: Proposed; implementation awaits explicit user approval.

## Request and baseline

The user requests that AgenticLens stop naming newly allocated local worktree directories sequentially as 1, 2, 3. Prefer an available worktree identifier, otherwise generate a short UUID, to avoid concurrency concerns. This ticket owns the local capture coordination namespace under the configured capture state directory, normally ~/.agenticlens/state; it does not create Git worktrees.

Baseline: dc44619ae15586fe1194f872743033f523635c92. Initial Git working tree was clean. Branch: chore/GEN-20-unique-worktree-identifiers. Workspace root: alspaces, from .agenticlens-config.yaml. GEN-20 had no prior workspace or plan-index entry.

Current store.go allocates a UNIQUE worktree directory row inside an immediate SQLite transaction. This grounding does not establish a race in committed numeric SQL allocation. The requested change removes coupling between those internal row IDs and filesystem names; it must preserve the existing transactional isolation.

## Proposed behavior and acceptance boundaries

1. New named-project worktree scopes persist a portable single-component coordination token. For linked Git worktrees, prefer the Git administrative worktree identifier observed through existing ownership metadata, qualified by a short digest of its canonical administrative identity so identical basenames from different repositories sharing project_name cannot collide. Do not trust a checkout basename, branch name, environment variable, or unverified external harness identifier as unique authority.
2. When no linked-worktree identifier is available, including ordinary checkouts and non-Git directories, allocate a cryptographically random short UUID representation of 16 hexadecimal characters. Persist it once; reopening the same worktree, adding Git later, and concurrent opens must reuse the committed token.
3. A UNIQUE persisted mapping from numeric worktree row ID to coordination token is authoritative. Keep numeric SQL IDs and row ownership semantics intact. Serialize allocation with the existing immediate transaction. Detect token collisions and reserved filesystem entries; use bounded random retry and return a contextual error on exhaustion. Never adopt another scope's directory or overwrite unknown existing state.
4. Existing registered worktrees retain their current numeric coordination paths, manifests, locks, registrations, receipts and pending relocation state. No automatic directory rename or reinstallation is needed. Preserve unassociated Open(directory) behavior. The compatibility design must prevent older binaries from silently reconstructing numeric paths for newly tokenized scopes; use an explicit registry protocol transition with transactional upgrade of supported protocol-2 registries. Protocol-1 recovery remains explicit. Read-only inspection of an unchanged protocol-2 registry resolves its existing numeric paths without migration or writes.
5. All path consumers resolve the same persisted mapping: writable and read-only opens, explicit coordination aliases, preview, installation, worker startup/successors, diagnostics, retained-state inspection and relocation/resumption. Accept only safe tokens beneath the validated namespace; preserve current containment, symlink and exact-worktree ownership checks for both numeric and token paths.
6. Preview/dry-run remains read-only. Existing scopes expose their persisted path. An unallocated scope is explicitly provisional, with no guessed SQL allocation or durable UUID claim. Callers must distinguish that state: actual installation opens the scope before rendering durable manifests; policy equivalence compares validated registry/worktree identity rather than two independently generated preview UUIDs. Missing state remains absent to workers and diagnostics.
7. Existing sessions and portable ledgers remain usable without the original machine's capture state. The new local mapping is operational state and must not become a new source authenticity requirement or change session identity, refresh precedence or recorded facts.

## Architecture and scope

Changes belong to Go capture adapters and their CLI consumers. Filesystem observation, entropy, SQLite transactions and upgrade handling stay in the imperative shell. Domain calculations remain pure; no frontend, ledger-schema, configuration-key or runtime dependency changes are proposed. Apply TECHNICAL_GUIDELINES.md: strong typing, zero any/interface{}, explicit errors and zero emojis.

The implementation may introduce a focused versioned registry-upgrade helper and one shared token/path resolver. Do not spread independent token generation or numeric path reconstruction across callers. Concrete helper/file names are implementation choices within these boundaries.

## Verification and completion

Extend coherent existing fixture-based coverage for concurrent allocation and reopening, same-name linked worktrees across distinct repositories, UUID fallback, collision handling, legacy numeric state, aliases/containment, installer previews and real installation, relocation/resume and read-only retained inspection. Preserve worker isolation and existing ownership-negative assertions. Tests must use temporary state and repositories, never this ticket workspace or the user's capture state.

Completion requires coder implementation, tester evidence, independent focused review, integrated make, final telemetry and normal PR delivery. No implementation, runtime validation, review approval or completion is claimed by this proposed contract.

# AL-79 iteration03 test-effectiveness review

**APPROVED within the assigned focus.** No blocking correctness finding. One nonblocking regression-test improvement, severity **3/10**.

Run: `i03_tests_sol`, independent discovery; requested profile `codex/gpt-6.1-sol/high` (runtime model/usage metadata unavailable). Base `9c39cf9d4f44bdb968cb92b0212dccf577d10255`; frozen 58-path candidate SHA-256 `c84c6b274d12ef53544596f8214c1e5501cf6022883276fc0184bf528320e5cd`, verified before and after investigation. Iteration03 PLAN supplies the approved manager/engine split and compatibility waiver.

## Finding

**TEST-SOL-01 — Ordinary cache identity rejection lacks a durable negative oracle (severity 3, nonblocking).** `manager/native_test.go` rejects a wrong-version artifact during installation, but that exercises `installEngine`, not the independent runtime comparison at `manager/run.go:206`. A disposable overlay disabling only that comparison leaves `go test -race ./internal/adapters/manager -count=1` green (6.955s) and all 22 distribution tests green (23.508s). A separate control puts a real v9.4.3 engine in the v9.4.2 cache leaf: the candidate rejects it (1.589s), while the mutant selects it successfully and fails the control assertion (0.759s). A focused negative case would protect the explicit requirement to reject mismatched cached identities. Current production behavior is correct; this is a test improvement, not an observed product defect or a coverage quota.

## Investigation and evidence

Read changed test bodies, helpers, new untracked manager tests and production selection/admission/recovery paths. The five deleted dispatcher CLI cases retain their meaningful assertions in `capture_user_hook_test.go`; enabled child callbacks now traverse the real manager. Protocol-1 execution and Windows dispatcher hash tests legitimately retire with waived compatibility and deleted activation code. Executable ownership/recovery assertions move to manager tests; hook tests retain independent homes, additive ownership, edited-hook refusal, migration and pending-state preservation. Native probes observe process transport; real binaries exercise admission and native hooks. Fixtures use isolated filesystem roots and no live ticket history.

Reused the supplied installation 10-test, distribution 22-test, capture/config/CLI race and checksum/role/diagnostic mutation evidence within their stated boundaries. Independently ran the unmodified manager race suite: passed, 6.832s. The exact-pin overlay runs above were additional diagnostic probes, not canonical integration gates.

An additional disposable bootstrap test exercised downloaded manager v9.8.8 with local pinned engine v9.8.7: checksum failure and partial transfer left home/cache/destination unchanged; successful installation preserved independent versions, exact release URLs and absence of an implicitly provisioned companion. Passed, 2.971s. Transfers use the fixture curl executable and real artifact bytes; this does not certify published GitHub assets.

Investigated later-prefix recovery conflicts: interrupting after checksum write and editing the executable causes partial rollback of unchanged metadata before refusal. The edited executable and journal survive; restoring only the recorded executable bytes lets the same journal finish recovery. An initial stronger no-mutation assertion failed, but the approved contract does not require that guarantee. Disposition: safe/resumable in the exercised case; provisional blocker withdrawn.

Disposable probes/overlays: `/tmp/al79-tests-sol-lavy9yjr/`. No candidate application/test edits, actual-home hook changes, commits or telemetry.

## Limits

No Windows runtime, hosted release execution or new live Codex trust exchange was claimed. Prior live source/binding/ledger evidence applies only to unchanged paths. Parent owns full `make`, consolidation, real installation and delivery; this scoped approval does not certify those pending gates.

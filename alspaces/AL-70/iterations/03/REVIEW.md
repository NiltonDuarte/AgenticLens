# AL-70 CI fixture-lifecycle correction

Baseline 07129d42 began clean. Consolidated delivery is INCOMPLETE while the hosted Backend Check failure is corrected. Prior source behavior approvals and local gates remain scoped to their tested revisions.

A12-TEST-02 (severity 3): CLI tester reproduced the hosted TempDir cleanup failure under race/count20 and traced reconcile-spawned detached workers whose parent commands had exited. The fixture has no serialization/join. Tester owns the one test-file correction using the existing fixture helper; source capture behavior remains unchanged. Raw evidence is /tmp/al70-ci-42988883853-failed.log and /tmp/al70-ci-tail-before-race.log.

Review coverage is test effectiveness with both gpt-6.1-sol/high and gpt-6-luna/medium. Product code, schema, owner/policy rules and frontend are unchanged, so the other focus approvals carry forward with this narrow dependency rationale. Parent owns integrated make, final telemetry and CI delivery. No retries, silent cleanup errors or changed product behavior are authorized.

Both test-effectiveness profiles approved the one-line frozen correction without new findings. Reviewed SHA-256 is 2e6a8d844233b33c4f0d258cfbc86834ded46edef91e4cd2aacb4bffc911947b. CLI tester's fixed count20 race run passed (15.093s), as did same-file siblings and the completion control (6.485s). Reviewers verified lease/log completion, late-start coverage and LIFO cleanup while every behavior assertion remains unchanged. Verification adds no discovery credit. Parent integrated make remains pending.

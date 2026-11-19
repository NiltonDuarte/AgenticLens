# Review scope and evidence

REV-8 covers seven merges to `main` from `16318142` through `9308ac77`. Review each squash commit against its first parent and confirm whether a finding remains at the final revision.

The user requested the next available review ID and authorized adding the explicit `REV-` configuration prefix for session binding. The approval applies to that configuration change and the audit. The coordinator owns the change's review artifacts and any shared verification. Reviewers work against the frozen revision and submit evidence from disposable fixtures.

| Focus | Primary risk |
| --- | --- |
| Behavior and integration | CLI and dashboard reachability, capture routing, live source recovery and documented defaults |
| Data integrity and failure handling | Physical ledger identity, portable facts, capture ownership and ambiguous or failed source reads |
| Architecture and maintainability | Configuration, transaction and release boundaries across callers and adapters |
| Test effectiveness | Whether fixtures exercise regressions and end-to-end controls |
| Release and installer verification | Manifest admission, published inventory and setup/runtime version selection |
| Live stream and dashboard behavior | Read ordering, connection lifecycle, backpressure and recovery from database failure |

Each focus uses the configured `gpt-6.1-sol/high` and `gpt-6-luna/medium` profiles. Discovery runs are independent. The final report records run outcomes and preserves unresolved issues.

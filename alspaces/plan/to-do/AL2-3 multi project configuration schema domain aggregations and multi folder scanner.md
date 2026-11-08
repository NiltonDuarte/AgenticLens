h3. Title: AL2-3: Multi-Project Configuration Schema, Domain Aggregations, and Multi-Folder Scanner

*GOAL*

Engineering leads can scan several local projects in one invocation and receive correct per-project and combined telemetry.

*CONTEXT*

The existing project name and single-project scanner do not provide multi-project identity or configuration isolation. Identical ticket IDs in two repositories must remain distinguishable. This foundation enables remote ingestion and the project dashboard.

*ACCEPTANCE CRITERIA*

* Configuration supports projects with stable IDs, names and local locations, including each project’s workspace/configuration location. scan and server accept repeatable --project <id>=<path> and --projects-config. Explicit CLI definitions override the same configured ID; invalid/duplicate definitions within one input fail clearly.
* Each project uses its own configuration, workspace root, ticket prefixes, roles and classification policy. Resolve configured relative locations from their declaring file; explicit relative CLI paths remain caller-relative.
* Scan results and CLI tables identify project origin. Per-project filtering and combined rollups count each project/ticket identity once; matching ticket IDs across projects never overwrite each other.
* A failed or unavailable project is diagnosed separately while valid projects remain usable. Preserve unavailable values and supported committed facts without original source logs or capture state.
* Existing single-project invocations remain usable with their prior semantics. Deliver local-project scanning and aggregation without requiring remote fetch or a new dashboard selector.

*Status*: Ready
*Blocked by*: None
*Work type*: feat

*IMPLEMENTATION NOTES*

AL2-4 owns remote synchronization and HTTP project registry/filtering; AL2-5 owns interactive project selection. Project identity is a read/aggregation context, not permission to rewrite committed ledger ownership. Concrete Go representations remain developer-owned.

Follow [engineering rules](../../../TECHNICAL_GUIDELINES.md) and [regression boundaries](../../../docs/REGRESSION_PREVENTION.md). [PO-19](../../PO-19/SPEC.md) preserves the earlier development reconciliation; historical observations require revalidation, not an assumption of current runtime support.

h3. Title: AL2-5: Multi-Project Web Dashboard Selector, Cross-Project Fleet Overview, and Comparative Analytics

*GOAL*

Dashboard users can select projects, compare their metrics and retain project context when opening individual tickets.

*CONTEXT*

AL2-3 and AL2-4 provide project identity, registry and query support. Existing views operate on one project. A global selection must update every affected view consistently and recover when a project or synchronization is unavailable.

*ACCEPTANCE CRITERIA*

* A global accessible multi-select lists configured projects from GET /api/projects. All Projects means every configured project; clearing the explicit selection restores All Projects. One or more IDs constrain all affected views to that set.
* Selection updates URL ?projects=... and survives refresh with local preference storage. Explicit URL selection takes precedence over stored preference; removed/invalid IDs are reported and can be cleared rather than yielding an unexplained empty view.
* Workspace rows show project identity in multi-project mode; detail breadcrumbs/header preserve it. Two projects with the same ticket ID navigate independently.
* Fleet views compare cost, velocity and anomalies by project and provide combined totals for the selected set. Unavailable costs retain their coverage rather than becoming zero; stale responses cannot overwrite a later project selection.
* Show synchronization state and allow on-demand refresh through the registry API. Loading, sync failure and empty results retain usable selection/retry controls; refresh never displays another project’s data under the current label.

*Status*: Blocked
*Blocked by*: AL2-3, AL2-4
*Work type*: feat

*IMPLEMENTATION NOTES*

Reuse existing view and query contracts rather than prescribing component filenames. AL-71 owns general request-state corrections; coordinate their invariants. No new ingestion or rate-card cost policy is owned here.

Follow [engineering rules](../../../TECHNICAL_GUIDELINES.md) and [regression boundaries](../../../docs/REGRESSION_PREVENTION.md). [PO-19](../../PO-19/SPEC.md) preserves the earlier development reconciliation; historical observations require revalidation, not an assumption of current runtime support.

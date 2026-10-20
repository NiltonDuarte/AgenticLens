h3. Title: AL2-4: Remote Git Repository Ingestion, Shallow Cache Manager, and Project Management API

*TARGET*: AgenticLens (adapters | git | http | server | tests)

*CONTEXT*
 
While AL2-3 enables multi-project scanning of local directories, engineering managers also need to monitor remote git repositories directly (e.g. `https://github.com/myorg/billing-service.git`) without requiring developers to pre-clone every project manually. To support this safely and efficiently, AgenticLens requires a dedicated git cache manager performing shallow clones, background synchronization, and a project registry REST API enabling live project discovery and on-demand syncing.

*GOAL*
 
Build a shallow git repository cache manager in `internal/adapters/git` that clones and updates remote repositories in `~/.cache/agenticlens/remotes/`, integrate it into the multi-project scanner, and expose `GET /api/projects` and `POST /api/projects/sync` endpoints with multi-project filtering on all telemetry queries.

*VERTICAL SLICE SCOPE*
 
* Boundary Adapter (Git Cache): Implement `GitCacheManager` in `internal/adapters/git/cache.go`. Manage shallow clones (`git clone --depth 1` / `git fetch --depth 1`) in a deterministic cache folder (`~/.cache/agenticlens/remotes/<project_id>/`). Validate remote repository URLs (HTTP/HTTPS/SSH) and isolate credentials. Support on-demand refresh and graceful fallback to cached state on network errors.
* Boundary Adapter (HTTP Server): Implement `ProjectsHandler` (`GET /api/projects`) returning list of monitored projects, source types, paths, ledger counts, and last synchronization timestamps. Implement `ProjectSyncHandler` (`POST /api/projects/sync`) triggering background fetch of remote repositories.
* Query Filtering: Update `WorkspacesHandler`, `FleetHandler`, `ComparisonHandler`, and `ProgressionHandler` in `internal/adapters/http/handlers.go` to parse `project` query parameters (e.g. `?project=billing&project=core`), applying domain multi-project filtering.
* Quality: Integration tests in `internal/adapters/git` using local bare git repositories verifying shallow cloning, fetching, error handling, and cache isolation. HTTP server tests in `internal/adapters/http` asserting project registry outputs, sync responses, and multi-project query filtering.

*ACCEPTANCE CRITERIA*
 
* Remote git URLs configured in `projects:` or passed via `--project <id>=<url>` are shallow-cloned into the cache directory.
* `GET /api/projects` returns the complete list of configured projects with connection status, workspace counts, and last sync timestamp.
* `POST /api/projects/sync` fetches remote repositories and updates cached workspaces without blocking active reads.
* `GET /api/workspaces?project=billing` returns only ledgers belonging to the billing project.
* When multiple projects are requested (`?project=p1,p2`), responses aggregate only those selected projects.
* All Go test suites pass with zero warnings or data races (`go test -race ./...`).

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.

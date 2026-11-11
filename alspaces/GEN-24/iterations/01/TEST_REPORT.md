# GEN-24 verification

The tester extended the existing fixture-driven Jekyll suite to publish beneath `/FixtureRepo/preview/24`, in addition to project and host roots. The same checks validate local asset/fragment destinations, accessible landmarks and illustrations, nested roadmap navigation, and source links pinned to the requested revision. Fixtures use temporary repositories rather than the live ticket workspace.

Final focused verification exercised the uncommitted GEN-24 candidate over baseline `167c204fc958da34eb82d97129dedd85ac9b8c58` after preview publication scripts were completed:

- `docker exec gen24-ruby bundle exec ruby .github/pages/tests/site_test.rb`: 8 runs, 453 assertions, 0 failures, 0 errors, 0 skips. The container uses Ruby 3.3 and the repository's pinned bundle.
- `python3 .github/pages/tests/preview_test.py`: 9 tests, all passed, no skips.
- `git diff --check -- .github/pages/tests/` and Python test byte-compilation: passed.

The Python suite exercises static composition, replacement, preservation of production and other previews, last-success retention while a PR head advances, close cleanup, stale/closed/fork/mismatched candidates, retargeted PR removal and rejection, symlinks/private payloads, invalid previous-state metadata and protected output paths. Rejected inputs preserve existing publication bytes; composition preserves source files.

The interrupted-publication test executes the actual `preview.py load-state` CLI against a real temporary bare Git remote, with an exact URL rewrite to the local fixture. It first loads clean committed publication history, then commits a pending-publication marker and confirms that the next load fails closed while preserving the marker and previous site. It uses no mocked domain or network response.

The tested `preview.py` content SHA-256 is `3d5f804ae7cb89020effaaca370f83b6e7f3c8302fb255ac3342cb34a265aa1c`. The parent owns the final candidate fingerprint, integration and delivery checks. No application behavior or embedded dashboard asset changes are included.

A real backward-compatibility probe executed the updated workflow’s complete candidate shell block against the baseline Pages source archived from `167c204fc958da34eb82d97129dedd85ac9b8c58`, which has neither new composer file. In an isolated temporary repository, the pinned bundle installed, baseline Jekyll tests passed (8 runs, 335 assertions, zero failures/errors/skips), and the production builder generated the preview with correct roadmap and stylesheet prefixes. Ruby YAML parsing also verified `queue: max` on publication and Python composer tests in the trusted resolve job. The tested workflow SHA-256 is `916044958bfce6bfb49ffe7a1883300bafe89718102e00d6f4bdb32864f9bc19`. Probe source and archives stayed outside Git. This execution, rather than text matching alone, establishes that an older PR does not need new composer test files to build.

These checks do not establish GitHub Actions event authorization, remote artifact provenance, concurrent hosted publication, deployment success or public preview availability. Those boundaries require workflow review and hosted evidence before delivery; no planned or unexecuted check is counted as passed.

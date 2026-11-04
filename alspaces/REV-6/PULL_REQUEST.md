# docs(review): measure test cost and recommend verification boundaries (REV-6)

Capture verification mixes pure calculations, real storage, native CLI setup and lifecycle waits. REV-6 records comparable local measurements, separates package/setup costs from individual tests, and recommends explicit fast gates before slower integration work.

Includes a capture-first boundary inventory, preserved assertion owners, resource/cache/race assumptions, provisional local budgets and prioritized follow-ups. Application, test, fixture, runtime and CI files remain unchanged; implementation requires separate tickets.

- [Contract](CONTRACT.md)
- [Measured review](iterations/01/TEST_REVIEW.md)
- [Summary and limits](SUMMARY.md)

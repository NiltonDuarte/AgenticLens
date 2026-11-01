# Research basis and application

Primary sources consulted in the preceding research turn, 2026-11-01. These are engineering accounts/product documentation, not controlled evidence that their results transfer here.

- [Anthropic Code Review](https://support.claude.com/en/articles/14233555-set-up-code-review-for-claude-code): correctness-focused specialized investigation, candidate verification and deduplication; adopted investigation/verification separation without mandatory parallelism.
- [Cursor: Building a better Bugbot](https://cursor.com/blog/building-bugbot): evaluates reviewer configurations against annotated bugs and observed resolution; adopted bounded old/new forward testing. Its historical multi-pass voting design is not treated as a requirement, and agreement is not proof.
- [OpenAI: Harness engineering](https://openai.com/index/harness-engineering/): concise navigable guidance, accessible evidence and mechanically enforced invariants; adopted one canonical procedure and links, preserving existing executable regression checks.
- [Aider linting/testing](https://aider.chat/docs/usage/lint-test.html): configurable checks feed failures back into editing; existing make gate remains authoritative, not duplicated in review prose.
- [Claude Code best practices](https://code.claude.com/docs/en/best-practices): distinguishes model instructions from deterministic enforcement; permission guidance explicitly acknowledges harness-specific limits.

No vendor performance statistic is used as a local quality claim. Broader automatic architecture enforcement or a production reviewer benchmark is not implemented in this documentation iteration.

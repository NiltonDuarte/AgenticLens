# REV-4: Test necessity audit

Baseline: `3481919832272d3a545f8d7a1c6c98a1f678fc58`. Initial working tree clean. Branch: `REV-4/reduce-low-value-tests`.

Review all tracked executable test files in backend and web, identify redundant or ineffective assertions, and explain individual removal recommendations with surviving coverage and risk. Preserve meaningful boundary, regression, provenance, security, and integration coverage. Do not infer author intent from passing tests.

This iteration delivers a documentation audit. Test deletion requires a concrete reviewed removal scope through the development workflow. No application, test, schema, or runtime script changes in this iteration. Benchmarks, fixtures, and helpers must be distinguished from executable tests. Record limitations rather than claiming a deletion was proved safe without running a reduced suite.

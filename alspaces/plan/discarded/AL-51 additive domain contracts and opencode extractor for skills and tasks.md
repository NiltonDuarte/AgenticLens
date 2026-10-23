# AL-51: Additive Domain Contracts and OpenCode Extractor for Skills and Tasks

- **STATUS**: Discarded (Superseded and Merged into AL-52)

## Discard Reasoning
This ticket was discarded and consolidated directly into **AL-52**. Separating additive domain model contracts and batch OpenCode extraction (AL-51) from real-time live SQLite stream extraction (AL-52) created an artificial division across the dual-plane telemetry ingestion pipeline. Both planes touch the exact same domain facts (`SkillFact`, `TaskLinkFact`, `LivePartFact`), consume the same underlying OpenCode SQLite databases, and require synchronized testing across Go backend and TypeScript frontend layers. Consolidating the full ingestion scope into AL-52 delivers an atomic, end-to-end backend vertical slice and accelerates frontend presentation work (AL-53 through AL-55).

## Original Scope
Refer to `alspaces/plan/to-do/AL-52 additive domain contracts opencode extractor and live stream ingestion for skills and tasks.md` for the unified implementation requirements.

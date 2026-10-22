# Contract: AL-1 Core Go Domain Models and Aggregations Engine

## 1. Problem Statement
Build the foundational Go domain model for AgenticLens representing atomic telemetry facts, workflow efficiency profiles, thread structures, and multi-ticket aggregations.

## 2. Invariants
- Functional Core, Imperative Shell (FCIS): Zero IO in domain layer.
- Static typing: Zero interface{} or any in Go domain packages.
- Zero emojis across all log strings, comments, and identifiers.

# AgenticLens Technical Guidelines

Mandatory technical guidelines for developing AgenticLens. All code must adhere to these rules.

---

## 1. Architectural Boundaries and Layer Separation

Dependencies flow strictly inward across three distinct layers:
1. **Presentation Layer** (`agentic_lens/presentation/`): Streamlit dashboard (`app.py`), modular view components (`views/`), formatters, and CLI (`cli.py`).
2. **Adapter Layer** (`agentic_lens/adapters/`): External boundary schemas with Pydantic validation (`schemas.py`) and filesystem repository scanner (`repository.py`).
3. **Domain Layer / Functional Core** (`agentic_lens/domain/`): Pure calculation functions (`aggregations.py`), frozen dataclasses (`models.py`), constants (`constants.py`), and enums (`enums.py`).

### Rules
- The functional core (`agentic_lens.domain`) must never import outer layers (`adapters`, `presentation`) or external framework libraries (`streamlit`, `pydantic`, `pandas`, `altair`).
- Adapters depend on Domain and external boundary libraries (Pydantic).
- Presentation depends on Adapters, Domain, and UI libraries (Streamlit, Altair, Pandas).
- **Zero-Bypass Policy**: Never introduce circular imports or leak framework types into inner layers.

---

## 2. Functional Core, Imperative Shell (FCIS)

Split operations into two clear parts:
1. **Functional Core (`agentic_lens.domain`):** Deterministic, pure functions. No disk, no network, no database calls, no clocks, no random generators. Takes immutable data in, returns computed metrics or typed domain models out. Tested entirely without mocks.
2. **Imperative Shell (`adapters` & `presentation`):** Scans the filesystem, validates external JSON ledgers, passes domain models to core functions, and renders visualizations.

### Aggregations & Calculations in the Core
- All mathematical rollups, context health categorizations, token formatting, turn velocity, and cost aggregations must be pure functions in `agentic_lens.domain.aggregations`.
- Defensively guard against division by zero (empty workspaces, zero tokens, zero cost, zero duration).

---

## 3. Stateless Classes, Inheritance Restrictions, and Protocols

The goal is preventing shared mutable state and fragile coupling.

### Rules
1. **Stateless Operations**:
   - Classes and repository helpers must remain strictly stateless. Methods take inputs and return outputs without mutating internal instance variables or accumulating run state.
2. **Inheritance is Strongly Discouraged**:
   - No inheritance in business logic or domain models.
   - Permitted solely at framework boundaries where library types mandate it (e.g. `pydantic.BaseModel` in `adapters/schemas.py` or `enum.StrEnum`).
3. **Avoid Abstract Base Classes (`abc.ABC`) — Use `typing.Protocol`**:
   - Use `typing.Protocol` or pure `Callable` type aliases for structural subtyping.
4. **Class Archetypes**:
   - **Data Containers:** `@dataclass(frozen=True, slots=True)` with zero business logic methods for all immutable domain data models.
   - **Boundary Models:** Pydantic `BaseModel` models in `adapters/schemas.py` for input validation and JSON deserialization.

---

## 4. Boundary Validation with Pydantic vs. Internal Dataclasses

Python dataclasses perform zero runtime type verification. They are fast and memory-efficient. Conversely, Pydantic provides runtime schema validation. We pay the validation cost **only at the boundaries**.

### Rules
1. **External Boundaries (Pydantic Required)**:
   - External JSON input (`agentic_metrics.json`) is parsed and validated exclusively via Pydantic models (`adapters/schemas.py`).
   - Pydantic models live exclusively in `adapters/` and must never leak into the domain core.
2. **Internal System (Dataclasses Mandate)**:
   - Data strictly internal to the system (domain entities, rollups, aggregations, scan reports) must be immutable dataclasses: `@dataclass(frozen=True, slots=True)`.
   - Once external data is validated at the boundary by Pydantic, the adapter immediately maps it into an immutable dataclass via `.to_domain()`.
3. **Dictionaries Strictly Banned**:
   - Raw dictionaries are permitted only at the immediate serialization boundary before Pydantic ingestion. Raw dictionaries must never be passed across internal layer boundaries.

---

## 5. Typing as the Primary Communication Layer

Docstrings rot. Types are verified on every check.

### Rules
- Target Python 3.14+ syntax and typing standards.
- Use Python PEP 695 `type` aliases (e.g. `type TicketId = str`).
- The `Any` type is banned across domain, adapter, and presentation layers.
- Model business states and outcomes as Discriminated Unions (Sum Types) using frozen slotted dataclasses.
- Enforce `enum.StrEnum` for all domain categorizations (e.g. `TimeGranularity`, `TimeWindowPreset`). Raw string primitives are banned for domain categories.

---

## 6. Presentation Layer & UI Invariants

AgenticLens uses Streamlit and Altair for data visualization.

### Rules
1. **Modern Streamlit Syntax**:
   - Streamlit >= 1.42.0: Avoid deprecated parameters such as `use_container_width`. Use `width="stretch"` or `width="large"`.
2. **Zero Emojis Mandate**:
   - Zero emojis permitted in code, docstrings, UI titles, tab labels, metric cards, tables, logs, or terminal output.
3. **Defensive Presentation**:
   - Views must handle missing or empty datasets gracefully with `st.info()` or empty state placeholders without raising uncaught exceptions.

---

## 7. Quality Checks & Verification

Before submitting changes, all checks must pass:
1. **Compilation**: `python3 -m py_compile agentic_lens/**/*.py scripts/*.py tests/*.py`
2. **Linting**: `uv run ruff check agentic_lens/`
3. **Testing**: `uv run pytest tests/` (zero silent skips)
4. **Deprecations**: `grep -rn "use_container_width" agentic_lens/` (must return zero matches)
5. **Emoji Audit**: Verify zero emojis in code and UI strings.

---

## Summary Checklist

| Topic | Decision | Rule |
|---|---|---|
| **Architecture** | Inward dependencies | `domain/` has zero external imports. `adapters/` handles validation. `presentation/` handles UI. |
| **Separation** | FCIS | Pure deterministic logic in domain; filesystem I/O and Pydantic in shell. |
| **Data Models** | Pydantic at Boundaries, Dataclasses Internally | Pydantic for ingress validation; `@dataclass(frozen=True, slots=True)` internally. Dictionaries banned. |
| **Typing** | Strict Python 3.14+ | PEP 695 type aliases, `StrEnum`, zero `Any`. |
| **UI & Styling** | Modern Streamlit | `width="stretch"`, zero emojis. |
| **Testing** | Deterministic Pure Core | Zero mocks for domain tests. Robust boundary tests for malformed JSON and path traversal. |

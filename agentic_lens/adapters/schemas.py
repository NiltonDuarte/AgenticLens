from __future__ import annotations

from pydantic import BaseModel, ConfigDict, Field

from agentic_lens.domain.models import (
    AgenticFactLedger,
    CodeFacts,
    InvocationFact,
    InvocationTokens,
    ToolFact,
)


class InvocationTokensSchema(BaseModel):
    model_config = ConfigDict(extra="ignore")

    input: int = 0
    output: int = 0
    cache_read: int = 0
    reasoning: int = 0
    peak_context: int = 0

    def to_domain(self) -> InvocationTokens:
        return InvocationTokens(
            input=self.input,
            output=self.output,
            cache_read=self.cache_read,
            reasoning=self.reasoning,
            peak_context=self.peak_context,
        )


class InvocationFactSchema(BaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    parent_id: str | None = None
    role: str = "orchestrator"
    model: str = "default"
    user: str = "unknown"
    started_at: str | None = None
    completed_at: str | None = None
    active_duration_seconds: int = 0
    turns: int = 0
    tokens: InvocationTokensSchema
    cost_usd: float = 0.0
    compactions: int = 0

    def to_domain(self) -> InvocationFact:
        return InvocationFact(
            id=self.id,
            parent_id=self.parent_id,
            role=self.role,
            model=self.model,
            user=self.user,
            started_at=self.started_at,
            completed_at=self.completed_at,
            active_duration_seconds=self.active_duration_seconds,
            turns=self.turns,
            tokens=self.tokens.to_domain(),
            cost_usd=self.cost_usd,
            compactions=self.compactions,
        )


class ToolFactSchema(BaseModel):
    model_config = ConfigDict(extra="ignore")

    name: str = "unknown"
    calls: int = 0
    errors: int = 0

    def to_domain(self) -> ToolFact:
        return ToolFact(
            name=self.name,
            calls=self.calls,
            errors=self.errors,
        )


class CodeFactsSchema(BaseModel):
    model_config = ConfigDict(extra="ignore")

    lines_added: int = 0
    lines_deleted: int = 0
    test_lines_added: int = 0
    test_lines_deleted: int = 0
    prod_lines_added: int = 0
    prod_lines_deleted: int = 0
    files_changed: int = 0
    sdlc_iterations: int = 1

    def to_domain(self) -> CodeFacts:
        return CodeFacts(
            lines_added=self.lines_added,
            lines_deleted=self.lines_deleted,
            test_lines_added=self.test_lines_added,
            test_lines_deleted=self.test_lines_deleted,
            prod_lines_added=self.prod_lines_added,
            prod_lines_deleted=self.prod_lines_deleted,
            files_changed=self.files_changed,
            sdlc_iterations=self.sdlc_iterations,
        )


class AgenticFactLedgerSchema(BaseModel):
    model_config = ConfigDict(extra="ignore")

    ticket_id: str | None = None
    ticket_title: str | None = None
    ticket_type: str | None = None
    started_at: str | None = None
    completed_at: str | None = None
    updated_at: str = ""
    code: CodeFactsSchema = Field(default_factory=CodeFactsSchema)
    invocations: list[InvocationFactSchema] = Field(default_factory=list)
    tools: list[ToolFactSchema] = Field(default_factory=list)

    def to_domain(self) -> AgenticFactLedger:
        return AgenticFactLedger(
            ticket_id=self.ticket_id,
            ticket_title=self.ticket_title,
            ticket_type=self.ticket_type,
            started_at=self.started_at,
            completed_at=self.completed_at,
            updated_at=self.updated_at,
            code=self.code.to_domain(),
            invocations=tuple(inv.to_domain() for inv in self.invocations),
            tools=tuple(t.to_domain() for t in self.tools),
        )

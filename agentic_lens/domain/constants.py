from __future__ import annotations

HEALTHY_CEILING_TOKENS: int = 200_000


def format_tokens_k(toks: int) -> str:
    """Format token count as human-readable string (e.g., 900.8k, 200k, 1.2M, 450)."""
    if toks >= 1_000_000:
        val = toks / 1_000_000
        return f"{val:.0f}M" if val.is_integer() else f"{val:.1f}M"
    if toks >= 1_000:
        val = toks / 1_000
        return f"{val:.0f}k" if val.is_integer() else f"{val:.1f}k"
    return str(toks)


def get_health_label(toks: int, ceiling: int = HEALTHY_CEILING_TOKENS) -> str:
    """Format granular health label with percentage for bloated/critical against ceiling."""
    if toks > 2 * ceiling:
        pct = round(toks / ceiling * 100)
        return f"CRITICAL ({pct}%)"
    if toks > ceiling:
        pct = round(toks / ceiling * 100)
        return f"BLOATED ({pct}%)"
    if toks >= 100_000:
        return "ELEVATED"
    return "HEALTHY"


def get_base_health_status(toks: int, ceiling: int = HEALTHY_CEILING_TOKENS) -> str:
    """Return base health category without percentage."""
    if toks > 2 * ceiling:
        return "CRITICAL"
    if toks > ceiling:
        return "BLOATED"
    if toks >= 100_000:
        return "ELEVATED"
    return "HEALTHY"

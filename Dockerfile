FROM python:3.14-slim

WORKDIR /app

COPY --from=ghcr.io/astral-sh/uv:latest /uv /bin/uv

COPY pyproject.toml /app/pyproject.toml
COPY README.md /app/README.md
COPY agentic_lens /app/agentic_lens
COPY scripts /app/scripts

RUN uv pip install --system --no-cache /app

EXPOSE 8501

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD python3 -c "import urllib.request; urllib.request.urlopen('http://localhost:8501/_stcore/health')" || exit 1

ENTRYPOINT ["streamlit", "run", "agentic_lens/presentation/app.py", "--server.port=8501", "--server.address=0.0.0.0", "--server.headless=true"]

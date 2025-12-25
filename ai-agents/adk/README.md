
https://google.github.io/adk-docs/get-started/python/

## Install (uv)

If you want to use provider-style models (e.g. `openai/gpt-4o-mini`) with an OpenAI-compatible endpoint, you must have **LiteLLM** installed.

From `adk/`:

```bash
uv pip install -r requirements.txt
```

## OpenAI-compatible endpoint config (for `widget_creator`)

The Python agent in `adk/widget_creator/agent.py` reads these environment variables:

- `OPENAI_API_KEY`: API key/token (some local proxies accept any non-empty value)
- `OPENAI_BASE_URL` (or `OPENAI_API_BASE`): base URL for your OpenAI-compatible endpoint (include `/v1`)
- `OPENAI_MODEL` (or `LLM_MODEL`): model id (defaults to `openai/gpt-4o-mini`)

Example (local OpenAI-compatible server):

```bash
cp adk/widget_creator/example.env adk/widget_creator/.env
# then edit `adk/widget_creator/.env` values as needed
```
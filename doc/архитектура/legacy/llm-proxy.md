# LLM Proxy — 5-й микросервис

> **⛔ УСТАРЁВШАЯ АРХИТЕКТУРА.** Этот файл — исторический артефакт (pre-M10).
> **LLM Proxy удалён.** Функциональность перенесена в `lkfl-server` → `internal/llm/`.
> **Не используйте** описанную здесь архитектуру для новой разработки.
>
> **→ Актуальная реализация:** [`пакеты-platform.md` → internal/llm/](./пакеты-platform.md)
> **→ ADR-022:** Note: M10 — merged into platform
> **→ ADR-024:** Modular Monolith — один бинарник, 17 internal пакетов

> **ADR-022 (исторический):** единая точка доступа ко всем LLM-моделям. Заменяет дублирование prompt/config/model tracking в Platform + Billing. После M10 T1002 функция перенесена in-process.

---

## Роль сервиса

LLM Proxy — **централизованный шлюз** для всех LLM-запросов платформы:

| Что делает | Почему не в Platform/Billing |
|------|------|
| Prompt management | Без дублирования config в каждом сервисе |
| Model routing | Agent → specific model (qwen3.6 для CEL, gpt-4 для moderation) |
| Cost tracking | Централизованный учёт token usage per tenant |
| Rate limiting | Единый limiter на все вызовы |
| Audit trail | ФСТЭК: все LLM-запросы logged в одном месте |
| Failover | Retry + fallback между ollama ↔ openai |

---

## Архитектура

```
┌─ Platform ────────┐    ┌─ Billing ────────┐
│ cel/generator.go  │    │ rule_engine.go   │
│ → LLMProxyClient  │    │ → LLMProxyClient │
│ POST /llm/v1/...  │    │ POST /llm/v1/... │
└─┬───────────────┘    └─┬─────────────┬────┘
  │ HTTP POST                 │ HTTP POST
  └─────────┬─────┬──────────┴──────────┘
            │     │
     ┌──────▼─────▼───────┐
     │   LLM Proxy        │
     │   port: :8085      │
     │                    │
     │ ┌── Agent Router ─┐│
     │ │ agent → model   ││
     │ │ prompt template ││
     │ │ temperature     ││
     │ └─────────────────┘│
     │                    │
     │ ┌── Providers ─────┐│
     │ │ OllamaClient     ││
     │ │ OpenAIClient     ││
     │ └─────────────────┘│
     │                    │
     │ ┌── Audit Log ────┐│
     │ │ PostgreSQL       ││
     │ │ (lkfl_llm)       ││
     │ └─────────────────┘│
     └───┬─────────┬──────┘
         │         │
    ┌────▼──┐  ┌───▼────┐
    │ollama │  │openai  │
    │:11434 │  │api.com │
    └───────┘  └────────┘
```

**Важно (pre-M10):** LLM Proxy — simple HTTP request/response. Не event-driven, не NATS. Вызывается редко (CRUD биллинг-правил, eligibility, recommendations — не hot path).

**M10 T1002:** все функции перенесены in-process как `internal/llm/LLMClient`. Agent router → in-memory, cost tracking → PostgreSQL audit table.

---

## API (pre-M10, reference)

> Следующие эндпоинты описывают pre-M10 REST API :8085. В mono-режиме эквивалент → `internal/llm/LLMClient.GenerateCEL(ctx, sourceText)`.

### POST /llm/v1/generate

Генерация CEL expression из русского текста.

```json
Request
{
    "agent": "cel-generator",
    "source_text": "бесплатный фитнес для директоров и удалённых",
    "tenant_id": "sdek",
    "context_schema": {
        "available_fields": ["grade", "years_of_service", "has_children", "tags.is_remote"]
    }
}

Response 200
{
    "cel_expression": "user.grade == 'Director' || tags.is_remote == true",
    "model": "ollama-qwen3.6:27b",
    "model_version": "v1.0.0",
    "token_usage": { "prompt": 120, "completion": 45 },
    "latency_ms": 340,
    "request_id": "a1b2c3d4-..."
}
```

### POST /llm/v1/validate

Валидация CEL expression (если сервис хочет перепроверить после генерации).

```json
Request
{
    "cel_expression": "user.grade in ['A', 'B']",
    "context_schema": { ... }
}

Response 200
{
    "valid": true,
    "errors": []
}
```

### GET /llm/v1/agents

Список доступных агентов (для Admin UI).

```json
Response 200
{
    "agents": [
        {
            "id": "cel-generator",
            "name": "CEL Expression Generator",
            "model": "ollama-qwen3.6:27b",
            "enabled": true
        },
        {
            "id": "content-moderation",
            "name": "Content Moderation",
            "model": "openai-gpt-4",
            "enabled": false
        }
    ]
}
```

### GET /llm/v1/metrics

Метрики для Grafana.

```json
{
    "total_requests": 1234,
    "total_tokens_used": 567890,
    "cost_today_usd": 1.23,
    "by_agent": {
        "cel-generator": { "requests": 1100, "tokens": 500000 },
        "content-moderation": { "requests": 134, "tokens": 67890 }
    },
    "by_tenant": {
        "sdek": { "requests": 800, "tokens": 340000 },
        "other": { "requests": 434, "tokens": 227890 }
    }
}
```

---

## Agent Router config

```yaml
# LLM Proxy: internal/agents/agent-configs.yaml

agents:
  cel-generator:
    model: "ollama-qwen3.6:27b"
    prompt_template: "templates/cel-generator.yaml"
    max_tokens: 512
    temperature: 0.0         # детерминизм критичен для CEL
    top_p: 0.1
    timeout: 10s
    retry:
      max_attempts: 2
      fallback_model: "openai-gpt-4o-mini"

  content-moderation:    # future
    model: "openai-gpt-4o"
    prompt_template: "templates/content-moderation.yaml"
    max_tokens: 256
    temperature: 0.1
    timeout: 15s

  analytics-summary:     # future
    model: "ollama-mistral:7b"
    prompt_template: "templates/analytics-summary.yaml"
    max_tokens: 2048
    temperature: 0.3
    timeout: 30s
```

---

## Prompt templates

```yaml
# LLM Proxy: templates/cel-generator.yaml

system: |
  You are a CEL (Common Expression Language) expression generator.
  Available context schema (nested layout):
  {{ .ContextSchema }}

  Custom functions:
  - date_diff_days(a, b) → int  // разница дат в днях
  - str_contains(str, substr) → bool

  RULES:
  - Output ONLY a valid CEL expression. No explanations, no quotes, no markdown.
  - Use == for equality, != for inequality
  - Use && for AND, || for OR, ! for NOT
  - Use 'string_literal' for string values (single quotes)
  - Use in for membership: user.grade in ['A', 'B']
  - Use date_diff_days() for date arithmetic
  - If condition is impossible to express, output: ERROR

user_template: |
  {{ .SourceText }}
```

**Важно:** prompt templates — один источник истины. Нигде не дублируются.

---

## DB Schema

```sql
-- lkfl_llm database

CREATE TABLE llm_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID REFERENCES tenants(id) NOT NULL,
    agent_id        VARCHAR(50) NOT NULL,          -- 'cel-generator', 'content-moderation'
    model_used      VARCHAR(100) NOT NULL,          -- 'ollama-qwen3.6:27b'
    model_version   VARCHAR(50) NOT NULL,           -- 'v1.0.0'
    prompt_tokens   INT NOT NULL,
    completion_tokens INT NOT NULL,
    latency_ms      INT NOT NULL,
    status          VARCHAR(20) NOT NULL,           -- 'success', 'error', 'fallback'
    error_message   TEXT,
    source_text     TEXT,                           -- для audit (НЕ хранится в logs)
    response_text   TEXT,                           -- сгенерированный CEL
    request_by      VARCHAR(50),                    -- 'platform', 'billing'
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_llm_requests_tenant ON llm_requests(tenant_id, created_at);
CREATE INDEX idx_llm_requests_agent ON llm_requests(agent_id, created_at);
CREATE INDEX idx_llm_requests_status ON llm_requests(status, created_at);
```

---

## Rate Limiting

| Agent | Limit | Reason |
|-------|-------|--------|
| cel-generator | 30 req/min per tenant | CRUD правил — не frequent |
| content-moderation | 60 req/min per tenant | Real-time moderation при публикации |
| analytics-summary | 10 req/min per tenant | Report generation — rare |

Rate limiting в LLM Proxy, не в вызывающих сервисах.

---

## Cost Tracking

```yaml
# Pricing config: internal/agents/pricing.yaml

pricing:
  ollama-localhost:
    cost_per_token: 0.00           # local, free
  openai-gpt-4o:
    cost_per_input_token: 0.0000025
    cost_per_output_token: 0.00001
  openai-gpt-4o-mini:
    cost_per_input_token: 0.00000015
    cost_per_output_token: 0.0000006
```

Monthly report per tenant: total tokens, cost, breakdown by agent.

---

## Security

- LLM Proxy — internal only, через Nginx, не public
- JWT validation (Keycloak) — API key от сервиса (service-account)
- Prompt injection protection: input sanitization перед отправкой в LLM
- Rate limiting global: 1000 req/min max (защита от abuse)
- LLM responses NOT cached (non-deterministic, но CEL validated после)

---

## Deployment (pre-M10, удалено)

> **M10 T1002:** LLM Proxy как standalone service :8085 удалён. Fункциональность перенесена в `lkfl-server` как `internal/llm/`.
> Docker-compose service `llm-proxy` и Nginx routing `/llm/` → :8085 больше не используются.
> Ollama и OpenAI остаются external dependencies и вызывается через `internal/llm/OllamaClient` / `internal/llm/OpenAIClient` in-process.

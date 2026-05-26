# ADR-022 — LLM Proxy как 5-й микросервис

## Статус

> ⚠️ **M10 T1002: merged into Platform `internal/llm/`.** Standalone LLM Proxy :8085 удалён. В mono-режиме LLM вызывается in-process через `internal/llm/LLMClient.GenerateCEL(ctx, source)`.
> Этот ADR описывает pre-M10 архитектуру отдельного 5-го микросервиса.

## Контекст

ADR-021 ввёл CEL Engine с LLM-генерацией. Изначально предполагалось, что Platform и Billing имеют собственные LLM clients — это создаёт дублирование:

| Ресурс | Platform | Billing | Дублирование? |
|--|--|--|:--:|
| System prompt (CEL schema + rules) | `internal/cel/prompts.go` | `llm/prompts.go` | ✅ |
| CEL validation after LLM response | `internal/cel/validator.go` | `llm/validator.go` | ✅ |
| Model version tracking | `internal/cel/` | `llm/` | ✅ |
| Rate limiting + cost tracking | — | — | ❌ Нет централизованного |
| Unit tests (hallucination cases) | `*_test.go` | `*_test.go` | ✅ |

По мере расширения LLM будет задействован в:
- CEL generation (Phase A)
- Content moderation (будущее: FAQ, баннеры)
- Recommendation personalization (будущее: LLM-generated descriptions)
- Analytics insights (будущее: автоматические report summaries)

Каждое новое применение требует prompt + model selection. Дублирование LLM-client в каждом сервисе не масштабируется.

## Решение

Ввести **LLM Proxy** — 5-й микросервис.

**Назначение:** единая точка доступа ко всем LLM-моделям (ollama, openai, cloud). Принимает запросы от всех сервисов, выбирает модель и prompt по типу агента, применяет rate limiting, tracking costs, audit trail.

**Почему отдельный сервис:**
- Single source of truth: 1 prompt config, 1 model routing, 1 rate limiter
- Cost tracking: централизованный учёт token usage per tenant
- Audit trail: все LLM-запросы logged в одном месте (ФСТЭК)
- Model routing: future — разные агенты → разные модели ( CEL → qwen3.6, moderation → gpt-4, analytics → mistral)
- Failover: LLM Proxy управляет retry + fallback между провайдерами

### Архитектура

```
┌─── Platform ────────┐    ┌─ Billing ────────┐
│                      │    │                  │
│ cel/generator.go     │    │ rule_engine.go   │
│ → LLMProxyClient     │    │ → LLMProxyClient │
│   .GenerateCEL(...)  │    │   .GenerateCEL(.  │
│                      │    │                  │
└───┬───────────┬─────┘    └───┬───────┬──────┘
    │ HTTP POST  │             │ HTTP POST │
    │ /llm/v1/  │             │ /llm/v1/  │
    └────────────┴────────────┴-----------┘
                       │
                ┌──────▼─────────────┐
                │   LLM Proxy        │
                │   (:8085)          │
                │                    │
                │ ┌─ Agent Router ──┐│
                 │ │ model select    ││
                 │ │ prompt mgmt     ││
                 │ │ cost tracking   ││
                 │ └─────────────────┘│
                │                  │
                │ ┌─ Providers ──────┐
                │ │ ollama client    │
                │ │ openai client    │
                │ └──────────────────┘
                └─────┬───────────┬──┘
                      │           │
               ┌──────▼──┐  ┌────▼────┐
               │  ollama │  │ openai  │
               │  local  │  │  cloud  │
               └─────────┘  └────────┘

```

### API

```
POST /llm/v1/generate
{
    "agent": "cel-generator",        // тип агента → модель + prompt
    "source_text": "директорам бесплатно",
    "tenant_id": "sdek",
    "context_schema": { ... }       // optional: preload для лучшего generation
}

Response:
{
    "cel_expression": "user.grade == 'Director'",
    "model": "ollama-qwen3.6",
    "model_version": "v1.0.0",
    "token_usage": { "prompt": 120, "completion": 45 },
    "latency_ms": 340,
    "request_id": "uuid"            // для audit trail
}
```

### Agent Router

```yaml
agents:
  cel-generator:
    model: "ollama-qwen3.6:27b"
    prompt_template: "prompts/cel-generator.yaml"
    max_tokens: 512
    temperature: 0.0               // детерминизм критичен для CEL
    timeout: 10s

  content-moderation:    # future
    model: "openai-gpt-4"
    prompt_template: "prompts/content-moderation.yaml"
    max_tokens: 256
    temperature: 0.1

  analytics-summary:     # future
    model: "ollama-mistral"
    prompt_template: "prompts/analytics.yaml"
    max_tokens: 2048
    temperature: 0.3
```

## Аргументы «за»

- Нет дублирования prompt/validate/model config
- Централизованный cost tracking + rate limiting
- ФСТЭК audit trail в одном месте
- Масштабируемо: новые агенты = новый YAML entry, не код в N сервисов
- Failover: LLM Proxy управляет retry/fallback между провайдерами

## Аргументы «против»

- Extra service = extra infra overhead
- Single point of failure — решается replication + health check
- Latency +1 HTTP hop — не критично: LLM вызывается редко (CRUD правил, не hot path)

## Вердикт

**За.** LLM Proxy — правильный паттерн. Без него дублирование не масштабируется при добавлении новых LLM-агентов.

## Следствия

- Все сервисы (Platform, Billing) используют `LLMProxyClient`, не прямой LLM API
- LLM Proxy: `:8085` (внутренний, через Nginx, не public)
- DB: `lkfl_llm` — audit log, agent config, prompt templates
- NATS subjects: не нужны (LLM Proxy — simple HTTP service, request/response pattern)

## Статус

✅ Accepted

# T1002 — Merge LLM Proxy → Platform internal/llm/

## Веха

M10-рефакторинг-по-результатам-аудита

## Контекст

LLM Proxy — 5-й Go-сервис в архитектуре (ADR-022, :8085, lkfl_llm DB).
На его основе работают:

| Agent | Статус |
|------|--|
| `cel-generator` | ✅ active — единственный consumer |
| `content-moderation` | ❌ future — not started |
| `analytics-summary` | ❌ future — not started |

Сейчас 1 active agent → 1 caller (Platform `cel/Generate()` calls HTTP :8085).

**Проблема:**
1. Overhead 5-го бинарника: Dockerfile, Nginx route (`/llm/v1/`), Keycloak client, go.mod, CI build
2. Отдельная БД `lkfl_llm` — 3 таблицы (llm_requests, agent_configs, prompt_templates) — аудит-лог ведётся редко (CEL gen вызывается при CRUD, не hot path)
3. 1 producer → 1 HTTP call = unnecessary network latency + serialization
4. "Scale when 3+ agents" — premature. Теперь 1 agent, 2 future = не факт. LLM Proxy можно вынести обратно простым рефакторингом (интерфейс `LLMProvider`)

**Решение — встроить в Platform:**
```
platform/
  internal/
    llm/                    # ← было отдельный сервис :8085
      router.go            # agent → model mapping
      providers.go         # Ollama + OpenAI clients
      audit.go             # request logging → Loki logs (не БД)
      rate_limiter.go      # per tenant per agent
```

```
platform cmd/server/main.go
  → internal/cel/generator.go
    → LLMProxyClient.GenerateCEL()
      → internal/llm/router.go
        → internal/llm/providers/ollama.go
```

Audit trail: переехать из `lkfl_llm.llm_requests` → structured JSON logs в Loki. Структура:

```json
{
  "ts": "...", "level": "info", "svc": "platform",
  "msg": "llm.request", "agent": "cel-generator",
  "tenant_id": "sdek", "model": "ollama-qwen3.6:27b",
  "prompt_tokens": 120, "completion_tokens": 45,
  "latency_ms": 340, "request_id": "uuid"
}
```

**Когда выносить обратно:** когда появится 2-й active agent + hot-path usage. Interface `LLMProvider` останется.

### Файлы-мишени

| Действие | Файл |
|---|---|
| Убрать LLM Proxy как сервис | `архитектура/модули.md` — 5 Go сервисов → 4 |
| Встроить в Platform | `архитектура/пакеты-platform.md` — добавить `llm/` как internal пакет |
| Убрать Nginx route `/llm/v1/` | `архитектура/модули.md` — Nginx routes table |
| Убрать Keycloak client | `архитектура/модули.md` — убрать `lkfl-llm-proxy` из clients |
| Убрать БД | `архитектура/модули.md` — lkfl_llm → audit → Loki logs |
| Обновить Redis | `архитектура/модули.md` — Redis DB 4: убрать LLM Proxy rate limiting |
| Обновить ADR-022 | `архитектура/adr/022-llm-proxy-service.md` → superceded by T1002 decision |
| Обновить стек | `архитектура/стек.md` — LLM Proxy не отдельный сервис |
| Обновить nats-subjects.md | `архитектура/nats-subjects.md` — нет NATS для LLM (was: нет, but clarify) |
| Обновить README архитектуры | `архитектура/README.md` — 5 → 4 services |
| Обновить зависимости между сервисами | `архитектура/модули.md` — "Platform + Billing → LLM Proxy" → "internal/llm/" |
| ADR | Создать `архитектура/adr/024-llm-in-process.md` — обоснование merge |
| Update cel-handler | `пакеты-platform.md` — cel_handler.go → internal/llm/ не через HTTP |

### Критерии приёмки

- [ ] `архитектура/модули.md` — 4 Go сервиса (platform, billing, integrations, payment-gateway)
- [ ] `llm/` задокументирован как internal/ пакет с 4 файлами (router, providers, audit, rate_limiter)
- [ ] Audit trail → structured Loki logs (не отдельная БД)
- [ ] Nginx routes: `/llm/v1/` удалён (internal call, не нужен)
- [ ] Keycloak clients: `lkfl-llm-proxy` удалён
- [ ] PostgreSQL: `lkfl_llm` удалена
- [ ] Redis DB 4: LLM Proxy rate limiting удалён (или перенесён в Platform)
- [ ] ADR-022: статус → superceded, ссылка на ADR-024
- [ ] Создан ADR-024: обоснование merge (1 active agent, не hot path, вынести обратно через interface)
- [ ] `архитектура/стек.md` — LLM Proxy не отдельный сервис
- [ ] Dependencies updated: Platform + Billing → internal/llm/ (not :8085)

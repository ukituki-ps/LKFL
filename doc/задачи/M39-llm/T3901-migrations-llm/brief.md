# T3901-T3906 — LLM Engine (ADR-021/022)

## Веха

M39-llm

## T3901 — Migrations: LLM
```sql
CREATE TABLE lkfl_platform.llm_prompts (
    id UUID PK, tenant_id UUID FK, agent VARCHAR(50), name VARCHAR(255),
    template TEXT, version INT, status VARCHAR(20), created_at, updated_at
);
CREATE TABLE lkfl_platform.llm_requests_log (
    id UUID PK, tenant_id UUID FK, agent VARCHAR(50),
    input_text TEXT, output_text TEXT, status VARCHAR(20),
    model VARCHAR(50), input_tokens INT, output_tokens INT,
    cost_usd DECIMAL(10,6), duration_ms INT, created_at
);
CREATE TABLE lkfl_platform.llm_cost_tracking (
    id UUID PK, tenant_id UUID FK, date DATE,
    total_requests INT, total_input_tokens INT, total_output_tokens INT,
    total_cost_usd DECIMAL(10,6), PRIMARY KEY (id), UNIQUE (tenant_id, date)
);
```

## T3902 — internal/llm/ (Engine)
- Agent router (cel-generator, moderation, analytics)
- Prompt management
- Cost tracking
- Audit trail

## T3903 — LLM + CEL integration
- Описание на русском → CEL expression
- Валидация (cel-go AST)
- Fallback на ручной CEL

## T3904 — Admin API
```
POST /admin/cel/generate     — русский текст → CEL
GET  /admin/llm/logs         — лог запросов
GET  /admin/llm/cost         — стоимость по tenant/дате
```

## T3905 — Fallback
- LLM недоступен → ручной ввод CEL (admin textarea)
- Graceful degradation

## T3906 — Prometheus metrics
- `llm_requests_total{agent}`
- `llm_tokens_total{type,model}`
- `llm_cost_today_usd{tenant}`

## Критерии приёмки
- [ ] Все 6 задач
- [ ] Русский → CEL → валидация
- [ ] Cost tracking
- [ ] Fallback работает

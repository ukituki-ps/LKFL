# T1002 — Merge LLM Proxy → Platform internal/llm/ — отчёт

## Статус

✅ завершена

## Что сделано

1. **`doc/архитектура/модули.md`**:
   - Секция "6. LLM Proxy" переписана на "6. LLM — in-process пакет Platform (M10 T1002)"
   - Модули: `router/`, `providers/`, `audit/` + `client.go` (in-process)
   - Коммуникация Billing: HTTP POST /llm/v1/generate → NATS `llm.generate`/`llm.result`
   - Порты: :8085 убран (всё в :8080)
   - DB: lkfl_llm → merged в lkfl_platform
   - Keycloak: отдельный client убран (использует lkfl-platform)
   - Nginx: /llm/v1/ route убран
   - Redis DB 4: LLM Proxy rate limiting → LLM agent configs + rate limiting
   - NATS subjects: добавлен `llm.generate`/`llm.result`
   - Зависимости между сервисами: Platform + Billing → internal/llm/ через NATS
2. **`doc/архитектура/пакеты-platform.md`**:
   - Добавлена секция `internal/llm/` — full описание (client.go, router.go, providers.go, audit.go)
   - `internal/cel/`: LLMProxyClient удалён, зависимости → `internal/llm/`
3. **`doc/архитектура/README.md`**:
   - ASCII-диаграмма: LLM Proxy box убран, Billing → NATS llm.generate → Platform
   - 5 сервисов → 4 сервиса
   - Содержание: llm-proxy.md помечен как "M10 T1002: Исторический ADR"
4. **`doc/архитектура/adr/011-monorepo.md`** — структура: llm-proxy удалён из go modules.

## Что НЕ трогать

- ADR-022 текст — остаётся как историческая справка
- Файлы Go-кода platform/internal/llm/*.go — фактическое создание пакета при реализации
- llm-proxy.md — остаётся как ADR reference

## Проблемы

Нет — задача чистая документация.
